// Package postgres_test (pruebas de integración) requiere PostgreSQL REAL
// con las seis migraciones aplicadas (incluida
// 20260813120000_create_auth_credentials_and_sessions.sql) y
// database/testdata/dos_barberias.sql +
// database/testdata/hu005_credenciales_sesiones.sql cargados. Conéctate como
// barberia_app, igual que internal/platform/idempotency
// (estrategia-pruebas.md §2 prohíbe mocks para RLS/privilegios).
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/postgres?sslmode=disable"
//	go test -race ./internal/modules/auth/postgres/...
package postgres_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"sync"
	"testing"
	"time"

	"system-barbershop/internal/modules/auth"
	authpostgres "system-barbershop/internal/modules/auth/postgres"
	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
)

// uniqueToken evita que reejecuciones de la suite contra la misma base de
// datos compartida colisionen con filas dejadas por una corrida anterior
// (staff_session_token_hash_uk es única): cada prueba reclama un token
// propio, nunca reutilizado, igual que idempotency/postgres_test.go.
func uniqueToken(t *testing.T, label string) string {
	t.Helper()
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	return "test-" + label + "-" + hex.EncodeToString(buf)
}

const (
	testDatabaseURL = "postgres://barberia_app@localhost:5432/barberia_test?sslmode=disable"

	shopA = "11111111-1111-1111-1111-111111111111"
	shopB = "22222222-2222-2222-2222-222222222222"

	emailActiveA   = "duena.a@ejemplo.test"
	emailActiveB   = "dueno.b@ejemplo.test"
	emailInactiveB = "barbero.b@ejemplo.test" // is_active=false en dos_barberias.sql
	emailUnknown   = "no-existe-en-ningun-lado@ejemplo.test"

	staffUserActiveA = "aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaa1"
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
	return db
}

func TestResolveLoginTenant_ActiveUser_ResolvesOwnShop(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.New(db)

	shop, found, err := repo.ResolveLoginTenant(context.Background(), emailActiveA)
	if err != nil {
		t.Fatalf("ResolveLoginTenant: %v", err)
	}
	if !found || shop != shopA {
		t.Fatalf("expected (shopA, true), got (%q, %v)", shop, found)
	}

	shop, found, err = repo.ResolveLoginTenant(context.Background(), emailActiveB)
	if err != nil {
		t.Fatalf("ResolveLoginTenant: %v", err)
	}
	if !found || shop != shopB {
		t.Fatalf("expected (shopB, true), got (%q, %v)", shop, found)
	}
}

// TestResolveLoginTenant_InactiveAndUnknown_BothReturnNotFound cubre
// CA-005-02/CA-005-07 a nivel de repositorio real: un usuario inactivo y un
// correo inexistente son indistinguibles (found=false para ambos).
func TestResolveLoginTenant_InactiveAndUnknown_BothReturnNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.New(db)

	_, foundInactive, err := repo.ResolveLoginTenant(context.Background(), emailInactiveB)
	if err != nil {
		t.Fatalf("ResolveLoginTenant (inactive): %v", err)
	}
	if foundInactive {
		t.Fatal("expected an inactive user to resolve as not found")
	}

	_, foundUnknown, err := repo.ResolveLoginTenant(context.Background(), emailUnknown)
	if err != nil {
		t.Fatalf("ResolveLoginTenant (unknown): %v", err)
	}
	if foundUnknown {
		t.Fatal("expected an unknown email to resolve as not found")
	}
}

func TestLookupCredential_ResolvedTenant_ReturnsMatchingCredential(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.New(db)

	cred, err := repo.LookupCredential(context.Background(), shopA, true, emailActiveA)
	if err != nil {
		t.Fatalf("LookupCredential: %v", err)
	}
	if !cred.Found {
		t.Fatal("expected the fixture credential to be found")
	}
	if cred.StaffUserID != staffUserActiveA {
		t.Fatalf("expected staff_user_id %q, got %q", staffUserActiveA, cred.StaffUserID)
	}
	if cred.PasswordAlgorithm != "argon2id" {
		t.Fatalf("expected argon2id, got %q", cred.PasswordAlgorithm)
	}
	if cred.PasswordHash == "" {
		t.Fatal("expected a non-empty password hash")
	}
}

// TestLookupCredential_UnresolvedTenant_UsesDecoyWithoutError es la
// evidencia de PostgreSQL real de CA-005-02: el camino "no resuelto" ejecuta
// la MISMA forma de consulta que el resuelto (decoyBarbershopID +
// decoyStaffUserID), sin error, y sin encontrar nada.
func TestLookupCredential_UnresolvedTenant_UsesDecoyWithoutError(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.New(db)

	cred, err := repo.LookupCredential(context.Background(), "00000000-0000-0000-0000-000000000000", false, emailUnknown)
	if err != nil {
		t.Fatalf("LookupCredential (decoy path): %v", err)
	}
	if cred.Found {
		t.Fatal("expected the decoy path to never find a credential")
	}
}

// TestLookupCredential_CrossTenant_NeverLeaksAcrossShops verifica RLS en
// PostgreSQL real: pedir el correo de B mientras el tenant abierto es A no
// puede devolver la credencial de B (RN-TEN-01), aunque el llamador pase
// resolved=true por error.
func TestLookupCredential_CrossTenant_NeverLeaksAcrossShops(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.New(db)

	cred, err := repo.LookupCredential(context.Background(), shopA, true, emailActiveB)
	if err != nil {
		t.Fatalf("LookupCredential: %v", err)
	}
	if cred.Found {
		t.Fatal("RN-TEN-01: found a shopB credential while scoped to shopA")
	}
}

func TestCreateSession_PersistsRetrievableSessionScopedToTenant(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.New(db)

	tokenHash := auth.HashToken(uniqueToken(t, "persist"))
	issuedAt := time.Now().UTC().Truncate(time.Second)
	expiresAt := issuedAt.Add(auth.SessionDuration)

	if err := repo.CreateSession(context.Background(), shopA, staffUserActiveA, tokenHash, issuedAt, expiresAt); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// Verifica, con el mismo patrón InTenantTx que usa el resto del
	// proyecto, que la fila existe dentro del tenant correcto y NO es
	// visible desde el tenant ajeno (RLS).
	var countInA int
	err := db.InTenantTx(context.Background(), database.BarbershopID(shopA), func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx, `SELECT count(*) FROM staff_session WHERE token_hash = $1`, tokenHash).Scan(&countInA)
	})
	if err != nil {
		t.Fatalf("verify in shopA: %v", err)
	}
	if countInA != 1 {
		t.Fatalf("expected exactly 1 session visible in shopA, got %d", countInA)
	}

	var countInB int
	err = db.InTenantTx(context.Background(), database.BarbershopID(shopB), func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx, `SELECT count(*) FROM staff_session WHERE token_hash = $1`, tokenHash).Scan(&countInB)
	})
	if err != nil {
		t.Fatalf("verify in shopB: %v", err)
	}
	if countInB != 0 {
		t.Fatalf("CA-005-05: session created for A is visible from B (%d rows)", countInB)
	}
}

// TestCreateSession_ConcurrentAcrossTenants_NoDataRace ejercita el
// repositorio con dos tenants distintos en paralelo, útil bajo `go test
// -race`: no protege ninguna invariante de negocio nueva (a diferencia de
// idempotency, el login no reclama un recurso compartido), pero confirma
// que el pool y InTenantTx no tienen residuo de contexto entre solicitudes
// concurrentes de tenants distintos (ya cubierto en general por
// internal/platform/database, reforzado aquí para el camino real de auth).
func TestCreateSession_ConcurrentAcrossTenants_NoDataRace(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.New(db)

	const staffUserB = "bbbbbbb1-bbbb-bbbb-bbbb-bbbbbbbbbbb1"

	// Los tokens únicos se generan en la goroutine de la prueba, ANTES de
	// lanzar las goroutines concurrentes: t.Fatalf (dentro de uniqueToken)
	// solo puede llamarse desde la goroutine que ejecuta la prueba, nunca
	// desde una goroutine creada por ella.
	const n = 10
	tokensA := make([]string, n)
	tokensB := make([]string, n)
	for i := 0; i < n; i++ {
		tokensA[i] = auth.HashToken(uniqueToken(t, "race-a"))
		tokensB[i] = auth.HashToken(uniqueToken(t, "race-b"))
	}

	var wg sync.WaitGroup
	errs := make(chan error, 2*n)
	for i := 0; i < n; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			issuedAt := time.Now().UTC()
			err := repo.CreateSession(context.Background(), shopA, staffUserActiveA,
				tokensA[i], issuedAt, issuedAt.Add(auth.SessionDuration))
			errs <- err
		}(i)
		go func(i int) {
			defer wg.Done()
			issuedAt := time.Now().UTC()
			err := repo.CreateSession(context.Background(), shopB, staffUserB,
				tokensB[i], issuedAt, issuedAt.Add(auth.SessionDuration))
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent CreateSession failed: %v", err)
		}
	}
}
