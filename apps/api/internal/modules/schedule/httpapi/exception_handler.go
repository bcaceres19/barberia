package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"system-barbershop/internal/modules/schedule"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
	"system-barbershop/internal/platform/idempotency"
)

// exceptionIDParam es el nombre del parámetro de ruta que cmd/api.buildRouter
// registra para /private/barbers/{barberId}/schedule-exceptions/{exceptionId}.
const exceptionIDParam = "exceptionId"

// exceptionBasePath es la base usada para construir el header Location de
// un alta exitosa, relativa a /api/v1.
func exceptionBasePath(barberID string) string {
	return "/private/barbers/" + barberID + "/schedule-exceptions/"
}

// GetHolidayCalendarHandler expone
// GET /private/barbers/{barberId}/holiday-calendar (CA-041-01/02).
type GetHolidayCalendarHandler struct {
	service *schedule.Service
}

// NewGetHolidayCalendarHandler construye el handler de lectura.
func NewGetHolidayCalendarHandler(service *schedule.Service) *GetHolidayCalendarHandler {
	return &GetHolidayCalendarHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *GetHolidayCalendarHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	barberID := httpserver.URLParam(r, barberIDParam)

	enabled, err := h.service.GetHolidayCalendar(r.Context(), principal.BarbershopID, barberID)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newHolidayCalendarResponse(enabled))
}

// UpdateHolidayCalendarHandler expone
// PATCH /private/barbers/{barberId}/holiday-calendar (CA-041-01/02).
type UpdateHolidayCalendarHandler struct {
	service *schedule.Service
}

// NewUpdateHolidayCalendarHandler construye el handler de actualización.
func NewUpdateHolidayCalendarHandler(service *schedule.Service) *UpdateHolidayCalendarHandler {
	return &UpdateHolidayCalendarHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *UpdateHolidayCalendarHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	barberID := httpserver.URLParam(r, barberIDParam)

	var req UpdateHolidayCalendarRequest
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

	enabled, err := h.service.SetHolidayCalendar(r.Context(), principal.BarbershopID, barberID, req.Enabled)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newHolidayCalendarResponse(enabled))
}

// ListScheduleExceptionsHandler expone
// GET /private/barbers/{barberId}/schedule-exceptions (CA-041-04/05).
type ListScheduleExceptionsHandler struct {
	service *schedule.Service
}

// NewListScheduleExceptionsHandler construye el handler de listado.
func NewListScheduleExceptionsHandler(service *schedule.Service) *ListScheduleExceptionsHandler {
	return &ListScheduleExceptionsHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *ListScheduleExceptionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.service.ListExceptions(r.Context(), principal.BarbershopID, barberID, cursor, limit)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newScheduleExceptionListResponse(result))
}

// GetScheduleExceptionHandler expone
// GET /private/barbers/{barberId}/schedule-exceptions/{exceptionId}
// (CA-041-06).
type GetScheduleExceptionHandler struct {
	service *schedule.Service
}

// NewGetScheduleExceptionHandler construye el handler de lectura individual.
func NewGetScheduleExceptionHandler(service *schedule.Service) *GetScheduleExceptionHandler {
	return &GetScheduleExceptionHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *GetScheduleExceptionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	barberID := httpserver.URLParam(r, barberIDParam)
	exceptionID := httpserver.URLParam(r, exceptionIDParam)

	e, err := h.service.GetException(r.Context(), principal.BarbershopID, barberID, exceptionID)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newScheduleExceptionResponse(e))
}

// CreateScheduleExceptionHandler expone
// POST /private/barbers/{barberId}/schedule-exceptions (CA-041-04/05),
// protegido por el protocolo de idempotencia reutilizable de HU-004
// (RN-IDE-01, DEC-043).
type CreateScheduleExceptionHandler struct {
	service *schedule.Service
}

// NewCreateScheduleExceptionHandler construye el handler de alta.
func NewCreateScheduleExceptionHandler(service *schedule.Service) *CreateScheduleExceptionHandler {
	return &CreateScheduleExceptionHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *CreateScheduleExceptionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	barberID := httpserver.URLParam(r, barberIDParam)

	key, err := httpserver.IdempotencyKeyFromRequest(r)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

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

	var req CreateScheduleExceptionRequest
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

	result, err := h.service.CreateException(r.Context(), principal.BarbershopID, barberID,
		req.EffectiveDate, req.IsClosed, req.Reason, toExceptionSegmentInputs(req.Segments), key, fingerprint)
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
		w.Header().Set("Location", exceptionBasePath(barberID)+parsed.ID)
		httpserver.WriteStoredResponse(w, result.Response)
	default:
		httpserver.WriteProblem(w, httpserver.Translate(result.Decision.AsError(), requestID))
	}
}

// UpdateScheduleExceptionHandler expone
// PATCH /private/barbers/{barberId}/schedule-exceptions/{exceptionId}
// (CA-041-04/05).
type UpdateScheduleExceptionHandler struct {
	service *schedule.Service
}

// NewUpdateScheduleExceptionHandler construye el handler de edición.
func NewUpdateScheduleExceptionHandler(service *schedule.Service) *UpdateScheduleExceptionHandler {
	return &UpdateScheduleExceptionHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *UpdateScheduleExceptionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	barberID := httpserver.URLParam(r, barberIDParam)
	exceptionID := httpserver.URLParam(r, exceptionIDParam)

	var req UpdateScheduleExceptionRequest
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

	e, err := h.service.UpdateException(r.Context(), principal.BarbershopID, barberID, exceptionID,
		req.EffectiveDate, req.IsClosed, req.Reason, toExceptionSegmentInputs(req.Segments))
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newScheduleExceptionResponse(e))
}

// DeleteScheduleExceptionHandler expone
// DELETE /private/barbers/{barberId}/schedule-exceptions/{exceptionId}
// (CA-041-06).
type DeleteScheduleExceptionHandler struct {
	service *schedule.Service
}

// NewDeleteScheduleExceptionHandler construye el handler de retiro.
func NewDeleteScheduleExceptionHandler(service *schedule.Service) *DeleteScheduleExceptionHandler {
	return &DeleteScheduleExceptionHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *DeleteScheduleExceptionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	barberID := httpserver.URLParam(r, barberIDParam)
	exceptionID := httpserver.URLParam(r, exceptionIDParam)

	if err := h.service.DeleteException(r.Context(), principal.BarbershopID, barberID, exceptionID); err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListColombianHolidaysHandler expone
// GET /private/schedule/colombian-holidays (RN-BLQ-02): dato de referencia
// calendárica, calculado de forma determinista, igual para toda barbería y
// todo barbero.
type ListColombianHolidaysHandler struct{}

// NewListColombianHolidaysHandler construye el handler de festivos.
func NewListColombianHolidaysHandler() *ListColombianHolidaysHandler {
	return &ListColombianHolidaysHandler{}
}

// ServeHTTP implementa http.Handler.
func (h *ListColombianHolidaysHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	if _, ok := principalOrInternalError(w, r, requestID); !ok {
		return
	}

	raw := r.URL.Query().Get("year")
	year, err := strconv.Atoi(raw)
	if err != nil || year < 1900 || year > 2200 {
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Invalid("el parámetro year es obligatorio y debe ser un año entre 1900 y 2200"), requestID))
		return
	}

	holidays := schedule.ColombianHolidaysForYear(year)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newColombianHolidayListResponse(holidays))
}
