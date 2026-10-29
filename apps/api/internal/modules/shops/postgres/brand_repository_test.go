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
				        professional_term_gender = 'masculine',
				        has_physical_location = true, address = NULL, latitude = NULL, longitude = NULL
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
		// La ubicación que lee una barbería intacta: interruptor activo, sin dirección.
		Location: shops.Location{HasPhysicalLocation: true},
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

	result, err := repo.Update(context.Background(), shopA, salonBrand(), shops.LocationWrite{})
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

	if _, err := brands.Update(context.Background(), shopA, salonBrand(), shops.LocationWrite{}); err != nil {
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

	if _, err := repo.Update(context.Background(), shopA, salonBrand(), shops.LocationWrite{}); err != nil {
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
		if _, err := repo.Update(context.Background(), shopA, b, shops.LocationWrite{}); err == nil {
			t.Fatalf("%s: la base debía rechazar el valor", name)
		}
	}

	got, _, err := repo.Get(context.Background(), shopA)
	if err != nil || got != shops.DefaultBrand {
		t.Fatalf("a rejected write must leave the row untouched, got %+v err=%v", got, err)
	}
}

func f64(v float64) *float64 { return &v }
func str(v string) *string   { return &v }

// TestBrandUpdate_Location_PersistsAndRoundTrips cubre CA-102-01 en la capa de
// base: la dirección y sus coordenadas se guardan juntas y se leen iguales.
func TestBrandUpdate_Location_PersistsAndRoundTrips(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.NewBrandRepository(db)
	restoreBrand(t, db, shopA)

	res, err := repo.Update(context.Background(), shopA, salonBrand(), shops.LocationWrite{
		ReplaceAddress: true, Address: str("Calle 10 # 5-20, Bogotá"), Latitude: f64(4.598056), Longitude: f64(-74.076111),
	})
	if err != nil || !res.Found {
		t.Fatalf("Update: found=%v err=%v", res.Found, err)
	}
	got, _, err := repo.Get(context.Background(), shopA)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	loc := got.Location
	if !loc.HasPhysicalLocation || loc.Address == nil || *loc.Address != "Calle 10 # 5-20, Bogotá" ||
		loc.Latitude == nil || *loc.Latitude != 4.598056 || loc.Longitude == nil || *loc.Longitude != -74.076111 {
		t.Fatalf("la ubicación no se guardó tal cual: %+v", loc)
	}
	if *res.Brand.Location.Address != *loc.Address {
		t.Fatalf("RETURNING y Get deben coincidir: %+v vs %+v", res.Brand.Location, loc)
	}
}

// TestBrandUpdate_Location_ToggleKeepsAddressAndCoordinates cubre CA-102-05:
// apagar el interruptor conserva el dato; reactivarlo lo devuelve sin recapturar.
func TestBrandUpdate_Location_ToggleKeepsAddressAndCoordinates(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.NewBrandRepository(db)
	restoreBrand(t, db, shopA)
	ctx := context.Background()

	if _, err := repo.Update(ctx, shopA, salonBrand(), shops.LocationWrite{
		ReplaceAddress: true, Address: str("Calle 10"), Latitude: f64(1.5), Longitude: f64(2.5),
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	off := false
	res, err := repo.Update(ctx, shopA, salonBrand(), shops.LocationWrite{HasPhysicalLocation: &off})
	if err != nil {
		t.Fatalf("apagar: %v", err)
	}
	loc := res.Brand.Location
	if loc.HasPhysicalLocation || loc.Address == nil || *loc.Address != "Calle 10" || loc.Latitude == nil || *loc.Latitude != 1.5 {
		t.Fatalf("apagar no debe borrar dirección ni coordenadas: %+v", loc)
	}

	// Un PATCH de la marca sin tocar la ubicación tampoco la altera.
	res, err = repo.Update(ctx, shopA, salonBrand(), shops.LocationWrite{})
	if err != nil {
		t.Fatalf("sin cambios de ubicación: %v", err)
	}
	if res.Brand.Location.HasPhysicalLocation || res.Brand.Location.Address == nil {
		t.Fatalf("un LocationWrite vacío no debe tocar la ubicación: %+v", res.Brand.Location)
	}

	on := true
	res, err = repo.Update(ctx, shopA, salonBrand(), shops.LocationWrite{HasPhysicalLocation: &on})
	if err != nil {
		t.Fatalf("reactivar: %v", err)
	}
	if !res.Brand.Location.Visible() {
		t.Fatalf("reactivar debe volver a mostrar la ubicación guardada: %+v", res.Brand.Location)
	}
}

func TestBrandUpdate_Location_ClearRemovesAddressAndCoordinatesTogether(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.NewBrandRepository(db)
	restoreBrand(t, db, shopA)
	ctx := context.Background()

	if _, err := repo.Update(ctx, shopA, salonBrand(), shops.LocationWrite{
		ReplaceAddress: true, Address: str("Calle 10"), Latitude: f64(1.5), Longitude: f64(2.5),
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	res, err := repo.Update(ctx, shopA, salonBrand(), shops.LocationWrite{ReplaceAddress: true})
	if err != nil {
		t.Fatalf("quitar: %v", err)
	}
	loc := res.Brand.Location
	if loc.Address != nil || loc.Latitude != nil || loc.Longitude != nil {
		t.Fatalf("quitar la dirección debe quitar también las coordenadas: %+v", loc)
	}
}

// TestBrandUpdate_Location_CheckRejectsInconsistentState: las restricciones
// de la migración son la última línea de defensa si un servicio fallara.
func TestBrandUpdate_Location_CheckRejectsInconsistentState(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.NewBrandRepository(db)
	restoreBrand(t, db, shopA)

	cases := map[string]shops.LocationWrite{
		"coordenadas sin dirección":  {ReplaceAddress: true, Latitude: f64(1), Longitude: f64(2)},
		"media coordenada":           {ReplaceAddress: true, Address: str("x"), Latitude: f64(1)},
		"latitud fuera de rango":     {ReplaceAddress: true, Address: str("x"), Latitude: f64(91), Longitude: f64(0)},
		"dirección de solo espacios": {ReplaceAddress: true, Address: str("   ")},
	}
	for name, w := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := repo.Update(context.Background(), shopA, salonBrand(), w); err == nil {
				t.Fatal("la base debe rechazar el estado inconsistente")
			}
		})
	}
}

// TestBrandUpdate_Location_TwoTenants_NeverCrossesBarbershops: la ubicación de
// una barbería no se lee ni se escribe desde la otra.
func TestBrandUpdate_Location_TwoTenants_NeverCrossesBarbershops(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.NewBrandRepository(db)
	restoreBrand(t, db, shopA)
	restoreBrand(t, db, shopB)
	ctx := context.Background()

	if _, err := repo.Update(ctx, shopA, salonBrand(), shops.LocationWrite{
		ReplaceAddress: true, Address: str("Solo de A"), Latitude: f64(1), Longitude: f64(2),
	}); err != nil {
		t.Fatalf("Update shopA: %v", err)
	}
	gotB, found, err := repo.Get(ctx, shopB)
	if err != nil || !found {
		t.Fatalf("Get shopB: found=%v err=%v", found, err)
	}
	if gotB.Location.Address != nil || gotB.Location.Latitude != nil {
		t.Fatalf("la ubicación de A no debe verse en B: %+v", gotB.Location)
	}
}
