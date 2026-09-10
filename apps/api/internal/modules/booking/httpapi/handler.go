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
	"strconv"

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

// appointmentIDParam es el nombre del parámetro de ruta que
// cmd/api.buildRouter registra para /private/appointments/{appointmentId}
// (HU-064).
const appointmentIDParam = "appointmentId"

// GetAppointmentDetailHandler expone
// GET /private/appointments/{appointmentId} (HU-064, CA-064-01 a CA-064-04):
// lectura de solo detalle, sin idempotencia (operación de solo lectura).
type GetAppointmentDetailHandler struct {
	service *booking.DetailService
}

// NewGetAppointmentDetailHandler construye el handler de lectura.
func NewGetAppointmentDetailHandler(service *booking.DetailService) *GetAppointmentDetailHandler {
	return &GetAppointmentDetailHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *GetAppointmentDetailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	appointmentID := httpserver.URLParam(r, appointmentIDParam)
	detail, err := h.service.GetDetail(r.Context(), principal.BarbershopID, appointmentID)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newAppointmentDetailResponse(detail))
}

func newAppointmentDetailResponse(d booking.AppointmentDetail) AppointmentDetailResponse {
	return AppointmentDetailResponse{
		ID:               d.ID,
		BarberID:         d.BarberID,
		BarberFullName:   d.BarberFullName,
		AttendeeName:     d.AttendeeName,
		CustomerFullName: d.CustomerFullName,
		CustomerPhone:    d.CustomerPhone,
		CustomerEmail:    d.CustomerEmail,
		CustomerNote:     d.CustomerNote,
		StartsAt:         d.StartsAt,
		EndsAt:           d.EndsAt,
		Status:           string(d.Status),
		Origin:           string(d.Origin),
		ServiceName:      d.ServiceNameSnapshot,
		DurationMinutes:  d.DurationMinutesSnapshot,
		PriceAmount:      formatPriceAmount(d.PriceAmountCentsSnapshot),
		Currency:         d.CurrencySnapshot,
		VersionToken:     d.VersionToken,
		CreatedAt:        d.CreatedAt,
	}
}

// ListAppointmentHistoryHandler expone
// GET /private/appointments/{appointmentId}/history (HU-064, CA-064-05):
// lectura paginada por cursor opaco, sin idempotencia (operación de solo
// lectura), mismo patrón de parámetros que ListBarbersHandler
// (staff/httpapi).
type ListAppointmentHistoryHandler struct {
	service *booking.DetailService
}

// NewListAppointmentHistoryHandler construye el handler de lectura.
func NewListAppointmentHistoryHandler(service *booking.DetailService) *ListAppointmentHistoryHandler {
	return &ListAppointmentHistoryHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *ListAppointmentHistoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	appointmentID := httpserver.URLParam(r, appointmentIDParam)
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

	page, err := h.service.ListHistory(r.Context(), principal.BarbershopID, appointmentID, cursor, limit)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newAppointmentHistoryResponse(page))
}

func newAppointmentHistoryResponse(page booking.HistoryPage) AppointmentHistoryResponse {
	items := make([]AppointmentHistoryEntryResponse, 0, len(page.Items))
	for _, entry := range page.Items {
		items = append(items, AppointmentHistoryEntryResponse{
			ID:         entry.ID,
			EventType:  string(entry.EventType),
			ActorType:  string(entry.ActorType),
			ActorLabel: entry.ActorLabel,
			Reason:     entry.Reason,
			OccurredAt: entry.OccurredAt,
			Changes:    newAppointmentHistoryChangeResponses(entry.Changes),
		})
	}
	var next *string
	if page.NextCursor != "" {
		v := page.NextCursor
		next = &v
	}
	return AppointmentHistoryResponse{Items: items, NextCursor: next}
}

func newAppointmentHistoryChangeResponses(changes []booking.HistoryChange) []AppointmentHistoryChangeResponse {
	items := make([]AppointmentHistoryChangeResponse, 0, len(changes))
	for _, c := range changes {
		items = append(items, AppointmentHistoryChangeResponse{
			FieldName:     c.FieldName,
			PreviousValue: c.PreviousValue,
			NewValue:      c.NewValue,
		})
	}
	return items
}

// ifMatchHeader transporta el token opaco de versión (HU-064) como
// precondición de concurrencia optimista (HU-065): mismo criterio de
// cabecera obligatoria para una escritura crítica que
// httpserver.IdempotencyKeyHeader, pero exclusivo de operaciones que leen
// una representación antes de mutarla.
const ifMatchHeader = "If-Match"

// RescheduleAppointmentHandler expone
// POST /private/appointments/{appointmentId}/reschedule (HU-065, T2,
// CA-065-01 a CA-065-08), protegido por el protocolo de idempotencia
// reutilizable de HU-004 (RN-IDE-01, DEC-043) y por la precondición de
// versión de HU-064 (cabecera If-Match).
type RescheduleAppointmentHandler struct {
	service *booking.RescheduleService
}

// NewRescheduleAppointmentHandler construye el handler de reprogramación.
func NewRescheduleAppointmentHandler(service *booking.RescheduleService) *RescheduleAppointmentHandler {
	return &RescheduleAppointmentHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *RescheduleAppointmentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	appointmentID := httpserver.URLParam(r, appointmentIDParam)

	// 1. Cabeceras: obligatorias, validadas antes de tocar PostgreSQL o leer
	//    el cuerpo (mismo orden que CreateManualAppointmentHandler).
	key, err := httpserver.IdempotencyKeyFromRequest(r)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	versionToken := r.Header.Get(ifMatchHeader)
	if versionToken == "" {
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Invalid("falta la cabecera If-Match con el token de versión del turno"), requestID))
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

	var req RescheduleAppointmentRequest
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

	result, err := h.service.RescheduleAppointment(r.Context(), principal.BarbershopID, booking.RescheduleAppointmentRequest{
		AppointmentID:        appointmentID,
		NewStartsAtLocal:     req.StartsAt,
		ExpectedVersionToken: versionToken,
		ActorStaffUserID:     principal.StaffUserID,
	}, key, fingerprint)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	switch result.Decision.Outcome {
	case idempotency.OutcomeProceed, idempotency.OutcomeReplay:
		httpserver.WriteStoredResponse(w, result.Response)
	default:
		// Conflicto de idempotencia (RN-IDE-01) u operación en curso
		// (DEC-043): Translate ya sabe convertirlo en 409. Los conflictos de
		// agenda, estado y versión ya se tradujeron arriba, antes de llegar
		// a esta rama (nunca llegan como una Decision distinta de Proceed).
		httpserver.WriteProblem(w, httpserver.Translate(result.Decision.AsError(), requestID))
	}
}

// CancelAppointmentByBarberHandler expone
// POST /private/appointments/{appointmentId}/cancel (HU-066, T6,
// CA-066-01 a CA-066-08), protegido por el protocolo de idempotencia
// reutilizable de HU-004 (RN-IDE-01, DEC-043) y por la precondición de
// versión de HU-064 (cabecera If-Match). Sin cuerpo de solicitud: T6 no
// acepta ningún campo (CA-066-02, "el servidor no acepta un estado o actor
// enviado por el body").
type CancelAppointmentByBarberHandler struct {
	service *booking.CancelAppointmentByBarberService
}

// NewCancelAppointmentByBarberHandler construye el handler de cancelación.
func NewCancelAppointmentByBarberHandler(service *booking.CancelAppointmentByBarberService) *CancelAppointmentByBarberHandler {
	return &CancelAppointmentByBarberHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *CancelAppointmentByBarberHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	appointmentID := httpserver.URLParam(r, appointmentIDParam)

	// 1. Cabeceras: obligatorias, validadas antes de tocar PostgreSQL o leer
	//    el cuerpo (mismo orden que RescheduleAppointmentHandler).
	key, err := httpserver.IdempotencyKeyFromRequest(r)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	versionToken := r.Header.Get(ifMatchHeader)
	if versionToken == "" {
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Invalid("falta la cabecera If-Match con el token de versión del turno"), requestID))
		return
	}

	// 2. Cuerpo CRUDO, leído una sola vez (mismo criterio que
	//    RescheduleAppointmentHandler: la huella de idempotencia debe
	//    calcularse sobre los bytes exactos recibidos). Este comando no
	//    acepta ningún campo (CA-066-02): un cuerpo vacío/ausente se acepta
	//    tal cual (mismo criterio que POST /private/auth/logout), y
	//    cualquier cuerpo no vacío debe decodificar como un objeto JSON sin
	//    propiedades — cualquier campo, conocido o no, se rechaza como
	//    "cerrado" (el comando no tiene ningún campo que aceptar).
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
	if len(body) > 0 {
		var empty struct{}
		dec := json.NewDecoder(bytes.NewReader(body))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&empty); err != nil {
			writeUnknownFieldOrInvalidJSONProblem(w, requestID)
			return
		}
		if dec.More() {
			writeUnknownFieldOrInvalidJSONProblem(w, requestID)
			return
		}
	}

	fingerprint := httpserver.IdempotencyFingerprint(r, body)

	result, err := h.service.CancelAppointmentByBarber(r.Context(), principal.BarbershopID, booking.CancelAppointmentByBarberRequest{
		AppointmentID:        appointmentID,
		ExpectedVersionToken: versionToken,
		ActorStaffUserID:     principal.StaffUserID,
	}, key, fingerprint)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	switch result.Decision.Outcome {
	case idempotency.OutcomeProceed, idempotency.OutcomeReplay:
		httpserver.WriteStoredResponse(w, result.Response)
	default:
		// Conflicto de idempotencia (RN-IDE-01) u operación en curso
		// (DEC-043): Translate ya sabe convertirlo en 409. Los conflictos de
		// estado y versión ya se tradujeron arriba, antes de llegar a esta
		// rama (nunca llegan como una Decision distinta de Proceed).
		httpserver.WriteProblem(w, httpserver.Translate(result.Decision.AsError(), requestID))
	}
}
