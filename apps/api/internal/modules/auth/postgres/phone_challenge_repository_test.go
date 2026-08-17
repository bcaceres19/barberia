// Pruebas de integración de HU-007 contra
// auth_phone_challenge_request/verify con PostgreSQL REAL. Requieren las
// siete migraciones y database/testdata/dos_barberias.sql +
// hu005_credenciales_sesiones.sql + hu007_reto_telefonico.sql cargados.
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/auth/postgres/...
package postgres_test

import (
	"context"
	"testing"

	"system-barbershop/internal/modules/auth"
	authpostgres "system-barbershop/internal/modules/auth/postgres"
)

// verifiedEmailA/verifiedEmailB corresponden a duena.a/dueno.b, con
// teléfono verificado por testdata/hu007_reto_telefonico.sql.
// unverifiedEmailA es barbero.a, SIN teléfono verificado a propósito.
const (
	verifiedEmailA   = "duena.a@ejemplo.test"
	unverifiedEmailA = "barbero.a@ejemplo.test"
)

func testChallengeCfg() auth.PhoneChallengeConfig {
	return auth.PhoneChallengeConfig{ExpiresSeconds: 300, RateWindowSeconds: 900, RateMaxActive: 3, ResendCooldownSeconds: 60}
}

func testCodeHash(t *testing.T, label string) string {
	t.Helper()
	return auth.HMACHex(uniqueToken(t, label), []byte("secreto-de-prueba-repositorio-challenge"))
}

func TestPhoneChallengeRepository_RequestChallenge_VerifiedPhoneAndEscalatedIP_Accepted(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	throttleRepo := authpostgres.NewThrottleRepository(db)
	challengeRepo := authpostgres.NewPhoneChallengeRepository(db)

	ipHash := testThrottleIPHash(t, "challenge-accepted")
	cfg := auth.ThrottleConfig{WindowSeconds: 900, EscalationSeconds: 86400, Threshold: 5, RetentionSeconds: 172800}
	for i := 0; i < 6; i++ {
		if _, _, _, err := throttleRepo.RegisterAttempt(context.Background(), ipHash, cfg); err != nil {
			t.Fatalf("escalate: %v", err)
		}
	}

	accepted, phone, err := challengeRepo.RequestChallenge(context.Background(), verifiedEmailA, ipHash, testCodeHash(t, "accepted"), testChallengeCfg())
	if err != nil {
		t.Fatalf("RequestChallenge: %v", err)
	}
	if !accepted {
		t.Fatal("expected accepted=true for a verified phone with an escalated IP")
	}
	if phone == "" {
		t.Fatal("expected a non-empty phone on the accepted path")
	}
}

func TestPhoneChallengeRepository_RequestChallenge_UnverifiedPhone_NotAccepted(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	throttleRepo := authpostgres.NewThrottleRepository(db)
	challengeRepo := authpostgres.NewPhoneChallengeRepository(db)

	ipHash := testThrottleIPHash(t, "challenge-unverified")
	cfg := auth.ThrottleConfig{WindowSeconds: 900, EscalationSeconds: 86400, Threshold: 5, RetentionSeconds: 172800}
	for i := 0; i < 6; i++ {
		if _, _, _, err := throttleRepo.RegisterAttempt(context.Background(), ipHash, cfg); err != nil {
			t.Fatalf("escalate: %v", err)
		}
	}

	accepted, phone, err := challengeRepo.RequestChallenge(context.Background(), unverifiedEmailA, ipHash, testCodeHash(t, "unverified"), testChallengeCfg())
	if err != nil {
		t.Fatalf("RequestChallenge: %v", err)
	}
	if accepted || phone != "" {
		t.Fatalf("expected accepted=false and empty phone for an unverified phone, got accepted=%v phone=%q", accepted, phone)
	}
}

func TestPhoneChallengeRepository_RequestChallenge_NonEscalatedIP_NotAccepted(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	challengeRepo := authpostgres.NewPhoneChallengeRepository(db)

	// IP nueva, jamás registrada en login_throttle: no está escalada.
	ipHash := testThrottleIPHash(t, "challenge-not-escalated")

	accepted, phone, err := challengeRepo.RequestChallenge(context.Background(), verifiedEmailA, ipHash, testCodeHash(t, "not-escalated"), testChallengeCfg())
	if err != nil {
		t.Fatalf("RequestChallenge: %v", err)
	}
	if accepted || phone != "" {
		t.Fatalf("expected accepted=false for a non-escalated IP, got accepted=%v phone=%q", accepted, phone)
	}
}

func TestPhoneChallengeRepository_RequestChallenge_UnknownAccount_NotAccepted(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	throttleRepo := authpostgres.NewThrottleRepository(db)
	challengeRepo := authpostgres.NewPhoneChallengeRepository(db)

	ipHash := testThrottleIPHash(t, "challenge-unknown-account")
	cfg := auth.ThrottleConfig{WindowSeconds: 900, EscalationSeconds: 86400, Threshold: 5, RetentionSeconds: 172800}
	for i := 0; i < 6; i++ {
		if _, _, _, err := throttleRepo.RegisterAttempt(context.Background(), ipHash, cfg); err != nil {
			t.Fatalf("escalate: %v", err)
		}
	}

	accepted, phone, err := challengeRepo.RequestChallenge(context.Background(), "no-existe@ejemplo.test", ipHash, testCodeHash(t, "unknown"), testChallengeCfg())
	if err != nil {
		t.Fatalf("RequestChallenge: %v", err)
	}
	if accepted || phone != "" {
		t.Fatalf("expected accepted=false for an unknown account, got accepted=%v phone=%q", accepted, phone)
	}
}

func TestPhoneChallengeRepository_RequestChallenge_ResendCooldown_RejectsImmediate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	throttleRepo := authpostgres.NewThrottleRepository(db)
	challengeRepo := authpostgres.NewPhoneChallengeRepository(db)

	ipHash := testThrottleIPHash(t, "challenge-cooldown")
	cfg := auth.ThrottleConfig{WindowSeconds: 900, EscalationSeconds: 86400, Threshold: 5, RetentionSeconds: 172800}
	for i := 0; i < 6; i++ {
		if _, _, _, err := throttleRepo.RegisterAttempt(context.Background(), ipHash, cfg); err != nil {
			t.Fatalf("escalate: %v", err)
		}
	}

	first, _, err := challengeRepo.RequestChallenge(context.Background(), verifiedEmailA, ipHash, testCodeHash(t, "cooldown-1"), testChallengeCfg())
	if err != nil || !first {
		t.Fatalf("expected first request accepted, got accepted=%v err=%v", first, err)
	}

	second, phone, err := challengeRepo.RequestChallenge(context.Background(), verifiedEmailA, ipHash, testCodeHash(t, "cooldown-2"), testChallengeCfg())
	if err != nil {
		t.Fatalf("RequestChallenge: %v", err)
	}
	if second || phone != "" {
		t.Fatalf("expected the immediate resend to be rejected by cooldown, got accepted=%v phone=%q", second, phone)
	}
}

func TestPhoneChallengeRepository_VerifyChallenge_CorrectCode_SucceedsAndClearsThrottle(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	throttleRepo := authpostgres.NewThrottleRepository(db)
	challengeRepo := authpostgres.NewPhoneChallengeRepository(db)

	ipHash := testThrottleIPHash(t, "challenge-verify-success")
	throttleCfg := auth.ThrottleConfig{WindowSeconds: 900, EscalationSeconds: 86400, Threshold: 5, RetentionSeconds: 172800}
	for i := 0; i < 6; i++ {
		if _, _, _, err := throttleRepo.RegisterAttempt(context.Background(), ipHash, throttleCfg); err != nil {
			t.Fatalf("escalate: %v", err)
		}
	}

	codeHash := testCodeHash(t, "verify-success")
	accepted, _, err := challengeRepo.RequestChallenge(context.Background(), verifiedEmailA, ipHash, codeHash, testChallengeCfg())
	if err != nil || !accepted {
		t.Fatalf("expected request accepted, got accepted=%v err=%v", accepted, err)
	}

	ok, err := challengeRepo.VerifyChallenge(context.Background(), verifiedEmailA, ipHash, codeHash)
	if err != nil {
		t.Fatalf("VerifyChallenge: %v", err)
	}
	if !ok {
		t.Fatal("expected VerifyChallenge to succeed with the correct code")
	}

	// El escalamiento debe haberse limpiado en la misma operación
	// (DEC-062): la siguiente solicitud de reto ya no debe aceptarse
	// porque la IP dejó de estar escalada.
	acceptedAfter, _, err := challengeRepo.RequestChallenge(context.Background(), verifiedEmailA, ipHash, testCodeHash(t, "after-clear"), testChallengeCfg())
	if err != nil {
		t.Fatalf("RequestChallenge after verify: %v", err)
	}
	if acceptedAfter {
		t.Fatal("expected the IP to no longer be escalated after a successful verify (DEC-062 clears login_throttle)")
	}
}

func TestPhoneChallengeRepository_VerifyChallenge_WrongCode_Fails(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	throttleRepo := authpostgres.NewThrottleRepository(db)
	challengeRepo := authpostgres.NewPhoneChallengeRepository(db)

	ipHash := testThrottleIPHash(t, "challenge-verify-wrong")
	throttleCfg := auth.ThrottleConfig{WindowSeconds: 900, EscalationSeconds: 86400, Threshold: 5, RetentionSeconds: 172800}
	for i := 0; i < 6; i++ {
		if _, _, _, err := throttleRepo.RegisterAttempt(context.Background(), ipHash, throttleCfg); err != nil {
			t.Fatalf("escalate: %v", err)
		}
	}

	codeHash := testCodeHash(t, "verify-wrong-correct")
	accepted, _, err := challengeRepo.RequestChallenge(context.Background(), verifiedEmailA, ipHash, codeHash, testChallengeCfg())
	if err != nil || !accepted {
		t.Fatalf("expected request accepted, got accepted=%v err=%v", accepted, err)
	}

	wrongHash := testCodeHash(t, "verify-wrong-attempt")
	ok, err := challengeRepo.VerifyChallenge(context.Background(), verifiedEmailA, ipHash, wrongHash)
	if err != nil {
		t.Fatalf("VerifyChallenge: %v", err)
	}
	if ok {
		t.Fatal("expected VerifyChallenge to fail with an incorrect code")
	}
}

func TestPhoneChallengeRepository_VerifyChallenge_UnknownAccount_Fails(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	challengeRepo := authpostgres.NewPhoneChallengeRepository(db)

	ok, err := challengeRepo.VerifyChallenge(context.Background(), "no-existe@ejemplo.test", testThrottleIPHash(t, "verify-unknown"), testCodeHash(t, "verify-unknown-code"))
	if err != nil {
		t.Fatalf("VerifyChallenge: %v", err)
	}
	if ok {
		t.Fatal("expected VerifyChallenge to fail for an unknown account")
	}
}

// TestPhoneChallengeRepository_VerifyChallenge_WrongIP_Fails cubre que el
// reto está atado a la IP concreta que lo solicitó (DEC-062): un código
// correcto desde OTRA IP no debe verificar.
func TestPhoneChallengeRepository_VerifyChallenge_WrongIP_Fails(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	throttleRepo := authpostgres.NewThrottleRepository(db)
	challengeRepo := authpostgres.NewPhoneChallengeRepository(db)

	ipHashRequester := testThrottleIPHash(t, "challenge-wrong-ip-requester")
	ipHashOther := testThrottleIPHash(t, "challenge-wrong-ip-other")
	throttleCfg := auth.ThrottleConfig{WindowSeconds: 900, EscalationSeconds: 86400, Threshold: 5, RetentionSeconds: 172800}
	for i := 0; i < 6; i++ {
		if _, _, _, err := throttleRepo.RegisterAttempt(context.Background(), ipHashRequester, throttleCfg); err != nil {
			t.Fatalf("escalate: %v", err)
		}
	}

	codeHash := testCodeHash(t, "wrong-ip")
	accepted, _, err := challengeRepo.RequestChallenge(context.Background(), verifiedEmailA, ipHashRequester, codeHash, testChallengeCfg())
	if err != nil || !accepted {
		t.Fatalf("expected request accepted, got accepted=%v err=%v", accepted, err)
	}

	ok, err := challengeRepo.VerifyChallenge(context.Background(), verifiedEmailA, ipHashOther, codeHash)
	if err != nil {
		t.Fatalf("VerifyChallenge: %v", err)
	}
	if ok {
		t.Fatal("expected VerifyChallenge to fail when the code is presented from a different IP than the one that requested it")
	}
}
