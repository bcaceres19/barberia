// Package postgres_test (pruebas de integración, HU-067) requiere PostgreSQL
// REAL con las migraciones aplicadas y los mismos fixtures que
// cancel_repository_test.go/reschedule_repository_test.go. El turno base de
// cada prueba se inserta con repo.CreateInternal (año 2030, siempre futuro
// respecto al reloj real): input.Now es un valor SINTÉTICO que cada prueba
// elige libremente relativo a ese starts_at fijo, nunca el reloj del
// sistema -- exactamente lo que CompleteAppointmentService/MarkNoShowService
// ya resuelven antes de llamar al repositorio (close.go), así que estas
// pruebas ejercitan booking.Repository.CompleteAppointment/MarkNoShow
// directamente con el Now que cada caso necesita.
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

func closeFingerprint(seed string) idempotency.Fingerprint {
	sum := sha256.Sum256([]byte(seed))
	return idempotency.Fingerprint(fmt.Sprintf("%x", sum))
}

func newCloseInput(appointmentID, versionToken string, now time.Time) booking.CloseAppointmentInput {
	return booking.CloseAppointmentInput{
		AppointmentID:        appointmentID,
		ExpectedVersionToken: versionToken,
		Actor:                booking.Actor{Type: booking.ActorTypeStaff, StaffUserID: strPtr(staffQ)},
		Now:                  now,
	}
}

// setAppointmentStatus fuerza status directamente por SQL para preparar el
// estado previo de un escenario (mismo criterio que markAppointmentTerminal
// en reschedule_repository_test.go, generalizado a cualquiera de los cuatro
// estados terminales/confirmed en vez de solo 'completed').
func setAppointmentStatus(t *testing.T, db *database.DB, shop database.BarbershopID, appointmentID, status string) {
	t.Helper()
	err := db.InTenantTx(context.Background(), shop, func(ctx context.Context, q database.Queries) error {
		_, err := q.Exec(ctx,
			`UPDATE appointment SET status = $3, resolved_at = now() WHERE barbershop_id = $1 AND id = $2`,
			string(shop), appointmentID, status)
		return err
	})
	if err != nil {
		t.Fatalf("setAppointmentStatus(%s): %v", status, err)
	}
}

func countHistoryEvents(t *testing.T, repo interface {
	ListAppointmentHistory(ctx context.Context, barbershopID, appointmentID string, cursor *booking.HistoryCursor, limit int) ([]booking.HistoryRow, *booking.HistoryCursor, bool, error)
}, shop database.BarbershopID, appointmentID string, eventType booking.EventType) int {
	t.Helper()
	items, _, _, err := repo.ListAppointmentHistory(context.Background(), string(shop), appointmentID, nil, 20)
	if err != nil {
		t.Fatalf("ListAppointmentHistory: %v", err)
	}
	count := 0
	for _, row := range items {
		if row.EventType == eventType {
			count++
		}
	}
	return count
}

// ---- CompleteAppointment (T4 manual) ----

func TestCompleteAppointment_AtExactStartsAt_UpdatesStatusAndInsertsHistory(t *testing.T) {
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

	// CA-067-03: exactamente en starts_at ya es válido.
	result, err := repo.CompleteAppointment(context.Background(), string(shopQ),
		newCloseInput(created.Appointment.ID, detail.VersionToken, start),
		idempotency.Key("complete-key-"+suffix), closeFingerprint("valid-"+suffix))
	if err != nil {
		t.Fatalf("CompleteAppointment: %v", err)
	}
	if result.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("Outcome = %v, want Proceed", result.Decision.Outcome)
	}
	if result.Appointment.Status != booking.StatusCompleted {
		t.Fatalf("Status = %v, want %v", result.Appointment.Status, booking.StatusCompleted)
	}
	if result.Appointment.VersionToken == detail.VersionToken {
		t.Fatalf("VersionToken no cambió tras la escritura")
	}
	// CA-067-04: intervalo y snapshots nunca cambian.
	if result.Appointment.BarberID != detail.BarberID ||
		!result.Appointment.StartsAt.Equal(detail.StartsAt) || !result.Appointment.EndsAt.Equal(detail.EndsAt) ||
		result.Appointment.DurationMinutesSnapshot != detail.DurationMinutesSnapshot {
		t.Fatalf("T4 modificó datos fuera del estado: %+v", result.Appointment)
	}

	if got := countHistoryEvents(t, repo, shopQ, created.Appointment.ID, booking.EventTypeAppointmentCompleted); got != 1 {
		t.Fatalf("appointment_completed insertado %d veces, want exactamente 1", got)
	}
}

// TestCompleteAppointment_BeforeStartsAt_ReturnsValidationWithoutBlockingRetry
// cubre CA-067-03: antes del instante exacto, ni la cita ni el historial se
// tocan, Y la MISMA clave de idempotencia sigue disponible para un reintento
// legítimo una vez llegado starts_at (la transacción entera revierte,
// incluida la reclamación de idempotencia -- TestInTenantTx_RollbackOnCallbackError
// en platform/database ya prueba esa garantía general).
func TestCompleteAppointment_BeforeStartsAt_ReturnsValidationWithoutBlockingRetry(t *testing.T) {
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

	key := idempotency.Key("complete-early-" + suffix)
	early := start.Add(-1 * time.Minute)
	_, err = repo.CompleteAppointment(context.Background(), string(shopQ),
		newCloseInput(created.Appointment.ID, detail.VersionToken, early),
		key, closeFingerprint("early-"+suffix))
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindValidation {
		t.Fatalf("err = %v, want apperr.KindValidation", err)
	}

	stillConfirmed, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}
	if stillConfirmed.Status != booking.StatusConfirmed || stillConfirmed.VersionToken != detail.VersionToken {
		t.Fatalf("el intento prematuro modificó la cita: %+v", stillConfirmed)
	}
	if got := countHistoryEvents(t, repo, shopQ, created.Appointment.ID, booking.EventTypeAppointmentCompleted); got != 0 {
		t.Fatalf("el intento prematuro insertó historial: %d eventos", got)
	}

	// Reintento legítimo con la MISMA clave, ya en starts_at: debe
	// proceder como si fuera la primera vez, no como un conflicto de
	// idempotencia ni una repetición vacía.
	retry, err := repo.CompleteAppointment(context.Background(), string(shopQ),
		newCloseInput(created.Appointment.ID, detail.VersionToken, start),
		key, closeFingerprint("early-"+suffix))
	if err != nil {
		t.Fatalf("CompleteAppointment (reintento tras starts_at): %v", err)
	}
	if retry.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("Outcome del reintento = %v, want Proceed (la clave no debió quedar bloqueada por el intento prematuro)", retry.Decision.Outcome)
	}
	if retry.Appointment.Status != booking.StatusCompleted {
		t.Fatalf("Status del reintento = %v, want %v", retry.Appointment.Status, booking.StatusCompleted)
	}
}

func TestCompleteAppointment_RepeatedByBarber_IsSuccessNoOpNoDuplicateEvent(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	original, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}

	first, err := repo.CompleteAppointment(context.Background(), string(shopQ),
		newCloseInput(created.Appointment.ID, original.VersionToken, start),
		idempotency.Key("complete-first-"+suffix), closeFingerprint("first-"+suffix))
	if err != nil {
		t.Fatalf("CompleteAppointment(1): %v", err)
	}

	// Segunda: clave nueva, token ORIGINAL (obsoleto). Debe tener éxito
	// de todos modos (CA-067-05, mismo resultado repetido).
	second, err := repo.CompleteAppointment(context.Background(), string(shopQ),
		newCloseInput(created.Appointment.ID, original.VersionToken, start),
		idempotency.Key("complete-second-"+suffix), closeFingerprint("second-"+suffix))
	if err != nil {
		t.Fatalf("CompleteAppointment(2) con token obsoleto debió tener éxito (CA-067-05): %v", err)
	}
	if second.Appointment.VersionToken != first.Appointment.VersionToken {
		t.Fatalf("VersionToken cambió en una repetición no-op: %q != %q", second.Appointment.VersionToken, first.Appointment.VersionToken)
	}

	if got := countHistoryEvents(t, repo, shopQ, created.Appointment.ID, booking.EventTypeAppointmentCompleted); got != 1 {
		t.Fatalf("appointment_completed insertado %d veces, want exactamente 1", got)
	}
}

func TestCompleteAppointment_OppositeResultAlreadyNoShow_ReturnsInvalidState(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	setAppointmentStatus(t, db, shopQ, created.Appointment.ID, "no_show")

	afterDetail, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}

	_, err = repo.CompleteAppointment(context.Background(), string(shopQ),
		newCloseInput(created.Appointment.ID, afterDetail.VersionToken, start),
		idempotency.Key("complete-opposite-"+suffix), closeFingerprint("opposite-"+suffix))
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInvalidState {
		t.Fatalf("err = %v, want apperr.KindInvalidState (resultado contrario, CA-067-05)", err)
	}
}

func TestCompleteAppointment_CancelledByBarber_ReturnsInvalidState(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	setAppointmentStatus(t, db, shopQ, created.Appointment.ID, "cancelled_by_barber")

	afterDetail, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}

	_, err = repo.CompleteAppointment(context.Background(), string(shopQ),
		newCloseInput(created.Appointment.ID, afterDetail.VersionToken, start),
		idempotency.Key("complete-cancelled-"+suffix), closeFingerprint("cancelled-"+suffix))
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInvalidState {
		t.Fatalf("err = %v, want apperr.KindInvalidState", err)
	}
}

func TestCompleteAppointment_StaleVersionTokenOnConfirmed_ReturnsVersionConflict(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	staleDetail, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}

	newStarts := start.Add(3 * time.Hour)
	_, err = repo.Reschedule(context.Background(), string(shopQ),
		booking.RescheduleInput{
			AppointmentID:        created.Appointment.ID,
			NewStartsAt:          newStarts,
			NewEndsAt:            newStarts.Add(30 * time.Minute),
			ExpectedVersionToken: staleDetail.VersionToken,
			Actor:                booking.Actor{Type: booking.ActorTypeStaff, StaffUserID: strPtr(staffQ)},
		},
		idempotency.Key("resched-for-complete-"+suffix), closeFingerprint("resched-for-complete-"+suffix))
	if err != nil {
		t.Fatalf("Reschedule (para invalidar el token): %v", err)
	}

	_, err = repo.CompleteAppointment(context.Background(), string(shopQ),
		newCloseInput(created.Appointment.ID, staleDetail.VersionToken, newStarts),
		idempotency.Key("complete-stale-"+suffix), closeFingerprint("stale-"+suffix))
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindVersionConflict {
		t.Fatalf("err = %v, want apperr.KindVersionConflict", err)
	}
}

func TestCompleteAppointment_TenantMismatch_ReturnsNotFound(t *testing.T) {
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

	_, err = repo.CompleteAppointment(context.Background(), string(shopR),
		newCloseInput(created.Appointment.ID, detail.VersionToken, start),
		idempotency.Key("complete-tenant-"+suffix), closeFingerprint("tenant-"+suffix))
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindNotFound {
		t.Fatalf("err = %v, want apperr.KindNotFound (RN-TEN-01)", err)
	}
}

func TestCompleteAppointment_RepeatedKeySameIntent_Replays(t *testing.T) {
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

	input := newCloseInput(created.Appointment.ID, detail.VersionToken, start)
	key := idempotency.Key("complete-replay-" + suffix)
	fingerprint := closeFingerprint("replay-" + suffix)

	first, err := repo.CompleteAppointment(context.Background(), string(shopQ), input, key, fingerprint)
	if err != nil {
		t.Fatalf("CompleteAppointment(1): %v", err)
	}
	second, err := repo.CompleteAppointment(context.Background(), string(shopQ), input, key, fingerprint)
	if err != nil {
		t.Fatalf("CompleteAppointment(2): %v", err)
	}
	if second.Response.Body != first.Response.Body {
		t.Fatalf("la repetición no reprodujo el mismo cuerpo byte a byte:\n1: %s\n2: %s", first.Response.Body, second.Response.Body)
	}
	if got := countHistoryEvents(t, repo, shopQ, created.Appointment.ID, booking.EventTypeAppointmentCompleted); got != 1 {
		t.Fatalf("appointment_completed insertado %d veces, want exactamente 1", got)
	}
}

// TestCompleteAppointment_ExclusionConserved_CrossingAppointmentStillConflicts
// cubre CA-067-04: a diferencia de cancelar (HU-066), completar NUNCA libera
// el intervalo -- otra cita que cruce la misma franja sigue rechazándose.
func TestCompleteAppointment_ExclusionConserved_CrossingAppointmentStillConflicts(t *testing.T) {
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
		idempotency.Key("complete-excl-"+suffix), closeFingerprint("excl-"+suffix)); err != nil {
		t.Fatalf("CompleteAppointment: %v", err)
	}

	_, err = repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix+"-cruce"))
	if err == nil {
		t.Fatalf("expected un cruce de agenda: completar nunca libera el intervalo (CA-067-04)")
	}
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindConflict {
		t.Fatalf("expected apperr.KindConflict, got %v (%T)", err, err)
	}
}

// ---- MarkNoShow (T7) ----

func TestMarkNoShow_AtExactStartsAt_UpdatesStatusAndInsertsHistory(t *testing.T) {
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

	result, err := repo.MarkNoShow(context.Background(), string(shopQ),
		newCloseInput(created.Appointment.ID, detail.VersionToken, start),
		idempotency.Key("noshow-key-"+suffix), closeFingerprint("valid-"+suffix))
	if err != nil {
		t.Fatalf("MarkNoShow: %v", err)
	}
	if result.Appointment.Status != booking.StatusNoShow {
		t.Fatalf("Status = %v, want %v", result.Appointment.Status, booking.StatusNoShow)
	}
	if !result.Appointment.StartsAt.Equal(detail.StartsAt) || !result.Appointment.EndsAt.Equal(detail.EndsAt) {
		t.Fatalf("T7 modificó el intervalo: %+v", result.Appointment)
	}
	if got := countHistoryEvents(t, repo, shopQ, created.Appointment.ID, booking.EventTypeAppointmentNoShow); got != 1 {
		t.Fatalf("appointment_no_show insertado %d veces, want exactamente 1", got)
	}
}

func TestMarkNoShow_BeforeStartsAt_ReturnsValidationWithoutBlockingRetry(t *testing.T) {
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

	key := idempotency.Key("noshow-early-" + suffix)
	early := start.Add(-1 * time.Minute)
	_, err = repo.MarkNoShow(context.Background(), string(shopQ),
		newCloseInput(created.Appointment.ID, detail.VersionToken, early),
		key, closeFingerprint("early-"+suffix))
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindValidation {
		t.Fatalf("err = %v, want apperr.KindValidation", err)
	}

	if got := countHistoryEvents(t, repo, shopQ, created.Appointment.ID, booking.EventTypeAppointmentNoShow); got != 0 {
		t.Fatalf("el intento prematuro insertó historial: %d eventos", got)
	}

	retry, err := repo.MarkNoShow(context.Background(), string(shopQ),
		newCloseInput(created.Appointment.ID, detail.VersionToken, start),
		key, closeFingerprint("early-"+suffix))
	if err != nil {
		t.Fatalf("MarkNoShow (reintento tras starts_at): %v", err)
	}
	if retry.Appointment.Status != booking.StatusNoShow {
		t.Fatalf("Status del reintento = %v, want %v", retry.Appointment.Status, booking.StatusNoShow)
	}
}

func TestMarkNoShow_RepeatedByBarber_IsSuccessNoOpNoDuplicateEvent(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	original, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}

	first, err := repo.MarkNoShow(context.Background(), string(shopQ),
		newCloseInput(created.Appointment.ID, original.VersionToken, start),
		idempotency.Key("noshow-first-"+suffix), closeFingerprint("first-"+suffix))
	if err != nil {
		t.Fatalf("MarkNoShow(1): %v", err)
	}

	second, err := repo.MarkNoShow(context.Background(), string(shopQ),
		newCloseInput(created.Appointment.ID, original.VersionToken, start),
		idempotency.Key("noshow-second-"+suffix), closeFingerprint("second-"+suffix))
	if err != nil {
		t.Fatalf("MarkNoShow(2) con token obsoleto debió tener éxito (CA-067-05): %v", err)
	}
	if second.Appointment.VersionToken != first.Appointment.VersionToken {
		t.Fatalf("VersionToken cambió en una repetición no-op: %q != %q", second.Appointment.VersionToken, first.Appointment.VersionToken)
	}
	if got := countHistoryEvents(t, repo, shopQ, created.Appointment.ID, booking.EventTypeAppointmentNoShow); got != 1 {
		t.Fatalf("appointment_no_show insertado %d veces, want exactamente 1", got)
	}
}

func TestMarkNoShow_OppositeResultAlreadyCompleted_ReturnsInvalidState(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	setAppointmentStatus(t, db, shopQ, created.Appointment.ID, "completed")

	afterDetail, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}

	_, err = repo.MarkNoShow(context.Background(), string(shopQ),
		newCloseInput(created.Appointment.ID, afterDetail.VersionToken, start),
		idempotency.Key("noshow-opposite-"+suffix), closeFingerprint("opposite-"+suffix))
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInvalidState {
		t.Fatalf("err = %v, want apperr.KindInvalidState (resultado contrario, CA-067-05)", err)
	}
}

func TestMarkNoShow_TenantMismatch_ReturnsNotFound(t *testing.T) {
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

	_, err = repo.MarkNoShow(context.Background(), string(shopR),
		newCloseInput(created.Appointment.ID, detail.VersionToken, start),
		idempotency.Key("noshow-tenant-"+suffix), closeFingerprint("tenant-"+suffix))
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindNotFound {
		t.Fatalf("err = %v, want apperr.KindNotFound (RN-TEN-01)", err)
	}
}

func TestMarkNoShow_RepeatedKeySameIntent_Replays(t *testing.T) {
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

	input := newCloseInput(created.Appointment.ID, detail.VersionToken, start)
	key := idempotency.Key("noshow-replay-" + suffix)
	fingerprint := closeFingerprint("replay-" + suffix)

	first, err := repo.MarkNoShow(context.Background(), string(shopQ), input, key, fingerprint)
	if err != nil {
		t.Fatalf("MarkNoShow(1): %v", err)
	}
	second, err := repo.MarkNoShow(context.Background(), string(shopQ), input, key, fingerprint)
	if err != nil {
		t.Fatalf("MarkNoShow(2): %v", err)
	}
	if second.Response.Body != first.Response.Body {
		t.Fatalf("la repetición no reprodujo el mismo cuerpo byte a byte:\n1: %s\n2: %s", first.Response.Body, second.Response.Body)
	}
}

// ---- Carrera de tres resultados (CA-067-06) ----

// TestClose_ConcurrentCompleteNoShowCancel_ExactlyOneTerminalAndOneEvent
// cubre la "carrera coordinada de tres resultados" de las pruebas
// obligatorias: completar, marcar inasistencia y cancelar disparados al
// mismo tiempo sobre LA MISMA cita `confirmed`, con claves de idempotencia
// (y por lo tanto operaciones) distintas. `FOR UPDATE` serializa las tres
// transacciones: quien gane la fila la deja en su propio estado terminal;
// las otras dos, al verla ya distinta de `confirmed`, responden
// invalid-state (CA-067-05) sin escribir nada. El resultado final debe ser
// EXACTAMENTE uno de los tres estados terminales, con EXACTAMENTE un evento
// de historial en total (nunca cero, nunca dos).
func TestClose_ConcurrentCompleteNoShowCancel_ExactlyOneTerminalAndOneEvent(t *testing.T) {
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

	var (
		wg           sync.WaitGroup
		startBarrier = make(chan struct{})

		completeErr, noShowErr, cancelErr error
		completeOK, noShowOK, cancelOK    bool
	)
	wg.Add(3)
	go func() {
		defer wg.Done()
		close(startBarrier)
		res, err := repo.CompleteAppointment(context.Background(), string(shopQ),
			newCloseInput(created.Appointment.ID, detail.VersionToken, start),
			idempotency.Key("race-complete-"+suffix), closeFingerprint("race-complete-"+suffix))
		completeErr = err
		completeOK = err == nil && res.Appointment.Status == booking.StatusCompleted
	}()
	go func() {
		defer wg.Done()
		<-startBarrier
		res, err := repo.MarkNoShow(context.Background(), string(shopQ),
			newCloseInput(created.Appointment.ID, detail.VersionToken, start),
			idempotency.Key("race-noshow-"+suffix), closeFingerprint("race-noshow-"+suffix))
		noShowErr = err
		noShowOK = err == nil && res.Appointment.Status == booking.StatusNoShow
	}()
	go func() {
		defer wg.Done()
		<-startBarrier
		res, err := repo.CancelByBarber(context.Background(), string(shopQ),
			booking.CancelAppointmentByBarberInput{
				AppointmentID:        created.Appointment.ID,
				ExpectedVersionToken: detail.VersionToken,
				Actor:                booking.Actor{Type: booking.ActorTypeStaff, StaffUserID: strPtr(staffQ)},
			},
			idempotency.Key("race-cancel-"+suffix), closeFingerprint("race-cancel-"+suffix))
		cancelErr = err
		cancelOK = err == nil && res.Appointment.Status == booking.StatusCancelledByBarber
	}()
	wg.Wait()

	wins := 0
	if completeOK {
		wins++
	}
	if noShowOK {
		wins++
	}
	if cancelOK {
		wins++
	}
	if wins != 1 {
		t.Fatalf("exactamente un resultado debió ganar la carrera, ganaron %d (complete=%v/%v noShow=%v/%v cancel=%v/%v)",
			wins, completeOK, completeErr, noShowOK, noShowErr, cancelOK, cancelErr)
	}

	// Las dos operaciones perdedoras deben haber respondido invalid-state
	// (CA-067-05/CA-066-01), nunca un error inesperado ni un éxito parcial.
	for _, loserErr := range []error{completeErr, noShowErr, cancelErr} {
		if loserErr == nil {
			continue
		}
		appErr, ok := apperr.As(loserErr)
		if !ok || appErr.Kind != apperr.KindInvalidState {
			t.Fatalf("err de la operación perdedora = %v, want apperr.KindInvalidState", loserErr)
		}
	}

	final, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}
	if final.Status == booking.StatusConfirmed {
		t.Fatalf("la cita quedó confirmed tras la carrera, want exactamente un terminal")
	}

	totalEvents := countHistoryEvents(t, repo, shopQ, created.Appointment.ID, booking.EventTypeAppointmentCompleted) +
		countHistoryEvents(t, repo, shopQ, created.Appointment.ID, booking.EventTypeAppointmentNoShow) +
		countHistoryEvents(t, repo, shopQ, created.Appointment.ID, booking.EventTypeAppointmentCancelledByBarber)
	if totalEvents != 1 {
		t.Fatalf("total de eventos terminales insertados bajo la carrera = %d, want exactamente 1", totalEvents)
	}
}
