// Package httpapi es el adaptador HTTP del módulo booking (HU-061):
// decodifica, invoca booking.ManualBookingService y traduce el resultado a
// HTTP (docs/03-desarrollo/estandar-backend-go.md §4), mismo patrón que
// schedule/httpapi y catalog/httpapi.
package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/modules/booking"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
	"system-barbershop/internal/platform/idempotency"
)

// appointmentBasePath es la base usada para construir el header Location
// de una cita creada, relativa a /api/v1.
const appointmentBasePath = "/private/appointments/"

func principalOrInternalError(w http.ResponseWriter, r *http.Request, requestID string) (auth.Principal, bool) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Internal(errors.New("booking: falta el principal de sesión en el contexto")), requestID))
		return auth.Principal{}, false
	}
	return principal, true
}

func writeUnknownFieldOrInvalidJSONProblem(w http.ResponseWriter, requestID string) {
	httpserver.WriteProblem(w, httpserver.Translate(
		apperr.Invalid("cuerpo JSON inválido o con un campo desconocido"), requestID))
}

// CreateManualAppointmentHandler expone POST /private/appointments
// (CA-061-01 a CA-061-08), protegido por el protocolo de idempotencia
// reutilizable de HU-004 (RN-IDE-01, DEC-043).
type CreateManualAppointmentHandler struct {
	service *booking.ManualBookingService
}

// NewCreateManualAppointmentHandler construye el handler de alta.
func NewCreateManualAppointmentHandler(service *booking.ManualBookingService) *CreateManualAppointmentHandler {
	return &CreateManualAppointmentHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *CreateManualAppointmentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

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

	var req CreateManualAppointmentRequest
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

	result, err := h.service.CreateManualAppointment(r.Context(), principal.BarbershopID, booking.CreateManualAppointmentRequest{
		BarberID:         req.BarberID,
		ServiceID:        req.ServiceID,
		AttendeeName:     req.AttendeeName,
		CustomerFullName: req.CustomerFullName,
		CustomerPhone:    req.CustomerPhone,
		CustomerEmail:    req.CustomerEmail,
		CustomerNote:     req.CustomerNote,
		StartsAtLocal:    req.StartsAt,
		ActorStaffUserID: principal.StaffUserID,
	}, key, fingerprint)
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
		w.Header().Set("Location", appointmentBasePath+parsed.ID)
		httpserver.WriteStoredResponse(w, result.Response)
	default:
		// Conflicto de contenido/operación (RN-IDE-01) o ejecución en curso
		// (DEC-043): Translate ya sabe convertirlo en 409.
		httpserver.WriteProblem(w, httpserver.Translate(result.Decision.AsError(), requestID))
	}
}
