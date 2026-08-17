package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/platform/database"
)

// PurgeRepository implementa auth.PurgeRepository contra
// login_throttle_purge_expired/auth_phone_challenge_purge_expired (HU-007,
// CA-007-06). Ambas funciones están concedidas exclusivamente a
// barberia_worker; db debe ser el pool del proceso worker
// (cfg.WorkerDatabaseURL, DEC-040), nunca el de api.
type PurgeRepository struct {
	db *database.DB
}

// NewPurgeRepository construye el repositorio.
func NewPurgeRepository(db *database.DB) *PurgeRepository {
	return &PurgeRepository{db: db}
}

var _ auth.PurgeRepository = (*PurgeRepository)(nil)

// PurgeLoginThrottle implementa auth.PurgeRepository.
func (r *PurgeRepository) PurgeLoginThrottle(ctx context.Context, limit int) (int, error) {
	var deleted int
	err := r.db.CallSecurityDefinerRow(ctx,
		`SELECT login_throttle_purge_expired($1)`, []any{limit},
		func(row pgx.Row) error { return row.Scan(&deleted) },
	)
	if err != nil {
		return 0, fmt.Errorf("auth/postgres: purge login_throttle: %w", err)
	}
	return deleted, nil
}

// PurgePhoneChallenges implementa auth.PurgeRepository.
func (r *PurgeRepository) PurgePhoneChallenges(ctx context.Context, limit int) (int, error) {
	var deleted int
	err := r.db.CallSecurityDefinerRow(ctx,
		`SELECT auth_phone_challenge_purge_expired($1)`, []any{limit},
		func(row pgx.Row) error { return row.Scan(&deleted) },
	)
	if err != nil {
		return 0, fmt.Errorf("auth/postgres: purge auth_phone_challenge: %w", err)
	}
	return deleted, nil
}
