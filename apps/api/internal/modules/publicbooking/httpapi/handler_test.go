package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"system-barbershop/internal/modules/publicbooking"
	"system-barbershop/internal/modules/publicbooking/httpapi"
	"system-barbershop/internal/platform/httpserver"
)

// fakeRepository es un doble en memoria de publicbooking.Repository: estas
// pruebas verifican el adaptador HTTP en aislamiento (extracción del
// parámetro de ruta, traducción de errores, forma de la respuesta), no
// PostgreSQL real (esa cobertura vive en
// internal/modules/publicbooking/postgres/repository_test.go).
type fakeRepository struct {
	profile publicbooking.BarbershopProfile
	found   bool
	err     error
	calls   []string

	listResult publicbooking.PublicServiceListResult
	listFound  bool
	listErr    error

	barbersResult publicbooking.PublicBarberListResult
	barbersFound  bool
	barbersErr    error
}

func (f *fakeRepository) ResolveBySlug(_ context.Context, slug string) (publicbooking.BarbershopProfile, bool, error) {
	f.calls = append(f.calls, slug)
	return f.profile, f.found, f.err
}

func (f *fakeRepository) ListPublicServices(_ context.Context, slug string, _ *publicbooking.ServiceCursor, _ int) (publicbooking.PublicServiceListResult, bool, error) {
	f.calls = append(f.calls, slug)
	return f.listResult, f.listFound, f.listErr
}

func (f *fakeRepository) ListPublicBarbers(_ context.Context, slug string, _ string) (publicbooking.PublicBarberListResult, bool, error) {
	f.calls = append(f.calls, slug)
	return f.barbersResult, f.barbersFound, f.barbersErr
}

var _ publicbooking.Repository = (*fakeRepository)(nil)

const validServiceID = "11111111-1111-1111-1111-111111111111"

func requestWithSlug(slug string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/barbershops/"+slug, nil)
	return httpserver.RequestWithURLParam(req, "slug", slug)
}

func TestResolveBarbershopHandler_Success_ReturnsPublicProfile(t *testing.T) {
	email := "contacto@ejemplo.test"
	phone := "+573001234567"
	repo := &fakeRepository{
		found: true,
		profile: publicbooking.BarbershopProfile{
			Name: "Barbería Ejemplo", Timezone: "America/Bogota",
			ContactEmail: &email, ContactPhone: &phone,
		},
	}
	h := httpapi.NewResolveBarbershopHandler(publicbooking.NewService(repo))

	req := requestWithSlug("barberia-ejemplo")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body httpapi.PublicBarbershopProfileResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body.Name != "Barbería Ejemplo" || body.Timezone != "America/Bogota" {
		t.Fatalf("unexpected response: %+v", body)
	}
	if body.ContactEmail == nil || *body.ContactEmail != email {
		t.Fatalf("expected contactEmail=%q, got %v", email, body.ContactEmail)
	}
	if len(repo.calls) != 1 || repo.calls[0] != "barberia-ejemplo" {
		t.Fatalf("expected exactly one ResolveBySlug call with the path slug, got %v", repo.calls)
	}

	// CA-090-04: el cuerpo nunca menciona un identificador interno ni el
	// slug mismo.
	raw := strings.ToLower(rec.Body.String())
	for _, forbidden := range []string{"barbershopid", "\"id\"", "slug"} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("CA-090-04: response body must never mention %q: %s", forbidden, raw)
		}
	}
}

func TestResolveBarbershopHandler_NoContact_ReturnsNullNotOmitted(t *testing.T) {
	repo := &fakeRepository{
		found:   true,
		profile: publicbooking.BarbershopProfile{Name: "Barbería Ejemplo", Timezone: "America/Bogota"},
	}
	h := httpapi.NewResolveBarbershopHandler(publicbooking.NewService(repo))

	req := requestWithSlug("barberia-ejemplo")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var raw map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&raw); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	for _, field := range []string{"contactEmail", "contactPhone"} {
		v, present := raw[field]
		if !present {
			t.Fatalf("expected %q to be present (as null), it was omitted", field)
		}
		if v != nil {
			t.Fatalf("expected %q=null, got %v", field, v)
		}
	}
}

func TestResolveBarbershopHandler_NotFound_ReturnsUniform404(t *testing.T) {
	repo := &fakeRepository{found: false}
	h := httpapi.NewResolveBarbershopHandler(publicbooking.NewService(repo))

	req := requestWithSlug("no-existe")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("expected Content-Type application/problem+json, got %q", ct)
	}
}

// TestResolveBarbershopHandler_MalformedSlug_ReturnsSameUniform404AsUnknown
// cubre CA-090-02: la respuesta HTTP para un slug con forma imposible es
// indistinguible de la de un slug desconocido.
func TestResolveBarbershopHandler_MalformedSlug_ReturnsSameUniform404AsUnknown(t *testing.T) {
	repo := &fakeRepository{found: false}
	h := httpapi.NewResolveBarbershopHandler(publicbooking.NewService(repo))

	req := requestWithSlug("Bad%20Slug!")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestResolveBarbershopHandler_RepositoryError_ReturnsSafe500(t *testing.T) {
	repo := &fakeRepository{err: errors.New("boom")}
	h := httpapi.NewResolveBarbershopHandler(publicbooking.NewService(repo))

	req := requestWithSlug("barberia-ejemplo")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "boom") {
		t.Fatalf("CA-003-02: the internal cause must never reach the response body: %s", rec.Body.String())
	}
}

func requestWithSlugAndQuery(slug, rawQuery string) *http.Request {
	target := "/api/v1/public/barbershops/" + slug + "/services"
	if rawQuery != "" {
		target += "?" + rawQuery
	}
	req := httptest.NewRequest(http.MethodGet, target, nil)
	return httpserver.RequestWithURLParam(req, "slug", slug)
}

func TestListPublicServicesHandler_Success_ReturnsItems(t *testing.T) {
	desc := "Con lavado incluido"
	repo := &fakeRepository{
		listFound: true,
		listResult: publicbooking.PublicServiceListResult{
			Items: []publicbooking.PublicService{
				{ID: "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4", Name: "Corte clásico", Description: &desc, DurationMinutes: 30, PriceCents: 4500000, Currency: "COP"},
			},
		},
	}
	h := httpapi.NewListPublicServicesHandler(publicbooking.NewService(repo))

	req := requestWithSlugAndQuery("barberia-ejemplo", "")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body httpapi.PublicServiceListResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if len(body.Items) != 1 {
		t.Fatalf("expected 1 item, got %+v", body.Items)
	}
	item := body.Items[0]
	if item.Name != "Corte clásico" || item.DurationMinutes != 30 || item.Price != "45000.00" || item.Currency != "COP" {
		t.Fatalf("unexpected item: %+v", item)
	}
	if item.Description == nil || *item.Description != desc {
		t.Fatalf("expected description=%q, got %v", desc, item.Description)
	}
	if body.NextCursor != nil {
		t.Fatalf("expected nextCursor=null for a single-item page, got %v", *body.NextCursor)
	}
}

func TestListPublicServicesHandler_Empty_ReturnsEmptyArrayNotNull(t *testing.T) {
	repo := &fakeRepository{listFound: true}
	h := httpapi.NewListPublicServicesHandler(publicbooking.NewService(repo))

	req := requestWithSlugAndQuery("barberia-ejemplo", "")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var raw map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&raw); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	items, ok := raw["items"].([]any)
	if !ok {
		t.Fatalf("expected items to be a JSON array, got %T: %v", raw["items"], raw["items"])
	}
	if len(items) != 0 {
		t.Fatalf("expected an empty array, got %v", items)
	}
}

func TestListPublicServicesHandler_NotFound_ReturnsUniform404(t *testing.T) {
	repo := &fakeRepository{listFound: false}
	h := httpapi.NewListPublicServicesHandler(publicbooking.NewService(repo))

	req := requestWithSlugAndQuery("no-existe", "")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListPublicServicesHandler_InvalidLimit_Returns400(t *testing.T) {
	repo := &fakeRepository{listFound: true}
	h := httpapi.NewListPublicServicesHandler(publicbooking.NewService(repo))

	req := requestWithSlugAndQuery("barberia-ejemplo", "limit=-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListPublicServicesHandler_InvalidCursor_Returns400(t *testing.T) {
	repo := &fakeRepository{listFound: true}
	h := httpapi.NewListPublicServicesHandler(publicbooking.NewService(repo))

	req := requestWithSlugAndQuery("barberia-ejemplo", "cursor=no-es-un-cursor-valido")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListPublicServicesHandler_RepositoryError_ReturnsSafe500(t *testing.T) {
	repo := &fakeRepository{listErr: errors.New("boom")}
	h := httpapi.NewListPublicServicesHandler(publicbooking.NewService(repo))

	req := requestWithSlugAndQuery("barberia-ejemplo", "")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "boom") {
		t.Fatalf("CA-003-02: the internal cause must never reach the response body: %s", rec.Body.String())
	}
}

func requestWithSlugAndServiceID(slug, serviceID string) *http.Request {
	target := "/api/v1/public/barbershops/" + slug + "/services/" + serviceID + "/barbers"
	req := httptest.NewRequest(http.MethodGet, target, nil)
	req = httpserver.RequestWithURLParam(req, "slug", slug)
	return httpserver.RequestWithURLParam(req, "serviceId", serviceID)
}

func TestListPublicBarbersHandler_OneBarber_ReturnsSingleItem(t *testing.T) {
	repo := &fakeRepository{
		barbersFound: true,
		barbersResult: publicbooking.PublicBarberListResult{
			Items: []publicbooking.PublicBarber{{ID: "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4", FullName: "Juan Pérez"}},
		},
	}
	h := httpapi.NewListPublicBarbersHandler(publicbooking.NewService(repo))

	req := requestWithSlugAndServiceID("barberia-ejemplo", validServiceID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body httpapi.PublicBarberListResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if len(body.Items) != 1 || body.Items[0].FullName != "Juan Pérez" {
		t.Fatalf("unexpected items: %+v", body.Items)
	}
}

func TestListPublicBarbersHandler_ZeroBarbers_ReturnsEmptyArrayNotNull(t *testing.T) {
	repo := &fakeRepository{barbersFound: true}
	h := httpapi.NewListPublicBarbersHandler(publicbooking.NewService(repo))

	req := requestWithSlugAndServiceID("barberia-ejemplo", validServiceID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var raw map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&raw); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	items, ok := raw["items"].([]any)
	if !ok {
		t.Fatalf("expected items to be a JSON array, got %T: %v", raw["items"], raw["items"])
	}
	if len(items) != 0 {
		t.Fatalf("expected an empty array, got %v", items)
	}
}

func TestListPublicBarbersHandler_ManyBarbers_ReturnsAllItemsInOrder(t *testing.T) {
	repo := &fakeRepository{
		barbersFound: true,
		barbersResult: publicbooking.PublicBarberListResult{
			Items: []publicbooking.PublicBarber{
				{ID: "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4", FullName: "Ana Gómez"},
				{ID: "1c9a7f2d-3b4e-4a5f-9c6d-7e8f9a0b1c2d", FullName: "Luis Rojas"},
			},
		},
	}
	h := httpapi.NewListPublicBarbersHandler(publicbooking.NewService(repo))

	req := requestWithSlugAndServiceID("barberia-ejemplo", validServiceID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var body httpapi.PublicBarberListResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if len(body.Items) != 2 || body.Items[0].FullName != "Ana Gómez" || body.Items[1].FullName != "Luis Rojas" {
		t.Fatalf("unexpected items: %+v", body.Items)
	}
}

func TestListPublicBarbersHandler_BarbershopNotFound_ReturnsUniform404(t *testing.T) {
	repo := &fakeRepository{barbersFound: false}
	h := httpapi.NewListPublicBarbersHandler(publicbooking.NewService(repo))

	req := requestWithSlugAndServiceID("no-existe", validServiceID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestListPublicBarbersHandler_ForeignOrMalformedServiceID_ReturnsSameSuccessAsZeroBarbers
// cubre CA-092-03: un serviceID ajeno, inexistente o malformado nunca
// produce un error distinto de "cero barberos elegibles" -misma respuesta
// 200 con items=[], sin importar la causa concreta.
func TestListPublicBarbersHandler_ForeignOrMalformedServiceID_ReturnsSameSuccessAsZeroBarbers(t *testing.T) {
	for _, serviceID := range []string{validServiceID, "no-es-un-uuid", "00000000-0000-0000-0000-000000000000"} {
		t.Run(serviceID, func(t *testing.T) {
			repo := &fakeRepository{barbersFound: true}
			h := httpapi.NewListPublicBarbersHandler(publicbooking.NewService(repo))

			req := requestWithSlugAndServiceID("barberia-ejemplo", serviceID)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestListPublicBarbersHandler_RepositoryError_ReturnsSafe500(t *testing.T) {
	repo := &fakeRepository{barbersErr: errors.New("boom")}
	h := httpapi.NewListPublicBarbersHandler(publicbooking.NewService(repo))

	req := requestWithSlugAndServiceID("barberia-ejemplo", validServiceID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "boom") {
		t.Fatalf("CA-003-02: the internal cause must never reach the response body: %s", rec.Body.String())
	}
}
