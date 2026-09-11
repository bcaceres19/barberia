// Pruebas de integración de HU-090 contra el router REAL de producción
// (buildRouter) y PostgreSQL real, mismo patrón que
// settings_integration_test.go (HU-020). Requieren todas las migraciones
// aplicadas (incluida 20260911045044_add_barbershop_public_slug.sql) y
// database/testdata/hu090_reserva_publica.sql cargado.
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./cmd/api/...
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	publicShopSlugUno = "barberia-hu090-uno"
	publicShopSlugDos = "barberia-hu090-dos"
	publicShopNameUno = "Barbería de prueba HU-090 Uno"
)

type publicBarbershopProfileBody struct {
	Name         string  `json:"name"`
	Timezone     string  `json:"timezone"`
	ContactEmail *string `json:"contactEmail"`
	ContactPhone *string `json:"contactPhone"`
}

func doResolvePublicBarbershopRequest(router http.Handler, slug string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/barbershops/"+slug, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// TestPublicBooking_HTTP_ValidSlug_ReturnsPublicProfile cubre CA-090-01 a
// nivel HTTP/PostgreSQL real, SIN ninguna cookie de sesión: la resolución
// pública abre el contexto exacto de la barbería.
func TestPublicBooking_HTTP_ValidSlug_ReturnsPublicProfile(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	rec := doResolvePublicBarbershopRequest(router, publicShopSlugUno)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body publicBarbershopProfileBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body.Name != publicShopNameUno || body.Timezone != "America/Bogota" {
		t.Fatalf("unexpected response: %+v", body)
	}
	if body.ContactEmail == nil || *body.ContactEmail != "contacto.hu090uno@ejemplo.test" {
		t.Fatalf("expected contactEmail from the fixture, got %v", body.ContactEmail)
	}
}

// TestPublicBooking_HTTP_UnknownSlug_ReturnsUniform404Problem cubre
// CA-090-02: un enlace inexistente responde el envelope RFC 9457 uniforme
// del resto del contrato, con X-Request-Id, sin distinguir la causa.
func TestPublicBooking_HTTP_UnknownSlug_ReturnsUniform404Problem(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	rec := doResolvePublicBarbershopRequest(router, "barberia-que-no-existe")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("expected Content-Type application/problem+json, got %q", ct)
	}
	if rec.Header().Get("X-Request-Id") == "" {
		t.Fatal("expected X-Request-Id on the problem response")
	}
}

// TestPublicBooking_HTTP_MalformedSlug_ReturnsSameUniform404AsUnknown cubre
// CA-090-02 punta a punta: la forma HTTP de un slug imposible es
// indistinguible de la de un slug desconocido.
func TestPublicBooking_HTTP_MalformedSlug_ReturnsSameUniform404AsUnknown(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	known := doResolvePublicBarbershopRequest(router, "barberia-que-no-existe")
	malformed := doResolvePublicBarbershopRequest(router, "Mal%20Formado!!!")

	if known.Code != malformed.Code {
		t.Fatalf("expected the same status for unknown and malformed slugs, got %d vs %d", known.Code, malformed.Code)
	}

	// instance/requestId varían por solicitud a propósito (correlación de
	// logs, RN-DAT-02): la comparación relevante es sobre el resto del
	// envelope RFC 9457, que debe ser idéntico para ambas causas
	// (CA-090-02).
	type problemBody struct {
		Type   string `json:"type"`
		Title  string `json:"title"`
		Status int    `json:"status"`
		Detail string `json:"detail"`
		Code   string `json:"code"`
	}
	var knownProblem, malformedProblem problemBody
	if err := json.NewDecoder(known.Body).Decode(&knownProblem); err != nil {
		t.Fatalf("decode known-slug problem body: %v", err)
	}
	if err := json.NewDecoder(malformed.Body).Decode(&malformedProblem); err != nil {
		t.Fatalf("decode malformed-slug problem body: %v", err)
	}
	if knownProblem != malformedProblem {
		t.Fatalf("expected the same problem envelope (minus instance/requestId) for unknown and malformed slugs, got %+v vs %+v",
			knownProblem, malformedProblem)
	}
}

// TestPublicBooking_HTTP_TwoTenants_NeverCrossesBarbershops confirma
// RN-TEN-01 a nivel HTTP: resolver el slug de una barbería nunca trae ni
// insinúa datos de la otra.
func TestPublicBooking_HTTP_TwoTenants_NeverCrossesBarbershops(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	uno := doResolvePublicBarbershopRequest(router, publicShopSlugUno)
	dos := doResolvePublicBarbershopRequest(router, publicShopSlugDos)
	if uno.Code != http.StatusOK || dos.Code != http.StatusOK {
		t.Fatalf("expected both slugs to resolve, got %d and %d", uno.Code, dos.Code)
	}
	if strings.Contains(dos.Body.String(), "hu090uno") {
		t.Fatalf("RN-TEN-01: resolving slugDos leaked shopUno's contact: %s", dos.Body.String())
	}
}

// TestPublicBooking_HTTP_NoSessionCookieRequired confirma CA-090-03/DEC-034:
// la operación nunca pasa por SessionMiddleware -está registrada sobre el
// router público completo, nunca sobre `private`- así que una solicitud sin
// ninguna cookie de sesión igual resuelve con éxito.
func TestPublicBooking_HTTP_NoSessionCookieRequired(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/barbershops/"+publicShopSlugUno, nil)
	// Deliberadamente sin ninguna cookie.
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 without any session cookie, got %d: %s", rec.Code, rec.Body.String())
	}
}
