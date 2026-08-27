package shops

import "context"

// TimezoneLookup adapta Service al puerto booking.TimezonePort (HU-061,
// RN-DIS-07): shops nunca importa booking, ni siquiera aquí; este
// adaptador satisface esa interfaz de forma puramente estructural (la
// firma solo usa tipos universales), mismo criterio que
// staff.NewBarberLookup frente a catalog.BarberPort/schedule.BarberPort.
// cmd/api construye este valor y lo pasa donde booking.TimezonePort lo
// espera.
type TimezoneLookup struct {
	service *Service
}

// NewTimezoneLookup construye el adaptador sobre el Service de HU-020 ya
// existente.
func NewTimezoneLookup(service *Service) TimezoneLookup {
	return TimezoneLookup{service: service}
}

// Timezone implementa el método que booking.TimezonePort exige.
func (a TimezoneLookup) Timezone(ctx context.Context, barbershopID string) (string, error) {
	barbershop, err := a.service.Get(ctx, barbershopID)
	if err != nil {
		return "", err
	}
	return barbershop.Timezone, nil
}
