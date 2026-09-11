package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/modules/shops"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
)

// bookingPolicyIfMatchHeader transporta el token opaco de versión (HU-093)
// como precondición de concurrencia optimista, mismo criterio que
// booking.ifMatchHeader (HU-064/HU-065): cabecera obligatoria para una
// escritura que primero lee una representación.
const bookingPolicyIfMatchHeader = "If-Match"

// GetBookingPolicyHandler expone GET /private/settings/booking-policy
// (HU-093, CA-093-01): lectura autenticada de la política pública de
// reserva y cancelación de la barbería activa.
type GetBookingPolicyHandler struct {
	service *shops.BookingPolicyService
}

// NewGetBookingPolicyHandler construye el handler de lectura.
func NewGetBookingPolicyHandler(service *shops.BookingPolicyService) *GetBookingPolicyHandler {
	return &GetBookingPolicyHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *GetBookingPolicyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())

	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Internal(errors.New("shops: falta el principal de sesión en el contexto")), requestID))
		return
	}

	policy, err := h.service.Get(r.Context(), principal.BarbershopID)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newBookingPolicyResponse(policy))
}

// UpdateBookingPolicyHandler expone PUT /private/settings/booking-policy
// (HU-093, CA-093-02 a CA-093-04): actualización autenticada de contrato
// cerrado, protegida por la precondición de versión de If-Match.
type UpdateBookingPolicyHandler struct {
	service *shops.BookingPolicyService
}

// NewUpdateBookingPolicyHandler construye el handler de actualización.
func NewUpdateBookingPolicyHandler(service *shops.BookingPolicyService) *UpdateBookingPolicyHandler {
	return &UpdateBookingPolicyHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *UpdateBookingPolicyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())

	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Internal(errors.New("shops: falta el principal de sesión en el contexto")), requestID))
		return
	}

	// La cabecera se valida antes de leer el cuerpo (mismo orden que
	// RescheduleAppointmentHandler): una precondición ausente no debe
	// gastar el trabajo de decodificar JSON.
	versionToken := r.Header.Get(bookingPolicyIfMatchHeader)
	if versionToken == "" {
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Invalid("falta la cabecera If-Match con el token de versión de la política de reserva"), requestID))
		return
	}

	var req UpdateBookingPolicyRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
			return
		}
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Invalid("cuerpo JSON inválido o con un campo desconocido"), requestID))
		return
	}
	if dec.More() {
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Invalid("cuerpo JSON inválido o con un campo desconocido"), requestID))
		return
	}

	// Normalización, validación de rango/coherencia y persistencia atómica
	// viven en shops.BookingPolicyService: este handler no repite ninguna
	// regla de negocio (docs/03-desarrollo/estandar-backend-go.md §4).
	policy, err := h.service.Update(r.Context(), principal.BarbershopID, shops.BookingPolicyInput{
		MinAdvanceMinutes:              req.MinAdvanceMinutes,
		MaxAdvanceDays:                 req.MaxAdvanceDays,
		SlotGridMinutes:                req.SlotGridMinutes,
		CancellationDeadlineMinutes:    req.CancellationDeadlineMinutes,
		LateCancellationClientAllowed:  req.LateCancellationClientAllowed,
		LateCancellationReasonRequired: req.LateCancellationReasonRequired,
		ExpectedVersionToken:           versionToken,
	})
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newBookingPolicyResponse(policy))
}

// newBookingPolicyResponse construye la representación canónica
// (CA-093-01): siempre los mismos siete campos, en la misma forma tanto
// para GET como para la respuesta 200 de PUT.
func newBookingPolicyResponse(p shops.BookingPolicy) BookingPolicyResponse {
	return BookingPolicyResponse{
		MinAdvanceMinutes:              p.MinAdvanceMinutes,
		MaxAdvanceDays:                 p.MaxAdvanceDays,
		SlotGridMinutes:                p.SlotGridMinutes,
		CancellationDeadlineMinutes:    p.CancellationDeadlineMinutes,
		LateCancellationClientAllowed:  p.LateCancellationClientAllowed,
		LateCancellationReasonRequired: p.LateCancellationReasonRequired,
		VersionToken:                   p.VersionToken,
	}
}
