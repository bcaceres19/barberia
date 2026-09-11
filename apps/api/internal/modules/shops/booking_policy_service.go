package shops

import (
	"context"
	"fmt"

	"system-barbershop/internal/platform/apperr"
)

// BookingPolicyService implementa el caso de uso de HU-093: leer y
// actualizar la configuración pública de reserva y cancelación de la
// barbería activa. Vive en el paquete shops porque shops es el módulo
// dueño de la configuración de la barbería (HU-020); no comparte Service
// (HU-020) para no acoplar la precondición de versión de HU-093 a un
// endpoint que nunca la tuvo.
type BookingPolicyService struct {
	repo BookingPolicyRepository
}

// NewBookingPolicyService construye el servicio.
func NewBookingPolicyService(repo BookingPolicyRepository) *BookingPolicyService {
	return &BookingPolicyService{repo: repo}
}

// Get lee la política de reserva de la barbería barbershopID. barbershopID
// llega siempre de un auth.Principal ya autenticado (CA-093-03); este
// método no acepta ni valida ningún identificador que el cliente pueda
// controlar.
func (s *BookingPolicyService) Get(ctx context.Context, barbershopID string) (BookingPolicy, error) {
	if err := ctx.Err(); err != nil {
		return BookingPolicy{}, apperr.Internal(fmt.Errorf("shops: contexto cancelado antes de leer la política de reserva: %w", err))
	}

	policy, found, err := s.repo.Get(ctx, barbershopID)
	if err != nil {
		return BookingPolicy{}, apperr.Internal(fmt.Errorf("shops: leer política de reserva: %w", err))
	}
	if !found {
		return BookingPolicy{}, apperr.NotFound("barbería no encontrada")
	}
	return policy, nil
}

// BookingPolicyInput es la entrada cruda del caso de uso de actualización,
// tal como llega del handler HTTP: sin validar todavía.
// ExpectedVersionToken llega de la cabecera If-Match (CA-093-02).
type BookingPolicyInput struct {
	MinAdvanceMinutes              int
	MaxAdvanceDays                 int
	SlotGridMinutes                int
	CancellationDeadlineMinutes    int
	LateCancellationClientAllowed  bool
	LateCancellationReasonRequired bool
	ExpectedVersionToken           string
}

// Update valida input contra los rangos de DEC-083 y las combinaciones
// coherentes exigidas por CA-093-02 y, si pasa todas las comprobaciones,
// delega en el repositorio la comparación de versión y la escritura
// atómica de contrato cerrado (PUT: los seis campos siempre presentes,
// nunca una actualización parcial). Los campos se evalúan en un orden
// fijo; el primer error encontrado es el que se devuelve, sin ejecutar
// ninguna escritura (mismo criterio que shops.Service.Update).
func (s *BookingPolicyService) Update(ctx context.Context, barbershopID string, input BookingPolicyInput) (BookingPolicy, error) {
	if err := ctx.Err(); err != nil {
		return BookingPolicy{}, apperr.Internal(fmt.Errorf("shops: contexto cancelado antes de actualizar la política de reserva: %w", err))
	}

	if input.ExpectedVersionToken == "" {
		return BookingPolicy{}, errBookingPolicyVersionTokenRequired()
	}

	switch {
	case input.MinAdvanceMinutes < MinAdvanceMinutesFloor || input.MinAdvanceMinutes > MinAdvanceMinutesCeil:
		return BookingPolicy{}, errBookingPolicyFieldOutOfRange("minAdvanceMinutes")
	case input.MaxAdvanceDays < MaxAdvanceDaysFloor || input.MaxAdvanceDays > MaxAdvanceDaysCeil:
		return BookingPolicy{}, errBookingPolicyFieldOutOfRange("maxAdvanceDays")
	case !IsAllowedSlotGridMinutes(input.SlotGridMinutes):
		return BookingPolicy{}, errBookingPolicyFieldOutOfRange("slotGridMinutes")
	case input.CancellationDeadlineMinutes < CancellationDeadlineMinutesFloor || input.CancellationDeadlineMinutes > CancellationDeadlineMinutesCeil:
		return BookingPolicy{}, errBookingPolicyFieldOutOfRange("cancellationDeadlineMinutes")
	}

	// CA-093-02: combinación incoherente (exigir motivo para una
	// cancelación tardía que el cliente ni siquiera puede hacer).
	if input.LateCancellationReasonRequired && !input.LateCancellationClientAllowed {
		return BookingPolicy{}, errBookingPolicyIncoherentCancellationPolicy()
	}
	// Mismo criterio que barbershop_min_advance_vs_window_ck: la
	// anticipación mínima no puede alcanzar ni superar toda la ventana, o
	// ninguna franja quedaría reservable.
	if input.MinAdvanceMinutes >= input.MaxAdvanceDays*1440 {
		return BookingPolicy{}, errBookingPolicyAdvanceExceedsWindow()
	}

	result, err := s.repo.Update(ctx, barbershopID, BookingPolicyUpdateInput{
		MinAdvanceMinutes:              input.MinAdvanceMinutes,
		MaxAdvanceDays:                 input.MaxAdvanceDays,
		SlotGridMinutes:                input.SlotGridMinutes,
		CancellationDeadlineMinutes:    input.CancellationDeadlineMinutes,
		LateCancellationClientAllowed:  input.LateCancellationClientAllowed,
		LateCancellationReasonRequired: input.LateCancellationReasonRequired,
		ExpectedVersionToken:           input.ExpectedVersionToken,
	})
	if err != nil {
		return BookingPolicy{}, apperr.Internal(fmt.Errorf("shops: actualizar política de reserva: %w", err))
	}
	switch result.Outcome {
	case BookingPolicyUpdateOutcomeVersionConflict:
		return BookingPolicy{}, errBookingPolicyVersionConflict()
	case BookingPolicyUpdateOutcomeNotFound:
		return BookingPolicy{}, apperr.NotFound("barbería no encontrada")
	}
	return result.Policy, nil
}
