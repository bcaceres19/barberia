// Package postgres_test (pruebas de integración, HU-094) requiere
// PostgreSQL REAL con todas las migraciones aplicadas y
// database/testdata/hu094_disponibilidad.sql cargado. Conéctate como
// barberia_app (docs/03-desarrollo/estrategia-pruebas.md §2 prohíbe mocks
// para RLS/resolución de tenant sin contexto).
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/publicbooking/postgres/...
package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	publicbookingpostgres "system-barbershop/internal/modules/publicbooking/postgres"
	"system-barbershop/internal/platform/database"
)

const (
	slugAvailUno = "barberia-hu094-uno"
	slugAvailDos = "barberia-hu094-dos"

	barbershopAvailUnoID = "00940001-0094-0094-0094-009400010001"
	barberAvailUnoID     = "00940011-0094-0094-0094-009400110011"
	barbershopAvailDosID = "00940002-0094-0094-0094-009400020002"
	barberAvailDosID     = "00940012-0094-0094-0094-009400120012"
	serviceAvailUnoID    = "00940101-0094-0094-0094-009401010101"
)

func TestResolveBarbershopID_KnownSlug_ReturnsInternalID(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)

	id, found, err := repo.ResolveBarbershopID(context.Background(), slugAvailUno)
	if err != nil {
		t.Fatalf("ResolveBarbershopID: %v", err)
	}
	if !found {
		t.Fatal("found = false, want true")
	}
	if id != barbershopAvailUnoID {
		t.Fatalf("id = %q, want %q", id, barbershopAvailUnoID)
	}
}

func TestResolveBarbershopID_UnknownSlug_ReturnsNotFoundWithoutError(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)

	id, found, err := repo.ResolveBarbershopID(context.Background(), "no-existe-nunca")
	if err != nil {
		t.Fatalf("ResolveBarbershopID: %v", err)
	}
	if found || id != "" {
		t.Fatalf("found = %v, id = %q, want false/\"\"", found, id)
	}
}

// createAppointmentForOccupancy inserta una cita directamente (sin pasar
// por internal/modules/booking, que no es el objeto bajo prueba aquí) para
// que ListOccupiedIntervals tenga algo real que leer, dentro de una
// transacción tenant-aware con el rol real.
// occupancyPhoneSeed ancla la unicidad del teléfono sintético de cada
// customer creado por esta suite en el instante de arranque del proceso de
// prueba: `customer`/`appointment` nunca admiten DELETE (RN-DAT-03), así
// que corridas repetidas contra el mismo PostgreSQL de desarrollo no deben
// competir por el mismo teléfono dentro de la misma barbería.
var occupancyPhoneSeed = time.Now().UnixNano()
var occupancyCustomerCounter int64

func createAppointmentForOccupancy(t *testing.T, db *database.DB, barbershopID, barberID, serviceID, status string, startsAt, endsAt time.Time) string {
	t.Helper()
	occupancyCustomerCounter++
	phone := fmt.Sprintf("+57%09d", (occupancyPhoneSeed+occupancyCustomerCounter)%1000000000)

	var appointmentID string
	err := db.InTenantTx(context.Background(), database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var customerID string
		if err := q.QueryRow(ctx, `
			INSERT INTO customer (barbershop_id, full_name, phone)
			VALUES ($1, 'Cliente de prueba HU-094', $2)
			RETURNING id`,
			barbershopID, phone,
		).Scan(&customerID); err != nil {
			return err
		}
		return q.QueryRow(ctx, `
			INSERT INTO appointment (
				barbershop_id, barber_id, service_id, customer_id, attendee_name,
				starts_at, ends_at, status, origin, resolved_at,
				service_name_snapshot, duration_minutes_snapshot,
				price_amount_snapshot, price_currency_snapshot
			) VALUES (
				$1, $2, $3, $4, 'Cliente de prueba HU-094',
				$5, $6, $7, 'manual', CASE WHEN $7 <> 'confirmed' THEN now() END,
				'Servicio', 30, 35000.00, 'COP'
			)
			RETURNING id`,
			barbershopID, barberID, serviceID, customerID, startsAt, endsAt, status,
		).Scan(&appointmentID)
	})
	if err != nil {
		t.Fatalf("crear appointment: %v", err)
	}
	return appointmentID
}

// randomFutureBase ancla cada corrida de prueba en un punto distinto y muy
// lejano en el futuro (`appointment`/`customer` nunca admiten DELETE,
// RN-DAT-03: corridas repetidas contra el mismo Postgres de desarrollo
// acumulan filas, así que dos corridas nunca deben competir por el mismo
// [starts_at, ends_at) del mismo barbero).
func randomFutureBase() time.Time {
	offsetHours := time.Duration(time.Now().UnixNano()%100000) * time.Hour
	return time.Now().UTC().Truncate(time.Minute).Add(1000*24*time.Hour + offsetHours)
}

func TestListOccupiedIntervals_OnlyOccupyingStatusesAndTenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)

	base := randomFutureBase()
	confirmedStart := base
	confirmedEnd := confirmedStart.Add(30 * time.Minute)
	cancelledStart := base.Add(2 * time.Hour)
	cancelledEnd := cancelledStart.Add(30 * time.Minute)
	otherShopStart := base.Add(4 * time.Hour)
	otherShopEnd := otherShopStart.Add(30 * time.Minute)

	createAppointmentForOccupancy(t, db, barbershopAvailUnoID, barberAvailUnoID, serviceAvailUnoID, "confirmed", confirmedStart, confirmedEnd)
	createAppointmentForOccupancy(t, db, barbershopAvailUnoID, barberAvailUnoID, serviceAvailUnoID, "cancelled_by_barber", cancelledStart, cancelledEnd)
	createAppointmentForOccupancy(t, db, barbershopAvailDosID, barberAvailDosID, "00940201-0094-0094-0094-009402010201", "confirmed", otherShopStart, otherShopEnd)

	from := base
	to := base.Add(10 * time.Hour)
	starts, ends, err := repo.ListOccupiedIntervals(context.Background(), barbershopAvailUnoID, barberAvailUnoID, from, to)
	if err != nil {
		t.Fatalf("ListOccupiedIntervals: %v", err)
	}
	if len(starts) != 1 || len(ends) != 1 {
		t.Fatalf("esperaba exactamente 1 intervalo (solo confirmed, tenant Uno), got starts=%v ends=%v", starts, ends)
	}
	if !starts[0].Equal(confirmedStart) || !ends[0].Equal(confirmedEnd) {
		t.Fatalf("intervalo = [%v, %v), want [%v, %v)", starts[0], ends[0], confirmedStart, confirmedEnd)
	}
}

func TestListOccupiedIntervals_RangeOutsideWindow_ReturnsEmpty(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)

	base := randomFutureBase()
	farStart := base.Add(48 * time.Hour)
	farEnd := farStart.Add(30 * time.Minute)
	createAppointmentForOccupancy(t, db, barbershopAvailUnoID, barberAvailUnoID, serviceAvailUnoID, "confirmed", farStart, farEnd)

	starts, ends, err := repo.ListOccupiedIntervals(context.Background(), barbershopAvailUnoID, barberAvailUnoID, base, base.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("ListOccupiedIntervals: %v", err)
	}
	if len(starts) != 0 || len(ends) != 0 {
		t.Fatalf("esperaba vacío fuera de rango, got starts=%v ends=%v", starts, ends)
	}
}
