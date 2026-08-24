// Package postgres es el adaptador de persistencia de staff.Repository
// sobre database.DB (CA-002-06: el núcleo de staff no importa este paquete
// ni pgx; es este paquete el que importa staff), mismo patrón que
// shops/postgres.
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"system-barbershop/internal/modules/staff"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

// createBarberOperation identifica, dentro de una barbería, la operación
// de idempotencia del alta de barberos (RN-IDE-01): una clave ya usada
// para OTRA operación (por ejemplo una futura "crear cita") nunca se
// confunde con esta.
const createBarberOperation idempotency.Operation = "create_barber"

// createBarberIdempotencyTTL es la vigencia de una reclamación de
// idempotencia sobre el alta de un barbero antes de tratarse como si nunca
// hubiera existido (CA-004-05). 24 horas cubre con margen cualquier
// reintento de red razonable sin dejar el mecanismo vigente indefinidamente.
const createBarberIdempotencyTTL = 24 * time.Hour

// Repository implementa staff.Repository sobre database.DB.
type Repository struct {
	db    *database.DB
	coord idempotency.Coordinator
}

// New construye el repositorio con el coordinador de idempotencia real
// (SQLCoordinator). El parámetro coord existe para pruebas: internal/
// platform/idempotency ya prueba SQLCoordinator de forma exhaustiva contra
// PostgreSQL real, así que este paquete no necesita repetir esas pruebas,
// pero sí necesita poder construir el repositorio con el mismo adaptador
// que producción usa.
func New(db *database.DB, coord idempotency.Coordinator) *Repository {
	return &Repository{db: db, coord: coord}
}

var _ staff.Repository = (*Repository)(nil)

// barberResponseWire es la forma JSON EXACTA que también usa
// httpapi.BarberResponse (mismos nombres de campo, mismo orden, mismos
// tipos): la respuesta que Complete persiste para una repetición exacta
// (CA-004-01) debe ser BYTE A BYTE idéntica a la que el handler ya
// construyó para la primera ejecución, y ambas deben tener EXACTAMENTE la
// misma forma que un GET/PATCH de este mismo barbero. Si cambia
// httpapi.BarberResponse, este struct debe cambiar igual en el mismo
// commit (ver TestCreate_StoredResponseBody_MatchesHTTPAPIWireShape en
// repository_test.go).
type barberResponseWire struct {
	ID        string    `json:"id"`
	FullName  string    `json:"fullName"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func marshalBarberResponse(b staff.Barber) ([]byte, error) {
	return json.Marshal(barberResponseWire{
		ID:        b.ID,
		FullName:  b.FullName,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	})
}

// List implementa staff.Repository.List: orden estable (created_at, id)
// dentro del tenant vigente, cursor opaco decodificado por el núcleo
// (staff.Cursor), paginación "pedir uno de más" para saber si hay página
// siguiente sin una segunda consulta COUNT.
func (r *Repository) List(ctx context.Context, barbershopID string, cursor *staff.Cursor, limit int) (staff.ListResult, error) {
	var result staff.ListResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var (
			rows pgx.Rows
			err  error
		)
		// fetchLimit pide una fila de más: si vuelve, hay página siguiente.
		fetchLimit := limit + 1

		if cursor == nil {
			rows, err = q.Query(ctx,
				`SELECT id, full_name, created_at, updated_at
				   FROM barber
				  WHERE barbershop_id = $1
				  ORDER BY created_at, id
				  LIMIT $2`,
				barbershopID, fetchLimit,
			)
		} else {
			rows, err = q.Query(ctx,
				`SELECT id, full_name, created_at, updated_at
				   FROM barber
				  WHERE barbershop_id = $1
				    AND (created_at, id) > ($2, $3)
				  ORDER BY created_at, id
				  LIMIT $4`,
				barbershopID, cursor.CreatedAt, cursor.ID, fetchLimit,
			)
		}
		if err != nil {
			return fmt.Errorf("list barbers: query: %w", err)
		}
		defer rows.Close()

		items := make([]staff.Barber, 0, fetchLimit)
		for rows.Next() {
			var b staff.Barber
			if err := rows.Scan(&b.ID, &b.FullName, &b.CreatedAt, &b.UpdatedAt); err != nil {
				return fmt.Errorf("list barbers: scan: %w", err)
			}
			items = append(items, b)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("list barbers: rows: %w", err)
		}

		hasMore := len(items) > limit
		if hasMore {
			items = items[:limit]
		}

		result.Items = items
		if hasMore {
			last := items[len(items)-1]
			result.NextCursor = staff.EncodeCursor(staff.Cursor{CreatedAt: last.CreatedAt, ID: last.ID})
		}
		return nil
	})
	if err != nil {
		return staff.ListResult{}, fmt.Errorf("staff/postgres: list barbers: %w", err)
	}
	return result, nil
}

// Get implementa staff.Repository.Get. RLS (barber_select_tenant_policy) ya
// restringe la fila visible a barbershop_id = current_setting(...); el
// filtro explícito WHERE barbershop_id = $2 es defensa en profundidad,
// mismo patrón que shops/postgres.Repository.Get.
func (r *Repository) Get(ctx context.Context, barbershopID, barberID string) (staff.Barber, bool, error) {
	var b staff.Barber
	found := false

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		b.ID = barberID
		err := q.QueryRow(ctx,
			`SELECT full_name, created_at, updated_at
			   FROM barber
			  WHERE id = $1 AND barbershop_id = $2`,
			barberID, barbershopID,
		).Scan(&b.FullName, &b.CreatedAt, &b.UpdatedAt)
		switch {
		case err == nil:
			found = true
		case errors.Is(err, pgx.ErrNoRows):
			// found queda false: cubre tanto "no existe" como "es de otra
			// barbería" (CA-021-05), sin distinguir la causa.
		default:
			return fmt.Errorf("select barber: %w", err)
		}
		return nil
	})
	if err != nil {
		return staff.Barber{}, false, fmt.Errorf("staff/postgres: get barber: %w", err)
	}
	if !found {
		return staff.Barber{}, false, nil
	}
	return b, true, nil
}

// Create implementa staff.Repository.Create: Begin, INSERT y Complete
// ocurren dentro de la MISMA InTenantTx (apps/api/README.md, "Patrón
// obligatorio: idempotencia reutilizable"). Si Begin no devuelve
// OutcomeProceed, el callback no ejecuta ningún efecto y la transacción
// igual hace COMMIT (Begin ya dejó su propio registro, cuando aplica); si
// el efecto o Complete fallan, InTenantTx hace ROLLBACK de TODA la
// transacción, incluido el INSERT que Begin reclamó, dejando la clave libre
// para un reintento legítimo (CA-004-06).
func (r *Repository) Create(
	ctx context.Context,
	barbershopID string,
	fullName string,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (staff.CreateResult, error) {
	var result staff.CreateResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		decision, err := r.coord.Begin(ctx, q, database.BarbershopID(barbershopID), key, createBarberOperation, fingerprint, createBarberIdempotencyTTL)
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

		var b staff.Barber
		if err := q.QueryRow(ctx,
			`INSERT INTO barber (barbershop_id, full_name)
			      VALUES ($1, $2)
			   RETURNING id, full_name, created_at, updated_at`,
			barbershopID, fullName,
		).Scan(&b.ID, &b.FullName, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return fmt.Errorf("insert barber: %w", err)
		}

		body, err := marshalBarberResponse(b)
		if err != nil {
			return fmt.Errorf("marshal created barber: %w", err)
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
			// No debería ocurrir: Begin reclamó la clave en esta misma
			// transacción hace un instante; ver el comentario de
			// Coordinator.Complete para el único caso legítimo (otra
			// llamada ya la completó), que no puede pasar dentro de una
			// única transacción secuencial como esta.
			return fmt.Errorf("idempotency complete: la reclamación ya no estaba in_progress")
		}

		result.Barber = b
		result.Response = stored
		return nil
	})
	if err != nil {
		return staff.CreateResult{}, fmt.Errorf("staff/postgres: create barber: %w", err)
	}
	return result, nil
}

// Rename implementa staff.Repository.Rename: UPDATE ... RETURNING dentro de
// una única sentencia, filtrando por id Y barbershop_id (defensa en
// profundidad sobre barber_update_tenant_policy). Nunca crea ni duplica una
// fila (CA-021-04); cero filas afectadas cubre tanto "no existe" como "es
// de otra barbería" (CA-021-05).
func (r *Repository) Rename(ctx context.Context, barbershopID, barberID, fullName string) (staff.RenameResult, error) {
	var result staff.RenameResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var b staff.Barber
		b.ID = barberID
		err := q.QueryRow(ctx,
			`UPDATE barber
			    SET full_name = $3
			  WHERE id = $1 AND barbershop_id = $2
			RETURNING full_name, created_at, updated_at`,
			barberID, barbershopID, fullName,
		).Scan(&b.FullName, &b.CreatedAt, &b.UpdatedAt)
		switch {
		case err == nil:
			result.Found = true
			result.Barber = b
		case errors.Is(err, pgx.ErrNoRows):
			// result.Found queda false: defensivo, ver RenameResult.Found.
		default:
			return fmt.Errorf("update barber: %w", err)
		}
		return nil
	})
	if err != nil {
		return staff.RenameResult{}, fmt.Errorf("staff/postgres: rename barber: %w", err)
	}
	return result, nil
}
