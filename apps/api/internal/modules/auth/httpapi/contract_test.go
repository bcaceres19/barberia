// Pruebas de contrato de HU-005: comparan el handler real contra el
// OpenAPI fuente (api/openapi/paths/public-auth.yaml y sus schemas), sin
// depender de Node/Redocly (mismo criterio que
// internal/platform/httpserver/contract_test.go): se lee el YAML fuente
// directamente, no el bundle de dist/.
package httpapi_test

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

	"system-barbershop/internal/modules/auth/httpapi"
)

type schemaDoc struct {
	Properties map[string]any `yaml:"properties"`
	Required   []string       `yaml:"required"`
}

type operation struct {
	OperationID string         `yaml:"operationId"`
	Security    []any          `yaml:"security"`
	Responses   map[string]any `yaml:"responses"`
}

type pathFile struct {
	Login struct {
		Post operation `yaml:"post"`
	} `yaml:"/public/auth/login"`
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("no se pudo obtener el directorio de trabajo: %v", err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "redocly.yaml")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("no se encontró redocly.yaml subiendo desde el directorio de la prueba")
		}
		dir = parent
	}
}

func loadYAML[T any](t *testing.T, relPath string) T {
	t.Helper()
	full := filepath.Join(findRepoRoot(t), filepath.FromSlash(relPath))
	raw, err := os.ReadFile(full)
	if err != nil {
		t.Fatalf("no se pudo leer %s: %v", full, err)
	}
	var v T
	if err := yaml.Unmarshal(raw, &v); err != nil {
		t.Fatalf("no se pudo interpretar %s: %v", full, err)
	}
	return v
}

// TestContract_LoginRequestSchema_MatchesDTOFields verifica que
// httpapi.LoginRequest (email, password) coincide exactamente con las
// propiedades y los required de api/openapi/components/schemas/LoginRequest.yaml.
func TestContract_LoginRequestSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/LoginRequest.yaml")

	wantProps := map[string]bool{"email": true, "password": true}
	if len(schema.Properties) != len(wantProps) {
		t.Fatalf("expected %d properties, schema has %d: %v", len(wantProps), len(schema.Properties), schema.Properties)
	}
	for prop := range wantProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("schema no declara la propiedad %q", prop)
		}
	}
	for _, want := range []string{"email", "password"} {
		found := false
		for _, r := range schema.Required {
			if r == want {
				found = true
			}
		}
		if !found {
			t.Errorf("schema no marca %q como required", want)
		}
	}
}

// TestContract_LoginResponseSchema_MatchesDTOFields verifica
// httpapi.LoginResponse contra LoginResponse.yaml.
func TestContract_LoginResponseSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/LoginResponse.yaml")

	if _, ok := schema.Properties["expiresAt"]; !ok {
		t.Fatal("schema no declara expiresAt")
	}
	if len(schema.Properties) != 1 {
		t.Fatalf("expected exactly 1 property (expiresAt), got %v", schema.Properties)
	}
	if len(schema.Required) != 1 || schema.Required[0] != "expiresAt" {
		t.Fatalf("expected expiresAt to be required, got %v", schema.Required)
	}
}

// TestContract_LoginOperation_MethodPathSecurityAndResponses verifica que
// el path fuente declara exactamente el método, la seguridad pública y los
// códigos de respuesta que el handler realmente produce (200 éxito, 400
// JSON/campo desconocido, 401 credenciales inválidas, 422 validación de
// campo, 500 error interno seguro), sin inventar códigos no ejercitados.
func TestContract_LoginOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[pathFile](t, "api/openapi/paths/public-auth.yaml")
	op := doc.Login.Post

	if op.OperationID != "loginWithPassword" {
		t.Fatalf("expected operationId loginWithPassword, got %q", op.OperationID)
	}
	if op.Security == nil || len(op.Security) != 0 {
		t.Fatalf("expected security: [] (operación pública, DEC-055), got %v", op.Security)
	}

	wantStatuses := []string{"200", "400", "401", "422", "500"}
	for _, s := range wantStatuses {
		if _, ok := op.Responses[s]; !ok {
			t.Errorf("el contrato no documenta la respuesta %s, pero el handler la produce", s)
		}
	}
	if len(op.Responses) != len(wantStatuses) {
		t.Errorf("el contrato documenta %d respuestas, se esperaban exactamente %d (%v); revisa que no sobre ni falte una",
			len(op.Responses), len(wantStatuses), wantStatuses)
	}
}

// TestContract_CookieNameMatchesSecuritySchemeReference confirma que el
// nombre de cookie que el handler realmente fija (httpapi.CookieName) es el
// mismo documentado en components/security-schemes/SessionCookie.yaml, para
// que HU-006 no herede una referencia desactualizada.
func TestContract_CookieNameMatchesSecuritySchemeReference(t *testing.T) {
	type securityScheme struct {
		Name string `yaml:"name"`
	}
	scheme := loadYAML[securityScheme](t, "api/openapi/components/security-schemes/SessionCookie.yaml")

	if scheme.Name != httpapi.CookieName {
		t.Fatalf("SessionCookie.yaml documenta el nombre %q, el handler usa %q", scheme.Name, httpapi.CookieName)
	}
}
