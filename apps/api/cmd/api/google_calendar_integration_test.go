// Pruebas de integración de la conexión con Google Calendar (issue #323,
// DEC-099, DEC-102) contra el router REAL (buildRouter), PostgreSQL real con dos
// barberías y un Google FALSO: ninguna prueba toca la red real. Requieren
// database/testdata/dos_barberias.sql y hu005_credenciales_sesiones.sql.
package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"golang.org/x/oauth2"

	"system-barbershop/internal/modules/googlecalendar"
	googlecalendargoogle "system-barbershop/internal/modules/googlecalendar/google"
	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
)

const googleStubRefreshToken = "1//refresh-token-de-prueba-no-es-real"

type googleStub struct {
	mu          sync.Mutex
	tokenForms  []url.Values
	revoked     []string
	tokenStatus int
	tokenBody   string
	server      *httptest.Server
}

func newGoogleStub(t *testing.T) *googleStub {
	t.Helper()
	enc := base64.RawURLEncoding.EncodeToString
	idToken := enc([]byte(`{"alg":"none"}`)) + "." + enc([]byte(`{"email":"barbero@ejemplo.test"}`)) + "." + enc([]byte("firma"))
	s := &googleStub{
		tokenStatus: http.StatusOK,
		tokenBody: `{"access_token":"a","token_type":"Bearer","expires_in":3600,"refresh_token":"` + googleStubRefreshToken +
			`","scope":"openid email https://www.googleapis.com/auth/calendar.events","id_token":"` + idToken + `"}`,
	}
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
		s.revoked = append(s.revoked, r.PostForm.Get("token"))
		s.mu.Unlock()
		w.WriteHeader(http.StatusOK)
	})
	s.server = httptest.NewServer(mux)
	t.Cleanup(s.server.Close)
	return s
}

func (s *googleStub) revokedTokens() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.revoked...)
}

func googleCalendarRouter(t *testing.T, db *database.DB, stub *googleStub) http.Handler {
	t.Helper()
	original := newGoogleCalendarProvider
	newGoogleCalendarProvider = func(cfg config.Config) googlecalendar.OAuthProvider {
		return googlecalendargoogle.New(googlecalendargoogle.Config{
			ClientID:     cfg.GoogleCalendarClientID,
			ClientSecret: cfg.GoogleCalendarClientSecret,
			RedirectURL:  cfg.GoogleCalendarRedirectURI,
			Endpoint: &oauth2.Endpoint{
				AuthURL:  "https://accounts.example.test/o/oauth2/v2/auth",
				TokenURL: stub.server.URL + "/token",
			},
			RevokeURL: stub.server.URL + "/revoke",
		})
	}
	t.Cleanup(func() { newGoogleCalendarProvider = original })

	cfg := testRouterConfig()
	cfg.GoogleCalendarClientID = "cliente-de-prueba.apps.googleusercontent.com"
	cfg.GoogleCalendarClientSecret = "secreto-de-prueba"
	cfg.GoogleCalendarRedirectURI = "http://localhost:5173/panel/barberia/google-calendar/callback"
	cfg.GoogleCalendarTokenEncryptionKey = base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{5}, 32))
	cfg.GoogleCalendarTokenKeyID = "v1"

	router, err := buildRouter(db, discardLogger(), cfg)
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	return router
}

func doGCal(router http.Handler, raw, method, path, body string) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body != "" {
		reader = bytes.NewReader([]byte(body))
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, "/api/v1/private/integrations/google-calendar"+path, reader)
	req.Header.Set("Content-Type", "application/json")
	if raw != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: raw})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func gcalJSON(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v (%s)", err, rec.Body.String())
	}
	return body
}

// startConnection inicia el flujo y devuelve el state que Google devolvería.
func startConnection(t *testing.T, router http.Handler, raw string) string {
	t.Helper()
	rec := doGCal(router, raw, http.MethodPost, "/connect", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("connect: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	authURL, err := url.Parse(gcalJSON(t, rec)["authorizationUrl"].(string))
	if err != nil {
		t.Fatal(err)
	}
	q := authURL.Query()
	if q.Get("code_challenge_method") != "S256" || q.Get("access_type") != "offline" ||
		q.Get("prompt") != "consent" || q.Get("state") == "" {
		t.Fatalf("la URL de consentimiento debe usar PKCE S256, acceso offline y state: %s", authURL)
	}
	if strings.Contains(authURL.String(), "secreto-de-prueba") {
		t.Fatal("la URL nunca lleva el secreto del cliente")
	}
	return q.Get("state")
}

func linkMe(t *testing.T, router http.Handler, raw, barberID string) {
	t.Helper()
	rec := doMyBarberRequest(router, raw, http.MethodPut, `{"barberId":"`+barberID+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("vincular: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGoogleCalendar_HTTP_ConnectJourney_EncryptedAtRestAndDisconnect(t *testing.T) {
	db := setupTestDB(t)
	stub := newGoogleStub(t)
	router := googleCalendarRouter(t, db, stub)
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "gcal-journey")
	t.Cleanup(func() {
		doGCal(router, raw, http.MethodDelete, "", "")
		doMyBarberRequest(router, raw, http.MethodDelete, "")
	})
	barber := createBarberVia(t, router, raw, "Gcal "+uniqueToken(t, "n"))

	// Sin barbero vinculado: el estado lo dice y conectar responde 409.
	rec := doGCal(router, raw, http.MethodGet, "", "")
	if body := gcalJSON(t, rec); rec.Code != http.StatusOK || body["barberLinked"] != false || body["enabled"] != true {
		t.Fatalf("sin vínculo: %d %v", rec.Code, body)
	}
	if rec := doGCal(router, raw, http.MethodPost, "/connect", ""); rec.Code != http.StatusConflict {
		t.Fatalf("conectar sin barbero: expected 409, got %d", rec.Code)
	}

	linkMe(t, router, raw, barber.ID)
	state := startConnection(t, router, raw)

	rec = doGCal(router, raw, http.MethodPost, "/callback", `{"state":"`+state+`","code":"codigo-de-google"}`)
	if rec.Code != http.StatusOK || gcalJSON(t, rec)["result"] != "connected" {
		t.Fatalf("callback: %d %s", rec.Code, rec.Body.String())
	}
	form := stub.tokenForms[0]
	if form.Get("grant_type") != "authorization_code" || form.Get("code") != "codigo-de-google" || form.Get("code_verifier") == "" {
		t.Fatalf("el canje debe llevar el código y el verificador PKCE: %v", form)
	}

	rec = doGCal(router, raw, http.MethodGet, "", "")
	body := gcalJSON(t, rec)
	if body["status"] != "connected" || body["accountEmail"] != "barbero@ejemplo.test" {
		t.Fatalf("estado conectado: %v", body)
	}
	if strings.Contains(rec.Body.String(), googleStubRefreshToken) {
		t.Fatal("ninguna respuesta contiene el refresh token")
	}

	// En la base solo hay texto cifrado, nunca el token en claro.
	var ciphertext []byte
	var keyID string
	if err := db.InTenantTx(context.Background(), database.BarbershopID(shopA), func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx,
			`SELECT refresh_token_ciphertext, token_key_id FROM google_calendar_connection
			  WHERE barbershop_id = $1 AND barber_id = $2`, shopA, barber.ID).Scan(&ciphertext, &keyID)
	}); err != nil {
		t.Fatal(err)
	}
	if len(ciphertext) == 0 || keyID != "v1" || bytes.Contains(ciphertext, []byte(googleStubRefreshToken)) {
		t.Fatalf("el refresh token debe persistirse cifrado con key_id (len=%d key=%q)", len(ciphertext), keyID)
	}

	rec = doGCal(router, raw, http.MethodPatch, "", `{"reminderMinutes":45}`)
	if rec.Code != http.StatusOK || gcalJSON(t, rec)["reminderMinutes"] != float64(45) {
		t.Fatalf("recordatorio: %d %s", rec.Code, rec.Body.String())
	}
	if rec := doGCal(router, raw, http.MethodPatch, "", `{"reminderMinutes":40321}`); rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("fuera de rango: expected 422, got %d", rec.Code)
	}

	if rec := doGCal(router, raw, http.MethodDelete, "", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("desconectar: expected 204, got %d", rec.Code)
	}
	if got := stub.revokedTokens(); len(got) != 1 || got[0] != googleStubRefreshToken {
		t.Fatalf("desconectar debe revocar el token en Google: %v", got)
	}
	rec = doGCal(router, raw, http.MethodGet, "", "")
	body = gcalJSON(t, rec)
	if body["status"] != "disconnected" || body["accountEmail"] != nil {
		t.Fatalf("desconectado: %v", body)
	}
	var left int
	_ = db.InTenantTx(context.Background(), database.BarbershopID(shopA), func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx,
			`SELECT count(*) FROM google_calendar_connection
			  WHERE barber_id = $1 AND refresh_token_ciphertext IS NOT NULL`, barber.ID).Scan(&left)
	})
	if left != 0 {
		t.Fatal("desconectar borra las credenciales de la base")
	}
	if rec := doGCal(router, raw, http.MethodDelete, "", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("desconectar de nuevo es idempotente, got %d", rec.Code)
	}
}

func TestGoogleCalendar_HTTP_StateIsBoundToSessionTenantAndSingleUse(t *testing.T) {
	db := setupTestDB(t)
	stub := newGoogleStub(t)
	router := googleCalendarRouter(t, db, stub)
	rawA := createSessionCookie(t, db, shopA, staffUserActiveA, "gcal-state-a")
	rawA2 := createSessionCookie(t, db, shopA, staffUserActiveA, "gcal-state-a2") // misma persona, otra sesión
	rawB := createSessionCookie(t, db, shopB, staffUserActiveB, "gcal-state-b")
	t.Cleanup(func() {
		doGCal(router, rawA, http.MethodDelete, "", "")
		doMyBarberRequest(router, rawA, http.MethodDelete, "")
		doMyBarberRequest(router, rawB, http.MethodDelete, "")
	})
	linkMe(t, router, rawA, createBarberVia(t, router, rawA, "Estado A "+uniqueToken(t, "n")).ID)
	linkMe(t, router, rawB, createBarberVia(t, router, rawB, "Estado B "+uniqueToken(t, "n")).ID)

	state := startConnection(t, router, rawA)
	callback := `{"state":"` + state + `","code":"codigo"}`

	// Otra sesión (aun de la misma persona) y otra barbería no pueden usarlo.
	for name, cookie := range map[string]string{"otra sesión": rawA2, "otra barbería": rawB, "sin cookie": ""} {
		rec := doGCal(router, cookie, http.MethodPost, "/callback", callback)
		want := http.StatusBadRequest
		if cookie == "" {
			want = http.StatusUnauthorized
		}
		if rec.Code != want {
			t.Fatalf("%s: expected %d, got %d: %s", name, want, rec.Code, rec.Body.String())
		}
	}
	if len(stub.tokenForms) != 0 {
		t.Fatal("un state ajeno nunca debe llegar a Google")
	}

	// Un intento de otra sesión de la MISMA barbería ya consumió ese state (se
	// consume antes de comprobar a quién pertenece): ni siquiera su dueño puede
	// reutilizarlo y debe iniciar de nuevo. Un state filtrado queda inservible.
	if rec := doGCal(router, rawA, http.MethodPost, "/callback", callback); rec.Code != http.StatusBadRequest {
		t.Fatalf("un state ya gastado por un intento ajeno no sirve ni a su dueño: %d", rec.Code)
	}

	// Con un state nuevo, el dueño legítimo sí puede; y una segunda vez ya no.
	state = startConnection(t, router, rawA)
	callback = `{"state":"` + state + `","code":"codigo"}`
	if rec := doGCal(router, rawA, http.MethodPost, "/callback", callback); rec.Code != http.StatusOK {
		t.Fatalf("la sesión legítima debe poder completar: %d %s", rec.Code, rec.Body.String())
	}
	if rec := doGCal(router, rawA, http.MethodPost, "/callback", callback); rec.Code != http.StatusBadRequest {
		t.Fatalf("un state usado no sirve de nuevo: %d", rec.Code)
	}

	// La conexión de A no es visible desde B.
	if body := gcalJSON(t, doGCal(router, rawB, http.MethodGet, "", "")); body["status"] != "not_connected" {
		t.Fatalf("B no debe ver la conexión de A: %v", body)
	}
}

func TestGoogleCalendar_HTTP_DeniedAndFailedExchangeConnectNothing(t *testing.T) {
	db := setupTestDB(t)
	stub := newGoogleStub(t)
	router := googleCalendarRouter(t, db, stub)
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "gcal-denied")
	t.Cleanup(func() {
		doGCal(router, raw, http.MethodDelete, "", "")
		doMyBarberRequest(router, raw, http.MethodDelete, "")
	})
	linkMe(t, router, raw, createBarberVia(t, router, raw, "Rechazo "+uniqueToken(t, "n")).ID)

	state := startConnection(t, router, raw)
	rec := doGCal(router, raw, http.MethodPost, "/callback", `{"state":"`+state+`","error":"access_denied"}`)
	if rec.Code != http.StatusOK || gcalJSON(t, rec)["result"] != "denied" {
		t.Fatalf("rechazo: %d %s", rec.Code, rec.Body.String())
	}

	stub.mu.Lock()
	stub.tokenStatus, stub.tokenBody = http.StatusBadRequest, `{"error":"invalid_grant"}`
	stub.mu.Unlock()
	state = startConnection(t, router, raw)
	rec = doGCal(router, raw, http.MethodPost, "/callback", `{"state":"`+state+`","code":"codigo-malo"}`)
	if rec.Code != http.StatusOK || gcalJSON(t, rec)["result"] != "failed" {
		t.Fatalf("canje fallido: %d %s", rec.Code, rec.Body.String())
	}
	if body := gcalJSON(t, doGCal(router, raw, http.MethodGet, "", "")); body["status"] != "not_connected" {
		t.Fatalf("ni el rechazo ni el canje fallido conectan: %v", body)
	}
}

// DEC-100: si el usuario cambia o quita su vínculo, la conexión del barbero
// anterior se desconecta y se revoca, para que nadie herede una cuenta ajena.
func TestGoogleCalendar_HTTP_ChangingTheBarberLinkDisconnectsThePreviousBarber(t *testing.T) {
	db := setupTestDB(t)
	stub := newGoogleStub(t)
	router := googleCalendarRouter(t, db, stub)
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "gcal-link-change")
	t.Cleanup(func() {
		doGCal(router, raw, http.MethodDelete, "", "")
		doMyBarberRequest(router, raw, http.MethodDelete, "")
	})
	first := createBarberVia(t, router, raw, "Primero "+uniqueToken(t, "n"))
	second := createBarberVia(t, router, raw, "Segundo "+uniqueToken(t, "n"))

	linkMe(t, router, raw, first.ID)
	state := startConnection(t, router, raw)
	if rec := doGCal(router, raw, http.MethodPost, "/callback", `{"state":"`+state+`","code":"c"}`); rec.Code != http.StatusOK {
		t.Fatalf("conectar: %d", rec.Code)
	}

	// Cambiar a otro barbero desconecta al primero.
	linkMe(t, router, raw, second.ID)
	if got := stub.revokedTokens(); len(got) != 1 || got[0] != googleStubRefreshToken {
		t.Fatalf("cambiar de barbero debe revocar el token del anterior: %v", got)
	}
	if body := gcalJSON(t, doGCal(router, raw, http.MethodGet, "", "")); body["status"] != "not_connected" {
		t.Fatalf("el nuevo barbero no hereda la conexión del anterior: %v", body)
	}
	var status string
	if err := db.InTenantTx(context.Background(), database.BarbershopID(shopA), func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx, `SELECT status FROM google_calendar_connection WHERE barber_id = $1`, first.ID).Scan(&status)
	}); err != nil || status != "disconnected" {
		t.Fatalf("la conexión del primer barbero debe quedar desconectada: %q %v", status, err)
	}

	// Quitar el vínculo también: reconecta el segundo y luego sin vínculo.
	state = startConnection(t, router, raw)
	if rec := doGCal(router, raw, http.MethodPost, "/callback", `{"state":"`+state+`","code":"c"}`); rec.Code != http.StatusOK {
		t.Fatalf("reconectar: %d", rec.Code)
	}
	doMyBarberRequest(router, raw, http.MethodDelete, "")
	if got := stub.revokedTokens(); len(got) != 2 {
		t.Fatalf("quitar el vínculo también revoca: %v", got)
	}
}

func TestGoogleCalendar_HTTP_DisabledIntegrationKeepsTheProductWorking(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig()) // sin GOOGLE_CALENDAR_*
	if err != nil {
		t.Fatalf("buildRouter debe arrancar sin credenciales de Google: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "gcal-disabled")

	rec := doGCal(router, raw, http.MethodGet, "", "")
	if body := gcalJSON(t, rec); rec.Code != http.StatusOK || body["enabled"] != false {
		t.Fatalf("desactivada: %d %v", rec.Code, body)
	}
	if rec := doGCal(router, raw, http.MethodPost, "/connect", ""); rec.Code != http.StatusConflict {
		t.Fatalf("conectar desactivada: expected 409, got %d", rec.Code)
	}
	// El resto del producto sigue funcionando.
	if rec := doListBarbersRequest(router, raw, "limit=1"); rec.Code != http.StatusOK {
		t.Fatalf("la lista de barberos debe seguir funcionando: %d", rec.Code)
	}
}

func TestGoogleCalendar_HTTP_NoCookie_Returns401(t *testing.T) {
	db := setupTestDB(t)
	router := googleCalendarRouter(t, db, newGoogleStub(t))
	for _, c := range []struct{ method, path string }{
		{http.MethodGet, ""}, {http.MethodPatch, ""}, {http.MethodDelete, ""},
		{http.MethodPost, "/connect"}, {http.MethodPost, "/callback"},
	} {
		if rec := doGCal(router, "", c.method, c.path, `{}`); rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s sin cookie: expected 401, got %d", c.method, c.path, rec.Code)
		}
	}
}
