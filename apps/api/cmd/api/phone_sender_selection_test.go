package main

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"system-barbershop/internal/platform/config"
)

func TestSelectWhatsAppOTPProvider_SelectsConfiguredProvider(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	secret := []byte("secreto-de-prueba-suficientemente-largo-0123456789")
	meta := selectWhatsAppOTPProvider(config.Config{OTPProvider: "meta", MetaWhatsAppPhoneNumberID: fakeMetaPhoneNumberID, MetaWhatsAppAccessToken: fakeMetaAccessToken, MetaWhatsAppTemplateName: fakeMetaTemplateName}, logger, secret)
	if got := fmt.Sprintf("%T", meta); got != "notification.MetaWhatsAppOTPProvider" {
		t.Fatalf("expected Meta provider, got %s", got)
	}
	twilio := selectWhatsAppOTPProvider(config.Config{OTPProvider: "twilio", TwilioAccountSID: "ACfake", TwilioAuthToken: "fake", TwilioVerifyServiceSID: "VAfake"}, logger, secret)
	if got := fmt.Sprintf("%T", twilio); got != "notification.TwilioVerifyOTPProvider" {
		t.Fatalf("expected Twilio provider, got %s", got)
	}
	sandbox := selectWhatsAppOTPProvider(config.Config{OTPProvider: "twilio_sandbox", TwilioAccountSID: "ACfake", TwilioAuthToken: "fake", TwilioWhatsAppSandboxFrom: "+14155238886"}, logger, secret)
	if got := fmt.Sprintf("%T", sandbox); got != "notification.TwilioSandboxWhatsAppOTPProvider" {
		t.Fatalf("expected Twilio Sandbox provider, got %s", got)
	}
}
