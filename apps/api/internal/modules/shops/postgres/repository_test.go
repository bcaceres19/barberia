// Package postgres_test (pruebas de integración, HU-020) requiere
// PostgreSQL REAL con las nueve migraciones aplicadas (incluida
// 20260823120000_add_barbershop_contact_info.sql) y
// database/testdata/dos_barberias.sql cargado. Conéctate como barberia_app
// (docs/03-desarrollo/estrategia-pruebas.md §2 prohíbe mocks para RLS/zona
// IANA/atomicidad).
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/shops/postgres/...
package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"system-barbershop/internal/modules/shops"
	shopspostgres "system-barbershop/internal/modules/shops/postgres"
	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
)

const (
	testDatabaseURL = "postgres://barberia_app@localhost:5432/barberia_test?sslmode=disable"

	// NO shopA/shopB reales (11111111.../22222222...) ni el par "shopC/shopD"
	// de hu021_barberos.sql (33333333.../44444444..., ajeno a este archivo):
	// este paquete corre como un binario de test SEPARADO del de cmd/api, y
	// `go test ./...` ejecuta binarios de paquete en paralelo por defecto.
	// cmd/api/settings_integration_test.go ya hace UPDATE confirmado sobre
	// las mismas columnas del par real A/B; compartir esa fila aquí produjo
	// una fuga de estado real entre paquetes (issue #158). Este es el par
	// dedicado de dos_barberias.sql reservado para ese caso.
	shopA = "12121212-1212-1212-1212-121212121212"
	shopB = "34343434-3434-3434-3434-343434343434"
)

func setupTestDB(t *testing.T) *database.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = testDatabaseURL
	}
	cfg := config.Config{
		Environment:              "test",
		DatabaseMaxConns:         10,
		DatabaseMinConns:         2,
		DatabaseMaxConnLifetime:  time.Hour,
		DatabaseMaxConnIdleTime:  30 * time.Minute,
		DatabaseConnectTimeout:   5 * time.Second,
		DatabaseStatementTimeout: 10 * time.Second,
	}
	db, err := database.NewDB(config.DatabaseDSN(url), cfg)
	if err != nil {
		t.Fatalf("database.NewDB: %v", err)
	}
	// Registrado con t.Cleanup (no un `defer db.Close()` en cada prueba) a
	// propósito: t.Cleanup ejecuta en orden LIFO junto con el de
	// restoreShopA, así que el cierre del pool registrado AQUÍ (primero)
	// corre DESPUÉS de que restoreShopA (registrado más tarde, en cada
	// prueba) todavía tenga un pool abierto para restaurar la fila.
	t.Cleanup(func() { db.Close() })
	return db
}

// sameBarbershop compara por VALOR, no por identidad de puntero:
// ContactEmail/ContactPhone son *string, y dos lecturas independientes
// (por ejemplo el RETURNING de un UPDATE y un SELECT posterior) nunca
// comparten la misma dirección de memoria aunque el contenido sea idéntico
// (comparar los struct con == compararía esas direcciones, no el texto).
func sameBarbershop(a, b shops.Barbershop) bool {
	if a.Name != b.Name || a.Timezone != b.Timezone {
		return false
	}
	return samePointerValue(a.ContactEmail, b.ContactEmail) && samePointerValue(a.ContactPhone, b.ContactPhone)
}

func samePointerValue(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// restoreShopA deja shopA en el mismo estado que dos_barberias.sql
// (name/timezone originales, sin contacto) al terminar cada prueba: las
// pruebas de este archivo comparten la fila de dos_barberias.sql en vez de
// crear una barbería propia, porque HU-020 no expone alta de barberías.
func restoreShopA(t *testing.T, db *database.DB) {
	t.Helper()
	t.Cleanup(func() {
		err := db.InTenantTx(context.Background(), database.BarbershopID(shopA), func(ctx context.Context, q database.Queries) error {
			_, err := q.Exec(ctx,
				`UPDATE barbershop SET name = $2, timezone = $3, contact_email = NULL, contact_phone = NULL WHERE id = $1`,
				shopA, "Barbería de prueba (aislamiento de paquete) 1", "America/Bogota",
			)
			return err
		})
		if err != nil {
			t.Fatalf("restoreShopA cleanup: %v", err)
		}
	})
}

func TestGet_ReturnsTheFourAuthorizedFields(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.New(db)
	restoreShopA(t, db)

	b, found, err := repo.Get(context.Background(), shopA)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found {
		t.Fatal("expected shopA to be found")
	}
	if b.Name != "Barbería de prueba (aislamiento de paquete) 1" || b.Timezone != "America/Bogota" {
		t.Fatalf("unexpected barbershop: %+v", b)
	}
	if b.ContactEmail != nil || b.ContactPhone != nil {
		t.Fatalf("expected no contact configured yet, got %+v", b)
	}
}

// TestUpdate_ValidContact_PersistsAndRoundTrips cubre CA-020-02/CA-020-06:
// el contacto guardado se lee de vuelta exactamente igual, sin una segunda
// consulta (la propia sentencia UPDATE ... RETURNING lo confirma).
func TestUpdate_ValidContact_PersistsAndRoundTrips(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.New(db)
	restoreShopA(t, db)

	email := "contacto@ejemplo.test"
	phone := "+573001234567"
	result, err := repo.Update(context.Background(), shopA, shops.UpdateInput{
		Name: "Barbería de prueba (aislamiento de paquete) 1 renombrada", Timezone: "America/Bogota",
		ContactEmail: &email, ContactPhone: &phone,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !result.TimezoneValid || !result.Found {
		t.Fatalf("expected a successful update, got %+v", result)
	}
	if result.Barbershop.Name != "Barbería de prueba (aislamiento de paquete) 1 renombrada" {
		t.Fatalf("unexpected name: %+v", result.Barbershop)
	}
	if result.Barbershop.ContactEmail == nil || *result.Barbershop.ContactEmail != email {
		t.Fatalf("expected contactEmail=%q, got %v", email, result.Barbershop.ContactEmail)
	}
	if result.Barbershop.ContactPhone == nil || *result.Barbershop.ContactPhone != phone {
		t.Fatalf("expected contactPhone=%q, got %v", phone, result.Barbershop.ContactPhone)
	}

	// Confirmación independiente con una lectura nueva (no la misma
	// sentencia): descarta que RETURNING mienta sobre lo persistido.
	got, found, err := repo.Get(context.Background(), shopA)
	if err != nil {
		t.Fatalf("Get after Update: %v", err)
	}
	if !found || !sameBarbershop(got, result.Barbershop) {
		t.Fatalf("expected the persisted row to match the RETURNING result, got %+v (found=%v)", got, found)
	}
}

// TestUpdate_NilContact_PersistsAsNullNotEmptyString cubre CA-020-06.
func TestUpdate_NilContact_PersistsAsNullNotEmptyString(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.New(db)
	restoreShopA(t, db)

	if _, err := repo.Update(context.Background(), shopA, shops.UpdateInput{
		Name: "Barbería de prueba (aislamiento de paquete) 1", Timezone: "America/Bogota",
		ContactEmail: nil, ContactPhone: nil,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, found, err := repo.Get(context.Background(), shopA)
	if err != nil || !found {
		t.Fatalf("Get: found=%v err=%v", found, err)
	}
	if got.ContactEmail != nil || got.ContactPhone != nil {
		t.Fatalf("expected NULL contact, got %+v", got)
	}
}

// TestUpdate_InvalidTimezone_WritesNothing cubre CA-020-03: una zona que
// pg_timezone_names no reconoce implica cero escritura dentro de la MISMA
// transacción; el nombre y el contacto anteriores permanecen intactos.
func TestUpdate_InvalidTimezone_WritesNothing(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.New(db)
	restoreShopA(t, db)

	email := "antes@ejemplo.test"
	if _, err := repo.Update(context.Background(), shopA, shops.UpdateInput{
		Name: "Nombre antes de la prueba", Timezone: "America/Bogota", ContactEmail: &email,
	}); err != nil {
		t.Fatalf("seed Update: %v", err)
	}

	for _, invalid := range []string{"COT", "UTC-5", "America/Bogota ", "not-a-zone"} {
		result, err := repo.Update(context.Background(), shopA, shops.UpdateInput{
			Name: "Nombre que NO debe guardarse", Timezone: invalid,
		})
		if err != nil {
			t.Fatalf("Update with invalid timezone %q: %v", invalid, err)
		}
		if result.TimezoneValid {
			t.Fatalf("expected TimezoneValid=false for %q", invalid)
		}
		if result.Found {
			t.Fatalf("expected Found=false (no row touched) for invalid timezone %q", invalid)
		}
	}

	got, found, err := repo.Get(context.Background(), shopA)
	if err != nil || !found {
		t.Fatalf("Get: found=%v err=%v", found, err)
	}
	if got.Name != "Nombre antes de la prueba" {
		t.Fatalf("CA-020-03: name changed despite an invalid timezone: %+v", got)
	}
	if got.ContactEmail == nil || *got.ContactEmail != email {
		t.Fatalf("CA-020-03: contact changed despite an invalid timezone: %+v", got)
	}
}

// TestGetAndUpdate_TenantAIsolatedFromTenantB repite, acotado a las
// columnas de HU-020, la garantía ya probada a nivel SQL en
// database/tests/hu020_configuracion_barberia.sql: el repositorio nunca
// mezcla el tenant de contexto con un identificador distinto, porque
// barbershopID SIEMPRE se usa a la vez como contexto de InTenantTx y como
// filtro (no existe un canal para pedir "la fila de B con el contexto de
// A" a través de este puerto).
func TestGetAndUpdate_TenantAIsolatedFromTenantB(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.New(db)
	restoreShopA(t, db)

	a, foundA, err := repo.Get(context.Background(), shopA)
	if err != nil || !foundA {
		t.Fatalf("Get shopA: found=%v err=%v", foundA, err)
	}
	b, foundB, err := repo.Get(context.Background(), shopB)
	if err != nil || !foundB {
		t.Fatalf("Get shopB: found=%v err=%v", foundB, err)
	}
	if a.Name == b.Name {
		t.Fatalf("test fixture assumption broken: shopA and shopB have the same name %q", a.Name)
	}

	if _, err := repo.Update(context.Background(), shopA, shops.UpdateInput{
		Name: "Solo A debe cambiar", Timezone: "America/Bogota",
	}); err != nil {
		t.Fatalf("Update shopA: %v", err)
	}

	bAfter, _, err := repo.Get(context.Background(), shopB)
	if err != nil {
		t.Fatalf("Get shopB after updating shopA: %v", err)
	}
	if bAfter.Name != b.Name {
		t.Fatalf("CA-020-05: updating shopA changed shopB's name from %q to %q", b.Name, bAfter.Name)
	}
}
