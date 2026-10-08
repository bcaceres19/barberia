// Pruebas de integración de shops.PublicLinkRepository (issue #304, DEC-117)
// contra PostgreSQL REAL con el par shopA/shopB: la generación en la primera
// lectura, su estabilidad y el aislamiento entre tenants solo se pueden
// verificar contra la restricción y las políticas RLS reales.
package postgres_test

import (
	"context"
	"regexp"
	"testing"

	shopspostgres "system-barbershop/internal/modules/shops/postgres"
	"system-barbershop/internal/platform/database"
)

// standardShopName es el nombre original de cada fixture: las pruebas de este
// archivo cambian el nombre para derivar el slug y deben dejarlo como estaba
// (repository_test.go lo asume).
var standardShopName = map[string]string{
	shopA: "Barbería de prueba (aislamiento de paquete) 1",
	shopB: "Barbería de prueba (aislamiento de paquete) 2",
}

// prepareShop deja id sin slug y con el nombre name, y al terminar lo
// devuelve a su nombre estándar sin slug.
func prepareShop(t *testing.T, db *database.DB, id, name string) {
	t.Helper()
	resetSlugFixture(t, db, id, standardShopName[id])
	err := db.InTenantTx(context.Background(), database.BarbershopID(id), func(ctx context.Context, q database.Queries) error {
		_, err := q.Exec(ctx, `UPDATE barbershop SET name = $2 WHERE id = $1`, id, name)
		return err
	})
	if err != nil {
		t.Fatalf("prepareShop(%s): %v", id, err)
	}
}

// resolveSlug usa la misma función SECURITY DEFINER que la ruta pública
// (public_resolve_barbershop_by_slug) para saber a qué barbería apunta un
// slug: es lo que decide si un enlace "entra" a otra plataforma.
func resolveSlug(t *testing.T, db *database.DB, slug string) *string {
	t.Helper()
	var id *string
	err := db.InTenantTx(context.Background(), database.BarbershopID(shopA), func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx, `SELECT public_resolve_barbershop_by_slug($1)::text`, slug).Scan(&id)
	})
	if err != nil {
		t.Fatalf("resolveSlug(%q): %v", slug, err)
	}
	return id
}

func TestEnsure_NullSlug_GeneratesNameWithRandomCodeAndPersistsIt(t *testing.T) {
	db := setupTestDB(t)
	prepareShop(t, db, shopA, "Corte Fino Estudio")
	repo := shopspostgres.NewPublicLinkRepository(db)

	slug, found, err := repo.Ensure(context.Background(), shopA)
	if err != nil || !found {
		t.Fatalf("Ensure: found=%v err=%v", found, err)
	}
	if !regexp.MustCompile(`^cortefinoestudio[a-z2-9]{8}$`).MatchString(slug) {
		t.Fatalf("expected cortefinoestudio<código de 8>, got %q", slug)
	}
	if stored := readPublicSlug(t, db, shopA); stored == nil || *stored != slug {
		t.Fatalf("the generated slug must be persisted, stored=%v", stored)
	}
}

func TestEnsure_ExistingSlug_IsReturnedUnchanged(t *testing.T) {
	db := setupTestDB(t)
	prepareShop(t, db, shopA, "Corte Fino Estudio")
	repo := shopspostgres.NewPublicLinkRepository(db)

	first, _, err := repo.Ensure(context.Background(), shopA)
	if err != nil {
		t.Fatalf("first Ensure: %v", err)
	}
	for range 3 {
		again, found, err := repo.Ensure(context.Background(), shopA)
		if err != nil || !found || again != first {
			t.Fatalf("Ensure must be stable: got %q (found=%v err=%v), want %q", again, found, err, first)
		}
	}
}

func TestEnsure_ConcurrentFirstReads_ProduceASingleSlug(t *testing.T) {
	db := setupTestDB(t)
	prepareShop(t, db, shopA, "Corte Fino Estudio")
	repo := shopspostgres.NewPublicLinkRepository(db)

	const readers = 8
	results := make(chan string, readers)
	errs := make(chan error, readers)
	for range readers {
		go func() {
			slug, _, err := repo.Ensure(context.Background(), shopA)
			results <- slug
			errs <- err
		}()
	}
	seen := map[string]struct{}{}
	for range readers {
		seen[<-results] = struct{}{}
		if err := <-errs; err != nil {
			t.Fatalf("concurrent Ensure: %v", err)
		}
	}
	if len(seen) != 1 {
		t.Fatalf("concurrent first reads must converge on one slug, got %v", seen)
	}
}

// TestEnsure_TwoTenantsSameName_GetDistinctUnrelatedLinksThatResolveOnlyToTheirOwner
// cubre el pedido del propietario: dos barberías con el MISMO nombre reciben
// enlaces únicos, y cada enlace lleva únicamente a su propia barbería.
func TestEnsure_TwoTenantsSameName_GetDistinctUnrelatedLinksThatResolveOnlyToTheirOwner(t *testing.T) {
	db := setupTestDB(t)
	prepareShop(t, db, shopA, "Barbería Gemela")
	prepareShop(t, db, shopB, "Barbería Gemela")
	repo := shopspostgres.NewPublicLinkRepository(db)

	slugA, _, err := repo.Ensure(context.Background(), shopA)
	if err != nil {
		t.Fatalf("Ensure shopA: %v", err)
	}
	slugB, _, err := repo.Ensure(context.Background(), shopB)
	if err != nil {
		t.Fatalf("Ensure shopB: %v", err)
	}
	if slugA == slugB {
		t.Fatalf("two tenants share the public slug %q", slugA)
	}

	if got := resolveSlug(t, db, slugA); got == nil || *got != shopA {
		t.Fatalf("slug A must resolve to shopA, got %v", got)
	}
	if got := resolveSlug(t, db, slugB); got == nil || *got != shopB {
		t.Fatalf("slug B must resolve to shopB, got %v", got)
	}
	// Un slug vecino inventado (el patrón de los viejos «-2», «-3») no
	// resuelve a ninguna barbería: el código aleatorio no se puede deducir.
	for _, guess := range []string{"barberiagemela", "barberiagemela2", "barberia-gemela", "barberia-gemela-2"} {
		if got := resolveSlug(t, db, guess); got != nil {
			t.Fatalf("guessed slug %q must not resolve, got %v", guess, *got)
		}
	}
}

// TestEnsure_OtherTenantRow_IsNeverTouched: generar el enlace de una barbería
// no escribe ni lee la fila de otra.
func TestEnsure_OtherTenantRow_IsNeverTouched(t *testing.T) {
	db := setupTestDB(t)
	prepareShop(t, db, shopA, "Barbería A")
	prepareShop(t, db, shopB, "Barbería B")
	repo := shopspostgres.NewPublicLinkRepository(db)

	if _, _, err := repo.Ensure(context.Background(), shopA); err != nil {
		t.Fatalf("Ensure shopA: %v", err)
	}
	if got := readPublicSlug(t, db, shopB); got != nil {
		t.Fatalf("Ensure for shopA must not generate a slug for shopB, got %q", *got)
	}
}

func TestEnsure_InexistentBarbershop_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.NewPublicLinkRepository(db)

	_, found, err := repo.Ensure(context.Background(), "99999999-9999-9999-9999-999999999999")
	if err != nil {
		t.Fatalf("Ensure: %v", err)
	}
	if found {
		t.Fatal("an inexistent barbershop must not be found")
	}
}
