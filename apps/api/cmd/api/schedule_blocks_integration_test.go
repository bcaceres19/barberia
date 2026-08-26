// Pruebas de integración de HU-042 (bloqueos de agenda) contra el router
// REAL de producción (buildRouter) y PostgreSQL real, mismo patrón que
// schedule_integration_test.go (HU-040)/schedule_exceptions_integration_test.go
// (HU-041). Requieren las quince migraciones aplicadas (incluida
// 20260826100000_create_time_block.sql) y database/testdata/dos_barberias.sql
// + testdata/hu005_credenciales_sesiones.sql cargados.
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./cmd/api/...
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type timeBlockBody struct {
	ID        string  `json:"id"`
	BlockType string  `json:"blockType"`
	Source    string  `json:"source"`
	StartsAt  string  `json:"startsAt"`
	EndsAt    string  `json:"endsAt"`
	Reason    *string `json:"reason"`
	DeletedAt *string `json:"deletedAt"`
	DeletedBy *string `json:"deletedBy"`
}

type timeBlockListBody struct {
	Items      []timeBlockBody `json:"items"`
	NextCursor *string         `json:"nextCursor"`
}

type seriesDateBody struct {
	BlockDate string `json:"blockDate"`
}

type seriesExceptionBody struct {
	ExcludedDate string `json:"excludedDate"`
}

type timeBlockSeriesBody struct {
	ID              string                `json:"id"`
	BlockType       string                `json:"blockType"`
	RecurrenceKind  string                `json:"recurrenceKind"`
	ISOWeekday      *int                  `json:"isoWeekday"`
	StartsTime      string                `json:"startsTime"`
	DurationMinutes int                   `json:"durationMinutes"`
	EffectiveFrom   string                `json:"effectiveFrom"`
	EffectiveUntil  *string               `json:"effectiveUntil"`
	Dates           []seriesDateBody      `json:"dates"`
	Exceptions      []seriesExceptionBody `json:"exceptions"`
}

func doJSONRequest(router http.Handler, method, rawToken, path, idempotencyKey, rawBody string) *httptest.ResponseRecorder {
	var body *bytes.Reader
	if rawBody != "" {
		body = bytes.NewReader([]byte(rawBody))
	} else {
		body = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, "/api/v1/private"+path, body)
	if rawBody != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// TestTimeBlocks_HTTP_CreateListGetDelete_FullJourney recorre crear (201,
// RN-IDE-01), listar (CA-042), consultar (200) y retirar (204, luego 404 al
// repetir: RN-BLQ-04 nunca borra físico), a través del router real.
func TestTimeBlocks_HTTP_CreateListGetDelete_FullJourney(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "blocks-journey-a")
	barber := createBarberViaRouter(t, router, raw, "Barbero Bloqueos HTTP "+uniqueToken(t, "journey"))

	key := uniqueToken(t, "block-create")
	body := `{"blockType":"emergency","startsAt":"2027-07-20T15:00:00-05:00","endsAt":"2027-07-20T19:00:00-05:00","reason":"Urgencia médica"}`
	first := doJSONRequest(router, http.MethodPost, raw, "/barbers/"+barber.ID+"/time-blocks", key, body)
	if first.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create, got %d: %s", first.Code, first.Body.String())
	}
	var created timeBlockBody
	if err := json.Unmarshal(first.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if created.BlockType != "emergency" || created.Source != "manual" {
		t.Fatalf("unexpected created block: %+v", created)
	}
	wantLoc := "/private/barbers/" + barber.ID + "/time-blocks/" + created.ID
	if loc := first.Header().Get("Location"); loc != wantLoc {
		t.Fatalf("expected Location %q, got %q", wantLoc, loc)
	}

	// RN-IDE-01: repetir exactamente la misma clave y cuerpo reproduce la
	// misma respuesta, sin crear un segundo bloqueo.
	replay := doJSONRequest(router, http.MethodPost, raw, "/barbers/"+barber.ID+"/time-blocks", key, body)
	if replay.Code != http.StatusCreated || replay.Body.String() != first.Body.String() {
		t.Fatalf("expected byte-identical replay, got %d: %s", replay.Code, replay.Body.String())
	}

	list := doJSONRequest(router, http.MethodGet, raw, "/barbers/"+barber.ID+"/time-blocks", "", "")
	if list.Code != http.StatusOK {
		t.Fatalf("expected 200 on list, got %d: %s", list.Code, list.Body.String())
	}
	var listed timeBlockListBody
	if err := json.Unmarshal(list.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	found := false
	for _, item := range listed.Items {
		if item.ID == created.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected the created block in the list, got %+v", listed.Items)
	}

	get := doJSONRequest(router, http.MethodGet, raw, "/barbers/"+barber.ID+"/time-blocks/"+created.ID, "", "")
	if get.Code != http.StatusOK {
		t.Fatalf("expected 200 on get, got %d: %s", get.Code, get.Body.String())
	}

	del := doJSONRequest(router, http.MethodDelete, raw, "/barbers/"+barber.ID+"/time-blocks/"+created.ID, "", "")
	if del.Code != http.StatusNoContent {
		t.Fatalf("expected 204 on delete, got %d: %s", del.Code, del.Body.String())
	}

	delAgain := doJSONRequest(router, http.MethodDelete, raw, "/barbers/"+barber.ID+"/time-blocks/"+created.ID, "", "")
	if delAgain.Code != http.StatusNotFound {
		t.Fatalf("expected 404 retrying delete (RN-BLQ-04, no physical delete), got %d: %s", delAgain.Code, delAgain.Body.String())
	}
}

// TestTimeBlocks_HTTP_CrossTenant_NeverLeaksAnotherShopsBlock verifica
// RN-TEN-01 a través del router real.
func TestTimeBlocks_HTTP_CrossTenant_NeverLeaksAnotherShopsBlock(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	rawA := createSessionCookie(t, db, shopA, staffUserActiveA, "blocks-tenant-a")
	barberA := createBarberViaRouter(t, router, rawA, "Barbero Bloqueos Tenant A "+uniqueToken(t, "cross"))

	key := uniqueToken(t, "block-cross")
	body := `{"blockType":"vacation","startsAt":"2027-08-01T00:00:00-05:00","endsAt":"2027-08-15T00:00:00-05:00"}`
	created := doJSONRequest(router, http.MethodPost, rawA, "/barbers/"+barberA.ID+"/time-blocks", key, body)
	if created.Code != http.StatusCreated {
		t.Fatalf("setup create: expected 201, got %d: %s", created.Code, created.Body.String())
	}
	var block timeBlockBody
	_ = json.Unmarshal(created.Body.Bytes(), &block)

	rawB := createSessionCookie(t, db, shopB, staffUserActiveB, "blocks-tenant-b")
	crossGet := doJSONRequest(router, http.MethodGet, rawB, "/barbers/"+barberA.ID+"/time-blocks/"+block.ID, "", "")
	if crossGet.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for a cross-tenant read, got %d: %s", crossGet.Code, crossGet.Body.String())
	}
}

// TestTimeBlockSeries_HTTP_WeeklyCreateUpdateSplitDelete_FullJourney recorre
// crear una serie weekly (201), editarla completa (200, scope whole),
// dividirla desde una fecha (200, scope this_and_following, RN-BLQ-01) y
// retirarla (204), a través del router real.
func TestTimeBlockSeries_HTTP_WeeklyCreateUpdateSplitDelete_FullJourney(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "series-journey-a")
	barber := createBarberViaRouter(t, router, raw, "Barbero Series HTTP "+uniqueToken(t, "journey"))

	key := uniqueToken(t, "series-create")
	body := `{"blockType":"lunch","recurrenceKind":"weekly","isoWeekday":1,"startsTime":"13:00","durationMinutes":60,"effectiveFrom":"2027-01-04"}`
	created := doJSONRequest(router, http.MethodPost, raw, "/barbers/"+barber.ID+"/time-block-series", key, body)
	if created.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create, got %d: %s", created.Code, created.Body.String())
	}
	var series timeBlockSeriesBody
	if err := json.Unmarshal(created.Body.Bytes(), &series); err != nil {
		t.Fatalf("decode: %v", err)
	}

	wholeBody := `{"scope":"whole","blockType":"lunch","startsTime":"13:30","durationMinutes":45,"effectiveFrom":"2027-01-04"}`
	updated := doJSONRequest(router, http.MethodPatch, raw, "/barbers/"+barber.ID+"/time-block-series/"+series.ID, "", wholeBody)
	if updated.Code != http.StatusOK {
		t.Fatalf("expected 200 on whole update, got %d: %s", updated.Code, updated.Body.String())
	}
	var wholeUpdated timeBlockSeriesBody
	_ = json.Unmarshal(updated.Body.Bytes(), &wholeUpdated)
	if wholeUpdated.StartsTime != "13:30" || wholeUpdated.DurationMinutes != 45 {
		t.Fatalf("unexpected series after whole update: %+v", wholeUpdated)
	}

	splitBody := `{"scope":"this_and_following","effectiveDate":"2027-06-01","blockType":"lunch","startsTime":"14:00","durationMinutes":30,"effectiveFrom":"2027-06-01"}`
	split := doJSONRequest(router, http.MethodPatch, raw, "/barbers/"+barber.ID+"/time-block-series/"+series.ID, "", splitBody)
	if split.Code != http.StatusOK {
		t.Fatalf("expected 200 on split update, got %d: %s", split.Code, split.Body.String())
	}
	var newSeries timeBlockSeriesBody
	_ = json.Unmarshal(split.Body.Bytes(), &newSeries)
	if newSeries.ID == series.ID {
		t.Fatal("this_and_following debe devolver una serie NUEVA distinta de la original")
	}
	if newSeries.StartsTime != "14:00" || newSeries.EffectiveFrom != "2027-06-01" {
		t.Fatalf("unexpected new series after split: %+v", newSeries)
	}

	originalAfter := doJSONRequest(router, http.MethodGet, raw, "/barbers/"+barber.ID+"/time-block-series/"+series.ID, "", "")
	var originalAfterBody timeBlockSeriesBody
	_ = json.Unmarshal(originalAfter.Body.Bytes(), &originalAfterBody)
	if originalAfterBody.EffectiveUntil == nil || *originalAfterBody.EffectiveUntil != "2027-05-31" {
		t.Fatalf("expected original series truncated to 2027-05-31, got %+v", originalAfterBody.EffectiveUntil)
	}

	delOriginal := doJSONRequest(router, http.MethodDelete, raw, "/barbers/"+barber.ID+"/time-block-series/"+series.ID, "", "")
	if delOriginal.Code != http.StatusNoContent {
		t.Fatalf("expected 204 deleting the original series, got %d: %s", delOriginal.Code, delOriginal.Body.String())
	}
	delNew := doJSONRequest(router, http.MethodDelete, raw, "/barbers/"+barber.ID+"/time-block-series/"+newSeries.ID, "", "")
	if delNew.Code != http.StatusNoContent {
		t.Fatalf("expected 204 deleting the split series, got %d: %s", delNew.Code, delNew.Body.String())
	}
}

// TestTimeBlockSeries_HTTP_DateList_AddRemoveDatesAndExceptions recorre una
// serie date_list: alta con fechas explícitas en una sola operación
// (RN-BLQ-01), agregar/retirar una fecha y agregar/retirar una excepción
// ("esta instancia no"), a través del router real.
func TestTimeBlockSeries_HTTP_DateList_AddRemoveDatesAndExceptions(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "series-datelist-a")
	barber := createBarberViaRouter(t, router, raw, "Barbero Series DateList HTTP "+uniqueToken(t, "journey"))

	key := uniqueToken(t, "series-datelist-create")
	body := `{"blockType":"vacation","recurrenceKind":"date_list","startsTime":"00:00","durationMinutes":1440,"effectiveFrom":"2027-12-01","effectiveUntil":"2027-12-31","explicitDates":["2027-12-15","2027-12-16"]}`
	created := doJSONRequest(router, http.MethodPost, raw, "/barbers/"+barber.ID+"/time-block-series", key, body)
	if created.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create, got %d: %s", created.Code, created.Body.String())
	}
	var series timeBlockSeriesBody
	_ = json.Unmarshal(created.Body.Bytes(), &series)
	if len(series.Dates) != 2 {
		t.Fatalf("expected 2 explicit dates created in one operation, got %+v", series.Dates)
	}

	addDate := doJSONRequest(router, http.MethodPost, raw, "/barbers/"+barber.ID+"/time-block-series/"+series.ID+"/dates", "", `{"blockDate":"2027-12-20"}`)
	if addDate.Code != http.StatusNoContent {
		t.Fatalf("expected 204 adding a date, got %d: %s", addDate.Code, addDate.Body.String())
	}
	removeDate := doJSONRequest(router, http.MethodDelete, raw, "/barbers/"+barber.ID+"/time-block-series/"+series.ID+"/dates/2027-12-15", "", "")
	if removeDate.Code != http.StatusNoContent {
		t.Fatalf("expected 204 removing a date, got %d: %s", removeDate.Code, removeDate.Body.String())
	}

	addException := doJSONRequest(router, http.MethodPost, raw, "/barbers/"+barber.ID+"/time-block-series/"+series.ID+"/exceptions", "", `{"excludedDate":"2027-12-16"}`)
	if addException.Code != http.StatusNoContent {
		t.Fatalf("expected 204 adding an exception, got %d: %s", addException.Code, addException.Body.String())
	}

	fetched := doJSONRequest(router, http.MethodGet, raw, "/barbers/"+barber.ID+"/time-block-series/"+series.ID, "", "")
	var fetchedBody timeBlockSeriesBody
	_ = json.Unmarshal(fetched.Body.Bytes(), &fetchedBody)
	if len(fetchedBody.Dates) != 2 { // 2027-12-16 y 2027-12-20 (2027-12-15 se retiró)
		t.Fatalf("unexpected dates after add/remove: %+v", fetchedBody.Dates)
	}
	if len(fetchedBody.Exceptions) != 1 || fetchedBody.Exceptions[0].ExcludedDate != "2027-12-16" {
		t.Fatalf("unexpected exceptions: %+v", fetchedBody.Exceptions)
	}

	removeException := doJSONRequest(router, http.MethodDelete, raw, "/barbers/"+barber.ID+"/time-block-series/"+series.ID+"/exceptions/2027-12-16", "", "")
	if removeException.Code != http.StatusNoContent {
		t.Fatalf("expected 204 removing an exception, got %d: %s", removeException.Code, removeException.Body.String())
	}
}
