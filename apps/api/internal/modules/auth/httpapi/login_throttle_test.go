// Pruebas de HU-007 sobre LoginHandler: la integración con el escalamiento
// de HU-007 (429, Retry-After, contraseña nunca evaluada).
package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/modules/auth/httpapi"
	"system-barbershop/internal/platform/clientip"
)

type escalatedThrottleRepository struct {
	retryAfter time.Time
}

func (r escalatedThrottleRepository) RegisterAttempt(context.Context, string, auth.ThrottleConfig) (int, bool, time.Time, error) {
	return 6, true, r.retryAfter, nil
}

func newLoginHandlerWithThrottle(t *testing.T, throttle *auth.ThrottleService, repo auth.Repository, hasherMatch bool) *httpapi.LoginHandler {
	t.Helper()
	svc, err := auth.NewLoginService(repo, stubHasher{match: hasherMatch}, stubTokens{token: "token-de-prueba-throttle"}, stubClock{now: time.Now()}, throttle)
	if err != nil {
		t.Fatalf("NewLoginService: %v", err)
	}
	return httpapi.NewLoginHandler(svc, httpapi.DefaultCookieConfig(), clientip.TrustedProxies{})
}

// TestLoginHandler_Escalated_Returns429WithRetryAfterAndNeverTouchesRepository
// cubre CA-007-02/CA-007-03: una IP escalada responde 429 con Retry-After,
// SIN resolver tenant ni evaluar contraseña (el repositorio de login nunca
// se toca).
func TestLoginHandler_Escalated_Returns429WithRetryAfterAndNeverTouchesRepository(t *testing.T) {
	now := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
	throttle := auth.NewThrottleService(escalatedThrottleRepository{retryAfter: now.Add(3600 * time.Second)}, auth.ThrottleConfig{WindowSeconds: 900, EscalationSeconds: 86400, Threshold: 5, RetentionSeconds: 172800}, []byte("secreto-de-prueba-suficientemente-largo"), stubClock{now: now})

	repo := &countingRepository{stubRepository: stubRepository{shop: "shop-1", found: true, cred: auth.Credential{Found: true, StaffUserID: "u-1", PasswordHash: "h"}}}
	h := newLoginHandlerWithThrottle(t, throttle, repo, true)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/auth/login", strings.NewReader(`{"email":"duena.a@ejemplo.test","password":"correcta"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	body := assertProblem(t, rec, http.StatusTooManyRequests, "challenge-required")
	_ = body

	if got := rec.Header().Get("Retry-After"); got != "3600" {
		t.Fatalf("expected Retry-After=3600, got %q", got)
	}
	if repo.resolveCalls != 0 {
		t.Fatalf("expected ResolveLoginTenant to never be called when escalated (CA-007-02), got %d calls", repo.resolveCalls)
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatal("expected no Set-Cookie when escalated")
	}
}

// TestLoginHandler_NotEscalated_ProceedsNormally confirma que, sin
// escalamiento, el login sigue funcionando exactamente igual que antes de
// HU-007 (regresión).
func TestLoginHandler_NotEscalated_ProceedsNormally(t *testing.T) {
	throttle := auth.NewThrottleService(stubThrottleRepository{}, auth.ThrottleConfig{WindowSeconds: 900, EscalationSeconds: 86400, Threshold: 5, RetentionSeconds: 172800}, []byte("secreto-de-prueba-suficientemente-largo"), stubClock{now: time.Now()})

	repo := stubRepository{shop: "shop-1", found: true, cred: auth.Credential{Found: true, StaffUserID: "u-1", PasswordHash: "h"}}
	h := newLoginHandlerWithThrottle(t, throttle, repo, true)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/auth/login", strings.NewReader(`{"email":"duena.a@ejemplo.test","password":"correcta"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 when not escalated, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(rec.Result().Cookies()) != 1 {
		t.Fatal("expected a session cookie when login succeeds normally")
	}
}

type countingRepository struct {
	stubRepository
	resolveCalls int
}

func (c *countingRepository) ResolveLoginTenant(ctx context.Context, email string) (string, bool, error) {
	c.resolveCalls++
	return c.stubRepository.ResolveLoginTenant(ctx, email)
}
