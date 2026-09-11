// Pruebas de integración de la generación automática del slug público
// (HU-090, DEC-082) dentro de shops.Repository.Update, contra PostgreSQL
// REAL: la unicidad GLOBAL entre dos tenants y el reintento ante
// unique_violation solo se pueden verificar contra la restricción real de
// la base (docs/03-desarrollo/estrategia-pruebas.md §2), nunca con un doble.
// Mismo par shopA/shopB reservado que repository_test.go (ver su
// comentario sobre el issue #158: paquete de test separado de cmd/api).
package postgres_test

import (
	"context"
	"testing"

	"system-barbershop/internal/modules/shops"
	shopspostgres "system-barbershop/internal/modules/shops/postgres"
	"system-barbershop/internal/platform/database"
)

// readPublicSlug lee public_slug directamente por SQL: el contrato de
// shops.Repository/shops.Barbershop deliberadamente NUNCA expone esta
// columna (CA-020-07, HU-090 fuera de alcance para editarla desde HU-020),
// así que solo una prueba puede necesitar verla.
func readPublicSlug(t *testing.T, db *database.DB, id string) *string {
	t.Helper()
	var slug *string
	err := db.InTenantTx(context.Background(), database.BarbershopID(id), func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx, `SELECT public_slug FROM barbershop WHERE id = $1`, id).Scan(&slug)
	})
	if err != nil {
		t.Fatalf("readPublicSlug(%s): %v", id, err)
	}
	return slug
}

// resetSlugFixture deja id con el nombre original y public_slug=NULL antes
// y después de la prueba (t.Cleanup), para que las pruebas de este archivo
// no dependan del orden de ejecución ni dejen estado para
// repository_test.go (que nunca toca public_slug).
func resetSlugFixture(t *testing.T, db *database.DB, id, originalName string) {
	t.Helper()
	reset := func() {
		err := db.InTenantTx(context.Background(), database.BarbershopID(id), func(ctx context.Context, q database.Queries) error {
			_, err := q.Exec(ctx,
				`UPDATE barbershop SET name = $2, timezone = 'America/Bogota', public_slug = NULL WHERE id = $1`,
				id, originalName,
			)
			return err
		})
		if err != nil {
			t.Fatalf("resetSlugFixture(%s): %v", id, err)
		}
	}
	reset()
	t.Cleanup(reset)
}

func TestUpdate_GeneratesSlugFromName_WhenSlugIsNull(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.New(db)
	resetSlugFixture(t, db, shopA, "Barbería de prueba (aislamiento de paquete) 1")

	if before := readPublicSlug(t, db, shopA); before != nil {
		t.Fatalf("test fixture assumption broken: expected NULL public_slug before Update, got %q", *before)
	}

	result, err := repo.Update(context.Background(), shopA, shops.UpdateInput{
		Name: "Barbería Slug Automático", Timezone: "America/Bogota",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !result.Found {
		t.Fatalf("expected Found=true, got %+v", result)
	}

	got := readPublicSlug(t, db, shopA)
	if got == nil || *got != "barberia-slug-automatico" {
		t.Fatalf("expected public_slug=%q, got %v", "barberia-slug-automatico", got)
	}
}

func TestUpdate_DoesNotRegenerateSlug_WhenAlreadySet(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.New(db)
	resetSlugFixture(t, db, shopA, "Barbería de prueba (aislamiento de paquete) 1")

	if _, err := repo.Update(context.Background(), shopA, shops.UpdateInput{
		Name: "Nombre Original Del Slug", Timezone: "America/Bogota",
	}); err != nil {
		t.Fatalf("seed Update: %v", err)
	}
	original := readPublicSlug(t, db, shopA)
	if original == nil {
		t.Fatal("expected a slug to already exist after the seed Update")
	}

	// DEC-082: el slug se genera UNA sola vez; renombrar la barbería
	// después nunca lo toca (editarlo es una capacidad fuera del alcance
	// de HU-090, aunque la columna ya la soporta).
	if _, err := repo.Update(context.Background(), shopA, shops.UpdateInput{
		Name: "Nombre Completamente Distinto", Timezone: "America/Bogota",
	}); err != nil {
		t.Fatalf("second Update: %v", err)
	}

	after := readPublicSlug(t, db, shopA)
	if after == nil || *after != *original {
		t.Fatalf("expected public_slug to remain %q after renaming, got %v", *original, after)
	}
}

// TestUpdate_SlugCollisionAcrossTenants_AppendsDeterministicSuffix cubre
// DEC-082 (unicidad GLOBAL, sufijo numérico determinístico): dos barberías
// DISTINTAS con nombres que producen la misma base de slug reciben
// identificadores distintos, resuelto por el reintento real ante
// unique_violation de idx_barbershop_public_slug.
func TestUpdate_SlugCollisionAcrossTenants_AppendsDeterministicSuffix(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.New(db)
	resetSlugFixture(t, db, shopA, "Barbería de prueba (aislamiento de paquete) 1")
	resetSlugFixture(t, db, shopB, "Barbería de prueba (aislamiento de paquete) 2")

	const collidingName = "Café Aroma Colisión"

	if _, err := repo.Update(context.Background(), shopA, shops.UpdateInput{
		Name: collidingName, Timezone: "America/Bogota",
	}); err != nil {
		t.Fatalf("Update shopA: %v", err)
	}
	slugA := readPublicSlug(t, db, shopA)
	if slugA == nil || *slugA != "cafe-aroma-colision" {
		t.Fatalf("expected shopA public_slug=%q, got %v", "cafe-aroma-colision", slugA)
	}

	if _, err := repo.Update(context.Background(), shopB, shops.UpdateInput{
		Name: collidingName, Timezone: "America/Bogota",
	}); err != nil {
		t.Fatalf("Update shopB: %v", err)
	}
	slugB := readPublicSlug(t, db, shopB)
	if slugB == nil || *slugB != "cafe-aroma-colision-2" {
		t.Fatalf("expected shopB public_slug=%q (deterministic suffix), got %v", "cafe-aroma-colision-2", slugB)
	}

	if *slugA == *slugB {
		t.Fatalf("CA-090-03/RN-TEN-01: two tenants ended up with the same public_slug %q", *slugA)
	}
}
