package appointmentnotice

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"system-barbershop/internal/platform/clock"
)

// Config agrupa los parámetros del despachador.
type Config struct {
	BatchSize    int
	LeaseSeconds int
	MaxAttempts  int
	PollInterval time.Duration
	SendTimeout  time.Duration
}

// DefaultConfig: hasta 5 intentos (DEC-142.6), lease de 2 minutos y un envío de
// hasta 20 segundos.
func DefaultConfig() Config {
	return Config{BatchSize: 20, LeaseSeconds: 120, MaxAttempts: 5, PollInterval: 15 * time.Second, SendTimeout: 20 * time.Second}
}

// retrySchedule es la espera, en segundos, tras el intento n (1-based): 1, 5,
// 15 y 60 minutos. Más allá del último valor se repite el último.
var retrySchedule = []int{60, 300, 900, 3600}

func retryAfterSeconds(attempts int) int {
	if attempts < 1 {
		attempts = 1
	}
	if attempts > len(retrySchedule) {
		attempts = len(retrySchedule)
	}
	return retrySchedule[attempts-1]
}

// Dispatcher consume la cola de avisos.
type Dispatcher struct {
	store  Store
	mailer Mailer
	clock  clock.Clock
	cfg    Config
	log    *slog.Logger
}

// Deps son las dependencias del despachador.
type Deps struct {
	Store  Store
	Mailer Mailer
	Clock  clock.Clock
	Config Config
	Logger *slog.Logger
}

// NewDispatcher construye el despachador.
func NewDispatcher(deps Deps) *Dispatcher {
	logger := deps.Logger
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	cfg := deps.Config
	if cfg.BatchSize == 0 {
		cfg = DefaultConfig()
	}
	return &Dispatcher{store: deps.Store, mailer: deps.Mailer, clock: deps.Clock, cfg: cfg, log: logger}
}

// Run procesa la cola hasta que ctx termina.
func (d *Dispatcher) Run(ctx context.Context) {
	for {
		processed := d.RunOnce(ctx)
		if ctx.Err() != nil {
			return
		}
		if processed > 0 {
			continue
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(d.cfg.PollInterval):
		}
	}
}

// RunOnce reclama un lote y lo procesa. Devuelve cuántos avisos reclamó.
func (d *Dispatcher) RunOnce(ctx context.Context) int {
	notices, err := d.store.ClaimNotices(ctx, d.cfg.BatchSize, d.cfg.LeaseSeconds, d.cfg.MaxAttempts, d.clock.Now())
	if err != nil {
		d.log.Error("appointment_notice.worker.claim_failed", "error", err)
		return 0
	}
	for _, n := range notices {
		if ctx.Err() != nil {
			return len(notices)
		}
		d.Process(ctx, n)
	}
	return len(notices)
}

// skipReason decide, con el estado VIGENTE del turno, si el aviso ya no debe
// enviarse. Al ser función del estado y no del momento en que se encoló, un
// reintento o una ejecución fuera de orden nunca avisa algo viejo (RN-REC-02,
// RN-REC-03).
func skipReason(c Context, now time.Time) string {
	switch {
	case c.AlreadyCovered:
		return CodeSuperseded
	case c.Kind == KindRescheduled && c.AppointmentStatus != "confirmed":
		return CodeNotConfirmed
	case c.Kind == KindCancelledByBarber && c.AppointmentStatus != "cancelled_by_barber":
		return CodeNotCancelled
	case c.Kind == KindReminder && c.AppointmentStatus != "confirmed":
		// RN-REC-02: un turno cancelado, completado o no asistido no se recuerda.
		return CodeNotConfirmed
	case c.Kind == KindReminder && !c.NoticeStateAt.Equal(c.StartsAt):
		// RN-REC-01/03: el turno cambió de hora; el recordatorio de la hora vieja
		// no se envía (el vigente lo programó la reprogramación).
		return CodeSuperseded
	case !c.StartsAt.After(now):
		return CodePast
	case c.CustomerEmail == "":
		return CodeNoEmail
	}
	return ""
}

// Process envía UN aviso reclamado y lo cierra por CAS sobre su claim_token. Si
// la reclamación se perdió (lease vencido y otro worker lo retomó), no escribe
// nada.
func (d *Dispatcher) Process(ctx context.Context, n Notice) {
	c, found, err := d.store.NoticeContext(ctx, n.ID, n.ClaimToken)
	if err != nil {
		// Sin poder leer el contexto no se cierra: el lease vence y se reintenta.
		d.log.Error("appointment_notice.context_failed", "notice_id", n.ID, "error", err)
		return
	}
	if !found {
		return
	}

	now := d.clock.Now()
	in := FinishInput{NoticeID: n.ID, ClaimToken: n.ClaimToken, MaxAttempts: d.cfg.MaxAttempts, Now: now}

	if reason := skipReason(c, now); reason != "" {
		in.Outcome, in.Code = OutcomeSkipped, reason
		d.finish(ctx, n, in)
		return
	}

	loc, err := time.LoadLocation(c.Timezone)
	if err != nil {
		in.Outcome, in.Code = OutcomePermanent, CodeInvalidTimezone
		d.finish(ctx, n, in)
		return
	}

	sendCtx, cancel := context.WithTimeout(ctx, d.cfg.SendTimeout)
	err = d.mailer.SendChangeNotice(sendCtx, c.Kind, c.CustomerEmail, c.BarbershopName, c.ServiceName, c.AttendeeName,
		c.StartsAt.In(loc).Format("2006-01-02 15:04")+" ("+loc.String()+")")
	cancel()

	switch {
	case err == nil:
		in.Outcome, in.CoversStateAt = OutcomeSent, c.StateAt
		if c.Kind == KindReminder {
			in.CoversStateAt = c.NoticeStateAt
		}
	case isPermanent(err):
		in.Outcome, in.Code = OutcomePermanent, CodeProviderRejected
	default:
		in.Outcome, in.Code, in.RetrySeconds = OutcomeRetry, CodeProviderUnavailable, retryAfterSeconds(n.Attempts)
	}
	d.finish(ctx, n, in)
}

func isPermanent(err error) bool {
	var p interface{ Permanent() bool }
	return errors.As(err, &p) && p.Permanent()
}

func (d *Dispatcher) finish(ctx context.Context, n Notice, in FinishInput) {
	// El resultado se persiste aunque el contexto del aviso ya venciera.
	finishCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	ok, err := d.store.FinishNotice(finishCtx, in)
	if err != nil {
		d.log.Error("appointment_notice.finish_failed", "notice_id", n.ID, "outcome", in.Outcome, "error", err)
		return
	}
	if !ok {
		d.log.Warn("appointment_notice.claim_lost", "notice_id", n.ID)
		return
	}
	// Nunca se registra el correo ni el nombre del cliente (RN-DAT-02).
	d.log.Info("appointment_notice.processed",
		"notice_id", n.ID, "attempt", n.Attempts, "outcome", in.Outcome, "code", in.Code)
}
