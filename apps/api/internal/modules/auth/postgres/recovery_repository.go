package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/platform/database"
)

// RecoveryRepository implementa auth.RecoveryRepository contra
// auth_recovery_request/verify/current_credential/change_password (HU-008,
// DDL-AUT-01): las cuatro funciones corren ANTES de resolver contexto de
// tenant (resuelven la cuenta por correo internamente), así que usan la
// misma excepción angosta database.DB.CallSecurityDefinerRow que
// ThrottleRepository y PhoneChallengeRepository.
type RecoveryRepository struct {
	db *database.DB
}

// NewRecoveryRepository construye el repositorio.
func NewRecoveryRepository(db *database.DB) *RecoveryRepository {
	return &RecoveryRepository{db: db}
}

var _ auth.RecoveryRepository = (*RecoveryRepository)(nil)

// RequestRecovery implementa auth.RecoveryRepository.
func (r *RecoveryRepository) RequestRecovery(ctx context.Context, email, codeHash string, cfg auth.RecoveryConfig) (bool, string, string, error) {
	var accepted bool
	var phone, resolvedEmail *string

	err := r.db.CallSecurityDefinerRow(ctx,
		`SELECT accepted, phone, email
		   FROM auth_recovery_request($1, $2, $3, $4, $5, $6, $7)`,
		[]any{
			email, codeHash,
			cfg.CodeExpiresSeconds, cfg.CodeMaxAttempts,
			cfg.ResendCooldownSeconds, cfg.ResendWindowSeconds, cfg.ResendMaxPerWindow,
		},
		func(row pgx.Row) error {
			return row.Scan(&accepted, &phone, &resolvedEmail)
		},
	)
	if err != nil {
		return false, "", "", fmt.Errorf("auth/postgres: request recovery: %w", err)
	}

	return accepted, derefOrEmpty(phone), derefOrEmpty(resolvedEmail), nil
}

// VerifyRecovery implementa auth.RecoveryRepository.
func (r *RecoveryRepository) VerifyRecovery(ctx context.Context, email, codeHash, resetTokenHash string, resetExpiresSeconds int) (bool, string, string, error) {
	var ok bool
	var phone, resolvedEmail *string

	err := r.db.CallSecurityDefinerRow(ctx,
		`SELECT ok, phone, email
		   FROM auth_recovery_verify($1, $2, $3, $4)`,
		[]any{email, codeHash, resetTokenHash, resetExpiresSeconds},
		func(row pgx.Row) error {
			return row.Scan(&ok, &phone, &resolvedEmail)
		},
	)
	if err != nil {
		return false, "", "", fmt.Errorf("auth/postgres: verify recovery: %w", err)
	}

	return ok, derefOrEmpty(phone), derefOrEmpty(resolvedEmail), nil
}

// CurrentCredential implementa auth.RecoveryRepository.
func (r *RecoveryRepository) CurrentCredential(ctx context.Context, email, resetTokenHash string) (bool, string, string, error) {
	var found bool
	var passwordHash, passwordAlgorithm *string

	err := r.db.CallSecurityDefinerRow(ctx,
		`SELECT found, password_hash, password_algorithm
		   FROM auth_recovery_current_credential($1, $2)`,
		[]any{email, resetTokenHash},
		func(row pgx.Row) error {
			return row.Scan(&found, &passwordHash, &passwordAlgorithm)
		},
	)
	if err != nil {
		return false, "", "", fmt.Errorf("auth/postgres: recovery current credential: %w", err)
	}

	return found, derefOrEmpty(passwordHash), derefOrEmpty(passwordAlgorithm), nil
}

// ChangePassword implementa auth.RecoveryRepository.
func (r *RecoveryRepository) ChangePassword(ctx context.Context, email, resetTokenHash, newPasswordHash string) (bool, error) {
	var ok bool

	err := r.db.CallSecurityDefinerRow(ctx,
		`SELECT auth_recovery_change_password($1, $2, $3)`,
		[]any{email, resetTokenHash, newPasswordHash},
		func(row pgx.Row) error {
			return row.Scan(&ok)
		},
	)
	if err != nil {
		return false, fmt.Errorf("auth/postgres: recovery change password: %w", err)
	}
	return ok, nil
}

func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
