package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"system-barbershop/internal/modules/catalog"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
)

// barberIDParam es el nombre del parámetro de ruta que cmd/api.buildRouter
// registra para /private/barbers/{barberId}/services[/{serviceId}].
// Deliberadamente igual al de staff/httpapi (mismo segmento de ruta), pero
// este paquete no lo importa: es una constante propia (mismo criterio de
// duplicación pequeña que catalog.barberIDPattern en assignment.go).
// serviceIDParam ya existe en handler.go (mismo paquete): se reutiliza tal
// cual, sin declararlo dos veces.
const barberIDParam = "barberId"

// assignmentBasePath es la base usada para construir el header Location de
// una asignación creada, relativa a /api/v1 (igual que el resto del
// contrato).
func assignmentBasePath(barberID string) string {
	return "/private/barbers/" + barberID + "/services/"
}

// ListAssignmentsHandler expone GET /private/barbers/{barberId}/services
// (CA-023-01).
type ListAssignmentsHandler struct {
	service *catalog.AssignmentService
}

// NewListAssignmentsHandler construye el handler de listado.
func NewListAssignmentsHandler(service *catalog.AssignmentService) *ListAssignmentsHandler {
	return &ListAssignmentsHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *ListAssignmentsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	barberID := httpserver.URLParam(r, barberIDParam)

	query := r.URL.Query()
	cursor := query.Get("cursor")

	limit := 0
	if raw := query.Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			httpserver.WriteProblem(w, httpserver.Translate(
				apperr.Invalid("el parámetro limit debe ser un entero positivo"), requestID))
			return
		}
		limit = parsed
	}

	result, err := h.service.List(r.Context(), principal.BarbershopID, barberID, cursor, limit)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newAssignmentListResponse(result))
}

// AssignServiceHandler expone
// PUT /private/barbers/{barberId}/services/{serviceId} (CA-023-02,
// CA-023-03). Semántica HTTP naturalmente repetible: no usa el protocolo de
// Idempotency-Key de RN-IDE-01 (esa cabecera protege un POST no idempotente
// por sí mismo; PUT sobre un recurso identificado por su propia clave
// -barberId+serviceId- ya lo es). Sin cuerpo: el contrato no declara
// ningún campo escribible más allá de los dos identificadores de la ruta.
type AssignServiceHandler struct {
	service *catalog.AssignmentService
}

// NewAssignServiceHandler construye el handler de asignación.
func NewAssignServiceHandler(service *catalog.AssignmentService) *AssignServiceHandler {
	return &AssignServiceHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *AssignServiceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	barberID := httpserver.URLParam(r, barberIDParam)
	serviceID := httpserver.URLParam(r, serviceIDParam)

	result, err := h.service.Assign(r.Context(), principal.BarbershopID, barberID, serviceID)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	status := http.StatusOK
	if result.Outcome == catalog.AssignOutcomeCreated {
		status = http.StatusCreated
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", assignmentBasePath(barberID)+serviceID)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(newAssignmentResponse(result.Assignment))
}

// UnassignServiceHandler expone
// DELETE /private/barbers/{barberId}/services/{serviceId} (CA-023-05,
// CA-023-06, DEC-068).
type UnassignServiceHandler struct {
	service *catalog.AssignmentService
}

// NewUnassignServiceHandler construye el handler de desasignación.
func NewUnassignServiceHandler(service *catalog.AssignmentService) *UnassignServiceHandler {
	return &UnassignServiceHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *UnassignServiceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	barberID := httpserver.URLParam(r, barberIDParam)
	serviceID := httpserver.URLParam(r, serviceIDParam)

	if _, err := h.service.Unassign(r.Context(), principal.BarbershopID, barberID, serviceID); err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func newAssignmentResponse(a catalog.Assignment) AssignmentResponse {
	return AssignmentResponse{
		BarberID:  a.BarberID,
		ServiceID: a.ServiceID,
		CreatedAt: a.CreatedAt,
	}
}

func newAssignmentListResponse(result catalog.AssignmentListResult) AssignmentListResponse {
	items := make([]AssignmentResponse, 0, len(result.Items))
	for _, a := range result.Items {
		items = append(items, newAssignmentResponse(a))
	}
	var next *string
	if result.NextCursor != "" {
		v := result.NextCursor
		next = &v
	}
	return AssignmentListResponse{Items: items, NextCursor: next}
}
