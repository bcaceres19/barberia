package booking

import (
	"context"
	"fmt"
	"strings"
	"time"

	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/idempotency"
)

// CancelAppointmentByBarberRequest es la entrada cruda de T6 (HU-066): el
// barbero autenticado cancela una cita `confirmed` en cualquier momento,
// sin ventana temporal (CA-066-02). ExpectedVersionToken es el token opaco
// que booking.EncodeVersionToken produjo para la representación que el
// cliente leyó (HU-064), mismo criterio que RescheduleAppointmentRequest.
type CancelAppointmentByBarberRequest struct {
	AppointmentID        string
	ExpectedVersionToken string
	ActorStaffUserID     string
}

// CancelAppointmentByBarberInput es la entrada ya validada que llega al
// repositorio: Actor ya se derivó del principal autenticado, nunca del
// cuerpo de la solicitud (CA-066-02, "el servidor no acepta un estado o
// actor enviado por el body").
type CancelAppointmentByBarberInput struct {
	AppointmentID        string
	ExpectedVersionToken string
	Actor                Actor
}

// CancelledAppointment es la representación de la cita ya cancelada (o, en
// una repetición no-op, la representación vigente sin cambios adicionales):
// mismo campo a campo que RescheduleAppointment, con VersionToken en vez de
// updatedAt crudo (HU-064).
type CancelledAppointment struct {
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

// CancelAppointmentByBarberResult es el desenlace de un intento de
// cancelación idempotente (RN-IDE-01, DEC-043), mismo criterio que
// RescheduleResult: Decision.Outcome distingue Proceed (aplicada ahora,
// incluida una cancelación repetida que ya encontró la cita
// `cancelled_by_barber`, CA-066-04) de Replay (repetición exacta por la
// misma Idempotency-Key) y cualquier otro desenlace que Decision.AsError()
// ya traduce.
type CancelAppointmentByBarberResult struct {
	Decision    idempotency.Decision
	Appointment CancelledAppointment
	Response    idempotency.StoredResponse
}

// CancelAppointmentByBarberService implementa T6 (HU-066): cancela una cita
// `confirmed`, sin cambiar ningún otro dato. A diferencia de
// RescheduleService, T6 no tiene ventana temporal ni depende de
// blocks/timezone (CA-066-02: "ningún reloj del cliente impone un límite"),
// así que el único colaborador es el repositorio.
type CancelAppointmentByBarberService struct {
	repo Repository
}

// NewCancelAppointmentByBarberService construye el caso de uso.
func NewCancelAppointmentByBarberService(repo Repository) *CancelAppointmentByBarberService {
	return &CancelAppointmentByBarberService{repo: repo}
}

// CancelAppointmentByBarber ejecuta el caso de uso completo. barbershopID ya
// llegó resuelto por el llamador (sesión autenticada); key/fingerprint ya
// fueron interpretados por la capa HTTP. Sin lectura previa del turno: T6 no
// necesita resolver duración, bloqueos ni zona horaria antes de abrir la
// transacción (a diferencia de T2), así que la verificación AUTORITATIVA de
// estado y versión ocurre íntegramente dentro del repositorio, con la fila
// bloqueada (`FOR UPDATE`).
func (s *CancelAppointmentByBarberService) CancelAppointmentByBarber(
	ctx context.Context,
	barbershopID string,
	req CancelAppointmentByBarberRequest,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (CancelAppointmentByBarberResult, error) {
	if err := ctx.Err(); err != nil {
		return CancelAppointmentByBarberResult{}, apperr.Internal(fmt.Errorf("booking: contexto cancelado antes de cancelar el turno: %w", err))
	}

	appointmentID := strings.TrimSpace(req.AppointmentID)
	if !LooksLikeAppointmentID(appointmentID) {
		return CancelAppointmentByBarberResult{}, errAppointmentNotFound()
	}

	expectedVersionToken := strings.TrimSpace(req.ExpectedVersionToken)
	if expectedVersionToken == "" {
		return CancelAppointmentByBarberResult{}, errVersionTokenRequired()
	}

	if req.ActorStaffUserID == "" {
		return CancelAppointmentByBarberResult{}, errActorInvalid()
	}

	actorID := req.ActorStaffUserID
	input := CancelAppointmentByBarberInput{
		AppointmentID:        appointmentID,
		ExpectedVersionToken: expectedVersionToken,
		Actor:                Actor{Type: ActorTypeStaff, StaffUserID: &actorID},
	}

	return s.repo.CancelByBarber(ctx, barbershopID, input, key, fingerprint)
}
