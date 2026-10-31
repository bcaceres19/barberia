// Pruebas de integración de HU-024 (Repository.Deactivate/Reactivate) contra
// PostgreSQL REAL, mismos requisitos que repository_test.go: migraciones
// aplicadas hasta 20260824150000_create_barber_service.sql y
// database/testdata/hu022_catalogo.sql cargado. Conéctate como
// barberia_app.
package postgres_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"system-barbershop/internal/modules/catalog"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

func deactivateFP(serviceID string) idempotency.Fingerprint {
	return idempotency.ComputeFingerprint("POST", "/api/v1/private/services/"+serviceID+"/deactivate", nil)
}

func reactivateFP(serviceID string) idempotency.Fingerprint {
	return idempotency.ComputeFingerprint("POST", "/api/v1/private/services/"+serviceID+"/reactivate", nil)
}

// --- CA-024-02/CA-024-03: desactivar conserva el recurso -------------------

func TestDeactivate_ActiveService_SetsInactiveWithTimestamp(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	svc := createService(t, repo, shopE, "Corte a desactivar "+suffix)
	key := idempotency.Key("deactivate-" + suffix)

	result, err := repo.Deactivate(context.Background(), string(shopE), svc.ID, key, deactivateFP(svc.ID))
	if err != nil {
		t.Fatalf("Deactivate: %v", err)
	}
	if result.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected OutcomeProceed, got %s", result.Decision.Outcome)
	}
	if result.InvalidTransition {
		t.Fatal("did not expect InvalidTransition on the first deactivation of an active service")
	}
	if result.Service.IsActive {
		t.Fatal("expected IsActive=false after Deactivate")
	}
	if result.Service.DeactivatedAt == nil {
		t.Fatal("expected DeactivatedAt to be set after Deactivate")
	}
	if result.Service.ID != svc.ID || result.Service.Name != svc.Name || result.Service.DurationMinutes != svc.DurationMinutes || result.Service.PriceCents != svc.PriceCents {
		t.Fatalf("CA-024-02: deactivate must not alter identity/name/duration/price: got %+v, want same as %+v", result.Service, svc)
	}

	got, found, err := repo.Get(context.Background(), string(shopE), svc.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found {
		t.Fatal("CA-024-02: the service must still exist (no physical delete) after deactivation")
	}
	if got.IsActive || got.DeactivatedAt == nil {
		t.Fatalf("expected persisted row to be inactive with a timestamp, got %+v", got)
	}
}

func TestDeactivate_Repeated_SameKey_ReturnsSameResponseWithoutSecondEffect(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	svc := createService(t, repo, shopE, "Corte repetido "+suffix)
	key := idempotency.Key("deactivate-repeat-" + suffix)
	fp := deactivateFP(svc.ID)

	first, err := repo.Deactivate(context.Background(), string(shopE), svc.ID, key, fp)
	if err != nil {
		t.Fatalf("Deactivate (first): %v", err)
	}
	if first.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected first call to proceed, got %s", first.Decision.Outcome)
	}

	second, err := repo.Deactivate(context.Background(), string(shopE), svc.ID, key, fp)
	if err != nil {
		t.Fatalf("Deactivate (second): %v", err)
	}
	if second.Decision.Outcome != idempotency.OutcomeReplay {
		t.Fatalf("CA-024-06: expected OutcomeReplay on repetition with the same key, got %s", second.Decision.Outcome)
	}
	if second.Response.Body != first.Response.Body {
		t.Fatalf("CA-004-01: expected byte-identical replay body, got %q vs %q", second.Response.Body, first.Response.Body)
	}
}

func TestDeactivate_AlreadyInactive_NewKey_ReturnsInvalidTransitionWithoutChangingRow(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	svc := createService(t, repo, shopE, "Corte ya inactivo "+suffix)
	first, err := repo.Deactivate(context.Background(), string(shopE), svc.ID, idempotency.Key("deactivate-a-"+suffix), deactivateFP(svc.ID))
	if err != nil || first.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("setup: expected the first deactivation to proceed, got %+v err=%v", first, err)
	}

	// CA-024-06: una clave NUEVA sobre una transición que ya no aplica
	// (el servicio ya está inactivo) es una transición inválida, no un
	// segundo efecto silencioso.
	second, err := repo.Deactivate(context.Background(), string(shopE), svc.ID, idempotency.Key("deactivate-b-"+suffix), deactivateFP(svc.ID))
	if err != nil {
		t.Fatalf("Deactivate (second, new key): %v", err)
	}
	if second.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected the new key to also reach OutcomeProceed (it is a fresh claim), got %s", second.Decision.Outcome)
	}
	if !second.Found {
		t.Fatal("expected Found=true: the service exists, it is just already inactive")
	}
	if !second.InvalidTransition {
		t.Fatal("expected InvalidTransition=true: the service was already inactive")
	}

	got, found, err := repo.Get(context.Background(), string(shopE), svc.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found || got.DeactivatedAt == nil || !first.Service.DeactivatedAt.Equal(*got.DeactivatedAt) {
		t.Fatalf("expected the row to keep its original deactivatedAt (rejected transition changed nothing), got %+v", got)
	}
}

func TestDeactivate_UnknownID_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	const unknownID = "00000000-0000-0000-0000-000000000000"
	result, err := repo.Deactivate(context.Background(), string(shopE), unknownID, idempotency.Key("deactivate-missing-"+suffix), deactivateFP(unknownID))
	if err != nil {
		t.Fatalf("Deactivate: %v", err)
	}
	if result.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected OutcomeProceed (fresh claim, rolled back), got %s", result.Decision.Outcome)
	}
	if result.Found {
		t.Fatal("CA-024-07: expected Found=false for an unknown serviceId")
	}
}

func TestDeactivate_CrossTenant_NeverLeaksAnotherShopsService(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	svc := createService(t, repo, shopF, "De la barbería F "+suffix)

	result, err := repo.Deactivate(context.Background(), string(shopE), svc.ID, idempotency.Key("deactivate-cross-"+suffix), deactivateFP(svc.ID))
	if err != nil {
		t.Fatalf("Deactivate: %v", err)
	}
	if result.Found {
		t.Fatal("CA-024-07/RN-TEN-01: shopE must not be able to deactivate shopF's service")
	}

	got, found, err := repo.Get(context.Background(), string(shopF), svc.ID)
	if err != nil {
		t.Fatalf("Get (owner tenant): %v", err)
	}
	if !found || !got.IsActive {
		t.Fatal("expected shopF's service to remain untouched and active")
	}
}

// --- CA-024-05: reactivar no recrea ni altera catálogo ----------------------

func TestReactivate_InactiveService_SetsActiveClearsTimestamp(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	svc := createService(t, repo, shopE, "Corte a reactivar "+suffix)
	deactivated, err := repo.Deactivate(context.Background(), string(shopE), svc.ID, idempotency.Key("deactivate-for-reactivate-"+suffix), deactivateFP(svc.ID))
	if err != nil || deactivated.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("setup: expected deactivation to proceed, got %+v err=%v", deactivated, err)
	}

	result, err := repo.Reactivate(context.Background(), string(shopE), svc.ID, idempotency.Key("reactivate-"+suffix), reactivateFP(svc.ID))
	if err != nil {
		t.Fatalf("Reactivate: %v", err)
	}
	if result.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("expected OutcomeProceed, got %s", result.Decision.Outcome)
	}
	if result.InvalidTransition {
		t.Fatal("did not expect InvalidTransition when reactivating an inactive service")
	}
	if !result.Service.IsActive {
		t.Fatal("expected IsActive=true after Reactivate")
	}
	if result.Service.DeactivatedAt != nil {
		t.Fatalf("expected DeactivatedAt=nil after Reactivate, got %v", *result.Service.DeactivatedAt)
	}
	// CA-024-05: nunca crea otra fila ni altera duración/precio.
	if result.Service.ID != svc.ID || result.Service.DurationMinutes != svc.DurationMinutes || result.Service.PriceCents != svc.PriceCents {
		t.Fatalf("expected the SAME row with unchanged duration/price, got %+v want id=%s duration=%d price=%d",
			result.Service, svc.ID, svc.DurationMinutes, svc.PriceCents)
	}
}

// RN-SER-01: mientras un servicio estuvo inactivo, otro activo pudo tomar su
// nombre; reactivarlo es un conflicto de nombre controlado, no un error de
// servidor, no escribe nada y deja libre la clave de idempotencia.
func TestReactivate_NameTakenByAnotherActiveService_ReportsNameTakenWithoutWriting(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	original := createService(t, repo, shopE, "Corte repetido "+suffix)
	if _, err := repo.Deactivate(context.Background(), string(shopE), original.ID,
		idempotency.Key("deactivate-dup-"+suffix), deactivateFP(original.ID)); err != nil {
		t.Fatalf("Deactivate: %v", err)
	}
	createService(t, repo, shopE, "Corte repetido "+suffix) // toma el nombre libre

	key := idempotency.Key("reactivate-dup-" + suffix)
	result, err := repo.Reactivate(context.Background(), string(shopE), original.ID, key, reactivateFP(original.ID))
	if err != nil {
		t.Fatalf("Reactivate must report the conflict, not fail: %v", err)
	}
	if !result.NameTaken {
		t.Fatalf("expected NameTaken, got %+v", result)
	}

	got, found, err := repo.Get(context.Background(), string(shopE), original.ID)
	if err != nil || !found {
		t.Fatalf("Get: found=%v err=%v", found, err)
	}
	if got.IsActive {
		t.Fatal("a name conflict must leave the service inactive")
	}

	// La clave quedó libre: tras liberar el nombre, el mismo reintento funciona.
	renamed := "Otro nombre " + suffix
	if _, err := repo.Update(context.Background(), string(shopE), original.ID, catalog.UpdateFields{Name: &renamed}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	retry, err := repo.Reactivate(context.Background(), string(shopE), original.ID, key, reactivateFP(original.ID))
	if err != nil || retry.NameTaken || !retry.Service.IsActive {
		t.Fatalf("expected the retry with the same key to reactivate, got %+v (err %v)", retry, err)
	}
}

func TestReactivate_AlreadyActive_NewKey_ReturnsInvalidTransition(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	svc := createService(t, repo, shopE, "Corte siempre activo "+suffix)

	result, err := repo.Reactivate(context.Background(), string(shopE), svc.ID, idempotency.Key("reactivate-noop-"+suffix), reactivateFP(svc.ID))
	if err != nil {
		t.Fatalf("Reactivate: %v", err)
	}
	if !result.Found {
		t.Fatal("expected Found=true: the service exists")
	}
	if !result.InvalidTransition {
		t.Fatal("expected InvalidTransition=true: the service was already active")
	}
}

func TestReactivate_UnknownID_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	const unknownID = "00000000-0000-0000-0000-000000000000"
	result, err := repo.Reactivate(context.Background(), string(shopE), unknownID, idempotency.Key("reactivate-missing-"+suffix), reactivateFP(unknownID))
	if err != nil {
		t.Fatalf("Reactivate: %v", err)
	}
	if result.Found {
		t.Fatal("CA-024-07: expected Found=false for an unknown serviceId")
	}
}

// --- CA-024-06/DEC-069: carrera real, sin sleep como sincronización --------

// TestDeactivate_TwoRealConcurrentConnections_ExactlyOneSucceeds confirma el
// mecanismo TOCTOU exigido por el trabajo requerido §2.4: dos conexiones
// reales confirmando la desactivación del MISMO servicio al mismo tiempo
// (dos claves de idempotencia DISTINTAS, así que ninguna puede resolverse
// por simple repetición) nunca dejan la fila en un estado inconsistente:
// exactamente una transiciona is_active; la otra descubre, tras esperar el
// FOR UPDATE de la primera, que la transición ya no aplica
// (InvalidTransition), nunca un error ni una segunda escritura.
func TestDeactivate_TwoRealConcurrentConnections_ExactlyOneSucceeds(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	svc := createService(t, repo, shopE, "Corte en carrera "+suffix)

	var (
		wg               sync.WaitGroup
		aResult, bResult catalog.LifecycleResult
		aErr, bErr       error
		aStarted         = make(chan struct{})
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		close(aStarted)
		aResult, aErr = repo.Deactivate(context.Background(), string(shopE), svc.ID, idempotency.Key("race-a-"+suffix), deactivateFP(svc.ID))
	}()
	go func() {
		defer wg.Done()
		<-aStarted
		bResult, bErr = repo.Deactivate(context.Background(), string(shopE), svc.ID, idempotency.Key("race-b-"+suffix), deactivateFP(svc.ID))
	}()
	wg.Wait()

	if aErr != nil {
		t.Fatalf("goroutine A: %v", aErr)
	}
	if bErr != nil {
		t.Fatalf("goroutine B: %v", bErr)
	}

	succeeded := aResult.Decision.Outcome == idempotency.OutcomeProceed && !aResult.InvalidTransition
	rejected := bResult.Decision.Outcome == idempotency.OutcomeProceed && bResult.InvalidTransition
	if !succeeded || !rejected {
		// El orden real de ejecución no está garantizado: si B ganó la
		// carrera, los roles se invierten.
		succeeded = bResult.Decision.Outcome == idempotency.OutcomeProceed && !bResult.InvalidTransition
		rejected = aResult.Decision.Outcome == idempotency.OutcomeProceed && aResult.InvalidTransition
	}
	if !succeeded || !rejected {
		t.Fatalf("expected exactly one winner (deactivated) and one InvalidTransition, got a=%+v b=%+v", aResult, bResult)
	}

	got, found, err := repo.Get(context.Background(), string(shopE), svc.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found || got.IsActive || got.DeactivatedAt == nil {
		t.Fatalf("expected the service to end up inactive exactly once, got %+v", got)
	}
}

// --- RN-SER-03 / CA-024-01 / CA-024-04: impacto real de desactivar ---------

// insertAppointmentFixture crea una cita para el recuento de impacto; el
// estado y el inicio son lo único que el recuento distingue.
func insertAppointmentFixture(t *testing.T, db *database.DB, shop database.BarbershopID, barberID, serviceID, customerID, status string, startsAt time.Time) {
	t.Helper()
	err := db.InTenantTx(context.Background(), shop, func(ctx context.Context, q database.Queries) error {
		_, err := q.Exec(ctx,
			`INSERT INTO appointment (barbershop_id, barber_id, service_id, customer_id, attendee_name,
			                          starts_at, ends_at, status, origin,
			                          service_name_snapshot, duration_minutes_snapshot,
			                          price_amount_snapshot, price_currency_snapshot, resolved_at)
			 VALUES ($1, $2, $3, $4, 'Persona atendida', $5, $6, $7, 'manual', 'Servicio', 30, 20000.00, 'COP',
			         CASE WHEN $7 = 'confirmed' THEN NULL ELSE now() END)`,
			string(shop), barberID, serviceID, customerID, startsAt, startsAt.Add(30*time.Minute), status)
		return err
	})
	if err != nil {
		t.Fatalf("insertAppointmentFixture(%s): %v", status, err)
	}
}

func createCustomerFixture(t *testing.T, db *database.DB, shop database.BarbershopID) string {
	t.Helper()
	var id string
	err := db.InTenantTx(context.Background(), shop, func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx,
			`INSERT INTO customer (barbershop_id, full_name) VALUES ($1, 'Cliente de prueba') RETURNING id`,
			string(shop)).Scan(&id)
	})
	if err != nil {
		t.Fatalf("createCustomerFixture: %v", err)
	}
	return id
}

// El recuento solo incluye citas `confirmed` futuras de ESE servicio y de la
// barbería de la sesión; desactivar no toca ninguna cita (CA-024-03).
func TestCountFutureAppointments_CountsOnlyConfirmedFutureOfThatServiceAndTenant(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	target := createService(t, repo, shopE, "Servicio con citas "+suffix)
	other := createService(t, repo, shopE, "Otro servicio "+suffix)
	barber := createBarberFixture(t, db, shopE, "Barbero impacto "+suffix)
	customer := createCustomerFixture(t, db, shopE)
	now := time.Now()

	insertAppointmentFixture(t, db, shopE, barber.ID, target.ID, customer, "confirmed", now.Add(48*time.Hour))
	insertAppointmentFixture(t, db, shopE, barber.ID, target.ID, customer, "confirmed", now.Add(72*time.Hour))
	insertAppointmentFixture(t, db, shopE, barber.ID, target.ID, customer, "confirmed", now.Add(-48*time.Hour))
	insertAppointmentFixture(t, db, shopE, barber.ID, target.ID, customer, "cancelled_by_barber", now.Add(96*time.Hour))
	insertAppointmentFixture(t, db, shopE, barber.ID, other.ID, customer, "confirmed", now.Add(120*time.Hour))

	got, err := repo.CountFutureAppointments(context.Background(), string(shopE), target.ID)
	if err != nil {
		t.Fatalf("CountFutureAppointments: %v", err)
	}
	if got != 2 {
		t.Fatalf("expected 2 confirmed future appointments of the service, got %d", got)
	}

	foreign, err := repo.CountFutureAppointments(context.Background(), string(shopF), target.ID)
	if err != nil {
		t.Fatalf("CountFutureAppointments from another tenant: %v", err)
	}
	if foreign != 0 {
		t.Fatalf("RN-TEN-01: another tenant must not count the appointments, got %d", foreign)
	}

	// CA-024-04: la confirmación repite la consulta en su propia transacción.
	result, err := repo.Deactivate(context.Background(), string(shopE), target.ID,
		idempotency.Key("deactivate-impact-"+suffix), deactivateFP(target.ID))
	if err != nil {
		t.Fatalf("Deactivate: %v", err)
	}
	if !strings.Contains(result.Response.Body, `"affectedAppointments":2`) {
		t.Fatalf("expected the confirmation to report 2 affected appointments, got %s", result.Response.Body)
	}

	// CA-024-03: ninguna cita se cancela ni se modifica.
	after, err := repo.CountFutureAppointments(context.Background(), string(shopE), target.ID)
	if err != nil || after != 2 {
		t.Fatalf("deactivating must not touch appointments: count=%d err=%v", after, err)
	}
}

// --- HU-029 / DEC-004: aplicar el precio nuevo a los turnos futuros ---------

const staffPriceE = "eeeeee01-eeee-4eee-8eee-eeeeeeeeee01"

func appointmentPrice(t *testing.T, db *database.DB, shop database.BarbershopID, appointmentID string) string {
	t.Helper()
	var price string
	err := db.InTenantTx(context.Background(), shop, func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx, `SELECT price_amount_snapshot::text FROM appointment WHERE id = $1`, appointmentID).Scan(&price)
	})
	if err != nil {
		t.Fatalf("appointmentPrice: %v", err)
	}
	return price
}

func appointmentIDs(t *testing.T, db *database.DB, shop database.BarbershopID, serviceID string) []string {
	t.Helper()
	var ids []string
	err := db.InTenantTx(context.Background(), shop, func(ctx context.Context, q database.Queries) error {
		rows, err := q.Query(ctx, `SELECT id::text FROM appointment WHERE service_id = $1 ORDER BY starts_at`, serviceID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}
			ids = append(ids, id)
		}
		return rows.Err()
	})
	if err != nil {
		t.Fatalf("appointmentIDs: %v", err)
	}
	return ids
}

func priceHistory(t *testing.T, db *database.DB, shop database.BarbershopID, appointmentID string) (events int, previous, next string) {
	t.Helper()
	err := db.InTenantTx(context.Background(), shop, func(ctx context.Context, q database.Queries) error {
		if err := q.QueryRow(ctx,
			`SELECT count(*) FROM appointment_history WHERE appointment_id = $1 AND event_type = 'appointment_service_changed'`,
			appointmentID).Scan(&events); err != nil {
			return err
		}
		if events == 0 {
			return nil
		}
		return q.QueryRow(ctx,
			`SELECT c.previous_value, c.new_value
			   FROM appointment_history_change c
			   JOIN appointment_history h ON h.barbershop_id = c.barbershop_id AND h.id = c.history_id
			  WHERE h.appointment_id = $1 AND c.field_name = 'price_amount_snapshot'`,
			appointmentID).Scan(&previous, &next)
	})
	if err != nil {
		t.Fatalf("priceHistory: %v", err)
	}
	return events, previous, next
}

func TestUpdate_ApplyPriceToUpcoming_ChangesOnlyConfirmedFutureOfThatService(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	target := createService(t, repo, shopE, "Precio a futuras "+suffix)
	other := createService(t, repo, shopE, "Precio otro "+suffix)
	barber := createBarberFixture(t, db, shopE, "Barbero precio "+suffix)
	customer := createCustomerFixture(t, db, shopE)
	now := time.Now()

	insertAppointmentFixture(t, db, shopE, barber.ID, target.ID, customer, "confirmed", now.Add(48*time.Hour))
	insertAppointmentFixture(t, db, shopE, barber.ID, target.ID, customer, "confirmed", now.Add(72*time.Hour))
	insertAppointmentFixture(t, db, shopE, barber.ID, target.ID, customer, "confirmed", now.Add(-48*time.Hour))
	insertAppointmentFixture(t, db, shopE, barber.ID, target.ID, customer, "cancelled_by_barber", now.Add(96*time.Hour))
	insertAppointmentFixture(t, db, shopE, barber.ID, other.ID, customer, "confirmed", now.Add(120*time.Hour))
	ids := appointmentIDs(t, db, shopE, target.ID)
	otherIDs := appointmentIDs(t, db, shopE, other.ID)

	newPrice := int64(3_500_000) // 35000.00
	result, err := repo.Update(context.Background(), string(shopE), target.ID, catalog.UpdateFields{
		PriceCents: &newPrice, ApplyPriceToUpcoming: true, ActorStaffUserID: staffPriceE,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if !result.Found || result.UpcomingPricesUpdated != 2 {
		t.Fatalf("expected found and 2 upcoming appointments updated, got %+v", result)
	}

	// ids viene ordenado por inicio: pasada, 48 h, 72 h, cancelada (96 h).
	wantPrices := []string{"20000.00", "35000.00", "35000.00", "20000.00"}
	if len(ids) != len(wantPrices) {
		t.Fatalf("expected %d appointments, got %d", len(wantPrices), len(ids))
	}
	for i, id := range ids {
		if got := appointmentPrice(t, db, shopE, id); got != wantPrices[i] {
			t.Errorf("appointment %d: expected price %s, got %s", i, wantPrices[i], got)
		}
	}
	if got := appointmentPrice(t, db, shopE, otherIDs[0]); got != "20000.00" {
		t.Errorf("another service's appointment must keep its price, got %s", got)
	}

	events, previous, next := priceHistory(t, db, shopE, ids[1])
	if events != 1 || previous != "20000.00" || next != "35000.00" {
		t.Errorf("expected one price event 20000.00 -> 35000.00, got %d %s -> %s", events, previous, next)
	}
	if events, _, _ := priceHistory(t, db, shopE, ids[0]); events != 0 {
		t.Errorf("a past appointment must get no history event, got %d", events)
	}
}

func TestUpdate_WithoutApplyPriceToUpcoming_LeavesAppointmentsUntouched(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	svc := createService(t, repo, shopE, "Precio sin propagar "+suffix)
	barber := createBarberFixture(t, db, shopE, "Barbero sin propagar "+suffix)
	customer := createCustomerFixture(t, db, shopE)
	insertAppointmentFixture(t, db, shopE, barber.ID, svc.ID, customer, "confirmed", time.Now().Add(48*time.Hour))

	newPrice := int64(4_000_000)
	result, err := repo.Update(context.Background(), string(shopE), svc.ID, catalog.UpdateFields{PriceCents: &newPrice})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if result.UpcomingPricesUpdated != 0 {
		t.Fatalf("expected no appointments updated, got %d", result.UpcomingPricesUpdated)
	}
	id := appointmentIDs(t, db, shopE, svc.ID)[0]
	if got := appointmentPrice(t, db, shopE, id); got != "20000.00" {
		t.Fatalf("RN-SER-04: the appointment must keep its price, got %s", got)
	}
	if events, _, _ := priceHistory(t, db, shopE, id); events != 0 {
		t.Fatalf("expected no history event, got %d", events)
	}
}

func TestUpdate_ApplyPriceToUpcoming_NeverCrossesTenants(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	svc := createService(t, repo, shopE, "Precio aislado "+suffix)
	barber := createBarberFixture(t, db, shopE, "Barbero aislado "+suffix)
	customer := createCustomerFixture(t, db, shopE)
	insertAppointmentFixture(t, db, shopE, barber.ID, svc.ID, customer, "confirmed", time.Now().Add(48*time.Hour))

	newPrice := int64(5_000_000)
	result, err := repo.Update(context.Background(), string(shopF), svc.ID, catalog.UpdateFields{
		PriceCents: &newPrice, ApplyPriceToUpcoming: true, ActorStaffUserID: staffPriceE,
	})
	if err != nil {
		t.Fatalf("Update from another tenant: %v", err)
	}
	if result.Found || result.UpcomingPricesUpdated != 0 {
		t.Fatalf("another tenant must neither find the service nor change its appointments, got %+v", result)
	}
	id := appointmentIDs(t, db, shopE, svc.ID)[0]
	if got := appointmentPrice(t, db, shopE, id); got != "20000.00" {
		t.Fatalf("the appointment of shop E must keep its price, got %s", got)
	}
}
