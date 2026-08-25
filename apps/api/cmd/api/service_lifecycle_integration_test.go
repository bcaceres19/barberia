// Pruebas de integración de HU-024 contra el router REAL de producción
// (buildRouter) y PostgreSQL real, mismo patrón que
// catalog_integration_test.go (HU-022). Requieren las doce migraciones
// aplicadas (incluida 20260824150000_create_barber_service.sql) y
// database/testdata/dos_barberias.sql + testdata/hu005_credenciales_sesiones.sql
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

type deactivationImpactBody struct {
	AffectedAppointments int `json:"affectedAppointments"`
}

type deactivationBody struct {
	Service              serviceBody `json:"service"`
	AffectedAppointments int         `json:"affectedAppointments"`
}

func doGetDeactivationImpactRequest(router http.Handler, rawToken, serviceID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/private/services/"+serviceID+"/deactivation-impact", nil)
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func doDeactivateServiceRequest(router http.Handler, rawToken, serviceID, idempotencyKey string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/private/services/"+serviceID+"/deactivate", nil)
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

func doReactivateServiceRequest(router http.Handler, rawToken, serviceID, idempotencyKey string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/private/services/"+serviceID+"/reactivate", nil)
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

// TestServiceLifecycle_HTTP_PreviewDeactivateReactivate_FullJourney recorre
// previsualizar (CA-024-01), desactivar (CA-024-02, CA-024-04), repetir con
// la misma clave (CA-024-06), rechazar una clave nueva sobre una transición
// que ya no aplica y reactivar (CA-024-05), a través del router real.
func TestServiceLifecycle_HTTP_PreviewDeactivateReactivate_FullJourney(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "lifecycle-journey-a")

	created := createServiceViaRouter(t, router, raw, "Servicio de ciclo de vida "+uniqueToken(t, "lifecycle"))

	// CA-024-01: previsualización real, siempre 0 en B1 (DEC-069).
	impactRec := doGetDeactivationImpactRequest(router, raw, created.ID)
	if impactRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on preview, got %d: %s", impactRec.Code, impactRec.Body.String())
	}
	var impact deactivationImpactBody
	if err := json.Unmarshal(impactRec.Body.Bytes(), &impact); err != nil {
		t.Fatalf("decode preview response: %v", err)
	}
	if impact.AffectedAppointments != 0 {
		t.Fatalf("DEC-069: expected affectedAppointments=0, got %d", impact.AffectedAppointments)
	}

	// CA-024-02/CA-024-04: confirmar la desactivación.
	deactivateKey := uniqueToken(t, "lifecycle-deactivate-key")
	deactivateRec := doDeactivateServiceRequest(router, raw, created.ID, deactivateKey)
	if deactivateRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on deactivate, got %d: %s", deactivateRec.Code, deactivateRec.Body.String())
	}
	var deactivated deactivationBody
	if err := json.Unmarshal(deactivateRec.Body.Bytes(), &deactivated); err != nil {
		t.Fatalf("decode deactivate response: %v", err)
	}
	if deactivated.Service.IsActive {
		t.Fatal("expected isActive=false after deactivate")
	}
	if deactivated.Service.DeactivatedAt == nil {
		t.Fatal("expected deactivatedAt set after deactivate")
	}
	if deactivated.AffectedAppointments != 0 {
		t.Fatalf("expected affectedAppointments=0 on confirmation, got %d", deactivated.AffectedAppointments)
	}
	if deactivated.Service.DurationMinutes != created.DurationMinutes || deactivated.Service.Price != created.Price {
		t.Fatalf("CA-024-02: deactivate must not alter duration/price, got %+v", deactivated.Service)
	}

	// CA-024-06: repetir con la MISMA clave reproduce la misma respuesta.
	replayRec := doDeactivateServiceRequest(router, raw, created.ID, deactivateKey)
	if replayRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on replay, got %d: %s", replayRec.Code, replayRec.Body.String())
	}
	if replayRec.Body.String() != deactivateRec.Body.String() {
		t.Fatalf("CA-004-01: expected byte-identical replay, got %q vs %q", replayRec.Body.String(), deactivateRec.Body.String())
	}

	// Una clave NUEVA sobre una transición que ya no aplica es un conflicto.
	conflictRec := doDeactivateServiceRequest(router, raw, created.ID, uniqueToken(t, "lifecycle-deactivate-key-2"))
	if conflictRec.Code != http.StatusConflict {
		t.Fatalf("expected 409 deactivating an already-inactive service with a new key, got %d: %s", conflictRec.Code, conflictRec.Body.String())
	}

	// GET independiente confirma la persistencia real.
	afterDeactivateRec := doGetServiceRequest(router, raw, created.ID)
	var afterDeactivate serviceBody
	if err := json.Unmarshal(afterDeactivateRec.Body.Bytes(), &afterDeactivate); err != nil {
		t.Fatalf("decode reloaded get response: %v", err)
	}
	if afterDeactivate.IsActive || afterDeactivate.DeactivatedAt == nil {
		t.Fatalf("expected the deactivation to persist across an independent GET, got %+v", afterDeactivate)
	}

	// Aparece en el listado con isActive=false, sin desaparecer (RN-SER-03).
	listRec := doListServicesRequest(router, raw, "limit=50")
	var page serviceListBody
	if err := json.Unmarshal(listRec.Body.Bytes(), &page); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	foundInactive := false
	for _, item := range page.Items {
		if item.ID == created.ID {
			foundInactive = true
			if item.IsActive {
				t.Fatalf("expected the listed item to reflect isActive=false, got %+v", item)
			}
		}
	}
	if !foundInactive {
		t.Fatal("RN-SER-03: expected the deactivated service to still appear in the list (no physical delete)")
	}

	// CA-024-05: reactivar vuelve a activo sin crear otra fila.
	reactivateRec := doReactivateServiceRequest(router, raw, created.ID, uniqueToken(t, "lifecycle-reactivate-key"))
	if reactivateRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on reactivate, got %d: %s", reactivateRec.Code, reactivateRec.Body.String())
	}
	var reactivated serviceBody
	if err := json.Unmarshal(reactivateRec.Body.Bytes(), &reactivated); err != nil {
		t.Fatalf("decode reactivate response: %v", err)
	}
	if !reactivated.IsActive || reactivated.DeactivatedAt != nil {
		t.Fatalf("expected isActive=true and deactivatedAt=nil after reactivate, got %+v", reactivated)
	}
	if reactivated.ID != created.ID || reactivated.DurationMinutes != created.DurationMinutes || reactivated.Price != created.Price {
		t.Fatalf("CA-024-05: reactivate must be the SAME row, unchanged, got %+v", reactivated)
	}

	// Reactivar de nuevo (ya activo) con una clave nueva es un conflicto.
	reactivateConflictRec := doReactivateServiceRequest(router, raw, created.ID, uniqueToken(t, "lifecycle-reactivate-key-2"))
	if reactivateConflictRec.Code != http.StatusConflict {
		t.Fatalf("expected 409 reactivating an already-active service with a new key, got %d: %s", reactivateConflictRec.Code, reactivateConflictRec.Body.String())
	}
}

// TestServiceLifecycle_HTTP_TwoTenants_CrossAccessReturns404WithoutLeaking
// cubre CA-024-07: conocer el UUID real de un servicio de otra barbería
// nunca permite previsualizar, desactivar ni reactivar, y nunca cambia su
// estado real.
func TestServiceLifecycle_HTTP_TwoTenants_CrossAccessReturns404WithoutLeaking(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	rawA := createSessionCookie(t, db, shopA, staffUserActiveA, "lifecycle-cross-a")
	rawB := createSessionCookie(t, db, shopB, staffUserActiveB, "lifecycle-cross-b")

	serviceA := createServiceViaRouter(t, router, rawA, "De la barbería A "+uniqueToken(t, "lifecycle-cross"))

	if rec := doGetDeactivationImpactRequest(router, rawB, serviceA.ID); rec.Code != http.StatusNotFound {
		t.Fatalf("CA-024-07: expected 404 previewing shopA's service from shopB, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec := doDeactivateServiceRequest(router, rawB, serviceA.ID, uniqueToken(t, "lifecycle-cross-key")); rec.Code != http.StatusNotFound {
		t.Fatalf("CA-024-07: expected 404 deactivating shopA's service from shopB, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec := doReactivateServiceRequest(router, rawB, serviceA.ID, uniqueToken(t, "lifecycle-cross-key-2")); rec.Code != http.StatusNotFound {
		t.Fatalf("CA-024-07: expected 404 reactivating shopA's service from shopB, got %d: %s", rec.Code, rec.Body.String())
	}

	// El servicio de A sigue activo e intacto.
	confirmRec := doGetServiceRequest(router, rawA, serviceA.ID)
	var confirm serviceBody
	if err := json.Unmarshal(confirmRec.Body.Bytes(), &confirm); err != nil {
		t.Fatalf("decode confirm response: %v", err)
	}
	if !confirm.IsActive || confirm.DeactivatedAt != nil {
		t.Fatalf("expected shopA's service to remain untouched and active, got %+v", confirm)
	}
}

func TestServiceLifecycle_HTTP_MissingIdempotencyKey_Returns400(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "lifecycle-missing-key")
	created := createServiceViaRouter(t, router, raw, "Sin clave "+uniqueToken(t, "lifecycle-missing"))

	if rec := doDeactivateServiceRequest(router, raw, created.ID, ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 without Idempotency-Key, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec := doReactivateServiceRequest(router, raw, created.ID, ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 without Idempotency-Key, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestServiceLifecycle_HTTP_UnknownServiceID_Returns404(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "lifecycle-unknown")
	const unknownID = "00000000-0000-0000-0000-000000000000"

	if rec := doGetDeactivationImpactRequest(router, raw, unknownID); rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 previewing an unknown service, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec := doDeactivateServiceRequest(router, raw, unknownID, uniqueToken(t, "lifecycle-unknown-key")); rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 deactivating an unknown service, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec := doReactivateServiceRequest(router, raw, unknownID, uniqueToken(t, "lifecycle-unknown-key-2")); rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 reactivating an unknown service, got %d: %s", rec.Code, rec.Body.String())
	}
}
