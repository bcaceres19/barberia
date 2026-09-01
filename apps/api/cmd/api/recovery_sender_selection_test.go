// Pruebas de la matriz de selección de HU-008 (issue #86,
// docs/10-backlog/prompts/test/issue-86-otp-correo-resend.md): confirman
// qué auth.RecoveryCodeSender concreto construye selectRecoverySender para
// cada combinación de ambiente y credenciales, sin llamadas de red reales
// (las credenciales usadas aquí son ficticias, nunca reales).
package main

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"system-barbershop/internal/platform/config"
)

const (
	fakeMetaPhoneNumberID = "fake-meta-phone-number-id"
	fakeMetaAccessToken   = "fake-meta-access-token"
	fakeMetaTemplateName  = "fake-meta-template"
	fakeResendAPIKey      = "fake-resend-api-key"
	fakeResendFromAddress = "no-responder@barberia.test"
)

func recoverySenderType(cfg config.Config) string {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	return fmt.Sprintf("%T", selectRecoverySender(cfg, logger))
}

func TestSelectRecoverySender_MetaAndResendComplete_UsesDualChannel(t *testing.T) {
	for _, env := range []string{"local", "test", "pilot", "production"} {
		cfg := config.Config{
			Environment:               env,
			MetaWhatsAppPhoneNumberID: fakeMetaPhoneNumberID,
			MetaWhatsAppAccessToken:   fakeMetaAccessToken,
			MetaWhatsAppTemplateName:  fakeMetaTemplateName,
			ResendAPIKey:              fakeResendAPIKey,
			ResendFromAddress:         fakeResendFromAddress,
		}
		got := recoverySenderType(cfg)
		if got != "notification.DualChannelRecoverySender" {
			t.Fatalf("env=%s: expected DualChannelRecoverySender, got %s", env, got)
		}
	}
}

func TestSelectRecoverySender_OnlyResendComplete_LocalOrTest_UsesEmailOnly(t *testing.T) {
	for _, env := range []string{"local", "test"} {
		cfg := config.Config{
			Environment:       env,
			ResendAPIKey:      fakeResendAPIKey,
			ResendFromAddress: fakeResendFromAddress,
		}
		got := recoverySenderType(cfg)
		if got != "notification.EmailOnlyRecoverySender" {
			t.Fatalf("env=%s: expected EmailOnlyRecoverySender, got %s", env, got)
		}
	}
}

func TestSelectRecoverySender_OnlyResendComplete_HardenedEnvironment_NeverUsesEmailOnly(t *testing.T) {
	for _, env := range []string{"pilot", "production"} {
		cfg := config.Config{
			Environment:       env,
			ResendAPIKey:      fakeResendAPIKey,
			ResendFromAddress: fakeResendFromAddress,
		}
		got := recoverySenderType(cfg)
		if got != "auth.LoggingRecoveryCodeSender" {
			t.Fatalf("env=%s: expected LoggingRecoveryCodeSender (never email-only outside local/test), got %s", env, got)
		}
	}
}

func TestSelectRecoverySender_OnlyMetaComplete_UsesLoggingPlaceholder(t *testing.T) {
	cfg := config.Config{
		Environment:               "local",
		MetaWhatsAppPhoneNumberID: fakeMetaPhoneNumberID,
		MetaWhatsAppAccessToken:   fakeMetaAccessToken,
		MetaWhatsAppTemplateName:  fakeMetaTemplateName,
	}
	got := recoverySenderType(cfg)
	if got != "auth.LoggingRecoveryCodeSender" {
		t.Fatalf("expected LoggingRecoveryCodeSender, got %s", got)
	}
}

func TestSelectRecoverySender_NoCredentials_UsesLoggingPlaceholder(t *testing.T) {
	cfg := config.Config{Environment: "local"}
	got := recoverySenderType(cfg)
	if got != "auth.LoggingRecoveryCodeSender" {
		t.Fatalf("expected LoggingRecoveryCodeSender, got %s", got)
	}
}
