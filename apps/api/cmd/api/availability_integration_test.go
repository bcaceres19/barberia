// Pruebas de integración de HU-094 contra el router REAL de producción
// (buildRouter) y PostgreSQL real, mismo patrón que
// public_booking_integration_test.go (HU-090/HU-091/HU-092). Requieren
// todas las migraciones aplicadas y
// database/testdata/hu094_disponibilidad.sql cargado.
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./cmd/api/...
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	availShopSlugUno  = "barberia-hu094-uno"
	availShopSlugDos  = "barberia-hu094-dos"
	availServiceIDUno = "00940101-0094-0094-0094-009401010101"
	availBarberIDUno  = "00940011-0094-0094-0094-009400110011"
	availBarberIDDos  = "00940012-0094-0094-0094-009400120012"
)

type availabilitySlotBody struct {
	StartsAt time.Time `json:"startsAt"`
}

type availabilityResponseBody struct {
	Slots           []availabilitySlotBody `json:"slots"`
	DurationMinutes int                    `json:"durationMinutes"`
	Timezone        string                 `json:"timezone"`
	SlotGridMinutes int                    `json:"slotGridMinutes"`
}

func doPublicAvailabilityRequest(router http.Handler, slug, serviceID, barberID string) *httptest.ResponseRecorder {
	url := "/api/v1/public/barbershops/" + slug + "/services/" + serviceID + "/barbers/" + barberID + "/availability"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// TestAvailability_HTTP_ValidSelection_ReturnsOrderedSlots cubre CA-094-01/
// CA-094-04 a nivel HTTP/PostgreSQL real: el fixture deja jornada 9:00-18:00
// los siete días ISO, sin bloqueos ni citas, anticipación 0 y ventana de 7
// días, así que SIEMPRE debe haber al menos una franja futura sin importar
// cuándo corra la prueba.
func TestAvailability_HTTP_ValidSelection_ReturnsOrderedSlots(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	before := time.Now().UTC()
	rec := doPublicAvailabilityRequest(router, availShopSlugUno, availServiceIDUno, availBarberIDUno)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body availabilityResponseBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body.DurationMinutes != 30 {
		t.Fatalf("durationMinutes = %d, want 30", body.DurationMinutes)
	}
	if body.Timezone != "America/Bogota" {
		t.Fatalf("timezone = %q, want America/Bogota", body.Timezone)
	}
	if body.SlotGridMinutes != 30 {
		t.Fatalf("slotGridMinutes = %d, want 30", body.SlotGridMinutes)
	}
	if len(body.Slots) == 0 {
		t.Fatal("esperaba al menos una franja (jornada 7 días sin ocupaciones)")
	}
	for i, slot := range body.Slots {
		if slot.StartsAt.Before(before) {
			t.Fatalf("franja %d = %v, anterior al instante de la solicitud %v (RN-DIS-03/CA-094-04)", i, slot.StartsAt, before)
		}
		if i > 0 && !slot.StartsAt.After(body.Slots[i-1].StartsAt) {
			t.Fatalf("orden no estrictamente creciente en índice %d: %+v", i, body.Slots)
		}
	}
}

// TestAvailability_HTTP_UnknownSlug_ReturnsUniform404Problem cubre
// CA-090-02 reutilizado por HU-094.
func TestAvailability_HTTP_UnknownSlug_ReturnsUniform404Problem(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	rec := doPublicAvailabilityRequest(router, "barberia-que-no-existe", availServiceIDUno, availBarberIDUno)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("expected Content-Type application/problem+json, got %q", ct)
	}
}

// TestAvailability_HTTP_ForeignOrMalformedIDs_ReturnsEmptySlotsNotError
// cubre el mismo criterio que CA-092-03: un serviceId/barberId con forma
// inválida, ajeno o sin asignación vigente nunca produce un error
// distinto, siempre una disponibilidad vacía.
func TestAvailability_HTTP_ForeignOrMalformedIDs_ReturnsEmptySlotsNotError(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	tests := []struct {
		name      string
		slug      string
		serviceID string
		barberID  string
	}{
		{"serviceId sin forma de UUID", availShopSlugUno, "no-es-uuid", availBarberIDUno},
		{"barberId sin forma de UUID", availShopSlugUno, availServiceIDUno, "no-es-uuid"},
		{"barberId de otra barbería (CA-094-05, aislamiento de tenant)", availShopSlugUno, availServiceIDUno, availBarberIDDos},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := doPublicAvailabilityRequest(router, tc.slug, tc.serviceID, tc.barberID)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			}
			var body availabilityResponseBody
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("decode response body: %v", err)
			}
			if len(body.Slots) != 0 {
				t.Fatalf("expected empty slots, got %+v", body.Slots)
			}
		})
	}
}

// TestAvailability_HTTP_TwoTenants_NeverCrossesBarbershops confirma
// RN-TEN-01/CA-094-05: la disponibilidad de un barbero real de la barbería
// Dos nunca aparece al consultarla bajo el slug de la barbería Uno, y
// viceversa (ambas resuelven con éxito bajo su propio slug).
func TestAvailability_HTTP_TwoTenants_NeverCrossesBarbershops(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	uno := doPublicAvailabilityRequest(router, availShopSlugUno, availServiceIDUno, availBarberIDUno)
	crossed := doPublicAvailabilityRequest(router, availShopSlugDos, availServiceIDUno, availBarberIDUno)
	if uno.Code != http.StatusOK || crossed.Code != http.StatusOK {
		t.Fatalf("expected both requests to resolve 200, got %d and %d", uno.Code, crossed.Code)
	}

	var unoBody, crossedBody availabilityResponseBody
	if err := json.NewDecoder(uno.Body).Decode(&unoBody); err != nil {
		t.Fatalf("decode uno body: %v", err)
	}
	if err := json.NewDecoder(crossed.Body).Decode(&crossedBody); err != nil {
		t.Fatalf("decode crossed body: %v", err)
	}
	if len(unoBody.Slots) == 0 {
		t.Fatal("esperaba disponibilidad real bajo el slug propio de la barbería Uno")
	}
	if len(crossedBody.Slots) != 0 {
		t.Fatalf("RN-TEN-01: el barbero/servicio de la barbería Uno produjo disponibilidad bajo el slug de la barbería Dos: %+v", crossedBody.Slots)
	}
}

// TestAvailability_HTTP_NoSessionCookieRequired confirma que esta operación
// está registrada sobre el router público completo, nunca sobre `private`
// (mismo criterio que HU-090/HU-091/HU-092).
func TestAvailability_HTTP_NoSessionCookieRequired(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/barbershops/"+availShopSlugUno+"/services/"+availServiceIDUno+"/barbers/"+availBarberIDUno+"/availability", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 without any session cookie, got %d: %s", rec.Code, rec.Body.String())
	}
}
