// Pruebas del adaptador HTTP de la cancelación del cliente (HU-099) en
// aislamiento: cabeceras obligatorias, cuerpo cerrado, traducción de cada
// rechazo a su `code` y reproducción idempotente. La transacción, la política y
// el historial reales se prueban con PostgreSQL en booking/postgres.
package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"system-barbershop/internal/modules/customeraccess"
	"system-barbershop/internal/modules/customeraccess/httpapi"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
	"system-barbershop/internal/platform/idempotency"
)

type fakeCanceller struct {
	calls    int
	version  string
	reason   string
	token    idempotency.Key
	err      error
	decision idempotency.Decision
	stored   idempotency.StoredResponse
}

func (f *fakeCanceller) CancelAppointment(
	_ context.Context, _, _, expectedVersionToken, reason string,
	key idempotency.Key, _ idempotency.Fingerprint,
) (idempotency.Decision, idempotency.StoredResponse, error) {
	f.calls++
	f.version, f.reason, f.token = expectedVersionToken, reason, key
	return f.decision, f.stored, f.err
}

func cancelHandler(canceller *fakeCanceller, found bool) *httpapi.CancelAppointmentHandler {
	repo := &fakeRepository{
		targetFound: found,
		target:      customeraccess.CancellationTarget{BarbershopID: "shop-1", AppointmentID: "appt-1"},
	}
	return httpapi.NewCancelAppointmentHandler(customeraccess.NewService(repo, customeraccess.WithCanceller(canceller)))
}

func cancelRequest(body string, headers map[string]string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/customer/appointments/token-en-claro/cancel", strings.NewReader(body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return httpserver.RequestWithURLParam(req, "token", "token-en-claro")
}

var okHeaders = map[string]string{"Idempotency-Key": "0b7d5c8e-3f4a-4d6b-9a1e-2c3d4e5f6a7b", "If-Match": "v1"}

func newProceed() *fakeCanceller {
	return &fakeCanceller{
		decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed},
		stored:   idempotency.StoredResponse{Status: 200, ContentType: "application/json", Body: `{"status":"cancelled_by_customer"}`},
	}
}

func problemOf(t *testing.T, rec *httptest.ResponseRecorder) httpserver.Problem {
	t.Helper()
	var p httpserver.Problem
	if err := json.NewDecoder(rec.Body).Decode(&p); err != nil {
		t.Fatalf("decode problem: %v", err)
	}
	return p
}

func TestCancelAppointmentHandler_Success_ReplaysStoredResponseAndForwardsReason(t *testing.T) {
	canceller := newProceed()
	rec := httptest.NewRecorder()
	cancelHandler(canceller, true).ServeHTTP(rec, cancelRequest(`{"reason":"  no puedo ir  "}`, okHeaders))

	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != `{"status":"cancelled_by_customer"}` {
		t.Fatalf("got %d %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("Cache-Control = %q", rec.Header().Get("Cache-Control"))
	}
	if canceller.calls != 1 || canceller.version != "v1" || canceller.reason != "  no puedo ir  " {
		t.Fatalf("canceller = %+v", canceller)
	}
}

func TestCancelAppointmentHandler_EmptyBodyAndNullReasonAreAccepted(t *testing.T) {
	for _, body := range []string{"", `{}`, `{"reason":null}`} {
		canceller := newProceed()
		rec := httptest.NewRecorder()
		cancelHandler(canceller, true).ServeHTTP(rec, cancelRequest(body, okHeaders))
		if rec.Code != http.StatusOK || canceller.reason != "" {
			t.Fatalf("body %q: got %d reason=%q", body, rec.Code, canceller.reason)
		}
	}
}

func TestCancelAppointmentHandler_ReplayOutcomeWritesStoredResponse(t *testing.T) {
	canceller := newProceed()
	canceller.decision = idempotency.Decision{Outcome: idempotency.OutcomeReplay}
	rec := httptest.NewRecorder()
	cancelHandler(canceller, true).ServeHTTP(rec, cancelRequest("", okHeaders))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "cancelled_by_customer") {
		t.Fatalf("got %d %s", rec.Code, rec.Body.String())
	}
}

func TestCancelAppointmentHandler_RejectsBadInputBeforeTouchingAnything(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		headers map[string]string
		status  int
	}{
		{"sin Idempotency-Key", "", map[string]string{"If-Match": "v1"}, http.StatusBadRequest},
		{"sin If-Match", "", map[string]string{"Idempotency-Key": okHeaders["Idempotency-Key"]}, http.StatusBadRequest},
		{"campo desconocido (estado)", `{"status":"completed"}`, okHeaders, http.StatusBadRequest},
		{"campo desconocido junto a reason", `{"reason":"x","actor":"staff"}`, okHeaders, http.StatusBadRequest},
		{"json inválido", `{`, okHeaders, http.StatusBadRequest},
		{"reason no es texto", `{"reason":5}`, okHeaders, http.StatusBadRequest},
		{"dos objetos", `{} {}`, okHeaders, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			canceller := newProceed()
			rec := httptest.NewRecorder()
			cancelHandler(canceller, true).ServeHTTP(rec, cancelRequest(tc.body, tc.headers))
			if rec.Code != tc.status {
				t.Fatalf("expected %d, got %d: %s", tc.status, rec.Code, rec.Body.String())
			}
			if canceller.calls != 0 {
				t.Fatal("no debió intentar cancelar")
			}
		})
	}
}

func TestCancelAppointmentHandler_UnusableTokenIsUniform404(t *testing.T) {
	canceller := newProceed()
	rec := httptest.NewRecorder()
	cancelHandler(canceller, false).ServeHTTP(rec, cancelRequest("", okHeaders))
	if rec.Code != http.StatusNotFound || problemOf(t, rec).Code != "not-found" {
		t.Fatalf("got %d %s", rec.Code, rec.Body.String())
	}
	if canceller.calls != 0 {
		t.Fatal("no debió intentar cancelar")
	}
}

func TestCancelAppointmentHandler_RejectionsKeepTheirDistinctCodes(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"plazo vencido", apperr.CancellationWindowClosed("fuera de plazo"), http.StatusUnprocessableEntity, "cancellation-window-closed"},
		{"motivo obligatorio", apperr.CancellationReasonRequired("falta motivo"), http.StatusUnprocessableEntity, "cancellation-reason-required"},
		{"motivo demasiado largo", apperr.Validation("el motivo excede el largo máximo"), http.StatusUnprocessableEntity, "validation-error"},
		{"versión obsoleta", apperr.VersionConflict("cambió"), http.StatusConflict, "version-conflict"},
		{"otro terminal", apperr.InvalidState("ya no está confirmado"), http.StatusConflict, "invalid-state"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			canceller := newProceed()
			canceller.err = tc.err
			rec := httptest.NewRecorder()
			cancelHandler(canceller, true).ServeHTTP(rec, cancelRequest("", okHeaders))
			if rec.Code != tc.status || problemOf(t, rec).Code != tc.code {
				t.Fatalf("expected %d/%s, got %d %s", tc.status, tc.code, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCancelAppointmentHandler_IdempotencyConflictIs409(t *testing.T) {
	canceller := newProceed()
	canceller.decision = idempotency.Decision{Outcome: idempotency.OutcomeConflictFingerprint}
	rec := httptest.NewRecorder()
	cancelHandler(canceller, true).ServeHTTP(rec, cancelRequest("", okHeaders))
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}
