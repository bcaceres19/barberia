// Pruebas de integración de HU-040 contra el router REAL de producción
// (buildRouter) y PostgreSQL real, mismo patrón que
// barber_services_integration_test.go (HU-023). Requieren las trece
// migraciones aplicadas (incluida 20260825160000_create_working_hour.sql)
// y database/testdata/dos_barberias.sql +
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

type workingHourBody struct {
	ID              string `json:"id"`
	ISOWeekday      int    `json:"isoWeekday"`
	StartsTime      string `json:"startsTime"`
	DurationMinutes int    `json:"durationMinutes"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

type workingHourListBody struct {
	Items      []workingHourBody `json:"items"`
	NextCursor *string           `json:"nextCursor"`
}

func doListWorkingHoursRequest(router http.Handler, rawToken, barberID, query string) *httptest.ResponseRecorder {
	target := "/api/v1/private/barbers/" + barberID + "/working-hours"
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

func doGetWorkingHourRequest(router http.Handler, rawToken, barberID, workingHourID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/private/barbers/"+barberID+"/working-hours/"+workingHourID, nil)
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func doCreateWorkingHourRequest(router http.Handler, rawToken, barberID, idempotencyKey, rawBody string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/private/barbers/"+barberID+"/working-hours", bytes.NewReader([]byte(rawBody)))
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

func doUpdateWorkingHourRequest(router http.Handler, rawToken, barberID, workingHourID, rawBody string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/private/barbers/"+barberID+"/working-hours/"+workingHourID, bytes.NewReader([]byte(rawBody)))
	req.Header.Set("Content-Type", "application/json")
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func doDeleteWorkingHourRequest(router http.Handler, rawToken, barberID, workingHourID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/private/barbers/"+barberID+"/working-hours/"+workingHourID, nil)
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// TestWorkingHours_HTTP_CreateListGetUpdateDelete_FullJourney recorre crear
// (201), repetir la misma clave (200, mismo cuerpo, RN-IDE-01), listar
// (CA-040-01), consultar (200), editar (200) y retirar (204, luego 404 al
// repetir), a través del router real.
func TestWorkingHours_HTTP_CreateListGetUpdateDelete_FullJourney(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "schedule-journey-a")
	barber := createBarberViaRouter(t, router, raw, "Barbero Horario HTTP "+uniqueToken(t, "journey"))

	key := uniqueToken(t, "schedule-create")
	body := `{"isoWeekday":1,"startsTime":"08:00","durationMinutes":240}`
	first := doCreateWorkingHourRequest(router, raw, barber.ID, key, body)
	if first.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create, got %d: %s", first.Code, first.Body.String())
	}
	var created workingHourBody
	if err := json.Unmarshal(first.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if created.ISOWeekday != 1 || created.StartsTime != "08:00" || created.DurationMinutes != 240 {
		t.Fatalf("unexpected created working hour: %+v", created)
	}
	wantLoc := "/private/barbers/" + barber.ID + "/working-hours/" + created.ID
	if loc := first.Header().Get("Location"); loc != wantLoc {
		t.Fatalf("expected Location %q, got %q", wantLoc, loc)
	}

	// RN-IDE-01: repetir exactamente la misma clave y cuerpo reproduce la
	// misma respuesta (mismo status, mismo cuerpo), sin crear un segundo
	// tramo.
	replay := doCreateWorkingHourRequest(router, raw, barber.ID, key, body)
	if replay.Code != http.StatusCreated {
		t.Fatalf("expected 201 on replay, got %d: %s", replay.Code, replay.Body.String())
	}
	if replay.Body.String() != first.Body.String() {
		t.Fatalf("expected byte-identical replay body, got different bodies")
	}

	// CA-040-01: aparece en el listado del barbero.
	listRec := doListWorkingHoursRequest(router, raw, barber.ID, "limit=50")
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on list, got %d: %s", listRec.Code, listRec.Body.String())
	}
	var page workingHourListBody
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
		t.Fatal("CA-040-01: expected the created working hour to appear in the list")
	}

	// Lectura individual.
	getRec := doGetWorkingHourRequest(router, raw, barber.ID, created.ID)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on get, got %d: %s", getRec.Code, getRec.Body.String())
	}

	// Edición: reemplaza el intervalo completo.
	updateRec := doUpdateWorkingHourRequest(router, raw, barber.ID, created.ID, `{"isoWeekday":1,"startsTime":"09:00","durationMinutes":180}`)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on update, got %d: %s", updateRec.Code, updateRec.Body.String())
	}
	var updated workingHourBody
	if err := json.Unmarshal(updateRec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode update: %v", err)
	}
	if updated.StartsTime != "09:00" || updated.DurationMinutes != 180 {
		t.Fatalf("expected the updated interval to persist, got %+v", updated)
	}

	// Retiro y reintento: 204 y luego 404 (CA-040-05).
	deleteRec := doDeleteWorkingHourRequest(router, raw, barber.ID, created.ID)
	if deleteRec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 on delete, got %d: %s", deleteRec.Code, deleteRec.Body.String())
	}
	repeatDelete := doDeleteWorkingHourRequest(router, raw, barber.ID, created.ID)
	if repeatDelete.Code != http.StatusNotFound {
		t.Fatalf("expected 404 deleting an already-removed working hour, got %d: %s", repeatDelete.Code, repeatDelete.Body.String())
	}
}

// TestWorkingHours_HTTP_OverlappingCreate_Returns409 cubre CA-040-04 a
// través del router real: un segundo tramo que se solapa con uno existente
// del mismo barbero y día se rechaza con 409, sin persistir nada.
func TestWorkingHours_HTTP_OverlappingCreate_Returns409(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "schedule-overlap-a")
	barber := createBarberViaRouter(t, router, raw, "Barbero Solape HTTP "+uniqueToken(t, "overlap"))

	first := doCreateWorkingHourRequest(router, raw, barber.ID, uniqueToken(t, "overlap-1"), `{"isoWeekday":2,"startsTime":"08:00","durationMinutes":120}`)
	if first.Code != http.StatusCreated {
		t.Fatalf("expected 201 creating the first segment, got %d: %s", first.Code, first.Body.String())
	}

	rec := doCreateWorkingHourRequest(router, raw, barber.ID, uniqueToken(t, "overlap-2"), `{"isoWeekday":2,"startsTime":"09:00","durationMinutes":120}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("CA-040-04: expected 409 on an overlapping segment, got %d: %s", rec.Code, rec.Body.String())
	}
	var problem map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
		t.Fatalf("decode conflict problem: %v", err)
	}
	if problem["code"] != "conflict" {
		t.Fatalf("expected code=conflict, got %v", problem["code"])
	}
}

// TestWorkingHours_HTTP_InvalidField_Returns422 cubre CA-040-04: día, hora
// o duración inválidos responden 422 sin persistir nada.
func TestWorkingHours_HTTP_InvalidField_Returns422(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "schedule-invalid-a")
	barber := createBarberViaRouter(t, router, raw, "Barbero Inválido HTTP "+uniqueToken(t, "invalid"))

	cases := []string{
		`{"isoWeekday":0,"startsTime":"08:00","durationMinutes":60}`,
		`{"isoWeekday":1,"startsTime":"8am","durationMinutes":60}`,
		`{"isoWeekday":1,"startsTime":"08:00","durationMinutes":0}`,
	}
	for _, body := range cases {
		rec := doCreateWorkingHourRequest(router, raw, barber.ID, uniqueToken(t, "invalid"), body)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422 for body %s, got %d: %s", body, rec.Code, rec.Body.String())
		}
	}
}

// TestWorkingHours_HTTP_CrossTenantBarber_Returns404 cubre RN-TEN-01: un
// barberId real de la barbería B responde 404 bajo el contexto de A, tanto
// para crear como para listar.
func TestWorkingHours_HTTP_CrossTenantBarber_Returns404(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	rawA := createSessionCookie(t, db, shopA, staffUserActiveA, "schedule-cross-a")
	rawB := createSessionCookie(t, db, shopB, staffUserActiveB, "schedule-cross-b")
	barberOfB := createBarberViaRouter(t, router, rawB, "Barbero de B HTTP "+uniqueToken(t, "cross"))

	rec := doListWorkingHoursRequest(router, rawA, barberOfB.ID, "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 listing another shop's barber, got %d: %s", rec.Code, rec.Body.String())
	}

	createRec := doCreateWorkingHourRequest(router, rawA, barberOfB.ID, uniqueToken(t, "cross-create"), `{"isoWeekday":1,"startsTime":"08:00","durationMinutes":60}`)
	if createRec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 creating under another shop's barber, got %d: %s", createRec.Code, createRec.Body.String())
	}
}

// TestWorkingHours_HTTP_Unauthenticated_Returns401 confirma que las cinco
// operaciones exigen sesión (CA-006-04 ya lo prueba de forma estructural
// para TODA ruta privada; esta prueba deja evidencia propia de HU-040).
func TestWorkingHours_HTTP_Unauthenticated_Returns401(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	const anyID = "00000000-0000-0000-0000-000000000001"

	if rec := doListWorkingHoursRequest(router, "", anyID, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on list, got %d", rec.Code)
	}
	if rec := doCreateWorkingHourRequest(router, "", anyID, "k", `{"isoWeekday":1,"startsTime":"08:00","durationMinutes":60}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on create, got %d", rec.Code)
	}
	if rec := doGetWorkingHourRequest(router, "", anyID, anyID); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on get, got %d", rec.Code)
	}
	if rec := doUpdateWorkingHourRequest(router, "", anyID, anyID, `{"isoWeekday":1,"startsTime":"08:00","durationMinutes":60}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on update, got %d", rec.Code)
	}
	if rec := doDeleteWorkingHourRequest(router, "", anyID, anyID); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on delete, got %d", rec.Code)
	}
}
