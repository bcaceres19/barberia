package httpserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// RequestIDHeader es el encabezado que transporta el identificador de
// solicitud entre el cliente, los logs y las respuestas de error, según
// docs/06-api/estandar-openapi.md.
const RequestIDHeader = "X-Request-Id"

type requestIDKey struct{}

// RequestID recorre el middleware base de todas las rutas: asigna un
// identificador si el cliente no envió uno, lo expone en el contexto tipado
// y lo refleja en la respuesta para poder correlacionar logs, transacción e
// intento de notificación, según docs/04-arquitectura/backend-go.md.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if id == "" {
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
