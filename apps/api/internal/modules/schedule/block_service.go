package schedule

import (
	"context"
	"fmt"
	"time"

	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/idempotency"
)

// clampBlockLimit aplica el mismo recorte que List (working_hour) a las
// listas de bloqueos/series: límite técnico de página, no un máximo de
// negocio de bloqueos por barbero.
func clampBlockLimit(limit int) int {
	switch {
	case limit <= 0:
		return DefaultListLimit
	case limit < MinListLimit:
		return MinListLimit
	case limit > MaxListLimit:
		return MaxListLimit
	default:
		return limit
	}
}

// ---------------------------------------------------------------------
// Bloqueos puntuales (HU-042)
// ---------------------------------------------------------------------

// ListBlocks lee una página de bloqueos puntuales del barbero de barberID
// dentro de barbershopID. Un barberID inexistente o de otra barbería
// produce apperr.NotFound.
func (s *Service) ListBlocks(ctx context.Context, barbershopID, barberID, cursorToken string, limit int, includeDeleted bool) (BlockListResult, error) {
	if err := ctx.Err(); err != nil {
		return BlockListResult{}, apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de listar bloqueos: %w", err))
	}
	if !LooksLikeBarberID(barberID) {
		return BlockListResult{}, errBarberNotFound()
	}
	exists, err := s.barbers.Exists(ctx, barbershopID, barberID)
	if err != nil {
		return BlockListResult{}, apperr.Internal(fmt.Errorf("schedule: verificar barbero para listar bloqueos: %w", err))
	}
	if !exists {
		return BlockListResult{}, errBarberNotFound()
	}

	limit = clampBlockLimit(limit)

	var cursor *BlockCursor
	if cursorToken != "" {
		decoded, err := DecodeBlockCursor(cursorToken)
		if err != nil {
			return BlockListResult{}, err
		}
		cursor = &decoded
	}

	result, err := s.repo.ListBlocks(ctx, barbershopID, barberID, cursor, limit, includeDeleted)
	if err != nil {
		return BlockListResult{}, apperr.Internal(fmt.Errorf("schedule: listar bloqueos: %w", err))
	}
	return result, nil
}

// GetBlock lee un bloqueo por id dentro del barbero y barbería activos,
// incluidos los retirados lógicamente (RN-BLQ-04: el registro se conserva).
func (s *Service) GetBlock(ctx context.Context, barbershopID, barberID, blockID string) (TimeBlock, error) {
	if err := ctx.Err(); err != nil {
		return TimeBlock{}, apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de leer bloqueo: %w", err))
	}
	if !LooksLikeBarberID(barberID) || !LooksLikeTimeBlockID(blockID) {
		return TimeBlock{}, errTimeBlockNotFound()
	}
	block, found, err := s.repo.GetBlock(ctx, barbershopID, barberID, blockID)
	if err != nil {
		return TimeBlock{}, apperr.Internal(fmt.Errorf("schedule: leer bloqueo: %w", err))
	}
	if !found {
		return TimeBlock{}, errTimeBlockNotFound()
	}
	return block, nil
}

// CreateBlock registra un bloqueo puntual, protegido por el protocolo de
// idempotencia reutilizable de HU-004 (RN-IDE-01, DEC-043). A diferencia de
// working_hour, nunca produce un conflicto de negocio: RN-BLQ-03/DEC-008
// exigen que crear un bloqueo jamás falle por chocar con otro estado
// existente.
func (s *Service) CreateBlock(
	ctx context.Context,
	barbershopID, barberID string,
	blockType, startsAtRaw, endsAtRaw string,
	reason *string,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (CreateBlockResult, error) {
	if err := ctx.Err(); err != nil {
		return CreateBlockResult{}, apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de crear bloqueo: %w", err))
	}
	if !LooksLikeBarberID(barberID) {
		return CreateBlockResult{}, errBarberNotFound()
	}

	input, err := validateBlockInput(blockType, startsAtRaw, endsAtRaw, reason)
	if err != nil {
		return CreateBlockResult{}, err
	}

	exists, err := s.barbers.Exists(ctx, barbershopID, barberID)
	if err != nil {
		return CreateBlockResult{}, apperr.Internal(fmt.Errorf("schedule: verificar barbero para crear bloqueo: %w", err))
	}
	if !exists {
		return CreateBlockResult{}, errBarberNotFound()
	}

	result, err := s.repo.CreateBlock(ctx, barbershopID, barberID, input, key, fingerprint)
	if err != nil {
		return CreateBlockResult{}, apperr.Internal(fmt.Errorf("schedule: crear bloqueo: %w", err))
	}
	return result, nil
}

// DeleteBlock retira lógicamente un bloqueo existente (RN-BLQ-04): fija
// deleted_at/deleted_by y deja de restar disponibilidad, pero el registro
// se conserva. actorID es el staff_user que ejecuta el retiro. Un
// barberID/blockID inexistente, de otro barbero, de otra barbería, o ya
// retirado antes, producen el mismo apperr.NotFound; reintentar tras ese
// error es seguro y no reescribe deleted_at/deleted_by.
func (s *Service) DeleteBlock(ctx context.Context, barbershopID, barberID, blockID, actorID string) error {
	if err := ctx.Err(); err != nil {
		return apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de retirar bloqueo: %w", err))
	}
	if !LooksLikeBarberID(barberID) || !LooksLikeTimeBlockID(blockID) {
		return errTimeBlockNotFound()
	}
	found, err := s.repo.DeleteBlock(ctx, barbershopID, barberID, blockID, actorID)
	if err != nil {
		return apperr.Internal(fmt.Errorf("schedule: retirar bloqueo: %w", err))
	}
	if !found {
		return errTimeBlockNotFound()
	}
	return nil
}

func validateBlockInput(blockType, startsAtRaw, endsAtRaw string, reason *string) (CreateBlockInput, error) {
	bt, err := ValidateBlockType(blockType)
	if err != nil {
		return CreateBlockInput{}, err
	}
	startsAt, err := ValidateInstant(startsAtRaw)
	if err != nil {
		return CreateBlockInput{}, err
	}
	endsAt, err := ValidateInstant(endsAtRaw)
	if err != nil {
		return CreateBlockInput{}, err
	}
	if err := ValidateBlockInterval(startsAt, endsAt); err != nil {
		return CreateBlockInput{}, err
	}
	r, err := ValidateReason(reason)
	if err != nil {
		return CreateBlockInput{}, err
	}
	return CreateBlockInput{BlockType: bt, StartsAt: startsAt, EndsAt: endsAt, Reason: r}, nil
}

// ---------------------------------------------------------------------
// Series recurrentes (HU-042)
// ---------------------------------------------------------------------

// ListSeries lee una página de series del barbero de barberID dentro de
// barbershopID, cada una con sus fechas/excepciones embebidas.
func (s *Service) ListSeries(ctx context.Context, barbershopID, barberID, cursorToken string, limit int, includeDeleted bool) (SeriesListResult, error) {
	if err := ctx.Err(); err != nil {
		return SeriesListResult{}, apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de listar series: %w", err))
	}
	if !LooksLikeBarberID(barberID) {
		return SeriesListResult{}, errBarberNotFound()
	}
	exists, err := s.barbers.Exists(ctx, barbershopID, barberID)
	if err != nil {
		return SeriesListResult{}, apperr.Internal(fmt.Errorf("schedule: verificar barbero para listar series: %w", err))
	}
	if !exists {
		return SeriesListResult{}, errBarberNotFound()
	}

	limit = clampBlockLimit(limit)

	var cursor *SeriesCursor
	if cursorToken != "" {
		decoded, err := DecodeSeriesCursor(cursorToken)
		if err != nil {
			return SeriesListResult{}, err
		}
		cursor = &decoded
	}

	result, err := s.repo.ListSeries(ctx, barbershopID, barberID, cursor, limit, includeDeleted)
	if err != nil {
		return SeriesListResult{}, apperr.Internal(fmt.Errorf("schedule: listar series: %w", err))
	}
	return result, nil
}

// GetSeries lee una serie por id dentro del barbero y barbería activos,
// incluidas las retiradas lógicamente.
func (s *Service) GetSeries(ctx context.Context, barbershopID, barberID, seriesID string) (TimeBlockSeries, error) {
	if err := ctx.Err(); err != nil {
		return TimeBlockSeries{}, apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de leer serie: %w", err))
	}
	if !LooksLikeBarberID(barberID) || !LooksLikeSeriesID(seriesID) {
		return TimeBlockSeries{}, errTimeBlockSeriesNotFound()
	}
	series, found, err := s.repo.GetSeries(ctx, barbershopID, barberID, seriesID)
	if err != nil {
		return TimeBlockSeries{}, apperr.Internal(fmt.Errorf("schedule: leer serie: %w", err))
	}
	if !found {
		return TimeBlockSeries{}, errTimeBlockSeriesNotFound()
	}
	return series, nil
}

// CreateSeries registra una definición recurrente semanal o una lista
// explícita de fechas (RN-BLQ-01, DEC-020), protegida por el protocolo de
// idempotencia reutilizable. explicitDates solo se acepta cuando
// recurrenceKind es date_list (RN-BLQ-01: "bloqueo de varios días debe
// poder crearse en una sola operación").
func (s *Service) CreateSeries(
	ctx context.Context,
	barbershopID, barberID string,
	blockType, recurrenceKind string,
	isoWeekday *int,
	startsTimeRaw string,
	durationMinutes int,
	effectiveFrom string,
	effectiveUntil *string,
	reason *string,
	explicitDates []string,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (CreateSeriesResult, error) {
	if err := ctx.Err(); err != nil {
		return CreateSeriesResult{}, apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de crear serie: %w", err))
	}
	if !LooksLikeBarberID(barberID) {
		return CreateSeriesResult{}, errBarberNotFound()
	}

	input, err := validateSeriesInput(blockType, recurrenceKind, isoWeekday, startsTimeRaw, durationMinutes, effectiveFrom, effectiveUntil, reason, explicitDates)
	if err != nil {
		return CreateSeriesResult{}, err
	}

	exists, err := s.barbers.Exists(ctx, barbershopID, barberID)
	if err != nil {
		return CreateSeriesResult{}, apperr.Internal(fmt.Errorf("schedule: verificar barbero para crear serie: %w", err))
	}
	if !exists {
		return CreateSeriesResult{}, errBarberNotFound()
	}

	result, err := s.repo.CreateSeries(ctx, barbershopID, barberID, input, key, fingerprint)
	if err != nil {
		return CreateSeriesResult{}, apperr.Internal(fmt.Errorf("schedule: crear serie: %w", err))
	}
	return result, nil
}

func validateSeriesInput(
	blockType, recurrenceKind string,
	isoWeekday *int,
	startsTimeRaw string,
	durationMinutes int,
	effectiveFrom string,
	effectiveUntil *string,
	reason *string,
	explicitDates []string,
) (CreateSeriesInput, error) {
	bt, err := ValidateBlockType(blockType)
	if err != nil {
		return CreateSeriesInput{}, err
	}
	rk, err := ValidateRecurrenceKind(recurrenceKind)
	if err != nil {
		return CreateSeriesInput{}, err
	}
	if err := ValidateWeekdayShape(rk, isoWeekday); err != nil {
		return CreateSeriesInput{}, err
	}
	startsTime, err := ValidateStartsTime(startsTimeRaw)
	if err != nil {
		return CreateSeriesInput{}, err
	}
	if durationMinutes < MinDurationMinutes || durationMinutes > MaxDurationMinutes {
		return CreateSeriesInput{}, errDurationInvalid()
	}
	from, err := ValidateCivilDate(effectiveFrom)
	if err != nil {
		return CreateSeriesInput{}, err
	}
	var until *string
	if effectiveUntil != nil {
		u, err := ValidateCivilDate(*effectiveUntil)
		if err != nil {
			return CreateSeriesInput{}, err
		}
		until = &u
	}
	if err := ValidateEffectiveRange(from, until); err != nil {
		return CreateSeriesInput{}, err
	}
	r, err := ValidateReason(reason)
	if err != nil {
		return CreateSeriesInput{}, err
	}

	if rk == RecurrenceKindWeekly {
		if len(explicitDates) > 0 {
			return CreateSeriesInput{}, errExplicitDatesOnlyDateList()
		}
		return CreateSeriesInput{
			BlockType: bt, RecurrenceKind: rk, ISOWeekday: isoWeekday, StartsTime: startsTime,
			DurationMinutes: durationMinutes, EffectiveFrom: from, EffectiveUntil: until, Reason: r,
		}, nil
	}

	dates := make([]string, 0, len(explicitDates))
	for _, raw := range explicitDates {
		d, err := ValidateCivilDate(raw)
		if err != nil {
			return CreateSeriesInput{}, err
		}
		if d < from || (until != nil && d > *until) {
			return CreateSeriesInput{}, errDateOutOfSeriesRange()
		}
		dates = append(dates, d)
	}

	return CreateSeriesInput{
		BlockType: bt, RecurrenceKind: rk, ISOWeekday: nil, StartsTime: startsTime,
		DurationMinutes: durationMinutes, EffectiveFrom: from, EffectiveUntil: until, Reason: r,
		ExplicitDates: dates,
	}, nil
}

// UpdateSeries edita una serie existente (RN-BLQ-01: "esta y las
// siguientes" o "toda la serie"; "esta instancia" vive en el sub-recurso de
// excepciones). scope == UpdateScopeWhole reemplaza los campos editables de
// la cabecera completa. scope == UpdateScopeThisAndFollowing solo se admite
// sobre una serie weekly: trunca la serie original en splitEffectiveDate-1
// y crea una serie nueva desde splitEffectiveDate con los campos nuevos,
// devolviendo esa serie nueva. splitEffectiveDate se ignora cuando scope es
// whole.
func (s *Service) UpdateSeries(
	ctx context.Context,
	barbershopID, barberID, seriesID string,
	scope string,
	blockType, startsTimeRaw string,
	durationMinutes int,
	effectiveFrom string,
	effectiveUntil *string,
	reason *string,
	splitEffectiveDate string,
) (TimeBlockSeries, error) {
	if err := ctx.Err(); err != nil {
		return TimeBlockSeries{}, apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de editar serie: %w", err))
	}
	if !LooksLikeBarberID(barberID) || !LooksLikeSeriesID(seriesID) {
		return TimeBlockSeries{}, errTimeBlockSeriesNotFound()
	}
	if scope != UpdateScopeWhole && scope != UpdateScopeThisAndFollowing {
		return TimeBlockSeries{}, errUpdateScopeInvalid()
	}

	bt, err := ValidateBlockType(blockType)
	if err != nil {
		return TimeBlockSeries{}, err
	}
	startsTime, err := ValidateStartsTime(startsTimeRaw)
	if err != nil {
		return TimeBlockSeries{}, err
	}
	if durationMinutes < MinDurationMinutes || durationMinutes > MaxDurationMinutes {
		return TimeBlockSeries{}, errDurationInvalid()
	}
	from, err := ValidateCivilDate(effectiveFrom)
	if err != nil {
		return TimeBlockSeries{}, err
	}
	var until *string
	if effectiveUntil != nil {
		u, err := ValidateCivilDate(*effectiveUntil)
		if err != nil {
			return TimeBlockSeries{}, err
		}
		until = &u
	}
	if err := ValidateEffectiveRange(from, until); err != nil {
		return TimeBlockSeries{}, err
	}
	r, err := ValidateReason(reason)
	if err != nil {
		return TimeBlockSeries{}, err
	}

	input := UpdateSeriesInput{
		BlockType: bt, StartsTime: startsTime, DurationMinutes: durationMinutes,
		EffectiveFrom: from, EffectiveUntil: until, Reason: r,
	}

	if scope == UpdateScopeWhole {
		result, err := s.repo.UpdateSeriesWhole(ctx, barbershopID, barberID, seriesID, input)
		if err != nil {
			return TimeBlockSeries{}, apperr.Internal(fmt.Errorf("schedule: editar serie: %w", err))
		}
		if !result.Found {
			return TimeBlockSeries{}, errTimeBlockSeriesNotFound()
		}
		return result.Series, nil
	}

	// this_and_following: primero se lee la serie original para validar su
	// forma (weekly, rango vigente) ANTES de tocar la base; el propio
	// repositorio vuelve a bloquear la fila dentro de la transacción de
	// SplitSeriesFrom, así que esta lectura es solo de validación, no la
	// fuente de verdad transaccional.
	existing, found, err := s.repo.GetSeries(ctx, barbershopID, barberID, seriesID)
	if err != nil {
		return TimeBlockSeries{}, apperr.Internal(fmt.Errorf("schedule: leer serie antes de dividir: %w", err))
	}
	if !found || existing.DeletedAt != nil {
		return TimeBlockSeries{}, errTimeBlockSeriesNotFound()
	}
	if existing.RecurrenceKind != RecurrenceKindWeekly {
		return TimeBlockSeries{}, errSplitOnlyWeekly()
	}

	splitDate, err := ValidateCivilDate(splitEffectiveDate)
	if err != nil {
		return TimeBlockSeries{}, err
	}
	if splitDate <= existing.EffectiveFrom || (existing.EffectiveUntil != nil && splitDate > *existing.EffectiveUntil) {
		return TimeBlockSeries{}, errSplitDateOutOfRange()
	}

	result, err := s.repo.SplitSeriesFrom(ctx, barbershopID, barberID, seriesID, splitDate, input)
	if err != nil {
		return TimeBlockSeries{}, apperr.Internal(fmt.Errorf("schedule: dividir serie: %w", err))
	}
	if !result.Found {
		return TimeBlockSeries{}, errTimeBlockSeriesNotFound()
	}
	return result.Series, nil
}

// DeleteSeries retira lógicamente la serie completa (RN-BLQ-04). No borra
// sus fechas/excepciones (RESTRICT/DEC-070): quedan huérfanas de una serie
// retirada, mismo criterio de auditoría que un bloqueo puntual retirado.
func (s *Service) DeleteSeries(ctx context.Context, barbershopID, barberID, seriesID string) error {
	if err := ctx.Err(); err != nil {
		return apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de retirar serie: %w", err))
	}
	if !LooksLikeBarberID(barberID) || !LooksLikeSeriesID(seriesID) {
		return errTimeBlockSeriesNotFound()
	}
	found, err := s.repo.DeleteSeries(ctx, barbershopID, barberID, seriesID)
	if err != nil {
		return apperr.Internal(fmt.Errorf("schedule: retirar serie: %w", err))
	}
	if !found {
		return errTimeBlockSeriesNotFound()
	}
	return nil
}

// AddSeriesDate agrega una fecha explícita a una serie date_list (CA-042).
func (s *Service) AddSeriesDate(ctx context.Context, barbershopID, barberID, seriesID, blockDateRaw string) error {
	if err := ctx.Err(); err != nil {
		return apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de agregar fecha: %w", err))
	}
	if !LooksLikeBarberID(barberID) || !LooksLikeSeriesID(seriesID) {
		return errTimeBlockSeriesNotFound()
	}
	blockDate, err := ValidateCivilDate(blockDateRaw)
	if err != nil {
		return err
	}

	series, found, err := s.repo.GetSeries(ctx, barbershopID, barberID, seriesID)
	if err != nil {
		return apperr.Internal(fmt.Errorf("schedule: leer serie para agregar fecha: %w", err))
	}
	if !found || series.DeletedAt != nil {
		return errTimeBlockSeriesNotFound()
	}
	if series.RecurrenceKind != RecurrenceKindDateList {
		return errDateOutOfSeriesRange()
	}
	if blockDate < series.EffectiveFrom || (series.EffectiveUntil != nil && blockDate > *series.EffectiveUntil) {
		return errDateOutOfSeriesRange()
	}

	_, conflict, err := s.repo.AddSeriesDate(ctx, barbershopID, barberID, seriesID, blockDate)
	if err != nil {
		return apperr.Internal(fmt.Errorf("schedule: agregar fecha: %w", err))
	}
	if conflict {
		return errDuplicateSeriesDate()
	}
	return nil
}

// RemoveSeriesDate retira físicamente una fecha explícita de una serie
// date_list. found=false cubre "no existía"/"ya se había retirado":
// reintentar tras un 404 es seguro.
func (s *Service) RemoveSeriesDate(ctx context.Context, barbershopID, barberID, seriesID, blockDateRaw string) error {
	if err := ctx.Err(); err != nil {
		return apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de retirar fecha: %w", err))
	}
	if !LooksLikeBarberID(barberID) || !LooksLikeSeriesID(seriesID) {
		return errTimeBlockSeriesNotFound()
	}
	blockDate, err := ValidateCivilDate(blockDateRaw)
	if err != nil {
		return err
	}
	found, err := s.repo.RemoveSeriesDate(ctx, barbershopID, barberID, seriesID, blockDate)
	if err != nil {
		return apperr.Internal(fmt.Errorf("schedule: retirar fecha: %w", err))
	}
	if !found {
		return errSeriesDateNotFound()
	}
	return nil
}

// AddSeriesException suprime una instancia puntual de cualquier serie
// (weekly o date_list): "esta instancia no" (RN-BLQ-01).
func (s *Service) AddSeriesException(ctx context.Context, barbershopID, barberID, seriesID, excludedDateRaw string, reason *string) error {
	if err := ctx.Err(); err != nil {
		return apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de agregar excepción: %w", err))
	}
	if !LooksLikeBarberID(barberID) || !LooksLikeSeriesID(seriesID) {
		return errTimeBlockSeriesNotFound()
	}
	excludedDate, err := ValidateCivilDate(excludedDateRaw)
	if err != nil {
		return err
	}
	r, err := ValidateReason(reason)
	if err != nil {
		return err
	}

	found, conflict, err := s.repo.AddSeriesException(ctx, barbershopID, barberID, seriesID, excludedDate, r)
	if err != nil {
		return apperr.Internal(fmt.Errorf("schedule: agregar excepción: %w", err))
	}
	if !found {
		return errTimeBlockSeriesNotFound()
	}
	if conflict {
		return errDuplicateSeriesException()
	}
	return nil
}

// RemoveSeriesException retira físicamente una excepción ("restaura" la
// instancia suprimida). found=false cubre "no existía"/"ya se había
// retirado": reintentar tras un 404 es seguro.
func (s *Service) RemoveSeriesException(ctx context.Context, barbershopID, barberID, seriesID, excludedDateRaw string) error {
	if err := ctx.Err(); err != nil {
		return apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de retirar excepción: %w", err))
	}
	if !LooksLikeBarberID(barberID) || !LooksLikeSeriesID(seriesID) {
		return errTimeBlockSeriesNotFound()
	}
	excludedDate, err := ValidateCivilDate(excludedDateRaw)
	if err != nil {
		return err
	}
	found, err := s.repo.RemoveSeriesException(ctx, barbershopID, barberID, seriesID, excludedDate)
	if err != nil {
		return apperr.Internal(fmt.Errorf("schedule: retirar excepción: %w", err))
	}
	if !found {
		return errSeriesExceptionNotFound()
	}
	return nil
}

// ---------------------------------------------------------------------
// Proyección efectiva (HU-042 §"Trabajo requerido, punto 1.5")
// ---------------------------------------------------------------------

// SeriesOccurrence es una ocurrencia expandida de una serie dentro del
// rango consultado, en forma civil (fecha + hora local de la barbería):
// mismo criterio que working_hour/schedule-exceptions, que tampoco emiten
// instantes absolutos porque el cliente ya conoce la zona IANA de la
// barbería (GET .../barbershop).
type SeriesOccurrence struct {
	SeriesID        string
	BlockType       string
	Date            string
	StartsTime      string
	DurationMinutes int
	Reason          *string
}

// EffectiveBlocksResult separa deliberadamente los bloqueos puntuales (ya
// timestamptz) de las ocurrencias de serie expandidas (civiles): unirlas en
// una sola línea de tiempo exige la zona IANA de la barbería para convertir
// las segundas a instantes absolutos, y ningún consumidor la necesita
// todavía (B3/B4 no existen aún). Documentado en el prompt HU-042, sección
// "Qué NO hace esta HU": la proyección efectiva final es responsabilidad
// del futuro consumidor, no de esta operación interna.
type EffectiveBlocksResult struct {
	ManualBlocks      []TimeBlock
	SeriesOccurrences []SeriesOccurrence
}

// EffectiveBlocks calcula, para el barbero de barberID, los bloqueos
// puntuales vigentes y las ocurrencias de serie dentro de [from, to]
// (fechas civiles, ambos límites inclusive). Es la "operación interna de
// proyección/consulta para disponibilidad futura sin materializar
// instancias de una serie en time_block" que el prompt de HU-042 exige.
func (s *Service) EffectiveBlocks(ctx context.Context, barbershopID, barberID, fromRaw, toRaw string) (EffectiveBlocksResult, error) {
	if err := ctx.Err(); err != nil {
		return EffectiveBlocksResult{}, apperr.Internal(fmt.Errorf("schedule: contexto cancelado antes de proyectar bloqueos: %w", err))
	}
	if !LooksLikeBarberID(barberID) {
		return EffectiveBlocksResult{}, errBarberNotFound()
	}

	fromDate, err := ValidateCivilDate(fromRaw)
	if err != nil {
		return EffectiveBlocksResult{}, err
	}
	toDate, err := ValidateCivilDate(toRaw)
	if err != nil {
		return EffectiveBlocksResult{}, err
	}
	if toDate < fromDate {
		return EffectiveBlocksResult{}, errQueryRangeInvalid()
	}

	exists, err := s.barbers.Exists(ctx, barbershopID, barberID)
	if err != nil {
		return EffectiveBlocksResult{}, apperr.Internal(fmt.Errorf("schedule: verificar barbero para proyectar bloqueos: %w", err))
	}
	if !exists {
		return EffectiveBlocksResult{}, errBarberNotFound()
	}

	fromInstant, err := time.Parse(dateLayout, fromDate)
	if err != nil {
		return EffectiveBlocksResult{}, apperr.Internal(fmt.Errorf("schedule: parsear from ya validado: %w", err))
	}
	toInstant, err := time.Parse(dateLayout, toDate)
	if err != nil {
		return EffectiveBlocksResult{}, apperr.Internal(fmt.Errorf("schedule: parsear to ya validado: %w", err))
	}
	untilExclusive := toInstant.AddDate(0, 0, 1)

	manual, err := s.repo.ListEffectiveManualBlocks(ctx, barbershopID, barberID, fromInstant, untilExclusive)
	if err != nil {
		return EffectiveBlocksResult{}, apperr.Internal(fmt.Errorf("schedule: listar bloqueos puntuales vigentes: %w", err))
	}

	series, err := s.repo.ListActiveSeriesForProjection(ctx, barbershopID, barberID, fromDate, toDate)
	if err != nil {
		return EffectiveBlocksResult{}, apperr.Internal(fmt.Errorf("schedule: listar series activas: %w", err))
	}

	return EffectiveBlocksResult{
		ManualBlocks:      manual,
		SeriesOccurrences: expandSeriesOccurrences(series, fromDate, toDate),
	}, nil
}

// expandSeriesOccurrences expande cada serie weekly/date_list dentro de
// [fromDate, toDate] (civil, ambos límites inclusive), recortado además por
// el rango efectivo propio de cada serie, y excluye las fechas de
// Exceptions. No fusiona con bloqueos puntuales ni con otra serie que se
// solape (ver EffectiveBlocksResult).
func expandSeriesOccurrences(series []TimeBlockSeries, fromDate, toDate string) []SeriesOccurrence {
	var out []SeriesOccurrence
	from, err := time.Parse(dateLayout, fromDate)
	if err != nil {
		return nil
	}
	to, err := time.Parse(dateLayout, toDate)
	if err != nil {
		return nil
	}

	for _, sr := range series {
		if sr.DeletedAt != nil {
			continue
		}
		excluded := make(map[string]bool, len(sr.Exceptions))
		for _, ex := range sr.Exceptions {
			excluded[ex.ExcludedDate] = true
		}

		rangeStart := from
		if sr.EffectiveFrom > fromDate {
			parsed, err := time.Parse(dateLayout, sr.EffectiveFrom)
			if err != nil {
				continue
			}
			rangeStart = parsed
		}
		rangeEnd := to
		if sr.EffectiveUntil != nil && *sr.EffectiveUntil < toDate {
			parsed, err := time.Parse(dateLayout, *sr.EffectiveUntil)
			if err != nil {
				continue
			}
			rangeEnd = parsed
		}
		if rangeStart.After(rangeEnd) {
			continue
		}

		switch sr.RecurrenceKind {
		case RecurrenceKindWeekly:
			if sr.ISOWeekday == nil {
				continue
			}
			for d := rangeStart; !d.After(rangeEnd); d = d.AddDate(0, 0, 1) {
				if isoWeekdayOf(d) != *sr.ISOWeekday {
					continue
				}
				dateStr := d.Format(dateLayout)
				if excluded[dateStr] {
					continue
				}
				out = append(out, SeriesOccurrence{
					SeriesID: sr.ID, BlockType: sr.BlockType, Date: dateStr,
					StartsTime: sr.StartsTime, DurationMinutes: sr.DurationMinutes, Reason: sr.Reason,
				})
			}
		case RecurrenceKindDateList:
			rangeStartStr, rangeEndStr := rangeStart.Format(dateLayout), rangeEnd.Format(dateLayout)
			for _, sd := range sr.Dates {
				if sd.BlockDate < rangeStartStr || sd.BlockDate > rangeEndStr {
					continue
				}
				if excluded[sd.BlockDate] {
					continue
				}
				out = append(out, SeriesOccurrence{
					SeriesID: sr.ID, BlockType: sr.BlockType, Date: sd.BlockDate,
					StartsTime: sr.StartsTime, DurationMinutes: sr.DurationMinutes, Reason: sr.Reason,
				})
			}
		}
	}
	return out
}

// isoWeekdayOf convierte time.Weekday (0=domingo…6=sábado) a ISO
// (1=lunes…7=domingo, DEC-020).
func isoWeekdayOf(d time.Time) int {
	wd := int(d.Weekday())
	if wd == 0 {
		return 7
	}
	return wd
}
