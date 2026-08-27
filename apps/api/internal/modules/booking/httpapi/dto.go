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
