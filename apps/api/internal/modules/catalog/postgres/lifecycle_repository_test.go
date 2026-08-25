// Pruebas de integración de HU-024 (Repository.Deactivate/Reactivate) contra
// PostgreSQL REAL, mismos requisitos que repository_test.go: migraciones
// aplicadas hasta 20260824150000_create_barber_service.sql y
// database/testdata/hu022_catalogo.sql cargado. Conéctate como
// barberia_app.
package postgres_test

import (
	"context"
	"sync"
	"testing"

	"system-barbershop/internal/modules/catalog"
	"system-barbershop/internal/platform/idempotency"
)

func deactivateFP(serviceID string) idempotency.Fingerprint {
	return idempotency.ComputeFingerprint("POST", "/api/v1/private/services/"+serviceID+"/deactivate", nil)
}

func reactivateFP(serviceID string) idempotency.Fingerprint {
	return idempotency.ComputeFingerprint("POST", "/api/v1/private/services/"+serviceID+"/reactivate", nil)
}

// --- CA-024-02/CA-024-03: desactivar conserva el recurso -------------------

func TestDeactivate_ActiveService_SetsInactiveWithTimestamp(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	svc := createService(t, repo, shopE, "Corte a desactivar "+suffix)
	key := idempotency.Key("deactivate-" + suffix)

	result, err := repo.Deactivate(context.Background(), string(shopE), svc.ID, key, deactivateFP(svc.ID))
	if err != nil {
		t.Fatalf("Deactivate: %v", err)
	}
	if result.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected OutcomeProceed, got %s", result.Decision.Outcome)
	}
	if result.InvalidTransition {
		t.Fatal("did not expect InvalidTransition on the first deactivation of an active service")
	}
	if result.Service.IsActive {
		t.Fatal("expected IsActive=false after Deactivate")
	}
	if result.Service.DeactivatedAt == nil {
		t.Fatal("expected DeactivatedAt to be set after Deactivate")
	}
	if result.Service.ID != svc.ID || result.Service.Name != svc.Name || result.Service.DurationMinutes != svc.DurationMinutes || result.Service.PriceCents != svc.PriceCents {
		t.Fatalf("CA-024-02: deactivate must not alter identity/name/duration/price: got %+v, want same as %+v", result.Service, svc)
	}

	got, found, err := repo.Get(context.Background(), string(shopE), svc.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found {
		t.Fatal("CA-024-02: the service must still exist (no physical delete) after deactivation")
	}
	if got.IsActive || got.DeactivatedAt == nil {
		t.Fatalf("expected persisted row to be inactive with a timestamp, got %+v", got)
	}
}

func TestDeactivate_Repeated_SameKey_ReturnsSameResponseWithoutSecondEffect(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	svc := createService(t, repo, shopE, "Corte repetido "+suffix)
	key := idempotency.Key("deactivate-repeat-" + suffix)
	fp := deactivateFP(svc.ID)

	first, err := repo.Deactivate(context.Background(), string(shopE), svc.ID, key, fp)
	if err != nil {
		t.Fatalf("Deactivate (first): %v", err)
	}
	if first.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected first call to proceed, got %s", first.Decision.Outcome)
	}

	second, err := repo.Deactivate(context.Background(), string(shopE), svc.ID, key, fp)
	if err != nil {
		t.Fatalf("Deactivate (second): %v", err)
	}
	if second.Decision.Outcome != idempotency.OutcomeReplay {
		t.Fatalf("CA-024-06: expected OutcomeReplay on repetition with the same key, got %s", second.Decision.Outcome)
	}
	if second.Response.Body != first.Response.Body {
		t.Fatalf("CA-004-01: expected byte-identical replay body, got %q vs %q", second.Response.Body, first.Response.Body)
	}
}

func TestDeactivate_AlreadyInactive_NewKey_ReturnsInvalidTransitionWithoutChangingRow(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	svc := createService(t, repo, shopE, "Corte ya inactivo "+suffix)
	first, err := repo.Deactivate(context.Background(), string(shopE), svc.ID, idempotency.Key("deactivate-a-"+suffix), deactivateFP(svc.ID))
	if err != nil || first.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("setup: expected the first deactivation to proceed, got %+v err=%v", first, err)
	}

	// CA-024-06: una clave NUEVA sobre una transición que ya no aplica
	// (el servicio ya está inactivo) es una transición inválida, no un
	// segundo efecto silencioso.
	second, err := repo.Deactivate(context.Background(), string(shopE), svc.ID, idempotency.Key("deactivate-b-"+suffix), deactivateFP(svc.ID))
	if err != nil {
		t.Fatalf("Deactivate (second, new key): %v", err)
	}
	if second.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected the new key to also reach OutcomeProceed (it is a fresh claim), got %s", second.Decision.Outcome)
	}
	if !second.Found {
		t.Fatal("expected Found=true: the service exists, it is just already inactive")
	}
	if !second.InvalidTransition {
		t.Fatal("expected InvalidTransition=true: the service was already inactive")
	}

	got, found, err := repo.Get(context.Background(), string(shopE), svc.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found || got.DeactivatedAt == nil || !first.Service.DeactivatedAt.Equal(*got.DeactivatedAt) {
		t.Fatalf("expected the row to keep its original deactivatedAt (rejected transition changed nothing), got %+v", got)
	}
}

func TestDeactivate_UnknownID_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	const unknownID = "00000000-0000-0000-0000-000000000000"
	result, err := repo.Deactivate(context.Background(), string(shopE), unknownID, idempotency.Key("deactivate-missing-"+suffix), deactivateFP(unknownID))
	if err != nil {
		t.Fatalf("Deactivate: %v", err)
	}
	if result.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected OutcomeProceed (fresh claim, rolled back), got %s", result.Decision.Outcome)
	}
	if result.Found {
		t.Fatal("CA-024-07: expected Found=false for an unknown serviceId")
	}
}

func TestDeactivate_CrossTenant_NeverLeaksAnotherShopsService(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	svc := createService(t, repo, shopF, "De la barbería F "+suffix)

	result, err := repo.Deactivate(context.Background(), string(shopE), svc.ID, idempotency.Key("deactivate-cross-"+suffix), deactivateFP(svc.ID))
	if err != nil {
		t.Fatalf("Deactivate: %v", err)
	}
	if result.Found {
		t.Fatal("CA-024-07/RN-TEN-01: shopE must not be able to deactivate shopF's service")
	}

	got, found, err := repo.Get(context.Background(), string(shopF), svc.ID)
	if err != nil {
		t.Fatalf("Get (owner tenant): %v", err)
	}
	if !found || !got.IsActive {
		t.Fatal("expected shopF's service to remain untouched and active")
	}
}

// --- CA-024-05: reactivar no recrea ni altera catálogo ----------------------

func TestReactivate_InactiveService_SetsActiveClearsTimestamp(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	svc := createService(t, repo, shopE, "Corte a reactivar "+suffix)
	deactivated, err := repo.Deactivate(context.Background(), string(shopE), svc.ID, idempotency.Key("deactivate-for-reactivate-"+suffix), deactivateFP(svc.ID))
	if err != nil || deactivated.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("setup: expected deactivation to proceed, got %+v err=%v", deactivated, err)
	}

	result, err := repo.Reactivate(context.Background(), string(shopE), svc.ID, idempotency.Key("reactivate-"+suffix), reactivateFP(svc.ID))
	if err != nil {
		t.Fatalf("Reactivate: %v", err)
	}
	if result.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected OutcomeProceed, got %s", result.Decision.Outcome)
	}
	if result.InvalidTransition {
		t.Fatal("did not expect InvalidTransition when reactivating an inactive service")
	}
	if !result.Service.IsActive {
		t.Fatal("expected IsActive=true after Reactivate")
	}
	if result.Service.DeactivatedAt != nil {
		t.Fatalf("expected DeactivatedAt=nil after Reactivate, got %v", *result.Service.DeactivatedAt)
	}
	// CA-024-05: nunca crea otra fila ni altera duración/precio.
	if result.Service.ID != svc.ID || result.Service.DurationMinutes != svc.DurationMinutes || result.Service.PriceCents != svc.PriceCents {
		t.Fatalf("expected the SAME row with unchanged duration/price, got %+v want id=%s duration=%d price=%d",
			result.Service, svc.ID, svc.DurationMinutes, svc.PriceCents)
	}
}

func TestReactivate_AlreadyActive_NewKey_ReturnsInvalidTransition(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	svc := createService(t, repo, shopE, "Corte siempre activo "+suffix)

	result, err := repo.Reactivate(context.Background(), string(shopE), svc.ID, idempotency.Key("reactivate-noop-"+suffix), reactivateFP(svc.ID))
	if err != nil {
		t.Fatalf("Reactivate: %v", err)
	}
	if !result.Found {
		t.Fatal("expected Found=true: the service exists")
	}
	if !result.InvalidTransition {
		t.Fatal("expected InvalidTransition=true: the service was already active")
	}
}

func TestReactivate_UnknownID_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	const unknownID = "00000000-0000-0000-0000-000000000000"
	result, err := repo.Reactivate(context.Background(), string(shopE), unknownID, idempotency.Key("reactivate-missing-"+suffix), reactivateFP(unknownID))
	if err != nil {
		t.Fatalf("Reactivate: %v", err)
	}
	if result.Found {
		t.Fatal("CA-024-07: expected Found=false for an unknown serviceId")
	}
}

// --- CA-024-06/DEC-069: carrera real, sin sleep como sincronización --------

// TestDeactivate_TwoRealConcurrentConnections_ExactlyOneSucceeds confirma el
// mecanismo TOCTOU exigido por el trabajo requerido §2.4: dos conexiones
// reales confirmando la desactivación del MISMO servicio al mismo tiempo
// (dos claves de idempotencia DISTINTAS, así que ninguna puede resolverse
// por simple repetición) nunca dejan la fila en un estado inconsistente:
// exactamente una transiciona is_active; la otra descubre, tras esperar el
// FOR UPDATE de la primera, que la transición ya no aplica
// (InvalidTransition), nunca un error ni una segunda escritura.
func TestDeactivate_TwoRealConcurrentConnections_ExactlyOneSucceeds(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	svc := createService(t, repo, shopE, "Corte en carrera "+suffix)

	var (
		wg               sync.WaitGroup
		aResult, bResult catalog.LifecycleResult
		aErr, bErr       error
		aStarted         = make(chan struct{})
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		close(aStarted)
		aResult, aErr = repo.Deactivate(context.Background(), string(shopE), svc.ID, idempotency.Key("race-a-"+suffix), deactivateFP(svc.ID))
	}()
	go func() {
		defer wg.Done()
		<-aStarted
		bResult, bErr = repo.Deactivate(context.Background(), string(shopE), svc.ID, idempotency.Key("race-b-"+suffix), deactivateFP(svc.ID))
	}()
	wg.Wait()

	if aErr != nil {
		t.Fatalf("goroutine A: %v", aErr)
	}
	if bErr != nil {
		t.Fatalf("goroutine B: %v", bErr)
	}

	succeeded := aResult.Decision.Outcome == idempotency.OutcomeProceed && !aResult.InvalidTransition
	rejected := bResult.Decision.Outcome == idempotency.OutcomeProceed && bResult.InvalidTransition
	if !succeeded || !rejected {
		// El orden real de ejecución no está garantizado: si B ganó la
		// carrera, los roles se invierten.
		succeeded = bResult.Decision.Outcome == idempotency.OutcomeProceed && !bResult.InvalidTransition
		rejected = aResult.Decision.Outcome == idempotency.OutcomeProceed && aResult.InvalidTransition
	}
	if !succeeded || !rejected {
		t.Fatalf("expected exactly one winner (deactivated) and one InvalidTransition, got a=%+v b=%+v", aResult, bResult)
	}

	got, found, err := repo.Get(context.Background(), string(shopE), svc.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found || got.IsActive || got.DeactivatedAt == nil {
		t.Fatalf("expected the service to end up inactive exactly once, got %+v", got)
	}
}
