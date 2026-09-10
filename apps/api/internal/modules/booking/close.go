package booking

import (
	"context"
	"fmt"
	"strings"
	"time"

	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/clock"
	"system-barbershop/internal/platform/idempotency"
)

// CloseAppointmentRequest es la entrada cruda, compartida por T4 manual
// (completar) y T7 (marcar inasistencia) de HU-067: el barbero autenticado
// cierra una cita `confirmed` cuyo starts_at ya pasó. Mismo criterio que
// CancelAppointmentByBarberRequest: sin ningún campo de estado/actor/tiempo
// aceptado desde el cliente (CA-067-*, "el servidor no acepta un estado o
// actor enviado por el body").
type CloseAppointmentRequest struct {
	AppointmentID        string
	ExpectedVersionToken string
	ActorStaffUserID     string
}

// CloseAppointmentInput es la entrada ya validada que llega al repositorio:
// Now es el instante que CompleteAppointmentService/MarkNoShowService
// resolvieron UNA sola vez desde su clock.Clock inyectado (nunca leído por
// el repositorio, que no conoce clock.Clock, CA-002-06); el repositorio lo
// compara contra el starts_at de la fila YA bloqueada (`FOR UPDATE`), nunca
// contra una lectura previa fuera de la transacción.
type CloseAppointmentInput struct {
	AppointmentID        string
	ExpectedVersionToken string
	Actor                Actor
	Now                  time.Time
}

// ClosedAppointment es la representación de la cita ya cerrada (o, en una
// repetición no-op, la representación vigente sin cambios adicionales):
// mismo campo a campo que CancelledAppointment/RescheduleAppointment, con
// VersionToken en vez de updatedAt crudo (HU-064). Compartida por T4 manual
// y T7 porque ambas producen exactamente la misma forma, solo con un
// Status final distinto.
type ClosedAppointment struct {
	ID                       string
	BarberID                 string
	ServiceID                string
	CustomerID               string
	AttendeeName             string
	StartsAt                 time.Time
	EndsAt                   time.Time
	Status                   Status
	Origin                   Origin
	ServiceNameSnapshot      string
	DurationMinutesSnapshot  int
	PriceAmountCentsSnapshot int64
	CurrencySnapshot         string
	CustomerNote             *string
	VersionToken             string
	CreatedAt                time.Time
}

// CompleteAppointmentResult es el desenlace de un intento de cierre como
// `completed` idempotente (RN-IDE-01, DEC-043), mismo criterio que
// CancelAppointmentByBarberResult.
type CompleteAppointmentResult struct {
	Decision    idempotency.Decision
	Appointment ClosedAppointment
	Response    idempotency.StoredResponse
}

// MarkNoShowResult es el desenlace de un intento de cierre como `no_show`
// idempotente (RN-IDE-01, DEC-043), mismo criterio que
// CompleteAppointmentResult.
type MarkNoShowResult struct {
	Decision    idempotency.Decision
	Appointment ClosedAppointment
	Response    idempotency.StoredResponse
}

// validateCloseRequest aplica, para ambos comandos, la validación de forma
// compartida antes de tocar el repositorio (mismo orden que
// CancelAppointmentByBarberService): un fallo aquí nunca reclama una clave
// de idempotencia.
func validateCloseRequest(req CloseAppointmentRequest) (appointmentID, expectedVersionToken string, actor Actor, err error) {
	appointmentID = strings.TrimSpace(req.AppointmentID)
	if !LooksLikeAppointmentID(appointmentID) {
		return "", "", Actor{}, errAppointmentNotFound()
	}

	expectedVersionToken = strings.TrimSpace(req.ExpectedVersionToken)
	if expectedVersionToken == "" {
		return "", "", Actor{}, errVersionTokenRequired()
	}

	if req.ActorStaffUserID == "" {
		return "", "", Actor{}, errActorInvalid()
	}

	actorID := req.ActorStaffUserID
	return appointmentID, expectedVersionToken, Actor{Type: ActorTypeStaff, StaffUserID: &actorID}, nil
}

// CompleteAppointmentService implementa T4 manual (HU-067): cierra una cita
// `confirmed` cuyo starts_at ya pasó como `completed`. clock es el único
// colaborador además del repositorio, inyectado para que las pruebas
// congelen el instante de cierre (docs/03-desarrollo/estandar-backend-go.md).
type CompleteAppointmentService struct {
	repo  Repository
	clock clock.Clock
}

// NewCompleteAppointmentService construye el caso de uso.
func NewCompleteAppointmentService(repo Repository, clk clock.Clock) *CompleteAppointmentService {
	return &CompleteAppointmentService{repo: repo, clock: clk}
}

// CompleteAppointment ejecuta el caso de uso completo. Sin lectura previa
// del turno: la frontera exacta de starts_at (CA-067-03) y el estado se
// verifican AUTORITATIVAMENTE dentro del repositorio, con la fila
// bloqueada (`FOR UPDATE`), nunca aquí -- así un intento prematuro nunca
// toca ninguna fila ni reclama una clave de idempotencia.
func (s *CompleteAppointmentService) CompleteAppointment(
	ctx context.Context,
	barbershopID string,
	req CloseAppointmentRequest,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (CompleteAppointmentResult, error) {
	if err := ctx.Err(); err != nil {
		return CompleteAppointmentResult{}, apperr.Internal(fmt.Errorf("booking: contexto cancelado antes de completar el turno: %w", err))
	}

	appointmentID, expectedVersionToken, actor, err := validateCloseRequest(req)
	if err != nil {
		return CompleteAppointmentResult{}, err
	}

	input := CloseAppointmentInput{
		AppointmentID:        appointmentID,
		ExpectedVersionToken: expectedVersionToken,
		Actor:                actor,
		Now:                  s.clock.Now(),
	}

	return s.repo.CompleteAppointment(ctx, barbershopID, input, key, fingerprint)
}

// MarkNoShowService implementa T7 (HU-067): cierra una cita `confirmed`
// cuyo starts_at ya pasó como `no_show`. Mismo criterio que
// CompleteAppointmentService, sin colaboradores adicionales.
type MarkNoShowService struct {
	repo  Repository
	clock clock.Clock
}

// NewMarkNoShowService construye el caso de uso.
func NewMarkNoShowService(repo Repository, clk clock.Clock) *MarkNoShowService {
	return &MarkNoShowService{repo: repo, clock: clk}
}

// MarkNoShow ejecuta el caso de uso completo. Mismo criterio de
// verificación autoritativa dentro del repositorio que CompleteAppointment.
func (s *MarkNoShowService) MarkNoShow(
	ctx context.Context,
	barbershopID string,
	req CloseAppointmentRequest,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (MarkNoShowResult, error) {
	if err := ctx.Err(); err != nil {
		return MarkNoShowResult{}, apperr.Internal(fmt.Errorf("booking: contexto cancelado antes de marcar inasistencia: %w", err))
	}

	appointmentID, expectedVersionToken, actor, err := validateCloseRequest(req)
	if err != nil {
		return MarkNoShowResult{}, err
	}

	input := CloseAppointmentInput{
		AppointmentID:        appointmentID,
		ExpectedVersionToken: expectedVersionToken,
		Actor:                actor,
		Now:                  s.clock.Now(),
	}

	return s.repo.MarkNoShow(ctx, barbershopID, input, key, fingerprint)
}
