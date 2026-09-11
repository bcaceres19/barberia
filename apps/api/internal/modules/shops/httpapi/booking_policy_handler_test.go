package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"system-barbershop/internal/modules/shops"
	"system-barbershop/internal/modules/shops/httpapi"
)

// fakeBookingPolicyRepository es un doble en memoria de
// shops.BookingPolicyRepository: estas pruebas verifican el adaptador HTTP
// en aislamiento (decodificación, forma cerrada, cabecera If-Match,
// extracción del principal, traducción de errores), no PostgreSQL real
// (esa cobertura vive en
// internal/modules/shops/postgres/booking_policy_repository_test.go).
type fakeBookingPolicyRepository struct {
	getPolicy shops.BookingPolicy
	getFound  bool
	getErr    error

	updateResult shops.BookingPolicyUpdateResult
	updateErr    error
	updateCalls  []struct {
		barbershopID string
		input        shops.BookingPolicyUpdateInput
	}
}

func (f *fakeBookingPolicyRepository) Get(_ context.Context, _ string) (shops.BookingPolicy, bool, error) {
	return f.getPolicy, f.getFound, f.getErr
}

func (f *fakeBookingPolicyRepository) Update(_ context.Context, barbershopID string, input shops.BookingPolicyUpdateInput) (shops.BookingPolicyUpdateResult, error) {
	f.updateCalls = append(f.updateCalls, struct {
		barbershopID string
		input        shops.BookingPolicyUpdateInput
	}{barbershopID, input})
	return f.updateResult, f.updateErr
}

var _ shops.BookingPolicyRepository = (*fakeBookingPolicyRepository)(nil)

const validBookingPolicyBody = `{
	"minAdvanceMinutes": 60,
	"maxAdvanceDays": 3,
	"slotGridMinutes": 15,
	"cancellationDeadlineMinutes": 20,
	"lateCancellationClientAllowed": true,
	"lateCancellationReasonRequired": true
}`

func requestWithPrincipalAndIfMatch(method, target string, body []byte, ifMatch string) *http.Request {
	req := requestWithPrincipal(method, target, body)
	if ifMatch != "" {
		req.Header.Set("If-Match", ifMatch)
	}
	return req
}

// --- GET ------------------------------------------------------------------

func TestGetBookingPolicyHandler_Success_ReturnsSevenFields(t *testing.T) {
	repo := &fakeBookingPolicyRepository{
		getFound: true,
		getPolicy: shops.BookingPolicy{
			MinAdvanceMinutes: 60, MaxAdvanceDays: 3, SlotGridMinutes: 15,
			CancellationDeadlineMinutes: 20, LateCancellationClientAllowed: true,
			LateCancellationReasonRequired: true, VersionToken: "tok-1",
		},
	}
	h := httpapi.NewGetBookingPolicyHandler(shops.NewBookingPolicyService(repo))

	req := requestWithPrincipal(http.MethodGet, "/api/v1/private/settings/booking-policy", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body httpapi.BookingPolicyResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body.MinAdvanceMinutes != 60 || body.MaxAdvanceDays != 3 || body.SlotGridMinutes != 15 ||
		body.CancellationDeadlineMinutes != 20 || !body.LateCancellationClientAllowed ||
		!body.LateCancellationReasonRequired || body.VersionToken != "tok-1" {
		t.Fatalf("unexpected response: %+v", body)
	}
}

func TestGetBookingPolicyHandler_MissingPrincipalInContext_ReturnsSafe500(t *testing.T) {
	repo := &fakeBookingPolicyRepository{}
	h := httpapi.NewGetBookingPolicyHandler(shops.NewBookingPolicyService(repo))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/private/settings/booking-policy", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetBookingPolicyHandler_NotFound_ReturnsUniform404(t *testing.T) {
	repo := &fakeBookingPolicyRepository{getFound: false}
	h := httpapi.NewGetBookingPolicyHandler(shops.NewBookingPolicyService(repo))

	req := requestWithPrincipal(http.MethodGet, "/api/v1/private/settings/booking-policy", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- PUT --------------------------------------------------------------------

func TestUpdateBookingPolicyHandler_Success_ReturnsSavedRepresentation(t *testing.T) {
	saved := shops.BookingPolicy{
		MinAdvanceMinutes: 60, MaxAdvanceDays: 3, SlotGridMinutes: 15,
		CancellationDeadlineMinutes: 20, LateCancellationClientAllowed: true,
		LateCancellationReasonRequired: true, VersionToken: "tok-2",
	}
	repo := &fakeBookingPolicyRepository{updateResult: shops.BookingPolicyUpdateResult{Outcome: shops.BookingPolicyUpdateOutcomeUpdated, Policy: saved}}
	h := httpapi.NewUpdateBookingPolicyHandler(shops.NewBookingPolicyService(repo))

	req := requestWithPrincipalAndIfMatch(http.MethodPut, "/api/v1/private/settings/booking-policy", []byte(validBookingPolicyBody), "tok-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(repo.updateCalls) != 1 {
		t.Fatalf("expected exactly one Update call, got %d", len(repo.updateCalls))
	}
	if repo.updateCalls[0].barbershopID != testShopID {
		t.Fatalf("expected barbershopID=%q from the principal, got %q", testShopID, repo.updateCalls[0].barbershopID)
	}
	if repo.updateCalls[0].input.ExpectedVersionToken != "tok-1" {
		t.Fatalf("expected the If-Match header to reach the service as ExpectedVersionToken, got %+v", repo.updateCalls[0].input)
	}
}

func TestUpdateBookingPolicyHandler_MissingIfMatch_ReturnsInvalidRequestWithoutCallingRepository(t *testing.T) {
	repo := &fakeBookingPolicyRepository{}
	h := httpapi.NewUpdateBookingPolicyHandler(shops.NewBookingPolicyService(repo))

	req := requestWithPrincipal(http.MethodPut, "/api/v1/private/settings/booking-policy", []byte(validBookingPolicyBody))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(repo.updateCalls) != 0 {
		t.Fatal("expected Repository.Update to never be called without If-Match")
	}
}

func TestUpdateBookingPolicyHandler_UnknownField_ReturnsInvalidRequest(t *testing.T) {
	repo := &fakeBookingPolicyRepository{}
	h := httpapi.NewUpdateBookingPolicyHandler(shops.NewBookingPolicyService(repo))

	body := []byte(`{"minAdvanceMinutes":60,"maxAdvanceDays":3,"slotGridMinutes":15,"cancellationDeadlineMinutes":20,"lateCancellationClientAllowed":true,"lateCancellationReasonRequired":true,"barbershopId":"22222222-2222-2222-2222-222222222222"}`)
	req := requestWithPrincipalAndIfMatch(http.MethodPut, "/api/v1/private/settings/booking-policy", body, "tok-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	// CA-093-03: barbershopId nunca se acepta como campo del cuerpo.
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(repo.updateCalls) != 0 {
		t.Fatal("expected Repository.Update to never be called for a request with an unknown field")
	}
}

func TestUpdateBookingPolicyHandler_MalformedJSON_ReturnsInvalidRequest(t *testing.T) {
	repo := &fakeBookingPolicyRepository{}
	h := httpapi.NewUpdateBookingPolicyHandler(shops.NewBookingPolicyService(repo))

	req := requestWithPrincipalAndIfMatch(http.MethodPut, "/api/v1/private/settings/booking-policy", []byte(`{not-json`), "tok-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateBookingPolicyHandler_FieldOutOfRange_ReturnsUnprocessableEntity(t *testing.T) {
	repo := &fakeBookingPolicyRepository{}
	h := httpapi.NewUpdateBookingPolicyHandler(shops.NewBookingPolicyService(repo))

	body := []byte(`{"minAdvanceMinutes":9999,"maxAdvanceDays":3,"slotGridMinutes":15,"cancellationDeadlineMinutes":20,"lateCancellationClientAllowed":true,"lateCancellationReasonRequired":true}`)
	req := requestWithPrincipalAndIfMatch(http.MethodPut, "/api/v1/private/settings/booking-policy", body, "tok-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(repo.updateCalls) != 0 {
		t.Fatal("expected Repository.Update to never be called for an out-of-range field")
	}
}

func TestUpdateBookingPolicyHandler_IncoherentPolicy_ReturnsUnprocessableEntity(t *testing.T) {
	repo := &fakeBookingPolicyRepository{}
	h := httpapi.NewUpdateBookingPolicyHandler(shops.NewBookingPolicyService(repo))

	body := []byte(`{"minAdvanceMinutes":60,"maxAdvanceDays":3,"slotGridMinutes":15,"cancellationDeadlineMinutes":20,"lateCancellationClientAllowed":false,"lateCancellationReasonRequired":true}`)
	req := requestWithPrincipalAndIfMatch(http.MethodPut, "/api/v1/private/settings/booking-policy", body, "tok-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateBookingPolicyHandler_VersionConflict_ReturnsConflict(t *testing.T) {
	repo := &fakeBookingPolicyRepository{updateResult: shops.BookingPolicyUpdateResult{Outcome: shops.BookingPolicyUpdateOutcomeVersionConflict}}
	h := httpapi.NewUpdateBookingPolicyHandler(shops.NewBookingPolicyService(repo))

	req := requestWithPrincipalAndIfMatch(http.MethodPut, "/api/v1/private/settings/booking-policy", []byte(validBookingPolicyBody), "tok-vieja")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateBookingPolicyHandler_MissingPrincipalInContext_ReturnsSafe500(t *testing.T) {
	repo := &fakeBookingPolicyRepository{}
	h := httpapi.NewUpdateBookingPolicyHandler(shops.NewBookingPolicyService(repo))

	req := httptest.NewRequest(http.MethodPut, "/api/v1/private/settings/booking-policy", nil)
	req.Header.Set("If-Match", "tok-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(repo.updateCalls) != 0 {
		t.Fatal("expected Repository.Update to never be invoked without a principal in context")
	}
}
