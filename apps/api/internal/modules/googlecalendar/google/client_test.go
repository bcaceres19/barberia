package google_test

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"golang.org/x/oauth2"

	"system-barbershop/internal/modules/googlecalendar"
	"system-barbershop/internal/modules/googlecalendar/google"
)

// googleStub es un Google falso: registra lo que recibe y responde lo que el
// caso configure. Ninguna prueba toca la red real.
type googleStub struct {
	mu          sync.Mutex
	tokenForms  []url.Values
	revokeForms []url.Values
	tokenStatus int
	tokenBody   string
	revokeCode  int
	server      *httptest.Server
}

func newStub(t *testing.T) *googleStub {
	t.Helper()
	s := &googleStub{tokenStatus: http.StatusOK, revokeCode: http.StatusOK}
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		s.mu.Lock()
		s.tokenForms = append(s.tokenForms, r.PostForm)
		status, body := s.tokenStatus, s.tokenBody
		s.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	})
	mux.HandleFunc("/revoke", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		s.mu.Lock()
		s.revokeForms = append(s.revokeForms, r.PostForm)
		code := s.revokeCode
		s.mu.Unlock()
		w.WriteHeader(code)
	})
	s.server = httptest.NewServer(mux)
	t.Cleanup(s.server.Close)
	return s
}

func (s *googleStub) client() *google.Client {
	return google.New(google.Config{
		ClientID:     "cliente-de-prueba.apps.googleusercontent.com",
		ClientSecret: "secreto-de-prueba",
		RedirectURL:  "http://localhost:5173/panel/barberia/google-calendar/callback",
		Endpoint: &oauth2.Endpoint{
			AuthURL:  "https://accounts.example.test/o/oauth2/v2/auth",
			TokenURL: s.server.URL + "/token",
		},
		RevokeURL: s.server.URL + "/revoke",
	})
}

func idToken(claims string) string {
	enc := base64.RawURLEncoding.EncodeToString
	return enc([]byte(`{"alg":"none"}`)) + "." + enc([]byte(claims)) + "." + enc([]byte("firma"))
}

func TestAuthorizationURL_UsesPKCEOfflineAccessMinimalScopesAndNoSecret(t *testing.T) {
	stub := newStub(t)
	raw := stub.client().AuthorizationURL("estado-uno", "verificador-pkce-de-prueba-0123456789-abcdefghij")
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()

	want := map[string]string{
		"response_type":         "code",
		"client_id":             "cliente-de-prueba.apps.googleusercontent.com",
		"redirect_uri":          "http://localhost:5173/panel/barberia/google-calendar/callback",
		"state":                 "estado-uno",
		"access_type":           "offline",
		"prompt":                "consent",
		"code_challenge_method": "S256",
	}
	for key, value := range want {
		if q.Get(key) != value {
			t.Errorf("%s = %q, se esperaba %q", key, q.Get(key), value)
		}
	}
	if q.Get("code_challenge") == "" || strings.Contains(raw, "verificador-pkce") {
		t.Fatal("la URL lleva el desafío S256, nunca el verificador")
	}
	if strings.Contains(raw, "secreto-de-prueba") || q.Has("client_secret") {
		t.Fatal("la URL de consentimiento nunca lleva el secreto del cliente")
	}
	scopes := strings.Fields(q.Get("scope"))
	wantScopes := map[string]bool{"openid": true, "email": true, google.ScopeCalendarEvents: true}
	if len(scopes) != len(wantScopes) {
		t.Fatalf("alcances = %v, solo se piden los mínimos (DEC-102)", scopes)
	}
	for _, s := range scopes {
		if !wantScopes[s] {
			t.Fatalf("alcance inesperado %q", s)
		}
	}
}

func TestExchange_SendsVerifierReadsRefreshTokenAndEmail(t *testing.T) {
	stub := newStub(t)
	stub.tokenBody = `{"access_token":"a","token_type":"Bearer","expires_in":3600,"refresh_token":"1//refresh","scope":"openid email https://www.googleapis.com/auth/calendar.events","id_token":"` +
		idToken(`{"email":"barbero@ejemplo.test"}`) + `"}`

	tokens, err := stub.client().Exchange(context.Background(), "codigo-de-google", "verificador-pkce-de-prueba-0123456789-abcdefghij")
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if tokens.RefreshToken != "1//refresh" || tokens.AccountEmail != "barbero@ejemplo.test" {
		t.Fatalf("tokens = %+v", tokens)
	}
	form := stub.tokenForms[0]
	if form.Get("grant_type") != "authorization_code" || form.Get("code") != "codigo-de-google" ||
		form.Get("code_verifier") != "verificador-pkce-de-prueba-0123456789-abcdefghij" {
		t.Fatalf("el canje debe enviar el código y el verificador PKCE: %v", form)
	}
}

func TestExchange_MissingCalendarScope_IsRejected(t *testing.T) {
	stub := newStub(t)
	stub.tokenBody = `{"access_token":"a","token_type":"Bearer","expires_in":3600,"refresh_token":"1//refresh","scope":"openid email"}`
	_, err := stub.client().Exchange(context.Background(), "c", "v")
	if !errors.Is(err, google.ErrScopeMissing) {
		t.Fatalf("sin el permiso de calendario no hay conexión: %v", err)
	}
}

func TestExchange_FailuresDoNotLeakTheCodeOrSecret(t *testing.T) {
	stub := newStub(t)
	stub.tokenStatus = http.StatusBadRequest
	stub.tokenBody = `{"error":"invalid_grant","error_description":"code CODIGO-SECRETO ya usado"}`

	_, err := stub.client().Exchange(context.Background(), "CODIGO-SECRETO", "v")
	if err == nil {
		t.Fatal("se esperaba un error")
	}
	for _, leak := range []string{"CODIGO-SECRETO", "secreto-de-prueba", "ya usado"} {
		if strings.Contains(err.Error(), leak) {
			t.Fatalf("el error no debe contener %q: %v", leak, err)
		}
	}
	if !strings.Contains(err.Error(), "invalid_grant") {
		t.Fatalf("el código de error estándar sí es útil: %v", err)
	}
}

func TestExchange_WithoutRefreshToken_Fails(t *testing.T) {
	stub := newStub(t)
	stub.tokenBody = `{"access_token":"a","token_type":"Bearer","expires_in":3600,"scope":"https://www.googleapis.com/auth/calendar.events"}`
	if _, err := stub.client().Exchange(context.Background(), "c", "v"); err == nil {
		t.Fatal("sin refresh token no hay conexión posible")
	}
}

func TestRefreshAccessToken_ReturnsAccessTokenAndExpiry(t *testing.T) {
	stub := newStub(t)
	stub.tokenBody = `{"access_token":"access-nuevo","token_type":"Bearer","expires_in":3600}`

	token, expiresAt, err := stub.client().RefreshAccessToken(context.Background(), "1//refresh")
	if err != nil || token != "access-nuevo" || expiresAt.IsZero() {
		t.Fatalf("RefreshAccessToken: %q %v %v", token, expiresAt, err)
	}
	form := stub.tokenForms[0]
	if form.Get("grant_type") != "refresh_token" || form.Get("refresh_token") != "1//refresh" {
		t.Fatalf("formulario inesperado: %v", form)
	}
}

func TestRefreshAccessToken_InvalidGrantIsPermanentOtherErrorsAreTransient(t *testing.T) {
	stub := newStub(t)

	stub.tokenStatus = http.StatusBadRequest
	stub.tokenBody = `{"error":"invalid_grant","error_description":"Token has been expired or revoked."}`
	_, _, err := stub.client().RefreshAccessToken(context.Background(), "1//refresh")
	if !errors.Is(err, googlecalendar.ErrTokenRevoked) {
		t.Fatalf("invalid_grant = permiso revocado: %v", err)
	}

	stub.tokenStatus = http.StatusServiceUnavailable
	stub.tokenBody = `{"error":"backend_error"}`
	_, _, err = stub.client().RefreshAccessToken(context.Background(), "1//refresh")
	if err == nil || errors.Is(err, googlecalendar.ErrTokenRevoked) {
		t.Fatalf("un 503 es transitorio, no una revocación: %v", err)
	}
	if strings.Contains(err.Error(), "1//refresh") {
		t.Fatalf("el error no debe contener el token: %v", err)
	}
}

func TestRevoke_SendsTheTokenAndTreatsAlreadyInvalidAsSuccess(t *testing.T) {
	stub := newStub(t)
	if err := stub.client().Revoke(context.Background(), "1//refresh"); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	if stub.revokeForms[0].Get("token") != "1//refresh" {
		t.Fatalf("debe enviar el token: %v", stub.revokeForms[0])
	}

	stub.revokeCode = http.StatusBadRequest // Google: token ya inválido
	if err := stub.client().Revoke(context.Background(), "1//viejo"); err != nil {
		t.Fatalf("un token que Google ya no reconoce es el estado buscado: %v", err)
	}

	stub.revokeCode = http.StatusInternalServerError
	err := stub.client().Revoke(context.Background(), "1//otro")
	if err == nil || strings.Contains(err.Error(), "1//otro") {
		t.Fatalf("un 500 es un error sin el token: %v", err)
	}
}
