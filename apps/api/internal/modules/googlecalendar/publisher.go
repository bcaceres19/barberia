package googlecalendar

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"system-barbershop/internal/platform/clock"
)

// PublisherConfig son los parámetros del worker de publicación. Valores por
// defecto razonables; ninguna decisión normativa fija las cadencias (DEC-101
// pide documentarlas en la entrega).
type PublisherConfig struct {
	// BatchSize es cuántos trabajos reclama por ciclo.
	BatchSize int
	// LeaseSeconds es la vigencia de una reclamación: si el worker cae, otro
	// retoma el trabajo cuando vence.
	LeaseSeconds int
	// MaxAttempts es el máximo de ejecuciones antes de dejar el trabajo `failed`.
	MaxAttempts int
	// PollInterval es la pausa entre ciclos cuando no hay trabajo.
	PollInterval time.Duration
	// RestoreIntervalSeconds es la cadencia del chequeo de eventos borrados por
	// conexión (DEC-101: «pocos minutos»).
	RestoreIntervalSeconds int
	// JobTimeout acota cada trabajo, para que nunca supere el lease.
	JobTimeout time.Duration
}

// DefaultPublisherConfig devuelve los valores documentados.
func DefaultPublisherConfig() PublisherConfig {
	return PublisherConfig{
		BatchSize:              5,
		LeaseSeconds:           300,
		MaxAttempts:            8,
		PollInterval:           5 * time.Second,
		RestoreIntervalSeconds: 300,
		JobTimeout:             45 * time.Second,
	}
}

const (
	backoffBase = 30 * time.Second
	backoffCap  = time.Hour
	// maxRestorePerCycle acota el chequeo de eventos borrados por ciclo para que
	// nunca retrase la publicación.
	maxRestorePerCycle = 3
	// maxEventGenerations acota cuántas veces se reintenta un id ya usado.
	maxEventGenerations = 4
)

// Backoff es la espera antes de reintentar la ejecución número `attempts`: 30 s
// duplicándose, con tope de 1 h (un trabajo agota 8 intentos en unas dos horas).
func Backoff(attempts int) time.Duration {
	if attempts < 1 {
		attempts = 1
	}
	d := backoffBase
	for i := 1; i < attempts; i++ {
		d *= 2
		if d >= backoffCap {
			return backoffCap
		}
	}
	return d
}

// Publisher reconcilia la cola con Google Calendar: lee el estado ACTUAL de cada
// recurso reclamado y deja al evento igual (DEC-101). Nunca llama a Google dentro
// de una transacción de negocio, y una caída de Google jamás pierde una cita.
type Publisher struct {
	store    JobStore
	api      EventsAPI
	provider OAuthProvider
	cipher   *Cipher
	clock    clock.Clock
	cfg      PublisherConfig
	log      *slog.Logger

	mu     sync.Mutex
	tokens map[string]cachedToken
}

type cachedToken struct {
	value     string
	expiresAt time.Time
}

// PublisherDeps agrupa las dependencias del publicador.
type PublisherDeps struct {
	Store    JobStore
	API      EventsAPI
	Provider OAuthProvider
	Cipher   *Cipher
	Clock    clock.Clock
	Config   PublisherConfig
	Logger   *slog.Logger
}

// NewPublisher construye el publicador.
func NewPublisher(deps PublisherDeps) *Publisher {
	logger := deps.Logger
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	cfg := deps.Config
	if cfg.BatchSize == 0 {
		cfg = DefaultPublisherConfig()
	}
	return &Publisher{
		store: deps.Store, api: deps.API, provider: deps.Provider, cipher: deps.Cipher,
		clock: deps.Clock, cfg: cfg, log: logger, tokens: map[string]cachedToken{},
	}
}

// Run procesa la cola hasta que ctx termina.
func (p *Publisher) Run(ctx context.Context) {
	for {
		processed := p.RunOnce(ctx)
		if ctx.Err() != nil {
			return
		}
		// Con trabajo pendiente se continúa sin esperar; sin él, se duerme.
		if processed > 0 {
			continue
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(p.cfg.PollInterval):
		}
	}
}

// RunOnce reclama un lote, lo procesa y avanza el chequeo de eventos borrados.
// Devuelve cuántos trabajos procesó.
func (p *Publisher) RunOnce(ctx context.Context) int {
	jobs, err := p.store.ClaimJobs(ctx, p.cfg.BatchSize, p.cfg.LeaseSeconds, p.clock.Now())
	if err != nil {
		p.log.Error("google_calendar.worker.claim_failed", "error", err)
		return 0
	}
	for _, job := range jobs {
		if ctx.Err() != nil {
			return len(jobs)
		}
		p.processWithTimeout(ctx, job)
	}
	for range maxRestorePerCycle {
		if ctx.Err() != nil || !p.RestoreOnce(ctx) {
			break
		}
	}
	return len(jobs)
}

func (p *Publisher) processWithTimeout(ctx context.Context, job Job) {
	jobCtx, cancel := context.WithTimeout(ctx, p.cfg.JobTimeout)
	defer cancel()
	p.Process(jobCtx, job)
}

type jobAction int

const (
	actionSkip jobAction = iota
	actionRemove
	actionPublish
)

// decide es la política de DEC-101 sobre el estado ACTUAL del recurso. Al ser
// una función de estado, un reintento o una ejecución fuera de orden nunca
// publica un estado viejo.
func decide(jc JobContext, now time.Time) jobAction {
	hasLink := jc.LinkEventID != ""
	// Un recurso borrado o que ya es de otro barbero deja de ser de esta conexión.
	if !jc.ResourceFound || !jc.BarberMatches {
		return actionRemove
	}
	switch jc.ResourceType {
	case ResourceAppointment:
		switch jc.ApptStatus {
		case "confirmed":
			// Lo pasado no se publica (DEC-101.8); si ya existe, se mantiene al día.
			if !hasLink && !jc.EndsAt.After(now) {
				return actionSkip
			}
			return actionPublish
		case "cancelled_by_customer", "cancelled_by_barber":
			return actionRemove
		default: // completed y no_show conservan el evento tal como está.
			return actionSkip
		}
	case ResourceTimeBlock:
		if jc.BlockDeleted || !IsPublishableBlock(jc.BlockType, jc.BlockSource) {
			return actionRemove
		}
		if !hasLink && !jc.EndsAt.After(now) {
			return actionSkip
		}
		return actionPublish
	}
	return actionSkip
}

// errReauth indica que Google ya no acepta el permiso de la conexión.
var errReauth = errors.New("googlecalendar: se requiere volver a autorizar")

// Process reconcilia un trabajo reclamado y lo finaliza por CAS sobre su
// claim_token. Si la reclamación se perdió (lease vencido y otro worker lo
// retomó), no escribe nada.
func (p *Publisher) Process(ctx context.Context, job Job) {
	jc, found, err := p.store.JobContext(ctx, job.ID, job.ClaimToken)
	if err != nil {
		// Sin poder leer el contexto no se finaliza: el lease vence y se reintenta.
		p.log.Error("google_calendar.job.context_failed", "job_id", job.ID, "error", err)
		return
	}
	if !found {
		return // la reclamación ya no es nuestra
	}

	finish := FinishInput{JobID: job.ID, ClaimToken: job.ClaimToken, MaxAttempts: p.cfg.MaxAttempts, Now: p.clock.Now()}

	if jc.ConnectionStatus != string(StatusConnected) && jc.ConnectionStatus != string(StatusError) {
		finish.Outcome = OutcomeSkipped
		p.finish(ctx, job, finish)
		return
	}

	switch decide(jc, p.clock.Now()) {
	case actionSkip:
		finish.Outcome = OutcomeSkipped
	case actionRemove:
		finish = p.remove(ctx, job, jc, finish)
	case actionPublish:
		finish = p.publish(ctx, job, jc, finish)
	}
	p.finish(ctx, job, finish)
}

func (p *Publisher) finish(ctx context.Context, job Job, in FinishInput) {
	// El resultado se persiste aunque el contexto del trabajo ya venciera.
	finishCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	ok, err := p.store.FinishJob(finishCtx, in)
	if err != nil {
		p.log.Error("google_calendar.job.finish_failed", "job_id", job.ID, "outcome", in.Outcome, "error", err)
		return
	}
	if !ok {
		p.log.Warn("google_calendar.job.claim_lost", "job_id", job.ID)
		return
	}
	p.log.Info("google_calendar.event.reconciled",
		"job_id", job.ID, "connection_id", job.ConnectionID, "resource_type", job.ResourceType,
		"resource_id", job.ResourceID, "outcome", in.Outcome, "attempt", job.Attempts)
}

func (p *Publisher) remove(ctx context.Context, job Job, jc JobContext, finish FinishInput) FinishInput {
	if jc.LinkEventID == "" {
		finish.Outcome = OutcomeSkipped
		return finish
	}
	// El cliente recibe la cancelación de Google solo si se le había invitado.
	notify := jc.ResourceType == ResourceAppointment && jc.CustomerEmail != ""
	err := p.withToken(ctx, job, jc, func(token string) error {
		err := p.api.Delete(ctx, token, jc.CalendarID, jc.LinkEventID, notify)
		if IsGone(err) {
			return nil // ya no existe: es lo que se buscaba
		}
		return err
	})
	if err != nil {
		return p.failure(finish, job, err)
	}
	finish.Outcome = OutcomeRemoved
	return finish
}

func (p *Publisher) publish(ctx context.Context, job Job, jc JobContext, finish FinishInput) FinishInput {
	ev, hash := BuildEvent(jc, job.ConnectionID)

	var published Published
	generation := jc.LinkGeneration
	err := p.withToken(ctx, job, jc, func(token string) error {
		var err error
		published, generation, err = p.upsert(ctx, token, job, jc, ev, hash)
		return err
	})
	if err != nil {
		return p.failure(finish, job, err)
	}
	finish.Outcome = OutcomePublished
	finish.EventID = published.EventID
	finish.ETag = published.ETag
	finish.Generation = generation
	finish.VisibleHash = hash
	return finish
}

// upsert actualiza el evento vinculado o lo crea. El vínculo persistido decide:
// reintentar nunca crea un segundo evento.
func (p *Publisher) upsert(ctx context.Context, token string, job Job, jc JobContext, ev Event, hash string) (Published, int, error) {
	if jc.LinkEventID != "" {
		// Solo se avisa al invitado si cambió algo que él ve.
		notify := ev.AttendeeEmail != "" && hash != jc.LinkVisibleHash
		published, err := p.api.Patch(ctx, token, jc.CalendarID, jc.LinkEventID, ev, notify)
		if err == nil {
			return published, jc.LinkGeneration, nil
		}
		if !IsGone(err) {
			return Published{}, 0, err
		}
		// El barbero lo borró en Google: se recrea con una generación nueva.
		return p.create(ctx, token, job, jc, ev, jc.LinkGeneration+1)
	}
	return p.create(ctx, token, job, jc, ev, 1)
}

func (p *Publisher) create(ctx context.Context, token string, job Job, jc JobContext, ev Event, generation int) (Published, int, error) {
	notify := ev.AttendeeEmail != ""
	for range maxEventGenerations {
		ev.ID = EventID(job.ConnectionID, job.ResourceType, job.ResourceID, generation)
		published, err := p.api.Insert(ctx, token, jc.CalendarID, ev, notify)
		if err == nil {
			return published, generation, nil
		}
		if !IsConflict(err) {
			return Published{}, 0, err
		}
		// El id ya existe: o es un reintento tras caerse entre crear y guardar el
		// vínculo (se adopta el evento), o es el de uno borrado (otra generación).
		adopted := ev
		adopted.ID = ""
		published, patchErr := p.api.Patch(ctx, token, jc.CalendarID, ev.ID, adopted, notify)
		if patchErr == nil {
			return published, generation, nil
		}
		if !IsGone(patchErr) {
			return Published{}, 0, patchErr
		}
		generation++
	}
	return Published{}, 0, &APIError{Status: 409, Reason: "event_id_exhausted"}
}

// failure traduce un error a la finalización correspondiente.
func (p *Publisher) failure(finish FinishInput, job Job, err error) FinishInput {
	if errors.Is(err, errReauth) || errors.Is(err, ErrTokenRevoked) {
		finish.Outcome = OutcomeReauth
		finish.ErrorCode = "token_revoked"
		return finish
	}
	if apiErr, ok := asAPIError(err); ok {
		switch {
		case apiErr.Status == 401:
			finish.Outcome = OutcomeReauth
			finish.ErrorCode = "token_revoked"
			return finish
		case apiErr.Status == 403 && isRateLimit(apiErr.Reason), apiErr.Status == 429, apiErr.Status >= 500:
			return p.retry(finish, job, "google_unavailable", apiErr.RetryAfter)
		case apiErr.Status == 403:
			finish.Outcome = OutcomePermanent
			finish.ErrorCode = "permission_denied"
			return finish
		case apiErr.Status == 404:
			// Con el calendario inexistente (el evento ya se trató aparte).
			finish.Outcome = OutcomePermanent
			finish.ErrorCode = "calendar_not_found"
			return finish
		case apiErr.Status == 400 || apiErr.Status == 409:
			finish.Outcome = OutcomePermanent
			finish.ErrorCode = "invalid_event"
			return finish
		}
		return p.retry(finish, job, "google_error", apiErr.RetryAfter)
	}
	// Red caída, tiempo agotado, no se pudo descifrar...: transitorio.
	return p.retry(finish, job, "transient", 0)
}

func (p *Publisher) retry(finish FinishInput, job Job, code string, retryAfter time.Duration) FinishInput {
	wait := Backoff(job.Attempts)
	if retryAfter > wait {
		wait = retryAfter
	}
	if wait > backoffCap {
		wait = backoffCap
	}
	finish.Outcome = OutcomeRetry
	finish.ErrorCode = code
	finish.RetrySeconds = int(wait.Seconds())
	return finish
}

func isRateLimit(reason string) bool {
	switch reason {
	case "rateLimitExceeded", "userRateLimitExceeded", "quotaExceeded", "dailyLimitExceeded":
		return true
	}
	return false
}

// withToken ejecuta fn con un access token vigente de la conexión. Un 401 fuerza
// UNA renovación y reintenta; si persiste, el permiso ya no sirve.
func (p *Publisher) withToken(ctx context.Context, job Job, jc JobContext, fn func(token string) error) error {
	token, err := p.accessToken(ctx, job, jc, false)
	if err != nil {
		return err
	}
	err = fn(token)
	if apiErr, ok := asAPIError(err); ok && apiErr.Status == 401 {
		if token, err = p.accessToken(ctx, job, jc, true); err != nil {
			return err
		}
		return fn(token)
	}
	return err
}

func (p *Publisher) accessToken(ctx context.Context, job Job, jc JobContext, force bool) (string, error) {
	return p.tokenFor(ctx, job.ConnectionID, job.BarbershopID, jc.BarberID, jc.TokenCiphertext, jc.TokenKeyID, force)
}

// tokenFor devuelve un access token de la conexión, con caché en memoria hasta
// poco antes de su vencimiento: nunca se persiste (DEC-102).
func (p *Publisher) tokenFor(ctx context.Context, connectionID, barbershopID, barberID string, ciphertext []byte, keyID string, force bool) (string, error) {
	now := p.clock.Now()
	p.mu.Lock()
	cached, ok := p.tokens[connectionID]
	p.mu.Unlock()
	if ok && !force && cached.expiresAt.After(now.Add(time.Minute)) {
		return cached.value, nil
	}
	if len(ciphertext) == 0 {
		return "", errReauth
	}
	refresh, err := p.cipher.Decrypt(ciphertext, keyID, tokenAAD(barbershopID, barberID))
	if err != nil {
		return "", fmt.Errorf("googlecalendar: descifrar refresh token: %w", err)
	}
	access, expiresAt, err := p.provider.RefreshAccessToken(ctx, string(refresh))
	if err != nil {
		return "", err // ErrTokenRevoked o transitorio
	}
	p.mu.Lock()
	p.tokens[connectionID] = cachedToken{value: access, expiresAt: expiresAt}
	p.mu.Unlock()
	return access, nil
}

// RestoreOnce reserva una conexión vencida y recrea los eventos que el barbero
// borró por error en Google (DEC-101.4): lista SOLO los eventos publicados por
// NAVA (propiedad privada navaConnectionId, sin syncToken ni webhook) y reencola
// los vínculos vigentes cuyo evento falta. Nunca modifica el dominio de NAVA ni
// restaura lo cancelado, terminal o pasado (RestoreLinks ya lo filtra). Devuelve
// false cuando no había ninguna conexión pendiente de comprobar.
func (p *Publisher) RestoreOnce(ctx context.Context) bool {
	now := p.clock.Now()
	target, found, err := p.store.RestoreClaimConnection(ctx, p.cfg.RestoreIntervalSeconds, now)
	if err != nil {
		p.log.Error("google_calendar.restore.claim_failed", "error", err)
		return false
	}
	if !found {
		return false
	}
	restoreCtx, cancel := context.WithTimeout(ctx, p.cfg.JobTimeout)
	defer cancel()

	token, err := p.tokenFor(restoreCtx, target.ConnectionID, target.BarbershopID, target.BarberID,
		target.TokenCiphertext, target.TokenKeyID, false)
	if err != nil {
		// Un permiso revocado lo descubre y gestiona el siguiente trabajo de publicación.
		p.log.Warn("google_calendar.restore.no_token", "connection_id", target.ConnectionID)
		return true
	}
	existing, err := p.api.ListEventIDs(restoreCtx, token, target.CalendarID, target.ConnectionID, now)
	if err != nil {
		p.log.Warn("google_calendar.restore.list_failed", "connection_id", target.ConnectionID, "error", err)
		return true
	}
	links, err := p.store.RestoreLinks(restoreCtx, target.ConnectionID, now)
	if err != nil {
		p.log.Error("google_calendar.restore.links_failed", "connection_id", target.ConnectionID, "error", err)
		return true
	}
	restored := 0
	for _, link := range links {
		if existing[link.EventID] {
			continue
		}
		queued, err := p.store.EnqueueMissing(restoreCtx, target.ConnectionID, link.ResourceType, link.ResourceID)
		if err != nil {
			p.log.Error("google_calendar.restore.enqueue_failed", "connection_id", target.ConnectionID, "error", err)
			continue
		}
		if queued {
			restored++
		}
	}
	if restored > 0 {
		p.log.Info("google_calendar.restore.requeued", "connection_id", target.ConnectionID, "events", restored)
	}
	return true
}
