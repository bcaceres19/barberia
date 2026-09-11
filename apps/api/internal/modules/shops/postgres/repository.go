// Package postgres es el adaptador de persistencia de shops.Repository
// sobre database.DB (CA-002-06: el núcleo de shops no importa este paquete
// ni pgx; es este paquete el que importa shops).
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"system-barbershop/internal/modules/shops"
	"system-barbershop/internal/platform/database"
)

// publicSlugUniqueIndex es el nombre del índice único parcial que
// 20260911045044_add_barbershop_public_slug.sql crea sobre
// lower(public_slug) (HU-090, DEC-082).
const publicSlugUniqueIndex = "idx_barbershop_public_slug"

// maxSlugGenerationAttempts acota el reintento de sufijo ante colisión real
// (DEC-082: '-2', '-3', ...). 50 barberías con exactamente el mismo nombre
// es un escenario que no ocurre en el MVP de un solo tenant por despliegue;
// el límite existe para que un defecto real nunca deje la transacción en
// un bucle infinito.
const maxSlugGenerationAttempts = 50

// isPublicSlugConflict informa si err es el unique_violation (23505) de
// publicSlugUniqueIndex. Cualquier otra violación devuelve false: solo este
// índice concreto dispara el reintento de sufijo de SlugWithSuffix.
func isPublicSlugConflict(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23505" && pgErr.ConstraintName == publicSlugUniqueIndex
}

// Repository implementa shops.Repository sobre database.DB.
type Repository struct {
	db *database.DB
}

// New construye el repositorio.
func New(db *database.DB) *Repository {
	return &Repository{db: db}
}

var _ shops.Repository = (*Repository)(nil)

// Get implementa shops.Repository.Get leyendo los cuatro campos autorizados
// dentro de una transacción tenant-aware. RLS
// (barbershop_select_tenant_policy) ya restringe la fila visible a
// id = current_setting('app.barbershop_id'); el filtro explícito
// WHERE id = $1 es defensa en profundidad, mismo patrón que
// auth/postgres.Repository.BarbershopName.
func (r *Repository) Get(ctx context.Context, barbershopID string) (shops.Barbershop, bool, error) {
	var b shops.Barbershop
	found := false

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		err := q.QueryRow(ctx,
			`SELECT name, timezone, contact_email, contact_phone
			   FROM barbershop
			  WHERE id = $1`,
			barbershopID,
		).Scan(&b.Name, &b.Timezone, &b.ContactEmail, &b.ContactPhone)
		switch {
		case err == nil:
			found = true
		case errors.Is(err, pgx.ErrNoRows):
			// found queda false: defensivo, ver shops.UpdateResult.Found.
		default:
			return fmt.Errorf("select barbershop: %w", err)
		}
		return nil
	})
	if err != nil {
		return shops.Barbershop{}, false, fmt.Errorf("shops/postgres: get barbershop: %w", err)
	}
	return b, found, nil
}

// Update implementa shops.Repository.Update. Dentro de UNA sola
// transacción tenant-aware: primero confirma input.Timezone contra
// pg_timezone_names con una consulta parametrizada exacta (nunca
// concatenada); si no existe, devuelve inmediatamente sin ejecutar ningún
// UPDATE (CA-020-03: cero escritura, Name y Contact* anteriores quedan
// intactos). Si la zona es válida, ejecuta el UPDATE filtrando también por
// barbershopID (defensa en profundidad sobre
// barbershop_update_tenant_policy) y devuelve la representación recién
// guardada en la misma sentencia (RETURNING), sin una segunda consulta.
//
// HU-090/DEC-082: si public_slug todavía es NULL, esta misma transacción lo
// genera a partir de input.Name (shops.SlugBase) y lo persiste junto con el
// resto de campos -el barbero nunca lo escribe, y esta operación nunca lo
// TOCA si ya existe (editar el slug es una capacidad fuera del alcance de
// HU-090, aunque la columna ya la soporta). Una colisión real de unicidad
// global (dos barberías cuyo nombre produce el mismo slug) se resuelve
// reintentando con shops.SlugWithSuffix, capturando el unique_violation
// exacto de idx_barbershop_public_slug -el mismo patrón ya usado por
// catalog/postgres.isNameConflict-, nunca revisando existencia antes con
// una segunda consulta (evita una condición de carrera contra otro tenant
// escribiendo su propio slug al mismo tiempo).
func (r *Repository) Update(ctx context.Context, barbershopID string, input shops.UpdateInput) (shops.UpdateResult, error) {
	var result shops.UpdateResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var validTimezone bool
		if err := q.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM pg_timezone_names WHERE name = $1)`,
			input.Timezone,
		).Scan(&validTimezone); err != nil {
			return fmt.Errorf("check timezone: %w", err)
		}
		if !validTimezone {
			// result.TimezoneValid queda false (valor cero); ningún UPDATE
			// se ejecuta en esta transacción, así que Name/Contact*
			// anteriores no se tocan.
			return nil
		}
		result.TimezoneValid = true

		// FOR UPDATE: la fila propia queda bloqueada por el resto de la
		// transacción, coherente con "UNA sola transacción tenant-aware"
		// (mismo criterio que el resto del módulo); RLS ya acota la fila a
		// la barbería vigente, así que este bloqueo nunca alcanza otro
		// tenant.
		var currentSlug *string
		if err := q.QueryRow(ctx,
			`SELECT public_slug FROM barbershop WHERE id = $1 FOR UPDATE`,
			barbershopID,
		).Scan(&currentSlug); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil // result.Found queda false: defensivo, ver shops.UpdateResult.Found.
			}
			return fmt.Errorf("lock barbershop before slug generation: %w", err)
		}

		if currentSlug != nil {
			var b shops.Barbershop
			err := q.QueryRow(ctx,
				`UPDATE barbershop
				    SET name = $2, timezone = $3, contact_email = $4, contact_phone = $5
				  WHERE id = $1
				RETURNING name, timezone, contact_email, contact_phone`,
				barbershopID, input.Name, input.Timezone, input.ContactEmail, input.ContactPhone,
			).Scan(&b.Name, &b.Timezone, &b.ContactEmail, &b.ContactPhone)
			switch {
			case err == nil:
				result.Found = true
				result.Barbershop = b
			case errors.Is(err, pgx.ErrNoRows):
				// result.Found queda false: defensivo, ver shops.UpdateResult.Found.
			default:
				return fmt.Errorf("update barbershop: %w", err)
			}
			return nil
		}

		base := shops.SlugBase(input.Name)
		candidate := base
		for attempt := 2; ; attempt++ {
			// SAVEPOINT: un unique_violation dentro de PostgreSQL aborta el
			// resto de la transacción hasta un ROLLBACK (SQLSTATE 25P02);
			// sin este punto de retorno, el segundo intento de UPDATE
			// fallaría con "current transaction is aborted" en lugar de
			// probar el siguiente candidato (confirmado contra PostgreSQL
			// real: ver TestUpdate_SlugCollisionAcrossTenants_AppendsDeterministicSuffix).
			if _, err := q.Exec(ctx, `SAVEPOINT slug_attempt`); err != nil {
				return fmt.Errorf("savepoint before slug attempt: %w", err)
			}

			var b shops.Barbershop
			err := q.QueryRow(ctx,
				`UPDATE barbershop
				    SET name = $2, timezone = $3, contact_email = $4, contact_phone = $5, public_slug = $6
				  WHERE id = $1
				RETURNING name, timezone, contact_email, contact_phone`,
				barbershopID, input.Name, input.Timezone, input.ContactEmail, input.ContactPhone, candidate,
			).Scan(&b.Name, &b.Timezone, &b.ContactEmail, &b.ContactPhone)
			switch {
			case err == nil:
				if _, releaseErr := q.Exec(ctx, `RELEASE SAVEPOINT slug_attempt`); releaseErr != nil {
					return fmt.Errorf("release savepoint after successful slug attempt: %w", releaseErr)
				}
				result.Found = true
				result.Barbershop = b
				return nil
			case errors.Is(err, pgx.ErrNoRows):
				if _, releaseErr := q.Exec(ctx, `RELEASE SAVEPOINT slug_attempt`); releaseErr != nil {
					return fmt.Errorf("release savepoint after no-op slug attempt: %w", releaseErr)
				}
				// result.Found queda false: defensivo, ver shops.UpdateResult.Found.
				return nil
			case isPublicSlugConflict(err):
				if _, rollbackErr := q.Exec(ctx, `ROLLBACK TO SAVEPOINT slug_attempt`); rollbackErr != nil {
					return fmt.Errorf("rollback to savepoint after slug conflict: %w", rollbackErr)
				}
				if attempt > maxSlugGenerationAttempts {
					return fmt.Errorf("generate unique public_slug: agotados %d intentos de sufijo", maxSlugGenerationAttempts)
				}
				candidate = shops.SlugWithSuffix(base, attempt)
				continue
			default:
				return fmt.Errorf("update barbershop with generated slug: %w", err)
			}
		}
	})
	if err != nil {
		return shops.UpdateResult{}, fmt.Errorf("shops/postgres: update barbershop: %w", err)
	}
	return result, nil
}
