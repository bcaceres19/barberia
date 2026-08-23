package notification_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"system-barbershop/internal/modules/notification"
)

func newResendSenderAgainstTestServer(t *testing.T, srv *httptest.Server, cfg notification.ResendEmailConfig) notification.ResendEmailSender {
	t.Helper()
	client := srv.Client()
	client.Transport = redirectTransport{target: srv.URL}
	return notification.NewResendEmailSender(cfg, client)
}

func TestResendEmailSender_Send_PostsExactCodeToRecipient(t *testing.T) {
	var gotAuth string
	var gotBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"resend-id"}`))
	}))
	defer srv.Close()

	sender := newResendSenderAgainstTestServer(t, srv, notification.ResendEmailConfig{
		APIKey: "resend-key", FromAddress: "no-responder@barberia.test", Subject: "Código de recuperación",
	})

	if err := sender.Send(context.Background(), "duena.a@ejemplo.test", "482913"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotAuth != "Bearer resend-key" {
		t.Fatalf("unexpected Authorization header: %q", gotAuth)
	}
	if gotBody["from"] != "no-responder@barberia.test" {
		t.Fatalf("unexpected from: %v", gotBody["from"])
	}
	to, ok := gotBody["to"].([]any)
	if !ok || len(to) != 1 || to[0] != "duena.a@ejemplo.test" {
		t.Fatalf("unexpected to: %v", gotBody["to"])
	}
	text, _ := gotBody["text"].(string)
	if !strings.Contains(text, "482913") {
		t.Fatal("expected the exact code to appear in the email body")
	}
}

func TestResendEmailSender_Send_ErrorStatus_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	sender := newResendSenderAgainstTestServer(t, srv, notification.ResendEmailConfig{APIKey: "bad", FromAddress: "a@b.test", Subject: "x"})

	if err := sender.Send(context.Background(), "a@b.test", "482913"); err == nil {
		t.Fatal("expected an error on a non-2xx response")
	}
}
