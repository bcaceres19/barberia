package auth

import (
	"context"
	"fmt"
	"time"

	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/clock"
)

// challengeRequiredMessage es el ÚNICO texto que ThrottleService devuelve
// cuando una IP está escalada (HU-007, CA-007-03): no revela si el correo
// existe ni distingue la razón exacta del escalamiento.
const challengeRequiredMessage = "se requiere verificar el teléfono antes de continuar"

// ThrottleConfig agrupa los valores configurables del límite por IP
// (CA-007-05, DEC-052, DEC-061). Ningún valor vive incrustado en el código:
// todos llegan desde config.Config, que ya los validó al arrancar.
type ThrottleConfig struct {
	WindowSeconds     int
	EscalationSeconds int
	// Threshold es la cantidad de solicitudes permitidas SIN reto dentro de
	// la ventana; la solicitud Threshold+1 lo exige (DEC-061).
	Threshold        int
	RetentionSeconds int
}

// ThrottleRepository es el puerto de persistencia del contador de HU-007. El
// núcleo no importa pgx: postgres/ traduce entre este contrato y
// login_throttle_register_attempt.
type ThrottleRepository interface {
	// RegisterAttempt registra un intento para ipHash y devuelve si, tras
	// este registro, la IP queda escalada. retryAfter solo es válido cuando
	// escalated=true.
	RegisterAttempt(ctx context.Context, ipHash string, cfg ThrottleConfig) (attemptCount int, escalated bool, retryAfter time.Time, err error)
}

// ThrottleService implementa el conteo por IP de HU-007. No conoce Chi,
// net/http ni PostgreSQL (CA-002-06).
type ThrottleService struct {
	repo   ThrottleRepository
	cfg    ThrottleConfig
	secret []byte
	clock  clock.Clock
}

// NewThrottleService construye el servicio. secret firma el HMAC de la IP
// (HU-007, DEC-062); nunca se registra ni se expone.
func NewThrottleService(repo ThrottleRepository, cfg ThrottleConfig, secret []byte, clk clock.Clock) *ThrottleService {
	return &ThrottleService{repo: repo, cfg: cfg, secret: secret, clock: clk}
}

// HashIP calcula el ip_hash estable de rawIP con el secreto de despliegue.
// Se expone para que el handler de HU-007 (challenge/verify) pueda derivar
// el mismo ip_hash sin duplicar el secreto en la capa HTTP.
func (s *ThrottleService) HashIP(rawIP string) string {
	return HMACHex(rawIP, s.secret)
}

// Check registra un intento desde rawIP y devuelve un error
// (apperr.KindChallengeRequired) si la IP queda escalada tras este
// registro. nil significa "continúa con la evaluación normal de la
// contraseña" (CA-007-01). Nunca resuelve tenant ni evalúa contraseña por sí
// mismo: eso es responsabilidad exclusiva del llamador, y solo debe ocurrir
// cuando Check devuelve nil (CA-007-02).
func (s *ThrottleService) Check(ctx context.Context, rawIP string) error {
	if err := ctx.Err(); err != nil {
		return apperr.Internal(fmt.Errorf("auth: contexto cancelado antes de registrar intento: %w", err))
	}

	ipHash := s.HashIP(rawIP)
	_, escalated, retryAfter, err := s.repo.RegisterAttempt(ctx, ipHash, s.cfg)
	if err != nil {
		return apperr.Internal(fmt.Errorf("auth: registrar intento de acceso: %w", err))
	}
	if !escalated {
		return nil
	}

	retrySeconds := int(retryAfter.Sub(s.clock.Now()).Seconds())
	if retrySeconds < 1 {
		retrySeconds = 1
	}
	return apperr.ChallengeRequired(challengeRequiredMessage, retrySeconds)
}
