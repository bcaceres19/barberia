package httpserver_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"system-barbershop/internal/platform/httpserver"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))
}

func TestNewRouter_MountsThreeAudienceGroups(t *testing.T) {
	router := httpserver.NewRouter(discardLogger())

	for _, prefix := range []string{"/api/v1/public/", "/api/v1/customer/", "/api/v1/private/"} {
		req := httptest.NewRequest(http.MethodGet, prefix+"cualquier-cosa", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		// Sin operaciones de negocio todavía, chi responde 404 dentro del
		// grupo montado; lo relevante es que el grupo existe (no un 404 de
		// "ruta desconocida" antes de llegar al router) y que el middleware
		// base ya corrió sobre esa respuesta.
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: expected 404 from the mounted-but-empty group, got %d", prefix, rec.Code)
		}
		if got := rec.Header().Get(httpserver.RequestIDHeader); got == "" {
			t.Errorf("%s: expected a request id even on a 404 from an empty audience group", prefix)
		}
		if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Errorf("%s: expected security headers on a 404 too, got %q", prefix, got)
		}
	}
}

func TestNewRouter_RecoversPanicInsideAudienceGroup(t *testing.T) {
	router := httpserver.NewRouter(discardLogger())
	router.Get("/api/v1/private/panics", func(w http.ResponseWriter, r *http.Request) {
		panic("boom dentro de una audiencia real")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/private/panics", nil)
	req.Header.Set(httpserver.RequestIDHeader, "trace-abc123")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("expected application/problem+json, got %q", ct)
	}

	var p struct {
		Instance  string `json:"instance"`
		RequestID string `json:"requestId"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	// La correlación sobrevive al panic: prueba que RequestID corrió antes
	// que Recover en la cadena y que ambos llegaron intactos hasta el
	// handler que hizo panic (orden observable de middleware).
	if p.Instance != "trace-abc123" || p.RequestID != "trace-abc123" {
		t.Fatalf("expected the client-supplied request id to survive the panic, got instance=%q requestId=%q", p.Instance, p.RequestID)
	}
}

func TestNewRouter_LogsMatchedRoutePatternNotRawPath(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	router := httpserver.NewRouter(logger)
	router.Get("/api/v1/customer/appointments/{token}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customer/appointments/un-token-secreto-de-verdad", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	logLine := buf.String()
	if bytes.Contains([]byte(logLine), []byte("un-token-secreto-de-verdad")) {
		t.Fatalf("expected the raw token to be absent from the log, got: %s", logLine)
	}

	var fields map[string]any
	if err := json.Unmarshal(buf.Bytes(), &fields); err != nil {
		t.Fatalf("log line is not valid JSON: %v", err)
	}
	if fields["route"] != "/api/v1/customer/appointments/{token}" {
		t.Fatalf("expected the route pattern, got %v", fields["route"])
	}
}

func TestNewRouter_NotFoundOutsideAnyAudienceIsStillUniformProblem(t *testing.T) {
	router := httpserver.NewRouter(discardLogger())

	req := httptest.NewRequest(http.MethodGet, "/no-existe", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assertUniformNotFoundProblem(t, rec)
}

func TestNewRouter_MethodNotAllowedIsAlsoUniformProblem(t *testing.T) {
	router := httpserver.NewRouter(discardLogger())
	router.Get("/api/v1/private/solo-get", func(w http.ResponseWriter, r *http.Request) {})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/private/solo-get", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// docs/06-api/estandar-openapi.md no incluye 405 en su tabla de códigos;
	// un método no soportado en una ruta conocida colapsa a 404 igual que un
	// recurso ajeno, por el mismo motivo de no confirmar la existencia de
	// algo que el cliente no debería poder inferir.
	assertUniformNotFoundProblem(t, rec)
}

func assertUniformNotFoundProblem(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("expected application/problem+json, got %q", ct)
	}
	if got := rec.Header().Get(httpserver.RequestIDHeader); got == "" {
		t.Error("expected a request id header")
	}

	var p struct {
		Type   string `json:"type"`
		Title  string `json:"title"`
		Status int    `json:"status"`
		Code   string `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	if p.Type == "" || p.Title == "" || p.Code == "" || p.Status != http.StatusNotFound {
		t.Fatalf("expected a fully populated problem, got %+v", p)
	}
}
