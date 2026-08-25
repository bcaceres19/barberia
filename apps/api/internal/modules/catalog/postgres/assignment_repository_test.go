// Package postgres_test (pruebas de integración, HU-023) requiere
// PostgreSQL REAL con las doce migraciones aplicadas (incluida
// 20260824150000_create_barber_service.sql) y database/testdata/
// dos_barberias.sql + database/testdata/hu023_asignaciones.sql cargados,
// además de un barber y un service reales por prueba. Conéctate como
// barberia_app (docs/03-desarrollo/estrategia-pruebas.md §2 prohíbe mocks
// para RLS/FK/carreras reales).
//
// Los barberos de fixture se insertan con SQL directo (no vía
// staffpostgres.Repository): este paquete no importa el módulo staff bajo
// ninguna circunstancia, ni siquiera en pruebas (internal/platform/archtest/
// module_boundary_test.go lo verifica) -mismo espíritu que HU-023 exige
// para el propio núcleo (colaboración solo por BarberPort). Los servicios sí
// se crean vía catalogpostgres.Repository.Create: ES el mismo módulo bajo
// prueba, no una dependencia cruzada.
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/catalog/postgres/...
package postgres_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"system-barbershop/internal/modules/catalog"
	catalogpostgres "system-barbershop/internal/modules/catalog/postgres"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

const (
	// shopG/shopH son las dos barberías dedicadas de
	// testdata/hu023_asignaciones.sql: aisladas de dos_barberias.sql/
	// hu021_barberos.sql/hu022_catalogo.sql para que un conteo exacto de
	// asignaciones no dependa de lo que otras suites hagan sobre
	// shopA-shopF.
	shopG = database.BarbershopID("77777777-7777-7777-7777-777777777777")
	shopH = database.BarbershopID("88888888-8888-8888-8888-888888888888")
)

func newAssignmentRepository(db *database.DB) *catalogpostgres.AssignmentRepository {
	return catalogpostgres.NewAssignmentRepository(db)
}

// fixtureBarber es el subconjunto mínimo de un barbero real que este
// paquete necesita para sus fixtures: solo el id, nunca importando
// staff.Barber (ver nota de paquete más arriba).
type fixtureBarber struct {
	ID string
}

func createBarberFixture(t *testing.T, db *database.DB, shop database.BarbershopID, name string) fixtureBarber {
	t.Helper()
	var id string
	err := db.InTenantTx(context.Background(), shop, func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx,
			`INSERT INTO barber (barbershop_id, full_name) VALUES ($1, $2) RETURNING id`,
			string(shop), name,
		).Scan(&id)
	})
	if err != nil {
		t.Fatalf("createBarberFixture: %v", err)
	}
	return fixtureBarber{ID: id}
}

func createServiceFixture(t *testing.T, db *database.DB, shop database.BarbershopID, name string) catalog.Service {
	t.Helper()
	repo := catalogpostgres.New(db, idempotency.NewSQLCoordinator())
	key := idempotency.Key("assignment-fixture-service-" + uniqueSuffix(t))
	fp := idempotency.ComputeFingerprint("POST", "/api/v1/private/services", []byte(`{"name":"`+name+`"}`))
	result, err := repo.Create(context.Background(), string(shop), validInput(name), key, fp)
	if err != nil {
		t.Fatalf("createServiceFixture: %v", err)
	}
	if result.NameTaken {
		t.Fatalf("createServiceFixture: unexpected name conflict for %q", name)
	}
	if result.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("createServiceFixture: expected OutcomeProceed, got %s", result.Decision.Outcome)
	}
	return result.Service
}

// --- Assign: CA-023-02, CA-023-03, CA-023-04 -----------------------------

func TestAssign_New_CreatesRowAndIsVisibleInList(t *testing.T) {
	db := setupTestDB(t)
	repo := newAssignmentRepository(db)
	suffix := uniqueSuffix(t)

	barber := createBarberFixture(t, db, shopG, "Barbero Asignación "+suffix)
	service := createServiceFixture(t, db, shopG, "Servicio Asignación "+suffix)

	result, err := repo.Assign(context.Background(), string(shopG), barber.ID, service.ID)
	if err != nil {
		t.Fatalf("Assign: %v", err)
	}
	if result.Outcome != catalog.AssignOutcomeCreated {
		t.Fatalf("expected AssignOutcomeCreated, got %v", result.Outcome)
	}
	if result.Assignment.BarberID != barber.ID || result.Assignment.ServiceID != service.ID {
		t.Fatalf("unexpected assignment: %+v", result.Assignment)
	}

	page, err := repo.List(context.Background(), string(shopG), barber.ID, nil, 50)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, item := range page.Items {
		if item.ServiceID == service.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected the new assignment to appear in List, got %+v", page.Items)
	}
}

func TestAssign_Repeated_DoesNotCreateASecondRow(t *testing.T) {
	db := setupTestDB(t)
	repo := newAssignmentRepository(db)
	suffix := uniqueSuffix(t)

	barber := createBarberFixture(t, db, shopG, "Barbero Repetido "+suffix)
	service := createServiceFixture(t, db, shopG, "Servicio Repetido "+suffix)

	first, err := repo.Assign(context.Background(), string(shopG), barber.ID, service.ID)
	if err != nil {
		t.Fatalf("Assign (first): %v", err)
	}
	if first.Outcome != catalog.AssignOutcomeCreated {
		t.Fatalf("expected first call to create, got %v", first.Outcome)
	}

	second, err := repo.Assign(context.Background(), string(shopG), barber.ID, service.ID)
	if err != nil {
		t.Fatalf("Assign (second): %v", err)
	}
	if second.Outcome != catalog.AssignOutcomeAlreadyExists {
		t.Fatalf("expected AssignOutcomeAlreadyExists, got %v", second.Outcome)
	}
	if !second.Assignment.CreatedAt.Equal(first.Assignment.CreatedAt) {
		t.Fatalf("expected the replay to report the ORIGINAL createdAt, got first=%v second=%v",
			first.Assignment.CreatedAt, second.Assignment.CreatedAt)
	}

	page, err := repo.List(context.Background(), string(shopG), barber.ID, nil, 50)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	count := 0
	for _, item := range page.Items {
		if item.ServiceID == service.ID {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("CA-023-02: expected exactly 1 row for this pair, got %d", count)
	}
}

func TestAssign_SameServiceToMultipleBarbers_EachIsAnIndependentResource(t *testing.T) {
	// CA-023-03: un mismo servicio puede asignarse a varios barberos de la
	// misma barbería, cada asociación es un recurso independiente.
	db := setupTestDB(t)
	repo := newAssignmentRepository(db)
	suffix := uniqueSuffix(t)

	service := createServiceFixture(t, db, shopG, "Servicio Compartido "+suffix)

	var barbers []fixtureBarber
	for i := 0; i < 4; i++ {
		barbers = append(barbers, createBarberFixture(t, db, shopG, "Equipo Compartido "+suffix+string(rune('A'+i))))
	}

	for _, b := range barbers {
		result, err := repo.Assign(context.Background(), string(shopG), b.ID, service.ID)
		if err != nil {
			t.Fatalf("Assign: %v", err)
		}
		if result.Outcome != catalog.AssignOutcomeCreated {
			t.Fatalf("expected AssignOutcomeCreated for barber %s, got %v", b.ID, result.Outcome)
		}
	}

	// Retirar uno no afecta a los otros tres (cada asociación es
	// independiente): el servicio sigue activo con 3 asignaciones
	// restantes, muy por encima de la última.
	unassignRepo := repo
	unassignResult, err := unassignRepo.Unassign(context.Background(), string(shopG), barbers[0].ID, service.ID)
	if err != nil {
		t.Fatalf("Unassign: %v", err)
	}
	if unassignResult.Outcome != catalog.UnassignOutcomeDeleted {
		t.Fatalf("expected UnassignOutcomeDeleted, got %v", unassignResult.Outcome)
	}
	for _, b := range barbers[1:] {
		page, err := repo.List(context.Background(), string(shopG), b.ID, nil, 50)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		found := false
		for _, item := range page.Items {
			if item.ServiceID == service.ID {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected barber %s to keep its own assignment untouched", b.ID)
		}
	}
}

func TestAssign_UnknownServiceInSameTenant_ServiceNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := newAssignmentRepository(db)
	suffix := uniqueSuffix(t)

	barber := createBarberFixture(t, db, shopG, "Barbero Sin Servicio "+suffix)

	result, err := repo.Assign(context.Background(), string(shopG), barber.ID, "99999999-9999-4999-8999-999999999999")
	if err != nil {
		t.Fatalf("Assign: %v", err)
	}
	if result.Outcome != catalog.AssignOutcomeServiceNotFound {
		t.Fatalf("expected AssignOutcomeServiceNotFound, got %v", result.Outcome)
	}
}

func TestAssign_ServiceFromAnotherTenant_ServiceNotFound(t *testing.T) {
	// CA-023-04: un servicio real, pero de OTRA barbería, es indistinguible
	// de uno inexistente: el filtro WHERE barbershop_id = $2 (defensa en
	// profundidad, además de RLS) lo descubre como "no existe" antes de
	// tocar la FK.
	db := setupTestDB(t)
	repo := newAssignmentRepository(db)
	suffix := uniqueSuffix(t)

	barberG := createBarberFixture(t, db, shopG, "Barbero G "+suffix)
	serviceH := createServiceFixture(t, db, shopH, "Servicio H "+suffix)

	result, err := repo.Assign(context.Background(), string(shopG), barberG.ID, serviceH.ID)
	if err != nil {
		t.Fatalf("Assign: %v", err)
	}
	if result.Outcome != catalog.AssignOutcomeServiceNotFound {
		t.Fatalf("expected AssignOutcomeServiceNotFound for a cross-tenant service, got %v", result.Outcome)
	}
}

// TestRawSQL_CompositeFK_RejectsCrossTenantAssociation confirma, con SQL
// directo (sin pasar por AssignmentRepository, que ya lo evita con su
// propio filtro), que barber_service_barber_fk/barber_service_service_fk
// (FK compuestas por barbershop_id) rechazan una asociación entre un
// barbero y un servicio de barberías distintas AUN bajo el rol de
// aplicación real (CA-023-04): "la FK tenant-aware ... rechaza la relación
// incluso con el rol real de aplicación".
func TestRawSQL_CompositeFK_RejectsCrossTenantAssociation(t *testing.T) {
	db := setupTestDB(t)
	suffix := uniqueSuffix(t)

	barberG := createBarberFixture(t, db, shopG, "FK Barbero G "+suffix)
	serviceH := createServiceFixture(t, db, shopH, "FK Servicio H "+suffix)

	err := db.InTenantTx(context.Background(), shopG, func(ctx context.Context, q database.Queries) error {
		_, execErr := q.Exec(ctx,
			`INSERT INTO barber_service (barbershop_id, barber_id, service_id) VALUES ($1, $2, $3)`,
			string(shopG), barberG.ID, serviceH.ID,
		)
		return execErr
	})
	if err == nil {
		t.Fatal("CA-023-04: expected the composite FK to reject a cross-tenant association, got no error")
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		t.Fatalf("expected a foreign_key_violation (23503), got %v", err)
	}
}

// --- Unassign: CA-023-05, CA-023-06, DEC-068 -----------------------------

func TestUnassign_NotLastAssignment_Deletes(t *testing.T) {
	db := setupTestDB(t)
	repo := newAssignmentRepository(db)
	suffix := uniqueSuffix(t)

	service := createServiceFixture(t, db, shopG, "Servicio No Última "+suffix)
	barberA := createBarberFixture(t, db, shopG, "No Última A "+suffix)
	barberB := createBarberFixture(t, db, shopG, "No Última B "+suffix)

	mustAssign(t, repo, shopG, barberA.ID, service.ID)
	mustAssign(t, repo, shopG, barberB.ID, service.ID)

	result, err := repo.Unassign(context.Background(), string(shopG), barberA.ID, service.ID)
	if err != nil {
		t.Fatalf("Unassign: %v", err)
	}
	if result.Outcome != catalog.UnassignOutcomeDeleted {
		t.Fatalf("expected UnassignOutcomeDeleted, got %v", result.Outcome)
	}

	page, err := repo.List(context.Background(), string(shopG), barberA.ID, nil, 50)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, item := range page.Items {
		if item.ServiceID == service.ID {
			t.Fatal("expected the assignment to be gone after Unassign")
		}
	}
}

func TestUnassign_LastActiveAssignment_RejectedWithoutDeleting(t *testing.T) {
	// DEC-068/CA-023-05/CA-023-06: un solo barbero asignado a un servicio
	// activo; retirarlo dejaría el servicio en cero, así que se rechaza.
	db := setupTestDB(t)
	repo := newAssignmentRepository(db)
	suffix := uniqueSuffix(t)

	service := createServiceFixture(t, db, shopG, "Servicio Última "+suffix)
	barber := createBarberFixture(t, db, shopG, "Última Único "+suffix)
	mustAssign(t, repo, shopG, barber.ID, service.ID)

	result, err := repo.Unassign(context.Background(), string(shopG), barber.ID, service.ID)
	if err != nil {
		t.Fatalf("Unassign: %v", err)
	}
	if result.Outcome != catalog.UnassignOutcomeLastActiveConflict {
		t.Fatalf("expected UnassignOutcomeLastActiveConflict, got %v", result.Outcome)
	}

	// Nada se borró: la fila sigue ahí.
	page, err := repo.List(context.Background(), string(shopG), barber.ID, nil, 50)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, item := range page.Items {
		if item.ServiceID == service.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("CA-023-06: the rejected unassign must NOT have deleted the row")
	}

	// La operación es segura ante repetición: reintentarla produce el
	// MISMO rechazo, nunca un borrado accidental en el segundo intento.
	again, err := repo.Unassign(context.Background(), string(shopG), barber.ID, service.ID)
	if err != nil {
		t.Fatalf("Unassign (retry): %v", err)
	}
	if again.Outcome != catalog.UnassignOutcomeLastActiveConflict {
		t.Fatalf("CA-023-05: expected the retry to be rejected the same way, got %v", again.Outcome)
	}
}

func TestUnassign_NoSuchAssignment_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := newAssignmentRepository(db)
	suffix := uniqueSuffix(t)

	service := createServiceFixture(t, db, shopG, "Servicio Sin Asignar "+suffix)
	barber := createBarberFixture(t, db, shopG, "Sin Asignar "+suffix)

	result, err := repo.Unassign(context.Background(), string(shopG), barber.ID, service.ID)
	if err != nil {
		t.Fatalf("Unassign: %v", err)
	}
	if result.Outcome != catalog.UnassignOutcomeNotFound {
		t.Fatalf("expected UnassignOutcomeNotFound, got %v", result.Outcome)
	}
}

func TestUnassign_UnknownService_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := newAssignmentRepository(db)
	suffix := uniqueSuffix(t)

	barber := createBarberFixture(t, db, shopG, "Servicio Fantasma "+suffix)

	result, err := repo.Unassign(context.Background(), string(shopG), barber.ID, "99999999-9999-4999-8999-999999999999")
	if err != nil {
		t.Fatalf("Unassign: %v", err)
	}
	if result.Outcome != catalog.UnassignOutcomeNotFound {
		t.Fatalf("expected UnassignOutcomeNotFound for an unknown service, got %v", result.Outcome)
	}
}

// TestUnassign_ConcurrentRaceOnLastTwoAssignments_ExactlyOneSucceeds es LA
// prueba de la carrera que DEC-068 exige explícitamente: dos
// desasignaciones concurrentes REALES (dos conexiones físicas) de las DOS
// ÚLTIMAS filas activas de un mismo servicio. Exactamente una debe tener
// éxito (UnassignOutcomeDeleted) y la otra debe ser rechazada
// (UnassignOutcomeLastActiveConflict): el servicio JAMÁS puede quedar en
// cero asignaciones como resultado de esta carrera. El bloqueo de fila
// (`SELECT ... FOR UPDATE` sobre `service`) dentro de
// AssignmentRepository.Unassign es lo que hace esto posible: ambas
// transacciones se serializan sobre la MISMA fila de `service`.
func TestUnassign_ConcurrentRaceOnLastTwoAssignments_ExactlyOneSucceeds(t *testing.T) {
	db := setupTestDB(t)
	repo := newAssignmentRepository(db)
	suffix := uniqueSuffix(t)

	service := createServiceFixture(t, db, shopG, "Servicio Carrera "+suffix)
	barberA := createBarberFixture(t, db, shopG, "Carrera A "+suffix)
	barberB := createBarberFixture(t, db, shopG, "Carrera B "+suffix)

	mustAssign(t, repo, shopG, barberA.ID, service.ID)
	mustAssign(t, repo, shopG, barberB.ID, service.ID)

	var (
		wg               sync.WaitGroup
		resultA, resultB catalog.UnassignResult
		errA, errB       error
		startA           = make(chan struct{})
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		close(startA)
		resultA, errA = repo.Unassign(context.Background(), string(shopG), barberA.ID, service.ID)
	}()
	go func() {
		defer wg.Done()
		<-startA
		resultB, errB = repo.Unassign(context.Background(), string(shopG), barberB.ID, service.ID)
	}()
	wg.Wait()

	if errA != nil {
		t.Fatalf("goroutine A: %v", errA)
	}
	if errB != nil {
		t.Fatalf("goroutine B: %v", errB)
	}

	outcomes := []catalog.UnassignOutcome{resultA.Outcome, resultB.Outcome}
	deletedCount, conflictCount := 0, 0
	for _, o := range outcomes {
		switch o {
		case catalog.UnassignOutcomeDeleted:
			deletedCount++
		case catalog.UnassignOutcomeLastActiveConflict:
			conflictCount++
		default:
			t.Fatalf("unexpected outcome in the race: %v", o)
		}
	}
	if deletedCount != 1 || conflictCount != 1 {
		t.Fatalf("DEC-068: expected exactly one Deleted and one LastActiveConflict, got A=%v B=%v",
			resultA.Outcome, resultB.Outcome)
	}

	// Verificación final, la que de verdad importa: el servicio conserva
	// AL MENOS una asignación activa, nunca cero (CA-023-06).
	remaining := 0
	for _, barberID := range []string{barberA.ID, barberB.ID} {
		page, err := repo.List(context.Background(), string(shopG), barberID, nil, 50)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		for _, item := range page.Items {
			if item.ServiceID == service.ID {
				remaining++
			}
		}
	}
	if remaining != 1 {
		t.Fatalf("DEC-068: expected exactly 1 remaining assignment for the service after the race, got %d", remaining)
	}
}

func mustAssign(t *testing.T, repo *catalogpostgres.AssignmentRepository, shop database.BarbershopID, barberID, serviceID string) {
	t.Helper()
	result, err := repo.Assign(context.Background(), string(shop), barberID, serviceID)
	if err != nil {
		t.Fatalf("mustAssign: %v", err)
	}
	if result.Outcome != catalog.AssignOutcomeCreated {
		t.Fatalf("mustAssign: expected AssignOutcomeCreated, got %v", result.Outcome)
	}
}
