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

	tests := []struct {
		name string
		cfg  config.Config
		want string
	}{
		{
			name: "plantilla completa usa Meta",
			cfg: config.Config{
				OTPProvider: "meta", MetaWhatsAppMode: "template",
				MetaWhatsAppPhoneNumberID: fakeMetaPhoneNumberID, MetaWhatsAppAccessToken: fakeMetaAccessToken,
				MetaWhatsAppTemplateName: fakeMetaTemplateName,
			},
			want: "notification.MetaWhatsAppOTPProvider",
		},
		{
			name: "texto con destinatarios en local usa Meta",
			cfg: config.Config{
				OTPProvider: "meta", MetaWhatsAppMode: "text", Environment: "local",
				MetaWhatsAppPhoneNumberID: fakeMetaPhoneNumberID, MetaWhatsAppAccessToken: fakeMetaAccessToken,
				MetaWhatsAppTestRecipients: []string{"+573001234567"},
			},
			want: "notification.MetaWhatsAppOTPProvider",
		},
		{
			name: "texto sin destinatarios en local no envía",
			cfg: config.Config{
				OTPProvider: "meta", MetaWhatsAppMode: "text", Environment: "local",
				MetaWhatsAppPhoneNumberID: fakeMetaPhoneNumberID, MetaWhatsAppAccessToken: fakeMetaAccessToken,
			},
			want: "auth.localWhatsAppOTPProvider",
		},
		{
			name: "texto sin destinatarios en producción usa Meta",
			cfg: config.Config{
				OTPProvider: "meta", MetaWhatsAppMode: "text", Environment: "production",
				MetaWhatsAppPhoneNumberID: fakeMetaPhoneNumberID, MetaWhatsAppAccessToken: fakeMetaAccessToken,
			},
			want: "notification.MetaWhatsAppOTPProvider",
		},
		{
			name: "modo plantilla sin plantilla no envía",
			cfg: config.Config{
				OTPProvider: "meta", MetaWhatsAppMode: "template",
				MetaWhatsAppPhoneNumberID: fakeMetaPhoneNumberID, MetaWhatsAppAccessToken: fakeMetaAccessToken,
			},
			want: "auth.localWhatsAppOTPProvider",
		},
		{
			name: "sin credenciales registra el código",
			cfg:  config.Config{OTPProvider: "meta", MetaWhatsAppMode: "template"},
			want: "auth.localWhatsAppOTPProvider",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fmt.Sprintf("%T", selectWhatsAppOTPProvider(tt.cfg, logger, secret))
			if got != tt.want {
				t.Fatalf("expected %s, got %s", tt.want, got)
			}
		})
	}
}
