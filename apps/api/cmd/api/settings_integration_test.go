// Pruebas de integración de HU-020 contra el router REAL de producción
// (buildRouter) y PostgreSQL real, mismo patrón que
// session_context_integration_test.go (HU-012). Requieren las nueve
// migraciones aplicadas (incluida
// 20260823120000_add_barbershop_contact_info.sql) y
// database/testdata/dos_barberias.sql + testdata/hu005_credenciales_sesiones.sql
// cargados.
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./cmd/api/...
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"system-barbershop/internal/platform/database"
)

type barbershopSettingsBody struct {
	Name         string  `json:"name"`
	Timezone     string  `json:"timezone"`
	ContactEmail *string `json:"contactEmail"`
	ContactPhone *string `json:"contactPhone"`
}

func doGetSettingsRequest(router http.Handler, rawToken string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/private/settings/barbershop", nil)
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func doPatchSettingsRequest(router http.Handler, rawToken string, body barbershopSettingsBody) *httptest.ResponseRecorder {
	raw, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/private/settings/barbershop", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// restoreBarbershopRow deja shop en el mismo estado que dos_barberias.sql
// al terminar la prueba: estas pruebas de integración comparten esas dos
// filas reales en vez de crear una barbería propia (HU-020 no expone alta
// de barberías).
func restoreBarbershopRow(t *testing.T, db *database.DB, shop, name string) {
	t.Helper()
	t.Cleanup(func() {
		err := db.InTenantTx(context.Background(), database.BarbershopID(shop), func(ctx context.Context, q database.Queries) error {
			_, err := q.Exec(ctx,
				`UPDATE barbershop SET name = $2, timezone = 'America/Bogota', contact_email = NULL, contact_phone = NULL WHERE id = $1`,
				shop, name,
			)
			return err
		})
		if err != nil {
			t.Fatalf("restoreBarbershopRow cleanup: %v", err)
		}
	})
}

// TestSettings_HTTP_ValidSession_ReturnsOwnBarbershop cubre CA-020-01 a
// nivel HTTP/PostgreSQL real: la lectura devuelve exactamente los cuatro
// campos autorizados de la barbería propietaria de la sesión.
func TestSettings_HTTP_ValidSession_ReturnsOwnBarbershop(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	restoreBarbershopRow(t, db, shopA, "Barbería de prueba A")

	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "settings-get-a")
	rec := doGetSettingsRequest(router, raw)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body barbershopSettingsBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v: %s", err, rec.Body.String())
	}
	if body.Name != "Barbería de prueba A" || body.Timezone != "America/Bogota" {
		t.Fatalf("unexpected body: %+v", body)
	}
	if body.ContactEmail != nil || body.ContactPhone != nil {
		t.Fatalf("expected no contact configured yet, got %+v", body)
	}
}

// TestSettings_HTTP_UpdateThenGet_PersistsAndRoundTrips cubre CA-020-02:
// guardar y volver a leer (una solicitud independiente) devuelve
// exactamente lo guardado.
func TestSettings_HTTP_UpdateThenGet_PersistsAndRoundTrips(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	restoreBarbershopRow(t, db, shopA, "Barbería de prueba A")

	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "settings-update-a")
	email := "contacto@ejemplo.test"
	phone := "+573001234567"

	patchRec := doPatchSettingsRequest(router, raw, barbershopSettingsBody{
		Name: "Barbería renombrada", Timezone: "America/Bogota",
		ContactEmail: &email, ContactPhone: &phone,
	})
	if patchRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on PATCH, got %d: %s", patchRec.Code, patchRec.Body.String())
	}

	getRec := doGetSettingsRequest(router, raw)
	var body barbershopSettingsBody
	if err := json.Unmarshal(getRec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode GET response: %v", err)
	}
	if body.Name != "Barbería renombrada" {
		t.Fatalf("expected the renamed value to round-trip, got %+v", body)
	}
	if body.ContactEmail == nil || *body.ContactEmail != email {
		t.Fatalf("expected contactEmail=%q to round-trip, got %v", email, body.ContactEmail)
	}
}

// TestSettings_HTTP_TwoTenants_UpdateNeverCrossesBarbershops es la prueba
// de aislamiento por tenant exigida por el prompt (CA-020-05): actualizar
// la barbería de A, contra el endpoint real con dos tenants reales, nunca
// afecta la de B.
func TestSettings_HTTP_TwoTenants_UpdateNeverCrossesBarbershops(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	restoreBarbershopRow(t, db, shopA, "Barbería de prueba A")
	restoreBarbershopRow(t, db, shopB, "Barbería de prueba B")

	rawA := createSessionCookie(t, db, shopA, staffUserActiveA, "settings-tenant-a")
	rawB := createSessionCookie(t, db, shopB, staffUserActiveB, "settings-tenant-b")

	beforeB := doGetSettingsRequest(router, rawB)
	var bodyBeforeB barbershopSettingsBody
	if err := json.Unmarshal(beforeB.Body.Bytes(), &bodyBeforeB); err != nil {
		t.Fatalf("decode shopB response before: %v", err)
	}

	patchRec := doPatchSettingsRequest(router, rawA, barbershopSettingsBody{
		Name: "Solo A debe cambiar", Timezone: "America/Bogota",
	})
	if patchRec.Code != http.StatusOK {
		t.Fatalf("expected 200 updating shopA, got %d: %s", patchRec.Code, patchRec.Body.String())
	}

	afterB := doGetSettingsRequest(router, rawB)
	var bodyAfterB barbershopSettingsBody
	if err := json.Unmarshal(afterB.Body.Bytes(), &bodyAfterB); err != nil {
		t.Fatalf("decode shopB response after: %v", err)
	}
	if bodyAfterB.Name != bodyBeforeB.Name {
		t.Fatalf("CA-020-05: updating shopA changed shopB's name from %q to %q", bodyBeforeB.Name, bodyAfterB.Name)
	}

	afterA := doGetSettingsRequest(router, rawA)
	var bodyAfterA barbershopSettingsBody
	if err := json.Unmarshal(afterA.Body.Bytes(), &bodyAfterA); err != nil {
		t.Fatalf("decode shopA response after: %v", err)
	}
	if bodyAfterA.Name != "Solo A debe cambiar" {
		t.Fatalf("expected shopA's own update to apply, got %+v", bodyAfterA)
	}
}

// TestSettings_HTTP_NoCookie_Returns401Uniform confirma que estas rutas
// comparten el mismo 401 uniforme que el resto de /api/v1/private (ya
// cubierto estructuralmente por TestPrivateRouteInventory_...).
func TestSettings_HTTP_NoCookie_Returns401Uniform(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	if rec := doGetSettingsRequest(router, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for GET without cookie, got %d", rec.Code)
	}
	if rec := doPatchSettingsRequest(router, "", barbershopSettingsBody{Name: "x", Timezone: "America/Bogota"}); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for PATCH without cookie, got %d", rec.Code)
	}
}

// TestSettings_HTTP_InvalidTimezone_Returns422WithoutPartialWrite cubre
// CA-020-03 a nivel HTTP/PostgreSQL real.
func TestSettings_HTTP_InvalidTimezone_Returns422WithoutPartialWrite(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	restoreBarbershopRow(t, db, shopA, "Barbería de prueba A")

	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "settings-invalid-tz")

	rec := doPatchSettingsRequest(router, raw, barbershopSettingsBody{
		Name: "Nombre que no debe guardarse", Timezone: "COT",
	})
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}

	getRec := doGetSettingsRequest(router, raw)
	var body barbershopSettingsBody
	if err := json.Unmarshal(getRec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode GET response: %v", err)
	}
	if body.Name != "Barbería de prueba A" {
		t.Fatalf("CA-020-03: an invalid timezone must not write anything, got name=%q", body.Name)
	}
}

// TestSettings_HTTP_UnknownField_Returns400 cubre CA-020-05: un intento de
// enviar barbershopId en el cuerpo se rechaza como forma inválida, nunca
// se interpreta como el tenant.
func TestSettings_HTTP_UnknownField_Returns400(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "settings-unknown-field")

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/private/settings/barbershop",
		bytes.NewReader([]byte(`{"name":"x","timezone":"America/Bogota","contactEmail":"","contactPhone":"","barbershopId":"`+shopB+`"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: cookieName, Value: raw})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestSettings_HTTP_ThroughFullRouter_NeverLogsContactInfo refuerza
// RN-DAT-02 para estas rutas específicas: ni el nombre, ni el correo ni el
// teléfono de contacto aparecen en el log estructurado de una operación
// exitosa.
func TestSettings_HTTP_ThroughFullRouter_NeverLogsContactInfo(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
	router, err := buildRouter(db, logger, testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	restoreBarbershopRow(t, db, shopA, "Barbería de prueba A")

	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "settings-log-safe")
	email := "no-debe-aparecer-en-logs@ejemplo.test"
	phone := "+573009998877"
	rec := doPatchSettingsRequest(router, raw, barbershopSettingsBody{
		Name: "Nombre no sensible", Timezone: "America/Bogota", ContactEmail: &email, ContactPhone: &phone,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	logLine := logBuf.String()
	for _, forbidden := range []string{email, phone, raw, cookieName + "="} {
		if strings.Contains(logLine, forbidden) {
			t.Fatalf("RN-DAT-02: sensitive data leaked into the log: %s", logLine)
		}
	}
}
