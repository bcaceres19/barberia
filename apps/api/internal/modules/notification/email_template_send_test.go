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

// Los cuatro correos transaccionales salen con la plantilla de marca: la misma
// petición lleva HTML y texto plano (issue #465), y el nombre del cliente, que
// es entrada de usuario, llega escapado en el HTML.
func TestTransactionalEmails_SendBrandedHTMLAndText(t *testing.T) {
	const evil = `Ana <script>alert(1)</script>`

	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = nil
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	client := srv.Client()
	client.Transport = redirectTransport{target: srv.URL}

	cases := []struct {
		name  string
		send  func() error
		wants []string
	}{
		{"recuperación", func() error {
			return notification.NewResendEmailSender(
				notification.ResendEmailConfig{APIKey: "k", FromAddress: "a@b.test", Subject: "Recuperación"}, client,
			).Send(context.Background(), "c@ejemplo.test", "482913")
		}, []string{"482913", "Recupera tu acceso"}},
		{"acceso", func() error {
			return notification.NewResendLoginChallengeEmailSender(
				notification.ResendEmailConfig{APIKey: "k", FromAddress: "a@b.test"}, client,
			).SendCode(context.Background(), "c@ejemplo.test", "482913")
		}, []string{"482913", "Verifica que eres tú"}},
		{"confirmación", func() error {
			return notification.NewResendConfirmationEmailSender(
				notification.ResendConfirmationEmailConfig{APIKey: "k", FromAddress: "a@b.test"}, client,
			).SendConfirmation(context.Background(), "c@ejemplo.test", "Barbería Centro", "Corte", evil,
				"jueves 7 de enero", "https://nava.test/turno/x")
		}, []string{"Barbería Centro", "Corte", "Ver mi turno", "https://nava.test/turno/x"}},
		{"cancelación", func() error {
			return newChangeNoticeSender(t, srv).SendChangeNotice(context.Background(),
				"cancelled_by_barber", "c@ejemplo.test", "Barbería Centro", "Corte", evil, "jueves 7 de enero")
		}, []string{"Barbería Centro", "Turno cancelado", "jueves 7 de enero"}},
		{"reprogramación", func() error {
			return newChangeNoticeSender(t, srv).SendChangeNotice(context.Background(),
				"rescheduled", "c@ejemplo.test", "Barbería Centro", "Corte", evil, "viernes 8 de enero")
		}, []string{"Barbería Centro", "Horario actualizado", "viernes 8 de enero"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.send(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			html, _ := got["html"].(string)
			text, _ := got["text"].(string)
			if html == "" || text == "" {
				t.Fatalf("la petición debe llevar html y text: html=%d text=%d", len(html), len(text))
			}
			if !strings.Contains(html, "NAVA") || !strings.Contains(text, "NAVA") {
				t.Error("ambas versiones llevan la marca")
			}
			for _, want := range tc.wants {
				if !strings.Contains(strings.ToLower(html), strings.ToLower(want)) || !strings.Contains(strings.ToLower(text), strings.ToLower(want)) {
					t.Errorf("%q debe estar en html y en text", want)
				}
			}
			if strings.Contains(html, "<script>") {
				t.Errorf("el nombre del cliente debe escaparse en el HTML:\n%s", html)
			}
		})
	}
}
