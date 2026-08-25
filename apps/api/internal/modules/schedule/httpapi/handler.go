// Package httpapi es el adaptador HTTP del módulo schedule: decodifica,
// invoca schedule.Service y traduce el resultado a HTTP
// (docs/03-desarrollo/estandar-backend-go.md §4), mismo patrón que
// staff/httpapi y catalog/httpapi.
package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/modules/schedule"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
	"system-barbershop/internal/platform/idempotency"
)

// barberIDParam y workingHourIDParam son los nombres de los parámetros de
// ruta que cmd/api.buildRouter registra para
// /private/barbers/{barberId}/working-hours[/{workingHourId}].
const (
	barberIDParam      = "barberId"
	workingHourIDParam = "workingHourId"
)

// workingHourBasePath es la base usada para construir el header Location
// de un alta exitosa, relativa a /api/v1.
func workingHourBasePath(barberID string) string {
	return "/private/barbers/" + barberID + "/working-hours/"
}

func principalOrInternalError(w http.ResponseWriter, r *http.Request, requestID string) (auth.Principal, bool) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		// Mismo defecto de wiring que staff/httpapi contempla: solo puede
		// ocurrir si esta ruta se montó fuera del subrouter protegido por
		// SessionMiddleware.
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Internal(errors.New("schedule: falta el principal de sesión en el contexto")), requestID))
		return auth.Principal{}, false
	}
	return principal, true
}

func writeUnknownFieldOrInvalidJSONProblem(w http.ResponseWriter, requestID string) {
	httpserver.WriteProblem(w, httpserver.Translate(
		apperr.Invalid("cuerpo JSON inválido o con un campo desconocido"), requestID))
}

// ListWorkingHoursHandler expone
// GET /private/barbers/{barberId}/working-hours (CA-040-01).
type ListWorkingHoursHandler struct {
	service *schedule.Service
}

// NewListWorkingHoursHandler construye el handler de listado.
func NewListWorkingHoursHandler(service *schedule.Service) *ListWorkingHoursHandler {
	return &ListWorkingHoursHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *ListWorkingHoursHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	_ = json.NewEncoder(w).Encode(newWorkingHourListResponse(result))
}

// GetWorkingHourHandler expone
// GET /private/barbers/{barberId}/working-hours/{workingHourId} (HU-040):
// la lectura individual permite demostrar de forma directa el 404 de un
// tramo de otro barbero o de otra barbería (CA-040-05).
type GetWorkingHourHandler struct {
	service *schedule.Service
}

// NewGetWorkingHourHandler construye el handler de lectura individual.
func NewGetWorkingHourHandler(service *schedule.Service) *GetWorkingHourHandler {
	return &GetWorkingHourHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *GetWorkingHourHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	barberID := httpserver.URLParam(r, barberIDParam)
	workingHourID := httpserver.URLParam(r, workingHourIDParam)

	wh, err := h.service.Get(r.Context(), principal.BarbershopID, barberID, workingHourID)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newWorkingHourResponse(wh))
}

// CreateWorkingHourHandler expone
// POST /private/barbers/{barberId}/working-hours (CA-040-02), protegido
// por el protocolo de idempotencia reutilizable de HU-004 (RN-IDE-01,
// DEC-043).
type CreateWorkingHourHandler struct {
	service *schedule.Service
}

// NewCreateWorkingHourHandler construye el handler de alta.
func NewCreateWorkingHourHandler(service *schedule.Service) *CreateWorkingHourHandler {
	return &CreateWorkingHourHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *CreateWorkingHourHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	barberID := httpserver.URLParam(r, barberIDParam)

	// 1. Cabecera: obligatoria, validada antes de tocar PostgreSQL o leer
	//    el cuerpo (apps/api/README.md, "Patrón obligatorio: idempotencia
	//    reutilizable").
	key, err := httpserver.IdempotencyKeyFromRequest(r)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	// 2. Cuerpo CRUDO, leído una sola vez: la huella de idempotencia debe
	//    calcularse sobre los bytes exactos recibidos.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
			return
		}
		httpserver.WriteProblem(w, httpserver.Translate(apperr.Invalid("cuerpo de la solicitud ilegible"), requestID))
		return
	}

	var req CreateWorkingHourRequest
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeUnknownFieldOrInvalidJSONProblem(w, requestID)
		return
	}
	if dec.More() {
		writeUnknownFieldOrInvalidJSONProblem(w, requestID)
		return
	}

	fingerprint := httpserver.IdempotencyFingerprint(r, body)

	result, err := h.service.Create(r.Context(), principal.BarbershopID, barberID, req.ISOWeekday, req.StartsTime, req.DurationMinutes, key, fingerprint)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	switch result.Decision.Outcome {
	case idempotency.OutcomeProceed, idempotency.OutcomeReplay:
		var parsed struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal([]byte(result.Response.Body), &parsed)
		w.Header().Set("Location", workingHourBasePath(barberID)+parsed.ID)
		httpserver.WriteStoredResponse(w, result.Response)
	default:
		// Conflicto de contenido/operación (RN-IDE-01) o ejecución en curso
		// (DEC-043): Translate ya sabe convertirlo en 409.
		httpserver.WriteProblem(w, httpserver.Translate(result.Decision.AsError(), requestID))
	}
}

// UpdateWorkingHourHandler expone
// PATCH /private/barbers/{barberId}/working-hours/{workingHourId}
// (CA-040-04).
type UpdateWorkingHourHandler struct {
	service *schedule.Service
}

// NewUpdateWorkingHourHandler construye el handler de edición.
func NewUpdateWorkingHourHandler(service *schedule.Service) *UpdateWorkingHourHandler {
	return &UpdateWorkingHourHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *UpdateWorkingHourHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	barberID := httpserver.URLParam(r, barberIDParam)
	workingHourID := httpserver.URLParam(r, workingHourIDParam)

	var req UpdateWorkingHourRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
			return
		}
		writeUnknownFieldOrInvalidJSONProblem(w, requestID)
		return
	}
	if dec.More() {
		writeUnknownFieldOrInvalidJSONProblem(w, requestID)
		return
	}

	wh, err := h.service.Update(r.Context(), principal.BarbershopID, barberID, workingHourID, req.ISOWeekday, req.StartsTime, req.DurationMinutes)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newWorkingHourResponse(wh))
}

// DeleteWorkingHourHandler expone
// DELETE /private/barbers/{barberId}/working-hours/{workingHourId}
// (CA-040-05).
type DeleteWorkingHourHandler struct {
	service *schedule.Service
}

// NewDeleteWorkingHourHandler construye el handler de retiro.
func NewDeleteWorkingHourHandler(service *schedule.Service) *DeleteWorkingHourHandler {
	return &DeleteWorkingHourHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *DeleteWorkingHourHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	barberID := httpserver.URLParam(r, barberIDParam)
	workingHourID := httpserver.URLParam(r, workingHourIDParam)

	if err := h.service.Delete(r.Context(), principal.BarbershopID, barberID, workingHourID); err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func newWorkingHourResponse(wh schedule.WorkingHour) WorkingHourResponse {
	return WorkingHourResponse{
		ID:              wh.ID,
		ISOWeekday:      wh.ISOWeekday,
		StartsTime:      wh.StartsTime,
		DurationMinutes: wh.DurationMinutes,
		CreatedAt:       wh.CreatedAt,
		UpdatedAt:       wh.UpdatedAt,
	}
}

func newWorkingHourListResponse(result schedule.ListResult) WorkingHourListResponse {
	items := make([]WorkingHourResponse, 0, len(result.Items))
	for _, wh := range result.Items {
		items = append(items, newWorkingHourResponse(wh))
	}
	var next *string
	if result.NextCursor != "" {
		v := result.NextCursor
		next = &v
	}
	return WorkingHourListResponse{Items: items, NextCursor: next}
}
