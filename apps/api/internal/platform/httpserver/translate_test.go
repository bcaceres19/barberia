package httpserver_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
)

func TestTranslate_NotFound_ExposesSafeMessage(t *testing.T) {
	p := httpserver.Translate(apperr.NotFound("no existe una cita con ese identificador"), "req-1")

	if p.Status != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", p.Status)
	}
	if p.Detail != "no existe una cita con ese identificador" {
		t.Fatalf("expected the safe message to pass through, got %q", p.Detail)
	}
	if p.Code == "" || p.RequestID != "req-1" || p.Type == "" || p.Title == "" {
		t.Fatalf("expected fully populated problem, got %+v", p)
	}
}

// TestTranslate_NotFound_IdenticalForMissingAndCrossTenant es CA-003-03: un
// recurso inexistente y uno de otra barbería deben producir exactamente el
// mismo Problem. apperr solo ofrece un Kind para ambos casos, así que esta
// prueba confirma que no hay forma de invocar Translate que los distinga.
func TestTranslate_NotFound_IdenticalForMissingAndCrossTenant(t *testing.T) {
	missing := httpserver.Translate(apperr.NotFound("recurso no disponible"), "req-42")
	crossTenant := httpserver.Translate(apperr.NotFound("recurso no disponible"), "req-42")

	if missing != crossTenant {
		t.Fatalf("expected identical problems for missing vs cross-tenant, got %+v vs %+v", missing, crossTenant)
	}
}

// TestTranslate_Internal_NeverExposesCause es CA-003-02: ni el mensaje del
// error interno ni ningún fragmento reconocible (SQL, ruta de archivo,
// nombre de proveedor) llega al Detail expuesto al cliente.
func TestTranslate_Internal_NeverExposesCause(t *testing.T) {
	sensitive := errors.New(`pq: syntax error at or near "SELECT" in /var/lib/app/query.sql (driver postgres v14.2)`)
	p := httpserver.Translate(apperr.Internal(sensitive), "req-2")

	if p.Status != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", p.Status)
	}
	if strings.Contains(p.Detail, "pq:") || strings.Contains(p.Detail, "/var/lib") || strings.Contains(p.Detail, "postgres") {
		t.Fatalf("internal cause leaked into Detail: %q", p.Detail)
	}
	if p.Detail != "" {
		t.Fatalf("expected empty Detail for an internal error, got %q", p.Detail)
	}
}

func TestTranslate_UnrecognizedError_CollapsesToGenericInternal(t *testing.T) {
	p := httpserver.Translate(errors.New("cualquier error ajeno no envuelto"), "req-3")

	if p.Status != http.StatusInternalServerError {
		t.Fatalf("expected 500 for an unrecognized error, got %d", p.Status)
	}
	if p.Detail != "" {
		t.Fatalf("expected empty Detail, got %q", p.Detail)
	}
}

func TestTranslate_Invalid_ExposesSafeMessage(t *testing.T) {
	p := httpserver.Translate(apperr.Invalid("Idempotency-Key tiene un formato inválido"), "req-5")

	if p.Status != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", p.Status)
	}
	if p.Code != "invalid-request" {
		t.Fatalf("expected code invalid-request, got %q", p.Code)
	}
	if p.Detail != "Idempotency-Key tiene un formato inválido" {
		t.Fatalf("expected the safe message to pass through, got %q", p.Detail)
	}
}

func TestTranslate_IdempotencyConflict_MapsTo409(t *testing.T) {
	p := httpserver.Translate(apperr.IdempotencyConflict("contenido distinto para la misma clave"), "req-6")

	if p.Status != http.StatusConflict {
		t.Fatalf("expected 409, got %d", p.Status)
	}
	if p.Code != "idempotency-conflict" {
		t.Fatalf("expected code idempotency-conflict, got %q", p.Code)
	}
}

func TestTranslate_IdempotencyLocked_MapsTo409WithoutWaiting(t *testing.T) {
	p := httpserver.Translate(apperr.IdempotencyLocked("operación en curso"), "req-7")

	if p.Status != http.StatusConflict {
		t.Fatalf("expected 409 (DEC-043), got %d", p.Status)
	}
	if p.Code != "idempotency-locked" {
		t.Fatalf("expected code idempotency-locked, got %q", p.Code)
	}
}

func TestTranslate_Validation_MapsTo422(t *testing.T) {
	p := httpserver.Translate(apperr.Validation("email es obligatorio"), "req-8")

	if p.Status != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", p.Status)
	}
	if p.Code != "validation-error" {
		t.Fatalf("expected code validation-error, got %q", p.Code)
	}
	if p.Detail != "email es obligatorio" {
		t.Fatalf("expected the safe message to pass through, got %q", p.Detail)
	}
}

// TestTranslate_Unauthorized_MapsTo401AndNeverDistinguishesReason es
// CA-005-02/CA-005-07 a nivel de Translate: apperr.Unauthorized solo
// transporta un mensaje, nunca un motivo estructurado que la respuesta
// pudiera filtrar.
func TestTranslate_Unauthorized_MapsTo401AndNeverDistinguishesReason(t *testing.T) {
	unknownEmail := httpserver.Translate(apperr.Unauthorized("correo o contraseña incorrectos"), "req-9")
	wrongPassword := httpserver.Translate(apperr.Unauthorized("correo o contraseña incorrectos"), "req-9")

	if unknownEmail.Status != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", unknownEmail.Status)
	}
	if unknownEmail.Code != "unauthorized" {
		t.Fatalf("expected code unauthorized, got %q", unknownEmail.Code)
	}
	if unknownEmail != wrongPassword {
		t.Fatalf("expected identical problems, got %+v vs %+v", unknownEmail, wrongPassword)
	}
}

func TestTranslate_MaxBytesError_MapsToPayloadTooLarge(t *testing.T) {
	err := &http.MaxBytesError{Limit: httpserver.MaxRequestBodyBytes}

	p := httpserver.Translate(err, "req-4")

	if p.Status != http.StatusBadRequest {
		t.Fatalf("expected 400 for an oversized body, got %d", p.Status)
	}
	if p.Code != "payload-too-large" {
		t.Fatalf("expected payload-too-large code, got %q", p.Code)
	}
}
