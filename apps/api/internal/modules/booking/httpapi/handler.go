// Package httpapi es el adaptador HTTP del módulo booking (HU-061):
// decodifica, invoca booking.ManualBookingService y traduce el resultado a
// HTTP (docs/03-desarrollo/estandar-backend-go.md §4), mismo patrón que
// schedule/httpapi y catalog/httpapi.
package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

// agendaBarberIDParam es el nombre del parámetro de ruta que
// cmd/api.buildRouter registra para
// /private/barbers/{barberId}/appointments/daily-agenda (HU-062).
const agendaBarberIDParam = "barberId"

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

// ListDailyAgendaHandler expone
// GET /private/barbers/{barberId}/appointments/daily-agenda (HU-062,
// CA-062-01 a CA-062-07): lectura de la agenda diaria de un único barbero,
// sin idempotencia (operación de solo lectura).
type ListDailyAgendaHandler struct {
	service *booking.AgendaService
}

// NewListDailyAgendaHandler construye el handler de lectura.
func NewListDailyAgendaHandler(service *booking.AgendaService) *ListDailyAgendaHandler {
	return &ListDailyAgendaHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *ListDailyAgendaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	barberID := httpserver.URLParam(r, agendaBarberIDParam)
	date := r.URL.Query().Get("date")

	entries, err := h.service.ListDailyAgenda(r.Context(), principal.BarbershopID, barberID, date)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newDailyAgendaResponse(entries))
}

func newDailyAgendaResponse(entries []booking.DailyAgendaEntry) DailyAgendaResponse {
	items := make([]DailyAgendaEntryResponse, 0, len(entries))
	for _, e := range entries {
		items = append(items, DailyAgendaEntryResponse{
			ID:              e.ID,
			AttendeeName:    e.AttendeeName,
			StartsAt:        e.StartsAt,
			EndsAt:          e.EndsAt,
			Status:          string(e.Status),
			Origin:          string(e.Origin),
			ServiceName:     e.ServiceNameSnapshot,
			DurationMinutes: e.DurationMinutesSnapshot,
			PriceAmount:     formatPriceAmount(e.PriceAmountCentsSnapshot),
			Currency:        e.CurrencySnapshot,
		})
	}
	return DailyAgendaResponse{Items: items}
}

// formatPriceAmount formatea centavos como el string decimal de dos
// decimales que el contrato expone (mismo criterio y mismo resultado que
// bookingpostgres.formatPriceAmount/catalog.FormatPriceCOP; se repite aquí,
// sin importar el paquete postgres, porque httpapi no depende de detalles
// de persistencia, mismo criterio de duplicación mínima que
// booking.LooksLikeBarberID frente a staff/schedule).
func formatPriceAmount(cents int64) string {
	whole := cents / 100
	frac := cents % 100
	return fmt.Sprintf("%d.%02d", whole, frac)
}
