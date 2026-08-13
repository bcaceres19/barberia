package httpserver_test

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
)

// problemSchema espeja lo único que esta prueba necesita del schema real de
// api/openapi/components/schemas/Problem.yaml: sus propiedades declaradas y
// cuáles son obligatorias. No se lee el bundle generado en dist/ (no es
// fuente y el job Go de CI no tiene Node/Redocly disponibles); se lee
// directamente el archivo fuente, que no tiene $ref externos que resolver.
type problemSchema struct {
	Properties map[string]any `yaml:"properties"`
	Required   []string       `yaml:"required"`
}

func loadProblemSchema(t *testing.T) problemSchema {
	t.Helper()

	repoRoot := findRepoRoot(t)
	path := filepath.Join(repoRoot, "api", "openapi", "components", "schemas", "Problem.yaml")

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("no se pudo leer %s: %v", path, err)
	}

	var schema problemSchema
	if err := yaml.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("no se pudo interpretar %s: %v", path, err)
	}
	if len(schema.Properties) == 0 {
		t.Fatalf("%s no declaró properties; ¿cambió la estructura del archivo?", path)
	}
	return schema
}

// findRepoRoot sube desde el directorio de la prueba hasta encontrar
// redocly.yaml, que solo vive en la raíz del repositorio.
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

// jsonFieldNames son las claves que Problem realmente serializa
// (docs/06-api/estandar-openapi.md sección 12): deben ser subconjunto de
// las properties del schema, y el schema no debe declarar una propiedad que
// Go nunca produce.
var jsonFieldNames = []string{"type", "title", "status", "detail", "instance", "code", "requestId"}

func TestContract_ProblemStructMatchesOpenAPISchema(t *testing.T) {
	schema := loadProblemSchema(t)

	for _, field := range jsonFieldNames {
		if _, ok := schema.Properties[field]; !ok {
			t.Errorf("Problem serializa %q pero el schema OpenAPI no lo declara", field)
		}
	}
	for prop := range schema.Properties {
		found := false
		for _, field := range jsonFieldNames {
			if field == prop {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("el schema OpenAPI declara %q pero Problem nunca lo serializa", prop)
		}
	}
}

// TestContract_TranslateAlwaysFillsRequiredFields recorre cada rama de
// Translate y confirma que ninguna deja vacío un campo que el schema marca
// required. Si el schema agrega un required nuevo sin que Translate lo
// llene, esta prueba lo detecta antes que un cliente real.
func TestContract_TranslateAlwaysFillsRequiredFields(t *testing.T) {
	schema := loadProblemSchema(t)

	cases := map[string]error{
		"not_found":       apperr.NotFound("recurso no disponible"),
		"internal":        apperr.Internal(errors.New("causa interna")),
		"unrecognized":    errors.New("error ajeno"),
		"max_bytes_error": &http.MaxBytesError{Limit: httpserver.MaxRequestBodyBytes},
	}

	for name, err := range cases {
		t.Run(name, func(t *testing.T) {
			p := httpserver.Translate(err, "req-contract")
			values := map[string]any{
				"type":      p.Type,
				"title":     p.Title,
				"status":    p.Status,
				"detail":    p.Detail,
				"instance":  p.Instance,
				"code":      p.Code,
				"requestId": p.RequestID,
			}
			for _, required := range schema.Required {
				v, ok := values[required]
				if !ok {
					t.Fatalf("campo requerido %q no existe en la proyección de la prueba", required)
				}
				if isZeroValue(v) {
					t.Errorf("campo requerido %q quedó vacío para el caso %s: %+v", required, name, p)
				}
			}
		})
	}
}

func isZeroValue(v any) bool {
	switch val := v.(type) {
	case string:
		return val == ""
	case int:
		return val == 0
	default:
		return v == nil
	}
}
