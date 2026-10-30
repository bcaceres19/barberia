package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/platform/database"
)

// PhoneChallengeRepository implementa auth.PhoneChallengeRepository contra
// auth_phone_challenge_request/verify (HU-007, DDL-AUT-01): ambas funciones
// corren ANTES de resolver contexto de tenant (resuelven la cuenta por
// correo internamente), así que usan la misma excepción angosta
// database.DB.CallSecurityDefinerRow que ThrottleRepository.
type PhoneChallengeRepository struct {
	db *database.DB
}

// NewPhoneChallengeRepository construye el repositorio.
func NewPhoneChallengeRepository(db *database.DB) *PhoneChallengeRepository {
	return &PhoneChallengeRepository{db: db}
}

var _ auth.PhoneChallengeRepository = (*PhoneChallengeRepository)(nil)

// RequestChallenge implementa auth.PhoneChallengeRepository. Los destinos
// salen de auth_phone_challenge_request según el canal de la barbería de la
// cuenta (DEC-139); este método no los calcula ni los filtra.
func (r *PhoneChallengeRepository) RequestChallenge(ctx context.Context, email, ipHash, codeHash string, cfg auth.PhoneChallengeConfig) (bool, auth.ChallengeDestinations, error) {
	var accepted bool
	var phone, accountEmail *string

	err := r.db.CallSecurityDefinerRow(ctx,
		`SELECT accepted, phone, email
		   FROM auth_phone_challenge_request($1, $2, $3, $4, $5, $6, $7)`,
		[]any{
			email, ipHash, codeHash,
			cfg.ExpiresSeconds, cfg.RateWindowSeconds, cfg.RateMaxActive, cfg.ResendCooldownSeconds,
		},
		func(row pgx.Row) error {
			return row.Scan(&accepted, &phone, &accountEmail)
		},
	)
	if err != nil {
		return false, auth.ChallengeDestinations{}, fmt.Errorf("auth/postgres: request phone challenge: %w", err)
	}

	var dest auth.ChallengeDestinations
	if phone != nil {
		dest.Phone = *phone
	}
	if accountEmail != nil {
		dest.Email = *accountEmail
	}
	return accepted, dest, nil
}

// VerifyChallenge implementa auth.PhoneChallengeRepository.
func (r *PhoneChallengeRepository) VerifyChallenge(ctx context.Context, email, ipHash, codeHash string) (bool, error) {
	var ok bool

	err := r.db.CallSecurityDefinerRow(ctx,
		`SELECT auth_phone_challenge_verify($1, $2, $3)`,
		[]any{email, ipHash, codeHash},
		func(row pgx.Row) error {
			return row.Scan(&ok)
		},
	)
	if err != nil {
		return false, fmt.Errorf("auth/postgres: verify phone challenge: %w", err)
	}
	return ok, nil
}
