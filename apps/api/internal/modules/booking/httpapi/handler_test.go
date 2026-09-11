package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/modules/booking"
	"system-barbershop/internal/modules/booking/httpapi"
	"system-barbershop/internal/platform/idempotency"
)

// fakeRepository/fakeCatalog/fakeBlocks/fakeTimezones son dobles en memoria
// de los cuatro colaboradores de booking.ManualBookingService: estas
// pruebas verifican el adaptador HTTP en aislamiento (decodificación, forma
// cerrada, extracción del principal, cabecera Idempotency-Key, headers,
// traducción de errores), no PostgreSQL real (esa cobertura vive en
// internal/modules/booking/postgres/manual_repository_test.go).
type fakeRepository struct {
	createFn func(ctx context.Context, barbershopID string, input booking.CreateInternalInput, key idempotency.Key, fingerprint idempotency.Fingerprint) (booking.CreateManualResult, error)
}

func (f *fakeRepository) CreateInternal(context.Context, string, booking.CreateInternalInput) (booking.CreateInternalResult, error) {
	panic("no usado")
}

func (f *fakeRepository) FindCustomerForReconciliation(context.Context, string, *string, *string) (booking.Customer, bool, error) {
	return booking.Customer{}, false, nil
}

func (f *fakeRepository) CreateManual(ctx context.Context, barbershopID string, input booking.CreateInternalInput, key idempotency.Key, fingerprint idempotency.Fingerprint) (booking.CreateManualResult, error) {
	return f.createFn(ctx, barbershopID, input, key, fingerprint)
}

func (f *fakeRepository) ListDailyAgenda(context.Context, string, string, time.Time, time.Time) ([]booking.DailyAgendaEntry, error) {
	panic("no usado")
}

func (f *fakeRepository) GetAppointmentDetail(context.Context, string, string) (booking.AppointmentDetail, bool, error) {
	panic("no usado")
}

func (f *fakeRepository) ListAppointmentHistory(context.Context, string, string, *booking.HistoryCursor, int) ([]booking.HistoryRow, *booking.HistoryCursor, bool, error) {
	panic("no usado")
}

func (f *fakeRepository) CustomerNames(context.Context, string, []string) (map[string]string, error) {
	panic("no usado")
}

func (f *fakeRepository) Reschedule(
	context.Context, string, booking.RescheduleInput, idempotency.Key, idempotency.Fingerprint,
) (booking.RescheduleResult, error) {
	panic("no usado")
}

func (f *fakeRepository) CancelByBarber(
	context.Context, string, booking.CancelAppointmentByBarberInput, idempotency.Key, idempotency.Fingerprint,
) (booking.CancelAppointmentByBarberResult, error) {
	panic("no usado")
}

func (f *fakeRepository) CompleteAppointment(
	context.Context, string, booking.CloseAppointmentInput, idempotency.Key, idempotency.Fingerprint,
) (booking.CompleteAppointmentResult, error) {
	panic("no usado")
}

func (f *fakeRepository) MarkNoShow(
	context.Context, string, booking.CloseAppointmentInput, idempotency.Key, idempotency.Fingerprint,
) (booking.MarkNoShowResult, error) {
	panic("no usado")
}

func (f *fakeRepository) CorrectAppointmentStatus(
	context.Context, string, booking.CorrectAppointmentStatusInput, idempotency.Key, idempotency.Fingerprint,
) (booking.CorrectAppointmentStatusResult, error) {
	panic("no usado")
}

var _ booking.Repository = (*fakeRepository)(nil)

type fakeCatalog struct{}

func (fakeCatalog) ActiveAssignedService(context.Context, string, string, string) (string, int, int64, string, bool, error) {
	return "Corte clásico", 30, 2000000, "COP", true, nil
}

type fakeBlocks struct{}

func (fakeBlocks) HasActiveBlock(context.Context, string, string, time.Time, time.Time, string) (bool, error) {
	return false, nil
}

type fakeTimezones struct{}

func (fakeTimezones) Timezone(context.Context, string) (string, error) { return "America/Bogota", nil }

const testShopID = "11111111-1111-1111-1111-111111111111"

func requestWithPrincipal(method, target string, body []byte) *http.Request {
	var req *http.Request
	if body == nil {
		req = httptest.NewRequest(method, target, nil)
	} else {
		req = httptest.NewRequest(method, target, bytes.NewReader(body))
	}
	principal := auth.Principal{SessionID: "s-1", StaffUserID: "u-1", BarbershopID: testShopID}
	return req.WithContext(auth.ContextWithPrincipal(req.Context(), principal))
}

func validRequestBody() []byte {
	return []byte(`{
		"barberId": "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4",
		"serviceId": "6f1a2b3c-4d5e-4f60-8172-8394a5b6c7d8",
		"attendeeName": "Carlos Restrepo",
		"customerFullName": "Carlos Restrepo",
		"startsAt": "2026-09-03T14:30:00"
	}`)
}

var fixtureTimestamp = time.Date(2026, 8, 27, 15, 4, 5, 0, time.UTC)

func fixedCreateFn(t *testing.T) func(context.Context, string, booking.CreateInternalInput, idempotency.Key, idempotency.Fingerprint) (booking.CreateManualResult, error) {
	return func(_ context.Context, _ string, input booking.CreateInternalInput, _ idempotency.Key, _ idempotency.Fingerprint) (booking.CreateManualResult, error) {
		body, err := json.Marshal(map[string]any{
			"id":              "appt-1",
			"barberId":        input.BarberID,
			"serviceId":       input.ServiceID,
			"customerId":      "cust-1",
			"attendeeName":    input.AttendeeName,
			"startsAt":        input.StartsAt,
			"endsAt":          input.EndsAt,
			"status":          "confirmed",
			"origin":          "manual",
			"serviceName":     input.Service.Name,
			"durationMinutes": input.Service.DurationMinutes,
			"priceAmount":     "20000.00",
			"currency":        input.Service.Currency,
			"customerNote":    nil,
			"createdAt":       fixtureTimestamp,
			"updatedAt":       fixtureTimestamp,
		})
		if err != nil {
			t.Fatalf("marshal fixture body: %v", err)
		}
		return booking.CreateManualResult{
			Decision:    idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Appointment: booking.Appointment{ID: "appt-1"},
			Response:    idempotency.StoredResponse{Status: 201, ContentType: "application/json", Body: string(body)},
		}, nil
	}
}

func newHandler(repo *fakeRepository) *httpapi.CreateManualAppointmentHandler {
	service := booking.NewManualBookingService(repo, fakeCatalog{}, fakeBlocks{}, fakeTimezones{})
	return httpapi.NewCreateManualAppointmentHandler(service)
}

func TestCreateManualAppointmentHandler_ValidRequest_Returns201WithLocation(t *testing.T) {
	repo := &fakeRepository{createFn: fixedCreateFn(t)}
	h := newHandler(repo)

	req := requestWithPrincipal(http.MethodPost, "/api/v1/private/appointments", validRequestBody())
	req.Header.Set("Idempotency-Key", "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if loc := rec.Header().Get("Location"); loc != "/private/appointments/appt-1" {
		t.Fatalf("expected Location /private/appointments/appt-1, got %q", loc)
	}
}

func TestCreateManualAppointmentHandler_MissingIdempotencyKey_Returns400(t *testing.T) {
	repo := &fakeRepository{createFn: fixedCreateFn(t)}
	h := newHandler(repo)

	req := requestWithPrincipal(http.MethodPost, "/api/v1/private/appointments", validRequestBody())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateManualAppointmentHandler_UnknownField_Returns400(t *testing.T) {
	repo := &fakeRepository{createFn: fixedCreateFn(t)}
	h := newHandler(repo)

	body := []byte(`{"barberId":"8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4","serviceId":"6f1a2b3c-4d5e-4f60-8172-8394a5b6c7d8","attendeeName":"Carlos","customerFullName":"Carlos","startsAt":"2026-09-03T14:30:00","unexpected":true}`)
	req := requestWithPrincipal(http.MethodPost, "/api/v1/private/appointments", body)
	req.Header.Set("Idempotency-Key", "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateManualAppointmentHandler_ServiceNotAssigned_Returns404(t *testing.T) {
	repo := &fakeRepository{createFn: fixedCreateFn(t)}
	service := booking.NewManualBookingService(repo, notFoundCatalog{}, fakeBlocks{}, fakeTimezones{})
	h := httpapi.NewCreateManualAppointmentHandler(service)

	req := requestWithPrincipal(http.MethodPost, "/api/v1/private/appointments", validRequestBody())
	req.Header.Set("Idempotency-Key", "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

type notFoundCatalog struct{}

func (notFoundCatalog) ActiveAssignedService(context.Context, string, string, string) (string, int, int64, string, bool, error) {
	return "", 0, 0, "", false, nil
}

func TestCreateManualAppointmentHandler_BlockedInterval_Returns409(t *testing.T) {
	repo := &fakeRepository{createFn: fixedCreateFn(t)}
	service := booking.NewManualBookingService(repo, fakeCatalog{}, blockedBlocks{}, fakeTimezones{})
	h := httpapi.NewCreateManualAppointmentHandler(service)

	req := requestWithPrincipal(http.MethodPost, "/api/v1/private/appointments", validRequestBody())
	req.Header.Set("Idempotency-Key", "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

type blockedBlocks struct{}

func (blockedBlocks) HasActiveBlock(context.Context, string, string, time.Time, time.Time, string) (bool, error) {
	return true, nil
}

func TestCreateManualAppointmentHandler_RepeatedKeySameBody_ReplaysSameStatus(t *testing.T) {
	calls := 0
	repo := &fakeRepository{createFn: func(ctx context.Context, barbershopID string, input booking.CreateInternalInput, key idempotency.Key, fp idempotency.Fingerprint) (booking.CreateManualResult, error) {
		calls++
		result, err := fixedCreateFn(t)(ctx, barbershopID, input, key, fp)
		if calls > 1 {
			result.Decision = idempotency.Decision{Outcome: idempotency.OutcomeReplay, Response: result.Response}
		}
		return result, err
	}}
	h := newHandler(repo)

	first := httptest.NewRecorder()
	req1 := requestWithPrincipal(http.MethodPost, "/api/v1/private/appointments", validRequestBody())
	req1.Header.Set("Idempotency-Key", "key-replay")
	h.ServeHTTP(first, req1)

	second := httptest.NewRecorder()
	req2 := requestWithPrincipal(http.MethodPost, "/api/v1/private/appointments", validRequestBody())
	req2.Header.Set("Idempotency-Key", "key-replay")
	h.ServeHTTP(second, req2)

	if first.Code != http.StatusCreated || second.Code != http.StatusCreated {
		t.Fatalf("expected both responses 201, got %d and %d", first.Code, second.Code)
	}
	if first.Body.String() != second.Body.String() {
		t.Fatalf("expected byte-identical replay body")
	}
	if calls != 2 {
		t.Fatalf("expected CreateManual called twice (once per request), got %d", calls)
	}
}
