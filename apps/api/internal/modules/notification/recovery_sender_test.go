package notification_test

import (
	"context"
	"errors"
	"testing"

	"system-barbershop/internal/modules/notification"
)

type spyChannel struct {
	calls []string
	err   error
}

func (s *spyChannel) Send(_ context.Context, dest, code string) error {
	s.calls = append(s.calls, dest+":"+code)
	return s.err
}

func TestDualChannelRecoverySender_SendsBothChannels(t *testing.T) {
	whatsapp := &spyChannel{}
	email := &spyChannel{}
	sender := notification.NewDualChannelRecoverySender(whatsapp, email)

	if err := sender.SendCode(context.Background(), "+573001234567", "a@b.test", "482913"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(whatsapp.calls) != 1 || whatsapp.calls[0] != "+573001234567:482913" {
		t.Fatalf("unexpected whatsapp calls: %v", whatsapp.calls)
	}
	if len(email.calls) != 1 || email.calls[0] != "a@b.test:482913" {
		t.Fatalf("unexpected email calls: %v", email.calls)
	}
}

// TestDualChannelRecoverySender_OneChannelFails_StillCallsTheOther cubre la
// tolerancia a fallo parcial de DEC-066: un canal que falla no debe impedir
// el intento por el otro.
func TestDualChannelRecoverySender_OneChannelFails_StillCallsTheOther(t *testing.T) {
	whatsapp := &spyChannel{err: errors.New("meta down")}
	email := &spyChannel{}
	sender := notification.NewDualChannelRecoverySender(whatsapp, email)

	err := sender.SendCode(context.Background(), "+573001234567", "a@b.test", "482913")
	if err == nil {
		t.Fatal("expected an error to be returned for internal logging")
	}
	if len(email.calls) != 1 {
		t.Fatal("expected the email channel to still be attempted despite the whatsapp failure")
	}
}

func TestDualChannelRecoverySender_BothChannelsFail_ReturnsCombinedError(t *testing.T) {
	whatsapp := &spyChannel{err: errors.New("meta down")}
	email := &spyChannel{err: errors.New("resend down")}
	sender := notification.NewDualChannelRecoverySender(whatsapp, email)

	err := sender.SendCode(context.Background(), "+573001234567", "a@b.test", "482913")
	if err == nil {
		t.Fatal("expected an error when both channels fail")
	}
}

// TestDualChannelRecoverySender_OnlyChosenChannel cubre CA-008-09 (DEC-092):
// con un solo destino lleno se entrega únicamente por ese canal y el otro no
// se invoca ni como respaldo, incluso si el elegido falla.
func TestDualChannelRecoverySender_OnlyChosenChannel(t *testing.T) {
	t.Run("whatsapp elegido no toca el correo", func(t *testing.T) {
		whatsapp := &spyChannel{}
		email := &spyChannel{}
		sender := notification.NewDualChannelRecoverySender(whatsapp, email)

		if err := sender.SendCode(context.Background(), "+573001234567", "", "482913"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(whatsapp.calls) != 1 || len(email.calls) != 0 {
			t.Fatalf("expected only whatsapp, got whatsapp=%v email=%v", whatsapp.calls, email.calls)
		}
	})

	t.Run("correo elegido no toca WhatsApp", func(t *testing.T) {
		whatsapp := &spyChannel{}
		email := &spyChannel{}
		sender := notification.NewDualChannelRecoverySender(whatsapp, email)

		if err := sender.SendCode(context.Background(), "", "a@b.test", "482913"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(email.calls) != 1 || len(whatsapp.calls) != 0 {
			t.Fatalf("expected only email, got whatsapp=%v email=%v", whatsapp.calls, email.calls)
		}
	})

	t.Run("el canal elegido que falla no cae al otro", func(t *testing.T) {
		whatsappErr := errors.New("meta down")
		whatsapp := &spyChannel{err: whatsappErr}
		email := &spyChannel{}
		sender := notification.NewDualChannelRecoverySender(whatsapp, email)

		err := sender.SendCode(context.Background(), "+573001234567", "", "482913")
		if !errors.Is(err, whatsappErr) {
			t.Fatalf("expected the whatsapp error, got %v", err)
		}
		if len(email.calls) != 0 {
			t.Fatalf("email must never be a fallback, got %v", email.calls)
		}
	})

	t.Run("sin destino no envía", func(t *testing.T) {
		whatsapp := &spyChannel{}
		email := &spyChannel{}
		sender := notification.NewDualChannelRecoverySender(whatsapp, email)

		if err := sender.SendCode(context.Background(), "", "", "482913"); err == nil {
			t.Fatal("expected an error when no destination is provided")
		}
		if len(whatsapp.calls)+len(email.calls) != 0 {
			t.Fatal("nothing must be sent without a destination")
		}
	})
}
