// Package httpapi es el adaptador HTTP del módulo staff: decodifica,
// invoca staff.Service y traduce el resultado a HTTP
// (docs/03-desarrollo/estandar-backend-go.md §4). No importa Chi
// directamente: usa httpserver.URLParam para leer {barberId}, igual que
// httpserver.RequestIDFromContext/IdempotencyKeyFromRequest para el resto
// de valores que Chi o net/http exponen (DEC-034,
// internal/platform/archtest/chi_boundary_test.go).
package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/modules/staff"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
	"system-barbershop/internal/platform/idempotency"
)

// barberIDParam es el nombre del parámetro de ruta que cmd/api.buildRouter
// registra para /private/barbers/{barberId}.
const barberIDParam = "barberId"

// barberBasePath es la base usada para construir el header Location de un
// alta exitosa, relativa a /api/v1 (igual que el resto del contrato).
const barberBasePath = "/private/barbers/"

func principalOrInternalError(w http.ResponseWriter, r *http.Request, requestID string) (auth.Principal, bool) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		// Mismo defecto de wiring que shops/httpapi contempla: solo puede
		// ocurrir si esta ruta se montó fuera del subrouter protegido por
		// SessionMiddleware (CA-006-04 lo evita estructuralmente).
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Internal(errors.New("staff: falta el principal de sesión en el contexto")), requestID))
		return auth.Principal{}, false
	}
	return principal, true
}

func writeUnknownFieldOrInvalidJSONProblem(w http.ResponseWriter, requestID string) {
	httpserver.WriteProblem(w, httpserver.Translate(
		apperr.Invalid("cuerpo JSON inválido o con un campo desconocido"), requestID))
}

// ListBarbersHandler expone GET /private/barbers (CA-021-01, CA-021-02).
type ListBarbersHandler struct {
	service *staff.Service
}

// NewListBarbersHandler construye el handler de listado.
func NewListBarbersHandler(service *staff.Service) *ListBarbersHandler {
	return &ListBarbersHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *ListBarbersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	_ = json.NewEncoder(w).Encode(newBarberListResponse(result))
}

// GetBarberHandler expone GET /private/barbers/{barberId} (CA-021-05: la
// lectura individual permite demostrar de forma directa el 404 de un
// barbero de otra barbería).
type GetBarberHandler struct {
	service *staff.Service
}

// NewGetBarberHandler construye el handler de lectura individual.
func NewGetBarberHandler(service *staff.Service) *GetBarberHandler {
	return &GetBarberHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *GetBarberHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	barberID := httpserver.URLParam(r, barberIDParam)
	barber, err := h.service.Get(r.Context(), principal.BarbershopID, barberID)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newBarberResponse(barber))
}

// CreateBarberHandler expone POST /private/barbers (CA-021-02), protegido
// por el protocolo de idempotencia reutilizable de HU-004 (RN-IDE-01,
// DEC-043).
type CreateBarberHandler struct {
	service *staff.Service
}

// NewCreateBarberHandler construye el handler de alta.
func NewCreateBarberHandler(service *staff.Service) *CreateBarberHandler {
	return &CreateBarberHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *CreateBarberHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	//    calcularse sobre los bytes exactos recibidos, nunca sobre una
	//    re-serialización del JSON ya decodificado.
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

	var req CreateBarberRequest
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

	result, err := h.service.Create(r.Context(), principal.BarbershopID, req.FullName, key, fingerprint)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	switch result.Decision.Outcome {
	case idempotency.OutcomeProceed, idempotency.OutcomeReplay:
		// El mismo result.Response, byte a byte, para la primera ejecución
		// y para cualquier repetición exacta (CA-004-01). Location se
		// reconstruye a partir del `id` del cuerpo ya persistido: funciona
		// igual para ambos desenlaces, sin necesitar un campo aparte.
		var parsed struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal([]byte(result.Response.Body), &parsed)
		w.Header().Set("Location", barberBasePath+parsed.ID)
		httpserver.WriteStoredResponse(w, result.Response)
	default:
		// Conflicto de contenido/operación (RN-IDE-01) o ejecución en curso
		// (DEC-043): Translate ya sabe convertirlo en 409.
		httpserver.WriteProblem(w, httpserver.Translate(result.Decision.AsError(), requestID))
	}
}

// RenameBarberHandler expone PATCH /private/barbers/{barberId} (CA-021-04).
type RenameBarberHandler struct {
	service *staff.Service
}

// NewRenameBarberHandler construye el handler de renombrado.
func NewRenameBarberHandler(service *staff.Service) *RenameBarberHandler {
	return &RenameBarberHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *RenameBarberHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}

	barberID := httpserver.URLParam(r, barberIDParam)

	var req UpdateBarberRequest
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

	barber, err := h.service.Rename(r.Context(), principal.BarbershopID, barberID, req.FullName)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newBarberResponse(barber))
}

func newBarberResponse(b staff.Barber) BarberResponse {
	return BarberResponse{
		ID:        b.ID,
		FullName:  b.FullName,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
}

func newBarberListResponse(result staff.ListResult) BarberListResponse {
	items := make([]BarberResponse, 0, len(result.Items))
	for _, b := range result.Items {
		items = append(items, newBarberResponse(b))
	}
	var next *string
	if result.NextCursor != "" {
		v := result.NextCursor
		next = &v
	}
	return BarberListResponse{Items: items, NextCursor: next}
}
