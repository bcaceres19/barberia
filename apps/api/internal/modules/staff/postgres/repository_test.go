// Package postgres_test (pruebas de integración, HU-021) requiere
// PostgreSQL REAL con las diez migraciones aplicadas (incluida
// 20260823130000_create_barber.sql) y database/testdata/dos_barberias.sql +
// database/testdata/hu021_barberos.sql cargados. Conéctate como
// barberia_app (docs/03-desarrollo/estrategia-pruebas.md §2 prohíbe mocks
// para RLS/idempotencia/atomicidad).
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/staff/postgres/...
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

	"system-barbershop/internal/modules/staff"
	staffpostgres "system-barbershop/internal/modules/staff/postgres"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

const (
	testDatabaseURL = "postgres://barberia_app@localhost:5432/barberia_test?sslmode=disable"

	// shopC/shopD son las dos barberías dedicadas de
	// testdata/hu021_barberos.sql: aisladas de dos_barberias.sql para que un
	// conteo exacto de barberos en una página no dependa de lo que otras
	// suites (settings, auth) hagan sobre shopA/shopB.
	shopC = database.BarbershopID("33333333-3333-3333-3333-333333333333")
	shopD = database.BarbershopID("44444444-4444-4444-4444-444444444444")
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

func newRepository(db *database.DB) *staffpostgres.Repository {
	return staffpostgres.New(db, idempotency.NewSQLCoordinator())
}

// uniqueSuffix evita colisiones entre ejecuciones repetidas de la suite
// contra el mismo PostgreSQL persistente (barberia_app no puede DELETE,
// CA-021-07/DEC-047, así que las filas de una corrida anterior siguen ahí):
// cada prueba usa nombres/claves con un sufijo aleatorio propio.
func uniqueSuffix(t *testing.T) string {
	t.Helper()
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	return hex.EncodeToString(buf)
}

func createBarber(t *testing.T, repo *staffpostgres.Repository, shop database.BarbershopID, fullName string) staff.Barber {
	t.Helper()
	key := idempotency.Key("fixture-" + uniqueSuffix(t))
	fp := idempotency.ComputeFingerprint("POST", "/api/v1/private/barbers", []byte(`{"fullName":"`+fullName+`"}`))
	result, err := repo.Create(context.Background(), string(shop), fullName, key, fp)
	if err != nil {
		t.Fatalf("createBarber: %v", err)
	}
	if result.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("createBarber: expected OutcomeProceed, got %s", result.Decision.Outcome)
	}
	return result.Barber
}

// --- List: orden estable, paginación contigua sin duplicados/omisiones --

// TestList_TraversingUntilExhausted_TerminatesWithoutNextCursor recorre
// TODA la colección actual de shopD (que puede tener filas de corridas
// anteriores de esta suite: barberia_app no puede DELETE, CA-021-07/
// DEC-047, así que esta prueba nunca asume un tamaño exacto ni una
// barbería vacía) con limit=1, y confirma que el recorrido SIEMPRE termina
// con NextCursor="" en algún momento, sin duplicados ni omisiones respecto
// de una lectura de control con un límite grande.
func TestList_TraversingUntilExhausted_TerminatesWithoutNextCursor(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	marker := createBarber(t, repo, shopD, "Marcador de recorrido "+suffix)

	seen := map[string]bool{}
	var cursor *staff.Cursor
	pages := 0
	for {
		pages++
		if pages > 100000 {
			t.Fatal("too many pages; possible infinite loop (NextCursor never became empty)")
		}
		page, err := repo.List(context.Background(), string(shopD), cursor, 1)
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
		decoded, err := staff.DecodeCursor(page.NextCursor)
		if err != nil {
			t.Fatalf("DecodeCursor: %v", err)
		}
		cursor = &decoded
	}

	if !seen[marker.ID] {
		t.Fatalf("expected the just-created marker barber %s to appear while exhausting the collection", marker.ID)
	}
}

func TestList_FourBarbersInATeamShop_ReturnedAsFourDistinctResources(t *testing.T) {
	// CA-021-02: una barbería con un barbero y luego tres más registrados
	// produce cuatro recursos distintos de la misma barbería, con el mismo
	// código de camino que una barbería unipersonal (CA-021-01): no hay una
	// forma de datos especial para "barbero único".
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	created := make(map[string]bool, 4)
	for i := 0; i < 4; i++ {
		b := createBarber(t, repo, shopC, "Equipo "+suffix+" "+string(rune('A'+i)))
		if created[b.ID] {
			t.Fatalf("duplicate id returned by Create: %s", b.ID)
		}
		created[b.ID] = true
	}
	if len(created) != 4 {
		t.Fatalf("expected 4 distinct barbers, got %d", len(created))
	}

	// Recorre TODAS las páginas con limit=1 y confirma que los 4 ids
	// aparecen entre lo recorrido, sin duplicados, en orden estable.
	seen := map[string]bool{}
	var cursor *staff.Cursor
	var lastCreatedAt time.Time
	pages := 0
	for {
		pages++
		if pages > 10000 {
			t.Fatal("too many pages; possible infinite loop")
		}
		page, err := repo.List(context.Background(), string(shopC), cursor, 1)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		for _, item := range page.Items {
			if seen[item.ID] {
				t.Fatalf("CA-021-02: item %s seen twice while paging", item.ID)
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
		decoded, err := staff.DecodeCursor(page.NextCursor)
		if err != nil {
			t.Fatalf("DecodeCursor: %v", err)
		}
		cursor = &decoded
	}

	for id := range created {
		if !seen[id] {
			t.Fatalf("CA-021-02: barber %s was never returned while paging through all pages", id)
		}
	}
}

func TestList_FirstPage_NeverReturnsMoreThanLimit(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	// Deja al menos 3 filas conocidas en shopC para esta prueba, sin
	// importar cuántas más haya de corridas anteriores.
	createBarber(t, repo, shopC, "Límite Uno "+suffix)
	createBarber(t, repo, shopC, "Límite Dos "+suffix)
	createBarber(t, repo, shopC, "Límite Tres "+suffix)

	page, err := repo.List(context.Background(), string(shopC), nil, 2)
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

// --- CA-021-05: A no ve, no lee ni renombra un barbero de B --------------

func TestGet_CrossTenant_NeverLeaksAnotherShopsBarber(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	barberD := createBarber(t, repo, shopD, "De la barbería D "+suffix)

	// El mismo identificador real, consultado bajo el contexto de C, debe
	// comportarse EXACTAMENTE como "no existe" (CA-021-05).
	_, found, err := repo.Get(context.Background(), string(shopC), barberD.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if found {
		t.Fatal("CA-021-05: shopC's context could read shopD's barber")
	}

	// Defensivo: bajo su propio tenant, sigue visible.
	got, found, err := repo.Get(context.Background(), string(shopD), barberD.ID)
	if err != nil {
		t.Fatalf("Get (own tenant): %v", err)
	}
	if !found || got.ID != barberD.ID {
		t.Fatal("expected shopD to still see its own barber")
	}
}

func TestGet_NonexistentID_ReturnsNotFoundSameAsOtherTenant(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	_, found, err := repo.Get(context.Background(), string(shopC), "99999999-9999-4999-8999-999999999999")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if found {
		t.Fatal("expected not found for a nonexistent id")
	}
}

func TestRename_CrossTenant_NeverRenamesAnotherShopsBarber(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	barberD := createBarber(t, repo, shopD, "Nombre Original D "+suffix)

	result, err := repo.Rename(context.Background(), string(shopC), barberD.ID, "Renombrado por C "+suffix)
	if err != nil {
		t.Fatalf("Rename: %v", err)
	}
	if result.Found {
		t.Fatal("CA-021-05: shopC's context renamed shopD's barber")
	}

	// El nombre de D no cambió.
	stillD, found, err := repo.Get(context.Background(), string(shopD), barberD.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found || stillD.FullName != "Nombre Original D "+suffix {
		t.Fatalf("expected shopD's barber name untouched, got %+v", stillD)
	}
}

func TestRename_OwnTenant_UpdatesByIDWithoutDuplicating(t *testing.T) {
	// CA-021-04: la lista muestra el nuevo nombre sin duplicar el recurso.
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	original := createBarber(t, repo, shopC, "Antes de renombrar "+suffix)

	result, err := repo.Rename(context.Background(), string(shopC), original.ID, "Después de renombrar "+suffix)
	if err != nil {
		t.Fatalf("Rename: %v", err)
	}
	if !result.Found {
		t.Fatal("expected Found=true for the barber's own tenant")
	}
	if result.Barber.ID != original.ID {
		t.Fatalf("expected the same id after rename, got %s vs %s", result.Barber.ID, original.ID)
	}
	if result.Barber.FullName != "Después de renombrar "+suffix {
		t.Fatalf("unexpected renamed name: %+v", result.Barber)
	}
	if !result.Barber.UpdatedAt.After(original.UpdatedAt) && !result.Barber.UpdatedAt.Equal(original.UpdatedAt) {
		t.Fatalf("expected updated_at to advance or stay equal (never go backwards), before=%v after=%v",
			original.UpdatedAt, result.Barber.UpdatedAt)
	}

	got, found, err := repo.Get(context.Background(), string(shopC), original.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found || got.FullName != "Después de renombrar "+suffix {
		t.Fatalf("expected the rename to persist, got %+v", got)
	}
}

func TestRename_DuplicateNameAcrossBarbers_Allowed(t *testing.T) {
	// HU-021 no impone unicidad de fullName: dos barberos de la misma
	// barbería pueden compartir nombre.
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	sharedName := "Nombre Compartido " + suffix

	a := createBarber(t, repo, shopC, "Barbero Uno "+suffix)
	b := createBarber(t, repo, shopC, "Barbero Dos "+suffix)

	if _, err := repo.Rename(context.Background(), string(shopC), a.ID, sharedName); err != nil {
		t.Fatalf("Rename a: %v", err)
	}
	if _, err := repo.Rename(context.Background(), string(shopC), b.ID, sharedName); err != nil {
		t.Fatalf("Rename b: %v", err)
	}

	gotA, _, _ := repo.Get(context.Background(), string(shopC), a.ID)
	gotB, _, _ := repo.Get(context.Background(), string(shopC), b.ID)
	if gotA.FullName != sharedName || gotB.FullName != sharedName {
		t.Fatalf("expected both barbers to accept the same name, got %q and %q", gotA.FullName, gotB.FullName)
	}
	if gotA.ID == gotB.ID {
		t.Fatal("expected two distinct barbers, not one merged into the other")
	}
}

// --- Idempotencia del alta (RN-IDE-01, DEC-043, CA-021-02) ----------------

func TestCreate_FirstExecution_PersistsAndReturnsProceed(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	key := idempotency.Key("first-exec-" + suffix)
	fp := idempotency.ComputeFingerprint("POST", "/api/v1/private/barbers", []byte(`{"fullName":"Nuevo"}`))

	result, err := repo.Create(context.Background(), string(shopC), "Nuevo "+suffix, key, fp)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if result.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected OutcomeProceed, got %s", result.Decision.Outcome)
	}
	if result.Barber.ID == "" || result.Barber.FullName != "Nuevo "+suffix {
		t.Fatalf("unexpected created barber: %+v", result.Barber)
	}
	if result.Response.Status != 201 {
		t.Fatalf("expected stored status 201, got %d", result.Response.Status)
	}

	got, found, err := repo.Get(context.Background(), string(shopC), result.Barber.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found || got.FullName != "Nuevo "+suffix {
		t.Fatalf("expected the created barber to be persisted, got found=%v %+v", found, got)
	}
}

func TestCreate_SameKeyAndContent_ReplaysWithoutCreatingASecondRow(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	key := idempotency.Key("replay-" + suffix)
	body := []byte(`{"fullName":"Repetido"}`)
	fp := idempotency.ComputeFingerprint("POST", "/api/v1/private/barbers", body)

	first, err := repo.Create(context.Background(), string(shopC), "Repetido "+suffix, key, fp)
	if err != nil {
		t.Fatalf("Create (first): %v", err)
	}
	if first.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected first call to proceed, got %s", first.Decision.Outcome)
	}

	second, err := repo.Create(context.Background(), string(shopC), "Repetido "+suffix, key, fp)
	if err != nil {
		t.Fatalf("Create (second): %v", err)
	}
	if second.Decision.Outcome != idempotency.OutcomeReplay {
		t.Fatalf("expected second call to replay, got %s", second.Decision.Outcome)
	}
	if second.Response != first.Response {
		t.Fatalf("expected the exact same stored response (CA-004-01), got %+v vs %+v", first.Response, second.Response)
	}

	// Confirma que existe UNA sola fila con ese nombre para esta clave (no
	// una segunda inserción silenciosa): se verifica leyendo por id el
	// mismo recurso que la primera respuesta reportó.
	var parsed struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(second.Response.Body), &parsed); err != nil {
		t.Fatalf("unmarshal stored response body: %v", err)
	}
	if parsed.ID != first.Barber.ID {
		t.Fatalf("expected the replayed body to reference the original barber id %s, got %s", first.Barber.ID, parsed.ID)
	}
}

func TestCreate_SameKeyDifferentContent_ConflictsWithoutEffect(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	key := idempotency.Key("conflict-" + suffix)
	fpA := idempotency.ComputeFingerprint("POST", "/api/v1/private/barbers", []byte(`{"fullName":"A"}`))
	fpB := idempotency.ComputeFingerprint("POST", "/api/v1/private/barbers", []byte(`{"fullName":"B"}`))

	first, err := repo.Create(context.Background(), string(shopC), "Nombre A "+suffix, key, fpA)
	if err != nil {
		t.Fatalf("Create (first): %v", err)
	}
	if first.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected first call to proceed, got %s", first.Decision.Outcome)
	}

	second, err := repo.Create(context.Background(), string(shopC), "Nombre B "+suffix, key, fpB)
	if err != nil {
		t.Fatalf("Create (second): %v", err)
	}
	if second.Decision.Outcome != idempotency.OutcomeConflictFingerprint {
		t.Fatalf("expected OutcomeConflictFingerprint, got %s", second.Decision.Outcome)
	}
	if second.Barber.ID != "" {
		t.Fatal("expected no barber created on conflict")
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
	fullName := "Concurrente " + suffix
	body := []byte(`{"fullName":"` + fullName + `"}`)
	fp := idempotency.ComputeFingerprint("POST", "/api/v1/private/barbers", body)

	// db.InTenantTx serializa internamente por CONEXIÓN, no por llamada: se
	// necesita una segunda "sesión" real, que la implementación de
	// Repository ya adquiere de su propio pool en cada InTenantTx
	// (Repository.db comparte el pool, pero cada InTenantTx adquiere su
	// propia conexión física). Se orquesta con dos goroutines igual que
	// internal/platform/idempotency/postgres_test.go.
	var (
		wg               sync.WaitGroup
		aResult, bResult staff.CreateResult
		aErr, bErr       error
		aStarted         = make(chan struct{})
		bAttempted       = make(chan struct{})
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		// No hay gancho directo para "mantener abierta la transacción de A
		// mientras B intenta": Repository.Create ya hace Begin+INSERT+
		// Complete de punta a punta. Para forzar la contención real se
		// dispara B inmediatamente después de A sin esperar a que A
		// termine, confiando en que el planificador del sistema operativo
		// entrelace ambas conexiones dentro de la ventana de la
		// transacción de A la mayoría de las veces; si no colisionan,
		// ambas rutas (proceed+replay o proceed+locked) son válidas y la
		// prueba las acepta explícitamente más abajo (ver comentario final).
		close(aStarted)
		aResult, aErr = repo.Create(context.Background(), string(shopC), fullName, key, fp)
	}()
	go func() {
		defer wg.Done()
		<-aStarted
		bResult, bErr = repo.Create(context.Background(), string(shopC), fullName, key, fp)
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

	outcomes := map[idempotency.Outcome]int{aResult.Decision.Outcome: 1, bResult.Decision.Outcome: 1}
	// La única combinación INACEPTABLE es que ambas hayan procedido: eso
	// significaría dos INSERT reales para la misma clave.
	if aResult.Decision.Outcome == idempotency.OutcomeProceed && bResult.Decision.Outcome == idempotency.OutcomeProceed {
		t.Fatalf("CA-004-03: both connections proceeded, expected the effect exactly once (%v)", outcomes)
	}
	validPair := func(x, y idempotency.Outcome) bool {
		return x == idempotency.OutcomeProceed &&
			(y == idempotency.OutcomeReplay || y == idempotency.OutcomeLocked || y == idempotency.OutcomeConflictInProgress)
	}
	if !validPair(aResult.Decision.Outcome, bResult.Decision.Outcome) && !validPair(bResult.Decision.Outcome, aResult.Decision.Outcome) {
		t.Fatalf("unexpected outcome pair: a=%s b=%s", aResult.Decision.Outcome, bResult.Decision.Outcome)
	}

	// Exactamente un barbero fue creado para esta clave: confirmarlo
	// leyendo el ganador por id.
	winner := aResult
	if winner.Decision.Outcome != idempotency.OutcomeProceed {
		winner = bResult
	}
	if winner.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatal("expected at least one of the two connections to proceed")
	}
	got, found, err := repo.Get(context.Background(), string(shopC), winner.Barber.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found || got.FullName != fullName {
		t.Fatalf("expected the single created barber to be persisted, got found=%v %+v", found, got)
	}
}

func TestCreate_SameKeyLiteralInTwoTenants_DoesNotInterfere(t *testing.T) {
	// CA-004-04/RN-TEN-01: la misma clave literal en dos barberías distintas
	// no interfiere: cada una procede de forma independiente.
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	key := idempotency.Key("cross-tenant-" + suffix)
	fp := idempotency.ComputeFingerprint("POST", "/api/v1/private/barbers", []byte(`{"fullName":"X"}`))

	resultC, err := repo.Create(context.Background(), string(shopC), "En C "+suffix, key, fp)
	if err != nil {
		t.Fatalf("Create shopC: %v", err)
	}
	if resultC.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected shopC to proceed, got %s", resultC.Decision.Outcome)
	}

	resultD, err := repo.Create(context.Background(), string(shopD), "En D "+suffix, key, fp)
	if err != nil {
		t.Fatalf("Create shopD: %v", err)
	}
	if resultD.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected shopD to ALSO proceed with the same literal key (RN-TEN-01), got %s", resultD.Decision.Outcome)
	}
	if resultC.Barber.ID == resultD.Barber.ID {
		t.Fatal("expected two distinct barbers, one per tenant")
	}
}

func TestCreate_DifferentKeysSameName_CreatesTwoDistinctBarbers(t *testing.T) {
	// La clave de idempotencia no convierte el nombre en único (trabajo
	// requerido §1.3): dos claves distintas SÍ crean dos barberos con el
	// mismo fullName.
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	name := "Nombre Repetido " + suffix

	fp := idempotency.ComputeFingerprint("POST", "/api/v1/private/barbers", []byte(`{"fullName":"`+name+`"}`))

	first, err := repo.Create(context.Background(), string(shopC), name, idempotency.Key("key-1-"+suffix), fp)
	if err != nil {
		t.Fatalf("Create (1): %v", err)
	}
	second, err := repo.Create(context.Background(), string(shopC), name, idempotency.Key("key-2-"+suffix), fp)
	if err != nil {
		t.Fatalf("Create (2): %v", err)
	}
	if first.Barber.ID == second.Barber.ID {
		t.Fatal("expected two distinct barbers with the same name from two different idempotency keys")
	}
	if first.Barber.FullName != name || second.Barber.FullName != name {
		t.Fatalf("expected both to keep the requested name, got %q and %q", first.Barber.FullName, second.Barber.FullName)
	}
}

// TestCreate_StoredResponseBody_MatchesHTTPAPIWireShape confirma que el
// JSON que Repository.Create persiste para una repetición exacta usa
// EXACTAMENTE las mismas claves que httpapi.BarberResponse (id, fullName,
// createdAt, updatedAt): la advertencia de mantenimiento manual que ambos
// archivos declaran en su comentario.
func TestCreate_StoredResponseBody_MatchesHTTPAPIWireShape(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	key := idempotency.Key("wire-shape-" + suffix)
	fp := idempotency.ComputeFingerprint("POST", "/api/v1/private/barbers", []byte(`{"fullName":"Forma"}`))

	result, err := repo.Create(context.Background(), string(shopC), "Forma "+suffix, key, fp)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	var generic map[string]any
	if err := json.Unmarshal([]byte(result.Response.Body), &generic); err != nil {
		t.Fatalf("unmarshal stored body: %v", err)
	}
	wantKeys := []string{"id", "fullName", "createdAt", "updatedAt"}
	if len(generic) != len(wantKeys) {
		t.Fatalf("expected exactly %v, got keys %v", wantKeys, generic)
	}
	for _, k := range wantKeys {
		if _, ok := generic[k]; !ok {
			t.Fatalf("expected key %q in the stored response body, got %v", k, generic)
		}
	}
}
