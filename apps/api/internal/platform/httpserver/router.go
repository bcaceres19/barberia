package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"system-barbershop/internal/platform/apperr"
)

// NewRouter monta las tres audiencias del contrato —/api/v1/public,
// /api/v1/customer y /api/v1/private— todavía sin operaciones de negocio
// (HU-003 no las incluye) y aplica el tramo del orden de middleware que no
// depende de autenticación, tenant ni idempotencia
// (docs/04-arquitectura/backend-go.md sección 6, pasos 1 a 6):
//
//  1. identificador de solicitud;
//  2. recuperación controlada de panic;
//  3. límite de tamaño del cuerpo;
//  4. timeout y cancelación por context.Context;
//  5. registro estructurado con datos personales redactados;
//  6. cabeceras de seguridad.
//
// Los pasos 7 a 11 (rate limit, autenticación, selección de barbería,
// tenant/RLS, idempotencia) llegan con las historias que los necesitan; no
// se agregan aquí sin esa necesidad real. cmd/api monta sobre el *chi.Mux
// resultante las rutas concretas (por ejemplo /health) a medida que
// existen, según el comentario de wiring en cmd/api/main.go.
func NewRouter(logger *slog.Logger) *chi.Mux {
	r := chi.NewRouter()
	r.Use(
		RequestID,
		Recover(logger),
		BodyLimit,
		Timeout(DefaultRequestTimeout),
		RequestLogger(logger),
		SecurityHeaders,
	)

	r.Route("/api/v1/public", func(chi.Router) {})
	r.Route("/api/v1/customer", func(chi.Router) {})
	r.Route("/api/v1/private", func(chi.Router) {})

	// Chi responde texto plano en sus manejadores por defecto de 404 y 405.
	// CA-003-01 exige application/problem+json en TODA respuesta de error
	// del API, sin excepción para una ruta desconocida, así que ambos se
	// reemplazan por el mismo traductor central que usa el resto del
	// middleware.
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		WriteProblem(w, Translate(apperr.NotFound("recurso no encontrado"), RequestIDFromContext(r.Context())))
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		WriteProblem(w, Translate(apperr.NotFound("recurso no encontrado"), RequestIDFromContext(r.Context())))
	})

	return r
}
