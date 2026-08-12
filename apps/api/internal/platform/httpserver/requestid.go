package httpserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"regexp"
)

// RequestIDHeader es el encabezado que transporta el identificador de
// solicitud entre el cliente, los logs y las respuestas de error, según
// docs/06-api/estandar-openapi.md.
const RequestIDHeader = "X-Request-Id"

// requestIDPattern acota lo que se acepta de un cliente: caracteres seguros
// para una cabecera y para un log estructurado, longitud razonable para un
// identificador de correlación (cubre UUID, hex e ids de rastreo típicos de
// proxies). Cualquier otra cosa —vacío, control characters, longitud
// arbitraria— se descarta y se genera un id propio.
var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

type requestIDKey struct{}

// RequestID recorre el middleware base de todas las rutas: asigna un
// identificador si el cliente no envió uno o si el que envió no tiene un
// formato seguro, lo expone en el contexto tipado y lo refleja en la
// respuesta para poder correlacionar logs, transacción e intento de
// notificación, según docs/04-arquitectura/backend-go.md.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if !requestIDPattern.MatchString(id) {
			id = newRequestID()
		}

		w.Header().Set(RequestIDHeader, id)
		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFromContext devuelve el identificador de solicitud del contexto
// tipado, o una cadena vacía si el middleware [RequestID] no se aplicó.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

func newRequestID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		// crypto/rand solo falla si el generador del sistema no está
		// disponible; en ese caso preferimos una respuesta degradada a
		// bloquear la solicitud.
		return "unavailable"
	}
	return hex.EncodeToString(buf)
}
