// Package postgres es el adaptador de persistencia de
// publicbooking.Repository sobre database.DB (CA-002-06: el núcleo de
// publicbooking no importa este paquete ni pgx; es este paquete el que
// importa publicbooking).
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"system-barbershop/internal/modules/publicbooking"
	"system-barbershop/internal/platform/database"
)

// Repository implementa publicbooking.Repository sobre database.DB.
type Repository struct {
	db *database.DB
}

// New construye el repositorio.
func New(db *database.DB) *Repository {
	return &Repository{db: db}
}

var _ publicbooking.Repository = (*Repository)(nil)

// ResolveBySlug implementa publicbooking.Repository.ResolveBySlug en dos
// pasos, exactamente el patrón ya establecido por
// authn_resolve_login_tenant/authn_resolve_session_tenant (HU-005,
// database.DB.ResolveTenant): primero resuelve slug a un identificador de
// barbería SIN contexto de tenant, mediante la función SECURITY DEFINER
// estrecha public_resolve_barbershop_by_slug (que responde igual -NULL- a
// forma inválida, slug desconocido o barbería no publicable); si resuelve,
// abre una transacción tenant-aware normal (InTenantTx) sobre ESE
// identificador para leer los campos públicos mínimos ya protegidos por
// barbershop_select_tenant_policy, sin necesitar ningún privilegio
// adicional sobre la tabla.
func (r *Repository) ResolveBySlug(ctx context.Context, slug string) (publicbooking.BarbershopProfile, bool, error) {
	barbershopID, found, err := r.db.ResolveTenant(ctx,
		`SELECT public_resolve_barbershop_by_slug($1)`,
		slug,
	)
	if err != nil {
		return publicbooking.BarbershopProfile{}, false, fmt.Errorf("publicbooking/postgres: resolve tenant by slug: %w", err)
	}
	if !found {
		return publicbooking.BarbershopProfile{}, false, nil
	}

	var profile publicbooking.BarbershopProfile
	profileFound := false
	err = r.db.InTenantTx(ctx, barbershopID, func(ctx context.Context, q database.Queries) error {
		err := q.QueryRow(ctx,
			`SELECT name, timezone, contact_email, contact_phone
			   FROM barbershop
			  WHERE id = $1`,
			string(barbershopID),
		).Scan(&profile.Name, &profile.Timezone, &profile.ContactEmail, &profile.ContactPhone)
		switch {
		case err == nil:
			profileFound = true
		case errors.Is(err, pgx.ErrNoRows):
			// profileFound queda false: defensivo (la función SECURITY
			// DEFINER ya confirmó la fila en la misma sentencia SQL que
			// esta transacción re-lee; una desaparición entre ambas solo
			// podría venir de un borrado físico que este esquema no
			// permite para barbershop).
		default:
			return fmt.Errorf("select public barbershop profile: %w", err)
		}
		return nil
	})
	if err != nil {
		return publicbooking.BarbershopProfile{}, false, fmt.Errorf("publicbooking/postgres: read public profile: %w", err)
	}
	return profile, profileFound, nil
}
