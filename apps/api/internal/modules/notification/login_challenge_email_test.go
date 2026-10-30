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

func newLoginChallengeSenderAgainstTestServer(t *testing.T, srv *httptest.Server) notification.ResendLoginChallengeEmailSender {
	t.Helper()
	client := srv.Client()
	client.Transport = redirectTransport{target: srv.URL}
	return notification.NewResendLoginChallengeEmailSender(notification.ResendEmailConfig{
		APIKey: "resend-key", FromAddress: "no-responder@barberia.test", Subject: "asunto de recuperación que no aplica",
	}, client)
}

// DEC-139: el correo del reto de acceso lleva el código exacto, su propio asunto
// (no el de recuperación) y un texto que no habla de recuperar el acceso.
func TestResendLoginChallengeEmailSender_SendCode_PostsCodeWithOwnSubjectAndText(t *testing.T) {
	var gotAuth string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	if err := newLoginChallengeSenderAgainstTestServer(t, srv).SendCode(context.Background(), "duena.a@ejemplo.test", "482913"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotAuth != "Bearer resend-key" {
		t.Fatalf("unexpected Authorization header: %q", gotAuth)
	}
	to, ok := gotBody["to"].([]any)
	if !ok || len(to) != 1 || to[0] != "duena.a@ejemplo.test" {
		t.Fatalf("unexpected to: %v", gotBody["to"])
	}
	if gotBody["subject"] != "Código de verificación de acceso" {
		t.Fatalf("the login challenge must use its own fixed subject, got %v", gotBody["subject"])
	}
	text, _ := gotBody["text"].(string)
	if !strings.Contains(text, "482913") || strings.Contains(text, "recuperación") {
		t.Fatalf("unexpected email text: %q", text)
	}
}

func TestResendLoginChallengeEmailSender_SendCode_ErrorStatus_DoesNotLeakRecipientOrCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"invalid to duena.a@ejemplo.test"}`))
	}))
	defer srv.Close()

	err := newLoginChallengeSenderAgainstTestServer(t, srv).SendCode(context.Background(), "duena.a@ejemplo.test", "482913")

	if err == nil {
		t.Fatal("expected an error on a non-2xx response")
	}
	for _, secret := range []string{"duena.a@ejemplo.test", "482913"} {
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("the error must not leak %q: %v", secret, err)
		}
	}
}
