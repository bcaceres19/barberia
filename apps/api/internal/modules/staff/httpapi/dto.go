package httpapi

import "time"

// BarberResponse es la representación canónica de un barbero (CA-021-01,
// CA-021-07): exactamente id, fullName, createdAt, updatedAt. Nunca active,
// deletedAt, sortOrder, staffUserId, servicios ni horarios (DEC-047).
//
// ADVERTENCIA: postgres.barberResponseWire (staff/postgres/repository.go)
// declara EXACTAMENTE la misma forma (mismos nombres de campo JSON, mismo
// orden, mismos tipos) para construir el cuerpo que la idempotencia
// persiste en el alta. Un cambio aquí debe reflejarse ahí en el mismo
// commit; ver TestCreate_StoredResponseBody_MatchesHTTPAPIWireShape en
// postgres/repository_test.go.
type BarberResponse struct {
	ID        string    `json:"id"`
	FullName  string    `json:"fullName"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// BarberListResponse es la página paginada por cursor de GET /private/barbers
// (CA-021-02). NextCursor es un puntero para que "sin página siguiente"
// serialice como `null` explícito, nunca como una cadena vacía.
type BarberListResponse struct {
	Items      []BarberResponse `json:"items"`
	NextCursor *string          `json:"nextCursor"`
}

// CreateBarberRequest es el cuerpo de POST /private/barbers (CA-021-02).
// Cerrado: el handler rechaza cualquier campo desconocido. Nunca declara
// barbershopId: el tenant se deriva exclusivamente de auth.Principal.
type CreateBarberRequest struct {
	FullName string `json:"fullName"`
}

// UpdateBarberRequest es el cuerpo de PATCH /private/barbers/{barberId}
// (CA-021-04). Cerrado y limitado al único campo editable de HU-021: nunca
// active, deletedAt, sortOrder, staffUserId, servicios ni horarios
// (DEC-047, CA-021-07).
type UpdateBarberRequest struct {
	FullName string `json:"fullName"`
}
