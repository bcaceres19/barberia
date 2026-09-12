// Package httpapi es el adaptador HTTP del módulo publicbooking: decodifica,
// invoca publicbooking.Service y traduce el resultado a HTTP
// (docs/03-desarrollo/estandar-backend-go.md §4). No importa Chi: la
// composición de rutas es responsabilidad exclusiva de cmd/api sobre el
// router público que internal/platform/httpserver ya construye.
package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"system-barbershop/internal/modules/publicbooking"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
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
