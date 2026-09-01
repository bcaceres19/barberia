package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/modules/auth/httpapi"
	"system-barbershop/internal/platform/clientip"
	"system-barbershop/internal/platform/httpserver"
)

// --- dobles mínimos del núcleo, para probar SOLO la capa HTTP -------------

type stubRepository struct {
	shop  string
	found bool
	cred  auth.Credential
}

func (s stubRepository) ResolveLoginTenant(context.Context, string) (string, bool, error) {
	return s.shop, s.found, nil
}
func (s stubRepository) LookupCredential(context.Context, string, bool, string) (auth.Credential, error) {
	return s.cred, nil
}
func (s stubRepository) CreateSession(context.Context, string, string, string, time.Time, time.Time) error {
	return nil
}
func (s stubRepository) StaffUserNames(context.Context, string, []string) (map[string]string, error) {
	return map[string]string{}, nil
}

type stubHasher struct{ match bool }

func (h stubHasher) Hash(string) (string, error) { return "hash-señuelo", nil }
func (h stubHasher) Verify(string, string) bool  { return h.match }

type stubTokens struct{ token string }

func (s stubTokens) New() (string, error) { return s.token, nil }

type stubClock struct{ now time.Time }

func (c stubClock) Now() time.Time { return c.now }

type failingRepository struct{ err error }

func (f failingRepository) ResolveLoginTenant(context.Context, string) (string, bool, error) {
	return "", false, f.err
}
func (f failingRepository) LookupCredential(context.Context, string, bool, string) (auth.Credential, error) {
	return auth.Credential{}, nil
}
func (f failingRepository) CreateSession(context.Context, string, string, string, time.Time, time.Time) error {
	return nil
}
func (f failingRepository) StaffUserNames(context.Context, string, []string) (map[string]string, error) {
	return map[string]string{}, nil
}

func newHandler(t *testing.T, repo auth.Repository, match bool) *httpapi.LoginHandler {
	t.Helper()
	// throttle=nil: estas pruebas verifican SOLO la capa HTTP del login de
	// HU-005; la integración con el escalamiento de HU-007 vive en
	// contract_test.go/handler_test.go de este mismo paquete, con su propio
	// doble de auth.ThrottleRepository.
	svc, err := auth.NewLoginService(repo, stubHasher{match: match}, stubTokens{token: "token-de-prueba-http"}, stubClock{now: time.Date(2026, 8, 13, 0, 0, 0, 0, time.UTC)}, nil)
	if err != nil {
		t.Fatalf("NewLoginService: %v", err)
	}
	return httpapi.NewLoginHandler(svc, httpapi.DefaultCookieConfig(), clientip.TrustedProxies{})
}

func doLogin(t *testing.T, h *httpapi.LoginHandler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// El middleware RequestID normalmente fija esto; se simula aquí porque
	// la prueba llama al handler directo, sin pasar por httpserver.NewRouter.
	ctx := req.Context()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req.WithContext(ctx))
	return rec
}

// --- pruebas ----------------------------------------------------------

func TestLoginHandler_Success_SetsCookieWithoutTokenInBody(t *testing.T) {
	repo := stubRepository{
		shop: "11111111-1111-1111-1111-111111111111", found: true,
		cred: auth.Credential{Found: true, StaffUserID: "u-1", PasswordHash: "h"},
	}
	h := newHandler(t, repo, true)

	rec := doLogin(t, h, `{"email":"duena.a@ejemplo.test","password":"correcta"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected exactly one cookie, got %d", len(cookies))
	}
	c := cookies[0]
	if c.Name != httpapi.CookieName {
		t.Fatalf("expected cookie name %q, got %q", httpapi.CookieName, c.Name)
	}
	if c.Value != "token-de-prueba-http" {
		t.Fatalf("expected the raw token as cookie value, got %q", c.Value)
	}
	if !c.HttpOnly {
		t.Fatal("expected HttpOnly")
	}
	if !c.Secure {
		t.Fatal("expected Secure (DEC-050)")
	}
	if c.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected SameSite=Lax (DP-SEG-07, provisional), got %v", c.SameSite)
	}
	if c.Path != httpapi.CookiePath {
		t.Fatalf("expected path %q, got %q", httpapi.CookiePath, c.Path)
	}

	if strings.Contains(rec.Body.String(), "token-de-prueba-http") {
		t.Fatal("CA-005-04: the raw token must never appear in the response body")
	}
	var resp httpapi.LoginResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	if resp.ExpiresAt.IsZero() {
		t.Fatal("expected a non-zero expiresAt in the response")
	}
}

func TestLoginHandler_InvalidJSON_Returns400(t *testing.T) {
	h := newHandler(t, stubRepository{}, false)

	rec := doLogin(t, h, `{"email": "sin-cerrar`)

	assertProblem(t, rec, http.StatusBadRequest, "invalid-request")
}

func TestLoginHandler_UnknownField_Returns400(t *testing.T) {
	h := newHandler(t, stubRepository{}, false)

	rec := doLogin(t, h, `{"email":"a@b.test","password":"x","barbershopId":"colado"}`)

	assertProblem(t, rec, http.StatusBadRequest, "invalid-request")
}

func TestLoginHandler_MissingField_Returns422(t *testing.T) {
	h := newHandler(t, stubRepository{}, false)

	rec := doLogin(t, h, `{"email":"a@b.test"}`)

	assertProblem(t, rec, http.StatusUnprocessableEntity, "validation-error")
}

func TestLoginHandler_EmptyPassword_Returns422(t *testing.T) {
	h := newHandler(t, stubRepository{}, false)

	rec := doLogin(t, h, `{"email":"a@b.test","password":""}`)

	assertProblem(t, rec, http.StatusUnprocessableEntity, "validation-error")
}

func TestLoginHandler_WrongPassword_Returns401(t *testing.T) {
	repo := stubRepository{
		shop: "shop-1", found: true,
		cred: auth.Credential{Found: true, StaffUserID: "u-1", PasswordHash: "h"},
	}
	h := newHandler(t, repo, false)

	rec := doLogin(t, h, `{"email":"duena.a@ejemplo.test","password":"incorrecta"}`)

	body := assertProblem(t, rec, http.StatusUnauthorized, "unauthorized")
	if len(rec.Result().Cookies()) != 0 {
		t.Fatal("expected no Set-Cookie on a failed login")
	}
	_ = body
}

func TestLoginHandler_UnknownEmail_ReturnsIdenticalProblemToWrongPassword(t *testing.T) {
	unknown := newHandler(t, stubRepository{found: false}, false)
	recUnknown := doLogin(t, unknown, `{"email":"no-existe@ejemplo.test","password":"cualquiera"}`)

	wrong := newHandler(t, stubRepository{
		shop: "shop-1", found: true,
		cred: auth.Credential{Found: true, StaffUserID: "u-1", PasswordHash: "h"},
	}, false)
	recWrong := doLogin(t, wrong, `{"email":"duena.a@ejemplo.test","password":"incorrecta"}`)

	if recUnknown.Code != recWrong.Code {
		t.Fatalf("expected identical status, got %d vs %d", recUnknown.Code, recWrong.Code)
	}
	if recUnknown.Body.String() != recWrong.Body.String() {
		t.Fatalf("expected identical body (except requestId, both empty here), got %q vs %q",
			recUnknown.Body.String(), recWrong.Body.String())
	}
	headersUnknown := recUnknown.Result().Header.Clone()
	headersWrong := recWrong.Result().Header.Clone()
	headersUnknown.Del("Date")
	headersWrong.Del("Date")
	if len(headersUnknown) != len(headersWrong) {
		t.Fatalf("expected the same functional headers, got %v vs %v", headersUnknown, headersWrong)
	}
}

// TestLoginHandler_InactiveUser_SameProblemAsUnknownEmail cubre CA-005-07 en
// la capa HTTP: un repositorio que representa un usuario inactivo
// (found=false, igual que ResolveLoginTenant lo modela en PostgreSQL real)
// produce la misma respuesta que un correo inexistente.
func TestLoginHandler_InactiveUser_SameProblemAsUnknownEmail(t *testing.T) {
	h := newHandler(t, stubRepository{found: false}, false)

	rec := doLogin(t, h, `{"email":"barbero.b@ejemplo.test","password":"cualquiera"}`)

	assertProblem(t, rec, http.StatusUnauthorized, "unauthorized")
}

func TestLoginHandler_InternalRepositoryError_ReturnsSafe500(t *testing.T) {
	h := newHandler(t, failingRepository{err: errors.New("pq: connection reset by peer at /var/lib/pg")}, false)

	rec := doLogin(t, h, `{"email":"duena.a@ejemplo.test","password":"x"}`)

	body := assertProblem(t, rec, http.StatusInternalServerError, "internal-error")
	if strings.Contains(body, "pq:") || strings.Contains(body, "/var/lib") {
		t.Fatalf("CA-005-04/CA-003-02: internal cause leaked into response: %s", body)
	}
}

func TestLoginHandler_OversizedBody_ReturnsPayloadTooLarge(t *testing.T) {
	h := newHandler(t, stubRepository{}, false)

	huge := `{"email":"a@b.test","password":"` + strings.Repeat("x", 2<<20) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/auth/login", strings.NewReader(huge))
	req.Body = http.MaxBytesReader(httptest.NewRecorder(), req.Body, httpserver.MaxRequestBodyBytes)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assertProblem(t, rec, http.StatusBadRequest, "payload-too-large")
}

// TestLoginHandler_ThroughFullRouter_NeverLogsSensitiveValues cubre la
// prueba obligatoria del prompt: "captura completa de logs del recorrido
// inspeccionada por correo, contraseña, cookie, token y hash: ocurrencias
// permitidas, cero". Monta el handler sobre httpserver.NewRouter (la cadena
// de middleware real, incluida RequestLogger) para que el log capturado sea
// el mismo que produciría cmd/api, no una llamada directa al handler que se
// saltaría el middleware.
func TestLoginHandler_ThroughFullRouter_NeverLogsSensitiveValues(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	repo := stubRepository{
		shop: "11111111-1111-1111-1111-111111111111", found: true,
		cred: auth.Credential{Found: true, StaffUserID: "u-1", PasswordHash: "h"},
	}
	successHandler := newHandler(t, repo, true)
	failHandler := newHandler(t, repo, false)

	router, _ := httpserver.NewRouter(logger)
	router.Post("/api/v1/public/auth/login", func(w http.ResponseWriter, r *http.Request) {
		// Alterna éxito y fallo en la misma ruta según el cuerpo, para
		// capturar ambos recorridos con un único router de prueba.
		body, _ := readAndRestore(r)
		if strings.Contains(body, "correcta") {
			successHandler.ServeHTTP(w, r)
			return
		}
		failHandler.ServeHTTP(w, r)
	})

	const sensitiveEmail = "correo-secreto-de-prueba@ejemplo.test"
	const sensitivePassword = "contraseña-super-secreta-12345"

	for _, body := range []string{
		`{"email":"` + sensitiveEmail + `","password":"correcta"}`,
		`{"email":"` + sensitiveEmail + `","password":"` + sensitivePassword + `"}`,
	} {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/public/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		var capturedToken string
		for _, c := range rec.Result().Cookies() {
			if c.Name == httpapi.CookieName {
				capturedToken = c.Value
			}
		}

		logLine := logBuf.String()
		forbidden := []string{sensitiveEmail, sensitivePassword}
		if capturedToken != "" {
			forbidden = append(forbidden, capturedToken, auth.HashToken(capturedToken))
		}
		for _, value := range forbidden {
			if strings.Contains(logLine, value) {
				t.Fatalf("CA-005-04: sensitive value %q leaked into the log: %s", value, logLine)
			}
		}
	}
}

func readAndRestore(r *http.Request) (string, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))
	return buf.String(), err
}

func assertProblem(t *testing.T, rec *httptest.ResponseRecorder, status int, code string) string {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("expected status %d, got %d: %s", status, rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("expected application/problem+json, got %q", ct)
	}
	var p struct {
		Code   string `json:"code"`
		Status int    `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if p.Code != code {
		t.Fatalf("expected code %q, got %q", code, p.Code)
	}
	return rec.Body.String()
}
