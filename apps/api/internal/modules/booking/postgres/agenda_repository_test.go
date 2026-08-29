// Package postgres_test (pruebas de integración, HU-062) requiere
// PostgreSQL REAL con las dieciséis migraciones aplicadas y
// database/testdata/dos_barberias.sql + database/testdata/hu060_citas.sql
// cargados: mismos fixtures que repository_test.go/manual_repository_test.go
// (HU-060/HU-061), reutilizados aquí porque ListDailyAgenda solo lee
// appointment, sin tabla nueva. Los turnos de estas pruebas se insertan con
// repo.CreateInternal, la misma primitiva probada en repository_test.go,
// nunca con SQL directo.
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/booking/postgres/...
package postgres_test

import (
	"context"
	"testing"
	"time"

	"system-barbershop/internal/modules/booking"
)

func containsID(entries []booking.DailyAgendaEntry, id string) bool {
	for _, e := range entries {
		if e.ID == id {
			return true
		}
	}
	return false
}

func TestListDailyAgenda_ReturnsOnlyIntersectingAppointmentsOrderedByStartsAt(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	day := baseStart(t).Truncate(24 * time.Hour)
	rangeStart := day
	rangeEnd := day.Add(24 * time.Hour)

	// before: termina exactamente en rangeStart -> NO interseca (semiabierto [)).
	before := validInput(t, barberQ1, serviceQ, rangeStart.Add(-30*time.Minute), 30, suffix+"-before")
	// firstOfDay: empieza justo en rangeStart -> SÍ interseca (extremo incluido).
	firstOfDay := validInput(t, barberQ1, serviceQ, rangeStart, 30, suffix+"-first")
	// midDay: claramente dentro del rango.
	midDay := validInput(t, barberQ1, serviceQ, rangeStart.Add(4*time.Hour), 45, suffix+"-mid")
	// atBoundary: empieza exactamente en rangeEnd -> NO interseca (extremo excluido).
	atBoundary := validInput(t, barberQ1, serviceQ, rangeEnd, 30, suffix+"-boundary")
	// otherBarber: mismo shop, mismo rango, pero otro barbero -> NO debe
	// aparecer en la agenda de barberQ1.
	otherBarber := validInput(t, barberQ2, serviceQ, rangeStart.Add(2*time.Hour), 30, suffix+"-other-barber")

	beforeResult, err := repo.CreateInternal(context.Background(), string(shopQ), before)
	if err != nil {
		t.Fatalf("CreateInternal(before): %v", err)
	}
	firstResult, err := repo.CreateInternal(context.Background(), string(shopQ), firstOfDay)
	if err != nil {
		t.Fatalf("CreateInternal(firstOfDay): %v", err)
	}
	midResult, err := repo.CreateInternal(context.Background(), string(shopQ), midDay)
	if err != nil {
		t.Fatalf("CreateInternal(midDay): %v", err)
	}
	if _, err := repo.CreateInternal(context.Background(), string(shopQ), atBoundary); err != nil {
		t.Fatalf("CreateInternal(atBoundary): %v", err)
	}
	otherBarberResult, err := repo.CreateInternal(context.Background(), string(shopQ), otherBarber)
	if err != nil {
		t.Fatalf("CreateInternal(otherBarber): %v", err)
	}

	entries, err := repo.ListDailyAgenda(context.Background(), string(shopQ), barberQ1, rangeStart, rangeEnd)
	if err != nil {
		t.Fatalf("ListDailyAgenda: %v", err)
	}

	idxFirst, idxMid := -1, -1
	for i, e := range entries {
		if e.ID == firstResult.Appointment.ID {
			idxFirst = i
		}
		if e.ID == midResult.Appointment.ID {
			idxMid = i
		}
	}
	if idxFirst == -1 || idxMid == -1 {
		t.Fatalf("faltan turnos esperados en la respuesta: %+v", entries)
	}
	if idxFirst > idxMid {
		t.Fatalf("orden incorrecto: firstOfDay (idx %d) debe preceder a midDay (idx %d)", idxFirst, idxMid)
	}
	if containsID(entries, otherBarberResult.Appointment.ID) {
		t.Fatalf("un turno de otro barbero (barberQ2) no debió aparecer en la agenda de barberQ1")
	}
	if containsID(entries, beforeResult.Appointment.ID) {
		t.Fatalf("el turno 'before' (termina exactamente en rangeStart) no debió intersecar el rango")
	}
}

// TestListDailyAgenda_NocturnalAppointment_VisibleInBothIntersectingDays
// verifica DEC-075: un turno que cruza medianoche aparece en la agenda del
// día en que empieza Y en la del día siguiente, nunca solo en uno.
func TestListDailyAgenda_NocturnalAppointment_VisibleInBothIntersectingDays(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	day1Start := baseStart(t).Truncate(24 * time.Hour)
	day1End := day1Start.Add(24 * time.Hour)
	day2End := day1End.Add(24 * time.Hour)

	nocturnal := validInput(t, barberQ1, serviceQ, day1Start.Add(23*time.Hour+30*time.Minute), 45, suffix+"-nocturnal")
	result, err := repo.CreateInternal(context.Background(), string(shopQ), nocturnal)
	if err != nil {
		t.Fatalf("CreateInternal(nocturnal): %v", err)
	}

	day1Entries, err := repo.ListDailyAgenda(context.Background(), string(shopQ), barberQ1, day1Start, day1End)
	if err != nil {
		t.Fatalf("ListDailyAgenda(day1): %v", err)
	}
	if !containsID(day1Entries, result.Appointment.ID) {
		t.Fatalf("el turno nocturno debió aparecer en la agenda del día en que empieza")
	}

	day2Entries, err := repo.ListDailyAgenda(context.Background(), string(shopQ), barberQ1, day1End, day2End)
	if err != nil {
		t.Fatalf("ListDailyAgenda(day2): %v", err)
	}
	if !containsID(day2Entries, result.Appointment.ID) {
		t.Fatalf("el turno nocturno debió aparecer también en la agenda del día siguiente (DEC-075)")
	}
}

// TestListDailyAgenda_TenantMismatch_ReturnsNoRows verifica RN-TEN-01: un
// turno real de shopQ/barberQ1 nunca aparece al consultar con el
// barbershopID de shopR, aunque el rango de fechas y el barberID coincidan
// literalmente (RLS + el filtro explícito por barbershop_id de la consulta).
func TestListDailyAgenda_TenantMismatch_ReturnsNoRows(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	day := baseStart(t).Truncate(24 * time.Hour)
	rangeStart := day
	rangeEnd := day.Add(24 * time.Hour)

	input := validInput(t, barberQ1, serviceQ, rangeStart.Add(3*time.Hour), 30, suffix+"-tenant")
	result, err := repo.CreateInternal(context.Background(), string(shopQ), input)
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}

	entries, err := repo.ListDailyAgenda(context.Background(), string(shopR), barberQ1, rangeStart, rangeEnd)
	if err != nil {
		t.Fatalf("ListDailyAgenda(shopR): %v", err)
	}
	if containsID(entries, result.Appointment.ID) {
		t.Fatalf("una cita de shopQ no debió aparecer al consultar shopR (RN-TEN-01)")
	}
}
