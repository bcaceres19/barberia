package postgres_test

import (
	"context"
	"testing"

	"system-barbershop/internal/modules/shops"
	shopspostgres "system-barbershop/internal/modules/shops/postgres"
	"system-barbershop/internal/platform/database"
)

// restoreBrand devuelve la marca de shop a los valores iniciales al terminar
// la prueba: comparten la fila de dos_barberias.sql en vez de crear una
// barbería propia (no existe alta de barberías).
func restoreBrand(t *testing.T, db *database.DB, shop string) {
	t.Helper()
	t.Cleanup(func() {
		err := db.InTenantTx(context.Background(), database.BarbershopID(shop), func(ctx context.Context, q database.Queries) error {
			_, err := q.Exec(ctx,
				`UPDATE barbershop
				    SET brand_accent = 'brass', business_term = 'barbería', business_term_gender = 'feminine',
				        professional_term = 'barbero', professional_term_plural = 'barberos',
				        professional_term_gender = 'masculine'
				  WHERE id = $1`, shop)
			return err
		})
		if err != nil {
			t.Fatalf("restoreBrand cleanup: %v", err)
		}
	})
}

func salonBrand() shops.Brand {
	return shops.Brand{
		Accent: "ruby", BusinessTerm: "salón de belleza", BusinessTermGender: shops.GenderMasculine,
		ProfessionalTerm: "estilista", ProfessionalTermPlural: "estilistas", ProfessionalTermGender: shops.GenderFeminine,
	}
}

// TestBrandGet_ReturnsTheInitialValuesForAnUntouchedBarbershop: una barbería
// que nunca cambió su marca lee exactamente la interfaz anterior.
func TestBrandGet_ReturnsTheInitialValuesForAnUntouchedBarbershop(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.NewBrandRepository(db)
	restoreBrand(t, db, shopA)

	got, found, err := repo.Get(context.Background(), shopA)
	if err != nil || !found {
		t.Fatalf("Get: found=%v err=%v", found, err)
	}
	if got != shops.DefaultBrand {
		t.Fatalf("expected the initial values %+v, got %+v", shops.DefaultBrand, got)
	}
}

func TestBrandUpdate_PersistsAndRoundTrips(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.NewBrandRepository(db)
	restoreBrand(t, db, shopA)

	result, err := repo.Update(context.Background(), shopA, salonBrand())
	if err != nil || !result.Found {
		t.Fatalf("Update: found=%v err=%v", result.Found, err)
	}
	if result.Brand != salonBrand() {
		t.Fatalf("RETURNING must echo what was stored, got %+v", result.Brand)
	}

	// Lectura independiente: descarta que RETURNING mienta sobre lo persistido.
	got, found, err := repo.Get(context.Background(), shopA)
	if err != nil || !found || got != salonBrand() {
		t.Fatalf("expected the persisted row to match, got %+v (found=%v err=%v)", got, found, err)
	}
}

// TestBrandUpdate_DoesNotTouchHU020OrHU093Columns: la marca vive en la misma
// fila que el nombre, la zona, el contacto y la política de reserva, y
// escribirla nunca debe alterarlos.
func TestBrandUpdate_DoesNotTouchOtherSettingsOfTheSameRow(t *testing.T) {
	db := setupTestDB(t)
	brands := shopspostgres.NewBrandRepository(db)
	settings := shopspostgres.New(db)
	policies := shopspostgres.NewBookingPolicyRepository(db)
	restoreBrand(t, db, shopA)

	nameBefore, _, err := settings.Get(context.Background(), shopA)
	if err != nil {
		t.Fatalf("settings.Get before: %v", err)
	}
	policyBefore, _, err := policies.Get(context.Background(), shopA)
	if err != nil {
		t.Fatalf("policies.Get before: %v", err)
	}

	if _, err := brands.Update(context.Background(), shopA, salonBrand()); err != nil {
		t.Fatalf("Update: %v", err)
	}

	nameAfter, _, err := settings.Get(context.Background(), shopA)
	if err != nil {
		t.Fatalf("settings.Get after: %v", err)
	}
	if !sameBarbershop(nameBefore, nameAfter) {
		t.Fatalf("HU-020 columns changed: before %+v after %+v", nameBefore, nameAfter)
	}
	policyAfter, _, err := policies.Get(context.Background(), shopA)
	if err != nil {
		t.Fatalf("policies.Get after: %v", err)
	}
	// El token de versión cambia a propósito (comparte updated_at, documentado);
	// los seis campos de la política no.
	policyBefore.VersionToken, policyAfter.VersionToken = "", ""
	if policyBefore != policyAfter {
		t.Fatalf("HU-093 columns changed: before %+v after %+v", policyBefore, policyAfter)
	}
}

// TestBrandUpdate_TwoTenants_NeverCrossesBarbershops es la prueba de
// aislamiento por tenant sobre PostgreSQL real con dos barberías.
func TestBrandUpdate_TwoTenants_NeverCrossesBarbershops(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.NewBrandRepository(db)
	restoreBrand(t, db, shopA)
	restoreBrand(t, db, shopB)

	if _, err := repo.Update(context.Background(), shopA, salonBrand()); err != nil {
		t.Fatalf("Update shopA: %v", err)
	}

	gotB, found, err := repo.Get(context.Background(), shopB)
	if err != nil || !found {
		t.Fatalf("Get shopB: found=%v err=%v", found, err)
	}
	if gotB != shops.DefaultBrand {
		t.Fatalf("updating shopA changed shopB's brand: %+v", gotB)
	}
}

// TestBrandUpdate_RLSBlocksWritingAnotherTenantsRow: con el contexto del
// tenant A, pedir a la base que escriba la fila de B no encuentra nada, aunque
// el identificador de B se pase explícitamente (defensa en profundidad: RLS,
// no solo el WHERE de la sentencia).
func TestBrandUpdate_RLSBlocksWritingAnotherTenantsRow(t *testing.T) {
	db := setupTestDB(t)
	restoreBrand(t, db, shopB)

	var rows int64
	err := db.InTenantTx(context.Background(), database.BarbershopID(shopA), func(ctx context.Context, q database.Queries) error {
		tag, err := q.Exec(ctx, `UPDATE barbershop SET brand_accent = 'ruby' WHERE id = $1`, shopB)
		rows = tag.RowsAffected()
		return err
	})
	if err != nil {
		t.Fatalf("InTenantTx: %v", err)
	}
	if rows != 0 {
		t.Fatalf("RLS must hide shopB from shopA's tenant context, %d rows were updated", rows)
	}
}

// TestBrandUpdate_DatabaseCheckIsTheLastLineOfDefense: aunque el servicio
// validara mal, la base rechaza un acento fuera de la lista y un término sin
// normalizar, y la sentencia atómica no deja una escritura parcial.
func TestBrandUpdate_DatabaseCheckIsTheLastLineOfDefense(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.NewBrandRepository(db)
	restoreBrand(t, db, shopA)

	for name, mutate := range map[string]func(*shops.Brand){
		"acento fuera de la lista": func(b *shops.Brand) { b.Accent = "rojo" },
		"término sin minúsculas":   func(b *shops.Brand) { b.BusinessTerm = "Salón" },
		"término con dígitos":      func(b *shops.Brand) { b.ProfessionalTermPlural = "estilistas2" },
		"género inválido":          func(b *shops.Brand) { b.ProfessionalTermGender = "x" },
	} {
		b := salonBrand()
		mutate(&b)
		if _, err := repo.Update(context.Background(), shopA, b); err == nil {
			t.Fatalf("%s: la base debía rechazar el valor", name)
		}
	}

	got, _, err := repo.Get(context.Background(), shopA)
	if err != nil || got != shops.DefaultBrand {
		t.Fatalf("a rejected write must leave the row untouched, got %+v err=%v", got, err)
	}
}
