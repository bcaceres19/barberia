package httpserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"system-barbershop/internal/platform/httpserver"
)

func TestSecurityHeaders_SetsExpectedHeaders(t *testing.T) {
	handler := httpserver.SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	cases := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "no-referrer",
	}
	for header, want := range cases {
		if got := rec.Header().Get(header); got != want {
			t.Errorf("expected %s: %s, got %q", header, want, got)
		}
	}
}
