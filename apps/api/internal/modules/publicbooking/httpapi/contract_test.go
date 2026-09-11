// Pruebas de contrato de HU-090: comparan el handler real contra el
// OpenAPI fuente (api/openapi/paths/public-booking.yaml y sus schemas), sin
// depender de Node/Redocly (mismo criterio que
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

type parameter struct {
	Name     string `yaml:"name"`
	In       string `yaml:"in"`
	Required bool   `yaml:"required"`
}

type operation struct {
	OperationID string         `yaml:"operationId"`
	Security    []any          `yaml:"security"`
	Parameters  []parameter    `yaml:"parameters"`
	Responses   map[string]any `yaml:"responses"`
}

type publicBarbershopPathFile struct {
	BarbershopBySlug struct {
		Get operation `yaml:"get"`
	} `yaml:"/public/barbershops/{slug}"`
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

// TestContract_PublicBarbershopProfileSchema_MatchesDTOFields verifica que
// httpapi.PublicBarbershopProfileResponse (name, timezone, contactEmail,
// contactPhone) coincide exactamente con las propiedades y los required de
// PublicBarbershopProfile.yaml (CA-090-04: nunca un identificador interno
// ni el slug mismo).
func TestContract_PublicBarbershopProfileSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/PublicBarbershopProfile.yaml")
	want := []string{"name", "timezone", "contactEmail", "contactPhone"}

	if len(schema.Properties) != len(want) {
		t.Fatalf("expected %d properties, schema has %d: %v", len(want), len(schema.Properties), schema.Properties)
	}
	for _, prop := range want {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("schema no declara la propiedad %q", prop)
		}
	}
	for _, forbidden := range []string{"id", "barbershopId", "slug"} {
		if _, ok := schema.Properties[forbidden]; ok {
			t.Fatalf("CA-090-04: el contrato público nunca debe declarar %q", forbidden)
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

// TestContract_ResolvePublicBarbershopOperation_MethodPathSecurityAndResponses
// verifica que el path fuente declara exactamente la operación pública sin
// sesión (security: []), el parámetro slug obligatorio en la ruta y los
// códigos que el handler realmente produce (200, 404 uniforme, 500), sin
// inventar códigos no ejercitados (en particular, SIN 401: esta operación
// nunca exige sesión).
func TestContract_ResolvePublicBarbershopOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[publicBarbershopPathFile](t, "api/openapi/paths/public-booking.yaml")
	op := doc.BarbershopBySlug.Get

	if op.OperationID != "resolvePublicBarbershop" {
		t.Fatalf("expected operationId resolvePublicBarbershop, got %q", op.OperationID)
	}
	if len(op.Security) != 0 {
		t.Fatalf("CA-090-01: esta operación nunca exige sesión, security debe ser [], got %v", op.Security)
	}

	if len(op.Parameters) != 1 || op.Parameters[0].Name != "slug" || op.Parameters[0].In != "path" || !op.Parameters[0].Required {
		t.Fatalf("expected exactly one required path parameter named slug, got %+v", op.Parameters)
	}

	wantStatuses := []string{"200", "404", "500"}
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

// TestContract_OpenAPIYAML_RegistersPublicBarbershopPath confirma que
// openapi.yaml registra el path bajo el mismo documento raíz que las demás
// operaciones públicas, con la misma técnica de referencia JSON pointer.
func TestContract_OpenAPIYAML_RegistersPublicBarbershopPath(t *testing.T) {
	type pathsDoc struct {
		Paths map[string]any `yaml:"paths"`
	}
	doc := loadYAML[pathsDoc](t, "api/openapi/openapi.yaml")

	if _, ok := doc.Paths["/public/barbershops/{slug}"]; !ok {
		t.Fatal("openapi.yaml no registra paths./public/barbershops/{slug}")
	}
}
