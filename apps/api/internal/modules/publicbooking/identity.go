package publicbooking

import (
	"regexp"
	"strings"

	"system-barbershop/internal/platform/apperr"
)

// Límites y formas iguales a booking.CustomerFullNameMaxLength/
// CustomerNoteMaxLength/AttendeeNameMaxLength y sus mismos patrones de
// teléfono/correo (customer_phone_ck/customer_email_ck). Se duplican aquí a
// propósito (CA-002-06, mismo criterio que serviceIDPattern frente a
// catalog.LooksLikeServiceID en domain.go): el núcleo de publicbooking no
// importa booking.
const (
	CustomerFullNameMaxLength = 120
	CustomerNoteMaxLength     = 500
	AttendeeNameMaxLength     = 120
)

var customerPhonePattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

// customerEmailShapePattern refleja customer_email_ck: al menos un
// carácter, '@', al menos un carácter, '.', al menos un carácter. La forma
// canónica (minúsculas, sin espacios) se valida aparte.
var customerEmailShapePattern = regexp.MustCompile(`^.+@.+\..+$`)

// CustomerIdentityInput son los datos que HU-096 captura del formulario
// público, todavía sin normalizar (HU-097 los persistirá, DP-PUB-05/CT-011
// resueltas). A diferencia de booking.NewCustomerInput (HU-061, teléfono y
// correo opcionales por RN-CIT-02), aquí los tres campos de contacto son
// obligatorios (CA-096-02): la reserva pública siempre necesita poder
// contactar al cliente.
type CustomerIdentityInput struct {
	FullName       string
	Phone          string
	Email          string
	Note           *string
	ForSomeoneElse bool
	AttendeeName   *string
}

// CustomerIdentity es la forma ya normalizada y validada: AttendeeName
// siempre queda resuelto (CA-096-01), nunca vacío.
type CustomerIdentity struct {
	FullName     string
	Phone        string
	Email        string
	Note         *string
	AttendeeName string
}

// NormalizeAndValidateCustomerIdentity aplica RN-RES-01 a RN-RES-03 y
// RN-DAT-01/RN-DAT-02: recorta y normaliza cada campo, exige nombre,
// teléfono y correo no vacíos con la forma aprobada, acota la nota y
// resuelve AttendeeName sin pedirlo dos veces cuando el cliente reserva
// para sí mismo (CA-096-01). No consulta la base de datos: la
// reconciliación de `customer` (DEC-085) es una decisión aparte
// (ReconcilePublicCustomer), porque depende de datos ya persistidos.
func NormalizeAndValidateCustomerIdentity(in CustomerIdentityInput) (CustomerIdentity, error) {
	fullName := strings.TrimSpace(in.FullName)
	if fullName == "" {
		return CustomerIdentity{}, errIdentityFullNameRequired()
	}
	if len(fullName) > CustomerFullNameMaxLength {
		return CustomerIdentity{}, errIdentityFullNameTooLong()
	}

	phone := strings.TrimSpace(in.Phone)
	if phone == "" {
		return CustomerIdentity{}, errIdentityPhoneRequired()
	}
	if !customerPhonePattern.MatchString(phone) {
		return CustomerIdentity{}, errIdentityPhoneInvalid()
	}

	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email == "" {
		return CustomerIdentity{}, errIdentityEmailRequired()
	}
	if strings.ContainsAny(email, " \t\n\r") || !customerEmailShapePattern.MatchString(email) || len(email) > 254 {
		return CustomerIdentity{}, errIdentityEmailInvalid()
	}

	var note *string
	if in.Note != nil {
		trimmed := strings.TrimSpace(*in.Note)
		if len(trimmed) > CustomerNoteMaxLength {
			return CustomerIdentity{}, errIdentityNoteTooLong()
		}
		if trimmed != "" {
			note = &trimmed
		}
	}

	// CA-096-01: para sí mismo, el nombre atendido se deriva del cliente sin
	// pedirlo dos veces. Para otra persona, un nombre no vacío adicional es
	// obligatorio -distinto del propio cliente en forma, aunque coincida por
	// coincidencia no se rechaza (RN-RES-03 no lo exige).
	attendeeName := fullName
	if in.ForSomeoneElse {
		if in.AttendeeName == nil {
			return CustomerIdentity{}, errIdentityAttendeeNameRequired()
		}
		trimmedAttendee := strings.TrimSpace(*in.AttendeeName)
		if trimmedAttendee == "" {
			return CustomerIdentity{}, errIdentityAttendeeNameRequired()
		}
		if len(trimmedAttendee) > AttendeeNameMaxLength {
			return CustomerIdentity{}, errIdentityAttendeeNameTooLong()
		}
		attendeeName = trimmedAttendee
	}

	return CustomerIdentity{
		FullName:     fullName,
		Phone:        phone,
		Email:        email,
		Note:         note,
		AttendeeName: attendeeName,
	}, nil
}

func errIdentityFullNameRequired() error {
	return apperr.Validation("el nombre del cliente es obligatorio")
}

func errIdentityFullNameTooLong() error {
	return apperr.Validation("el nombre del cliente excede el largo máximo")
}

func errIdentityPhoneRequired() error {
	return apperr.Validation("el teléfono del cliente es obligatorio")
}

func errIdentityPhoneInvalid() error {
	return apperr.Validation("el teléfono del cliente tiene un formato inválido")
}

func errIdentityEmailRequired() error {
	return apperr.Validation("el correo del cliente es obligatorio")
}

func errIdentityEmailInvalid() error {
	return apperr.Validation("el correo del cliente tiene un formato inválido")
}

func errIdentityNoteTooLong() error {
	return apperr.Validation("la nota excede el largo máximo")
}

func errIdentityAttendeeNameRequired() error {
	return apperr.Validation("el nombre de la persona atendida es obligatorio")
}

func errIdentityAttendeeNameTooLong() error {
	return apperr.Validation("el nombre de la persona atendida excede el largo máximo")
}

// CustomerReconciliation es la decisión de identidad de DEC-085 (resuelve
// DP-PUB-04), ya calculada: si ReuseCustomerID es nil, HU-097 crea un
// customer nuevo con los datos dados (sin coincidencia, o conflicto entre
// dos filas distintas). Si no es nil, HU-097 reutiliza esa fila y
// sobrescribe el campo que no participó en la coincidencia con el valor
// nuevo dado -sobrescribir con el mismo valor cuando también coincide es un
// no-op inofensivo, así que no hace falta distinguir ese caso aparte.
type CustomerReconciliation struct {
	ReuseCustomerID *string
	UpdatePhone     *string
	UpdateEmail     *string
}

// ReconcilePublicCustomer decide la reconciliación pública de `customer`
// según DEC-085 a partir de los identificadores ya encontrados (o nil) por
// teléfono y por correo dentro de la misma barbería (RN-TEN-01,
// CustomerRepository.FindCustomerMatches). phone/email son los valores ya
// normalizados que el cliente dio ahora.
func ReconcilePublicCustomer(phoneMatchID, emailMatchID *string, phone, email string) CustomerReconciliation {
	switch {
	case phoneMatchID == nil && emailMatchID == nil:
		return CustomerReconciliation{}
	case phoneMatchID != nil && emailMatchID != nil && *phoneMatchID != *emailMatchID:
		// Conflicto: cada campo coincide con una fila distinta. DEC-085 no
		// fusiona -se crea un customer nuevo con los datos dados en vez de
		// adivinar cuál identidad es la correcta.
		return CustomerReconciliation{}
	case phoneMatchID != nil:
		id := *phoneMatchID
		e := email
		return CustomerReconciliation{ReuseCustomerID: &id, UpdateEmail: &e}
	default:
		id := *emailMatchID
		p := phone
		return CustomerReconciliation{ReuseCustomerID: &id, UpdatePhone: &p}
	}
}
