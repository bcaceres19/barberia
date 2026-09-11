// Package postgres_test (pruebas de integración, HU-093) requiere
// PostgreSQL REAL con las migraciones aplicadas (incluida
// 20260911060000_add_barbershop_booking_policy.sql) y
// database/testdata/hu093_politica_reserva.sql cargado. Conéctate como
// barberia_app (docs/03-desarrollo/estrategia-pruebas.md §2 prohíbe mocks
// para RLS/concurrencia optimista real).
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/shops/postgres/...
package postgres_test

import (
	"context"
	"testing"

	"system-barbershop/internal/modules/shops"
	shopspostgres "system-barbershop/internal/modules/shops/postgres"
)

// Par dedicado de hu093_politica_reserva.sql (HU-093): ambas barberías
// quedan en los defaults de DEC-083 (CA-093-01), sin overrides.
const (
	bookingPolicyShopUno = "00930001-0093-0093-0093-009300010001"
	bookingPolicyShopDos = "00930002-0093-0093-0093-009300020002"
)

func TestBookingPolicyGet_Defaults_MatchesDEC083(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.NewBookingPolicyRepository(db)

	policy, found, err := repo.Get(context.Background(), bookingPolicyShopUno)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found {
		t.Fatal("expected bookingPolicyShopUno to be found")
	}
	want := shops.BookingPolicy{
		MinAdvanceMinutes: 60, MaxAdvanceDays: 3, SlotGridMinutes: 15,
		CancellationDeadlineMinutes: 20, LateCancellationClientAllowed: true,
		LateCancellationReasonRequired: true,
	}
	got := policy
	got.VersionToken = "" // comparado aparte: no es un default, es derivado de updated_at.
	if got != want {
		t.Fatalf("expected defaults %+v, got %+v", want, got)
	}
	if policy.VersionToken == "" {
		t.Fatal("expected a non-empty VersionToken")
	}
}

func TestBookingPolicyGet_UnknownBarbershop_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.NewBookingPolicyRepository(db)

	_, found, err := repo.Get(context.Background(), "99999999-9999-9999-9999-999999999999")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if found {
		t.Fatal("expected an unknown barbershop not to be found")
	}
}

// TestBookingPolicyUpdate_ValidVersionToken_PersistsAndRoundTrips cubre
// CA-093-02/CA-093-04: una actualización con el token vigente escribe los
// seis campos y devuelve un token NUEVO (la escritura tocó updated_at);
// una lectura posterior confirma que el valor persistido es exactamente el
// escrito, no una mutación parcial.
func TestBookingPolicyUpdate_ValidVersionToken_PersistsAndRoundTrips(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.NewBookingPolicyRepository(db)

	before, found, err := repo.Get(context.Background(), bookingPolicyShopUno)
	if err != nil || !found {
		t.Fatalf("Get (before): found=%v err=%v", found, err)
	}

	result, err := repo.Update(context.Background(), bookingPolicyShopUno, shops.BookingPolicyUpdateInput{
		MinAdvanceMinutes: 90, MaxAdvanceDays: 5, SlotGridMinutes: 30,
		CancellationDeadlineMinutes: 60, LateCancellationClientAllowed: false,
		LateCancellationReasonRequired: false, ExpectedVersionToken: before.VersionToken,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if result.Outcome != shops.BookingPolicyUpdateOutcomeUpdated {
		t.Fatalf("expected BookingPolicyUpdateOutcomeUpdated, got %v", result.Outcome)
	}
	if result.Policy.MinAdvanceMinutes != 90 || result.Policy.MaxAdvanceDays != 5 ||
		result.Policy.SlotGridMinutes != 30 || result.Policy.CancellationDeadlineMinutes != 60 ||
		result.Policy.LateCancellationClientAllowed || result.Policy.LateCancellationReasonRequired {
		t.Fatalf("unexpected returned policy: %+v", result.Policy)
	}
	if result.Policy.VersionToken == before.VersionToken {
		t.Fatal("expected a new VersionToken after a real write")
	}

	after, found, err := repo.Get(context.Background(), bookingPolicyShopUno)
	if err != nil || !found {
		t.Fatalf("Get (after): found=%v err=%v", found, err)
	}
	if after != result.Policy {
		t.Fatalf("expected the re-read policy to match the update result, got %+v vs %+v", after, result.Policy)
	}

	// Restaura los defaults para no dejar estado mutado entre ejecuciones de
	// la suite (mismo criterio de higiene que restoreShopA en
	// repository_test.go, aplicado aquí con un Update normal en vez de SQL
	// directo: ejercita el mismo camino de código que el resto de la
	// prueba).
	if _, err := repo.Update(context.Background(), bookingPolicyShopUno, shops.BookingPolicyUpdateInput{
		MinAdvanceMinutes: 60, MaxAdvanceDays: 3, SlotGridMinutes: 15,
		CancellationDeadlineMinutes: 20, LateCancellationClientAllowed: true,
		LateCancellationReasonRequired: true, ExpectedVersionToken: after.VersionToken,
	}); err != nil {
		t.Fatalf("Update (restore defaults): %v", err)
	}
}

// TestBookingPolicyUpdate_StaleVersionToken_ReturnsVersionConflictWithoutWriting
// cubre CA-093-02: un token que ya no coincide con la representación
// vigente (porque otra escritura tocó la fila primero) se rechaza SIN
// escribir nada, verificado releyendo el valor tras el intento fallido.
func TestBookingPolicyUpdate_StaleVersionToken_ReturnsVersionConflictWithoutWriting(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.NewBookingPolicyRepository(db)

	staleRead, found, err := repo.Get(context.Background(), bookingPolicyShopDos)
	if err != nil || !found {
		t.Fatalf("Get (stale read): found=%v err=%v", found, err)
	}

	// Una escritura real intermedia invalida staleRead.VersionToken.
	fresh, err := repo.Update(context.Background(), bookingPolicyShopDos, shops.BookingPolicyUpdateInput{
		MinAdvanceMinutes: 45, MaxAdvanceDays: 7, SlotGridMinutes: 20,
		CancellationDeadlineMinutes: 30, LateCancellationClientAllowed: true,
		LateCancellationReasonRequired: true, ExpectedVersionToken: staleRead.VersionToken,
	})
	if err != nil || fresh.Outcome != shops.BookingPolicyUpdateOutcomeUpdated {
		t.Fatalf("Update (intermediate write): outcome=%v err=%v", fresh.Outcome, err)
	}

	result, err := repo.Update(context.Background(), bookingPolicyShopDos, shops.BookingPolicyUpdateInput{
		MinAdvanceMinutes: 999, MaxAdvanceDays: 999, SlotGridMinutes: 5,
		CancellationDeadlineMinutes: 999, LateCancellationClientAllowed: true,
		LateCancellationReasonRequired: true, ExpectedVersionToken: staleRead.VersionToken,
	})
	if err != nil {
		t.Fatalf("Update (stale): %v", err)
	}
	if result.Outcome != shops.BookingPolicyUpdateOutcomeVersionConflict {
		t.Fatalf("expected BookingPolicyUpdateOutcomeVersionConflict, got %v", result.Outcome)
	}

	current, found, err := repo.Get(context.Background(), bookingPolicyShopDos)
	if err != nil || !found {
		t.Fatalf("Get (after stale attempt): found=%v err=%v", found, err)
	}
	if current.MinAdvanceMinutes != 45 || current.MaxAdvanceDays != 7 || current.SlotGridMinutes != 20 || current.CancellationDeadlineMinutes != 30 {
		t.Fatalf("expected the stale update to write nothing, got %+v", current)
	}

	// Restaura los defaults.
	if _, err := repo.Update(context.Background(), bookingPolicyShopDos, shops.BookingPolicyUpdateInput{
		MinAdvanceMinutes: 60, MaxAdvanceDays: 3, SlotGridMinutes: 15,
		CancellationDeadlineMinutes: 20, LateCancellationClientAllowed: true,
		LateCancellationReasonRequired: true, ExpectedVersionToken: current.VersionToken,
	}); err != nil {
		t.Fatalf("Update (restore defaults): %v", err)
	}
}

// TestBookingPolicyUpdate_UnknownBarbershop_ReturnsNotFound cubre el
// desenlace defensivo BookingPolicyUpdateOutcomeNotFound.
func TestBookingPolicyUpdate_UnknownBarbershop_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.NewBookingPolicyRepository(db)

	result, err := repo.Update(context.Background(), "99999999-9999-9999-9999-999999999999", shops.BookingPolicyUpdateInput{
		MinAdvanceMinutes: 60, MaxAdvanceDays: 3, SlotGridMinutes: 15,
		CancellationDeadlineMinutes: 20, LateCancellationClientAllowed: true,
		LateCancellationReasonRequired: true, ExpectedVersionToken: "cualquier-token",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if result.Outcome != shops.BookingPolicyUpdateOutcomeNotFound {
		t.Fatalf("expected BookingPolicyUpdateOutcomeNotFound, got %v", result.Outcome)
	}
}

// TestBookingPolicyGetAndUpdate_TenantAIsolatedFromTenantB confirma que
// actualizar la política de Uno nunca afecta la de Dos (RN-TEN-01),
// verificado con dos tenants reales.
func TestBookingPolicyGetAndUpdate_TenantAIsolatedFromTenantB(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.NewBookingPolicyRepository(db)

	dosBefore, found, err := repo.Get(context.Background(), bookingPolicyShopDos)
	if err != nil || !found {
		t.Fatalf("Get(shopDos, before): found=%v err=%v", found, err)
	}

	unoBefore, found, err := repo.Get(context.Background(), bookingPolicyShopUno)
	if err != nil || !found {
		t.Fatalf("Get(shopUno): found=%v err=%v", found, err)
	}
	result, err := repo.Update(context.Background(), bookingPolicyShopUno, shops.BookingPolicyUpdateInput{
		MinAdvanceMinutes: 120, MaxAdvanceDays: 10, SlotGridMinutes: 60,
		CancellationDeadlineMinutes: 120, LateCancellationClientAllowed: false,
		LateCancellationReasonRequired: false, ExpectedVersionToken: unoBefore.VersionToken,
	})
	if err != nil || result.Outcome != shops.BookingPolicyUpdateOutcomeUpdated {
		t.Fatalf("Update(shopUno): outcome=%v err=%v", result.Outcome, err)
	}

	dosAfter, found, err := repo.Get(context.Background(), bookingPolicyShopDos)
	if err != nil || !found {
		t.Fatalf("Get(shopDos, after): found=%v err=%v", found, err)
	}
	if dosAfter != dosBefore {
		t.Fatalf("RN-TEN-01: updating shopUno changed shopDos's policy: before=%+v after=%+v", dosBefore, dosAfter)
	}

	// Restaura los defaults de Uno.
	if _, err := repo.Update(context.Background(), bookingPolicyShopUno, shops.BookingPolicyUpdateInput{
		MinAdvanceMinutes: 60, MaxAdvanceDays: 3, SlotGridMinutes: 15,
		CancellationDeadlineMinutes: 20, LateCancellationClientAllowed: true,
		LateCancellationReasonRequired: true, ExpectedVersionToken: result.Policy.VersionToken,
	}); err != nil {
		t.Fatalf("Update (restore shopUno defaults): %v", err)
	}
}

// TestBookingPolicyUpdate_InvalidValueBypassingService_HitsCheckConstraint
// es una prueba de defensa en profundidad: aunque BookingPolicyService ya
// valida rangos y coherencia antes de llegar aquí, esta prueba confirma
// que la base de datos también los exige de forma independiente
// (barbershop_slot_grid_minutes_ck) si esa validación previa alguna vez
// tuviera un defecto.
func TestBookingPolicyUpdate_InvalidValueBypassingService_HitsCheckConstraint(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.NewBookingPolicyRepository(db)

	current, found, err := repo.Get(context.Background(), bookingPolicyShopUno)
	if err != nil || !found {
		t.Fatalf("Get: found=%v err=%v", found, err)
	}

	_, err = repo.Update(context.Background(), bookingPolicyShopUno, shops.BookingPolicyUpdateInput{
		MinAdvanceMinutes: 60, MaxAdvanceDays: 3, SlotGridMinutes: 7, // fuera del conjunto discreto.
		CancellationDeadlineMinutes: 20, LateCancellationClientAllowed: true,
		LateCancellationReasonRequired: true, ExpectedVersionToken: current.VersionToken,
	})
	if err == nil {
		t.Fatal("expected the database CHECK constraint to reject slotGridMinutes=7")
	}

	unchanged, found, getErr := repo.Get(context.Background(), bookingPolicyShopUno)
	if getErr != nil || !found {
		t.Fatalf("Get (after rejected write): found=%v err=%v", found, getErr)
	}
	if unchanged.SlotGridMinutes != current.SlotGridMinutes {
		t.Fatalf("expected zero write on a CHECK violation, got %+v", unchanged)
	}
}
