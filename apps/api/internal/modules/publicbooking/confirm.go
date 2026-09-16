package publicbooking

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/clock"
	"system-barbershop/internal/platform/idempotency"
)

// MaxAlternatives acota las franjas alternativas que RN-CON-05/DEC-090
// devuelven cuando el cliente pierde la carrera de confirmación: "hasta 3",
// aplicado literalmente, sin reinterpretar.
const MaxAlternatives = 3

// accessTokenBytes es la entropía cruda del token de acceso al turno antes
// de codificar: 32 bytes (256 bits, DEC-089), mismo criterio que
// auth.CryptoTokenGenerator (token opaco de sesión).
const accessTokenBytes = 32

// accessTokenValidity es la vigencia fija del token de acceso desde su
// emisión (DEC-089, resuelve DP-PUB-05): 90 días, sin rotación.
const accessTokenValidity = 90 * 24 * time.Hour

// ConfirmPublicAppointmentInput es la entrada cruda (ya decodificada del
// JSON, todavía sin normalizar) del caso de uso de HU-097: slug/serviceId/
// barberId llegan de la ruta -no confiados, revalidados de nuevo aquí
// exactamente como HU-094/HU-096 ya exigían-, startsAtRaw es el instante
// absoluto ISO 8601 que el cliente eligió de entre los `slots` de HU-094
// (nunca una hora civil: a diferencia de la creación manual, el público
// siempre parte de un inicio ya resuelto), e Identity son los datos de
// HU-096 sin normalizar todavía (CustomerIdentityInput, identity.go).
type ConfirmPublicAppointmentInput struct {
	Slug        string
	ServiceID   string
	BarberID    string
	StartsAtRaw string
	Identity    CustomerIdentityInput
}

// ConfirmResult es el desenlace de un intento de confirmación pública
// idempotente (RN-IDE-01, DEC-043), mismo criterio que
// booking.CreateManualResult: Decision.Outcome distingue Proceed/Replay/
// cualquier otro desenlace que Decision.AsError() ya traduce. Response ya
// trae, byte a byte, el cuerpo público completo -incluido el token de
// acceso en claro (DEC-089)- que httpapi reenvía sin volver a construirlo.
type ConfirmResult struct {
	Decision idempotency.Decision
	Response idempotency.StoredResponse
}

// AlternativeSlot es una franja alternativa ya calculada (RN-CON-05,
// DEC-090): el mismo instante absoluto que AvailabilitySlot, expuesto como
// tipo propio para que ScheduleConflictError no dependa de un tipo pensado
// para HU-094.
type AlternativeSlot struct {
	StartsAt time.Time
}

// ScheduleConflictError es el desenlace de RN-CON-05/DEC-090: la franja
// elegida ya no está disponible (perdió la carrera de exclusión de
// PostgreSQL, o la revalidación de servidor la descartó antes de intentar
// persistir, CA-097-02) y el servidor ya calculó hasta MaxAlternatives
// franjas cronológicamente cercanas con el mismo barbero y servicio. Cause
// es el *apperr.Error (Kind apperr.KindScheduleConflict) que
// httpserver.Translate ya sabe traducir a la parte segura y común del
// Problem 409; httpapi extiende esa respuesta con Alternatives (RFC 9457,
// extensión abierta) al escribirla.
type ScheduleConflictError struct {
	Cause        error
	Alternatives []AlternativeSlot
}

func (e *ScheduleConflictError) Error() string { return e.Cause.Error() }
func (e *ScheduleConflictError) Unwrap() error { return e.Cause }

func errScheduleConflict() error {
	return apperr.ScheduleConflict("la franja elegida ya no está disponible")
}

// errAppointmentResourceNotAvailable cubre CA-097-02: barbería no publicable
// (mismo criterio que errBarbershopNotPublic), o serviceId/barberId sin
// forma de UUID, ajeno, inexistente, inactivo o sin asignación vigente. Se
// unifica en un único 404 uniforme, sin distinguir la causa, exactamente
// como CA-090-02/CA-092-03: a diferencia de la disponibilidad de solo
// lectura (donde esa misma causa produce una lista vacía), una operación de
// escritura no tiene un "vacío" que devolver, así que el 404 uniforme es la
// respuesta equivalente.
func errAppointmentResourceNotAvailable() error {
	return apperr.NotFound("no existe un servicio activo asignado a ese barbero en esa barbería pública")
}

func errStartsAtInvalid() error {
	return apperr.Invalid("startsAt debe ser un instante absoluto ISO 8601 válido")
}

// ConfirmationService implementa T1 pública (HU-097, prompt
// PROMPT-HU-097-v1): revalida todo contra PostgreSQL real -barbería,
// servicio/asignación, jornada, bloqueo, política de reserva y franja-,
// decide la reconciliación de cliente de DEC-085, emite el token de acceso
// de DEC-089 y persiste todo en una sola transacción atómica a través de
// PublicAppointmentPort. Colabora con booking exclusivamente mediante ese
// puerto pequeño (CA-002-06): nunca importa booking.
type ConfirmationService struct {
	profiles         Repository
	availabilityRepo AvailabilityRepository
	availability     *AvailabilityService
	assignments      ServiceAssignmentPort
	customers        CustomerRepository
	appointments     PublicAppointmentPort
	email            ConfirmationEmailPort
	clock            clock.Clock
	webBaseURL       string
}

// NewConfirmationService construye el caso de uso. profiles resuelve el
// perfil público por slug (mismo Repository que Service, HU-090: necesario
// aquí solo para BarbershopProfile.Name, el resumen/correo de confirmación
// lo necesita); availabilityRepo resuelve el identificador interno de la
// barbería (mismo AvailabilityRepository que AvailabilityService, HU-094).
// webBaseURL es el origen del frontend público (config.Config.
// PublicWebBaseURL, DEC-089/DEC-091); una cadena vacía es válida (el
// correo se envía sin enlace clicable, nunca bloquea la confirmación).
func NewConfirmationService(
	profiles Repository,
	availabilityRepo AvailabilityRepository,
	availability *AvailabilityService,
	assignments ServiceAssignmentPort,
	customers CustomerRepository,
	appointments PublicAppointmentPort,
	email ConfirmationEmailPort,
	clk clock.Clock,
	webBaseURL string,
) *ConfirmationService {
	return &ConfirmationService{
		profiles:         profiles,
		availabilityRepo: availabilityRepo,
		availability:     availability,
		assignments:      assignments,
		customers:        customers,
		appointments:     appointments,
		email:            email,
		clock:            clk,
		webBaseURL:       strings.TrimRight(webBaseURL, "/"),
	}
}

// ConfirmAppointment ejecuta el caso de uso completo. Un error de envío del
// correo de confirmación (DEC-091) NUNCA revierte ni oculta un
// ConfirmResult ya exitoso: se devuelve junto al resultado para que
// httpapi lo registre sin destinatario, código ni token (RN-DAT-02), mismo
// criterio que auth.RecoveryService.Request frente a RecoveryCodeSender.
func (s *ConfirmationService) ConfirmAppointment(
	ctx context.Context,
	in ConfirmPublicAppointmentInput,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (result ConfirmResult, emailErr error, err error) {
	if err := ctx.Err(); err != nil {
		return ConfirmResult{}, nil, apperr.Internal(fmt.Errorf("publicbooking: contexto cancelado antes de confirmar: %w", err))
	}

	slug := strings.TrimSpace(in.Slug)
	if slug == "" || len(slug) > MaxSlugLength {
		return ConfirmResult{}, nil, errBarbershopNotPublic()
	}
	if !LooksLikePublicServiceID(in.ServiceID) || !LooksLikePublicBarberID(in.BarberID) {
		return ConfirmResult{}, nil, errAppointmentResourceNotAvailable()
	}

	profile, found, err := s.profiles.ResolveBySlug(ctx, slug)
	if err != nil {
		return ConfirmResult{}, nil, apperr.Internal(fmt.Errorf("publicbooking: resolver perfil de barbería para confirmar: %w", err))
	}
	if !found {
		return ConfirmResult{}, nil, errBarbershopNotPublic()
	}

	barbershopID, found, err := s.availabilityRepo.ResolveBarbershopID(ctx, slug)
	if err != nil {
		return ConfirmResult{}, nil, apperr.Internal(fmt.Errorf("publicbooking: resolver barbería para confirmar: %w", err))
	}
	if !found {
		return ConfirmResult{}, nil, errBarbershopNotPublic()
	}

	serviceName, durationMinutes, priceCents, currency, assigned, err := s.assignments.ActiveAssignedService(ctx, barbershopID, in.BarberID, in.ServiceID)
	if err != nil {
		return ConfirmResult{}, nil, apperr.Internal(fmt.Errorf("publicbooking: verificar asignación para confirmar: %w", err))
	}
	if !assigned || durationMinutes <= 0 {
		return ConfirmResult{}, nil, errAppointmentResourceNotAvailable()
	}

	startsAt, err := time.Parse(time.RFC3339, strings.TrimSpace(in.StartsAtRaw))
	if err != nil {
		return ConfirmResult{}, nil, errStartsAtInvalid()
	}
	startsAt = startsAt.UTC()
	endsAt := startsAt.Add(time.Duration(durationMinutes) * time.Minute)

	identity, err := NormalizeAndValidateCustomerIdentity(in.Identity)
	if err != nil {
		return ConfirmResult{}, nil, err
	}

	starts, timezone, _, err := s.availability.computeAvailableStarts(ctx, barbershopID, in.BarberID, durationMinutes)
	if err != nil {
		return ConfirmResult{}, nil, err
	}
	loc, locErr := time.LoadLocation(timezone)
	if locErr != nil {
		return ConfirmResult{}, nil, apperr.Internal(fmt.Errorf("publicbooking: zona horaria irresoluble para confirmar: %w", locErr))
	}

	if !containsInstant(starts, startsAt) {
		return ConfirmResult{}, nil, &ScheduleConflictError{
			Cause:        errScheduleConflict(),
			Alternatives: computeAlternatives(starts, startsAt, loc),
		}
	}

	phoneMatchID, emailMatchID, err := s.customers.FindCustomerMatches(ctx, barbershopID, identity.Phone, identity.Email)
	if err != nil {
		return ConfirmResult{}, nil, apperr.Internal(fmt.Errorf("publicbooking: buscar coincidencias de cliente para confirmar: %w", err))
	}
	reconciliation := ReconcilePublicCustomer(phoneMatchID, emailMatchID, identity.Phone, identity.Email)

	tokenPlain, tokenHash, err := generateAccessToken()
	if err != nil {
		return ConfirmResult{}, nil, apperr.Internal(fmt.Errorf("publicbooking: generar token de acceso: %w", err))
	}
	issuedAt := s.clock.Now()
	expiresAt := issuedAt.Add(accessTokenValidity)

	var customerNote *string
	if identity.Note != nil {
		note := *identity.Note
		customerNote = &note
	}

	decision, stored, err := s.appointments.CreatePublicAppointment(
		ctx,
		barbershopID, profile.Name, timezone,
		in.BarberID, in.ServiceID,
		startsAt, endsAt,
		identity.AttendeeName,
		serviceName, durationMinutes, priceCents, currency,
		customerNote,
		reconciliation.ReuseCustomerID,
		reconciliation.UpdatePhone,
		reconciliation.UpdateEmail,
		identity.FullName, identity.Phone, identity.Email,
		tokenPlain, tokenHash,
		issuedAt, expiresAt,
		key, fingerprint,
	)
	if err != nil {
		if appErr, ok := apperr.As(err); ok && appErr.Kind == apperr.KindScheduleConflict {
			// La revalidación de arriba pasó, pero la restricción de
			// exclusión de PostgreSQL (RN-CON-03) rechazó la franja: otra
			// confirmación ganó la carrera en el instante exacto entre esa
			// lectura y este INSERT. Recalcula alternativas con una lectura
			// fresca -la anterior ya está obsoleta- antes de responder
			// (RN-CON-05, DEC-090).
			freshStarts, freshTimezone, _, freshErr := s.availability.computeAvailableStarts(ctx, barbershopID, in.BarberID, durationMinutes)
			if freshErr != nil {
				return ConfirmResult{}, nil, freshErr
			}
			freshLoc, freshLocErr := time.LoadLocation(freshTimezone)
			if freshLocErr != nil {
				return ConfirmResult{}, nil, apperr.Internal(fmt.Errorf("publicbooking: zona horaria irresoluble tras conflicto: %w", freshLocErr))
			}
			return ConfirmResult{}, nil, &ScheduleConflictError{
				Cause:        err,
				Alternatives: computeAlternatives(freshStarts, startsAt, freshLoc),
			}
		}
		return ConfirmResult{}, nil, err
	}

	result = ConfirmResult{Decision: decision, Response: stored}

	if decision.Outcome == idempotency.OutcomeProceed {
		emailErr = s.sendConfirmationEmail(ctx, identity, profile, serviceName, startsAt, loc, tokenPlain)
	}
	return result, emailErr, nil
}

func containsInstant(starts []time.Time, target time.Time) bool {
	for _, t := range starts {
		if t.Equal(target) {
			return true
		}
	}
	return false
}

// computeAlternatives aplica DEC-090 literalmente: hasta MaxAlternatives
// franjas del MISMO día civil (en loc) más cercanas cronológicamente a
// target (antes o después); si ese día ya no tiene ninguna, salta al primer
// inicio del siguiente día civil con franjas.
func computeAlternatives(starts []time.Time, target time.Time, loc *time.Location) []AlternativeSlot {
	targetDay := target.In(loc).Format("2006-01-02")

	var sameDay []time.Time
	for _, t := range starts {
		if t.Equal(target) {
			continue
		}
		if t.In(loc).Format("2006-01-02") == targetDay {
			sameDay = append(sameDay, t)
		}
	}

	if len(sameDay) > 0 {
		sort.Slice(sameDay, func(i, j int) bool {
			di, dj := durationAbs(sameDay[i].Sub(target)), durationAbs(sameDay[j].Sub(target))
			if di != dj {
				return di < dj
			}
			return sameDay[i].Before(sameDay[j])
		})
		if len(sameDay) > MaxAlternatives {
			sameDay = sameDay[:MaxAlternatives]
		}
		alternatives := make([]AlternativeSlot, 0, len(sameDay))
		for _, t := range sameDay {
			alternatives = append(alternatives, AlternativeSlot{StartsAt: t})
		}
		return alternatives
	}

	var next time.Time
	found := false
	for _, t := range starts {
		day := t.In(loc).Format("2006-01-02")
		if day <= targetDay {
			continue
		}
		if !found || t.Before(next) {
			next = t
			found = true
		}
	}
	if !found {
		return nil
	}
	return []AlternativeSlot{{StartsAt: next}}
}

func durationAbs(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

// generateAccessToken produce el par (valor en claro, hash) del token de
// acceso al turno (DEC-089): 32 bytes de crypto/rand codificados en
// base64.RawURLEncoding (apto para viajar como segmento de URL, mismo
// criterio que auth.CryptoTokenGenerator) y SHA-256 hexadecimal minúsculo
// del valor en claro, exactamente la forma que
// appointment_access_token_token_hash_ck exige.
func generateAccessToken() (plain, hash string, err error) {
	buf := make([]byte, accessTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("generar entropía del token: %w", err)
	}
	plain = base64.RawURLEncoding.EncodeToString(buf)
	sum := sha256.Sum256([]byte(plain))
	hash = hex.EncodeToString(sum[:])
	return plain, hash, nil
}

// sendConfirmationEmail construye el enlace de acceso (cadena vacía cuando
// webBaseURL no está configurado, DEC-091) y delega en ConfirmationEmailPort.
// Nunca se llama para OutcomeReplay: el correo original ya se envió (o falló
// y quedó registrado) la primera vez; reenviarlo en cada repetición
// idempotente violaría "un único intento" (DEC-091).
func (s *ConfirmationService) sendConfirmationEmail(
	ctx context.Context,
	identity CustomerIdentity,
	profile BarbershopProfile,
	serviceName string,
	startsAt time.Time,
	loc *time.Location,
	tokenPlain string,
) error {
	var link string
	if s.webBaseURL != "" {
		link = s.webBaseURL + "/mi-turno/" + tokenPlain
	}
	formatted := startsAt.In(loc).Format("2006-01-02 15:04") + " (" + loc.String() + ")"
	return s.email.SendConfirmation(ctx, identity.Email, profile.Name, serviceName, identity.AttendeeName, formatted, link)
}
