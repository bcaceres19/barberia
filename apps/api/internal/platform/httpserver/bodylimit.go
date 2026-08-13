package httpserver

import "net/http"

// MaxRequestBodyBytes acota el cuerpo de cualquier solicitud. 1 MiB cubre
// con margen los payloads JSON del MVP (ninguna operación de negocio sube
// archivos); una ruta futura con una necesidad distinta la documenta y la
// ajusta en su propio middleware, según docs/04-arquitectura/backend-go.md
// sección 6 paso 3.
const MaxRequestBodyBytes = 1 << 20 // 1 MiB

// BodyLimit envuelve el cuerpo de la solicitud con http.MaxBytesReader. No
// responde nada por sí mismo: el primer intento de leer más allá del
// límite falla con *http.MaxBytesError, y [Translate] lo traduce a un
// problema seguro cuando el handler downstream reporta ese error.
func BodyLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodyBytes)
		next.ServeHTTP(w, r)
	})
}
