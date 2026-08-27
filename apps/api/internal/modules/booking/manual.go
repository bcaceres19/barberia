package booking

import (
	"context"
	"strings"
	"time"

	"system-barbershop/internal/platform/idempotency"
)

// CreateManualAppointmentRequest es la entrada cruda (ya decodificada del
// JSON, todavía sin normalizar) del caso de uso de HU-061: el barbero
// autenticado registra un turno manual recibido por teléfono, WhatsApp o en
// persona. StartsAtLocal es el instante civil "AAAA-MM-DDTHH:MM:SS" (sin
// desplazamiento de zona) que el barbero eligió en su propio reloj: se
// interpreta contra la zona IANA vigente de la barbería (RN-DIS-07), nunca
// contra la zona del servidor o del dispositivo.
type CreateManualAppointmentRequest struct {
	BarberID         string
	ServiceID        string
	AttendeeName     string
	CustomerFullName string
	CustomerPhone    *string
	CustomerEmail    *string
	CustomerNote     *string
	StartsAtLocal    string
	ActorStaffUserID string
}

// civilLocalLayout es el único formato que StartsAtLocal admite: fecha y
// hora civiles completas, sin zona (RFC 3339 sin el sufijo de
// desplazamiento). Un cliente que envíe un sufijo de zona ("Z", "+05:00")
// se rechaza: esta operación exige el reloj de pared que el barbero vio,
// no un instante que el cliente ya convirtió.
const civilLocalLayout = "2006-01-02T15:04:05"

// ManualBookingService implementa el caso de uso completo de HU-061:
// resuelve la zona de la barbería, interpreta el instante civil, verifica
// DEC-072 (servicio activo y asignado) y DEC-073 (sin bloqueo vigente),
// decide la reconciliación de cliente de DEC-071 y delega en la primitiva
// idempotente de persistencia. Colabora con catalog/schedule/shops
// exclusivamente mediante los puertos pequeños declarados en ports.go,
// nunca importando esos módulos (mismo criterio que catalog frente a
// staff, HU-023).
type ManualBookingService struct {
	repo      Repository
	catalog   BarberServicePort
	blocks    BlockCheckPort
	timezones TimezonePort
}

// NewManualBookingService construye el caso de uso a partir de sus cuatro
// colaboradores.
func NewManualBookingService(repo Repository, catalog BarberServicePort, blocks BlockCheckPort, timezones TimezonePort) *ManualBookingService {
	return &ManualBookingService{repo: repo, catalog: catalog, blocks: blocks, timezones: timezones}
}

// CreateManualAppointment ejecuta el caso de uso completo (§"Trabajo
// requerido, punto 2" del prompt de HU-061). barbershopID ya llegó
// resuelto por el llamador (sesión autenticada); key/fingerprint ya fueron
// interpretados por la capa HTTP.
func (s *ManualBookingService) CreateManualAppointment(
	ctx context.Context,
	barbershopID string,
	req CreateManualAppointmentRequest,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (CreateManualResult, error) {
	barberID := strings.TrimSpace(req.BarberID)
	if barberID == "" {
		return CreateManualResult{}, errBarberIDRequired()
	}
	serviceID := strings.TrimSpace(req.ServiceID)
	if serviceID == "" {
		return CreateManualResult{}, errServiceIDRequired()
	}
	attendeeName := strings.TrimSpace(req.AttendeeName)
	if attendeeName == "" {
		return CreateManualResult{}, errAttendeeNameRequired()
	}
	if len(attendeeName) > AttendeeNameMaxLength {
		return CreateManualResult{}, errAttendeeNameTooLong()
	}

	customer, err := NewCustomerInput{
		FullName: req.CustomerFullName,
		Phone:    normalizeOptional(req.CustomerPhone),
		Email:    normalizeOptionalEmail(req.CustomerEmail),
	}.normalizedOrError()
	if err != nil {
		return CreateManualResult{}, err
	}

	note := normalizeOptional(req.CustomerNote)
	if note != nil && len(*note) > CustomerNoteMaxLength {
		return CreateManualResult{}, errCustomerNoteTooLong()
	}

	timezone, err := s.timezones.Timezone(ctx, barbershopID)
	if err != nil {
		return CreateManualResult{}, err
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return CreateManualResult{}, errTimezoneUnresolvable()
	}

	startsAt, err := parseCivilLocal(req.StartsAtLocal, loc)
	if err != nil {
		return CreateManualResult{}, err
	}

	name, durationMinutes, priceCents, currency, found, err := s.catalog.ActiveAssignedService(ctx, barbershopID, barberID, serviceID)
	if err != nil {
		return CreateManualResult{}, err
	}
	if !found {
		return CreateManualResult{}, errServiceNotAssigned()
	}

	endsAt := startsAt.Add(time.Duration(durationMinutes) * time.Minute)

	blocked, err := s.blocks.HasActiveBlock(ctx, barbershopID, barberID, startsAt, endsAt, timezone)
	if err != nil {
		return CreateManualResult{}, err
	}
	if blocked {
		return CreateManualResult{}, errBlockedInterval()
	}

	customerInput, err := s.resolveCustomer(ctx, barbershopID, customer)
	if err != nil {
		return CreateManualResult{}, err
	}

	if req.ActorStaffUserID == "" {
		return CreateManualResult{}, errActorInvalid()
	}
	actorID := req.ActorStaffUserID

	input := CreateInternalInput{
		BarberID:     barberID,
		ServiceID:    serviceID,
		AttendeeName: attendeeName,
		StartsAt:     startsAt,
		EndsAt:       endsAt,
		Origin:       OriginManual,
		Service: ServiceSnapshot{
			Name:             name,
			DurationMinutes:  durationMinutes,
			PriceAmountCents: priceCents,
			Currency:         currency,
		},
		CustomerNote: note,
		Customer:     customerInput,
		Actor:        Actor{Type: ActorTypeStaff, StaffUserID: &actorID},
	}
	if err := input.validate(); err != nil {
		return CreateManualResult{}, err
	}

	return s.repo.CreateManual(ctx, barbershopID, input, key, fingerprint)
}

// resolveCustomer aplica DEC-071: si el cliente dio teléfono, busca/reutiliza
// por teléfono (DEC-045); si no, pero dio correo, busca/reutiliza por
// correo (DEC-046); si faltan ambos, siempre construye un cliente nuevo, sin
// reconciliar por nombre.
func (s *ManualBookingService) resolveCustomer(ctx context.Context, barbershopID string, c NewCustomerInput) (CustomerInput, error) {
	if c.Phone != nil || c.Email != nil {
		existing, found, err := s.repo.FindCustomerForReconciliation(ctx, barbershopID, c.Phone, c.Email)
		if err != nil {
			return CustomerInput{}, err
		}
		if found {
			id := existing.ID
			return CustomerInput{ExistingID: &id}, nil
		}
	}
	input := c
	return CustomerInput{New: &input}, nil
}

func normalizeOptional(raw *string) *string {
	if raw == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeOptionalEmail(raw *string) *string {
	v := normalizeOptional(raw)
	if v == nil {
		return nil
	}
	lower := strings.ToLower(*v)
	return &lower
}

// normalizedOrError valida NewCustomerInput ya normalizado, reutilizando
// exactamente las mismas reglas de forma que CreateInternalInput.validate
// aplicará después (mismo criterio de una sola fuente de verdad).
func (c NewCustomerInput) normalizedOrError() (NewCustomerInput, error) {
	c.FullName = strings.TrimSpace(c.FullName)
	if err := c.validate(); err != nil {
		return NewCustomerInput{}, err
	}
	return c, nil
}

// parseCivilLocal interpreta raw como fecha/hora civil en loc. Un sufijo de
// zona explícito (Z, +05:00) o un formato que no coincida exactamente con
// civilLocalLayout se rechaza como entrada inválida.
func parseCivilLocal(raw string, loc *time.Location) (time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	t, err := time.ParseInLocation(civilLocalLayout, trimmed, loc)
	if err != nil {
		return time.Time{}, errStartsAtInvalid()
	}
	return t, nil
}
