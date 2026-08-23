// Pruebas de integración de HU-007 (CA-007-06) contra
// login_throttle_purge_expired/auth_phone_challenge_purge_expired con
// PostgreSQL REAL, conectado como barberia_worker (DEC-040): estas
// funciones NO están concedidas a barberia_app.
//
//	export TEST_WORKER_DATABASE_URL="postgres://barberia_worker:worker_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/auth/postgres/...
package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	authpostgres "system-barbershop/internal/modules/auth/postgres"
	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
)

const testWorkerDatabaseURL = "postgres://barberia_worker@localhost:5432/barberia_test?sslmode=disable"

func setupWorkerTestDB(t *testing.T) *database.DB {
	t.Helper()
	url := os.Getenv("TEST_WORKER_DATABASE_URL")
	if url == "" {
		url = testWorkerDatabaseURL
	}
	cfg := config.Config{
		Environment:              "test",
		DatabaseMaxConns:         5,
		DatabaseMinConns:         1,
		DatabaseMaxConnLifetime:  time.Hour,
		DatabaseMaxConnIdleTime:  30 * time.Minute,
		DatabaseConnectTimeout:   5 * time.Second,
		DatabaseStatementTimeout: 10 * time.Second,
	}
	db, err := database.NewDB(config.DatabaseDSN(url), cfg)
	if err != nil {
		t.Fatalf("database.NewDB (worker): %v", err)
	}
	return db
}

func TestPurgeRepository_PurgeLoginThrottle_DeletesOnlyExpired(t *testing.T) {
	appDB := setupTestDB(t)
	defer appDB.Close()
	throttleRepo := authpostgres.NewThrottleRepository(appDB)

	// Fila vigente: no debe purgarse.
	ipHash := testThrottleIPHash(t, "purge-not-expired")
	if _, _, _, err := throttleRepo.RegisterAttempt(context.Background(), ipHash, testThrottleCfg()); err != nil {
		t.Fatalf("RegisterAttempt: %v", err)
	}

	workerDB := setupWorkerTestDB(t)
	defer workerDB.Close()
	purgeRepo := authpostgres.NewPurgeRepository(workerDB)

	// No se afirma nada sobre el conteo devuelto (otras pruebas del
	// paquete pueden haber dejado filas vencidas de ejecuciones previas);
	// lo que importa es que la fila recién creada, vigente, siga
	// existiendo después de purgar.
	if _, err := purgeRepo.PurgeLoginThrottle(context.Background(), 1000); err != nil {
		t.Fatalf("PurgeLoginThrottle: %v", err)
	}

	count, _, _, err := throttleRepo.RegisterAttempt(context.Background(), ipHash, testThrottleCfg())
	if err != nil {
		t.Fatalf("RegisterAttempt (post-purge check): %v", err)
	}
	if count != 2 {
		t.Fatalf("expected the still-valid row to survive purge (count=2 on next attempt), got %d", count)
	}
}

func TestPurgeRepository_PurgePhoneChallenges_RunsWithoutError(t *testing.T) {
	workerDB := setupWorkerTestDB(t)
	defer workerDB.Close()
	purgeRepo := authpostgres.NewPurgeRepository(workerDB)

	if _, err := purgeRepo.PurgePhoneChallenges(context.Background(), 1000); err != nil {
		t.Fatalf("PurgePhoneChallenges: %v", err)
	}
}

// TestPurgeRepository_PurgeRecoveryCodes_RunsWithoutError cubre HU-008:
// auth_recovery_purge_expired concedida exclusivamente a barberia_worker.
func TestPurgeRepository_PurgeRecoveryCodes_RunsWithoutError(t *testing.T) {
	workerDB := setupWorkerTestDB(t)
	defer workerDB.Close()
	purgeRepo := authpostgres.NewPurgeRepository(workerDB)

	if _, err := purgeRepo.PurgeRecoveryCodes(context.Background(), 1000); err != nil {
		t.Fatalf("PurgeRecoveryCodes: %v", err)
	}
}
