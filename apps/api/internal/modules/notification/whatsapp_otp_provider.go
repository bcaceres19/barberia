package notification

import (
	"context"
	"fmt"

	"system-barbershop/internal/modules/auth"
)

// MetaWhatsAppOTPProvider preserves Meta's application-managed OTP model:
// it generates a local code, sends it through the existing sender and later
// yields its HMAC digest for the auth repository to validate.
type MetaWhatsAppOTPProvider struct {
	sender authPhoneSender
	codes  auth.PhoneCodeGenerator
	secret []byte
}

type authPhoneSender interface {
	Send(ctx context.Context, phoneE164, code string) error
}

func NewMetaWhatsAppOTPProvider(sender WhatsAppSender, codes auth.PhoneCodeGenerator, secret []byte) MetaWhatsAppOTPProvider {
	return MetaWhatsAppOTPProvider{sender: sender, codes: codes, secret: secret}
}

var _ auth.WhatsAppOTPProvider = MetaWhatsAppOTPProvider{}

func (p MetaWhatsAppOTPProvider) Prepare(_ context.Context) (auth.PreparedWhatsAppOTP, error) {
	code, err := p.codes.New()
	if err != nil {
		return auth.PreparedWhatsAppOTP{}, fmt.Errorf("notification: generar OTP Meta: %w", err)
	}
	return auth.NewLocalWhatsAppOTP(auth.HMACHex(code, p.secret), code), nil
}

func (p MetaWhatsAppOTPProvider) Deliver(ctx context.Context, phone string, prepared auth.PreparedWhatsAppOTP) error {
	return p.sender.Send(ctx, phone, prepared.Code())
}

func (p MetaWhatsAppOTPProvider) VerificationDigest(_ context.Context, _ string, code string) (string, error) {
	return auth.HMACHex(code, p.secret), nil
}
