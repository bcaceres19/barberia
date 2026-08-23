// Package postgres es el adaptador de persistencia de shops.Repository
// sobre database.DB (CA-002-06: el núcleo de shops no importa este paquete
// ni pgx; es este paquete el que importa shops).
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"system-barbershop/internal/modules/shops"
	"system-barbershop/internal/platform/database"
)

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
	})
	if err != nil {
		return shops.UpdateResult{}, fmt.Errorf("shops/postgres: update barbershop: %w", err)
	}
	return result, nil
}
