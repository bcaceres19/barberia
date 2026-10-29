package postgres_test

import (
	"context"
	"testing"
	"time"

	"system-barbershop/internal/modules/booking"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

// Servicio a domicilio (DEC-131, HU-103) contra PostgreSQL real y dos tenants:
// la dirección se persiste junto al turno, el panel de esa barbería la lee en la
// agenda y en el detalle, y nunca es visible desde otra barbería.

func publicWithHome(t *testing.T, home *booking.HomeService) (booking.CreatePublicInput, string) {
	t.Helper()
	suffix := uniqueSuffix(t)
	input := publicInput(t, barberQ1, serviceQ, baseStart(t), suffix, booking.CustomerInput{
		New: &booking.NewCustomerInput{FullName: "Cliente domicilio " + suffix},
	})
	input.HomeService = home
	return input, suffix
}

func appointmentIDAt(t *testing.T, db *database.DB, shop database.BarbershopID, barberID string, start time.Time) string {
	t.Helper()
	var id string
	err := db.InTenantTx(context.Background(), shop, func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx,
			`SELECT id FROM appointment WHERE barbershop_id = $1 AND barber_id = $2 AND starts_at = $3`,
			string(shop), barberID, start).Scan(&id)
	})
	if err != nil {
		t.Fatalf("find appointment: %v", err)
	}
	return id
}

func TestCreatePublic_HomeService_PersistsAddressAndIsReadableByItsShopOnly(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	home := &booking.HomeService{Address: "Carrera 7 # 71-21, Bogotá", Latitude: 4.6486, Longitude: -74.0628}
	input, suffix := publicWithHome(t, home)

	if _, err := repo.CreatePublic(context.Background(), string(shopQ), input,
		idempotency.Key("home-"+suffix), idempotency.Fingerprint("8123456789abcdef0123456789abcdef")); err != nil {
		t.Fatalf("CreatePublic: %v", err)
	}
	id := appointmentIDAt(t, db, shopQ, barberQ1, input.StartsAt)

	// Detalle del panel de la barbería dueña.
	detail, found, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), id)
	if err != nil || !found {
		t.Fatalf("GetAppointmentDetail: found=%v err=%v", found, err)
	}
	if detail.HomeService == nil || *detail.HomeService != *home {
		t.Fatalf("detail.HomeService = %+v, want %+v", detail.HomeService, home)
	}

	// Agenda del barbero ese día.
	entries, err := repo.ListDailyAgenda(context.Background(), string(shopQ), barberQ1,
		input.StartsAt.Add(-time.Hour), input.StartsAt.Add(time.Hour))
	if err != nil {
		t.Fatalf("ListDailyAgenda: %v", err)
	}
	var seen bool
	for _, e := range entries {
		if e.ID == id {
			seen = true
			if e.HomeService == nil || *e.HomeService != *home {
				t.Fatalf("agenda entry HomeService = %+v, want %+v", e.HomeService, home)
			}
		}
	}
	if !seen {
		t.Fatal("the home appointment is missing from the agenda")
	}

	// RN-TEN-01: la otra barbería ni encuentra el turno ni su dirección.
	_, foundOther, err := repo.GetAppointmentDetail(context.Background(), string(shopR), id)
	if err != nil {
		t.Fatalf("GetAppointmentDetail from the other shop: %v", err)
	}
	if foundOther {
		t.Fatal("RN-TEN-01: shopR must not read shopQ's home appointment")
	}
	err = db.InTenantTx(context.Background(), shopR, func(ctx context.Context, q database.Queries) error {
		var n int
		if err := q.QueryRow(ctx, `SELECT count(*) FROM appointment WHERE service_address = $1`, home.Address).Scan(&n); err != nil {
			return err
		}
		if n != 0 {
			t.Fatalf("RN-TEN-01: shopR sees %d rows with shopQ's address", n)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("cross-tenant read: %v", err)
	}
}

func TestCreatePublic_WithoutHomeService_StoresNoAddress(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	input, suffix := publicWithHome(t, nil)

	if _, err := repo.CreatePublic(context.Background(), string(shopQ), input,
		idempotency.Key("local-"+suffix), idempotency.Fingerprint("9123456789abcdef0123456789abcdef")); err != nil {
		t.Fatalf("CreatePublic: %v", err)
	}
	id := appointmentIDAt(t, db, shopQ, barberQ1, input.StartsAt)

	detail, _, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), id)
	if err != nil {
		t.Fatalf("GetAppointmentDetail: %v", err)
	}
	if detail.HomeService != nil {
		t.Fatalf("a turn in the shop must carry no address, got %+v", detail.HomeService)
	}
	err = db.InTenantTx(context.Background(), shopQ, func(ctx context.Context, q database.Queries) error {
		var isHome bool
		var address *string
		var lat, lng *float64
		if err := q.QueryRow(ctx,
			`SELECT is_home_service, service_address, service_latitude, service_longitude FROM appointment WHERE id = $1`, id,
		).Scan(&isHome, &address, &lat, &lng); err != nil {
			return err
		}
		if isHome || address != nil || lat != nil || lng != nil {
			t.Fatalf("non-home turn arrastra datos de dirección: %v %v %v %v", isHome, address, lat, lng)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("read columns: %v", err)
	}
}

// La base impide un turno parcial aunque la aplicación se equivoque (CA-103-04).
func TestAppointmentHomeServiceCheck_RejectsPartialAddressAtTheDatabase(t *testing.T) {
	db := setupTestDB(t)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	err := db.InTenantTx(context.Background(), shopQ, func(ctx context.Context, q database.Queries) error {
		_, execErr := q.Exec(ctx, `
			INSERT INTO appointment (
				barbershop_id, barber_id, service_id, customer_id, attendee_name,
				starts_at, ends_at, status, origin,
				service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot,
				is_home_service, service_address)
			SELECT $1, $2, $3, c.id, $4, $5, $6, 'confirmed', 'public', 'Corte', 30, 18000.00, 'COP', true, 'Solo texto'
			  FROM customer c WHERE c.barbershop_id = $1 LIMIT 1`,
			string(shopQ), barberQ1, serviceQ, "Parcial "+suffix, start, start.Add(30*time.Minute))
		return execErr
	})
	if err == nil {
		t.Fatal("appointment_home_service_ck must reject a home turn without coordinates")
	}
}
