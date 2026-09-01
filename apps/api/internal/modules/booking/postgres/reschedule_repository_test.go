// Package postgres_test (pruebas de integración, HU-065) requiere
// PostgreSQL REAL con las diecisiete migraciones aplicadas y
// database/testdata/dos_barberias.sql + database/testdata/hu060_citas.sql
// cargados: mismos fixtures que repository_test.go/detail_repository_test.go
// (HU-060/HU-064). El turno base de cada prueba se inserta con
// repo.CreateInternal, la misma primitiva probada en repository_test.go.
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/booking/postgres/...
package postgres_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"testing"
	"time"

	"system-barbershop/internal/modules/booking"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

// rescheduleFingerprint deriva un Fingerprint hexadecimal válido (32-128
// caracteres en minúsculas, idempotency_record_request_fingerprint_ck) a
// partir de una semilla legible: ComputeFingerprint produce la forma real
// en producción, pero estas pruebas ejercitan booking.Repository.Reschedule
// directamente, sin pasar por la capa HTTP que lo calcularía.
func rescheduleFingerprint(seed string) idempotency.Fingerprint {
	sum := sha256.Sum256([]byte(seed))
	return idempotency.Fingerprint(fmt.Sprintf("%x", sum))
}

func newRescheduleInput(appointmentID, versionToken string, newStartsAt, newEndsAt time.Time) booking.RescheduleInput {
	return booking.RescheduleInput{
		AppointmentID:        appointmentID,
		NewStartsAt:          newStartsAt,
		NewEndsAt:            newEndsAt,
		ExpectedVersionToken: versionToken,
		Actor:                booking.Actor{Type: booking.ActorTypeStaff, StaffUserID: strPtr(staffQ)},
	}
}

// markAppointmentTerminal marca directamente (barberia_app tiene UPDATE
// sobre appointment) una cita como resuelta, sin pasar por ninguna
// transición todavía no construida (T3/cancelación/cierre no existen en
// esta HU): solo sirve para fabricar el estado previo que
// TestReschedule_NotConfirmed_ReturnsInvalidState necesita.
func markAppointmentTerminal(t *testing.T, db *database.DB, shop database.BarbershopID, appointmentID string) {
	t.Helper()
	err := db.InTenantTx(context.Background(), shop, func(ctx context.Context, q database.Queries) error {
		_, err := q.Exec(ctx,
			`UPDATE appointment SET status = 'completed', resolved_at = now() WHERE barbershop_id = $1 AND id = $2`,
			string(shop), appointmentID)
		return err
	})
	if err != nil {
		t.Fatalf("markAppointmentTerminal: %v", err)
	}
}

func TestReschedule_ValidRequest_UpdatesIntervalAndInsertsHistory(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	detail, found, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil || !found {
		t.Fatalf("GetAppointmentDetail: found=%v err=%v", found, err)
	}

	newStarts := start.Add(3 * time.Hour)
	newEnds := newStarts.Add(30 * time.Minute)
	input := newRescheduleInput(created.Appointment.ID, detail.VersionToken, newStarts, newEnds)

	result, err := repo.Reschedule(context.Background(), string(shopQ), input,
		idempotency.Key("resched-key-"+suffix), rescheduleFingerprint("valid-"+suffix))
	if err != nil {
		t.Fatalf("Reschedule: %v", err)
	}
	if result.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("Outcome = %v, want Proceed", result.Decision.Outcome)
	}
	if !result.Appointment.StartsAt.Equal(newStarts) || !result.Appointment.EndsAt.Equal(newEnds) {
		t.Fatalf("intervalo no aplicado: got [%v,%v), want [%v,%v)",
			result.Appointment.StartsAt, result.Appointment.EndsAt, newStarts, newEnds)
	}
	if result.Appointment.VersionToken == detail.VersionToken {
		t.Fatalf("VersionToken no cambió tras la escritura")
	}
	// Barbero, servicio, cliente, persona, nota, origen y estado NUNCA
	// cambian en T2 (trabajo requerido §1.2).
	if result.Appointment.BarberID != detail.BarberID || result.Appointment.Status != detail.Status {
		t.Fatalf("T2 modificó datos fuera del intervalo: %+v", result.Appointment)
	}

	items, _, found, err := repo.ListAppointmentHistory(context.Background(), string(shopQ), created.Appointment.ID, nil, 20)
	if err != nil || !found {
		t.Fatalf("ListAppointmentHistory: found=%v err=%v", found, err)
	}
	var rescheduled *booking.HistoryRow
	for i := range items {
		if items[i].EventType == booking.EventTypeAppointmentRescheduled {
			rescheduled = &items[i]
		}
	}
	if rescheduled == nil {
		t.Fatalf("no se insertó el evento appointment_rescheduled: %+v", items)
	}
	if len(rescheduled.Changes) != 2 {
		t.Fatalf("Changes = %+v, want starts_at y ends_at", rescheduled.Changes)
	}
}

func TestReschedule_SameInterval_IsNoOpWithoutHistoryOrVersionChange(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	detail, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}

	input := newRescheduleInput(created.Appointment.ID, detail.VersionToken, created.Appointment.StartsAt, created.Appointment.EndsAt)
	result, err := repo.Reschedule(context.Background(), string(shopQ), input,
		idempotency.Key("resched-noop-"+suffix), rescheduleFingerprint("noop-"+suffix))
	if err != nil {
		t.Fatalf("Reschedule: %v", err)
	}
	if result.Appointment.VersionToken != detail.VersionToken {
		t.Fatalf("VersionToken cambió en un no-op: %q != %q", result.Appointment.VersionToken, detail.VersionToken)
	}

	items, _, _, err := repo.ListAppointmentHistory(context.Background(), string(shopQ), created.Appointment.ID, nil, 20)
	if err != nil {
		t.Fatalf("ListAppointmentHistory: %v", err)
	}
	for _, row := range items {
		if row.EventType == booking.EventTypeAppointmentRescheduled {
			t.Fatalf("un no-op no debió insertar appointment_rescheduled: %+v", items)
		}
	}
}

func TestReschedule_StaleVersionToken_ReturnsVersionConflict(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}

	newStarts := start.Add(3 * time.Hour)
	input := newRescheduleInput(created.Appointment.ID, "un-token-que-nunca-produjo-encodeversiontoken", newStarts, newStarts.Add(30*time.Minute))

	_, err = repo.Reschedule(context.Background(), string(shopQ), input,
		idempotency.Key("resched-stale-"+suffix), rescheduleFingerprint("stale-"+suffix))
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindVersionConflict {
		t.Fatalf("err = %v, want apperr.KindVersionConflict", err)
	}
}

func TestReschedule_NotConfirmed_ReturnsInvalidState(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	detail, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}
	markAppointmentTerminal(t, db, shopQ, created.Appointment.ID)

	newStarts := start.Add(3 * time.Hour)
	// El token de versión ya cambió al marcar la cita terminal (updated_at
	// se movió), así que se recalcula contra el detalle real para aislar
	// el chequeo de ESTADO del chequeo de versión.
	afterDetail, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail(2): %v", err)
	}
	if afterDetail.VersionToken == detail.VersionToken {
		t.Fatalf("VersionToken no cambió tras marcar la cita terminal")
	}

	input := newRescheduleInput(created.Appointment.ID, afterDetail.VersionToken, newStarts, newStarts.Add(30*time.Minute))
	_, err = repo.Reschedule(context.Background(), string(shopQ), input,
		idempotency.Key("resched-terminal-"+suffix), rescheduleFingerprint("terminal-"+suffix))
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInvalidState {
		t.Fatalf("err = %v, want apperr.KindInvalidState", err)
	}
}

func TestReschedule_CrossesAnotherAppointment_ReturnsConflict(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	moving, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix+"-moving"))
	if err != nil {
		t.Fatalf("CreateInternal(moving): %v", err)
	}
	occupying := start.Add(3 * time.Hour)
	if _, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, occupying, 30, suffix+"-occupying")); err != nil {
		t.Fatalf("CreateInternal(occupying): %v", err)
	}

	detail, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), moving.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}

	// Mueve "moving" a un intervalo que cruza exactamente con "occupying".
	input := newRescheduleInput(moving.Appointment.ID, detail.VersionToken, occupying.Add(10*time.Minute), occupying.Add(40*time.Minute))
	_, err = repo.Reschedule(context.Background(), string(shopQ), input,
		idempotency.Key("resched-cross-"+suffix), rescheduleFingerprint("cross-"+suffix))
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindConflict {
		t.Fatalf("err = %v, want apperr.KindConflict", err)
	}

	// Verificación final: el intervalo original de "moving" sigue intacto
	// (la transacción completa se revirtió). Se compara contra `detail`
	// (también leído de PostgreSQL, precisión de microsegundos), nunca
	// contra `start` (time.Time de Go en memoria, precisión de
	// nanosegundos): comparar contra el valor crudo produciría un falso
	// desajuste por el simple truncamiento de timestamptz, no por ningún
	// defecto real de Reschedule.
	after, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), moving.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail(after): %v", err)
	}
	if !after.StartsAt.Equal(detail.StartsAt) {
		t.Fatalf("el intervalo original se modificó a pesar del conflicto: %v != %v", after.StartsAt, detail.StartsAt)
	}
}

// TestReschedule_TenantMismatch_ReturnsNotFound verifica RN-TEN-01: una
// cita real de shopQ nunca se reprograma consultándola con el
// barbershopID de shopR, aunque el identificador coincida literalmente.
func TestReschedule_TenantMismatch_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	detail, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}

	input := newRescheduleInput(created.Appointment.ID, detail.VersionToken, start.Add(3*time.Hour), start.Add(3*time.Hour+30*time.Minute))
	_, err = repo.Reschedule(context.Background(), string(shopR), input,
		idempotency.Key("resched-tenant-"+suffix), rescheduleFingerprint("tenant-"+suffix))
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindNotFound {
		t.Fatalf("err = %v, want apperr.KindNotFound (RN-TEN-01)", err)
	}
}

func TestReschedule_RepeatedKeySameIntent_Replays(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	detail, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}

	newStarts := start.Add(3 * time.Hour)
	input := newRescheduleInput(created.Appointment.ID, detail.VersionToken, newStarts, newStarts.Add(30*time.Minute))
	key := idempotency.Key("resched-replay-" + suffix)
	fingerprint := rescheduleFingerprint("replay-" + suffix)

	first, err := repo.Reschedule(context.Background(), string(shopQ), input, key, fingerprint)
	if err != nil {
		t.Fatalf("Reschedule(1): %v", err)
	}
	second, err := repo.Reschedule(context.Background(), string(shopQ), input, key, fingerprint)
	if err != nil {
		t.Fatalf("Reschedule(2): %v", err)
	}
	if second.Response.Body != first.Response.Body {
		t.Fatalf("la repetición no reprodujo el mismo cuerpo byte a byte:\n1: %s\n2: %s", first.Response.Body, second.Response.Body)
	}

	items, _, _, err := repo.ListAppointmentHistory(context.Background(), string(shopQ), created.Appointment.ID, nil, 20)
	if err != nil {
		t.Fatalf("ListAppointmentHistory: %v", err)
	}
	count := 0
	for _, row := range items {
		if row.EventType == booking.EventTypeAppointmentRescheduled {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("appointment_rescheduled insertado %d veces, want exactamente 1", count)
	}
}

// TestReschedule_ConcurrentReschedulesToOverlappingIntervals_ExactlyOneSucceeds
// cubre RN-CON-01/RN-CON-03 para T2: dos citas confirmadas distintas del
// mismo barbero se reprograman EN PARALELO hacia intervalos que se
// cruzarían entre sí. La restricción de exclusión GiST, evaluada también en
// UPDATE, garantiza que como máximo una de las dos escrituras persista.
func TestReschedule_ConcurrentReschedulesToOverlappingIntervals_ExactlyOneSucceeds(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	createdA, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix+"-A"))
	if err != nil {
		t.Fatalf("CreateInternal(A): %v", err)
	}
	createdB, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start.Add(1*time.Hour), 30, suffix+"-B"))
	if err != nil {
		t.Fatalf("CreateInternal(B): %v", err)
	}
	detailA, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), createdA.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail(A): %v", err)
	}
	detailB, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), createdB.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail(B): %v", err)
	}

	target := start.Add(6 * time.Hour)
	inputA := newRescheduleInput(createdA.Appointment.ID, detailA.VersionToken, target, target.Add(30*time.Minute))
	inputB := newRescheduleInput(createdB.Appointment.ID, detailB.VersionToken, target.Add(10*time.Minute), target.Add(40*time.Minute))

	var (
		wg               sync.WaitGroup
		resultA, resultB booking.RescheduleResult
		errA, errB       error
		startBarrier     = make(chan struct{})
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		close(startBarrier)
		resultA, errA = repo.Reschedule(context.Background(), string(shopQ), inputA,
			idempotency.Key("resched-race-a-"+suffix), rescheduleFingerprint("race-a-"+suffix))
	}()
	go func() {
		defer wg.Done()
		<-startBarrier
		resultB, errB = repo.Reschedule(context.Background(), string(shopQ), inputB,
			idempotency.Key("resched-race-b-"+suffix), rescheduleFingerprint("race-b-"+suffix))
	}()
	wg.Wait()

	succeeded, conflicted := 0, 0
	for _, res := range []struct {
		result booking.RescheduleResult
		err    error
	}{{resultA, errA}, {resultB, errB}} {
		switch {
		case res.err == nil:
			succeeded++
		default:
			appErr, ok := apperr.As(res.err)
			if !ok || appErr.Kind != apperr.KindConflict {
				t.Fatalf("expected apperr.KindConflict for the losing goroutine, got %v (%T)", res.err, res.err)
			}
			conflicted++
		}
	}
	if succeeded != 1 || conflicted != 1 {
		t.Fatalf("RN-CON-01/RN-CON-03 en T2: expected exactly one success and one conflict, got succeeded=%d conflicted=%d (errA=%v errB=%v)",
			succeeded, conflicted, errA, errB)
	}
}
