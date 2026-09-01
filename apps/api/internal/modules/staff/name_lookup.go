package staff

import (
	"context"

	"system-barbershop/internal/platform/apperr"
)

// BarberNameLookup adapta *Service a un puerto mínimo y estable (booking.
// BarberNamePort, HU-064) para que otro módulo resuelva el nombre visible de
// un barbero sin importar staff ni su repositorio: mismo criterio que
// BarberLookup frente a catalog.BarberPort. staff NUNCA importa booking: la
// composición ocurre en cmd/api.
type BarberNameLookup struct {
	service *Service
}

// NewBarberNameLookup construye el adaptador sobre un *Service ya existente
// (el mismo que BarberLookup y staff/httpapi reutilizan).
func NewBarberNameLookup(service *Service) BarberNameLookup {
	return BarberNameLookup{service: service}
}

// Name informa el nombre visible de un barbero por id, dentro de esa
// barbería. found es false tanto si el barbero no existe como si pertenece
// a otra barbería (RN-TEN-01), mismo criterio que BarberLookup.Exists.
func (l BarberNameLookup) Name(ctx context.Context, barbershopID, barberID string) (string, bool, error) {
	barber, err := l.service.Get(ctx, barbershopID, barberID)
	if err != nil {
		if appErr, ok := apperr.As(err); ok && appErr.Kind == apperr.KindNotFound {
			return "", false, nil
		}
		return "", false, err
	}
	return barber.FullName, true, nil
}
