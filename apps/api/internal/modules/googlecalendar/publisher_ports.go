package googlecalendar

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Job es un trabajo reclamado de la cola: «reconcilia este recurso».
type Job struct {
	ID           string
	ClaimToken   string
	BarbershopID string
	ConnectionID string
	ResourceType string
	ResourceID   string
	Attempts     int
}

// JobContext es lo mínimo que el worker necesita saber para reconciliar UN
// trabajo reclamado: el estado actual del recurso, el vínculo y la conexión.
type JobContext struct {
	BarberID         string
	ConnectionStatus string
	CalendarID       string
	ReminderMinutes  *int
	TokenCiphertext  []byte
	TokenKeyID       string
	Timezone         string

	LinkEventID     string
	LinkGeneration  int
	LinkVisibleHash string

	ResourceType  string
	ResourceID    string
	ResourceFound bool
	BarberMatches bool
	ApptStatus    string
	StartsAt      time.Time
	EndsAt        time.Time
	AttendeeName  string
	ServiceName   string
	CustomerEmail string
	BlockType     string
	BlockSource   string
	BlockDeleted  bool
}

// Resultados con los que se finaliza un trabajo.
const (
	OutcomePublished = "published"
	OutcomeRemoved   = "removed"
	OutcomeSkipped   = "skipped"
	OutcomeRetry     = "retry"
	OutcomeReauth    = "reauth"
	OutcomePermanent = "permanent"
)

// FinishInput finaliza un trabajo con CAS sobre su claim_token.
type FinishInput struct {
	JobID        string
	ClaimToken   string
	Outcome      string
	EventID      string
	ETag         string
	Generation   int
	VisibleHash  string
	ErrorCode    string
	RetrySeconds int
	MaxAttempts  int
	Now          time.Time
}

// RestoreTarget es la conexión que el chequeo periódico reservó.
type RestoreTarget struct {
	ConnectionID    string
	BarbershopID    string
	BarberID        string
	CalendarID      string
	TokenCiphertext []byte
	TokenKeyID      string
}

// RestoreLink es un vínculo cuyo recurso sigue vigente y futuro.
type RestoreLink struct {
	ResourceType string
	ResourceID   string
	EventID      string
}

// JobStore es el puerto de persistencia del WORKER: sin contexto de tenant, solo
// funciones de reclamo y finalización (DEC-040, DEC-102). Es distinto de
// Repository, que opera con la sesión de un barbero.
type JobStore interface {
	ClaimJobs(ctx context.Context, limit, leaseSeconds int, now time.Time) ([]Job, error)
	JobContext(ctx context.Context, jobID, claimToken string) (JobContext, bool, error)
	FinishJob(ctx context.Context, in FinishInput) (bool, error)
	RestoreClaimConnection(ctx context.Context, intervalSeconds int, now time.Time) (RestoreTarget, bool, error)
	RestoreLinks(ctx context.Context, connectionID string, now time.Time) ([]RestoreLink, error)
	EnqueueMissing(ctx context.Context, connectionID, resourceType, resourceID string) (bool, error)
}

// Published es lo que Google devuelve al crear o actualizar un evento.
type Published struct {
	EventID string
	ETag    string
}

// EventsAPI es el puerto hacia la API de eventos de Google Calendar. El adaptador
// (google/) lo implementa con HTTP; las pruebas con un Google falso.
type EventsAPI interface {
	// Insert crea el evento con ev.ID. Devuelve un *APIError con Status 409 si ese
	// id ya existe. notify envía la invitación al asistente.
	Insert(ctx context.Context, accessToken, calendarID string, ev Event, notify bool) (Published, error)
	// Patch actualiza el evento al estado deseado. *APIError 404/410 si ya no existe.
	Patch(ctx context.Context, accessToken, calendarID, eventID string, ev Event, notify bool) (Published, error)
	// Delete elimina el evento. 404/410 los trata el llamador como éxito.
	Delete(ctx context.Context, accessToken, calendarID, eventID string, notify bool) error
	// ListEventIDs devuelve los ids de los eventos publicados por NAVA para esa
	// conexión (propiedad privada navaConnectionId) que terminan después de timeMin.
	ListEventIDs(ctx context.Context, accessToken, calendarID, connectionID string, timeMin time.Time) (map[string]bool, error)
}

// APIError es un error de la API de Google reducido a lo que importa para
// decidir: estado HTTP, motivo estándar y espera sugerida. Nunca contiene el
// cuerpo de la respuesta ni el token.
type APIError struct {
	Status     int
	Reason     string
	RetryAfter time.Duration
}

func (e *APIError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("google calendar: estado %d (%s)", e.Status, e.Reason)
	}
	return fmt.Sprintf("google calendar: estado %d", e.Status)
}

func asAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}

// IsGone indica que el evento ya no existe en Google (404 o 410).
func IsGone(err error) bool {
	apiErr, ok := asAPIError(err)
	return ok && (apiErr.Status == 404 || apiErr.Status == 410)
}

// IsConflict indica que el id propuesto ya existe (409).
func IsConflict(err error) bool {
	apiErr, ok := asAPIError(err)
	return ok && apiErr.Status == 409
}
