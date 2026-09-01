// Package postgres_test (pruebas de integración, HU-064) requiere
// PostgreSQL REAL con las dieciséis migraciones aplicadas y
// database/testdata/dos_barberias.sql + database/testdata/hu060_citas.sql
// cargados: mismos fixtures que repository_test.go/agenda_repository_test.go
// (HU-060/HU-062), reutilizados aquí porque HU-064 solo LEE appointment,
// customer, appointment_history y appointment_history_change, sin tabla
// nueva. El turno base de cada prueba se inserta con repo.CreateInternal (la
// misma primitiva probada en repository_test.go): esa inserción ya emite un
// evento appointment_created real, que estas pruebas también verifican. Las
// filas adicionales de historial que la paginación necesita (HU-064 no
// expone todavía ninguna mutación que las genere) se insertan con SQL
// directo dentro de InTenantTx, nunca contra la validación de dominio.
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/booking/postgres/...
package postgres_test

import (
	"context"
	"testing"
	"time"

	"system-barbershop/internal/modules/booking"
	"system-barbershop/internal/platform/database"
)

// insertSyntheticHistoryRow agrega, dentro de la misma barbería que
// appointmentID, una fila de appointment_history con actor 'system' (sin FK
// que resolver) para poder probar paginación/orden sin depender de una
// mutación que HU-064 no expone todavía. occurred_at se fija explícitamente
// para controlar el orden esperado.
func insertSyntheticHistoryRow(t *testing.T, db *database.DB, shop database.BarbershopID, appointmentID string, occurredAt time.Time, id string) {
	t.Helper()
	err := db.InTenantTx(context.Background(), shop, func(ctx context.Context, q database.Queries) error {
		// appointment_history_reason_ck exige reason NOT NULL cuando
		// event_type = 'appointment_status_corrected' (única razón por la
		// que este evento se eligió como sintético: no exige ningún otro
		// dato ni FK que resolver).
		_, err := q.Exec(ctx, `
			INSERT INTO appointment_history (id, barbershop_id, appointment_id, event_type, actor_type, reason, occurred_at)
			VALUES ($1, $2, $3, 'appointment_status_corrected', 'system', 'corrección sintética de prueba', $4)`,
			id, string(shop), appointmentID, occurredAt)
		return err
	})
	if err != nil {
		t.Fatalf("insertSyntheticHistoryRow: %v", err)
	}
}

func TestGetAppointmentDetail_Found_ReturnsCustomerAndServiceSnapshot(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}

	detail, found, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}
	if !found {
		t.Fatalf("found = false, want true")
	}
	if detail.ID != created.Appointment.ID || detail.BarberID != barberQ1 {
		t.Fatalf("detalle inesperado: %+v", detail)
	}
	if detail.CustomerFullName != "Cliente de prueba "+suffix {
		t.Fatalf("CustomerFullName = %q", detail.CustomerFullName)
	}
	if detail.ServiceNameSnapshot != "Corte de prueba "+suffix || detail.DurationMinutesSnapshot != 30 {
		t.Fatalf("snapshot de servicio inesperado: %+v", detail)
	}
	if detail.PriceAmountCentsSnapshot != 2000000 || detail.CurrencySnapshot != "COP" {
		t.Fatalf("precio/moneda inesperados: %+v", detail)
	}
	if detail.VersionToken == "" {
		t.Fatalf("VersionToken vacío")
	}
	// BarberFullName SIEMPRE llega vacío del repositorio: se resuelve aparte
	// vía BarberNamePort (booking.DetailService), no aquí (CA-002-06).
	if detail.BarberFullName != "" {
		t.Fatalf("BarberFullName = %q, want vacío (se resuelve en el caso de uso)", detail.BarberFullName)
	}
}

func TestGetAppointmentDetail_VersionToken_StableAcrossReadsWithoutWrite(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}

	first, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail(1): %v", err)
	}
	second, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail(2): %v", err)
	}
	if first.VersionToken != second.VersionToken {
		t.Fatalf("VersionToken cambió sin ninguna escritura: %q != %q", first.VersionToken, second.VersionToken)
	}
}

func TestGetAppointmentDetail_NonexistentID_ReturnsFoundFalse(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	_, found, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}
	if found {
		t.Fatalf("found = true, want false para un appointmentId inexistente")
	}
}

// TestGetAppointmentDetail_TenantMismatch_ReturnsFoundFalse verifica
// RN-TEN-01: una cita real de shopQ nunca aparece al consultarla con
// barbershopID de shopR, aunque el identificador coincida literalmente.
func TestGetAppointmentDetail_TenantMismatch_ReturnsFoundFalse(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}

	_, found, err := repo.GetAppointmentDetail(context.Background(), string(shopR), created.Appointment.ID)
	if err != nil {
		t.Fatalf("GetAppointmentDetail(shopR): %v", err)
	}
	if found {
		t.Fatalf("found = true, want false: una cita de shopQ no debió verse desde shopR")
	}
}

func TestListAppointmentHistory_AfterCreate_ReturnsTheCreatedEvent(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}

	items, next, found, err := repo.ListAppointmentHistory(context.Background(), string(shopQ), created.Appointment.ID, nil, 20)
	if err != nil {
		t.Fatalf("ListAppointmentHistory: %v", err)
	}
	if !found {
		t.Fatalf("found = false, want true")
	}
	if next != nil {
		t.Fatalf("next = %+v, want nil (una sola página)", next)
	}
	if len(items) != 1 {
		t.Fatalf("items = %+v, want exactamente 1 (appointment_created)", items)
	}
	row := items[0]
	if row.EventType != booking.EventTypeAppointmentCreated {
		t.Fatalf("EventType = %q, want %q", row.EventType, booking.EventTypeAppointmentCreated)
	}
	if row.ActorType != booking.ActorTypeStaff || row.ActorStaffUserID == nil || *row.ActorStaffUserID != staffQ {
		t.Fatalf("actor inesperado: %+v", row)
	}
	if row.ActorCustomerID != nil {
		t.Fatalf("ActorCustomerID = %v, want nil para un actor staff", row.ActorCustomerID)
	}
}

func TestListAppointmentHistory_NonexistentAppointment_ReturnsFoundFalse(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	items, next, found, err := repo.ListAppointmentHistory(context.Background(), string(shopQ), "00000000-0000-0000-0000-000000000000", nil, 20)
	if err != nil {
		t.Fatalf("ListAppointmentHistory: %v", err)
	}
	if found || items != nil || next != nil {
		t.Fatalf("found/items/next = %v/%v/%v, want false/nil/nil", found, items, next)
	}
}

// TestListAppointmentHistory_TenantMismatch_ReturnsFoundFalse verifica
// RN-TEN-01, mismo criterio que TestGetAppointmentDetail_TenantMismatch.
func TestListAppointmentHistory_TenantMismatch_ReturnsFoundFalse(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}

	_, _, found, err := repo.ListAppointmentHistory(context.Background(), string(shopR), created.Appointment.ID, nil, 20)
	if err != nil {
		t.Fatalf("ListAppointmentHistory(shopR): %v", err)
	}
	if found {
		t.Fatalf("found = true, want false: el historial de una cita de shopQ no debió verse desde shopR")
	}
}

// TestListAppointmentHistory_Pagination_TwoPagesWithoutOverlapOrDuplicate
// cubre CA-064-05: orden estable (occurred_at, id) y páginas consecutivas
// que ni omiten ni repiten filas.
func TestListAppointmentHistory_Pagination_TwoPagesWithoutOverlapOrDuplicate(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	appointmentID := created.Appointment.ID

	// appointment_created ya ocupa un instante propio (occurred_at de
	// CreateInternal); se agregan tres filas sintéticas más tarde para
	// tener 4 en total y forzar dos páginas de tamaño 2.
	base := start.Add(1 * time.Hour)
	insertSyntheticHistoryRow(t, db, shopQ, appointmentID, base, "aaaaaaaa-0000-0000-0000-000000000001")
	insertSyntheticHistoryRow(t, db, shopQ, appointmentID, base.Add(1*time.Minute), "aaaaaaaa-0000-0000-0000-000000000002")
	insertSyntheticHistoryRow(t, db, shopQ, appointmentID, base.Add(2*time.Minute), "aaaaaaaa-0000-0000-0000-000000000003")

	page1, next1, found, err := repo.ListAppointmentHistory(context.Background(), string(shopQ), appointmentID, nil, 2)
	if err != nil {
		t.Fatalf("ListAppointmentHistory(page1): %v", err)
	}
	if !found || len(page1) != 2 {
		t.Fatalf("page1 = %+v (found=%v), want 2 filas", page1, found)
	}
	if next1 == nil {
		t.Fatalf("next1 = nil, want un cursor (quedan más filas)")
	}

	page2, next2, found, err := repo.ListAppointmentHistory(context.Background(), string(shopQ), appointmentID, next1, 2)
	if err != nil {
		t.Fatalf("ListAppointmentHistory(page2): %v", err)
	}
	if !found || len(page2) != 2 {
		t.Fatalf("page2 = %+v (found=%v), want 2 filas", page2, found)
	}
	if next2 != nil {
		t.Fatalf("next2 = %+v, want nil (última página)", next2)
	}

	seen := make(map[string]bool, 4)
	for _, row := range append(append([]booking.HistoryRow{}, page1...), page2...) {
		if seen[row.ID] {
			t.Fatalf("id %q apareció en ambas páginas (duplicado)", row.ID)
		}
		seen[row.ID] = true
	}
	if len(seen) != 4 {
		t.Fatalf("total de filas únicas = %d, want 4", len(seen))
	}

	// Orden estable: cada fila de page1 debe preceder (occurred_at) a cada
	// fila de page2.
	for _, a := range page1 {
		for _, b := range page2 {
			if !a.OccurredAt.Before(b.OccurredAt) {
				t.Fatalf("orden incorrecto: %+v no precede a %+v", a, b)
			}
		}
	}
}

func TestListAppointmentHistory_ChangesAttachedToTheirOwnEntry(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	appointmentID := created.Appointment.ID
	historyID := "bbbbbbbb-0000-0000-0000-000000000001"
	insertSyntheticHistoryRow(t, db, shopQ, appointmentID, start.Add(1*time.Hour), historyID)

	err = db.InTenantTx(context.Background(), shopQ, func(ctx context.Context, q database.Queries) error {
		_, err := q.Exec(ctx, `
			INSERT INTO appointment_history_change (barbershop_id, history_id, field_name, previous_value, new_value)
			VALUES ($1, $2, 'status', 'confirmed', 'cancelled_by_barber')`,
			string(shopQ), historyID)
		return err
	})
	if err != nil {
		t.Fatalf("insertar appointment_history_change: %v", err)
	}

	items, _, _, err := repo.ListAppointmentHistory(context.Background(), string(shopQ), appointmentID, nil, 20)
	if err != nil {
		t.Fatalf("ListAppointmentHistory: %v", err)
	}
	var synthetic *booking.HistoryRow
	for i := range items {
		if items[i].ID == historyID {
			synthetic = &items[i]
		}
	}
	if synthetic == nil {
		t.Fatalf("no se encontró la fila sintética %q en %+v", historyID, items)
	}
	if len(synthetic.Changes) != 1 || synthetic.Changes[0].FieldName != "status" {
		t.Fatalf("Changes = %+v, want un cambio de status", synthetic.Changes)
	}
	if synthetic.Changes[0].PreviousValue == nil || *synthetic.Changes[0].PreviousValue != "confirmed" {
		t.Fatalf("PreviousValue = %v, want 'confirmed'", synthetic.Changes[0].PreviousValue)
	}

	// El evento appointment_created (sin fila en appointment_history_change)
	// no debe traer ningún Changes prestado del sintético.
	for _, row := range items {
		if row.ID != historyID && len(row.Changes) != 0 {
			t.Fatalf("la fila %q trae Changes = %+v, want ninguno", row.ID, row.Changes)
		}
	}
}

func TestCustomerNames_ReturnsMapForRequestedIDs(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}

	names, err := repo.CustomerNames(context.Background(), string(shopQ), []string{created.Customer.ID, "00000000-0000-0000-0000-000000000000"})
	if err != nil {
		t.Fatalf("CustomerNames: %v", err)
	}
	if got := names[created.Customer.ID]; got != "Cliente de prueba "+suffix {
		t.Fatalf("names[%q] = %q, want %q", created.Customer.ID, got, "Cliente de prueba "+suffix)
	}
	if _, ok := names["00000000-0000-0000-0000-000000000000"]; ok {
		t.Fatalf("un id inexistente no debió aparecer en el mapa")
	}
}

func TestCustomerNames_EmptyInput_ReturnsEmptyMapWithoutQuerying(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	names, err := repo.CustomerNames(context.Background(), string(shopQ), nil)
	if err != nil {
		t.Fatalf("CustomerNames: %v", err)
	}
	if len(names) != 0 {
		t.Fatalf("names = %+v, want vacío", names)
	}
}
