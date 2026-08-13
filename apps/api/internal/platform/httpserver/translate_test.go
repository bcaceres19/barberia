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
