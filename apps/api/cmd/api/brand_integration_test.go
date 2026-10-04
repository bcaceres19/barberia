// Pruebas de integración de marca y vocabulario (issue #292, DEC-110) contra
// el router REAL de producción (buildRouter) y PostgreSQL real, mismo patrón
// que settings_integration_test.go. Requieren la migración
// 20261003120000_add_barbershop_brand.sql aplicada y
// database/testdata/dos_barberias.sql + testdata/hu005_credenciales_sesiones.sql
// cargados.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"system-barbershop/internal/platform/database"
)

type brandBody struct {
	Accent                 string `json:"accent"`
	BusinessTerm           string `json:"businessTerm"`
	BusinessTermGender     string `json:"businessTermGender"`
	ProfessionalTerm       string `json:"professionalTerm"`
	ProfessionalTermPlural string `json:"professionalTermPlural"`
	ProfessionalTermGender string `json:"professionalTermGender"`
}

var initialBrand = brandBody{
	Accent: "brass", BusinessTerm: "barbería", BusinessTermGender: "feminine",
	ProfessionalTerm: "barbero", ProfessionalTermPlural: "barberos", ProfessionalTermGender: "masculine",
}

func doBrandRequest(handler http.Handler, method, rawToken string, body any) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, "/api/v1/private/settings/brand", reader)
	req.Header.Set("Content-Type", "application/json")
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

// restoreBrandRow devuelve la marca de shop a los valores iniciales al
// terminar la prueba: estas pruebas comparten las filas de dos_barberias.sql.
func restoreBrandRow(t *testing.T, db *database.DB, shop string) {
	t.Helper()
	t.Cleanup(func() {
		err := db.InTenantTx(context.Background(), database.BarbershopID(shop), func(ctx context.Context, q database.Queries) error {
			_, err := q.Exec(ctx,
				`UPDATE barbershop
				    SET brand_accent = 'brass', business_term = 'barbería', business_term_gender = 'feminine',
				        professional_term = 'barbero', professional_term_plural = 'barberos',
				        professional_term_gender = 'masculine'
				  WHERE id = $1`, shop)
			return err
		})
		if err != nil {
			t.Fatalf("restoreBrandRow cleanup: %v", err)
		}
	})
}

func decodeBrand(t *testing.T, rec *httptest.ResponseRecorder) brandBody {
	t.Helper()
	var body brandBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode brand response: %v: %s", err, rec.Body.String())
	}
	return body
}

// TestBrand_HTTP_UntouchedBarbershop_ReturnsTheInitialValues: la barbería que
// nunca cambió su marca recibe exactamente la interfaz anterior.
func TestBrand_HTTP_UntouchedBarbershop_ReturnsTheInitialValues(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	restoreBrandRow(t, db, shopA)

	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "brand-get-a")
	rec := doBrandRequest(router, http.MethodGet, raw, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeBrand(t, rec); got != initialBrand {
		t.Fatalf("expected the initial brand %+v, got %+v", initialBrand, got)
	}
}

// TestBrand_HTTP_PatchThenGet_NormalizesPersistsAndRoundTrips: el servidor
// normaliza los términos y una lectura independiente devuelve lo guardado.
func TestBrand_HTTP_PatchThenGet_NormalizesPersistsAndRoundTrips(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	restoreBrandRow(t, db, shopA)

	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "brand-patch-a")
	patch := doBrandRequest(router, http.MethodPatch, raw, brandBody{
		Accent: "emerald", BusinessTerm: "  Salón   de Belleza ", BusinessTermGender: "masculine",
		ProfessionalTerm: "Estilista", ProfessionalTermPlural: "Estilistas", ProfessionalTermGender: "feminine",
	})
	if patch.Code != http.StatusOK {
		t.Fatalf("expected 200 on PATCH, got %d: %s", patch.Code, patch.Body.String())
	}
	want := brandBody{
		Accent: "emerald", BusinessTerm: "salón de belleza", BusinessTermGender: "masculine",
		ProfessionalTerm: "estilista", ProfessionalTermPlural: "estilistas", ProfessionalTermGender: "feminine",
	}
	if got := decodeBrand(t, patch); got != want {
		t.Fatalf("PATCH must answer the normalized brand %+v, got %+v", want, got)
	}
	if got := decodeBrand(t, doBrandRequest(router, http.MethodGet, raw, nil)); got != want {
		t.Fatalf("GET must round-trip the saved brand %+v, got %+v", want, got)
	}
}

// TestBrand_HTTP_TwoTenants_PatchNeverCrossesBarbershops: guardar la marca de
// A, contra el endpoint real con dos tenants reales, nunca afecta a B.
func TestBrand_HTTP_TwoTenants_PatchNeverCrossesBarbershops(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	restoreBrandRow(t, db, shopA)
	restoreBrandRow(t, db, shopB)

	rawA := createSessionCookie(t, db, shopA, staffUserActiveA, "brand-tenant-a")
	rawB := createSessionCookie(t, db, shopB, staffUserActiveB, "brand-tenant-b")

	if rec := doBrandRequest(router, http.MethodPatch, rawA, brandBody{
		Accent: "ruby", BusinessTerm: "salón", BusinessTermGender: "masculine",
		ProfessionalTerm: "estilista", ProfessionalTermPlural: "estilistas", ProfessionalTermGender: "feminine",
	}); rec.Code != http.StatusOK {
		t.Fatalf("expected 200 updating shopA, got %d: %s", rec.Code, rec.Body.String())
	}

	if got := decodeBrand(t, doBrandRequest(router, http.MethodGet, rawB, nil)); got != initialBrand {
		t.Fatalf("updating shopA changed shopB's brand: %+v", got)
	}
}

func TestBrand_HTTP_InvalidValue_Returns422WithoutWriting(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	restoreBrandRow(t, db, shopA)

	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "brand-invalid")
	rec := doBrandRequest(router, http.MethodPatch, raw, brandBody{
		Accent: "#ff0000", BusinessTerm: "salón", BusinessTermGender: "masculine",
		ProfessionalTerm: "estilista", ProfessionalTermPlural: "estilistas", ProfessionalTermGender: "feminine",
	})
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeBrand(t, doBrandRequest(router, http.MethodGet, raw, nil)); got != initialBrand {
		t.Fatalf("a rejected PATCH must not write anything, got %+v", got)
	}
}

func TestBrand_HTTP_NoCookie_Returns401Uniform(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	if rec := doBrandRequest(router, http.MethodGet, "", nil); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for GET without cookie, got %d", rec.Code)
	}
	if rec := doBrandRequest(router, http.MethodPatch, "", initialBrand); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for PATCH without cookie, got %d", rec.Code)
	}
}
