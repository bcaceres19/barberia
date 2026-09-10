package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"system-barbershop/internal/modules/booking"
	"system-barbershop/internal/modules/booking/httpapi"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
	"system-barbershop/internal/platform/idempotency"
)

// fakeRescheduleRepository es un doble mínimo para probar
// RescheduleAppointmentHandler en aislamiento (cabeceras, decodificación,
// forma de la respuesta, traducción de errores/Decision), no PostgreSQL
// real (esa cobertura vive en
// internal/modules/booking/postgres/reschedule_repository_test.go) ni la
// orquestación de RescheduleService (esa vive en
// internal/modules/booking/reschedule_test.go).
type fakeRescheduleRepository struct {
	fakeRepository

	detail      booking.AppointmentDetail
	detailFound bool

	rescheduleResult booking.RescheduleResult
	rescheduleErr    error
}

func (f *fakeRescheduleRepository) GetAppointmentDetail(context.Context, string, string) (booking.AppointmentDetail, bool, error) {
	return f.detail, f.detailFound, nil
}

func (f *fakeRescheduleRepository) Reschedule(
	context.Context, string, booking.RescheduleInput, idempotency.Key, idempotency.Fingerprint,
) (booking.RescheduleResult, error) {
	if f.rescheduleErr != nil {
		return booking.RescheduleResult{}, f.rescheduleErr
	}
	return f.rescheduleResult, nil
}

const validRescheduleAppointmentID = "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"

func newRescheduleHandler(repo *fakeRescheduleRepository) *httpapi.RescheduleAppointmentHandler {
	service := booking.NewRescheduleService(repo, fakeBlocks{}, fakeTimezones{}, fakeAgendaClock{now: time.Now()})
	return httpapi.NewRescheduleAppointmentHandler(service)
}

func rescheduleRequest(body []byte, idempotencyKey, ifMatch string) *http.Request {
	req := requestWithPrincipal(
		http.MethodPost,
		"/api/v1/private/appointments/"+validRescheduleAppointmentID+"/reschedule",
		body,
	)
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	if ifMatch != "" {
		req.Header.Set("If-Match", ifMatch)
	}
	req = httpserver.RequestWithURLParam(req, "appointmentId", validRescheduleAppointmentID)
	return req
}

func validRescheduleBody() []byte {
	return []byte(`{"startsAt": "2099-01-01T14:30:00"}`)
}

func readyRescheduleDetail() booking.AppointmentDetail {
	return booking.AppointmentDetail{
		ID:                      validRescheduleAppointmentID,
		BarberID:                "barber-1",
		Status:                  booking.StatusConfirmed,
		DurationMinutesSnapshot: 30,
	}
}

func TestRescheduleAppointmentHandler_ValidRequest_Returns200(t *testing.T) {
	repo := &fakeRescheduleRepository{
		detail:      readyRescheduleDetail(),
		detailFound: true,
		rescheduleResult: booking.RescheduleResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Response: idempotency.StoredResponse{
				Status:      200,
				ContentType: "application/json",
				Body:        `{"id":"` + validRescheduleAppointmentID + `","versionToken":"new-token"}`,
			},
		},
	}
	h := newRescheduleHandler(repo)

	req := rescheduleRequest(validRescheduleBody(), "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got httpapi.AppointmentRescheduledResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got.VersionToken != "new-token" {
		t.Fatalf("VersionToken = %q, want %q", got.VersionToken, "new-token")
	}
}

func TestRescheduleAppointmentHandler_MissingIdempotencyKey_Returns400(t *testing.T) {
	repo := &fakeRescheduleRepository{detail: readyRescheduleDetail(), detailFound: true}
	h := newRescheduleHandler(repo)

	req := rescheduleRequest(validRescheduleBody(), "", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRescheduleAppointmentHandler_MissingIfMatch_Returns400(t *testing.T) {
	repo := &fakeRescheduleRepository{detail: readyRescheduleDetail(), detailFound: true}
	h := newRescheduleHandler(repo)

	req := rescheduleRequest(validRescheduleBody(), "key-1", "")
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

func TestRescheduleAppointmentHandler_UnknownField_Returns400(t *testing.T) {
	repo := &fakeRescheduleRepository{detail: readyRescheduleDetail(), detailFound: true}
	h := newRescheduleHandler(repo)

	req := rescheduleRequest([]byte(`{"startsAt":"2099-01-01T14:30:00","barberId":"x"}`), "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRescheduleAppointmentHandler_AppointmentNotFound_Returns404(t *testing.T) {
	repo := &fakeRescheduleRepository{detailFound: false}
	h := newRescheduleHandler(repo)

	req := rescheduleRequest(validRescheduleBody(), "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRescheduleAppointmentHandler_VersionConflict_Returns409WithDistinguishableCode(t *testing.T) {
	repo := &fakeRescheduleRepository{
		detail: readyRescheduleDetail(), detailFound: true,
		rescheduleErr: apperr.VersionConflict("el turno cambió desde que se leyó; recarga antes de reintentar"),
	}
	h := newRescheduleHandler(repo)

	req := rescheduleRequest(validRescheduleBody(), "key-1", "old-token")
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

func TestRescheduleAppointmentHandler_InvalidState_Returns409WithDistinguishableCode(t *testing.T) {
	repo := &fakeRescheduleRepository{
		detail: readyRescheduleDetail(), detailFound: true,
		rescheduleErr: apperr.InvalidState("el turno ya no está confirmado; recarga para ver su estado actual"),
	}
	h := newRescheduleHandler(repo)

	req := rescheduleRequest(validRescheduleBody(), "key-1", "old-token")
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

func TestRescheduleAppointmentHandler_ScheduleConflict_Returns409WithConflictCode(t *testing.T) {
	repo := &fakeRescheduleRepository{
		detail: readyRescheduleDetail(), detailFound: true,
		rescheduleErr: apperr.Conflict("el barbero ya tiene una cita en ese intervalo"),
	}
	h := newRescheduleHandler(repo)

	req := rescheduleRequest(validRescheduleBody(), "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	var problem httpserver.Problem
	if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
		t.Fatalf("unmarshal problem: %v", err)
	}
	if problem.Code != "conflict" {
		t.Fatalf("code = %q, want %q", problem.Code, "conflict")
	}
}

func TestRescheduleAppointmentHandler_NotFuture_Returns422(t *testing.T) {
	repo := &fakeRescheduleRepository{detail: readyRescheduleDetail(), detailFound: true}
	h := newRescheduleHandler(repo)

	req := rescheduleRequest([]byte(`{"startsAt":"2020-01-01T00:00:00"}`), "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRescheduleAppointmentHandler_IdempotencyConflictDecision_Returns409(t *testing.T) {
	repo := &fakeRescheduleRepository{
		detail: readyRescheduleDetail(), detailFound: true,
		rescheduleResult: booking.RescheduleResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeConflictFingerprint},
		},
	}
	h := newRescheduleHandler(repo)

	req := rescheduleRequest(validRescheduleBody(), "key-1", "old-token")
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

func TestRescheduleAppointmentHandler_ReplayDecision_ReturnsStoredStatus(t *testing.T) {
	repo := &fakeRescheduleRepository{
		detail: readyRescheduleDetail(), detailFound: true,
		rescheduleResult: booking.RescheduleResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeReplay},
			Response: idempotency.StoredResponse{
				Status:      200,
				ContentType: "application/json",
				Body:        `{"id":"` + validRescheduleAppointmentID + `","versionToken":"replayed-token"}`,
			},
		},
	}
	h := newRescheduleHandler(repo)

	req := rescheduleRequest(validRescheduleBody(), "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got httpapi.AppointmentRescheduledResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got.VersionToken != "replayed-token" {
		t.Fatalf("VersionToken = %q, want %q (repetición byte a byte)", got.VersionToken, "replayed-token")
	}
}

func TestRescheduleAppointmentHandler_NoPrincipal_Returns500(t *testing.T) {
	repo := &fakeRescheduleRepository{detail: readyRescheduleDetail(), detailFound: true}
	h := newRescheduleHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/private/appointments/"+validRescheduleAppointmentID+"/reschedule", nil)
	req.Header.Set("Idempotency-Key", "key-1")
	req.Header.Set("If-Match", "old-token")
	req = httpserver.RequestWithURLParam(req, "appointmentId", validRescheduleAppointmentID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}
