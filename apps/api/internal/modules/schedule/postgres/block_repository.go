package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"system-barbershop/internal/modules/schedule"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

// createTimeBlockOperation/createTimeBlockSeriesOperation identifican,
// dentro de una barbería, las operaciones de idempotencia de HU-042
// (RN-IDE-01): una clave ya usada para OTRA operación (incluidas las de
// HU-040/HU-041) nunca se confunde con estas.
const (
	createTimeBlockOperation       idempotency.Operation = "create_time_block"
	createTimeBlockSeriesOperation idempotency.Operation = "create_time_block_series"
	timeBlockIdempotencyTTL                              = 24 * time.Hour
)

// timeBlockSeriesDatePKConstraint/timeBlockSeriesExceptionPKConstraint son
// los nombres de las PK compuestas de
// 20260826100000_create_time_block.sql: un unique_violation sobre CUALQUIER
// otra restricción se propaga como error interno, mismo criterio que
// workingHourWeekdayStartUniqueConstraint.
const (
	timeBlockSeriesDatePKConstraint      = "time_block_series_date_pk"
	timeBlockSeriesExceptionPKConstraint = "time_block_series_exception_pk"
)

var errSeriesDateConflictInternal = errors.New("schedule/postgres: la fecha ya está registrada en la serie")
var errSeriesExceptionConflictInternal = errors.New("schedule/postgres: la excepción ya está registrada en la serie")

// isConstraintViolation(err, code, constraint) ya existe en
// exception_repository.go: se reutiliza tal cual en vez de duplicarla.

// ---------------------------------------------------------------------
// Wire shapes — DEBEN coincidir byte a byte con httpapi (ver ADVERTENCIA
// en repository.go sobre workingHourResponseWire).
// ---------------------------------------------------------------------

type timeBlockResponseWire struct {
	ID        string     `json:"id"`
	BlockType string     `json:"blockType"`
	Source    string     `json:"source"`
	StartsAt  time.Time  `json:"startsAt"`
	EndsAt    time.Time  `json:"endsAt"`
	Reason    *string    `json:"reason"`
	DeletedAt *time.Time `json:"deletedAt"`
	DeletedBy *string    `json:"deletedBy"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func marshalTimeBlockResponse(b schedule.TimeBlock) ([]byte, error) {
	return json.Marshal(timeBlockResponseWire{
		ID: b.ID, BlockType: b.BlockType, Source: b.Source, StartsAt: b.StartsAt, EndsAt: b.EndsAt,
		Reason: b.Reason, DeletedAt: b.DeletedAt, DeletedBy: b.DeletedBy, CreatedAt: b.CreatedAt, UpdatedAt: b.UpdatedAt,
	})
}

type seriesDateWire struct {
	BlockDate string `json:"blockDate"`
}

type seriesExceptionWire struct {
	ExcludedDate string    `json:"excludedDate"`
	Reason       *string   `json:"reason"`
	CreatedAt    time.Time `json:"createdAt"`
}

type timeBlockSeriesResponseWire struct {
	ID              string                `json:"id"`
	BlockType       string                `json:"blockType"`
	RecurrenceKind  string                `json:"recurrenceKind"`
	ISOWeekday      *int                  `json:"isoWeekday"`
	StartsTime      string                `json:"startsTime"`
	DurationMinutes int                   `json:"durationMinutes"`
	EffectiveFrom   string                `json:"effectiveFrom"`
	EffectiveUntil  *string               `json:"effectiveUntil"`
	Reason          *string               `json:"reason"`
	DeletedAt       *time.Time            `json:"deletedAt"`
	CreatedAt       time.Time             `json:"createdAt"`
	UpdatedAt       time.Time             `json:"updatedAt"`
	Dates           []seriesDateWire      `json:"dates"`
	Exceptions      []seriesExceptionWire `json:"exceptions"`
}

func marshalTimeBlockSeriesResponse(sr schedule.TimeBlockSeries) ([]byte, error) {
	dates := make([]seriesDateWire, 0, len(sr.Dates))
	for _, d := range sr.Dates {
		dates = append(dates, seriesDateWire{BlockDate: d.BlockDate})
	}
	exceptions := make([]seriesExceptionWire, 0, len(sr.Exceptions))
	for _, e := range sr.Exceptions {
		exceptions = append(exceptions, seriesExceptionWire{ExcludedDate: e.ExcludedDate, Reason: e.Reason, CreatedAt: e.CreatedAt})
	}
	return json.Marshal(timeBlockSeriesResponseWire{
		ID: sr.ID, BlockType: sr.BlockType, RecurrenceKind: sr.RecurrenceKind, ISOWeekday: sr.ISOWeekday,
		StartsTime: sr.StartsTime, DurationMinutes: sr.DurationMinutes, EffectiveFrom: sr.EffectiveFrom,
		EffectiveUntil: sr.EffectiveUntil, Reason: sr.Reason, DeletedAt: sr.DeletedAt,
		CreatedAt: sr.CreatedAt, UpdatedAt: sr.UpdatedAt, Dates: dates, Exceptions: exceptions,
	})
}

// ---------------------------------------------------------------------
// Bloqueos puntuales
// ---------------------------------------------------------------------

func scanTimeBlock(row interface {
	Scan(dest ...any) error
}) (schedule.TimeBlock, error) {
	var b schedule.TimeBlock
	err := row.Scan(&b.ID, &b.BlockType, &b.Source, &b.StartsAt, &b.EndsAt, &b.Reason, &b.DeletedAt, &b.DeletedBy, &b.CreatedAt, &b.UpdatedAt)
	return b, err
}

const timeBlockColumns = `id, block_type, source, starts_at, ends_at, reason, deleted_at, deleted_by, created_at, updated_at`

// ListBlocks implementa schedule.BlockRepository.ListBlocks: orden estable
// (starts_at, id), paginación "pedir uno de más" (mismo patrón que
// Repository.List de working_hour).
func (r *Repository) ListBlocks(ctx context.Context, barbershopID, barberID string, cursor *schedule.BlockCursor, limit int, includeDeleted bool) (schedule.BlockListResult, error) {
	var result schedule.BlockListResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		fetchLimit := limit + 1
		deletedClause := "AND deleted_at IS NULL"
		if includeDeleted {
			deletedClause = ""
		}

		var (
			rows pgx.Rows
			err  error
		)
		if cursor == nil {
			rows, err = q.Query(ctx,
				`SELECT `+timeBlockColumns+`
				   FROM time_block
				  WHERE barbershop_id = $1 AND barber_id = $2 `+deletedClause+`
				  ORDER BY starts_at, id
				  LIMIT $3`,
				barbershopID, barberID, fetchLimit,
			)
		} else {
			rows, err = q.Query(ctx,
				`SELECT `+timeBlockColumns+`
				   FROM time_block
				  WHERE barbershop_id = $1 AND barber_id = $2 `+deletedClause+`
				    AND (starts_at, id) > ($3::timestamptz, $4)
				  ORDER BY starts_at, id
				  LIMIT $5`,
				barbershopID, barberID, cursor.StartsAt, cursor.ID, fetchLimit,
			)
		}
		if err != nil {
			return fmt.Errorf("list time blocks: query: %w", err)
		}
		defer rows.Close()

		items := make([]schedule.TimeBlock, 0, fetchLimit)
		for rows.Next() {
			b, err := scanTimeBlock(rows)
			if err != nil {
				return fmt.Errorf("list time blocks: scan: %w", err)
			}
			items = append(items, b)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("list time blocks: rows: %w", err)
		}

		hasMore := len(items) > limit
		if hasMore {
			items = items[:limit]
		}
		result.Items = items
		if hasMore {
			last := items[len(items)-1]
			result.NextCursor = schedule.EncodeBlockCursor(schedule.BlockCursor{StartsAt: last.StartsAt.Format(time.RFC3339Nano), ID: last.ID})
		}
		return nil
	})
	if err != nil {
		return schedule.BlockListResult{}, fmt.Errorf("schedule/postgres: list time blocks: %w", err)
	}
	return result, nil
}

// GetBlock implementa schedule.BlockRepository.GetBlock.
func (r *Repository) GetBlock(ctx context.Context, barbershopID, barberID, blockID string) (schedule.TimeBlock, bool, error) {
	var b schedule.TimeBlock
	found := false

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		row := q.QueryRow(ctx,
			`SELECT `+timeBlockColumns+`
			   FROM time_block
			  WHERE id = $1 AND barbershop_id = $2 AND barber_id = $3`,
			blockID, barbershopID, barberID,
		)
		scanned, err := scanTimeBlock(row)
		switch {
		case err == nil:
			b = scanned
			found = true
		case errors.Is(err, pgx.ErrNoRows):
		default:
			return fmt.Errorf("select time block: %w", err)
		}
		return nil
	})
	if err != nil {
		return schedule.TimeBlock{}, false, fmt.Errorf("schedule/postgres: get time block: %w", err)
	}
	return b, found, nil
}

// CreateBlock implementa schedule.BlockRepository.CreateBlock: Begin, el
// INSERT y Complete dentro de UNA sola InTenantTx. Nunca verifica solape
// (RN-BLQ-03/DEC-008): un bloqueo puntual siempre se crea.
func (r *Repository) CreateBlock(
	ctx context.Context,
	barbershopID, barberID string,
	input schedule.CreateBlockInput,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (schedule.CreateBlockResult, error) {
	var result schedule.CreateBlockResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		decision, err := r.coord.Begin(ctx, q, database.BarbershopID(barbershopID), key, createTimeBlockOperation, fingerprint, timeBlockIdempotencyTTL)
		if err != nil {
			return fmt.Errorf("idempotency begin: %w", err)
		}
		result.Decision = decision
		if decision.Outcome != idempotency.OutcomeProceed {
			if decision.Outcome == idempotency.OutcomeReplay {
				result.Response = decision.Response
			}
			return nil
		}

		row := q.QueryRow(ctx,
			`INSERT INTO time_block (barbershop_id, barber_id, block_type, source, starts_at, ends_at, reason)
			      VALUES ($1, $2, $3, 'manual', $4, $5, $6)
			   RETURNING `+timeBlockColumns,
			barbershopID, barberID, input.BlockType, input.StartsAt, input.EndsAt, input.Reason,
		)
		b, err := scanTimeBlock(row)
		if err != nil {
			return fmt.Errorf("insert time block: %w", err)
		}

		body, err := marshalTimeBlockResponse(b)
		if err != nil {
			return fmt.Errorf("marshal created time block: %w", err)
		}
		stored := idempotency.StoredResponse{Status: 201, ContentType: "application/json", Body: string(body)}

		ok, err := r.coord.Complete(ctx, q, database.BarbershopID(barbershopID), key, stored)
		if err != nil {
			return fmt.Errorf("idempotency complete: %w", err)
		}
		if !ok {
			return fmt.Errorf("idempotency complete: la reclamación ya no estaba in_progress")
		}

		result.Block = b
		result.Response = stored
		return nil
	})
	if err != nil {
		return schedule.CreateBlockResult{}, fmt.Errorf("schedule/postgres: create time block: %w", err)
	}
	return result, nil
}

// DeleteBlock implementa schedule.BlockRepository.DeleteBlock: retiro
// lógico (RN-BLQ-04), nunca DELETE físico. found=false cubre "no
// existe"/"de otro barbero"/"de otra barbería"/"ya estaba retirado".
func (r *Repository) DeleteBlock(ctx context.Context, barbershopID, barberID, blockID, actorID string) (bool, error) {
	found := false

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		tag, err := q.Exec(ctx,
			`UPDATE time_block
			    SET deleted_at = now(), deleted_by = $4
			  WHERE id = $1 AND barbershop_id = $2 AND barber_id = $3 AND deleted_at IS NULL`,
			blockID, barbershopID, barberID, actorID,
		)
		if err != nil {
			return fmt.Errorf("soft delete time block: %w", err)
		}
		found = tag.RowsAffected() > 0
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("schedule/postgres: delete time block: %w", err)
	}
	return found, nil
}

// ListEffectiveManualBlocks implementa
// schedule.BlockRepository.ListEffectiveManualBlocks.
func (r *Repository) ListEffectiveManualBlocks(ctx context.Context, barbershopID, barberID string, from, until time.Time) ([]schedule.TimeBlock, error) {
	var items []schedule.TimeBlock

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		rows, err := q.Query(ctx,
			`SELECT `+timeBlockColumns+`
			   FROM time_block
			  WHERE barbershop_id = $1 AND barber_id = $2 AND deleted_at IS NULL
			    AND starts_at < $4 AND ends_at > $3
			  ORDER BY starts_at`,
			barbershopID, barberID, from, until,
		)
		if err != nil {
			return fmt.Errorf("list effective manual blocks: query: %w", err)
		}
		defer rows.Close()
		items = make([]schedule.TimeBlock, 0)
		for rows.Next() {
			b, err := scanTimeBlock(rows)
			if err != nil {
				return fmt.Errorf("list effective manual blocks: scan: %w", err)
			}
			items = append(items, b)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("schedule/postgres: list effective manual blocks: %w", err)
	}
	return items, nil
}

// ---------------------------------------------------------------------
// Series recurrentes
// ---------------------------------------------------------------------

// seriesSelectSQL usa to_char para starts_time/effective_from/effective_until
// en vez de depender del formato de escaneo de pgx sobre `time`/`date`
// (mismo criterio que to_char(starts_time, 'HH24:MI') en working_hour): un
// scan directo a time.Time exige declarar el tipo exacto que pgx usa para
// `time`/`date`, que varía según el driver; to_char es explícito y estable.
const timeBlockSeriesSelectSQL = `SELECT id, block_type, recurrence_kind, iso_weekday,
       to_char(starts_time, 'HH24:MI') AS starts_time, duration_minutes,
       to_char(effective_from, 'YYYY-MM-DD') AS effective_from,
       to_char(effective_until, 'YYYY-MM-DD') AS effective_until,
       reason, deleted_at, created_at, updated_at
  FROM time_block_series`

func scanTimeBlockSeriesText(row interface {
	Scan(dest ...any) error
}) (schedule.TimeBlockSeries, error) {
	var sr schedule.TimeBlockSeries
	var effectiveUntil *string
	err := row.Scan(
		&sr.ID, &sr.BlockType, &sr.RecurrenceKind, &sr.ISOWeekday, &sr.StartsTime, &sr.DurationMinutes,
		&sr.EffectiveFrom, &effectiveUntil, &sr.Reason, &sr.DeletedAt, &sr.CreatedAt, &sr.UpdatedAt,
	)
	sr.EffectiveUntil = effectiveUntil
	return sr, err
}

// loadSeriesChildren rellena Dates/Exceptions de cada serie en items
// (mismo tenant/barbero ya filtrado por el llamador). Dos consultas por
// lote, no una por serie: el número de series de una página está acotado
// por MaxListLimit.
func loadSeriesChildren(ctx context.Context, q database.Queries, barbershopID string, items []schedule.TimeBlockSeries) error {
	if len(items) == 0 {
		return nil
	}
	byID := make(map[string]*schedule.TimeBlockSeries, len(items))
	ids := make([]string, 0, len(items))
	for i := range items {
		byID[items[i].ID] = &items[i]
		ids = append(ids, items[i].ID)
	}

	dateRows, err := q.Query(ctx,
		`SELECT series_id, to_char(block_date, 'YYYY-MM-DD')
		   FROM time_block_series_date
		  WHERE barbershop_id = $1 AND series_id = ANY($2)
		  ORDER BY series_id, block_date`,
		barbershopID, ids,
	)
	if err != nil {
		return fmt.Errorf("load series dates: %w", err)
	}
	defer dateRows.Close()
	for dateRows.Next() {
		var seriesID, blockDate string
		if err := dateRows.Scan(&seriesID, &blockDate); err != nil {
			return fmt.Errorf("scan series date: %w", err)
		}
		if sr, ok := byID[seriesID]; ok {
			sr.Dates = append(sr.Dates, schedule.SeriesDate{BlockDate: blockDate})
		}
	}
	if err := dateRows.Err(); err != nil {
		return fmt.Errorf("iterate series dates: %w", err)
	}

	excRows, err := q.Query(ctx,
		`SELECT series_id, to_char(excluded_date, 'YYYY-MM-DD'), reason, created_at
		   FROM time_block_series_exception
		  WHERE barbershop_id = $1 AND series_id = ANY($2)
		  ORDER BY series_id, excluded_date`,
		barbershopID, ids,
	)
	if err != nil {
		return fmt.Errorf("load series exceptions: %w", err)
	}
	defer excRows.Close()
	for excRows.Next() {
		var seriesID string
		var ex schedule.SeriesException
		if err := excRows.Scan(&seriesID, &ex.ExcludedDate, &ex.Reason, &ex.CreatedAt); err != nil {
			return fmt.Errorf("scan series exception: %w", err)
		}
		if sr, ok := byID[seriesID]; ok {
			sr.Exceptions = append(sr.Exceptions, ex)
		}
	}
	return excRows.Err()
}

// ListSeries implementa schedule.BlockRepository.ListSeries.
func (r *Repository) ListSeries(ctx context.Context, barbershopID, barberID string, cursor *schedule.SeriesCursor, limit int, includeDeleted bool) (schedule.SeriesListResult, error) {
	var result schedule.SeriesListResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		fetchLimit := limit + 1
		deletedClause := "AND deleted_at IS NULL"
		if includeDeleted {
			deletedClause = ""
		}

		var (
			rows pgx.Rows
			err  error
		)
		if cursor == nil {
			rows, err = q.Query(ctx,
				timeBlockSeriesSelectSQL+`
				  WHERE barbershop_id = $1 AND barber_id = $2 `+deletedClause+`
				  ORDER BY effective_from, id
				  LIMIT $3`,
				barbershopID, barberID, fetchLimit,
			)
		} else {
			rows, err = q.Query(ctx,
				timeBlockSeriesSelectSQL+`
				  WHERE barbershop_id = $1 AND barber_id = $2 `+deletedClause+`
				    AND (effective_from, id) > ($3::date, $4)
				  ORDER BY effective_from, id
				  LIMIT $5`,
				barbershopID, barberID, cursor.EffectiveFrom, cursor.ID, fetchLimit,
			)
		}
		if err != nil {
			return fmt.Errorf("list time block series: query: %w", err)
		}
		defer rows.Close()

		items := make([]schedule.TimeBlockSeries, 0, fetchLimit)
		for rows.Next() {
			sr, err := scanTimeBlockSeriesText(rows)
			if err != nil {
				return fmt.Errorf("list time block series: scan: %w", err)
			}
			items = append(items, sr)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("list time block series: rows: %w", err)
		}

		hasMore := len(items) > limit
		if hasMore {
			items = items[:limit]
		}
		if err := loadSeriesChildren(ctx, q, barbershopID, items); err != nil {
			return err
		}

		result.Items = items
		if hasMore {
			last := items[len(items)-1]
			result.NextCursor = schedule.EncodeSeriesCursor(schedule.SeriesCursor{EffectiveFrom: last.EffectiveFrom, ID: last.ID})
		}
		return nil
	})
	if err != nil {
		return schedule.SeriesListResult{}, fmt.Errorf("schedule/postgres: list time block series: %w", err)
	}
	return result, nil
}

// GetSeries implementa schedule.BlockRepository.GetSeries.
func (r *Repository) GetSeries(ctx context.Context, barbershopID, barberID, seriesID string) (schedule.TimeBlockSeries, bool, error) {
	var sr schedule.TimeBlockSeries
	found := false

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		row := q.QueryRow(ctx,
			timeBlockSeriesSelectSQL+` WHERE id = $1 AND barbershop_id = $2 AND barber_id = $3`,
			seriesID, barbershopID, barberID,
		)
		scanned, err := scanTimeBlockSeriesText(row)
		switch {
		case err == nil:
			sr = scanned
			found = true
		case errors.Is(err, pgx.ErrNoRows):
		default:
			return fmt.Errorf("select time block series: %w", err)
		}
		if !found {
			return nil
		}
		items := []schedule.TimeBlockSeries{sr}
		if err := loadSeriesChildren(ctx, q, barbershopID, items); err != nil {
			return err
		}
		sr = items[0]
		return nil
	})
	if err != nil {
		return schedule.TimeBlockSeries{}, false, fmt.Errorf("schedule/postgres: get time block series: %w", err)
	}
	return sr, found, nil
}

// CreateSeries implementa schedule.BlockRepository.CreateSeries: Begin, el
// INSERT de la cabecera, el INSERT de cada fecha explícita y Complete
// dentro de UNA sola InTenantTx.
func (r *Repository) CreateSeries(
	ctx context.Context,
	barbershopID, barberID string,
	input schedule.CreateSeriesInput,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (schedule.CreateSeriesResult, error) {
	var result schedule.CreateSeriesResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		decision, err := r.coord.Begin(ctx, q, database.BarbershopID(barbershopID), key, createTimeBlockSeriesOperation, fingerprint, timeBlockIdempotencyTTL)
		if err != nil {
			return fmt.Errorf("idempotency begin: %w", err)
		}
		result.Decision = decision
		if decision.Outcome != idempotency.OutcomeProceed {
			if decision.Outcome == idempotency.OutcomeReplay {
				result.Response = decision.Response
			}
			return nil
		}

		row := q.QueryRow(ctx,
			`INSERT INTO time_block_series
			        (barbershop_id, barber_id, block_type, recurrence_kind, iso_weekday, starts_time, duration_minutes, effective_from, effective_until, reason)
			 VALUES ($1, $2, $3, $4, $5, $6::time, $7, $8::date, $9::date, $10)
			 RETURNING id, block_type, recurrence_kind, iso_weekday, to_char(starts_time, 'HH24:MI'), duration_minutes,
			           to_char(effective_from, 'YYYY-MM-DD'), to_char(effective_until, 'YYYY-MM-DD'), reason, deleted_at, created_at, updated_at`,
			barbershopID, barberID, input.BlockType, input.RecurrenceKind, input.ISOWeekday, input.StartsTime,
			input.DurationMinutes, input.EffectiveFrom, input.EffectiveUntil, input.Reason,
		)
		sr, err := scanTimeBlockSeriesText(row)
		if err != nil {
			return fmt.Errorf("insert time block series: %w", err)
		}

		for _, d := range input.ExplicitDates {
			if _, err := q.Exec(ctx,
				`INSERT INTO time_block_series_date (barbershop_id, series_id, block_date) VALUES ($1, $2, $3::date)`,
				barbershopID, sr.ID, d,
			); err != nil {
				return fmt.Errorf("insert series explicit date: %w", err)
			}
			sr.Dates = append(sr.Dates, schedule.SeriesDate{BlockDate: d})
		}

		body, err := marshalTimeBlockSeriesResponse(sr)
		if err != nil {
			return fmt.Errorf("marshal created time block series: %w", err)
		}
		stored := idempotency.StoredResponse{Status: 201, ContentType: "application/json", Body: string(body)}

		ok, err := r.coord.Complete(ctx, q, database.BarbershopID(barbershopID), key, stored)
		if err != nil {
			return fmt.Errorf("idempotency complete: %w", err)
		}
		if !ok {
			return fmt.Errorf("idempotency complete: la reclamación ya no estaba in_progress")
		}

		result.Series = sr
		result.Response = stored
		return nil
	})
	if err != nil {
		return schedule.CreateSeriesResult{}, fmt.Errorf("schedule/postgres: create time block series: %w", err)
	}
	return result, nil
}

// UpdateSeriesWhole implementa schedule.BlockRepository.UpdateSeriesWhole.
func (r *Repository) UpdateSeriesWhole(ctx context.Context, barbershopID, barberID, seriesID string, input schedule.UpdateSeriesInput) (schedule.UpdateSeriesResult, error) {
	var result schedule.UpdateSeriesResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		row := q.QueryRow(ctx,
			`UPDATE time_block_series
			    SET block_type = $4, starts_time = $5::time, duration_minutes = $6,
			        effective_from = $7::date, effective_until = $8::date, reason = $9
			  WHERE id = $1 AND barbershop_id = $2 AND barber_id = $3 AND deleted_at IS NULL
			RETURNING id, block_type, recurrence_kind, iso_weekday, to_char(starts_time, 'HH24:MI'), duration_minutes,
			          to_char(effective_from, 'YYYY-MM-DD'), to_char(effective_until, 'YYYY-MM-DD'), reason, deleted_at, created_at, updated_at`,
			seriesID, barbershopID, barberID, input.BlockType, input.StartsTime, input.DurationMinutes,
			input.EffectiveFrom, input.EffectiveUntil, input.Reason,
		)
		sr, err := scanTimeBlockSeriesText(row)
		switch {
		case err == nil:
			result.Found = true
		case errors.Is(err, pgx.ErrNoRows):
			return nil
		default:
			return fmt.Errorf("update time block series: %w", err)
		}
		items := []schedule.TimeBlockSeries{sr}
		if err := loadSeriesChildren(ctx, q, barbershopID, items); err != nil {
			return err
		}
		result.Series = items[0]
		return nil
	})
	if err != nil {
		return schedule.UpdateSeriesResult{}, fmt.Errorf("schedule/postgres: update time block series: %w", err)
	}
	return result, nil
}

// SplitSeriesFrom implementa schedule.BlockRepository.SplitSeriesFrom
// ("esta y las siguientes", RN-BLQ-01): bloquea la serie original,
// captura su iso_weekday y su effective_until ORIGINAL antes de truncarla,
// la trunca en effectiveDate-1, e inserta la serie nueva desde
// effectiveDate hasta ese effective_until original. Devuelve la serie
// NUEVA. found=false cubre "no existe"/"de otro barbero"/"de otra
// barbería"/"ya retirada".
func (r *Repository) SplitSeriesFrom(ctx context.Context, barbershopID, barberID, seriesID, effectiveDate string, input schedule.UpdateSeriesInput) (schedule.UpdateSeriesResult, error) {
	var result schedule.UpdateSeriesResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var isoWeekday *int
		var originalUntil *string
		row := q.QueryRow(ctx,
			`SELECT iso_weekday, to_char(effective_until, 'YYYY-MM-DD')
			   FROM time_block_series
			  WHERE id = $1 AND barbershop_id = $2 AND barber_id = $3 AND deleted_at IS NULL
			    FOR UPDATE`,
			seriesID, barbershopID, barberID,
		)
		if err := row.Scan(&isoWeekday, &originalUntil); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("lock original series: %w", err)
		}

		if _, err := q.Exec(ctx,
			`UPDATE time_block_series SET effective_until = ($4::date - INTERVAL '1 day')::date
			  WHERE id = $1 AND barbershop_id = $2 AND barber_id = $3`,
			seriesID, barbershopID, barberID, effectiveDate,
		); err != nil {
			return fmt.Errorf("truncate original series: %w", err)
		}

		newRow := q.QueryRow(ctx,
			`INSERT INTO time_block_series
			        (barbershop_id, barber_id, block_type, recurrence_kind, iso_weekday, starts_time, duration_minutes, effective_from, effective_until, reason)
			 VALUES ($1, $2, $3, 'weekly', $4, $5::time, $6, $7::date, $8::date, $9)
			 RETURNING id, block_type, recurrence_kind, iso_weekday, to_char(starts_time, 'HH24:MI'), duration_minutes,
			           to_char(effective_from, 'YYYY-MM-DD'), to_char(effective_until, 'YYYY-MM-DD'), reason, deleted_at, created_at, updated_at`,
			barbershopID, barberID, input.BlockType, isoWeekday, input.StartsTime, input.DurationMinutes,
			effectiveDate, originalUntil, input.Reason,
		)
		sr, err := scanTimeBlockSeriesText(newRow)
		if err != nil {
			return fmt.Errorf("insert split series: %w", err)
		}
		result.Found = true
		result.Series = sr
		return nil
	})
	if err != nil {
		return schedule.UpdateSeriesResult{}, fmt.Errorf("schedule/postgres: split time block series: %w", err)
	}
	return result, nil
}

// DeleteSeries implementa schedule.BlockRepository.DeleteSeries: retiro
// lógico de la cabecera (RN-BLQ-04). No toca dates/exceptions (RESTRICT).
func (r *Repository) DeleteSeries(ctx context.Context, barbershopID, barberID, seriesID string) (bool, error) {
	found := false

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		tag, err := q.Exec(ctx,
			`UPDATE time_block_series SET deleted_at = now()
			  WHERE id = $1 AND barbershop_id = $2 AND barber_id = $3 AND deleted_at IS NULL`,
			seriesID, barbershopID, barberID,
		)
		if err != nil {
			return fmt.Errorf("soft delete time block series: %w", err)
		}
		found = tag.RowsAffected() > 0
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("schedule/postgres: delete time block series: %w", err)
	}
	return found, nil
}

// AddSeriesDate implementa schedule.BlockRepository.AddSeriesDate.
func (r *Repository) AddSeriesDate(ctx context.Context, barbershopID, barberID, seriesID, blockDate string) (bool, bool, error) {
	found, conflict := false, false

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var exists bool
		if err := q.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM time_block_series WHERE id = $1 AND barbershop_id = $2 AND barber_id = $3 AND deleted_at IS NULL)`,
			seriesID, barbershopID, barberID,
		).Scan(&exists); err != nil {
			return fmt.Errorf("check series exists: %w", err)
		}
		if !exists {
			return nil
		}
		found = true

		if _, err := q.Exec(ctx,
			`INSERT INTO time_block_series_date (barbershop_id, series_id, block_date) VALUES ($1, $2, $3::date)`,
			barbershopID, seriesID, blockDate,
		); err != nil {
			if isConstraintViolation(err, "23505", timeBlockSeriesDatePKConstraint) {
				conflict = true
				return errSeriesDateConflictInternal
			}
			return fmt.Errorf("insert series date: %w", err)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, errSeriesDateConflictInternal) {
			return found, conflict, nil
		}
		return false, false, fmt.Errorf("schedule/postgres: add series date: %w", err)
	}
	return found, conflict, nil
}

// RemoveSeriesDate implementa schedule.BlockRepository.RemoveSeriesDate:
// retiro físico (fila hija sin ciclo de vida propio).
func (r *Repository) RemoveSeriesDate(ctx context.Context, barbershopID, barberID, seriesID, blockDate string) (bool, error) {
	found := false

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		tag, err := q.Exec(ctx,
			`DELETE FROM time_block_series_date d
			  USING time_block_series s
			 WHERE d.series_id = s.id AND d.barbershop_id = s.barbershop_id
			   AND d.barbershop_id = $1 AND d.series_id = $2 AND d.block_date = $3::date
			   AND s.barber_id = $4`,
			barbershopID, seriesID, blockDate, barberID,
		)
		if err != nil {
			return fmt.Errorf("delete series date: %w", err)
		}
		found = tag.RowsAffected() > 0
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("schedule/postgres: remove series date: %w", err)
	}
	return found, nil
}

// AddSeriesException implementa schedule.BlockRepository.AddSeriesException.
func (r *Repository) AddSeriesException(ctx context.Context, barbershopID, barberID, seriesID, excludedDate string, reason *string) (bool, bool, error) {
	found, conflict := false, false

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var exists bool
		if err := q.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM time_block_series WHERE id = $1 AND barbershop_id = $2 AND barber_id = $3 AND deleted_at IS NULL)`,
			seriesID, barbershopID, barberID,
		).Scan(&exists); err != nil {
			return fmt.Errorf("check series exists: %w", err)
		}
		if !exists {
			return nil
		}
		found = true

		if _, err := q.Exec(ctx,
			`INSERT INTO time_block_series_exception (barbershop_id, series_id, excluded_date, reason) VALUES ($1, $2, $3::date, $4)`,
			barbershopID, seriesID, excludedDate, reason,
		); err != nil {
			if isConstraintViolation(err, "23505", timeBlockSeriesExceptionPKConstraint) {
				conflict = true
				return errSeriesExceptionConflictInternal
			}
			return fmt.Errorf("insert series exception: %w", err)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, errSeriesExceptionConflictInternal) {
			return found, conflict, nil
		}
		return false, false, fmt.Errorf("schedule/postgres: add series exception: %w", err)
	}
	return found, conflict, nil
}

// RemoveSeriesException implementa
// schedule.BlockRepository.RemoveSeriesException: retiro físico.
func (r *Repository) RemoveSeriesException(ctx context.Context, barbershopID, barberID, seriesID, excludedDate string) (bool, error) {
	found := false

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		tag, err := q.Exec(ctx,
			`DELETE FROM time_block_series_exception e
			  USING time_block_series s
			 WHERE e.series_id = s.id AND e.barbershop_id = s.barbershop_id
			   AND e.barbershop_id = $1 AND e.series_id = $2 AND e.excluded_date = $3::date
			   AND s.barber_id = $4`,
			barbershopID, seriesID, excludedDate, barberID,
		)
		if err != nil {
			return fmt.Errorf("delete series exception: %w", err)
		}
		found = tag.RowsAffected() > 0
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("schedule/postgres: remove series exception: %w", err)
	}
	return found, nil
}

// ListActiveSeriesForProjection implementa
// schedule.BlockRepository.ListActiveSeriesForProjection.
func (r *Repository) ListActiveSeriesForProjection(ctx context.Context, barbershopID, barberID, from, until string) ([]schedule.TimeBlockSeries, error) {
	var items []schedule.TimeBlockSeries

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		rows, err := q.Query(ctx,
			timeBlockSeriesSelectSQL+`
			  WHERE barbershop_id = $1 AND barber_id = $2 AND deleted_at IS NULL
			    AND effective_from <= $4::date
			    AND (effective_until IS NULL OR effective_until >= $3::date)
			  ORDER BY effective_from`,
			barbershopID, barberID, from, until,
		)
		if err != nil {
			return fmt.Errorf("list active series: query: %w", err)
		}
		defer rows.Close()
		items = make([]schedule.TimeBlockSeries, 0)
		for rows.Next() {
			sr, err := scanTimeBlockSeriesText(rows)
			if err != nil {
				return fmt.Errorf("list active series: scan: %w", err)
			}
			items = append(items, sr)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		return loadSeriesChildren(ctx, q, barbershopID, items)
	})
	if err != nil {
		return nil, fmt.Errorf("schedule/postgres: list active series for projection: %w", err)
	}
	return items, nil
}
