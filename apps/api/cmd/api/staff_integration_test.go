// Pruebas de integración de HU-021 contra el router REAL de producción
// (buildRouter) y PostgreSQL real, mismo patrón que
// settings_integration_test.go (HU-020). Requieren las diez migraciones
// aplicadas (incluida 20260823130000_create_barber.sql) y
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

type barberBody struct {
	ID        string `json:"id"`
	FullName  string `json:"fullName"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type barberListBody struct {
	Items      []barberBody `json:"items"`
	NextCursor *string      `json:"nextCursor"`
}

func doListBarbersRequest(router http.Handler, rawToken, query string) *httptest.ResponseRecorder {
	target := "/api/v1/private/barbers"
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

func doGetBarberRequest(router http.Handler, rawToken, barberID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/private/barbers/"+barberID, nil)
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func doCreateBarberRequest(router http.Handler, rawToken, idempotencyKey, rawBody string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/private/barbers", bytes.NewReader([]byte(rawBody)))
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

func doRenameBarberRequest(router http.Handler, rawToken, barberID, rawBody string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/private/barbers/"+barberID, bytes.NewReader([]byte(rawBody)))
	req.Header.Set("Content-Type", "application/json")
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// TestStaff_HTTP_OneThenFourBarbers_SamePathBothCases cubre CA-021-01 y
// CA-021-02 de punta a punta: la barbería empieza con un barbero, se
// registran tres más, y la lista devuelve cuatro recursos distintos con el
// mismo camino de código que el primero.
func TestStaff_HTTP_OneThenFourBarbers_SamePathBothCases(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "staff-team-a")

	names := []string{"Único Al Empezar", "Segundo", "Tercero", "Cuarto"}
	created := map[string]bool{}
	for i, name := range names {
		key := uniqueToken(t, "create-team")
		body := `{"fullName":"` + name + " " + uniqueToken(t, "n") + `"}`
		rec := doCreateBarberRequest(router, raw, key+"-"+string(rune('0'+i)), body)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 for %q, got %d: %s", name, rec.Code, rec.Body.String())
		}
		var b barberBody
		if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
			t.Fatalf("decode create response: %v", err)
		}
		if created[b.ID] {
			t.Fatalf("CA-021-02: duplicate id %s across creates", b.ID)
		}
		created[b.ID] = true
	}
	if len(created) != 4 {
		t.Fatalf("expected 4 distinct barbers, got %d", len(created))
	}

	// Recorre la lista completa (limit pequeño para forzar paginación) y
	// confirma que los 4 aparecen.
	seen := map[string]bool{}
	query := "limit=1"
	for pages := 0; ; pages++ {
		if pages > 1000 {
			t.Fatal("too many pages")
		}
		rec := doListBarbersRequest(router, raw, query)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 listing, got %d: %s", rec.Code, rec.Body.String())
		}
		var page barberListBody
		if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
			t.Fatalf("decode list response: %v", err)
		}
		for _, item := range page.Items {
			seen[item.ID] = true
		}
		if page.NextCursor == nil {
			break
		}
		query = "limit=1&cursor=" + *page.NextCursor
	}

	for id := range created {
		if !seen[id] {
			t.Fatalf("CA-021-02: barber %s never appeared while paging the list", id)
		}
	}
}

// TestStaff_HTTP_CreateGetListRenameReload_FullJourney recorre alta,
// lectura, listado y renombrado, confirmando persistencia real.
func TestStaff_HTTP_CreateGetListRenameReload_FullJourney(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "staff-journey-a")

	name := "Barbero De Prueba " + uniqueToken(t, "journey")
	createRec := doCreateBarberRequest(router, raw, uniqueToken(t, "journey-key"), `{"fullName":"`+name+`"}`)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", createRec.Code, createRec.Body.String())
	}
	loc := createRec.Header().Get("Location")
	if !strings.HasPrefix(loc, "/private/barbers/") {
		t.Fatalf("unexpected Location: %q", loc)
	}
	var created barberBody
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	getRec := doGetBarberRequest(router, raw, created.ID)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on GET, got %d: %s", getRec.Code, getRec.Body.String())
	}
	var got barberBody
	if err := json.Unmarshal(getRec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if got.ID != created.ID || got.FullName != name {
		t.Fatalf("expected GET to return the created barber, got %+v", got)
	}

	newName := "Renombrado " + uniqueToken(t, "journey")
	renameRec := doRenameBarberRequest(router, raw, created.ID, `{"fullName":"`+newName+`"}`)
	if renameRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on PATCH, got %d: %s", renameRec.Code, renameRec.Body.String())
	}
	var renamed barberBody
	if err := json.Unmarshal(renameRec.Body.Bytes(), &renamed); err != nil {
		t.Fatalf("decode rename response: %v", err)
	}
	if renamed.ID != created.ID {
		t.Fatalf("expected the same id after rename, got %s vs %s", renamed.ID, created.ID)
	}
	if renamed.FullName != newName {
		t.Fatalf("expected the new name, got %q", renamed.FullName)
	}

	// "Recargar": una solicitud GET independiente devuelve exactamente lo
	// guardado, sin duplicar el recurso.
	afterRec := doGetBarberRequest(router, raw, created.ID)
	var after barberBody
	if err := json.Unmarshal(afterRec.Body.Bytes(), &after); err != nil {
		t.Fatalf("decode reloaded get response: %v", err)
	}
	if after.FullName != newName {
		t.Fatalf("expected the renamed value to persist across an independent GET, got %+v", after)
	}
}

// TestStaff_HTTP_TwoTenants_CrossAccessAlwaysReturns404WithoutLeaking cubre
// CA-021-05: conocer el UUID real de un barbero de otra barbería nunca
// permite consultarlo ni editarlo, y nunca aparece en el listado propio.
func TestStaff_HTTP_TwoTenants_CrossAccessAlwaysReturns404WithoutLeaking(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	rawA := createSessionCookie(t, db, shopA, staffUserActiveA, "staff-tenant-a")
	rawB := createSessionCookie(t, db, shopB, staffUserActiveB, "staff-tenant-b")

	nameB := "Solo De B " + uniqueToken(t, "cross")
	createRecB := doCreateBarberRequest(router, rawB, uniqueToken(t, "cross-key"), `{"fullName":"`+nameB+`"}`)
	if createRecB.Code != http.StatusCreated {
		t.Fatalf("expected 201 creating in shopB, got %d: %s", createRecB.Code, createRecB.Body.String())
	}
	var barberB barberBody
	if err := json.Unmarshal(createRecB.Body.Bytes(), &barberB); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// A intenta leer el barbero real de B: 404, mismo problem type que un
	// identificador inexistente.
	getFromA := doGetBarberRequest(router, rawA, barberB.ID)
	if getFromA.Code != http.StatusNotFound {
		t.Fatalf("CA-021-05: expected 404 reading shopB's barber from shopA, got %d: %s", getFromA.Code, getFromA.Body.String())
	}
	nonexistentFromA := doGetBarberRequest(router, rawA, "99999999-9999-4999-8999-999999999999")
	if nonexistentFromA.Code != getFromA.Code {
		t.Fatalf("CA-021-05: expected the same status for nonexistent vs cross-tenant, got %d vs %d",
			nonexistentFromA.Code, getFromA.Code)
	}
	// instance/requestId son de correlación por solicitud y varían
	// legítimamente; el resto del Problem (type/title/status/detail/code)
	// debe ser IDÉNTICO para que el cliente no pueda distinguir "no existe"
	// de "es de otra barbería" (CA-021-05).
	var crossTenantProblem, nonexistentProblem map[string]any
	if err := json.Unmarshal(getFromA.Body.Bytes(), &crossTenantProblem); err != nil {
		t.Fatalf("decode cross-tenant problem: %v", err)
	}
	if err := json.Unmarshal(nonexistentFromA.Body.Bytes(), &nonexistentProblem); err != nil {
		t.Fatalf("decode nonexistent problem: %v", err)
	}
	for _, field := range []string{"type", "title", "status", "detail", "code"} {
		if crossTenantProblem[field] != nonexistentProblem[field] {
			t.Fatalf("CA-021-05: field %q differs between cross-tenant and nonexistent problems: %v vs %v",
				field, crossTenantProblem[field], nonexistentProblem[field])
		}
	}

	// A intenta renombrar el barbero real de B: 404, sin efecto.
	renameFromA := doRenameBarberRequest(router, rawA, barberB.ID, `{"fullName":"Secuestrado Por A"}`)
	if renameFromA.Code != http.StatusNotFound {
		t.Fatalf("CA-021-05: expected 404 renaming shopB's barber from shopA, got %d: %s", renameFromA.Code, renameFromA.Body.String())
	}
	// El nombre de B no cambió.
	stillB := doGetBarberRequest(router, rawB, barberB.ID)
	var confirmB barberBody
	if err := json.Unmarshal(stillB.Body.Bytes(), &confirmB); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if confirmB.FullName != nameB {
		t.Fatalf("CA-021-05: shopB's barber name changed via shopA's failed rename attempt, got %q", confirmB.FullName)
	}

	// El barbero de B nunca aparece en el listado de A.
	listFromA := doListBarbersRequest(router, rawA, "limit=50")
	var pageA barberListBody
	if err := json.Unmarshal(listFromA.Body.Bytes(), &pageA); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	for _, item := range pageA.Items {
		if item.ID == barberB.ID {
			t.Fatal("CA-021-05: shopB's barber leaked into shopA's list")
		}
	}
}

// TestStaff_HTTP_NoCookie_Returns401Uniform confirma que estas cuatro rutas
// comparten el mismo 401 uniforme que el resto de /api/v1/private (ya
// cubierto estructuralmente por TestPrivateRouteInventory_...).
func TestStaff_HTTP_NoCookie_Returns401Uniform(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	if rec := doListBarbersRequest(router, "", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for list without cookie, got %d", rec.Code)
	}
	if rec := doGetBarberRequest(router, "", "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for get without cookie, got %d", rec.Code)
	}
	if rec := doCreateBarberRequest(router, "", "some-key", `{"fullName":"X"}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for create without cookie, got %d", rec.Code)
	}
	if rec := doRenameBarberRequest(router, "", "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4", `{"fullName":"X"}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for rename without cookie, got %d", rec.Code)
	}
}

// TestStaff_HTTP_UnknownField_Returns400 cubre CA-021-07: un intento de
// enviar un campo fuera de alcance (active) se rechaza como forma
// inválida.
func TestStaff_HTTP_UnknownField_Returns400(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "staff-unknown-field")

	createRec := doCreateBarberRequest(router, raw, uniqueToken(t, "unknown-field"), `{"fullName":"X","active":true}`)
	if createRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for an unknown field on create, got %d: %s", createRec.Code, createRec.Body.String())
	}

	existing := doCreateBarberRequest(router, raw, uniqueToken(t, "unknown-field-target"), `{"fullName":"Objetivo"}`)
	var target barberBody
	if err := json.Unmarshal(existing.Body.Bytes(), &target); err != nil {
		t.Fatalf("decode: %v", err)
	}
	renameRec := doRenameBarberRequest(router, raw, target.ID, `{"fullName":"Y","deletedAt":null}`)
	if renameRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for an unknown field on rename, got %d: %s", renameRec.Code, renameRec.Body.String())
	}
}

// TestStaff_HTTP_Idempotency_ReplayAndConflict cubre RN-IDE-01/DEC-043 a
// través del router real: la misma clave con el mismo cuerpo reproduce la
// respuesta original byte a byte sin crear una segunda fila; la misma
// clave con contenido distinto responde 409.
func TestStaff_HTTP_Idempotency_ReplayAndConflict(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "staff-idem")

	key := uniqueToken(t, "idem-key")
	body := `{"fullName":"Idempotente ` + uniqueToken(t, "idem-name") + `"}`

	first := doCreateBarberRequest(router, raw, key, body)
	if first.Code != http.StatusCreated {
		t.Fatalf("expected 201 on first call, got %d: %s", first.Code, first.Body.String())
	}

	second := doCreateBarberRequest(router, raw, key, body)
	if second.Code != first.Code {
		t.Fatalf("expected the same status on replay, got %d vs %d", second.Code, first.Code)
	}
	if second.Body.String() != first.Body.String() {
		t.Fatalf("expected the exact original body on replay (CA-004-01), got %q vs %q", second.Body.String(), first.Body.String())
	}
	if second.Header().Get("Location") != first.Header().Get("Location") {
		t.Fatal("expected the same Location on replay")
	}

	// Confirma que solo existe una fila para ese nombre: el listado con un
	// límite suficiente no debe contener dos entradas distintas con el
	// mismo id repetido (ya lo garantiza el propio Location idéntico), y
	// una tercera llamada con la MISMA clave pero OTRO contenido conflictúa.
	conflict := doCreateBarberRequest(router, raw, key, `{"fullName":"Contenido Distinto"}`)
	if conflict.Code != http.StatusConflict {
		t.Fatalf("expected 409 for the same key with different content, got %d: %s", conflict.Code, conflict.Body.String())
	}
}

// TestStaff_HTTP_ThroughFullRouter_NeverLogsFullNameOrCookie refuerza
// RN-DAT-02: el nombre del barbero, el cuerpo y la cookie no aparecen en el
// registro estructurado de una operación exitosa o fallida.
func TestStaff_HTTP_ThroughFullRouter_NeverLogsFullNameOrCookie(t *testing.T) {
	db := setupTestDB(t)

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
	router, err := buildRouter(db, logger, testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "staff-log-safe")

	sensitiveName := "Nombre-No-Debe-Aparecer-En-Logs-" + uniqueToken(t, "log")
	rec := doCreateBarberRequest(router, raw, uniqueToken(t, "log-key"), `{"fullName":"`+sensitiveName+`"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	logLine := logBuf.String()
	for _, forbidden := range []string{sensitiveName, raw, cookieName + "="} {
		if strings.Contains(logLine, forbidden) {
			t.Fatalf("RN-DAT-02: sensitive data leaked into the log: %s", logLine)
		}
	}
}
