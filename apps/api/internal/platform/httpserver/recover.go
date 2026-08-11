package httpserver

import (
	"log/slog"
	"net/http"
)

// Recover intercepta un panic de un handler downstream, lo registra con el
// request_id vigente y responde 500 sin filtrar detalles internos, según el
// orden de middleware de docs/04-arquitectura/backend-go.md.
func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recuperado",
						"request_id", RequestIDFromContext(r.Context()),
						"panic", rec,
					)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
