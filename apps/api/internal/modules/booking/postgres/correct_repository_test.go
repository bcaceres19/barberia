// Package postgres_test (pruebas de integración, HU-068) requiere PostgreSQL
// REAL con las migraciones aplicadas y los mismos fixtures que
// close_repository_test.go/cancel_repository_test.go. Mismo criterio que
// close_repository_test.go: input.Now es un valor SINTÉTICO que cada prueba
// elige libremente relativo al starts_at fijo de baseStart, nunca el reloj
// del sistema.
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/booking/postgres/...
package postgres_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"system-barbershop/internal/modules/booking"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/idempotency"
)

func newCorrectInput(appointmentID, destination, reason, versionToken string, now time.Time) booking.CorrectAppointmentStatusInput {
	return booking.CorrectAppointmentStatusInput{
		AppointmentID:        appointmentID,
		DestinationStatus:    booking.Status(destination),
		Reason:               reason,
		ExpectedVersionToken: versionToken,
		Actor:                booking.Actor{Type: booking.ActorTypeStaff, StaffUserID: strPtr(staffQ)},
		Now:                  now,
	}
}

// ---- Matriz de dominio: los 12 cambios posibles entre los cuatro
// terminales (CA-068-01, CA-068-03) ----

func TestCorrectAppointmentStatus_TwelveTerminalTransitions_UpdateStatusAndAppendSingleHistoryEvent(t *testing.T) {
	terminals := []string{"completed", "no_show", "cancelled_by_customer", "cancelled_by_barber"}

	for _, from := range terminals {
		for _, to := range terminals {
			if from == to {
				continue
			}
			from, to := from, to
			t.Run(from+"_to_"+to, func(t *testing.T) {
				db := setupTestDB(t)
				repo := newRepository(db)
				suffix := uniqueSuffix(t)
				start := baseStart(t)

				created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
				if err != nil {
					t.Fatalf("CreateInternal: %v", err)
				}
				setAppointmentStatus(t, db, shopQ, created.Appointment.ID, from)

				before, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
				if err != nil {
					t.Fatalf("GetAppointmentDetail: %v", err)
				}

				reason := "Corrección de " + from + " a " + to + " por error de clasificación."
				result, err := repo.CorrectAppointmentStatus(context.Background(), string(shopQ),
					newCorrectInput(created.Appointment.ID, to, reason, before.VersionToken, start),
					idempotency.Key("correct-"+from+"-"+to+"-"+suffix), closeFingerprint("correct-"+from+"-"+to+"-"+suffix))
				if err != nil {
					t.Fatalf("CorrectAppointmentStatus(%s->%s): %v", from, to, err)
				}
				if string(result.Appointment.Status) != to {
					t.Fatalf("Status = %v, want %v", result.Appointment.Status, to)
				}
				if result.Appointment.VersionToken == before.VersionToken {
					t.Fatalf("VersionToken no cambió tras la escritura")
				}
				// CA-068-03: el intervalo y los snapshots nunca cambian, T8
				// solo corrige el estado.
				if !result.Appointment.StartsAt.Equal(before.StartsAt) || !result.Appointment.EndsAt.Equal(before.EndsAt) {
					t.Fatalf("T8 modificó el intervalo: %+v", result.Appointment)
				}

				items, _, _, err := repo.ListAppointmentHistory(context.Background(), string(shopQ), created.Appointment.ID, nil, 20)
				if err != nil {
					t.Fatalf("ListAppointmentHistory: %v", err)
				}
				var corrections []booking.HistoryRow
				for _, row := range items {
					if row.EventType == booking.EventTypeAppointmentStatusCorrected {
						corrections = append(corrections, row)
					}
				}
				if len(corrections) != 1 {
					t.Fatalf("appointment_status_corrected insertado %d veces, want exactamente 1", len(corrections))
				}
				got := corrections[0]
				if got.Reason == nil || *got.Reason != reason {
					t.Fatalf("Reason = %v, want %q", got.Reason, reason)
				}
				if len(got.Changes) != 1 || got.Changes[0].FieldName != "status" {
					t.Fatalf("Changes = %+v, want un único cambio de status", got.Changes)
				}
				if got.Changes[0].PreviousValue == nil || *got.Changes[0].PreviousValue != from {
					t.Fatalf("PreviousValue = %v, want %q", got.Changes[0].PreviousValue, from)
				}
				if got.Changes[0].NewValue == nil || *got.Changes[0].NewValue != to {
					t.Fatalf("NewValue = %v, want %q", got.Changes[0].NewValue, to)
				}
			})
		}
	}
}

// ---- Casos fuera de la matriz feliz ----

func TestCorrectAppointmentStatus_StillConfirmed_ReturnsInvalidState(t *testing.T) {
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

	_, err = repo.CorrectAppointmentStatus(context.Background(), string(shopQ),
		newCorrectInput(created.Appointment.ID, "no_show", "motivo", detail.VersionToken, start),
		idempotency.Key("correct-confirmed-"+suffix), closeFingerprint("confirmed-"+suffix))
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInvalidState {
		t.Fatalf("err = %v, want apperr.KindInvalidState (todavía confirmed, CA-068-01)", err)
	}
}

// TestCorrectAppointmentStatus_SameDestination_IsNoOpSuccessNoDuplicateEvent
// cubre CA-068-02: una solicitud nueva hacia el MISMO estado vigente
// responde éxito sin efecto, sin comparar versionToken ni duplicar el
// evento.
func TestCorrectAppointmentStatus_SameDestination_IsNoOpSuccessNoDuplicateEvent(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	setAppointmentStatus(t, db, shopQ, created.Appointment.ID, "completed")
	before, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}

	first, err := repo.CorrectAppointmentStatus(context.Background(), string(shopQ),
		newCorrectInput(created.Appointment.ID, "no_show", "primer motivo", before.VersionToken, start),
		idempotency.Key("correct-noop-1-"+suffix), closeFingerprint("noop-1-"+suffix))
	if err != nil {
		t.Fatalf("CorrectAppointmentStatus(1): %v", err)
	}

	// Segunda: clave nueva, token OBSOLETO (el de antes de la primera
	// corrección), pero hacia el MISMO estado ya vigente (no_show). Debe
	// tener éxito de todos modos.
	second, err := repo.CorrectAppointmentStatus(context.Background(), string(shopQ),
		newCorrectInput(created.Appointment.ID, "no_show", "segundo motivo", before.VersionToken, start),
		idempotency.Key("correct-noop-2-"+suffix), closeFingerprint("noop-2-"+suffix))
	if err != nil {
		t.Fatalf("CorrectAppointmentStatus(2) hacia el mismo estado debió tener éxito (CA-068-02): %v", err)
	}
	if second.Appointment.VersionToken != first.Appointment.VersionToken {
		t.Fatalf("VersionToken cambió en una repetición no-op: %q != %q", second.Appointment.VersionToken, first.Appointment.VersionToken)
	}

	items, _, _, err := repo.ListAppointmentHistory(context.Background(), string(shopQ), created.Appointment.ID, nil, 20)
	if err != nil {
		t.Fatalf("ListAppointmentHistory: %v", err)
	}
	count := 0
	for _, row := range items {
		if row.EventType == booking.EventTypeAppointmentStatusCorrected {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("appointment_status_corrected insertado %d veces, want exactamente 1 (el no-op no duplica)", count)
	}
}

// TestCorrectAppointmentStatus_CancelledToOccupying_BeforeStartsAt_ReturnsValidation
// cubre CA-068-04: de un terminal cancelado hacia completed/no_show,
// starts_at todavía debe pasar.
func TestCorrectAppointmentStatus_CancelledToOccupying_BeforeStartsAt_ReturnsValidation(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	setAppointmentStatus(t, db, shopQ, created.Appointment.ID, "cancelled_by_barber")
	before, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}

	early := start.Add(-1 * time.Minute)
	_, err = repo.CorrectAppointmentStatus(context.Background(), string(shopQ),
		newCorrectInput(created.Appointment.ID, "completed", "motivo", before.VersionToken, early),
		idempotency.Key("correct-early-"+suffix), closeFingerprint("early-"+suffix))
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindValidation {
		t.Fatalf("err = %v, want apperr.KindValidation", err)
	}

	stillCancelled, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}
	if stillCancelled.Status != booking.StatusCancelledByBarber || stillCancelled.VersionToken != before.VersionToken {
		t.Fatalf("el intento prematuro modificó la cita: %+v", stillCancelled)
	}
}

// TestCorrectAppointmentStatus_CancelledToOccupying_CrossesAnotherAppointment_ReturnsConflictAndRollsBack
// cubre CA-068-04: corregir desde un cancelado hacia completed/no_show que
// volvería a ocupar una franja ya ocupada por OTRA cita del mismo barbero
// responde 409 y la transacción completa hace rollback (sin escribir nada).
func TestCorrectAppointmentStatus_CancelledToOccupying_CrossesAnotherAppointment_ReturnsConflictAndRollsBack(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	// A: cancelada, ya no ocupa agenda.
	a, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix+"-a"))
	if err != nil {
		t.Fatalf("CreateInternal(A): %v", err)
	}
	setAppointmentStatus(t, db, shopQ, a.Appointment.ID, "cancelled_by_barber")
	beforeA, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), a.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail(A): %v", err)
	}

	// B: mismo barbero, MISMO intervalo exacto que A -- puede crearse porque
	// A ya no ocupa agenda.
	b, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix+"-b"))
	if err != nil {
		t.Fatalf("CreateInternal(B): %v", err)
	}
	_ = b

	// Corregir A hacia completed volvería a ocupar la misma franja que B ya
	// ocupa (B sigue confirmed): debe responder Conflict, sin tocar A.
	_, err = repo.CorrectAppointmentStatus(context.Background(), string(shopQ),
		newCorrectInput(a.Appointment.ID, "completed", "motivo", beforeA.VersionToken, start),
		idempotency.Key("correct-cross-"+suffix), closeFingerprint("cross-"+suffix))
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindConflict {
		t.Fatalf("err = %v, want apperr.KindConflict (CA-068-04)", err)
	}

	afterA, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), a.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail(A after): %v", err)
	}
	if afterA.Status != booking.StatusCancelledByBarber || afterA.VersionToken != beforeA.VersionToken {
		t.Fatalf("el cruce rechazado modificó A: %+v", afterA)
	}

	items, _, _, err := repo.ListAppointmentHistory(context.Background(), string(shopQ), a.Appointment.ID, nil, 20)
	if err != nil {
		t.Fatalf("ListAppointmentHistory: %v", err)
	}
	for _, row := range items {
		if row.EventType == booking.EventTypeAppointmentStatusCorrected {
			t.Fatalf("el cruce rechazado insertó historial de todos modos: %+v", row)
		}
	}
}

// TestCorrectAppointmentStatus_OccupyingToCancelled_ReleasesExclusion_NewAppointmentCanUseFreedSlot
// cubre CA-068-05: corregir desde completed/no_show hacia un cancelado
// libera la exclusión en la misma confirmación.
func TestCorrectAppointmentStatus_OccupyingToCancelled_ReleasesExclusion_NewAppointmentCanUseFreedSlot(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	setAppointmentStatus(t, db, shopQ, created.Appointment.ID, "completed")
	before, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}

	if _, err := repo.CorrectAppointmentStatus(context.Background(), string(shopQ),
		newCorrectInput(created.Appointment.ID, "cancelled_by_barber", "motivo", before.VersionToken, start),
		idempotency.Key("correct-release-"+suffix), closeFingerprint("release-"+suffix)); err != nil {
		t.Fatalf("CorrectAppointmentStatus: %v", err)
	}

	if _, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix+"-new")); err != nil {
		t.Fatalf("CreateInternal tras liberar la franja debió tener éxito (CA-068-05): %v", err)
	}
}

func TestCorrectAppointmentStatus_StaleVersionToken_ReturnsVersionConflict(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	setAppointmentStatus(t, db, shopQ, created.Appointment.ID, "completed")
	stale, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}

	// Invalida el token con una corrección real distinta del destino que se
	// probará después.
	if _, err := repo.CorrectAppointmentStatus(context.Background(), string(shopQ),
		newCorrectInput(created.Appointment.ID, "no_show", "primer motivo", stale.VersionToken, start),
		idempotency.Key("correct-invalidate-"+suffix), closeFingerprint("invalidate-"+suffix)); err != nil {
		t.Fatalf("CorrectAppointmentStatus (para invalidar el token): %v", err)
	}

	_, err = repo.CorrectAppointmentStatus(context.Background(), string(shopQ),
		newCorrectInput(created.Appointment.ID, "cancelled_by_barber", "motivo", stale.VersionToken, start),
		idempotency.Key("correct-stale-"+suffix), closeFingerprint("stale-"+suffix))
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindVersionConflict {
		t.Fatalf("err = %v, want apperr.KindVersionConflict", err)
	}
}

func TestCorrectAppointmentStatus_TenantMismatch_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	setAppointmentStatus(t, db, shopQ, created.Appointment.ID, "completed")
	detail, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}

	_, err = repo.CorrectAppointmentStatus(context.Background(), string(shopR),
		newCorrectInput(created.Appointment.ID, "no_show", "motivo", detail.VersionToken, start),
		idempotency.Key("correct-tenant-"+suffix), closeFingerprint("tenant-"+suffix))
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindNotFound {
		t.Fatalf("err = %v, want apperr.KindNotFound (RN-TEN-01)", err)
	}
}

func TestCorrectAppointmentStatus_RepeatedKeySameIntent_Replays(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	setAppointmentStatus(t, db, shopQ, created.Appointment.ID, "completed")
	detail, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}

	input := newCorrectInput(created.Appointment.ID, "no_show", "motivo repetido", detail.VersionToken, start)
	key := idempotency.Key("correct-replay-" + suffix)
	fingerprint := closeFingerprint("replay-" + suffix)

	first, err := repo.CorrectAppointmentStatus(context.Background(), string(shopQ), input, key, fingerprint)
	if err != nil {
		t.Fatalf("CorrectAppointmentStatus(1): %v", err)
	}
	second, err := repo.CorrectAppointmentStatus(context.Background(), string(shopQ), input, key, fingerprint)
	if err != nil {
		t.Fatalf("CorrectAppointmentStatus(2): %v", err)
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
		if row.EventType == booking.EventTypeAppointmentStatusCorrected {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("appointment_status_corrected insertado %d veces, want exactamente 1", count)
	}
}

// TestCorrectAppointmentStatus_AppendOnly_PreviousEventsUntouched cubre
// CA-068-03: los eventos anteriores (aquí, appointment_completed de T4
// manual) permanecen intactos y en orden, la corrección solo agrega
// evidencia.
func TestCorrectAppointmentStatus_AppendOnly_PreviousEventsUntouched(t *testing.T) {
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

	if _, err := repo.CompleteAppointment(context.Background(), string(shopQ),
		newCloseInput(created.Appointment.ID, detail.VersionToken, start),
		idempotency.Key("append-complete-"+suffix), closeFingerprint("append-complete-"+suffix)); err != nil {
		t.Fatalf("CompleteAppointment: %v", err)
	}
	afterComplete, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}

	if _, err := repo.CorrectAppointmentStatus(context.Background(), string(shopQ),
		newCorrectInput(created.Appointment.ID, "no_show", "motivo append-only", afterComplete.VersionToken, start),
		idempotency.Key("append-correct-"+suffix), closeFingerprint("append-correct-"+suffix)); err != nil {
		t.Fatalf("CorrectAppointmentStatus: %v", err)
	}

	items, _, _, err := repo.ListAppointmentHistory(context.Background(), string(shopQ), created.Appointment.ID, nil, 20)
	if err != nil {
		t.Fatalf("ListAppointmentHistory: %v", err)
	}
	var eventTypes []booking.EventType
	for _, row := range items {
		eventTypes = append(eventTypes, row.EventType)
	}
	wantOrder := []booking.EventType{
		booking.EventTypeAppointmentCreated,
		booking.EventTypeAppointmentCompleted,
		booking.EventTypeAppointmentStatusCorrected,
	}
	if len(eventTypes) != len(wantOrder) {
		t.Fatalf("eventos = %v, want %v", eventTypes, wantOrder)
	}
	for i, want := range wantOrder {
		if eventTypes[i] != want {
			t.Fatalf("eventos[%d] = %v, want %v (orden = %v)", i, eventTypes[i], want, eventTypes)
		}
	}
}

// TestCorrectAppointmentStatus_ConcurrentCorrections_ExactlyOneWinsWithVersionProtection
// cubre la carrera coordinada: dos correcciones concurrentes sobre LA MISMA
// cita, ambas partiendo del MISMO versionToken leído antes de la carrera
// (`FOR UPDATE` serializa las dos transacciones). Quien gane deja la fila en
// su propio destino; la otra, al ver la fila ya distinta bajo el MISMO
// token obsoleto, responde VersionConflict (nunca un error inesperado ni un
// segundo evento de historial).
func TestCorrectAppointmentStatus_ConcurrentCorrections_ExactlyOneWinsWithVersionProtection(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	setAppointmentStatus(t, db, shopQ, created.Appointment.ID, "cancelled_by_barber")
	before, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}

	var (
		wg           sync.WaitGroup
		startBarrier = make(chan struct{})

		toNoShowErr, toCustomerErr error
		toNoShowOK, toCustomerOK   bool
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		close(startBarrier)
		res, err := repo.CorrectAppointmentStatus(context.Background(), string(shopQ),
			newCorrectInput(created.Appointment.ID, "no_show", "motivo carrera A", before.VersionToken, start),
			idempotency.Key("race-correct-noshow-"+suffix), closeFingerprint("race-correct-noshow-"+suffix))
		toNoShowErr = err
		toNoShowOK = err == nil && res.Appointment.Status == booking.StatusNoShow
	}()
	go func() {
		defer wg.Done()
		<-startBarrier
		res, err := repo.CorrectAppointmentStatus(context.Background(), string(shopQ),
			newCorrectInput(created.Appointment.ID, "cancelled_by_customer", "motivo carrera B", before.VersionToken, start),
			idempotency.Key("race-correct-customer-"+suffix), closeFingerprint("race-correct-customer-"+suffix))
		toCustomerErr = err
		toCustomerOK = err == nil && res.Appointment.Status == booking.StatusCancelledByCustomer
	}()
	wg.Wait()

	wins := 0
	if toNoShowOK {
		wins++
	}
	if toCustomerOK {
		wins++
	}
	if wins != 1 {
		t.Fatalf("exactamente una corrección debió ganar la carrera, ganaron %d (noShow=%v/%v customer=%v/%v)",
			wins, toNoShowOK, toNoShowErr, toCustomerOK, toCustomerErr)
	}

	for _, loserErr := range []error{toNoShowErr, toCustomerErr} {
		if loserErr == nil {
			continue
		}
		appErr, ok := apperr.As(loserErr)
		if !ok || appErr.Kind != apperr.KindVersionConflict {
			t.Fatalf("err de la corrección perdedora = %v, want apperr.KindVersionConflict", loserErr)
		}
	}

	items, _, _, err := repo.ListAppointmentHistory(context.Background(), string(shopQ), created.Appointment.ID, nil, 20)
	if err != nil {
		t.Fatalf("ListAppointmentHistory: %v", err)
	}
	count := 0
	for _, row := range items {
		if row.EventType == booking.EventTypeAppointmentStatusCorrected {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("total de correcciones insertadas bajo la carrera = %d, want exactamente 1", count)
	}
}
