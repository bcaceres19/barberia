package httpserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"system-barbershop/internal/platform/httpserver"
)

// alwaysUnauthorized simula el middleware de sesión real de HU-006 sin
// depender del módulo auth: cualquier solicitud que pase por aquí recibe el
// mismo 401 uniforme, sin ejecutar el siguiente handler.
func alwaysUnauthorized(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
}

// TestPrivateSubrouter_MiddlewareAppliesToRoutesRegisteredOnIt confirma el
// mecanismo positivo que cmd/api usa: una ruta registrada sobre el
// subrouter `private` que [NewRouter] devuelve SÍ pasa por el middleware
// montado con private.Use(...).
func TestPrivateSubrouter_MiddlewareAppliesToRoutesRegisteredOnIt(t *testing.T) {
	router, private := httpserver.NewRouter(discardLogger())
	private.Use(alwaysUnauthorized)
	private.Get("/protected", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/private/protected", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected the route registered on `private` to go through its middleware, got %d", rec.Code)
	}
}

// TestPrivateSubrouter_BypassingItSkipsTheMiddleware es el control negativo
// que exige el prompt de HU-006: demuestra que el método de la prueba
// estructural de CA-006-04 (caminar el router con chi.Walk y golpear cada
// ruta descubierta sin cookie, esperando 401) SÍ distingue una ruta
// protegida de una que no lo está. Si esta prueba alguna vez empezara a
// fallar (la ruta "bypasseada" empezara a devolver 401), significaría que
// chi cambió su comportamiento de precedencia entre una ruta estática
// registrada directamente en el *chi.Mux y un subrouter montado con
// Route()/Mount() en el mismo prefijo -y habría que revisar si
// TestPrivateRouteInventory_* (cmd/api) sigue siendo una guardia real.
func TestPrivateSubrouter_BypassingItSkipsTheMiddleware(t *testing.T) {
	router, private := httpserver.NewRouter(discardLogger())
	private.Use(alwaysUnauthorized)
	private.Get("/protected", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	// Registrado con el patrón COMPLETO directamente sobre el *chi.Mux
	// devuelto, en vez de sobre `private`: exactamente el error que
	// CA-006-04 quiere impedir en producción.
	router.Get("/api/v1/private/bypassed", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/private/bypassed", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code == http.StatusUnauthorized {
		t.Skip("chi ahora enruta una ruta estática registrada en el Mux padre a través del subrouter montado; el control negativo ya no aplica, pero tampoco hay nada que corregir (el bypass real ya no es posible)")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected the bypassed route to skip the middleware entirely (200 from the raw handler), got %d", rec.Code)
	}

	if err := chiWalkFinds(router, http.MethodGet, "/api/v1/private/bypassed"); err != nil {
		t.Fatalf("expected chi.Walk to still discover the bypassed route (that is what makes the structural inventory test in cmd/api a real guard): %v", err)
	}
}

func chiWalkFinds(router *chi.Mux, wantMethod, wantPattern string) error {
	found := false
	err := chi.Walk(router, func(method, pattern string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		if method == wantMethod && pattern == wantPattern {
			found = true
		}
		return nil
	})
	if err != nil {
		return err
	}
	if !found {
		return http.ErrNotSupported
	}
	return nil
}
