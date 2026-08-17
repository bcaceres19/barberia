package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/modules/auth/httpapi"
	"system-barbershop/internal/platform/clientip"
)

type stubPhoneChallengeRepository struct {
	requestAccepted bool
	requestPhone    string
	verifyOK        bool
}

func (s stubPhoneChallengeRepository) RequestChallenge(context.Context, string, string, string, auth.PhoneChallengeConfig) (bool, string, error) {
	return s.requestAccepted, s.requestPhone, nil
}

func (s stubPhoneChallengeRepository) VerifyChallenge(context.Context, string, string, string) (bool, error) {
	return s.verifyOK, nil
}

type stubCodeGenerator struct{}

func (stubCodeGenerator) New() (string, error) { return "123456", nil }

type stubSender struct{ calls int }

func (s *stubSender) SendCode(context.Context, string, string) error { s.calls++; return nil }

func testChallengeCfg() auth.PhoneChallengeConfig {
	return auth.PhoneChallengeConfig{ExpiresSeconds: 300, RateWindowSeconds: 900, RateMaxActive: 3, ResendCooldownSeconds: 60}
}

func testThrottleServiceForHandler() *auth.ThrottleService {
	return auth.NewThrottleService(stubThrottleRepository{}, auth.ThrottleConfig{WindowSeconds: 900, EscalationSeconds: 86400, Threshold: 5, RetentionSeconds: 172800}, []byte("secreto-de-prueba-suficientemente-largo"), stubClock{})
}

type stubThrottleRepository struct{}

func (stubThrottleRepository) RegisterAttempt(context.Context, string, auth.ThrottleConfig) (int, bool, time.Time, error) {
	return 0, false, time.Time{}, nil
}

func newChallengeHandler(t *testing.T, accepted bool, sender *stubSender) *httpapi.ChallengeHandler {
	t.Helper()
	svc := auth.NewPhoneChallengeService(stubPhoneChallengeRepository{requestAccepted: accepted, requestPhone: "+573001234567"}, stubCodeGenerator{}, sender, testChallengeCfg(), []byte("secreto-de-prueba-suficientemente-largo"))
	return httpapi.NewChallengeHandler(svc, testThrottleServiceForHandler(), clientip.TrustedProxies{})
}

func TestChallengeHandler_Accepted_Returns202WithGenericMessage(t *testing.T) {
	sender := &stubSender{}
	h := newChallengeHandler(t, true, sender)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/auth/challenge", strings.NewReader(`{"email":"barbero@ejemplo.test"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp httpapi.ChallengeAcceptedResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp.Message == "" {
		t.Fatal("expected a non-empty generic message")
	}
	if sender.calls != 1 {
		t.Fatalf("expected exactly one send when accepted, got %d", sender.calls)
	}
}

// TestChallengeHandler_NotAccepted_StillReturns202 es la prueba central de
// no enumeración (DEC-062): un correo inexistente, no verificado o una IP
// no escalada produce EXACTAMENTE la misma respuesta 202.
func TestChallengeHandler_NotAccepted_StillReturns202(t *testing.T) {
	sender := &stubSender{}
	h := newChallengeHandler(t, false, sender)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/auth/challenge", strings.NewReader(`{"email":"no-existe@ejemplo.test"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202 even when not accepted internally, got %d: %s", rec.Code, rec.Body.String())
	}
	if sender.calls != 0 {
		t.Fatal("expected no send when the repository does not accept the request")
	}
}

func TestChallengeHandler_AcceptedVsNotAccepted_IdenticalResponse(t *testing.T) {
	acceptedRec := httptest.NewRecorder()
	newChallengeHandler(t, true, &stubSender{}).ServeHTTP(acceptedRec, httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"email":"real@ejemplo.test"}`)))

	notAcceptedRec := httptest.NewRecorder()
	newChallengeHandler(t, false, &stubSender{}).ServeHTTP(notAcceptedRec, httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"email":"no-existe@ejemplo.test"}`)))

	if acceptedRec.Code != notAcceptedRec.Code {
		t.Fatalf("expected identical status, got %d vs %d", acceptedRec.Code, notAcceptedRec.Code)
	}
	if acceptedRec.Body.String() != notAcceptedRec.Body.String() {
		t.Fatalf("expected identical body, got %q vs %q", acceptedRec.Body.String(), notAcceptedRec.Body.String())
	}
}

func TestChallengeHandler_MissingEmail_Returns422(t *testing.T) {
	h := newChallengeHandler(t, false, &stubSender{})
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assertProblem(t, rec, http.StatusUnprocessableEntity, "validation-error")
}

func TestChallengeHandler_UnknownField_Returns400(t *testing.T) {
	h := newChallengeHandler(t, false, &stubSender{})
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"email":"a@b.test","extra":"x"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assertProblem(t, rec, http.StatusBadRequest, "invalid-request")
}

func newVerifyHandler(t *testing.T, verifyOK bool) *httpapi.ChallengeVerifyHandler {
	t.Helper()
	svc := auth.NewPhoneChallengeService(stubPhoneChallengeRepository{verifyOK: verifyOK}, stubCodeGenerator{}, &stubSender{}, testChallengeCfg(), []byte("secreto-de-prueba-suficientemente-largo"))
	return httpapi.NewChallengeVerifyHandler(svc, testThrottleServiceForHandler(), clientip.TrustedProxies{})
}

func TestChallengeVerifyHandler_Success_Returns204(t *testing.T) {
	h := newVerifyHandler(t, true)
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"email":"a@b.test","code":"123456"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Body.Len() != 0 {
		t.Fatal("expected an empty body on 204")
	}
}

func TestChallengeVerifyHandler_Failure_Returns401Uniform(t *testing.T) {
	h := newVerifyHandler(t, false)
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"email":"a@b.test","code":"000000"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assertProblem(t, rec, http.StatusUnauthorized, "unauthorized")
}

func TestChallengeVerifyHandler_MalformedCode_Returns422(t *testing.T) {
	h := newVerifyHandler(t, true)
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"email":"a@b.test","code":"12"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assertProblem(t, rec, http.StatusUnprocessableEntity, "validation-error")
}

func TestChallengeVerifyHandler_MissingFields_Returns422(t *testing.T) {
	h := newVerifyHandler(t, true)
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{"email":"a@b.test"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assertProblem(t, rec, http.StatusUnprocessableEntity, "validation-error")
}
