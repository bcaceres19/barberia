package httpapi

import (
	"time"

	"system-barbershop/internal/platform/httpserver"
)

// PublicBarbershopProfileResponse es la representación pública mínima de
// una barbería habilitada (HU-090, CA-090-01, CA-090-04), forma exacta del
// componente OpenAPI PublicBarbershopProfile.yaml. contactEmail/
// contactPhone viajan siempre presentes, con valor null explícito cuando la
// barbería no tiene ese contacto configurado, igual que
// shops/httpapi.BarbershopSettingsResponse.
type PublicBarbershopProfileResponse struct {
	Name         string  `json:"name"`
	Timezone     string  `json:"timezone"`
	ContactEmail *string `json:"contactEmail"`
	ContactPhone *string `json:"contactPhone"`
}

// PublicServiceResponse es la proyección pública mínima de un servicio
// (HU-091, CA-091-01), forma exacta del componente OpenAPI
// PublicServiceResponse.yaml: nunca isActive, deactivatedAt, createdAt ni
// updatedAt (esos campos son alcance exclusivo del catálogo privado,
// HU-022/HU-024).
type PublicServiceResponse struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Description     *string `json:"description"`
	DurationMinutes int     `json:"durationMinutes"`
	Price           string  `json:"price"`
	Currency        string  `json:"currency"`
}

// PublicServiceListResponse es una página del catálogo público de
// servicios (HU-091), misma forma que ServiceListResponse.yaml del
// catálogo privado.
type PublicServiceListResponse struct {
	Items      []PublicServiceResponse `json:"items"`
	NextCursor *string                 `json:"nextCursor"`
}

// PublicBarberResponse es la proyección pública mínima de un barbero con
// asignación vigente (HU-092, CA-092-02), forma exacta del componente
// OpenAPI PublicBarberResponse.yaml: nunca foto, biografía ni ningún otro
// campo fuera de alcance de HU-092.
type PublicBarberResponse struct {
	ID       string `json:"id"`
	FullName string `json:"fullName"`
}

// PublicBarberListResponse es la lista completa de barberos elegibles para
// el servicio activo resuelto (HU-092). Sin paginar: ver
// publicbooking.PublicBarberListResult.
type PublicBarberListResponse struct {
	Items []PublicBarberResponse `json:"items"`
}

// AvailabilitySlotResponse es un inicio público válido (HU-094, CA-094-01),
// forma exacta del componente OpenAPI AvailabilitySlotResponse.yaml.
type AvailabilitySlotResponse struct {
	StartsAt time.Time `json:"startsAt"`
}

// ConfirmPublicAppointmentRequest es el cuerpo de POST /public/barbershops/
// {slug}/services/{serviceId}/barbers/{barberId}/appointments (HU-097).
// Cerrado: el handler rechaza cualquier campo desconocido. Nunca declara
// barbershopId, customerId, status, origin, precio, duración ni ningún
// otro dato que el servidor derive o revalide (CA-097-02); startsAt es el
// instante absoluto ISO 8601 que el cliente eligió de entre los `slots` de
// HU-094 (nunca una hora civil, a diferencia de CreateManualAppointmentRequest).
type ConfirmPublicAppointmentRequest struct {
	StartsAt       string  `json:"startsAt"`
	FullName       string  `json:"fullName"`
	Phone          string  `json:"phone"`
	Email          string  `json:"email"`
	Note           *string `json:"note"`
	ForSomeoneElse bool    `json:"forSomeoneElse"`
	AttendeeName   *string `json:"attendeeName"`
}

// ConfirmedPublicAppointmentResponse es la respuesta 201 de HU-097, forma
// exacta que booking/postgres.publicAppointmentResponseWire serializa
// dentro de la transacción atómica (idempotency_record.response_body):
// ambas deben coincidir campo a campo (mismo criterio que
// AppointmentResponse.yaml frente a manualAppointmentResponseWire).
// accessToken viaja EN CLARO, la única vez que existe fuera de la
// transacción (DEC-089); nunca vuelve a aparecer en ninguna otra respuesta.
type ConfirmedPublicAppointmentResponse struct {
	AttendeeName    string  `json:"attendeeName"`
	BarbershopName  string  `json:"barbershopName"`
	ServiceName     string  `json:"serviceName"`
	DurationMinutes int     `json:"durationMinutes"`
	PriceAmount     string  `json:"priceAmount"`
	Currency        string  `json:"currency"`
	StartsAt        string  `json:"startsAt"`
	EndsAt          string  `json:"endsAt"`
	Timezone        string  `json:"timezone"`
	AccessToken     string  `json:"accessToken"`
	CustomerNote    *string `json:"customerNote"`
}

// AlternativeSlotResponse es una franja alternativa (RN-CON-05, DEC-090),
// misma forma que AvailabilitySlotResponse.
type AlternativeSlotResponse struct {
	StartsAt time.Time `json:"startsAt"`
}

// ScheduleConflictProblemResponse extiende httpserver.Problem con
// `alternatives` (RFC 9457, extensión abierta: Problem.yaml declara
// additionalProperties abierto exactamente para este caso). Nunca se
// construye a mano fuera de writeScheduleConflictProblem: Translate sigue
// siendo el único punto que decide type/title/code/status/detail seguros
// (CA-003-02); esta forma solo agrega el dato público adicional que
// Translate no puede conocer.
type ScheduleConflictProblemResponse struct {
	httpserver.Problem
	Alternatives []AlternativeSlotResponse `json:"alternatives"`
}

// AvailabilityResponse es la respuesta completa de HU-094: los inicios
// válidos ya ordenados, más el contexto mínimo para interpretarlos sin
// ambigüedad (RN-DIS-07). Un arreglo `slots` vacío cubre por igual "sin
// disponibilidad hoy" y serviceId/barberId ajeno, inexistente, inactivo o
// sin asignación vigente: la causa nunca se distingue (mismo criterio que
// CA-092-03).
type AvailabilityResponse struct {
	Slots           []AvailabilitySlotResponse `json:"slots"`
	DurationMinutes int                        `json:"durationMinutes"`
	Timezone        string                     `json:"timezone"`
	SlotGridMinutes int                        `json:"slotGridMinutes"`
}
