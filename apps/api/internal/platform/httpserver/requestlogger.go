package httpserver

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// RequestLogger registra cada solicitud con una lista de campos cerrada:
// request_id, método, patrón de ruta, estado y duración. Nunca registra
// r.URL.Path: una audiencia como /api/v1/customer/appointments/{token}/...
// lleva un token en la ruta, y el patrón coincidente (p. ej.
// "/api/v1/customer/appointments/{token}") identifica la operación sin
// exponer el valor real (RN-DAT-02, CA-003-04). Tampoco registra query,
// cabeceras ni cuerpo: ninguno de esos campos está en la lista permitida.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			route := "unmatched"
			if rctx := chi.RouteContext(r.Context()); rctx != nil {
				if pattern := rctx.RoutePattern(); pattern != "" {
					route = pattern
				}
			}

			logger.Info("solicitud http",
				"request_id", RequestIDFromContext(r.Context()),
				"method", r.Method,
				"route", route,
				"status", ww.Status(),
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
