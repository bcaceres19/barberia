package auth_test

import (
	"context"
	"testing"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/platform/apperr"
)

// fakeOTPProvider reemplaza al transporte de WhatsApp sin tráfico real. El
// digest es determinista para comprobar qué valor llega al repositorio.
type fakeOTPProvider struct {
	delivered []struct{ phone, code string }
}

func (p *fakeOTPProvider) Prepare(context.Context) (auth.PreparedWhatsAppOTP, error) {
	return auth.NewLocalWhatsAppOTP("digest-of-482913", "482913"), nil
}

func (p *fakeOTPProvider) Deliver(_ context.Context, phone string, prepared auth.PreparedWhatsAppOTP) error {
	p.delivered = append(p.delivered, struct{ phone, code string }{phone, prepared.Code()})
	return nil
}

func (p *fakeOTPProvider) VerificationDigest(code string) string { return "digest-of-" + code }

func newProviderPhoneService(repo *fakePhoneChallengeRepository, provider auth.WhatsAppOTPProvider) *auth.PhoneChallengeService {
	return auth.NewPhoneChallengeServiceWithOTPProvider(repo, provider, testChallengeCfg())
}

func TestPhoneChallengeService_Provider_Request_PersistsDigestAndDeliversTheCode(t *testing.T) {
	repo := &fakePhoneChallengeRepository{requestAccepted: true, requestPhone: "+573001234567"}
	provider := &fakeOTPProvider{}

	if err := newProviderPhoneService(repo, provider).Request(context.Background(), "barbero@ejemplo.test", "ip-hash"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := repo.requestCalls[0].codeHash; got != "digest-of-482913" {
		t.Fatalf("expected the provider digest to be persisted, got %q", got)
	}
	if len(provider.delivered) != 1 || provider.delivered[0].phone != "+573001234567" || provider.delivered[0].code != "482913" {
		t.Fatalf("expected the NAVA code to be delivered to the stored phone, got %+v", provider.delivered)
	}
}

func TestPhoneChallengeService_Provider_Verify_ConsumesChallengeWithTheTypedCodeDigest(t *testing.T) {
	repo := &fakePhoneChallengeRepository{verifyOK: true}

	if err := newProviderPhoneService(repo, &fakeOTPProvider{}).Verify(context.Background(), "Barbero@Ejemplo.TEST", "ip-hash", "482913"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.verifyCall.codeHash != "digest-of-482913" {
		t.Fatalf("expected the typed code digest to reach the repository, got %q", repo.verifyCall.codeHash)
	}
	if repo.verifyCall.email != "barbero@ejemplo.test" {
		t.Fatalf("expected a normalized email, got %q", repo.verifyCall.email)
	}
}

func TestPhoneChallengeService_Provider_Verify_RejectedCodeIsUniformUnauthorized(t *testing.T) {
	repo := &fakePhoneChallengeRepository{}

	err := newProviderPhoneService(repo, &fakeOTPProvider{}).Verify(context.Background(), "barbero@ejemplo.test", "ip-hash", "000000")

	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindUnauthorized {
		t.Fatalf("expected KindUnauthorized, got %v (%T)", err, err)
	}
}

func newProviderRecoveryService(repo *fakeRecoveryRepository, provider auth.WhatsAppOTPProvider, sender *spyRecoverySender) *auth.RecoveryService {
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

func TestRecoveryService_Provider_Request_PhoneUsesProviderAndEmailKeepsSender(t *testing.T) {
	repo := &fakeRecoveryRepository{
		resolveEmail: "duena.a@ejemplo.test", resolveFound: true,
		requestAccepted: true, requestPhone: "+573001234567", requestEmail: "duena.a@ejemplo.test",
	}
	provider := &fakeOTPProvider{}
	sender := &spyRecoverySender{}
	svc := newProviderRecoveryService(repo, provider, sender)

	if err := svc.Request(context.Background(), phoneTarget("+573001234567")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.requestCalls[0].codeHash != "digest-of-482913" {
		t.Fatalf("expected the provider digest to be persisted, got %q", repo.requestCalls[0].codeHash)
	}
	if len(provider.delivered) != 1 || provider.delivered[0].phone != "+573001234567" || provider.delivered[0].code != "482913" {
		t.Fatalf("expected a delivery of the NAVA code to the stored phone, got %+v", provider.delivered)
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

func TestRecoveryService_Provider_Verify_PhoneIssuesTokenWithTheTypedCodeDigest(t *testing.T) {
	repo := &fakeRecoveryRepository{
		resolveEmail: "duena.a@ejemplo.test", resolveFound: true,
		verifyOK: true, verifyPhone: "+573001234567", verifyEmail: "duena.a@ejemplo.test",
	}
	svc := newProviderRecoveryService(repo, &fakeOTPProvider{}, &spyRecoverySender{})

	token, _, _, err := svc.Verify(context.Background(), phoneTarget("+573001234567"), "482913")
	if err != nil || token != "reset-token-value" {
		t.Fatalf("expected a reset token, got %q, %v", token, err)
	}
	if repo.verifyCalls[0].codeHash != "digest-of-482913" {
		t.Fatalf("expected the typed code digest to reach the repository, got %q", repo.verifyCalls[0].codeHash)
	}
}

func TestRecoveryService_Provider_Verify_RejectedCodeIsUniformUnauthorized(t *testing.T) {
	repo := &fakeRecoveryRepository{resolveEmail: "duena.a@ejemplo.test", resolveFound: true}
	svc := newProviderRecoveryService(repo, &fakeOTPProvider{}, &spyRecoverySender{})

	_, _, _, err := svc.Verify(context.Background(), phoneTarget("+573001234567"), "000000")

	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindUnauthorized {
		t.Fatalf("expected KindUnauthorized, got %v (%T)", err, err)
	}
}
