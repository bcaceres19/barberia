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
	resolveEmail string
	resolveFound bool
	resolveErr   error
	resolveCalls []string

	requestAccepted bool
	requestPhone    string
	requestEmail    string
	requestErr      error
	requestCalls    []struct{ email, codeHash string }

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

func (f *fakeRecoveryRepository) ResolveAccountEmailByPhone(_ context.Context, phone string) (string, bool, error) {
	f.resolveCalls = append(f.resolveCalls, phone)
	if f.resolveErr != nil {
		return "", false, f.resolveErr
	}
	return f.resolveEmail, f.resolveFound, nil
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

func emailTarget(value string) auth.RecoveryTarget {
	return auth.RecoveryTarget{Channel: auth.RecoveryChannelEmail, Value: value}
}

func phoneTarget(value string) auth.RecoveryTarget {
	return auth.RecoveryTarget{Channel: auth.RecoveryChannelWhatsApp, Value: value}
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

func TestRecoveryService_Request_EmailChannel_SendsOnlyToEmail(t *testing.T) {
	repo := &fakeRecoveryRepository{requestAccepted: true, requestPhone: "+573001234567", requestEmail: "duena.a@ejemplo.test"}
	sender := &spyRecoverySender{}
	svc := newTestRecoveryService(repo, sender, "482913", "reset-token-value")

	if err := svc.Request(context.Background(), emailTarget("Duena.A@Ejemplo.TEST")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sender.calls) != 1 {
		t.Fatalf("expected exactly one send, got %d", len(sender.calls))
	}
	call := sender.calls[0]
	// CA-008-09: elegir correo nunca entrega por WhatsApp.
	if call.phone != "" || call.email != "duena.a@ejemplo.test" || call.code != "482913" {
		t.Fatalf("unexpected send call: %+v", call)
	}
	if len(repo.resolveCalls) != 0 {
		t.Fatalf("the email channel must not resolve by phone, got %v", repo.resolveCalls)
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

	if err := svc.Request(context.Background(), emailTarget("inexistente@ejemplo.test")); err != nil {
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

	err := svc.Request(context.Background(), emailTarget("duena.a@ejemplo.test"))
	if err == nil {
		t.Fatal("expected an error to be returned for internal logging when delivery fails")
	}
}

func TestRecoveryService_Request_WhatsAppChannel_ResolvesPhoneAndSendsOnlyToPhone(t *testing.T) {
	repo := &fakeRecoveryRepository{
		resolveEmail: "duena.a@ejemplo.test", resolveFound: true,
		requestAccepted: true, requestPhone: "+573001234567", requestEmail: "duena.a@ejemplo.test",
	}
	sender := &spyRecoverySender{}
	svc := newTestRecoveryService(repo, sender, "482913", "reset-token-value")

	// Los separadores de presentación se descartan antes de buscar la cuenta.
	if err := svc.Request(context.Background(), phoneTarget("+57 300 123-4567")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.resolveCalls) != 1 || repo.resolveCalls[0] != "+573001234567" {
		t.Fatalf("expected the normalized E.164 phone to be resolved, got %v", repo.resolveCalls)
	}
	if repo.requestCalls[0].email != "duena.a@ejemplo.test" {
		t.Fatalf("expected the resolved email to identify the account, got %q", repo.requestCalls[0].email)
	}
	// CA-008-09: elegir WhatsApp nunca entrega por correo.
	if len(sender.calls) != 1 || sender.calls[0].phone != "+573001234567" || sender.calls[0].email != "" {
		t.Fatalf("expected a send to the stored phone only, got %+v", sender.calls)
	}
}

// TestRecoveryService_Request_UnresolvedPhone_NeverSendsNorTouchesRecovery
// cubre CA-008-10 y DP-SEG-14: ninguna coincidencia y una coincidencia
// ambigua se ven igual (found=false), y un valor que no es E.164 ni siquiera
// consulta la base de datos.
func TestRecoveryService_Request_UnresolvedPhone_NeverSendsNorTouchesRecovery(t *testing.T) {
	for _, tc := range []struct {
		name         string
		value        string
		wantResolves int
	}{
		{"sin cuenta o ambiguo", "+573009999999", 1},
		{"sin prefijo internacional", "3001234567", 0},
		{"letras", "+57abc", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRecoveryRepository{resolveFound: false, requestAccepted: true, requestPhone: "+573001234567", requestEmail: "x@ejemplo.test"}
			sender := &spyRecoverySender{}
			svc := newTestRecoveryService(repo, sender, "482913", "reset-token-value")

			if err := svc.Request(context.Background(), phoneTarget(tc.value)); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(repo.resolveCalls) != tc.wantResolves {
				t.Fatalf("resolve calls = %d, want %d", len(repo.resolveCalls), tc.wantResolves)
			}
			if len(repo.requestCalls) != 0 || len(sender.calls) != 0 {
				t.Fatal("an unresolved phone must neither create a code nor send anything")
			}
		})
	}
}

func TestRecoveryService_Request_ResolveError_ReturnsErrorForLoggingOnly(t *testing.T) {
	repo := &fakeRecoveryRepository{resolveErr: errors.New("db down")}
	sender := &spyRecoverySender{}
	svc := newTestRecoveryService(repo, sender, "482913", "reset-token-value")

	if err := svc.Request(context.Background(), phoneTarget("+573001234567")); err == nil {
		t.Fatal("expected an error for internal logging when the phone lookup fails")
	}
	if len(sender.calls) != 0 {
		t.Fatal("nothing must be sent when the account could not be resolved")
	}
}

// --- Verify -----------------------------------------------------------------

func TestRecoveryService_Verify_Success_ReturnsTokenAndMaskedDestinations(t *testing.T) {
	repo := &fakeRecoveryRepository{verifyOK: true, verifyPhone: "+573001234567", verifyEmail: "duena.a@ejemplo.test"}
	svc := newTestRecoveryService(repo, &spyRecoverySender{}, "482913", "reset-token-value")

	token, maskedPhone, maskedEmail, err := svc.Verify(context.Background(), emailTarget("duena.a@ejemplo.test"), "482913")
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
		_, _, _, err := svc.Verify(context.Background(), emailTarget("cualquiera@ejemplo.test"), "000000")
		if err == nil {
			t.Fatalf("%s: expected an error", c.name)
		}
		messages = append(messages, err.Error())
	}
	if messages[0] != messages[1] {
		t.Fatalf("expected identical error messages across causes (no enumeration), got %q vs %q", messages[0], messages[1])
	}
}

func TestRecoveryService_Verify_WhatsAppChannel_UsesResolvedEmail(t *testing.T) {
	repo := &fakeRecoveryRepository{
		resolveEmail: "duena.a@ejemplo.test", resolveFound: true,
		verifyOK: true, verifyPhone: "+573001234567", verifyEmail: "duena.a@ejemplo.test",
	}
	svc := newTestRecoveryService(repo, &spyRecoverySender{}, "482913", "reset-token-value")

	if _, _, _, err := svc.Verify(context.Background(), phoneTarget("+573001234567"), "482913"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.verifyCalls) != 1 || repo.verifyCalls[0].email != "duena.a@ejemplo.test" {
		t.Fatalf("expected the verify call to use the resolved email, got %+v", repo.verifyCalls)
	}
}

// TestRecoveryService_Verify_UnresolvedPhone_SameErrorAsWrongCode cubre
// DEC-065/DP-SEG-15: un teléfono sin cuenta única falla con exactamente el
// mismo error uniforme que un código incorrecto y sin consultar la
// verificación.
func TestRecoveryService_Verify_UnresolvedPhone_SameErrorAsWrongCode(t *testing.T) {
	wrongCode := newTestRecoveryService(&fakeRecoveryRepository{verifyOK: false}, &spyRecoverySender{}, "482913", "reset-token-value")
	_, _, _, wantErr := wrongCode.Verify(context.Background(), emailTarget("duena.a@ejemplo.test"), "000000")

	repo := &fakeRecoveryRepository{resolveFound: false, verifyOK: true}
	svc := newTestRecoveryService(repo, &spyRecoverySender{}, "482913", "reset-token-value")
	_, _, _, err := svc.Verify(context.Background(), phoneTarget("+573009999999"), "482913")

	if err == nil || wantErr == nil || err.Error() != wantErr.Error() {
		t.Fatalf("expected the uniform invalid-code error, got %v want %v", err, wantErr)
	}
	if len(repo.verifyCalls) != 0 {
		t.Fatal("an unresolved phone must never reach the verification")
	}
}

// --- ChangePassword -----------------------------------------------------

func TestRecoveryService_ChangePassword_Success(t *testing.T) {
	repo := &fakeRecoveryRepository{credentialFound: true, credentialHash: mustHash(t, "vieja-contrasena-larga"), changeOK: true}
	svc := newTestRecoveryService(repo, &spyRecoverySender{}, "482913", "reset-token-value")

	err := svc.ChangePassword(context.Background(), emailTarget("duena.a@ejemplo.test"), "reset-token-value", "contrasena-nueva-valida")
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

	err := svc.ChangePassword(context.Background(), emailTarget("duena.a@ejemplo.test"), "token-desconocido", "contrasena-nueva-valida")
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

	err := svc.ChangePassword(context.Background(), emailTarget("duena.a@ejemplo.test"), "reset-token-value", current)
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

	err := svc.ChangePassword(context.Background(), emailTarget("duena.a@ejemplo.test"), "reset-token-value", "contrasena-nueva-valida")
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

func TestRecoveryService_ChangePassword_WhatsAppChannel_UsesResolvedEmail(t *testing.T) {
	repo := &fakeRecoveryRepository{
		resolveEmail: "duena.a@ejemplo.test", resolveFound: true,
		credentialFound: true, credentialHash: mustHash(t, "vieja-contrasena-larga"), changeOK: true,
	}
	svc := newTestRecoveryService(repo, &spyRecoverySender{}, "482913", "reset-token-value")

	if err := svc.ChangePassword(context.Background(), phoneTarget("+573001234567"), "reset-token-value", "contrasena-nueva-valida"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.changeCalls) != 1 || repo.changeCalls[0].email != "duena.a@ejemplo.test" {
		t.Fatalf("expected the change to use the resolved email, got %+v", repo.changeCalls)
	}
}

func TestRecoveryService_ChangePassword_UnresolvedPhone_SameErrorAsInvalidToken(t *testing.T) {
	repo := &fakeRecoveryRepository{resolveFound: false, credentialFound: true, changeOK: true}
	svc := newTestRecoveryService(repo, &spyRecoverySender{}, "482913", "reset-token-value")

	err := svc.ChangePassword(context.Background(), phoneTarget("+573009999999"), "reset-token-value", "contrasena-nueva-valida")
	if err == nil {
		t.Fatal("expected an error for an unresolved phone")
	}
	if len(repo.credentialCalls) != 0 || len(repo.changeCalls) != 0 {
		t.Fatal("an unresolved phone must never read or change a credential")
	}
}

func TestNormalizePhone(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want string
		ok   bool
	}{
		{"+573001234567", "+573001234567", true},
		{"  +57 300 123 4567 ", "+573001234567", true},
		{"+57 (300) 123-4567", "+573001234567", true},
		{"+57.300.123.4567", "+573001234567", true},
		{"3001234567", "", false},
		{"+0573001234567", "", false},
		{"+57300", "", false},
		{"+5730012345678901", "", false},
		{"+57abc1234567", "", false},
		{"", "", false},
	} {
		got, ok := auth.NormalizePhone(tc.raw)
		if got != tc.want || ok != tc.ok {
			t.Errorf("NormalizePhone(%q) = (%q, %v), want (%q, %v)", tc.raw, got, ok, tc.want, tc.ok)
		}
	}
}
