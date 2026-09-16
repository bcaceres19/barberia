// Pruebas del caso de uso EN AISLAMIENTO, sin PostgreSQL real: cubren
// revalidación de recursos, membresía de la franja elegida contra el
// cálculo compartido de HU-094, la traducción a ScheduleConflictError con
// alternativas (RN-CON-05/DEC-090), la generación del token y el
// comportamiento no bloqueante del correo de confirmación (DEC-091). La
// persistencia atómica real (customer/appointment/appointment_history/
// appointment_access_token, la carrera de exclusión de PostgreSQL) se
// prueba contra PostgreSQL real en
// internal/modules/booking/postgres/public_repository_test.go.
package publicbooking_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"system-barbershop/internal/modules/publicbooking"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/idempotency"
)

type fakeCustomerRepository struct {
	phoneMatchID, emailMatchID *string
	err                        error
}

func (f *fakeCustomerRepository) FindCustomerMatches(_ context.Context, _, _, _ string) (*string, *string, error) {
	return f.phoneMatchID, f.emailMatchID, f.err
}

type capturedAppointmentCall struct {
	barbershopID, barbershopName, timezone         string
	barberID, serviceID                            string
	startsAt, endsAt                               time.Time
	attendeeName, serviceName                      string
	durationMinutes                                int
	priceAmountCents                               int64
	currency                                       string
	customerNote                                   *string
	customerExistingID                             *string
	customerUpdatePhone, customerUpdateEmail       *string
	customerFullName, customerPhone, customerEmail string
	tokenPlain, tokenHash                          string
	tokenIssuedAt, tokenExpiresAt                  time.Time
}

type fakePublicAppointmentPort struct {
	decision idempotency.Decision
	stored   idempotency.StoredResponse
	err      error
	captured *capturedAppointmentCall
}

func (f *fakePublicAppointmentPort) CreatePublicAppointment(
	_ context.Context,
	barbershopID, barbershopName, timezone string,
	barberID, serviceID string,
	startsAt, endsAt time.Time,
	attendeeName string,
	serviceName string,
	durationMinutes int,
	priceAmountCents int64,
	currency string,
	customerNote *string,
	customerExistingID *string,
	customerUpdatePhone *string,
	customerUpdateEmail *string,
	customerFullName, customerPhone, customerEmail string,
	tokenPlain, tokenHash string,
	tokenIssuedAt, tokenExpiresAt time.Time,
	_ idempotency.Key,
	_ idempotency.Fingerprint,
) (idempotency.Decision, idempotency.StoredResponse, error) {
	if f.captured != nil {
		*f.captured = capturedAppointmentCall{
			barbershopID: barbershopID, barbershopName: barbershopName, timezone: timezone,
			barberID: barberID, serviceID: serviceID,
			startsAt: startsAt, endsAt: endsAt,
			attendeeName: attendeeName, serviceName: serviceName,
			durationMinutes: durationMinutes, priceAmountCents: priceAmountCents, currency: currency,
			customerNote:        customerNote,
			customerExistingID:  customerExistingID,
			customerUpdatePhone: customerUpdatePhone, customerUpdateEmail: customerUpdateEmail,
			customerFullName: customerFullName, customerPhone: customerPhone, customerEmail: customerEmail,
			tokenPlain: tokenPlain, tokenHash: tokenHash,
			tokenIssuedAt: tokenIssuedAt, tokenExpiresAt: tokenExpiresAt,
		}
	}
	if f.err != nil {
		return idempotency.Decision{}, idempotency.StoredResponse{}, f.err
	}
	return f.decision, f.stored, nil
}

type confirmationEmailCall struct {
	email, barbershopName, serviceName, attendeeName, startsAtFormatted, accessLink string
}

type fakeConfirmationEmailPort struct {
	err   error
	calls []confirmationEmailCall
}

func (f *fakeConfirmationEmailPort) SendConfirmation(_ context.Context, email, barbershopName, serviceName, attendeeName, startsAtFormatted, accessLink string) error {
	f.calls = append(f.calls, confirmationEmailCall{email, barbershopName, serviceName, attendeeName, startsAtFormatted, accessLink})
	return f.err
}

const (
	confirmSlug      = "barberia-confirmacion"
	confirmServiceID = "33333333-3333-3333-3333-333333333333"
	confirmBarberID  = "44444444-4444-4444-4444-444444444444"
)

// newConfirmationFixture arma un ConfirmationService con todos los dobles
// razonables por defecto: barbería resuelta, servicio de 30 min asignado,
// jornada 9:00-18:00 hoy en America/Bogota, sin ocupaciones, sin
// coincidencia de cliente, puerto de persistencia que siempre Proceed.
func newConfirmationFixture(now time.Time) (
	*publicbooking.ConfirmationService,
	*fakeRepository,
	*fakeCustomerRepository,
	*fakePublicAppointmentPort,
	*fakeConfirmationEmailPort,
) {
	profiles := &fakeRepository{found: true, profile: publicbooking.BarbershopProfile{Name: "Barbería Confirmación", Timezone: "America/Bogota"}}
	availRepo, effectiveDay, blocks, policy, assignments, timezones := newAvailabilityFixture(now)
	availability := publicbooking.NewAvailabilityService(availRepo, effectiveDay, blocks, policy, assignments, timezones, fixedAvailabilityClock{now: now})
	customers := &fakeCustomerRepository{}
	appointments := &fakePublicAppointmentPort{decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed}, stored: idempotency.StoredResponse{Status: 201, ContentType: "application/json", Body: "{}"}}
	email := &fakeConfirmationEmailPort{}

	svc := publicbooking.NewConfirmationService(
		profiles, availRepo, availability, assignments, customers, appointments, email,
		fixedAvailabilityClock{now: now}, "https://reservas.ejemplo.test",
	)
	return svc, profiles, customers, appointments, email
}

func validConfirmInput(startsAt time.Time) publicbooking.ConfirmPublicAppointmentInput {
	return publicbooking.ConfirmPublicAppointmentInput{
		Slug:        confirmSlug,
		ServiceID:   confirmServiceID,
		BarberID:    confirmBarberID,
		StartsAtRaw: startsAt.Format(time.RFC3339),
		Identity: publicbooking.CustomerIdentityInput{
			FullName: "Carlos Restrepo",
			Phone:    "+573001234567",
			Email:    "carlos@example.com",
		},
	}
}

func TestConfirmAppointment_ValidSlotAndIdentity_ProceedsAndSendsEmail(t *testing.T) {
	loc, _ := time.LoadLocation("America/Bogota")
	now := time.Date(2026, time.September, 14, 8, 0, 0, 0, loc)
	startsAt := time.Date(2026, time.September, 14, 9, 0, 0, 0, loc)

	svc, _, _, _, email := newConfirmationFixture(now)
	result, emailErr, err := svc.ConfirmAppointment(context.Background(), validConfirmInput(startsAt), idempotency.Key("k-1"), idempotency.Fingerprint("f-1"))
	if err != nil {
		t.Fatalf("ConfirmAppointment: %v", err)
	}
	if emailErr != nil {
		t.Fatalf("emailErr = %v, want nil", emailErr)
	}
	if result.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("Outcome = %v, want OutcomeProceed", result.Decision.Outcome)
	}
	if len(email.calls) != 1 {
		t.Fatalf("expected exactly one confirmation email sent, got %d", len(email.calls))
	}
	if email.calls[0].accessLink == "" {
		t.Fatal("expected a non-empty access link when webBaseURL is configured")
	}
}

func TestConfirmAppointment_UnknownSlug_ReturnsUniformNotFound(t *testing.T) {
	loc, _ := time.LoadLocation("America/Bogota")
	now := time.Date(2026, time.September, 14, 8, 0, 0, 0, loc)
	startsAt := time.Date(2026, time.September, 14, 9, 0, 0, 0, loc)

	svc, profiles, _, _, _ := newConfirmationFixture(now)
	profiles.found = false
	in := validConfirmInput(startsAt)

	_, _, err := svc.ConfirmAppointment(context.Background(), in, idempotency.Key("k-1"), idempotency.Fingerprint("f-1"))
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindNotFound {
		t.Fatalf("expected apperr.KindNotFound, got %v", err)
	}
}

func TestConfirmAppointment_SlotNoLongerAvailable_ReturnsScheduleConflictWithAlternatives(t *testing.T) {
	loc, _ := time.LoadLocation("America/Bogota")
	now := time.Date(2026, time.September, 14, 8, 0, 0, 0, loc)
	// Fuera de la jornada 9:00-18:00 del fixture: nunca es un inicio válido.
	startsAt := time.Date(2026, time.September, 14, 20, 0, 0, 0, loc)

	svc, _, _, _, _ := newConfirmationFixture(now)
	_, _, err := svc.ConfirmAppointment(context.Background(), validConfirmInput(startsAt), idempotency.Key("k-1"), idempotency.Fingerprint("f-1"))

	var conflict *publicbooking.ScheduleConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("expected *ScheduleConflictError, got %v (%T)", err, err)
	}
	appErr, ok := apperr.As(conflict)
	if !ok || appErr.Kind != apperr.KindScheduleConflict {
		t.Fatalf("expected apperr.KindScheduleConflict via Unwrap, got %v", conflict)
	}
	if len(conflict.Alternatives) == 0 {
		t.Fatal("expected at least one alternative when the same day still has free slots")
	}
	if len(conflict.Alternatives) > publicbooking.MaxAlternatives {
		t.Fatalf("expected at most %d alternatives, got %d", publicbooking.MaxAlternatives, len(conflict.Alternatives))
	}
}

func TestConfirmAppointment_RepositoryReportsScheduleConflict_RecomputesFreshAlternatives(t *testing.T) {
	loc, _ := time.LoadLocation("America/Bogota")
	now := time.Date(2026, time.September, 14, 8, 0, 0, 0, loc)
	startsAt := time.Date(2026, time.September, 14, 9, 0, 0, 0, loc)

	svc, _, _, appointments, email := newConfirmationFixture(now)
	appointments.err = apperr.ScheduleConflict("la franja elegida ya no está disponible")

	_, _, err := svc.ConfirmAppointment(context.Background(), validConfirmInput(startsAt), idempotency.Key("k-1"), idempotency.Fingerprint("f-1"))

	var conflict *publicbooking.ScheduleConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("expected *ScheduleConflictError, got %v (%T)", err, err)
	}
	if len(email.calls) != 0 {
		t.Fatal("a lost race must never send a confirmation email")
	}
}

func TestConfirmAppointment_InvalidIdentity_RejectsWithoutPersisting(t *testing.T) {
	loc, _ := time.LoadLocation("America/Bogota")
	now := time.Date(2026, time.September, 14, 8, 0, 0, 0, loc)
	startsAt := time.Date(2026, time.September, 14, 9, 0, 0, 0, loc)

	svc, _, _, appointments, _ := newConfirmationFixture(now)
	in := validConfirmInput(startsAt)
	in.Identity.Email = "correo-invalido"

	_, _, err := svc.ConfirmAppointment(context.Background(), in, idempotency.Key("k-1"), idempotency.Fingerprint("f-1"))
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindValidation {
		t.Fatalf("expected apperr.KindValidation, got %v", err)
	}
	if appointments.captured != nil {
		t.Fatal("invalid identity must never reach the persistence port")
	}
}

func TestConfirmAppointment_ReplayOutcome_NeverResendsEmail(t *testing.T) {
	loc, _ := time.LoadLocation("America/Bogota")
	now := time.Date(2026, time.September, 14, 8, 0, 0, 0, loc)
	startsAt := time.Date(2026, time.September, 14, 9, 0, 0, 0, loc)

	svc, _, _, appointments, email := newConfirmationFixture(now)
	appointments.decision = idempotency.Decision{Outcome: idempotency.OutcomeReplay}

	_, emailErr, err := svc.ConfirmAppointment(context.Background(), validConfirmInput(startsAt), idempotency.Key("k-1"), idempotency.Fingerprint("f-1"))
	if err != nil {
		t.Fatalf("ConfirmAppointment: %v", err)
	}
	if emailErr != nil {
		t.Fatalf("emailErr = %v, want nil", emailErr)
	}
	if len(email.calls) != 0 {
		t.Fatalf("a replay must never resend the confirmation email, got %d calls", len(email.calls))
	}
}

func TestConfirmAppointment_EmailDeliveryFails_StillReturnsSuccessfulResult(t *testing.T) {
	loc, _ := time.LoadLocation("America/Bogota")
	now := time.Date(2026, time.September, 14, 8, 0, 0, 0, loc)
	startsAt := time.Date(2026, time.September, 14, 9, 0, 0, 0, loc)

	svc, _, _, _, email := newConfirmationFixture(now)
	email.err = errors.New("resend: 500")

	result, emailErr, err := svc.ConfirmAppointment(context.Background(), validConfirmInput(startsAt), idempotency.Key("k-1"), idempotency.Fingerprint("f-1"))
	if err != nil {
		t.Fatalf("ConfirmAppointment: %v", err)
	}
	if emailErr == nil {
		t.Fatal("expected emailErr to surface the delivery failure")
	}
	if result.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("a failed email must never turn a successful booking into an error result, got %+v", result)
	}
}

func TestConfirmAppointment_NoWebBaseURL_SendsEmailWithoutLink(t *testing.T) {
	loc, _ := time.LoadLocation("America/Bogota")
	now := time.Date(2026, time.September, 14, 8, 0, 0, 0, loc)
	startsAt := time.Date(2026, time.September, 14, 9, 0, 0, 0, loc)

	profiles := &fakeRepository{found: true, profile: publicbooking.BarbershopProfile{Name: "Barbería Confirmación", Timezone: "America/Bogota"}}
	availRepo, effectiveDay, blocks, policy, assignments, timezones := newAvailabilityFixture(now)
	availability := publicbooking.NewAvailabilityService(availRepo, effectiveDay, blocks, policy, assignments, timezones, fixedAvailabilityClock{now: now})
	appointments := &fakePublicAppointmentPort{decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed}, stored: idempotency.StoredResponse{Status: 201, ContentType: "application/json", Body: "{}"}}
	email := &fakeConfirmationEmailPort{}

	svc := publicbooking.NewConfirmationService(
		profiles, availRepo, availability, assignments, &fakeCustomerRepository{}, appointments, email,
		fixedAvailabilityClock{now: now}, "",
	)

	_, emailErr, err := svc.ConfirmAppointment(context.Background(), validConfirmInput(startsAt), idempotency.Key("k-1"), idempotency.Fingerprint("f-1"))
	if err != nil {
		t.Fatalf("ConfirmAppointment: %v", err)
	}
	if emailErr != nil {
		t.Fatalf("emailErr = %v, want nil", emailErr)
	}
	if len(email.calls) != 1 || email.calls[0].accessLink != "" {
		t.Fatalf("expected exactly one email with an empty access link, got %+v", email.calls)
	}
}
