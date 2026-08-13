package httpserver

import (
	"net/http"

	"system-barbershop/internal/platform/idempotency"
)

// IdempotencyKeyHeader es la cabecera HTTP que transporta la clave de
// idempotencia de una escritura crítica (RN-IDE-01, DEC-043). Requests de
// creación crítica la declaran obligatoria (docs/06-api/estandar-openapi.md
// sección 10).
const IdempotencyKeyHeader = "Idempotency-Key"

// IdempotencyKeyFromRequest valida la cabecera Idempotency-Key de r. Es el
// único punto donde ese valor cruza de HTTP hacia idempotency.Key —el
// paquete idempotency nunca importa net/http—, igual que RequestID es el
// único punto que traduce X-Request-Id hacia el contexto tipado.
func IdempotencyKeyFromRequest(r *http.Request) (idempotency.Key, error) {
	return idempotency.ParseKey(r.Header.Get(IdempotencyKeyHeader))
}

// IdempotencyFingerprint calcula la huella canónica de r a partir de su
// método, su ruta ya resuelta (r.URL.Path, sin query string) y el cuerpo
// EXACTO recibido en body. Ver idempotency.ComputeFingerprint para la forma
// canónica completa y por qué el cuerpo debe ser el crudo, no una
// re-serialización.
func IdempotencyFingerprint(r *http.Request, body []byte) idempotency.Fingerprint {
	return idempotency.ComputeFingerprint(r.Method, r.URL.Path, body)
}

// WriteStoredResponse reproduce byte a byte una StoredResponse
// (idempotency.OutcomeReplay, CA-004-01): mismo status, mismo Content-Type
// y mismo cuerpo que la ejecución original, sin volver a ejecutar el
// efecto.
func WriteStoredResponse(w http.ResponseWriter, resp idempotency.StoredResponse) {
	w.Header().Set("Content-Type", resp.ContentType)
	w.WriteHeader(resp.Status)
	_, _ = w.Write([]byte(resp.Body))
}
