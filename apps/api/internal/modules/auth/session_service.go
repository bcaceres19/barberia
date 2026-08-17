package auth

import (
	"context"
	"fmt"
	"time"

	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/clock"
)

// invalidSessionMessage es el ÚNICO texto que este archivo devuelve para
// cualquier fallo de validación de sesión: cookie ausente, con forma
// inválida, token desconocido, vencido, revocado o de un usuario que dejó
// de estar activo (CA-006-02, CA-006-05). httpapi la usa también para
// rechazar una forma de cookie evidentemente inválida ANTES de calcular su
// hash o tocar PostgreSQL, sin construir un mensaje distinto para ese caso.
const invalidSessionMessage = "sesión inválida o expirada"

// ErrInvalidSession construye el error uniforme de sesión inválida. Se
// exporta porque httpapi.SessionMiddleware lo necesita para el rechazo
// temprano de una cookie con forma inválida, antes de invocar
// [SessionService.Validate] (CA-006-05: ni el hash ni PostgreSQL ven una
// cookie evidentemente hostil).
func ErrInvalidSession() error {
	return apperr.Unauthorized(invalidSessionMessage)
}

// Principal es la identidad ya autenticada de una solicitud privada:
// identificadores opacos únicamente, nunca correo, nombre ni ningún otro
// dato personal (RN-DAT-02). Los handlers privados lo leen del contexto con
// [PrincipalFromContext]; nunca lo reciben de una URL, un cuerpo o una
// cabecera controlada por el cliente.
type Principal struct {
	SessionID    string
	StaffUserID  string
	BarbershopID string
	// ExpiresAt es la expiración YA renovada por esta misma solicitud
	// (DEC-050, renovación deslizante); no es dato personal, es el mismo
	// valor no sensible que LoginResponse ya expone (CA-005-04). HU-012
	// (DEC-060) lo reexpone tal cual en GET /private/auth/session.
	ExpiresAt time.Time
}

type principalContextKey struct{}

// ContextWithPrincipal adjunta p al contexto. Lo usa
// httpapi.SessionMiddleware tras validar la sesión; ningún otro punto del
// código debe construir este valor de contexto.
func ContextWithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, p)
}

// PrincipalFromContext recupera el [Principal] que [ContextWithPrincipal]
// dejó en ctx. ok=false indica que el middleware de sesión no corrió sobre
// esta solicitud: un handler privado nunca debe continuar en ese caso, es
// síntoma de una ruta montada fuera del subrouter protegido (CA-006-04).
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalContextKey{}).(Principal)
	return p, ok
}

// SessionRepository es el puerto de persistencia de HU-006: validar y
// renovar una sesión existente, y revocarla. Deliberadamente separado de
// Repository (HU-005): el flujo de login nunca necesita renovar ni
// revocar, así que unirlos obligaría a cualquier implementación futura a
// resolver operaciones que no le corresponden.
type SessionRepository interface {
	// ResolveSessionTenant llama a la función estrecha
	// authn_resolve_session_tenant ANTES de que exista contexto de tenant
	// (igual que ResolveLoginTenant para el login). found=false cubre un
	// token desconocido, revocado o vencido; esta interfaz no tiene forma
	// de distinguirlos entre sí (CA-006-02).
	ResolveSessionTenant(ctx context.Context, tokenHash string) (barbershopID string, found bool, err error)

	// ValidateAndRenewSession reconfirma la sesión DENTRO de la
	// transacción tenant-aware -la resolución previa no sustituye esta
	// autorización final- y, si sigue vigente, extiende
	// last_used_at/expires_at en la MISMA operación atómica (DEC-050,
	// renovación deslizante de 30 días desde el último uso). now y
	// newExpiresAt llegan ya calculados por el reloj inyectado del
	// servicio, nunca de now() de PostgreSQL, para que las pruebas puedan
	// congelar el tiempo. found=false cubre vencida, revocada, o de un
	// usuario que dejó de estar activo entre la resolución previa y esta
	// comprobación; una sesión vencida nunca se renueva ni se reabre
	// (CA-006-03).
	ValidateAndRenewSession(ctx context.Context, barbershopID, tokenHash string, now, newExpiresAt time.Time) (Principal, bool, error)

	// RevokeSession fija revoked_at = now en la sesión sessionID, solo
	// dentro de barbershopID y solo si no estaba revocada ya. El efecto
	// observable es idempotente (CA-006-02): revocar una sesión ya
	// revocada, o una que ya no existe dentro de ese tenant, no es un
	// error. barbershopID y sessionID llegan siempre de un [Principal] ya
	// validado por el middleware, nunca de un parámetro que el cliente
	// pueda controlar (CA-006-07).
	RevokeSession(ctx context.Context, barbershopID, sessionID string, now time.Time) error

	// BarbershopName devuelve barbershop.name de la barbería activa del
	// principal ya autenticado, aislado por tenant vía RLS
	// (barbershop_select_tenant_policy, DEC-024). Usada exclusivamente por
	// [SessionService.Context] (HU-012, DEC-060); ninguna otra operación
	// necesita este dato hoy.
	BarbershopName(ctx context.Context, barbershopID string) (string, error)
}

// SessionService implementa el caso de uso de HU-006: validar una sesión
// existente con renovación deslizante, y cerrarla. No conoce Chi, net/http,
// JSON ni PostgreSQL (CA-002-06): recibe sus dependencias por constructor,
// igual que [LoginService].
type SessionService struct {
	repo  SessionRepository
	clock clock.Clock
}

// NewSessionService construye el servicio.
func NewSessionService(repo SessionRepository, clk clock.Clock) *SessionService {
	return &SessionService{repo: repo, clock: clk}
}

// Validate valida rawToken (el valor EN CLARO leído de la cookie) y, si
// corresponde a una sesión vigente de un usuario activo, la renueva
// (last_used_at/expires_at) y devuelve el [Principal] resuelto. Devuelve
// exactamente el mismo error ([ErrInvalidSession]) para token desconocido,
// vencido, revocado o de un usuario inactivo: ninguna rama de este método
// puede producir un resultado observable distinto entre esos casos
// (CA-006-02, CA-006-03).
func (s *SessionService) Validate(ctx context.Context, rawToken string) (Principal, error) {
	if err := ctx.Err(); err != nil {
		return Principal{}, apperr.Internal(fmt.Errorf("auth: contexto cancelado antes de validar sesión: %w", err))
	}

	tokenHash := HashToken(rawToken)

	barbershopID, found, err := s.repo.ResolveSessionTenant(ctx, tokenHash)
	if err != nil {
		return Principal{}, apperr.Internal(fmt.Errorf("auth: resolver tenant de sesión: %w", err))
	}
	if !found {
		return Principal{}, ErrInvalidSession()
	}

	now := s.clock.Now()
	newExpiresAt := now.Add(SessionDuration)

	principal, valid, err := s.repo.ValidateAndRenewSession(ctx, barbershopID, tokenHash, now, newExpiresAt)
	if err != nil {
		return Principal{}, apperr.Internal(fmt.Errorf("auth: renovar sesión: %w", err))
	}
	if !valid {
		return Principal{}, ErrInvalidSession()
	}

	return principal, nil
}

// SessionContext es el resultado de la operación de contexto de sesión de
// HU-012 (DEC-060): el payload mínimo que GET /private/auth/session
// devuelve. Nunca incluye StaffUserID, correo ni nombre del barbero
// (RN-DAT-02, mismo patrón que [Principal]).
type SessionContext struct {
	BarbershopID   string
	BarbershopName string
	ExpiresAt      time.Time
}

// Context arma el contexto de sesión mínimo que HU-012 rehidrata al abrir o
// recargar la aplicación (DEC-060): el nombre de la barbería activa y la
// expiración ya renovada por [SessionMiddleware.RequireSession] para esta
// misma solicitud. principal llega siempre ya validado por el middleware;
// este método no vuelve a validar la sesión ni acepta un identificador que
// el cliente pueda controlar.
func (s *SessionService) Context(ctx context.Context, principal Principal) (SessionContext, error) {
	if err := ctx.Err(); err != nil {
		return SessionContext{}, apperr.Internal(fmt.Errorf("auth: contexto cancelado antes de leer contexto de sesión: %w", err))
	}

	name, err := s.repo.BarbershopName(ctx, principal.BarbershopID)
	if err != nil {
		return SessionContext{}, apperr.Internal(fmt.Errorf("auth: leer nombre de barbería: %w", err))
	}

	return SessionContext{
		BarbershopID:   principal.BarbershopID,
		BarbershopName: name,
		ExpiresAt:      principal.ExpiresAt,
	}, nil
}

// Logout revoca la sesión de principal (nunca un identificador que llegue
// del cliente: principal ya fue resuelto y validado por el middleware para
// ESTA solicitud). Idempotente en su efecto observable: [SessionMiddleware]
// ya rechaza con el mismo 401 uniforme un segundo intento con el mismo
// material, así que este método ni siquiera se vuelve a alcanzar; a nivel
// de repositorio, revocar una sesión ya revocada tampoco es un error
// (CA-006-02).
func (s *SessionService) Logout(ctx context.Context, principal Principal) error {
	if err := ctx.Err(); err != nil {
		return apperr.Internal(fmt.Errorf("auth: contexto cancelado antes de cerrar sesión: %w", err))
	}

	now := s.clock.Now()
	if err := s.repo.RevokeSession(ctx, principal.BarbershopID, principal.SessionID, now); err != nil {
		return apperr.Internal(fmt.Errorf("auth: revocar sesión: %w", err))
	}
	return nil
}
