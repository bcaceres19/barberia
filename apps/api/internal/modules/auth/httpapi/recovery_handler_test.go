package httpapi_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/modules/auth/httpapi"
)

func discardRecoveryLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

type stubRecoveryRepository struct {
	requestAccepted bool
	requestPhone    string
	requestEmail    string
	verifyOK        bool
	verifyPhone     string
	verifyEmail     string
	credentialFound bool
	credentialHash  string
	changeOK        bool
}

func (s stubRecoveryRepository) RequestRecovery(context.Context, string, string, auth.RecoveryConfig) (bool, string, string, error) {
	return s.requestAccepted, s.requestPhone, s.requestEmail, nil
}

func (s stubRecoveryRepository) VerifyRecovery(context.Context, string, string, string, int) (bool, string, string, error) {
	return s.verifyOK, s.verifyPhone, s.verifyEmail, nil
}

func (s stubRecoveryRepository) CurrentCredential(context.Context, string, string) (bool, string, string, error) {
	return s.credentialFound, s.credentialHash, "argon2id", nil
}

func (s stubRecoveryRepository) ChangePassword(context.Context, string, string, string) (bool, error) {
	return s.changeOK, nil
}

type stubRecoverySender struct{ calls int }

func (s *stubRecoverySender) SendCode(context.Context, string, string, string) error {
	s.calls++
	return nil
}

func testRecoveryCfg() auth.RecoveryConfig {
	return auth.RecoveryConfig{
		CodeExpiresSeconds: 900, CodeMaxAttempts: 5,
		ResendCooldownSeconds: 60, ResendWindowSeconds: 3600, ResendMaxPerWindow: 3,
		ResetTokenExpiresSeconds: 300,
	}
}

func newRecoveryService(repo auth.RecoveryRepository, sender auth.RecoveryCodeSender) *auth.RecoveryService {
	return auth.NewRecoveryService(repo, stubCodeGenerator{}, stubTokenGenerator{}, sender, auth.NewArgon2Hasher(), testRecoveryCfg(), []byte("secreto-de-prueba-suficientemente-largo"))
}

type stubTokenGenerator struct{}

func (stubTokenGenerator) New() (string, error) { return "reset-token-de-prueba", nil }

func TestRecoveryRequestHandler_Accepted_Returns202AndSends(t *testing.T) {
	sender := &stubRecoverySender{}
	svc := newRecoveryService(stubRecoveryRepository{requestAccepted: true, requestPhone: "+573001234567", requestEmail: "a@b.test"}, sender)
	h := httpapi.NewRecoveryRequestHandler(svc, discardRecoveryLogger())

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"email":"a@b.test"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp httpapi.RecoveryRequestAcceptedResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp.Message == "" {
		t.Fatal("expected a non-empty generic message")
	}
	if sender.calls != 1 {
		t.Fatalf("expected exactly one send, got %d", sender.calls)
	}
}

// TestRecoveryRequestHandler_AcceptedVsNotAccepted_IdenticalResponse cubre
// CA-008-01 (no enumeración) al nivel HTTP.
func TestRecoveryRequestHandler_AcceptedVsNotAccepted_IdenticalResponse(t *testing.T) {
	acceptedSvc := newRecoveryService(stubRecoveryRepository{requestAccepted: true, requestPhone: "+573001234567", requestEmail: "real@ejemplo.test"}, &stubRecoverySender{})
	acceptedRec := httptest.NewRecorder()
	httpapi.NewRecoveryRequestHandler(acceptedSvc, discardRecoveryLogger()).ServeHTTP(acceptedRec, httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"email":"real@ejemplo.test"}`)))

	notAcceptedSvc := newRecoveryService(stubRecoveryRepository{requestAccepted: false}, &stubRecoverySender{})
	notAcceptedRec := httptest.NewRecorder()
	httpapi.NewRecoveryRequestHandler(notAcceptedSvc, discardRecoveryLogger()).ServeHTTP(notAcceptedRec, httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"email":"no-existe@ejemplo.test"}`)))

	if acceptedRec.Code != notAcceptedRec.Code {
		t.Fatalf("expected identical status, got %d vs %d", acceptedRec.Code, notAcceptedRec.Code)
	}
	if acceptedRec.Body.String() != notAcceptedRec.Body.String() {
		t.Fatalf("expected identical body, got %q vs %q", acceptedRec.Body.String(), notAcceptedRec.Body.String())
	}
}

func TestRecoveryRequestHandler_MissingEmail_Returns422(t *testing.T) {
	svc := newRecoveryService(stubRecoveryRepository{}, &stubRecoverySender{})
	h := httpapi.NewRecoveryRequestHandler(svc, discardRecoveryLogger())
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assertProblem(t, rec, http.StatusUnprocessableEntity, "validation-error")
}

func TestRecoveryRequestHandler_UnknownField_Returns400(t *testing.T) {
	svc := newRecoveryService(stubRecoveryRepository{}, &stubRecoverySender{})
	h := httpapi.NewRecoveryRequestHandler(svc, discardRecoveryLogger())
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"email":"a@b.test","extra":"x"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assertProblem(t, rec, http.StatusBadRequest, "invalid-request")
}

func TestRecoveryVerifyHandler_Success_ReturnsTokenAndMaskedDestinations(t *testing.T) {
	svc := newRecoveryService(stubRecoveryRepository{verifyOK: true, verifyPhone: "+573001234567", verifyEmail: "a@b.test"}, &stubRecoverySender{})
	h := httpapi.NewRecoveryVerifyHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"email":"a@b.test","code":"123456"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp httpapi.RecoveryVerifyResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp.ResetToken == "" {
		t.Fatal("expected a non-empty reset token")
	}
	if resp.MaskedPhone == "+573001234567" || resp.MaskedEmail == "a@b.test" {
		t.Fatal("expected masked destinations, never the full value")
	}
}

func TestRecoveryVerifyHandler_Failure_Returns401Uniform(t *testing.T) {
	svc := newRecoveryService(stubRecoveryRepository{verifyOK: false}, &stubRecoverySender{})
	h := httpapi.NewRecoveryVerifyHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"email":"a@b.test","code":"000000"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assertProblem(t, rec, http.StatusUnauthorized, "unauthorized")
}

func TestRecoveryVerifyHandler_MalformedCode_Returns422(t *testing.T) {
	svc := newRecoveryService(stubRecoveryRepository{verifyOK: true}, &stubRecoverySender{})
	h := httpapi.NewRecoveryVerifyHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"email":"a@b.test","code":"12"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assertProblem(t, rec, http.StatusUnprocessableEntity, "validation-error")
}

func TestRecoveryResetPasswordHandler_Success_Returns204(t *testing.T) {
	svc := newRecoveryService(stubRecoveryRepository{credentialFound: true, credentialHash: mustHashHTTP(t, "contrasena-actual-larga"), changeOK: true}, &stubRecoverySender{})
	h := httpapi.NewRecoveryResetPasswordHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"email":"a@b.test","resetToken":"tok","newPassword":"contrasena-nueva-larga"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRecoveryResetPasswordHandler_InvalidToken_Returns401Uniform(t *testing.T) {
	svc := newRecoveryService(stubRecoveryRepository{credentialFound: false}, &stubRecoverySender{})
	h := httpapi.NewRecoveryResetPasswordHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"email":"a@b.test","resetToken":"tok-invalido","newPassword":"contrasena-nueva-larga"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assertProblem(t, rec, http.StatusUnauthorized, "unauthorized")
}

func TestRecoveryResetPasswordHandler_WeakPassword_Returns422(t *testing.T) {
	svc := newRecoveryService(stubRecoveryRepository{credentialFound: true, credentialHash: mustHashHTTP(t, "contrasena-actual-larga")}, &stubRecoverySender{})
	h := httpapi.NewRecoveryResetPasswordHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"email":"a@b.test","resetToken":"tok","newPassword":"corta"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assertProblem(t, rec, http.StatusUnprocessableEntity, "validation-error")
}

func TestRecoveryResetPasswordHandler_MissingFields_Returns422(t *testing.T) {
	svc := newRecoveryService(stubRecoveryRepository{}, &stubRecoverySender{})
	h := httpapi.NewRecoveryResetPasswordHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"email":"a@b.test"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assertProblem(t, rec, http.StatusUnprocessableEntity, "validation-error")
}

func mustHashHTTP(t *testing.T, password string) string {
	t.Helper()
	hash, err := auth.NewArgon2Hasher().Hash(password)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	return hash
}
