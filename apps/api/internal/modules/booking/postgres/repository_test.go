// Package postgres_test (pruebas de integración, HU-060) requiere
// PostgreSQL REAL con las dieciséis migraciones aplicadas (incluida
// 20260827110000_create_appointment_core.sql) y database/testdata/dos_barberias.sql +
// database/testdata/hu060_citas.sql cargados. Conéctate como barberia_app
// (docs/03-desarrollo/estrategia-pruebas.md §2 prohíbe mocks para
// RLS/exclusión/atomicidad/concurrencia).
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/booking/postgres/...
package postgres_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"system-barbershop/internal/modules/booking"
	bookingpostgres "system-barbershop/internal/modules/booking/postgres"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

const (
	testDatabaseURL = "postgres://barberia_app@localhost:5432/barberia_test?sslmode=disable"

	// shopQ/shopR son las dos barberías dedicadas de
	// testdata/hu060_citas.sql: aisladas de dos_barberias.sql y de las
	// demás suites para que un conteo o una carrera de exclusión no
	// dependa de lo que otras suites hagan.
	shopQ = database.BarbershopID("c17a0001-c17a-c17a-c17a-c17a00010001")
	shopR = database.BarbershopID("c17a0002-c17a-c17a-c17a-c17a00020002")

	barberQ1 = "c17a1001-1001-1001-1001-100110011001"
	barberQ2 = "c17a1002-1002-1002-1002-100210021002"
	barberR1 = "c17a2001-2001-2001-2001-200120012001"

	serviceQ = "c17a5001-5001-5001-5001-500150015001"
	serviceR = "c17a5002-5002-5002-5002-500250025002"

	staffQ = "c17a9001-9001-9001-9001-900190019001"
	staffR = "c17a9002-9002-9002-9002-900290029002"
)

func setupTestDB(t *testing.T) *database.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = testDatabaseURL
	}
	cfg := config.Config{
		Environment:              "test",
		DatabaseMaxConns:         10,
		DatabaseMinConns:         2,
		DatabaseMaxConnLifetime:  time.Hour,
		DatabaseMaxConnIdleTime:  30 * time.Minute,
		DatabaseConnectTimeout:   5 * time.Second,
		DatabaseStatementTimeout: 10 * time.Second,
	}
	db, err := database.NewDB(config.DatabaseDSN(url), cfg)
	if err != nil {
		t.Fatalf("database.NewDB: %v", err)
	}
	t.Cleanup(db.Close)
	return db
}

func newRepository(db *database.DB) *bookingpostgres.Repository {
	return bookingpostgres.New(db, idempotency.NewSQLCoordinator())
}

// uniqueSuffix evita colisiones entre ejecuciones repetidas de la suite
// contra el mismo PostgreSQL persistente (barberia_app no puede DELETE
// customer/appointment, RN-DAT-03/RN-HIS-02, así que las filas de una
// corrida anterior siguen ahí): cada prueba usa nombres y una fecha propia.
func uniqueSuffix(t *testing.T) string {
	t.Helper()
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	return hex.EncodeToString(buf)
}

// testStartOffset deriva, de bytes aleatorios frescos, un desplazamiento en
// minutos sobre una fecha base lejana en el futuro dentro de una ventana de
// ~500 años (~2.6*10^8 minutos): cada ejecución de la suite reserva su
// propia franja con probabilidad de colisión despreciable incluso después
// de miles de corridas contra el mismo PostgreSQL persistente (customer y
// appointment son append-only para barberia_app, RN-DAT-03/RN-HIS-02: nada
// limpia las filas de una corrida anterior). Un día de amplitud, como se
// usó en una primera versión de esta prueba, resultó insuficiente y
// produjo un falso conflicto real tras acumular corridas locales.
func testStartOffset(t *testing.T) time.Duration {
	t.Helper()
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	var n uint64
	for _, b := range buf {
		n = n<<8 | uint64(b)
	}
	const windowMinutes = 262_800_000 // ~500 años
	return time.Duration(n%windowMinutes) * time.Minute
}

func baseStart(t *testing.T) time.Time {
	return time.Date(2030, 1, 1, 9, 0, 0, 0, time.UTC).Add(testStartOffset(t))
}

func strPtr(s string) *string { return &s }

func validInput(t *testing.T, barberID, serviceID string, start time.Time, durationMinutes int, suffix string) booking.CreateInternalInput {
	return booking.CreateInternalInput{
		BarberID:     barberID,
		ServiceID:    serviceID,
		AttendeeName: "Cliente de prueba " + suffix,
		StartsAt:     start,
		EndsAt:       start.Add(time.Duration(durationMinutes) * time.Minute),
		Origin:       booking.OriginManual,
		Service: booking.ServiceSnapshot{
			Name:             "Corte de prueba " + suffix,
			DurationMinutes:  durationMinutes,
			PriceAmountCents: 2000000,
			Currency:         "COP",
		},
		Customer: booking.CustomerInput{
			New: &booking.NewCustomerInput{FullName: "Cliente de prueba " + suffix},
		},
		Actor: booking.Actor{Type: booking.ActorTypeStaff, StaffUserID: strPtr(staffQ)},
	}
}

func TestCreateInternal_NewCustomer_PersistsAppointmentAndHistory(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	input := validInput(t, barberQ1, serviceQ, baseStart(t), 30, suffix)
	result, err := repo.CreateInternal(context.Background(), string(shopQ), input)
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}

	if result.Appointment.Status != booking.StatusConfirmed {
		t.Fatalf("expected status confirmed, got %v", result.Appointment.Status)
	}
	if !result.Appointment.OccupiesSchedule {
		t.Fatalf("expected a confirmed appointment to occupy the schedule")
	}
	if result.Appointment.Origin != booking.OriginManual {
		t.Fatalf("expected origin manual, got %v", result.Appointment.Origin)
	}
	if result.Appointment.DurationMinutesSnapshot != 30 {
		t.Fatalf("expected duration snapshot 30, got %d", result.Appointment.DurationMinutesSnapshot)
	}
	if result.Appointment.PriceAmountCentsSnapshot != 2000000 {
		t.Fatalf("expected price snapshot 2000000, got %d", result.Appointment.PriceAmountCentsSnapshot)
	}
	if result.Customer.ID == "" {
		t.Fatalf("expected a persisted customer id")
	}
	if result.Appointment.CustomerID != result.Customer.ID {
		t.Fatalf("appointment.CustomerID (%s) does not match the created customer (%s)",
			result.Appointment.CustomerID, result.Customer.ID)
	}

	// El historial appointment_created quedó persistido en la misma
	// transacción (CA-060-05): lo confirma una lectura directa con el rol
	// real, filtrando por tenant.
	err = db.InTenantTx(context.Background(), shopQ, func(ctx context.Context, q database.Queries) error {
		var count int
		scanErr := q.QueryRow(ctx,
			`SELECT count(*) FROM appointment_history
			  WHERE barbershop_id = $1 AND appointment_id = $2
			    AND event_type = 'appointment_created' AND actor_type = 'staff' AND actor_staff_user_id = $3`,
			string(shopQ), result.Appointment.ID, staffQ,
		).Scan(&count)
		if scanErr != nil {
			return scanErr
		}
		if count != 1 {
			t.Fatalf("expected exactly one appointment_created history row, got %d", count)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("verify history: %v", err)
	}
}

func TestCreateInternal_ExistingCustomer_Links(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	// Crea un cliente primero, con una cita, y reutiliza ese cliente en una
	// segunda cita (misma persona, otro horario).
	first := validInput(t, barberQ1, serviceQ, baseStart(t), 30, suffix)
	firstResult, err := repo.CreateInternal(context.Background(), string(shopQ), first)
	if err != nil {
		t.Fatalf("CreateInternal (first): %v", err)
	}

	second := validInput(t, barberQ1, serviceQ, baseStart(t).Add(2*time.Hour), 30, suffix)
	second.Customer = booking.CustomerInput{ExistingID: &firstResult.Customer.ID}
	secondResult, err := repo.CreateInternal(context.Background(), string(shopQ), second)
	if err != nil {
		t.Fatalf("CreateInternal (second, existing customer): %v", err)
	}

	if secondResult.Customer.ID != firstResult.Customer.ID {
		t.Fatalf("expected the second appointment to link the same customer, got %s vs %s",
			secondResult.Customer.ID, firstResult.Customer.ID)
	}
}

func TestCreateInternal_ExistingCustomerFromOtherTenant_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	// Cliente real, pero de shopR: RN-TEN-01 exige que shopQ no pueda
	// vincularlo, ni siquiera conociendo su id real.
	rInput := validInput(t, barberR1, serviceR, baseStart(t), 45, suffix)
	rInput.Actor.StaffUserID = strPtr(staffR)
	rResult, err := repo.CreateInternal(context.Background(), string(shopR), rInput)
	if err != nil {
		t.Fatalf("CreateInternal (shopR fixture): %v", err)
	}

	qInput := validInput(t, barberQ1, serviceQ, baseStart(t), 30, suffix)
	qInput.Customer = booking.CustomerInput{ExistingID: &rResult.Customer.ID}
	_, err = repo.CreateInternal(context.Background(), string(shopQ), qInput)
	if err == nil {
		t.Fatalf("expected an error linking a customer from another tenant")
	}
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindNotFound {
		t.Fatalf("expected apperr.KindNotFound, got %v (%T)", err, err)
	}
}

func TestCreateInternal_ScheduleConflict_RollsBackNewCustomer(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	base := validInput(t, barberQ1, serviceQ, start, 30, suffix+"-base")
	if _, err := repo.CreateInternal(context.Background(), string(shopQ), base); err != nil {
		t.Fatalf("CreateInternal (base): %v", err)
	}

	// Se cruza totalmente con la cita base, mismo barbero. El cliente NUEVO
	// de este intento nunca debe quedar persistido: la transacción entera
	// revierte (CA-060-05, "error tras cliente").
	overlap := validInput(t, barberQ1, serviceQ, start, 30, suffix+"-overlap")
	_, err := repo.CreateInternal(context.Background(), string(shopQ), overlap)
	if err == nil {
		t.Fatalf("expected a schedule conflict for an overlapping interval")
	}
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindConflict {
		t.Fatalf("expected apperr.KindConflict, got %v (%T)", err, err)
	}

	err = db.InTenantTx(context.Background(), shopQ, func(ctx context.Context, q database.Queries) error {
		var count int
		scanErr := q.QueryRow(ctx,
			`SELECT count(*) FROM customer WHERE barbershop_id = $1 AND full_name = $2`,
			string(shopQ), "Cliente de prueba "+suffix+"-overlap",
		).Scan(&count)
		if scanErr != nil {
			return scanErr
		}
		if count != 0 {
			t.Fatalf("CA-060-05: the customer of the failed attempt should not be persisted, found %d rows", count)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("verify rollback: %v", err)
	}
}

// TestCreateInternal_HistoryInsertFails_RollsBackAppointmentAndCustomer
// cubre el cuarto caso obligatorio del prompt de HU-060 ("error tras
// cita"): un actor de otro tenant hace que el INSERT de
// appointment_history choque contra su FK compuesta tenant-aware DESPUÉS
// de que el cliente y la cita ya se insertaron dentro de la misma
// transacción. Ninguno de los tres debe quedar persistido.
func TestCreateInternal_HistoryInsertFails_RollsBackAppointmentAndCustomer(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	input := validInput(t, barberQ1, serviceQ, baseStart(t), 30, suffix)
	// staffR pertenece a shopR, no a shopQ: la FK compuesta
	// appointment_history_barbershop_id_actor_staff_user_id_fk debe
	// rechazarla.
	input.Actor.StaffUserID = strPtr(staffR)

	_, err := repo.CreateInternal(context.Background(), string(shopQ), input)
	if err == nil {
		t.Fatalf("expected an error when the actor belongs to another tenant")
	}

	verifyErr := db.InTenantTx(context.Background(), shopQ, func(ctx context.Context, q database.Queries) error {
		var customerCount, appointmentCount int
		if scanErr := q.QueryRow(ctx,
			`SELECT count(*) FROM customer WHERE barbershop_id = $1 AND full_name = $2`,
			string(shopQ), "Cliente de prueba "+suffix,
		).Scan(&customerCount); scanErr != nil {
			return scanErr
		}
		if scanErr := q.QueryRow(ctx,
			`SELECT count(*) FROM appointment WHERE barbershop_id = $1 AND attendee_name = $2`,
			string(shopQ), "Cliente de prueba "+suffix,
		).Scan(&appointmentCount); scanErr != nil {
			return scanErr
		}
		if customerCount != 0 {
			t.Fatalf("expected no customer written when the history insert fails, found %d rows", customerCount)
		}
		if appointmentCount != 0 {
			t.Fatalf("expected no appointment written when the history insert fails, found %d rows", appointmentCount)
		}
		return nil
	})
	if verifyErr != nil {
		t.Fatalf("verify rollback: %v", verifyErr)
	}
}

func TestCreateInternal_ContiguousInterval_Succeeds(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	base := validInput(t, barberQ1, serviceQ, start, 30, suffix+"-a")
	if _, err := repo.CreateInternal(context.Background(), string(shopQ), base); err != nil {
		t.Fatalf("CreateInternal (base): %v", err)
	}

	// Contigua: empieza exactamente cuando la base termina. RN-DIS-05.
	contiguous := validInput(t, barberQ1, serviceQ, start.Add(30*time.Minute), 30, suffix+"-b")
	if _, err := repo.CreateInternal(context.Background(), string(shopQ), contiguous); err != nil {
		t.Fatalf("expected a contiguous appointment to succeed, got: %v", err)
	}
}

func TestCreateInternal_ContextCancelled_NoPartialWrite(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.CreateInternal(ctx, string(shopQ), validInput(t, barberQ1, serviceQ, baseStart(t), 30, suffix))
	if err == nil {
		t.Fatalf("expected an error for a cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Logf("underlying error (informational, may be wrapped by pgx): %v", err)
	}

	err = db.InTenantTx(context.Background(), shopQ, func(ctx context.Context, q database.Queries) error {
		var count int
		scanErr := q.QueryRow(ctx,
			`SELECT count(*) FROM customer WHERE barbershop_id = $1 AND full_name = $2`,
			string(shopQ), "Cliente de prueba "+suffix,
		).Scan(&count)
		if scanErr != nil {
			return scanErr
		}
		if count != 0 {
			t.Fatalf("expected no customer written after a cancelled context, found %d rows", count)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("verify no write: %v", err)
	}
}

// TestCreateInternal_ConcurrentOverlap_ExactlyOneSucceeds es la carrera de
// dos conexiones reales exigida por el prompt de HU-060 (sin sleeps, sin
// dobles): dos goroutines, cada una con su propia conexión adquirida del
// pool real, intentan crear al mismo tiempo dos citas que se cruzan para el
// mismo barbero. PostgreSQL debe aceptar exactamente una
// (appointment_barber_interval_excl); la otra debe recibir el conflicto
// tipado. Ejecutar con -race, mismo patrón que
// TestUnassign_ConcurrentRaceOnLastTwoAssignments_ExactlyOneSucceeds
// (catalog/postgres/assignment_repository_test.go, HU-023).
func TestCreateInternal_ConcurrentOverlap_ExactlyOneSucceeds(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	inputA := validInput(t, barberQ1, serviceQ, start, 30, suffix+"-A")
	inputB := validInput(t, barberQ1, serviceQ, start.Add(10*time.Minute), 30, suffix+"-B")

	var (
		wg               sync.WaitGroup
		resultA, resultB booking.CreateInternalResult
		errA, errB       error
		startBarrier     = make(chan struct{})
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		close(startBarrier)
		resultA, errA = repo.CreateInternal(context.Background(), string(shopQ), inputA)
	}()
	go func() {
		defer wg.Done()
		<-startBarrier
		resultB, errB = repo.CreateInternal(context.Background(), string(shopQ), inputB)
	}()
	wg.Wait()

	succeeded, conflicted := 0, 0
	for _, res := range []struct {
		result booking.CreateInternalResult
		err    error
	}{{resultA, errA}, {resultB, errB}} {
		switch {
		case res.err == nil:
			succeeded++
			if res.result.Appointment.ID == "" {
				t.Fatalf("a successful result must carry a persisted appointment id")
			}
		default:
			appErr, ok := apperr.As(res.err)
			if !ok || appErr.Kind != apperr.KindConflict {
				t.Fatalf("expected apperr.KindConflict for the losing goroutine, got %v (%T)", res.err, res.err)
			}
			conflicted++
		}
	}

	if succeeded != 1 || conflicted != 1 {
		t.Fatalf("RN-CON-01/RN-CON-03: expected exactly one success and one conflict in the race, got succeeded=%d conflicted=%d (errA=%v errB=%v)",
			succeeded, conflicted, errA, errB)
	}

	// Verificación final: exactamente una cita cruzada persiste para ese
	// barbero en esa ventana, consultado con el rol real.
	err := db.InTenantTx(context.Background(), shopQ, func(ctx context.Context, q database.Queries) error {
		var count int
		scanErr := q.QueryRow(ctx,
			`SELECT count(*) FROM appointment
			  WHERE barbershop_id = $1 AND barber_id = $2
			    AND tstzrange(starts_at, ends_at, '[)') && tstzrange($3, $4, '[)')
			    AND occupies_schedule`,
			string(shopQ), barberQ1, start.Add(-time.Hour), start.Add(time.Hour),
		).Scan(&count)
		if scanErr != nil {
			return scanErr
		}
		if count != 1 {
			t.Fatalf("expected exactly one persisted crossing appointment, found %d", count)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("verify persisted count: %v", err)
	}
}

func TestCreateInternal_DifferentBarber_SameWindow_Succeeds(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	first := validInput(t, barberQ1, serviceQ, start, 30, suffix+"-1")
	if _, err := repo.CreateInternal(context.Background(), string(shopQ), first); err != nil {
		t.Fatalf("CreateInternal (barberQ1): %v", err)
	}

	second := validInput(t, barberQ2, serviceQ, start, 30, suffix+"-2")
	if _, err := repo.CreateInternal(context.Background(), string(shopQ), second); err != nil {
		t.Fatalf("expected a different barber in the same window to succeed, got: %v", err)
	}
}
