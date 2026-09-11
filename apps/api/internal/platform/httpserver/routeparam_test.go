package httpserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"system-barbershop/internal/platform/httpserver"
)

// TestRequestWithURLParam_Chained_AccumulatesBothParams cubre HU-092: una
// ruta con dos parámetros ({slug} y {serviceId}) necesita encadenar dos
// llamadas a RequestWithURLParam en las pruebas de httpapi; la segunda
// llamada no debe borrar el parámetro que dejó la primera.
func TestRequestWithURLParam_Chained_AccumulatesBothParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = httpserver.RequestWithURLParam(req, "slug", "barberia-ejemplo")
	req = httpserver.RequestWithURLParam(req, "serviceId", "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4")

	if got := httpserver.URLParam(req, "slug"); got != "barberia-ejemplo" {
		t.Fatalf("expected slug=%q to survive chaining, got %q", "barberia-ejemplo", got)
	}
	if got := httpserver.URLParam(req, "serviceId"); got != "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4" {
		t.Fatalf("expected serviceId to be set, got %q", got)
	}
}
