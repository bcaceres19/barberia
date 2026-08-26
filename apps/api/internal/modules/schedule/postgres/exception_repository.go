package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"system-barbershop/internal/modules/schedule"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

// createExceptionOperation identifica, dentro de una barbería, la
// operación de idempotencia del alta de una excepción de jornada
// (RN-IDE-01), mismo criterio que createWorkingHourOperation de HU-040.
const createExceptionOperation idempotency.Operation = "create_schedule_exception"

const createExceptionIdempotencyTTL = 24 * time.Hour

// exceptionDateUniqueConstraint y exceptionSegmentOverlapConstraint son los
// nombres de las restricciones de 20260826090000_create_working_hour_override.sql
// que un unique_violation/exclusion_violation puede reportar. Solo estos
// dos se traducen a un conflicto de negocio: cualquier otra violación se
// propaga como error interno.
const (
	exceptionDateUniqueConstraint     = "working_hour_override_shop_barber_date_uk"
	exceptionSegmentOverlapConstraint = "working_hour_override_segment_no_overlap_excl"
)

// errExceptionConflictInternal es el mismo patrón de sentinela que
// errOverlapConflictInternal (schedule/postgres/repository.go, HU-040):
// solo existe para forzar ROLLBACK completo -incluida la reclamación de
// idempotencia, cuando aplica- desde dentro de una función anónima.
var errExceptionConflictInternal = errors.New("schedule/postgres: la excepción entra en conflicto con el estado existente")

// exceptionResponseWire es la forma JSON EXACTA que también usa
// httpapi.ScheduleExceptionResponse: la respuesta que Complete persiste
// para una repetición exacta (CA-004-01) debe ser BYTE A BYTE idéntica a
// la que el handler ya construyó, mismo criterio que
// workingHourResponseWire.
type exceptionResponseWire struct {
	ID            string                 `json:"id"`
	EffectiveDate string                 `json:"effectiveDate"`
	IsClosed      bool                   `json:"isClosed"`
	Reason        *string                `json:"reason"`
	Segments      []exceptionSegmentWire `json:"segments"`
	CreatedAt     time.Time              `json:"createdAt"`
	UpdatedAt     time.Time              `json:"updatedAt"`
}

type exceptionSegmentWire struct {
	ID              string `json:"id"`
	StartsTime      string `json:"startsTime"`
	DurationMinutes int    `json:"durationMinutes"`
}

func marshalExceptionResponse(e schedule.ScheduleException) ([]byte, error) {
	segments := make([]exceptionSegmentWire, 0, len(e.Segments))
	for _, seg := range e.Segments {
		segments = append(segments, exceptionSegmentWire{ID: seg.ID, StartsTime: seg.StartsTime, DurationMinutes: seg.DurationMinutes})
	}
	return json.Marshal(exceptionResponseWire{
		ID:            e.ID,
		EffectiveDate: e.EffectiveDate,
		IsClosed:      e.IsClosed,
		Reason:        e.Reason,
		Segments:      segments,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	})
}

func isConstraintViolation(err error, code, constraint string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == code && pgErr.ConstraintName == constraint
}

func isExceptionDateConflict(err error) bool {
	return isConstraintViolation(err, "23505", exceptionDateUniqueConstraint)
}

func isExceptionSegmentOverlap(err error) bool {
	return isConstraintViolation(err, "23P01", exceptionSegmentOverlapConstraint)
}

// loadExceptionSegments lee, ordenados por hora de inicio, los tramos de
// UNA excepción.
func loadExceptionSegments(ctx context.Context, q database.Queries, barbershopID, overrideID string) ([]schedule.ExceptionSegment, error) {
	rows, err := q.Query(ctx,
		`SELECT id, to_char(starts_time, 'HH24:MI'), duration_minutes
		   FROM working_hour_override_segment
		  WHERE barbershop_id = $1 AND override_id = $2
		  ORDER BY starts_time`,
		barbershopID, overrideID,
	)
	if err != nil {
		return nil, fmt.Errorf("select exception segments: %w", err)
	}
	defer rows.Close()

	segments := make([]schedule.ExceptionSegment, 0, 4)
	for rows.Next() {
		var seg schedule.ExceptionSegment
		if err := rows.Scan(&seg.ID, &seg.StartsTime, &seg.DurationMinutes); err != nil {
			return nil, fmt.Errorf("scan exception segment: %w", err)
		}
		segments = append(segments, seg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate exception segments: %w", err)
	}
	return segments, nil
}

// loadExceptionSegmentsBatch lee los tramos de VARIAS excepciones a la vez
// (ListExceptions): evita N+1 consultas, una por fila de la página.
func loadExceptionSegmentsBatch(ctx context.Context, q database.Queries, barbershopID string, overrideIDs []string) (map[string][]schedule.ExceptionSegment, error) {
	byOverride := make(map[string][]schedule.ExceptionSegment, len(overrideIDs))
	if len(overrideIDs) == 0 {
		return byOverride, nil
	}

	rows, err := q.Query(ctx,
		`SELECT id, override_id, to_char(starts_time, 'HH24:MI'), duration_minutes
		   FROM working_hour_override_segment
		  WHERE barbershop_id = $1 AND override_id = ANY($2)
		  ORDER BY override_id, starts_time`,
		barbershopID, overrideIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("select exception segments batch: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var overrideID string
		var seg schedule.ExceptionSegment
		if err := rows.Scan(&seg.ID, &overrideID, &seg.StartsTime, &seg.DurationMinutes); err != nil {
			return nil, fmt.Errorf("scan exception segment batch: %w", err)
		}
		byOverride[overrideID] = append(byOverride[overrideID], seg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate exception segments batch: %w", err)
	}
	return byOverride, nil
}

func insertExceptionSegments(ctx context.Context, q database.Queries, barbershopID, overrideID string, segments []schedule.CreateExceptionSegmentInput) ([]schedule.ExceptionSegment, error) {
	inserted := make([]schedule.ExceptionSegment, 0, len(segments))
	for _, seg := range segments {
		var id string
		if err := q.QueryRow(ctx,
			`INSERT INTO working_hour_override_segment (barbershop_id, override_id, starts_time, duration_minutes)
			      VALUES ($1, $2, $3::time, $4)
			   RETURNING id`,
			barbershopID, overrideID, seg.StartsTime, seg.DurationMinutes,
		).Scan(&id); err != nil {
			return nil, err
		}
		inserted = append(inserted, schedule.ExceptionSegment{ID: id, StartsTime: seg.StartsTime, DurationMinutes: seg.DurationMinutes})
	}
	return inserted, nil
}

// GetHolidayCalendarEnabled implementa schedule.Repository.GetHolidayCalendarEnabled.
func (r *Repository) GetHolidayCalendarEnabled(ctx context.Context, barbershopID, barberID string) (bool, bool, error) {
	var enabled bool
	found := false

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		err := q.QueryRow(ctx,
			`SELECT holiday_calendar_enabled FROM barber WHERE id = $1 AND barbershop_id = $2`,
			barberID, barbershopID,
		).Scan(&enabled)
		switch {
		case err == nil:
			found = true
		case errors.Is(err, pgx.ErrNoRows):
			// found queda false.
		default:
			return fmt.Errorf("select holiday_calendar_enabled: %w", err)
		}
		return nil
	})
	if err != nil {
		return false, false, fmt.Errorf("schedule/postgres: get holiday calendar enabled: %w", err)
	}
	return enabled, found, nil
}

// SetHolidayCalendarEnabled implementa schedule.Repository.SetHolidayCalendarEnabled.
func (r *Repository) SetHolidayCalendarEnabled(ctx context.Context, barbershopID, barberID string, enabled bool) (schedule.HolidayCalendarResult, error) {
	var result schedule.HolidayCalendarResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		err := q.QueryRow(ctx,
			`UPDATE barber SET holiday_calendar_enabled = $3
			  WHERE id = $1 AND barbershop_id = $2
			RETURNING holiday_calendar_enabled`,
			barberID, barbershopID, enabled,
		).Scan(&result.Enabled)
		switch {
		case err == nil:
			result.Found = true
		case errors.Is(err, pgx.ErrNoRows):
			// result.Found queda false.
		default:
			return fmt.Errorf("update holiday_calendar_enabled: %w", err)
		}
		return nil
	})
	if err != nil {
		return schedule.HolidayCalendarResult{}, fmt.Errorf("schedule/postgres: set holiday calendar enabled: %w", err)
	}
	return result, nil
}

// ListExceptions implementa schedule.Repository.ListExceptions: orden
// estable (effective_date, id), paginación "pedir uno de más" (mismo
// patrón que List de HU-040), tramos cargados en un segundo lote sin N+1.
func (r *Repository) ListExceptions(ctx context.Context, barbershopID, barberID string, cursor *schedule.ExceptionCursor, limit int) (schedule.ExceptionListResult, error) {
	var result schedule.ExceptionListResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var (
			rows pgx.Rows
			err  error
		)
		fetchLimit := limit + 1

		if cursor == nil {
			rows, err = q.Query(ctx,
				`SELECT id, to_char(effective_date, 'YYYY-MM-DD'), is_closed, reason, created_at, updated_at
				   FROM working_hour_override
				  WHERE barbershop_id = $1 AND barber_id = $2
				  ORDER BY effective_date, id
				  LIMIT $3`,
				barbershopID, barberID, fetchLimit,
			)
		} else {
			rows, err = q.Query(ctx,
				`SELECT id, to_char(effective_date, 'YYYY-MM-DD'), is_closed, reason, created_at, updated_at
				   FROM working_hour_override
				  WHERE barbershop_id = $1 AND barber_id = $2
				    AND (effective_date, id) > ($3::date, $4)
				  ORDER BY effective_date, id
				  LIMIT $5`,
				barbershopID, barberID, cursor.EffectiveDate, cursor.ID, fetchLimit,
			)
		}
		if err != nil {
			return fmt.Errorf("list exceptions: query: %w", err)
		}

		items := make([]schedule.ScheduleException, 0, fetchLimit)
		ids := make([]string, 0, fetchLimit)
		for rows.Next() {
			var e schedule.ScheduleException
			if err := rows.Scan(&e.ID, &e.EffectiveDate, &e.IsClosed, &e.Reason, &e.CreatedAt, &e.UpdatedAt); err != nil {
				rows.Close()
				return fmt.Errorf("list exceptions: scan: %w", err)
			}
			items = append(items, e)
			ids = append(ids, e.ID)
		}
		rowsErr := rows.Err()
		rows.Close()
		if rowsErr != nil {
			return fmt.Errorf("list exceptions: rows: %w", rowsErr)
		}

		hasMore := len(items) > limit
		if hasMore {
			items = items[:limit]
			ids = ids[:limit]
		}

		segmentsByOverride, err := loadExceptionSegmentsBatch(ctx, q, barbershopID, ids)
		if err != nil {
			return fmt.Errorf("list exceptions: %w", err)
		}
		for i := range items {
			items[i].Segments = segmentsByOverride[items[i].ID]
		}

		result.Items = items
		if hasMore {
			last := items[len(items)-1]
			result.NextCursor = schedule.EncodeExceptionCursor(schedule.ExceptionCursor{EffectiveDate: last.EffectiveDate, ID: last.ID})
		}
		return nil
	})
	if err != nil {
		return schedule.ExceptionListResult{}, fmt.Errorf("schedule/postgres: list exceptions: %w", err)
	}
	return result, nil
}

// GetException implementa schedule.Repository.GetException.
func (r *Repository) GetException(ctx context.Context, barbershopID, barberID, exceptionID string) (schedule.ScheduleException, bool, error) {
	return r.getExceptionBy(ctx, barbershopID,
		`SELECT id, to_char(effective_date, 'YYYY-MM-DD'), is_closed, reason, created_at, updated_at
		   FROM working_hour_override
		  WHERE id = $1 AND barbershop_id = $2 AND barber_id = $3`,
		exceptionID, barberID,
	)
}

// GetExceptionByDate implementa schedule.Repository.GetExceptionByDate.
func (r *Repository) GetExceptionByDate(ctx context.Context, barbershopID, barberID, effectiveDate string) (schedule.ScheduleException, bool, error) {
	return r.getExceptionBy(ctx, barbershopID,
		`SELECT id, to_char(effective_date, 'YYYY-MM-DD'), is_closed, reason, created_at, updated_at
		   FROM working_hour_override
		  WHERE effective_date = $1::date AND barbershop_id = $2 AND barber_id = $3`,
		effectiveDate, barberID,
	)
}

func (r *Repository) getExceptionBy(ctx context.Context, barbershopID, query, arg1, barberID string) (schedule.ScheduleException, bool, error) {
	var e schedule.ScheduleException
	found := false

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		err := q.QueryRow(ctx, query, arg1, barbershopID, barberID).
			Scan(&e.ID, &e.EffectiveDate, &e.IsClosed, &e.Reason, &e.CreatedAt, &e.UpdatedAt)
		switch {
		case err == nil:
			found = true
		case errors.Is(err, pgx.ErrNoRows):
			return nil
		default:
			return fmt.Errorf("select exception: %w", err)
		}

		segments, err := loadExceptionSegments(ctx, q, barbershopID, e.ID)
		if err != nil {
			return err
		}
		e.Segments = segments
		return nil
	})
	if err != nil {
		return schedule.ScheduleException{}, false, fmt.Errorf("schedule/postgres: get exception: %w", err)
	}
	if !found {
		return schedule.ScheduleException{}, false, nil
	}
	return e, true, nil
}

// CreateException implementa schedule.Repository.CreateException: Begin,
// el INSERT de la cabecera y de cada tramo, y Complete ocurren dentro de
// la MISMA InTenantTx. Un unique_violation de fecha duplicada o un
// exclusion_violation de solape hacen ROLLBACK de toda la transacción
// -incluida la reclamación de idempotencia- y se traducen a
// CreateExceptionResult.Conflict (mismo criterio que Create de HU-040).
func (r *Repository) CreateException(
	ctx context.Context,
	barbershopID, barberID string,
	input schedule.CreateExceptionInput,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (schedule.CreateExceptionResult, error) {
	var result schedule.CreateExceptionResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		decision, err := r.coord.Begin(ctx, q, database.BarbershopID(barbershopID), key, createExceptionOperation, fingerprint, createExceptionIdempotencyTTL)
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

		var e schedule.ScheduleException
		row := q.QueryRow(ctx,
			`INSERT INTO working_hour_override (barbershop_id, barber_id, effective_date, is_closed, reason)
			      VALUES ($1, $2, $3::date, $4, $5)
			   RETURNING id, to_char(effective_date, 'YYYY-MM-DD'), is_closed, reason, created_at, updated_at`,
			barbershopID, barberID, input.EffectiveDate, input.IsClosed, input.Reason,
		)
		if err := row.Scan(&e.ID, &e.EffectiveDate, &e.IsClosed, &e.Reason, &e.CreatedAt, &e.UpdatedAt); err != nil {
			if isExceptionDateConflict(err) {
				result.Conflict = true
				return errExceptionConflictInternal
			}
			return fmt.Errorf("insert exception: %w", err)
		}

		segments, err := insertExceptionSegments(ctx, q, barbershopID, e.ID, input.Segments)
		if err != nil {
			if isExceptionSegmentOverlap(err) {
				result.Conflict = true
				return errExceptionConflictInternal
			}
			return fmt.Errorf("insert exception segments: %w", err)
		}
		e.Segments = segments

		body, err := marshalExceptionResponse(e)
		if err != nil {
			return fmt.Errorf("marshal created exception: %w", err)
		}
		stored := idempotency.StoredResponse{
			Status:      201,
			ContentType: "application/json",
			Body:        string(body),
		}

		ok, err := r.coord.Complete(ctx, q, database.BarbershopID(barbershopID), key, stored)
		if err != nil {
			return fmt.Errorf("idempotency complete: %w", err)
		}
		if !ok {
			return fmt.Errorf("idempotency complete: la reclamación ya no estaba in_progress")
		}

		result.Exception = e
		result.Response = stored
		return nil
	})
	if err != nil {
		if errors.Is(err, errExceptionConflictInternal) {
			return result, nil
		}
		return schedule.CreateExceptionResult{}, fmt.Errorf("schedule/postgres: create exception: %w", err)
	}
	return result, nil
}

// UpdateException implementa schedule.Repository.UpdateException:
// reemplaza la cabecera con `UPDATE ... RETURNING`, filtrando por id,
// barbershopID y barberID, borra los tramos existentes y crea los nuevos,
// todo dentro de UNA sola transacción. Cero filas afectadas en el UPDATE
// de la cabecera cubre "no existe"/"de otro barbero"/"de otra barbería"
// (CA-041-06). Un conflicto detectado DESPUÉS de que el UPDATE de la
// cabecera o el DELETE de los tramos previos ya se ejecutaron fuerza
// ROLLBACK de la transacción COMPLETA mediante el mismo sentinela que
// CreateException (errExceptionConflictInternal): nunca deja una
// cabecera ya modificada con sus tramos a medio reemplazar.
func (r *Repository) UpdateException(ctx context.Context, barbershopID, barberID, exceptionID string, input schedule.UpdateExceptionInput) (schedule.UpdateExceptionResult, error) {
	var result schedule.UpdateExceptionResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var e schedule.ScheduleException
		e.ID = exceptionID
		row := q.QueryRow(ctx,
			`UPDATE working_hour_override
			    SET effective_date = $4::date, is_closed = $5, reason = $6
			  WHERE id = $1 AND barbershop_id = $2 AND barber_id = $3
			RETURNING to_char(effective_date, 'YYYY-MM-DD'), is_closed, reason, created_at, updated_at`,
			exceptionID, barbershopID, barberID, input.EffectiveDate, input.IsClosed, input.Reason,
		)
		err := row.Scan(&e.EffectiveDate, &e.IsClosed, &e.Reason, &e.CreatedAt, &e.UpdatedAt)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil // result.Found queda false; nada se escribió todavía.
		case err != nil:
			if isExceptionDateConflict(err) {
				result.Conflict = true
				return errExceptionConflictInternal
			}
			return fmt.Errorf("update exception: %w", err)
		}
		result.Found = true

		if _, err := q.Exec(ctx,
			`DELETE FROM working_hour_override_segment WHERE barbershop_id = $1 AND override_id = $2`,
			barbershopID, exceptionID,
		); err != nil {
			return fmt.Errorf("delete existing exception segments: %w", err)
		}

		segments, err := insertExceptionSegments(ctx, q, barbershopID, exceptionID, input.Segments)
		if err != nil {
			if isExceptionSegmentOverlap(err) {
				result.Conflict = true
				return errExceptionConflictInternal
			}
			return fmt.Errorf("insert updated exception segments: %w", err)
		}
		e.Segments = segments
		result.Exception = e
		return nil
	})
	if err != nil {
		if errors.Is(err, errExceptionConflictInternal) {
			// El ROLLBACK ya deshizo el UPDATE de la cabecera y el DELETE
			// de los tramos previos: Found se limpia para reflejar que no
			// se persistió ningún cambio, solo Conflict=true importa aquí.
			return schedule.UpdateExceptionResult{Conflict: true}, nil
		}
		return schedule.UpdateExceptionResult{}, fmt.Errorf("schedule/postgres: update exception: %w", err)
	}
	return result, nil
}

// DeleteException implementa schedule.Repository.DeleteException: borra
// primero los tramos de la cabecera (si existe y pertenece a este barbero
// y barbería) y luego la cabecera misma, dentro de UNA sola transacción.
func (r *Repository) DeleteException(ctx context.Context, barbershopID, barberID, exceptionID string) (bool, error) {
	found := false

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		if _, err := q.Exec(ctx,
			`DELETE FROM working_hour_override_segment
			  WHERE barbershop_id = $1
			    AND override_id IN (
			      SELECT id FROM working_hour_override
			       WHERE barbershop_id = $1 AND barber_id = $2 AND id = $3
			    )`,
			barbershopID, barberID, exceptionID,
		); err != nil {
			return fmt.Errorf("delete exception segments: %w", err)
		}

		tag, err := q.Exec(ctx,
			`DELETE FROM working_hour_override WHERE id = $1 AND barbershop_id = $2 AND barber_id = $3`,
			exceptionID, barbershopID, barberID,
		)
		if err != nil {
			return fmt.Errorf("delete exception: %w", err)
		}
		found = tag.RowsAffected() > 0
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("schedule/postgres: delete exception: %w", err)
	}
	return found, nil
}

// ListWorkingHoursForWeekday implementa
// schedule.Repository.ListWorkingHoursForWeekday: sin paginar (lo usa
// exclusivamente ResolveEffectiveDay, nunca un handler HTTP).
func (r *Repository) ListWorkingHoursForWeekday(ctx context.Context, barbershopID, barberID string, isoWeekday int) ([]schedule.WorkingHour, error) {
	var items []schedule.WorkingHour

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		rows, err := q.Query(ctx,
			`SELECT id, iso_weekday, to_char(starts_time, 'HH24:MI'), duration_minutes, created_at, updated_at
			   FROM working_hour
			  WHERE barbershop_id = $1 AND barber_id = $2 AND iso_weekday = $3
			  ORDER BY starts_time`,
			barbershopID, barberID, isoWeekday,
		)
		if err != nil {
			return fmt.Errorf("list working hours for weekday: query: %w", err)
		}
		defer rows.Close()

		items = make([]schedule.WorkingHour, 0, 4)
		for rows.Next() {
			var wh schedule.WorkingHour
			if err := rows.Scan(&wh.ID, &wh.ISOWeekday, &wh.StartsTime, &wh.DurationMinutes, &wh.CreatedAt, &wh.UpdatedAt); err != nil {
				return fmt.Errorf("list working hours for weekday: scan: %w", err)
			}
			items = append(items, wh)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("schedule/postgres: list working hours for weekday: %w", err)
	}
	return items, nil
}
