// Package postgres_test (pruebas de integración, HU-096) requiere
// PostgreSQL REAL con todas las migraciones aplicadas y
// database/testdata/dos_barberias.sql cargado (shop A/B: FindCustomerMatches
// solo lee `customer` por barbershopID+phone/barbershopID+email, ninguna
// columna mutable de `barbershop`, así que reutilizar A/B es seguro incluso
// con paquetes concurrentes -docs/03-desarrollo/estrategia-pruebas.md §2).
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/publicbooking/postgres/...
package postgres_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"testing"

	publicbookingpostgres "system-barbershop/internal/modules/publicbooking/postgres"
	"system-barbershop/internal/platform/database"
)

const (
	shopIdentityA = "11111111-1111-1111-1111-111111111111"
	shopIdentityB = "22222222-2222-2222-2222-222222222222"
)

// identityUniqueDigits genera dígitos aleatorios (nunca letras), mismo
// criterio que booking/postgres.uniqueDigits: un teléfono E.164 solo admite
// dígitos tras el '+'.
func identityUniqueDigits(t *testing.T, n int) string {
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

// insertTestCustomer inserta directamente una fila `customer` dentro de
// barbershopID, para tener una identidad ya existente contra la que
// reconciliar (HU-096 no expone ninguna operación de escritura propia: la
// persistencia real de customer vive en booking, HU-097 la reutilizará).
func insertTestCustomer(t *testing.T, db *database.DB, barbershopID, fullName, phone, email string) string {
	t.Helper()
	var id string
	err := db.InTenantTx(context.Background(), database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx,
			`INSERT INTO customer (barbershop_id, full_name, phone, email)
			 VALUES ($1, $2, $3, $4)
			 RETURNING id`,
			barbershopID, fullName, phone, email,
		).Scan(&id)
	})
	if err != nil {
		t.Fatalf("insertTestCustomer: %v", err)
	}
	return id
}

func TestFindCustomerMatches_PhoneMatch_ReturnsOnlyPhoneMatchID(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)
	suffix := identityUniqueDigits(t, 6)
	phone := "+57300" + suffix
	email := fmt.Sprintf("cliente.%s@ejemplo.test", suffix)
	existingID := insertTestCustomer(t, db, shopIdentityA, "Cliente HU-096 "+suffix, phone, email)

	phoneMatch, emailMatch, err := repo.FindCustomerMatches(context.Background(), shopIdentityA, phone, "otro."+email)
	if err != nil {
		t.Fatalf("FindCustomerMatches: %v", err)
	}
	if phoneMatch == nil || *phoneMatch != existingID {
		t.Fatalf("phoneMatch = %v, want %q", phoneMatch, existingID)
	}
	if emailMatch != nil {
		t.Fatalf("emailMatch = %v, want nil", *emailMatch)
	}
}

func TestFindCustomerMatches_EmailMatch_ReturnsOnlyEmailMatchID(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)
	suffix := identityUniqueDigits(t, 6)
	phone := "+57301" + suffix
	email := fmt.Sprintf("cliente.%s@ejemplo.test", suffix)
	existingID := insertTestCustomer(t, db, shopIdentityA, "Cliente HU-096 "+suffix, phone, email)

	phoneMatch, emailMatch, err := repo.FindCustomerMatches(context.Background(), shopIdentityA, "+57302"+suffix, email)
	if err != nil {
		t.Fatalf("FindCustomerMatches: %v", err)
	}
	if emailMatch == nil || *emailMatch != existingID {
		t.Fatalf("emailMatch = %v, want %q", emailMatch, existingID)
	}
	if phoneMatch != nil {
		t.Fatalf("phoneMatch = %v, want nil", *phoneMatch)
	}
}

func TestFindCustomerMatches_ConflictBetweenDifferentRows_ReturnsBothDistinctIDs(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)
	suffix := identityUniqueDigits(t, 6)
	phone := "+57303" + suffix
	email := fmt.Sprintf("email.%s@ejemplo.test", suffix)
	otherPhone := "+57304" + suffix
	otherEmail := fmt.Sprintf("phone.%s@ejemplo.test", suffix)

	phoneOwnerID := insertTestCustomer(t, db, shopIdentityA, "Dueño de teléfono "+suffix, phone, otherEmail)
	emailOwnerID := insertTestCustomer(t, db, shopIdentityA, "Dueño de correo "+suffix, otherPhone, email)

	phoneMatch, emailMatch, err := repo.FindCustomerMatches(context.Background(), shopIdentityA, phone, email)
	if err != nil {
		t.Fatalf("FindCustomerMatches: %v", err)
	}
	if phoneMatch == nil || *phoneMatch != phoneOwnerID {
		t.Fatalf("phoneMatch = %v, want %q", phoneMatch, phoneOwnerID)
	}
	if emailMatch == nil || *emailMatch != emailOwnerID {
		t.Fatalf("emailMatch = %v, want %q", emailMatch, emailOwnerID)
	}
	if *phoneMatch == *emailMatch {
		t.Fatalf("expected distinct ids, both were %q", *phoneMatch)
	}
}

func TestFindCustomerMatches_NoMatch_ReturnsBothNil(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)
	suffix := identityUniqueDigits(t, 6)

	phoneMatch, emailMatch, err := repo.FindCustomerMatches(context.Background(), shopIdentityA, "+57305"+suffix, fmt.Sprintf("nadie.%s@ejemplo.test", suffix))
	if err != nil {
		t.Fatalf("FindCustomerMatches: %v", err)
	}
	if phoneMatch != nil || emailMatch != nil {
		t.Fatalf("phoneMatch = %v, emailMatch = %v, want both nil", phoneMatch, emailMatch)
	}
}

// TestFindCustomerMatches_NeverCrossesTenant prueba RN-TEN-01: un customer
// de shop B con el mismo teléfono/correo que se busca en shop A nunca
// aparece como coincidencia (aislamiento fuerte, DEC-024).
func TestFindCustomerMatches_NeverCrossesTenant(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)
	suffix := identityUniqueDigits(t, 6)
	phone := "+57306" + suffix
	email := fmt.Sprintf("cliente.%s@ejemplo.test", suffix)
	insertTestCustomer(t, db, shopIdentityB, "Cliente de otra barbería "+suffix, phone, email)

	phoneMatch, emailMatch, err := repo.FindCustomerMatches(context.Background(), shopIdentityA, phone, email)
	if err != nil {
		t.Fatalf("FindCustomerMatches: %v", err)
	}
	if phoneMatch != nil || emailMatch != nil {
		t.Fatalf("phoneMatch = %v, emailMatch = %v, want both nil (cross-tenant match)", phoneMatch, emailMatch)
	}
}
