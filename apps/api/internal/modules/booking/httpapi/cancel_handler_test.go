package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"system-barbershop/internal/modules/booking"
	"system-barbershop/internal/modules/booking/httpapi"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
	"system-barbershop/internal/platform/idempotency"
)

// fakeCancelRepository es un doble mínimo para probar
// CancelAppointmentByBarberHandler en aislamiento (cabeceras, forma cerrada
// del cuerpo, forma de la respuesta, traducción de errores/Decision), no
// PostgreSQL real (esa cobertura vive en
// internal/modules/booking/postgres/cancel_repository_test.go) ni la
// orquestación de CancelAppointmentByBarberService (esa vive en
// internal/modules/booking/cancel_test.go). Mismo criterio que
// fakeRescheduleRepository en reschedule_handler_test.go.
type fakeCancelRepository struct {
	fakeRepository

	cancelResult booking.CancelAppointmentByBarberResult
	cancelErr    error
}

func (f *fakeCancelRepository) CancelByBarber(
	context.Context, string, booking.CancelAppointmentByBarberInput, idempotency.Key, idempotency.Fingerprint,
) (booking.CancelAppointmentByBarberResult, error) {
	if f.cancelErr != nil {
		return booking.CancelAppointmentByBarberResult{}, f.cancelErr
	}
	return f.cancelResult, nil
}

const validCancelAppointmentID = "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"

func newCancelHandler(repo *fakeCancelRepository) *httpapi.CancelAppointmentByBarberHandler {
	service := booking.NewCancelAppointmentByBarberService(repo)
	return httpapi.NewCancelAppointmentByBarberHandler(service)
}

func cancelRequest(body []byte, idempotencyKey, ifMatch string) *http.Request {
	req := requestWithPrincipal(
		http.MethodPost,
		"/api/v1/private/appointments/"+validCancelAppointmentID+"/cancel",
		body,
	)
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	if ifMatch != "" {
		req.Header.Set("If-Match", ifMatch)
	}
	req = httpserver.RequestWithURLParam(req, "appointmentId", validCancelAppointmentID)
	return req
}

func TestCancelAppointmentByBarberHandler_ValidRequest_NoBody_Returns200(t *testing.T) {
	repo := &fakeCancelRepository{
		cancelResult: booking.CancelAppointmentByBarberResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Response: idempotency.StoredResponse{
				Status:      200,
				ContentType: "application/json",
				Body:        `{"id":"` + validCancelAppointmentID + `","status":"cancelled_by_barber","versionToken":"new-token"}`,
			},
		},
	}
	h := newCancelHandler(repo)

	req := cancelRequest(nil, "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got httpapi.CancelAppointmentByBarberResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got.VersionToken != "new-token" || got.Status != "cancelled_by_barber" {
		t.Fatalf("respuesta inesperada: %+v", got)
	}
}

func TestCancelAppointmentByBarberHandler_ValidRequest_EmptyJSONBody_Returns200(t *testing.T) {
	repo := &fakeCancelRepository{
		cancelResult: booking.CancelAppointmentByBarberResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Response: idempotency.StoredResponse{
				Status:      200,
				ContentType: "application/json",
				Body:        `{"id":"` + validCancelAppointmentID + `","status":"cancelled_by_barber","versionToken":"new-token"}`,
			},
		},
	}
	h := newCancelHandler(repo)

	req := cancelRequest([]byte(`{}`), "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCancelAppointmentByBarberHandler_MissingIdempotencyKey_Returns400(t *testing.T) {
	repo := &fakeCancelRepository{}
	h := newCancelHandler(repo)

	req := cancelRequest(nil, "", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCancelAppointmentByBarberHandler_MissingIfMatch_Returns400(t *testing.T) {
	repo := &fakeCancelRepository{}
	h := newCancelHandler(repo)

	req := cancelRequest(nil, "key-1", "")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	var problem httpserver.Problem
	if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
		t.Fatalf("unmarshal problem: %v", err)
	}
	if problem.Code != "invalid-request" {
		t.Fatalf("code = %q, want %q", problem.Code, "invalid-request")
	}
}

func TestCancelAppointmentByBarberHandler_UnknownField_Returns400(t *testing.T) {
	repo := &fakeCancelRepository{}
	h := newCancelHandler(repo)

	req := cancelRequest([]byte(`{"reason":"x"}`), "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCancelAppointmentByBarberHandler_AppointmentNotFound_Returns404(t *testing.T) {
	repo := &fakeCancelRepository{cancelErr: apperr.NotFound("no existe una cita con ese identificador")}
	h := newCancelHandler(repo)

	req := cancelRequest(nil, "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCancelAppointmentByBarberHandler_VersionConflict_Returns409WithDistinguishableCode(t *testing.T) {
	repo := &fakeCancelRepository{
		cancelErr: apperr.VersionConflict("el turno cambió desde que se leyó; recarga antes de reintentar"),
	}
	h := newCancelHandler(repo)

	req := cancelRequest(nil, "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	var problem httpserver.Problem
	if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
		t.Fatalf("unmarshal problem: %v", err)
	}
	if problem.Code != "version-conflict" {
		t.Fatalf("code = %q, want %q", problem.Code, "version-conflict")
	}
}

func TestCancelAppointmentByBarberHandler_InvalidState_Returns409WithDistinguishableCode(t *testing.T) {
	repo := &fakeCancelRepository{
		cancelErr: apperr.InvalidState("el turno ya no está confirmado; recarga para ver su estado actual"),
	}
	h := newCancelHandler(repo)

	req := cancelRequest(nil, "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	var problem httpserver.Problem
	if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
		t.Fatalf("unmarshal problem: %v", err)
	}
	if problem.Code != "invalid-state" {
		t.Fatalf("code = %q, want %q", problem.Code, "invalid-state")
	}
}

func TestCancelAppointmentByBarberHandler_IdempotencyConflictDecision_Returns409(t *testing.T) {
	repo := &fakeCancelRepository{
		cancelResult: booking.CancelAppointmentByBarberResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeConflictFingerprint},
		},
	}
	h := newCancelHandler(repo)

	req := cancelRequest(nil, "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	var problem httpserver.Problem
	if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
		t.Fatalf("unmarshal problem: %v", err)
	}
	if problem.Code != "idempotency-conflict" {
		t.Fatalf("code = %q, want %q", problem.Code, "idempotency-conflict")
	}
}

func TestCancelAppointmentByBarberHandler_ReplayDecision_ReturnsStoredStatus(t *testing.T) {
	repo := &fakeCancelRepository{
		cancelResult: booking.CancelAppointmentByBarberResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeReplay},
			Response: idempotency.StoredResponse{
				Status:      200,
				ContentType: "application/json",
				Body:        `{"id":"` + validCancelAppointmentID + `","versionToken":"replayed-token"}`,
			},
		},
	}
	h := newCancelHandler(repo)

	req := cancelRequest(nil, "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got httpapi.CancelAppointmentByBarberResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got.VersionToken != "replayed-token" {
		t.Fatalf("VersionToken = %q, want %q (repetición byte a byte)", got.VersionToken, "replayed-token")
	}
}

func TestCancelAppointmentByBarberHandler_NoPrincipal_Returns500(t *testing.T) {
	repo := &fakeCancelRepository{}
	h := newCancelHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/private/appointments/"+validCancelAppointmentID+"/cancel", nil)
	req.Header.Set("Idempotency-Key", "key-1")
	req.Header.Set("If-Match", "old-token")
	req = httpserver.RequestWithURLParam(req, "appointmentId", validCancelAppointmentID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}
