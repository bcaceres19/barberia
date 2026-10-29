package booking

import (
	"context"

	"system-barbershop/internal/platform/idempotency"
)

// CustomerAppointmentCancellerAdapter satisface, de forma puramente
// estructural, customeraccess.AppointmentCanceller (HU-099): traduce los tipos
// universales de ese puerto hacia CancelAppointmentByCustomerRequest. Mismo
// criterio que PublicAppointmentAdapter: customeraccess nunca importa booking
// (CA-002-06); el adaptador vive aquí porque booking sí puede conocer los tipos
// universales que el puerto expone.
type CustomerAppointmentCancellerAdapter struct {
	service *CancelAppointmentByCustomerService
}

// NewCustomerAppointmentCancellerAdapter construye el adaptador.
func NewCustomerAppointmentCancellerAdapter(service *CancelAppointmentByCustomerService) CustomerAppointmentCancellerAdapter {
	return CustomerAppointmentCancellerAdapter{service: service}
}

// CancelAppointment implementa customeraccess.AppointmentCanceller.
func (a CustomerAppointmentCancellerAdapter) CancelAppointment(
	ctx context.Context,
	barbershopID, appointmentID, expectedVersionToken, reason string,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (idempotency.Decision, idempotency.StoredResponse, error) {
	result, err := a.service.CancelAppointmentByCustomer(ctx, barbershopID, CancelAppointmentByCustomerRequest{
		AppointmentID:        appointmentID,
		ExpectedVersionToken: expectedVersionToken,
		Reason:               reason,
	}, key, fingerprint)
	if err != nil {
		return idempotency.Decision{}, idempotency.StoredResponse{}, err
	}
	return result.Decision, result.Response, nil
}
