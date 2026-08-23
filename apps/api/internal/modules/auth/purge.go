package auth

import (
	"context"
	"fmt"
)

// PurgeRepository es el puerto de mantenimiento de HU-007/HU-008
// (CA-007-06, DEC-064): purga en lote las filas vencidas de
// login_throttle, auth_phone_challenge y staff_recovery_code. Exclusivo
// del proceso worker, conectado como barberia_worker (DEC-040);
// barberia_app no tiene GRANT sobre estas funciones.
type PurgeRepository interface {
	PurgeLoginThrottle(ctx context.Context, limit int) (deleted int, err error)
	PurgePhoneChallenges(ctx context.Context, limit int) (deleted int, err error)
	PurgeRecoveryCodes(ctx context.Context, limit int) (deleted int, err error)
}

// PurgeService implementa el mantenimiento periódico de HU-007/HU-008. No
// conoce Chi, net/http ni PostgreSQL.
type PurgeService struct {
	repo                PurgeRepository
	loginThrottleLimit  int
	phoneChallengeLimit int
	recoveryCodeLimit   int
}

// NewPurgeService construye el servicio.
func NewPurgeService(repo PurgeRepository, loginThrottleLimit, phoneChallengeLimit, recoveryCodeLimit int) *PurgeService {
	return &PurgeService{
		repo:                repo,
		loginThrottleLimit:  loginThrottleLimit,
		phoneChallengeLimit: phoneChallengeLimit,
		recoveryCodeLimit:   recoveryCodeLimit,
	}
}

// PurgeOnce ejecuta un lote de purga de las tres tablas. Cada llamada a las
// funciones SQL ya es su propia transacción corta con SKIP LOCKED
// (estandar-base-datos.md §11): PurgeOnce no abre una transacción propia
// que las envuelva.
func (s *PurgeService) PurgeOnce(ctx context.Context) (loginThrottleDeleted, phoneChallengeDeleted, recoveryCodeDeleted int, err error) {
	if err := ctx.Err(); err != nil {
		return 0, 0, 0, fmt.Errorf("auth: contexto cancelado antes de purgar: %w", err)
	}

	loginThrottleDeleted, err = s.repo.PurgeLoginThrottle(ctx, s.loginThrottleLimit)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("auth: purgar login_throttle: %w", err)
	}

	phoneChallengeDeleted, err = s.repo.PurgePhoneChallenges(ctx, s.phoneChallengeLimit)
	if err != nil {
		return loginThrottleDeleted, 0, 0, fmt.Errorf("auth: purgar auth_phone_challenge: %w", err)
	}

	recoveryCodeDeleted, err = s.repo.PurgeRecoveryCodes(ctx, s.recoveryCodeLimit)
	if err != nil {
		return loginThrottleDeleted, phoneChallengeDeleted, 0, fmt.Errorf("auth: purgar staff_recovery_code: %w", err)
	}

	return loginThrottleDeleted, phoneChallengeDeleted, recoveryCodeDeleted, nil
}
