// Package httpapi es el adaptador HTTP del módulo reviews (HU-100): decodifica,
// invoca reviews.Service y traduce el resultado a HTTP. No importa Chi; cmd/api
// compone las rutas.
package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/modules/reviews"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
)

const (
	tokenParam    = "token"
	slugParam     = "slug"
	barberIDParam = "barberId"
	reviewIDParam = "reviewId"
)

func decodeStrict(w http.ResponseWriter, r *http.Request, requestID string, into any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(into); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
			return false
		}
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Invalid("cuerpo JSON inválido o con un campo desconocido"), requestID))
		return false
	}
	if dec.More() {
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Invalid("cuerpo JSON inválido o con un campo desconocido"), requestID))
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// CreateReviewHandler expone POST /customer/appointments/{token}/review (HU-100).
// Sin sesión: la única credencial es el token del enlace. Cache-Control: no-store
// como el resto de la audiencia customer (RN-DAT-02).
type CreateReviewHandler struct{ service *reviews.Service }

// NewCreateReviewHandler construye el handler.
func NewCreateReviewHandler(service *reviews.Service) *CreateReviewHandler {
	return &CreateReviewHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *CreateReviewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	w.Header().Set("Cache-Control", "no-store")

	var req CreateReviewRequest
	if !decodeStrict(w, r, requestID, &req) {
		return
	}
	rating := 0 // fuera de 1..5: el servicio lo rechaza con 422 si el campo falta
	if req.Rating != nil {
		rating = *req.Rating
	}

	review, err := h.service.Create(r.Context(), httpserver.URLParam(r, tokenParam), rating, req.Comment)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	writeJSON(w, http.StatusCreated, CustomerReviewResponse{
		Rating: review.Rating, Comment: review.Comment, CreatedAt: review.CreatedAt,
	})
}

// ListPublicReviewsHandler expone GET /public/barbershops/{slug}/barbers/{barberId}/reviews.
type ListPublicReviewsHandler struct{ service *reviews.Service }

// NewListPublicReviewsHandler construye el handler.
func NewListPublicReviewsHandler(service *reviews.Service) *ListPublicReviewsHandler {
	return &ListPublicReviewsHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *ListPublicReviewsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	list, err := h.service.ListPublic(r.Context(), httpserver.URLParam(r, slugParam), httpserver.URLParam(r, barberIDParam))
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	items := make([]PublicReviewResponse, 0, len(list.Items))
	for _, it := range list.Items {
		items = append(items, PublicReviewResponse{Rating: it.Rating, Comment: it.Comment, CreatedAt: it.CreatedAt})
	}
	writeJSON(w, http.StatusOK, PublicReviewListResponse{
		Rating: list.Summary.Rating, RatingCount: list.Summary.RatingCount, Items: items,
	})
}

func principalOrInternalError(w http.ResponseWriter, r *http.Request, requestID string) (auth.Principal, bool) {
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Internal(errors.New("reviews: falta el principal de sesión en el contexto")), requestID))
		return auth.Principal{}, false
	}
	return principal, true
}

func newReviewResponse(r reviews.Review) ReviewResponse {
	return ReviewResponse{ID: r.ID, Rating: r.Rating, Comment: r.Comment, HiddenAt: r.HiddenAt, CreatedAt: r.CreatedAt}
}

// ListReviewsHandler expone GET /private/barbers/{barberId}/reviews (panel).
type ListReviewsHandler struct{ service *reviews.Service }

// NewListReviewsHandler construye el handler.
func NewListReviewsHandler(service *reviews.Service) *ListReviewsHandler {
	return &ListReviewsHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *ListReviewsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}
	list, err := h.service.ListForBarber(r.Context(), principal.BarbershopID, httpserver.URLParam(r, barberIDParam))
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	items := make([]ReviewResponse, 0, len(list.Items))
	for _, it := range list.Items {
		items = append(items, newReviewResponse(it))
	}
	writeJSON(w, http.StatusOK, ReviewListResponse{
		Rating: list.Summary.Rating, RatingCount: list.Summary.RatingCount, Items: items,
	})
}

// UpdateReviewHandler expone PATCH /private/barbers/{barberId}/reviews/{reviewId}:
// oculta o vuelve a mostrar una reseña. PATCH con el estado deseado es
// idempotente por naturaleza, sin Idempotency-Key.
type UpdateReviewHandler struct{ service *reviews.Service }

// NewUpdateReviewHandler construye el handler.
func NewUpdateReviewHandler(service *reviews.Service) *UpdateReviewHandler {
	return &UpdateReviewHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *UpdateReviewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := principalOrInternalError(w, r, requestID)
	if !ok {
		return
	}
	var req UpdateReviewRequest
	if !decodeStrict(w, r, requestID, &req) {
		return
	}
	if req.Hidden == nil {
		httpserver.WriteProblem(w, httpserver.Translate(apperr.Validation("hidden es obligatorio"), requestID))
		return
	}
	review, err := h.service.SetHidden(r.Context(), principal.BarbershopID,
		httpserver.URLParam(r, barberIDParam), httpserver.URLParam(r, reviewIDParam), *req.Hidden)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	writeJSON(w, http.StatusOK, newReviewResponse(review))
}
