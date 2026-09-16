package booking

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/idempotency"
)

// tokenHashPattern refleja appointment_access_token_token_hash_ck:
// hexadecimal minúscula de 64 caracteres (SHA-256), mismo criterio de
// verificación de forma que el resto del dominio antes de tocar PostgreSQL.
var tokenHashPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// CreatePublicInput es la entrada ya autorizada de T1 pública (HU-097,
// prompt PROMPT-HU-097-v1 §"Trabajo requerido"): quien la construye
// (publicbooking.ConfirmationService) ya revalidó servicio/asignación,
// jornada, bloqueo, política de reserva y franja contra PostgreSQL real, y
// ya decidió la reconciliación de cliente de DEC-085. A diferencia de
// CreateInternalInput, no declara Actor: el actor de
// appointment_history SIEMPRE es el cliente (ActorTypeCustomer) y su
// CustomerID solo se conoce DESPUÉS de resolver/crear la fila `customer`
// dentro de la misma transacción (Repository.CreatePublic lo construye
// internamente), nunca antes.
type CreatePublicInput struct {
	BarberID     string
	ServiceID    string
	AttendeeName string
	StartsAt     time.Time
	EndsAt       time.Time
	Service      ServiceSnapshot
	CustomerNote *string

	// BarbershopName y Timezone viajan aquí, ya resueltos por
	// publicbooking, ÚNICAMENTE para que la respuesta almacenada de
	// idempotencia (publicAppointmentResponseWire) pueda construirse
	// dentro de la misma transacción: booking no lee `barbershop` (no es su
	// tabla) ni conoce la zona IANA por sí mismo (RN-DIS-07 es un concepto
	// de shops). Nunca se persisten ni se comparan: son datos de
	// presentación de paso.
	BarbershopName string
	Timezone       string

	// TokenPlain es el valor del token EN CLARO (DEC-089): viaja hasta aquí
	// solo para incluirse, una única vez, en la respuesta almacenada de
	// idempotencia (publicAppointmentResponseWire.AccessToken). Nunca se
	// persiste: PostgreSQL solo recibe TokenHash.
	TokenPlain string

	// Customer decide, igual que CreateInternalInput.Customer, si se
	// reutiliza un cliente existente (ExistingID) o se crea uno nuevo (New)
	// — la decisión de DEC-085 ya tomada por ReconcilePublicCustomer.
	Customer CustomerInput
	// CustomerUpdatePhone/CustomerUpdateEmail acompañan Customer.ExistingID
	// no nil cuando DEC-085 exige sobrescribir el campo que NO participó en
	// la coincidencia con el valor nuevo que el cliente dio (por ejemplo,
	// coincidió por teléfono pero trajo un correo distinto): nil cuando no
	// aplica ninguna sobrescritura. Ambos nil junto con Customer.New no nil
	// es el caso "sin coincidencia, cliente nuevo".
	CustomerUpdatePhone *string
	CustomerUpdateEmail *string

	// TokenHash es SHA-256(token) en hexadecimal minúscula (DEC-089): el
	// valor en claro NUNCA llega a este paquete ni a PostgreSQL, solo su
	// hash. TokenIssuedAt/TokenExpiresAt ya llegan calculados
	// (issuedAt + 90 días, DEC-089); TokenExpiresAt siempre es posterior a
	// TokenIssuedAt.
	TokenHash      string
	TokenIssuedAt  time.Time
	TokenExpiresAt time.Time
}

func (in CreatePublicInput) validate() error {
	if strings.TrimSpace(in.BarberID) == "" {
		return errBarberIDRequired()
	}
	if strings.TrimSpace(in.ServiceID) == "" {
		return errServiceIDRequired()
	}
	name := strings.TrimSpace(in.AttendeeName)
	if name == "" {
		return errAttendeeNameRequired()
	}
	if len(name) > AttendeeNameMaxLength {
		return errAttendeeNameTooLong()
	}
	if !in.EndsAt.After(in.StartsAt) {
		return errIntervalInvalid()
	}
	if int64(in.EndsAt.Sub(in.StartsAt).Seconds()) != int64(in.Service.DurationMinutes)*60 {
		return errIntervalDurationMismatch()
	}
	if err := in.Service.validate(); err != nil {
		return err
	}
	if in.CustomerNote != nil && len(*in.CustomerNote) > CustomerNoteMaxLength {
		return errCustomerNoteTooLong()
	}
	if err := in.Customer.validate(); err != nil {
		return err
	}
	if !tokenHashPattern.MatchString(in.TokenHash) {
		return errPublicTokenHashInvalid()
	}
	if !in.TokenExpiresAt.After(in.TokenIssuedAt) {
		return errPublicTokenExpiryInvalid()
	}
	if strings.TrimSpace(in.TokenPlain) == "" || strings.TrimSpace(in.BarbershopName) == "" || strings.TrimSpace(in.Timezone) == "" {
		return apperr.Internal(errors.New("booking: CreatePublicInput llegó con un dato de presentación vacío"))
	}
	return nil
}

func errPublicTokenHashInvalid() error {
	return apperr.Internal(errors.New("booking: token_hash con forma inválida antes de persistir"))
}

func errPublicTokenExpiryInvalid() error {
	return apperr.Internal(errors.New("booking: expiresAt del token debe ser posterior a issuedAt"))
}

// CreatePublicResult es el desenlace de un intento de alta pública
// idempotente (RN-IDE-01, DEC-043), mismo criterio que CreateManualResult:
// Decision.Outcome distingue Proceed/Replay/cualquier otro desenlace que
// Decision.AsError() ya traduce. Response es la única fuente que
// httpapi reenvía al cliente (mismo patrón que CreateManualResult): ya
// incluye el token de acceso EN CLARO, la única vez que existe fuera de
// esta transacción (DEC-089).
type CreatePublicResult struct {
	Decision idempotency.Decision
	Response idempotency.StoredResponse
}
