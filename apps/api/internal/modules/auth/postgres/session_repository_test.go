// Package postgres_test (pruebas de integración, HU-006) requiere
// PostgreSQL REAL con las seis migraciones aplicadas y
// database/testdata/dos_barberias.sql + testdata/hu005_credenciales_sesiones.sql
// cargados, igual que repository_test.go. Conéctate como barberia_app.
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/auth/postgres/...
package postgres_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"system-barbershop/internal/modules/auth"
	authpostgres "system-barbershop/internal/modules/auth/postgres"
	"system-barbershop/internal/platform/database"
)

const (
	staffUserActiveB   = "bbbbbbb1-bbbb-bbbb-bbbb-bbbbbbbbbbb1"
	staffUserInactiveB = "bbbbbbb2-bbbb-bbbb-bbbb-bbbbbbbbbbb2" // is_active=false en dos_barberias.sql
)

// createTestSession inserta una sesión propia (token único por prueba) y
// devuelve su hash y, consultando la fila insertada, su id real: CreateSession
// no lo devuelve (staff_session.id usa gen_random_uuid() por defecto).
func createTestSession(t *testing.T, db *database.DB, repo *authpostgres.Repository, shop, staffUserID string, issuedAt, expiresAt time.Time, label string) (tokenHash, sessionID string) {
	t.Helper()
	tokenHash = auth.HashToken(uniqueToken(t, label))
	if err := repo.CreateSession(context.Background(), shop, staffUserID, tokenHash, issuedAt, expiresAt); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	err := db.InTenantTx(context.Background(), database.BarbershopID(shop), func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx, `SELECT id FROM staff_session WHERE token_hash = $1`, tokenHash).Scan(&sessionID)
	})
	if err != nil {
		t.Fatalf("lookup created session id: %v", err)
	}
	return tokenHash, sessionID
}

func readSessionRow(t *testing.T, db *database.DB, shop, sessionID string) (revokedAt *time.Time, expiresAt, lastUsedAt time.Time) {
	t.Helper()
	err := db.InTenantTx(context.Background(), database.BarbershopID(shop), func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx,
			`SELECT revoked_at, expires_at, last_used_at FROM staff_session WHERE id = $1`, sessionID,
		).Scan(&revokedAt, &expiresAt, &lastUsedAt)
	})
	if err != nil {
		t.Fatalf("read session row: %v", err)
	}
	return revokedAt, expiresAt, lastUsedAt
}

func TestValidateAndRenewSession_ValidSession_RenewsLastUsedAndExpiresAt(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.New(db)

	issuedAt := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second)
	oldExpiresAt := issuedAt.Add(auth.SessionDuration)
	tokenHash, sessionID := createTestSession(t, db, repo, shopA, staffUserActiveA, issuedAt, oldExpiresAt, "renew")

	now := time.Now().UTC().Truncate(time.Second)
	newExpiresAt := now.Add(auth.SessionDuration)

	principal, found, err := repo.ValidateAndRenewSession(context.Background(), shopA, tokenHash, now, newExpiresAt)
	if err != nil {
		t.Fatalf("ValidateAndRenewSession: %v", err)
	}
	if !found {
		t.Fatal("expected the fresh session to be found and renewed")
	}
	if principal.SessionID != sessionID || principal.StaffUserID != staffUserActiveA || principal.BarbershopID != shopA {
		t.Fatalf("unexpected principal: %+v (want session=%s user=%s shop=%s)", principal, sessionID, staffUserActiveA, shopA)
	}

	_, gotExpiresAt, gotLastUsedAt := readSessionRow(t, db, shopA, sessionID)
	if !gotExpiresAt.Equal(newExpiresAt) {
		t.Fatalf("expected expires_at to advance to %v, got %v", newExpiresAt, gotExpiresAt)
	}
	if !gotLastUsedAt.Equal(now) {
		t.Fatalf("expected last_used_at to advance to %v, got %v", now, gotLastUsedAt)
	}
}

func TestValidateAndRenewSession_UnknownTokenHash_ReturnsNotFoundWithoutError(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.New(db)

	_, found, err := repo.ValidateAndRenewSession(context.Background(), shopA, auth.HashToken(uniqueToken(t, "unknown")), time.Now().UTC(), time.Now().UTC().Add(auth.SessionDuration))
	if err != nil {
		t.Fatalf("ValidateAndRenewSession: %v", err)
	}
	if found {
		t.Fatal("expected an unknown token hash to never be found")
	}
}

// TestValidateAndRenewSession_ExpiredSession_NeverRenewsOrReopens cubre
// CA-006-03: una sesión vencida no se renueva sola, y la fila permanece sin
// cambios (ninguna actualización "revive" expires_at ni last_used_at).
func TestValidateAndRenewSession_ExpiredSession_NeverRenewsOrReopens(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.New(db)

	issuedAt := time.Now().UTC().Add(-40 * 24 * time.Hour).Truncate(time.Second)
	expiredAt := issuedAt.Add(10 * 24 * time.Hour) // ya vencida respecto a "ahora"
	tokenHash, sessionID := createTestSession(t, db, repo, shopA, staffUserActiveA, issuedAt, expiredAt, "expired")

	now := time.Now().UTC().Truncate(time.Second)
	_, found, err := repo.ValidateAndRenewSession(context.Background(), shopA, tokenHash, now, now.Add(auth.SessionDuration))
	if err != nil {
		t.Fatalf("ValidateAndRenewSession: %v", err)
	}
	if found {
		t.Fatal("CA-006-03: an expired session must never be found/renewed")
	}

	_, gotExpiresAt, gotLastUsedAt := readSessionRow(t, db, shopA, sessionID)
	if !gotExpiresAt.Equal(expiredAt) {
		t.Fatalf("expected expires_at to stay at %v, got %v (renewal must not touch an expired row)", expiredAt, gotExpiresAt)
	}
	if !gotLastUsedAt.Equal(issuedAt) {
		t.Fatalf("expected last_used_at to stay at %v, got %v", issuedAt, gotLastUsedAt)
	}
}

// TestValidateAndRenewSession_RevokedSession_NeverRenewsOrReopens cubre
// CA-006-02: reutilizar el material tras un cierre de sesión falla, y la
// renovación nunca borra revoked_at.
func TestValidateAndRenewSession_RevokedSession_NeverRenewsOrReopens(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.New(db)

	issuedAt := time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Second)
	tokenHash, sessionID := createTestSession(t, db, repo, shopA, staffUserActiveA, issuedAt, issuedAt.Add(auth.SessionDuration), "revoked")

	revokedAt := time.Now().UTC().Truncate(time.Second)
	if err := repo.RevokeSession(context.Background(), shopA, sessionID, revokedAt); err != nil {
		t.Fatalf("RevokeSession: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	_, found, err := repo.ValidateAndRenewSession(context.Background(), shopA, tokenHash, now, now.Add(auth.SessionDuration))
	if err != nil {
		t.Fatalf("ValidateAndRenewSession: %v", err)
	}
	if found {
		t.Fatal("expected a revoked session to never be found/renewed")
	}

	gotRevokedAt, _, _ := readSessionRow(t, db, shopA, sessionID)
	if gotRevokedAt == nil || !gotRevokedAt.Equal(revokedAt) {
		t.Fatalf("expected revoked_at to remain %v, got %v (renewal must never clear it)", revokedAt, gotRevokedAt)
	}
}

// TestValidateAndRenewSession_InactiveUser_NeverRenews cubre la parte de
// CA-005-07 que HU-005 dejó bloqueada: un usuario desactivado no conserva
// sesiones vigentes, ahora verificable de extremo a extremo porque
// ValidateAndRenewSession reconfirma is_active en cada uso.
func TestValidateAndRenewSession_InactiveUser_NeverRenews(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.New(db)

	issuedAt := time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Second)
	tokenHash, _ := createTestSession(t, db, repo, shopB, staffUserInactiveB, issuedAt, issuedAt.Add(auth.SessionDuration), "inactive-user")

	now := time.Now().UTC().Truncate(time.Second)
	_, found, err := repo.ValidateAndRenewSession(context.Background(), shopB, tokenHash, now, now.Add(auth.SessionDuration))
	if err != nil {
		t.Fatalf("ValidateAndRenewSession: %v", err)
	}
	if found {
		t.Fatal("expected a session belonging to an inactive user to never be found/renewed")
	}
}

// TestValidateAndRenewSession_WrongTenantContext_NeverLeaksAcrossShops
// verifica RLS en PostgreSQL real: abrir la transacción con el tenant
// equivocado nunca encuentra (ni renueva) la sesión real de otro tenant
// (RN-TEN-01), incluso si el token hash coincide exactamente.
func TestValidateAndRenewSession_WrongTenantContext_NeverLeaksAcrossShops(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.New(db)

	issuedAt := time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Second)
	tokenHash, sessionID := createTestSession(t, db, repo, shopA, staffUserActiveA, issuedAt, issuedAt.Add(auth.SessionDuration), "cross-tenant")

	now := time.Now().UTC().Truncate(time.Second)
	_, found, err := repo.ValidateAndRenewSession(context.Background(), shopB, tokenHash, now, now.Add(auth.SessionDuration))
	if err != nil {
		t.Fatalf("ValidateAndRenewSession: %v", err)
	}
	if found {
		t.Fatal("RN-TEN-01: found/renewed shopA's session while scoped to shopB")
	}

	// La fila real de A permanece intacta.
	_, gotExpiresAt, _ := readSessionRow(t, db, shopA, sessionID)
	if !gotExpiresAt.Equal(issuedAt.Add(auth.SessionDuration)) {
		t.Fatal("expected shopA's session to remain untouched by a shopB-scoped attempt")
	}
}

func TestRevokeSession_RevokesOwnSession_AndIsIdempotent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.New(db)

	issuedAt := time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Second)
	_, sessionID := createTestSession(t, db, repo, shopA, staffUserActiveA, issuedAt, issuedAt.Add(auth.SessionDuration), "idempotent-logout")

	first := time.Now().UTC().Truncate(time.Second)
	if err := repo.RevokeSession(context.Background(), shopA, sessionID, first); err != nil {
		t.Fatalf("RevokeSession (first): %v", err)
	}
	gotRevokedAt, _, _ := readSessionRow(t, db, shopA, sessionID)
	if gotRevokedAt == nil || !gotRevokedAt.Equal(first) {
		t.Fatalf("expected revoked_at=%v, got %v", first, gotRevokedAt)
	}

	// CA-006-02: repetir el logout con la misma sesión ya revocada no debe
	// fallar ni volver a mover revoked_at.
	second := first.Add(time.Minute)
	if err := repo.RevokeSession(context.Background(), shopA, sessionID, second); err != nil {
		t.Fatalf("RevokeSession (second, already revoked): %v", err)
	}
	gotRevokedAt, _, _ = readSessionRow(t, db, shopA, sessionID)
	if gotRevokedAt == nil || !gotRevokedAt.Equal(first) {
		t.Fatalf("expected revoked_at to stay at the FIRST revocation (%v), got %v", first, gotRevokedAt)
	}
}

// TestRevokeSession_WrongTenant_NeverRevokesAnotherShopsSession es la
// prueba de repositorio que respalda CA-006-07/DEC-058: aunque
// RevokeSession recibiera el id real de una sesión de otra barbería (nunca
// ocurre en producción: sessionID sale siempre de un auth.Principal ya
// validado para ESE tenant), RLS y el filtro explícito de barbershop_id
// hacen que la operación no afecte ninguna fila.
func TestRevokeSession_WrongTenant_NeverRevokesAnotherShopsSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.New(db)

	issuedAt := time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Second)
	_, sessionIDOfA := createTestSession(t, db, repo, shopA, staffUserActiveA, issuedAt, issuedAt.Add(auth.SessionDuration), "victim-a")

	// Intento con el tenant B, pero el id real de la sesión de A.
	if err := repo.RevokeSession(context.Background(), shopB, sessionIDOfA, time.Now().UTC()); err != nil {
		t.Fatalf("RevokeSession (cross-tenant attempt): %v", err)
	}

	gotRevokedAt, _, _ := readSessionRow(t, db, shopA, sessionIDOfA)
	if gotRevokedAt != nil {
		t.Fatalf("CA-006-07: shopA's session was revoked by a shopB-scoped call, revoked_at=%v", gotRevokedAt)
	}
}

// TestValidateAndRenewSession_ConcurrentWithRevoke_FinalStateAlwaysRevoked
// es la prueba de carrera real exigida por el prompt: sin sleep, varias
// renovaciones concurrentes contra una revocación real, coordinadas con un
// WaitGroup para maximizar la superposición. El resultado final siempre
// debe quedar revocado -ninguna renovación posterior reabre la sesión ni
// borra revoked_at- porque ValidateAndRenewSession solo actualiza
// last_used_at/expires_at cuando su propio WHERE (revoked_at IS NULL)
// todavía se cumple en el momento en que PostgreSQL evalúa la sentencia.
func TestValidateAndRenewSession_ConcurrentWithRevoke_FinalStateAlwaysRevoked(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.New(db)

	issuedAt := time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Second)
	tokenHash, sessionID := createTestSession(t, db, repo, shopA, staffUserActiveA, issuedAt, issuedAt.Add(auth.SessionDuration), "race")

	const renewers = 20
	var ready sync.WaitGroup
	ready.Add(renewers + 1)
	start := make(chan struct{})
	var wg sync.WaitGroup
	errs := make(chan error, renewers+1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		ready.Done()
		<-start
		errs <- repo.RevokeSession(context.Background(), shopA, sessionID, time.Now().UTC())
	}()

	for i := 0; i < renewers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ready.Done()
			<-start
			now := time.Now().UTC()
			_, _, err := repo.ValidateAndRenewSession(context.Background(), shopA, tokenHash, now, now.Add(auth.SessionDuration))
			errs <- err
		}()
	}

	ready.Wait()
	close(start)
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent call failed: %v", err)
		}
	}

	gotRevokedAt, _, _ := readSessionRow(t, db, shopA, sessionID)
	if gotRevokedAt == nil {
		t.Fatal("expected the session to end up revoked after the race")
	}

	// Una renovación posterior, ya asentada la carrera, confirma que nada
	// reabrió la sesión.
	now := time.Now().UTC()
	_, found, err := repo.ValidateAndRenewSession(context.Background(), shopA, tokenHash, now, now.Add(auth.SessionDuration))
	if err != nil {
		t.Fatalf("post-race ValidateAndRenewSession: %v", err)
	}
	if found {
		t.Fatal("expected the session to remain revoked after the race, but a later renewal found it valid")
	}
}

// --- BarbershopName (HU-012, DEC-060) ------------------------------------

// TestBarbershopName_ReturnsRealNamePerTenant confirma que el nombre
// devuelto es la columna real barbershop.name de database/testdata/dos_barberias.sql
// para cada tenant, no un valor inventado ni compartido entre barberías.
func TestBarbershopName_ReturnsRealNamePerTenant(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.New(db)

	nameA, err := repo.BarbershopName(context.Background(), shopA)
	if err != nil {
		t.Fatalf("BarbershopName(shopA): %v", err)
	}
	if nameA != "Barbería de prueba A" {
		t.Fatalf("expected %q, got %q", "Barbería de prueba A", nameA)
	}

	nameB, err := repo.BarbershopName(context.Background(), shopB)
	if err != nil {
		t.Fatalf("BarbershopName(shopB): %v", err)
	}
	if nameB != "Barbería de prueba B" {
		t.Fatalf("expected %q, got %q", "Barbería de prueba B", nameB)
	}
}

// TestBarbershopName_UnknownID_ReturnsError confirma que un identificador
// de barbería inexistente no produce un nombre vacío silencioso: falla de
// forma explícita, igual que el resto de operaciones tenant-aware
// (CA-001-03).
func TestBarbershopName_UnknownID_ReturnsError(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.New(db)

	const unknownShop = "99999999-9999-9999-9999-999999999999"
	if _, err := repo.BarbershopName(context.Background(), unknownShop); err == nil {
		t.Fatal("expected an error for an unknown barbershop id, got nil")
	}
}
