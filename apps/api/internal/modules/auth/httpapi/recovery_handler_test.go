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
	resolveEmail    string
	resolveFound    bool
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

func (s stubRecoveryRepository) ResolveAccountEmailByPhone(context.Context, string) (string, bool, error) {
	return s.resolveEmail, s.resolveFound, nil
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

type stubRecoverySender struct {
	calls        int
	phone, email string
}

func (s *stubRecoverySender) SendCode(_ context.Context, phone, email, _ string) error {
	s.calls++
	s.phone, s.email = phone, email
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

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"channel":"email","email":"a@b.test"}`))
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
	httpapi.NewRecoveryRequestHandler(acceptedSvc, discardRecoveryLogger()).ServeHTTP(acceptedRec, httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"channel":"email","email":"real@ejemplo.test"}`)))

	notAcceptedSvc := newRecoveryService(stubRecoveryRepository{requestAccepted: false}, &stubRecoverySender{})
	notAcceptedRec := httptest.NewRecorder()
	httpapi.NewRecoveryRequestHandler(notAcceptedSvc, discardRecoveryLogger()).ServeHTTP(notAcceptedRec, httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"channel":"email","email":"no-existe@ejemplo.test"}`)))

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
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"channel":"email","email":"a@b.test","extra":"x"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assertProblem(t, rec, http.StatusBadRequest, "invalid-request")
}

func TestRecoveryVerifyHandler_Success_ReturnsTokenAndMaskedDestinations(t *testing.T) {
	svc := newRecoveryService(stubRecoveryRepository{verifyOK: true, verifyPhone: "+573001234567", verifyEmail: "a@b.test"}, &stubRecoverySender{})
	h := httpapi.NewRecoveryVerifyHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"channel":"email","email":"a@b.test","code":"123456"}`))
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

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"channel":"email","email":"a@b.test","code":"000000"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assertProblem(t, rec, http.StatusUnauthorized, "unauthorized")
}

func TestRecoveryVerifyHandler_MalformedCode_Returns422(t *testing.T) {
	svc := newRecoveryService(stubRecoveryRepository{verifyOK: true}, &stubRecoverySender{})
	h := httpapi.NewRecoveryVerifyHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"channel":"email","email":"a@b.test","code":"12"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assertProblem(t, rec, http.StatusUnprocessableEntity, "validation-error")
}

func TestRecoveryResetPasswordHandler_Success_Returns204(t *testing.T) {
	svc := newRecoveryService(stubRecoveryRepository{credentialFound: true, credentialHash: mustHashHTTP(t, "contrasena-actual-larga"), changeOK: true}, &stubRecoverySender{})
	h := httpapi.NewRecoveryResetPasswordHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"channel":"email","email":"a@b.test","resetToken":"tok","newPassword":"contrasena-nueva-larga"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRecoveryResetPasswordHandler_InvalidToken_Returns401Uniform(t *testing.T) {
	svc := newRecoveryService(stubRecoveryRepository{credentialFound: false}, &stubRecoverySender{})
	h := httpapi.NewRecoveryResetPasswordHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"channel":"email","email":"a@b.test","resetToken":"tok-invalido","newPassword":"contrasena-nueva-larga"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assertProblem(t, rec, http.StatusUnauthorized, "unauthorized")
}

func TestRecoveryResetPasswordHandler_WeakPassword_Returns422(t *testing.T) {
	svc := newRecoveryService(stubRecoveryRepository{credentialFound: true, credentialHash: mustHashHTTP(t, "contrasena-actual-larga")}, &stubRecoverySender{})
	h := httpapi.NewRecoveryResetPasswordHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"channel":"email","email":"a@b.test","resetToken":"tok","newPassword":"corta"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assertProblem(t, rec, http.StatusUnprocessableEntity, "validation-error")
}

func TestRecoveryResetPasswordHandler_MissingFields_Returns422(t *testing.T) {
	svc := newRecoveryService(stubRecoveryRepository{}, &stubRecoverySender{})
	h := httpapi.NewRecoveryResetPasswordHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"channel":"email","email":"a@b.test"}`))
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

// --- Canal elegido (DEC-092, DEC-093) ---------------------------------------

func TestRecoveryRequestHandler_EmailChannel_SendsOnlyToEmail(t *testing.T) {
	sender := &stubRecoverySender{}
	svc := newRecoveryService(stubRecoveryRepository{requestAccepted: true, requestPhone: "+573001234567", requestEmail: "a@b.test"}, sender)
	rec := httptest.NewRecorder()
	httpapi.NewRecoveryRequestHandler(svc, discardRecoveryLogger()).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"channel":"email","email":"a@b.test"}`)))

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", rec.Code, rec.Body.String())
	}
	if sender.calls != 1 || sender.phone != "" || sender.email != "a@b.test" {
		t.Fatalf("expected a single send to the email only, got %+v", sender)
	}
}

func TestRecoveryRequestHandler_WhatsAppChannel_SendsOnlyToPhone(t *testing.T) {
	sender := &stubRecoverySender{}
	svc := newRecoveryService(stubRecoveryRepository{resolveEmail: "a@b.test", resolveFound: true, requestAccepted: true, requestPhone: "+573001234567", requestEmail: "a@b.test"}, sender)
	rec := httptest.NewRecorder()
	httpapi.NewRecoveryRequestHandler(svc, discardRecoveryLogger()).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"channel":"whatsapp","phone":"+57 300 123 4567"}`)))

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", rec.Code, rec.Body.String())
	}
	if sender.calls != 1 || sender.phone != "+573001234567" || sender.email != "" {
		t.Fatalf("expected a single send to the phone only, got %+v", sender)
	}
}

// TestRecoveryRequestHandler_WhatsApp_UnknownAmbiguousAndKnown_IdenticalResponse
// cubre CA-008-10 y DEC-065: un número con cuenta, sin cuenta o ambiguo entre
// barberías produce exactamente la misma respuesta.
func TestRecoveryRequestHandler_WhatsApp_UnknownAmbiguousAndKnown_IdenticalResponse(t *testing.T) {
	body := `{"channel":"whatsapp","phone":"+573001234567"}`
	serve := func(repo stubRecoveryRepository) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		httpapi.NewRecoveryRequestHandler(newRecoveryService(repo, &stubRecoverySender{}), discardRecoveryLogger()).
			ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(body)))
		return rec
	}
	known := serve(stubRecoveryRepository{resolveEmail: "a@b.test", resolveFound: true, requestAccepted: true, requestPhone: "+573001234567", requestEmail: "a@b.test"})
	unknownOrAmbiguous := serve(stubRecoveryRepository{resolveFound: false})

	if known.Code != unknownOrAmbiguous.Code || known.Body.String() != unknownOrAmbiguous.Body.String() {
		t.Fatalf("expected identical responses, got %d %q vs %d %q", known.Code, known.Body.String(), unknownOrAmbiguous.Code, unknownOrAmbiguous.Body.String())
	}
}

func TestRecoveryHandlers_InvalidTarget_Returns422(t *testing.T) {
	svc := newRecoveryService(stubRecoveryRepository{}, &stubRecoverySender{})
	handlers := map[string]http.Handler{
		"request": httpapi.NewRecoveryRequestHandler(svc, discardRecoveryLogger()),
		"verify":  httpapi.NewRecoveryVerifyHandler(svc),
		"reset":   httpapi.NewRecoveryResetPasswordHandler(svc),
	}
	extra := map[string]string{
		"request": "",
		"verify":  `,"code":"123456"`,
		"reset":   `,"resetToken":"tok","newPassword":"contrasena-nueva-larga"`,
	}
	targets := map[string]string{
		"sin canal":                  `{"email":"a@b.test"`,
		"canal desconocido":          `{"channel":"sms","phone":"+573001234567"`,
		"email sin correo":           `{"channel":"email"`,
		"whatsapp sin teléfono":      `{"channel":"whatsapp"`,
		"email con phone":            `{"channel":"email","email":"a@b.test","phone":"+573001234567"`,
		"whatsapp con email":         `{"channel":"whatsapp","phone":"+573001234567","email":"a@b.test"`,
		"teléfono sin internacional": `{"channel":"whatsapp","phone":"3001234567"`,
		"teléfono con letras":        `{"channel":"whatsapp","phone":"+57abc1234567"`,
	}
	for op, h := range handlers {
		for name, target := range targets {
			t.Run(op+"/"+name, func(t *testing.T) {
				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(target+extra[op]+"}")))
				assertProblem(t, rec, http.StatusUnprocessableEntity, "validation-error")
			})
		}
	}
}

func TestRecoveryVerifyAndReset_WhatsAppChannel_UseResolvedAccount(t *testing.T) {
	repo := stubRecoveryRepository{
		resolveEmail: "a@b.test", resolveFound: true,
		verifyOK: true, verifyPhone: "+573001234567", verifyEmail: "a@b.test",
		credentialFound: true, credentialHash: mustHashHTTP(t, "contrasena-actual-larga"), changeOK: true,
	}
	svc := newRecoveryService(repo, &stubRecoverySender{})

	rec := httptest.NewRecorder()
	httpapi.NewRecoveryVerifyHandler(svc).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"channel":"whatsapp","phone":"+573001234567","code":"123456"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("verify: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	httpapi.NewRecoveryResetPasswordHandler(svc).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"channel":"whatsapp","phone":"+573001234567","resetToken":"tok","newPassword":"contrasena-nueva-larga"}`)))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("reset: expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRecoveryVerifyAndReset_UnresolvedPhone_Returns401Uniform(t *testing.T) {
	svc := newRecoveryService(stubRecoveryRepository{resolveFound: false, verifyOK: true, credentialFound: true, changeOK: true}, &stubRecoverySender{})

	rec := httptest.NewRecorder()
	httpapi.NewRecoveryVerifyHandler(svc).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"channel":"whatsapp","phone":"+573009999999","code":"123456"}`)))
	assertProblem(t, rec, http.StatusUnauthorized, "unauthorized")

	rec = httptest.NewRecorder()
	httpapi.NewRecoveryResetPasswordHandler(svc).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"channel":"whatsapp","phone":"+573009999999","resetToken":"tok","newPassword":"contrasena-nueva-larga"}`)))
	assertProblem(t, rec, http.StatusUnauthorized, "unauthorized")
}
