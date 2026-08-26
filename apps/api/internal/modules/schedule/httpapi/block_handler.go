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

// blockIDParam, seriesIDParam, blockDateParam y excludedDateParam son los
// nombres de los parámetros de ruta que cmd/api.main registra para los
// recursos de HU-042 (bloqueo puntual, serie recurrente, fecha explícita y
// excepción).
const (
	blockIDParam      = "blockId"
	seriesIDParam     = "seriesId"
	blockDateParam    = "blockDate"
	excludedDateParam = "excludedDate"
)

func timeBlockBasePath(barberID string) string {
	return "/private/barbers/" + barberID + "/time-blocks/"
}

func timeBlockSeriesBasePath(barberID string) string {
	return "/private/barbers/" + barberID + "/time-block-series/"
}

func newTimeBlockResponse(b schedule.TimeBlock) TimeBlockResponse {
	return TimeBlockResponse{
		ID: b.ID, BlockType: b.BlockType, Source: b.Source, StartsAt: b.StartsAt, EndsAt: b.EndsAt,
		Reason: b.Reason, DeletedAt: b.DeletedAt, DeletedBy: b.DeletedBy, CreatedAt: b.CreatedAt, UpdatedAt: b.UpdatedAt,
	}
}

func newTimeBlockListResponse(result schedule.BlockListResult) TimeBlockListResponse {
	items := make([]TimeBlockResponse, 0, len(result.Items))
	for _, b := range result.Items {
		items = append(items, newTimeBlockResponse(b))
	}
	var next *string
	if result.NextCursor != "" {
		v := result.NextCursor
		next = &v
	}
	return TimeBlockListResponse{Items: items, NextCursor: next}
}

func newTimeBlockSeriesResponse(sr schedule.TimeBlockSeries) TimeBlockSeriesResponse {
	dates := make([]SeriesDateResponse, 0, len(sr.Dates))
	for _, d := range sr.Dates {
		dates = append(dates, SeriesDateResponse{BlockDate: d.BlockDate})
	}
	exceptions := make([]SeriesExceptionResponse, 0, len(sr.Exceptions))
	for _, e := range sr.Exceptions {
		exceptions = append(exceptions, SeriesExceptionResponse{ExcludedDate: e.ExcludedDate, Reason: e.Reason, CreatedAt: e.CreatedAt})
	}
	return TimeBlockSeriesResponse{
		ID: sr.ID, BlockType: sr.BlockType, RecurrenceKind: sr.RecurrenceKind, ISOWeekday: sr.ISOWeekday,
		StartsTime: sr.StartsTime, DurationMinutes: sr.DurationMinutes, EffectiveFrom: sr.EffectiveFrom,
		EffectiveUntil: sr.EffectiveUntil, Reason: sr.Reason, DeletedAt: sr.DeletedAt,
		CreatedAt: sr.CreatedAt, UpdatedAt: sr.UpdatedAt, Dates: dates, Exceptions: exceptions,
	}
}

func newTimeBlockSeriesListResponse(result schedule.SeriesListResult) TimeBlockSeriesListResponse {
	items := make([]TimeBlockSeriesResponse, 0, len(result.Items))
	for _, sr := range result.Items {
		items = append(items, newTimeBlockSeriesResponse(sr))
	}
	var next *string
	if result.NextCursor != "" {
		v := result.NextCursor
		next = &v
	}
	return TimeBlockSeriesListResponse{Items: items, NextCursor: next}
}

func parseIncludeDeleted(query map[string][]string) bool {
	values, ok := query["includeDeleted"]
	return ok && len(values) > 0 && values[0] == "true"
}

func decodeJSONBody[T any](w http.ResponseWriter, r *http.Request, requestID string) (T, bool) {
	var req T
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
			return req, false
		}
		writeUnknownFieldOrInvalidJSONProblem(w, requestID)
		return req, false
	}
	if dec.More() {
		writeUnknownFieldOrInvalidJSONProblem(w, requestID)
		return req, false
	}
	return req, true
}

// --- Bloqueos puntuales (HU-042) ------------------------------------------

// ListTimeBlocksHandler expone GET /private/barbers/{barberId}/time-blocks.
type ListTimeBlocksHandler struct{ service *schedule.Service }

func NewListTimeBlocksHandler(service *schedule.Service) *ListTimeBlocksHandler {
	return &ListTimeBlocksHandler{service: service}
}

func (h *ListTimeBlocksHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}
	barberID := httpserver.URLParam(r, barberIDParam)
	query := r.URL.Query()

	limit := 0
	if raw := query.Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			httpserver.WriteProblem(w, httpserver.Translate(apperr.Invalid("el parámetro limit debe ser un entero positivo"), requestID))
			return
		}
		limit = parsed
	}

	result, err := h.service.ListBlocks(r.Context(), principal.BarbershopID, barberID, query.Get("cursor"), limit, parseIncludeDeleted(query))
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newTimeBlockListResponse(result))
}

// GetTimeBlockHandler expone
// GET /private/barbers/{barberId}/time-blocks/{blockId}.
type GetTimeBlockHandler struct{ service *schedule.Service }

func NewGetTimeBlockHandler(service *schedule.Service) *GetTimeBlockHandler {
	return &GetTimeBlockHandler{service: service}
}

func (h *GetTimeBlockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}
	barberID := httpserver.URLParam(r, barberIDParam)
	blockID := httpserver.URLParam(r, blockIDParam)

	block, err := h.service.GetBlock(r.Context(), principal.BarbershopID, barberID, blockID)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newTimeBlockResponse(block))
}

// CreateTimeBlockHandler expone
// POST /private/barbers/{barberId}/time-blocks (RN-BLQ-01/03), protegido
// por el protocolo de idempotencia reutilizable (RN-IDE-01, DEC-043).
type CreateTimeBlockHandler struct{ service *schedule.Service }

func NewCreateTimeBlockHandler(service *schedule.Service) *CreateTimeBlockHandler {
	return &CreateTimeBlockHandler{service: service}
}

func (h *CreateTimeBlockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	var req CreateTimeBlockRequest
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil || dec.More() {
		writeUnknownFieldOrInvalidJSONProblem(w, requestID)
		return
	}

	fingerprint := httpserver.IdempotencyFingerprint(r, body)

	result, err := h.service.CreateBlock(r.Context(), principal.BarbershopID, barberID, req.BlockType, req.StartsAt, req.EndsAt, req.Reason, key, fingerprint)
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
		w.Header().Set("Location", timeBlockBasePath(barberID)+parsed.ID)
		httpserver.WriteStoredResponse(w, result.Response)
	default:
		httpserver.WriteProblem(w, httpserver.Translate(result.Decision.AsError(), requestID))
	}
}

// DeleteTimeBlockHandler expone
// DELETE /private/barbers/{barberId}/time-blocks/{blockId} (RN-BLQ-04):
// retiro lógico, nunca físico.
type DeleteTimeBlockHandler struct{ service *schedule.Service }

func NewDeleteTimeBlockHandler(service *schedule.Service) *DeleteTimeBlockHandler {
	return &DeleteTimeBlockHandler{service: service}
}

func (h *DeleteTimeBlockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}
	barberID := httpserver.URLParam(r, barberIDParam)
	blockID := httpserver.URLParam(r, blockIDParam)

	if err := h.service.DeleteBlock(r.Context(), principal.BarbershopID, barberID, blockID, principal.StaffUserID); err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// EffectiveBlocksHandler expone
// GET /private/barbers/{barberId}/time-blocks/effective?from&to: la
// operación interna de proyección que el prompt de HU-042 exige, para que
// B3/B4 (y la propia pantalla, "próximos bloqueos") puedan leer el efecto
// combinado sin recalcular la expansión de series por su cuenta.
type EffectiveBlocksHandler struct{ service *schedule.Service }

func NewEffectiveBlocksHandler(service *schedule.Service) *EffectiveBlocksHandler {
	return &EffectiveBlocksHandler{service: service}
}

func (h *EffectiveBlocksHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}
	barberID := httpserver.URLParam(r, barberIDParam)
	query := r.URL.Query()

	result, err := h.service.EffectiveBlocks(r.Context(), principal.BarbershopID, barberID, query.Get("from"), query.Get("to"))
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	manual := make([]TimeBlockResponse, 0, len(result.ManualBlocks))
	for _, b := range result.ManualBlocks {
		manual = append(manual, newTimeBlockResponse(b))
	}
	occurrences := make([]SeriesOccurrenceResponse, 0, len(result.SeriesOccurrences))
	for _, occ := range result.SeriesOccurrences {
		occurrences = append(occurrences, SeriesOccurrenceResponse{
			SeriesID: occ.SeriesID, BlockType: occ.BlockType, Date: occ.Date,
			StartsTime: occ.StartsTime, DurationMinutes: occ.DurationMinutes, Reason: occ.Reason,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(EffectiveBlocksResponse{ManualBlocks: manual, SeriesOccurrences: occurrences})
}

// --- Series recurrentes (HU-042) ------------------------------------------

// ListTimeBlockSeriesHandler expone
// GET /private/barbers/{barberId}/time-block-series.
type ListTimeBlockSeriesHandler struct{ service *schedule.Service }

func NewListTimeBlockSeriesHandler(service *schedule.Service) *ListTimeBlockSeriesHandler {
	return &ListTimeBlockSeriesHandler{service: service}
}

func (h *ListTimeBlockSeriesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}
	barberID := httpserver.URLParam(r, barberIDParam)
	query := r.URL.Query()

	limit := 0
	if raw := query.Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			httpserver.WriteProblem(w, httpserver.Translate(apperr.Invalid("el parámetro limit debe ser un entero positivo"), requestID))
			return
		}
		limit = parsed
	}

	result, err := h.service.ListSeries(r.Context(), principal.BarbershopID, barberID, query.Get("cursor"), limit, parseIncludeDeleted(query))
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newTimeBlockSeriesListResponse(result))
}

// GetTimeBlockSeriesHandler expone
// GET /private/barbers/{barberId}/time-block-series/{seriesId}.
type GetTimeBlockSeriesHandler struct{ service *schedule.Service }

func NewGetTimeBlockSeriesHandler(service *schedule.Service) *GetTimeBlockSeriesHandler {
	return &GetTimeBlockSeriesHandler{service: service}
}

func (h *GetTimeBlockSeriesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}
	barberID := httpserver.URLParam(r, barberIDParam)
	seriesID := httpserver.URLParam(r, seriesIDParam)

	series, err := h.service.GetSeries(r.Context(), principal.BarbershopID, barberID, seriesID)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newTimeBlockSeriesResponse(series))
}

// CreateTimeBlockSeriesHandler expone
// POST /private/barbers/{barberId}/time-block-series, protegido por el
// protocolo de idempotencia reutilizable.
type CreateTimeBlockSeriesHandler struct{ service *schedule.Service }

func NewCreateTimeBlockSeriesHandler(service *schedule.Service) *CreateTimeBlockSeriesHandler {
	return &CreateTimeBlockSeriesHandler{service: service}
}

func (h *CreateTimeBlockSeriesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	var req CreateTimeBlockSeriesRequest
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil || dec.More() {
		writeUnknownFieldOrInvalidJSONProblem(w, requestID)
		return
	}

	fingerprint := httpserver.IdempotencyFingerprint(r, body)

	result, err := h.service.CreateSeries(r.Context(), principal.BarbershopID, barberID,
		req.BlockType, req.RecurrenceKind, req.ISOWeekday, req.StartsTime, req.DurationMinutes,
		req.EffectiveFrom, req.EffectiveUntil, req.Reason, req.ExplicitDates, key, fingerprint)
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
		w.Header().Set("Location", timeBlockSeriesBasePath(barberID)+parsed.ID)
		httpserver.WriteStoredResponse(w, result.Response)
	default:
		httpserver.WriteProblem(w, httpserver.Translate(result.Decision.AsError(), requestID))
	}
}

// UpdateTimeBlockSeriesHandler expone
// PATCH /private/barbers/{barberId}/time-block-series/{seriesId}
// (RN-BLQ-01: scope whole o this_and_following).
type UpdateTimeBlockSeriesHandler struct{ service *schedule.Service }

func NewUpdateTimeBlockSeriesHandler(service *schedule.Service) *UpdateTimeBlockSeriesHandler {
	return &UpdateTimeBlockSeriesHandler{service: service}
}

func (h *UpdateTimeBlockSeriesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}
	barberID := httpserver.URLParam(r, barberIDParam)
	seriesID := httpserver.URLParam(r, seriesIDParam)

	req, ok := decodeJSONBody[UpdateTimeBlockSeriesRequest](w, r, requestID)
	if !ok {
		return
	}

	series, err := h.service.UpdateSeries(r.Context(), principal.BarbershopID, barberID, seriesID,
		req.Scope, req.BlockType, req.StartsTime, req.DurationMinutes, req.EffectiveFrom, req.EffectiveUntil, req.Reason, req.EffectiveDate)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newTimeBlockSeriesResponse(series))
}

// DeleteTimeBlockSeriesHandler expone
// DELETE /private/barbers/{barberId}/time-block-series/{seriesId}
// (RN-BLQ-04): retiro lógico de la serie completa.
type DeleteTimeBlockSeriesHandler struct{ service *schedule.Service }

func NewDeleteTimeBlockSeriesHandler(service *schedule.Service) *DeleteTimeBlockSeriesHandler {
	return &DeleteTimeBlockSeriesHandler{service: service}
}

func (h *DeleteTimeBlockSeriesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}
	barberID := httpserver.URLParam(r, barberIDParam)
	seriesID := httpserver.URLParam(r, seriesIDParam)

	if err := h.service.DeleteSeries(r.Context(), principal.BarbershopID, barberID, seriesID); err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AddSeriesDateHandler expone
// POST /private/barbers/{barberId}/time-block-series/{seriesId}/dates.
type AddSeriesDateHandler struct{ service *schedule.Service }

func NewAddSeriesDateHandler(service *schedule.Service) *AddSeriesDateHandler {
	return &AddSeriesDateHandler{service: service}
}

func (h *AddSeriesDateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}
	barberID := httpserver.URLParam(r, barberIDParam)
	seriesID := httpserver.URLParam(r, seriesIDParam)

	req, ok := decodeJSONBody[AddSeriesDateRequest](w, r, requestID)
	if !ok {
		return
	}

	if err := h.service.AddSeriesDate(r.Context(), principal.BarbershopID, barberID, seriesID, req.BlockDate); err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RemoveSeriesDateHandler expone
// DELETE .../time-block-series/{seriesId}/dates/{blockDate}.
type RemoveSeriesDateHandler struct{ service *schedule.Service }

func NewRemoveSeriesDateHandler(service *schedule.Service) *RemoveSeriesDateHandler {
	return &RemoveSeriesDateHandler{service: service}
}

func (h *RemoveSeriesDateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}
	barberID := httpserver.URLParam(r, barberIDParam)
	seriesID := httpserver.URLParam(r, seriesIDParam)
	blockDate := httpserver.URLParam(r, blockDateParam)

	if err := h.service.RemoveSeriesDate(r.Context(), principal.BarbershopID, barberID, seriesID, blockDate); err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AddSeriesExceptionHandler expone
// POST .../time-block-series/{seriesId}/exceptions ("esta instancia no").
type AddSeriesExceptionHandler struct{ service *schedule.Service }

func NewAddSeriesExceptionHandler(service *schedule.Service) *AddSeriesExceptionHandler {
	return &AddSeriesExceptionHandler{service: service}
}

func (h *AddSeriesExceptionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}
	barberID := httpserver.URLParam(r, barberIDParam)
	seriesID := httpserver.URLParam(r, seriesIDParam)

	req, ok := decodeJSONBody[AddSeriesExceptionRequest](w, r, requestID)
	if !ok {
		return
	}

	if err := h.service.AddSeriesException(r.Context(), principal.BarbershopID, barberID, seriesID, req.ExcludedDate, req.Reason); err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RemoveSeriesExceptionHandler expone
// DELETE .../time-block-series/{seriesId}/exceptions/{excludedDate}
// ("restaura" la instancia).
type RemoveSeriesExceptionHandler struct{ service *schedule.Service }

func NewRemoveSeriesExceptionHandler(service *schedule.Service) *RemoveSeriesExceptionHandler {
	return &RemoveSeriesExceptionHandler{service: service}
}

func (h *RemoveSeriesExceptionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}
	barberID := httpserver.URLParam(r, barberIDParam)
	seriesID := httpserver.URLParam(r, seriesIDParam)
	excludedDate := httpserver.URLParam(r, excludedDateParam)

	if err := h.service.RemoveSeriesException(r.Context(), principal.BarbershopID, barberID, seriesID, excludedDate); err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
