package schedule

import (
	"context"
	"time"
)

// AvailabilityLookup adapta Service a los puertos que
// publicbooking.Service declara para HU-094 (EffectiveDayPort,
// BusyBlocksPort): mismo criterio estructural que ManualBookingBlocks
// (HU-061, DEC-073) frente a booking.BlockCheckPort -schedule nunca importa
// publicbooking, y las firmas usan solo tipos universales para que este
// adaptador satisfaga esos puertos sin que publicbooking importe schedule.
// cmd/api construye este valor y lo pasa donde publicbooking lo espera.
type AvailabilityLookup struct {
	service *Service
}

// NewAvailabilityLookup construye el adaptador sobre el Service de casos de
// uso ya existente (HU-040/HU-041/HU-042).
func NewAvailabilityLookup(service *Service) AvailabilityLookup {
	return AvailabilityLookup{service: service}
}

// ResolveEffectiveDay implementa el método que
// publicbooking.EffectiveDayPort exige: reutiliza Service.ResolveEffectiveDay
// (CA-041-07) sin duplicar la precedencia excepción-manual > festivo
// automático > horario semanal, y aplana EffectiveDay a pares paralelos de
// hora civil de inicio y duración en minutos (mismo criterio de "tipos
// universales" que booking.BarberServicePort). dateCivil llega en formato
// "AAAA-MM-DD"; se parsea en UTC porque ResolveEffectiveDay solo necesita
// Format/Weekday de la fecha, ambos independientes de zona para una fecha
// sin componente de hora.
func (a AvailabilityLookup) ResolveEffectiveDay(ctx context.Context, barbershopID, barberID, dateCivil string) (
	isWorking bool, segmentStarts []string, segmentDurations []int, err error,
) {
	date, err := time.Parse(dateLayout, dateCivil)
	if err != nil {
		return false, nil, nil, err
	}
	day, err := a.service.ResolveEffectiveDay(ctx, barbershopID, barberID, date)
	if err != nil {
		return false, nil, nil, err
	}
	for _, seg := range day.Segments {
		segmentStarts = append(segmentStarts, seg.StartsTime)
		segmentDurations = append(segmentDurations, seg.DurationMinutes)
	}
	return day.IsWorking, segmentStarts, segmentDurations, nil
}

// BusyIntervals implementa el método que publicbooking.BusyBlocksPort
// exige: reutiliza Service.EffectiveBlocks (HU-042) para proyectar
// bloqueos puntuales y ocurrencias de serie dentro de [fromDate, toDate], y
// los aplana a pares paralelos de instantes absolutos de inicio/fin. Las
// ocurrencias de serie llegan en forma civil (fecha + hora local): se
// convierten a instantes absolutos con timezone, mismo criterio de
// conversión que ManualBookingBlocks.HasActiveBlock.
func (a AvailabilityLookup) BusyIntervals(ctx context.Context, barbershopID, barberID, fromDate, toDate, timezone string) (
	starts []time.Time, ends []time.Time, err error,
) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, nil, err
	}

	result, err := a.service.EffectiveBlocks(ctx, barbershopID, barberID, fromDate, toDate)
	if err != nil {
		return nil, nil, err
	}

	for _, b := range result.ManualBlocks {
		starts = append(starts, b.StartsAt)
		ends = append(ends, b.EndsAt)
	}
	for _, occ := range result.SeriesOccurrences {
		occStart, err := time.ParseInLocation("2006-01-02 15:04", occ.Date+" "+occ.StartsTime, loc)
		if err != nil {
			return nil, nil, err
		}
		occEnd := occStart.Add(time.Duration(occ.DurationMinutes) * time.Minute)
		starts = append(starts, occStart)
		ends = append(ends, occEnd)
	}
	return starts, ends, nil
}
