package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/modules/googlecalendar"
	"system-barbershop/internal/modules/googlecalendar/httpapi"
)

const (
	shop    = "11111111-1111-1111-1111-111111111111"
	user    = "aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2"
	barber  = "c0000001-0000-4000-8000-000000000001"
	session = "5e550001-0000-4000-8000-000000000001"
)

type memRepo struct {
	mu     sync.Mutex
	linked bool
	conn   *googlecalendar.Connection
	states map[string]googlecalendar.OAuthState
	used   map[string]bool
}

func newRepo(linked bool) *memRepo {
	return &memRepo{linked: linked, states: map[string]googlecalendar.OAuthState{}, used: map[string]bool{}}
}

func (r *memRepo) GetConnection(context.Context, string, string) (googlecalendar.Connection, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.conn == nil {
		return googlecalendar.Connection{}, false, nil
	}
	return *r.conn, true, nil
}

func (r *memRepo) SaveConnected(_ context.Context, s, b string, d googlecalendar.ConnectedData) (googlecalendar.Connection, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	at := d.At
	r.conn = &googlecalendar.Connection{BarbershopID: s, BarberID: b, Status: googlecalendar.StatusConnected,
		AccountEmail: d.AccountEmail, CalendarID: "primary", Credentials: d.Credentials, KeyID: d.KeyID, ConnectedAt: &at}
	return *r.conn, nil
}

func (r *memRepo) MarkStatus(_ context.Context, _, _ string, status googlecalendar.Status, _ string, _ time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.conn == nil {
		return false, nil
	}
	r.conn.Status = status
	if status != googlecalendar.StatusError {
		r.conn.Credentials, r.conn.KeyID = nil, ""
	}
	return true, nil
}

func (r *memRepo) SetReminder(_ context.Context, _, _ string, m *int) (googlecalendar.Connection, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.conn == nil || r.conn.Status == googlecalendar.StatusDisconnected {
		return googlecalendar.Connection{}, false, nil
	}
	r.conn.ReminderMinutes = m
	return *r.conn, true, nil
}

func (r *memRepo) CreateState(_ context.Context, s googlecalendar.OAuthState, _ time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.states[s.StateHash] = s
	return nil
}

func (r *memRepo) ConsumeState(_ context.Context, _, hash string, now time.Time) (googlecalendar.OAuthState, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.states[hash]
	if !ok || r.used[hash] || !s.ExpiresAt.After(now) {
		return googlecalendar.OAuthState{}, false, nil
	}
	r.used[hash] = true
	return s, true, nil
}

func (r *memRepo) BarberOfUser(context.Context, string, string) (string, bool, error) {
	return barber, r.linked, nil
}

type stubProvider struct{ tokens googlecalendar.Tokens }

func (p stubProvider) AuthorizationURL(state, _ string) string {
	return "https://accounts.example.test/auth?state=" + state
}
func (p stubProvider) Exchange(context.Context, string, string) (googlecalendar.Tokens, error) {
	return p.tokens, nil
}
func (p stubProvider) RefreshAccessToken(context.Context, string) (string, time.Time, error) {
	return "access", time.Now().Add(time.Hour), nil
}
func (p stubProvider) Revoke(context.Context, string) error { return nil }

type fixedClock struct{}

func (fixedClock) Now() time.Time { return time.Date(2026, 10, 9, 15, 0, 0, 0, time.UTC) }

func newHandlers(t *testing.T, repo *memRepo, enabled bool) *httpapi.Handlers {
	t.Helper()
	deps := googlecalendar.Deps{Repo: repo, Clock: fixedClock{}}
	if enabled {
		cipher, err := googlecalendar.NewCipher("v1", map[string][]byte{"v1": bytes.Repeat([]byte{3}, 32)})
		if err != nil {
			t.Fatal(err)
		}
		deps.Cipher = cipher
		deps.Provider = stubProvider{tokens: googlecalendar.Tokens{RefreshToken: "1//refresh-secreto", AccountEmail: "barbero@ejemplo.test"}}
	}
	return httpapi.New(googlecalendar.NewService(deps))
}

func request(method, body string) *http.Request {
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, "/api/v1/private/integrations/google-calendar", nil)
	} else {
		r = httptest.NewRequest(method, "/api/v1/private/integrations/google-calendar", strings.NewReader(body))
	}
	return r.WithContext(auth.ContextWithPrincipal(r.Context(),
		auth.Principal{SessionID: session, StaffUserID: user, BarbershopID: shop}))
}

func do(h http.HandlerFunc, r *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h(rec, r)
	return rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v (%s)", err, rec.Body.String())
	}
	return body
}

func TestGet_ReportsNotConnectedThenConnectedAndNeverLeaksCredentials(t *testing.T) {
	repo := newRepo(true)
	h := newHandlers(t, repo, true)

	rec := do(h.Get, request(http.MethodGet, ""))
	body := decode(t, rec)
	if rec.Code != http.StatusOK || body["status"] != "not_connected" || body["enabled"] != true || body["barberLinked"] != true {
		t.Fatalf("sin conexión: %d %v", rec.Code, body)
	}

	// Conecta de punta a punta por la API y vuelve a leer.
	rec = do(h.Connect, request(http.MethodPost, ""))
	state := strings.TrimPrefix(decode(t, rec)["authorizationUrl"].(string), "https://accounts.example.test/auth?state=")
	rec = do(h.Callback, request(http.MethodPost, `{"state":"`+state+`","code":"codigo-de-google"}`))
	if rec.Code != http.StatusOK || decode(t, rec)["result"] != "connected" {
		t.Fatalf("callback: %d %s", rec.Code, rec.Body.String())
	}

	rec = do(h.Get, request(http.MethodGet, ""))
	body = decode(t, rec)
	if body["status"] != "connected" || body["accountEmail"] != "barbero@ejemplo.test" {
		t.Fatalf("conectado: %v", body)
	}
	for _, forbidden := range []string{"refresh", "token", "secret", "credential", "ciphertext", "keyId", "state", "code"} {
		for key := range body {
			if strings.Contains(strings.ToLower(key), forbidden) {
				t.Fatalf("la respuesta no debe exponer %q (campo %q)", forbidden, key)
			}
		}
	}
	if strings.Contains(rec.Body.String(), "1//refresh-secreto") {
		t.Fatal("la respuesta jamás contiene el refresh token")
	}
}

func TestGet_WithoutBarberOrWithIntegrationOff(t *testing.T) {
	rec := do(newHandlers(t, newRepo(false), true).Get, request(http.MethodGet, ""))
	if body := decode(t, rec); body["barberLinked"] != false || body["status"] != "not_connected" {
		t.Fatalf("sin barbero: %v", body)
	}
	rec = do(newHandlers(t, newRepo(true), false).Get, request(http.MethodGet, ""))
	if body := decode(t, rec); body["enabled"] != false {
		t.Fatalf("integración desactivada: %v", body)
	}
}

func TestConnect_DisabledOrWithoutBarber_Is409(t *testing.T) {
	for name, h := range map[string]*httpapi.Handlers{
		"desactivada": newHandlers(t, newRepo(true), false),
		"sin barbero": newHandlers(t, newRepo(false), true),
	} {
		t.Run(name, func(t *testing.T) {
			rec := do(h.Connect, request(http.MethodPost, ""))
			if rec.Code != http.StatusConflict || decode(t, rec)["code"] != "invalid-state" {
				t.Fatalf("expected 409 invalid-state, got %d %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCallback_StrictBodyAndUniformInvalidStateError(t *testing.T) {
	h := newHandlers(t, newRepo(true), true)

	for name, body := range map[string]string{
		"JSON inválido":       `{`,
		"campo desconocido":   `{"state":"x","barberId":"b"}`,
		"identificador ajeno": `{"state":"x","staffUserId":"u"}`,
		"dos documentos":      `{"state":"x"}{}`,
	} {
		t.Run(name, func(t *testing.T) {
			rec := do(h.Callback, request(http.MethodPost, body))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}

	rec := do(h.Callback, request(http.MethodPost, `{"state":"inventado","code":"c"}`))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("un state inexistente es 400, got %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "inventado") {
		t.Fatal("el error nunca repite el state recibido")
	}
}

func TestCallback_DeniedByTheBarber(t *testing.T) {
	repo := newRepo(true)
	h := newHandlers(t, repo, true)
	rec := do(h.Connect, request(http.MethodPost, ""))
	state := strings.TrimPrefix(decode(t, rec)["authorizationUrl"].(string), "https://accounts.example.test/auth?state=")

	rec = do(h.Callback, request(http.MethodPost, `{"state":"`+state+`","error":"access_denied"}`))
	if rec.Code != http.StatusOK || decode(t, rec)["result"] != "denied" {
		t.Fatalf("rechazo: %d %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateReminder_RequiredFieldNullRangeAndNoConnection(t *testing.T) {
	repo := newRepo(true)
	h := newHandlers(t, repo, true)

	// Sin conexión: 404.
	if rec := do(h.UpdateReminder, request(http.MethodPatch, `{"reminderMinutes":30}`)); rec.Code != http.StatusNotFound {
		t.Fatalf("sin conexión se espera 404, got %d", rec.Code)
	}

	rec := do(h.Connect, request(http.MethodPost, ""))
	state := strings.TrimPrefix(decode(t, rec)["authorizationUrl"].(string), "https://accounts.example.test/auth?state=")
	do(h.Callback, request(http.MethodPost, `{"state":"`+state+`","code":"c"}`))

	rec = do(h.UpdateReminder, request(http.MethodPatch, `{"reminderMinutes":30}`))
	if rec.Code != http.StatusOK || decode(t, rec)["reminderMinutes"] != float64(30) {
		t.Fatalf("guardar 30: %d %s", rec.Code, rec.Body.String())
	}
	rec = do(h.UpdateReminder, request(http.MethodPatch, `{"reminderMinutes":null}`))
	if rec.Code != http.StatusOK || decode(t, rec)["reminderMinutes"] != nil {
		t.Fatalf("null vuelve a los predeterminados: %d %s", rec.Code, rec.Body.String())
	}
	for name, body := range map[string]string{"ausente": `{}`, "texto": `{"reminderMinutes":"30"}`, "extra": `{"reminderMinutes":5,"status":"connected"}`} {
		if rec := do(h.UpdateReminder, request(http.MethodPatch, body)); rec.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d", name, rec.Code)
		}
	}
	for _, body := range []string{`{"reminderMinutes":-1}`, `{"reminderMinutes":40321}`} {
		if rec := do(h.UpdateReminder, request(http.MethodPatch, body)); rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("%s: expected 422, got %d", body, rec.Code)
		}
	}
}

func TestDisconnect_Is204AndIdempotent(t *testing.T) {
	h := newHandlers(t, newRepo(true), true)
	for range 2 {
		rec := do(h.Disconnect, request(http.MethodDelete, ""))
		if rec.Code != http.StatusNoContent || rec.Body.Len() != 0 {
			t.Fatalf("expected an empty 204, got %d %q", rec.Code, rec.Body.String())
		}
	}
}

// --- Contrato ---------------------------------------------------------------

type operation struct {
	OperationID string         `yaml:"operationId"`
	Security    []any          `yaml:"security"`
	Responses   map[string]any `yaml:"responses"`
}

type pathsFile struct {
	Item struct {
		Get    operation `yaml:"get"`
		Patch  operation `yaml:"patch"`
		Delete operation `yaml:"delete"`
	} `yaml:"/private/integrations/google-calendar"`
	Connect struct {
		Post operation `yaml:"post"`
	} `yaml:"/private/integrations/google-calendar/connect"`
	Callback struct {
		Post operation `yaml:"post"`
	} `yaml:"/private/integrations/google-calendar/callback"`
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "redocly.yaml")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("no se encontró redocly.yaml")
		}
		dir = parent
	}
}

func loadYAML[T any](t *testing.T, rel string) T {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	var v T
	if err := yaml.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	return v
}

func requireOperation(t *testing.T, op operation, id string, statuses ...string) {
	t.Helper()
	if op.OperationID != id {
		t.Fatalf("operationId = %q, se esperaba %q", op.OperationID, id)
	}
	if len(op.Security) != 1 {
		t.Fatalf("%s debe exigir exactamente SessionCookie", id)
	}
	if len(op.Responses) != len(statuses) {
		t.Errorf("%s documenta %d respuestas, se esperaban %v", id, len(op.Responses), statuses)
	}
	for _, s := range statuses {
		if _, ok := op.Responses[s]; !ok {
			t.Errorf("%s no documenta la respuesta %s que el handler produce", id, s)
		}
	}
}

func TestContract_OperationsAndResponses(t *testing.T) {
	doc := loadYAML[pathsFile](t, "api/openapi/paths/google-calendar.yaml")
	requireOperation(t, doc.Item.Get, "getGoogleCalendarConnection", "200", "401", "500")
	requireOperation(t, doc.Item.Patch, "updateGoogleCalendarReminder", "200", "400", "401", "404", "422", "500")
	requireOperation(t, doc.Item.Delete, "disconnectGoogleCalendar", "204", "401", "500")
	requireOperation(t, doc.Connect.Post, "startGoogleCalendarConnection", "200", "401", "409", "500")
	requireOperation(t, doc.Callback.Post, "completeGoogleCalendarConnection", "200", "400", "401", "409", "500")
}

type schemaDoc struct {
	Properties map[string]any `yaml:"properties"`
	Required   []string       `yaml:"required"`
}

func TestContract_SchemasMatchTheDTOsAndNeverCarrySecrets(t *testing.T) {
	cases := map[string][]string{
		"GoogleCalendarConnectionResponse":        {"enabled", "barberLinked", "status", "accountEmail", "reminderMinutes", "connectedAt", "lastSyncedAt"},
		"GoogleCalendarAuthorizationResponse":     {"authorizationUrl"},
		"CompleteGoogleCalendarConnectionRequest": {"state", "code", "error"},
		"GoogleCalendarCallbackResponse":          {"result"},
		"UpdateGoogleCalendarReminderRequest":     {"reminderMinutes"},
	}
	for name, want := range cases {
		schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/"+name+".yaml")
		if len(schema.Properties) != len(want) {
			t.Errorf("%s declara %d propiedades, se esperaban %v", name, len(schema.Properties), want)
		}
		for _, prop := range want {
			if _, ok := schema.Properties[prop]; !ok {
				t.Errorf("%s no declara %q", name, prop)
			}
		}
		for prop := range schema.Properties {
			lower := strings.ToLower(prop)
			for _, forbidden := range []string{"barberid", "staffuserid", "barbershopid", "refreshtoken", "accesstoken", "secret", "ciphertext"} {
				if lower == forbidden {
					t.Errorf("%s nunca debe declarar %q", name, prop)
				}
			}
		}
	}
}

func TestContract_OpenAPIRegistersTheIntegrationPaths(t *testing.T) {
	type root struct {
		Paths map[string]any `yaml:"paths"`
	}
	doc := loadYAML[root](t, "api/openapi/openapi.yaml")
	for _, p := range []string{
		"/private/integrations/google-calendar",
		"/private/integrations/google-calendar/connect",
		"/private/integrations/google-calendar/callback",
	} {
		if _, ok := doc.Paths[p]; !ok {
			t.Errorf("openapi.yaml no registra %s", p)
		}
	}
}
