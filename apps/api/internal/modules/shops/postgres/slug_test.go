// Pruebas de integración de la generación automática del slug público
// (HU-090, DEC-082, DEC-117) dentro de shops.Repository.Update, contra PostgreSQL
// REAL: la unicidad GLOBAL entre dos tenants y el reintento ante
// unique_violation solo se pueden verificar contra la restricción real de
// la base (docs/03-desarrollo/estrategia-pruebas.md §2), nunca con un doble.
// Mismo par shopA/shopB reservado que repository_test.go (ver su
// comentario sobre el issue #158: paquete de test separado de cmd/api).
package postgres_test

import (
	"context"
	"errors"
	"regexp"
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
	if got == nil || !regexp.MustCompile(`^barberiaslugautomatico[a-z2-9]{8}$`).MatchString(*got) {
		t.Fatalf("expected public_slug=barberiaslugautomatico<código de 8>, got %v", got)
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

// fixedCodes devuelve un generador que entrega los códigos en orden: permite
// forzar una colisión real contra idx_barbershop_public_slug sin depender del
// azar.
func fixedCodes(codes ...string) func() (string, error) {
	i := 0
	return func() (string, error) {
		if i >= len(codes) {
			return "", errors.New("fixedCodes: sin más códigos")
		}
		code := codes[i]
		i++
		return code, nil
	}
}

// TestUpdate_SlugCollision_RetriesWithAnotherCode cubre DEC-082/DEC-117
// (unicidad GLOBAL): dos barberías DISTINTAS con el mismo nombre y el mismo
// código sorteado chocan en idx_barbershop_public_slug; la segunda reintenta
// con otro código dentro de la misma transacción y termina con un slug
// distinto, sin error para quien guarda.
func TestUpdate_SlugCollision_RetriesWithAnotherCode(t *testing.T) {
	db := setupTestDB(t)
	resetSlugFixture(t, db, shopA, "Barbería de prueba (aislamiento de paquete) 1")
	resetSlugFixture(t, db, shopB, "Barbería de prueba (aislamiento de paquete) 2")

	const collidingName = "Café Aroma Colisión"

	repoA := shopspostgres.New(db).WithSlugCode(fixedCodes("aaaaaaaa"))
	if _, err := repoA.Update(context.Background(), shopA, shops.UpdateInput{
		Name: collidingName, Timezone: "America/Bogota",
	}); err != nil {
		t.Fatalf("Update shopA: %v", err)
	}
	slugA := readPublicSlug(t, db, shopA)
	if slugA == nil || *slugA != "cafearomacolisionaaaaaaaa" {
		t.Fatalf("expected shopA public_slug=%q, got %v", "cafearomacolisionaaaaaaaa", slugA)
	}

	// shopB sortea primero el mismo código (colisión real) y luego otro.
	repoB := shopspostgres.New(db).WithSlugCode(fixedCodes("aaaaaaaa", "bbbbbbbb"))
	if _, err := repoB.Update(context.Background(), shopB, shops.UpdateInput{
		Name: collidingName, Timezone: "America/Bogota",
	}); err != nil {
		t.Fatalf("Update shopB must survive the collision: %v", err)
	}
	slugB := readPublicSlug(t, db, shopB)
	if slugB == nil || *slugB != "cafearomacolisionbbbbbbbb" {
		t.Fatalf("expected shopB public_slug=%q after retrying, got %v", "cafearomacolisionbbbbbbbb", slugB)
	}

	if *slugA == *slugB {
		t.Fatalf("CA-090-03/RN-TEN-01: two tenants ended up with the same public_slug %q", *slugA)
	}
}

// TestUpdate_SlugCollision_ExhaustedAttempts_FailsWithoutWriting: si todos los
// intentos chocan, la actualización falla entera (error, no bucle) y no deja
// el nombre a medias.
func TestUpdate_SlugCollision_ExhaustedAttempts_FailsWithoutWriting(t *testing.T) {
	db := setupTestDB(t)
	resetSlugFixture(t, db, shopA, "Barbería de prueba (aislamiento de paquete) 1")
	resetSlugFixture(t, db, shopB, "Barbería de prueba (aislamiento de paquete) 2")

	if _, err := shopspostgres.New(db).WithSlugCode(fixedCodes("aaaaaaaa")).Update(context.Background(), shopA, shops.UpdateInput{
		Name: "Mismo Nombre", Timezone: "America/Bogota",
	}); err != nil {
		t.Fatalf("Update shopA: %v", err)
	}

	always := func() (string, error) { return "aaaaaaaa", nil }
	_, err := shopspostgres.New(db).WithSlugCode(always).Update(context.Background(), shopB, shops.UpdateInput{
		Name: "Mismo Nombre", Timezone: "America/Bogota",
	})
	if err == nil {
		t.Fatal("expected an error when every attempt collides")
	}
	if got := readPublicSlug(t, db, shopB); got != nil {
		t.Fatalf("a failed update must leave public_slug NULL, got %q", *got)
	}
}
