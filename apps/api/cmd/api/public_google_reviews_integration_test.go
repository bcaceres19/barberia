// Pruebas de integración de HU-101 (DEC-129) contra el router REAL y PostgreSQL
// real: el endpoint público de reseñas de Google resuelve la barbería por slug,
// usa SU ficha y degrada a la estructura vacía. En APP_ENVIRONMENT=test, sin
// clave, el proveedor es el simulado: ninguna prueba toca la red.
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"system-barbershop/internal/platform/database"
)

type googleReviewsBody struct {
	Rating      *float64        `json:"rating"`
	RatingCount int             `json:"ratingCount"`
	MapsURL     *string         `json:"mapsUrl"`
	Reviews     []googleReviewB `json:"reviews"`
}

type googleReviewB struct {
	AuthorName string `json:"authorName"`
	Rating     int    `json:"rating"`
}

func doGoogleReviewsRequest(router http.Handler, slug string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/barbershops/"+slug+"/reviews", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func setGooglePlaceID(t *testing.T, db *database.DB, slug string, placeID *string) {
	t.Helper()
	ctx := context.Background()
	id, found, err := db.ResolveTenant(ctx, `SELECT public_resolve_barbershop_by_slug($1)`, slug)
	if err != nil || !found {
		t.Fatalf("resolver %s: found=%v err=%v", slug, found, err)
	}
	err = db.InTenantTx(ctx, id, func(ctx context.Context, q database.Queries) error {
		_, err := q.Exec(ctx, `UPDATE barbershop SET google_place_id = $2 WHERE id = $1`, string(id), placeID)
		return err
	})
	if err != nil {
		t.Fatalf("guardar ficha de %s: %v", slug, err)
	}
}

// TestPublicGoogleReviews_HTTP_FollowsTheShopListingAndStaysEmptyOtherwise cubre
// CA-101-01/02 y RN-TEN-01: sin ficha responde la estructura vacía; con ella, la
// calificación y las reseñas (aquí del proveedor simulado); la ficha de una
// barbería nunca aparece en la otra; y el Place ID nunca viaja en la respuesta.
func TestPublicGoogleReviews_HTTP_FollowsTheShopListingAndStaysEmptyOtherwise(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	place := "ChIJficha_de_prueba-1"
	t.Cleanup(func() { setGooglePlaceID(t, db, publicShopSlugUno, nil) })

	rec := doGoogleReviewsRequest(router, publicShopSlugUno)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"rating":null,"ratingCount":0,"mapsUrl":null,"reviews":[]}` {
		t.Fatalf("CA-101-01: sin ficha debe ser la estructura vacía, fue %s", got)
	}

	setGooglePlaceID(t, db, publicShopSlugUno, &place)
	rec = doGoogleReviewsRequest(router, publicShopSlugUno)
	var body googleReviewsBody
	if rec.Code != http.StatusOK || json.Unmarshal(rec.Body.Bytes(), &body) != nil {
		t.Fatalf("expected 200 with a body, got %d: %s", rec.Code, rec.Body.String())
	}
	if body.Rating == nil || body.RatingCount == 0 || body.MapsURL == nil || len(body.Reviews) == 0 || len(body.Reviews) > 5 {
		t.Fatalf("CA-101-02: con ficha debe traer calificación, total, enlace y hasta 5 reseñas: %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), place) {
		t.Fatal("el Place ID nunca debe viajar al navegador")
	}

	other := doGoogleReviewsRequest(router, publicShopSlugDos)
	if got := strings.TrimSpace(other.Body.String()); other.Code != http.StatusOK || got != `{"rating":null,"ratingCount":0,"mapsUrl":null,"reviews":[]}` {
		t.Fatalf("RN-TEN-01: la ficha de uno se filtró al otro: %d %s", other.Code, other.Body.String())
	}

	if rec := doGoogleReviewsRequest(router, "barberia-que-no-existe"); rec.Code != http.StatusNotFound {
		t.Fatalf("un enlace inexistente debe ser 404, fue %d", rec.Code)
	}
}
