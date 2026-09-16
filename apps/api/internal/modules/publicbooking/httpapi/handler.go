// Package httpapi es el adaptador HTTP del módulo publicbooking: decodifica,
// invoca publicbooking.Service y traduce el resultado a HTTP
// (docs/03-desarrollo/estandar-backend-go.md §4). No importa Chi: la
// composición de rutas es responsabilidad exclusiva de cmd/api sobre el
// router público que internal/platform/httpserver ya construye.
package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"system-barbershop/internal/modules/publicbooking"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
	"system-barbershop/internal/platform/idempotency"
)

// slugParam es el nombre del parámetro de ruta que cmd/api.buildRouter
// registra para /public/barbershops/{slug} (HU-090).
const slugParam = "slug"

// serviceIDParam es el nombre del parámetro de ruta que cmd/api.buildRouter
// registra para /public/barbershops/{slug}/services/{serviceId}/barbers
// (HU-092).
const serviceIDParam = "serviceId"

// barberIDParam es el nombre del parámetro de ruta que cmd/api.buildRouter
// registra para
// /public/barbershops/{slug}/services/{serviceId}/barbers/{barberId}/availability
// (HU-094).
const barberIDParam = "barberId"

// ResolveBarbershopHandler expone GET /public/barbershops/{slug}
// (CA-090-01 a CA-090-04): resolución pública sin sesión, sin que el
// cliente pueda fijar barbershopId.
type ResolveBarbershopHandler struct {
	service *publicbooking.Service
}

// NewResolveBarbershopHandler construye el handler de resolución pública.
func NewResolveBarbershopHandler(service *publicbooking.Service) *ResolveBarbershopHandler {
	return &ResolveBarbershopHandler{service: service}
}

// ServeHTTP implementa http.Handler. Deliberadamente NO lee ningún
// principal de sesión (CA-090-03): el único dato de entrada es el slug de
// la ruta, y el servicio resuelve el tenant por sí mismo antes de leer
// nada.
func (h *ResolveBarbershopHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())

	slug := httpserver.URLParam(r, slugParam)
	profile, err := h.service.ResolveBarbershop(r.Context(), slug)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newPublicBarbershopProfileResponse(profile))
}

func newPublicBarbershopProfileResponse(p publicbooking.BarbershopProfile) PublicBarbershopProfileResponse {
	return PublicBarbershopProfileResponse{
		Name:         p.Name,
		Timezone:     p.Timezone,
		ContactEmail: p.ContactEmail,
		ContactPhone: p.ContactPhone,
	}
}

// ListPublicServicesHandler expone GET /public/barbershops/{slug}/services
// (HU-091, CA-091-01 a CA-091-03): catálogo público de servicios activos y
// asignados de la barbería resuelta por slug, sin sesión.
type ListPublicServicesHandler struct {
	service *publicbooking.Service
}

// NewListPublicServicesHandler construye el handler del catálogo público.
func NewListPublicServicesHandler(service *publicbooking.Service) *ListPublicServicesHandler {
	return &ListPublicServicesHandler{service: service}
}

// ServeHTTP implementa http.Handler. Igual que ResolveBarbershopHandler, no
// lee ningún principal de sesión: slug (de la ruta) y cursor/limit (de la
// query) son toda la entrada.
func (h *ListPublicServicesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())

	slug := httpserver.URLParam(r, slugParam)
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

	result, err := h.service.ListPublicServices(r.Context(), slug, cursor, limit)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newPublicServiceListResponse(result))
}

func newPublicServiceResponse(s publicbooking.PublicService) PublicServiceResponse {
	return PublicServiceResponse{
		ID:              s.ID,
		Name:            s.Name,
		Description:     s.Description,
		DurationMinutes: s.DurationMinutes,
		Price:           publicbooking.FormatServicePriceCOP(s.PriceCents),
		Currency:        s.Currency,
	}
}

func newPublicServiceListResponse(result publicbooking.PublicServiceListResult) PublicServiceListResponse {
	items := make([]PublicServiceResponse, 0, len(result.Items))
	for _, svc := range result.Items {
		items = append(items, newPublicServiceResponse(svc))
	}
	var nextCursor *string
	if result.NextCursor != "" {
		nextCursor = &result.NextCursor
	}
	return PublicServiceListResponse{Items: items, NextCursor: nextCursor}
}

// ListPublicBarbersHandler expone
// GET /public/barbershops/{slug}/services/{serviceId}/barbers (HU-092,
// CA-092-01 a CA-092-03): barberos con asignación vigente al servicio activo
// resuelto, sin sesión.
type ListPublicBarbersHandler struct {
	service *publicbooking.Service
}

// NewListPublicBarbersHandler construye el handler de selección pública de
// barbero.
func NewListPublicBarbersHandler(service *publicbooking.Service) *ListPublicBarbersHandler {
	return &ListPublicBarbersHandler{service: service}
}

// ServeHTTP implementa http.Handler. Igual que los otros handlers públicos,
// no lee ningún principal de sesión: slug y serviceId (ambos de la ruta) son
// toda la entrada.
func (h *ListPublicBarbersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())

	slug := httpserver.URLParam(r, slugParam)
	serviceID := httpserver.URLParam(r, serviceIDParam)

	result, err := h.service.ListPublicBarbers(r.Context(), slug, serviceID)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newPublicBarberListResponse(result))
}

func newPublicBarberResponse(b publicbooking.PublicBarber) PublicBarberResponse {
	return PublicBarberResponse{ID: b.ID, FullName: b.FullName}
}

func newPublicBarberListResponse(result publicbooking.PublicBarberListResult) PublicBarberListResponse {
	items := make([]PublicBarberResponse, 0, len(result.Items))
	for _, barber := range result.Items {
		items = append(items, newPublicBarberResponse(barber))
	}
	return PublicBarberListResponse{Items: items}
}

// ListPublicAvailabilityHandler expone
// GET /public/barbershops/{slug}/services/{serviceId}/barbers/{barberId}/availability
// (HU-094, CA-094-01 a CA-094-06): inicios públicos válidos del servicio
// activo con el barbero elegido, sin sesión.
type ListPublicAvailabilityHandler struct {
	service *publicbooking.AvailabilityService
}

// NewListPublicAvailabilityHandler construye el handler de disponibilidad
// pública.
func NewListPublicAvailabilityHandler(service *publicbooking.AvailabilityService) *ListPublicAvailabilityHandler {
	return &ListPublicAvailabilityHandler{service: service}
}

// ServeHTTP implementa http.Handler. Igual que los otros handlers públicos,
// no lee ningún principal de sesión: slug, serviceId y barberId (los tres
// de la ruta) son toda la entrada.
func (h *ListPublicAvailabilityHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())

	slug := httpserver.URLParam(r, slugParam)
	serviceID := httpserver.URLParam(r, serviceIDParam)
	barberID := httpserver.URLParam(r, barberIDParam)

	result, err := h.service.ListPublicAvailability(r.Context(), slug, serviceID, barberID)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newAvailabilityResponse(result))
}

func newAvailabilityResponse(result publicbooking.AvailabilityResult) AvailabilityResponse {
	slots := make([]AvailabilitySlotResponse, 0, len(result.Slots))
	for _, slot := range result.Slots {
		slots = append(slots, AvailabilitySlotResponse{StartsAt: slot.StartsAt})
	}
	return AvailabilityResponse{
		Slots:           slots,
		DurationMinutes: result.DurationMinutes,
		Timezone:        result.Timezone,
		SlotGridMinutes: result.SlotGridMinutes,
	}
}

// writeUnknownFieldOrInvalidJSONProblem cubre un cuerpo JSON malformado o
// con un campo desconocido, mismo criterio de rechazo estricto que
// bookinghttpapi.writeUnknownFieldOrInvalidJSONProblem (duplicado a
// propósito: publicbooking/httpapi no importa booking/httpapi, CA-002-06).
func writeUnknownFieldOrInvalidJSONProblem(w http.ResponseWriter, requestID string) {
	httpserver.WriteProblem(w, httpserver.Translate(
		apperr.Invalid("cuerpo JSON inválido o con un campo desconocido"), requestID))
}

// ConfirmPublicAppointmentHandler expone POST /public/barbershops/{slug}/
// services/{serviceId}/barbers/{barberId}/appointments (HU-097, T1
// pública), protegido por el protocolo de idempotencia reutilizable de
// HU-004 (RN-IDE-01, DEC-043). Igual que los otros handlers públicos, no
// lee ningún principal de sesión: slug/serviceId/barberId (de la ruta) y el
// cuerpo son toda la entrada.
type ConfirmPublicAppointmentHandler struct {
	service *publicbooking.ConfirmationService
	logger  *slog.Logger
}

// NewConfirmPublicAppointmentHandler construye el handler de confirmación
// pública. logger registra únicamente un fallo de entrega del correo de
// confirmación, sin destinatario, código ni enlace (RN-DAT-02, DEC-091),
// mismo criterio que auth/httpapi.RecoveryRequestHandler.
func NewConfirmPublicAppointmentHandler(service *publicbooking.ConfirmationService, logger *slog.Logger) *ConfirmPublicAppointmentHandler {
	return &ConfirmPublicAppointmentHandler{service: service, logger: logger}
}

// ServeHTTP implementa http.Handler.
func (h *ConfirmPublicAppointmentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())

	slug := httpserver.URLParam(r, slugParam)
	serviceID := httpserver.URLParam(r, serviceIDParam)
	barberID := httpserver.URLParam(r, barberIDParam)

	// 1. Cabecera: obligatoria, validada antes de tocar PostgreSQL o leer el
	//    cuerpo (mismo orden que bookinghttpapi.CreateManualAppointmentHandler).
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

	var req ConfirmPublicAppointmentRequest
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

	result, emailErr, err := h.service.ConfirmAppointment(r.Context(), publicbooking.ConfirmPublicAppointmentInput{
		Slug:        slug,
		ServiceID:   serviceID,
		BarberID:    barberID,
		StartsAtRaw: req.StartsAt,
		Identity: publicbooking.CustomerIdentityInput{
			FullName:       req.FullName,
			Phone:          req.Phone,
			Email:          req.Email,
			Note:           req.Note,
			ForSomeoneElse: req.ForSomeoneElse,
			AttendeeName:   req.AttendeeName,
		},
	}, key, fingerprint)
	if err != nil {
		var conflict *publicbooking.ScheduleConflictError
		if errors.As(err, &conflict) {
			writeScheduleConflictProblem(w, requestID, conflict)
			return
		}
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	if emailErr != nil {
		// El resultado se descarta deliberadamente para la respuesta (la
		// cita ya quedó confirmada, DEC-091); solo se registra sin
		// destinatario, código ni enlace (RN-DAT-02).
		h.logger.WarnContext(r.Context(), "publicbooking: fallo al enviar el correo de confirmación", "requestId", requestID)
	}

	switch result.Decision.Outcome {
	case idempotency.OutcomeProceed, idempotency.OutcomeReplay:
		httpserver.WriteStoredResponse(w, result.Response)
	default:
		httpserver.WriteProblem(w, httpserver.Translate(result.Decision.AsError(), requestID))
	}
}

// writeScheduleConflictProblem escribe el 409 de RN-CON-05/DEC-090:
// httpserver.Translate calcula la parte segura y común (type/title/code/
// status/detail); esta función solo agrega `alternatives` (RFC 9457,
// extensión abierta que Problem.yaml ya declara) antes de codificar.
func writeScheduleConflictProblem(w http.ResponseWriter, requestID string, conflict *publicbooking.ScheduleConflictError) {
	problem := httpserver.Translate(conflict, requestID)
	alternatives := make([]AlternativeSlotResponse, 0, len(conflict.Alternatives))
	for _, alt := range conflict.Alternatives {
		alternatives = append(alternatives, AlternativeSlotResponse{StartsAt: alt.StartsAt})
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(problem.Status)
	_ = json.NewEncoder(w).Encode(ScheduleConflictProblemResponse{Problem: problem, Alternatives: alternatives})
}
