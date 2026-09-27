package notification

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"system-barbershop/internal/modules/auth"
)

const (
	twilioMessagingBaseURL = "https://api.twilio.com/2010-04-01"
	twilioSandboxTimeout   = 5 * time.Second
)

// TwilioSandboxConfig holds development-only credentials for Twilio's shared
// WhatsApp Sandbox sender. It must never be used in a production environment.
type TwilioSandboxConfig struct {
	AccountSID   string
	AuthToken    string
	APIKeySID    string
	APIKeySecret string
	FromE164     string
}

// TwilioSandboxWhatsAppOTPProvider uses Programmable Messaging to exercise
// the WhatsApp delivery path in local/test. Unlike Verify, Sandbox does not
// manage the OTP, so the application remains its sole code authority.
type TwilioSandboxWhatsAppOTPProvider struct {
	httpClient *http.Client
	cfg        TwilioSandboxConfig
	codes      auth.PhoneCodeGenerator
	secret     []byte
}

func NewTwilioSandboxWhatsAppOTPProvider(cfg TwilioSandboxConfig, codes auth.PhoneCodeGenerator, secret []byte, httpClient *http.Client) TwilioSandboxWhatsAppOTPProvider {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return TwilioSandboxWhatsAppOTPProvider{httpClient: httpClient, cfg: cfg, codes: codes, secret: secret}
}

var _ auth.WhatsAppOTPProvider = TwilioSandboxWhatsAppOTPProvider{}

func (p TwilioSandboxWhatsAppOTPProvider) Prepare(_ context.Context) (auth.PreparedWhatsAppOTP, error) {
	code, err := p.codes.New()
	if err != nil {
		return auth.PreparedWhatsAppOTP{}, fmt.Errorf("notification: generar OTP Sandbox Twilio: %w", err)
	}
	return auth.NewLocalWhatsAppOTP(auth.HMACHex(code, p.secret), code), nil
}

func (p TwilioSandboxWhatsAppOTPProvider) Deliver(ctx context.Context, phone string, prepared auth.PreparedWhatsAppOTP) error {
	ctx, cancel := context.WithTimeout(ctx, twilioSandboxTimeout)
	defer cancel()

	values := url.Values{
		"From": {"whatsapp:" + p.cfg.FromE164},
		"To":   {"whatsapp:" + phone},
		"Body": {"Tu código de verificación de Barbería es: " + prepared.Code()},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/Accounts/%s/Messages.json", twilioMessagingBaseURL, p.cfg.AccountSID),
		strings.NewReader(values.Encode()))
	if err != nil {
		return fmt.Errorf("notification: construir solicitud Sandbox Twilio: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	setTwilioBasicAuth(req, p.cfg.AccountSID, p.cfg.AuthToken, p.cfg.APIKeySID, p.cfg.APIKeySecret)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("notification: solicitar Sandbox Twilio: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("notification: Sandbox Twilio respondió %d", resp.StatusCode)
	}
	return nil
}

func (p TwilioSandboxWhatsAppOTPProvider) VerificationDigest(_ context.Context, _ string, code string) (string, error) {
	return auth.HMACHex(code, p.secret), nil
}
