// Pruebas de integración de HU-041 (excepciones de jornada y festivos):
// mismo criterio que repository_test.go (HU-040) -PostgreSQL REAL, sin
// mocks para RLS/idempotencia/atomicidad (docs/03-desarrollo/
// estrategia-pruebas.md §2)- pero con su propio fixture aislado
// (testdata/hu041_excepciones.sql) para que un conteo exacto de
// excepciones por barbero no dependa del estado que dejen otras suites.
//
//	export TEST_DATABASE_URL="postgres://barberia_app@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/schedule/postgres/...
package postgres_test

import (
	"context"
	"encoding/json"
	"testing"

	"system-barbershop/internal/modules/schedule"
	schedulepostgres "system-barbershop/internal/modules/schedule/postgres"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

const (
	// shopM/shopN y sus barberos son los de testdata/hu041_excepciones.sql:
	// aislados de dos_barberias.sql/hu040_horario.sql/etc. por el mismo
	// motivo que esos fixtures ya documentan.
	shopM    = database.BarbershopID("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
	shopN    = database.BarbershopID("ffffffff-ffff-ffff-ffff-ffffffffffff")
	barberM1 = "eeee0001-0001-0001-0001-000100010001"
	barberM2 = "eeee0002-0002-0002-0002-000200020002"
	barberN1 = "ffff0001-0001-0001-0001-000100010001"
)

// createException crea una excepción con una clave de idempotencia única y
// programa su retiro al terminar la prueba.
func createException(t *testing.T, repo *schedulepostgres.Repository, shop database.BarbershopID, barberID string, input schedule.CreateExceptionInput) schedule.ScheduleException {
	t.Helper()
	key := idempotency.Key("exc-fixture-" + uniqueSuffix(t))
	fp := idempotency.ComputeFingerprint("POST", "/api/v1/private/barbers/"+barberID+"/schedule-exceptions", []byte(uniqueSuffix(t)))

	result, err := repo.CreateException(context.Background(), string(shop), barberID, input, key, fp)
	if err != nil {
		t.Fatalf("createException: %v", err)
	}
	if result.Decision.Outcome != idempotency.OutcomeProceed || result.Conflict {
		t.Fatalf("createException: expected a clean OutcomeProceed, got decision=%s conflict=%v", result.Decision.Outcome, result.Conflict)
	}
	e := result.Exception
	t.Cleanup(func() {
		_, _ = repo.DeleteException(context.Background(), string(shop), barberID, e.ID)
	})
	return e
}

// --- CA-004-01/RN-IDE-01: idempotencia real, wire shape --------------------

func TestCreateException_StoredResponseBody_MatchesHTTPAPIWireShape(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	key := idempotency.Key("exc-wire-shape-" + uniqueSuffix(t))
	fp := idempotency.ComputeFingerprint("POST", "/x", []byte("exc-wire-shape"))
	input := schedule.CreateExceptionInput{
		EffectiveDate: "2027-03-15",
		IsClosed:      true,
	}
	result, err := repo.CreateException(context.Background(), string(shopM), barberM1, input, key, fp)
	if err != nil {
		t.Fatalf("CreateException: %v", err)
	}
	if result.Conflict {
		t.Fatalf("unexpected conflict creating the wire-shape fixture")
	}
	t.Cleanup(func() {
		_, _ = repo.DeleteException(context.Background(), string(shopM), barberM1, result.Exception.ID)
	})

	var decoded map[string]any
	if err := json.Unmarshal([]byte(result.Response.Body), &decoded); err != nil {
		t.Fatalf("unmarshal stored response: %v", err)
	}
	for _, field := range []string{"id", "effectiveDate", "isClosed", "reason", "segments", "createdAt", "updatedAt"} {
		if _, ok := decoded[field]; !ok {
			t.Fatalf("expected stored response to contain field %q, got %v", field, decoded)
		}
	}
}

func TestCreateException_RepeatedKeySameBody_ReplaysWithoutSecondRow(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	key := idempotency.Key("exc-replay-" + uniqueSuffix(t))
	input := schedule.CreateExceptionInput{EffectiveDate: "2027-03-16", IsClosed: true}
	fp := idempotency.ComputeFingerprint("POST", "/x", []byte("exc-replay-body"))

	first, err := repo.CreateException(context.Background(), string(shopM), barberM1, input, key, fp)
	if err != nil {
		t.Fatalf("first CreateException: %v", err)
	}
	t.Cleanup(func() { _, _ = repo.DeleteException(context.Background(), string(shopM), barberM1, first.Exception.ID) })
	if first.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected first call to proceed, got %s", first.Decision.Outcome)
	}

	second, err := repo.CreateException(context.Background(), string(shopM), barberM1, input, key, fp)
	if err != nil {
		t.Fatalf("second CreateException: %v", err)
	}
	if second.Decision.Outcome != idempotency.OutcomeReplay {
		t.Fatalf("expected second call to replay, got %s", second.Decision.Outcome)
	}
	if second.Response.Body != first.Response.Body {
		t.Fatalf("expected byte-identical replayed body, got different bodies")
	}
}

// --- Calendario de festivos (CA-041-01/02) ---------------------------------

func TestHolidayCalendarEnabled_UnknownBarber_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	_, found, err := repo.GetHolidayCalendarEnabled(context.Background(), string(shopM), "00000000-0000-0000-0000-000000000099")
	if err != nil {
		t.Fatalf("GetHolidayCalendarEnabled: %v", err)
	}
	if found {
		t.Fatal("expected found=false for an unknown barber")
	}
}

func TestSetHolidayCalendarEnabled_ThenGet_PersistsToggle(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	setResult, err := repo.SetHolidayCalendarEnabled(context.Background(), string(shopM), barberM2, false)
	if err != nil {
		t.Fatalf("SetHolidayCalendarEnabled(false): %v", err)
	}
	if !setResult.Found || setResult.Enabled {
		t.Fatalf("expected found=true enabled=false, got %+v", setResult)
	}
	t.Cleanup(func() { _, _ = repo.SetHolidayCalendarEnabled(context.Background(), string(shopM), barberM2, true) })

	enabled, found, err := repo.GetHolidayCalendarEnabled(context.Background(), string(shopM), barberM2)
	if err != nil {
		t.Fatalf("GetHolidayCalendarEnabled: %v", err)
	}
	if !found || enabled {
		t.Fatalf("expected found=true enabled=false after set, got found=%v enabled=%v", found, enabled)
	}

	setResult, err = repo.SetHolidayCalendarEnabled(context.Background(), string(shopM), barberM2, true)
	if err != nil {
		t.Fatalf("SetHolidayCalendarEnabled(true): %v", err)
	}
	if !setResult.Found || !setResult.Enabled {
		t.Fatalf("expected found=true enabled=true, got %+v", setResult)
	}
}

func TestSetHolidayCalendarEnabled_UnknownBarber_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	result, err := repo.SetHolidayCalendarEnabled(context.Background(), string(shopM), "00000000-0000-0000-0000-000000000098", true)
	if err != nil {
		t.Fatalf("SetHolidayCalendarEnabled: %v", err)
	}
	if result.Found {
		t.Fatal("expected found=false for an unknown barber")
	}
}

// --- CA-041-04/05: forma, fecha duplicada y solape de tramos ---------------

func TestCreateException_DuplicateDate_ReturnsConflict(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	createException(t, repo, shopM, barberM1, schedule.CreateExceptionInput{EffectiveDate: "2027-04-01", IsClosed: true})

	key := idempotency.Key("exc-dup-date-" + uniqueSuffix(t))
	fp := idempotency.ComputeFingerprint("POST", "/x", []byte("dup-date"))
	result, err := repo.CreateException(context.Background(), string(shopM), barberM1, schedule.CreateExceptionInput{EffectiveDate: "2027-04-01", IsClosed: true}, key, fp)
	if err != nil {
		t.Fatalf("CreateException: %v", err)
	}
	if !result.Conflict {
		t.Fatal("expected Conflict=true for a duplicate effective_date for the same barber")
	}
}

func TestCreateException_OverlappingSegments_ReturnsConflictAndNothingPersists(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	key := idempotency.Key("exc-seg-overlap-" + uniqueSuffix(t))
	fp := idempotency.ComputeFingerprint("POST", "/x", []byte("seg-overlap"))
	input := schedule.CreateExceptionInput{
		EffectiveDate: "2027-04-02",
		IsClosed:      false,
		Segments: []schedule.CreateExceptionSegmentInput{
			{StartsTime: "08:00", DurationMinutes: 120},
			{StartsTime: "09:00", DurationMinutes: 60},
		},
	}
	result, err := repo.CreateException(context.Background(), string(shopM), barberM1, input, key, fp)
	if err != nil {
		t.Fatalf("CreateException: %v", err)
	}
	if !result.Conflict {
		t.Fatal("expected Conflict=true for overlapping segments")
	}

	_, found, err := repo.GetExceptionByDate(context.Background(), string(shopM), barberM1, "2027-04-02")
	if err != nil {
		t.Fatalf("GetExceptionByDate: %v", err)
	}
	if found {
		t.Fatal("expected nothing to persist when segment insertion conflicts: the ROLLBACK must cover the header too")
	}
}

func TestCreateException_SplitOpenDay_TwoNonOverlappingSegments(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	e := createException(t, repo, shopM, barberM1, schedule.CreateExceptionInput{
		EffectiveDate: "2027-04-03",
		IsClosed:      false,
		Segments: []schedule.CreateExceptionSegmentInput{
			{StartsTime: "08:00", DurationMinutes: 120},
			{StartsTime: "14:00", DurationMinutes: 120},
		},
	})
	if len(e.Segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(e.Segments))
	}
}

// --- CA-041-05: aislamiento de tenant y de barbero --------------------------

func TestGetException_CrossBarber_NeverLeaksAnotherBarbersException(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	e := createException(t, repo, shopM, barberM1, schedule.CreateExceptionInput{EffectiveDate: "2027-05-01", IsClosed: true})

	_, found, err := repo.GetException(context.Background(), string(shopM), barberM2, e.ID)
	if err != nil {
		t.Fatalf("GetException: %v", err)
	}
	if found {
		t.Fatal("CA-041-05: barberM2's context could read barberM1's exception")
	}
}

func TestGetException_CrossTenant_NeverLeaksAnotherShopsException(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	e := createException(t, repo, shopN, barberN1, schedule.CreateExceptionInput{EffectiveDate: "2027-05-02", IsClosed: true})

	_, found, err := repo.GetException(context.Background(), string(shopM), barberM1, e.ID)
	if err != nil {
		t.Fatalf("GetException: %v", err)
	}
	if found {
		t.Fatal("CA-041-05: shopM's context could read shopN's exception")
	}
}

func TestUpdateException_CrossTenant_NeverEditsAnotherShopsException(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	e := createException(t, repo, shopN, barberN1, schedule.CreateExceptionInput{EffectiveDate: "2027-05-03", IsClosed: true})

	result, err := repo.UpdateException(context.Background(), string(shopM), barberM1, e.ID, schedule.UpdateExceptionInput{EffectiveDate: "2027-05-04", IsClosed: true})
	if err != nil {
		t.Fatalf("UpdateException: %v", err)
	}
	if result.Found {
		t.Fatal("CA-041-05: shopM's context could update shopN's exception")
	}
}

func TestDeleteException_CrossTenant_NeverDeletesAnotherShopsException(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	e := createException(t, repo, shopN, barberN1, schedule.CreateExceptionInput{EffectiveDate: "2027-05-05", IsClosed: true})

	found, err := repo.DeleteException(context.Background(), string(shopM), barberM1, e.ID)
	if err != nil {
		t.Fatalf("DeleteException: %v", err)
	}
	if found {
		t.Fatal("CA-041-05: shopM's context could delete shopN's exception")
	}

	_, found, err = repo.GetException(context.Background(), string(shopN), barberN1, e.ID)
	if err != nil {
		t.Fatalf("GetException (own tenant): %v", err)
	}
	if !found {
		t.Fatal("expected shopN to still see its own exception")
	}
}

// --- UpdateException: reemplazo completo ------------------------------------

func TestUpdateException_NoConflict_ReplacesDateAndSegments(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	e := createException(t, repo, shopM, barberM1, schedule.CreateExceptionInput{
		EffectiveDate: "2027-06-01",
		IsClosed:      false,
		Segments:      []schedule.CreateExceptionSegmentInput{{StartsTime: "08:00", DurationMinutes: 60}},
	})

	result, err := repo.UpdateException(context.Background(), string(shopM), barberM1, e.ID, schedule.UpdateExceptionInput{
		EffectiveDate: "2027-06-02",
		IsClosed:      true,
	})
	if err != nil {
		t.Fatalf("UpdateException: %v", err)
	}
	if !result.Found || result.Conflict {
		t.Fatalf("expected a clean update, got found=%v conflict=%v", result.Found, result.Conflict)
	}
	if result.Exception.EffectiveDate != "2027-06-02" || !result.Exception.IsClosed || len(result.Exception.Segments) != 0 {
		t.Fatalf("expected the exception to become closed on the new date with no segments, got %+v", result.Exception)
	}
}

func TestUpdateException_ConflictWithAnotherDate_LeavesOriginalUntouched(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	createException(t, repo, shopM, barberM2, schedule.CreateExceptionInput{EffectiveDate: "2027-06-10", IsClosed: true})
	moving := createException(t, repo, shopM, barberM2, schedule.CreateExceptionInput{
		EffectiveDate: "2027-06-11",
		IsClosed:      false,
		Segments:      []schedule.CreateExceptionSegmentInput{{StartsTime: "09:00", DurationMinutes: 60}},
	})

	result, err := repo.UpdateException(context.Background(), string(shopM), barberM2, moving.ID, schedule.UpdateExceptionInput{EffectiveDate: "2027-06-10", IsClosed: true})
	if err != nil {
		t.Fatalf("UpdateException: %v", err)
	}
	if !result.Conflict {
		t.Fatal("expected Conflict=true when the new date collides with another exception of the same barber")
	}

	unchanged, found, err := repo.GetException(context.Background(), string(shopM), barberM2, moving.ID)
	if err != nil {
		t.Fatalf("GetException: %v", err)
	}
	if !found || unchanged.EffectiveDate != "2027-06-11" || len(unchanged.Segments) != 1 {
		t.Fatalf("expected the moving exception to keep its original date and segment, got %+v", unchanged)
	}
}

func TestUpdateException_NonexistentID_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	result, err := repo.UpdateException(context.Background(), string(shopM), barberM1, "00000000-0000-0000-0000-000000000001", schedule.UpdateExceptionInput{EffectiveDate: "2027-06-20", IsClosed: true})
	if err != nil {
		t.Fatalf("UpdateException: %v", err)
	}
	if result.Found {
		t.Fatal("expected Found=false for a nonexistent id")
	}
}

// --- DeleteException: retiro físico, seguro ante reintento ------------------

func TestDeleteException_ThenGet_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	e := createException(t, repo, shopM, barberM1, schedule.CreateExceptionInput{EffectiveDate: "2027-07-01", IsClosed: true})

	found, err := repo.DeleteException(context.Background(), string(shopM), barberM1, e.ID)
	if err != nil {
		t.Fatalf("DeleteException: %v", err)
	}
	if !found {
		t.Fatal("expected the first DeleteException to find and remove the exception")
	}

	_, found, err = repo.GetException(context.Background(), string(shopM), barberM1, e.ID)
	if err != nil {
		t.Fatalf("GetException after delete: %v", err)
	}
	if found {
		t.Fatal("expected the exception to be gone after DeleteException")
	}
}

func TestDeleteException_Retried_SecondCallIsSafeNoOp(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	e := createException(t, repo, shopM, barberM1, schedule.CreateExceptionInput{EffectiveDate: "2027-07-02", IsClosed: true})

	first, err := repo.DeleteException(context.Background(), string(shopM), barberM1, e.ID)
	if err != nil || !first {
		t.Fatalf("first DeleteException: found=%v err=%v", first, err)
	}
	second, err := repo.DeleteException(context.Background(), string(shopM), barberM1, e.ID)
	if err != nil {
		t.Fatalf("second DeleteException: %v", err)
	}
	if second {
		t.Fatal("expected the second DeleteException of an already-removed exception to report found=false")
	}
}

// --- ListExceptions: orden estable (effective_date, id) ---------------------

func TestListExceptions_OrderedByEffectiveDateThenID(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	late := createException(t, repo, shopN, barberN1, schedule.CreateExceptionInput{EffectiveDate: "2027-08-20", IsClosed: true})
	early := createException(t, repo, shopN, barberN1, schedule.CreateExceptionInput{EffectiveDate: "2027-08-10", IsClosed: true})

	page, err := repo.ListExceptions(context.Background(), string(shopN), barberN1, nil, 50)
	if err != nil {
		t.Fatalf("ListExceptions: %v", err)
	}

	indexOf := func(id string) int {
		for i, item := range page.Items {
			if item.ID == id {
				return i
			}
		}
		return -1
	}
	iEarly, iLate := indexOf(early.ID), indexOf(late.ID)
	if iEarly == -1 || iLate == -1 {
		t.Fatalf("expected both exceptions to appear in the listing")
	}
	if iEarly >= iLate {
		t.Fatalf("expected the 2027-08-10 exception (index %d) to sort before the 2027-08-20 one (index %d)", iEarly, iLate)
	}
}

func TestListExceptions_FirstPage_NeverReturnsMoreThanLimit(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	createException(t, repo, shopN, barberN1, schedule.CreateExceptionInput{EffectiveDate: "2027-09-01", IsClosed: true})
	createException(t, repo, shopN, barberN1, schedule.CreateExceptionInput{EffectiveDate: "2027-09-02", IsClosed: true})
	createException(t, repo, shopN, barberN1, schedule.CreateExceptionInput{EffectiveDate: "2027-09-03", IsClosed: true})

	page, err := repo.ListExceptions(context.Background(), string(shopN), barberN1, nil, 2)
	if err != nil {
		t.Fatalf("ListExceptions: %v", err)
	}
	if len(page.Items) != 2 {
		t.Fatalf("expected exactly 2 items for limit=2, got %d", len(page.Items))
	}
	if page.NextCursor == "" {
		t.Fatal("expected a next cursor: there are at least 3 known items and limit was 2")
	}
	decoded, err := schedule.DecodeExceptionCursor(page.NextCursor)
	if err != nil {
		t.Fatalf("DecodeExceptionCursor: %v", err)
	}

	next, err := repo.ListExceptions(context.Background(), string(shopN), barberN1, &decoded, 50)
	if err != nil {
		t.Fatalf("ListExceptions (second page): %v", err)
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

// --- GetExceptionByDate: la usa ResolveEffectiveDay internamente -----------

func TestGetExceptionByDate_FoundAndNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	e := createException(t, repo, shopM, barberM1, schedule.CreateExceptionInput{EffectiveDate: "2027-10-05", IsClosed: true})

	found, ok, err := repo.GetExceptionByDate(context.Background(), string(shopM), barberM1, "2027-10-05")
	if err != nil {
		t.Fatalf("GetExceptionByDate: %v", err)
	}
	if !ok || found.ID != e.ID {
		t.Fatalf("expected to find the exception created for 2027-10-05, got ok=%v found=%+v", ok, found)
	}

	_, ok, err = repo.GetExceptionByDate(context.Background(), string(shopM), barberM1, "2027-10-06")
	if err != nil {
		t.Fatalf("GetExceptionByDate: %v", err)
	}
	if ok {
		t.Fatal("expected no exception for a date without one")
	}
}

// --- ListWorkingHoursForWeekday: la usa ResolveEffectiveDay internamente ---

func TestListWorkingHoursForWeekday_OnlyThatBarberAndWeekday(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	monday := createWorkingHour(t, repo, shopM, barberM1, schedule.CreateInput{ISOWeekday: 1, StartsTime: "08:00", DurationMinutes: 240})
	createWorkingHour(t, repo, shopM, barberM1, schedule.CreateInput{ISOWeekday: 2, StartsTime: "08:00", DurationMinutes: 240})
	createWorkingHour(t, repo, shopM, barberM2, schedule.CreateInput{ISOWeekday: 1, StartsTime: "10:00", DurationMinutes: 60})

	items, err := repo.ListWorkingHoursForWeekday(context.Background(), string(shopM), barberM1, 1)
	if err != nil {
		t.Fatalf("ListWorkingHoursForWeekday: %v", err)
	}
	if len(items) != 1 || items[0].ID != monday.ID {
		t.Fatalf("expected exactly the Monday segment of barberM1, got %+v", items)
	}
}
