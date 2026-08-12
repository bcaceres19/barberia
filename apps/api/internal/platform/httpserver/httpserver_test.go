package httpserver_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"system-barbershop/internal/platform/httpserver"
)

func TestRequestID_AcceptsValidClientID(t *testing.T) {
	handler := httpserver.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Seen-Id", httpserver.RequestIDFromContext(r.Context()))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(httpserver.RequestIDHeader, "client-supplied-id-123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(httpserver.RequestIDHeader); got != "client-supplied-id-123" {
		t.Fatalf("expected valid client id to be respected, got %q", got)
	}
	if got := rec.Header().Get("X-Seen-Id"); got != "client-supplied-id-123" {
		t.Fatalf("expected context to carry client id, got %q", got)
	}
}

func TestRequestID_RejectsInvalidClientID(t *testing.T) {
	cases := []string{
		"",
		strings.Repeat("a", 200),
		"contains space",
		"contains\nnewline",
		"contains\ttab",
	}

	for _, invalid := range cases {
		handler := httpserver.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if invalid != "" {
			req.Header.Set(httpserver.RequestIDHeader, invalid)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		got := rec.Header().Get(httpserver.RequestIDHeader)
		if got == invalid {
			t.Fatalf("expected invalid client id %q to be replaced, but it was echoed back", invalid)
		}
		if got == "" {
			t.Fatalf("expected a generated id for invalid input %q, got empty", invalid)
		}
	}
}

func TestRecover_DoesNotLogRawPanicValue(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	sensitive := "correo=cliente@ejemplo.com telefono=555-0000"

	handler := httpserver.Recover(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(sensitive)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(req.Context())
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	logOutput := buf.String()
	if strings.Contains(logOutput, sensitive) {
		t.Fatalf("raw panic value leaked into log: %s", logOutput)
	}
	if !strings.Contains(logOutput, "panic_type") {
		t.Fatalf("expected panic_type field in log, got: %s", logOutput)
	}
}

func TestRecover_RespondsWithProblemJSON(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))

	handler := httpserver.RequestID(httpserver.Recover(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("expected application/problem+json, got %q", ct)
	}

	var p httprobProblem
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	if p.Status != http.StatusInternalServerError {
		t.Fatalf("expected status field 500, got %d", p.Status)
	}
	if p.Title == "" {
		t.Fatal("expected non-empty title")
	}
	if p.Instance == "" {
		t.Fatal("expected instance to carry the request id")
	}
}

// httprobProblem espeja httpserver.Problem para no depender de exportar
// más de lo necesario en la prueba.
type httprobProblem struct {
	Type     string `json:"type,omitempty"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

func TestWriteProblem_SetsContentTypeAndStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	httpserver.WriteProblem(rec, httpserver.Problem{
		Title:  "No encontrado",
		Status: http.StatusNotFound,
	})

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("expected application/problem+json, got %q", ct)
	}
}
