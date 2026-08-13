package auth

import (
	"context"
	"fmt"

	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/clock"
)

// decoyBarbershopID es un identificador de barbería que NUNCA existe en
// datos reales (gen_random_uuid() jamás produce el UUID nulo). Se usa como
// tenant señuelo cuando ResolveLoginTenant no resolvió ninguna barbería, así
// LookupCredential ejecuta la MISMA forma de transacción y de consultas que
// cuando sí hay un tenant real: un correo inexistente implica el mismo
// trabajo de base de datos que una barbería resuelta con contraseña
// incorrecta (CA-005-02). Es una elección de implementación, no una
// decisión de producto que requiera registro en docs/00-control
// (estandar-backend-go.md §1: "una excepción es válida cuando simplifica un
// caso real y queda explicada en la revisión").
const decoyBarbershopID = "00000000-0000-0000-0000-000000000000"

// dummyPassword es una contraseña ficticia fija, usada SOLO para calcular el
// hash señuelo al construir el servicio. No corresponde a ninguna cuenta
// real ni se compara jamás contra la contraseña de un cliente.
const dummyPassword = "correo-inexistente-o-credencial-ausente-nunca-coincide-con-nada"

// LoginService implementa el caso de uso de HU-005: inicio de sesión con
// correo y contraseña. No conoce Chi, net/http, JSON ni PostgreSQL
// (CA-002-06): recibe sus dependencias por constructor.
type LoginService struct {
	repo      Repository
	hasher    PasswordHasher
	tokens    TokenGenerator
	clock     clock.Clock
	dummyHash string
}

// NewLoginService construye el servicio. Calcula el hash señuelo una sola
// vez, no en cada Login: cada solicitud fallida ya paga el costo
// criptográfico completo de una verificación real; recalcular la sal del
// señuelo en cada llamada no aporta nada y solo desplaza cuándo ocurre el
// costo de inicialización.
func NewLoginService(repo Repository, hasher PasswordHasher, tokens TokenGenerator, clk clock.Clock) (*LoginService, error) {
	dummyHash, err := hasher.Hash(dummyPassword)
	if err != nil {
		return nil, fmt.Errorf("auth: preparar hash señuelo: %w", err)
	}
	return &LoginService{repo: repo, hasher: hasher, tokens: tokens, clock: clk, dummyHash: dummyHash}, nil
}

// Login verifica correo y contraseña y, si son válidos, emite una sesión
// nueva. Devuelve exactamente el mismo error (apperr.KindUnauthorized,
// invalidCredentialsMessage) para correo inexistente, usuario inactivo y
// contraseña incorrecta (CA-005-02, CA-005-07): ninguna rama de este método
// puede producir un resultado observable distinto entre esos tres casos.
// LookupCredential iguala la forma del trabajo de base de datos entre el
// camino resuelto y el no resuelto; esta función siempre ejecuta exactamente
// una llamada a Verify, con un hash del mismo costo (real o señuelo), sin
// importar qué rama se tomó antes.
func (s *LoginService) Login(ctx context.Context, rawEmail, password string) (Session, error) {
	if err := ctx.Err(); err != nil {
		return Session{}, apperr.Internal(fmt.Errorf("auth: contexto cancelado antes de iniciar sesión: %w", err))
	}

	email := NormalizeEmail(rawEmail)

	barbershopID, resolved, err := s.repo.ResolveLoginTenant(ctx, email)
	if err != nil {
		return Session{}, apperr.Internal(fmt.Errorf("auth: resolver tenant de acceso: %w", err))
	}

	lookupShop := barbershopID
	if !resolved {
		lookupShop = decoyBarbershopID
	}

	cred, err := s.repo.LookupCredential(ctx, lookupShop, resolved, email)
	if err != nil {
		return Session{}, apperr.Internal(fmt.Errorf("auth: buscar credencial: %w", err))
	}

	hashToVerify := s.dummyHash
	if cred.Found {
		hashToVerify = cred.PasswordHash
	}
	passwordMatches := s.hasher.Verify(hashToVerify, password)

	if !resolved || !cred.Found || !passwordMatches {
		return Session{}, errInvalidCredentials()
	}

	rawToken, err := s.tokens.New()
	if err != nil {
		return Session{}, apperr.Internal(fmt.Errorf("auth: generar token de sesión: %w", err))
	}

	now := s.clock.Now()
	expiresAt := now.Add(SessionDuration)

	if err := s.repo.CreateSession(ctx, barbershopID, cred.StaffUserID, HashToken(rawToken), now, expiresAt); err != nil {
		return Session{}, apperr.Internal(fmt.Errorf("auth: persistir sesión: %w", err))
	}

	return Session{
		Token:        rawToken,
		BarbershopID: barbershopID,
		StaffUserID:  cred.StaffUserID,
		IssuedAt:     now,
		ExpiresAt:    expiresAt,
	}, nil
}
