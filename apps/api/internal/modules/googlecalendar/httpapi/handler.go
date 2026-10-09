package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/modules/googlecalendar"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
)

func principalOrInternalError(w http.ResponseWriter, r *http.Request, requestID string) (auth.Principal, bool) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		// Defecto de wiring: solo ocurre si la ruta se montó fuera del
		// subrouter protegido por SessionMiddleware.
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Internal(errors.New("googlecalendar: falta el principal de sesión en el contexto")), requestID))
		return auth.Principal{}, false
	}
	return principal, true
}

func writeInvalidJSON(w http.ResponseWriter, requestID string) {
	httpserver.WriteProblem(w, httpserver.Translate(
		apperr.Invalid("cuerpo JSON inválido o con un campo desconocido"), requestID))
}

// decodeStrict decodifica un único objeto JSON rechazando campos desconocidos y
// datos sobrantes. Devuelve false tras escribir el problema.
func decodeStrict(w http.ResponseWriter, r *http.Request, requestID string, into any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(into); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
			return false
		}
		writeInvalidJSON(w, requestID)
		return false
	}
	if dec.More() {
		writeInvalidJSON(w, requestID)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func newConnectionResponse(view googlecalendar.StatusView) ConnectionResponse {
	response := ConnectionResponse{
		Enabled:      view.Enabled,
		BarberLinked: view.BarberLinked,
		Status:       string(googlecalendar.StatusNotConnected),
	}
	if view.Connection != nil {
		fillFromConnection(&response, *view.Connection)
		response.PendingSyncJobs = view.Jobs.Pending
		response.FailedSyncJobs = view.Jobs.Failed
	}
	return response
}

func fillFromConnection(response *ConnectionResponse, conn googlecalendar.Connection) {
	response.Status = string(conn.Status)
	if conn.AccountEmail != "" {
		email := conn.AccountEmail
		response.AccountEmail = &email
	}
	response.ReminderMinutes = conn.ReminderMinutes
	response.ConnectedAt = conn.ConnectedAt
	response.LastSyncedAt = conn.LastSyncedAt
}

// SyncNow expone POST /private/integrations/google-calendar/sync.
func (h *Handlers) SyncNow(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}
	counts, err := h.service.SyncNow(r.Context(), principal.BarbershopID, principal.StaffUserID)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	writeJSON(w, http.StatusOK, SyncResponse{PendingSyncJobs: counts.Pending, FailedSyncJobs: counts.Failed})
}

// Handlers reúne los handlers de la integración.
type Handlers struct {
	service *googlecalendar.Service
}

// New construye los handlers sobre el servicio.
func New(service *googlecalendar.Service) *Handlers {
	return &Handlers{service: service}
}

// Get expone GET /private/integrations/google-calendar.
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}
	view, err := h.service.Status(r.Context(), principal.BarbershopID, principal.StaffUserID)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	writeJSON(w, http.StatusOK, newConnectionResponse(view))
}

// Connect expone POST /private/integrations/google-calendar/connect.
func (h *Handlers) Connect(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}
	url, err := h.service.StartConnect(r.Context(), principal.BarbershopID, principal.StaffUserID, principal.SessionID)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	writeJSON(w, http.StatusOK, AuthorizationResponse{AuthorizationURL: url})
}

// Callback expone POST /private/integrations/google-calendar/callback.
func (h *Handlers) Callback(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}
	var req CallbackRequest
	if !decodeStrict(w, r, requestID, &req) {
		return
	}
	result, err := h.service.CompleteCallback(r.Context(),
		principal.BarbershopID, principal.StaffUserID, principal.SessionID, req.State, req.Code, req.Error)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	writeJSON(w, http.StatusOK, CallbackResponse{Result: string(result)})
}

// reminderRequest distingue «ausente» de «null»: reminderMinutes es obligatorio
// y null significa «recordatorios predeterminados».
type reminderRequest struct {
	ReminderMinutes json.RawMessage `json:"reminderMinutes"`
}

// UpdateReminder expone PATCH /private/integrations/google-calendar.
func (h *Handlers) UpdateReminder(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}
	var req reminderRequest
	if !decodeStrict(w, r, requestID, &req) {
		return
	}
	if len(req.ReminderMinutes) == 0 {
		writeInvalidJSON(w, requestID)
		return
	}
	var minutes *int
	if string(req.ReminderMinutes) != "null" {
		var value int
		if err := json.Unmarshal(req.ReminderMinutes, &value); err != nil {
			writeInvalidJSON(w, requestID)
			return
		}
		minutes = &value
	}

	conn, err := h.service.UpdateReminder(r.Context(), principal.BarbershopID, principal.StaffUserID, minutes)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	response := ConnectionResponse{Enabled: h.service.Enabled(), BarberLinked: true}
	fillFromConnection(&response, conn)
	writeJSON(w, http.StatusOK, response)
}

// Disconnect expone DELETE /private/integrations/google-calendar.
func (h *Handlers) Disconnect(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}
	if err := h.service.Disconnect(r.Context(), principal.BarbershopID, principal.StaffUserID); err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
