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

// RescheduleAppointmentRequest es la entrada cruda (ya decodificada del
// JSON, todavía sin normalizar) de T2 (HU-065): el barbero autenticado
// mueve una cita `confirmed` a otro inicio. NewStartsAtLocal usa el mismo
// formato civil sin zona que CreateManualAppointmentRequest.StartsAtLocal
// (RN-DIS-07). ExpectedVersionToken es el token opaco que
// booking.EncodeVersionToken produjo para la representación que el cliente
// leyó (HU-064): la precondición de concurrencia optimista de esta
// operación.
type RescheduleAppointmentRequest struct {
	AppointmentID        string
	NewStartsAtLocal     string
	ExpectedVersionToken string
	ActorStaffUserID     string
}

// RescheduleInput es la entrada ya validada y resuelta que llega al
// repositorio: NewEndsAt ya se derivó de duration_minutes_snapshot (nunca
// del cliente ni del catálogo), y ExpectedVersionToken viaja intacto para
// que el repositorio compare dentro de la misma transacción que bloquea la
// fila (§2.4 del prompt de HU-065, "vuelve a verificar estado/versión").
type RescheduleInput struct {
	AppointmentID        string
	NewStartsAt          time.Time
	NewEndsAt            time.Time
	ExpectedVersionToken string
	Actor                Actor
}

// RescheduleAppointment es la representación de la cita ya reprogramada
// (o, en un no-op, la representación vigente sin cambios): deliberadamente
// SIN BarberFullName resuelto (igual que AppointmentDetail desde el
// repositorio, CA-002-06) y con VersionToken en vez de updatedAt crudo,
// mismo criterio que HU-064.
type RescheduleAppointment struct {
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

// RescheduleResult es el desenlace de un intento de reprogramación
// idempotente (RN-IDE-01, DEC-043), mismo criterio que CreateManualResult:
// Decision.Outcome distingue Proceed (aplicada ahora, incluido un no-op del
// mismo intervalo) de Replay (repetición exacta) y cualquier otro desenlace
// que Decision.AsError() ya traduce.
type RescheduleResult struct {
	Decision    idempotency.Decision
	Appointment RescheduleAppointment
	Response    idempotency.StoredResponse
}

// RescheduleService implementa T2 (HU-065): mueve el intervalo de una cita
// `confirmed`, sin cambiar ningún otro dato. Colabora con schedule/shops
// EXCLUSIVAMENTE mediante BlockCheckPort/TimezonePort (mismo criterio que
// ManualBookingService): nunca importa esos paquetes.
type RescheduleService struct {
	repo      Repository
	blocks    BlockCheckPort
	timezones TimezonePort
	clock     clock.Clock
}

// NewRescheduleService construye el caso de uso a partir de sus cuatro
// colaboradores.
func NewRescheduleService(repo Repository, blocks BlockCheckPort, timezones TimezonePort, clk clock.Clock) *RescheduleService {
	return &RescheduleService{repo: repo, blocks: blocks, timezones: timezones, clock: clk}
}

// RescheduleAppointment ejecuta el caso de uso completo. barbershopID ya
// llegó resuelto por el llamador (sesión autenticada); key/fingerprint ya
// fueron interpretados por la capa HTTP. La lectura previa del turno
// (GetAppointmentDetail) solo sirve para resolver BarberID/duración y
// ejecutar el chequeo de bloqueo ANTES de abrir la transacción, igual que
// ManualBookingService consulta catalog/blocks antes de escribir: la
// verificación AUTORITATIVA de estado y versión ocurre dentro de la
// transacción del repositorio, con la fila bloqueada (`FOR UPDATE`), nunca
// aquí.
func (s *RescheduleService) RescheduleAppointment(
	ctx context.Context,
	barbershopID string,
	req RescheduleAppointmentRequest,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (RescheduleResult, error) {
	if err := ctx.Err(); err != nil {
		return RescheduleResult{}, apperr.Internal(fmt.Errorf("booking: contexto cancelado antes de reprogramar: %w", err))
	}

	appointmentID := strings.TrimSpace(req.AppointmentID)
	if !LooksLikeAppointmentID(appointmentID) {
		return RescheduleResult{}, errAppointmentNotFound()
	}

	expectedVersionToken := strings.TrimSpace(req.ExpectedVersionToken)
	if expectedVersionToken == "" {
		return RescheduleResult{}, errVersionTokenRequired()
	}

	if req.ActorStaffUserID == "" {
		return RescheduleResult{}, errActorInvalid()
	}

	timezone, err := s.timezones.Timezone(ctx, barbershopID)
	if err != nil {
		return RescheduleResult{}, err
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return RescheduleResult{}, errTimezoneUnresolvable()
	}

	newStartsAt, err := parseCivilLocal(req.NewStartsAtLocal, loc)
	if err != nil {
		return RescheduleResult{}, err
	}
	if !newStartsAt.After(s.clock.Now()) {
		return RescheduleResult{}, errStartsAtNotFuture()
	}

	detail, found, err := s.repo.GetAppointmentDetail(ctx, barbershopID, appointmentID)
	if err != nil {
		return RescheduleResult{}, apperr.Internal(fmt.Errorf("booking: leer la cita antes de reprogramar: %w", err))
	}
	if !found {
		return RescheduleResult{}, errAppointmentNotFound()
	}

	newEndsAt := newStartsAt.Add(time.Duration(detail.DurationMinutesSnapshot) * time.Minute)

	blocked, err := s.blocks.HasActiveBlock(ctx, barbershopID, detail.BarberID, newStartsAt, newEndsAt, timezone)
	if err != nil {
		return RescheduleResult{}, err
	}
	if blocked {
		// DEC-076: un bloqueo vigente en el nuevo intervalo rechaza T2 con
		// el mismo tratamiento que un cruce de citas, sin importar cuál de
		// los dos aplique primero.
		return RescheduleResult{}, errBlockedInterval()
	}

	actorID := req.ActorStaffUserID
	input := RescheduleInput{
		AppointmentID:        appointmentID,
		NewStartsAt:          newStartsAt,
		NewEndsAt:            newEndsAt,
		ExpectedVersionToken: expectedVersionToken,
		Actor:                Actor{Type: ActorTypeStaff, StaffUserID: &actorID},
	}

	return s.repo.Reschedule(ctx, barbershopID, input, key, fingerprint)
}
