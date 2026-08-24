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

func TestInvalid_CarriesSafeMessage(t *testing.T) {
	err := apperr.Invalid("Idempotency-Key tiene un formato inválido")

	got, ok := apperr.As(err)
	if !ok {
		t.Fatal("expected apperr.As to recognize the error")
	}
	if got.Kind != apperr.KindInvalid {
		t.Fatalf("expected KindInvalid, got %q", got.Kind)
	}
	if got.Message != "Idempotency-Key tiene un formato inválido" {
		t.Fatalf("unexpected message: %q", got.Message)
	}
}

func TestIdempotencyConflict_CarriesSafeMessage(t *testing.T) {
	err := apperr.IdempotencyConflict("la clave ya se usó con contenido distinto")

	got, ok := apperr.As(err)
	if !ok {
		t.Fatal("expected apperr.As to recognize the error")
	}
	if got.Kind != apperr.KindIdempotencyConflict {
		t.Fatalf("expected KindIdempotencyConflict, got %q", got.Kind)
	}
}

func TestIdempotencyLocked_CarriesSafeMessage(t *testing.T) {
	err := apperr.IdempotencyLocked("operación en curso")

	got, ok := apperr.As(err)
	if !ok {
		t.Fatal("expected apperr.As to recognize the error")
	}
	if got.Kind != apperr.KindIdempotencyLocked {
		t.Fatalf("expected KindIdempotencyLocked, got %q", got.Kind)
	}
}

func TestValidation_CarriesSafeMessage(t *testing.T) {
	err := apperr.Validation("email es obligatorio")

	got, ok := apperr.As(err)
	if !ok {
		t.Fatal("expected apperr.As to recognize the error")
	}
	if got.Kind != apperr.KindValidation {
		t.Fatalf("expected KindValidation, got %q", got.Kind)
	}
}

func TestUnauthorized_CarriesSafeMessage(t *testing.T) {
	err := apperr.Unauthorized("correo o contraseña incorrectos")

	got, ok := apperr.As(err)
	if !ok {
		t.Fatal("expected apperr.As to recognize the error")
	}
	if got.Kind != apperr.KindUnauthorized {
		t.Fatalf("expected KindUnauthorized, got %q", got.Kind)
	}
	if got.Message != "correo o contraseña incorrectos" {
		t.Fatalf("unexpected message: %q", got.Message)
	}
}

func TestConflict_CarriesSafeMessage(t *testing.T) {
	err := apperr.Conflict("ya existe un servicio activo con ese nombre")

	got, ok := apperr.As(err)
	if !ok {
		t.Fatal("expected apperr.As to recognize the error")
	}
	if got.Kind != apperr.KindConflict {
		t.Fatalf("expected KindConflict, got %q", got.Kind)
	}
	if got.Message != "ya existe un servicio activo con ese nombre" {
		t.Fatalf("unexpected message: %q", got.Message)
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
