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
}

func (f *fakeRepository) ResolveBySlug(_ context.Context, slug string) (publicbooking.BarbershopProfile, bool, error) {
	f.calls = append(f.calls, slug)
	return f.profile, f.found, f.err
}

var _ publicbooking.Repository = (*fakeRepository)(nil)

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
