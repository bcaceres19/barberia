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

// correctionReasonMaxLength refleja EXACTAMENTE el límite de
// appointment_history_reason_ck (database/migrations/20260827110000_create_appointment_core.sql):
// 500 caracteres para el motivo de appointment_status_corrected.
const correctionReasonMaxLength = 500

// CorrectAppointmentStatusRequest es la entrada cruda de T8 (HU-068): el
// barbero autenticado corrige una clasificación terminal errónea hacia otra
// terminal distinta, con motivo obligatorio. A diferencia de T6/T7, el
// cuerpo SÍ lleva campos (DestinationStatus, Reason): sin ellos no hay forma
// de expresar hacia qué terminal corregir ni por qué.
type CorrectAppointmentStatusRequest struct {
	AppointmentID        string
	DestinationStatus    string
	Reason               string
	ExpectedVersionToken string
	ActorStaffUserID     string
}

// CorrectAppointmentStatusInput es la entrada ya validada que llega al
// repositorio: DestinationStatus ya es un Status cerrado (nunca `confirmed`,
// CA-068-02), Reason ya viene recortado y no vacío, y Now es el instante que
// CorrectAppointmentStatusService resolvió UNA sola vez desde su
// clock.Clock inyectado (nunca leído por el repositorio), mismo criterio que
// CloseAppointmentInput.
type CorrectAppointmentStatusInput struct {
	AppointmentID        string
	DestinationStatus    Status
	Reason               string
	ExpectedVersionToken string
	Actor                Actor
	Now                  time.Time
}

// CorrectedAppointment es la representación de la cita ya corregida (o, en
// una repetición no-op hacia el mismo estado, la representación vigente sin
// cambios adicionales): mismo campo a campo que ClosedAppointment/
// CancelledAppointment.
type CorrectedAppointment struct {
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

// CorrectAppointmentStatusResult es el desenlace de un intento de
// corrección idempotente (RN-IDE-01, DEC-043), mismo criterio que
// CompleteAppointmentResult.
type CorrectAppointmentStatusResult struct {
	Decision    idempotency.Decision
	Appointment CorrectedAppointment
	Response    idempotency.StoredResponse
}

// parseCorrectionDestination reconoce únicamente los cuatro terminales
// válidos como destino de T8 (CA-068-02): `confirmed` y cualquier otro
// valor desconocido se rechazan aquí, antes de tocar el repositorio.
func parseCorrectionDestination(raw string) (Status, bool) {
	switch Status(raw) {
	case StatusCompleted, StatusNoShow, StatusCancelledByCustomer, StatusCancelledByBarber:
		return Status(raw), true
	default:
		return "", false
	}
}

// CorrectAppointmentStatusService implementa T8 (HU-068): corrige una cita
// que ya tiene un resultado terminal hacia otro terminal distinto, con
// motivo obligatorio. clock es el único colaborador además del repositorio,
// inyectado para que las pruebas congelen el instante de corrección (mismo
// criterio que CompleteAppointmentService/MarkNoShowService, necesario aquí
// porque una corrección hacia `completed`/`no_show` revalida la misma
// frontera temporal que T4/T7, CA-068-04).
type CorrectAppointmentStatusService struct {
	repo  Repository
	clock clock.Clock
}

// NewCorrectAppointmentStatusService construye el caso de uso.
func NewCorrectAppointmentStatusService(repo Repository, clk clock.Clock) *CorrectAppointmentStatusService {
	return &CorrectAppointmentStatusService{repo: repo, clock: clk}
}

// CorrectAppointmentStatus ejecuta el caso de uso completo. Sin lectura
// previa del turno: origen, destino y frontera temporal se verifican
// AUTORITATIVAMENTE dentro del repositorio, con la fila bloqueada
// (`FOR UPDATE`), nunca aquí — mismo criterio que CompleteAppointment.
func (s *CorrectAppointmentStatusService) CorrectAppointmentStatus(
	ctx context.Context,
	barbershopID string,
	req CorrectAppointmentStatusRequest,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (CorrectAppointmentStatusResult, error) {
	if err := ctx.Err(); err != nil {
		return CorrectAppointmentStatusResult{}, apperr.Internal(fmt.Errorf("booking: contexto cancelado antes de corregir el turno: %w", err))
	}

	appointmentID := strings.TrimSpace(req.AppointmentID)
	if !LooksLikeAppointmentID(appointmentID) {
		return CorrectAppointmentStatusResult{}, errAppointmentNotFound()
	}

	destination, ok := parseCorrectionDestination(req.DestinationStatus)
	if !ok {
		return CorrectAppointmentStatusResult{}, errCorrectionDestinationInvalid()
	}

	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return CorrectAppointmentStatusResult{}, errCorrectionReasonRequired()
	}
	if len(reason) > correctionReasonMaxLength {
		return CorrectAppointmentStatusResult{}, errCorrectionReasonTooLong()
	}

	expectedVersionToken := strings.TrimSpace(req.ExpectedVersionToken)
	if expectedVersionToken == "" {
		return CorrectAppointmentStatusResult{}, errVersionTokenRequired()
	}

	if req.ActorStaffUserID == "" {
		return CorrectAppointmentStatusResult{}, errActorInvalid()
	}

	actorID := req.ActorStaffUserID
	input := CorrectAppointmentStatusInput{
		AppointmentID:        appointmentID,
		DestinationStatus:    destination,
		Reason:               reason,
		ExpectedVersionToken: expectedVersionToken,
		Actor:                Actor{Type: ActorTypeStaff, StaffUserID: &actorID},
		Now:                  s.clock.Now(),
	}

	return s.repo.CorrectAppointmentStatus(ctx, barbershopID, input, key, fingerprint)
}
