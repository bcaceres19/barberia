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
	"strings"
	"testing"

	"system-barbershop/internal/modules/auth"
	authpostgres "system-barbershop/internal/modules/auth/postgres"
	"system-barbershop/internal/platform/database"
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

	accepted, dest, err := challengeRepo.RequestChallenge(context.Background(), verifiedEmailA, ipHash, testCodeHash(t, "accepted"), testChallengeCfg())
	if err != nil {
		t.Fatalf("RequestChallenge: %v", err)
	}
	if !accepted {
		t.Fatal("expected accepted=true for a verified phone with an escalated IP")
	}
	if dest.Phone == "" || dest.Email != "" {
		t.Fatalf("expected only the phone for the whatsapp channel (DEC-139), got %+v", dest)
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

	accepted, dest, err := challengeRepo.RequestChallenge(context.Background(), unverifiedEmailA, ipHash, testCodeHash(t, "unverified"), testChallengeCfg())
	if err != nil {
		t.Fatalf("RequestChallenge: %v", err)
	}
	if accepted || !dest.None() {
		t.Fatalf("expected accepted=false and empty phone for an unverified phone, got accepted=%v dest=%+v", accepted, dest)
	}
}

func TestPhoneChallengeRepository_RequestChallenge_NonEscalatedIP_NotAccepted(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	challengeRepo := authpostgres.NewPhoneChallengeRepository(db)

	// IP nueva, jamás registrada en login_throttle: no está escalada.
	ipHash := testThrottleIPHash(t, "challenge-not-escalated")

	accepted, dest, err := challengeRepo.RequestChallenge(context.Background(), verifiedEmailA, ipHash, testCodeHash(t, "not-escalated"), testChallengeCfg())
	if err != nil {
		t.Fatalf("RequestChallenge: %v", err)
	}
	if accepted || !dest.None() {
		t.Fatalf("expected accepted=false for a non-escalated IP, got accepted=%v dest=%+v", accepted, dest)
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

	accepted, dest, err := challengeRepo.RequestChallenge(context.Background(), "no-existe@ejemplo.test", ipHash, testCodeHash(t, "unknown"), testChallengeCfg())
	if err != nil {
		t.Fatalf("RequestChallenge: %v", err)
	}
	if accepted || !dest.None() {
		t.Fatalf("expected accepted=false for an unknown account, got accepted=%v dest=%+v", accepted, dest)
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

	second, dest, err := challengeRepo.RequestChallenge(context.Background(), verifiedEmailA, ipHash, testCodeHash(t, "cooldown-2"), testChallengeCfg())
	if err != nil {
		t.Fatalf("RequestChallenge: %v", err)
	}
	if second || !dest.None() {
		t.Fatalf("expected the immediate resend to be rejected by cooldown, got accepted=%v dest=%+v", second, dest)
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

// escalateAndRequest deja un reto activo para email atado a la IP indicada.
func escalateAndRequest(t *testing.T, throttleRepo *authpostgres.ThrottleRepository, challengeRepo *authpostgres.PhoneChallengeRepository, email, ipHash, label string) string {
	t.Helper()
	throttleCfg := auth.ThrottleConfig{WindowSeconds: 900, EscalationSeconds: 86400, Threshold: 5, RetentionSeconds: 172800}
	for i := 0; i < 6; i++ {
		if _, _, _, err := throttleRepo.RegisterAttempt(context.Background(), ipHash, throttleCfg); err != nil {
			t.Fatalf("escalate: %v", err)
		}
	}
	codeHash := testCodeHash(t, label)
	accepted, _, err := challengeRepo.RequestChallenge(context.Background(), email, ipHash, codeHash, testChallengeCfg())
	if err != nil || !accepted {
		t.Fatalf("expected request accepted, got accepted=%v err=%v", accepted, err)
	}
	return codeHash
}

// Un digest que nunca coincide avanza el contador hasta invalidar el reto.
func TestPhoneChallengeRepository_VerifyChallenge_WrongDigest_ExhaustsAttempts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	throttleRepo := authpostgres.NewThrottleRepository(db)
	challengeRepo := authpostgres.NewPhoneChallengeRepository(db)

	ip := testThrottleIPHash(t, "challenge-phone-exhaust")
	codeHash := escalateAndRequest(t, throttleRepo, challengeRepo, verifiedEmailA, ip, "phone-exhaust")

	never := strings.Repeat("0", 64)
	for i := 0; i < 5; i++ {
		if ok, err := challengeRepo.VerifyChallenge(context.Background(), verifiedEmailA, ip, never); err != nil || ok {
			t.Fatalf("attempt %d: ok=%v err=%v", i+1, ok, err)
		}
	}
	if ok, err := challengeRepo.VerifyChallenge(context.Background(), verifiedEmailA, ip, codeHash); err != nil || ok {
		t.Fatalf("the correct digest must fail once attempts are exhausted: ok=%v err=%v", ok, err)
	}
}

// --- DEC-139: el destino sale del canal de la barbería de la cuenta -------------

const (
	channelShopA    = "11111111-1111-1111-1111-111111111111"
	channelShopB    = "22222222-2222-2222-2222-222222222222"
	inactiveEmailB  = "barbero.b@ejemplo.test"
	channelRestored = "whatsapp"
)

// setLoginChannel fija el canal de una barbería dentro de su propio tenant (la
// misma ruta que usa la aplicación) y lo restablece a WhatsApp al terminar la
// prueba, que es el valor con el que testdata/hu007_reto_telefonico.sql deja
// las dos barberías.
func setLoginChannel(t *testing.T, db *database.DB, shopID, channel string) {
	t.Helper()
	write := func(db *database.DB, value string) {
		err := db.InTenantTx(context.Background(), database.BarbershopID(shopID), func(ctx context.Context, q database.Queries) error {
			_, err := q.Exec(ctx, `UPDATE barbershop SET login_otp_channel = $2 WHERE id = $1`, shopID, value)
			return err
		})
		if err != nil {
			t.Fatalf("set login_otp_channel=%s: %v", value, err)
		}
	}
	write(db, channel)
	// El cierre del pool de la prueba (defer) corre antes que t.Cleanup, así que la
	// restauración abre su propia conexión.
	t.Cleanup(func() {
		fresh := setupTestDB(t)
		defer fresh.Close()
		write(fresh, channelRestored)
	})
}

func escalatedIP(t *testing.T, db *database.DB, label string) string {
	t.Helper()
	ipHash := testThrottleIPHash(t, label)
	cfg := auth.ThrottleConfig{WindowSeconds: 900, EscalationSeconds: 86400, Threshold: 5, RetentionSeconds: 172800}
	throttleRepo := authpostgres.NewThrottleRepository(db)
	for i := 0; i < 6; i++ {
		if _, _, _, err := throttleRepo.RegisterAttempt(context.Background(), ipHash, cfg); err != nil {
			t.Fatalf("escalate: %v", err)
		}
	}
	return ipHash
}

func TestPhoneChallengeRepository_RequestChallenge_ChannelDecidesDestinations(t *testing.T) {
	cases := []struct {
		name      string
		channel   string
		account   string
		accepted  bool
		wantPhone bool
		wantEmail bool
	}{
		{"email, phone verified", "email", verifiedEmailA, true, false, true},
		{"email, phone NOT verified still receives by email", "email", unverifiedEmailA, true, false, true},
		{"whatsapp, phone verified", "whatsapp", verifiedEmailA, true, true, false},
		{"whatsapp, phone NOT verified: no destination and no email fallback", "whatsapp", unverifiedEmailA, false, false, false},
		{"both, phone verified: both destinations", "both", verifiedEmailA, true, true, true},
		{"both, phone NOT verified: only the email", "both", unverifiedEmailA, true, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()
			setLoginChannel(t, db, channelShopA, tc.channel)
			ip := escalatedIP(t, db, "channel-"+tc.name)

			accepted, dest, err := authpostgres.NewPhoneChallengeRepository(db).RequestChallenge(
				context.Background(), tc.account, ip, testCodeHash(t, "channel-"+tc.name), testChallengeCfg())
			if err != nil {
				t.Fatalf("RequestChallenge: %v", err)
			}
			if accepted != tc.accepted {
				t.Fatalf("accepted = %v, want %v (dest=%+v)", accepted, tc.accepted, dest)
			}
			if (dest.Phone != "") != tc.wantPhone || (dest.Email != "") != tc.wantEmail {
				t.Fatalf("destinations = %+v, want phone=%v email=%v", dest, tc.wantPhone, tc.wantEmail)
			}
			if dest.Email != "" && dest.Email != tc.account {
				t.Fatalf("the email destination must be the account's own email, got %q", dest.Email)
			}
		})
	}
}

// El canal de una barbería no cambia el destino de la otra (RN-TEN-01).
func TestPhoneChallengeRepository_RequestChallenge_ChannelIsPerBarbershop(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	setLoginChannel(t, db, channelShopA, "both")
	setLoginChannel(t, db, channelShopB, "whatsapp")
	repo := authpostgres.NewPhoneChallengeRepository(db)

	accepted, dest, err := repo.RequestChallenge(context.Background(), verifiedEmailB, escalatedIP(t, db, "per-shop-b"), testCodeHash(t, "per-shop-b"), testChallengeCfg())
	if err != nil || !accepted {
		t.Fatalf("expected B accepted, got accepted=%v err=%v", accepted, err)
	}
	if dest.Phone == "" || dest.Email != "" {
		t.Fatalf("barbershop B is on whatsapp and must not inherit A's channel, got %+v", dest)
	}

	accepted, dest, err = repo.RequestChallenge(context.Background(), verifiedEmailA, escalatedIP(t, db, "per-shop-a"), testCodeHash(t, "per-shop-a"), testChallengeCfg())
	if err != nil || !accepted {
		t.Fatalf("expected A accepted, got accepted=%v err=%v", accepted, err)
	}
	if dest.Phone == "" || dest.Email == "" {
		t.Fatalf("barbershop A is on both channels, got %+v", dest)
	}
}

// Una cuenta inactiva no recibe destino con ningún canal.
func TestPhoneChallengeRepository_RequestChallenge_InactiveAccount_NoDestination(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	setLoginChannel(t, db, channelShopB, "both")

	accepted, dest, err := authpostgres.NewPhoneChallengeRepository(db).RequestChallenge(
		context.Background(), inactiveEmailB, escalatedIP(t, db, "inactive"), testCodeHash(t, "inactive"), testChallengeCfg())
	if err != nil {
		t.Fatalf("RequestChallenge: %v", err)
	}
	if accepted || !dest.None() {
		t.Fatalf("an inactive account must look like an unknown one, got accepted=%v dest=%+v", accepted, dest)
	}
}
