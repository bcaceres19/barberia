package publicbooking

import (
	"context"
	"fmt"
	"strings"
	"time"

	"system-barbershop/internal/modules/availability"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/clock"
)

// AvailabilityService implementa el caso de uso de HU-094: proyectar los
// inicios públicos válidos de un servicio ya elegido, para un barbero ya
// elegido, dentro de la barbería resuelta por slug. Deliberadamente
// SEPARADO de Service (que resuelve HU-090/HU-091/HU-092): necesita
// colaborar con schedule/shops/catalog a través de puertos pequeños
// (EffectiveDayPort/BusyBlocksPort/BookingPolicyPort/ServiceAssignmentPort/
// AvailabilityTimezonePort, mismo criterio de HU-061), algo que las otras
// tres operaciones -de una sola tabla- nunca necesitaron.
type AvailabilityService struct {
	repo         AvailabilityRepository
	effectiveDay EffectiveDayPort
	blocks       BusyBlocksPort
	policy       BookingPolicyPort
	assignments  ServiceAssignmentPort
	timezones    AvailabilityTimezonePort
	clock        clock.Clock
}

// NewAvailabilityService construye el caso de uso a partir de sus seis
// colaboradores.
func NewAvailabilityService(
	repo AvailabilityRepository,
	effectiveDay EffectiveDayPort,
	blocks BusyBlocksPort,
	policy BookingPolicyPort,
	assignments ServiceAssignmentPort,
	timezones AvailabilityTimezonePort,
	clk clock.Clock,
) *AvailabilityService {
	return &AvailabilityService{
		repo:         repo,
		effectiveDay: effectiveDay,
		blocks:       blocks,
		policy:       policy,
		assignments:  assignments,
		timezones:    timezones,
		clock:        clk,
	}
}

// civilDateLayout es el formato de fecha civil que EffectiveDayPort y
// BusyBlocksPort esperan ("AAAA-MM-DD"), mismo layout que
// schedule.dateLayout.
const civilDateLayout = "2006-01-02"

// civilDateTimeLayout combina civilDateLayout con la hora "HH:MM" que
// EffectiveDayPort devuelve para cada tramo (RN-DIS-07: siempre hora local
// de la barbería, nunca del dispositivo).
const civilDateTimeLayout = "2006-01-02 15:04"

// ListPublicAvailability resuelve rawSlug -mismo criterio de recorte y
// forma que Service.ResolveBarbershop- y, si resuelve, proyecta los
// inicios válidos del servicio rawServiceID con el barbero rawBarberID
// (CA-094-01 a CA-094-06). Un slug inválido, desconocido o de una barbería
// no publicable produce el mismo apperr.NotFound uniforme que las demás
// operaciones públicas (CA-090-02). Un rawServiceID/rawBarberID sin forma
// de UUID, ajeno, inexistente, inactivo o sin asignación vigente NUNCA
// produce un error distinto (mismo criterio que CA-092-03): el resultado es
// una disponibilidad vacía, exactamente igual que "sin franjas hoy".
func (s *AvailabilityService) ListPublicAvailability(ctx context.Context, rawSlug, rawServiceID, rawBarberID string) (AvailabilityResult, error) {
	if err := ctx.Err(); err != nil {
		return AvailabilityResult{}, apperr.Internal(fmt.Errorf("publicbooking: contexto cancelado antes de calcular disponibilidad: %w", err))
	}

	slug := strings.TrimSpace(rawSlug)
	if slug == "" || len(slug) > MaxSlugLength {
		return AvailabilityResult{}, errBarbershopNotPublic()
	}

	barbershopID, found, err := s.repo.ResolveBarbershopID(ctx, slug)
	if err != nil {
		return AvailabilityResult{}, apperr.Internal(fmt.Errorf("publicbooking: resolver barbería para disponibilidad: %w", err))
	}
	if !found {
		return AvailabilityResult{}, errBarbershopNotPublic()
	}

	if !LooksLikePublicServiceID(rawServiceID) || !LooksLikePublicBarberID(rawBarberID) {
		return AvailabilityResult{}, nil
	}

	_, durationMinutes, _, _, assigned, err := s.assignments.ActiveAssignedService(ctx, barbershopID, rawBarberID, rawServiceID)
	if err != nil {
		return AvailabilityResult{}, apperr.Internal(fmt.Errorf("publicbooking: verificar asignación para disponibilidad: %w", err))
	}
	if !assigned || durationMinutes <= 0 {
		return AvailabilityResult{}, nil
	}

	timezone, err := s.timezones.Timezone(ctx, barbershopID)
	if err != nil {
		return AvailabilityResult{}, apperr.Internal(fmt.Errorf("publicbooking: resolver zona horaria para disponibilidad: %w", err))
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return AvailabilityResult{}, apperr.Internal(fmt.Errorf("publicbooking: zona horaria irresoluble para disponibilidad: %w", err))
	}

	minAdvanceMinutes, maxAdvanceDays, slotGridMinutes, err := s.policy.BookingPolicy(ctx, barbershopID)
	if err != nil {
		return AvailabilityResult{}, apperr.Internal(fmt.Errorf("publicbooking: resolver política de reserva para disponibilidad: %w", err))
	}
	if maxAdvanceDays <= 0 || slotGridMinutes <= 0 {
		return AvailabilityResult{}, apperr.Internal(fmt.Errorf("publicbooking: política de reserva con valores no positivos"))
	}

	now := s.clock.Now()
	earliestStart := now.Add(time.Duration(minAdvanceMinutes) * time.Minute)
	latestStart := now.AddDate(0, 0, maxAdvanceDays)

	fromDate := civilDateOnly(now.In(loc))
	toDate := civilDateOnly(latestStart.In(loc))

	segments, err := s.resolveSegments(ctx, barbershopID, rawBarberID, fromDate, toDate, loc)
	if err != nil {
		return AvailabilityResult{}, err
	}

	busy, err := s.resolveBusyIntervals(ctx, barbershopID, rawBarberID, fromDate, toDate, timezone, now, latestStart)
	if err != nil {
		return AvailabilityResult{}, err
	}

	starts := availability.GenerateStarts(segments, busy, availability.Policy{
		ServiceDuration: time.Duration(durationMinutes) * time.Minute,
		GridStep:        time.Duration(slotGridMinutes) * time.Minute,
		EarliestStart:   earliestStart,
		LatestStart:     latestStart,
	})

	slots := make([]AvailabilitySlot, 0, len(starts))
	for _, t := range starts {
		slots = append(slots, AvailabilitySlot{StartsAt: t})
	}

	return AvailabilityResult{
		Slots:           slots,
		DurationMinutes: durationMinutes,
		Timezone:        timezone,
		SlotGridMinutes: slotGridMinutes,
	}, nil
}

// civilDateOnly trunca t a medianoche civil en su propia ubicación: la
// misma fecha civil que t.Format(civilDateLayout) representaría.
func civilDateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// resolveSegments consulta EffectiveDayPort para cada fecha civil de
// [fromDate, toDate] (ambos límites inclusive) y convierte cada tramo
// (hora civil + duración) a un availability.Interval de instantes
// absolutos en loc. No duplica la precedencia excepción manual > festivo
// automático > horario semanal de B2: solo consume su resultado ya
// resuelto.
func (s *AvailabilityService) resolveSegments(ctx context.Context, barbershopID, barberID string, fromDate, toDate time.Time, loc *time.Location) ([]availability.Interval, error) {
	var segments []availability.Interval
	for d := fromDate; !d.After(toDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format(civilDateLayout)
		isWorking, starts, durations, err := s.effectiveDay.ResolveEffectiveDay(ctx, barbershopID, barberID, dateStr)
		if err != nil {
			return nil, apperr.Internal(fmt.Errorf("publicbooking: resolver jornada efectiva para disponibilidad: %w", err))
		}
		if !isWorking {
			continue
		}
		if len(durations) != len(starts) || len(starts) > EffectiveDaySegmentsLimit {
			return nil, apperr.Internal(fmt.Errorf("publicbooking: adaptador de jornada efectiva devolvió una forma inesperada para %s", dateStr))
		}
		for i, startsTime := range starts {
			segStart, err := time.ParseInLocation(civilDateTimeLayout, dateStr+" "+startsTime, loc)
			if err != nil {
				return nil, apperr.Internal(fmt.Errorf("publicbooking: tramo con hora civil inválida (%s %s): %w", dateStr, startsTime, err))
			}
			if durations[i] <= 0 {
				continue
			}
			segEnd := segStart.Add(time.Duration(durations[i]) * time.Minute)
			segments = append(segments, availability.Interval{Start: segStart, End: segEnd})
		}
	}
	return segments, nil
}

// resolveBusyIntervals combina las ocupaciones de BusyBlocksPort (bloqueos,
// RN-BLQ-01 a RN-BLQ-04) con las de AvailabilityRepository.ListOccupiedIntervals
// (citas que ocupan agenda, RN-CON-01/RN-CAN-04) dentro de
// [now, latestStart]: la ventana exacta que puede producir un candidato
// válido (CA-094-05).
func (s *AvailabilityService) resolveBusyIntervals(
	ctx context.Context,
	barbershopID, barberID string,
	fromDate, toDate time.Time,
	timezone string,
	now, latestStart time.Time,
) ([]availability.Interval, error) {
	blockStarts, blockEnds, err := s.blocks.BusyIntervals(ctx, barbershopID, barberID, fromDate.Format(civilDateLayout), toDate.Format(civilDateLayout), timezone)
	if err != nil {
		return nil, apperr.Internal(fmt.Errorf("publicbooking: resolver bloqueos para disponibilidad: %w", err))
	}
	if len(blockStarts) != len(blockEnds) {
		return nil, apperr.Internal(fmt.Errorf("publicbooking: adaptador de bloqueos devolvió una forma inesperada"))
	}

	appointmentStarts, appointmentEnds, err := s.repo.ListOccupiedIntervals(ctx, barbershopID, barberID, now, latestStart)
	if err != nil {
		return nil, apperr.Internal(fmt.Errorf("publicbooking: resolver citas para disponibilidad: %w", err))
	}
	if len(appointmentStarts) != len(appointmentEnds) {
		return nil, apperr.Internal(fmt.Errorf("publicbooking: consulta de ocupación devolvió una forma inesperada"))
	}

	busy := make([]availability.Interval, 0, len(blockStarts)+len(appointmentStarts))
	for i := range blockStarts {
		busy = append(busy, availability.Interval{Start: blockStarts[i], End: blockEnds[i]})
	}
	for i := range appointmentStarts {
		busy = append(busy, availability.Interval{Start: appointmentStarts[i], End: appointmentEnds[i]})
	}
	return busy, nil
}
