package schedule

import (
	"context"
	"time"
)

// ManualBookingBlocks adapta Service al puerto booking.BlockCheckPort
// (HU-061, DEC-073): schedule nunca importa booking, ni siquiera aquí; este
// adaptador satisface esa interfaz de forma puramente estructural (todos
// los tipos de la firma son universales), mismo criterio que
// catalog.ManualBookingCatalog frente a booking.BarberServicePort. cmd/api
// construye este valor y lo pasa donde booking.BlockCheckPort lo espera.
type ManualBookingBlocks struct {
	service *Service
}

// NewManualBookingBlocks construye el adaptador sobre el Service de
// casos de uso ya existente (HU-042).
func NewManualBookingBlocks(service *Service) ManualBookingBlocks {
	return ManualBookingBlocks{service: service}
}

// HasActiveBlock implementa el método que booking.BlockCheckPort exige:
// proyecta los bloqueos puntuales y las ocurrencias de serie vigentes
// dentro de la fecha civil (o par de fechas civiles, si el intervalo cruza
// medianoche local) que corresponde a [startsAt, endsAt) en timezone,
// reutilizando Service.EffectiveBlocks (CA-042), y compara cada uno contra
// el intervalo semiabierto pedido (RN-DIS-05).
func (a ManualBookingBlocks) HasActiveBlock(ctx context.Context, barbershopID, barberID string, startsAt, endsAt time.Time, timezone string) (bool, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return false, err
	}

	fromDate := startsAt.In(loc).Format(dateLayout)
	toDate := endsAt.Add(-time.Nanosecond).In(loc).Format(dateLayout)
	if toDate < fromDate {
		toDate = fromDate
	}

	result, err := a.service.EffectiveBlocks(ctx, barbershopID, barberID, fromDate, toDate)
	if err != nil {
		return false, err
	}

	for _, b := range result.ManualBlocks {
		if startsAt.Before(b.EndsAt) && b.StartsAt.Before(endsAt) {
			return true, nil
		}
	}

	for _, occ := range result.SeriesOccurrences {
		occStart, err := time.ParseInLocation("2006-01-02 15:04", occ.Date+" "+occ.StartsTime, loc)
		if err != nil {
			return false, err
		}
		occEnd := occStart.Add(time.Duration(occ.DurationMinutes) * time.Minute)
		if startsAt.Before(occEnd) && occStart.Before(endsAt) {
			return true, nil
		}
	}

	return false, nil
}
