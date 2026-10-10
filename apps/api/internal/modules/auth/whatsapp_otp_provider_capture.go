package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// CapturingWhatsAppOTPProvider exists only for local/test E2E flows using a
// locally generated provider. Provider-managed OTPs intentionally cannot be
// captured because the application never learns their code.
type CapturingWhatsAppOTPProvider struct {
	inner WhatsAppOTPProvider
	path  string
}

func NewCapturingWhatsAppOTPProvider(inner WhatsAppOTPProvider, path string) CapturingWhatsAppOTPProvider {
	return CapturingWhatsAppOTPProvider{inner: inner, path: path}
}

func (p CapturingWhatsAppOTPProvider) Prepare(ctx context.Context) (PreparedWhatsAppOTP, error) {
	return p.inner.Prepare(ctx)
}

func (p CapturingWhatsAppOTPProvider) Deliver(ctx context.Context, phone string, prepared PreparedWhatsAppOTP) error {
	if err := p.inner.Deliver(ctx, phone, prepared); err != nil {
		return err
	}
	data, err := json.Marshal(struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}{Phone: phone, Code: prepared.Code()})
	if err != nil {
		return fmt.Errorf("auth: serializar captura OTP: %w", err)
	}
	if err := os.WriteFile(p.path, data, 0o600); err != nil {
		return fmt.Errorf("auth: escribir captura OTP: %w", err)
	}
	return nil
}

func (p CapturingWhatsAppOTPProvider) VerificationDigest(code string) string {
	return p.inner.VerificationDigest(code)
}
