package googlecalendar

import (
	"time"

	"system-barbershop/internal/platform/apperr"
)

// Status es el estado de la conexión de un barbero con Google Calendar
// (DEC-101). NotConnected no se persiste: es lo que se informa cuando el
// barbero nunca conectó.
type Status string

const (
	StatusNotConnected   Status = "not_connected"
	StatusConnected      Status = "connected"
	StatusReauthRequired Status = "reauth_required"
	StatusError          Status = "error"
	StatusDisconnected   Status = "disconnected"
)

// Límites de reminder_minutes: 0 a 40320 (cuatro semanas, el máximo de Google).
const (
	MinReminderMinutes = 0
	MaxReminderMinutes = 40320
)

// CalendarPrimary es el calendario principal de la cuenta conectada.
const CalendarPrimary = "primary"

// Connection es la conexión de un barbero. Las credenciales cifradas nunca
// salen del módulo: Credentials es el texto cifrado y KeyID la clave que lo
// produjo; ninguna capa HTTP los expone.
type Connection struct {
	ID              string
	BarbershopID    string
	BarberID        string
	Status          Status
	AccountEmail    string
	CalendarID      string
	ReminderMinutes *int
	ConnectedAt     *time.Time
	DisconnectedAt  *time.Time
	LastSyncedAt    *time.Time
	LastErrorCode   string
	Credentials     []byte
	KeyID           string
}

// HasCredentials informa si la conexión conserva un refresh token utilizable.
func (c Connection) HasCredentials() bool {
	return len(c.Credentials) > 0
}

// OAuthState es el estado de un solo uso de una autorización en curso: liga
// la vuelta de Google a la barbería, el barbero, el usuario y la sesión que
// la iniciaron. El valor en claro del `state` nunca se persiste (solo su hash).
type OAuthState struct {
	BarbershopID      string
	BarberID          string
	StaffUserID       string
	SessionID         string
	StateHash         string
	VerifierEncrypted []byte
	VerifierKeyID     string
	ExpiresAt         time.Time
}

// Tokens es lo que Google devuelve al canjear el código de autorización.
type Tokens struct {
	RefreshToken string
	AccountEmail string
}

// ValidateReminderMinutes comprueba la anticipación del recordatorio. nil es
// válido: significa «usar los recordatorios predeterminados de Google».
func ValidateReminderMinutes(minutes *int) error {
	if minutes == nil {
		return nil
	}
	if *minutes < MinReminderMinutes || *minutes > MaxReminderMinutes {
		return apperr.Validation("la anticipación del recordatorio debe estar entre 0 y 40320 minutos")
	}
	return nil
}

func errIntegrationDisabled() error {
	return apperr.InvalidState("la integración con Google Calendar no está disponible en este entorno")
}

func errNoLinkedBarber() error {
	return apperr.InvalidState("vincula tu usuario con un barbero antes de conectar Google Calendar")
}

func errNotConnected() error {
	return apperr.NotFound("no hay una conexión con Google Calendar")
}

// errAuthorizationInvalid cubre un state ausente, vencido, ya usado o ajeno:
// el mensaje es el mismo en todos los casos para no revelar cuál ocurrió.
func errAuthorizationInvalid() error {
	return apperr.Invalid("la autorización con Google no es válida o ya venció; inicia la conexión de nuevo")
}
