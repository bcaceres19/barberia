package notification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"system-barbershop/internal/modules/auth"
)

const (
	twilioVerifyBaseURL = "https://verify.twilio.com/v2"
	twilioVerifyTimeout = 5 * time.Second
)

type TwilioVerifyConfig struct {
	AccountSID       string
	AuthToken        string
	APIKeySID        string
	APIKeySecret     string
	VerifyServiceSID string
	Channel          string
}

// TwilioVerifyOTPProvider delegates code generation and validation to Twilio
// Verify. Its persistence digest is an opaque app marker, never an OTP.
type TwilioVerifyOTPProvider struct {
	httpClient *http.Client
	cfg        TwilioVerifyConfig
	marker     string
}

func NewTwilioVerifyOTPProvider(cfg TwilioVerifyConfig, marker string, httpClient *http.Client) TwilioVerifyOTPProvider {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if cfg.Channel == "" {
		cfg.Channel = "sms"
	}
	return TwilioVerifyOTPProvider{httpClient: httpClient, cfg: cfg, marker: marker}
}

var _ auth.WhatsAppOTPProvider = TwilioVerifyOTPProvider{}

func (p TwilioVerifyOTPProvider) Prepare(_ context.Context) (auth.PreparedWhatsAppOTP, error) {
	return auth.NewProviderManagedWhatsAppOTP(p.marker), nil
}

func (p TwilioVerifyOTPProvider) Deliver(ctx context.Context, phone string, _ auth.PreparedWhatsAppOTP) error {
	status, err := p.post(ctx, "Verifications", url.Values{"To": {phone}, "Channel": {p.cfg.Channel}})
	if err != nil {
		return err
	}
	if status != "pending" {
		return errors.New("notification: Twilio Verify no dejó el OTP pendiente")
	}
	return nil
}

func (p TwilioVerifyOTPProvider) VerificationDigest(ctx context.Context, phone, code string) (string, error) {
	status, err := p.post(ctx, "VerificationCheck", url.Values{"To": {phone}, "Code": {code}})
	if err != nil {
		return "", err
	}
	if status != "approved" {
		return "", auth.ErrWhatsAppOTPRejected
	}
	return p.marker, nil
}

func (p TwilioVerifyOTPProvider) post(ctx context.Context, resource string, values url.Values) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, twilioVerifyTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/Services/%s/%s", twilioVerifyBaseURL, p.cfg.VerifyServiceSID, resource),
		strings.NewReader(values.Encode()))
	if err != nil {
		return "", fmt.Errorf("notification: construir solicitud Twilio Verify: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	setTwilioBasicAuth(req, p.cfg.AccountSID, p.cfg.AuthToken, p.cfg.APIKeySID, p.cfg.APIKeySecret)
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("notification: solicitar Twilio Verify: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, resp.Body)
		if resource == "VerificationCheck" && (resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound) {
			return "", auth.ErrWhatsAppOTPRejected
		}
		return "", fmt.Errorf("notification: Twilio Verify respondió %d", resp.StatusCode)
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("notification: decodificar respuesta Twilio Verify: %w", err)
	}
	return body.Status, nil
}

func setTwilioBasicAuth(req *http.Request, accountSID, authToken, apiKeySID, apiKeySecret string) {
	if apiKeySID != "" && apiKeySecret != "" {
		req.SetBasicAuth(apiKeySID, apiKeySecret)
		return
	}
	req.SetBasicAuth(accountSID, authToken)
}
