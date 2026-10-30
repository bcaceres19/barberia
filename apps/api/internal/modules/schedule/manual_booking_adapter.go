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

// WithinWorkingHours implementa el método que booking.BlockCheckPort exige
// (CA-061-02, CA-065-04): el intervalo debe caber completo en un segmento de
// la jornada efectiva (Service.ResolveEffectiveDay, CA-041-07). Un segmento
// pertenece a su fecha civil de inicio y, si es nocturno (DEC-020), termina
// al día siguiente; por eso se revisan la fecha de inicio del intervalo y la
// anterior (un turno de madrugada cae en el segmento nocturno de ayer).
func (a ManualBookingBlocks) WithinWorkingHours(ctx context.Context, barbershopID, barberID string, startsAt, endsAt time.Time, timezone string) (bool, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return false, err
	}

	startDay := startsAt.In(loc)
	for _, candidate := range []time.Time{startDay, startDay.AddDate(0, 0, -1)} {
		date := candidate.Format(dateLayout)
		civil, err := time.Parse(dateLayout, date)
		if err != nil {
			return false, err
		}
		day, err := a.service.ResolveEffectiveDay(ctx, barbershopID, barberID, civil)
		if err != nil {
			return false, err
		}
		if !day.IsWorking {
			continue
		}
		for _, seg := range day.Segments {
			segStart, err := parseSegmentStart(date, seg.StartsTime, loc)
			if err != nil {
				return false, err
			}
			segEnd := segStart.Add(time.Duration(seg.DurationMinutes) * time.Minute)
			if !startsAt.Before(segStart) && !endsAt.After(segEnd) {
				return true, nil
			}
		}
	}
	return false, nil
}

// parseSegmentStart acepta la hora civil del segmento con o sin segundos
// ("09:00" o "09:00:00"), según la entregue la jornada efectiva.
func parseSegmentStart(date, startsTime string, loc *time.Location) (time.Time, error) {
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", date+" "+startsTime, loc); err == nil {
		return t, nil
	}
	return time.ParseInLocation("2006-01-02 15:04", date+" "+startsTime, loc)
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
