package auth_test

import (
	"context"
	"errors"
	"testing"

	"system-barbershop/internal/modules/auth"
)

// --- dobles de prueba ---------------------------------------------------

// fixedCodeGenerator ya está declarado en phone_challenge_test.go; se
// reutiliza aquí sin duplicarlo.

type fixedTokenGenerator struct{ token string }

func (g fixedTokenGenerator) New() (string, error) { return g.token, nil }

// spyRecoverySender registra CADA llamada con sus argumentos exactos
// (CA-008-01 exige poder afirmar que el envío real ocurrió con el
// contenido correcto, no solo que "no falló").
type spyRecoverySender struct {
	calls []recoverySendCall
	err   error
}

type recoverySendCall struct{ phone, email, code string }

func (s *spyRecoverySender) SendCode(_ context.Context, phone, email, code string) error {
	s.calls = append(s.calls, recoverySendCall{phone, email, code})
	return s.err
}

type fakeRecoveryRepository struct {
	requestAccepted   bool
	requestPhone      string
	requestEmail      string
	requestErr        error
	requestCalls      []struct{ email, codeHash string }

	verifyOK    bool
	verifyPhone string
	verifyEmail string
	verifyErr   error
	verifyCalls []struct{ email, codeHash, resetTokenHash string }

	credentialFound bool
	credentialHash  string
	credentialAlg   string
	credentialErr   error
	credentialCalls []struct{ email, resetTokenHash string }

	changeOK    bool
	changeErr   error
	changeCalls []struct{ email, resetTokenHash, newHash string }
}

func (f *fakeRecoveryRepository) RequestRecovery(_ context.Context, email, codeHash string, _ auth.RecoveryConfig) (bool, string, string, error) {
	f.requestCalls = append(f.requestCalls, struct{ email, codeHash string }{email, codeHash})
	if f.requestErr != nil {
		return false, "", "", f.requestErr
	}
	return f.requestAccepted, f.requestPhone, f.requestEmail, nil
}

func (f *fakeRecoveryRepository) VerifyRecovery(_ context.Context, email, codeHash, resetTokenHash string, _ int) (bool, string, string, error) {
	f.verifyCalls = append(f.verifyCalls, struct{ email, codeHash, resetTokenHash string }{email, codeHash, resetTokenHash})
	if f.verifyErr != nil {
		return false, "", "", f.verifyErr
	}
	return f.verifyOK, f.verifyPhone, f.verifyEmail, nil
}

func (f *fakeRecoveryRepository) CurrentCredential(_ context.Context, email, resetTokenHash string) (bool, string, string, error) {
	f.credentialCalls = append(f.credentialCalls, struct{ email, resetTokenHash string }{email, resetTokenHash})
	if f.credentialErr != nil {
		return false, "", "", f.credentialErr
	}
	return f.credentialFound, f.credentialHash, f.credentialAlg, nil
}

func (f *fakeRecoveryRepository) ChangePassword(_ context.Context, email, resetTokenHash, newHash string) (bool, error) {
	f.changeCalls = append(f.changeCalls, struct{ email, resetTokenHash, newHash string }{email, resetTokenHash, newHash})
	if f.changeErr != nil {
		return false, f.changeErr
	}
	return f.changeOK, nil
}

const testRecoverySecret = "prueba-recovery-no-es-un-secreto-real-0123456789"

func newTestRecoveryService(repo *fakeRecoveryRepository, sender *spyRecoverySender, code, token string) *auth.RecoveryService {
	return auth.NewRecoveryService(
		repo,
		fixedCodeGenerator{code: code},
		fixedTokenGenerator{token: token},
		sender,
		auth.NewArgon2Hasher(),
		auth.RecoveryConfig{
			CodeExpiresSeconds:       900,
			CodeMaxAttempts:          5,
			ResendCooldownSeconds:    60,
			ResendWindowSeconds:      3600,
			ResendMaxPerWindow:       3,
			ResetTokenExpiresSeconds: 300,
		},
		[]byte(testRecoverySecret),
	)
}

// --- Request --------------------------------------------------------------

func TestRecoveryService_Request_Accepted_SendsExactCodeToBothChannels(t *testing.T) {
	repo := &fakeRecoveryRepository{requestAccepted: true, requestPhone: "+573001234567", requestEmail: "duena.a@ejemplo.test"}
	sender := &spyRecoverySender{}
	svc := newTestRecoveryService(repo, sender, "482913", "reset-token-value")

	if err := svc.Request(context.Background(), "Duena.A@Ejemplo.TEST"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sender.calls) != 1 {
		t.Fatalf("expected exactly one send, got %d", len(sender.calls))
	}
	call := sender.calls[0]
	if call.phone != "+573001234567" || call.email != "duena.a@ejemplo.test" || call.code != "482913" {
		t.Fatalf("unexpected send call: %+v", call)
	}
	if repo.requestCalls[0].email != "duena.a@ejemplo.test" {
		t.Fatalf("expected normalized email passed to repository, got %q", repo.requestCalls[0].email)
	}
	if len(repo.requestCalls[0].codeHash) != 64 {
		t.Fatalf("expected a 64-hex HMAC code hash, got %d chars", len(repo.requestCalls[0].codeHash))
	}
}

func TestRecoveryService_Request_NotAccepted_NeverSends(t *testing.T) {
	repo := &fakeRecoveryRepository{requestAccepted: false}
	sender := &spyRecoverySender{}
	svc := newTestRecoveryService(repo, sender, "482913", "reset-token-value")

	if err := svc.Request(context.Background(), "inexistente@ejemplo.test"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sender.calls) != 0 {
		t.Fatal("expected no send when the repository does not accept the request")
	}
}

func TestRecoveryService_Request_SendFailure_ReturnsErrorForLoggingOnly(t *testing.T) {
	repo := &fakeRecoveryRepository{requestAccepted: true, requestPhone: "+573001234567", requestEmail: "duena.a@ejemplo.test"}
	sender := &spyRecoverySender{err: errors.New("meta unavailable")}
	svc := newTestRecoveryService(repo, sender, "482913", "reset-token-value")

	err := svc.Request(context.Background(), "duena.a@ejemplo.test")
	if err == nil {
		t.Fatal("expected an error to be returned for internal logging when delivery fails")
	}
}

// --- Verify -----------------------------------------------------------------

func TestRecoveryService_Verify_Success_ReturnsTokenAndMaskedDestinations(t *testing.T) {
	repo := &fakeRecoveryRepository{verifyOK: true, verifyPhone: "+573001234567", verifyEmail: "duena.a@ejemplo.test"}
	svc := newTestRecoveryService(repo, &spyRecoverySender{}, "482913", "reset-token-value")

	token, maskedPhone, maskedEmail, err := svc.Verify(context.Background(), "duena.a@ejemplo.test", "482913")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "reset-token-value" {
		t.Fatalf("expected the raw reset token to be returned, got %q", token)
	}
	if maskedPhone != "+57 *** *** 67" {
		t.Fatalf("unexpected masked phone: %q", maskedPhone)
	}
	if maskedEmail != "d***@e***.test" {
		t.Fatalf("unexpected masked email: %q", maskedEmail)
	}
	if len(repo.verifyCalls[0].resetTokenHash) != 64 {
		t.Fatalf("expected a 64-hex SHA-256 reset token hash passed to the repository")
	}
}

func TestRecoveryService_Verify_Failure_ReturnsUniformError(t *testing.T) {
	cases := []struct {
		name string
		repo *fakeRecoveryRepository
	}{
		{"código incorrecto", &fakeRecoveryRepository{verifyOK: false}},
		{"cuenta inexistente", &fakeRecoveryRepository{verifyOK: false}},
	}
	var messages []string
	for _, c := range cases {
		svc := newTestRecoveryService(c.repo, &spyRecoverySender{}, "482913", "reset-token-value")
		_, _, _, err := svc.Verify(context.Background(), "cualquiera@ejemplo.test", "000000")
		if err == nil {
			t.Fatalf("%s: expected an error", c.name)
		}
		messages = append(messages, err.Error())
	}
	if messages[0] != messages[1] {
		t.Fatalf("expected identical error messages across causes (no enumeration), got %q vs %q", messages[0], messages[1])
	}
}

// --- ChangePassword -----------------------------------------------------

func TestRecoveryService_ChangePassword_Success(t *testing.T) {
	repo := &fakeRecoveryRepository{credentialFound: true, credentialHash: mustHash(t, "vieja-contrasena-larga"), changeOK: true}
	svc := newTestRecoveryService(repo, &spyRecoverySender{}, "482913", "reset-token-value")

	err := svc.ChangePassword(context.Background(), "duena.a@ejemplo.test", "reset-token-value", "contrasena-nueva-valida")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.changeCalls) != 1 {
		t.Fatal("expected ChangePassword to be called exactly once")
	}
	if repo.changeCalls[0].newHash == "" || repo.changeCalls[0].newHash == "contrasena-nueva-valida" {
		t.Fatal("expected a hashed password to be persisted, never the plain value")
	}
}

func TestRecoveryService_ChangePassword_InvalidToken_NeverHashesOrWrites(t *testing.T) {
	repo := &fakeRecoveryRepository{credentialFound: false}
	svc := newTestRecoveryService(repo, &spyRecoverySender{}, "482913", "reset-token-value")

	err := svc.ChangePassword(context.Background(), "duena.a@ejemplo.test", "token-desconocido", "contrasena-nueva-valida")
	if err == nil {
		t.Fatal("expected an error for an unknown reset token")
	}
	if len(repo.changeCalls) != 0 {
		t.Fatal("expected ChangePassword never to be called when the token was not found")
	}
}

func TestRecoveryService_ChangePassword_SameAsCurrent_Rejected(t *testing.T) {
	current := "misma-contrasena-actual-larga"
	repo := &fakeRecoveryRepository{credentialFound: true, credentialHash: mustHash(t, current)}
	svc := newTestRecoveryService(repo, &spyRecoverySender{}, "482913", "reset-token-value")

	err := svc.ChangePassword(context.Background(), "duena.a@ejemplo.test", "reset-token-value", current)
	if err == nil {
		t.Fatal("expected rejection when the new password equals the current one")
	}
	if len(repo.changeCalls) != 0 {
		t.Fatal("expected ChangePassword never to be called when the policy rejects the password")
	}
}

func TestRecoveryService_ChangePassword_RaceLosesToken_ReturnsUniformError(t *testing.T) {
	repo := &fakeRecoveryRepository{credentialFound: true, credentialHash: mustHash(t, "vieja-contrasena-larga"), changeOK: false}
	svc := newTestRecoveryService(repo, &spyRecoverySender{}, "482913", "reset-token-value")

	err := svc.ChangePassword(context.Background(), "duena.a@ejemplo.test", "reset-token-value", "contrasena-nueva-valida")
	if err == nil {
		t.Fatal("expected an error when the repository reports the token was already consumed")
	}
}

func mustHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := auth.NewArgon2Hasher().Hash(password)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	return hash
}

// --- ValidateNewPassword (DEC-063) --------------------------------------

func TestValidateNewPassword(t *testing.T) {
	hasher := auth.NewArgon2Hasher()
	currentHash := mustHash(t, "contrasena-actual-larga")

	cases := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"demasiado corta", "corta1234", true},
		{"exactamente el minimo", "1234567890", false},
		{"demasiado larga", string(make([]byte, 129)), true},
		{"igual al correo", "duena.a@ejemplo.test", true},
		{"igual a la actual", "contrasena-actual-larga", true},
		{"valida", "contrasena-completamente-nueva", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := auth.ValidateNewPassword(c.password, "duena.a@ejemplo.test", currentHash, hasher)
			if (err != nil) != c.wantErr {
				t.Fatalf("ValidateNewPassword(%q) error=%v, wantErr=%v", c.password, err, c.wantErr)
			}
		})
	}
}
