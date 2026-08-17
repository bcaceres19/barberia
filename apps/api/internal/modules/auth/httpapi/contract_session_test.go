// Pruebas de contrato de HU-012 (DEC-060): comparan la operación de
// contexto de sesión contra el OpenAPI fuente
// (api/openapi/paths/private-auth.yaml), con el mismo criterio que
// contract_logout_test.go: se lee el YAML fuente directamente, no el
// bundle de dist/.
package httpapi_test

import "testing"

type sessionContextOperation struct {
	OperationID string           `yaml:"operationId"`
	Security    []map[string]any `yaml:"security"`
	Responses   map[string]any   `yaml:"responses"`
}

type sessionContextPathFile struct {
	SessionContext struct {
		Get sessionContextOperation `yaml:"get"`
	} `yaml:"/private/auth/session"`
}

// TestContract_SessionContextOperation_MethodPathSecurityAndResponses
// verifica que el path fuente declara exactamente la operación de contexto
// de sesión: el esquema de seguridad SessionCookie (igual que logout) y los
// códigos de respuesta que el handler realmente produce (200 éxito, 401
// sesión inválida, 500 error interno seguro), sin inventar códigos no
// ejercitados.
func TestContract_SessionContextOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[sessionContextPathFile](t, "api/openapi/paths/private-auth.yaml")
	op := doc.SessionContext.Get

	if op.OperationID != "getSessionContext" {
		t.Fatalf("expected operationId getSessionContext, got %q", op.OperationID)
	}
	if len(op.Security) != 1 {
		t.Fatalf("expected exactly one security requirement (SessionCookie), got %v", op.Security)
	}
	if _, ok := op.Security[0]["SessionCookie"]; !ok {
		t.Fatalf("expected the SessionCookie security scheme, got %v", op.Security[0])
	}

	wantStatuses := []string{"200", "401", "500"}
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

// TestContract_OpenAPIYAML_RegistersSessionContextPath confirma que
// openapi.yaml registra el path bajo el mismo document raíz que las demás
// operaciones de auth, con la misma técnica de referencia JSON pointer.
func TestContract_OpenAPIYAML_RegistersSessionContextPath(t *testing.T) {
	type pathsDoc struct {
		Paths map[string]any `yaml:"paths"`
	}
	doc := loadYAML[pathsDoc](t, "api/openapi/openapi.yaml")

	if _, ok := doc.Paths["/private/auth/session"]; !ok {
		t.Fatal("openapi.yaml no registra paths./private/auth/session")
	}
}
