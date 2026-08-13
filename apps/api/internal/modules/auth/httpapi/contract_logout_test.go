// Pruebas de contrato de HU-006: comparan la operación de logout contra el
// OpenAPI fuente (api/openapi/paths/private-auth.yaml y
// api/openapi/components/security-schemes/SessionCookie.yaml), con el mismo
// criterio que contract_test.go (HU-005): se lee el YAML fuente
// directamente, no el bundle de dist/.
package httpapi_test

import (
	"testing"

	"system-barbershop/internal/modules/auth/httpapi"
)

type logoutOperation struct {
	OperationID string           `yaml:"operationId"`
	Security    []map[string]any `yaml:"security"`
	Responses   map[string]any   `yaml:"responses"`
}

type logoutPathFile struct {
	Logout struct {
		Post logoutOperation `yaml:"post"`
	} `yaml:"/private/auth/logout"`
}

// TestContract_LogoutOperation_MethodPathSecurityAndResponses verifica que
// el path fuente declara exactamente la operación de cierre de sesión: el
// esquema de seguridad SessionCookie (nunca security: [], a diferencia del
// login) y los códigos de respuesta que el handler realmente produce (204
// éxito, 401 sesión inválida, 500 error interno seguro), sin inventar
// códigos no ejercitados.
func TestContract_LogoutOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[logoutPathFile](t, "api/openapi/paths/private-auth.yaml")
	op := doc.Logout.Post

	if op.OperationID != "logout" {
		t.Fatalf("expected operationId logout, got %q", op.OperationID)
	}
	if len(op.Security) != 1 {
		t.Fatalf("expected exactly one security requirement (SessionCookie), got %v", op.Security)
	}
	if _, ok := op.Security[0]["SessionCookie"]; !ok {
		t.Fatalf("expected the SessionCookie security scheme, got %v", op.Security[0])
	}

	wantStatuses := []string{"204", "401", "500"}
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

// TestContract_SessionCookieSecurityScheme_MatchesRuntimeConfig confirma
// que el esquema de seguridad registrado (nombre, ubicación) coincide con
// la cookie que SessionMiddleware realmente lee, para que HU-006 no
// registre un esquema desconectado del código.
func TestContract_SessionCookieSecurityScheme_MatchesRuntimeConfig(t *testing.T) {
	type securityScheme struct {
		Type string `yaml:"type"`
		In   string `yaml:"in"`
		Name string `yaml:"name"`
	}
	scheme := loadYAML[securityScheme](t, "api/openapi/components/security-schemes/SessionCookie.yaml")

	if scheme.Type != "apiKey" || scheme.In != "cookie" {
		t.Fatalf("expected apiKey/cookie, got type=%q in=%q", scheme.Type, scheme.In)
	}
	if scheme.Name != httpapi.CookieName {
		t.Fatalf("SessionCookie.yaml documenta el nombre %q, el runtime usa %q", scheme.Name, httpapi.CookieName)
	}
}

// TestContract_OpenAPIYAML_RegistersSessionCookieSecurityScheme confirma
// que openapi.yaml registra el esquema bajo components/securitySchemes con
// la misma clave que la operación de logout referencia
// (SessionCookie.yaml ya no queda "listo pero sin registrar" como durante
// HU-005; ver DEC-058).
func TestContract_OpenAPIYAML_RegistersSessionCookieSecurityScheme(t *testing.T) {
	type componentsDoc struct {
		Components struct {
			SecuritySchemes map[string]any `yaml:"securitySchemes"`
		} `yaml:"components"`
	}
	doc := loadYAML[componentsDoc](t, "api/openapi/openapi.yaml")

	if _, ok := doc.Components.SecuritySchemes["SessionCookie"]; !ok {
		t.Fatal("openapi.yaml no registra components.securitySchemes.SessionCookie")
	}
}
