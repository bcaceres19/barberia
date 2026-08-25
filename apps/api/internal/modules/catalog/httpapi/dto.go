package httpapi

import "time"

// ServiceResponse es la representación canónica de un servicio del catálogo
// (CA-022-01, CA-022-07, CA-024-02, CA-024-05): id, name, description,
// durationMinutes, price, currency, isActive, deactivatedAt, createdAt,
// updatedAt. isActive/deactivatedAt son de solo lectura (HU-024): ningún
// endpoint de HU-022 (create/update) los acepta como entrada; solo cambian
// mediante DeactivateServiceHandler/ReactivateServiceHandler. Nunca
// asignaciones ni citas (alcance excluido).
//
// ADVERTENCIA: postgres.serviceResponseWire
// (catalog/postgres/repository.go) declara EXACTAMENTE la misma forma
// (mismos nombres de campo JSON, mismo orden, mismos tipos) para construir
// el cuerpo que la idempotencia persiste en el alta/desactivación/
// reactivación. Un cambio aquí debe reflejarse ahí en el mismo commit; ver
// TestCreate_StoredResponseBody_MatchesHTTPAPIWireShape en
// postgres/repository_test.go.
type ServiceResponse struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Description     *string    `json:"description"`
	DurationMinutes int        `json:"durationMinutes"`
	Price           string     `json:"price"`
	Currency        string     `json:"currency"`
	IsActive        bool       `json:"isActive"`
	DeactivatedAt   *time.Time `json:"deactivatedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// ServiceDeactivationImpactResponse es el cuerpo de
// GET /private/services/{serviceId}/deactivation-impact (CA-024-01). Siempre
// AffectedAppointments: 0 en B1 (DEC-069).
type ServiceDeactivationImpactResponse struct {
	AffectedAppointments int `json:"affectedAppointments"`
}

// ServiceDeactivationResponse es el cuerpo de
// POST /private/services/{serviceId}/deactivate (CA-024-02, CA-024-04):
// el servicio ya desactivado junto con el impacto vuelto a consultar en la
// misma confirmación.
type ServiceDeactivationResponse struct {
	Service              ServiceResponse `json:"service"`
	AffectedAppointments int             `json:"affectedAppointments"`
}

// ServiceListResponse es la página paginada por cursor de
// GET /private/services (CA-022-01). NextCursor es un puntero para que "sin
// página siguiente" serialice como `null` explícito, nunca como una cadena
// vacía.
type ServiceListResponse struct {
	Items      []ServiceResponse `json:"items"`
	NextCursor *string           `json:"nextCursor"`
}

// CreateServiceRequest es el cuerpo de POST /private/services (CA-022-02,
// CA-022-03, CA-022-04). Cerrado: el handler rechaza cualquier campo
// desconocido. Nunca declara barbershopId (el tenant se deriva de
// SessionCookie), currency (fija en COP, DEC-067) ni isActive.
type CreateServiceRequest struct {
	Name            string  `json:"name"`
	Description     *string `json:"description,omitempty"`
	DurationMinutes int     `json:"durationMinutes"`
	Price           string  `json:"price"`
}

// UpdateServiceRequest es el cuerpo de PATCH /private/services/{serviceId}
// (CA-022-04, CA-022-05). Cerrado, edición parcial: cada puntero nil
// significa "no editar este campo"; un objeto sin ningún campo se rechaza
// (errUpdateEmptyBody). Nunca acepta isActive, asignaciones, citas ni
// alcance de propagación (RN-SER-04) ni currency (DEC-067).
type UpdateServiceRequest struct {
	Name            *string `json:"name,omitempty"`
	Description     *string `json:"description,omitempty"`
	DurationMinutes *int    `json:"durationMinutes,omitempty"`
	Price           *string `json:"price,omitempty"`
}
