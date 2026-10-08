package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"system-barbershop/internal/modules/shops"
	"system-barbershop/internal/platform/database"
)

// PublicLinkRepository implementa shops.PublicLinkRepository sobre
// database.DB. Comparte con Repository.Update la generación del slug
// (claimSlug) para que haya una sola forma de producirlo.
type PublicLinkRepository struct {
	db      *database.DB
	newCode slugCodeFunc
}

// NewPublicLinkRepository construye el repositorio del enlace público.
func NewPublicLinkRepository(db *database.DB) *PublicLinkRepository {
	return &PublicLinkRepository{db: db, newCode: shops.NewSlugCode}
}

var _ shops.PublicLinkRepository = (*PublicLinkRepository)(nil)

// Ensure implementa shops.PublicLinkRepository.Ensure. FOR UPDATE serializa
// dos lecturas simultáneas de la misma barbería: la segunda espera, ve el
// slug que dejó la primera y no genera otro. RLS acota la fila al tenant
// vigente; el filtro explícito WHERE id = $1 es defensa en profundidad.
func (r *PublicLinkRepository) Ensure(ctx context.Context, barbershopID string) (string, bool, error) {
	var (
		slug  string
		found bool
	)

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var (
			name        string
			currentSlug *string
		)
		err := q.QueryRow(ctx,
			`SELECT name, public_slug FROM barbershop WHERE id = $1 FOR UPDATE`,
			barbershopID,
		).Scan(&name, &currentSlug)
		switch {
		case err == nil:
		case errors.Is(err, pgx.ErrNoRows):
			return nil // found queda false: defensivo.
		default:
			return fmt.Errorf("lock barbershop for public link: %w", err)
		}
		found = true

		if currentSlug != nil {
			slug = *currentSlug
			return nil
		}
		generated, err := claimSlug(ctx, q, r.newCode, barbershopID, name)
		if err != nil {
			return err
		}
		slug = generated
		return nil
	})
	if err != nil {
		return "", false, fmt.Errorf("shops/postgres: ensure public link: %w", err)
	}
	return slug, found, nil
}
