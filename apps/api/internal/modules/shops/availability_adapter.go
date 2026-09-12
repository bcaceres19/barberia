package shops

import "context"

// AvailabilityBookingPolicy adapta BookingPolicyService al puerto que
// publicbooking.BookingPolicyPort declara para HU-094: mismo criterio
// estructural que TimezoneLookup frente a booking.TimezonePort -shops
// nunca importa publicbooking, y la firma usa solo tipos universales.
type AvailabilityBookingPolicy struct {
	service *BookingPolicyService
}

// NewAvailabilityBookingPolicy construye el adaptador sobre el
// BookingPolicyService ya existente (HU-093).
func NewAvailabilityBookingPolicy(service *BookingPolicyService) AvailabilityBookingPolicy {
	return AvailabilityBookingPolicy{service: service}
}

// BookingPolicy implementa el método que publicbooking.BookingPolicyPort
// exige: reutiliza BookingPolicyService.Get (HU-093, DEC-083) y aplana
// BookingPolicy a los tres enteros que HU-094 necesita (RN-DIS-04,
// RN-DIS-06). El resto de campos (plazo/política de cancelación) no
// participa en el cálculo de disponibilidad.
func (a AvailabilityBookingPolicy) BookingPolicy(ctx context.Context, barbershopID string) (
	minAdvanceMinutes, maxAdvanceDays, slotGridMinutes int, err error,
) {
	policy, err := a.service.Get(ctx, barbershopID)
	if err != nil {
		return 0, 0, 0, err
	}
	return policy.MinAdvanceMinutes, policy.MaxAdvanceDays, policy.SlotGridMinutes, nil
}
