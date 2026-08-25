package httpapi

import "time"

// WorkingHourResponse es la representación canónica de un tramo de
// horario laboral (HU-040, CA-040-01/02/03): exactamente id, isoWeekday,
// startsTime, durationMinutes, createdAt, updatedAt. Nunca barbershopId ni
// barberId: el tenant se deriva de SessionCookie y el barbero de la ruta.
//
// ADVERTENCIA: postgres.workingHourResponseWire (schedule/postgres/
// repository.go) declara EXACTAMENTE la misma forma (mismos nombres de
// campo JSON, mismo orden, mismos tipos) para construir el cuerpo que la
// idempotencia persiste en el alta. Un cambio aquí debe reflejarse ahí en
// el mismo commit.
type WorkingHourResponse struct {
	ID              string    `json:"id"`
	ISOWeekday      int       `json:"isoWeekday"`
	StartsTime      string    `json:"startsTime"`
	DurationMinutes int       `json:"durationMinutes"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// WorkingHourListResponse es la página paginada por cursor de
// GET /private/barbers/{barberId}/working-hours (CA-040-01). NextCursor es
// un puntero para que "sin página siguiente" serialice como `null`
// explícito, nunca como una cadena vacía.
type WorkingHourListResponse struct {
	Items      []WorkingHourResponse `json:"items"`
	NextCursor *string               `json:"nextCursor"`
}

// CreateWorkingHourRequest es el cuerpo de
// POST /private/barbers/{barberId}/working-hours (CA-040-02). Cerrado: el
// handler rechaza cualquier campo desconocido. Nunca declara barbershopId
// ni barberId: el tenant se deriva de SessionCookie, el barbero de la ruta.
type CreateWorkingHourRequest struct {
	ISOWeekday      int    `json:"isoWeekday"`
	StartsTime      string `json:"startsTime"`
	DurationMinutes int    `json:"durationMinutes"`
}

// UpdateWorkingHourRequest es el cuerpo de
// PATCH /private/barbers/{barberId}/working-hours/{workingHourId}
// (CA-040-04). Los tres campos son obligatorios: reemplaza el intervalo
// completo, mismo criterio que UpdateWorkingHourRequest del contrato
// OpenAPI.
type UpdateWorkingHourRequest struct {
	ISOWeekday      int    `json:"isoWeekday"`
	StartsTime      string `json:"startsTime"`
	DurationMinutes int    `json:"durationMinutes"`
}
