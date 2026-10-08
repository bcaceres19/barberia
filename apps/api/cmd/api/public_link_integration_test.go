// Pruebas de integración del enlace público de reservas (issue #304, DEC-117)
// contra el router REAL de producción (buildRouter) y PostgreSQL real, mismo
// patrón que brand_integration_test.go. Requieren
// database/testdata/dos_barberias.sql + testdata/hu005_credenciales_sesiones.sql
// cargados.
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"system-barbershop/internal/platform/database"
)

func doPublicLinkRequest(handler http.Handler, rawToken string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/private/settings/public-link", nil)
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func decodePublicLinkSlug(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode public link: %v: %s", err, rec.Body.String())
	}
	if len(body) != 1 {
		t.Fatalf("expected exactly the slug field, got %v", body)
	}
	return body["slug"]
}

// clearPublicSlugs deja sin slug a las dos barberías de la fixture antes y
// después de la prueba: otras pruebas del paquete (HU-020) generan el suyo al
// guardar y no deben condicionar a esta.
func clearPublicSlugs(t *testing.T, db *database.DB) {
	t.Helper()
	clear := func() {
		for _, shop := range []string{shopA, shopB} {
			err := db.InTenantTx(context.Background(), database.BarbershopID(shop), func(ctx context.Context, q database.Queries) error {
				_, err := q.Exec(ctx, `UPDATE barbershop SET public_slug = NULL WHERE id = $1`, shop)
				return err
			})
			if err != nil {
				t.Fatalf("clearPublicSlugs(%s): %v", shop, err)
			}
		}
	}
	clear()
	t.Cleanup(clear)
}

func TestPublicLink_HTTP_WithoutSession_Is401(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	if rec := doPublicLinkRequest(router, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without a session, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestPublicLink_HTTP_FirstReadGeneratesAndLaterReadsKeepIt: una barbería sin
// enlace lo recibe al abrirlo y no cambia en lecturas posteriores.
func TestPublicLink_HTTP_FirstReadGeneratesAndLaterReadsKeepIt(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	clearPublicSlugs(t, db)

	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "public-link-a")
	first := doPublicLinkRequest(router, raw)
	slug := decodePublicLinkSlug(t, first)
	if !regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{1,38}[a-z0-9])$`).MatchString(slug) {
		t.Fatalf("generated slug %q does not satisfy barbershop_public_slug_ck", slug)
	}
	if got := first.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("expected Cache-Control no-store, got %q", got)
	}
	if again := decodePublicLinkSlug(t, doPublicLinkRequest(router, raw)); again != slug {
		t.Fatalf("a later read must keep the link: %q -> %q", slug, again)
	}
}

// TestPublicLink_HTTP_TwoTenants_EachSeesOnlyItsOwnLinkAndItOpensOnlyItsOwnShop
// cubre el pedido del propietario de extremo a extremo: cada barbería recibe
// un enlace distinto, su sesión nunca devuelve el de la otra y, abierto sin
// sesión, cada enlace muestra únicamente la plataforma de su dueño.
func TestPublicLink_HTTP_TwoTenants_EachSeesOnlyItsOwnLinkAndItOpensOnlyItsOwnShop(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	clearPublicSlugs(t, db)

	rawA := createSessionCookie(t, db, shopA, staffUserActiveA, "public-link-tenant-a")
	rawB := createSessionCookie(t, db, shopB, staffUserActiveB, "public-link-tenant-b")

	slugA := decodePublicLinkSlug(t, doPublicLinkRequest(router, rawA))
	slugB := decodePublicLinkSlug(t, doPublicLinkRequest(router, rawB))
	if slugA == slugB {
		t.Fatalf("two tenants share the public slug %q", slugA)
	}
	// La sesión B repetida sigue devolviendo el suyo, nunca el de A.
	if again := decodePublicLinkSlug(t, doPublicLinkRequest(router, rawB)); again != slugB {
		t.Fatalf("session B must keep reading its own slug %q, got %q", slugB, again)
	}

	nameOf := func(slug string) string {
		rec := doResolvePublicBarbershopRequest(router, slug)
		if rec.Code != http.StatusOK {
			t.Fatalf("public link %q must open without a session, got %d: %s", slug, rec.Code, rec.Body.String())
		}
		var body publicBarbershopProfileBody
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("decode public profile: %v", err)
		}
		return body.Name
	}
	nameA, nameB := nameOf(slugA), nameOf(slugB)
	if nameA == nameB {
		t.Fatalf("each link must open its own barbershop, both opened %q", nameA)
	}
}
