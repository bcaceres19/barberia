// Pruebas de integración de HU-022 contra el router REAL de producción
// (buildRouter) y PostgreSQL real, mismo patrón que
// staff_integration_test.go (HU-021). Requieren las once migraciones
// aplicadas (incluida 20260824140000_create_service.sql) y
// database/testdata/dos_barberias.sql + testdata/hu005_credenciales_sesiones.sql
// cargados.
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./cmd/api/...
package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type serviceBody struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Description     *string `json:"description"`
	DurationMinutes int     `json:"durationMinutes"`
	Price           string  `json:"price"`
	Currency        string  `json:"currency"`
	IsActive        bool    `json:"isActive"`
	DeactivatedAt   *string `json:"deactivatedAt"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

type serviceListBody struct {
	Items      []serviceBody `json:"items"`
	NextCursor *string       `json:"nextCursor"`
}

func doListServicesRequest(router http.Handler, rawToken, query string) *httptest.ResponseRecorder {
	target := "/api/v1/private/services"
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

func doGetServiceRequest(router http.Handler, rawToken, serviceID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/private/services/"+serviceID, nil)
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func doCreateServiceRequest(router http.Handler, rawToken, idempotencyKey, rawBody string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/private/services", bytes.NewReader([]byte(rawBody)))
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

func doUpdateServiceRequest(router http.Handler, rawToken, serviceID, rawBody string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/private/services/"+serviceID, bytes.NewReader([]byte(rawBody)))
	req.Header.Set("Content-Type", "application/json")
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// TestCatalog_HTTP_CreateGetListUpdateReload_FullJourney recorre alta,
// lectura, listado y edición, confirmando persistencia real (CA-022-01,
// CA-022-02, CA-022-04, CA-022-05).
func TestCatalog_HTTP_CreateGetListUpdateReload_FullJourney(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "catalog-journey-a")

	name := "Servicio De Prueba " + uniqueToken(t, "journey")
	createRec := doCreateServiceRequest(router, raw, uniqueToken(t, "journey-key"),
		`{"name":"`+name+`","description":"Descripción de prueba","durationMinutes":30,"price":"45000.00"}`)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", createRec.Code, createRec.Body.String())
	}
	loc := createRec.Header().Get("Location")
	if !strings.HasPrefix(loc, "/private/services/") {
		t.Fatalf("unexpected Location: %q", loc)
	}
	var created serviceBody
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.Price != "45000.00" || created.Currency != "COP" {
		t.Fatalf("unexpected price/currency: %+v", created)
	}

	getRec := doGetServiceRequest(router, raw, created.ID)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on GET, got %d: %s", getRec.Code, getRec.Body.String())
	}
	var got serviceBody
	if err := json.Unmarshal(getRec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if got.ID != created.ID || got.Name != name {
		t.Fatalf("expected GET to return the created service, got %+v", got)
	}

	// Aparece en el listado.
	listRec := doListServicesRequest(router, raw, "limit=50")
	var page serviceListBody
	if err := json.Unmarshal(listRec.Body.Bytes(), &page); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	found := false
	for _, item := range page.Items {
		if item.ID == created.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("CA-022-01: expected the created service to appear in the list")
	}

	// Edición parcial: solo precio y duración.
	updateRec := doUpdateServiceRequest(router, raw, created.ID, `{"durationMinutes":45,"price":"50000.00"}`)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on PATCH, got %d: %s", updateRec.Code, updateRec.Body.String())
	}
	var updated serviceBody
	if err := json.Unmarshal(updateRec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode update response: %v", err)
	}
	if updated.ID != created.ID {
		t.Fatalf("expected the same id after update, got %s vs %s", updated.ID, created.ID)
	}
	if updated.DurationMinutes != 45 || updated.Price != "50000.00" {
		t.Fatalf("expected updated duration/price, got %+v", updated)
	}
	if updated.Name != name {
		t.Fatalf("CA-022-05: expected name untouched by an unrelated edit, got %q", updated.Name)
	}

	// "Recargar": una solicitud GET independiente devuelve exactamente lo
	// guardado, sin duplicar el recurso.
	afterRec := doGetServiceRequest(router, raw, created.ID)
	var after serviceBody
	if err := json.Unmarshal(afterRec.Body.Bytes(), &after); err != nil {
		t.Fatalf("decode reloaded get response: %v", err)
	}
	if after.DurationMinutes != 45 || after.Price != "50000.00" {
		t.Fatalf("expected the edit to persist across an independent GET, got %+v", after)
	}
}

// TestCatalog_HTTP_TwoTenants_CrossAccessAlwaysReturns404WithoutLeaking cubre
// CA-022-06: conocer el UUID real de un servicio de otra barbería nunca
// permite consultarlo ni editarlo, y nunca aparece en el listado propio.
func TestCatalog_HTTP_TwoTenants_CrossAccessAlwaysReturns404WithoutLeaking(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	rawA := createSessionCookie(t, db, shopA, staffUserActiveA, "catalog-tenant-a")
	rawB := createSessionCookie(t, db, shopB, staffUserActiveB, "catalog-tenant-b")

	nameB := "Solo De B " + uniqueToken(t, "cross")
	createRecB := doCreateServiceRequest(router, rawB, uniqueToken(t, "cross-key"),
		`{"name":"`+nameB+`","durationMinutes":30,"price":"45000.00"}`)
	if createRecB.Code != http.StatusCreated {
		t.Fatalf("expected 201 creating in shopB, got %d: %s", createRecB.Code, createRecB.Body.String())
	}
	var serviceB serviceBody
	if err := json.Unmarshal(createRecB.Body.Bytes(), &serviceB); err != nil {
		t.Fatalf("decode: %v", err)
	}

	getFromA := doGetServiceRequest(router, rawA, serviceB.ID)
	if getFromA.Code != http.StatusNotFound {
		t.Fatalf("CA-022-06: expected 404 reading shopB's service from shopA, got %d: %s", getFromA.Code, getFromA.Body.String())
	}
	nonexistentFromA := doGetServiceRequest(router, rawA, "99999999-9999-4999-8999-999999999999")
	if nonexistentFromA.Code != getFromA.Code {
		t.Fatalf("CA-022-06: expected the same status for nonexistent vs cross-tenant, got %d vs %d",
			nonexistentFromA.Code, getFromA.Code)
	}
	var crossTenantProblem, nonexistentProblem map[string]any
	if err := json.Unmarshal(getFromA.Body.Bytes(), &crossTenantProblem); err != nil {
		t.Fatalf("decode cross-tenant problem: %v", err)
	}
	if err := json.Unmarshal(nonexistentFromA.Body.Bytes(), &nonexistentProblem); err != nil {
		t.Fatalf("decode nonexistent problem: %v", err)
	}
	for _, field := range []string{"type", "title", "status", "detail", "code"} {
		if crossTenantProblem[field] != nonexistentProblem[field] {
			t.Fatalf("CA-022-06: field %q differs between cross-tenant and nonexistent problems: %v vs %v",
				field, crossTenantProblem[field], nonexistentProblem[field])
		}
	}

	updateFromA := doUpdateServiceRequest(router, rawA, serviceB.ID, `{"name":"Secuestrado Por A"}`)
	if updateFromA.Code != http.StatusNotFound {
		t.Fatalf("CA-022-06: expected 404 editing shopB's service from shopA, got %d: %s", updateFromA.Code, updateFromA.Body.String())
	}
	stillB := doGetServiceRequest(router, rawB, serviceB.ID)
	var confirmB serviceBody
	if err := json.Unmarshal(stillB.Body.Bytes(), &confirmB); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if confirmB.Name != nameB {
		t.Fatalf("CA-022-06: shopB's service name changed via shopA's failed edit attempt, got %q", confirmB.Name)
	}

	listFromA := doListServicesRequest(router, rawA, "limit=50")
	var pageA serviceListBody
	if err := json.Unmarshal(listFromA.Body.Bytes(), &pageA); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	for _, item := range pageA.Items {
		if item.ID == serviceB.ID {
			t.Fatal("CA-022-06: shopB's service leaked into shopA's list")
		}
	}
}

// TestCatalog_HTTP_NoCookie_Returns401Uniform confirma que estas cuatro
// rutas comparten el mismo 401 uniforme que el resto de /api/v1/private.
func TestCatalog_HTTP_NoCookie_Returns401Uniform(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	if rec := doListServicesRequest(router, "", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for list without cookie, got %d", rec.Code)
	}
	if rec := doGetServiceRequest(router, "", "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for get without cookie, got %d", rec.Code)
	}
	if rec := doCreateServiceRequest(router, "", "some-key", `{"name":"X","durationMinutes":30,"price":"1.00"}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for create without cookie, got %d", rec.Code)
	}
	if rec := doUpdateServiceRequest(router, "", "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4", `{"name":"X"}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for update without cookie, got %d", rec.Code)
	}
}

// TestCatalog_HTTP_UnknownField_Returns400 cubre CA-022-07: un intento de
// enviar un campo fuera de alcance (isActive) se rechaza como forma
// inválida.
func TestCatalog_HTTP_UnknownField_Returns400(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "catalog-unknown-field")

	createRec := doCreateServiceRequest(router, raw, uniqueToken(t, "unknown-field"),
		`{"name":"X","durationMinutes":30,"price":"1.00","isActive":true}`)
	if createRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for an unknown field on create, got %d: %s", createRec.Code, createRec.Body.String())
	}

	existing := doCreateServiceRequest(router, raw, uniqueToken(t, "unknown-field-target"),
		`{"name":"Objetivo `+uniqueToken(t, "target")+`","durationMinutes":30,"price":"1.00"}`)
	var target serviceBody
	if err := json.Unmarshal(existing.Body.Bytes(), &target); err != nil {
		t.Fatalf("decode: %v", err)
	}
	updateRec := doUpdateServiceRequest(router, raw, target.ID, `{"name":"Y","isActive":false}`)
	if updateRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for an unknown field on update, got %d: %s", updateRec.Code, updateRec.Body.String())
	}
}

// TestCatalog_HTTP_Idempotency_ReplayAndConflict cubre RN-IDE-01/DEC-043 a
// través del router real: la misma clave con el mismo cuerpo reproduce la
// respuesta original byte a byte sin crear una segunda fila; la misma
// clave con contenido distinto responde 409.
func TestCatalog_HTTP_Idempotency_ReplayAndConflict(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "catalog-idem")

	key := uniqueToken(t, "idem-key")
	body := `{"name":"Idempotente ` + uniqueToken(t, "idem-name") + `","durationMinutes":30,"price":"1.00"}`

	first := doCreateServiceRequest(router, raw, key, body)
	if first.Code != http.StatusCreated {
		t.Fatalf("expected 201 on first call, got %d: %s", first.Code, first.Body.String())
	}

	second := doCreateServiceRequest(router, raw, key, body)
	if second.Code != first.Code {
		t.Fatalf("expected the same status on replay, got %d vs %d", second.Code, first.Code)
	}
	if second.Body.String() != first.Body.String() {
		t.Fatalf("expected the exact original body on replay (CA-004-01), got %q vs %q", second.Body.String(), first.Body.String())
	}
	if second.Header().Get("Location") != first.Header().Get("Location") {
		t.Fatal("expected the same Location on replay")
	}

	conflict := doCreateServiceRequest(router, raw, key, `{"name":"Contenido Distinto","durationMinutes":30,"price":"1.00"}`)
	if conflict.Code != http.StatusConflict {
		t.Fatalf("expected 409 for the same key with different content, got %d: %s", conflict.Code, conflict.Body.String())
	}
}

// TestCatalog_HTTP_DuplicateActiveName_Returns409 cubre DEC-067 a través del
// router real: un segundo alta con el mismo nombre exacto (y una clave de
// idempotencia distinta) responde 409 sin crear una segunda fila.
func TestCatalog_HTTP_DuplicateActiveName_Returns409(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "catalog-dup-name")

	name := "Nombre Único HTTP " + uniqueToken(t, "dup")
	first := doCreateServiceRequest(router, raw, uniqueToken(t, "dup-key-1"),
		`{"name":"`+name+`","durationMinutes":30,"price":"1.00"}`)
	if first.Code != http.StatusCreated {
		t.Fatalf("expected 201 for the first service, got %d: %s", first.Code, first.Body.String())
	}

	second := doCreateServiceRequest(router, raw, uniqueToken(t, "dup-key-2"),
		`{"name":"`+name+`","durationMinutes":45,"price":"2.00"}`)
	if second.Code != http.StatusConflict {
		t.Fatalf("DEC-067: expected 409 for a duplicate active name, got %d: %s", second.Code, second.Body.String())
	}
	problem := map[string]any{}
	if err := json.Unmarshal(second.Body.Bytes(), &problem); err != nil {
		t.Fatalf("decode conflict problem: %v", err)
	}
	if problem["code"] != "conflict" {
		t.Fatalf("expected code=conflict (distinct from idempotency-conflict), got %v", problem["code"])
	}
}

// TestCatalog_HTTP_ThroughFullRouter_NeverLogsNameDescriptionOrCookie
// refuerza RN-DAT-02: el nombre, la descripción, el cuerpo y la cookie no
// aparecen en el registro estructurado de una operación exitosa.
func TestCatalog_HTTP_ThroughFullRouter_NeverLogsNameDescriptionOrCookie(t *testing.T) {
	db := setupTestDB(t)

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
	router, err := buildRouter(db, logger, testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "catalog-log-safe")

	sensitiveName := "Nombre-No-Debe-Aparecer-En-Logs-" + uniqueToken(t, "log")
	sensitiveDescription := "Descripcion-No-Debe-Aparecer-" + uniqueToken(t, "log-desc")
	rec := doCreateServiceRequest(router, raw, uniqueToken(t, "log-key"),
		`{"name":"`+sensitiveName+`","description":"`+sensitiveDescription+`","durationMinutes":30,"price":"1.00"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	logLine := logBuf.String()
	for _, forbidden := range []string{sensitiveName, sensitiveDescription, raw, cookieName + "="} {
		if strings.Contains(logLine, forbidden) {
			t.Fatalf("RN-DAT-02: sensitive data leaked into the log: %s", logLine)
		}
	}
}
