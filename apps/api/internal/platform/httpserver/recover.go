package httpserver

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

// Recover intercepta un panic de un handler downstream, lo registra con el
// request_id vigente y responde 500 en application/problem+json sin filtrar
// detalles internos, según el orden de middleware de
// docs/04-arquitectura/backend-go.md.
func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					requestID := RequestIDFromContext(r.Context())

					// NO se registra rec crudo: si un handler futuro hace panic
					// con un valor que interpole datos de negocio (p. ej. un
					// fmt.Errorf con el cuerpo de la solicitud), ese valor no
					// debe llegar al log solo por haber capturado el panic. El
					// tipo basta para diagnosticar dónde mirar.
					logger.Error("panic recuperado",
						"request_id", requestID,
						"panic_type", fmt.Sprintf("%T", rec),
					)

					WriteProblem(w, Translate(errors.New("panic recuperado en el límite HTTP"), requestID))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
