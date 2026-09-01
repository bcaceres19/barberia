package httpapi

import "time"

// CreateManualAppointmentRequest es el cuerpo JSON de
// POST /private/appointments (HU-061): exactamente los campos que el
// barbero aporta al registrar un turno manual. Ningún campo de estado,
// origen, actor, barbershopId, snapshot de servicio ni identificador de
// cliente existente: el servidor los deriva o los decide (DEC-071,
// DEC-072).
type CreateManualAppointmentRequest struct {
	BarberID         string  `json:"barberId"`
	ServiceID        string  `json:"serviceId"`
	AttendeeName     string  `json:"attendeeName"`
	CustomerFullName string  `json:"customerFullName"`
	CustomerPhone    *string `json:"customerPhone"`
	CustomerEmail    *string `json:"customerEmail"`
	CustomerNote     *string `json:"customerNote"`
	// StartsAt es el instante civil "AAAA-MM-DDTHH:MM:SS" (sin zona) que el
	// barbero eligió en su propio reloj, interpretado contra la zona IANA
	// vigente de la barbería (RN-DIS-07).
	StartsAt string `json:"startsAt"`
}

// AppointmentResponse es la representación canónica de la cita recién
// creada. Los nombres de campo coinciden exactamente con
// bookingpostgres.manualAppointmentResponseWire: ambos serializan la misma
// forma, una para el cliente HTTP en vivo y otra para el cuerpo almacenado
// de una repetición idempotente (RN-IDE-01).
type AppointmentResponse struct {
	ID              string    `json:"id"`
	BarberID        string    `json:"barberId"`
	ServiceID       string    `json:"serviceId"`
	CustomerID      string    `json:"customerId"`
	AttendeeName    string    `json:"attendeeName"`
	StartsAt        time.Time `json:"startsAt"`
	EndsAt          time.Time `json:"endsAt"`
	Status          string    `json:"status"`
	Origin          string    `json:"origin"`
	ServiceName     string    `json:"serviceName"`
	DurationMinutes int       `json:"durationMinutes"`
	PriceAmount     string    `json:"priceAmount"`
	Currency        string    `json:"currency"`
	CustomerNote    *string   `json:"customerNote"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// RescheduleAppointmentRequest es el cuerpo JSON de
// POST /private/appointments/{appointmentId}/reschedule (HU-065, T2):
// exactamente el nuevo inicio. Ningún otro campo: barbero, servicio,
// duración, precio, cliente, persona, origen, nota y estado permanecen
// iguales (el servidor los conserva, nunca los acepta de este cuerpo).
type RescheduleAppointmentRequest struct {
	// StartsAt es el instante civil "AAAA-MM-DDTHH:MM:SS" (sin zona) que el
	// barbero eligió en su propio reloj, mismo formato que
	// CreateManualAppointmentRequest.StartsAt (RN-DIS-07).
	StartsAt string `json:"startsAt"`
}

// AppointmentRescheduledResponse es la representación de la cita ya
// reprogramada (o, en un no-op del mismo intervalo, la representación
// vigente sin cambios). Deliberadamente SIN updatedAt crudo (HU-064):
// versionToken es el único dato de concurrencia expuesto, ya actualizado
// para una reprogramación posterior sin volver a pedir el detalle.
type AppointmentRescheduledResponse struct {
	ID              string    `json:"id"`
	BarberID        string    `json:"barberId"`
	ServiceID       string    `json:"serviceId"`
	CustomerID      string    `json:"customerId"`
	AttendeeName    string    `json:"attendeeName"`
	StartsAt        time.Time `json:"startsAt"`
	EndsAt          time.Time `json:"endsAt"`
	Status          string    `json:"status"`
	Origin          string    `json:"origin"`
	ServiceName     string    `json:"serviceName"`
	DurationMinutes int       `json:"durationMinutes"`
	PriceAmount     string    `json:"priceAmount"`
	Currency        string    `json:"currency"`
	CustomerNote    *string   `json:"customerNote"`
	VersionToken    string    `json:"versionToken"`
	CreatedAt       time.Time `json:"createdAt"`
}

// DailyAgendaResponse es el cuerpo de
// GET /private/barbers/{barberId}/appointments/daily-agenda (HU-062):
// lista cronológica cerrada, sin cursor ni siguiente página (CA-062-07,
// un solo día de un solo barbero está acotado por diseño).
type DailyAgendaResponse struct {
	Items []DailyAgendaEntryResponse `json:"items"`
}

// DailyAgendaEntryResponse es la proyección mínima de un turno que HU-062
// expone (CA-062-05): nunca teléfono, correo, nota ni customerId. barberId
// no aparece: la ruta ya lo fija (DEC-074) y repetirlo sería un campo
// interno innecesario.
type DailyAgendaEntryResponse struct {
	ID              string    `json:"id"`
	AttendeeName    string    `json:"attendeeName"`
	StartsAt        time.Time `json:"startsAt"`
	EndsAt          time.Time `json:"endsAt"`
	Status          string    `json:"status"`
	Origin          string    `json:"origin"`
	ServiceName     string    `json:"serviceName"`
	DurationMinutes int       `json:"durationMinutes"`
	PriceAmount     string    `json:"priceAmount"`
	Currency        string    `json:"currency"`
}

// AppointmentDetailResponse es el cuerpo de
// GET /private/appointments/{appointmentId} (HU-064, CA-064-01 a CA-064-04).
// Deliberadamente SIN customerId, IDs de actor ni updatedAt crudo (trabajo
// requerido, "Fuera de alcance"): VersionToken es el único dato de
// concurrencia expuesto, opaco (booking.EncodeVersionToken).
type AppointmentDetailResponse struct {
	ID               string    `json:"id"`
	BarberID         string    `json:"barberId"`
	BarberFullName   string    `json:"barberFullName"`
	AttendeeName     string    `json:"attendeeName"`
	CustomerFullName string    `json:"customerFullName"`
	CustomerPhone    *string   `json:"customerPhone"`
	CustomerEmail    *string   `json:"customerEmail"`
	CustomerNote     *string   `json:"customerNote"`
	StartsAt         time.Time `json:"startsAt"`
	EndsAt           time.Time `json:"endsAt"`
	Status           string    `json:"status"`
	Origin           string    `json:"origin"`
	ServiceName      string    `json:"serviceName"`
	DurationMinutes  int       `json:"durationMinutes"`
	PriceAmount      string    `json:"priceAmount"`
	Currency         string    `json:"currency"`
	VersionToken     string    `json:"versionToken"`
	CreatedAt        time.Time `json:"createdAt"`
}

// AppointmentHistoryResponse es el cuerpo de
// GET /private/appointments/{appointmentId}/history (HU-064, CA-064-05):
// página paginada por cursor opaco, mismo criterio que BarberListResponse
// (staff/httpapi). NextCursor es nil cuando esta página es la última.
type AppointmentHistoryResponse struct {
	Items      []AppointmentHistoryEntryResponse `json:"items"`
	NextCursor *string                           `json:"nextCursor"`
}

// AppointmentHistoryEntryResponse es una entrada del historial ya traducida
// a un nombre visible de actor: nunca correo ni ningún identificador interno
// (RN-HIS-01).
type AppointmentHistoryEntryResponse struct {
	ID         string                             `json:"id"`
	EventType  string                             `json:"eventType"`
	ActorType  string                             `json:"actorType"`
	ActorLabel string                             `json:"actorLabel"`
	Reason     *string                            `json:"reason"`
	OccurredAt time.Time                          `json:"occurredAt"`
	Changes    []AppointmentHistoryChangeResponse `json:"changes"`
}

// AppointmentHistoryChangeResponse es un campo modificado con su valor
// anterior y nuevo (RN-HIS-01), proyección directa de
// appointment_history_change.
type AppointmentHistoryChangeResponse struct {
	FieldName     string  `json:"fieldName"`
	PreviousValue *string `json:"previousValue"`
	NewValue      *string `json:"newValue"`
}
