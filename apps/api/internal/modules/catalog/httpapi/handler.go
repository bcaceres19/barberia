// Package httpapi es el adaptador HTTP del módulo catalog: decodifica,
// invoca catalog.CatalogService y traduce el resultado a HTTP
// (docs/03-desarrollo/estandar-backend-go.md §4). No importa Chi
// directamente: usa httpserver.URLParam para leer {serviceId}, igual que
// httpserver.RequestIDFromContext/IdempotencyKeyFromRequest para el resto
// de valores que Chi o net/http exponen (DEC-034), mismo patrón que
// staff/httpapi.
package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/modules/catalog"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
	"system-barbershop/internal/platform/idempotency"
)

// serviceIDParam es el nombre del parámetro de ruta que cmd/api.buildRouter
// registra para /private/services/{serviceId}.
const serviceIDParam = "serviceId"

// serviceBasePath es la base usada para construir el header Location de un
// alta exitosa, relativa a /api/v1 (igual que el resto del contrato).
const serviceBasePath = "/private/services/"

func principalOrInternalError(w http.ResponseWriter, r *http.Request, requestID string) (auth.Principal, bool) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		// Mismo defecto de wiring que staff/httpapi contempla: solo puede
		// ocurrir si esta ruta se montó fuera del subrouter protegido por
		// SessionMiddleware (CA-006-04 lo evita estructuralmente).
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Internal(errors.New("catalog: falta el principal de sesión en el contexto")), requestID))
		return auth.Principal{}, false
	}
	return principal, true
}

func writeUnknownFieldOrInvalidJSONProblem(w http.ResponseWriter, requestID string) {
	httpserver.WriteProblem(w, httpserver.Translate(
		apperr.Invalid("cuerpo JSON inválido o con un campo desconocido"), requestID))
}

// ListServicesHandler expone GET /private/services (CA-022-01).
type ListServicesHandler struct {
	service *catalog.CatalogService
}

// NewListServicesHandler construye el handler de listado.
func NewListServicesHandler(service *catalog.CatalogService) *ListServicesHandler {
	return &ListServicesHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *ListServicesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

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

	result, err := h.service.List(r.Context(), principal.BarbershopID, cursor, limit)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newServiceListResponse(result))
}

// GetServiceHandler expone GET /private/services/{serviceId} (CA-022-06: la
// lectura individual permite demostrar de forma directa el 404 de un
// servicio de otra barbería).
type GetServiceHandler struct {
	service *catalog.CatalogService
}

// NewGetServiceHandler construye el handler de lectura individual.
func NewGetServiceHandler(service *catalog.CatalogService) *GetServiceHandler {
	return &GetServiceHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *GetServiceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	serviceID := httpserver.URLParam(r, serviceIDParam)
	svc, err := h.service.Get(r.Context(), principal.BarbershopID, serviceID)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newServiceResponse(svc))
}

// CreateServiceHandler expone POST /private/services (CA-022-02),
// protegido por el protocolo de idempotencia reutilizable de HU-004
// (RN-IDE-01, DEC-043).
type CreateServiceHandler struct {
	service *catalog.CatalogService
}

// NewCreateServiceHandler construye el handler de alta.
func NewCreateServiceHandler(service *catalog.CatalogService) *CreateServiceHandler {
	return &CreateServiceHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *CreateServiceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	var req CreateServiceRequest
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

	raw := catalog.CreateInputRaw{
		Name:            req.Name,
		DurationMinutes: req.DurationMinutes,
		Price:           req.Price,
	}
	if req.Description != nil {
		raw.Description = *req.Description
	}

	result, err := h.service.Create(r.Context(), principal.BarbershopID, raw, key, fingerprint)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	switch result.Decision.Outcome {
	case idempotency.OutcomeProceed, idempotency.OutcomeReplay:
		// El mismo result.Response, byte a byte, para la primera ejecución
		// y para cualquier repetición exacta (CA-004-01). Location se
		// reconstruye a partir del `id` del cuerpo ya persistido.
		var parsed struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal([]byte(result.Response.Body), &parsed)
		w.Header().Set("Location", serviceBasePath+parsed.ID)
		httpserver.WriteStoredResponse(w, result.Response)
	default:
		// Conflicto de contenido/operación (RN-IDE-01) o ejecución en curso
		// (DEC-043): Translate ya sabe convertirlo en 409.
		httpserver.WriteProblem(w, httpserver.Translate(result.Decision.AsError(), requestID))
	}
}

// UpdateServiceHandler expone PATCH /private/services/{serviceId}
// (CA-022-04, CA-022-05).
type UpdateServiceHandler struct {
	service *catalog.CatalogService
}

// NewUpdateServiceHandler construye el handler de edición parcial.
func NewUpdateServiceHandler(service *catalog.CatalogService) *UpdateServiceHandler {
	return &UpdateServiceHandler{service: service}
}

// ServeHTTP implementa http.Handler. La ausencia de una clave en el cuerpo
// (json.Decode no toca ese puntero) es la única señal de "no editar este
// campo": un valor explícito `null` de un cliente fuera del contrato (que
// declara los cuatro campos como `string`/`integer`, nunca nullable) se
// interpreta de la misma forma segura ("no editar"), sin distinguirlo de la
// ausencia. Ninguna de las dos formas puede tocar isActive, asignaciones,
// citas ni alcance de propagación: el contrato ni siquiera declara esos
// campos (DEC-047-style closed schema, RN-SER-04).
func (h *UpdateServiceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	serviceID := httpserver.URLParam(r, serviceIDParam)

	var req UpdateServiceRequest
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

	raw := catalog.UpdateInputRaw{
		Name:            req.Name,
		DescriptionSet:  req.Description != nil,
		Description:     req.Description,
		DurationMinutes: req.DurationMinutes,
		Price:           req.Price,
	}

	svc, err := h.service.Update(r.Context(), principal.BarbershopID, serviceID, raw)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newServiceResponse(svc))
}

func newServiceResponse(svc catalog.Service) ServiceResponse {
	return ServiceResponse{
		ID:              svc.ID,
		Name:            svc.Name,
		Description:     svc.Description,
		DurationMinutes: svc.DurationMinutes,
		Price:           catalog.FormatPriceCOP(svc.PriceCents),
		Currency:        svc.Currency,
		CreatedAt:       svc.CreatedAt,
		UpdatedAt:       svc.UpdatedAt,
	}
}

func newServiceListResponse(result catalog.ListResult) ServiceListResponse {
	items := make([]ServiceResponse, 0, len(result.Items))
	for _, svc := range result.Items {
		items = append(items, newServiceResponse(svc))
	}
	var next *string
	if result.NextCursor != "" {
		v := result.NextCursor
		next = &v
	}
	return ServiceListResponse{Items: items, NextCursor: next}
}
