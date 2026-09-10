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

// fakeCloseRepository es un doble mínimo para probar
// CompleteAppointmentHandler/MarkAppointmentNoShowHandler en aislamiento
// (cabeceras, forma cerrada del cuerpo, forma de la respuesta, traducción de
// errores/Decision), no PostgreSQL real (esa cobertura vive en
// internal/modules/booking/postgres/close_repository_test.go) ni la
// orquestación de CompleteAppointmentService/MarkNoShowService (esa vive en
// internal/modules/booking/close_test.go). Mismo criterio que
// fakeCancelRepository en cancel_handler_test.go.
type fakeCloseRepository struct {
	fakeRepository

	completeResult booking.CompleteAppointmentResult
	completeErr    error
	noShowResult   booking.MarkNoShowResult
	noShowErr      error
}

func (f *fakeCloseRepository) CompleteAppointment(
	context.Context, string, booking.CloseAppointmentInput, idempotency.Key, idempotency.Fingerprint,
) (booking.CompleteAppointmentResult, error) {
	if f.completeErr != nil {
		return booking.CompleteAppointmentResult{}, f.completeErr
	}
	return f.completeResult, nil
}

func (f *fakeCloseRepository) MarkNoShow(
	context.Context, string, booking.CloseAppointmentInput, idempotency.Key, idempotency.Fingerprint,
) (booking.MarkNoShowResult, error) {
	if f.noShowErr != nil {
		return booking.MarkNoShowResult{}, f.noShowErr
	}
	return f.noShowResult, nil
}

const validCloseAppointmentID = "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"

func newCompleteHandler(repo *fakeCloseRepository) *httpapi.CompleteAppointmentHandler {
	service := booking.NewCompleteAppointmentService(repo, clock.System{})
	return httpapi.NewCompleteAppointmentHandler(service)
}

func newNoShowHandler(repo *fakeCloseRepository) *httpapi.MarkAppointmentNoShowHandler {
	service := booking.NewMarkNoShowService(repo, clock.System{})
	return httpapi.NewMarkAppointmentNoShowHandler(service)
}

func closeRequest(path string, body []byte, idempotencyKey, ifMatch string) *http.Request {
	req := requestWithPrincipal(
		http.MethodPost,
		"/api/v1/private/appointments/"+validCloseAppointmentID+"/"+path,
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

func TestCompleteAppointmentHandler_ValidRequest_NoBody_Returns200(t *testing.T) {
	repo := &fakeCloseRepository{
		completeResult: booking.CompleteAppointmentResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Response: idempotency.StoredResponse{
				Status:      200,
				ContentType: "application/json",
				Body:        `{"id":"` + validCloseAppointmentID + `","status":"completed","versionToken":"new-token"}`,
			},
		},
	}
	h := newCompleteHandler(repo)

	req := closeRequest("complete", nil, "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got httpapi.AppointmentCompletedResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got.VersionToken != "new-token" || got.Status != "completed" {
		t.Fatalf("respuesta inesperada: %+v", got)
	}
}

func TestCompleteAppointmentHandler_ValidRequest_EmptyJSONBody_Returns200(t *testing.T) {
	repo := &fakeCloseRepository{
		completeResult: booking.CompleteAppointmentResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Response: idempotency.StoredResponse{Status: 200, ContentType: "application/json", Body: `{}`},
		},
	}
	h := newCompleteHandler(repo)

	req := closeRequest("complete", []byte(`{}`), "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCompleteAppointmentHandler_MissingIdempotencyKey_Returns400(t *testing.T) {
	h := newCompleteHandler(&fakeCloseRepository{})

	req := closeRequest("complete", nil, "", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCompleteAppointmentHandler_MissingIfMatch_Returns400(t *testing.T) {
	h := newCompleteHandler(&fakeCloseRepository{})

	req := closeRequest("complete", nil, "key-1", "")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCompleteAppointmentHandler_UnknownField_Returns400(t *testing.T) {
	h := newCompleteHandler(&fakeCloseRepository{})

	req := closeRequest("complete", []byte(`{"status":"completed"}`), "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCompleteAppointmentHandler_AppointmentNotFound_Returns404(t *testing.T) {
	repo := &fakeCloseRepository{completeErr: apperr.NotFound("no existe una cita con ese identificador")}
	h := newCompleteHandler(repo)

	req := closeRequest("complete", nil, "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCompleteAppointmentHandler_NotStarted_Returns422(t *testing.T) {
	repo := &fakeCloseRepository{
		completeErr: apperr.Validation("el turno todavía no comienza; espera hasta su hora de inicio"),
	}
	h := newCompleteHandler(repo)

	req := closeRequest("complete", nil, "key-1", "old-token")
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

func TestCompleteAppointmentHandler_VersionConflict_Returns409WithDistinguishableCode(t *testing.T) {
	repo := &fakeCloseRepository{
		completeErr: apperr.VersionConflict("el turno cambió desde que se leyó; recarga antes de reintentar"),
	}
	h := newCompleteHandler(repo)

	req := closeRequest("complete", nil, "key-1", "old-token")
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

func TestCompleteAppointmentHandler_AlreadyClosedDifferently_Returns409WithDistinguishableCode(t *testing.T) {
	repo := &fakeCloseRepository{
		completeErr: apperr.InvalidState("el turno ya tiene un resultado terminal registrado; una futura corrección (T8) permitirá cambiarlo"),
	}
	h := newCompleteHandler(repo)

	req := closeRequest("complete", nil, "key-1", "old-token")
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

func TestCompleteAppointmentHandler_IdempotencyConflictDecision_Returns409(t *testing.T) {
	repo := &fakeCloseRepository{
		completeResult: booking.CompleteAppointmentResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeConflictFingerprint},
		},
	}
	h := newCompleteHandler(repo)

	req := closeRequest("complete", nil, "key-1", "old-token")
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

func TestCompleteAppointmentHandler_ReplayDecision_ReturnsStoredStatus(t *testing.T) {
	repo := &fakeCloseRepository{
		completeResult: booking.CompleteAppointmentResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeReplay},
			Response: idempotency.StoredResponse{
				Status:      200,
				ContentType: "application/json",
				Body:        `{"id":"` + validCloseAppointmentID + `","versionToken":"replayed-token"}`,
			},
		},
	}
	h := newCompleteHandler(repo)

	req := closeRequest("complete", nil, "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got httpapi.AppointmentCompletedResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got.VersionToken != "replayed-token" {
		t.Fatalf("VersionToken = %q, want %q (repetición byte a byte)", got.VersionToken, "replayed-token")
	}
}

func TestCompleteAppointmentHandler_NoPrincipal_Returns500(t *testing.T) {
	h := newCompleteHandler(&fakeCloseRepository{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/private/appointments/"+validCloseAppointmentID+"/complete", nil)
	req.Header.Set("Idempotency-Key", "key-1")
	req.Header.Set("If-Match", "old-token")
	req = httpserver.RequestWithURLParam(req, "appointmentId", validCloseAppointmentID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}

// A partir de aquí, la cobertura equivalente para
// MarkAppointmentNoShowHandler: mismo criterio exacto que
// CompleteAppointmentHandler arriba, sin repetir cada variante de cabeceras
// (ya cubiertas del lado de complete, ambos handlers comparten
// readClosedAppointmentBody).

func TestMarkAppointmentNoShowHandler_ValidRequest_NoBody_Returns200(t *testing.T) {
	repo := &fakeCloseRepository{
		noShowResult: booking.MarkNoShowResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Response: idempotency.StoredResponse{
				Status:      200,
				ContentType: "application/json",
				Body:        `{"id":"` + validCloseAppointmentID + `","status":"no_show","versionToken":"new-token"}`,
			},
		},
	}
	h := newNoShowHandler(repo)

	req := closeRequest("no-show", nil, "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got httpapi.AppointmentNoShowResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got.VersionToken != "new-token" || got.Status != "no_show" {
		t.Fatalf("respuesta inesperada: %+v", got)
	}
}

func TestMarkAppointmentNoShowHandler_AppointmentNotFound_Returns404(t *testing.T) {
	repo := &fakeCloseRepository{noShowErr: apperr.NotFound("no existe una cita con ese identificador")}
	h := newNoShowHandler(repo)

	req := closeRequest("no-show", nil, "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMarkAppointmentNoShowHandler_NotStarted_Returns422(t *testing.T) {
	repo := &fakeCloseRepository{
		noShowErr: apperr.Validation("el turno todavía no comienza; espera hasta su hora de inicio"),
	}
	h := newNoShowHandler(repo)

	req := closeRequest("no-show", nil, "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMarkAppointmentNoShowHandler_AlreadyClosedDifferently_Returns409WithDistinguishableCode(t *testing.T) {
	repo := &fakeCloseRepository{
		noShowErr: apperr.InvalidState("el turno ya tiene un resultado terminal registrado; una futura corrección (T8) permitirá cambiarlo"),
	}
	h := newNoShowHandler(repo)

	req := closeRequest("no-show", nil, "key-1", "old-token")
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

func TestMarkAppointmentNoShowHandler_ReplayDecision_ReturnsStoredStatus(t *testing.T) {
	repo := &fakeCloseRepository{
		noShowResult: booking.MarkNoShowResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeReplay},
			Response: idempotency.StoredResponse{
				Status:      200,
				ContentType: "application/json",
				Body:        `{"id":"` + validCloseAppointmentID + `","versionToken":"replayed-token"}`,
			},
		},
	}
	h := newNoShowHandler(repo)

	req := closeRequest("no-show", nil, "key-1", "old-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMarkAppointmentNoShowHandler_NoPrincipal_Returns500(t *testing.T) {
	h := newNoShowHandler(&fakeCloseRepository{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/private/appointments/"+validCloseAppointmentID+"/no-show", nil)
	req.Header.Set("Idempotency-Key", "key-1")
	req.Header.Set("If-Match", "old-token")
	req = httpserver.RequestWithURLParam(req, "appointmentId", validCloseAppointmentID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}
