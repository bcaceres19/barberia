package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"system-barbershop/internal/modules/shops"
	"system-barbershop/internal/modules/shops/httpapi"
)

// fakePublicLinkRepository registra a qué barbería se le pidió el enlace:
// estas pruebas verifican el adaptador HTTP en aislamiento; la generación y
// el aislamiento reales viven en postgres/public_link_repository_test.go y
// cmd/api/public_link_integration_test.go.
type fakePublicLinkRepository struct {
	slug      string
	found     bool
	gotShopID string
}

func (f *fakePublicLinkRepository) Ensure(_ context.Context, barbershopID string) (string, bool, error) {
	f.gotShopID = barbershopID
	return f.slug, f.found, nil
}

func TestGetPublicLinkHandler_Success_ReturnsOnlyTheSlugAndForbidsCaching(t *testing.T) {
	repo := &fakePublicLinkRepository{slug: "corte-fino-k7x2m9", found: true}
	h := httpapi.NewGetPublicLinkHandler(shops.NewPublicLinkService(repo))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, requestWithPrincipal(http.MethodGet, "/api/v1/private/settings/public-link", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body) != 1 || body["slug"] != "corte-fino-k7x2m9" {
		t.Fatalf("expected exactly {slug}, got %v", body)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("expected Cache-Control no-store, got %q", got)
	}
	// El tenant sale del principal de sesión, nunca de la solicitud.
	if repo.gotShopID != testShopID {
		t.Fatalf("expected the session barbershop %q, got %q", testShopID, repo.gotShopID)
	}
}

func TestGetPublicLinkHandler_NoPrincipal_Is500(t *testing.T) {
	h := httpapi.NewGetPublicLinkHandler(shops.NewPublicLinkService(&fakePublicLinkRepository{}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/private/settings/public-link", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 without a principal, got %d", rec.Code)
	}
}

func TestGetPublicLinkHandler_BarbershopNotVisible_Is404(t *testing.T) {
	h := httpapi.NewGetPublicLinkHandler(shops.NewPublicLinkService(&fakePublicLinkRepository{found: false}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, requestWithPrincipal(http.MethodGet, "/api/v1/private/settings/public-link", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}
