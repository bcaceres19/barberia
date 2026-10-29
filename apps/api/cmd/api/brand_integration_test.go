// Pruebas de integración de marca y vocabulario (issue #292, DEC-110) contra
// el router REAL de producción (buildRouter) y PostgreSQL real, mismo patrón
// que settings_integration_test.go. Requieren la migración
// 20261003120000_add_barbershop_brand.sql y
// 20261029130000_create_staff_user_role.sql aplicadas y
// database/testdata/dos_barberias.sql + testdata/hu005_credenciales_sesiones.sql
// + testdata/staff_user_role_365.sql cargados.
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
	// Solo existe en la respuesta: el PATCH ya no lo admite (DEC-134), así que
	// omitempty evita enviarlo.
	PanelProfile string `json:"panelProfile,omitempty"`
}

var initialBrand = brandBody{
	Accent: "brass", BusinessTerm: "barbería", BusinessTermGender: "feminine",
	ProfessionalTerm: "barbero", ProfessionalTermPlural: "barberos", ProfessionalTermGender: "masculine",
	PanelProfile: "shop",
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
		PanelProfile: "shop",
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

// Pares de usuarios de database/testdata/staff_user_role_365.sql sobre el par de
// barberías dedicado a las pruebas de roles.
const (
	rolesShopA          = "12121212-1212-1212-1212-121212121212"
	rolesShopB          = "34343434-3434-3434-3434-343434343434"
	rolesUserEmpleado   = "365a0001-0000-4000-8000-000000000001"
	rolesUserMixto      = "365a0001-0000-4000-8000-000000000002"
	rolesUserIndep      = "365a0001-0000-4000-8000-000000000003"
	rolesUserEmpresa    = "365a0001-0000-4000-8000-000000000004"
	rolesUserSinRol     = "365a0001-0000-4000-8000-000000000005"
	rolesUserOtraBarber = "365a0001-0000-4000-8000-000000000006"
)

// TestBrand_HTTP_PanelProfile_ComesFromTheRolesOfTheSessionUser: el panel (DEC-134)
// lo deciden los roles de quien consulta, no la barbería: en la misma barbería
// cada usuario recibe el suyo, y el de otra barbería nunca se mezcla.
func TestBrand_HTTP_PanelProfile_ComesFromTheRolesOfTheSessionUser(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	cases := []struct {
		name string
		shop string
		user string
		want string
	}{
		{"empleado", rolesShopA, rolesUserEmpleado, "employee"},
		{"empleado e independiente a la vez", rolesShopA, rolesUserMixto, "solo"},
		{"independiente", rolesShopA, rolesUserIndep, "solo"},
		{"empresa", rolesShopA, rolesUserEmpresa, "shop"},
		{"sin rol conserva el panel completo", rolesShopA, rolesUserSinRol, "shop"},
		{"independiente de otra barbería", rolesShopB, rolesUserOtraBarber, "solo"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := createSessionCookie(t, db, tc.shop, tc.user, "brand-role-"+tc.user)
			rec := doBrandRequest(router, http.MethodGet, raw, nil)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			}
			if got := decodeBrand(t, rec).PanelProfile; got != tc.want {
				t.Fatalf("expected panelProfile %q, got %q", tc.want, got)
			}
		})
	}
}

// El PATCH ya no admite panelProfile: el contrato cerrado lo rechaza sin escribir.
func TestBrand_HTTP_PatchRejectsPanelProfile(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	restoreBrandRow(t, db, shopA)

	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "brand-profile-rejected")
	body := initialBrand
	body.PanelProfile = "solo"
	if rec := doBrandRequest(router, http.MethodPatch, raw, body); rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for panelProfile, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeBrand(t, doBrandRequest(router, http.MethodGet, raw, nil)); got != initialBrand {
		t.Fatalf("a rejected PATCH must not write anything, got %+v", got)
	}
}
