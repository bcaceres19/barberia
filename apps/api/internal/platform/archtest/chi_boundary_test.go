// Package archtest contiene pruebas estructurales sobre el árbol de
// paquetes de apps/api, no lógica de producción.
package archtest

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const chiImportPrefix = "github.com/go-chi/chi"

// packagesAllowedToImportChi es la única excepción: httpserver es la capa
// HTTP que DEC-034 autoriza a depender de Chi. archtest se excluye porque
// este archivo menciona el prefijo del import como cadena, no como import
// real.
var packagesAllowedToImportChi = map[string]bool{
	filepath.FromSlash("internal/platform/httpserver"): true,
	filepath.FromSlash("internal/platform/archtest"):   true,
}

// TestDomainAndServicesDoNotImportChi recorre internal/ y falla si algún
// paquete fuera de la excepción importa Chi, directa o indirectamente vía
// alias. Protege docs/04-arquitectura/backend-go.md sección 5 ("el servicio
// no conoce Chi") y CA-003-07.
func TestDomainAndServicesDoNotImportChi(t *testing.T) {
	root := findInternalDir(t)

	fset := token.NewFileSet()
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		rel, relErr := filepath.Rel(filepath.Dir(root), path)
		if relErr != nil {
			return relErr
		}
		pkgDir := filepath.ToSlash(filepath.Dir(rel))
		for allowed := range packagesAllowedToImportChi {
			if filepath.ToSlash(allowed) == pkgDir {
				return nil
			}
		}

		file, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			return parseErr
		}
		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			if strings.HasPrefix(importPath, chiImportPrefix) {
				t.Errorf("%s importa Chi (%s); el dominio y los servicios no deben conocer el router HTTP", path, importPath)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("no se pudo recorrer %s: %v", root, err)
	}
}

// findInternalDir ubica internal/ subiendo desde el directorio de trabajo
// de la prueba (internal/platform/archtest), sin asumir una ruta absoluta
// del checkout.
func findInternalDir(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("no se pudo obtener el directorio de trabajo: %v", err)
	}
	for {
		candidate := filepath.Join(dir, "internal")
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("no se encontró el directorio internal/ subiendo desde archtest")
		}
		dir = parent
	}
}
