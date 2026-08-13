package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/modules/auth/httpapi"
)

func newLogoutHandler(t *testing.T, repo *spySessionRepo) *httpapi.LogoutHandler {
	t.Helper()
	svc := auth.NewSessionService(repo, stubClock{now: time.Date(2026, 8, 13, 0, 0, 0, 0, time.UTC)})
	return httpapi.NewLogoutHandler(svc, httpapi.DefaultCookieConfig())
}

// TestLogoutHandler_Success_RevokesAndClearsCookie cubre CA-006-02: la
// respuesta exitosa limpia la cookie con Max-Age=0 y los mismos atributos
// con que HU-005 la emitió, sin cuerpo (204).
func TestLogoutHandler_Success_RevokesAndClearsCookie(t *testing.T) {
	repo := &spySessionRepo{}
	h := newLogoutHandler(t, repo)

	principal := auth.Principal{SessionID: "s-1", StaffUserID: "u-1", BarbershopID: "shop-a"}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/private/auth/logout", nil)
	req = req.WithContext(auth.ContextWithPrincipal(req.Context(), principal))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("expected an empty body for 204, got %q", rec.Body.String())
	}
	if repo.revokeCalls != 1 {
		t.Fatalf("expected exactly one RevokeSession call, got %d", repo.revokeCalls)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected exactly one Set-Cookie, got %d", len(cookies))
	}
	c := cookies[0]
	if c.Name != httpapi.CookieName {
		t.Fatalf("expected cookie name %q, got %q", httpapi.CookieName, c.Name)
	}
	if c.Value != "" {
		t.Fatalf("expected an empty cookie value when clearing, got %q", c.Value)
	}
	if c.MaxAge != -1 && c.MaxAge != 0 {
		t.Fatalf("expected MaxAge<=0 (delete immediately), got %d", c.MaxAge)
	}
	if c.Path != httpapi.CookiePath {
		t.Fatalf("expected path %q, got %q", httpapi.CookiePath, c.Path)
	}
	if !c.HttpOnly || !c.Secure || c.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected the same attributes as the original cookie, got %+v", c)
	}
}

// TestLogoutHandler_MissingPrincipalInContext_ReturnsSafe500 cubre el caso
// defensivo de una ruta montada sin pasar por SessionMiddleware (nunca
// debería ocurrir, CA-006-04 lo evita estructuralmente): el handler no debe
// asumir un bypass silencioso, ni ejecutar Logout con datos inventados.
func TestLogoutHandler_MissingPrincipalInContext_ReturnsSafe500(t *testing.T) {
	repo := &spySessionRepo{}
	h := newLogoutHandler(t, repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/private/auth/logout", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	if repo.revokeCalls != 0 {
		t.Fatal("expected Logout to never be invoked without a principal in context")
	}
}

func TestLogoutHandler_ServiceFailure_ReturnsSafe500(t *testing.T) {
	repo := &spySessionRepo{revokeErr: context.DeadlineExceeded}
	h := newLogoutHandler(t, repo)

	principal := auth.Principal{SessionID: "s-1", StaffUserID: "u-1", BarbershopID: "shop-a"}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/private/auth/logout", nil)
	req = req.WithContext(auth.ContextWithPrincipal(req.Context(), principal))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatal("expected no Set-Cookie on a failed logout")
	}
}
