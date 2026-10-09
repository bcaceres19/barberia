// Package httpapi es el adaptador HTTP de googlecalendar: decodifica, invoca
// googlecalendar.Service y traduce el resultado a HTTP
// (docs/03-desarrollo/estandar-backend-go.md §4). Ninguna respuesta contiene
// tokens, secretos, el código OAuth ni el `state`.
package httpapi

import "time"

// ConnectionResponse es GoogleCalendarConnectionResponse (OpenAPI).
type ConnectionResponse struct {
	Enabled         bool       `json:"enabled"`
	BarberLinked    bool       `json:"barberLinked"`
	Status          string     `json:"status"`
	AccountEmail    *string    `json:"accountEmail"`
	ReminderMinutes *int       `json:"reminderMinutes"`
	ConnectedAt     *time.Time `json:"connectedAt"`
	LastSyncedAt    *time.Time `json:"lastSyncedAt"`
}

// AuthorizationResponse es GoogleCalendarAuthorizationResponse (OpenAPI).
type AuthorizationResponse struct {
	AuthorizationURL string `json:"authorizationUrl"`
}

// CallbackRequest es CompleteGoogleCalendarConnectionRequest (OpenAPI).
// Cerrado: nunca acepta identificadores de barbero, usuario ni barbería.
type CallbackRequest struct {
	State string `json:"state"`
	Code  string `json:"code"`
	Error string `json:"error"`
}

// CallbackResponse es GoogleCalendarCallbackResponse (OpenAPI).
type CallbackResponse struct {
	Result string `json:"result"`
}
