// Pruebas de integración de HU-042 (bloqueos de agenda): mismo criterio que
// repository_test.go (HU-040)/exception_repository_test.go (HU-041) -
// PostgreSQL REAL, sin mocks para RLS/idempotencia/atomicidad
// (docs/03-desarrollo/estrategia-pruebas.md §2)- con su propio fixture
// aislado (testdata/hu042_bloqueos.sql).
//
//	export TEST_DATABASE_URL="postgres://barberia_app@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/schedule/postgres/...
package postgres_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"system-barbershop/internal/modules/schedule"
	schedulepostgres "system-barbershop/internal/modules/schedule/postgres"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

const (
	// shopO/shopP y sus barberos son los de testdata/hu042_bloqueos.sql:
	// aislados de dos_barberias.sql/hu040_horario.sql/hu041_excepciones.sql
	// por el mismo motivo que esos fixtures ya documentan.
	shopO    = database.BarbershopID("b10c0001-b10c-b10c-b10c-b10c00010001")
	shopP    = database.BarbershopID("b10c0002-b10c-b10c-b10c-b10c00020002")
	barberO1 = "b10c1001-1001-1001-1001-100110011001"
	barberO2 = "b10c1002-1002-1002-1002-100210021002"
	barberP1 = "b10c2001-2001-2001-2001-200120012001"
	ownerO   = "b10c9001-9001-9001-9001-900190019001"
)

func createBlock(t *testing.T, repo *schedulepostgres.Repository, shop database.BarbershopID, barberID string, input schedule.CreateBlockInput) schedule.TimeBlock {
	t.Helper()
	key := idempotency.Key("block-fixture-" + uniqueSuffix(t))
	fp := idempotency.ComputeFingerprint("POST", "/api/v1/private/barbers/"+barberID+"/time-blocks", []byte(uniqueSuffix(t)))

	result, err := repo.CreateBlock(context.Background(), string(shop), barberID, input, key, fp)
	if err != nil {
		t.Fatalf("createBlock: %v", err)
	}
	if result.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("createBlock: expected a clean OutcomeProceed, got %s", result.Decision.Outcome)
	}
	return result.Block
}

func createWeeklySeries(t *testing.T, repo *schedulepostgres.Repository, shop database.BarbershopID, barberID string, input schedule.CreateSeriesInput) schedule.TimeBlockSeries {
	t.Helper()
	key := idempotency.Key("series-fixture-" + uniqueSuffix(t))
	fp := idempotency.ComputeFingerprint("POST", "/api/v1/private/barbers/"+barberID+"/time-block-series", []byte(uniqueSuffix(t)))

	result, err := repo.CreateSeries(context.Background(), string(shop), barberID, input, key, fp)
	if err != nil {
		t.Fatalf("createWeeklySeries: %v", err)
	}
	if result.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("createWeeklySeries: expected a clean OutcomeProceed, got %s", result.Decision.Outcome)
	}
	return result.Series
}

// --- CA-042: idempotencia real, wire shape (bloqueo puntual) --------------

func TestCreateBlock_StoredResponseBody_MatchesHTTPAPIWireShape(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	key := idempotency.Key("block-wire-shape-" + uniqueSuffix(t))
	fp := idempotency.ComputeFingerprint("POST", "/x", []byte("block-wire-shape"))
	starts := time.Date(2027, 4, 10, 15, 0, 0, 0, time.UTC)
	input := schedule.CreateBlockInput{BlockType: schedule.BlockTypeEmergency, StartsAt: starts, EndsAt: starts.Add(2 * time.Hour)}

	result, err := repo.CreateBlock(context.Background(), string(shopO), barberO1, input, key, fp)
	if err != nil {
		t.Fatalf("CreateBlock: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal([]byte(result.Response.Body), &decoded); err != nil {
		t.Fatalf("unmarshal stored response: %v", err)
	}
	for _, field := range []string{"id", "blockType", "source", "startsAt", "endsAt", "reason", "deletedAt", "deletedBy", "createdAt", "updatedAt"} {
		if _, ok := decoded[field]; !ok {
			t.Fatalf("expected stored response to contain field %q, got %v", field, decoded)
		}
	}
	if decoded["source"] != "manual" {
		t.Fatalf("expected source=manual, got %v", decoded["source"])
	}
}

func TestCreateBlock_RepeatedKeySameBody_ReplaysWithoutSecondRow(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	key := idempotency.Key("block-replay-" + uniqueSuffix(t))
	starts := time.Date(2027, 4, 11, 15, 0, 0, 0, time.UTC)
	input := schedule.CreateBlockInput{BlockType: schedule.BlockTypeBreak, StartsAt: starts, EndsAt: starts.Add(15 * time.Minute)}
	fp := idempotency.ComputeFingerprint("POST", "/x", []byte("block-replay-body"))

	first, err := repo.CreateBlock(context.Background(), string(shopO), barberO1, input, key, fp)
	if err != nil {
		t.Fatalf("first CreateBlock: %v", err)
	}
	if first.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected first call to proceed, got %s", first.Decision.Outcome)
	}

	second, err := repo.CreateBlock(context.Background(), string(shopO), barberO1, input, key, fp)
	if err != nil {
		t.Fatalf("second CreateBlock: %v", err)
	}
	if second.Decision.Outcome != idempotency.OutcomeReplay {
		t.Fatalf("expected second call to replay, got %s", second.Decision.Outcome)
	}
	if second.Response.Body != first.Response.Body {
		t.Fatal("expected byte-identical replayed body")
	}
}

// --- RN-BLQ-03/DEC-008: un bloqueo puntual nunca falla por solape ---------

func TestCreateBlock_OverlappingIntervals_BothSucceed(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	starts := time.Date(2027, 4, 12, 15, 0, 0, 0, time.UTC)
	a := createBlock(t, repo, shopO, barberO1, schedule.CreateBlockInput{BlockType: schedule.BlockTypeUnavailable, StartsAt: starts, EndsAt: starts.Add(4 * time.Hour)})
	b := createBlock(t, repo, shopO, barberO1, schedule.CreateBlockInput{BlockType: schedule.BlockTypeEmergency, StartsAt: starts.Add(time.Hour), EndsAt: starts.Add(2 * time.Hour)})

	if a.ID == b.ID {
		t.Fatal("expected two distinct blocks, both persisted")
	}
}

// --- RN-BLQ-04: eliminación lógica, nunca física ---------------------------

func TestDeleteBlock_SoftDeletes_RecordSurvivesAndStopsBeingActive(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	starts := time.Date(2027, 4, 13, 15, 0, 0, 0, time.UTC)
	b := createBlock(t, repo, shopO, barberO1, schedule.CreateBlockInput{BlockType: schedule.BlockTypeVacation, StartsAt: starts, EndsAt: starts.Add(24 * time.Hour)})

	found, err := repo.DeleteBlock(context.Background(), string(shopO), barberO1, b.ID, ownerO)
	if err != nil {
		t.Fatalf("DeleteBlock: %v", err)
	}
	if !found {
		t.Fatal("expected found=true on first delete")
	}

	after, found, err := repo.GetBlock(context.Background(), string(shopO), barberO1, b.ID)
	if err != nil {
		t.Fatalf("GetBlock after delete: %v", err)
	}
	if !found {
		t.Fatal("RN-BLQ-04: el registro debe conservarse tras el retiro lógico")
	}
	if after.DeletedAt == nil || after.DeletedBy == nil || *after.DeletedBy != ownerO {
		t.Fatalf("expected DeletedAt/DeletedBy set, got %+v", after)
	}

	// Reintentar tras el retiro es seguro: found=false, no reescribe nada.
	foundAgain, err := repo.DeleteBlock(context.Background(), string(shopO), barberO1, b.ID, ownerO)
	if err != nil {
		t.Fatalf("DeleteBlock (retry): %v", err)
	}
	if foundAgain {
		t.Fatal("expected found=false on retry after already deleted")
	}

	// La lista normal (includeDeleted=false) ya no lo incluye.
	list, err := repo.ListBlocks(context.Background(), string(shopO), barberO1, nil, 50, false)
	if err != nil {
		t.Fatalf("ListBlocks: %v", err)
	}
	for _, item := range list.Items {
		if item.ID == b.ID {
			t.Fatal("un bloqueo retirado no debe aparecer en la lista activa (RN-BLQ-04)")
		}
	}

	// includeDeleted=true sí lo trae, para auditoría.
	listWithDeleted, err := repo.ListBlocks(context.Background(), string(shopO), barberO1, nil, 50, true)
	if err != nil {
		t.Fatalf("ListBlocks(includeDeleted): %v", err)
	}
	seen := false
	for _, item := range listWithDeleted.Items {
		if item.ID == b.ID {
			seen = true
		}
	}
	if !seen {
		t.Fatal("includeDeleted=true debe incluir el bloqueo retirado")
	}
}

// --- Aislamiento de tenant/barbero -----------------------------------------

func TestGetBlock_CrossTenant_NeverLeaksAnotherShopsBlock(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	starts := time.Date(2027, 4, 14, 15, 0, 0, 0, time.UTC)
	b := createBlock(t, repo, shopO, barberO1, schedule.CreateBlockInput{BlockType: schedule.BlockTypeDayOff, StartsAt: starts, EndsAt: starts.Add(24 * time.Hour)})

	_, found, err := repo.GetBlock(context.Background(), string(shopP), barberP1, b.ID)
	if err != nil {
		t.Fatalf("GetBlock cross-tenant: %v", err)
	}
	if found {
		t.Fatal("un bloqueo de otra barbería nunca debe ser visible (RN-TEN-01)")
	}
}

func TestGetBlock_CrossBarber_NeverLeaksAnotherBarbersBlock(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	starts := time.Date(2027, 4, 15, 15, 0, 0, 0, time.UTC)
	b := createBlock(t, repo, shopO, barberO1, schedule.CreateBlockInput{BlockType: schedule.BlockTypeHoliday, StartsAt: starts, EndsAt: starts.Add(24 * time.Hour)})

	_, found, err := repo.GetBlock(context.Background(), string(shopO), barberO2, b.ID)
	if err != nil {
		t.Fatalf("GetBlock cross-barber: %v", err)
	}
	if found {
		t.Fatal("un bloqueo de otro barbero nunca debe ser visible")
	}
}

// --- Series: weekly, date_list, dates/exceptions embebidos ----------------

func TestCreateSeries_Weekly_PersistsAndEmbedsEmptyChildren(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	monday := 1
	sr := createWeeklySeries(t, repo, shopO, barberO1, schedule.CreateSeriesInput{
		BlockType: schedule.BlockTypeLunch, RecurrenceKind: schedule.RecurrenceKindWeekly, ISOWeekday: &monday,
		StartsTime: "13:00", DurationMinutes: 60, EffectiveFrom: "2027-01-04",
	})
	if sr.RecurrenceKind != schedule.RecurrenceKindWeekly || sr.StartsTime != "13:00" || sr.DurationMinutes != 60 {
		t.Fatalf("unexpected series shape: %+v", sr)
	}
	if len(sr.Dates) != 0 || len(sr.Exceptions) != 0 {
		t.Fatalf("expected no children on a fresh series, got %+v", sr)
	}

	fetched, found, err := repo.GetSeries(context.Background(), string(shopO), barberO1, sr.ID)
	if err != nil {
		t.Fatalf("GetSeries: %v", err)
	}
	if !found || fetched.ID != sr.ID {
		t.Fatalf("expected to find the just-created series, got found=%v %+v", found, fetched)
	}
}

func TestCreateSeries_DateList_WithExplicitDates_CreatesAllInOneOperation(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	until := "2027-01-05"
	sr := createWeeklySeries(t, repo, shopO, barberO1, schedule.CreateSeriesInput{
		BlockType: schedule.BlockTypeVacation, RecurrenceKind: schedule.RecurrenceKindDateList,
		StartsTime: "00:00", DurationMinutes: 1440, EffectiveFrom: "2027-01-01", EffectiveUntil: &until,
		ExplicitDates: []string{"2027-01-01", "2027-01-02", "2027-01-05"},
	})
	if len(sr.Dates) != 3 {
		t.Fatalf("expected 3 explicit dates created in one operation (RN-BLQ-01), got %+v", sr.Dates)
	}
}

func TestAddSeriesDate_DuplicateDate_ReturnsConflictWithoutDuplicating(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	sr := createWeeklySeries(t, repo, shopO, barberO1, schedule.CreateSeriesInput{
		BlockType: schedule.BlockTypeVacation, RecurrenceKind: schedule.RecurrenceKindDateList,
		StartsTime: "00:00", DurationMinutes: 1440, EffectiveFrom: "2027-02-01",
	})

	found, conflict, err := repo.AddSeriesDate(context.Background(), string(shopO), barberO1, sr.ID, "2027-02-10")
	if err != nil {
		t.Fatalf("AddSeriesDate: %v", err)
	}
	if !found || conflict {
		t.Fatalf("expected found=true conflict=false on first add, got found=%v conflict=%v", found, conflict)
	}

	_, conflict, err = repo.AddSeriesDate(context.Background(), string(shopO), barberO1, sr.ID, "2027-02-10")
	if err != nil {
		t.Fatalf("AddSeriesDate (duplicate): %v", err)
	}
	if !conflict {
		t.Fatal("expected conflict=true on duplicate date")
	}

	fetched, _, err := repo.GetSeries(context.Background(), string(shopO), barberO1, sr.ID)
	if err != nil {
		t.Fatalf("GetSeries: %v", err)
	}
	if len(fetched.Dates) != 1 {
		t.Fatalf("expected exactly one date after a rejected duplicate, got %+v", fetched.Dates)
	}
}

func TestAddSeriesException_ThenRemove_RestoresInstance(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	friday := 5
	sr := createWeeklySeries(t, repo, shopO, barberO1, schedule.CreateSeriesInput{
		BlockType: schedule.BlockTypeUnavailable, RecurrenceKind: schedule.RecurrenceKindWeekly, ISOWeekday: &friday,
		StartsTime: "17:00", DurationMinutes: 120, EffectiveFrom: "2027-01-01",
	})

	found, conflict, err := repo.AddSeriesException(context.Background(), string(shopO), barberO1, sr.ID, "2027-01-08", nil)
	if err != nil || !found || conflict {
		t.Fatalf("AddSeriesException: found=%v conflict=%v err=%v", found, conflict, err)
	}

	fetched, _, _ := repo.GetSeries(context.Background(), string(shopO), barberO1, sr.ID)
	if len(fetched.Exceptions) != 1 || fetched.Exceptions[0].ExcludedDate != "2027-01-08" {
		t.Fatalf("expected one exception on 2027-01-08, got %+v", fetched.Exceptions)
	}

	removed, err := repo.RemoveSeriesException(context.Background(), string(shopO), barberO1, sr.ID, "2027-01-08")
	if err != nil || !removed {
		t.Fatalf("RemoveSeriesException: removed=%v err=%v", removed, err)
	}

	fetchedAfter, _, _ := repo.GetSeries(context.Background(), string(shopO), barberO1, sr.ID)
	if len(fetchedAfter.Exceptions) != 0 {
		t.Fatalf("expected the instance restored (no exceptions left), got %+v", fetchedAfter.Exceptions)
	}
}

// --- UpdateSeriesWhole / SplitSeriesFrom (RN-BLQ-01: alcance de edición) --

func TestUpdateSeriesWhole_ReplacesEditableFields(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	tuesday := 2
	sr := createWeeklySeries(t, repo, shopO, barberO1, schedule.CreateSeriesInput{
		BlockType: schedule.BlockTypeBreak, RecurrenceKind: schedule.RecurrenceKindWeekly, ISOWeekday: &tuesday,
		StartsTime: "10:00", DurationMinutes: 15, EffectiveFrom: "2027-01-01",
	})

	result, err := repo.UpdateSeriesWhole(context.Background(), string(shopO), barberO1, sr.ID, schedule.UpdateSeriesInput{
		BlockType: schedule.BlockTypeBreak, StartsTime: "10:30", DurationMinutes: 20, EffectiveFrom: "2027-01-01",
	})
	if err != nil {
		t.Fatalf("UpdateSeriesWhole: %v", err)
	}
	if !result.Found || result.Series.StartsTime != "10:30" || result.Series.DurationMinutes != 20 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestSplitSeriesFrom_TruncatesOriginalAndCreatesNewFromDate(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	wednesday := 3
	original := createWeeklySeries(t, repo, shopO, barberO1, schedule.CreateSeriesInput{
		BlockType: schedule.BlockTypeLunch, RecurrenceKind: schedule.RecurrenceKindWeekly, ISOWeekday: &wednesday,
		StartsTime: "13:00", DurationMinutes: 60, EffectiveFrom: "2027-01-01",
	})

	result, err := repo.SplitSeriesFrom(context.Background(), string(shopO), barberO1, original.ID, "2027-06-01", schedule.UpdateSeriesInput{
		BlockType: schedule.BlockTypeLunch, StartsTime: "14:00", DurationMinutes: 45, EffectiveFrom: "2027-06-01",
	})
	if err != nil {
		t.Fatalf("SplitSeriesFrom: %v", err)
	}
	if !result.Found {
		t.Fatal("expected found=true")
	}
	newSeries := result.Series
	if newSeries.ID == original.ID {
		t.Fatal("split debe crear una serie NUEVA distinta de la original")
	}
	if newSeries.StartsTime != "14:00" || newSeries.DurationMinutes != 45 || newSeries.EffectiveFrom != "2027-06-01" {
		t.Fatalf("unexpected new series shape: %+v", newSeries)
	}
	if newSeries.ISOWeekday == nil || *newSeries.ISOWeekday != wednesday {
		t.Fatalf("la serie nueva debe heredar iso_weekday de la original, got %+v", newSeries.ISOWeekday)
	}

	originalAfter, found, err := repo.GetSeries(context.Background(), string(shopO), barberO1, original.ID)
	if err != nil || !found {
		t.Fatalf("GetSeries(original) after split: found=%v err=%v", found, err)
	}
	if originalAfter.StartsTime != "13:00" {
		t.Fatal("la serie original NO debe cambiar sus campos propios, solo truncarse")
	}
	if originalAfter.EffectiveUntil == nil || *originalAfter.EffectiveUntil != "2027-05-31" {
		t.Fatalf("expected original.effectiveUntil = 2027-05-31 (effectiveDate - 1 día), got %+v", originalAfter.EffectiveUntil)
	}
}

// --- Proyección efectiva -----------------------------------------------

func TestListEffectiveManualBlocks_OnlyOverlappingRangeAndNotDeleted(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	inRange := createBlock(t, repo, shopO, barberO1, schedule.CreateBlockInput{
		BlockType: schedule.BlockTypeEmergency,
		StartsAt:  time.Date(2027, 3, 10, 10, 0, 0, 0, time.UTC),
		EndsAt:    time.Date(2027, 3, 10, 11, 0, 0, 0, time.UTC),
	})
	outOfRange := createBlock(t, repo, shopO, barberO1, schedule.CreateBlockInput{
		BlockType: schedule.BlockTypeEmergency,
		StartsAt:  time.Date(2027, 5, 10, 10, 0, 0, 0, time.UTC),
		EndsAt:    time.Date(2027, 5, 10, 11, 0, 0, 0, time.UTC),
	})
	deleted := createBlock(t, repo, shopO, barberO1, schedule.CreateBlockInput{
		BlockType: schedule.BlockTypeEmergency,
		StartsAt:  time.Date(2027, 3, 15, 10, 0, 0, 0, time.UTC),
		EndsAt:    time.Date(2027, 3, 15, 11, 0, 0, 0, time.UTC),
	})
	if _, err := repo.DeleteBlock(context.Background(), string(shopO), barberO1, deleted.ID, ownerO); err != nil {
		t.Fatalf("DeleteBlock: %v", err)
	}

	items, err := repo.ListEffectiveManualBlocks(context.Background(), string(shopO), barberO1,
		time.Date(2027, 3, 1, 0, 0, 0, 0, time.UTC), time.Date(2027, 4, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("ListEffectiveManualBlocks: %v", err)
	}
	got := map[string]bool{}
	for _, b := range items {
		got[b.ID] = true
	}
	if !got[inRange.ID] {
		t.Error("expected the in-range block to be present")
	}
	if got[outOfRange.ID] {
		t.Error("expected the out-of-range block to be absent")
	}
	if got[deleted.ID] {
		t.Error("expected the soft-deleted block to be absent (RN-BLQ-04)")
	}
}
