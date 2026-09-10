package notification_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/modules/notification"
)

var _ auth.PhoneCodeSender = notification.PhoneChallengeSender{}

type recordingWhatsAppSender struct {
	phone string
	code  string
	calls int
	err   error
}

func (s *recordingWhatsAppSender) Send(_ context.Context, phone, code string) error {
	s.phone = phone
	s.code = code
	s.calls++
	return s.err
}

func TestPhoneChallengeSender_SendCode_DelegatesExactlyOnce(t *testing.T) {
	channel := &recordingWhatsAppSender{}
	sender := notification.NewPhoneChallengeSender(channel)

	if err := sender.SendCode(context.Background(), "+573001234567", "482913"); err != nil {
		t.Fatalf("SendCode: %v", err)
	}
	if channel.calls != 1 || channel.phone != "+573001234567" || channel.code != "482913" {
		t.Fatalf("unexpected delegation: calls=%d phone=%q code=%q", channel.calls, channel.phone, channel.code)
	}
}

func TestPhoneChallengeSender_SendCode_PropagatesChannelError(t *testing.T) {
	errChannel := errors.New("notification: delivery unavailable")
	channel := &recordingWhatsAppSender{err: errChannel}
	sender := notification.NewPhoneChallengeSender(channel)

	err := sender.SendCode(context.Background(), "+573001234567", "482913")
	if !errors.Is(err, errChannel) {
		t.Fatalf("expected channel error, got %v", err)
	}
	if strings.Contains(err.Error(), "573001234567") || strings.Contains(err.Error(), "482913") {
		t.Fatalf("sender error must not contain delivery inputs: %v", err)
	}
}
