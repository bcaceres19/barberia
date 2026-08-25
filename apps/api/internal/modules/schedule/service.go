package schedule

import (
	"context"
	"fmt"

	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/idempotency"
)

// Service implementa los cinco casos de uso de HU-040: listar, consultar,
// crear, editar y retirar tramos de horario laboral recurrente de un
// barbero de la barbería activa.
type Service struct {
	repo    Repository
	barbers BarberPort
}

// NewService construye el servicio de casos de uso. barbers colabora con
// staff exclusivamente a través de este puerto (cmd/api lo conecta con
// staff.NewBarberLookup): schedule nunca importa el núcleo ni el
// repositorio de staff.
func NewService(repo Repository, barbers BarberPort) *Service {
	return &Service{repo: repo, barbers: barbers}
}

// List lee una página de tramos del barbero de barberID dentro de
// barbershopID (CA-040-01). Un barberID inexistente o de otra barbería
// produce apperr.NotFound, verificado mediante BarberPort antes de tocar
// el repositorio (mismo criterio que catalog.AssignmentService.List).
func (s *Service) List(ctx context.Context, barbershopID, barberID, cursorToken string, limit int) (ListResult, error) {
	if err := ctx.Err(); err != nil {
		return ListResult{}, apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de listar tramos: %w", err))
	}
	if !LooksLikeBarberID(barberID) {
		return ListResult{}, errBarberNotFound()
	}

	exists, err := s.barbers.Exists(ctx, barbershopID, barberID)
	if err != nil {
		return ListResult{}, apperr.Internal(fmt.Errorf("schedule: verificar barbero para listar tramos: %w", err))
	}
	if !exists {
		return ListResult{}, errBarberNotFound()
	}

	switch {
	case limit <= 0:
		limit = DefaultListLimit
	case limit < MinListLimit:
		limit = MinListLimit
	case limit > MaxListLimit:
		limit = MaxListLimit
	}

	var cursor *Cursor
	if cursorToken != "" {
		decoded, err := DecodeCursor(cursorToken)
		if err != nil {
			return ListResult{}, err
		}
		cursor = &decoded
	}

	result, err := s.repo.List(ctx, barbershopID, barberID, cursor, limit)
	if err != nil {
		return ListResult{}, apperr.Internal(fmt.Errorf("schedule: listar tramos: %w", err))
	}
	return result, nil
}

// Get lee un tramo por id dentro del barbero y barbería activos
// (CA-040-05). Un barberID/workingHourID inexistente, de otro barbero o de
// otra barbería produce el mismo apperr.NotFound.
func (s *Service) Get(ctx context.Context, barbershopID, barberID, workingHourID string) (WorkingHour, error) {
	if err := ctx.Err(); err != nil {
		return WorkingHour{}, apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de leer tramo: %w", err))
	}
	if !LooksLikeBarberID(barberID) || !LooksLikeWorkingHourID(workingHourID) {
		return WorkingHour{}, errWorkingHourNotFound()
	}

	wh, found, err := s.repo.Get(ctx, barbershopID, barberID, workingHourID)
	if err != nil {
		return WorkingHour{}, apperr.Internal(fmt.Errorf("schedule: leer tramo: %w", err))
	}
	if !found {
		return WorkingHour{}, errWorkingHourNotFound()
	}
	return wh, nil
}

// Create registra un tramo (CA-040-02), protegido por el protocolo de
// idempotencia reutilizable de HU-004 (RN-IDE-01, DEC-043). isoWeekday,
// startsTimeRaw y durationMinutes son los valores crudos del cuerpo; key y
// fingerprint ya fueron interpretados por la capa HTTP. La validación de
// los tres campos ocurre ANTES de tocar BarberPort o el repositorio: un
// tramo inválido nunca reclama ni consume la clave de idempotencia, porque
// no es el efecto que RN-IDE-01 protege (mismo criterio que
// staff.Service.Create).
func (s *Service) Create(
	ctx context.Context,
	barbershopID, barberID string,
	isoWeekday int,
	startsTimeRaw string,
	durationMinutes int,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (CreateResult, error) {
	if err := ctx.Err(); err != nil {
		return CreateResult{}, apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de crear tramo: %w", err))
	}
	if !LooksLikeBarberID(barberID) {
		return CreateResult{}, errBarberNotFound()
	}

	input, err := validateInterval(isoWeekday, startsTimeRaw, durationMinutes)
	if err != nil {
		return CreateResult{}, err
	}

	exists, err := s.barbers.Exists(ctx, barbershopID, barberID)
	if err != nil {
		return CreateResult{}, apperr.Internal(fmt.Errorf("schedule: verificar barbero para crear tramo: %w", err))
	}
	if !exists {
		return CreateResult{}, errBarberNotFound()
	}

	result, err := s.repo.Create(ctx, barbershopID, barberID, input, key, fingerprint)
	if err != nil {
		return CreateResult{}, apperr.Internal(fmt.Errorf("schedule: crear tramo: %w", err))
	}
	if result.Conflict {
		return CreateResult{}, errOverlapConflict()
	}
	return result, nil
}

// Update reemplaza el intervalo completo de un tramo existente
// (CA-040-04). Un barberID/workingHourID inexistente, de otro barbero o de
// otra barbería produce el mismo apperr.NotFound que Get (CA-040-05); un
// intervalo nuevo que se solapa con otro tramo existente del mismo barbero
// y día produce apperr.Conflict.
func (s *Service) Update(
	ctx context.Context,
	barbershopID, barberID, workingHourID string,
	isoWeekday int,
	startsTimeRaw string,
	durationMinutes int,
) (WorkingHour, error) {
	if err := ctx.Err(); err != nil {
		return WorkingHour{}, apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de editar tramo: %w", err))
	}
	if !LooksLikeBarberID(barberID) || !LooksLikeWorkingHourID(workingHourID) {
		return WorkingHour{}, errWorkingHourNotFound()
	}

	input, err := validateInterval(isoWeekday, startsTimeRaw, durationMinutes)
	if err != nil {
		return WorkingHour{}, err
	}

	result, err := s.repo.Update(ctx, barbershopID, barberID, workingHourID, input)
	if err != nil {
		return WorkingHour{}, apperr.Internal(fmt.Errorf("schedule: editar tramo: %w", err))
	}
	if !result.Found {
		return WorkingHour{}, errWorkingHourNotFound()
	}
	if result.Conflict {
		return WorkingHour{}, errOverlapConflict()
	}
	return result.WorkingHour, nil
}

// Delete retira físicamente un tramo existente (CA-040-05). Un
// barberID/workingHourID inexistente, de otro barbero, de otra barbería, o
// ya retirado antes, producen el mismo apperr.NotFound; reintentar la
// misma operación tras ese error es seguro.
func (s *Service) Delete(ctx context.Context, barbershopID, barberID, workingHourID string) error {
	if err := ctx.Err(); err != nil {
		return apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de retirar tramo: %w", err))
	}
	if !LooksLikeBarberID(barberID) || !LooksLikeWorkingHourID(workingHourID) {
		return errWorkingHourNotFound()
	}

	found, err := s.repo.Delete(ctx, barbershopID, barberID, workingHourID)
	if err != nil {
		return apperr.Internal(fmt.Errorf("schedule: retirar tramo: %w", err))
	}
	if !found {
		return errWorkingHourNotFound()
	}
	return nil
}

// validateInterval valida los tres campos de un intervalo (CA-040-04):
// día fuera de 1-7, hora con formato inválido o duración fuera de
// [1, 1440] se rechazan sin tocar BarberPort ni el repositorio.
func validateInterval(isoWeekday int, startsTimeRaw string, durationMinutes int) (CreateInput, error) {
	if isoWeekday < MinISOWeekday || isoWeekday > MaxISOWeekday {
		return CreateInput{}, errISOWeekdayInvalid()
	}
	startsTime, err := ValidateStartsTime(startsTimeRaw)
	if err != nil {
		return CreateInput{}, err
	}
	if durationMinutes < MinDurationMinutes || durationMinutes > MaxDurationMinutes {
		return CreateInput{}, errDurationInvalid()
	}
	return CreateInput{ISOWeekday: isoWeekday, StartsTime: startsTime, DurationMinutes: durationMinutes}, nil
}
