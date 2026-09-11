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
	"system-barbershop/internal/platform/clock"
	"system-barbershop/internal/platform/httpserver"
	"system-barbershop/internal/platform/idempotency"
)

// fakeCorrectRepository es un doble mínimo para probar
// CorrectAppointmentStatusHandler en aislamiento (cabeceras, forma cerrada
// del cuerpo, forma de la respuesta, traducción de errores/Decision), no
// PostgreSQL real (esa cobertura vive en
// internal/modules/booking/postgres/correct_repository_test.go) ni la
// orquestación de CorrectAppointmentStatusService (esa vive en
// internal/modules/booking/correct_test.go). Mismo criterio que
// fakeCloseRepository en close_handler_test.go.
type fakeCorrectRepository struct {
	fakeRepository

	correctResult booking.CorrectAppointmentStatusResult
	correctErr    error
}

func (f *fakeCorrectRepository) CorrectAppointmentStatus(
	context.Context, string, booking.CorrectAppointmentStatusInput, idempotency.Key, idempotency.Fingerprint,
) (booking.CorrectAppointmentStatusResult, error) {
	if f.correctErr != nil {
		return booking.CorrectAppointmentStatusResult{}, f.correctErr
	}
	return f.correctResult, nil
}

func newCorrectHandler(repo *fakeCorrectRepository) *httpapi.CorrectAppointmentStatusHandler {
	service := booking.NewCorrectAppointmentStatusService(repo, clock.System{})
	return httpapi.NewCorrectAppointmentStatusHandler(service)
}

func correctRequest(body []byte, idempotencyKey, ifMatch string) *http.Request {
	req := requestWithPrincipal(
		http.MethodPost,
		"/api/v1/private/appointments/"+validCloseAppointmentID+"/correct-status",
		body,
	)
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	if ifMatch != "" {
		req.Header.Set("If-Match", ifMatch)
	}
	req = httpserver.RequestWithURLParam(req, "appointmentId", validCloseAppointmentID)
	return req
}

var validCorrectBody = []byte(`{"status":"no_show","reason":"El barbero marcó completed por error, el cliente nunca llegó."}`)

func TestCorrectAppointmentStatusHandler_ValidRequest_Returns200(t *testing.T) {
	repo := &fakeCorrectRepository{
		correctResult: booking.CorrectAppointmentStatusResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Response: idempotency.StoredResponse{
				Status:      200,
				ContentType: "application/json",
				Body:        `{"id":"` + validCloseAppointmentID + `","status":"no_show","versionToken":"new-token"}`,
			},
		},
	}
	h := newCorrectHandler(repo)

	req := correctRequest(validCorrectBody, "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got httpapi.AppointmentStatusCorrectedResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got.VersionToken != "new-token" || got.Status != "no_show" {
		t.Fatalf("respuesta inesperada: %+v", got)
	}
}

func TestCorrectAppointmentStatusHandler_MissingIdempotencyKey_Returns400(t *testing.T) {
	h := newCorrectHandler(&fakeCorrectRepository{})

	req := correctRequest(validCorrectBody, "", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCorrectAppointmentStatusHandler_MissingIfMatch_Returns400(t *testing.T) {
	h := newCorrectHandler(&fakeCorrectRepository{})

	req := correctRequest(validCorrectBody, "key-1", "")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCorrectAppointmentStatusHandler_UnknownField_Returns400(t *testing.T) {
	h := newCorrectHandler(&fakeCorrectRepository{})

	req := correctRequest([]byte(`{"status":"no_show","reason":"motivo","actorStaffUserId":"x"}`), "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestCorrectAppointmentStatusHandler_DestinationConfirmed_Returns422 cubre
// CA-068-02: destino `confirmed` es un cuerpo bien formado que incumple una
// validación de campo (422), no un JSON inválido (400) — el servicio real
// (no el fake) rechaza esto antes de tocar el repositorio.
func TestCorrectAppointmentStatusHandler_DestinationConfirmed_Returns422(t *testing.T) {
	h := newCorrectHandler(&fakeCorrectRepository{})

	req := correctRequest([]byte(`{"status":"confirmed","reason":"motivo"}`), "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
	var problem httpserver.Problem
	if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
		t.Fatalf("unmarshal problem: %v", err)
	}
	if problem.Code != "validation-error" {
		t.Fatalf("code = %q, want %q", problem.Code, "validation-error")
	}
}

func TestCorrectAppointmentStatusHandler_ReasonBlank_Returns422(t *testing.T) {
	h := newCorrectHandler(&fakeCorrectRepository{})

	req := correctRequest([]byte(`{"status":"no_show","reason":"   "}`), "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCorrectAppointmentStatusHandler_AppointmentNotFound_Returns404(t *testing.T) {
	repo := &fakeCorrectRepository{correctErr: apperr.NotFound("no existe una cita con ese identificador")}
	h := newCorrectHandler(repo)

	req := correctRequest(validCorrectBody, "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCorrectAppointmentStatusHandler_StillConfirmed_Returns409WithDistinguishableCode(t *testing.T) {
	repo := &fakeCorrectRepository{
		correctErr: apperr.InvalidState("el turno todavía no tiene un resultado terminal para corregir"),
	}
	h := newCorrectHandler(repo)

	req := correctRequest(validCorrectBody, "key-1", "old-token")
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

func TestCorrectAppointmentStatusHandler_VersionConflict_Returns409WithDistinguishableCode(t *testing.T) {
	repo := &fakeCorrectRepository{
		correctErr: apperr.VersionConflict("el turno cambió desde que se leyó; recarga antes de reintentar"),
	}
	h := newCorrectHandler(repo)

	req := correctRequest(validCorrectBody, "key-1", "old-token")
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

// TestCorrectAppointmentStatusHandler_ScheduleConflict_Returns409WithDistinguishableCode
// cubre CA-068-04: corregir hacia un terminal que vuelve a ocupar una franja
// ya ocupada por otra cita.
func TestCorrectAppointmentStatusHandler_ScheduleConflict_Returns409WithDistinguishableCode(t *testing.T) {
	repo := &fakeCorrectRepository{
		correctErr: apperr.Conflict("el barbero ya tiene una cita en ese intervalo"),
	}
	h := newCorrectHandler(repo)

	req := correctRequest(validCorrectBody, "key-1", "old-token")
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

func TestCorrectAppointmentStatusHandler_IdempotencyConflictDecision_Returns409(t *testing.T) {
	repo := &fakeCorrectRepository{
		correctResult: booking.CorrectAppointmentStatusResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeConflictFingerprint},
		},
	}
	h := newCorrectHandler(repo)

	req := correctRequest(validCorrectBody, "key-1", "old-token")
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

func TestCorrectAppointmentStatusHandler_ReplayDecision_ReturnsStoredStatus(t *testing.T) {
	repo := &fakeCorrectRepository{
		correctResult: booking.CorrectAppointmentStatusResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeReplay},
			Response: idempotency.StoredResponse{
				Status:      200,
				ContentType: "application/json",
				Body:        `{"id":"` + validCloseAppointmentID + `","versionToken":"replayed-token"}`,
			},
		},
	}
	h := newCorrectHandler(repo)

	req := correctRequest(validCorrectBody, "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got httpapi.AppointmentStatusCorrectedResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got.VersionToken != "replayed-token" {
		t.Fatalf("VersionToken = %q, want %q (repetición byte a byte)", got.VersionToken, "replayed-token")
	}
}

func TestCorrectAppointmentStatusHandler_NoPrincipal_Returns500(t *testing.T) {
	h := newCorrectHandler(&fakeCorrectRepository{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/private/appointments/"+validCloseAppointmentID+"/correct-status", nil)
	req.Header.Set("Idempotency-Key", "key-1")
	req.Header.Set("If-Match", "old-token")
	req = httpserver.RequestWithURLParam(req, "appointmentId", validCloseAppointmentID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}
