package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"system-barbershop/internal/modules/auth"
)

func TestLoggingChallengeEmailSender_NeverLogsDestinationOrCode(t *testing.T) {
	var buf bytes.Buffer
	sender := auth.NewLoggingChallengeEmailSender(slog.New(slog.NewTextHandler(&buf, nil)))

	if err := sender.SendCode(context.Background(), "duena.a@ejemplo.test", "482913"); err != nil {
		t.Fatalf("the placeholder must never fail: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected a log line stating that a delivery was simulated")
	}
	if strings.Contains(buf.String(), "duena.a@ejemplo.test") || strings.Contains(buf.String(), "482913") {
		t.Fatalf("the log must not contain the email or the code (RN-DAT-02): %s", buf.String())
	}
}

func TestCapturingChallengeEmailSender_WritesCodeOnlyAfterSuccessfulDelivery(t *testing.T) {
	path := filepath.Join(t.TempDir(), "capture.json")
	inner := &fakeEmailSender{}

	if err := auth.NewCapturingChallengeEmailSender(inner, path).SendCode(context.Background(), "duena.a@ejemplo.test", "482913"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected a capture file: %v", err)
	}
	var got struct{ Email, Code string }
	if err := json.Unmarshal(data, &got); err != nil || got.Code != "482913" || got.Email != "duena.a@ejemplo.test" {
		t.Fatalf("unexpected capture %s (%v)", data, err)
	}
	if len(inner.calls) != 1 {
		t.Fatalf("expected the inner sender to be used once, got %d", len(inner.calls))
	}

	failPath := filepath.Join(t.TempDir(), "never.json")
	failing := &fakeEmailSender{err: errors.New("resend: 500")}
	if err := auth.NewCapturingChallengeEmailSender(failing, failPath).SendCode(context.Background(), "a@b.test", "111111"); err == nil {
		t.Fatal("expected the inner error to propagate")
	}
	if _, err := os.Stat(failPath); err == nil {
		t.Fatal("a failed delivery must not leave a captured code")
	}
}
