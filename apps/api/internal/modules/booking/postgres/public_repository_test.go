// Package postgres_test (pruebas de integración, HU-097) requiere
// PostgreSQL REAL con las diecisiete migraciones aplicadas (incluida
// 20260916100000_create_appointment_access_token.sql) y
// database/testdata/dos_barberias.sql + database/testdata/hu060_citas.sql
// cargados: mismos fixtures que repository_test.go/manual_repository_test.go
// (HU-060/HU-061), reutilizados aquí porque CreatePublic no depende de
// ninguna tabla adicional aparte de appointment_access_token.
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/booking/postgres/...
package postgres_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"system-barbershop/internal/modules/booking"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

// uniqueToken produce un par (plain, hash) con la MISMA forma que
// publicbooking.generateAccessToken (32 bytes codificados, SHA-256
// hexadecimal), suficiente para estas pruebas: no reutiliza esa función
// (vive en otro paquete, deliberadamente no importado aquí -CA-002-06 no
// distingue producción de pruebas, pero booking/postgres nunca depende de
// publicbooking en ningún caso).
func uniqueToken(t *testing.T) (plain, hash string) {
	t.Helper()
	plain = "test-token-" + uniqueSuffix(t)
	sum := sha256.Sum256([]byte(plain))
	return plain, hex.EncodeToString(sum[:])
}

func publicInput(t *testing.T, barberID, serviceID string, start time.Time, suffix string, customer booking.CustomerInput) booking.CreatePublicInput {
	tokenPlain, tokenHash := uniqueToken(t)
	issuedAt := time.Now().UTC()
	return booking.CreatePublicInput{
		BarberID:     barberID,
		ServiceID:    serviceID,
		AttendeeName: "Persona atendida " + suffix,
		StartsAt:     start,
		EndsAt:       start.Add(30 * time.Minute),
		Service: booking.ServiceSnapshot{
			Name:             "Corte público " + suffix,
			DurationMinutes:  30,
			PriceAmountCents: 1800000,
			Currency:         "COP",
		},
		BarbershopName: "Barbería de prueba " + suffix,
		Timezone:       "America/Bogota",
		TokenPlain:     tokenPlain,
		Customer:       customer,
		TokenHash:      tokenHash,
		TokenIssuedAt:  issuedAt,
		TokenExpiresAt: issuedAt.Add(90 * 24 * time.Hour),
	}
}

func TestCreatePublic_NewCustomer_PersistsAppointmentHistoryAndToken(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	input := publicInput(t, barberQ1, serviceQ, start, suffix, booking.CustomerInput{
		New: &booking.NewCustomerInput{FullName: "Cliente público " + suffix, Phone: strPtr("+57300" + uniqueDigits(t, 7)), Email: strPtr(suffix + "@example.com")},
	})
	key := idempotency.Key("public-" + suffix)
	fingerprint := idempotency.Fingerprint("3123456789abcdef0123456789abcdef")

	result, err := repo.CreatePublic(context.Background(), string(shopQ), input, key, fingerprint)
	if err != nil {
		t.Fatalf("CreatePublic: %v", err)
	}
	if result.Decision.Outcome != idempotency.OutcomeProceed {
		t.Fatalf("Outcome = %v, want OutcomeProceed", result.Decision.Outcome)
	}

	var parsed struct {
		AttendeeName    string `json:"attendeeName"`
		BarbershopName  string `json:"barbershopName"`
		ServiceName     string `json:"serviceName"`
		DurationMinutes int    `json:"durationMinutes"`
		Currency        string `json:"currency"`
		AccessToken     string `json:"accessToken"`
	}
	if err := json.Unmarshal([]byte(result.Response.Body), &parsed); err != nil {
		t.Fatalf("unmarshal stored response: %v", err)
	}
	if parsed.AccessToken != input.TokenPlain {
		t.Fatalf("AccessToken = %q, want %q (el valor en claro debe viajar tal cual una sola vez)", parsed.AccessToken, input.TokenPlain)
	}
	if parsed.AttendeeName != input.AttendeeName || parsed.ServiceName != input.Service.Name || parsed.DurationMinutes != 30 {
		t.Fatalf("stored response fields mismatch: %+v", parsed)
	}

	err = db.InTenantTx(context.Background(), shopQ, func(ctx context.Context, q database.Queries) error {
		var origin, status, actorType string
		var customerID string
		scanErr := q.QueryRow(ctx,
			`SELECT origin, status, customer_id FROM appointment
			  WHERE barbershop_id = $1 AND barber_id = $2 AND starts_at = $3`,
			string(shopQ), barberQ1, start,
		).Scan(&origin, &status, &customerID)
		if scanErr != nil {
			return scanErr
		}
		if origin != "public" || status != "confirmed" {
			t.Fatalf("origin/status = %s/%s, want public/confirmed", origin, status)
		}

		scanErr = q.QueryRow(ctx,
			`SELECT actor_type FROM appointment_history
			  WHERE barbershop_id = $1 AND actor_customer_id = $2 AND event_type = 'appointment_created'`,
			string(shopQ), customerID,
		).Scan(&actorType)
		if scanErr != nil {
			return scanErr
		}
		if actorType != "customer" {
			t.Fatalf("actor_type = %s, want customer", actorType)
		}

		var tokenHash string
		var expiresAt time.Time
		scanErr = q.QueryRow(ctx,
			`SELECT token_hash, expires_at FROM appointment_access_token
			  WHERE barbershop_id = $1 AND appointment_id = (
			    SELECT id FROM appointment WHERE barbershop_id = $1 AND customer_id = $2
			  )`,
			string(shopQ), customerID,
		).Scan(&tokenHash, &expiresAt)
		if scanErr != nil {
			return scanErr
		}
		if tokenHash != input.TokenHash {
			t.Fatalf("token_hash = %s, want %s", tokenHash, input.TokenHash)
		}
		if tokenHash == input.TokenPlain {
			t.Fatalf("token_hash nunca debe coincidir con el valor en claro")
		}
		if !expiresAt.After(input.TokenIssuedAt) {
			t.Fatalf("expires_at debe ser posterior a issuedAt")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("verify persisted rows: %v", err)
	}
}

func TestCreatePublic_SameKeyAndFingerprint_Replays(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	input := publicInput(t, barberQ1, serviceQ, start, suffix, booking.CustomerInput{
		New: &booking.NewCustomerInput{FullName: "Cliente réplica " + suffix},
	})
	key := idempotency.Key("public-replay-" + suffix)
	fingerprint := idempotency.Fingerprint("4123456789abcdef0123456789abcdef")

	first, err := repo.CreatePublic(context.Background(), string(shopQ), input, key, fingerprint)
	if err != nil {
		t.Fatalf("first CreatePublic: %v", err)
	}
	second, err := repo.CreatePublic(context.Background(), string(shopQ), input, key, fingerprint)
	if err != nil {
		t.Fatalf("second CreatePublic: %v", err)
	}
	if second.Decision.Outcome != idempotency.OutcomeReplay {
		t.Fatalf("Outcome = %v, want OutcomeReplay", second.Decision.Outcome)
	}
	if first.Response.Body != second.Response.Body {
		t.Fatalf("replay must reproduce byte-identical body:\nfirst=%s\nsecond=%s", first.Response.Body, second.Response.Body)
	}

	err = db.InTenantTx(context.Background(), shopQ, func(ctx context.Context, q database.Queries) error {
		var count int
		scanErr := q.QueryRow(ctx,
			`SELECT count(*) FROM appointment WHERE barbershop_id = $1 AND barber_id = $2 AND starts_at = $3`,
			string(shopQ), barberQ1, start,
		).Scan(&count)
		if scanErr != nil {
			return scanErr
		}
		if count != 1 {
			t.Fatalf("replay must not duplicate the appointment, found %d rows", count)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("verify no duplicate: %v", err)
	}
}

// TestCreatePublic_AccessToken_TenantIsolated cubre RN-TEN-01 para la tabla
// nueva `appointment_access_token`: un token emitido dentro de shopQ nunca
// es visible leyendo con el contexto de tenant de shopR, verificado con el
// rol real `barberia_app` (RLS forzada), nunca con un rol administrativo.
func TestCreatePublic_AccessToken_TenantIsolated(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	input := publicInput(t, barberQ1, serviceQ, start, suffix, booking.CustomerInput{
		New: &booking.NewCustomerInput{FullName: "Cliente aislado " + suffix},
	})
	key := idempotency.Key("public-isolation-" + suffix)
	fingerprint := idempotency.Fingerprint("7123456789abcdef0123456789abcdef")

	if _, err := repo.CreatePublic(context.Background(), string(shopQ), input, key, fingerprint); err != nil {
		t.Fatalf("CreatePublic: %v", err)
	}

	err := db.InTenantTx(context.Background(), shopR, func(ctx context.Context, q database.Queries) error {
		var count int
		scanErr := q.QueryRow(ctx,
			`SELECT count(*) FROM appointment_access_token WHERE token_hash = $1`,
			input.TokenHash,
		).Scan(&count)
		if scanErr != nil {
			return scanErr
		}
		if count != 0 {
			t.Fatalf("RN-TEN-01: shopR must never see a token issued for shopQ, found %d rows", count)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("verify tenant isolation: %v", err)
	}
}

// TestCreatePublic_ConcurrentOverlap_ExactlyOneSucceeds es la prueba
// central de RN-CON-01/RN-CON-02/RN-CON-03 aplicada a T1 pública: dos
// confirmaciones simultáneas del mismo barbero, con intervalos que se
// cruzan, deben dejar exactamente una cita persistida y traducir la
// perdedora a apperr.KindScheduleConflict (RN-CON-05/DEC-090) -NUNCA
// apperr.KindConflict genérico, que publicbooking no sabría distinguir de
// un conflicto de unicidad de `customer`.
func TestCreatePublic_ConcurrentOverlap_ExactlyOneSucceeds(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	inputA := publicInput(t, barberQ2, serviceQ, start, suffix+"-A", booking.CustomerInput{
		New: &booking.NewCustomerInput{FullName: "Cliente A " + suffix},
	})
	inputB := publicInput(t, barberQ2, serviceQ, start.Add(10*time.Minute), suffix+"-B", booking.CustomerInput{
		New: &booking.NewCustomerInput{FullName: "Cliente B " + suffix},
	})
	keyA := idempotency.Key("public-race-a-" + suffix)
	keyB := idempotency.Key("public-race-b-" + suffix)
	fingerprintA := idempotency.Fingerprint("5123456789abcdef0123456789abcdef")
	fingerprintB := idempotency.Fingerprint("6123456789abcdef0123456789abcdef")

	var (
		wg               sync.WaitGroup
		resultA, resultB booking.CreatePublicResult
		errA, errB       error
		startBarrier     = make(chan struct{})
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		close(startBarrier)
		resultA, errA = repo.CreatePublic(context.Background(), string(shopQ), inputA, keyA, fingerprintA)
	}()
	go func() {
		defer wg.Done()
		<-startBarrier
		resultB, errB = repo.CreatePublic(context.Background(), string(shopQ), inputB, keyB, fingerprintB)
	}()
	wg.Wait()

	succeeded, conflicted := 0, 0
	for _, res := range []struct {
		result booking.CreatePublicResult
		err    error
	}{{resultA, errA}, {resultB, errB}} {
		switch {
		case res.err == nil:
			succeeded++
			if res.result.Decision.Outcome != idempotency.OutcomeProceed {
				t.Fatalf("a successful result must carry OutcomeProceed, got %v", res.result.Decision.Outcome)
			}
		default:
			appErr, ok := apperr.As(res.err)
			if !ok || appErr.Kind != apperr.KindScheduleConflict {
				t.Fatalf("expected apperr.KindScheduleConflict for the losing goroutine, got %v (%T)", res.err, res.err)
			}
			conflicted++
		}
	}

	if succeeded != 1 || conflicted != 1 {
		t.Fatalf("RN-CON-01/RN-CON-03: expected exactly one success and one conflict, got succeeded=%d conflicted=%d (errA=%v errB=%v)",
			succeeded, conflicted, errA, errB)
	}

	err := db.InTenantTx(context.Background(), shopQ, func(ctx context.Context, q database.Queries) error {
		var count int
		scanErr := q.QueryRow(ctx,
			`SELECT count(*) FROM appointment
			  WHERE barbershop_id = $1 AND barber_id = $2
			    AND tstzrange(starts_at, ends_at, '[)') && tstzrange($3, $4, '[)')
			    AND occupies_schedule`,
			string(shopQ), barberQ2, start.Add(-time.Hour), start.Add(time.Hour),
		).Scan(&count)
		if scanErr != nil {
			return scanErr
		}
		if count != 1 {
			t.Fatalf("expected exactly one persisted crossing appointment, found %d", count)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("verify persisted count: %v", err)
	}
}
