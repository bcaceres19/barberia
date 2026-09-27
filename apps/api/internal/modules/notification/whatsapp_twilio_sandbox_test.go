package notification_test

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/modules/notification"
)

type fixedCodeGenerator struct{}

func (fixedCodeGenerator) New() (string, error) { return "482913", nil }

func sandboxProvider(t *testing.T, fn roundTripFunc) notification.TwilioSandboxWhatsAppOTPProvider {
	t.Helper()
	return notification.NewTwilioSandboxWhatsAppOTPProvider(notification.TwilioSandboxConfig{
		AccountSID: "ACfake", AuthToken: "fake-token", FromE164: "+14155238886",
	}, fixedCodeGenerator{}, []byte("test-secret"), &http.Client{Transport: fn})
}

func TestTwilioSandboxProvider_DeliversLocalOTPThroughMessagingAPI(t *testing.T) {
	p := sandboxProvider(t, func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost || r.URL.Path != "/2010-04-01/Accounts/ACfake/Messages.json" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))
		if values.Get("From") != "whatsapp:+14155238886" || values.Get("To") != "whatsapp:+573001234567" {
			t.Fatalf("unexpected sender or recipient: %s", body)
		}
		if !strings.Contains(values.Get("Body"), "482913") {
			t.Fatal("expected locally generated test code in Sandbox message")
		}
		return response(http.StatusCreated, `{"status":"queued"}`), nil
	})
	prepared, err := p.Prepare(context.Background())
	if err != nil || prepared.PersistenceDigest() == "" {
		t.Fatalf("Prepare: prepared=%+v err=%v", prepared, err)
	}
	if err := p.Deliver(context.Background(), "+573001234567", prepared); err != nil {
		t.Fatalf("Deliver: %v", err)
	}
	digest, err := p.VerificationDigest(context.Background(), "+573001234567", "482913")
	if err != nil || digest != auth.HMACHex("482913", []byte("test-secret")) {
		t.Fatalf("VerificationDigest: digest=%q err=%v", digest, err)
	}
}

func TestTwilioSandboxProvider_ExternalFailureDoesNotValidateOTP(t *testing.T) {
	p := sandboxProvider(t, func(*http.Request) (*http.Response, error) {
		return response(http.StatusServiceUnavailable, `{}`), nil
	})
	prepared, err := p.Prepare(context.Background())
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if err := p.Deliver(context.Background(), "+573001234567", prepared); err == nil {
		t.Fatal("expected provider failure")
	}
}
func TestTwilioSandboxProvider_UsesAPIKeyCredentials(t *testing.T) {
	p := notification.NewTwilioSandboxWhatsAppOTPProvider(notification.TwilioSandboxConfig{
		AccountSID: "ACaccount", APIKeySID: "SKfake", APIKeySecret: "api-key-secret", FromE164: "+14155238886",
	}, fixedCodeGenerator{}, []byte("test-secret"), &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "SKfake" || password != "api-key-secret" {
			t.Fatal("expected API Key basic authentication")
		}
		return response(http.StatusCreated, `{"status":"queued"}`), nil
	})})
	prepared, err := p.Prepare(context.Background())
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if err := p.Deliver(context.Background(), "+573001234567", prepared); err != nil {
		t.Fatalf("Deliver: %v", err)
	}
}
