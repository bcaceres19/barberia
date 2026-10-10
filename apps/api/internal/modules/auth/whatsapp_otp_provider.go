package auth

import "context"

// PreparedWhatsAppOTP contains only the material needed to persist and
// deliver a challenge. The code never reaches a repository or a log.
type PreparedWhatsAppOTP struct {
	persistenceDigest string
	code              string
}

func NewLocalWhatsAppOTP(digest, code string) PreparedWhatsAppOTP {
	return PreparedWhatsAppOTP{persistenceDigest: digest, code: code}
}

func (p PreparedWhatsAppOTP) PersistenceDigest() string { return p.persistenceDigest }
func (p PreparedWhatsAppOTP) Code() string              { return p.code }

// WhatsAppOTPProvider is the delivery boundary consumed by auth. NAVA always
// generates and validates the code; the provider only transports it and
// derives the digest the repository stores and compares.
type WhatsAppOTPProvider interface {
	Prepare(ctx context.Context) (PreparedWhatsAppOTP, error)
	Deliver(ctx context.Context, phoneE164 string, prepared PreparedWhatsAppOTP) error
	VerificationDigest(code string) string
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

func (p localWhatsAppOTPProvider) VerificationDigest(code string) string {
	return HMACHex(code, p.secret)
}
