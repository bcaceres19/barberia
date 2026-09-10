package main

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"system-barbershop/internal/platform/config"
)

func phoneSenderType(cfg config.Config) string {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	return fmt.Sprintf("%T", selectPhoneChallengeSender(cfg, logger))
}

func TestSelectPhoneChallengeSender_MetaComplete_UsesRealSender(t *testing.T) {
	for _, env := range []string{"local", "test", "pilot", "production"} {
		cfg := config.Config{
			Environment:               env,
			MetaWhatsAppPhoneNumberID: fakeMetaPhoneNumberID,
			MetaWhatsAppAccessToken:   fakeMetaAccessToken,
			MetaWhatsAppTemplateName:  fakeMetaTemplateName,
		}
		if got := phoneSenderType(cfg); got != "notification.PhoneChallengeSender" {
			t.Fatalf("env=%s: expected PhoneChallengeSender, got %s", env, got)
		}
	}
}

func TestSelectPhoneChallengeSender_NoMeta_LocalOrTest_UsesLoggingPlaceholder(t *testing.T) {
	for _, env := range []string{"local", "test"} {
		if got := phoneSenderType(config.Config{Environment: env}); got != "auth.LoggingPhoneCodeSender" {
			t.Fatalf("env=%s: expected LoggingPhoneCodeSender, got %s", env, got)
		}
	}
}
