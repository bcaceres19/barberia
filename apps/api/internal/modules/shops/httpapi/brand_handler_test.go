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

func (f *fakeBrandRepository) Update(_ context.Context, barbershopID string, brand shops.Brand) (shops.BrandUpdateResult, error) {
	f.updateCalls = append(f.updateCalls, brandUpdateCall{barbershopID, brand})
	return f.updateResult, nil
}

var _ shops.BrandRepository = (*fakeBrandRepository)(nil)

const validBrandBody = `{"accent":"emerald","businessTerm":"Salón","businessTermGender":"masculine","professionalTerm":"Estilista","professionalTermPlural":"estilistas","professionalTermGender":"feminine"}`

func TestGetBrandHandler_Success_ReturnsTheSevenFields(t *testing.T) {
	repo := &fakeBrandRepository{getFound: true, getBrand: shops.DefaultBrand}
	h := httpapi.NewGetBrandHandler(shops.NewBrandService(repo))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, requestWithPrincipal(http.MethodGet, "/api/v1/private/settings/brand", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := map[string]string{
		"accent": "brass", "businessTerm": "barbería", "businessTermGender": "feminine",
		"professionalTerm": "barbero", "professionalTermPlural": "barberos", "professionalTermGender": "masculine",
		"panelProfile": "shop",
	}
	if len(body) != len(want) {
		t.Fatalf("expected exactly %d fields, got %v", len(want), body)
	}
	for k, v := range want {
		if body[k] != v {
			t.Fatalf("field %s: expected %q, got %q", k, v, body[k])
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
	h := httpapi.NewGetBrandHandler(shops.NewBrandService(&fakeBrandRepository{}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, requestWithPrincipal(http.MethodGet, "/api/v1/private/settings/brand", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetBrandHandler_MissingPrincipal_ReturnsSafe500(t *testing.T) {
	h := httpapi.NewGetBrandHandler(shops.NewBrandService(&fakeBrandRepository{}))

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
	h := httpapi.NewUpdateBrandHandler(shops.NewBrandService(repo))

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
	h := httpapi.NewUpdateBrandHandler(shops.NewBrandService(repo))

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
	h := httpapi.NewUpdateBrandHandler(shops.NewBrandService(repo))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, requestWithPrincipal(http.MethodPatch, "/api/v1/private/settings/brand", []byte(validBrandBody+`{}`)))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateBrandHandler_MalformedJSON_Returns400(t *testing.T) {
	h := httpapi.NewUpdateBrandHandler(shops.NewBrandService(&fakeBrandRepository{}))

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
			h := httpapi.NewUpdateBrandHandler(shops.NewBrandService(repo))

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
	h := httpapi.NewUpdateBrandHandler(shops.NewBrandService(repo))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPatch, "/api/v1/private/settings/brand", bytes.NewReader([]byte(validBrandBody))))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(repo.updateCalls) != 0 {
		t.Fatal("no principal must never write")
	}
}

func TestUpdateBrandHandler_PanelProfile(t *testing.T) {
	const prefix = `{"accent":"brass","businessTerm":"barbería","businessTermGender":"feminine","professionalTerm":"barbero","professionalTermPlural":"barberos","professionalTermGender":"masculine"`

	t.Run("solo se guarda y se devuelve", func(t *testing.T) {
		saved := shops.DefaultBrand
		saved.PanelProfile = shops.PanelProfileSolo
		repo := &fakeBrandRepository{updateResult: shops.BrandUpdateResult{Brand: saved, Found: true}}
		h := httpapi.NewUpdateBrandHandler(shops.NewBrandService(repo))

		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, requestWithPrincipal(http.MethodPatch, "/api/v1/private/settings/brand", []byte(prefix+`,"panelProfile":"solo"}`)))

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		if got := repo.updateCalls[0].brand.PanelProfile; got != shops.PanelProfileSolo {
			t.Fatalf("expected solo to reach the repository, got %q", got)
		}
		var body httpapi.BrandResponse
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.PanelProfile != "solo" {
			t.Fatalf("expected panelProfile solo in the response, got %q", body.PanelProfile)
		}
	})

	t.Run("omitido conserva el guardado", func(t *testing.T) {
		repo := &fakeBrandRepository{updateResult: shops.BrandUpdateResult{Brand: shops.DefaultBrand, Found: true}}
		h := httpapi.NewUpdateBrandHandler(shops.NewBrandService(repo))

		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, requestWithPrincipal(http.MethodPatch, "/api/v1/private/settings/brand", []byte(prefix+`}`)))

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for a client that does not know the profile, got %d: %s", rec.Code, rec.Body.String())
		}
		if got := repo.updateCalls[0].brand.PanelProfile; got != "" {
			t.Fatalf("an omitted profile must reach the repository empty, got %q", got)
		}
	})

	for name, value := range map[string]string{
		"valor desconocido": `"individual"`,
		"vacío explícito":   `""`,
		"nulo":              `null`,
		"número":            `1`,
	} {
		t.Run("rechaza "+name, func(t *testing.T) {
			repo := &fakeBrandRepository{}
			h := httpapi.NewUpdateBrandHandler(shops.NewBrandService(repo))

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, requestWithPrincipal(http.MethodPatch, "/api/v1/private/settings/brand", []byte(prefix+`,"panelProfile":`+value+`}`)))

			// Un número no es un texto: 400 del decodificador; el resto, 422 del servicio.
			// `null` decodifica como puntero nulo, es decir, «omitido»: 200 sin cambiar nada.
			switch name {
			case "número":
				if rec.Code != http.StatusBadRequest {
					t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
				}
			case "nulo":
				if rec.Code != http.StatusOK && rec.Code != http.StatusNotFound {
					t.Fatalf("expected null to behave as omitted, got %d: %s", rec.Code, rec.Body.String())
				}
			default:
				if rec.Code != http.StatusUnprocessableEntity {
					t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
				}
				if len(repo.updateCalls) != 0 {
					t.Fatal("a rejected profile must never be written")
				}
			}
		})
	}
}
