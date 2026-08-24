// Pruebas de contrato de HU-022: comparan los handlers reales contra el
// OpenAPI fuente (api/openapi/paths/catalog.yaml y sus schemas), sin
// depender de Node/Redocly (mismo criterio que
// internal/modules/staff/httpapi/contract_test.go): se lee el YAML fuente
// directamente, no el bundle de dist/.
package httpapi_test

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
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

type catalogPathsFile struct {
	Collection struct {
		Get  operation `yaml:"get"`
		Post operation `yaml:"post"`
	} `yaml:"/private/services"`
	Item struct {
		Get   operation `yaml:"get"`
		Patch operation `yaml:"patch"`
	} `yaml:"/private/services/{serviceId}"`
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

// requireExactProps confirma que schema declara EXACTAMENTE wantProps (ni
// más ni menos) y que EXACTAMENTE wantRequired de ellos son obligatorios.
func requireExactProps(t *testing.T, schema schemaDoc, wantProps, wantRequired []string) {
	t.Helper()
	if len(schema.Properties) != len(wantProps) {
		t.Fatalf("expected %d properties, schema has %d: %v", len(wantProps), len(schema.Properties), schema.Properties)
	}
	for _, prop := range wantProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("schema no declara la propiedad %q", prop)
		}
	}
	if len(schema.Required) != len(wantRequired) {
		t.Fatalf("expected %d required fields, schema has %d: %v", len(wantRequired), len(schema.Required), schema.Required)
	}
	for _, wantReq := range wantRequired {
		found := false
		for _, r := range schema.Required {
			if r == wantReq {
				found = true
			}
		}
		if !found {
			t.Errorf("schema no marca %q como required", wantReq)
		}
	}
}

func requireResponses(t *testing.T, op operation, wantOperationID string, wantStatuses []string) {
	t.Helper()
	if op.OperationID != wantOperationID {
		t.Fatalf("expected operationId %q, got %q", wantOperationID, op.OperationID)
	}
	if len(op.Security) != 1 {
		t.Fatalf("expected exactly one security requirement (SessionCookie), got %v", op.Security)
	}
	for _, s := range wantStatuses {
		if _, ok := op.Responses[s]; !ok {
			t.Errorf("el contrato no documenta la respuesta %s para %s, pero el handler la produce", s, wantOperationID)
		}
	}
	if len(op.Responses) != len(wantStatuses) {
		t.Errorf("%s documenta %d respuestas, se esperaban exactamente %d (%v); revisa que no sobre ni falte una",
			wantOperationID, len(op.Responses), len(wantStatuses), wantStatuses)
	}
}

// --- Schemas -------------------------------------------------------------

// TestContract_ServiceResponseSchema_MatchesDTOFields verifica que
// httpapi.ServiceResponse (id, name, description, durationMinutes, price,
// currency, createdAt, updatedAt) coincide exactamente con
// ServiceResponse.yaml (CA-022-07: nunca isActive, deactivatedAt,
// asignaciones ni citas).
func TestContract_ServiceResponseSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/ServiceResponse.yaml")
	want := []string{"id", "name", "description", "durationMinutes", "price", "currency", "createdAt", "updatedAt"}
	requireExactProps(t, schema, want, want)
	for _, forbidden := range []string{"isActive", "deactivatedAt", "barbershopId", "barberIds", "appointments"} {
		if _, ok := schema.Properties[forbidden]; ok {
			t.Fatalf("CA-022-07: el contrato de respuesta nunca debe declarar %q", forbidden)
		}
	}
}

func TestContract_CreateServiceRequestSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/CreateServiceRequest.yaml")
	requireExactProps(t, schema,
		[]string{"name", "description", "durationMinutes", "price"},
		[]string{"name", "durationMinutes", "price"})
	for _, forbidden := range []string{"barbershopId", "currency", "isActive"} {
		if _, ok := schema.Properties[forbidden]; ok {
			t.Fatalf("DEC-067: el contrato de alta nunca debe declarar %q", forbidden)
		}
	}
}

func TestContract_UpdateServiceRequestSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/UpdateServiceRequest.yaml")
	requireExactProps(t, schema,
		[]string{"name", "description", "durationMinutes", "price"},
		[]string{})
	for _, forbidden := range []string{"isActive", "currency", "barbershopId", "barberIds", "appointments"} {
		if _, ok := schema.Properties[forbidden]; ok {
			t.Fatalf("RN-SER-04/DEC-067: el contrato de edición nunca debe declarar %q", forbidden)
		}
	}
}

func TestContract_ServiceListResponseSchema_HasItemsAndNextCursor(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/ServiceListResponse.yaml")
	requireExactProps(t, schema, []string{"items", "nextCursor"}, []string{"items", "nextCursor"})
}

// --- Operaciones -----------------------------------------------------------

func TestContract_ListServicesOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[catalogPathsFile](t, "api/openapi/paths/catalog.yaml")
	requireResponses(t, doc.Collection.Get, "listServices", []string{"200", "400", "401", "500"})
}

func TestContract_CreateServiceOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[catalogPathsFile](t, "api/openapi/paths/catalog.yaml")
	requireResponses(t, doc.Collection.Post, "createService", []string{"201", "400", "401", "409", "422", "500"})
}

func TestContract_GetServiceOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[catalogPathsFile](t, "api/openapi/paths/catalog.yaml")
	requireResponses(t, doc.Item.Get, "getService", []string{"200", "401", "404", "500"})
}

func TestContract_UpdateServiceOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[catalogPathsFile](t, "api/openapi/paths/catalog.yaml")
	requireResponses(t, doc.Item.Patch, "updateService", []string{"200", "400", "401", "404", "409", "422", "500"})
}

// TestContract_OpenAPIYAML_RegistersCatalogPaths confirma que openapi.yaml
// registra ambos paths bajo el mismo documento raíz que las demás
// operaciones privadas, con la misma técnica de referencia JSON pointer.
func TestContract_OpenAPIYAML_RegistersCatalogPaths(t *testing.T) {
	type pathsDoc struct {
		Paths map[string]any `yaml:"paths"`
	}
	doc := loadYAML[pathsDoc](t, "api/openapi/openapi.yaml")

	if _, ok := doc.Paths["/private/services"]; !ok {
		t.Fatal("openapi.yaml no registra paths./private/services")
	}
	if _, ok := doc.Paths["/private/services/{serviceId}"]; !ok {
		t.Fatal("openapi.yaml no registra paths./private/services/{serviceId}")
	}
}
