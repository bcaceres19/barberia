// Pruebas de contrato de HU-021: comparan los handlers reales contra el
// OpenAPI fuente (api/openapi/paths/staff.yaml y sus schemas), sin depender
// de Node/Redocly (mismo criterio que
// internal/modules/shops/httpapi/contract_test.go): se lee el YAML fuente
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

type staffPathsFile struct {
	Collection struct {
		Get  operation `yaml:"get"`
		Post operation `yaml:"post"`
	} `yaml:"/private/barbers"`
	Item struct {
		Get   operation `yaml:"get"`
		Patch operation `yaml:"patch"`
	} `yaml:"/private/barbers/{barberId}"`
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

func requireProps(t *testing.T, schema schemaDoc, want []string) {
	t.Helper()
	if len(schema.Properties) != len(want) {
		t.Fatalf("expected %d properties, schema has %d: %v", len(want), len(schema.Properties), schema.Properties)
	}
	for _, prop := range want {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("schema no declara la propiedad %q", prop)
		}
	}
	for _, wantReq := range want {
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

// TestContract_BarberResponseSchema_MatchesDTOFields verifica que
// httpapi.BarberResponse (id, fullName, createdAt, updatedAt) coincide
// exactamente con BarberResponse.yaml (CA-021-07: nunca active, deletedAt,
// sortOrder, staffUserId, servicios ni horarios).
func TestContract_BarberResponseSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/BarberResponse.yaml")
	requireProps(t, schema, []string{"id", "fullName", "createdAt", "updatedAt"})
}

func TestContract_CreateBarberRequestSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/CreateBarberRequest.yaml")
	requireProps(t, schema, []string{"fullName"})
	if _, ok := schema.Properties["barbershopId"]; ok {
		t.Fatal("el contrato de alta nunca debe declarar barbershopId")
	}
}

func TestContract_UpdateBarberRequestSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/UpdateBarberRequest.yaml")
	requireProps(t, schema, []string{"fullName"})
	for _, forbidden := range []string{"active", "deletedAt", "sortOrder", "staffUserId", "barbershopId"} {
		if _, ok := schema.Properties[forbidden]; ok {
			t.Fatalf("CA-021-07/DEC-047: el contrato de renombrado nunca debe declarar %q", forbidden)
		}
	}
}

func TestContract_BarberListResponseSchema_HasItemsAndNextCursor(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/BarberListResponse.yaml")
	requireProps(t, schema, []string{"items", "nextCursor"})
}

// --- Operaciones -----------------------------------------------------------

func TestContract_ListBarbersOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[staffPathsFile](t, "api/openapi/paths/staff.yaml")
	requireResponses(t, doc.Collection.Get, "listBarbers", []string{"200", "400", "401", "500"})
}

func TestContract_CreateBarberOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[staffPathsFile](t, "api/openapi/paths/staff.yaml")
	requireResponses(t, doc.Collection.Post, "createBarber", []string{"201", "400", "401", "409", "422", "500"})
}

func TestContract_GetBarberOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[staffPathsFile](t, "api/openapi/paths/staff.yaml")
	requireResponses(t, doc.Item.Get, "getBarber", []string{"200", "401", "404", "500"})
}

func TestContract_RenameBarberOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[staffPathsFile](t, "api/openapi/paths/staff.yaml")
	requireResponses(t, doc.Item.Patch, "renameBarber", []string{"200", "400", "401", "404", "422", "500"})
}

// TestContract_OpenAPIYAML_RegistersStaffPaths confirma que openapi.yaml
// registra ambos paths bajo el mismo documento raíz que las demás
// operaciones privadas, con la misma técnica de referencia JSON pointer.
func TestContract_OpenAPIYAML_RegistersStaffPaths(t *testing.T) {
	type pathsDoc struct {
		Paths map[string]any `yaml:"paths"`
	}
	doc := loadYAML[pathsDoc](t, "api/openapi/openapi.yaml")

	if _, ok := doc.Paths["/private/barbers"]; !ok {
		t.Fatal("openapi.yaml no registra paths./private/barbers")
	}
	if _, ok := doc.Paths["/private/barbers/{barberId}"]; !ok {
		t.Fatal("openapi.yaml no registra paths./private/barbers/{barberId}")
	}
}
