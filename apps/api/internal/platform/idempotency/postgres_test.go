// Package idempotency_test (pruebas de integración) requiere PostgreSQL
// REAL con las cinco migraciones aplicadas y database/testdata/dos_barberias.sql
// cargado — igual que internal/platform/database (estrategia-pruebas.md §2
// prohíbe mocks para reglas concurrentes; ddl_hardening_patterns.md, memoria
// del proyecto). Conéctate como barberia_app.
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/postgres?sslmode=disable"
//	go test -race ./internal/platform/idempotency/...
package idempotency_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

const (
	testDatabaseURL = "postgres://barberia_app@localhost:5432/barberia_test?sslmode=disable"
	shopA           = database.BarbershopID("11111111-1111-1111-1111-111111111111")
	shopB           = database.BarbershopID("22222222-2222-2222-2222-222222222222")
)

// setupTestDB abre un *database.DB real por la vía pública (database.NewDB),
// igual que cmd/api: no existe un atajo propio de este paquete para saltarse
// InTenantTx ni sus hooks de residuo de contexto.
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

// uniqueKey evita que reejecuciones de la suite contra la misma base de
// datos compartida colisionen con filas dejadas por una corrida anterior:
// cada prueba reclama una clave propia, nunca reutilizada.
func uniqueKey(t *testing.T, label string) idempotency.Key {
	t.Helper()
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	raw := fmt.Sprintf("test-%s-%s", label, hex.EncodeToString(buf))
	key, err := idempotency.ParseKey(raw)
	if err != nil {
		t.Fatalf("unexpected invalid generated key %q: %v", raw, err)
	}
	return key
}

func fingerprintOf(body string) idempotency.Fingerprint {
	return idempotency.ComputeFingerprint("POST", "/api/v1/private/test-effect", []byte(body))
}

// TestBegin_FirstExecution_ProceedsAndComplete_PersistsExactResponse cubre
// CA-004-01 (primera mitad: ejecución nueva) verificando que Complete deja
// la fila lista para una reproducción posterior.
func TestBegin_FirstExecution_ProceedsAndComplete_PersistsExactResponse(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	coord := idempotency.NewSQLCoordinator()

	key := uniqueKey(t, "first-exec")
	fp := fingerprintOf(`{"resource":"x"}`)
	stored := idempotency.StoredResponse{Status: 201, ContentType: "application/json", Body: `{"id":"r-1"}`}

	err := db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
		decision, err := coord.Begin(ctx, q, shopA, key, "test_op", fp, 60*time.Second)
		if err != nil {
			return err
		}
		if decision.Outcome != idempotency.OutcomeProceed {
			return fmt.Errorf("expected OutcomeProceed, got %s", decision.Outcome)
		}

		ok, err := coord.Complete(ctx, q, shopA, key, stored)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("expected Complete to report ok=true for a freshly claimed key")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("InTenantTx: %v", err)
	}
}

// TestBegin_SameKeyAndFingerprint_Replays cubre CA-004-01 completo: repetir
// la misma clave con el mismo contenido reproduce la respuesta original sin
// ejecutar el efecto de nuevo (aquí, sin volver a llamar Complete).
func TestBegin_SameKeyAndFingerprint_Replays(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	coord := idempotency.NewSQLCoordinator()

	key := uniqueKey(t, "replay")
	fp := fingerprintOf(`{"resource":"y"}`)
	stored := idempotency.StoredResponse{Status: 201, ContentType: "application/json", Body: `{"id":"r-2"}`}

	// Primera ejecución: reclama y completa.
	err := db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
		decision, err := coord.Begin(ctx, q, shopA, key, "test_op", fp, 60*time.Second)
		if err != nil {
			return err
		}
		if decision.Outcome != idempotency.OutcomeProceed {
			return fmt.Errorf("expected OutcomeProceed, got %s", decision.Outcome)
		}
		_, err = coord.Complete(ctx, q, shopA, key, stored)
		return err
	})
	if err != nil {
		t.Fatalf("first InTenantTx: %v", err)
	}

	// Repetición: misma clave, mismo fingerprint.
	var replay idempotency.Decision
	err = db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
		decision, err := coord.Begin(ctx, q, shopA, key, "test_op", fp, 60*time.Second)
		if err != nil {
			return err
		}
		replay = decision
		return nil
	})
	if err != nil {
		t.Fatalf("second InTenantTx: %v", err)
	}

	if replay.Outcome != idempotency.OutcomeReplay {
		t.Fatalf("expected OutcomeReplay, got %s", replay.Outcome)
	}
	if replay.Response != stored {
		t.Fatalf("expected the exact original response %+v, got %+v", stored, replay.Response)
	}
	if err := replay.AsError(); err != nil {
		t.Fatalf("expected AsError to be nil for a replay, got %v", err)
	}
}

// TestBegin_DifferentFingerprint_ConflictsWithoutEffect cubre CA-004-02.
func TestBegin_DifferentFingerprint_ConflictsWithoutEffect(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	coord := idempotency.NewSQLCoordinator()

	key := uniqueKey(t, "fp-conflict")
	original := fingerprintOf(`{"resource":"a"}`)
	different := fingerprintOf(`{"resource":"b"}`)
	stored := idempotency.StoredResponse{Status: 201, ContentType: "application/json", Body: `{"id":"r-3"}`}

	err := db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
		decision, err := coord.Begin(ctx, q, shopA, key, "test_op", original, 60*time.Second)
		if err != nil {
			return err
		}
		if decision.Outcome != idempotency.OutcomeProceed {
			return fmt.Errorf("expected OutcomeProceed, got %s", decision.Outcome)
		}
		_, err = coord.Complete(ctx, q, shopA, key, stored)
		return err
	})
	if err != nil {
		t.Fatalf("first InTenantTx: %v", err)
	}

	var conflict idempotency.Decision
	err = db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
		decision, err := coord.Begin(ctx, q, shopA, key, "test_op", different, 60*time.Second)
		if err != nil {
			return err
		}
		conflict = decision
		return nil
	})
	if err != nil {
		t.Fatalf("second InTenantTx: %v", err)
	}

	if conflict.Outcome != idempotency.OutcomeConflictFingerprint {
		t.Fatalf("expected OutcomeConflictFingerprint, got %s", conflict.Outcome)
	}
	appErr, ok := apperr.As(conflict.AsError())
	if !ok || appErr.Kind != apperr.KindIdempotencyConflict {
		t.Fatalf("expected KindIdempotencyConflict, got %v", conflict.AsError())
	}
}

// TestBegin_DifferentOperation_Conflicts cubre RN-IDE-01: una clave ya
// usada no se reutiliza para otra operación.
func TestBegin_DifferentOperation_Conflicts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	coord := idempotency.NewSQLCoordinator()

	key := uniqueKey(t, "op-conflict")
	fp := fingerprintOf(`{"resource":"c"}`)

	err := db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
		decision, err := coord.Begin(ctx, q, shopA, key, "operation_a", fp, 60*time.Second)
		if err != nil {
			return err
		}
		if decision.Outcome != idempotency.OutcomeProceed {
			return fmt.Errorf("expected OutcomeProceed, got %s", decision.Outcome)
		}
		_, err = coord.Complete(ctx, q, shopA, key, idempotency.StoredResponse{
			Status: 200, ContentType: "application/json", Body: `{}`,
		})
		return err
	})
	if err != nil {
		t.Fatalf("first InTenantTx: %v", err)
	}

	var conflict idempotency.Decision
	err = db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
		decision, err := coord.Begin(ctx, q, shopA, key, "operation_b", fp, 60*time.Second)
		if err != nil {
			return err
		}
		conflict = decision
		return nil
	})
	if err != nil {
		t.Fatalf("second InTenantTx: %v", err)
	}

	if conflict.Outcome != idempotency.OutcomeConflictOperation {
		t.Fatalf("expected OutcomeConflictOperation, got %s", conflict.Outcome)
	}
}

// TestBegin_RollbackUndoesReclaim_AllowsLegitimateRetry cubre CA-004-06 por
// la vía natural: un efecto que falla devuelve error desde el callback,
// InTenantTx hace ROLLBACK de toda la transacción (incluido el INSERT que
// Begin hizo), y un reintento legítimo encuentra la clave libre de nuevo
// sin necesidad de llamar Abort.
func TestBegin_RollbackUndoesReclaim_AllowsLegitimateRetry(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	coord := idempotency.NewSQLCoordinator()

	key := uniqueKey(t, "rollback-retry")
	fp := fingerprintOf(`{"resource":"d"}`)
	simulatedFailure := errors.New("fallo simulado del efecto")

	err := db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
		decision, err := coord.Begin(ctx, q, shopA, key, "test_op", fp, 60*time.Second)
		if err != nil {
			return err
		}
		if decision.Outcome != idempotency.OutcomeProceed {
			return fmt.Errorf("expected OutcomeProceed, got %s", decision.Outcome)
		}
		return simulatedFailure // fuerza ROLLBACK
	})
	if !errors.Is(err, simulatedFailure) {
		t.Fatalf("expected the simulated failure to propagate, got %v", err)
	}

	var retry idempotency.Decision
	err = db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
		decision, err := coord.Begin(ctx, q, shopA, key, "test_op", fp, 60*time.Second)
		if err != nil {
			return err
		}
		retry = decision
		return nil
	})
	if err != nil {
		t.Fatalf("retry InTenantTx: %v", err)
	}
	if retry.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected the retry to see a free key (OutcomeProceed), got %s", retry.Outcome)
	}
}

// TestAbort_ReleasesInProgressClaim_ButNeverCompleted cubre CA-004-06 (vía
// explícita) y el cierre de DDL-IDEM-01: Abort limpia una reclamación
// 'in_progress' propia, pero jamás borra una fila ya 'completed'.
func TestAbort_ReleasesInProgressClaim_ButNeverCompleted(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	coord := idempotency.NewSQLCoordinator()

	t.Run("libera in_progress y permite reintento en la misma transacción", func(t *testing.T) {
		key := uniqueKey(t, "abort-in-progress")
		fp := fingerprintOf(`{"resource":"e"}`)

		err := db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
			decision, err := coord.Begin(ctx, q, shopA, key, "test_op", fp, 60*time.Second)
			if err != nil {
				return err
			}
			if decision.Outcome != idempotency.OutcomeProceed {
				return fmt.Errorf("expected OutcomeProceed, got %s", decision.Outcome)
			}

			ok, err := coord.Abort(ctx, q, shopA, key)
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("expected Abort to report ok=true for an in_progress claim")
			}

			retry, err := coord.Begin(ctx, q, shopA, key, "test_op", fp, 60*time.Second)
			if err != nil {
				return err
			}
			if retry.Outcome != idempotency.OutcomeProceed {
				return fmt.Errorf("expected OutcomeProceed after Abort, got %s", retry.Outcome)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("InTenantTx: %v", err)
		}
	})

	t.Run("nunca borra una fila completed", func(t *testing.T) {
		key := uniqueKey(t, "abort-completed")
		fp := fingerprintOf(`{"resource":"f"}`)
		stored := idempotency.StoredResponse{Status: 200, ContentType: "application/json", Body: `{"ok":true}`}

		err := db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
			decision, err := coord.Begin(ctx, q, shopA, key, "test_op", fp, 60*time.Second)
			if err != nil {
				return err
			}
			if decision.Outcome != idempotency.OutcomeProceed {
				return fmt.Errorf("expected OutcomeProceed, got %s", decision.Outcome)
			}
			_, err = coord.Complete(ctx, q, shopA, key, stored)
			return err
		})
		if err != nil {
			t.Fatalf("complete InTenantTx: %v", err)
		}

		var abortOK bool
		var replay idempotency.Decision
		err = db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
			var err error
			abortOK, err = coord.Abort(ctx, q, shopA, key)
			if err != nil {
				return err
			}
			replay, err = coord.Begin(ctx, q, shopA, key, "test_op", fp, 60*time.Second)
			return err
		})
		if err != nil {
			t.Fatalf("abort+begin InTenantTx: %v", err)
		}

		if abortOK {
			t.Fatal("expected Abort to report ok=false for an already-completed row")
		}
		if replay.Outcome != idempotency.OutcomeReplay {
			t.Fatalf("expected the completed row to survive intact (OutcomeReplay), got %s", replay.Outcome)
		}
		if replay.Response != stored {
			t.Fatalf("expected the original response to survive Abort untouched, got %+v", replay.Response)
		}
	})
}

// TestBegin_CrossTenantIsolation cubre CA-004-04: la misma clave literal en
// dos barberías no interfiere.
func TestBegin_CrossTenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	coord := idempotency.NewSQLCoordinator()

	raw := uniqueKey(t, "cross-tenant")
	fpA := fingerprintOf(`{"shop":"a"}`)
	fpB := fingerprintOf(`{"shop":"b"}`)
	storedA := idempotency.StoredResponse{Status: 201, ContentType: "application/json", Body: `{"shop":"a"}`}

	err := db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
		decision, err := coord.Begin(ctx, q, shopA, raw, "test_op", fpA, 60*time.Second)
		if err != nil {
			return err
		}
		if decision.Outcome != idempotency.OutcomeProceed {
			return fmt.Errorf("expected OutcomeProceed for shop A, got %s", decision.Outcome)
		}
		_, err = coord.Complete(ctx, q, shopA, raw, storedA)
		return err
	})
	if err != nil {
		t.Fatalf("shop A InTenantTx: %v", err)
	}

	// La misma clave literal, en la OTRA barbería, con contenido distinto:
	// si el aislamiento fallara, esto vería la fila de A y devolvería
	// conflict_fingerprint o replay en vez de proceed.
	var decisionB idempotency.Decision
	err = db.InTenantTx(context.Background(), shopB, func(ctx context.Context, q database.Queries) error {
		decision, err := coord.Begin(ctx, q, shopB, raw, "test_op", fpB, 60*time.Second)
		if err != nil {
			return err
		}
		decisionB = decision
		return nil
	})
	if err != nil {
		t.Fatalf("shop B InTenantTx: %v", err)
	}
	if decisionB.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected shop B to see a free key despite the literal match with shop A, got %s", decisionB.Outcome)
	}
}

// TestBegin_ExpiredRecord_ReevaluatesInsteadOfReplaying cubre CA-004-05. El
// TTL usado es el mínimo permitido por idempotency_begin (1 segundo,
// idempotency_begin: p_ttl_seconds 1-86400); la espera es una constante
// acotada y conocida (algo mayor que ese TTL), no una duración adivinada.
func TestBegin_ExpiredRecord_ReevaluatesInsteadOfReplaying(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	coord := idempotency.NewSQLCoordinator()

	key := uniqueKey(t, "expired")
	original := fingerprintOf(`{"attempt":1}`)
	later := fingerprintOf(`{"attempt":2}`)

	err := db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
		decision, err := coord.Begin(ctx, q, shopA, key, "test_op", original, 1*time.Second)
		if err != nil {
			return err
		}
		if decision.Outcome != idempotency.OutcomeProceed {
			return fmt.Errorf("expected OutcomeProceed, got %s", decision.Outcome)
		}
		_, err = coord.Complete(ctx, q, shopA, key, idempotency.StoredResponse{
			Status: 200, ContentType: "application/json", Body: `{"attempt":1}`,
		})
		return err
	})
	if err != nil {
		t.Fatalf("first InTenantTx: %v", err)
	}

	time.Sleep(1200 * time.Millisecond)

	var afterExpiry idempotency.Decision
	err = db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
		decision, err := coord.Begin(ctx, q, shopA, key, "test_op", later, 60*time.Second)
		if err != nil {
			return err
		}
		afterExpiry = decision
		return nil
	})
	if err != nil {
		t.Fatalf("second InTenantTx: %v", err)
	}

	if afterExpiry.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected a vencida clave to be re-evaluated as OutcomeProceed, got %s", afterExpiry.Outcome)
	}
}

// TestBegin_InvalidFingerprint_RejectedBeforePostgres verifica que un
// Fingerprint construido fuera de ComputeFingerprint (formato inválido) se
// rechaza en Go, sin depender de que PostgreSQL lo detecte.
func TestBegin_InvalidFingerprint_RejectedBeforePostgres(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	coord := idempotency.NewSQLCoordinator()

	key := uniqueKey(t, "invalid-fingerprint")

	err := db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
		_, err := coord.Begin(ctx, q, shopA, key, "test_op", idempotency.Fingerprint("no-es-hexadecimal"), 60*time.Second)
		return err
	})
	if err == nil {
		t.Fatal("expected an error for an invalid fingerprint")
	}
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected KindInternal (defecto del llamador, no del cliente), got %v", err)
	}
}

// TestBegin_TwoRealConcurrentConnections_SingleEffectAndCoherentResponses
// cubre CA-004-03 con concurrencia REAL: dos conexiones distintas del pool
// (dos sesiones PostgreSQL reales) intentan la misma clave mientras la
// primera transacción sigue abierta. Ejecutar con -race.
func TestBegin_TwoRealConcurrentConnections_SingleEffectAndCoherentResponses(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	coord := idempotency.NewSQLCoordinator()

	key := uniqueKey(t, "two-conn")
	fp := fingerprintOf(`{"concurrent":true}`)
	stored := idempotency.StoredResponse{Status: 201, ContentType: "application/json", Body: `{"id":"only-once"}`}

	aReady := make(chan struct{})
	bDone := make(chan struct{})

	var (
		aOutcome, bOutcome idempotency.Outcome
		aErr, bErr         error
		wg                 sync.WaitGroup
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		aErr = db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
			decision, err := coord.Begin(ctx, q, shopA, key, "test_op", fp, 60*time.Second)
			if err != nil {
				return err
			}
			aOutcome = decision.Outcome
			close(aReady)
			<-bDone // mantiene la transacción de A abierta mientras B intenta

			if decision.Outcome == idempotency.OutcomeProceed {
				if _, err := coord.Complete(ctx, q, shopA, key, stored); err != nil {
					return err
				}
			}
			return nil
		})
	}()

	<-aReady
	bErr = db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
		decision, err := coord.Begin(ctx, q, shopA, key, "test_op", fp, 60*time.Second)
		if err != nil {
			return err
		}
		bOutcome = decision.Outcome
		return nil
	})
	close(bDone)
	wg.Wait()

	if aErr != nil {
		t.Fatalf("session A: %v", aErr)
	}
	if bErr != nil {
		t.Fatalf("session B: %v", bErr)
	}
	if aOutcome != idempotency.OutcomeProceed {
		t.Fatalf("expected session A to proceed, got %s", aOutcome)
	}
	if bOutcome != idempotency.OutcomeLocked {
		t.Fatalf("expected session B to be locked immediately (DEC-043), got %s", bOutcome)
	}
	if _, ok := apperr.As(idempotency.Decision{Outcome: bOutcome}.AsError()); !ok {
		t.Fatal("expected session B's outcome to translate into a client-facing error")
	}

	// Coherencia (CA-004-03): tras A confirmar, un reintento legítimo con la
	// misma clave y el mismo contenido reproduce exactamente esa única
	// ejecución — nunca una segunda.
	var afterward idempotency.Decision
	err := db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
		decision, err := coord.Begin(ctx, q, shopA, key, "test_op", fp, 60*time.Second)
		if err != nil {
			return err
		}
		afterward = decision
		return nil
	})
	if err != nil {
		t.Fatalf("afterward InTenantTx: %v", err)
	}
	if afterward.Outcome != idempotency.OutcomeReplay {
		t.Fatalf("expected OutcomeReplay after the winning session completed, got %s", afterward.Outcome)
	}
	if afterward.Response != stored {
		t.Fatalf("expected the single execution's exact response, got %+v", afterward.Response)
	}
}

// TestBegin_AbandonedClaim_ReportsConflictInProgress cubre el estado
// conflict_in_progress: el lock se libera al terminar la transacción de A
// (COMMIT), pero A nunca llamó Complete ni Abort, así que la fila sigue
// 'in_progress' y vigente cuando B llega después, ya sin contención de
// lock.
func TestBegin_AbandonedClaim_ReportsConflictInProgress(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	coord := idempotency.NewSQLCoordinator()

	key := uniqueKey(t, "abandoned")
	fp := fingerprintOf(`{"abandoned":true}`)

	err := db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
		decision, err := coord.Begin(ctx, q, shopA, key, "test_op", fp, 60*time.Second)
		if err != nil {
			return err
		}
		if decision.Outcome != idempotency.OutcomeProceed {
			return fmt.Errorf("expected OutcomeProceed, got %s", decision.Outcome)
		}
		return nil // COMMIT sin Complete ni Abort: reclamación abandonada.
	})
	if err != nil {
		t.Fatalf("first InTenantTx: %v", err)
	}

	var decision idempotency.Decision
	err = db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
		var err error
		decision, err = coord.Begin(ctx, q, shopA, key, "test_op", fp, 60*time.Second)
		return err
	})
	if err != nil {
		t.Fatalf("second InTenantTx: %v", err)
	}
	if decision.Outcome != idempotency.OutcomeConflictInProgress {
		t.Fatalf("expected OutcomeConflictInProgress, got %s", decision.Outcome)
	}
}
