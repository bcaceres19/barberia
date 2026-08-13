package httpserver

import (
	"context"
	"net/http"
	"time"
)

// DefaultRequestTimeout limita cuánto puede tardar el procesamiento de una
// solicitud antes de que su context.Context se cancele, para que una
// consulta PostgreSQL downstream se cancele con ella
// (docs/04-arquitectura/backend-go.md sección 9). Es menor que los timeouts
// de servidor de [New] para que la cancelación por contexto llegue primero
// y deje registrar la causa real.
const DefaultRequestTimeout = 10 * time.Second

// Timeout deriva un context.Context con plazo d para cada solicitud. No
// escribe ninguna respuesta por sí mismo: cede el control al handler
// downstream, que debe observar ctx.Done() en su trabajo de red o base de
// datos.
func Timeout(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
