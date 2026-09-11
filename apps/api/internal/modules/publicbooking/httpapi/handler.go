// Package httpapi es el adaptador HTTP del módulo publicbooking: decodifica,
// invoca publicbooking.Service y traduce el resultado a HTTP
// (docs/03-desarrollo/estandar-backend-go.md §4). No importa Chi: la
// composición de rutas es responsabilidad exclusiva de cmd/api sobre el
// router público que internal/platform/httpserver ya construye.
package httpapi

import (
	"encoding/json"
	"net/http"

	"system-barbershop/internal/modules/publicbooking"
	"system-barbershop/internal/platform/httpserver"
)

// slugParam es el nombre del parámetro de ruta que cmd/api.buildRouter
// registra para /public/barbershops/{slug} (HU-090).
const slugParam = "slug"

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
