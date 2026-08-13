package httpserver

import (
	"encoding/json"
	"net/http"
)

// Problem es el cuerpo de error uniforme según RFC 9457
// (application/problem+json), con los campos adicionales que
// docs/06-api/estandar-openapi.md sección 12 exige para todo el proyecto:
// code (para lógica del cliente) y requestId (correlación con logs y
// soporte). Ningún llamador debe construir un Problem a mano fuera de
// [Translate]: eso es lo único que garantiza que Detail nunca lleve SQL,
// rutas de archivo, nombres de proveedor ni versiones (CA-003-02).
type Problem struct {
	Type      string `json:"type,omitempty"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
	Detail    string `json:"detail,omitempty"`
	Instance  string `json:"instance,omitempty"`
	Code      string `json:"code"`
	RequestID string `json:"requestId"`
}

// WriteProblem escribe p como application/problem+json con el código de
// estado de p.Status. No decide qué información es segura para el
// cliente: eso es responsabilidad del llamador (CA-003-02: nunca SQL,
// rutas de archivo, nombres de excepción ni versiones de dependencias).
func WriteProblem(w http.ResponseWriter, p Problem) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	_ = json.NewEncoder(w).Encode(p)
}
