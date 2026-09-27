package auth_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/platform/apperr"
)

// providerMarker imita la marca opaca de un proveedor que administra el OTP
// (Twilio Verify): 64 hex sin relación con el código.
const providerMarker = "a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1"

// fakeManagedOTPProvider reemplaza a un proveedor externo sin tráfico real.
type fakeManagedOTPProvider struct {
	verifyErr error
	delivered []struct{ phone, code string }
	checked   []struct{ phone, code string }
}

func (p *fakeManagedOTPProvider) Prepare(context.Context) (auth.PreparedWhatsAppOTP, error) {
	return auth.NewProviderManagedWhatsAppOTP(providerMarker), nil
}

func (p *fakeManagedOTPProvider) Deliver(_ context.Context, phone string, prepared auth.PreparedWhatsAppOTP) error {
	p.delivered = append(p.delivered, struct{ phone, code string }{phone, prepared.Code()})
	return nil
}

func (p *fakeManagedOTPProvider) VerificationDigest(_ context.Context, phone, code string) (string, error) {
	p.checked = append(p.checked, struct{ phone, code string }{phone, code})
	if p.verifyErr != nil {
		return "", p.verifyErr
	}
	return providerMarker, nil
}

func newManagedPhoneService(repo *fakePhoneChallengeRepository, provider auth.WhatsAppOTPProvider) *auth.PhoneChallengeService {
	return auth.NewPhoneChallengeServiceWithOTPProvider(repo, provider, testChallengeCfg())
}

func TestPhoneChallengeService_ManagedProvider_Request_PersistsMarkerAndNeverACode(t *testing.T) {
	repo := &fakePhoneChallengeRepository{requestAccepted: true, requestPhone: "+573001234567"}
	provider := &fakeManagedOTPProvider{}

	if err := newManagedPhoneService(repo, provider).Request(context.Background(), "barbero@ejemplo.test", "ip-hash"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := repo.requestCalls[0].codeHash; got != providerMarker {
		t.Fatalf("expected the provider marker to be persisted, got %q", got)
	}
	if len(provider.delivered) != 1 || provider.delivered[0].code != "" {
		t.Fatalf("a provider-managed OTP must carry no code from the application, got %+v", provider.delivered)
	}
}

func TestPhoneChallengeService_ManagedProvider_Verify_ApprovedConsumesLocalChallenge(t *testing.T) {
	repo := &fakePhoneChallengeRepository{requestPhone: "+573001234567", verifyOK: true}
	provider := &fakeManagedOTPProvider{}

	if err := newManagedPhoneService(repo, provider).Verify(context.Background(), "Barbero@Ejemplo.TEST", "ip-hash", "482913"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(provider.checked) != 1 || provider.checked[0].phone != "+573001234567" || provider.checked[0].code != "482913" {
		t.Fatalf("expected the provider to validate the stored destination and typed code, got %+v", provider.checked)
	}
	if repo.verifyCall.codeHash != providerMarker {
		t.Fatalf("expected the local challenge to be consumed with the provider marker, got %q", repo.verifyCall.codeHash)
	}
}

// Un rechazo del proveedor debe seguir contando como intento fallido local:
// max_attempts es una protección de la aplicación, no del proveedor.
func TestPhoneChallengeService_ManagedProvider_Verify_RejectedCountsLocalAttempt(t *testing.T) {
	repo := &fakePhoneChallengeRepository{requestPhone: "+573001234567"}
	provider := &fakeManagedOTPProvider{verifyErr: auth.ErrWhatsAppOTPRejected}

	err := newManagedPhoneService(repo, provider).Verify(context.Background(), "barbero@ejemplo.test", "ip-hash", "000000")

	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindUnauthorized {
		t.Fatalf("expected KindUnauthorized, got %v (%T)", err, err)
	}
	got := repo.verifyCall.codeHash
	if len(got) != 64 || got == providerMarker || strings.Trim(got, "0") != "" {
		t.Fatalf("expected a never-matching 64-hex digest to advance the attempt counter, got %q", got)
	}
}

func TestPhoneChallengeService_ManagedProvider_Verify_NoActiveChallengeNeverCallsProvider(t *testing.T) {
	repo := &fakePhoneChallengeRepository{}
	provider := &fakeManagedOTPProvider{}

	err := newManagedPhoneService(repo, provider).Verify(context.Background(), "barbero@ejemplo.test", "ip-hash", "482913")

	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindUnauthorized {
		t.Fatalf("expected KindUnauthorized, got %v (%T)", err, err)
	}
	if len(provider.checked) != 0 {
		t.Fatal("expected no provider call without an active challenge")
	}
}

func TestPhoneChallengeService_ManagedProvider_Verify_ProviderOutageIsInternalAndDoesNotCountAttempt(t *testing.T) {
	repo := &fakePhoneChallengeRepository{requestPhone: "+573001234567"}
	provider := &fakeManagedOTPProvider{verifyErr: errors.New("provider unavailable")}

	err := newManagedPhoneService(repo, provider).Verify(context.Background(), "barbero@ejemplo.test", "ip-hash", "482913")

	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected KindInternal, got %v (%T)", err, err)
	}
	if repo.verifyCall.codeHash != "" {
		t.Fatal("a provider outage must not burn one of the user's attempts")
	}
}

func newManagedRecoveryService(repo *fakeRecoveryRepository, provider auth.WhatsAppOTPProvider, sender *spyRecoverySender) *auth.RecoveryService {
	return auth.NewRecoveryServiceWithWhatsAppOTPProvider(
		repo,
		fixedCodeGenerator{code: "482913"},
		fixedTokenGenerator{token: "reset-token-value"},
		sender,
		provider,
		auth.NewArgon2Hasher(),
		auth.RecoveryConfig{
			CodeExpiresSeconds: 600, CodeMaxAttempts: 5, ResendCooldownSeconds: 60,
			ResendWindowSeconds: 3600, ResendMaxPerWindow: 3, ResetTokenExpiresSeconds: 300,
		},
		[]byte(testRecoverySecret),
	)
}

func TestRecoveryService_ManagedProvider_Request_PhoneUsesProviderAndEmailKeepsSender(t *testing.T) {
	repo := &fakeRecoveryRepository{
		resolveEmail: "duena.a@ejemplo.test", resolveFound: true,
		requestAccepted: true, requestPhone: "+573001234567", requestEmail: "duena.a@ejemplo.test",
	}
	provider := &fakeManagedOTPProvider{}
	sender := &spyRecoverySender{}
	svc := newManagedRecoveryService(repo, provider, sender)

	if err := svc.Request(context.Background(), phoneTarget("+573001234567")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.requestCalls[0].codeHash != providerMarker {
		t.Fatalf("expected the provider marker to be persisted, got %q", repo.requestCalls[0].codeHash)
	}
	if len(provider.delivered) != 1 || provider.delivered[0].phone != "+573001234567" || provider.delivered[0].code != "" {
		t.Fatalf("expected a provider delivery to the stored phone with no app code, got %+v", provider.delivered)
	}
	if len(sender.calls) != 0 {
		t.Fatalf("the phone channel must never reach the email sender, got %+v", sender.calls)
	}

	if err := svc.Request(context.Background(), emailTarget("duena.a@ejemplo.test")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(provider.delivered) != 1 {
		t.Fatal("the email channel must never reach the phone provider")
	}
	if len(sender.calls) != 1 || sender.calls[0].email != "duena.a@ejemplo.test" || sender.calls[0].code != "482913" {
		t.Fatalf("expected the email sender to keep the app-generated code, got %+v", sender.calls)
	}
}

func TestRecoveryService_ManagedProvider_Verify_RejectedCountsLocalAttempt(t *testing.T) {
	repo := &fakeRecoveryRepository{resolveEmail: "duena.a@ejemplo.test", resolveFound: true}
	provider := &fakeManagedOTPProvider{verifyErr: auth.ErrWhatsAppOTPRejected}
	svc := newManagedRecoveryService(repo, provider, &spyRecoverySender{})

	_, _, _, err := svc.Verify(context.Background(), phoneTarget("+573001234567"), "000000")

	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindUnauthorized {
		t.Fatalf("expected KindUnauthorized, got %v (%T)", err, err)
	}
	if len(repo.verifyCalls) != 1 {
		t.Fatalf("expected exactly one local attempt to be recorded, got %+v", repo.verifyCalls)
	}
	got := repo.verifyCalls[0].codeHash
	if len(got) != 64 || got == providerMarker || strings.Trim(got, "0") != "" {
		t.Fatalf("expected a never-matching 64-hex digest, got %q", got)
	}
}

func TestRecoveryService_ManagedProvider_Verify_ApprovedIssuesToken(t *testing.T) {
	repo := &fakeRecoveryRepository{
		resolveEmail: "duena.a@ejemplo.test", resolveFound: true,
		verifyOK: true, verifyPhone: "+573001234567", verifyEmail: "duena.a@ejemplo.test",
	}
	provider := &fakeManagedOTPProvider{}
	svc := newManagedRecoveryService(repo, provider, &spyRecoverySender{})

	token, _, _, err := svc.Verify(context.Background(), phoneTarget("+573001234567"), "482913")
	if err != nil || token != "reset-token-value" {
		t.Fatalf("expected a reset token, got %q, %v", token, err)
	}
	if repo.verifyCalls[0].codeHash != providerMarker {
		t.Fatalf("expected the provider marker as the verification digest, got %q", repo.verifyCalls[0].codeHash)
	}
}

func TestRecoveryService_ManagedProvider_Verify_ProviderOutageIsInternal(t *testing.T) {
	repo := &fakeRecoveryRepository{resolveEmail: "duena.a@ejemplo.test", resolveFound: true}
	provider := &fakeManagedOTPProvider{verifyErr: errors.New("provider unavailable")}
	svc := newManagedRecoveryService(repo, provider, &spyRecoverySender{})

	_, _, _, err := svc.Verify(context.Background(), phoneTarget("+573001234567"), "482913")

	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected KindInternal, got %v (%T)", err, err)
	}
	if len(repo.verifyCalls) != 0 {
		t.Fatal("a provider outage must not burn one of the user's attempts")
	}
}
