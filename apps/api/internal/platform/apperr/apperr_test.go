package apperr_test

import (
	"errors"
	"fmt"
	"testing"

	"system-barbershop/internal/platform/apperr"
)

func TestNotFound_CarriesSafeMessage(t *testing.T) {
	err := apperr.NotFound("recurso no disponible")

	got, ok := apperr.As(err)
	if !ok {
		t.Fatal("expected apperr.As to recognize the error")
	}
	if got.Kind != apperr.KindNotFound {
		t.Fatalf("expected KindNotFound, got %q", got.Kind)
	}
	if got.Message != "recurso no disponible" {
		t.Fatalf("unexpected message: %q", got.Message)
	}
}

func TestInternal_WrapsCauseWithoutExposingIt(t *testing.T) {
	cause := errors.New("pq: syntax error at or near \"SELEC\"")
	err := apperr.Internal(cause)

	got, ok := apperr.As(err)
	if !ok {
		t.Fatal("expected apperr.As to recognize the error")
	}
	if got.Kind != apperr.KindInternal {
		t.Fatalf("expected KindInternal, got %q", got.Kind)
	}
	if !errors.Is(err, cause) {
		t.Fatal("expected Unwrap to expose the cause for internal error-chain inspection")
	}
}

func TestAs_ReturnsFalseForForeignError(t *testing.T) {
	_, ok := apperr.As(fmt.Errorf("plain error"))
	if ok {
		t.Fatal("expected apperr.As to return false for a non-apperr error")
	}
}

func TestAs_UnwrapsWrappedError(t *testing.T) {
	base := apperr.NotFound("x")
	wrapped := fmt.Errorf("contexto adicional: %w", base)

	got, ok := apperr.As(wrapped)
	if !ok {
		t.Fatal("expected apperr.As to unwrap through fmt.Errorf %w")
	}
	if got.Kind != apperr.KindNotFound {
		t.Fatalf("expected KindNotFound, got %q", got.Kind)
	}
}
