// Prueba estructural CA-002-06: verifica que ningún módulo de dominio
// importe internal/platform/database o github.com/jackc/pgx. Esta prueba
// existe para el import que alguien agregará dentro de seis meses con
// prisa. Debe fallar con un mensaje que diga qué paquete importó qué y por
// qué está prohibido.
//
// Vive en package database_test (no en un package "testdeps" propio):
// un directorio de Go solo admite el paquete de producción y, como mucho,
// su paquete de prueba externo "_test"; un tercer nombre de paquete en el
// mismo directorio impide compilar todo el módulo.
package database_test

import (
	"fmt"
	"go/build"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoDatabaseImportsInDomain verifica la dirección de dependencias
// CA-002-06 (estandar-backend-go.md §3). El dominio y los servicios no
// importan el paquete de base de datos ni tipos de pgx.
func TestNoDatabaseImportsInDomain(t *testing.T) {
	pkg, err := build.Import("system-barbershop/internal/platform/database", "", build.FindOnly)
	if err != nil {
		t.Fatalf("cannot find database package: %v", err)
	}

	// Buscar todos los paquetes bajo internal/modules/
	modulesDir := filepath.Join(filepath.Dir(pkg.Dir), "..", "modules")

	pkgs, err := filepath.Glob(filepath.Join(modulesDir, "*/*.go"))
	if err != nil {
		t.Fatalf("glob modules: %v", err)
	}

	var violations []string

	for _, file := range pkgs {
		// Solo archivos Go, no de prueba
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		bpkg, err := build.ImportDir(filepath.Dir(file), 0)
		if err != nil {
			// Ignorar directorios que no son paquetes válidos
			continue
		}

		// Verificar imports
		for _, imp := range bpkg.Imports {
			if imp == "system-barbershop/internal/platform/database" ||
				strings.HasPrefix(imp, "github.com/jackc/pgx") {
				violations = append(violations,
					fmt.Sprintf("%s importa %s (PROHIBIDO: dominio no puede importar base de datos ni driver pgx)",
						bpkg.ImportPath, imp))
			}
		}

		// Verificar test imports también
		for _, imp := range bpkg.TestImports {
			if imp == "system-barbershop/internal/platform/database" ||
				strings.HasPrefix(imp, "github.com/jackc/pgx") {
				violations = append(violations,
					fmt.Sprintf("%s (test) importa %s (PROHIBIDO)",
						bpkg.ImportPath, imp))
			}
		}
	}

	if len(violations) > 0 {
		t.Fatalf("Violaciones de CA-002-06 (dirección de dependencias):\n%s",
			strings.Join(violations, "\n"))
	}
}
