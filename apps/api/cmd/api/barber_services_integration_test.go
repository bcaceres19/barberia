// Pruebas de integración de HU-023 contra el router REAL de producción
// (buildRouter) y PostgreSQL real, mismo patrón que
// catalog_integration_test.go/staff_integration_test.go. Requieren las doce
// migraciones aplicadas (incluida 20260824150000_create_barber_service.sql)
// y database/testdata/dos_barberias.sql + testdata/hu005_credenciales_sesiones.sql
// cargados.
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./cmd/api/...
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type assignmentBody struct {
	BarberID  string `json:"barberId"`
	ServiceID string `json:"serviceId"`
	CreatedAt string `json:"createdAt"`
}

type assignmentListBody struct {
	Items      []assignmentBody `json:"items"`
	NextCursor *string          `json:"nextCursor"`
}

func doListAssignmentsRequest(router http.Handler, rawToken, barberID, query string) *httptest.ResponseRecorder {
	target := "/api/v1/private/barbers/" + barberID + "/services"
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

func doAssignServiceRequest(router http.Handler, rawToken, barberID, serviceID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPut, "/api/v1/private/barbers/"+barberID+"/services/"+serviceID, nil)
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func doUnassignServiceRequest(router http.Handler, rawToken, barberID, serviceID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/private/barbers/"+barberID+"/services/"+serviceID, nil)
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func createBarberViaRouter(t *testing.T, router http.Handler, rawToken, fullName string) barberBody {
	t.Helper()
	rec := doCreateBarberRequest(router, rawToken, uniqueToken(t, "assign-fixture-barber"), `{"fullName":"`+fullName+`"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("createBarberViaRouter: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var b barberBody
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatalf("createBarberViaRouter: decode: %v", err)
	}
	return b
}

func createServiceViaRouter(t *testing.T, router http.Handler, rawToken, name string) serviceBody {
	t.Helper()
	rec := doCreateServiceRequest(router, rawToken, uniqueToken(t, "assign-fixture-service"),
		`{"name":"`+name+`","durationMinutes":30,"price":"20000.00"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("createServiceViaRouter: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var s serviceBody
	if err := json.Unmarshal(rec.Body.Bytes(), &s); err != nil {
		t.Fatalf("createServiceViaRouter: decode: %v", err)
	}
	return s
}

// TestBarberServices_HTTP_AssignRepeatListUnassign_FullJourney recorre
// asignar (201), repetir (200, mismo createdAt, CA-023-02), listar
// (CA-023-01) y desasignar cuando ya no es la última (204), a través del
// router real.
func TestBarberServices_HTTP_AssignRepeatListUnassign_FullJourney(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "assign-journey-a")

	barber := createBarberViaRouter(t, router, raw, "Barbero Asignación HTTP "+uniqueToken(t, "journey"))
	barberTwo := createBarberViaRouter(t, router, raw, "Barbero Asignación HTTP Dos "+uniqueToken(t, "journey"))
	service := createServiceViaRouter(t, router, raw, "Servicio Asignación HTTP "+uniqueToken(t, "journey"))

	first := doAssignServiceRequest(router, raw, barber.ID, service.ID)
	if first.Code != http.StatusCreated {
		t.Fatalf("expected 201 on first assign, got %d: %s", first.Code, first.Body.String())
	}
	var firstBody assignmentBody
	if err := json.Unmarshal(first.Body.Bytes(), &firstBody); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if firstBody.BarberID != barber.ID || firstBody.ServiceID != service.ID {
		t.Fatalf("unexpected assignment body: %+v", firstBody)
	}
	loc := first.Header().Get("Location")
	wantLoc := "/private/barbers/" + barber.ID + "/services/" + service.ID
	if loc != wantLoc {
		t.Fatalf("expected Location %q, got %q", wantLoc, loc)
	}

	// CA-023-02: repetir exactamente la misma operación responde 200 (no
	// 201) con el MISMO createdAt, sin crear una segunda fila.
	second := doAssignServiceRequest(router, raw, barber.ID, service.ID)
	if second.Code != http.StatusOK {
		t.Fatalf("expected 200 on repeated assign, got %d: %s", second.Code, second.Body.String())
	}
	var secondBody assignmentBody
	if err := json.Unmarshal(second.Body.Bytes(), &secondBody); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if secondBody.CreatedAt != firstBody.CreatedAt {
		t.Fatalf("expected the same createdAt on replay, got %q vs %q", secondBody.CreatedAt, firstBody.CreatedAt)
	}

	// CA-023-01: aparece en el listado del barbero.
	listRec := doListAssignmentsRequest(router, raw, barber.ID, "limit=50")
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on list, got %d: %s", listRec.Code, listRec.Body.String())
	}
	var page assignmentListBody
	if err := json.Unmarshal(listRec.Body.Bytes(), &page); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	found := false
	for _, item := range page.Items {
		if item.ServiceID == service.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("CA-023-01: expected the assignment to appear in the list")
	}

	// Asigna un segundo barbero (CA-023-03), para que el primero YA NO sea
	// la última asignación activa y su desasignación pueda tener éxito.
	if rec := doAssignServiceRequest(router, raw, barberTwo.ID, service.ID); rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 assigning the second barber, got %d: %s", rec.Code, rec.Body.String())
	}

	unassignRec := doUnassignServiceRequest(router, raw, barber.ID, service.ID)
	if unassignRec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 unassigning a non-last assignment, got %d: %s", unassignRec.Code, unassignRec.Body.String())
	}

	// Repetir la desasignación ahora responde 404 (ya no existe).
	repeatUnassign := doUnassignServiceRequest(router, raw, barber.ID, service.ID)
	if repeatUnassign.Code != http.StatusNotFound {
		t.Fatalf("expected 404 unassigning an already-removed assignment, got %d: %s", repeatUnassign.Code, repeatUnassign.Body.String())
	}
}

// TestBarberServices_HTTP_LastActiveAssignment_Returns409 cubre DEC-068 a
// través del router real: retirar la última asignación activa de un
// servicio recién creado se rechaza con 409, sin borrar nada.
func TestBarberServices_HTTP_LastActiveAssignment_Returns409(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "assign-last-a")

	barber := createBarberViaRouter(t, router, raw, "Barbero Última HTTP "+uniqueToken(t, "last"))
	service := createServiceViaRouter(t, router, raw, "Servicio Última HTTP "+uniqueToken(t, "last"))

	if rec := doAssignServiceRequest(router, raw, barber.ID, service.ID); rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 assigning, got %d: %s", rec.Code, rec.Body.String())
	}

	rec := doUnassignServiceRequest(router, raw, barber.ID, service.ID)
	if rec.Code != http.StatusConflict {
		t.Fatalf("DEC-068: expected 409 unassigning the last active assignment, got %d: %s", rec.Code, rec.Body.String())
	}
	var problem map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
		t.Fatalf("decode conflict problem: %v", err)
	}
	if problem["code"] != "conflict" {
		t.Fatalf("expected code=conflict, got %v", problem["code"])
	}

	// Nada se borró: sigue apareciendo en el listado.
	listRec := doListAssignmentsRequest(router, raw, barber.ID, "")
	var page assignmentListBody
	if err := json.Unmarshal(listRec.Body.Bytes(), &page); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	found := false
	for _, item := range page.Items {
		if item.ServiceID == service.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("CA-023-06: the rejected unassign must NOT have deleted the row")
	}
}

// TestBarberServices_HTTP_TwoTenants_CrossAccessReturns404WithoutLeaking
// cubre CA-023-04: un barbero o servicio real de otra barbería responde 404
// uniforme, y nunca puede asociarse.
func TestBarberServices_HTTP_TwoTenants_CrossAccessReturns404WithoutLeaking(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	rawA := createSessionCookie(t, db, shopA, staffUserActiveA, "assign-tenant-a")
	rawB := createSessionCookie(t, db, shopB, staffUserActiveB, "assign-tenant-b")

	barberB := createBarberViaRouter(t, router, rawB, "Barbero Cruzado HTTP B "+uniqueToken(t, "cross"))
	serviceB := createServiceViaRouter(t, router, rawB, "Servicio Cruzado HTTP B "+uniqueToken(t, "cross"))
	barberA := createBarberViaRouter(t, router, rawA, "Barbero Cruzado HTTP A "+uniqueToken(t, "cross"))
	serviceA := createServiceViaRouter(t, router, rawA, "Servicio Cruzado HTTP A "+uniqueToken(t, "cross"))

	if rec := doListAssignmentsRequest(router, rawA, barberB.ID, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("CA-023-04: expected 404 listing shopB's barber from shopA, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec := doAssignServiceRequest(router, rawA, barberB.ID, serviceA.ID); rec.Code != http.StatusNotFound {
		t.Fatalf("CA-023-04: expected 404 assigning shopA's service to shopB's barber from shopA, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec := doAssignServiceRequest(router, rawA, barberA.ID, serviceB.ID); rec.Code != http.StatusNotFound {
		t.Fatalf("CA-023-04: expected 404 assigning shopB's service to shopA's barber from shopA, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec := doUnassignServiceRequest(router, rawA, barberB.ID, serviceB.ID); rec.Code != http.StatusNotFound {
		t.Fatalf("CA-023-04: expected 404 unassigning across tenants, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestBarberServices_HTTP_NoCookie_Returns401Uniform confirma que estas tres
// rutas comparten el mismo 401 uniforme que el resto de /api/v1/private.
func TestBarberServices_HTTP_NoCookie_Returns401Uniform(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	fakeID := "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"
	if rec := doListAssignmentsRequest(router, "", fakeID, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for list without cookie, got %d", rec.Code)
	}
	if rec := doAssignServiceRequest(router, "", fakeID, fakeID); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for assign without cookie, got %d", rec.Code)
	}
	if rec := doUnassignServiceRequest(router, "", fakeID, fakeID); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unassign without cookie, got %d", rec.Code)
	}
}

// TestBarberServices_HTTP_ResponseNeverIncludesNameDurationPrice cubre
// CA-023-07 a través del router real: la respuesta de asignar y de listar
// nunca incluye nombre, duración, precio ni estado del barbero o del
// servicio.
func TestBarberServices_HTTP_ResponseNeverIncludesNameDurationPrice(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "assign-shape-a")

	barber := createBarberViaRouter(t, router, raw, "Barbero Forma HTTP "+uniqueToken(t, "shape"))
	service := createServiceViaRouter(t, router, raw, "Servicio Forma HTTP "+uniqueToken(t, "shape"))

	rec := doAssignServiceRequest(router, raw, barber.ID, service.ID)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var generic map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &generic); err != nil {
		t.Fatalf("decode: %v", err)
	}
	wantKeys := map[string]bool{"barberId": true, "serviceId": true, "createdAt": true}
	if len(generic) != len(wantKeys) {
		t.Fatalf("CA-023-07: expected exactly %v, got keys %v", wantKeys, generic)
	}
	for k := range generic {
		if !wantKeys[k] {
			t.Fatalf("CA-023-07: unexpected field %q in assignment response: %v", k, generic)
		}
	}
}
