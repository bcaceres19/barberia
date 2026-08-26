// Pruebas de integración de HU-041 contra el router REAL de producción
// (buildRouter) y PostgreSQL real, mismo patrón que
// schedule_integration_test.go (HU-040). Requieren las catorce migraciones
// aplicadas (incluida 20260826090000_create_working_hour_override.sql) y
// database/testdata/dos_barberias.sql +
// testdata/hu005_credenciales_sesiones.sql cargados.
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

type holidayCalendarBody struct {
	Enabled bool `json:"enabled"`
}

type scheduleExceptionSegmentBody struct {
	ID              string `json:"id"`
	StartsTime      string `json:"startsTime"`
	DurationMinutes int    `json:"durationMinutes"`
}

type scheduleExceptionBody struct {
	ID            string                         `json:"id"`
	EffectiveDate string                         `json:"effectiveDate"`
	IsClosed      bool                           `json:"isClosed"`
	Reason        *string                        `json:"reason"`
	Segments      []scheduleExceptionSegmentBody `json:"segments"`
	CreatedAt     string                         `json:"createdAt"`
	UpdatedAt     string                         `json:"updatedAt"`
}

type scheduleExceptionListBody struct {
	Items      []scheduleExceptionBody `json:"items"`
	NextCursor *string                 `json:"nextCursor"`
}

type colombianHolidayBody struct {
	Date string `json:"date"`
	Name string `json:"name"`
}

type colombianHolidayListBody struct {
	Items []colombianHolidayBody `json:"items"`
}

func doGetHolidayCalendarRequest(router http.Handler, rawToken, barberID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/private/barbers/"+barberID+"/holiday-calendar", nil)
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func doUpdateHolidayCalendarRequest(router http.Handler, rawToken, barberID, rawBody string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/private/barbers/"+barberID+"/holiday-calendar", bytes.NewReader([]byte(rawBody)))
	req.Header.Set("Content-Type", "application/json")
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func doListScheduleExceptionsRequest(router http.Handler, rawToken, barberID, query string) *httptest.ResponseRecorder {
	target := "/api/v1/private/barbers/" + barberID + "/schedule-exceptions"
	if query != "" {
		target += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, target, nil)
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func doGetScheduleExceptionRequest(router http.Handler, rawToken, barberID, exceptionID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/private/barbers/"+barberID+"/schedule-exceptions/"+exceptionID, nil)
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func doCreateScheduleExceptionRequest(router http.Handler, rawToken, barberID, idempotencyKey, rawBody string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/private/barbers/"+barberID+"/schedule-exceptions", bytes.NewReader([]byte(rawBody)))
	req.Header.Set("Content-Type", "application/json")
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

func doUpdateScheduleExceptionRequest(router http.Handler, rawToken, barberID, exceptionID, rawBody string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/private/barbers/"+barberID+"/schedule-exceptions/"+exceptionID, bytes.NewReader([]byte(rawBody)))
	req.Header.Set("Content-Type", "application/json")
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func doDeleteScheduleExceptionRequest(router http.Handler, rawToken, barberID, exceptionID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/private/barbers/"+barberID+"/schedule-exceptions/"+exceptionID, nil)
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func doListColombianHolidaysRequest(router http.Handler, rawToken, query string) *httptest.ResponseRecorder {
	target := "/api/v1/private/schedule/colombian-holidays"
	if query != "" {
		target += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, target, nil)
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// TestScheduleExceptions_HTTP_CreateListGetUpdateDelete_FullJourney recorre
// crear (201), repetir la misma clave (201, mismo cuerpo, RN-IDE-01),
// listar (CA-041-04/05), consultar (200), editar (200) y retirar (204,
// luego 404 al repetir), a través del router real.
func TestScheduleExceptions_HTTP_CreateListGetUpdateDelete_FullJourney(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "exception-journey-a")
	barber := createBarberViaRouter(t, router, raw, "Barbero Excepción HTTP "+uniqueToken(t, "journey"))

	key := uniqueToken(t, "exception-create")
	body := `{"effectiveDate":"2027-11-15","isClosed":false,"reason":"Turno especial","segments":[{"startsTime":"09:00","durationMinutes":180}]}`
	first := doCreateScheduleExceptionRequest(router, raw, barber.ID, key, body)
	if first.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create, got %d: %s", first.Code, first.Body.String())
	}
	var created scheduleExceptionBody
	if err := json.Unmarshal(first.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if created.EffectiveDate != "2027-11-15" || created.IsClosed || len(created.Segments) != 1 {
		t.Fatalf("unexpected created exception: %+v", created)
	}
	if created.Reason == nil || *created.Reason != "Turno especial" {
		t.Fatalf("expected reason to persist, got %+v", created.Reason)
	}
	wantLoc := "/private/barbers/" + barber.ID + "/schedule-exceptions/" + created.ID
	if loc := first.Header().Get("Location"); loc != wantLoc {
		t.Fatalf("expected Location %q, got %q", wantLoc, loc)
	}

	// RN-IDE-01: repetir exactamente la misma clave y cuerpo reproduce la
	// misma respuesta, sin crear una segunda excepción.
	replay := doCreateScheduleExceptionRequest(router, raw, barber.ID, key, body)
	if replay.Code != http.StatusCreated {
		t.Fatalf("expected 201 on replay, got %d: %s", replay.Code, replay.Body.String())
	}
	if replay.Body.String() != first.Body.String() {
		t.Fatalf("expected byte-identical replay body, got different bodies")
	}

	// CA-041-04: aparece en el listado del barbero.
	listRec := doListScheduleExceptionsRequest(router, raw, barber.ID, "limit=50")
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on list, got %d: %s", listRec.Code, listRec.Body.String())
	}
	var page scheduleExceptionListBody
	if err := json.Unmarshal(listRec.Body.Bytes(), &page); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	found := false
	for _, item := range page.Items {
		if item.ID == created.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("CA-041-04: expected the created exception to appear in the list")
	}

	// Lectura individual.
	getRec := doGetScheduleExceptionRequest(router, raw, barber.ID, created.ID)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on get, got %d: %s", getRec.Code, getRec.Body.String())
	}

	// Edición: reemplaza la excepción completa, ahora cerrada.
	updateRec := doUpdateScheduleExceptionRequest(router, raw, barber.ID, created.ID, `{"effectiveDate":"2027-11-16","isClosed":true}`)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on update, got %d: %s", updateRec.Code, updateRec.Body.String())
	}
	var updated scheduleExceptionBody
	if err := json.Unmarshal(updateRec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode update: %v", err)
	}
	if updated.EffectiveDate != "2027-11-16" || !updated.IsClosed || len(updated.Segments) != 0 {
		t.Fatalf("expected the exception to become closed on the new date, got %+v", updated)
	}

	// Retiro y reintento: 204 y luego 404 (CA-041-06).
	deleteRec := doDeleteScheduleExceptionRequest(router, raw, barber.ID, created.ID)
	if deleteRec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 on delete, got %d: %s", deleteRec.Code, deleteRec.Body.String())
	}
	repeatDelete := doDeleteScheduleExceptionRequest(router, raw, barber.ID, created.ID)
	if repeatDelete.Code != http.StatusNotFound {
		t.Fatalf("expected 404 deleting an already-removed exception, got %d: %s", repeatDelete.Code, repeatDelete.Body.String())
	}
}

// TestScheduleExceptions_HTTP_DuplicateDate_Returns409 cubre CA-041-05: una
// segunda excepción del mismo barbero para la misma fecha se rechaza con
// 409, sin persistir nada.
func TestScheduleExceptions_HTTP_DuplicateDate_Returns409(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "exception-dupdate-a")
	barber := createBarberViaRouter(t, router, raw, "Barbero Fecha Duplicada HTTP "+uniqueToken(t, "dupdate"))

	first := doCreateScheduleExceptionRequest(router, raw, barber.ID, uniqueToken(t, "dupdate-1"), `{"effectiveDate":"2027-12-24","isClosed":true}`)
	if first.Code != http.StatusCreated {
		t.Fatalf("expected 201 creating the first exception, got %d: %s", first.Code, first.Body.String())
	}

	rec := doCreateScheduleExceptionRequest(router, raw, barber.ID, uniqueToken(t, "dupdate-2"), `{"effectiveDate":"2027-12-24","isClosed":true}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("CA-041-05: expected 409 for a duplicate effective date, got %d: %s", rec.Code, rec.Body.String())
	}
	var problem map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
		t.Fatalf("decode conflict problem: %v", err)
	}
	if problem["code"] != "conflict" {
		t.Fatalf("expected code=conflict, got %v", problem["code"])
	}
}

// TestScheduleExceptions_HTTP_InvalidShape_Returns422 cubre CA-041-04:
// fecha inválida, día cerrado con tramos, día abierto sin tramos y tramos
// solapados responden 422 sin persistir nada.
func TestScheduleExceptions_HTTP_InvalidShape_Returns422(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "exception-invalid-a")
	barber := createBarberViaRouter(t, router, raw, "Barbero Inválido HTTP "+uniqueToken(t, "invalid"))

	cases := []string{
		`{"effectiveDate":"2027-13-01","isClosed":true}`,
		`{"effectiveDate":"2027-11-20","isClosed":true,"segments":[{"startsTime":"08:00","durationMinutes":60}]}`,
		`{"effectiveDate":"2027-11-21","isClosed":false}`,
		`{"effectiveDate":"2027-11-22","isClosed":false,"segments":[{"startsTime":"08:00","durationMinutes":120},{"startsTime":"09:00","durationMinutes":60}]}`,
	}
	for _, body := range cases {
		rec := doCreateScheduleExceptionRequest(router, raw, barber.ID, uniqueToken(t, "invalid"), body)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422 for body %s, got %d: %s", body, rec.Code, rec.Body.String())
		}
	}
}

// TestScheduleExceptions_HTTP_CrossTenantBarber_Returns404 cubre RN-TEN-01:
// un barberId real de la barbería B responde 404 bajo el contexto de A.
func TestScheduleExceptions_HTTP_CrossTenantBarber_Returns404(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	rawA := createSessionCookie(t, db, shopA, staffUserActiveA, "exception-cross-a")
	rawB := createSessionCookie(t, db, shopB, staffUserActiveB, "exception-cross-b")
	barberOfB := createBarberViaRouter(t, router, rawB, "Barbero de B HTTP "+uniqueToken(t, "cross"))

	if rec := doListScheduleExceptionsRequest(router, rawA, barberOfB.ID, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 listing another shop's barber, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec := doCreateScheduleExceptionRequest(router, rawA, barberOfB.ID, uniqueToken(t, "cross-create"), `{"effectiveDate":"2027-11-23","isClosed":true}`); rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 creating under another shop's barber, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec := doGetHolidayCalendarRequest(router, rawA, barberOfB.ID); rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 reading another shop's barber holiday calendar, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestHolidayCalendar_HTTP_GetThenUpdate_PersistsToggle recorre GET
// (por defecto desactivado), PATCH a activado y GET de nuevo (CA-041-01/02).
func TestHolidayCalendar_HTTP_GetThenUpdate_PersistsToggle(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "holiday-toggle-a")
	barber := createBarberViaRouter(t, router, raw, "Barbero Festivos HTTP "+uniqueToken(t, "toggle"))

	getRec := doGetHolidayCalendarRequest(router, raw, barber.ID)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on get, got %d: %s", getRec.Code, getRec.Body.String())
	}
	var initial holidayCalendarBody
	if err := json.Unmarshal(getRec.Body.Bytes(), &initial); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if initial.Enabled {
		t.Fatal("expected the holiday calendar to be disabled by default for a new barber")
	}

	updateRec := doUpdateHolidayCalendarRequest(router, raw, barber.ID, `{"enabled":true}`)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on update, got %d: %s", updateRec.Code, updateRec.Body.String())
	}
	var updated holidayCalendarBody
	if err := json.Unmarshal(updateRec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode update: %v", err)
	}
	if !updated.Enabled {
		t.Fatal("expected the holiday calendar to become enabled")
	}

	confirmRec := doGetHolidayCalendarRequest(router, raw, barber.ID)
	var confirmed holidayCalendarBody
	if err := json.Unmarshal(confirmRec.Body.Bytes(), &confirmed); err != nil {
		t.Fatalf("decode confirm: %v", err)
	}
	if !confirmed.Enabled {
		t.Fatal("expected the toggle to persist across requests")
	}
}

// TestColombianHolidays_HTTP_KnownYear_ContainsIndependenceDay confirma que
// la consulta de referencia (sin barbero ni excepción de por medio) trae el
// 20 de julio como "Día de la Independencia" para 2026.
func TestColombianHolidays_HTTP_KnownYear_ContainsIndependenceDay(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "holidays-list-a")

	rec := doListColombianHolidaysRequest(router, raw, "year=2026")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var list colombianHolidayListBody
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	found := false
	for _, h := range list.Items {
		if h.Date == "2026-07-20" && h.Name == "Día de la Independencia" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected 2026-07-20 Día de la Independencia among the holidays, got %+v", list.Items)
	}
}

// TestColombianHolidays_HTTP_MissingYear_Returns400 cubre el parámetro year
// obligatorio.
func TestColombianHolidays_HTTP_MissingYear_Returns400(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "holidays-missing-year-a")

	rec := doListColombianHolidaysRequest(router, raw, "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for a missing year, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestScheduleExceptions_HTTP_Unauthenticated_Returns401 confirma que las
// operaciones de HU-041 exigen sesión.
func TestScheduleExceptions_HTTP_Unauthenticated_Returns401(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	const anyID = "00000000-0000-0000-0000-000000000001"

	if rec := doGetHolidayCalendarRequest(router, "", anyID); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on get holiday calendar, got %d", rec.Code)
	}
	if rec := doUpdateHolidayCalendarRequest(router, "", anyID, `{"enabled":true}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on update holiday calendar, got %d", rec.Code)
	}
	if rec := doListScheduleExceptionsRequest(router, "", anyID, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on list exceptions, got %d", rec.Code)
	}
	if rec := doCreateScheduleExceptionRequest(router, "", anyID, "k", `{"effectiveDate":"2027-11-24","isClosed":true}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on create exception, got %d", rec.Code)
	}
	if rec := doGetScheduleExceptionRequest(router, "", anyID, anyID); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on get exception, got %d", rec.Code)
	}
	if rec := doUpdateScheduleExceptionRequest(router, "", anyID, anyID, `{"effectiveDate":"2027-11-24","isClosed":true}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on update exception, got %d", rec.Code)
	}
	if rec := doDeleteScheduleExceptionRequest(router, "", anyID, anyID); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on delete exception, got %d", rec.Code)
	}
	if rec := doListColombianHolidaysRequest(router, "", "year=2026"); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on list colombian holidays, got %d", rec.Code)
	}
}
