package notification_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"system-barbershop/internal/modules/notification"
)

func newChangeNoticeSender(t *testing.T, srv *httptest.Server) notification.ResendChangeNoticeSender {
	t.Helper()
	client := srv.Client()
	client.Transport = redirectTransport{target: srv.URL}
	return notification.NewResendChangeNoticeSender(
		notification.ResendChangeNoticeConfig{APIKey: "resend-key", FromAddress: "no-responder@barberia.test"}, client)
}

func TestResendChangeNoticeSender_Cancelled_PostsTheCurrentAppointment(t *testing.T) {
	var gotAuth string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	err := newChangeNoticeSender(t, srv).SendChangeNotice(context.Background(),
		"cancelled_by_barber", "cliente.a@ejemplo.test", "Barbería Centro", "Corte clásico", "Ana", "2099-01-07 05:00 (America/Bogota)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAuth != "Bearer resend-key" || gotBody["subject"] != "Tu turno fue cancelado" {
		t.Fatalf("unexpected request: auth=%q subject=%v", gotAuth, gotBody["subject"])
	}
	to, _ := gotBody["to"].([]any)
	if len(to) != 1 || to[0] != "cliente.a@ejemplo.test" {
		t.Fatalf("unexpected to: %v", gotBody["to"])
	}
	text, _ := gotBody["text"].(string)
	for _, want := range []string{"Ana", "Barbería Centro", "canceló", "Corte clásico", "2099-01-07 05:00 (America/Bogota)"} {
		if !strings.Contains(text, want) {
			t.Fatalf("el texto debe contener %q:\n%s", want, text)
		}
	}
}

func TestResendChangeNoticeSender_Rescheduled_SaysTheTimeIsTheCurrentOne(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	if err := newChangeNoticeSender(t, srv).SendChangeNotice(context.Background(),
		"rescheduled", "cliente.a@ejemplo.test", "Barbería Centro", "Corte", "Ana", "2099-01-07 16:30 (America/Bogota)"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["subject"] != "Tu turno cambió de horario" {
		t.Fatalf("unexpected subject: %v", gotBody["subject"])
	}
	text, _ := gotBody["text"].(string)
	if !strings.Contains(text, "horario vigente") || !strings.Contains(text, "2099-01-07 16:30 (America/Bogota)") {
		t.Fatalf("el texto debe dar la hora vigente:\n%s", text)
	}
}

func TestResendChangeNoticeSender_UnknownKind_FailsWithoutCallingResend(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }))
	defer srv.Close()

	if err := newChangeNoticeSender(t, srv).SendChangeNotice(context.Background(),
		"inventado", "cliente.a@ejemplo.test", "B", "S", "A", "t"); err == nil {
		t.Fatal("un tipo desconocido debe fallar")
	}
	if called {
		t.Fatal("no debe llamar a Resend con un tipo desconocido")
	}
}

func TestResendChangeNoticeSender_ProviderStatus_ClassifiesPermanentAndTransient(t *testing.T) {
	cases := []struct {
		status    int
		permanent bool
	}{
		{http.StatusUnprocessableEntity, true},
		{http.StatusUnauthorized, true},
		{http.StatusTooManyRequests, false},
		{http.StatusRequestTimeout, false},
		{http.StatusInternalServerError, false},
		{http.StatusBadGateway, false},
	}
	for _, c := range cases {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(c.status)
			// Resend puede repetir el destinatario en el cuerpo del error.
			_, _ = w.Write([]byte(`{"message":"invalid to cliente.a@ejemplo.test"}`))
		}))
		err := newChangeNoticeSender(t, srv).SendChangeNotice(context.Background(),
			"cancelled_by_barber", "cliente.a@ejemplo.test", "B", "S", "A", "t")
		srv.Close()

		if err == nil {
			t.Fatalf("status %d: debe fallar", c.status)
		}
		var p interface{ Permanent() bool }
		if !errors.As(err, &p) || p.Permanent() != c.permanent {
			t.Fatalf("status %d: permanent esperado %v, error %v", c.status, c.permanent, err)
		}
		if strings.Contains(err.Error(), "cliente.a@ejemplo.test") {
			t.Fatalf("status %d: el error no debe incluir el destinatario (RN-DAT-02): %v", c.status, err)
		}
	}
}

func TestResendChangeNoticeSender_TransportFailure_IsTransientAndHidesTheRecipient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	sender := newChangeNoticeSender(t, srv)
	srv.Close() // la conexión falla

	err := sender.SendChangeNotice(context.Background(),
		"rescheduled", "cliente.a@ejemplo.test", "B", "S", "A", "t")
	if err == nil {
		t.Fatal("debe fallar")
	}
	var p interface{ Permanent() bool }
	if errors.As(err, &p) && p.Permanent() {
		t.Fatal("un fallo de red no es permanente")
	}
	if strings.Contains(err.Error(), "cliente.a@ejemplo.test") {
		t.Fatalf("el error no debe incluir el destinatario: %v", err)
	}
}
