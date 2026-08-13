// Pruebas de integración de HU-006 contra el router REAL de producción
// (buildRouter) y PostgreSQL real. Requieren las seis migraciones
// aplicadas y database/testdata/dos_barberias.sql +
// testdata/hu005_credenciales_sesiones.sql cargados, igual que
// internal/modules/auth/postgres/*_test.go.
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./cmd/api/...
package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"system-barbershop/internal/modules/auth"
	authpostgres "system-barbershop/internal/modules/auth/postgres"
	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
)

const (
	testDatabaseURL = "postgres://barberia_app@localhost:5432/barberia_test?sslmode=disable"

	shopA = "11111111-1111-1111-1111-111111111111"
	shopB = "22222222-2222-2222-2222-222222222222"

	staffUserActiveA = "aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaa1"
	staffUserActiveB = "bbbbbbb1-bbbb-bbbb-bbbb-bbbbbbbbbbb1"
)

func setupTestDB(t *testing.T) *database.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = testDatabaseURL
	}
	cfg := config.Config{
		Environment:              "test",
		DatabaseMaxConns:         10,
		DatabaseMinConns:         2,
		DatabaseMaxConnLifetime:  time.Hour,
		DatabaseMaxConnIdleTime:  30 * time.Minute,
		DatabaseConnectTimeout:   5 * time.Second,
		DatabaseStatementTimeout: 10 * time.Second,
	}
	db, err := database.NewDB(config.DatabaseDSN(url), cfg)
	if err != nil {
		t.Fatalf("database.NewDB: %v", err)
	}
	return db
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))
}

func uniqueToken(t *testing.T, label string) string {
	t.Helper()
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	return "test-" + label + "-" + hex.EncodeToString(buf)
}

// createSessionCookie inserta una sesión real (vía el mismo repositorio que
// usa buildRouter) y devuelve el valor EN CLARO para usarlo como cookie en
// una solicitud de prueba.
func createSessionCookie(t *testing.T, db *database.DB, shop, staffUserID, label string) string {
	t.Helper()
	repo := authpostgres.New(db)
	raw := uniqueToken(t, label)
	now := time.Now().UTC()
	if err := repo.CreateSession(t.Context(), shop, staffUserID, auth.HashToken(raw), now, now.Add(auth.SessionDuration)); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	return raw
}

const cookieName = "barberia_session"

func doLogoutRequest(router *chi.Mux, rawToken string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/private/auth/logout", nil)
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// TestPrivateRouteInventory_AllRegisteredRoutesRequireSession es la prueba
// estructural de CA-006-04: camina el router REAL de producción
// (buildRouter, el mismo que run() usa) con chi.Walk y confirma, para
// CADA ruta registrada bajo /api/v1/private -sin una lista manual
// desconectada del router-, que llamarla sin cookie responde el mismo 401
// uniforme. Si alguien registrara una ruta privada nueva por fuera del
// subrouter protegido (con el patrón completo sobre el *chi.Mux, en vez de
// sobre `private`), chi.Walk igual la descubre -recorre TODO lo que el
// router responde, sin importar cómo se montó- pero la solicitud sin
// cookie no obtendría 401, y esta prueba fallaría (ver el control negativo
// en internal/platform/httpserver/router_test.go, que demuestra que esta
// técnica sí distingue una ruta protegida de una que no lo está).
func TestPrivateRouteInventory_AllRegisteredRoutesRequireSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	router, err := buildRouter(db, discardLogger())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	type route struct{ method, pattern string }
	var privateRoutes []route
	err = chi.Walk(router, func(method, pattern string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		if strings.HasPrefix(pattern, "/api/v1/private") {
			privateRoutes = append(privateRoutes, route{method, pattern})
		}
		return nil
	})
	if err != nil {
		t.Fatalf("chi.Walk: %v", err)
	}
	if len(privateRoutes) == 0 {
		t.Fatal("expected at least one private route (POST /api/v1/private/auth/logout) to be registered")
	}

	for _, rt := range privateRoutes {
		// chi.Walk devuelve el patrón con parámetros ({id}); ninguna ruta
		// privada actual tiene parámetros, pero se sustituye por un valor
		// cualquiera si algún día los hay, para no romper la prueba en vez
		// de fortalecerla.
		path := strings.NewReplacer("{", "", "}", "").Replace(rt.pattern)
		req := httptest.NewRequest(rt.method, path, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("CA-006-04: %s %s sin cookie devolvió %d (se esperaba 401 uniforme, la ruta escapó del middleware de sesión)",
				rt.method, rt.pattern, rec.Code)
			continue
		}
		var p struct {
			Code string `json:"code"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil || p.Code != "unauthorized" {
			t.Errorf("%s %s: expected the uniform unauthorized problem, got %s", rt.method, rt.pattern, rec.Body.String())
		}
	}
}

// TestLogout_HTTP_ReusedCookieAfterLogout_Returns401AndNeverRunsTwice cubre
// CA-006-02: revoca en el servidor y, al reutilizar el mismo material,
// obtiene el mismo 401 uniforme (no un error distinto que revele que ya se
// había cerrado esa sesión), y una segunda llamada nunca vuelve a ejecutar
// la operación privada (el middleware la detiene antes del handler).
func TestLogout_HTTP_ReusedCookieAfterLogout_Returns401AndNeverRunsTwice(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	router, err := buildRouter(db, discardLogger())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "logout-once")

	first := doLogoutRequest(router, raw)
	if first.Code != http.StatusNoContent {
		t.Fatalf("expected 204 on first logout, got %d: %s", first.Code, first.Body.String())
	}

	second := doLogoutRequest(router, raw)
	if second.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on reused material, got %d: %s", second.Code, second.Body.String())
	}
	var p struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(second.Body.Bytes(), &p); err != nil || p.Code != "unauthorized" {
		t.Fatalf("expected the uniform unauthorized problem, got %s", second.Body.String())
	}
}

// TestLogout_HTTP_ClosingOneDeviceDoesNotAffectAnother cubre CA-006-06: dos
// sesiones del MISMO usuario (dos "dispositivos"); cerrar una no invalida
// la otra.
func TestLogout_HTTP_ClosingOneDeviceDoesNotAffectAnother(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	router, err := buildRouter(db, discardLogger())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	deviceOne := createSessionCookie(t, db, shopA, staffUserActiveA, "device-one")
	deviceTwo := createSessionCookie(t, db, shopA, staffUserActiveA, "device-two")

	closeOne := doLogoutRequest(router, deviceOne)
	if closeOne.Code != http.StatusNoContent {
		t.Fatalf("expected 204 closing device one, got %d: %s", closeOne.Code, closeOne.Body.String())
	}

	// El dispositivo dos sigue autenticado: la prueba de esa autenticación
	// es que también puede cerrar SU propia sesión con éxito.
	closeTwo := doLogoutRequest(router, deviceTwo)
	if closeTwo.Code != http.StatusNoContent {
		t.Fatalf("CA-006-06: device two was invalidated by closing device one, got %d: %s", closeTwo.Code, closeTwo.Body.String())
	}
}

// TestLogout_HTTP_SessionOfShopA_NeverExecutesShopBsLogout es la prueba
// end-to-end de CA-006-07 (DEC-058): la sesión de la barbería A nunca
// ejecuta el logout (ni ninguna otra operación privada) de la barbería B,
// verificado contra el endpoint real con dos tenants reales. Completa
// CA-005-05/CA-005-01 de HU-005, que solo se probaron a nivel
// PostgreSQL/RLS por no existir todavía ningún endpoint privado real.
func TestLogout_HTTP_SessionOfShopA_NeverExecutesShopBsLogout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	router, err := buildRouter(db, discardLogger())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	rawA := createSessionCookie(t, db, shopA, staffUserActiveA, "tenant-a")
	rawB := createSessionCookie(t, db, shopB, staffUserActiveB, "tenant-b")

	// La sesión de A cierra sesión: solo puede afectar su propia fila, sin
	// que exista ningún canal en la API para apuntar a la sesión de B (el
	// contrato no acepta un identificador de sesión objetivo; el tenant se
	// deriva exclusivamente de la cookie ya autenticada).
	recA := doLogoutRequest(router, rawA)
	if recA.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for shopA's own logout, got %d: %s", recA.Code, recA.Body.String())
	}

	// La sesión de B nunca se vio afectada: sigue autenticada y puede
	// cerrar SU PROPIA sesión con éxito.
	recB := doLogoutRequest(router, rawB)
	if recB.Code != http.StatusNoContent {
		t.Fatalf("CA-006-07: shopB's session was affected by shopA's logout, got %d: %s", recB.Code, recB.Body.String())
	}

	// Ambas quedan revocadas ahora (cada una por su propia llamada), y
	// ninguna repite la operación.
	if rec := doLogoutRequest(router, rawA); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected shopA's session to already be revoked, got %d", rec.Code)
	}
	if rec := doLogoutRequest(router, rawB); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected shopB's session to already be revoked, got %d", rec.Code)
	}
}

// TestSession_HTTP_ReusedCookieOnFreshRequest_StaysAuthenticatedWithoutCredentials
// cubre CA-006-01: una cookie de sesión persistida sobrevive a "cerrar y
// reabrir el navegador" -modelado aquí como una solicitud completamente
// nueva (nuevo http.Request/ResponseRecorder, sin ningún estado de cliente
// salvo el valor de la cookie)- dentro de la vigencia, sin reenviar
// credenciales. La evidencia observable es que esa cookie sigue
// autorizando una operación privada real (logout) tal cual la dejó el
// login original.
func TestSession_HTTP_ReusedCookieOnFreshRequest_StaysAuthenticatedWithoutCredentials(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	router, err := buildRouter(db, discardLogger())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "persistent")

	// "Nuevo navegador": una solicitud completamente independiente de la
	// que creó la sesión, reutilizando solo el valor persistido de la
	// cookie.
	rec := doLogoutRequest(router, raw)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("CA-006-01: reused cookie on a fresh request was not authenticated, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestPrivateRoute_ThroughFullRouter_NeverLogsSessionMaterial refuerza
// CA-006-05 a nivel del router de producción ensamblado: ni el token en
// claro ni su hash aparecen en el log estructurado.
func TestPrivateRoute_ThroughFullRouter_NeverLogsSessionMaterial(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
	router, err := buildRouter(db, logger)
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "log-safe")
	rec := doLogoutRequest(router, raw)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	logLine := logBuf.String()
	for _, forbidden := range []string{raw, auth.HashToken(raw), cookieName + "="} {
		if strings.Contains(logLine, forbidden) {
			t.Fatalf("CA-006-05: session material leaked into the log: %s", logLine)
		}
	}
}
