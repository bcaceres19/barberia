package auth

import (
	"context"
	"time"
)

// PasswordHasher deriva y verifica contraseñas con una función de derivación
// de clave resistente a fuerza bruta (CA-005-03). El núcleo no sabe qué
// algoritmo concreto se usa ni sus parámetros: eso vive en el adaptador
// ([Argon2Hasher]).
type PasswordHasher interface {
	// Hash deriva password con una sal aleatoria nueva en cada llamada y
	// devuelve el valor codificado (algoritmo, parámetros, sal y hash en un
	// único texto autocontenido, formato PHC). Dos llamadas con el mismo
	// password devuelven valores distintos.
	Hash(password string) (string, error)
	// Verify compara password contra encodedHash, en tiempo proporcional al
	// costo configurado, no al contenido. Un encodedHash con formato
	// inválido se trata como no coincidente, nunca como error: el llamador
	// no debe poder distinguir "credencial corrupta" de "contraseña
	// incorrecta" (CA-005-02).
	Verify(encodedHash, password string) bool
}

// TokenGenerator produce el token opaco de sesión (DEC-050) con entropía
// criptográfica suficiente, ya codificado en un alfabeto apto para cookie.
type TokenGenerator interface {
	New() (string, error)
}

// Credential es el material de autenticación de un staff_user, tal como lo
// devuelve el repositorio dentro de la transacción tenant-aware. Found=false
// significa "nada que verificar": el llamador debe verificar igualmente
// contra un valor ficticio con el mismo costo (CA-005-02), nunca saltarse la
// verificación.
type Credential struct {
	Found             bool
	StaffUserID       string
	PasswordHash      string
	PasswordAlgorithm string
}

// Repository es el puerto de persistencia del módulo auth. El núcleo no
// importa internal/platform/database ni pgx (CA-002-06): postgres/ traduce
// entre este contrato y database.DB.
type Repository interface {
	// ResolveLoginTenant llama a la función estrecha
	// authn_resolve_login_tenant ANTES de que exista contexto de tenant
	// (modelo-fisico-referencia.sql sección A.0). found=false cubre tanto un
	// correo inexistente como un usuario inactivo (CA-005-02, CA-005-07);
	// esta interfaz no tiene forma de distinguirlos.
	ResolveLoginTenant(ctx context.Context, email string) (barbershopID string, found bool, err error)

	// LookupCredential ejecuta dentro de una transacción tenant-aware (el
	// tenant real cuando resolved=true, o un tenant señuelo estable cuando
	// resolved=false) exactamente la misma forma de consulta en ambos
	// casos, para que un correo inexistente y una barbería resuelta con
	// contraseña incorrecta impliquen el mismo trabajo de base de datos
	// (CA-005-02). resolved debe ser el mismo valor que devolvió
	// ResolveLoginTenant para esta solicitud.
	LookupCredential(ctx context.Context, barbershopID string, resolved bool, email string) (Credential, error)

	// CreateSession persiste una sesión nueva para un usuario ya
	// autenticado, en su propia transacción tenant-aware. tokenHash es el
	// SHA-256 hexadecimal del token en claro (ver [HashToken]); el valor en
	// claro nunca llega a este puerto.
	CreateSession(ctx context.Context, barbershopID, staffUserID, tokenHash string, issuedAt, expiresAt time.Time) error
}
