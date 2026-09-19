package notification_test

import (
	"context"
	"errors"
	"testing"

	"system-barbershop/internal/modules/notification"
)

func TestEmailOnlyRecoverySender_SendsExactlyOneEmailIgnoringPhone(t *testing.T) {
	email := &spyChannel{}
	sender := notification.NewEmailOnlyRecoverySender(email)

	if err := sender.SendCode(context.Background(), "+573001234567", "a@b.test", "482913"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(email.calls) != 1 || email.calls[0] != "a@b.test:482913" {
		t.Fatalf("unexpected email calls: %v", email.calls)
	}
}

func TestEmailOnlyRecoverySender_ChannelFails_PropagatesError(t *testing.T) {
	channelErr := errors.New("resend down")
	email := &spyChannel{err: channelErr}
	sender := notification.NewEmailOnlyRecoverySender(email)

	err := sender.SendCode(context.Background(), "+573001234567", "a@b.test", "482913")
	if !errors.Is(err, channelErr) {
		t.Fatalf("expected the channel error to propagate, got: %v", err)
	}
}

// TestEmailOnlyRecoverySender_WhatsAppChoice_NeverFallsBackToEmail cubre
// CA-008-09: sin proveedor de WhatsApp en el ambiente, elegir ese canal no
// envía nada por correo.
func TestEmailOnlyRecoverySender_WhatsAppChoice_NeverFallsBackToEmail(t *testing.T) {
	email := &spyChannel{}
	sender := notification.NewEmailOnlyRecoverySender(email)

	if err := sender.SendCode(context.Background(), "+573001234567", "", "482913"); err == nil {
		t.Fatal("expected an error for internal logging when whatsapp is not configured")
	}
	if len(email.calls) != 0 {
		t.Fatalf("no email must be sent, got %v", email.calls)
	}
}
