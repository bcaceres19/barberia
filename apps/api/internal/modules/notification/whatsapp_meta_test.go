package notification_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"system-barbershop/internal/modules/notification"
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

func TestMetaWhatsAppSender_Send_PostsTemplateWithExactCode(t *testing.T) {
	var gotAuth, gotPath string
	var gotBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"messages":[{"id":"wamid.test"}]}`))
	}))
	defer srv.Close()

	sender := newMetaSenderAgainstTestServer(t, srv, notification.MetaWhatsAppConfig{
		APIVersion: "v21.0", PhoneNumberID: "1234567890", AccessToken: "token-de-prueba",
		TemplateName: "recuperacion_acceso", LanguageCode: "es",
	})

	if err := sender.Send(context.Background(), "+573001234567", "482913"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotAuth != "Bearer token-de-prueba" {
		t.Fatalf("unexpected Authorization header: %q", gotAuth)
	}
	if !strings.Contains(gotPath, "1234567890") || !strings.Contains(gotPath, "v21.0") {
		t.Fatalf("unexpected request path: %q", gotPath)
	}
	if gotBody["to"] != "573001234567" {
		t.Fatalf("expected 'to' without the leading '+', got %v", gotBody["to"])
	}
	template, ok := gotBody["template"].(map[string]any)
	if !ok {
		t.Fatalf("expected a template object, got %v", gotBody["template"])
	}
	if template["name"] != "recuperacion_acceso" {
		t.Fatalf("unexpected template name: %v", template["name"])
	}
	body, _ := json.Marshal(gotBody)
	if !strings.Contains(string(body), "482913") {
		t.Fatal("expected the exact code to appear in the template parameters")
	}
}

func TestMetaWhatsAppSender_Send_ErrorStatus_ReturnsErrorWithoutLeakingBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"número inválido +573001234567"}}`))
	}))
	defer srv.Close()

	sender := newMetaSenderAgainstTestServer(t, srv, notification.MetaWhatsAppConfig{
		APIVersion: "v21.0", PhoneNumberID: "1", AccessToken: "t", TemplateName: "x", LanguageCode: "es",
	})

	err := sender.Send(context.Background(), "+573001234567", "482913")
	if err == nil {
		t.Fatal("expected an error on a non-2xx response")
	}
	if strings.Contains(err.Error(), "573001234567") {
		t.Fatal("expected the error message to never include the phone number from the provider body")
	}
}
