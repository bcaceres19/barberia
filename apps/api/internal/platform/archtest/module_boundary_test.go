package archtest

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// modulePairsForbiddenToImport documenta cada par de módulos que HU-023
// (trabajo requerido §3.1, "colaboración explícita ... mediante APIs/
// puertos pequeños ... no mediante acoplamiento del núcleo a su
// repositorio") exige mantener SIN dependencia de compilación directa en
// ninguna dirección: catalog.AssignmentService consulta staff solo a
// través de catalog.BarberPort (una interfaz que catalog declara), y
// cmd/api (la raíz de composición) es quien construye
// staff.NewBarberLookup(...) y lo pasa donde ese puerto lo espera. Ningún
// archivo de internal/modules/catalog debe importar
// system-barbershop/internal/modules/staff, y viceversa.
var modulePairsForbiddenToImport = [][2]string{
	{"internal/modules/catalog", "system-barbershop/internal/modules/staff"},
	{"internal/modules/staff", "system-barbershop/internal/modules/catalog"},
}

// TestModulesDoNotImportEachOther recorre internal/modules/<módulo> y falla
// si algún archivo .go de ese árbol importa, directa o indirectamente vía
// alias, el paquete del otro módulo de cada par declarado arriba. Protege
// la separación de HU-023 (catalog/staff colaboran solo por puerto
// explícito) de una regresión futura que reintroduzca un import cruzado
// "por comodidad".
func TestModulesDoNotImportEachOther(t *testing.T) {
	internalDir := findInternalDir(t)
	repoRoot := filepath.Dir(internalDir)

	fset := token.NewFileSet()
	for _, pair := range modulePairsForbiddenToImport {
		moduleDir := filepath.Join(repoRoot, filepath.FromSlash(pair[0]))
		forbiddenImport := pair[1]

		err := filepath.Walk(moduleDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}

			file, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
			if parseErr != nil {
				return parseErr
			}
			for _, imp := range file.Imports {
				importPath := strings.Trim(imp.Path.Value, `"`)
				if importPath == forbiddenImport || strings.HasPrefix(importPath, forbiddenImport+"/") {
					t.Errorf("%s importa %s; %s y %s deben colaborar solo mediante un puerto explícito (HU-023)",
						path, importPath, pair[0], forbiddenImport)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("no se pudo recorrer %s: %v", moduleDir, err)
		}
	}
}
