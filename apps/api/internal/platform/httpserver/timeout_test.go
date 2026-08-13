package httpserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"system-barbershop/internal/platform/httpserver"
)

func TestTimeout_CancelsContextWithinConfiguredDuration(t *testing.T) {
	const budget = 30 * time.Millisecond

	done := make(chan struct{})
	handler := httpserver.Timeout(budget)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		close(done)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	start := time.Now()
	handler.ServeHTTP(rec, req)
	elapsed := time.Since(start)

	select {
	case <-done:
	default:
		t.Fatal("expected the handler to observe context cancellation")
	}
	if elapsed > budget+500*time.Millisecond {
		t.Fatalf("expected cancellation close to the %s budget, took %s", budget, elapsed)
	}
}

func TestTimeout_DoesNotCancelBeforeHandlerFinishes(t *testing.T) {
	handler := httpserver.Timeout(time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			t.Error("context was canceled before the handler finished")
		default:
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
