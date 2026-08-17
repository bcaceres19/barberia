// Pruebas de integración de HU-007 contra login_throttle_register_attempt
// con PostgreSQL REAL. Requieren las siete migraciones aplicadas (incluida
// 20260817180000_create_login_throttle_and_phone_challenge.sql). Conéctate
// como barberia_app, igual que repository_test.go/session_repository_test.go.
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/auth/postgres/...
package postgres_test

import (
	"context"
	"sync"
	"testing"

	"system-barbershop/internal/modules/auth"
	authpostgres "system-barbershop/internal/modules/auth/postgres"
)

func testThrottleIPHash(t *testing.T, label string) string {
	t.Helper()
	return auth.HMACHex(uniqueToken(t, label), []byte("secreto-de-prueba-repositorio-throttle"))
}

func testThrottleCfg() auth.ThrottleConfig {
	return auth.ThrottleConfig{WindowSeconds: 900, EscalationSeconds: 86400, Threshold: 5, RetentionSeconds: 172800}
}

// TestThrottleRepository_SixthAttempt_Escalates cubre DEC-061 (CT-005)
// contra PostgreSQL real: las cinco primeras solicitudes NO escalan; la
// sexta sí.
func TestThrottleRepository_SixthAttempt_Escalates(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.NewThrottleRepository(db)
	ipHash := testThrottleIPHash(t, "sixth-escalates")
	cfg := testThrottleCfg()

	for i := 1; i <= 5; i++ {
		count, escalated, _, err := repo.RegisterAttempt(context.Background(), ipHash, cfg)
		if err != nil {
			t.Fatalf("attempt %d: unexpected error: %v", i, err)
		}
		if count != i {
			t.Fatalf("attempt %d: expected attempt_count=%d, got %d", i, i, count)
		}
		if escalated {
			t.Fatalf("attempt %d: expected NOT escalated (DEC-061: threshold=5 allows 5 normal requests)", i)
		}
	}

	count, escalated, retryAfter, err := repo.RegisterAttempt(context.Background(), ipHash, cfg)
	if err != nil {
		t.Fatalf("6th attempt: unexpected error: %v", err)
	}
	if count != 6 {
		t.Fatalf("expected attempt_count=6, got %d", count)
	}
	if !escalated {
		t.Fatal("expected the 6th attempt to escalate (DEC-061/CT-005)")
	}
	if retryAfter.IsZero() {
		t.Fatal("expected a non-zero retryAfter once escalated")
	}
}

// TestThrottleRepository_ConcurrentAttempts_SameIP_NoLostIncrements es la
// prueba de concurrencia real exigida por el prompt: 40+ llamadas
// concurrentes coordinadas sin sleep, sin incrementos perdidos.
func TestThrottleRepository_ConcurrentAttempts_SameIP_NoLostIncrements(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.NewThrottleRepository(db)
	ipHash := testThrottleIPHash(t, "concurrent-same-ip")
	// Umbral alto para que la carrera no dependa de en qué orden llegan las
	// escrituras a superar el umbral: lo único que se mide aquí es que las
	// 40 escrituras concurrentes se reflejen todas, sin perder ninguna.
	cfg := auth.ThrottleConfig{WindowSeconds: 900, EscalationSeconds: 86400, Threshold: 1000, RetentionSeconds: 172800}

	const n = 45
	var wg sync.WaitGroup
	errCh := make(chan error, n)
	wg.Add(n)
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			<-start
			_, _, _, err := repo.RegisterAttempt(context.Background(), ipHash, cfg)
			errCh <- err
		}()
	}
	close(start)
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent RegisterAttempt failed: %v", err)
		}
	}

	finalCount, _, _, err := repo.RegisterAttempt(context.Background(), ipHash, cfg)
	if err != nil {
		t.Fatalf("final RegisterAttempt: %v", err)
	}
	if finalCount != n+1 {
		t.Fatalf("expected attempt_count=%d after %d concurrent + 1 final call (no lost increments), got %d", n+1, n, finalCount)
	}
}

// TestThrottleRepository_ConcurrentAttempts_DifferentIPs_Independent
// confirma que dos IP distintas no interfieren entre sí bajo concurrencia.
func TestThrottleRepository_ConcurrentAttempts_DifferentIPs_Independent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.NewThrottleRepository(db)
	ipHashA := testThrottleIPHash(t, "independent-a")
	ipHashB := testThrottleIPHash(t, "independent-b")
	cfg := testThrottleCfg()

	var wg sync.WaitGroup
	wg.Add(2)
	var errA, errB error
	go func() {
		defer wg.Done()
		_, _, _, errA = repo.RegisterAttempt(context.Background(), ipHashA, cfg)
	}()
	go func() {
		defer wg.Done()
		_, _, _, errB = repo.RegisterAttempt(context.Background(), ipHashB, cfg)
	}()
	wg.Wait()

	if errA != nil || errB != nil {
		t.Fatalf("unexpected errors: a=%v b=%v", errA, errB)
	}

	countA, _, _, err := repo.RegisterAttempt(context.Background(), ipHashA, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if countA != 2 {
		t.Fatalf("expected ipHashA to have attempt_count=2, got %d (independent counters must not merge)", countA)
	}
}
