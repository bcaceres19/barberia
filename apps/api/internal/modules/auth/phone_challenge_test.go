package auth_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/platform/apperr"
)

type requestChallengeCall struct {
	email, ipHash, codeHash string
	cfg                     auth.PhoneChallengeConfig
}

type fakePhoneChallengeRepository struct {
	requestAccepted bool
	requestPhone    string
	requestEmail    string
	requestErr      error
	requestCalls    []requestChallengeCall

	verifyOK   bool
	verifyErr  error
	verifyCall struct{ email, ipHash, codeHash string }
}

func (f *fakePhoneChallengeRepository) RequestChallenge(_ context.Context, email, ipHash, codeHash string, cfg auth.PhoneChallengeConfig) (bool, auth.ChallengeDestinations, error) {
	f.requestCalls = append(f.requestCalls, requestChallengeCall{email, ipHash, codeHash, cfg})
	return f.requestAccepted, auth.ChallengeDestinations{Phone: f.requestPhone, Email: f.requestEmail}, f.requestErr
}

func (f *fakePhoneChallengeRepository) VerifyChallenge(_ context.Context, email, ipHash, codeHash string) (bool, error) {
	f.verifyCall = struct{ email, ipHash, codeHash string }{email, ipHash, codeHash}
	return f.verifyOK, f.verifyErr
}

type fixedCodeGenerator struct{ code string }

func (g fixedCodeGenerator) New() (string, error) { return g.code, nil }

type failingCodeGenerator struct{ err error }

func (g failingCodeGenerator) New() (string, error) { return "", g.err }

type sendCodeCall struct{ phone, code string }

type fakeSender struct {
	err   error
	calls []sendCodeCall
}

func (s *fakeSender) SendCode(_ context.Context, phone, code string) error {
	s.calls = append(s.calls, sendCodeCall{phone, code})
	return s.err
}

func testChallengeCfg() auth.PhoneChallengeConfig {
	return auth.PhoneChallengeConfig{ExpiresSeconds: 300, RateWindowSeconds: 900, RateMaxActive: 3, ResendCooldownSeconds: 60}
}

const testHMACSecret = "secreto-de-prueba-suficientemente-largo-0123456789"

func TestPhoneChallengeService_Request_Accepted_SendsCode(t *testing.T) {
	repo := &fakePhoneChallengeRepository{requestAccepted: true, requestPhone: "+573001234567"}
	sender := &fakeSender{}
	svc := auth.NewPhoneChallengeService(repo, fixedCodeGenerator{code: "482913"}, sender, testChallengeCfg(), []byte(testHMACSecret))

	if err := svc.Request(context.Background(), "  Barbero@Ejemplo.TEST  ", "ip-hash-de-prueba"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sender.calls) != 1 {
		t.Fatalf("expected exactly one SendCode call, got %d", len(sender.calls))
	}
	if sender.calls[0].phone != "+573001234567" || sender.calls[0].code != "482913" {
		t.Fatalf("unexpected send call: %+v", sender.calls[0])
	}
	if len(repo.requestCalls) != 1 {
		t.Fatalf("expected exactly one RequestChallenge call, got %d", len(repo.requestCalls))
	}
	if repo.requestCalls[0].email != "barbero@ejemplo.test" {
		t.Fatalf("expected normalized email, got %q", repo.requestCalls[0].email)
	}
	if len(repo.requestCalls[0].codeHash) != 64 {
		t.Fatalf("expected a 64-char hex HMAC code hash, got %q", repo.requestCalls[0].codeHash)
	}
	if repo.requestCalls[0].codeHash == "482913" {
		t.Fatal("the raw code must never be passed to the repository as-is")
	}
}

func TestPhoneChallengeService_Request_NotAccepted_NeverSends(t *testing.T) {
	repo := &fakePhoneChallengeRepository{requestAccepted: false, requestPhone: ""}
	sender := &fakeSender{}
	svc := auth.NewPhoneChallengeService(repo, fixedCodeGenerator{code: "482913"}, sender, testChallengeCfg(), []byte(testHMACSecret))

	if err := svc.Request(context.Background(), "no-existe@ejemplo.test", "ip-hash-de-prueba"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sender.calls) != 0 {
		t.Fatal("expected no SendCode call when the repository does not accept the request")
	}
}

// TestPhoneChallengeService_Request_AcceptedButEmptyPhone_NeverSends cubre
// una inconsistencia defensiva: accepted=true con phone vacío nunca debe
// intentar un envío (evita un SendCode con destino vacío).
func TestPhoneChallengeService_Request_AcceptedButEmptyPhone_NeverSends(t *testing.T) {
	repo := &fakePhoneChallengeRepository{requestAccepted: true, requestPhone: ""}
	sender := &fakeSender{}
	svc := auth.NewPhoneChallengeService(repo, fixedCodeGenerator{code: "482913"}, sender, testChallengeCfg(), []byte(testHMACSecret))

	_ = svc.Request(context.Background(), "barbero@ejemplo.test", "ip-hash-de-prueba")
	if len(sender.calls) != 0 {
		t.Fatal("expected no SendCode call when phone is empty")
	}
}

func TestPhoneChallengeService_Request_SenderFails_ReturnsInternalButStillRegistered(t *testing.T) {
	repo := &fakePhoneChallengeRepository{requestAccepted: true, requestPhone: "+573001234567"}
	sender := &fakeSender{err: errors.New("whatsapp: timeout")}
	svc := auth.NewPhoneChallengeService(repo, fixedCodeGenerator{code: "482913"}, sender, testChallengeCfg(), []byte(testHMACSecret))

	err := svc.Request(context.Background(), "barbero@ejemplo.test", "ip-hash-de-prueba")

	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected KindInternal for a delivery failure, got %v (%T)", err, err)
	}
	if len(repo.requestCalls) != 1 {
		t.Fatal("expected the challenge to still be registered even if delivery fails")
	}
}

func TestPhoneChallengeService_Request_CodeGeneratorFails_ReturnsInternalWithoutCallingRepository(t *testing.T) {
	repo := &fakePhoneChallengeRepository{requestAccepted: true, requestPhone: "+573001234567"}
	svc := auth.NewPhoneChallengeService(repo, failingCodeGenerator{err: errors.New("rand exhausted")}, &fakeSender{}, testChallengeCfg(), []byte(testHMACSecret))

	err := svc.Request(context.Background(), "barbero@ejemplo.test", "ip-hash-de-prueba")

	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected KindInternal, got %v (%T)", err, err)
	}
	if len(repo.requestCalls) != 0 {
		t.Fatal("expected no RequestChallenge call when code generation fails")
	}
}

func TestPhoneChallengeService_Verify_Success(t *testing.T) {
	repo := &fakePhoneChallengeRepository{verifyOK: true}
	svc := auth.NewPhoneChallengeService(repo, fixedCodeGenerator{}, &fakeSender{}, testChallengeCfg(), []byte(testHMACSecret))

	if err := svc.Verify(context.Background(), "Barbero@Ejemplo.TEST", "ip-hash-de-prueba", "482913"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.verifyCall.email != "barbero@ejemplo.test" {
		t.Fatalf("expected normalized email, got %q", repo.verifyCall.email)
	}
	if len(repo.verifyCall.codeHash) != 64 || repo.verifyCall.codeHash == "482913" {
		t.Fatalf("expected a 64-char hex HMAC code hash, got %q", repo.verifyCall.codeHash)
	}
}

func TestPhoneChallengeService_Verify_Failure_ReturnsUnauthorized(t *testing.T) {
	repo := &fakePhoneChallengeRepository{verifyOK: false}
	svc := auth.NewPhoneChallengeService(repo, fixedCodeGenerator{}, &fakeSender{}, testChallengeCfg(), []byte(testHMACSecret))

	err := svc.Verify(context.Background(), "barbero@ejemplo.test", "ip-hash-de-prueba", "000000")

	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindUnauthorized {
		t.Fatalf("expected KindUnauthorized, got %v (%T)", err, err)
	}
}

// TestPhoneChallengeService_Verify_UnknownAccountAndWrongCode_SameError
// cubre la no enumeración del reto (DEC-062): una cuenta inexistente y un
// código incorrecto para una cuenta existente producen el mismo error.
func TestPhoneChallengeService_Verify_UnknownAccountAndWrongCode_SameError(t *testing.T) {
	unknownRepo := &fakePhoneChallengeRepository{verifyOK: false}
	unknownSvc := auth.NewPhoneChallengeService(unknownRepo, fixedCodeGenerator{}, &fakeSender{}, testChallengeCfg(), []byte(testHMACSecret))
	unknownErr := unknownSvc.Verify(context.Background(), "no-existe@ejemplo.test", "ip-hash", "000000")

	wrongRepo := &fakePhoneChallengeRepository{verifyOK: false}
	wrongSvc := auth.NewPhoneChallengeService(wrongRepo, fixedCodeGenerator{}, &fakeSender{}, testChallengeCfg(), []byte(testHMACSecret))
	wrongErr := wrongSvc.Verify(context.Background(), "barbero@ejemplo.test", "ip-hash", "111111")

	unknownApp, _ := apperr.As(unknownErr)
	wrongApp, _ := apperr.As(wrongErr)
	if unknownApp.Kind != wrongApp.Kind || unknownApp.Message != wrongApp.Message {
		t.Fatalf("expected identical apperr, got %+v vs %+v", unknownApp, wrongApp)
	}
}

func TestPhoneChallengeService_Verify_RepositoryError_ReturnsInternal(t *testing.T) {
	repo := &fakePhoneChallengeRepository{verifyErr: errors.New("db unavailable")}
	svc := auth.NewPhoneChallengeService(repo, fixedCodeGenerator{}, &fakeSender{}, testChallengeCfg(), []byte(testHMACSecret))

	err := svc.Verify(context.Background(), "barbero@ejemplo.test", "ip-hash", "000000")
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected KindInternal, got %v (%T)", err, err)
	}
}

func TestCryptoPhoneCodeGenerator_ProducesSixDigits(t *testing.T) {
	gen := auth.NewCryptoPhoneCodeGenerator()
	for i := 0; i < 50; i++ {
		code, err := gen.New()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(code) != 6 {
			t.Fatalf("expected a 6-digit code, got %q", code)
		}
		for _, c := range code {
			if c < '0' || c > '9' {
				t.Fatalf("expected only digits, got %q", code)
			}
		}
	}
}

type emailSendCall struct{ email, code string }

type fakeEmailSender struct {
	err   error
	calls []emailSendCall
}

func (s *fakeEmailSender) SendCode(_ context.Context, email, code string) error {
	s.calls = append(s.calls, emailSendCall{email, code})
	return s.err
}

func newChannelService(repo *fakePhoneChallengeRepository, wa *fakeSender, mail *fakeEmailSender) *auth.PhoneChallengeService {
	svc := auth.NewPhoneChallengeService(repo, fixedCodeGenerator{code: "482913"}, wa, testChallengeCfg(), []byte(testHMACSecret))
	if mail != nil {
		svc.WithEmailSender(mail)
	}
	return svc
}

// DEC-139: con el canal de correo el código sale solo por correo.
func TestPhoneChallengeService_Request_EmailChannel_SendsOnlyEmail(t *testing.T) {
	repo := &fakePhoneChallengeRepository{requestAccepted: true, requestEmail: "duena@ejemplo.test"}
	wa, mail := &fakeSender{}, &fakeEmailSender{}

	if err := newChannelService(repo, wa, mail).Request(context.Background(), "duena@ejemplo.test", "ip"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wa.calls) != 0 {
		t.Fatalf("whatsapp must not be used with the email channel, got %d calls", len(wa.calls))
	}
	if len(mail.calls) != 1 || mail.calls[0] != (emailSendCall{"duena@ejemplo.test", "482913"}) {
		t.Fatalf("unexpected email calls: %+v", mail.calls)
	}
}

// Con ambos canales el MISMO código sale por WhatsApp y por correo, tras una
// sola solicitud al repositorio (una operación lógica).
func TestPhoneChallengeService_Request_BothChannels_SameCodeOneLogicalRequest(t *testing.T) {
	repo := &fakePhoneChallengeRepository{requestAccepted: true, requestPhone: "+573001234567", requestEmail: "duena@ejemplo.test"}
	wa, mail := &fakeSender{}, &fakeEmailSender{}

	if err := newChannelService(repo, wa, mail).Request(context.Background(), "duena@ejemplo.test", "ip"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.requestCalls) != 1 {
		t.Fatalf("expected one logical request, got %d", len(repo.requestCalls))
	}
	if len(wa.calls) != 1 || len(mail.calls) != 1 {
		t.Fatalf("expected one delivery per channel, got whatsapp=%d email=%d", len(wa.calls), len(mail.calls))
	}
	if wa.calls[0].code != mail.calls[0].code || wa.calls[0].code != "482913" {
		t.Fatalf("both channels must carry the same code, got %q and %q", wa.calls[0].code, mail.calls[0].code)
	}
}

// Fallo parcial: el otro canal igual se intenta; el error es solo diagnóstico,
// nombra el canal y no incluye destinatario ni código.
func TestPhoneChallengeService_Request_BothChannels_PartialFailure(t *testing.T) {
	repo := &fakePhoneChallengeRepository{requestAccepted: true, requestPhone: "+573001234567", requestEmail: "duena@ejemplo.test"}
	wa := &fakeSender{err: errors.New("whatsapp: timeout")}
	mail := &fakeEmailSender{}

	err := newChannelService(repo, wa, mail).Request(context.Background(), "duena@ejemplo.test", "ip")

	if len(mail.calls) != 1 {
		t.Fatal("a failing channel must not prevent the other one from being attempted")
	}
	if err == nil {
		t.Fatal("expected a diagnostic error for the failed channel")
	}
	msg := err.Error()
	if !strings.Contains(msg, "whatsapp") || strings.Contains(msg, "correo") {
		t.Fatalf("the error must name only the failed channel, got %q", msg)
	}
	for _, secret := range []string{"+573001234567", "duena@ejemplo.test", "482913"} {
		if strings.Contains(msg, secret) {
			t.Fatalf("the error must not leak %q: %q", secret, msg)
		}
	}
}

func TestPhoneChallengeService_Request_BothChannels_BothFail(t *testing.T) {
	repo := &fakePhoneChallengeRepository{requestAccepted: true, requestPhone: "+573001234567", requestEmail: "duena@ejemplo.test"}
	wa := &fakeSender{err: errors.New("whatsapp: down")}
	mail := &fakeEmailSender{err: errors.New("resend: 500")}

	err := newChannelService(repo, wa, mail).Request(context.Background(), "duena@ejemplo.test", "ip")

	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected KindInternal, got %v", err)
	}
	if !strings.Contains(err.Error(), "whatsapp") || !strings.Contains(err.Error(), "correo") {
		t.Fatalf("expected both channels named, got %q", err.Error())
	}
}

// Un destino de correo sin remitente configurado es un fallo interno de ese
// canal: no se sustituye por WhatsApp (DEC-081).
func TestPhoneChallengeService_Request_EmailWithoutSender_NoFallbackToWhatsApp(t *testing.T) {
	repo := &fakePhoneChallengeRepository{requestAccepted: true, requestEmail: "duena@ejemplo.test"}
	wa := &fakeSender{}

	err := newChannelService(repo, wa, nil).Request(context.Background(), "duena@ejemplo.test", "ip")

	if err == nil {
		t.Fatal("expected an internal error when there is no email sender")
	}
	if len(wa.calls) != 0 {
		t.Fatal("the email channel must never fall back to WhatsApp")
	}
}

// Con WhatsApp configurado, un destino de correo vacío nunca dispara correo.
func TestPhoneChallengeService_Request_WhatsAppChannel_NeverSendsEmail(t *testing.T) {
	repo := &fakePhoneChallengeRepository{requestAccepted: true, requestPhone: "+573001234567"}
	wa, mail := &fakeSender{}, &fakeEmailSender{}

	if err := newChannelService(repo, wa, mail).Request(context.Background(), "duena@ejemplo.test", "ip"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wa.calls) != 1 || len(mail.calls) != 0 {
		t.Fatalf("expected only whatsapp, got whatsapp=%d email=%d", len(wa.calls), len(mail.calls))
	}
}
