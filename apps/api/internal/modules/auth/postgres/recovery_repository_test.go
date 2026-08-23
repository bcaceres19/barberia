// Pruebas de integración de HU-008 contra
// auth_recovery_request/verify/current_credential/change_password con
// PostgreSQL REAL. Requieren las ocho migraciones y
// database/testdata/dos_barberias.sql + hu005_credenciales_sesiones.sql +
// hu007_reto_telefonico.sql cargados.
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/auth/postgres/...
//
// A propósito NO se agregan filas nuevas a staff_user: internal/platform/
// database/database_test.go verifica el aislamiento RLS contando
// exactamente 2 usuarios por barbería sobre dos_barberias.sql, y una fila
// nueva rompería esa prueba, ajena a HU-008. El cooldown/límite de reenvío
// de DEC-064 es por cuenta, así que cada una de las dos únicas cuentas con
// teléfono verificado (duena.a, dueno.b, de hu007_reto_telefonico.sql) se
// usa en UN solo escenario secuencial de punta a punta, nunca repartida
// entre varias funciones de prueba independientes.
package postgres_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"system-barbershop/internal/modules/auth"
	authpostgres "system-barbershop/internal/modules/auth/postgres"
)

// verifiedEmailB es dueno.b, con teléfono verificado por
// testdata/hu007_reto_telefonico.sql (+573000000003). verifiedEmailA y
// unverifiedEmailA ya están declaradas en phone_challenge_repository_test.go.
const verifiedEmailB = "dueno.b@ejemplo.test"

func testRecoveryCodeHash(t *testing.T, label string) string {
	t.Helper()
	return auth.HMACHex(uniqueToken(t, label), []byte("secreto-de-prueba-repositorio-recovery"))
}

func testResetTokenHash(t *testing.T, label string) string {
	t.Helper()
	return auth.HashToken(uniqueToken(t, label))
}

func testRecoveryCfg() auth.RecoveryConfig {
	return auth.RecoveryConfig{
		CodeExpiresSeconds: 900, CodeMaxAttempts: 5,
		ResendCooldownSeconds: 60, ResendWindowSeconds: 3600, ResendMaxPerWindow: 3,
		ResetTokenExpiresSeconds: 300,
	}
}

// requestRecoveryEventually reintenta RequestRecovery hasta que sea
// aceptada o se agote un plazo generoso. duena.a/dueno.b son las ÚNICAS dos
// cuentas con teléfono verificado del fixture compartido
// (hu007_reto_telefonico.sql), y cmd/api ejercita el mismo recorrido de
// sistema sobre las MISMAS cuentas (recovery_integration_test.go); `go test
// ./...` ejecuta paquetes distintos en paralelo contra el MISMO PostgreSQL,
// así que el cooldown real de 60 s (DEC-064) puede rechazar momentáneamente
// la primera solicitud si ese otro paquete acaba de crear un código para la
// misma cuenta. Esto no prueba el cooldown en sí —eso ya lo cubre,
// determinista, el subtest "immediate resend rejected by cooldown" de más
// abajo—: es solo tolerancia a la contención entre paquetes de prueba
// independientes sobre un fixture escaso (solo 2 cuentas verificadas en
// todo el módulo).
func requestRecoveryEventually(t *testing.T, repo *authpostgres.RecoveryRepository, email, codeHash string, cfg auth.RecoveryConfig) (phone, resolvedEmail string) {
	t.Helper()
	deadline := time.Now().Add(65 * time.Second)
	for {
		accepted, p, e, err := repo.RequestRecovery(context.Background(), email, codeHash, cfg)
		if err != nil {
			t.Fatalf("RequestRecovery: %v", err)
		}
		if accepted {
			return p, e
		}
		if time.Now().After(deadline) {
			t.Fatalf("RequestRecovery: no se aceptó dentro del plazo de espera (contención de fixture entre paquetes de prueba)")
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// --- Escenarios que no requieren una cuenta verificada dedicada ----------
// (no crean ni consumen ningún código real: pueden repetirse libremente).

func TestRecoveryRepository_RequestRecovery_UnverifiedPhone_NotAccepted(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.NewRecoveryRepository(db)

	// barbero.a (unverifiedEmailA) no tiene teléfono verificado
	// (hu007_reto_telefonico.sql lo deja así a propósito).
	accepted, phone, email, err := repo.RequestRecovery(context.Background(), unverifiedEmailA, testRecoveryCodeHash(t, "noverificado"), testRecoveryCfg())
	if err != nil {
		t.Fatalf("RequestRecovery: %v", err)
	}
	if accepted || phone != "" || email != "" {
		t.Fatalf("expected accepted=false for an unverified phone, got accepted=%v phone=%q email=%q", accepted, phone, email)
	}
}

func TestRecoveryRepository_RequestRecovery_UnknownAccount_NotAccepted(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.NewRecoveryRepository(db)

	accepted, phone, email, err := repo.RequestRecovery(context.Background(), "no-existe@ejemplo.test", testRecoveryCodeHash(t, "unknown"), testRecoveryCfg())
	if err != nil {
		t.Fatalf("RequestRecovery: %v", err)
	}
	if accepted || phone != "" || email != "" {
		t.Fatalf("expected accepted=false for an unknown account, got accepted=%v", accepted)
	}
}

func TestRecoveryRepository_VerifyRecovery_UnknownAccount_Fails(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.NewRecoveryRepository(db)

	ok, _, _, err := repo.VerifyRecovery(context.Background(), "no-existe@ejemplo.test", testRecoveryCodeHash(t, "verify-unknown"), testResetTokenHash(t, "verify-unknown"), 300)
	if err != nil {
		t.Fatalf("VerifyRecovery: %v", err)
	}
	if ok {
		t.Fatal("expected VerifyRecovery to fail for an unknown account")
	}
}

func TestRecoveryRepository_CurrentCredential_UnknownToken_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.NewRecoveryRepository(db)

	found, _, _, err := repo.CurrentCredential(context.Background(), verifiedEmailA, testResetTokenHash(t, "unknown-token"))
	if err != nil {
		t.Fatalf("CurrentCredential: %v", err)
	}
	if found {
		t.Fatal("expected found=false for an unknown reset token")
	}
}

// --- Escenario completo con duena.a (verifiedEmailA) ----------------------
// Un único recorrido secuencial que cubre solicitud, cooldown de reenvío,
// código incorrecto, código correcto, lectura de credencial, cambio de
// contraseña y revocación de sesión, y reutilización del token ya
// consumido — todo sobre el MISMO código/token para no necesitar una
// segunda cuenta verificada.

func TestRecoveryRepository_FullScenario_DuenaA(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.NewRecoveryRepository(db)
	credentialRepo := authpostgres.New(db)

	codeHash := testRecoveryCodeHash(t, "duena-a-code")

	t.Run("RequestRecovery accepted", func(t *testing.T) {
		phone, email := requestRecoveryEventually(t, repo, verifiedEmailA, codeHash, testRecoveryCfg())
		if phone == "" || email != verifiedEmailA {
			t.Fatalf("expected phone/email, got phone=%q email=%q", phone, email)
		}
	})

	t.Run("immediate resend rejected by cooldown", func(t *testing.T) {
		accepted, phone, _, err := repo.RequestRecovery(context.Background(), verifiedEmailA, testRecoveryCodeHash(t, "duena-a-resend"), testRecoveryCfg())
		if err != nil {
			t.Fatalf("RequestRecovery: %v", err)
		}
		if accepted || phone != "" {
			t.Fatalf("expected the immediate resend to be rejected by cooldown (DEC-064), got accepted=%v", accepted)
		}
	})

	t.Run("verify with wrong code fails", func(t *testing.T) {
		ok, _, _, err := repo.VerifyRecovery(context.Background(), verifiedEmailA, testRecoveryCodeHash(t, "duena-a-wrong"), testResetTokenHash(t, "duena-a-wrong-token"), 300)
		if err != nil {
			t.Fatalf("VerifyRecovery: %v", err)
		}
		if ok {
			t.Fatal("expected VerifyRecovery to fail with an incorrect code")
		}
	})

	resetHash := testResetTokenHash(t, "duena-a-reset")

	t.Run("verify with correct code succeeds", func(t *testing.T) {
		ok, phone, email, err := repo.VerifyRecovery(context.Background(), verifiedEmailA, codeHash, resetHash, 300)
		if err != nil || !ok || phone == "" || email != verifiedEmailA {
			t.Fatalf("expected ok=true with phone/email, got ok=%v phone=%q email=%q err=%v", ok, phone, email, err)
		}
	})

	t.Run("current credential returns the hash for the valid token", func(t *testing.T) {
		found, hash, alg, err := repo.CurrentCredential(context.Background(), verifiedEmailA, resetHash)
		if err != nil || !found || hash == "" || alg != "argon2id" {
			t.Fatalf("expected found=true with a hash and algorithm, got found=%v hash=%q alg=%q err=%v", found, hash, alg, err)
		}
	})

	newHash := "fixture-nuevo-hash-de-prueba-suficientemente-largo-000"

	t.Run("change password succeeds and updates the credential", func(t *testing.T) {
		changed, err := repo.ChangePassword(context.Background(), verifiedEmailA, resetHash, newHash)
		if err != nil || !changed {
			t.Fatalf("expected ChangePassword to succeed, got %v err=%v", changed, err)
		}

		cred, err := credentialRepo.LookupCredential(context.Background(), shopA, true, verifiedEmailA)
		if err != nil {
			t.Fatalf("LookupCredential: %v", err)
		}
		if !cred.Found || cred.PasswordHash != newHash {
			t.Fatalf("expected the credential to be updated to the new hash, got found=%v hash=%q", cred.Found, cred.PasswordHash)
		}
	})

	t.Run("the fixture session gets revoked (CA-008-05)", func(t *testing.T) {
		// La sesión fija de hu005_credenciales_sesiones.sql para duena.a
		// (cccccca1) debe quedar revocada por el cambio de contraseña de
		// arriba, que ya se ejecutó en el subtest anterior.
		revokedAt, _, _ := readSessionRow(t, db, shopA, "cccccca1-cccc-cccc-cccc-ccccccccccc1")
		if revokedAt == nil {
			t.Fatal("expected the fixture session to be revoked after ChangePassword")
		}
	})

	t.Run("reusing the already-consumed token fails", func(t *testing.T) {
		second, err := repo.ChangePassword(context.Background(), verifiedEmailA, resetHash, "fixture-hash-reuso-suficientemente-largo")
		if err != nil {
			t.Fatalf("ChangePassword (reuse): %v", err)
		}
		if second {
			t.Fatal("expected reusing an already-consumed reset token to fail")
		}
	})
}

// --- Escenario de concurrencia con dueno.b (verifiedEmailB) ---------------
// Cuenta separada para no interferir con el escenario secuencial de
// duena.a: aquí el propio código se solicita y verifica una sola vez, y
// luego se dispara la carrera real de ChangePassword.

func TestRecoveryRepository_ChangePassword_ConcurrentSameToken_OnlyOneWinner(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := authpostgres.NewRecoveryRepository(db)

	codeHash := testRecoveryCodeHash(t, "duenob-race-code")
	requestRecoveryEventually(t, repo, verifiedEmailB, codeHash, testRecoveryCfg())
	resetHash := testResetTokenHash(t, "duenob-race-token")
	ok, _, _, err := repo.VerifyRecovery(context.Background(), verifiedEmailB, codeHash, resetHash, 300)
	if err != nil || !ok {
		t.Fatalf("expected verify to succeed, got ok=%v err=%v", ok, err)
	}

	const n = 10
	var wg sync.WaitGroup
	results := make([]bool, n)
	errs := make([]error, n)
	start := make(chan struct{})
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			<-start
			results[i], errs[i] = repo.ChangePassword(context.Background(), verifiedEmailB, resetHash, "fixture-hash-duenob-concurrente-suficientemente-largo")
		}(i)
	}
	close(start)
	wg.Wait()

	winners := 0
	for i := 0; i < n; i++ {
		if errs[i] != nil {
			t.Fatalf("goroutine %d: unexpected error: %v", i, errs[i])
		}
		if results[i] {
			winners++
		}
	}
	if winners != 1 {
		t.Fatalf("expected exactly 1 winner among %d concurrent ChangePassword calls with the same token, got %d", n, winners)
	}
}

// El cooldown y el límite de reenvío por ventana dependen de comparar
// contra pg_catalog.now() DENTRO de la función SQL: no hay forma de
// inyectar un reloj falso desde Go (a diferencia de la capa de servicio,
// que sí usa dobles deterministas). Esos casos de frontera (ventana
// vencida, tercer/cuarto código dentro de la hora) se cubren en
// database/tests/hu008_recuperacion_acceso.sql, donde el rol administrador
// puede reescribir created_at directamente sin esperar. La prueba de
// arriba (immediate resend) ya cubre el caso determinista y rápido: el
// cooldown SÍ bloquea un reenvío inmediato.
