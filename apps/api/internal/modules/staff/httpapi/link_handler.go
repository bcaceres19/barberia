package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"system-barbershop/internal/modules/staff"
	"system-barbershop/internal/platform/httpserver"
)

// GetMyBarberHandler expone GET /private/me/barber (DEC-100): el barbero que
// el usuario autenticado declaró ser, o 404 si no tiene ninguno.
type GetMyBarberHandler struct {
	service *staff.Service
}

// NewGetMyBarberHandler construye el handler de lectura del barbero propio.
func NewGetMyBarberHandler(service *staff.Service) *GetMyBarberHandler {
	return &GetMyBarberHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *GetMyBarberHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	barber, err := h.service.LinkedBarber(r.Context(), principal.BarbershopID, principal.StaffUserID)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	writeBarber(w, barber)
}

// LinkMyBarberHandler expone PUT /private/me/barber (DEC-100). Actúa siempre
// sobre el principal de la sesión; el cuerpo solo nombra al barbero.
type LinkMyBarberHandler struct {
	service *staff.Service
}

// NewLinkMyBarberHandler construye el handler de vínculo propio.
func NewLinkMyBarberHandler(service *staff.Service) *LinkMyBarberHandler {
	return &LinkMyBarberHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *LinkMyBarberHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	var req LinkMyBarberRequest
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

	barber, err := h.service.LinkMyBarber(r.Context(), principal.BarbershopID, principal.StaffUserID, req.BarberID)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	writeBarber(w, barber)
}

// UnlinkMyBarberHandler expone DELETE /private/me/barber (DEC-100).
type UnlinkMyBarberHandler struct {
	service *staff.Service
}

// NewUnlinkMyBarberHandler construye el handler que quita el vínculo propio.
func NewUnlinkMyBarberHandler(service *staff.Service) *UnlinkMyBarberHandler {
	return &UnlinkMyBarberHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *UnlinkMyBarberHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	if err := h.service.UnlinkMyBarber(r.Context(), principal.BarbershopID, principal.StaffUserID); err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeBarber(w http.ResponseWriter, barber staff.Barber) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newBarberResponse(barber))
}
