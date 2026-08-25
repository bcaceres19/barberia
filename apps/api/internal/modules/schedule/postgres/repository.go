// Package postgres es el adaptador de persistencia de schedule.Repository
// sobre database.DB (CA-002-06: el núcleo de schedule no importa este
// paquete ni pgx), mismo patrón que staff/postgres y catalog/postgres.
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

// createWorkingHourOperation identifica, dentro de una barbería, la
// operación de idempotencia del alta de un tramo (RN-IDE-01): una clave ya
// usada para OTRA operación nunca se confunde con esta.
const createWorkingHourOperation idempotency.Operation = "create_working_hour"

// createWorkingHourIdempotencyTTL es la vigencia de una reclamación de
// idempotencia sobre el alta de un tramo antes de tratarse como si nunca
// hubiera existido (CA-004-05), mismo valor que staff/catalog.
const createWorkingHourIdempotencyTTL = 24 * time.Hour

// workingHourWeekdayStartUniqueConstraint es el nombre del UNIQUE de
// 20260825160000_create_working_hour.sql que rechaza repetir exactamente
// el mismo día e inicio para un mismo barbero. Solo un unique_violation
// sobre ESTE constraint se traduce a un conflicto de negocio: cualquier
// otro unique_violation (que hoy no debería existir en working_hour) se
// propaga como error interno, mismo criterio que
// catalog/postgres.nameUniqueConstraint.
const workingHourWeekdayStartUniqueConstraint = "working_hour_shop_barber_weekday_start_uk"

// Repository implementa schedule.Repository sobre database.DB.
type Repository struct {
	db    *database.DB
	coord idempotency.Coordinator
}

// New construye el repositorio con el coordinador de idempotencia real
// (SQLCoordinator). El parámetro coord existe para pruebas, mismo criterio
// que staff/postgres.New.
func New(db *database.DB, coord idempotency.Coordinator) *Repository {
	return &Repository{db: db, coord: coord}
}

var _ schedule.Repository = (*Repository)(nil)

// errOverlapConflictInternal es un error interno (sentinela) que solo
// existe para hacer ROLLBACK de la transacción completa -incluida la
// reclamación de idempotencia que Begin ya había tomado, cuando aplica-
// desde dentro de una función anónima que solo puede comunicarse con su
// llamador mediante el valor de retorno error. Nunca se propaga fuera de
// este paquete: Create/Update lo interceptan y lo traducen a la bandera
// Conflict en el resultado, mismo patrón que
// catalog/postgres.errNameConflictInternal.
var errOverlapConflictInternal = errors.New("schedule/postgres: el tramo se solapa con otro existente")

// workingHourResponseWire es la forma JSON EXACTA que también usa
// httpapi.WorkingHourResponse: la respuesta que Complete persiste para una
// repetición exacta (CA-004-01) debe ser BYTE A BYTE idéntica a la que el
// handler ya construyó para la primera ejecución, mismo criterio que
// staff/postgres.barberResponseWire (ver
// TestCreate_StoredResponseBody_MatchesHTTPAPIWireShape en
// repository_test.go).
type workingHourResponseWire struct {
	ID              string    `json:"id"`
	ISOWeekday      int       `json:"isoWeekday"`
	StartsTime      string    `json:"startsTime"`
	DurationMinutes int       `json:"durationMinutes"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func marshalWorkingHourResponse(wh schedule.WorkingHour) ([]byte, error) {
	return json.Marshal(workingHourResponseWire{
		ID:              wh.ID,
		ISOWeekday:      wh.ISOWeekday,
		StartsTime:      wh.StartsTime,
		DurationMinutes: wh.DurationMinutes,
		CreatedAt:       wh.CreatedAt,
		UpdatedAt:       wh.UpdatedAt,
	})
}

// isWeekdayStartConflict informa si err es el unique_violation (23505) de
// workingHourWeekdayStartUniqueConstraint.
func isWeekdayStartConflict(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23505" && pgErr.ConstraintName == workingHourWeekdayStartUniqueConstraint
}

// lockBarberForScheduleWrite bloquea la fila de barber (SELECT ... FOR
// UPDATE) antes de verificar solape e insertar/editar un tramo: sin este
// bloqueo, dos transacciones concurrentes que parten de un conjunto de
// tramos existentes VACÍO para el mismo barbero y día podrían leer "cero
// tramos" ambas, no ver nada que bloquear con un SELECT ... FOR UPDATE
// sobre working_hour, e insertar dos tramos que se solapan entre sí (el
// clásico "phantom read" de una transacción a nivel READ COMMITTED). Mismo
// criterio que DEC-068 bloqueando la fila de service para resistir la
// carrera de dos desasignaciones concurrentes. No falla si el barbero no
// existe (cero filas): el llamador ya lo verificó mediante BarberPort antes
// de esta transacción; si de todas formas no existe, la verificación de
// solape no encuentra nada y el INSERT/UPDATE final falla por su cuenta
// (FK o cero filas afectadas).
func lockBarberForScheduleWrite(ctx context.Context, q database.Queries, barbershopID, barberID string) error {
	rows, err := q.Query(ctx, `SELECT 1 FROM barber WHERE barbershop_id = $1 AND id = $2 FOR UPDATE`, barbershopID, barberID)
	if err != nil {
		return fmt.Errorf("lock barber: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
	}
	return rows.Err()
}

// overlapsExisting bloquea (SELECT ... FOR UPDATE) y lee los tramos
// existentes del mismo (barbershopID, barberID, isoWeekday), excluyendo
// excludeID (vacío en Create, el propio tramo en Update), y verifica si
// alguno se solapa con [startsTime, startsTime+durationMinutes) mediante
// schedule.IntervalsOverlap.
func overlapsExisting(
	ctx context.Context, q database.Queries,
	barbershopID, barberID string, isoWeekday int, excludeID string,
	startsTime string, durationMinutes int,
) (bool, error) {
	rows, err := q.Query(ctx,
		`SELECT to_char(starts_time, 'HH24:MI'), duration_minutes
		   FROM working_hour
		  WHERE barbershop_id = $1 AND barber_id = $2 AND iso_weekday = $3 AND id <> $4
		    FOR UPDATE`,
		barbershopID, barberID, isoWeekday, excludeID,
	)
	if err != nil {
		return false, fmt.Errorf("lock existing working hours: %w", err)
	}
	defer rows.Close()

	newStart := minutesOfDayExported(startsTime)
	for rows.Next() {
		var existingStart string
		var existingDuration int
		if err := rows.Scan(&existingStart, &existingDuration); err != nil {
			return false, fmt.Errorf("scan existing working hour: %w", err)
		}
		if schedule.IntervalsOverlap(newStart, durationMinutes, minutesOfDayExported(existingStart), existingDuration) {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("iterate existing working hours: %w", err)
	}
	return false, nil
}

// List implementa schedule.Repository.List: orden estable (iso_weekday,
// starts_time, id) dentro del tenant y barbero vigentes, cursor opaco
// decodificado por el núcleo (schedule.Cursor), paginación "pedir uno de
// más" para saber si hay página siguiente sin una segunda consulta COUNT
// (mismo patrón que staff/postgres.Repository.List).
func (r *Repository) List(ctx context.Context, barbershopID, barberID string, cursor *schedule.Cursor, limit int) (schedule.ListResult, error) {
	var result schedule.ListResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var (
			rows pgx.Rows
			err  error
		)
		fetchLimit := limit + 1

		if cursor == nil {
			rows, err = q.Query(ctx,
				`SELECT id, iso_weekday, to_char(starts_time, 'HH24:MI'), duration_minutes, created_at, updated_at
				   FROM working_hour
				  WHERE barbershop_id = $1 AND barber_id = $2
				  ORDER BY iso_weekday, starts_time, id
				  LIMIT $3`,
				barbershopID, barberID, fetchLimit,
			)
		} else {
			rows, err = q.Query(ctx,
				`SELECT id, iso_weekday, to_char(starts_time, 'HH24:MI'), duration_minutes, created_at, updated_at
				   FROM working_hour
				  WHERE barbershop_id = $1 AND barber_id = $2
				    AND (iso_weekday, starts_time, id) > ($3, $4::time, $5)
				  ORDER BY iso_weekday, starts_time, id
				  LIMIT $6`,
				barbershopID, barberID, cursor.ISOWeekday, cursor.StartsTime, cursor.ID, fetchLimit,
			)
		}
		if err != nil {
			return fmt.Errorf("list working hours: query: %w", err)
		}
		defer rows.Close()

		items := make([]schedule.WorkingHour, 0, fetchLimit)
		for rows.Next() {
			var wh schedule.WorkingHour
			if err := rows.Scan(&wh.ID, &wh.ISOWeekday, &wh.StartsTime, &wh.DurationMinutes, &wh.CreatedAt, &wh.UpdatedAt); err != nil {
				return fmt.Errorf("list working hours: scan: %w", err)
			}
			items = append(items, wh)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("list working hours: rows: %w", err)
		}

		hasMore := len(items) > limit
		if hasMore {
			items = items[:limit]
		}

		result.Items = items
		if hasMore {
			last := items[len(items)-1]
			result.NextCursor = schedule.EncodeCursor(schedule.Cursor{ISOWeekday: last.ISOWeekday, StartsTime: last.StartsTime, ID: last.ID})
		}
		return nil
	})
	if err != nil {
		return schedule.ListResult{}, fmt.Errorf("schedule/postgres: list working hours: %w", err)
	}
	return result, nil
}

// Get implementa schedule.Repository.Get. RLS (working_hour_select_tenant_policy)
// ya restringe la fila visible a barbershop_id = current_setting(...); el
// filtro explícito WHERE barbershop_id = $2 AND barber_id = $3 es defensa
// en profundidad, mismo patrón que staff/postgres.Repository.Get.
func (r *Repository) Get(ctx context.Context, barbershopID, barberID, workingHourID string) (schedule.WorkingHour, bool, error) {
	var wh schedule.WorkingHour
	found := false

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		wh.ID = workingHourID
		err := q.QueryRow(ctx,
			`SELECT iso_weekday, to_char(starts_time, 'HH24:MI'), duration_minutes, created_at, updated_at
			   FROM working_hour
			  WHERE id = $1 AND barbershop_id = $2 AND barber_id = $3`,
			workingHourID, barbershopID, barberID,
		).Scan(&wh.ISOWeekday, &wh.StartsTime, &wh.DurationMinutes, &wh.CreatedAt, &wh.UpdatedAt)
		switch {
		case err == nil:
			found = true
		case errors.Is(err, pgx.ErrNoRows):
			// found queda false: cubre "no existe", "es de otro barbero" y "es
			// de otra barbería" (CA-040-05), sin distinguir la causa.
		default:
			return fmt.Errorf("select working hour: %w", err)
		}
		return nil
	})
	if err != nil {
		return schedule.WorkingHour{}, false, fmt.Errorf("schedule/postgres: get working hour: %w", err)
	}
	if !found {
		return schedule.WorkingHour{}, false, nil
	}
	return wh, true, nil
}

// Create implementa schedule.Repository.Create: Begin, el bloqueo del
// barbero, la verificación de solape, el INSERT y Complete ocurren dentro
// de la MISMA InTenantTx (apps/api/README.md, "Patrón obligatorio:
// idempotencia reutilizable"). Un solape (o un unique_violation defensivo
// de workingHourWeekdayStartUniqueConstraint) hace ROLLBACK de toda la
// transacción -incluida la reclamación de idempotencia que Begin ya había
// tomado- y se traduce a CreateResult.Conflict, dejando la clave libre
// para un reintento legítimo con un intervalo distinto (mismo criterio que
// CA-004-06/catalog.CreateResult.NameTaken).
func (r *Repository) Create(
	ctx context.Context,
	barbershopID, barberID string,
	input schedule.CreateInput,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (schedule.CreateResult, error) {
	var result schedule.CreateResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		decision, err := r.coord.Begin(ctx, q, database.BarbershopID(barbershopID), key, createWorkingHourOperation, fingerprint, createWorkingHourIdempotencyTTL)
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

		if err := lockBarberForScheduleWrite(ctx, q, barbershopID, barberID); err != nil {
			return err
		}

		overlaps, err := overlapsExisting(ctx, q, barbershopID, barberID, input.ISOWeekday, "00000000-0000-0000-0000-000000000000", input.StartsTime, input.DurationMinutes)
		if err != nil {
			return err
		}
		if overlaps {
			result.Conflict = true
			return errOverlapConflictInternal
		}

		var wh schedule.WorkingHour
		row := q.QueryRow(ctx,
			`INSERT INTO working_hour (barbershop_id, barber_id, iso_weekday, starts_time, duration_minutes)
			      VALUES ($1, $2, $3, $4::time, $5)
			   RETURNING id, iso_weekday, to_char(starts_time, 'HH24:MI'), duration_minutes, created_at, updated_at`,
			barbershopID, barberID, input.ISOWeekday, input.StartsTime, input.DurationMinutes,
		)
		if err := row.Scan(&wh.ID, &wh.ISOWeekday, &wh.StartsTime, &wh.DurationMinutes, &wh.CreatedAt, &wh.UpdatedAt); err != nil {
			if isWeekdayStartConflict(err) {
				result.Conflict = true
				return errOverlapConflictInternal
			}
			return fmt.Errorf("insert working hour: %w", err)
		}

		body, err := marshalWorkingHourResponse(wh)
		if err != nil {
			return fmt.Errorf("marshal created working hour: %w", err)
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

		result.WorkingHour = wh
		result.Response = stored
		return nil
	})
	if err != nil {
		if errors.Is(err, errOverlapConflictInternal) {
			return result, nil
		}
		return schedule.CreateResult{}, fmt.Errorf("schedule/postgres: create working hour: %w", err)
	}
	return result, nil
}

// Update implementa schedule.Repository.Update: bloquea el barbero,
// verifica solape contra los demás tramos del mismo (barbershopID,
// barberID, input.ISOWeekday) EXCLUYENDO workingHourID, y
// `UPDATE ... RETURNING` filtrando por id, barbershopID y barberID. Cero
// filas afectadas cubre "no existe"/"de otro barbero"/"de otra barbería"
// (CA-040-05).
func (r *Repository) Update(ctx context.Context, barbershopID, barberID, workingHourID string, input schedule.UpdateInput) (schedule.UpdateResult, error) {
	var result schedule.UpdateResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		if err := lockBarberForScheduleWrite(ctx, q, barbershopID, barberID); err != nil {
			return err
		}

		overlaps, err := overlapsExisting(ctx, q, barbershopID, barberID, input.ISOWeekday, workingHourID, input.StartsTime, input.DurationMinutes)
		if err != nil {
			return err
		}
		if overlaps {
			result.Conflict = true
			return nil
		}

		var wh schedule.WorkingHour
		wh.ID = workingHourID
		row := q.QueryRow(ctx,
			`UPDATE working_hour
			    SET iso_weekday = $4, starts_time = $5::time, duration_minutes = $6
			  WHERE id = $1 AND barbershop_id = $2 AND barber_id = $3
			RETURNING iso_weekday, to_char(starts_time, 'HH24:MI'), duration_minutes, created_at, updated_at`,
			workingHourID, barbershopID, barberID, input.ISOWeekday, input.StartsTime, input.DurationMinutes,
		)
		err = row.Scan(&wh.ISOWeekday, &wh.StartsTime, &wh.DurationMinutes, &wh.CreatedAt, &wh.UpdatedAt)
		switch {
		case err == nil:
			result.Found = true
			result.WorkingHour = wh
		case errors.Is(err, pgx.ErrNoRows):
			// result.Found queda false.
		case isWeekdayStartConflict(err):
			// Defensa en profundidad: overlapsExisting ya debería haber
			// capturado este caso (un mismo inicio exacto es un solape de
			// duración >= 1), pero un unique_violation aquí se traduce igual
			// a Conflict en vez de propagarse como error interno.
			result.Conflict = true
		default:
			return fmt.Errorf("update working hour: %w", err)
		}
		return nil
	})
	if err != nil {
		return schedule.UpdateResult{}, fmt.Errorf("schedule/postgres: update working hour: %w", err)
	}
	return result, nil
}

// Delete implementa schedule.Repository.Delete: retiro físico (working_hour
// no tiene eliminación lógica), filtrando por id, barbershopID y barberID.
func (r *Repository) Delete(ctx context.Context, barbershopID, barberID, workingHourID string) (bool, error) {
	found := false

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		tag, err := q.Exec(ctx,
			`DELETE FROM working_hour WHERE id = $1 AND barbershop_id = $2 AND barber_id = $3`,
			workingHourID, barbershopID, barberID,
		)
		if err != nil {
			return fmt.Errorf("delete working hour: %w", err)
		}
		found = tag.RowsAffected() > 0
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("schedule/postgres: delete working hour: %w", err)
	}
	return found, nil
}

// minutesOfDayExported convierte una hora "HH:MM" (ya canónica, producida
// por to_char(starts_time, 'HH24:MI') o ya validada por
// schedule.ValidateStartsTime) en minutos desde medianoche. Duplicada aquí
// porque schedule.minutesOfDay es un detalle interno no exportado del
// núcleo: este adaptador solo necesita la misma aritmética simple para
// alimentar schedule.IntervalsOverlap, no acceso a ningún estado del
// núcleo.
func minutesOfDayExported(hhmm string) int {
	hours := int(hhmm[0]-'0')*10 + int(hhmm[1]-'0')
	minutes := int(hhmm[3]-'0')*10 + int(hhmm[4]-'0')
	return hours*60 + minutes
}
