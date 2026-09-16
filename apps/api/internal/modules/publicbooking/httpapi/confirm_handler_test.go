// Pruebas del adaptador HTTP de HU-097 EN AISLAMIENTO: decodificación,
// cabecera Idempotency-Key obligatoria, forma cerrada del cuerpo,
// traducción de errores (incluida la extensión `alternatives` del 409 de
// RN-CON-05/DEC-090), reenvío byte a byte del cuerpo almacenado en éxito.
// La orquestación real del caso de uso se prueba en
// internal/modules/publicbooking/confirm_test.go; la persistencia atómica
// real en internal/modules/booking/postgres/public_repository_test.go.
package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"system-barbershop/internal/modules/publicbooking"
	"system-barbershop/internal/modules/publicbooking/httpapi"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
	"system-barbershop/internal/platform/idempotency"
)

type confirmFakeAvailabilityRepository struct {
	barbershopID string
	found        bool
}

func (f *confirmFakeAvailabilityRepository) ResolveBarbershopID(_ context.Context, _ string) (string, bool, error) {
	return f.barbershopID, f.found, nil
}

func (f *confirmFakeAvailabilityRepository) ListOccupiedIntervals(context.Context, string, string, time.Time, time.Time) ([]time.Time, []time.Time, error) {
	return nil, nil, nil
}

type confirmFakeEffectiveDay struct{}

func (confirmFakeEffectiveDay) ResolveEffectiveDay(context.Context, string, string, string) (bool, []string, []int, error) {
	return true, []string{"09:00"}, []int{540}, nil
}

type confirmFakeBusyBlocks struct{}

func (confirmFakeBusyBlocks) BusyIntervals(context.Context, string, string, string, string, string) ([]time.Time, []time.Time, error) {
	return nil, nil, nil
}

type confirmFakeBookingPolicy struct{}

func (confirmFakeBookingPolicy) BookingPolicy(context.Context, string) (int, int, int, error) {
	return 0, 1, 30, nil
}

type confirmFakeServiceAssignment struct{}

func (confirmFakeServiceAssignment) ActiveAssignedService(context.Context, string, string, string) (string, int, int64, string, bool, error) {
	return "Corte clásico", 30, 1800000, "COP", true, nil
}

type confirmFakeAvailabilityTimezone struct{}

func (confirmFakeAvailabilityTimezone) Timezone(context.Context, string) (string, error) {
	return "America/Bogota", nil
}

type confirmFakeClock struct{ now time.Time }

func (c confirmFakeClock) Now() time.Time { return c.now }

type confirmFakeCustomerRepository struct{}

func (confirmFakeCustomerRepository) FindCustomerMatches(context.Context, string, string, string) (*string, *string, error) {
	return nil, nil, nil
}

type confirmFakePublicAppointmentPort struct {
	decision idempotency.Decision
	stored   idempotency.StoredResponse
	err      error
}

func (f *confirmFakePublicAppointmentPort) CreatePublicAppointment(
	context.Context,
	string, string, string,
	string, string,
	time.Time, time.Time,
	string,
	string,
	int,
	int64,
	string,
	*string,
	*string,
	*string,
	*string,
	string, string, string,
	string, string,
	time.Time, time.Time,
	idempotency.Key,
	idempotency.Fingerprint,
) (idempotency.Decision, idempotency.StoredResponse, error) {
	if f.err != nil {
		return idempotency.Decision{}, idempotency.StoredResponse{}, f.err
	}
	return f.decision, f.stored, nil
}

type confirmFakeEmailPort struct{ calls int }

func (f *confirmFakeEmailPort) SendConfirmation(context.Context, string, string, string, string, string, string) error {
	f.calls++
	return nil
}

const (
	confirmHandlerSlug      = "barberia-confirmacion"
	confirmHandlerServiceID = "55555555-5555-5555-5555-555555555555"
	confirmHandlerBarberID  = "66666666-6666-6666-6666-666666666666"
)

func newConfirmHandler(appointments *confirmFakePublicAppointmentPort, email *confirmFakeEmailPort, now time.Time) *httpapi.ConfirmPublicAppointmentHandler {
	profiles := &fakeRepository{found: true, profile: publicbooking.BarbershopProfile{Name: "Barbería Confirmación", Timezone: "America/Bogota"}}
	availRepo := &confirmFakeAvailabilityRepository{barbershopID: "shop-1", found: true}
	availability := publicbooking.NewAvailabilityService(
		availRepo, confirmFakeEffectiveDay{}, confirmFakeBusyBlocks{}, confirmFakeBookingPolicy{}, confirmFakeServiceAssignment{}, confirmFakeAvailabilityTimezone{}, confirmFakeClock{now: now},
	)
	svc := publicbooking.NewConfirmationService(
		profiles, availRepo, availability, confirmFakeServiceAssignment{}, confirmFakeCustomerRepository{}, appointments, email,
		confirmFakeClock{now: now}, "https://reservas.ejemplo.test",
	)
	return httpapi.NewConfirmPublicAppointmentHandler(svc, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func confirmRequest(t *testing.T, body string, idempotencyKey string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/public/barbershops/"+confirmHandlerSlug+"/services/"+confirmHandlerServiceID+"/barbers/"+confirmHandlerBarberID+"/appointments",
		bytes.NewBufferString(body))
	if idempotencyKey != "" {
		req.Header.Set(httpserver.IdempotencyKeyHeader, idempotencyKey)
	}
	req = httpserver.RequestWithURLParam(req, "slug", confirmHandlerSlug)
	req = httpserver.RequestWithURLParam(req, "serviceId", confirmHandlerServiceID)
	req = httpserver.RequestWithURLParam(req, "barberId", confirmHandlerBarberID)
	return req
}

func validConfirmBody(startsAt time.Time) string {
	return `{"startsAt":"` + startsAt.Format(time.RFC3339) + `","fullName":"Carlos Restrepo","phone":"+573001234567","email":"carlos@example.com","forSomeoneElse":false}`
}

func TestConfirmPublicAppointmentHandler_MissingIdempotencyKey_Returns400(t *testing.T) {
	loc, _ := time.LoadLocation("America/Bogota")
	now := time.Date(2026, time.September, 14, 8, 0, 0, 0, loc)
	startsAt := time.Date(2026, time.September, 14, 9, 0, 0, 0, loc)

	h := newConfirmHandler(&confirmFakePublicAppointmentPort{decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed}}, &confirmFakeEmailPort{}, now)
	req := confirmRequest(t, validConfirmBody(startsAt), "")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestConfirmPublicAppointmentHandler_UnknownField_Returns400(t *testing.T) {
	loc, _ := time.LoadLocation("America/Bogota")
	now := time.Date(2026, time.September, 14, 8, 0, 0, 0, loc)

	h := newConfirmHandler(&confirmFakePublicAppointmentPort{decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed}}, &confirmFakeEmailPort{}, now)
	req := confirmRequest(t, `{"unknownField":true}`, "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestConfirmPublicAppointmentHandler_Success_ForwardsStoredResponseAndSendsEmail(t *testing.T) {
	loc, _ := time.LoadLocation("America/Bogota")
	now := time.Date(2026, time.September, 14, 8, 0, 0, 0, loc)
	startsAt := time.Date(2026, time.September, 14, 9, 0, 0, 0, loc)

	stored := idempotency.StoredResponse{Status: 201, ContentType: "application/json", Body: `{"attendeeName":"Carlos Restrepo","accessToken":"abc123"}`}
	appointments := &confirmFakePublicAppointmentPort{decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed}, stored: stored}
	email := &confirmFakeEmailPort{}
	h := newConfirmHandler(appointments, email, now)

	req := confirmRequest(t, validConfirmBody(startsAt), "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != stored.Body {
		t.Fatalf("expected byte-identical stored body, got %s", rec.Body.String())
	}
	if email.calls != 1 {
		t.Fatalf("expected exactly one confirmation email, got %d", email.calls)
	}
}

func TestConfirmPublicAppointmentHandler_ScheduleConflict_Returns409WithAlternatives(t *testing.T) {
	loc, _ := time.LoadLocation("America/Bogota")
	now := time.Date(2026, time.September, 14, 8, 0, 0, 0, loc)
	// Fuera de la jornada 9:00-18:00 del fixture: nunca es un inicio válido,
	// pero el mismo día sigue teniendo franjas libres (RN-CON-05).
	startsAt := time.Date(2026, time.September, 14, 20, 0, 0, 0, loc)

	h := newConfirmHandler(&confirmFakePublicAppointmentPort{decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed}}, &confirmFakeEmailPort{}, now)
	req := confirmRequest(t, validConfirmBody(startsAt), "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("expected Content-Type application/problem+json, got %q", ct)
	}
	var body struct {
		Code         string                            `json:"code"`
		Alternatives []httpapi.AlternativeSlotResponse `json:"alternatives"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body.Code != "slot-conflict" {
		t.Fatalf("expected code=slot-conflict, got %q", body.Code)
	}
	if len(body.Alternatives) == 0 {
		t.Fatal("expected at least one alternative slot")
	}
}

func TestConfirmPublicAppointmentHandler_RepositoryScheduleConflict_Returns409(t *testing.T) {
	loc, _ := time.LoadLocation("America/Bogota")
	now := time.Date(2026, time.September, 14, 8, 0, 0, 0, loc)
	startsAt := time.Date(2026, time.September, 14, 9, 0, 0, 0, loc)

	appointments := &confirmFakePublicAppointmentPort{err: apperr.ScheduleConflict("la franja elegida ya no está disponible")}
	h := newConfirmHandler(appointments, &confirmFakeEmailPort{}, now)
	req := confirmRequest(t, validConfirmBody(startsAt), "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}
