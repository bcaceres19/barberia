package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"system-barbershop/internal/modules/catalog"
	"system-barbershop/internal/platform/database"
)

// AssignmentRepository implementa catalog.AssignmentRepository sobre
// database.DB. Separado de Repository (repository.go, dueño de `service`)
// porque opera sobre una tabla distinta (`barber_service`) con su propia
// disciplina transaccional (DEC-068); comparte el mismo *database.DB, sin
// necesitar el coordinador de idempotencia (asignar/desasignar usan
// semántica HTTP naturalmente repetible -PUT/DELETE-, no el protocolo de
// Idempotency-Key de RN-IDE-01).
type AssignmentRepository struct {
	db *database.DB
}

// NewAssignmentRepository construye el repositorio de asignaciones.
func NewAssignmentRepository(db *database.DB) *AssignmentRepository {
	return &AssignmentRepository{db: db}
}

var _ catalog.AssignmentRepository = (*AssignmentRepository)(nil)

// Exists implementa catalog.AssignmentRepository.Exists (HU-061, DEC-072).
func (r *AssignmentRepository) Exists(ctx context.Context, barbershopID, barberID, serviceID string) (bool, error) {
	var found bool
	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		row := q.QueryRow(ctx,
			`SELECT EXISTS (
			   SELECT 1 FROM barber_service
			    WHERE barbershop_id = $1 AND barber_id = $2 AND service_id = $3
			 )`,
			barbershopID, barberID, serviceID,
		)
		return row.Scan(&found)
	})
	if err != nil {
		return false, fmt.Errorf("catalog/postgres: verificar asignación: %w", err)
	}
	return found, nil
}

// List implementa catalog.AssignmentRepository.List: orden estable
// (created_at, service_id) dentro del tenant vigente y del barbero
// solicitado, cursor opaco decodificado por el núcleo (catalog.Cursor,
// reutilizado con ID=serviceID), paginación "pedir uno de más" igual que
// Repository.List.
func (r *AssignmentRepository) List(ctx context.Context, barbershopID, barberID string, cursor *catalog.Cursor, limit int) (catalog.AssignmentListResult, error) {
	var result catalog.AssignmentListResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var (
			rows pgx.Rows
			err  error
		)
		fetchLimit := limit + 1

		if cursor == nil {
			rows, err = q.Query(ctx,
				`SELECT service_id, created_at
				   FROM barber_service
				  WHERE barbershop_id = $1 AND barber_id = $2
				  ORDER BY created_at, service_id
				  LIMIT $3`,
				barbershopID, barberID, fetchLimit,
			)
		} else {
			rows, err = q.Query(ctx,
				`SELECT service_id, created_at
				   FROM barber_service
				  WHERE barbershop_id = $1 AND barber_id = $2
				    AND (created_at, service_id) > ($3, $4)
				  ORDER BY created_at, service_id
				  LIMIT $5`,
				barbershopID, barberID, cursor.CreatedAt, cursor.ID, fetchLimit,
			)
		}
		if err != nil {
			return fmt.Errorf("list barber_service: query: %w", err)
		}
		defer rows.Close()

		items := make([]catalog.Assignment, 0, fetchLimit)
		for rows.Next() {
			a := catalog.Assignment{BarberID: barberID}
			if err := rows.Scan(&a.ServiceID, &a.CreatedAt); err != nil {
				return fmt.Errorf("list barber_service: scan: %w", err)
			}
			items = append(items, a)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("list barber_service: rows: %w", err)
		}

		hasMore := len(items) > limit
		if hasMore {
			items = items[:limit]
		}

		result.Items = items
		if hasMore {
			last := items[len(items)-1]
			result.NextCursor = catalog.EncodeCursor(catalog.Cursor{CreatedAt: last.CreatedAt, ID: last.ServiceID})
		}
		return nil
	})
	if err != nil {
		return catalog.AssignmentListResult{}, fmt.Errorf("catalog/postgres: list assignments: %w", err)
	}
	return result, nil
}

// Assign implementa catalog.AssignmentRepository.Assign dentro de UNA sola
// InTenantTx: primero confirma que serviceID exista en barbershopID (dueño
// de esa verificación: catalog es dueño de `service`), luego
// INSERT ... ON CONFLICT DO NOTHING (CA-023-02: repetir exactamente la
// misma operación no crea una segunda fila). La FK compuesta
// barber_service_barber_fk es defensa en profundidad sobre barberID
// (CA-023-04): AssignmentService ya lo verificó mediante BarberPort antes
// de esta llamada.
func (r *AssignmentRepository) Assign(ctx context.Context, barbershopID, barberID, serviceID string) (catalog.AssignResult, error) {
	var result catalog.AssignResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var serviceExists bool
		if err := q.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM service WHERE id = $1 AND barbershop_id = $2)`,
			serviceID, barbershopID,
		).Scan(&serviceExists); err != nil {
			return fmt.Errorf("check service exists: %w", err)
		}
		if !serviceExists {
			result.Outcome = catalog.AssignOutcomeServiceNotFound
			return nil
		}

		a := catalog.Assignment{BarberID: barberID, ServiceID: serviceID}
		err := q.QueryRow(ctx,
			`INSERT INTO barber_service (barbershop_id, barber_id, service_id)
			      VALUES ($1, $2, $3)
			 ON CONFLICT (barbershop_id, barber_id, service_id) DO NOTHING
			  RETURNING created_at`,
			barbershopID, barberID, serviceID,
		).Scan(&a.CreatedAt)
		switch {
		case err == nil:
			result.Outcome = catalog.AssignOutcomeCreated
			result.Assignment = a
		case errors.Is(err, pgx.ErrNoRows):
			// ON CONFLICT DO NOTHING no devolvió fila: ya existía. Se lee el
			// created_at real de la fila existente (CA-023-02: la respuesta
			// de una repetición refleja la asignación original, no un
			// instante nuevo inventado).
			if err := q.QueryRow(ctx,
				`SELECT created_at FROM barber_service
				  WHERE barbershop_id = $1 AND barber_id = $2 AND service_id = $3`,
				barbershopID, barberID, serviceID,
			).Scan(&a.CreatedAt); err != nil {
				return fmt.Errorf("select existing barber_service: %w", err)
			}
			result.Outcome = catalog.AssignOutcomeAlreadyExists
			result.Assignment = a
		default:
			return fmt.Errorf("insert barber_service: %w", err)
		}
		return nil
	})
	if err != nil {
		return catalog.AssignResult{}, fmt.Errorf("catalog/postgres: assign service: %w", err)
	}
	return result, nil
}

// Unassign implementa catalog.AssignmentRepository.Unassign dentro de UNA
// sola InTenantTx (DEC-068):
//
//  1. `SELECT is_active FROM service ... FOR UPDATE` bloquea la fila de
//     `service` hasta el fin de esta transacción: dos desasignaciones
//     concurrentes del mismo service_id se SERIALIZAN aquí, sin importar
//     qué barber_id retiren cada una. Sin fila -> UnassignOutcomeNotFound.
//  2. Confirma que la asociación (barbershop_id, barber_id, service_id)
//     exista -> UnassignOutcomeNotFound si no.
//  3. Cuenta cuántas filas activas quedan para ese service_id (incluida la
//     que se retiraría). Si el servicio está activo y esa cuenta es <= 1,
//     esta sería la última: UnassignOutcomeLastActiveConflict, sin borrar
//     nada.
//  4. En cualquier otro caso, DELETE y UnassignOutcomeDeleted.
//
// Porque el paso 1 mantiene el lock hasta el COMMIT/ROLLBACK de esta
// transacción, una segunda transacción concurrente que intente retirar la
// penúltima fila del MISMO servicio espera aquí; cuando se reanuda, ve la
// cuenta YA actualizada (una menos) y rechaza correctamente si le toca ser
// la última, sin ninguna ventana de carrera.
func (r *AssignmentRepository) Unassign(ctx context.Context, barbershopID, barberID, serviceID string) (catalog.UnassignResult, error) {
	var result catalog.UnassignResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var isActive bool
		err := q.QueryRow(ctx,
			`SELECT is_active FROM service WHERE id = $1 AND barbershop_id = $2 FOR UPDATE`,
			serviceID, barbershopID,
		).Scan(&isActive)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			result.Outcome = catalog.UnassignOutcomeNotFound
			return nil
		case err != nil:
			return fmt.Errorf("lock service: %w", err)
		}

		var targetExists bool
		if err := q.QueryRow(ctx,
			`SELECT EXISTS(
			   SELECT 1 FROM barber_service
			    WHERE barbershop_id = $1 AND barber_id = $2 AND service_id = $3
			 )`,
			barbershopID, barberID, serviceID,
		).Scan(&targetExists); err != nil {
			return fmt.Errorf("check target assignment: %w", err)
		}
		if !targetExists {
			result.Outcome = catalog.UnassignOutcomeNotFound
			return nil
		}

		var activeCount int
		if err := q.QueryRow(ctx,
			`SELECT count(*) FROM barber_service WHERE barbershop_id = $1 AND service_id = $2`,
			barbershopID, serviceID,
		).Scan(&activeCount); err != nil {
			return fmt.Errorf("count assignments: %w", err)
		}

		if isActive && activeCount <= 1 {
			result.Outcome = catalog.UnassignOutcomeLastActiveConflict
			return nil
		}

		tag, err := q.Exec(ctx,
			`DELETE FROM barber_service WHERE barbershop_id = $1 AND barber_id = $2 AND service_id = $3`,
			barbershopID, barberID, serviceID,
		)
		if err != nil {
			return fmt.Errorf("delete barber_service: %w", err)
		}
		if tag.RowsAffected() == 0 {
			// No debería ocurrir: targetExists ya confirmó la fila dentro
			// de esta misma transacción, y ninguna otra transacción puede
			// haberla borrado mientras el lock FOR UPDATE de service sigue
			// vigente en esta conexión (DEC-068 solo protege ESE service_id,
			// pero esta fila concreta solo puede desaparecer a través de
			// esta misma operación).
			return fmt.Errorf("delete barber_service: no rows affected despite existing target")
		}
		result.Outcome = catalog.UnassignOutcomeDeleted
		return nil
	})
	if err != nil {
		return catalog.UnassignResult{}, fmt.Errorf("catalog/postgres: unassign service: %w", err)
	}
	return result, nil
}
