package httpserver_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"system-barbershop/internal/platform/httpserver"
)

// echoBodyHandler es un handler solo de prueba —nunca se monta en el router
// de producción (HU-003 no agrega endpoints ficticios)— que existe
// únicamente para demostrar que BodyLimit + Translate rechazan un cuerpo
// que excede el límite, igual que TestRecover_* usa un handler que hace
// panic solo para ejercitar Recover.
func echoBodyHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := io.Copy(io.Discard, r.Body); err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, httpserver.RequestIDFromContext(r.Context())))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func TestBodyLimit_RejectsOversizedBody(t *testing.T) {
	handler := httpserver.RequestID(httpserver.BodyLimit(http.HandlerFunc(echoBodyHandler)))

	oversized := bytes.Repeat([]byte("a"), httpserver.MaxRequestBodyBytes+1)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(oversized))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for an oversized body, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("expected application/problem+json, got %q", ct)
	}

	var p struct {
		Code   string `json:"code"`
		Status int    `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	if p.Code != "payload-too-large" {
		t.Fatalf("expected payload-too-large code, got %q", p.Code)
	}
}

func TestBodyLimit_AllowsBodyWithinLimit(t *testing.T) {
	handler := httpserver.RequestID(httpserver.BodyLimit(http.HandlerFunc(echoBodyHandler)))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("cuerpo pequeño"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for a body within the limit, got %d", rec.Code)
	}
}

func TestBodyLimit_ProducesRecognizableMaxBytesError(t *testing.T) {
	// Confirma la premisa de la que depende TestBodyLimit_RejectsOversizedBody:
	// http.MaxBytesReader efectivamente produce un *http.MaxBytesError, no un
	// io.ErrUnexpectedEOF u otro error genérico que Translate no reconocería.
	rec := httptest.NewRecorder()
	body := http.MaxBytesReader(rec, io.NopCloser(bytes.NewReader(bytes.Repeat([]byte("a"), 10))), 1)

	_, err := io.ReadAll(body)

	var mbe *http.MaxBytesError
	if !errors.As(err, &mbe) {
		t.Fatalf("expected *http.MaxBytesError, got %T: %v", err, err)
	}
}
