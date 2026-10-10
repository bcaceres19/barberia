package notification_test

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
	"sync/atomic"
	"testing"
	"time"

	"system-barbershop/internal/modules/notification"
)

const (
	testToken     = "token-secreto-de-prueba"
	testRecipient = "+573001234567"
	testCode      = "583921"
)

// newMetaSenderAgainstTestServer construye un MetaWhatsAppSender cuyo host
// real de Meta se sustituye por un httptest.Server: el adaptador SIEMPRE
// llama a metaGraphBaseURL, así que la sustitución del host se hace
// interceptando el Transport del *http.Client, nunca configurando otro
// host en el adaptador (que no lo expone, a propósito).
func newMetaSenderAgainstTestServer(t *testing.T, srv *httptest.Server, cfg notification.MetaWhatsAppConfig) notification.MetaWhatsAppSender {
	t.Helper()
	client := srv.Client()
	client.Transport = redirectTransport{target: srv.URL}
	return notification.NewMetaWhatsAppSender(cfg, client)
}

// redirectTransport reescribe cualquier solicitud hacia target,
// preservando método/cuerpo/cabeceras, para no depender de un host real.
type redirectTransport struct{ target string }

func (t redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	newURL := t.target + req.URL.Path
	newReq := req.Clone(req.Context())
	u, err := req.URL.Parse(newURL)
	if err != nil {
		return nil, err
	}
	newReq.URL = u
	newReq.Host = u.Host
	return http.DefaultTransport.RoundTrip(newReq)
}

func templateConfig() notification.MetaWhatsAppConfig {
	return notification.MetaWhatsAppConfig{
		APIVersion: "v24.0", PhoneNumberID: "1315949954943476", AccessToken: testToken,
		Mode: notification.MetaModeTemplate, TemplateName: "codigo_verificacion", LanguageCode: "es",
	}
}

func developmentConfig() notification.MetaWhatsAppConfig {
	return notification.MetaWhatsAppConfig{
		APIVersion: "v24.0", PhoneNumberID: "1315949954943476", AccessToken: testToken,
		Mode: notification.MetaModeDevelopment, TestRecipients: []string{testRecipient},
	}
}

// capturedRequest recoge lo que el adaptador envió al servidor falso.
type capturedRequest struct {
	auth, contentType, method, path string
	body                            map[string]any
	calls                           atomic.Int32
}

func acceptingServer(t *testing.T, c *capturedRequest) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.calls.Add(1)
		c.auth, c.contentType, c.method, c.path = r.Header.Get("Authorization"), r.Header.Get("Content-Type"), r.Method, r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&c.body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"messaging_product":"whatsapp","contacts":[{"wa_id":"573001234567"}],"messages":[{"id":"wamid.HBgM"}]}`))
	}))
}

func TestMetaSender_DevelopmentMode_PostsFreeTextWithExactCode(t *testing.T) {
	var got capturedRequest
	srv := acceptingServer(t, &got)
	defer srv.Close()
	sender := newMetaSenderAgainstTestServer(t, srv, developmentConfig())

	res, err := sender.SendOTP(context.Background(), testRecipient, testCode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.MessageID != "wamid.HBgM" {
		t.Fatalf("expected the wamid from Meta, got %q", res.MessageID)
	}
	if got.method != http.MethodPost || got.path != "/v24.0/1315949954943476/messages" {
		t.Fatalf("unexpected request: %s %s", got.method, got.path)
	}
	if got.auth != "Bearer "+testToken || got.contentType != "application/json" {
		t.Fatalf("unexpected headers: auth=%q content-type=%q", got.auth, got.contentType)
	}
	if got.body["messaging_product"] != "whatsapp" || got.body["recipient_type"] != "individual" ||
		got.body["to"] != "573001234567" || got.body["type"] != "text" {
		t.Fatalf("unexpected payload: %v", got.body)
	}
	text := got.body["text"].(map[string]any)
	if text["preview_url"] != false {
		t.Fatalf("expected preview_url=false, got %v", text["preview_url"])
	}
	want := "NAVA — Prueba de integración\n\nTu código de prueba es: 583921\n\nEste mensaje es una simulación de desarrollo."
	if text["body"] != want {
		t.Fatalf("unexpected body:\n%q", text["body"])
	}
	if _, hasTemplate := got.body["template"]; hasTemplate {
		t.Fatal("development mode must not send a template")
	}
}

func TestMetaSender_TemplateMode_PostsAuthenticationTemplateWithSameCodeInBodyAndButton(t *testing.T) {
	var got capturedRequest
	srv := acceptingServer(t, &got)
	defer srv.Close()
	sender := newMetaSenderAgainstTestServer(t, srv, templateConfig())

	if err := sender.Send(context.Background(), testRecipient, testCode); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.body["type"] != "template" {
		t.Fatalf("expected a template message, got %v", got.body["type"])
	}
	if _, hasText := got.body["text"]; hasText {
		t.Fatal("template mode must never fall back to free text")
	}
	template := got.body["template"].(map[string]any)
	if template["name"] != "codigo_verificacion" || template["language"].(map[string]any)["code"] != "es" {
		t.Fatalf("unexpected template: %v", template)
	}
	components := template["components"].([]any)
	if len(components) != 2 {
		t.Fatalf("expected body and button components, got %d", len(components))
	}
	for i, wantType := range []string{"body", "button"} {
		c := components[i].(map[string]any)
		if c["type"] != wantType {
			t.Fatalf("component %d: expected %s, got %v", i, wantType, c["type"])
		}
		param := c["parameters"].([]any)[0].(map[string]any)
		if param["type"] != "text" || param["text"] != testCode {
			t.Fatalf("component %s must carry the NAVA code, got %v", wantType, param)
		}
	}
	button := components[1].(map[string]any)
	if button["sub_type"] != "url" || button["index"] != "0" {
		t.Fatalf("unexpected button component: %v", button)
	}
}

func TestMetaSender_InvalidRecipient_NeverCallsMeta(t *testing.T) {
	for _, phone := range []string{"", "3001234567", "+57 300 123 4567", "+0123456789", "+57300abc4567", "+1234567"} {
		var got capturedRequest
		srv := acceptingServer(t, &got)
		sender := newMetaSenderAgainstTestServer(t, srv, templateConfig())
		err := sender.Send(context.Background(), phone, testCode)
		srv.Close()
		if !errors.Is(err, notification.ErrInvalidRecipient) {
			t.Fatalf("phone %q: expected ErrInvalidRecipient, got %v", phone, err)
		}
		if got.calls.Load() != 0 {
			t.Fatalf("phone %q: Meta must not be called for an invalid recipient", phone)
		}
	}
}

func TestMetaSender_DevelopmentMode_RejectsRecipientOutsideAllowlist(t *testing.T) {
	var got capturedRequest
	srv := acceptingServer(t, &got)
	defer srv.Close()
	sender := newMetaSenderAgainstTestServer(t, srv, developmentConfig())

	err := sender.Send(context.Background(), "+573009999999", testCode)
	if !errors.Is(err, notification.ErrRecipientNotAuthorized) {
		t.Fatalf("expected ErrRecipientNotAuthorized, got %v", err)
	}
	if got.calls.Load() != 0 {
		t.Fatal("an unauthorized recipient must never reach Meta")
	}
}

func TestMetaSender_TemplateMode_IgnoresDevelopmentAllowlist(t *testing.T) {
	var got capturedRequest
	srv := acceptingServer(t, &got)
	defer srv.Close()
	cfg := templateConfig()
	cfg.TestRecipients = []string{"+573000000000"}
	sender := newMetaSenderAgainstTestServer(t, srv, cfg)

	if err := sender.Send(context.Background(), testRecipient, testCode); err != nil {
		t.Fatalf("template mode must send to any valid recipient: %v", err)
	}
}

func TestMetaSender_GraphAPIErrors_AreMappedToDomainErrors(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{"ventana cerrada", 400, `{"error":{"type":"OAuthException","code":131047,"fbtrace_id":"AbC"}}`, notification.ErrConversationWindowClosed},
		{"destinatario no permitido", 400, `{"error":{"code":131030}}`, notification.ErrRecipientNotAllowed},
		{"destinatario no alcanzable", 400, `{"error":{"code":131026}}`, notification.ErrRecipientUnreachable},
		{"token vencido", 401, `{"error":{"type":"OAuthException","code":190}}`, notification.ErrMetaCredentials},
		{"sin permiso", 403, `{"error":{"code":10,"error_subcode":2388185}}`, notification.ErrMetaCredentials},
		{"límite por HTTP", 429, `{}`, notification.ErrMetaRateLimited},
		{"límite por código", 400, `{"error":{"code":130429}}`, notification.ErrMetaRateLimited},
		{"plantilla inexistente", 404, `{"error":{"code":132001}}`, notification.ErrTemplateUnavailable},
		{"parámetros de plantilla", 400, `{"error":{"code":132000}}`, notification.ErrTemplateUnavailable},
		{"error genérico 400", 400, `{"error":{"code":100}}`, nil},
		{"error 500", 500, `no es json`, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()
			sender := newMetaSenderAgainstTestServer(t, srv, templateConfig())

			err := sender.Send(context.Background(), testRecipient, testCode)
			if err == nil {
				t.Fatal("expected an error")
			}
			var apiErr *notification.MetaAPIError
			if !errors.As(err, &apiErr) || apiErr.HTTPStatus != tt.status {
				t.Fatalf("expected a MetaAPIError with status %d, got %v", tt.status, err)
			}
			if tt.want != nil && !errors.Is(err, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, err)
			}
		})
	}
}

func TestMetaSender_Errors_NeverLeakSecretsRecipientOrCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"número inválido +573001234567 código 583921 token token-secreto-de-prueba","code":100}}`))
	}))
	defer srv.Close()
	sender := newMetaSenderAgainstTestServer(t, srv, templateConfig())

	err := sender.Send(context.Background(), testRecipient, testCode)
	for _, secret := range []string{"573001234567", testCode, testToken} {
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("error leaked %q: %v", secret, err)
		}
	}
}

func TestMetaSender_Success_WithoutMessageID_IsNotReportedAsAccepted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"messages":[]}`))
	}))
	defer srv.Close()
	sender := newMetaSenderAgainstTestServer(t, srv, templateConfig())

	if _, err := sender.SendOTP(context.Background(), testRecipient, testCode); err == nil {
		t.Fatal("a 200 without a wamid must not count as an accepted message")
	}
}

func TestMetaSender_NetworkFailureAndServerError_AreNeverRetried(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadGateway)
	}))
	sender := newMetaSenderAgainstTestServer(t, srv, templateConfig())
	if err := sender.Send(context.Background(), testRecipient, testCode); err == nil {
		t.Fatal("expected an error on 502")
	}
	srv.Close()
	if calls.Load() != 1 {
		t.Fatalf("a 5xx must not be retried, got %d requests", calls.Load())
	}

	// Conexión cerrada tras aceptar la solicitud: el resultado es desconocido.
	var dropped atomic.Int32
	drop := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		dropped.Add(1)
		hj, _ := w.(http.Hijacker)
		conn, _, _ := hj.Hijack()
		_ = conn.Close()
	}))
	defer drop.Close()
	sender = newMetaSenderAgainstTestServer(t, drop, templateConfig())
	err := sender.Send(context.Background(), testRecipient, testCode)
	if err == nil {
		t.Fatal("expected an error when the connection drops")
	}
	for _, secret := range []string{testToken, "573001234567", testCode} {
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("network error leaked %q: %v", secret, err)
		}
	}
	if dropped.Load() != 1 {
		t.Fatalf("a dropped connection must not be retried, got %d requests", dropped.Load())
	}
}

func TestMetaSender_Timeout_ReturnsErrorWithoutRetry(t *testing.T) {
	var calls atomic.Int32
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()
	defer close(release)
	sender := newMetaSenderAgainstTestServer(t, srv, templateConfig())

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := sender.Send(ctx, testRecipient, testCode); err == nil {
		t.Fatal("expected a timeout error")
	}
	if calls.Load() != 1 {
		t.Fatalf("a timeout must not be retried, got %d requests", calls.Load())
	}
}

func TestMetaSender_Success_LogsWamidButNeverSecrets(t *testing.T) {
	var got capturedRequest
	srv := acceptingServer(t, &got)
	defer srv.Close()
	var logs bytes.Buffer
	cfg := templateConfig()
	cfg.Logger = slog.New(slog.NewTextHandler(&logs, nil))
	sender := newMetaSenderAgainstTestServer(t, srv, cfg)

	if err := sender.Send(context.Background(), testRecipient, testCode); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := logs.String()
	if !strings.Contains(out, "wamid.HBgM") {
		t.Fatalf("expected the wamid in the log, got %q", out)
	}
	for _, secret := range []string{testToken, "573001234567", testCode} {
		if strings.Contains(out, secret) {
			t.Fatalf("log leaked %q: %s", secret, out)
		}
	}
}

func TestMetaSender_ResponseBodyIsBounded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, strings.Repeat("x", 1<<20))
	}))
	defer srv.Close()
	sender := newMetaSenderAgainstTestServer(t, srv, templateConfig())

	if err := sender.Send(context.Background(), testRecipient, testCode); err == nil {
		t.Fatal("expected an error")
	}
}
