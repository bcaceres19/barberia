package auth

import (
	"strings"
	"time"
)

// SessionDuration es la vigencia de una sesión desde su último uso (DEC-050:
// 30 días, renovación deslizante). La renovación en cada solicitud
// autenticada es responsabilidad de HU-006; HU-005 solo fija expires_at al
// emitir la sesión.
const SessionDuration = 30 * 24 * time.Hour

// Session es el resultado de un inicio de sesión exitoso. Token es el valor
// EN CLARO del token opaco: existe únicamente para que el adaptador HTTP lo
// escriba una sola vez en la cookie Set-Cookie. Nunca se registra, nunca se
// serializa a JSON, nunca se persiste tal cual (ver [HashToken]).
type Session struct {
	Token        string
	BarbershopID string
	StaffUserID  string
	IssuedAt     time.Time
	ExpiresAt    time.Time
}

// NormalizeEmail aplica la misma normalización que
// staff_user_email_ck y authn_resolve_login_tenant (recorte de espacios y
// minúsculas), como defensa en profundidad: la base de datos ya la aplica,
// pero normalizar antes evita una consulta con espacios sobrantes que nunca
// podría coincidir.
func NormalizeEmail(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}
