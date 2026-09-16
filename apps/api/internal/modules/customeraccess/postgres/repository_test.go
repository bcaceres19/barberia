// Package postgres_test (pruebas de integración, HU-098) requiere
// PostgreSQL REAL con todas las migraciones aplicadas (incluida
// 20260916120000_add_public_resolve_appointment_token_tenant.sql) y
// database/testdata/{hu098_acceso_turno.sql} cargado. Cada prueba crea su
// propia cita con booking/postgres.Repository.CreateInternal (la misma
// primitiva ya probada en booking/postgres/repository_test.go) sobre las
// barberías dedicadas del fixture, y agrega su propia fila
// appointment_access_token con SQL directo: esa tabla es responsabilidad
// de customeraccess/publicbooking, no de booking (docs/03-desarrollo/
// estrategia-pruebas.md §2 prohíbe mocks para RLS/resolución de tenant sin
// contexto).
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/customeraccess/postgres/...
package postgres_test

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"os"
	"testing"
	"time"

	"system-barbershop/internal/modules/booking"
	bookingpostgres "system-barbershop/internal/modules/booking/postgres"
	customeraccesspostgres "system-barbershop/internal/modules/customeraccess/postgres"
	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

const (
	testDatabaseURL = "postgres://barberia_app@localhost:5432/barberia_test?sslmode=disable"

	// Par dedicado de hu098_acceso_turno.sql: shopUno exige razón NUNCA
	// (client_allowed=true, reason_required=false, deadline=30) y shopDos
	// prohíbe la cancelación tardía del cliente por completo
	// (client_allowed=false, reason_required=false, deadline=90) -valores
	// deliberadamente distintos para que una fuga cruzada de política sea
	// detectable (CA-098-05).
	shopUno    = "00980001-0098-0098-0098-009800010001"
	barberUno  = "00980011-0098-0098-0098-009800110011"
	serviceUno = "00980012-0098-0098-0098-009800120012"

	shopDos    = "00980002-0098-0098-0098-009800020002"
	barberDos  = "00980021-0098-0098-0098-009800210021"
	serviceDos = "00980022-0098-0098-0098-009800220022"
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
	t.Cleanup(func() { db.Close() })
	return db
}

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

// createTestAppointment crea una cita real (booking.CreateInternal) dentro
// de shopID, con un cliente nuevo propio de la prueba, y devuelve la cita
// creada.
func createTestAppointment(t *testing.T, db *database.DB, shopID, barberID, serviceID string, status booking.Status) booking.Appointment {
	t.Helper()
	repo := bookingpostgres.New(db, idempotency.NewSQLCoordinator())

	fullName := "Cliente de prueba HU-098 " + uniqueDigits(t, 6)
	// Desplazamiento aleatorio de días/minutos para que llamadas repetidas
	// (incluso dentro de la misma prueba, mismo barbero) nunca choquen con
	// appointment_barber_interval_excl ni con una corrida anterior contra el
	// mismo PostgreSQL persistente (barberia_app no puede DELETE
	// appointment, RN-HIS-02, mismo motivo que uniqueSuffix en
	// booking/postgres/repository_test.go).
	offsetDays, err := rand.Int(rand.Reader, big.NewInt(3650))
	if err != nil {
		t.Fatalf("rand.Int: %v", err)
	}
	offsetMinutes, err := rand.Int(rand.Reader, big.NewInt(20))
	if err != nil {
		t.Fatalf("rand.Int: %v", err)
	}
	starts := time.Date(2027, 3, 15, 8, 0, 0, 0, time.UTC).
		AddDate(0, 0, int(offsetDays.Int64())).
		Add(time.Duration(offsetMinutes.Int64()) * time.Hour)

	created, err := repo.CreateInternal(context.Background(), shopID, booking.CreateInternalInput{
		BarberID:     barberID,
		ServiceID:    serviceID,
		AttendeeName: fullName,
		StartsAt:     starts,
		EndsAt:       starts.Add(30 * time.Minute),
		Origin:       booking.OriginPublic,
		Service: booking.ServiceSnapshot{
			Name:             "Corte de prueba HU-098",
			DurationMinutes:  30,
			PriceAmountCents: 2000000,
			Currency:         "COP",
		},
		Customer: booking.CustomerInput{
			New: &booking.NewCustomerInput{FullName: fullName},
		},
		Actor: booking.Actor{Type: booking.ActorTypeSystem},
	})
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	appointment := created.Appointment

	if status != booking.StatusConfirmed {
		err := db.InTenantTx(context.Background(), database.BarbershopID(shopID), func(ctx context.Context, q database.Queries) error {
			_, err := q.Exec(ctx, `UPDATE appointment SET status = $1 WHERE id = $2`, string(status), appointment.ID)
			return err
		})
		if err != nil {
			t.Fatalf("forzar status %q: %v", status, err)
		}
		appointment.Status = status
	}
	return appointment
}

// tokenState describe cómo insertar la fila appointment_access_token de una
// prueba: vigente, vencida o revocada (DEC-089, CA-098-02).
type tokenState struct {
	expiresIn time.Duration // 0 = sin vencimiento
	expired   bool
	revoked   bool
}

// insertAccessToken genera un token en claro nuevo, inserta su hash con el
// estado pedido y devuelve el valor en claro (lo único que
// ResolveAppointmentByTokenHash acepta).
func insertAccessToken(t *testing.T, db *database.DB, shopID, appointmentID string, st tokenState) string {
	t.Helper()
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	plain := hex.EncodeToString(buf)
	sum := sha256.Sum256([]byte(plain))
	hash := hex.EncodeToString(sum[:])

	issuedAt := time.Now().UTC().Add(-1 * time.Hour)
	if st.expired {
		issuedAt = time.Now().UTC().Add(-120 * 24 * time.Hour)
	}

	var expiresAt *time.Time
	switch {
	case st.expired:
		e := issuedAt.Add(10 * 24 * time.Hour)
		expiresAt = &e
	case st.expiresIn > 0:
		e := issuedAt.Add(st.expiresIn)
		expiresAt = &e
	}

	var revokedAt *time.Time
	if st.revoked {
		r := issuedAt.Add(time.Minute)
		revokedAt = &r
	}

	err := db.InTenantTx(context.Background(), database.BarbershopID(shopID), func(ctx context.Context, q database.Queries) error {
		_, err := q.Exec(ctx,
			`INSERT INTO appointment_access_token (barbershop_id, appointment_id, token_hash, issued_at, expires_at, revoked_at)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			shopID, appointmentID, hash, issuedAt, expiresAt, revokedAt,
		)
		return err
	})
	if err != nil {
		t.Fatalf("insertAccessToken: %v", err)
	}
	return plain
}

func TestResolveAppointmentByTokenHash_ValidToken_ReturnsView(t *testing.T) {
	db := setupTestDB(t)
	repo := customeraccesspostgres.New(db)

	appointment := createTestAppointment(t, db, shopUno, barberUno, serviceUno, booking.StatusConfirmed)
	tokenPlain := insertAccessToken(t, db, shopUno, appointment.ID, tokenState{expiresIn: 90 * 24 * time.Hour})

	sum := sha256.Sum256([]byte(tokenPlain))
	hash := hex.EncodeToString(sum[:])

	view, found, err := repo.ResolveAppointmentByTokenHash(context.Background(), hash)
	if err != nil {
		t.Fatalf("ResolveAppointmentByTokenHash: %v", err)
	}
	if !found {
		t.Fatal("expected found = true for a valid token")
	}
	if view.BarbershopName != "Barbería de prueba HU-098 Uno" {
		t.Fatalf("BarbershopName = %q", view.BarbershopName)
	}
	if view.Timezone != "America/Bogota" {
		t.Fatalf("Timezone = %q", view.Timezone)
	}
	if view.AttendeeName != appointment.AttendeeName {
		t.Fatalf("AttendeeName = %q, want %q", view.AttendeeName, appointment.AttendeeName)
	}
	if view.ServiceName != "Corte de prueba HU-098" || view.DurationMinutes != 30 {
		t.Fatalf("service snapshot inesperado: %+v", view)
	}
	if view.BarberName != "Barbero HU-098 Uno" {
		t.Fatalf("BarberName = %q", view.BarberName)
	}
	if !view.StartsAt.Equal(appointment.StartsAt) || !view.EndsAt.Equal(appointment.EndsAt) {
		t.Fatalf("intervalo inesperado: %+v", view)
	}
	if view.Status != string(booking.StatusConfirmed) {
		t.Fatalf("Status = %q", view.Status)
	}
	if view.CancellationDeadlineMinutes != 30 || !view.LateCancellationClientAllowed || view.LateCancellationReasonRequired {
		t.Fatalf("política de cancelación inesperada: %+v", view)
	}
}

func TestResolveAppointmentByTokenHash_ExpiredToken_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := customeraccesspostgres.New(db)

	appointment := createTestAppointment(t, db, shopUno, barberUno, serviceUno, booking.StatusConfirmed)
	tokenPlain := insertAccessToken(t, db, shopUno, appointment.ID, tokenState{expired: true})

	sum := sha256.Sum256([]byte(tokenPlain))
	hash := hex.EncodeToString(sum[:])

	_, found, err := repo.ResolveAppointmentByTokenHash(context.Background(), hash)
	if err != nil {
		t.Fatalf("ResolveAppointmentByTokenHash: %v", err)
	}
	if found {
		t.Fatal("un token vencido no debe resolver (CA-098-02)")
	}
}

func TestResolveAppointmentByTokenHash_RevokedToken_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := customeraccesspostgres.New(db)

	appointment := createTestAppointment(t, db, shopUno, barberUno, serviceUno, booking.StatusConfirmed)
	tokenPlain := insertAccessToken(t, db, shopUno, appointment.ID, tokenState{expiresIn: 90 * 24 * time.Hour, revoked: true})

	sum := sha256.Sum256([]byte(tokenPlain))
	hash := hex.EncodeToString(sum[:])

	_, found, err := repo.ResolveAppointmentByTokenHash(context.Background(), hash)
	if err != nil {
		t.Fatalf("ResolveAppointmentByTokenHash: %v", err)
	}
	if found {
		t.Fatal("un token revocado no debe resolver (CA-098-02)")
	}
}

func TestResolveAppointmentByTokenHash_UnknownToken_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := customeraccesspostgres.New(db)

	sum := sha256.Sum256([]byte("token-que-nunca-se-emitio"))
	hash := hex.EncodeToString(sum[:])

	_, found, err := repo.ResolveAppointmentByTokenHash(context.Background(), hash)
	if err != nil {
		t.Fatalf("ResolveAppointmentByTokenHash: %v", err)
	}
	if found {
		t.Fatal("un hash nunca emitido no debe resolver")
	}
}

// TestResolveAppointmentByTokenHash_TenantIsolation cubre CA-098-05: un
// token de shopDos jamás expone datos ni política de shopUno, y viceversa,
// aunque ambas citas existan al mismo tiempo con políticas deliberadamente
// distintas (ver comentario de shopUno/shopDos arriba).
func TestResolveAppointmentByTokenHash_TenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	repo := customeraccesspostgres.New(db)

	apptUno := createTestAppointment(t, db, shopUno, barberUno, serviceUno, booking.StatusConfirmed)
	tokenUno := insertAccessToken(t, db, shopUno, apptUno.ID, tokenState{expiresIn: 90 * 24 * time.Hour})

	apptDos := createTestAppointment(t, db, shopDos, barberDos, serviceDos, booking.StatusConfirmed)
	tokenDos := insertAccessToken(t, db, shopDos, apptDos.ID, tokenState{expiresIn: 90 * 24 * time.Hour})

	sumUno := sha256.Sum256([]byte(tokenUno))
	viewUno, foundUno, err := repo.ResolveAppointmentByTokenHash(context.Background(), hex.EncodeToString(sumUno[:]))
	if err != nil || !foundUno {
		t.Fatalf("resolver token de shopUno: found=%v err=%v", foundUno, err)
	}
	if viewUno.BarbershopName != "Barbería de prueba HU-098 Uno" {
		t.Fatalf("el token de shopUno resolvió otra barbería: %+v", viewUno)
	}
	if viewUno.CancellationDeadlineMinutes != 30 || !viewUno.LateCancellationClientAllowed {
		t.Fatalf("el token de shopUno trae la política de otra barbería: %+v", viewUno)
	}

	sumDos := sha256.Sum256([]byte(tokenDos))
	viewDos, foundDos, err := repo.ResolveAppointmentByTokenHash(context.Background(), hex.EncodeToString(sumDos[:]))
	if err != nil || !foundDos {
		t.Fatalf("resolver token de shopDos: found=%v err=%v", foundDos, err)
	}
	if viewDos.BarbershopName != "Barbería de prueba HU-098 Dos" {
		t.Fatalf("el token de shopDos resolvió otra barbería: %+v", viewDos)
	}
	if viewDos.CancellationDeadlineMinutes != 90 || viewDos.LateCancellationClientAllowed {
		t.Fatalf("el token de shopDos trae la política de otra barbería: %+v", viewDos)
	}
}
