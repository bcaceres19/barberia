package httpapi_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"system-barbershop/internal/modules/notification"
	notificationhttpapi "system-barbershop/internal/modules/notification/httpapi"
	"system-barbershop/internal/platform/httpserver"
)

const (
	verifyToken = "token-de-verificacion-de-prueba"
	appSecret   = "app-secret-de-prueba-0123456789"
	numberID    = "1315949954943476"
	webhookPath = "/api/v1/public/webhooks/meta/whatsapp"
)

type countingStore struct {
	inbound  atomic.Int32
	statuses atomic.Int32
	err      error
}

func (s *countingStore) RecordInbound(context.Context, string, time.Time) error {
	s.inbound.Add(1)
	return s.err
}

func (s *countingStore) RecordStatus(context.Context, string, string, time.Time, *int) (bool, error) {
	s.statuses.Add(1)
	return true, s.err
}

func (s *countingStore) ConversationOpen(context.Context, string) (bool, error) { return false, nil }

func newServer(t *testing.T, store *countingStore) *httptest.Server {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := notificationhttpapi.NewMetaWebhookHandler(
		notification.NewMetaWebhookService(store, numberID, []byte("secreto-hmac-de-prueba-0123456789"), logger),
		verifyToken, appSecret, logger,
	)
	router, _ := httpserver.NewRouter(logger)
	router.Get(webhookPath, handler.Verify)
	router.Post(webhookPath, handler.Receive)
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return srv
}

func sign(body string) string {
	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write([]byte(body))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func get(t *testing.T, srv *httptest.Server, query url.Values) (*http.Response, string) {
	t.Helper()
	resp, err := http.Get(srv.URL + webhookPath + "?" + query.Encode())
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp, string(body)
}

func post(t *testing.T, srv *httptest.Server, body, signature string) (*http.Response, string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, srv.URL+webhookPath, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if signature != "" {
		req.Header.Set("X-Hub-Signature-256", signature)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return resp, string(out)
}

const validBody = `{"object":"whatsapp_business_account","entry":[{"changes":[{"field":"messages","value":{` +
	`"metadata":{"phone_number_id":"` + numberID + `"},` +
	`"messages":[{"from":"573016823624","id":"wamid.in","timestamp":"1790000000"}],` +
	`"statuses":[{"id":"wamid.out","status":"delivered","timestamp":"1790000000"}]}}]}]}`

func TestVerify_CorrectTokenEchoesTheChallenge(t *testing.T) {
	srv := newServer(t, &countingStore{})
	resp, body := get(t, srv, url.Values{"hub.mode": {"subscribe"}, "hub.verify_token": {verifyToken}, "hub.challenge": {"1158201444"}})

	if resp.StatusCode != http.StatusOK || body != "1158201444" {
		t.Fatalf("expected 200 with the challenge, got %d %q", resp.StatusCode, body)
	}
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/plain") {
		t.Fatalf("expected text/plain, got %q", resp.Header.Get("Content-Type"))
	}
}

func TestVerify_RejectsWrongTokenWrongModeAndUnsafeChallenge(t *testing.T) {
	srv := newServer(t, &countingStore{})
	tests := []struct {
		name  string
		query url.Values
		want  int
	}{
		{"token incorrecto", url.Values{"hub.mode": {"subscribe"}, "hub.verify_token": {"otro"}, "hub.challenge": {"1"}}, http.StatusUnauthorized},
		{"sin token", url.Values{"hub.mode": {"subscribe"}, "hub.challenge": {"1"}}, http.StatusUnauthorized},
		{"modo distinto", url.Values{"hub.mode": {"unsubscribe"}, "hub.verify_token": {verifyToken}, "hub.challenge": {"1"}}, http.StatusUnauthorized},
		{"desafío con HTML", url.Values{"hub.mode": {"subscribe"}, "hub.verify_token": {verifyToken}, "hub.challenge": {"<script>alert(1)</script>"}}, http.StatusBadRequest},
		{"desafío vacío", url.Values{"hub.mode": {"subscribe"}, "hub.verify_token": {verifyToken}}, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := get(t, srv, tt.query)
			if resp.StatusCode != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, resp.StatusCode)
			}
			if !strings.HasPrefix(resp.Header.Get("Content-Type"), "application/problem+json") {
				t.Fatalf("errors must be problem+json, got %q", resp.Header.Get("Content-Type"))
			}
			if strings.Contains(body, "<script>") || strings.Contains(body, verifyToken) {
				t.Fatalf("the response must not reflect input or secrets: %s", body)
			}
		})
	}
}

func TestReceive_ValidSignatureRecordsEventsAndReturns200(t *testing.T) {
	store := &countingStore{}
	srv := newServer(t, store)

	resp, _ := post(t, srv, validBody, sign(validBody))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if store.inbound.Load() != 1 || store.statuses.Load() != 1 {
		t.Fatalf("expected one inbound and one status, got %d and %d", store.inbound.Load(), store.statuses.Load())
	}
}

func TestReceive_InvalidOrMissingSignatureNeverTouchesTheStore(t *testing.T) {
	store := &countingStore{}
	srv := newServer(t, store)

	tampered := strings.Replace(validBody, "delivered", "failed", 1)
	for name, signature := range map[string]string{
		"sin firma":        "",
		"firma de otro":    sign(`{"otro":true}`),
		"cuerpo alterado":  sign(tampered),
		"sin prefijo":      strings.TrimPrefix(sign(validBody), "sha256="),
		"firma incompleta": sign(validBody)[:20],
	} {
		// En «cuerpo alterado» la firma corresponde al cuerpo manipulado, no al enviado.
		resp, out := post(t, srv, validBody, signature)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("%s: expected 401, got %d", name, resp.StatusCode)
		}
		if strings.Contains(out, appSecret) {
			t.Fatalf("%s: the response leaked the secret", name)
		}
	}
	if store.inbound.Load() != 0 || store.statuses.Load() != 0 {
		t.Fatal("an unauthenticated notification must never reach the store")
	}
}

func TestReceive_SignedButMalformedBodyIsRejected(t *testing.T) {
	store := &countingStore{}
	srv := newServer(t, store)

	resp, _ := post(t, srv, "no es json", sign("no es json"))
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestReceive_StoreFailureReturns500SoMetaRetries(t *testing.T) {
	store := &countingStore{err: errors.New("db caída")}
	srv := newServer(t, store)

	resp, out := post(t, srv, validBody, sign(validBody))
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
	if strings.Contains(out, "db caída") {
		t.Fatalf("the response must not include the internal cause: %s", out)
	}
}

func TestReceive_OversizedBodyIsRejectedBeforeAnyWork(t *testing.T) {
	store := &countingStore{}
	srv := newServer(t, store)

	big := `{"object":"` + strings.Repeat("a", httpserver.MaxRequestBodyBytes+10) + `"}`
	resp, _ := post(t, srv, big, sign(big))
	if resp.StatusCode < 400 || resp.StatusCode >= 500 {
		t.Fatalf("expected a 4xx for an oversized body, got %d", resp.StatusCode)
	}
	if store.inbound.Load() != 0 {
		t.Fatal("an oversized body must not reach the store")
	}
}

func TestRoute_AbsentWhenNotRegistered(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router, _ := httpserver.NewRouter(logger)
	srv := httptest.NewServer(router)
	defer srv.Close()

	resp, err := http.Post(srv.URL+webhookPath, "application/json", bytes.NewReader([]byte(validBody)))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 when the webhook is not configured, got %d", resp.StatusCode)
	}
}
