package httpapi_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/modules/auth/httpapi"
	"system-barbershop/internal/platform/httpserver"
)

// --- doble mínimo del núcleo, para probar SOLO la capa HTTP ---------------

type spySessionRepo struct {
	resolveShop  string
	resolveFound bool
	resolveErr   error
	resolveCalls int

	renewPrincipal auth.Principal
	renewValid     bool
	renewErr       error
	renewCalls     int

	revokeErr   error
	revokeCalls int

	barbershopName      string
	barbershopNameErr   error
	barbershopNameCalls int
}

func (s *spySessionRepo) ResolveSessionTenant(context.Context, string) (string, bool, error) {
	s.resolveCalls++
	return s.resolveShop, s.resolveFound, s.resolveErr
}

func (s *spySessionRepo) ValidateAndRenewSession(context.Context, string, string, time.Time, time.Time) (auth.Principal, bool, error) {
	s.renewCalls++
	return s.renewPrincipal, s.renewValid, s.renewErr
}

func (s *spySessionRepo) RevokeSession(context.Context, string, string, time.Time) error {
	s.revokeCalls++
	return s.revokeErr
}

func (s *spySessionRepo) BarbershopName(context.Context, string) (string, error) {
	s.barbershopNameCalls++
	return s.barbershopName, s.barbershopNameErr
}

func newSessionMiddleware(t *testing.T, repo *spySessionRepo) *httpapi.SessionMiddleware {
	t.Helper()
	svc := auth.NewSessionService(repo, stubClock{now: time.Date(2026, 8, 13, 0, 0, 0, 0, time.UTC)})
	return httpapi.NewSessionMiddleware(svc, httpapi.DefaultCookieConfig())
}

func protectedHandler(called *bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		*called = true
		if _, ok := auth.PrincipalFromContext(r.Context()); !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func doPrivateRequest(t *testing.T, mw *httpapi.SessionMiddleware, next http.Handler, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/private/anything", nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	mw.RequireSession(next).ServeHTTP(rec, req)
	return rec
}

// --- pruebas ----------------------------------------------------------

func TestSessionMiddleware_NoCookie_Returns401UniformWithoutCallingService(t *testing.T) {
	repo := &spySessionRepo{}
	mw := newSessionMiddleware(t, repo)
	var called bool

	rec := doPrivateRequest(t, mw, protectedHandler(&called), nil)

	assertProblem(t, rec, http.StatusUnauthorized, "unauthorized")
	if called {
		t.Fatal("expected next handler to never run without a cookie (CA-006-04)")
	}
	if repo.resolveCalls != 0 {
		t.Fatal("expected no repository call for a missing cookie (CA-006-05)")
	}
}

func TestSessionMiddleware_MalformedCookie_Returns401WithoutTouchingRepository(t *testing.T) {
	repo := &spySessionRepo{}
	mw := newSessionMiddleware(t, repo)
	var called bool

	// Caracteres fuera del alfabeto base64.RawURLEncoding.
	rec := doPrivateRequest(t, mw, protectedHandler(&called), &http.Cookie{Name: httpapi.CookieName, Value: "tiene espacios y ó"})

	assertProblem(t, rec, http.StatusUnauthorized, "unauthorized")
	if called {
		t.Fatal("expected next handler to never run with a malformed cookie")
	}
	if repo.resolveCalls != 0 {
		t.Fatal("CA-006-05: a malformed cookie must never reach the repository (no hash, no DB call)")
	}
}

func TestSessionMiddleware_OversizedCookie_Returns401WithoutTouchingRepository(t *testing.T) {
	repo := &spySessionRepo{}
	mw := newSessionMiddleware(t, repo)
	var called bool

	huge := strings.Repeat("a", 10_000)
	rec := doPrivateRequest(t, mw, protectedHandler(&called), &http.Cookie{Name: httpapi.CookieName, Value: huge})

	assertProblem(t, rec, http.StatusUnauthorized, "unauthorized")
	if called {
		t.Fatal("expected next handler to never run with an oversized cookie")
	}
	if repo.resolveCalls != 0 {
		t.Fatal("CA-006-05: an oversized cookie must never reach the repository")
	}
}

func TestSessionMiddleware_UnknownRevokedOrExpiredToken_Returns401Uniform(t *testing.T) {
	for _, tc := range []struct {
		name string
		repo *spySessionRepo
	}{
		{"unknown-tenant", &spySessionRepo{resolveFound: false}},
		{"revoked-or-expired-or-inactive", &spySessionRepo{resolveShop: "shop-1", resolveFound: true, renewValid: false}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mw := newSessionMiddleware(t, tc.repo)
			var called bool

			rec := doPrivateRequest(t, mw, protectedHandler(&called), &http.Cookie{Name: httpapi.CookieName, Value: "un-token-con-forma-valida-1234"})

			assertProblem(t, rec, http.StatusUnauthorized, "unauthorized")
			if called {
				t.Fatal("expected next handler to never run for an invalid session")
			}
		})
	}
}

func TestSessionMiddleware_ValidSession_CallsNextWithPrincipalInContext(t *testing.T) {
	principal := auth.Principal{SessionID: "s-1", StaffUserID: "u-1", BarbershopID: "shop-1"}
	repo := &spySessionRepo{resolveShop: "shop-1", resolveFound: true, renewValid: true, renewPrincipal: principal}
	mw := newSessionMiddleware(t, repo)
	var called bool

	rec := doPrivateRequest(t, mw, protectedHandler(&called), &http.Cookie{Name: httpapi.CookieName, Value: "un-token-con-forma-valida-1234"})

	if !called {
		t.Fatal("expected next handler to run for a valid session")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from the downstream handler (principal was in context), got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSessionMiddleware_RepositoryFailure_ReturnsSafe500(t *testing.T) {
	repo := &spySessionRepo{resolveErr: errors.New("pq: connection reset by peer")}
	mw := newSessionMiddleware(t, repo)
	var called bool

	rec := doPrivateRequest(t, mw, protectedHandler(&called), &http.Cookie{Name: httpapi.CookieName, Value: "un-token-con-forma-valida-1234"})

	body := assertProblem(t, rec, http.StatusInternalServerError, "internal-error")
	if strings.Contains(body, "pq:") {
		t.Fatalf("internal cause leaked into response: %s", body)
	}
	if called {
		t.Fatal("expected next handler to never run on a repository failure")
	}
}

// TestSessionMiddleware_ThroughFullRouter_NeverLogsSessionMaterial cubre
// CA-006-05: el token de la cookie (en claro o hasheado) nunca aparece en
// los logs, montado sobre httpserver.NewRouter como cmd/api realmente lo
// hace.
func TestSessionMiddleware_ThroughFullRouter_NeverLogsSessionMaterial(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	principal := auth.Principal{SessionID: "s-1", StaffUserID: "u-1", BarbershopID: "shop-1"}
	repo := &spySessionRepo{resolveShop: "shop-1", resolveFound: true, renewValid: true, renewPrincipal: principal}
	mw := newSessionMiddleware(t, repo)

	router, private := httpserver.NewRouter(logger)
	private.Use(mw.RequireSession)
	private.Get("/whoami", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	const sensitiveToken = "token-secreto-de-sesion-1234567890"
	for _, cookieVal := range []string{sensitiveToken, "otro-token-invalido"} {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/private/whoami", nil)
		req.AddCookie(&http.Cookie{Name: httpapi.CookieName, Value: cookieVal})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
	}

	logLine := logBuf.String()
	for _, forbidden := range []string{sensitiveToken, auth.HashToken(sensitiveToken), httpapi.CookieName + "="} {
		if strings.Contains(logLine, forbidden) {
			t.Fatalf("CA-006-05: session material leaked into the log: %s", logLine)
		}
	}
}
