package auth_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"system-barbershop/internal/modules/auth"
)

func TestCapturingPhoneCodeSender_WritesPhoneAndCodeToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "capture.json")
	sender := auth.NewCapturingPhoneCodeSender(&fakeSender{}, path)

	if err := sender.SendCode(context.Background(), "+573001234567", "482913"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected the capture file to exist: %v", err)
	}
	var captured struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	if err := json.Unmarshal(raw, &captured); err != nil {
		t.Fatalf("capture file is not valid JSON: %v", err)
	}
	if captured.Phone != "+573001234567" || captured.Code != "482913" {
		t.Fatalf("unexpected capture content: %+v", captured)
	}
}

func TestCapturingPhoneCodeSender_InnerFailure_NeverWritesCapture(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "capture.json")
	sender := auth.NewCapturingPhoneCodeSender(&fakeSender{err: errors.New("whatsapp: timeout")}, path)

	if err := sender.SendCode(context.Background(), "+573001234567", "482913"); err == nil {
		t.Fatal("expected the inner sender's error to propagate")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("expected no capture file when the inner sender fails")
	}
}
