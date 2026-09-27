package notification_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/modules/notification"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func twilioProvider(t *testing.T, fn roundTripFunc) notification.TwilioVerifyOTPProvider {
	t.Helper()
	return notification.NewTwilioVerifyOTPProvider(notification.TwilioVerifyConfig{
		AccountSID: "ACfake", AuthToken: "fake-token", VerifyServiceSID: "VAfake",
	}, "provider-managed-digest", &http.Client{Transport: fn})
}

func response(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}

func TestTwilioVerifyProvider_DeliverPostsSMSVerificationByDefault(t *testing.T) {
	p := twilioProvider(t, func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v2/Services/VAfake/Verifications" || r.Method != http.MethodPost {
			t.Fatalf("unexpected endpoint: %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))
		if values.Get("To") != "+573001234567" || values.Get("Channel") != "sms" {
			t.Fatalf("unexpected form: %s", body)
		}
		return response(http.StatusCreated, `{"status":"pending"}`), nil
	})
	prepared, err := p.Prepare(context.Background())
	if err != nil || prepared.Code() != "" {
		t.Fatalf("unexpected provider-managed preparation: %+v, %v", prepared, err)
	}
	if err := p.Deliver(context.Background(), "+573001234567", prepared); err != nil {
		t.Fatalf("Deliver: %v", err)
	}
}

func TestTwilioVerifyProvider_DeliverPostsConfiguredWhatsAppVerification(t *testing.T) {
	p := notification.NewTwilioVerifyOTPProvider(notification.TwilioVerifyConfig{
		AccountSID: "ACfake", AuthToken: "fake-token", VerifyServiceSID: "VAfake", Channel: "whatsapp",
	}, "provider-managed-digest", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))
		if values.Get("Channel") != "whatsapp" {
			t.Fatalf("expected reversible WhatsApp channel, got %q", values.Get("Channel"))
		}
		return response(http.StatusCreated, `{"status":"pending"}`), nil
	})})
	prepared, err := p.Prepare(context.Background())
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if err := p.Deliver(context.Background(), "+573001234567", prepared); err != nil {
		t.Fatalf("Deliver: %v", err)
	}
}

func TestTwilioVerifyProvider_ApprovedCheckReturnsPersistenceDigest(t *testing.T) {
	p := twilioProvider(t, func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v2/Services/VAfake/VerificationCheck" {
			t.Fatalf("unexpected endpoint: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !bytes.Contains(body, []byte("Code=482913")) {
			t.Fatalf("verification code was not sent in the check request")
		}
		return response(http.StatusOK, `{"status":"approved"}`), nil
	})
	digest, err := p.VerificationDigest(context.Background(), "+573001234567", "482913")
	if err != nil || digest != "provider-managed-digest" {
		t.Fatalf("unexpected check result digest=%q err=%v", digest, err)
	}
}

func TestTwilioVerifyProvider_RejectedAndUnavailableAreTranslated(t *testing.T) {
	rejected := twilioProvider(t, func(*http.Request) (*http.Response, error) { return response(http.StatusNotFound, `{}`), nil })
	if _, err := rejected.VerificationDigest(context.Background(), "+573001234567", "000000"); !errors.Is(err, auth.ErrWhatsAppOTPRejected) {
		t.Fatalf("expected rejected OTP, got %v", err)
	}
	unavailable := twilioProvider(t, func(*http.Request) (*http.Response, error) { return nil, errors.New("timeout") })
	if _, err := unavailable.VerificationDigest(context.Background(), "+573001234567", "000000"); errors.Is(err, auth.ErrWhatsAppOTPRejected) || err == nil {
		t.Fatalf("expected external failure, got %v", err)
	}
}
