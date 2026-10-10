// Prueba de contrato del webhook de Meta (issue #355, DEC-126): compara las
// operaciones con api/openapi/paths/public-webhooks.yaml, mismo criterio que
// las pruebas de contrato de los demás módulos.
package httpapi_test

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"gopkg.in/yaml.v3"
)

type webhookOperation struct {
	OperationID string         `yaml:"operationId"`
	Security    []any          `yaml:"security"`
	Responses   map[string]any `yaml:"responses"`
}

type webhookPathFile struct {
	Meta struct {
		Get  webhookOperation `yaml:"get"`
		Post webhookOperation `yaml:"post"`
	} `yaml:"/public/webhooks/meta/whatsapp"`
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
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

func loadWebhookPaths(t *testing.T) webhookPathFile {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), "api", "openapi", "paths", "public-webhooks.yaml"))
	if err != nil {
		t.Fatalf("leer el contrato: %v", err)
	}
	var doc webhookPathFile
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("interpretar el contrato: %v", err)
	}
	return doc
}

func TestContract_WebhookOperationsAreDeclaredAsPublicWithTheirStatuses(t *testing.T) {
	doc := loadWebhookPaths(t)

	for name, tc := range map[string]struct {
		op   webhookOperation
		id   string
		want []string
	}{
		"GET":  {doc.Meta.Get, "verifyMetaWhatsAppWebhook", []string{"200", "400", "401", "500"}},
		"POST": {doc.Meta.Post, "receiveMetaWhatsAppWebhook", []string{"200", "400", "401", "500"}},
	} {
		if tc.op.OperationID != tc.id {
			t.Errorf("%s: expected operationId %s, got %q", name, tc.id, tc.op.OperationID)
		}
		if tc.op.Security == nil || len(tc.op.Security) != 0 {
			t.Errorf("%s: expected security: [] (sin sesión), got %v", name, tc.op.Security)
		}
		for _, status := range tc.want {
			if _, ok := tc.op.Responses[status]; !ok {
				t.Errorf("%s: falta la respuesta %s en el contrato", name, status)
			}
		}
	}
}

// Cada código que el contrato declara se ejercita contra el handler real: si el
// handler devolviera uno no declarado, esta prueba lo detecta.
func TestContract_HandlerStatusesAreAllDeclared(t *testing.T) {
	doc := loadWebhookPaths(t)
	store := &countingStore{}
	srv := newServer(t, store)

	declared := func(op webhookOperation, status int) bool {
		_, ok := op.Responses[strconv.Itoa(status)]
		return ok
	}

	resp, _ := get(t, srv, url.Values{"hub.mode": {"subscribe"}, "hub.verify_token": {verifyToken}, "hub.challenge": {"1"}})
	if !declared(doc.Meta.Get, resp.StatusCode) {
		t.Errorf("GET %d no está declarado", resp.StatusCode)
	}
	resp, _ = get(t, srv, url.Values{"hub.mode": {"subscribe"}, "hub.verify_token": {"mal"}, "hub.challenge": {"1"}})
	if !declared(doc.Meta.Get, resp.StatusCode) {
		t.Errorf("GET %d no está declarado", resp.StatusCode)
	}
	resp, _ = get(t, srv, url.Values{"hub.mode": {"subscribe"}, "hub.verify_token": {verifyToken}, "hub.challenge": {"<x>"}})
	if !declared(doc.Meta.Get, resp.StatusCode) {
		t.Errorf("GET %d no está declarado", resp.StatusCode)
	}

	resp, _ = post(t, srv, validBody, sign(validBody))
	if resp.StatusCode != http.StatusOK || !declared(doc.Meta.Post, resp.StatusCode) {
		t.Errorf("POST %d no está declarado", resp.StatusCode)
	}
	resp, _ = post(t, srv, validBody, "")
	if !declared(doc.Meta.Post, resp.StatusCode) {
		t.Errorf("POST %d no está declarado", resp.StatusCode)
	}
	resp, _ = post(t, srv, "no es json", sign("no es json"))
	if !declared(doc.Meta.Post, resp.StatusCode) {
		t.Errorf("POST %d no está declarado", resp.StatusCode)
	}
}
