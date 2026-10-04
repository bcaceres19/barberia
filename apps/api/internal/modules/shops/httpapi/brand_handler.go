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

// GetBrandHandler expone GET /private/settings/brand (issue #292): lectura
// autenticada de la marca y el vocabulario de la barbería activa.
type GetBrandHandler struct {
	service *shops.BrandService
}

// NewGetBrandHandler construye el handler de lectura.
func NewGetBrandHandler(service *shops.BrandService) *GetBrandHandler {
	return &GetBrandHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *GetBrandHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())

	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Internal(errors.New("shops: falta el principal de sesión en el contexto")), requestID))
		return
	}

	brand, err := h.service.Get(r.Context(), principal.BarbershopID)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newBrandResponse(brand))
}

// UpdateBrandHandler expone PATCH /private/settings/brand: actualización
// autenticada de contrato cerrado, con los seis campos siempre presentes.
type UpdateBrandHandler struct {
	service *shops.BrandService
}

// NewUpdateBrandHandler construye el handler de actualización.
func NewUpdateBrandHandler(service *shops.BrandService) *UpdateBrandHandler {
	return &UpdateBrandHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *UpdateBrandHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())

	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Internal(errors.New("shops: falta el principal de sesión en el contexto")), requestID))
		return
	}

	var req UpdateBrandRequest
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

	// Normalización, validación y persistencia viven en shops.BrandService:
	// este handler no repite ninguna regla (estandar-backend-go.md §4).
	brand, err := h.service.Update(r.Context(), principal.BarbershopID, shops.Brand{
		Accent:                 req.Accent,
		BusinessTerm:           req.BusinessTerm,
		BusinessTermGender:     shops.Gender(req.BusinessTermGender),
		ProfessionalTerm:       req.ProfessionalTerm,
		ProfessionalTermPlural: req.ProfessionalTermPlural,
		ProfessionalTermGender: shops.Gender(req.ProfessionalTermGender),
	})
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newBrandResponse(brand))
}

// newBrandResponse construye la representación canónica: siempre los mismos
// seis campos, en la misma forma para GET y para la respuesta 200 de PATCH.
func newBrandResponse(b shops.Brand) BrandResponse {
	return BrandResponse{
		Accent:                 b.Accent,
		BusinessTerm:           b.BusinessTerm,
		BusinessTermGender:     string(b.BusinessTermGender),
		ProfessionalTerm:       b.ProfessionalTerm,
		ProfessionalTermPlural: b.ProfessionalTermPlural,
		ProfessionalTermGender: string(b.ProfessionalTermGender),
	}
}
