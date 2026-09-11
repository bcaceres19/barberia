package httpserver

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// URLParam es el único punto donde un handler de módulo puede leer un
// parámetro de ruta ({barberId}, y cualquier otro que HU-021 en adelante
// necesite): httpserver es la ÚNICA capa autorizada a importar Chi
// (DEC-034, TestDomainAndServicesDoNotImportChi en
// internal/platform/archtest/chi_boundary_test.go), así que ningún
// internal/modules/*/httpapi importa chi directamente, igual que ya hace
// con RequestIDFromContext e IdempotencyKeyFromRequest para otras
// cabeceras/valores que Chi o net/http exponen.
func URLParam(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}

// RequestWithURLParam devuelve una copia de r con name=value inyectado como
// si Chi ya lo hubiera resuelto de la ruta. Existe para que las pruebas de
// internal/modules/*/httpapi (que NO pueden importar Chi,
// TestDomainAndServicesDoNotImportChi cubre también archivos _test.go)
// puedan construir un *http.Request de prueba con un parámetro de ruta sin
// levantar el router real; producción nunca la llama (chi.Mux ya hace este
// trabajo). Reutiliza el *chi.Context ya presente en r (si lo hay) en vez de
// reemplazarlo: encadenar dos llamadas para inyectar dos parámetros de ruta
// distintos (p. ej. `slug` y `serviceId`, HU-092) acumula ambos en el mismo
// contexto en lugar de que la segunda llamada borre el primero.
func RequestWithURLParam(r *http.Request, name, value string) *http.Request {
	routeCtx, ok := r.Context().Value(chi.RouteCtxKey).(*chi.Context)
	if !ok || routeCtx == nil {
		routeCtx = chi.NewRouteContext()
	}
	routeCtx.URLParams.Add(name, value)
	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx)
	return r.WithContext(ctx)
}
