package schedule

import (
	"context"
	"fmt"
	"time"

	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/idempotency"
)

// GetHolidayCalendar lee el interruptor de calendario colombiano de
// festivos del barbero de barberID dentro de barbershopID (CA-041-01/02).
// Un barberID inexistente o de otra barbería produce apperr.NotFound.
func (s *Service) GetHolidayCalendar(ctx context.Context, barbershopID, barberID string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de leer el calendario de festivos: %w", err))
	}
	if !LooksLikeBarberID(barberID) {
		return false, errBarberNotFound()
	}

	enabled, found, err := s.repo.GetHolidayCalendarEnabled(ctx, barbershopID, barberID)
	if err != nil {
		return false, apperr.Internal(fmt.Errorf("schedule: leer calendario de festivos: %w", err))
	}
	if !found {
		return false, errBarberNotFound()
	}
	return enabled, nil
}

// SetHolidayCalendar activa o desactiva el interruptor (CA-041-01/02). Un
// barberID inexistente o de otra barbería produce apperr.NotFound.
func (s *Service) SetHolidayCalendar(ctx context.Context, barbershopID, barberID string, enabled bool) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de actualizar el calendario de festivos: %w", err))
	}
	if !LooksLikeBarberID(barberID) {
		return false, errBarberNotFound()
	}

	result, err := s.repo.SetHolidayCalendarEnabled(ctx, barbershopID, barberID, enabled)
	if err != nil {
		return false, apperr.Internal(fmt.Errorf("schedule: actualizar calendario de festivos: %w", err))
	}
	if !result.Found {
		return false, errBarberNotFound()
	}
	return result.Enabled, nil
}

// ListExceptions lee una página de excepciones del barbero de barberID
// dentro de barbershopID (CA-041-04/05). Un barberID inexistente o de otra
// barbería produce apperr.NotFound, verificado mediante BarberPort antes
// de tocar el repositorio (mismo criterio que List de HU-040).
func (s *Service) ListExceptions(ctx context.Context, barbershopID, barberID, cursorToken string, limit int) (ExceptionListResult, error) {
	if err := ctx.Err(); err != nil {
		return ExceptionListResult{}, apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de listar excepciones: %w", err))
	}
	if !LooksLikeBarberID(barberID) {
		return ExceptionListResult{}, errBarberNotFound()
	}

	exists, err := s.barbers.Exists(ctx, barbershopID, barberID)
	if err != nil {
		return ExceptionListResult{}, apperr.Internal(fmt.Errorf("schedule: verificar barbero para listar excepciones: %w", err))
	}
	if !exists {
		return ExceptionListResult{}, errBarberNotFound()
	}

	switch {
	case limit <= 0:
		limit = DefaultListLimit
	case limit < MinListLimit:
		limit = MinListLimit
	case limit > MaxListLimit:
		limit = MaxListLimit
	}

	var cursor *ExceptionCursor
	if cursorToken != "" {
		decoded, err := DecodeExceptionCursor(cursorToken)
		if err != nil {
			return ExceptionListResult{}, err
		}
		cursor = &decoded
	}

	result, err := s.repo.ListExceptions(ctx, barbershopID, barberID, cursor, limit)
	if err != nil {
		return ExceptionListResult{}, apperr.Internal(fmt.Errorf("schedule: listar excepciones: %w", err))
	}
	return result, nil
}

// GetException lee una excepción por id dentro del barbero y barbería
// activos (CA-041-06).
func (s *Service) GetException(ctx context.Context, barbershopID, barberID, exceptionID string) (ScheduleException, error) {
	if err := ctx.Err(); err != nil {
		return ScheduleException{}, apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de leer la excepción: %w", err))
	}
	if !LooksLikeBarberID(barberID) || !LooksLikeExceptionID(exceptionID) {
		return ScheduleException{}, errExceptionNotFound()
	}

	exception, found, err := s.repo.GetException(ctx, barbershopID, barberID, exceptionID)
	if err != nil {
		return ScheduleException{}, apperr.Internal(fmt.Errorf("schedule: leer excepción: %w", err))
	}
	if !found {
		return ScheduleException{}, errExceptionNotFound()
	}
	return exception, nil
}

// CreateException registra una excepción de jornada (CA-041-04, CA-041-05),
// protegida por el protocolo de idempotencia reutilizable de HU-004
// (RN-IDE-01, DEC-043). effectiveDateRaw, isClosed, reasonRaw y
// segmentsRaw son los valores crudos del cuerpo; key y fingerprint ya
// fueron interpretados por la capa HTTP. La validación de forma ocurre
// ANTES de tocar BarberPort o el repositorio (mismo criterio que Create de
// HU-040): una excepción inválida nunca reclama ni consume la clave de
// idempotencia.
func (s *Service) CreateException(
	ctx context.Context,
	barbershopID, barberID string,
	effectiveDateRaw string,
	isClosed bool,
	reasonRaw *string,
	segmentsRaw []CreateExceptionSegmentInput,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (CreateExceptionResult, error) {
	if err := ctx.Err(); err != nil {
		return CreateExceptionResult{}, apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de crear excepción: %w", err))
	}
	if !LooksLikeBarberID(barberID) {
		return CreateExceptionResult{}, errBarberNotFound()
	}

	input, err := validateExceptionInput(effectiveDateRaw, isClosed, reasonRaw, segmentsRaw)
	if err != nil {
		return CreateExceptionResult{}, err
	}

	exists, err := s.barbers.Exists(ctx, barbershopID, barberID)
	if err != nil {
		return CreateExceptionResult{}, apperr.Internal(fmt.Errorf("schedule: verificar barbero para crear excepción: %w", err))
	}
	if !exists {
		return CreateExceptionResult{}, errBarberNotFound()
	}

	result, err := s.repo.CreateException(ctx, barbershopID, barberID, input, key, fingerprint)
	if err != nil {
		return CreateExceptionResult{}, apperr.Internal(fmt.Errorf("schedule: crear excepción: %w", err))
	}
	if result.Conflict {
		return CreateExceptionResult{}, errExceptionDateConflict()
	}
	return result, nil
}

// UpdateException reemplaza la excepción completa (CA-041-04, CA-041-05).
// Un barberID/exceptionID inexistente, de otro barbero o de otra barbería
// produce el mismo apperr.NotFound que GetException (CA-041-06); una fecha
// que ya tiene otra excepción del mismo barbero, o tramos que se solapan,
// producen apperr.Conflict.
func (s *Service) UpdateException(
	ctx context.Context,
	barbershopID, barberID, exceptionID string,
	effectiveDateRaw string,
	isClosed bool,
	reasonRaw *string,
	segmentsRaw []CreateExceptionSegmentInput,
) (ScheduleException, error) {
	if err := ctx.Err(); err != nil {
		return ScheduleException{}, apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de editar excepción: %w", err))
	}
	if !LooksLikeBarberID(barberID) || !LooksLikeExceptionID(exceptionID) {
		return ScheduleException{}, errExceptionNotFound()
	}

	input, err := validateExceptionInput(effectiveDateRaw, isClosed, reasonRaw, segmentsRaw)
	if err != nil {
		return ScheduleException{}, err
	}

	result, err := s.repo.UpdateException(ctx, barbershopID, barberID, exceptionID, input)
	if err != nil {
		return ScheduleException{}, apperr.Internal(fmt.Errorf("schedule: editar excepción: %w", err))
	}
	if !result.Found {
		return ScheduleException{}, errExceptionNotFound()
	}
	if result.Conflict {
		return ScheduleException{}, errExceptionDateConflict()
	}
	return result.Exception, nil
}

// DeleteException retira físicamente una excepción existente (CA-041-06).
func (s *Service) DeleteException(ctx context.Context, barbershopID, barberID, exceptionID string) error {
	if err := ctx.Err(); err != nil {
		return apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de retirar excepción: %w", err))
	}
	if !LooksLikeBarberID(barberID) || !LooksLikeExceptionID(exceptionID) {
		return errExceptionNotFound()
	}

	found, err := s.repo.DeleteException(ctx, barbershopID, barberID, exceptionID)
	if err != nil {
		return apperr.Internal(fmt.Errorf("schedule: retirar excepción: %w", err))
	}
	if !found {
		return errExceptionNotFound()
	}
	return nil
}

func validateExceptionInput(effectiveDateRaw string, isClosed bool, reasonRaw *string, segmentsRaw []CreateExceptionSegmentInput) (CreateExceptionInput, error) {
	effectiveDate, err := ValidateEffectiveDate(effectiveDateRaw)
	if err != nil {
		return CreateExceptionInput{}, err
	}
	reason, err := ValidateReason(reasonRaw)
	if err != nil {
		return CreateExceptionInput{}, err
	}
	segments, err := ValidateExceptionShape(isClosed, segmentsRaw)
	if err != nil {
		return CreateExceptionInput{}, err
	}
	return CreateExceptionInput{EffectiveDate: effectiveDate, IsClosed: isClosed, Reason: reason, Segments: segments}, nil
}

// --- Puerto interno de resolución de jornada efectiva (CA-041-07) --------

// EffectiveDaySource identifica qué regla decidió la jornada resuelta,
// según la precedencia de CA-041-07: excepción manual abierta/cerrada >
// festivo automático > horario semanal.
type EffectiveDaySource string

const (
	EffectiveDaySourceManualOpen   EffectiveDaySource = "manual-open"
	EffectiveDaySourceManualClosed EffectiveDaySource = "manual-closed"
	EffectiveDaySourceHolidayAuto  EffectiveDaySource = "holiday-auto"
	EffectiveDaySourceWeekly       EffectiveDaySource = "weekly"
)

// EffectiveDaySegment es un tramo trabajable, sin identificador propio:
// solo lo que un consumidor futuro (disponibilidad, B4) necesita para
// calcular huecos.
type EffectiveDaySegment struct {
	StartsTime      string
	DurationMinutes int
}

// EffectiveDay es el resultado de resolver la jornada de un barbero para
// una fecha civil concreta (CA-041-07): puerto interno para que la
// disponibilidad (B4) lo consuma después. No crea citas ni disponibilidad
// pública; no se expone por HTTP.
type EffectiveDay struct {
	Date      string
	IsWorking bool
	Segments  []EffectiveDaySegment
	Source    EffectiveDaySource
}

// ResolveEffectiveDay aplica la precedencia de CA-041-07: una excepción
// manual para esa fecha (abierta o cerrada) prevalece siempre; en su
// ausencia, un festivo colombiano bloquea el día SOLO si el barbero
// activó su calendario (CA-041-01/02); en cualquier otro caso, se usa el
// horario semanal de working_hour para el día ISO correspondiente
// (HU-040). date se interpreta siempre como fecha civil, nunca como un
// instante que dependa de la zona del dispositivo (CA-041-07): el
// llamador es responsable de construirla ya en la zona IANA de la
// barbería (HU-020).
func (s *Service) ResolveEffectiveDay(ctx context.Context, barbershopID, barberID string, date time.Time) (EffectiveDay, error) {
	if err := ctx.Err(); err != nil {
		return EffectiveDay{}, apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de resolver la jornada efectiva: %w", err))
	}
	if !LooksLikeBarberID(barberID) {
		return EffectiveDay{}, errBarberNotFound()
	}

	dateStr := date.Format("2006-01-02")

	exception, found, err := s.repo.GetExceptionByDate(ctx, barbershopID, barberID, dateStr)
	if err != nil {
		return EffectiveDay{}, apperr.Internal(fmt.Errorf("schedule: resolver jornada efectiva: leer excepción: %w", err))
	}
	if found {
		if exception.IsClosed {
			return EffectiveDay{Date: dateStr, IsWorking: false, Source: EffectiveDaySourceManualClosed}, nil
		}
		segments := make([]EffectiveDaySegment, 0, len(exception.Segments))
		for _, seg := range exception.Segments {
			segments = append(segments, EffectiveDaySegment{StartsTime: seg.StartsTime, DurationMinutes: seg.DurationMinutes})
		}
		return EffectiveDay{Date: dateStr, IsWorking: true, Segments: segments, Source: EffectiveDaySourceManualOpen}, nil
	}

	holidayEnabled, foundBarber, err := s.repo.GetHolidayCalendarEnabled(ctx, barbershopID, barberID)
	if err != nil {
		return EffectiveDay{}, apperr.Internal(fmt.Errorf("schedule: resolver jornada efectiva: leer calendario de festivos: %w", err))
	}
	if !foundBarber {
		return EffectiveDay{}, errBarberNotFound()
	}
	if holidayEnabled {
		if _, isHoliday := IsColombianHoliday(date); isHoliday {
			return EffectiveDay{Date: dateStr, IsWorking: false, Source: EffectiveDaySourceHolidayAuto}, nil
		}
	}

	isoWeekday := int(date.Weekday())
	if isoWeekday == 0 {
		isoWeekday = 7 // time.Sunday == 0; DEC-020 usa 1=lunes...7=domingo.
	}
	weekly, err := s.repo.ListWorkingHoursForWeekday(ctx, barbershopID, barberID, isoWeekday)
	if err != nil {
		return EffectiveDay{}, apperr.Internal(fmt.Errorf("schedule: resolver jornada efectiva: leer horario semanal: %w", err))
	}
	segments := make([]EffectiveDaySegment, 0, len(weekly))
	for _, wh := range weekly {
		segments = append(segments, EffectiveDaySegment{StartsTime: wh.StartsTime, DurationMinutes: wh.DurationMinutes})
	}
	return EffectiveDay{Date: dateStr, IsWorking: len(segments) > 0, Segments: segments, Source: EffectiveDaySourceWeekly}, nil
}
