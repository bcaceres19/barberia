// Package postgres_test (pruebas de integración, HU-040) requiere
// PostgreSQL REAL con las trece migraciones aplicadas (incluida
// 20260825160000_create_working_hour.sql) y
// database/testdata/hu040_horario.sql cargado. Conéctate como
// barberia_app (docs/03-desarrollo/estrategia-pruebas.md §2 prohíbe mocks
// para RLS/idempotencia/atomicidad).
//
//	export TEST_DATABASE_URL="postgres://barberia_app@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/schedule/postgres/...
package postgres_test

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"system-barbershop/internal/modules/schedule"
	schedulepostgres "system-barbershop/internal/modules/schedule/postgres"
	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

const (
	testDatabaseURL = "postgres://barberia_app@localhost:5432/barberia_test?sslmode=disable"

	// shopK/shopL y sus barberos son los de testdata/hu040_horario.sql:
	// aislados de dos_barberias.sql/hu021_barberos.sql/etc. por el mismo
	// motivo que esos fixtures ya documentan.
	shopK    = database.BarbershopID("cccccccc-cccc-cccc-cccc-cccccccccccc")
	shopL    = database.BarbershopID("dddddddd-dddd-dddd-dddd-dddddddddddd")
	barberK1 = "cccc0001-0001-0001-0001-000100010001"
	barberK2 = "cccc0002-0002-0002-0002-000200020002"
	barberL1 = "dddd0001-0001-0001-0001-000100010001"
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

func newRepository(db *database.DB) *schedulepostgres.Repository {
	return schedulepostgres.New(db, idempotency.NewSQLCoordinator())
}

// createWorkingHour crea un tramo con una clave de idempotencia única y
// programa su retiro al terminar la prueba (working_hour SÍ admite DELETE,
// a diferencia de barber/service): cada prueba deja la base tal como la
// encontró.
func createWorkingHour(t *testing.T, repo *schedulepostgres.Repository, shop database.BarbershopID, barberID string, input schedule.CreateInput) schedule.WorkingHour {
	t.Helper()
	key := idempotency.Key("fixture-" + uniqueSuffix(t))
	body := `{"isoWeekday":` + itoa(input.ISOWeekday) + `,"startsTime":"` + input.StartsTime + `","durationMinutes":` + itoa(input.DurationMinutes) + `}`
	fp := idempotency.ComputeFingerprint("POST", "/api/v1/private/barbers/"+barberID+"/working-hours", []byte(body))

	result, err := repo.Create(context.Background(), string(shop), barberID, input, key, fp)
	if err != nil {
		t.Fatalf("createWorkingHour: %v", err)
	}
	if result.Decision.Outcome != idempotency.OutcomeProceed || result.Conflict {
		t.Fatalf("createWorkingHour: expected a clean OutcomeProceed, got decision=%s conflict=%v", result.Decision.Outcome, result.Conflict)
	}
	wh := result.WorkingHour
	t.Cleanup(func() {
		_, _ = repo.Delete(context.Background(), string(shop), barberID, wh.ID)
	})
	return wh
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func uniqueSuffix(t *testing.T) string {
	t.Helper()
	// time.Now().UnixNano() más el nombre de la prueba basta para no
	// colisionar entre pruebas de la misma corrida; cada prueba además
	// limpia sus propias filas al terminar.
	return t.Name() + "-" + time.Now().Format("150405.000000000")
}

// --- CA-004-01/RN-IDE-01: idempotencia real, wire shape ------------------

// TestCreate_StoredResponseBody_MatchesHTTPAPIWireShape confirma que el
// cuerpo que Complete persiste (workingHourResponseWire) tiene EXACTAMENTE
// la misma forma que httpapi.WorkingHourResponse: mismos nombres de campo,
// mismo orden. Un cambio en cualquiera de los dos structs sin el otro
// rompería una repetición exacta de la clave de idempotencia (CA-004-01)
// de forma sutil, solo visible en el byte a byte de esta prueba.
func TestCreate_StoredResponseBody_MatchesHTTPAPIWireShape(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	key := idempotency.Key("wire-shape-" + uniqueSuffix(t))
	body := `{"isoWeekday":1,"startsTime":"08:00","durationMinutes":60}`
	fp := idempotency.ComputeFingerprint("POST", "/x", []byte(body))
	result, err := repo.Create(context.Background(), string(shopK), barberK1, schedule.CreateInput{ISOWeekday: 1, StartsTime: "08:00", DurationMinutes: 60}, key, fp)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if result.Conflict {
		t.Fatalf("unexpected conflict creating the wire-shape fixture")
	}
	t.Cleanup(func() { _, _ = repo.Delete(context.Background(), string(shopK), barberK1, result.WorkingHour.ID) })

	var decoded map[string]any
	if err := json.Unmarshal([]byte(result.Response.Body), &decoded); err != nil {
		t.Fatalf("unmarshal stored response: %v", err)
	}
	for _, field := range []string{"id", "isoWeekday", "startsTime", "durationMinutes", "createdAt", "updatedAt"} {
		if _, ok := decoded[field]; !ok {
			t.Fatalf("expected stored response to contain field %q, got %v", field, decoded)
		}
	}
}

func TestCreate_RepeatedKeySameBody_ReplaysWithoutSecondRow(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	key := idempotency.Key("replay-" + uniqueSuffix(t))
	input := schedule.CreateInput{ISOWeekday: 2, StartsTime: "09:00", DurationMinutes: 60}
	body := `{"isoWeekday":2,"startsTime":"09:00","durationMinutes":60}`
	fp := idempotency.ComputeFingerprint("POST", "/x", []byte(body))

	first, err := repo.Create(context.Background(), string(shopK), barberK1, input, key, fp)
	if err != nil {
		t.Fatalf("first Create: %v", err)
	}
	t.Cleanup(func() { _, _ = repo.Delete(context.Background(), string(shopK), barberK1, first.WorkingHour.ID) })
	if first.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected first call to proceed, got %s", first.Decision.Outcome)
	}

	second, err := repo.Create(context.Background(), string(shopK), barberK1, input, key, fp)
	if err != nil {
		t.Fatalf("second Create: %v", err)
	}
	if second.Decision.Outcome != idempotency.OutcomeReplay {
		t.Fatalf("expected second call to replay, got %s", second.Decision.Outcome)
	}
	if second.Response.Body != first.Response.Body {
		t.Fatalf("expected byte-identical replayed body, got different bodies")
	}
}

// --- CA-040-02/03: jornada partida y tramo nocturno -----------------------

func TestCreate_SplitShift_TwoNonOverlappingSegmentsSameDay(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	morning := createWorkingHour(t, repo, shopK, barberK1, schedule.CreateInput{ISOWeekday: 3, StartsTime: "08:00", DurationMinutes: 240})
	afternoon := createWorkingHour(t, repo, shopK, barberK1, schedule.CreateInput{ISOWeekday: 3, StartsTime: "14:00", DurationMinutes: 240})

	if morning.ID == afternoon.ID {
		t.Fatal("expected two distinct segments for a split shift")
	}
}

func TestCreate_NightShift_CrossesMidnightWithoutError(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	wh := createWorkingHour(t, repo, shopK, barberK1, schedule.CreateInput{ISOWeekday: 5, StartsTime: "22:00", DurationMinutes: 300})
	if wh.DurationMinutes != 300 {
		t.Fatalf("expected duration 300, got %d", wh.DurationMinutes)
	}
}

// --- CA-040-04: solape y repetición de hora de inicio ---------------------

func TestCreate_ExactSameStart_ReturnsConflict(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	createWorkingHour(t, repo, shopK, barberK2, schedule.CreateInput{ISOWeekday: 4, StartsTime: "08:00", DurationMinutes: 60})

	key := idempotency.Key("dup-start-" + uniqueSuffix(t))
	fp := idempotency.ComputeFingerprint("POST", "/x", []byte("dup-start"))
	result, err := repo.Create(context.Background(), string(shopK), barberK2, schedule.CreateInput{ISOWeekday: 4, StartsTime: "08:00", DurationMinutes: 30}, key, fp)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !result.Conflict {
		t.Fatal("expected Conflict=true for an exact duplicate day+start")
	}
}

func TestCreate_PartialOverlap_ReturnsConflict(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	createWorkingHour(t, repo, shopK, barberK2, schedule.CreateInput{ISOWeekday: 6, StartsTime: "08:00", DurationMinutes: 120})

	key := idempotency.Key("overlap-" + uniqueSuffix(t))
	fp := idempotency.ComputeFingerprint("POST", "/x", []byte("overlap"))
	// [08:00,10:00) ya existe; [09:00,11:00) se solapa en [09:00,10:00).
	result, err := repo.Create(context.Background(), string(shopK), barberK2, schedule.CreateInput{ISOWeekday: 6, StartsTime: "09:00", DurationMinutes: 120}, key, fp)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !result.Conflict {
		t.Fatal("expected Conflict=true for a partial overlap")
	}
}

func TestCreate_ContiguousSegments_NoConflict(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	createWorkingHour(t, repo, shopK, barberK2, schedule.CreateInput{ISOWeekday: 7, StartsTime: "08:00", DurationMinutes: 120})
	// [10:00,10:30) empieza justo donde termina el anterior: contiguo, no
	// solapado (semántica semiabierta [inicio, fin), CA-040-03).
	second := createWorkingHour(t, repo, shopK, barberK2, schedule.CreateInput{ISOWeekday: 7, StartsTime: "10:00", DurationMinutes: 30})
	if second.StartsTime != "10:00" {
		t.Fatalf("expected the contiguous segment to be created, got %+v", second)
	}
}

// TestCreate_ConcurrentOverlappingCreates_ExactlyOneSucceeds es la prueba
// de la carrera que el bloqueo de la fila de barber (SELECT ... FOR UPDATE
// dentro de Repository.Create) existe para resolver: dos altas
// concurrentes REALES para el MISMO barbero y día, partiendo de CERO tramos
// existentes (el caso "phantom read" que un SELECT ... FOR UPDATE sobre
// working_hour por sí solo no bloquearía, porque no habría fila que
// bloquear). Exactamente una debe tener éxito; la otra debe volver con
// Conflict=true. El barbero jamás termina con dos tramos solapados.
func TestCreate_ConcurrentOverlappingCreates_ExactlyOneSucceeds(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	var (
		wg         sync.WaitGroup
		resultA    schedule.CreateResult
		resultB    schedule.CreateResult
		errA, errB error
		start      = make(chan struct{})
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		keyA := idempotency.Key("race-a-" + uniqueSuffix(t))
		fpA := idempotency.ComputeFingerprint("POST", "/x", []byte("race-a"))
		resultA, errA = repo.Create(context.Background(), string(shopL), barberL1, schedule.CreateInput{ISOWeekday: 2, StartsTime: "08:00", DurationMinutes: 120}, keyA, fpA)
	}()
	go func() {
		defer wg.Done()
		<-start
		keyB := idempotency.Key("race-b-" + uniqueSuffix(t))
		fpB := idempotency.ComputeFingerprint("POST", "/x", []byte("race-b"))
		resultB, errB = repo.Create(context.Background(), string(shopL), barberL1, schedule.CreateInput{ISOWeekday: 2, StartsTime: "09:00", DurationMinutes: 120}, keyB, fpB)
	}()
	close(start)
	wg.Wait()

	if errA != nil {
		t.Fatalf("goroutine A: %v", errA)
	}
	if errB != nil {
		t.Fatalf("goroutine B: %v", errB)
	}

	succeeded, conflicted := 0, 0
	var survivorID string
	for _, r := range []schedule.CreateResult{resultA, resultB} {
		switch {
		case r.Decision.Outcome == idempotency.OutcomeProceed && !r.Conflict:
			succeeded++
			survivorID = r.WorkingHour.ID
		case r.Conflict:
			conflicted++
		default:
			t.Fatalf("unexpected outcome in the race: decision=%s conflict=%v", r.Decision.Outcome, r.Conflict)
		}
	}
	if succeeded != 1 || conflicted != 1 {
		t.Fatalf("expected exactly one success and one conflict, got succeeded=%d conflicted=%d", succeeded, conflicted)
	}
	if survivorID != "" {
		t.Cleanup(func() { _, _ = repo.Delete(context.Background(), string(shopL), barberL1, survivorID) })
	}
}

// --- CA-040-05: aislamiento de tenant y de barbero ------------------------

func TestGet_CrossBarber_NeverLeaksAnotherBarbersWorkingHour(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	wh := createWorkingHour(t, repo, shopK, barberK1, schedule.CreateInput{ISOWeekday: 1, StartsTime: "07:00", DurationMinutes: 30})

	_, found, err := repo.Get(context.Background(), string(shopK), barberK2, wh.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if found {
		t.Fatal("CA-040-05: barberK2's context could read barberK1's working hour")
	}
}

func TestGet_CrossTenant_NeverLeaksAnotherShopsWorkingHour(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	wh := createWorkingHour(t, repo, shopL, barberL1, schedule.CreateInput{ISOWeekday: 1, StartsTime: "06:00", DurationMinutes: 30})

	_, found, err := repo.Get(context.Background(), string(shopK), barberK1, wh.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if found {
		t.Fatal("CA-040-05: shopK's context could read shopL's working hour")
	}
}

func TestUpdate_CrossTenant_NeverEditsAnotherShopsWorkingHour(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	wh := createWorkingHour(t, repo, shopL, barberL1, schedule.CreateInput{ISOWeekday: 3, StartsTime: "06:00", DurationMinutes: 30})

	result, err := repo.Update(context.Background(), string(shopK), barberK1, wh.ID, schedule.UpdateInput{ISOWeekday: 3, StartsTime: "07:00", DurationMinutes: 30})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if result.Found {
		t.Fatal("CA-040-05: shopK's context could update shopL's working hour")
	}
}

func TestDelete_CrossTenant_NeverDeletesAnotherShopsWorkingHour(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	wh := createWorkingHour(t, repo, shopL, barberL1, schedule.CreateInput{ISOWeekday: 4, StartsTime: "06:00", DurationMinutes: 30})

	found, err := repo.Delete(context.Background(), string(shopK), barberK1, wh.ID)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if found {
		t.Fatal("CA-040-05: shopK's context could delete shopL's working hour")
	}

	// Sigue existiendo bajo su propio tenant.
	_, found, err = repo.Get(context.Background(), string(shopL), barberL1, wh.ID)
	if err != nil {
		t.Fatalf("Get (own tenant): %v", err)
	}
	if !found {
		t.Fatal("expected shopL to still see its own working hour")
	}
}

// --- Update: edición sin solape, con solape, no encontrado ----------------

func TestUpdate_NoConflict_ChangesInterval(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	wh := createWorkingHour(t, repo, shopK, barberK1, schedule.CreateInput{ISOWeekday: 2, StartsTime: "11:00", DurationMinutes: 60})

	result, err := repo.Update(context.Background(), string(shopK), barberK1, wh.ID, schedule.UpdateInput{ISOWeekday: 2, StartsTime: "12:00", DurationMinutes: 90})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !result.Found || result.Conflict {
		t.Fatalf("expected a clean update, got found=%v conflict=%v", result.Found, result.Conflict)
	}
	if result.WorkingHour.StartsTime != "12:00" || result.WorkingHour.DurationMinutes != 90 {
		t.Fatalf("expected the updated interval to persist, got %+v", result.WorkingHour)
	}
}

func TestUpdate_SameSegmentNoRealChange_DoesNotConflictWithItself(t *testing.T) {
	// Editar un tramo sin mover su intervalo (por ejemplo, solo para
	// confirmar el mismo valor) no debe chocar consigo mismo: la
	// verificación de solape EXCLUYE explícitamente workingHourID.
	db := setupTestDB(t)
	repo := newRepository(db)

	wh := createWorkingHour(t, repo, shopK, barberK1, schedule.CreateInput{ISOWeekday: 5, StartsTime: "13:00", DurationMinutes: 45})

	result, err := repo.Update(context.Background(), string(shopK), barberK1, wh.ID, schedule.UpdateInput{ISOWeekday: 5, StartsTime: "13:00", DurationMinutes: 45})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !result.Found || result.Conflict {
		t.Fatalf("expected updating a segment to its own current value to succeed, got found=%v conflict=%v", result.Found, result.Conflict)
	}
}

func TestUpdate_OverlapsAnotherSegment_ReturnsConflictWithoutChangingAnything(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	kept := createWorkingHour(t, repo, shopK, barberK2, schedule.CreateInput{ISOWeekday: 1, StartsTime: "15:00", DurationMinutes: 60})
	moving := createWorkingHour(t, repo, shopK, barberK2, schedule.CreateInput{ISOWeekday: 1, StartsTime: "17:00", DurationMinutes: 60})

	result, err := repo.Update(context.Background(), string(shopK), barberK2, moving.ID, schedule.UpdateInput{ISOWeekday: 1, StartsTime: "15:30", DurationMinutes: 60})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !result.Conflict {
		t.Fatal("expected Conflict=true when the new interval overlaps another existing segment")
	}

	// El tramo movido conserva su intervalo original: nada se persistió.
	unchanged, found, err := repo.Get(context.Background(), string(shopK), barberK2, moving.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found || unchanged.StartsTime != "17:00" {
		t.Fatalf("expected the moving segment to keep its original interval, got %+v", unchanged)
	}
	_ = kept
}

func TestUpdate_NonexistentID_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	result, err := repo.Update(context.Background(), string(shopK), barberK1, "00000000-0000-0000-0000-000000000001", schedule.UpdateInput{ISOWeekday: 1, StartsTime: "08:00", DurationMinutes: 30})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if result.Found {
		t.Fatal("expected Found=false for a nonexistent id")
	}
}

// --- Delete: retiro físico, seguro ante reintento -------------------------

func TestDelete_ThenGet_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	wh := createWorkingHour(t, repo, shopK, barberK1, schedule.CreateInput{ISOWeekday: 6, StartsTime: "16:00", DurationMinutes: 30})

	found, err := repo.Delete(context.Background(), string(shopK), barberK1, wh.ID)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !found {
		t.Fatal("expected the first Delete to find and remove the segment")
	}

	_, found, err = repo.Get(context.Background(), string(shopK), barberK1, wh.ID)
	if err != nil {
		t.Fatalf("Get after delete: %v", err)
	}
	if found {
		t.Fatal("expected the segment to be gone after Delete")
	}
}

func TestDelete_Retried_SecondCallIsSafeNoOp(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	wh := createWorkingHour(t, repo, shopK, barberK1, schedule.CreateInput{ISOWeekday: 7, StartsTime: "16:00", DurationMinutes: 30})

	first, err := repo.Delete(context.Background(), string(shopK), barberK1, wh.ID)
	if err != nil || !first {
		t.Fatalf("first Delete: found=%v err=%v", first, err)
	}
	second, err := repo.Delete(context.Background(), string(shopK), barberK1, wh.ID)
	if err != nil {
		t.Fatalf("second Delete: %v", err)
	}
	if second {
		t.Fatal("expected the second Delete of an already-removed segment to report found=false")
	}
}

// --- List: orden estable por (iso_weekday, starts_time, id) --------------

func TestList_OrderedByWeekdayThenStartsTime(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	a := createWorkingHour(t, repo, shopL, barberL1, schedule.CreateInput{ISOWeekday: 7, StartsTime: "07:00", DurationMinutes: 30})
	b := createWorkingHour(t, repo, shopL, barberL1, schedule.CreateInput{ISOWeekday: 7, StartsTime: "18:00", DurationMinutes: 30})

	page, err := repo.List(context.Background(), string(shopL), barberL1, nil, 50)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	indexOf := func(id string) int {
		for i, item := range page.Items {
			if item.ID == id {
				return i
			}
		}
		return -1
	}
	ia, ib := indexOf(a.ID), indexOf(b.ID)
	if ia == -1 || ib == -1 {
		t.Fatalf("expected both segments to appear in the listing")
	}
	if ia >= ib {
		t.Fatalf("expected the 07:00 segment (index %d) to sort before the 18:00 segment (index %d)", ia, ib)
	}
}

func TestList_FirstPage_NeverReturnsMoreThanLimit(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	createWorkingHour(t, repo, shopL, barberL1, schedule.CreateInput{ISOWeekday: 6, StartsTime: "05:00", DurationMinutes: 10})
	createWorkingHour(t, repo, shopL, barberL1, schedule.CreateInput{ISOWeekday: 6, StartsTime: "06:00", DurationMinutes: 10})
	createWorkingHour(t, repo, shopL, barberL1, schedule.CreateInput{ISOWeekday: 6, StartsTime: "07:00", DurationMinutes: 10})

	page, err := repo.List(context.Background(), string(shopL), barberL1, nil, 2)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(page.Items) != 2 {
		t.Fatalf("expected exactly 2 items for limit=2, got %d", len(page.Items))
	}
	if page.NextCursor == "" {
		t.Fatal("expected a next cursor: there are at least 3 known items and limit was 2")
	}
	decoded, err := schedule.DecodeCursor(page.NextCursor)
	if err != nil {
		t.Fatalf("DecodeCursor: %v", err)
	}

	next, err := repo.List(context.Background(), string(shopL), barberL1, &decoded, 50)
	if err != nil {
		t.Fatalf("List (second page): %v", err)
	}
	seen := map[string]bool{}
	for _, item := range page.Items {
		seen[item.ID] = true
	}
	for _, item := range next.Items {
		if seen[item.ID] {
			t.Fatalf("item %s appeared in both pages", item.ID)
		}
	}
}
