// Package httpapi es el adaptador HTTP del módulo shops: decodifica,
// invoca shops.Service y traduce el resultado a HTTP
// (docs/03-desarrollo/estandar-backend-go.md §4). No importa Chi: la
// composición de rutas es responsabilidad exclusiva de cmd/api sobre el
// subrouter privado que internal/platform/httpserver ya construye.
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

// GetBarbershopSettingsHandler expone GET /private/settings/barbershop
// (CA-020-01): lectura autenticada de la configuración básica de la
// barbería activa.
type GetBarbershopSettingsHandler struct {
	service *shops.Service
}

// NewGetBarbershopSettingsHandler construye el handler de lectura.
func NewGetBarbershopSettingsHandler(service *shops.Service) *GetBarbershopSettingsHandler {
	return &GetBarbershopSettingsHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *GetBarbershopSettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())

	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		// Mismo defecto de wiring que SessionContextHandler contempla: solo
		// puede ocurrir si esta ruta se montó fuera del subrouter protegido
		// por SessionMiddleware (CA-006-04 lo evita estructuralmente).
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Internal(errors.New("shops: falta el principal de sesión en el contexto")), requestID))
		return
	}

	barbershop, err := h.service.Get(r.Context(), principal.BarbershopID)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newBarbershopSettingsResponse(barbershop))
}

// UpdateBarbershopSettingsHandler expone PATCH /private/settings/barbershop
// (CA-020-02, CA-020-03, CA-020-06): actualización autenticada de la
// configuración básica de la barbería activa.
type UpdateBarbershopSettingsHandler struct {
	service *shops.Service
}

// NewUpdateBarbershopSettingsHandler construye el handler de actualización.
func NewUpdateBarbershopSettingsHandler(service *shops.Service) *UpdateBarbershopSettingsHandler {
	return &UpdateBarbershopSettingsHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *UpdateBarbershopSettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())

	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Internal(errors.New("shops: falta el principal de sesión en el contexto")), requestID))
		return
	}

	var req UpdateBarbershopSettingsRequest
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
		// Más de un valor JSON en el cuerpo (p. ej. "{}{}"): mismo
		// tratamiento que un JSON malformado (igual que LoginHandler).
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Invalid("cuerpo JSON inválido o con un campo desconocido"), requestID))
		return
	}

	// Normalización, validación de campo y persistencia atómica viven en
	// shops.Service: este handler no repite ninguna regla de negocio
	// (docs/03-desarrollo/estandar-backend-go.md §4).
	barbershop, err := h.service.Update(r.Context(), principal.BarbershopID, shops.Input{
		Name:         req.Name,
		Timezone:     req.Timezone,
		ContactEmail: req.ContactEmail,
		ContactPhone: req.ContactPhone,
	})
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newBarbershopSettingsResponse(barbershop))
}

// newBarbershopSettingsResponse construye la representación canónica
// (CA-020-07): siempre los mismos cuatro campos, en la misma forma tanto
// para GET como para la respuesta 200 de PATCH.
func newBarbershopSettingsResponse(b shops.Barbershop) BarbershopSettingsResponse {
	return BarbershopSettingsResponse{
		Name:         b.Name,
		Timezone:     b.Timezone,
		ContactEmail: b.ContactEmail,
		ContactPhone: b.ContactPhone,
	}
}
