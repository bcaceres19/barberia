package shops

import (
	"context"
	"fmt"
	"regexp"
	"unicode/utf8"

	"system-barbershop/internal/platform/apperr"
)

// contactEmailShapePattern refleja la misma forma mínima que
// barbershop_contact_email_ck exige en la base (al menos un carácter antes
// y después de '@', un '.' con al menos un carácter después): sin
// espacios, con '@' y con un '.' posterior. No es una validación RFC 5322
// completa a propósito (docs/06-api/estandar-openapi.md: "no crear reglas
// más restrictivas que las fuentes"); la normalización a minúsculas ocurre
// antes, en [NormalizeContactEmail].
var contactEmailShapePattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// contactPhoneE164Pattern es EXACTAMENTE la misma expresión que
// barbershop_contact_phone_ck y staff_user_phone_ck ya aplican en la base
// (20260817180000_create_login_throttle_and_phone_challenge.sql): '+',
// indicativo sin cero inicial, 8 a 15 dígitos en total. Validar aquí evita
// que un valor mal formado llegue a producir una violación de CHECK sin
// traducir (500) en vez de un 422 de campo.
var contactPhoneE164Pattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

// contactEmailMaxLength coincide con barbershop_contact_email_ck
// (char_length <= 254), mismo límite que staff_user.email.
const contactEmailMaxLength = 254

// Service implementa el caso de uso de HU-020: leer y actualizar la
// configuración básica de la barbería activa.
type Service struct {
	repo Repository
}

// NewService construye el servicio.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Get lee la configuración de la barbería barbershopID. barbershopID llega
// siempre de un auth.Principal ya autenticado (CA-020-01); este método no
// acepta ni valida ningún identificador que el cliente pueda controlar.
func (s *Service) Get(ctx context.Context, barbershopID string) (Barbershop, error) {
	if err := ctx.Err(); err != nil {
		return Barbershop{}, apperr.Internal(fmt.Errorf("shops: contexto cancelado antes de leer barbería: %w", err))
	}

	barbershop, found, err := s.repo.Get(ctx, barbershopID)
	if err != nil {
		return Barbershop{}, apperr.Internal(fmt.Errorf("shops: leer barbería: %w", err))
	}
	if !found {
		return Barbershop{}, apperr.NotFound("barbería no encontrada")
	}
	return barbershop, nil
}

// UpdateInput es la entrada cruda del caso de uso de actualización, tal
// como llega del handler HTTP: sin normalizar ni validar todavía.
type Input struct {
	Name         string
	Timezone     string
	ContactEmail string
	ContactPhone string
}

// Update normaliza y valida input y, si pasa todas las comprobaciones de
// campo, delega en el repositorio la confirmación final de la zona IANA y
// la escritura atómica (CA-020-02, CA-020-03, CA-020-06). barbershopID
// llega siempre de un auth.Principal ya autenticado, nunca de un campo del
// cuerpo (CA-020-05: el contrato ni siquiera declara un campo
// barbershopId). Los campos se evalúan en el orden fijo del contrato
// (name, timezone, contactEmail, contactPhone): el primer error de campo
// encontrado es el que se devuelve, sin ejecutar ninguna escritura.
func (s *Service) Update(ctx context.Context, barbershopID string, input Input) (Barbershop, error) {
	if err := ctx.Err(); err != nil {
		return Barbershop{}, apperr.Internal(fmt.Errorf("shops: contexto cancelado antes de actualizar barbería: %w", err))
	}

	name := NormalizeName(input.Name)
	switch {
	case name == "":
		return Barbershop{}, errNameRequired()
	case utf8.RuneCountInString(name) > NameMaxLength:
		return Barbershop{}, errNameTooLong()
	}

	timezone := NormalizeName(input.Timezone) // mismo recorte que el nombre; sin normalización de mayúsculas (un identificador IANA es sensible a caso).
	switch {
	case timezone == "":
		return Barbershop{}, errTimezoneRequired()
	case utf8.RuneCountInString(timezone) > TimezoneMaxLength:
		return Barbershop{}, errTimezoneTooLong()
	}

	contactEmail := NormalizeContactEmail(input.ContactEmail)
	if contactEmail != nil {
		switch {
		case utf8.RuneCountInString(*contactEmail) > contactEmailMaxLength:
			return Barbershop{}, errContactEmailInvalid()
		case !contactEmailShapePattern.MatchString(*contactEmail):
			return Barbershop{}, errContactEmailInvalid()
		}
	}

	contactPhone := NormalizeContact(input.ContactPhone)
	if contactPhone != nil && !contactPhoneE164Pattern.MatchString(*contactPhone) {
		return Barbershop{}, errContactPhoneInvalid()
	}

	result, err := s.repo.Update(ctx, barbershopID, UpdateInput{
		Name:         name,
		Timezone:     timezone,
		ContactEmail: contactEmail,
		ContactPhone: contactPhone,
	})
	if err != nil {
		return Barbershop{}, apperr.Internal(fmt.Errorf("shops: actualizar barbería: %w", err))
	}
	if !result.TimezoneValid {
		return Barbershop{}, errTimezoneInvalid()
	}
	if !result.Found {
		return Barbershop{}, apperr.NotFound("barbería no encontrada")
	}
	return result.Barbershop, nil
}
