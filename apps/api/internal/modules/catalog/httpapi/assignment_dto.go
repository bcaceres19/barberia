package httpapi

import "time"

// AssignmentResponse es la representación canónica de una asignación
// barbero-servicio (HU-023, CA-023-07): exactamente barberId, serviceId y
// createdAt. Nunca nombre, duración, precio ni estado del barbero o del
// servicio: un consumidor que necesite mostrarlos cruza por identificador
// contra GET /private/barbers y GET /private/services, que ya los exponen.
type AssignmentResponse struct {
	BarberID  string    `json:"barberId"`
	ServiceID string    `json:"serviceId"`
	CreatedAt time.Time `json:"createdAt"`
}

// AssignmentListResponse es la página paginada por cursor de
// GET /private/barbers/{barberId}/services (CA-023-01). NextCursor es un
// puntero para que "sin página siguiente" serialice como `null` explícito,
// nunca como una cadena vacía (mismo patrón que BarberListResponse/
// ServiceListResponse).
type AssignmentListResponse struct {
	Items      []AssignmentResponse `json:"items"`
	NextCursor *string              `json:"nextCursor"`
}
