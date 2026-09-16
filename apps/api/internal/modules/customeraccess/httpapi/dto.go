package httpapi

import "time"

// CustomerAppointmentResponse es la proyección mínima que HU-098 expone
// (CA-098-01, CA-098-03), forma exacta del componente OpenAPI
// CustomerAppointmentResponse.yaml: nunca un identificador interno,
// contacto del cliente, nota, historial ni ningún otro turno.
// cancellationDeadlineMinutes/lateCancellation* son los mismos nombres que
// shops/httpapi.BookingPolicyResponse (HU-093): la política vigente de
// cancelación, no un snapshot fijado al reservar.
type CustomerAppointmentResponse struct {
	BarbershopName                 string    `json:"barbershopName"`
	Timezone                       string    `json:"timezone"`
	AttendeeName                   string    `json:"attendeeName"`
	ServiceName                    string    `json:"serviceName"`
	DurationMinutes                int       `json:"durationMinutes"`
	BarberName                     string    `json:"barberName"`
	StartsAt                       time.Time `json:"startsAt"`
	EndsAt                         time.Time `json:"endsAt"`
	Status                         string    `json:"status"`
	CancellationDeadlineMinutes    int       `json:"cancellationDeadlineMinutes"`
	LateCancellationClientAllowed  bool      `json:"lateCancellationClientAllowed"`
	LateCancellationReasonRequired bool      `json:"lateCancellationReasonRequired"`
}
