package auth

import (
	"context"
	"errors"
	"strings"
)

// ErrWhatsAppOTPRejected is returned by a provider when a submitted code is
// not valid. Services translate it to their existing uniform public errors.
var ErrWhatsAppOTPRejected = errors.New("auth: código OTP de WhatsApp rechazado")

// rejectedOTPDigest is handed to the repository after a provider rejects a
// code. It has the 64-hex shape the SQL functions require but can never equal
// a stored digest (HMAC-SHA256 output or provider marker), so the local
// attempt counter still advances and max_attempts invalidates the challenge
// even though the provider, not the database, judged the code.
var rejectedOTPDigest = strings.Repeat("0", 64)

// PreparedWhatsAppOTP contains only the material needed to persist and
// deliver a challenge. The code is empty for provider-managed challenges
// (a provider that validates codes itself) and never reaches a repository or a log.
type PreparedWhatsAppOTP struct {
	persistenceDigest string
	code              string
}

func NewLocalWhatsAppOTP(digest, code string) PreparedWhatsAppOTP {
	return PreparedWhatsAppOTP{persistenceDigest: digest, code: code}
}

func NewProviderManagedWhatsAppOTP(digest string) PreparedWhatsAppOTP {
	return PreparedWhatsAppOTP{persistenceDigest: digest}
}

func (p PreparedWhatsAppOTP) PersistenceDigest() string { return p.persistenceDigest }
func (p PreparedWhatsAppOTP) Code() string              { return p.code }

// WhatsAppOTPProvider is the provider boundary consumed by auth. It owns
// delivery and the representation used to verify a code; auth only persists
// that representation for its own rate limits and business flow.
type WhatsAppOTPProvider interface {
	Prepare(ctx context.Context) (PreparedWhatsAppOTP, error)
	Deliver(ctx context.Context, phoneE164 string, prepared PreparedWhatsAppOTP) error
	VerificationDigest(ctx context.Context, phoneE164, code string) (string, error)
}

type localWhatsAppOTPProvider struct {
	codes  PhoneCodeGenerator
	sender PhoneCodeSender
	secret []byte
}

func NewLocalWhatsAppOTPProvider(codes PhoneCodeGenerator, sender PhoneCodeSender, secret []byte) WhatsAppOTPProvider {
	return localWhatsAppOTPProvider{codes: codes, sender: sender, secret: secret}
}

func (p localWhatsAppOTPProvider) Prepare(_ context.Context) (PreparedWhatsAppOTP, error) {
	code, err := p.codes.New()
	if err != nil {
		return PreparedWhatsAppOTP{}, err
	}
	return NewLocalWhatsAppOTP(HMACHex(code, p.secret), code), nil
}

func (p localWhatsAppOTPProvider) Deliver(ctx context.Context, phone string, prepared PreparedWhatsAppOTP) error {
	return p.sender.SendCode(ctx, phone, prepared.Code())
}

func (p localWhatsAppOTPProvider) VerificationDigest(_ context.Context, _ string, code string) (string, error) {
	return HMACHex(code, p.secret), nil
}
