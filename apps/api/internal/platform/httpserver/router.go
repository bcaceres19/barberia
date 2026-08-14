package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"system-barbershop/internal/platform/apperr"
)

// NewRouter monta las tres audiencias del contrato —/api/v1/public,
// /api/v1/customer y /api/v1/private— y aplica el tramo del orden de
// middleware que no depende de autenticación, tenant ni idempotencia
// (docs/04-arquitectura/backend-go.md sección 6, pasos 1 a 6):
//
//  1. identificador de solicitud;
//  2. recuperación controlada de panic;
//  3. límite de tamaño del cuerpo;
//  4. timeout y cancelación por context.Context;
//  5. registro estructurado con datos personales redactados;
//  6. cabeceras de seguridad.
//
// El segundo valor devuelto es el subrouter real montado en
// /api/v1/private: es el ÚNICO lugar donde cmd/api debe registrar
// middleware y rutas privadas (paso 8, autenticación) para que
// docs/04-arquitectura/backend-go.md CA-006-04 se cumpla sin excepciones:
// toda ruta bajo ese prefijo pasa por el middleware que se monte con
// private.Use(...) ANTES de registrar ninguna ruta sobre él. Público y
// cliente no necesitan ese mismo mecanismo todavía (ninguna historia les
// exige middleware propio); sus grupos quedan como marcadores vacíos, igual
// que antes de HU-006, y cmd/api sigue registrando sus rutas con el patrón
// completo sobre el *chi.Mux devuelto como primer valor (mismo patrón que
// /api/v1/public/auth/login desde HU-005).
func NewRouter(logger *slog.Logger) (*chi.Mux, chi.Router) {
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

	var private chi.Router
	r.Route("/api/v1/private", func(pr chi.Router) {
		private = pr
	})

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

	return r, private
}
