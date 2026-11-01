package appointmentnotice_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"system-barbershop/internal/modules/appointmentnotice"
)

var now = time.Date(2026, 1, 5, 12, 0, 0, 0, time.UTC)

type fixedClock struct{}

func (fixedClock) Now() time.Time { return now }

type fakeStore struct {
	mu          sync.Mutex
	notices     []appointmentnotice.Notice
	claimErr    error
	ctx         appointmentnotice.Context
	found       bool
	contextErr  error
	finishOK    bool
	finishErr   error
	finishes    []appointmentnotice.FinishInput
	claimedWith [4]int
}

func (s *fakeStore) ClaimNotices(_ context.Context, limit, lease, maxAttempts int, _ time.Time) ([]appointmentnotice.Notice, error) {
	s.claimedWith = [4]int{limit, lease, maxAttempts, 0}
	return s.notices, s.claimErr
}

func (s *fakeStore) NoticeContext(context.Context, string, string) (appointmentnotice.Context, bool, error) {
	return s.ctx, s.found, s.contextErr
}

func (s *fakeStore) FinishNotice(_ context.Context, in appointmentnotice.FinishInput) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.finishes = append(s.finishes, in)
	return s.finishOK, s.finishErr
}

type sentMail struct{ kind, email, shop, service, attendee, when string }

type fakeMailer struct {
	err   error
	calls []sentMail
}

func (m *fakeMailer) SendChangeNotice(_ context.Context, kind, email, shop, service, attendee, when string) error {
	m.calls = append(m.calls, sentMail{kind, email, shop, service, attendee, when})
	return m.err
}

type permanentErr struct{}

func (permanentErr) Error() string   { return "rechazo" }
func (permanentErr) Permanent() bool { return true }

func validContext() appointmentnotice.Context {
	return appointmentnotice.Context{
		Kind:              appointmentnotice.KindRescheduled,
		AppointmentStatus: "confirmed",
		StartsAt:          time.Date(2099, 1, 7, 21, 30, 0, 0, time.UTC), // 16:30 en Bogotá
		StateAt:           time.Date(2026, 1, 5, 11, 59, 0, 0, time.UTC),
		CustomerEmail:     "cliente.a@ejemplo.test",
		AttendeeName:      "Ana Prueba",
		ServiceName:       "Corte clásico",
		BarbershopName:    "Barbería Centro",
		Timezone:          "America/Bogota",
	}
}

func newDispatcher(store *fakeStore, mailer *fakeMailer, logs *bytes.Buffer) *appointmentnotice.Dispatcher {
	return appointmentnotice.NewDispatcher(appointmentnotice.Deps{
		Store: store, Mailer: mailer, Clock: fixedClock{}, Config: appointmentnotice.DefaultConfig(),
		Logger: slog.New(slog.NewTextHandler(logs, &slog.HandlerOptions{Level: slog.LevelDebug})),
	})
}

func process(t *testing.T, store *fakeStore, mailer *fakeMailer, attempts int) (appointmentnotice.FinishInput, string) {
	t.Helper()
	var logs bytes.Buffer
	store.found, store.finishOK = true, true
	newDispatcher(store, mailer, &logs).Process(context.Background(),
		appointmentnotice.Notice{ID: "n1", ClaimToken: "t1", Attempts: attempts})
	if len(store.finishes) != 1 {
		t.Fatalf("debe cerrarse exactamente una vez, se cerró %d", len(store.finishes))
	}
	return store.finishes[0], logs.String()
}

func TestProcess_Sent_BuildsTheMessageFromTheCurrentStateInTheShopTimezone(t *testing.T) {
	store, mailer := &fakeStore{ctx: validContext()}, &fakeMailer{}
	fin, _ := process(t, store, mailer, 1)

	if len(mailer.calls) != 1 {
		t.Fatalf("debe enviar un correo, envió %d", len(mailer.calls))
	}
	got := mailer.calls[0]
	want := sentMail{"rescheduled", "cliente.a@ejemplo.test", "Barbería Centro", "Corte clásico", "Ana Prueba", "2099-01-07 16:30 (America/Bogota)"}
	if got != want {
		t.Fatalf("correo = %+v, quiero %+v", got, want)
	}
	if fin.Outcome != appointmentnotice.OutcomeSent || !fin.CoversStateAt.Equal(validContext().StateAt) {
		t.Fatalf("cierre = %+v, debe ser sent cubriendo el estado leído", fin)
	}
}

func TestProcess_TheMessageFollowsTheStateAtSendTime(t *testing.T) {
	// RN-REC-03: el mismo aviso, reclamado de nuevo tras otro cambio, redacta la hora nueva.
	store, mailer := &fakeStore{ctx: validContext()}, &fakeMailer{}
	process(t, store, mailer, 1)

	moved := validContext()
	moved.StartsAt = moved.StartsAt.Add(90 * time.Minute) // 18:00 en Bogotá
	store2, mailer2 := &fakeStore{ctx: moved}, &fakeMailer{}
	process(t, store2, mailer2, 1)

	if mailer.calls[0].when == mailer2.calls[0].when || !strings.Contains(mailer2.calls[0].when, "18:00") {
		t.Fatalf("la hora debe seguir al estado vigente: %q frente a %q", mailer.calls[0].when, mailer2.calls[0].when)
	}
}

func TestProcess_Skips_WithoutSendingAndWithTheReason(t *testing.T) {
	cancelled := validContext()
	cancelled.Kind, cancelled.AppointmentStatus = appointmentnotice.KindCancelledByBarber, "cancelled_by_barber"

	cases := []struct {
		name   string
		mutate func(c *appointmentnotice.Context)
		base   appointmentnotice.Context
		code   string
	}{
		{"otro aviso ya envió este estado", func(c *appointmentnotice.Context) { c.AlreadyCovered = true }, validContext(), appointmentnotice.CodeSuperseded},
		{"reprogramación de un turno ya cancelado (RN-REC-02)", func(c *appointmentnotice.Context) { c.AppointmentStatus = "cancelled_by_customer" }, validContext(), appointmentnotice.CodeNotConfirmed},
		{"cancelación de un turno que ya no está cancelado por el barbero", func(c *appointmentnotice.Context) { c.AppointmentStatus = "confirmed" }, cancelled, appointmentnotice.CodeNotCancelled},
		{"el turno ya empezó", func(c *appointmentnotice.Context) { c.StartsAt = now.Add(-time.Minute) }, validContext(), appointmentnotice.CodePast},
		{"el turno empieza justo ahora", func(c *appointmentnotice.Context) { c.StartsAt = now }, validContext(), appointmentnotice.CodePast},
		{"cliente sin correo (RN-REC-06)", func(c *appointmentnotice.Context) { c.CustomerEmail = "" }, validContext(), appointmentnotice.CodeNoEmail},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := c.base
			c.mutate(&ctx)
			store, mailer := &fakeStore{ctx: ctx}, &fakeMailer{}
			fin, _ := process(t, store, mailer, 1)
			if len(mailer.calls) != 0 {
				t.Fatalf("no debe enviar, envió %d", len(mailer.calls))
			}
			if fin.Outcome != appointmentnotice.OutcomeSkipped || fin.Code != c.code {
				t.Fatalf("cierre = %s/%s, quiero skipped/%s", fin.Outcome, fin.Code, c.code)
			}
		})
	}
}

func TestProcess_CancellationIsSentWithTheCancelledKind(t *testing.T) {
	ctx := validContext()
	ctx.Kind, ctx.AppointmentStatus = appointmentnotice.KindCancelledByBarber, "cancelled_by_barber"
	store, mailer := &fakeStore{ctx: ctx}, &fakeMailer{}
	fin, _ := process(t, store, mailer, 1)
	if len(mailer.calls) != 1 || mailer.calls[0].kind != "cancelled_by_barber" || fin.Outcome != appointmentnotice.OutcomeSent {
		t.Fatalf("debe enviar la cancelación: %+v / %+v", mailer.calls, fin)
	}
}

func TestProcess_TransientFailure_RetriesWithGrowingWait(t *testing.T) {
	for attempts, wantSeconds := range map[int]int{1: 60, 2: 300, 3: 900, 4: 3600, 9: 3600} {
		store, mailer := &fakeStore{ctx: validContext()}, &fakeMailer{err: errors.New("timeout")}
		fin, _ := process(t, store, mailer, attempts)
		if fin.Outcome != appointmentnotice.OutcomeRetry || fin.Code != appointmentnotice.CodeProviderUnavailable ||
			fin.RetrySeconds != wantSeconds || fin.MaxAttempts != 5 {
			t.Fatalf("intento %d: cierre = %+v, quiero retry con %d s", attempts, fin, wantSeconds)
		}
	}
}

func TestProcess_PermanentRejection_DoesNotRetry(t *testing.T) {
	store, mailer := &fakeStore{ctx: validContext()}, &fakeMailer{err: permanentErr{}}
	fin, _ := process(t, store, mailer, 1)
	if fin.Outcome != appointmentnotice.OutcomePermanent || fin.Code != appointmentnotice.CodeProviderRejected || fin.RetrySeconds != 0 {
		t.Fatalf("cierre = %+v, quiero permanent/provider_rejected", fin)
	}
}

func TestProcess_PermanentRejectionWrapped_IsStillPermanent(t *testing.T) {
	wrapped := errors.Join(errors.New("contexto"), permanentErr{})
	store, mailer := &fakeStore{ctx: validContext()}, &fakeMailer{err: wrapped}
	fin, _ := process(t, store, mailer, 1)
	if fin.Outcome != appointmentnotice.OutcomePermanent {
		t.Fatalf("un rechazo envuelto sigue siendo permanente: %+v", fin)
	}
}

func TestProcess_InvalidTimezone_IsPermanentAndNotSent(t *testing.T) {
	ctx := validContext()
	ctx.Timezone = "Marte/Olimpo"
	store, mailer := &fakeStore{ctx: ctx}, &fakeMailer{}
	fin, _ := process(t, store, mailer, 1)
	if len(mailer.calls) != 0 || fin.Outcome != appointmentnotice.OutcomePermanent || fin.Code != appointmentnotice.CodeInvalidTimezone {
		t.Fatalf("una zona inválida no se reintenta ni se envía: %+v", fin)
	}
}

func TestProcess_LostClaim_DoesNothing(t *testing.T) {
	store, mailer := &fakeStore{ctx: validContext(), found: false}, &fakeMailer{}
	var logs bytes.Buffer
	newDispatcher(store, mailer, &logs).Process(context.Background(), appointmentnotice.Notice{ID: "n1", ClaimToken: "viejo", Attempts: 1})
	if len(mailer.calls) != 0 || len(store.finishes) != 0 {
		t.Fatal("con la reclamación perdida no se envía ni se cierra")
	}
}

func TestProcess_ContextReadFailure_DoesNotSendNorFinish(t *testing.T) {
	store, mailer := &fakeStore{ctx: validContext(), contextErr: errors.New("db caída")}, &fakeMailer{}
	var logs bytes.Buffer
	newDispatcher(store, mailer, &logs).Process(context.Background(), appointmentnotice.Notice{ID: "n1", ClaimToken: "t1", Attempts: 1})
	if len(mailer.calls) != 0 || len(store.finishes) != 0 {
		t.Fatal("sin contexto el lease vence y se reintenta: ni se envía ni se cierra")
	}
}

func TestProcess_LostClaimOnFinish_IsLoggedAsWarning(t *testing.T) {
	store, mailer := &fakeStore{ctx: validContext(), found: true, finishOK: false}, &fakeMailer{}
	var logs bytes.Buffer
	newDispatcher(store, mailer, &logs).Process(context.Background(), appointmentnotice.Notice{ID: "n1", ClaimToken: "t1", Attempts: 1})
	if !strings.Contains(logs.String(), "appointment_notice.claim_lost") {
		t.Fatalf("debe avisar de la reclamación perdida: %s", logs.String())
	}
}

func TestProcess_Logs_NeverContainPersonalData(t *testing.T) {
	for _, mailerErr := range []error{nil, errors.New("timeout"), permanentErr{}} {
		store, mailer := &fakeStore{ctx: validContext()}, &fakeMailer{err: mailerErr}
		_, logs := process(t, store, mailer, 1)
		for _, secret := range []string{"cliente.a@ejemplo.test", "Ana Prueba", "Corte clásico", "Barbería Centro"} {
			if strings.Contains(logs, secret) {
				t.Fatalf("el log no debe contener %q (RN-DAT-02):\n%s", secret, logs)
			}
		}
	}
}

func TestRunOnce_ClaimsWithTheConfiguredLimitsAndProcessesEachNotice(t *testing.T) {
	store := &fakeStore{
		ctx: validContext(), found: true, finishOK: true,
		notices: []appointmentnotice.Notice{{ID: "n1", ClaimToken: "t1", Attempts: 1}, {ID: "n2", ClaimToken: "t2", Attempts: 1}},
	}
	mailer := &fakeMailer{}
	var logs bytes.Buffer
	got := newDispatcher(store, mailer, &logs).RunOnce(context.Background())

	if got != 2 || len(mailer.calls) != 2 || len(store.finishes) != 2 {
		t.Fatalf("RunOnce = %d, correos %d, cierres %d; quiero 2/2/2", got, len(mailer.calls), len(store.finishes))
	}
	cfg := appointmentnotice.DefaultConfig()
	if store.claimedWith != [4]int{cfg.BatchSize, cfg.LeaseSeconds, cfg.MaxAttempts, 0} {
		t.Fatalf("reclamo con %v, quiero los valores por defecto %+v", store.claimedWith, cfg)
	}
}

func TestRunOnce_ClaimFailure_ReturnsZero(t *testing.T) {
	store, mailer := &fakeStore{claimErr: errors.New("db caída")}, &fakeMailer{}
	var logs bytes.Buffer
	if got := newDispatcher(store, mailer, &logs).RunOnce(context.Background()); got != 0 {
		t.Fatalf("RunOnce = %d, quiero 0", got)
	}
}

func TestRunOnce_StopsProcessingWhenTheContextIsCancelled(t *testing.T) {
	store := &fakeStore{
		ctx: validContext(), found: true, finishOK: true,
		notices: []appointmentnotice.Notice{{ID: "n1", ClaimToken: "t1", Attempts: 1}, {ID: "n2", ClaimToken: "t2", Attempts: 1}},
	}
	mailer := &fakeMailer{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var logs bytes.Buffer
	newDispatcher(store, mailer, &logs).RunOnce(ctx)
	if len(mailer.calls) != 0 {
		t.Fatal("con el contexto cancelado no se envía nada; el lease devolverá los avisos")
	}
}

// HU-131 (DEC-144): el recordatorio decide con el estado vigente del turno.
func reminderContext() appointmentnotice.Context {
	c := validContext()
	c.Kind = appointmentnotice.KindReminder
	c.NoticeStateAt = c.StartsAt // programado para la hora vigente
	return c
}

func TestProcess_Reminder_IsSentForTheCurrentHourAndCoversThatInstant(t *testing.T) {
	store, mailer := &fakeStore{ctx: reminderContext()}, &fakeMailer{}
	fin, _ := process(t, store, mailer, 1)

	if len(mailer.calls) != 1 || mailer.calls[0].kind != "reminder" ||
		mailer.calls[0].when != "2099-01-07 16:30 (America/Bogota)" {
		t.Fatalf("debe enviar el recordatorio con la hora vigente: %+v", mailer.calls)
	}
	if fin.Outcome != appointmentnotice.OutcomeSent || !fin.CoversStateAt.Equal(reminderContext().StartsAt) {
		t.Fatalf("un recordatorio cubre el instante del turno, no el updated_at: %+v", fin)
	}
}

func TestProcess_Reminder_Skips(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(c *appointmentnotice.Context)
		code   string
	}{
		{"turno cancelado (RN-REC-02)", func(c *appointmentnotice.Context) { c.AppointmentStatus = "cancelled_by_customer" }, appointmentnotice.CodeNotConfirmed},
		{"turno cancelado por el barbero", func(c *appointmentnotice.Context) { c.AppointmentStatus = "cancelled_by_barber" }, appointmentnotice.CodeNotConfirmed},
		{"turno completado", func(c *appointmentnotice.Context) { c.AppointmentStatus = "completed" }, appointmentnotice.CodeNotConfirmed},
		{"el turno cambió de hora: recordatorio de la hora vieja (RN-REC-03)", func(c *appointmentnotice.Context) { c.NoticeStateAt = c.StartsAt.Add(-3 * time.Hour) }, appointmentnotice.CodeSuperseded},
		{"ya enviado para esa hora", func(c *appointmentnotice.Context) { c.AlreadyCovered = true }, appointmentnotice.CodeSuperseded},
		{"el turno ya empezó", func(c *appointmentnotice.Context) {
			c.StartsAt = now.Add(-time.Minute)
			c.NoticeStateAt = c.StartsAt
		}, appointmentnotice.CodePast},
		{"cliente sin correo", func(c *appointmentnotice.Context) { c.CustomerEmail = "" }, appointmentnotice.CodeNoEmail},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := reminderContext()
			c.mutate(&ctx)
			store, mailer := &fakeStore{ctx: ctx}, &fakeMailer{}
			fin, _ := process(t, store, mailer, 1)
			if len(mailer.calls) != 0 {
				t.Fatalf("no debe enviar, envió %d", len(mailer.calls))
			}
			if fin.Outcome != appointmentnotice.OutcomeSkipped || fin.Code != c.code {
				t.Fatalf("cierre = %s/%s, quiero skipped/%s", fin.Outcome, fin.Code, c.code)
			}
		})
	}
}
