package httpserver

import (
	"encoding/json"
	"net/http"
)

// Problem es el cuerpo de error uniforme según RFC 9457
// (application/problem+json). Es la base reutilizable mínima para
// cualquier respuesta de error del API; HU-003 la completa con el
// catálogo de tipos de error del contrato OpenAPI (CA-003-01, CA-003-06).
type Problem struct {
	Type     string `json:"type,omitempty"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
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
