// Package postgres_test (pruebas de integración, HU-061) requiere
// PostgreSQL REAL con las dieciséis migraciones aplicadas y
// database/testdata/dos_barberias.sql + database/testdata/hu060_citas.sql
// cargados: mismos fixtures que repository_test.go (HU-060), reutilizados
// aquí porque CreateManual/FindCustomerForReconciliation no dependen de
// ninguna tabla nueva.
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/booking/postgres/...
package postgres_test

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"testing"
	"time"

	"system-barbershop/internal/modules/booking"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

// uniqueDigits genera dígitos aleatorios (nunca letras hexadecimales, a
// diferencia de uniqueSuffix): un teléfono E.164 solo admite dígitos tras
// el '+' (customer_phone_ck).
func uniqueDigits(t *testing.T, n int) string {
	t.Helper()
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	digits := make([]byte, n)
	for i, b := range buf {
		digits[i] = '0' + b%10
	}
	return string(digits)
}

func manualInput(t *testing.T, barberID, serviceID string, suffix string, customer booking.CustomerInput) booking.CreateInternalInput {
	start := baseStart(t)
	return booking.CreateInternalInput{
		BarberID:     barberID,
		ServiceID:    serviceID,
		AttendeeName: "Persona atendida " + suffix,
		StartsAt:     start,
		EndsAt:       start.Add(30 * time.Minute),
		Origin:       booking.OriginManual,
		Service: booking.ServiceSnapshot{
			Name:             "Corte manual " + suffix,
			DurationMinutes:  30,
			PriceAmountCents: 1500000,
			Currency:         "COP",
		},
		Customer: customer,
		Actor:    booking.Actor{Type: booking.ActorTypeStaff, StaffUserID: strPtr(staffQ)},
	}
}

func TestCreateManual_NewCustomerNoContact_Proceeds_PersistsAppointment(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	input := manualInput(t, barberQ1, serviceQ, suffix, booking.CustomerInput{
		New: &booking.NewCustomerInput{FullName: "Cliente manual " + suffix},
	})
	key := idempotency.Key("manual-" + suffix)
	fingerprint := idempotency.Fingerprint("0123456789abcdef0123456789abcdef")

	result, err := repo.CreateManual(context.Background(), string(shopQ), input, key, fingerprint)
	if err != nil {
		t.Fatalf("CreateManual: %v", err)
	}
	if result.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("Outcome = %v, want OutcomeProceed", result.Decision.Outcome)
	}
	if result.Appointment.Status != booking.StatusConfirmed {
		t.Fatalf("Status = %v, want confirmed", result.Appointment.Status)
	}
	if result.Appointment.Origin != booking.OriginManual {
		t.Fatalf("Origin = %v, want manual", result.Appointment.Origin)
	}
	if result.Response.Status != 201 {
		t.Fatalf("Response.Status = %d, want 201", result.Response.Status)
	}
	var parsed struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(result.Response.Body), &parsed); err != nil {
		t.Fatalf("unmarshal stored response: %v", err)
	}
	if parsed.ID != result.Appointment.ID {
		t.Fatalf("Response.Body id = %q, want %q", parsed.ID, result.Appointment.ID)
	}
}

func TestCreateManual_RepeatedKeySameBody_Replays_NeverCreatesSecondAppointment(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	input := manualInput(t, barberQ1, serviceQ, suffix, booking.CustomerInput{
		New: &booking.NewCustomerInput{FullName: "Cliente repetido " + suffix},
	})
	key := idempotency.Key("manual-replay-" + suffix)
	fingerprint := idempotency.Fingerprint("1123456789abcdef0123456789abcdef")

	first, err := repo.CreateManual(context.Background(), string(shopQ), input, key, fingerprint)
	if err != nil {
		t.Fatalf("primer CreateManual: %v", err)
	}
	second, err := repo.CreateManual(context.Background(), string(shopQ), input, key, fingerprint)
	if err != nil {
		t.Fatalf("segundo CreateManual: %v", err)
	}
	if second.Decision.Outcome != idempotency.OutcomeReplay {
		t.Fatalf("Outcome = %v, want OutcomeReplay", second.Decision.Outcome)
	}
	if second.Response.Body != first.Response.Body {
		t.Fatalf("Response.Body de la repetición no coincide byte a byte con el original")
	}

	err = db.InTenantTx(context.Background(), shopQ, func(ctx context.Context, q database.Queries) error {
		var count int
		scanErr := q.QueryRow(ctx,
			`SELECT count(*) FROM appointment WHERE barbershop_id = $1 AND attendee_name = $2`,
			string(shopQ), input.AttendeeName,
		).Scan(&count)
		if scanErr != nil {
			return scanErr
		}
		if count != 1 {
			t.Fatalf("expected exactly one appointment row after replay, got %d", count)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("verify single row: %v", err)
	}
}

func TestFindCustomerForReconciliation_ByPhone_FindsExistingCustomer(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	phone := "+573" + uniqueDigits(t, 7)

	seed := manualInput(t, barberQ1, serviceQ, suffix, booking.CustomerInput{
		New: &booking.NewCustomerInput{FullName: "Cliente con teléfono " + suffix, Phone: &phone},
	})
	created, err := repo.CreateInternal(context.Background(), string(shopQ), seed)
	if err != nil {
		t.Fatalf("seed CreateInternal: %v", err)
	}

	found, ok, err := repo.FindCustomerForReconciliation(context.Background(), string(shopQ), &phone, nil)
	if err != nil {
		t.Fatalf("FindCustomerForReconciliation: %v", err)
	}
	if !ok {
		t.Fatal("found = false, want true (DEC-045)")
	}
	if found.ID != created.Customer.ID {
		t.Fatalf("found.ID = %q, want %q", found.ID, created.Customer.ID)
	}
}

func TestFindCustomerForReconciliation_ByEmail_FindsExistingCustomer(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	email := "cliente." + suffix + "@example.com"

	seed := manualInput(t, barberQ1, serviceQ, suffix, booking.CustomerInput{
		New: &booking.NewCustomerInput{FullName: "Cliente con correo " + suffix, Email: &email},
	})
	created, err := repo.CreateInternal(context.Background(), string(shopQ), seed)
	if err != nil {
		t.Fatalf("seed CreateInternal: %v", err)
	}

	found, ok, err := repo.FindCustomerForReconciliation(context.Background(), string(shopQ), nil, &email)
	if err != nil {
		t.Fatalf("FindCustomerForReconciliation: %v", err)
	}
	if !ok {
		t.Fatal("found = false, want true (DEC-046/DEC-071)")
	}
	if found.ID != created.Customer.ID {
		t.Fatalf("found.ID = %q, want %q", found.ID, created.Customer.ID)
	}
}

func TestFindCustomerForReconciliation_NoMatch_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	phone := "+573" + uniqueDigits(t, 7)

	_, ok, err := repo.FindCustomerForReconciliation(context.Background(), string(shopQ), &phone, nil)
	if err != nil {
		t.Fatalf("FindCustomerForReconciliation: %v", err)
	}
	if ok {
		t.Fatal("found = true, want false: ningún cliente tiene este teléfono todavía")
	}
}

func TestFindCustomerForReconciliation_NoPhoneNoEmail_ReturnsNotFoundWithoutQuerying(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)

	_, ok, err := repo.FindCustomerForReconciliation(context.Background(), string(shopQ), nil, nil)
	if err != nil {
		t.Fatalf("FindCustomerForReconciliation: %v", err)
	}
	if ok {
		t.Fatal("found = true, want false: sin teléfono ni correo nunca reconcilia (DEC-071)")
	}
}

func TestCreateManual_ExistingCustomerFromOtherTenant_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)

	other := manualInput(t, barberR1, serviceR, "other"+suffix, booking.CustomerInput{
		New: &booking.NewCustomerInput{FullName: "Cliente de otra barbería " + suffix},
	})
	other.Actor = booking.Actor{Type: booking.ActorTypeStaff, StaffUserID: strPtr(staffR)}
	otherCreated, err := repo.CreateInternal(context.Background(), string(shopR), other)
	if err != nil {
		t.Fatalf("seed en shopR: %v", err)
	}

	input := manualInput(t, barberQ1, serviceQ, suffix, booking.CustomerInput{
		ExistingID: &otherCreated.Customer.ID,
	})
	_, err = repo.CreateManual(context.Background(), string(shopQ), input, idempotency.Key("manual-cross-"+suffix), idempotency.Fingerprint("2123456789abcdef0123456789abcdef"))
	if err == nil {
		t.Fatal("expected an error linking a customer from another tenant")
	}
}
