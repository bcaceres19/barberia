package catalog

import (
	"context"

	"system-barbershop/internal/platform/apperr"
)

// ManualBookingCatalog adapta CatalogService y AssignmentService al puerto
// booking.BarberServicePort (HU-061, DEC-072): catalog nunca importa
// booking, ni siquiera aquí; este adaptador satisface esa interfaz de
// forma puramente estructural (todos los tipos de la firma son
// universales), exactamente igual que staff.NewBarberLookup frente a
// catalog.BarberPort. cmd/api (la raíz de composición) construye este
// valor y lo pasa donde booking.BarberServicePort lo espera.
type ManualBookingCatalog struct {
	services    *CatalogService
	assignments *AssignmentService
}

// NewManualBookingCatalog construye el adaptador sobre los dos servicios de
// casos de uso ya existentes.
func NewManualBookingCatalog(services *CatalogService, assignments *AssignmentService) ManualBookingCatalog {
	return ManualBookingCatalog{services: services, assignments: assignments}
}

// ActiveAssignedService implementa el método que booking.BarberServicePort
// exige: found=false cubre servicio inexistente, inactivo o no asignado a
// barberID, sin distinguir la causa (RN-TEN-01, DEC-072). No verifica
// barberID por separado: una asignación solo puede existir para un barbero
// real de esta barbería (barber_service_barbershop_id_barber_id_fk),
// así que "no asignado" ya cubre implícitamente "barbero inexistente".
func (a ManualBookingCatalog) ActiveAssignedService(ctx context.Context, barbershopID, barberID, serviceID string) (
	name string, durationMinutes int, priceAmountCents int64, currency string, found bool, err error,
) {
	if !LooksLikeServiceID(serviceID) {
		return "", 0, 0, "", false, nil
	}

	assigned, err := a.assignments.IsAssigned(ctx, barbershopID, barberID, serviceID)
	if err != nil {
		return "", 0, 0, "", false, err
	}
	if !assigned {
		return "", 0, 0, "", false, nil
	}

	service, err := a.services.Get(ctx, barbershopID, serviceID)
	if err != nil {
		if appErr, ok := apperr.As(err); ok && appErr.Kind == apperr.KindNotFound {
			return "", 0, 0, "", false, nil
		}
		return "", 0, 0, "", false, err
	}
	if !service.IsActive {
		return "", 0, 0, "", false, nil
	}

	return service.Name, service.DurationMinutes, service.PriceCents, service.Currency, true, nil
}
