package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"system-barbershop/internal/modules/shops"
	"system-barbershop/internal/modules/shops/httpapi"
)

type brandUpdateCall struct {
	barbershopID string
	brand        shops.Brand
	location     shops.LocationWrite
}

// fakeBrandRepository es un doble en memoria de shops.BrandRepository: estas
// pruebas verifican el adaptador HTTP en aislamiento (decodificación, forma
// cerrada, principal, traducción de errores), no PostgreSQL real (esa
// cobertura vive en internal/modules/shops/postgres/brand_repository_test.go).
type fakeBrandRepository struct {
	getBrand shops.Brand
	getFound bool

	updateResult shops.BrandUpdateResult
	updateCalls  []brandUpdateCall
}

func (f *fakeBrandRepository) Get(_ context.Context, _ string) (shops.Brand, bool, error) {
	return f.getBrand, f.getFound, nil
}

func (f *fakeBrandRepository) Update(_ context.Context, barbershopID string, brand shops.Brand, location shops.LocationWrite) (shops.BrandUpdateResult, error) {
	f.updateCalls = append(f.updateCalls, brandUpdateCall{barbershopID, brand, location})
	return f.updateResult, nil
}

var _ shops.BrandRepository = (*fakeBrandRepository)(nil)

// fakePanelProfileRepository es un doble de shops.PanelProfileRepository: devuelve
// los roles fijados y recuerda a quién se los pidieron.
type fakePanelProfileRepository struct {
	roles []shops.Role
	calls [][2]string
}

func (f *fakePanelProfileRepository) RolesOf(_ context.Context, barbershopID, staffUserID string) ([]shops.Role, error) {
	f.calls = append(f.calls, [2]string{barbershopID, staffUserID})
	return f.roles, nil
}

var _ shops.PanelProfileRepository = (*fakePanelProfileRepository)(nil)

// profilesFor construye el servicio de panel con los roles dados.
func profilesFor(roles ...shops.Role) *shops.PanelProfileService {
	return shops.NewPanelProfileService(&fakePanelProfileRepository{roles: roles})
}

const validBrandBody = `{"accent":"emerald","businessTerm":"Salón","businessTermGender":"masculine","professionalTerm":"Estilista","professionalTermPlural":"estilistas","professionalTermGender":"feminine"}`

func TestGetBrandHandler_Success_ReturnsTheElevenFields(t *testing.T) {
	repo := &fakeBrandRepository{getFound: true, getBrand: shops.DefaultBrand}
	h := httpapi.NewGetBrandHandler(shops.NewBrandService(repo, nil), profilesFor())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, requestWithPrincipal(http.MethodGet, "/api/v1/private/settings/brand", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// La ubicación (DEC-130) viaja siempre: sin dirección ni coordenadas, con
	// null explícito, nunca omitida.
	want := map[string]any{
		"accent": "brass", "businessTerm": "barbería", "businessTermGender": "feminine",
		"professionalTerm": "barbero", "professionalTermPlural": "barberos", "professionalTermGender": "masculine",
		"panelProfile":        "shop",
		"hasPhysicalLocation": true, "address": nil, "latitude": nil, "longitude": nil,
	}
	if len(body) != len(want) {
		t.Fatalf("expected exactly %d fields, got %v", len(want), body)
	}
	for k, v := range want {
		got, present := body[k]
		if !present || got != v {
			t.Fatalf("field %s: expected %v, got %v (present=%v)", k, v, got, present)
		}
	}
	// El recurso nunca arrastra campos de HU-020 ni el identificador del tenant.
	raw := strings.ToLower(rec.Body.String())
	for _, forbidden := range []string{"barbershopid", "timezone", "contactemail"} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("response must never mention %q: %s", forbidden, raw)
		}
	}
}

func TestGetBrandHandler_NotFound_Returns404(t *testing.T) {
	h := httpapi.NewGetBrandHandler(shops.NewBrandService(&fakeBrandRepository{}, nil), profilesFor())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, requestWithPrincipal(http.MethodGet, "/api/v1/private/settings/brand", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetBrandHandler_MissingPrincipal_ReturnsSafe500(t *testing.T) {
	h := httpapi.NewGetBrandHandler(shops.NewBrandService(&fakeBrandRepository{}, nil), profilesFor())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/private/settings/brand", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateBrandHandler_Success_ReturnsSavedRepresentationAndUsesPrincipalTenant(t *testing.T) {
	saved := shops.Brand{
		Accent: "emerald", BusinessTerm: "salón", BusinessTermGender: shops.GenderMasculine,
		ProfessionalTerm: "estilista", ProfessionalTermPlural: "estilistas", ProfessionalTermGender: shops.GenderFeminine,
	}
	repo := &fakeBrandRepository{updateResult: shops.BrandUpdateResult{Brand: saved, Found: true}}
	h := httpapi.NewUpdateBrandHandler(shops.NewBrandService(repo, nil), profilesFor())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, requestWithPrincipal(http.MethodPatch, "/api/v1/private/settings/brand", []byte(validBrandBody)))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(repo.updateCalls) != 1 || repo.updateCalls[0].barbershopID != testShopID {
		t.Fatalf("expected one write for the principal's barbershop, got %+v", repo.updateCalls)
	}
	// El servicio normaliza ("Salón" → "salón") antes de llegar al repositorio.
	if repo.updateCalls[0].brand.BusinessTerm != "salón" || repo.updateCalls[0].brand.ProfessionalTerm != "estilista" {
		t.Fatalf("terms were not normalized before the write: %+v", repo.updateCalls[0].brand)
	}
	var body httpapi.BrandResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Accent != "emerald" || body.ProfessionalTermPlural != "estilistas" {
		t.Fatalf("unexpected response: %+v", body)
	}
}

func TestUpdateBrandHandler_UnknownField_Returns400(t *testing.T) {
	repo := &fakeBrandRepository{}
	h := httpapi.NewUpdateBrandHandler(shops.NewBrandService(repo, nil), profilesFor())

	body := strings.TrimSuffix(validBrandBody, "}") + `,"barbershopId":"22222222-2222-2222-2222-222222222222"}`
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, requestWithPrincipal(http.MethodPatch, "/api/v1/private/settings/brand", []byte(body)))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(repo.updateCalls) != 0 {
		t.Fatal("a tenant identifier in the body must never reach the repository")
	}
}

func TestUpdateBrandHandler_TrailingJSON_Returns400(t *testing.T) {
	repo := &fakeBrandRepository{}
	h := httpapi.NewUpdateBrandHandler(shops.NewBrandService(repo, nil), profilesFor())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, requestWithPrincipal(http.MethodPatch, "/api/v1/private/settings/brand", []byte(validBrandBody+`{}`)))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateBrandHandler_MalformedJSON_Returns400(t *testing.T) {
	h := httpapi.NewUpdateBrandHandler(shops.NewBrandService(&fakeBrandRepository{}, nil), profilesFor())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, requestWithPrincipal(http.MethodPatch, "/api/v1/private/settings/brand", []byte(`{no-json`)))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateBrandHandler_InvalidValue_Returns422WithoutWriting(t *testing.T) {
	cases := map[string]string{
		"acento fuera de la paleta": `{"accent":"#ff0000","businessTerm":"salón","businessTermGender":"masculine","professionalTerm":"estilista","professionalTermPlural":"estilistas","professionalTermGender":"feminine"}`,
		"término con dígitos":       `{"accent":"brass","businessTerm":"salón 24","businessTermGender":"masculine","professionalTerm":"estilista","professionalTermPlural":"estilistas","professionalTermGender":"feminine"}`,
		"género inválido":           `{"accent":"brass","businessTerm":"salón","businessTermGender":"x","professionalTerm":"estilista","professionalTermPlural":"estilistas","professionalTermGender":"feminine"}`,
		"campo omitido":             `{"accent":"brass"}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &fakeBrandRepository{}
			h := httpapi.NewUpdateBrandHandler(shops.NewBrandService(repo, nil), profilesFor())

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, requestWithPrincipal(http.MethodPatch, "/api/v1/private/settings/brand", []byte(body)))

			if rec.Code != http.StatusUnprocessableEntity {
				t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
			}
			if len(repo.updateCalls) != 0 {
				t.Fatal("a rejected value must never be written")
			}
		})
	}
}

func TestUpdateBrandHandler_MissingPrincipal_ReturnsSafe500(t *testing.T) {
	repo := &fakeBrandRepository{}
	h := httpapi.NewUpdateBrandHandler(shops.NewBrandService(repo, nil), profilesFor())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPatch, "/api/v1/private/settings/brand", bytes.NewReader([]byte(validBrandBody))))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(repo.updateCalls) != 0 {
		t.Fatal("no principal must never write")
	}
}

func TestBrandHandlers_PanelProfileComesFromTheRolesOfTheAuthenticatedUser(t *testing.T) {
	cases := []struct {
		name  string
		roles []shops.Role
		want  string
	}{
		{"sin roles", nil, "shop"},
		{"empresa", []shops.Role{shops.RoleEmpresa}, "shop"},
		{"independiente", []shops.Role{shops.RoleIndependiente}, "solo"},
		{"empleado", []shops.Role{shops.RoleEmpleado}, "employee"},
		{"empleado e independiente", []shops.Role{shops.RoleEmpleado, shops.RoleIndependiente}, "solo"},
	}
	for _, tc := range cases {
		t.Run("GET "+tc.name, func(t *testing.T) {
			repo := &fakeBrandRepository{getFound: true, getBrand: shops.DefaultBrand}
			rolesRepo := &fakePanelProfileRepository{roles: tc.roles}
			h := httpapi.NewGetBrandHandler(shops.NewBrandService(repo, nil), shops.NewPanelProfileService(rolesRepo))

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, requestWithPrincipal(http.MethodGet, "/api/v1/private/settings/brand", nil))

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			}
			var body httpapi.BrandResponse
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if body.PanelProfile != tc.want {
				t.Fatalf("expected panelProfile %q, got %q", tc.want, body.PanelProfile)
			}
			// El panel se pregunta por la persona de la sesión, en su barbería.
			if len(rolesRepo.calls) != 1 || rolesRepo.calls[0] != [2]string{testShopID, "u-1"} {
				t.Fatalf("roles must be read for the principal, got %v", rolesRepo.calls)
			}
		})
	}

	t.Run("PATCH responde con el panel del usuario", func(t *testing.T) {
		repo := &fakeBrandRepository{updateResult: shops.BrandUpdateResult{Brand: shops.DefaultBrand, Found: true}}
		h := httpapi.NewUpdateBrandHandler(shops.NewBrandService(repo, nil), profilesFor(shops.RoleEmpleado))

		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, requestWithPrincipal(http.MethodPatch, "/api/v1/private/settings/brand", []byte(validBrandBody)))

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var body httpapi.BrandResponse
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.PanelProfile != "employee" {
			t.Fatalf("expected panelProfile employee, got %q", body.PanelProfile)
		}
	})
}

// DEC-134: el panel ya no es editable; el contrato cerrado rechaza el campo.
func TestUpdateBrandHandler_RejectsPanelProfile(t *testing.T) {
	const body = `{"accent":"brass","businessTerm":"barbería","businessTermGender":"feminine","professionalTerm":"barbero","professionalTermPlural":"barberos","professionalTermGender":"masculine","panelProfile":"solo"}`
	repo := &fakeBrandRepository{}
	h := httpapi.NewUpdateBrandHandler(shops.NewBrandService(repo, nil), profilesFor())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, requestWithPrincipal(http.MethodPatch, "/api/v1/private/settings/brand", []byte(body)))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for a field the contract no longer declares, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(repo.updateCalls) != 0 {
		t.Fatal("a rejected body must never be written")
	}
}

func TestGetBrandHandler_WithLocation_ReturnsAddressAndCoordinates(t *testing.T) {
	brand := shops.DefaultBrand
	addr, lat, lng := "Calle 10 # 5-20, Bogotá", 4.598056, -74.076111
	brand.Location = shops.Location{HasPhysicalLocation: true, Address: &addr, Latitude: &lat, Longitude: &lng}
	h := httpapi.NewGetBrandHandler(shops.NewBrandService(&fakeBrandRepository{getFound: true, getBrand: brand}, nil), profilesFor())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, requestWithPrincipal(http.MethodGet, "/api/v1/private/settings/brand", nil))

	var body httpapi.BrandResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !body.HasPhysicalLocation || body.Address == nil || *body.Address != addr ||
		body.Latitude == nil || *body.Latitude != lat || body.Longitude == nil || *body.Longitude != lng {
		t.Fatalf("unexpected location: %+v", body)
	}
}

func TestUpdateBrandHandler_Location_DistinguishesAbsentFromNullAndValue(t *testing.T) {
	base := strings.TrimSuffix(validBrandBody, "}")
	cases := map[string]struct {
		extra      string
		wantSet    bool
		wantAddr   *string
		wantSwitch *bool
	}{
		"ausente no toca la ubicación": {extra: ``},
		"null quita la dirección":      {extra: `,"address":null`, wantSet: true},
		"texto fija la dirección": {extra: `,"address":"Calle 1 # 2-3"`, wantSet: true,
			wantAddr: func() *string { s := "Calle 1 # 2-3"; return &s }()},
		"solo el interruptor": {extra: `,"hasPhysicalLocation":false`,
			wantSwitch: func() *bool { b := false; return &b }()},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &fakeBrandRepository{getFound: true, updateResult: shops.BrandUpdateResult{Found: true}}
			geo := shops.GeocoderFunc(func(context.Context, string) (shops.Coordinates, error) {
				return shops.Coordinates{Latitude: 1, Longitude: 2}, nil
			})
			h := httpapi.NewUpdateBrandHandler(shops.NewBrandService(repo, geo), profilesFor())

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, requestWithPrincipal(http.MethodPatch, "/api/v1/private/settings/brand", []byte(base+tc.extra+"}")))

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			}
			w := repo.updateCalls[0].location
			if w.ReplaceAddress != tc.wantSet {
				t.Fatalf("ReplaceAddress = %v, quería %v", w.ReplaceAddress, tc.wantSet)
			}
			if (w.Address == nil) != (tc.wantAddr == nil) || (w.Address != nil && *w.Address != *tc.wantAddr) {
				t.Fatalf("Address = %v, quería %v", w.Address, tc.wantAddr)
			}
			if (w.HasPhysicalLocation == nil) != (tc.wantSwitch == nil) {
				t.Fatalf("HasPhysicalLocation = %v, quería %v", w.HasPhysicalLocation, tc.wantSwitch)
			}
		})
	}
}

func TestUpdateBrandHandler_UnresolvableAddress_Returns422WithoutWriting(t *testing.T) {
	repo := &fakeBrandRepository{getFound: true, updateResult: shops.BrandUpdateResult{Found: true}}
	geo := shops.GeocoderFunc(func(context.Context, string) (shops.Coordinates, error) {
		return shops.Coordinates{}, shops.ErrAddressUnresolvable
	})
	h := httpapi.NewUpdateBrandHandler(shops.NewBrandService(repo, geo), profilesFor())

	body := strings.TrimSuffix(validBrandBody, "}") + `,"address":"Ninguna parte 999"}`
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, requestWithPrincipal(http.MethodPatch, "/api/v1/private/settings/brand", []byte(body)))

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(repo.updateCalls) != 0 {
		t.Fatal("CA-102-01: una dirección no geocodificable no debe escribir")
	}
}

func TestUpdateBrandHandler_ClientCannotSendCoordinates(t *testing.T) {
	repo := &fakeBrandRepository{}
	h := httpapi.NewUpdateBrandHandler(shops.NewBrandService(repo, nil), profilesFor())

	for _, extra := range []string{`,"latitude":4.6`, `,"longitude":-74.0`} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, requestWithPrincipal(http.MethodPatch, "/api/v1/private/settings/brand",
			[]byte(strings.TrimSuffix(validBrandBody, "}")+extra+"}")))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s: las coordenadas las calcula el servidor; expected 400, got %d", extra, rec.Code)
		}
	}
	if len(repo.updateCalls) != 0 {
		t.Fatal("nada debe llegar al repositorio")
	}
}
