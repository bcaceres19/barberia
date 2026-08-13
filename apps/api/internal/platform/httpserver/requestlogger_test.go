package httpserver_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"system-barbershop/internal/platform/httpserver"
)

// allowedLogFields es la lista cerrada de RN-DAT-02: cualquier otra clave en
// una línea de log del request logger es una fuga. time/level/msg los añade
// slog.JSONHandler siempre; el resto los añade RequestLogger.
var allowedLogFields = map[string]bool{
	"time": true, "level": true, "msg": true,
	"request_id": true, "method": true, "route": true,
	"status": true, "duration_ms": true,
}

// forbiddenPatterns detecta los datos que RN-DAT-02 prohíbe en un registro
// técnico: correo, teléfono y las palabras clave de credenciales/tokens que
// CA-003-04 exige vigilar.
var forbiddenPatterns = []*regexp.Regexp{
	regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`), // correo
	regexp.MustCompile(`\+\d[\d\s-]{6,}\d`),                              // teléfono con prefijo internacional, p. ej. +57 300 123 4567
	regexp.MustCompile(`(?i)password`),
	regexp.MustCompile(`(?i)authorization`),
	regexp.MustCompile(`(?i)cookie`),
	regexp.MustCompile(`(?i)token`),
}

func findForbiddenPattern(logLine string) (string, bool) {
	for _, p := range forbiddenPatterns {
		if p.MatchString(logLine) {
			return p.String(), true
		}
	}
	return "", false
}

// TestFindForbiddenPattern_DetectsPoisonedLine es la prueba de la prueba: si
// findForbiddenPattern nunca detectara nada, TestRequestLogger_NoForbiddenFields
// pasaría vacuamente. Aquí se confirma que un campo prohibido real sí dispara
// la detección.
func TestFindForbiddenPattern_DetectsPoisonedLine(t *testing.T) {
	poisoned := []string{
		`{"msg":"solicitud http","cliente":"cliente@ejemplo.com"}`,
		`{"msg":"solicitud http","telefono":"+57 300 123 4567"}`,
		`{"msg":"solicitud http","password":"hunter2"}`,
		`{"msg":"solicitud http","Authorization":"Bearer abc"}`,
		`{"msg":"solicitud http","Cookie":"session=abc"}`,
		`{"msg":"solicitud http","access_token":"abc123"}`,
	}
	for _, line := range poisoned {
		if _, found := findForbiddenPattern(line); !found {
			t.Errorf("expected findForbiddenPattern to flag: %s", line)
		}
	}
}

func TestRequestLogger_OnlyAllowedFieldsAndNoForbiddenContent(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	handler := httpserver.RequestID(httpserver.RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})))

	req := httptest.NewRequest(http.MethodGet, "/cualquier/ruta?correo=cliente@ejemplo.com", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	line := buf.String()
	if line == "" {
		t.Fatal("expected the request logger to emit a line")
	}

	if pattern, found := findForbiddenPattern(line); found {
		t.Fatalf("log line matched forbidden pattern %s: %s", pattern, line)
	}

	var fields map[string]any
	if err := json.Unmarshal(buf.Bytes(), &fields); err != nil {
		t.Fatalf("log line is not valid JSON: %v", err)
	}
	for key := range fields {
		if !allowedLogFields[key] {
			t.Errorf("unexpected field %q in log line: %s", key, line)
		}
	}
	if fields["status"].(float64) != http.StatusTeapot {
		t.Errorf("expected status %d in log, got %v", http.StatusTeapot, fields["status"])
	}
	if fields["route"] != "unmatched" {
		t.Errorf("expected route %q for a request outside chi routing, got %v", "unmatched", fields["route"])
	}
}
