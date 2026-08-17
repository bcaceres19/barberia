package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/platform/database"
)

// ThrottleRepository implementa auth.ThrottleRepository contra
// login_throttle_register_attempt (HU-007, DDL-AUT-01): sin
// barbershop_id/RLS por diseño, así que corre fuera de InTenantTx mediante
// la excepción angosta database.DB.CallSecurityDefinerRow, igual que
// ResolveLoginTenant.
type ThrottleRepository struct {
	db *database.DB
}

// NewThrottleRepository construye el repositorio.
func NewThrottleRepository(db *database.DB) *ThrottleRepository {
	return &ThrottleRepository{db: db}
}

var _ auth.ThrottleRepository = (*ThrottleRepository)(nil)

// RegisterAttempt implementa auth.ThrottleRepository.
func (r *ThrottleRepository) RegisterAttempt(ctx context.Context, ipHash string, cfg auth.ThrottleConfig) (int, bool, time.Time, error) {
	var attemptCount int
	var escalated bool
	var retryAfter *time.Time

	err := r.db.CallSecurityDefinerRow(ctx,
		`SELECT attempt_count, escalated, retry_after
		   FROM login_throttle_register_attempt($1, $2, $3, $4, $5)`,
		[]any{ipHash, cfg.WindowSeconds, cfg.EscalationSeconds, cfg.Threshold, cfg.RetentionSeconds},
		func(row pgx.Row) error {
			return row.Scan(&attemptCount, &escalated, &retryAfter)
		},
	)
	if err != nil {
		return 0, false, time.Time{}, fmt.Errorf("auth/postgres: register throttle attempt: %w", err)
	}

	var retry time.Time
	if retryAfter != nil {
		retry = *retryAfter
	}
	return attemptCount, escalated, retry, nil
}
