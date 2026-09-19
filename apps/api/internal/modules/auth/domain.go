package auth

import (
	"regexp"
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

// RecoveryChannel es el canal que la persona elige en el paso 1 de la
// recuperación de acceso (DEC-092). El valor que escribe solo identifica la
// cuenta: el código se entrega únicamente por este canal y al contacto
// verificado almacenado, nunca a un destino escrito por la persona.
type RecoveryChannel string

const (
	// RecoveryChannelEmail identifica la cuenta por su correo y entrega por correo.
	RecoveryChannelEmail RecoveryChannel = "email"
	// RecoveryChannelWhatsApp identifica la cuenta por su teléfono verificado
	// y entrega por WhatsApp oficial.
	RecoveryChannelWhatsApp RecoveryChannel = "whatsapp"
)

// Valid indica si c es un canal soportado.
func (c RecoveryChannel) Valid() bool {
	return c == RecoveryChannelEmail || c == RecoveryChannelWhatsApp
}

// RecoveryTarget agrupa el canal elegido y el valor escrito (correo o número
// de WhatsApp) que viajan sin cambios por los tres pasos de la recuperación
// (DEC-093, DP-SEG-15).
type RecoveryTarget struct {
	Channel RecoveryChannel
	Value   string
}

// phonePattern es el formato E.164 que exige staff_user_phone_ck.
var phonePattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

// phoneSeparators son los separadores de presentación que se descartan antes
// de validar: espacios, guiones, puntos y paréntesis.
var phoneSeparators = strings.NewReplacer(" ", "", "-", "", ".", "", "(", "", ")", "")

// NormalizePhone devuelve el número en E.164 (`+573001234567`) tras quitar
// separadores de presentación. ok=false cuando el resultado no es E.164: es
// un error de forma de la solicitud, no una señal sobre la existencia de una
// cuenta.
func NormalizePhone(raw string) (string, bool) {
	normalized := phoneSeparators.Replace(strings.TrimSpace(raw))
	if !phonePattern.MatchString(normalized) {
		return "", false
	}
	return normalized, true
}
