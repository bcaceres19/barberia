// Pruebas de integración de HU-012 (DEC-060) contra el router REAL de
// producción (buildRouter) y PostgreSQL real, mismo patrón que
// session_integration_test.go (HU-006). Requieren las seis migraciones
// aplicadas y database/testdata/dos_barberias.sql +
// testdata/hu005_credenciales_sesiones.sql cargados.
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

	"system-barbershop/internal/modules/auth"
)

func doSessionContextRequest(router http.Handler, rawToken string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/private/auth/session", nil)
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// TestSessionContext_HTTP_ValidSession_ReturnsRealBarbershopName cubre
// CA-012-01/CA-012-04 a nivel HTTP/PostgreSQL real: el nombre devuelto es la
// columna real de la barbería propietaria de la sesión, nunca un valor
// fijo, y la respuesta no fija Set-Cookie (es una lectura, no una mutación,
// a diferencia de login/logout).
func TestSessionContext_HTTP_ValidSession_ReturnsRealBarbershopName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	router, err := buildRouter(db, discardLogger())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "session-context-a")
	rec := doSessionContextRequest(router, raw)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatal("expected no Set-Cookie on a session-context read")
	}

	var body struct {
		Barbershop struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"barbershop"`
		ExpiresAt string `json:"expiresAt"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v: %s", err, rec.Body.String())
	}
	if body.Barbershop.ID != shopA {
		t.Fatalf("expected barbershop.id=%q, got %q", shopA, body.Barbershop.ID)
	}
	if body.Barbershop.Name != "Barbería de prueba A" {
		t.Fatalf("expected barbershop.name=%q (real column, database/testdata/dos_barberias.sql), got %q",
			"Barbería de prueba A", body.Barbershop.Name)
	}
	if body.ExpiresAt == "" {
		t.Fatal("expected a non-empty expiresAt")
	}
}

// TestSessionContext_HTTP_TwoTenants_NeverCrossesBarbershopNames es la
// prueba de aislamiento por tenant exigida por el prompt: la sesión de A
// nunca ve el nombre de B ni viceversa, contra el endpoint real con dos
// tenants reales (mismo patrón que TestLogout_HTTP_SessionOfShopA_NeverExecutesShopBsLogout).
func TestSessionContext_HTTP_TwoTenants_NeverCrossesBarbershopNames(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	router, err := buildRouter(db, discardLogger())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	rawA := createSessionCookie(t, db, shopA, staffUserActiveA, "context-tenant-a")
	rawB := createSessionCookie(t, db, shopB, staffUserActiveB, "context-tenant-b")

	recA := doSessionContextRequest(router, rawA)
	recB := doSessionContextRequest(router, rawB)

	var bodyA, bodyB struct {
		Barbershop struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"barbershop"`
	}
	if err := json.Unmarshal(recA.Body.Bytes(), &bodyA); err != nil {
		t.Fatalf("decode shopA response: %v", err)
	}
	if err := json.Unmarshal(recB.Body.Bytes(), &bodyB); err != nil {
		t.Fatalf("decode shopB response: %v", err)
	}

	if bodyA.Barbershop.ID != shopA || bodyA.Barbershop.Name == bodyB.Barbershop.Name {
		t.Fatalf("RN-TEN-01: expected distinct barbershop identities, got A=%+v B=%+v", bodyA, bodyB)
	}
	if bodyB.Barbershop.ID != shopB {
		t.Fatalf("expected shopB's own id, got %q", bodyB.Barbershop.ID)
	}
}

// TestSessionContext_HTTP_NoCookie_Returns401Uniform confirma que esta
// ruta comparte el mismo 401 uniforme que el resto de /api/v1/private
// (ya cubierto estructuralmente por TestPrivateRouteInventory_..., esta
// prueba deja además explícito el caso concreto de este endpoint).
func TestSessionContext_HTTP_NoCookie_Returns401Uniform(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	router, err := buildRouter(db, discardLogger())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	rec := doSessionContextRequest(router, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestSessionContext_HTTP_RevokedSession_Returns401AndNeverExposesName
// confirma que cerrar sesión invalida también esta lectura: el nombre de la
// barbería nunca se filtra a través de una cookie ya revocada.
func TestSessionContext_HTTP_RevokedSession_Returns401AndNeverExposesName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	router, err := buildRouter(db, discardLogger())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "context-revoked")

	if rec := doLogoutRequest(router, raw); rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 revoking the session first, got %d: %s", rec.Code, rec.Body.String())
	}

	rec := doSessionContextRequest(router, raw)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for a revoked session, got %d: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "Barbería") {
		t.Fatalf("a revoked session must never expose the barbershop name: %s", rec.Body.String())
	}
}

// TestSessionContext_HTTP_ThroughFullRouter_NeverLogsSessionMaterial refuerza
// CA-006-05 para esta ruta específica, mismo patrón que
// TestPrivateRoute_ThroughFullRouter_NeverLogsSessionMaterial.
func TestSessionContext_HTTP_ThroughFullRouter_NeverLogsSessionMaterial(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
	router, err := buildRouter(db, logger)
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "context-log-safe")
	rec := doSessionContextRequest(router, raw)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	logLine := logBuf.String()
	for _, forbidden := range []string{raw, auth.HashToken(raw), cookieName + "="} {
		if strings.Contains(logLine, forbidden) {
			t.Fatalf("CA-006-05: session material leaked into the log: %s", logLine)
		}
	}
}
