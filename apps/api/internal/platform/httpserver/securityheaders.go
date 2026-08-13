package httpserver

import "net/http"

// SecurityHeaders añade el set mínimo de cabeceras de seguridad razonable
// para una API que solo sirve JSON (nunca HTML de terceros ni recursos
// incrustables): evita que un navegador adivine un Content-Type distinto al
// declarado, evita que la respuesta se incruste en un frame ajeno y evita
// filtrar la URL de origen en la cabecera Referer de una navegación
// saliente. No sustituye HSTS ni CSP, que dependen de dónde y cómo se
// exponga el servicio en producción y no están decididos todavía
// (docs/04-arquitectura/backend-go.md sección 6 paso 6).
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}
