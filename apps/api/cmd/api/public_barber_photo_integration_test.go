package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Fixture hu092_seleccion_barbero.sql (DEC-127): Uno-1 tiene fotografía y un
// servicio activo asignado; Uno-2 no tiene fotografía; Uno-3 tiene fotografía
// pero ninguna asignación vigente.
const (
	barberPhotoSlug           = "barberia-hu092-uno"
	barberPhotoWithPhotoID    = "00920011-0092-0092-0092-009200110011"
	barberPhotoWithoutPhotoID = "00920012-0092-0092-0092-009200120012"
	barberPhotoNotEligibleID  = "00920014-0092-0092-0092-009200140014"
)

// TestPublicBarberPhoto_HTTP_ServesEligibleBarberWithoutSession comprueba el
// cableado de la ruta pública: sin cookie, 200 con la imagen y su ETag, 304 a
// la solicitud condicional y el mismo 404 para quien no tiene fotografía o no
// es elegible.
func TestPublicBarberPhoto_HTTP_ServesEligibleBarberWithoutSession(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	get := func(barberID, ifNoneMatch string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/public/barbershops/"+barberPhotoSlug+"/barbers/"+barberID+"/photo", nil)
		if ifNoneMatch != "" {
			req.Header.Set("If-None-Match", ifNoneMatch)
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}

	ok := get(barberPhotoWithPhotoID, "")
	if ok.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", ok.Code, ok.Body.String())
	}
	if got := ok.Header().Get("Content-Type"); got != "image/png" {
		t.Fatalf("expected image/png, got %q", got)
	}
	etag := ok.Header().Get("ETag")
	if etag == "" {
		t.Fatal("expected an ETag")
	}

	if rec := get(barberPhotoWithPhotoID, etag); rec.Code != http.StatusNotModified {
		t.Fatalf("expected 304 for a matching If-None-Match, got %d", rec.Code)
	}

	for name, barberID := range map[string]string{
		"sin fotografía":    barberPhotoWithoutPhotoID,
		"no elegible":       barberPhotoNotEligibleID,
		"id sin forma uuid": "no-es-un-uuid",
	} {
		if rec := get(barberID, ""); rec.Code != http.StatusNotFound {
			t.Errorf("%s: expected 404, got %d", name, rec.Code)
		}
	}
}
