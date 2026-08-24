// Package postgres_test (pruebas de integración, HU-022) requiere
// PostgreSQL REAL con las once migraciones aplicadas (incluida
// 20260824140000_create_service.sql) y database/testdata/dos_barberias.sql +
// database/testdata/hu022_catalogo.sql cargados. Conéctate como
// barberia_app (docs/03-desarrollo/estrategia-pruebas.md §2 prohíbe mocks
// para RLS/idempotencia/atomicidad).
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/catalog/postgres/...
package postgres_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"system-barbershop/internal/modules/catalog"
	catalogpostgres "system-barbershop/internal/modules/catalog/postgres"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

const (
	testDatabaseURL = "postgres://barberia_app@localhost:5432/barberia_test?sslmode=disable"

	// shopE/shopF son las dos barberías dedicadas de
	// testdata/hu022_catalogo.sql: aisladas de dos_barberias.sql y
	// hu021_barberos.sql para que un conteo exacto de servicios en una
	// página no dependa de lo que otras suites hagan sobre shopA-shopD.
	shopE = database.BarbershopID("55555555-5555-5555-5555-555555555555")
	shopF = database.BarbershopID("66666666-6666-6666-6666-666666666666")
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
	t.Cleanup(func() { db.Close() })
	return db
}

func newRepository(db *database.DB) *catalogpostgres.Repository {
	return catalogpostgres.New(db, idempotency.NewSQLCoordinator())
}

// uniqueSuffix evita colisiones entre ejecuciones repetidas de la suite
// contra el mismo PostgreSQL persistente (barberia_app no puede DELETE,
// RN-SER-03, así que las filas de una corrida anterior siguen ahí): cada
// prueba usa nombres/claves con un sufijo aleatorio propio.
func uniqueSuffix(t *testing.T) string {
	t.Helper()
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	return hex.EncodeToString(buf)
}

func validInput(name string) catalog.CreateInput {
	return catalog.CreateInput{Name: name, DurationMinutes: 30, PriceCents: 4500000}
}

func createService(t *testing.T, repo *catalogpostgres.Repository, shop database.BarbershopID, name string) catalog.Service {
	t.Helper()
	key := idempotency.Key("fixture-" + uniqueSuffix(t))
	fp := idempotency.ComputeFingerprint("POST", "/api/v1/private/services", []byte(`{"name":"`+name+`"}`))
	result, err := repo.Create(context.Background(), string(shop), validInput(name), key, fp)
	if err != nil {
		t.Fatalf("createService: %v", err)
	}
	if result.NameTaken {
		t.Fatalf("createService: unexpected name conflict for %q", name)
	}
	if result.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("createService: expected OutcomeProceed, got %s", result.Decision.Outcome)
	}
	return result.Service
}

// --- List: orden estable, paginación contigua sin duplicados/omisiones --

func TestList_TraversingUntilExhausted_TerminatesWithoutNextCursor(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	marker := createService(t, repo, shopF, "Marcador de recorrido "+suffix)

	seen := map[string]bool{}
	var cursor *catalog.Cursor
	pages := 0
	for {
		pages++
		if pages > 100000 {
			t.Fatal("too many pages; possible infinite loop (NextCursor never became empty)")
		}
		page, err := repo.List(context.Background(), string(shopF), cursor, 1)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		for _, item := range page.Items {
			if seen[item.ID] {
				t.Fatalf("item %s seen twice while paging", item.ID)
			}
			seen[item.ID] = true
		}
		if page.NextCursor == "" {
			break
		}
		decoded, err := catalog.DecodeCursor(page.NextCursor)
		if err != nil {
			t.Fatalf("DecodeCursor: %v", err)
		}
		cursor = &decoded
	}

	if !seen[marker.ID] {
		t.Fatalf("expected the just-created marker service %s to appear while exhausting the collection", marker.ID)
	}
}

func TestList_FourServicesInACatalog_ReturnedAsFourDistinctResources(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	created := make(map[string]bool, 4)
	for i := 0; i < 4; i++ {
		s := createService(t, repo, shopE, "Catálogo "+suffix+" "+string(rune('A'+i)))
		if created[s.ID] {
			t.Fatalf("duplicate id returned by Create: %s", s.ID)
		}
		created[s.ID] = true
	}
	if len(created) != 4 {
		t.Fatalf("expected 4 distinct services, got %d", len(created))
	}

	seen := map[string]bool{}
	var cursor *catalog.Cursor
	var lastCreatedAt time.Time
	pages := 0
	for {
		pages++
		if pages > 10000 {
			t.Fatal("too many pages; possible infinite loop")
		}
		page, err := repo.List(context.Background(), string(shopE), cursor, 1)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		for _, item := range page.Items {
			if seen[item.ID] {
				t.Fatalf("CA-022-01: item %s seen twice while paging", item.ID)
			}
			seen[item.ID] = true
			if item.CreatedAt.Before(lastCreatedAt) {
				t.Fatalf("expected non-decreasing created_at order, got %v after %v", item.CreatedAt, lastCreatedAt)
			}
			lastCreatedAt = item.CreatedAt
		}
		if page.NextCursor == "" {
			break
		}
		decoded, err := catalog.DecodeCursor(page.NextCursor)
		if err != nil {
			t.Fatalf("DecodeCursor: %v", err)
		}
		cursor = &decoded
	}

	for id := range created {
		if !seen[id] {
			t.Fatalf("CA-022-01: service %s was never returned while paging through all pages", id)
		}
	}
}

func TestList_FirstPage_NeverReturnsMoreThanLimit(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	createService(t, repo, shopE, "Límite Uno "+suffix)
	createService(t, repo, shopE, "Límite Dos "+suffix)
	createService(t, repo, shopE, "Límite Tres "+suffix)

	page, err := repo.List(context.Background(), string(shopE), nil, 2)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(page.Items) != 2 {
		t.Fatalf("expected exactly 2 items for limit=2, got %d", len(page.Items))
	}
	if page.NextCursor == "" {
		t.Fatal("expected a next cursor: there are at least 3 known items and limit was 2")
	}
}

// --- CA-022-06: A no ve, no lee ni edita un servicio de B -----------------

func TestGet_CrossTenant_NeverLeaksAnotherShopsService(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	serviceF := createService(t, repo, shopF, "De la barbería F "+suffix)

	_, found, err := repo.Get(context.Background(), string(shopE), serviceF.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if found {
		t.Fatal("CA-022-06: shopE's context could read shopF's service")
	}

	got, found, err := repo.Get(context.Background(), string(shopF), serviceF.ID)
	if err != nil {
		t.Fatalf("Get (own tenant): %v", err)
	}
	if !found || got.ID != serviceF.ID {
		t.Fatal("expected shopF to still see its own service")
	}
}

func TestGet_NonexistentID_ReturnsNotFoundSameAsOtherTenant(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	_, found, err := repo.Get(context.Background(), string(shopE), "99999999-9999-4999-8999-999999999999")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if found {
		t.Fatal("expected not found for a nonexistent id")
	}
}

func TestUpdate_CrossTenant_NeverEditsAnotherShopsService(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	serviceF := createService(t, repo, shopF, "Nombre Original F "+suffix)
	newName := "Renombrado por E " + suffix

	result, err := repo.Update(context.Background(), string(shopE), serviceF.ID, catalog.UpdateFields{Name: &newName})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if result.Found {
		t.Fatal("CA-022-06: shopE's context edited shopF's service")
	}

	stillF, found, err := repo.Get(context.Background(), string(shopF), serviceF.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found || stillF.Name != "Nombre Original F "+suffix {
		t.Fatalf("expected shopF's service name untouched, got %+v", stillF)
	}
}

// --- Update: edición parcial, sin duplicar el recurso ---------------------

func TestUpdate_OwnTenant_UpdatesByIDWithoutDuplicating(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	original := createService(t, repo, shopE, "Antes de editar "+suffix)
	newName := "Después de editar " + suffix

	result, err := repo.Update(context.Background(), string(shopE), original.ID, catalog.UpdateFields{Name: &newName})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !result.Found {
		t.Fatal("expected Found=true for the service's own tenant")
	}
	if result.Service.ID != original.ID {
		t.Fatalf("expected the same id after update, got %s vs %s", result.Service.ID, original.ID)
	}
	if result.Service.Name != newName {
		t.Fatalf("unexpected edited name: %+v", result.Service)
	}
	// Campos no editados permanecen intactos.
	if result.Service.DurationMinutes != original.DurationMinutes || result.Service.PriceCents != original.PriceCents {
		t.Fatalf("expected untouched fields to remain, got %+v (original %+v)", result.Service, original)
	}

	got, found, err := repo.Get(context.Background(), string(shopE), original.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found || got.Name != newName {
		t.Fatalf("expected the edit to persist, got %+v", got)
	}
}

func TestUpdate_OnlyDuration_LeavesNameDescriptionPriceUntouched(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	original := createService(t, repo, shopE, "Solo duración "+suffix)
	newDuration := 90

	result, err := repo.Update(context.Background(), string(shopE), original.ID, catalog.UpdateFields{DurationMinutes: &newDuration})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !result.Found {
		t.Fatal("expected Found=true")
	}
	if result.Service.DurationMinutes != newDuration {
		t.Fatalf("expected duration %d, got %d", newDuration, result.Service.DurationMinutes)
	}
	if result.Service.Name != original.Name || result.Service.PriceCents != original.PriceCents {
		t.Fatalf("expected name/price untouched, got %+v (original %+v)", result.Service, original)
	}
}

func TestUpdate_DescriptionCleared_PersistsAsNull(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	description := "Descripción inicial"
	key := idempotency.Key("fixture-desc-" + suffix)
	fp := idempotency.ComputeFingerprint("POST", "/api/v1/private/services", []byte(`{}`))
	created, err := repo.Create(context.Background(), string(shopE), catalog.CreateInput{
		Name: "Con descripción " + suffix, Description: &description, DurationMinutes: 30, PriceCents: 100,
	}, key, fp)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Service.Description == nil || *created.Service.Description != description {
		t.Fatalf("expected the fixture to have a description, got %+v", created.Service)
	}

	result, err := repo.Update(context.Background(), string(shopE), created.Service.ID,
		catalog.UpdateFields{Description: catalog.OptionalDescription{Set: true, Value: nil}})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !result.Found {
		t.Fatal("expected Found=true")
	}
	if result.Service.Description != nil {
		t.Fatalf("expected description cleared to nil, got %q", *result.Service.Description)
	}
}

func TestUpdate_NotFound_ReturnsFoundFalse(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	newName := "No importa"
	result, err := repo.Update(context.Background(), string(shopE), "99999999-9999-4999-8999-999999999999",
		catalog.UpdateFields{Name: &newName})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if result.Found {
		t.Fatal("expected Found=false for a nonexistent id")
	}
}

// --- DEC-067: nombre único entre servicios ACTIVOS de la misma barbería --

func TestCreate_DuplicateActiveName_ReturnsNameTakenWithoutPersisting(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	name := "Nombre Único " + suffix

	first := createService(t, repo, shopE, name)

	key := idempotency.Key("dup-name-" + suffix)
	fp := idempotency.ComputeFingerprint("POST", "/api/v1/private/services", []byte(`{"name":"`+name+`"}`))
	second, err := repo.Create(context.Background(), string(shopE), validInput(name), key, fp)
	if err != nil {
		t.Fatalf("Create (duplicate name): %v", err)
	}
	if !second.NameTaken {
		t.Fatal("DEC-067: expected NameTaken=true for a duplicate active name")
	}
	if second.Service.ID != "" {
		t.Fatal("expected no service to be created on a name conflict")
	}

	// El primero sigue existiendo intacto; no se creó una segunda fila.
	got, found, err := repo.Get(context.Background(), string(shopE), first.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found || got.Name != name {
		t.Fatalf("expected the original service untouched, got found=%v %+v", found, got)
	}

	// La clave de idempotencia usada en el intento fallido queda libre: un
	// reintento legítimo con un nombre distinto SÍ debe proceder (CA-004-06
	// aplicado al conflicto de nombre: el ROLLBACK deshace también el Begin).
	fp2 := idempotency.ComputeFingerprint("POST", "/api/v1/private/services", []byte(`{"name":"otro"}`))
	retry, err := repo.Create(context.Background(), string(shopE), validInput("Nombre Distinto "+suffix), key, fp2)
	if err != nil {
		t.Fatalf("Create (retry with same key, different name): %v", err)
	}
	if retry.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected the idempotency key to be free for a legitimate retry, got %s", retry.Decision.Outcome)
	}
}

func TestCreate_SameNameDifferentTenants_BothSucceed(t *testing.T) {
	// DEC-067 restringe la unicidad a la MISMA barbería: dos barberías
	// distintas pueden usar el mismo nombre exacto sin conflicto.
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	name := "Compartido Entre Tenants " + suffix

	e := createService(t, repo, shopE, name)
	f := createService(t, repo, shopF, name)

	if e.ID == f.ID {
		t.Fatal("expected two distinct services, one per tenant")
	}
	if e.Name != name || f.Name != name {
		t.Fatalf("expected both to keep the requested name, got %q and %q", e.Name, f.Name)
	}
}

func TestUpdate_RenameToAnotherActiveServicesName_ReturnsNameTakenWithoutPersisting(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	taken := createService(t, repo, shopE, "Ya Existe "+suffix)
	other := createService(t, repo, shopE, "Para Renombrar "+suffix)

	conflictingName := taken.Name
	result, err := repo.Update(context.Background(), string(shopE), other.ID, catalog.UpdateFields{Name: &conflictingName})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !result.NameTaken {
		t.Fatal("DEC-067: expected NameTaken=true when renaming into another active service's name")
	}

	// `other` conserva su nombre original: el UPDATE no aplicó nada parcial.
	got, found, err := repo.Get(context.Background(), string(shopE), other.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found || got.Name != other.Name {
		t.Fatalf("expected other's name untouched after a conflicting rename attempt, got %+v", got)
	}
}

// --- Idempotencia del alta (RN-IDE-01, DEC-043, CA-022-02) ----------------

func TestCreate_FirstExecution_PersistsAndReturnsProceed(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	name := "Nuevo " + suffix
	key := idempotency.Key("first-exec-" + suffix)
	fp := idempotency.ComputeFingerprint("POST", "/api/v1/private/services", []byte(`{"name":"Nuevo"}`))

	result, err := repo.Create(context.Background(), string(shopE), validInput(name), key, fp)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if result.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected OutcomeProceed, got %s", result.Decision.Outcome)
	}
	if result.Service.ID == "" || result.Service.Name != name {
		t.Fatalf("unexpected created service: %+v", result.Service)
	}
	if result.Service.Currency != catalog.CurrencyCOP {
		t.Fatalf("expected currency %q, got %q", catalog.CurrencyCOP, result.Service.Currency)
	}
	if result.Response.Status != 201 {
		t.Fatalf("expected stored status 201, got %d", result.Response.Status)
	}

	got, found, err := repo.Get(context.Background(), string(shopE), result.Service.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found || got.Name != name {
		t.Fatalf("expected the created service to be persisted, got found=%v %+v", found, got)
	}
}

func TestCreate_SameKeyAndContent_ReplaysWithoutCreatingASecondRow(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	name := "Repetido " + suffix
	key := idempotency.Key("replay-" + suffix)
	body := []byte(`{"name":"Repetido"}`)
	fp := idempotency.ComputeFingerprint("POST", "/api/v1/private/services", body)

	first, err := repo.Create(context.Background(), string(shopE), validInput(name), key, fp)
	if err != nil {
		t.Fatalf("Create (first): %v", err)
	}
	if first.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected first call to proceed, got %s", first.Decision.Outcome)
	}

	second, err := repo.Create(context.Background(), string(shopE), validInput(name), key, fp)
	if err != nil {
		t.Fatalf("Create (second): %v", err)
	}
	if second.Decision.Outcome != idempotency.OutcomeReplay {
		t.Fatalf("expected second call to replay, got %s", second.Decision.Outcome)
	}
	if second.Response != first.Response {
		t.Fatalf("expected the exact same stored response (CA-004-01), got %+v vs %+v", first.Response, second.Response)
	}

	var parsed struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(second.Response.Body), &parsed); err != nil {
		t.Fatalf("unmarshal stored response body: %v", err)
	}
	if parsed.ID != first.Service.ID {
		t.Fatalf("expected the replayed body to reference the original service id %s, got %s", first.Service.ID, parsed.ID)
	}
}

func TestCreate_SameKeyDifferentContent_ConflictsWithoutEffect(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	key := idempotency.Key("conflict-" + suffix)
	fpA := idempotency.ComputeFingerprint("POST", "/api/v1/private/services", []byte(`{"name":"A"}`))
	fpB := idempotency.ComputeFingerprint("POST", "/api/v1/private/services", []byte(`{"name":"B"}`))

	first, err := repo.Create(context.Background(), string(shopE), validInput("Nombre A "+suffix), key, fpA)
	if err != nil {
		t.Fatalf("Create (first): %v", err)
	}
	if first.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected first call to proceed, got %s", first.Decision.Outcome)
	}

	second, err := repo.Create(context.Background(), string(shopE), validInput("Nombre B "+suffix), key, fpB)
	if err != nil {
		t.Fatalf("Create (second): %v", err)
	}
	if second.Decision.Outcome != idempotency.OutcomeConflictFingerprint {
		t.Fatalf("expected OutcomeConflictFingerprint, got %s", second.Decision.Outcome)
	}
	if second.Service.ID != "" {
		t.Fatal("expected no service created on conflict")
	}
	if err := second.Decision.AsError(); err == nil {
		t.Fatal("expected Decision.AsError() to be non-nil for a conflict")
	} else if appErr, ok := apperr.As(err); !ok || appErr.Kind != apperr.KindIdempotencyConflict {
		t.Fatalf("expected apperr.KindIdempotencyConflict, got %v", err)
	}
}

func TestCreate_TwoRealConcurrentConnections_SingleEffect(t *testing.T) {
	// CA-004-03/RN-IDE-01: dos conexiones reales y simultáneas con la misma
	// clave producen el efecto exactamente una vez; la perdedora recibe
	// OutcomeLocked de inmediato, sin esperar.
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	key := idempotency.Key("two-conn-" + suffix)
	name := "Concurrente " + suffix
	body := []byte(`{"name":"` + name + `"}`)
	fp := idempotency.ComputeFingerprint("POST", "/api/v1/private/services", body)

	var (
		wg               sync.WaitGroup
		aResult, bResult catalog.CreateResult
		aErr, bErr       error
		aStarted         = make(chan struct{})
		bAttempted       = make(chan struct{})
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		close(aStarted)
		aResult, aErr = repo.Create(context.Background(), string(shopE), validInput(name), key, fp)
	}()
	go func() {
		defer wg.Done()
		<-aStarted
		bResult, bErr = repo.Create(context.Background(), string(shopE), validInput(name), key, fp)
		close(bAttempted)
	}()
	wg.Wait()
	<-bAttempted

	if aErr != nil {
		t.Fatalf("goroutine A: %v", aErr)
	}
	if bErr != nil {
		t.Fatalf("goroutine B: %v", bErr)
	}

	if aResult.Decision.Outcome == idempotency.OutcomeProceed && bResult.Decision.Outcome == idempotency.OutcomeProceed {
		t.Fatalf("CA-004-03: both connections proceeded, expected the effect exactly once (a=%s b=%s)",
			aResult.Decision.Outcome, bResult.Decision.Outcome)
	}
	validPair := func(x, y idempotency.Outcome) bool {
		return x == idempotency.OutcomeProceed &&
			(y == idempotency.OutcomeReplay || y == idempotency.OutcomeLocked || y == idempotency.OutcomeConflictInProgress)
	}
	if !validPair(aResult.Decision.Outcome, bResult.Decision.Outcome) && !validPair(bResult.Decision.Outcome, aResult.Decision.Outcome) {
		t.Fatalf("unexpected outcome pair: a=%s b=%s", aResult.Decision.Outcome, bResult.Decision.Outcome)
	}

	winner := aResult
	if winner.Decision.Outcome != idempotency.OutcomeProceed {
		winner = bResult
	}
	if winner.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatal("expected at least one of the two connections to proceed")
	}
	got, found, err := repo.Get(context.Background(), string(shopE), winner.Service.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found || got.Name != name {
		t.Fatalf("expected the single created service to be persisted, got found=%v %+v", found, got)
	}
}

func TestCreate_SameKeyLiteralInTwoTenants_DoesNotInterfere(t *testing.T) {
	// CA-004-04/RN-TEN-01: la misma clave literal en dos barberías distintas
	// no interfiere: cada una procede de forma independiente.
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	key := idempotency.Key("cross-tenant-" + suffix)
	fp := idempotency.ComputeFingerprint("POST", "/api/v1/private/services", []byte(`{"name":"X"}`))

	resultE, err := repo.Create(context.Background(), string(shopE), validInput("En E "+suffix), key, fp)
	if err != nil {
		t.Fatalf("Create shopE: %v", err)
	}
	if resultE.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected shopE to proceed, got %s", resultE.Decision.Outcome)
	}

	resultF, err := repo.Create(context.Background(), string(shopF), validInput("En F "+suffix), key, fp)
	if err != nil {
		t.Fatalf("Create shopF: %v", err)
	}
	if resultF.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected shopF to ALSO proceed with the same literal key (RN-TEN-01), got %s", resultF.Decision.Outcome)
	}
	if resultE.Service.ID == resultF.Service.ID {
		t.Fatal("expected two distinct services, one per tenant")
	}
}

// TestCreate_StoredResponseBody_MatchesHTTPAPIWireShape confirma que el
// JSON que Repository.Create persiste para una repetición exacta usa
// EXACTAMENTE las mismas claves que httpapi.ServiceResponse: la advertencia
// de mantenimiento manual que ambos archivos declaran en su comentario.
func TestCreate_StoredResponseBody_MatchesHTTPAPIWireShape(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	key := idempotency.Key("wire-shape-" + suffix)
	fp := idempotency.ComputeFingerprint("POST", "/api/v1/private/services", []byte(`{"name":"Forma"}`))

	result, err := repo.Create(context.Background(), string(shopE), validInput("Forma "+suffix), key, fp)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	var generic map[string]any
	if err := json.Unmarshal([]byte(result.Response.Body), &generic); err != nil {
		t.Fatalf("unmarshal stored body: %v", err)
	}
	wantKeys := []string{"id", "name", "description", "durationMinutes", "price", "currency", "createdAt", "updatedAt"}
	if len(generic) != len(wantKeys) {
		t.Fatalf("expected exactly %v, got keys %v", wantKeys, generic)
	}
	for _, k := range wantKeys {
		if _, ok := generic[k]; !ok {
			t.Fatalf("expected key %q in the stored response body, got %v", k, generic)
		}
	}
	if generic["price"] != "45000.00" {
		t.Fatalf("expected price as decimal string \"45000.00\", got %v (%T)", generic["price"], generic["price"])
	}
	if generic["currency"] != "COP" {
		t.Fatalf("expected currency COP, got %v", generic["currency"])
	}
}

// --- Precisión monetaria exacta (numeric, nunca coma flotante) -----------

func TestCreateGet_PriceRoundTripsExactly(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	cases := []int64{1, 100, 4500000, 999999999999, 4500099}
	for _, cents := range cases {
		key := idempotency.Key("price-" + suffix + "-" + catalog.FormatPriceCOP(cents))
		fp := idempotency.ComputeFingerprint("POST", "/api/v1/private/services", []byte(catalog.FormatPriceCOP(cents)))
		input := catalog.CreateInput{Name: "Precio " + suffix + " " + catalog.FormatPriceCOP(cents), DurationMinutes: 30, PriceCents: cents}

		result, err := repo.Create(context.Background(), string(shopE), input, key, fp)
		if err != nil {
			t.Fatalf("Create(%d): %v", cents, err)
		}
		if result.NameTaken {
			t.Fatalf("Create(%d): unexpected name conflict", cents)
		}
		if result.Service.PriceCents != cents {
			t.Fatalf("expected exact round-trip of %d cents, got %d", cents, result.Service.PriceCents)
		}

		got, found, err := repo.Get(context.Background(), string(shopE), result.Service.ID)
		if err != nil {
			t.Fatalf("Get(%d): %v", cents, err)
		}
		if !found || got.PriceCents != cents {
			t.Fatalf("expected Get to round-trip %d cents exactly, got found=%v %+v", cents, found, got)
		}
	}
}
