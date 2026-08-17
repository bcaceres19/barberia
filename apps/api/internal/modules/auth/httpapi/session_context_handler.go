package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
)

// SessionContextHandler expone GET /api/v1/private/auth/session (HU-012,
// DEC-060): lectura no destructiva que rehidrata la cookie de sesión
// HttpOnly y devuelve el contexto mínimo de sesión. No revalida nada por su
// cuenta: SessionMiddleware ya validó y renovó la sesión para esta
// solicitud antes de que este handler se ejecute, igual que LogoutHandler.
// A diferencia de logout, este handler nunca escribe Set-Cookie: es una
// consulta, no una mutación.
type SessionContextHandler struct {
	service *auth.SessionService
}

// NewSessionContextHandler construye el handler de contexto de sesión.
func NewSessionContextHandler(service *auth.SessionService) *SessionContextHandler {
	return &SessionContextHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *SessionContextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())

	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		// Mismo defecto de wiring que LogoutHandler contempla: solo puede
		// ocurrir si esta ruta se montó fuera del subrouter protegido por
		// SessionMiddleware (CA-006-04 lo evita estructuralmente).
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Internal(errors.New("auth: falta el principal de sesión en el contexto")), requestID))
		return
	}

	sessionContext, err := h.service.Context(r.Context(), principal)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(SessionContextResponse{
		Barbershop: BarbershopSummary{
			ID:   sessionContext.BarbershopID,
			Name: sessionContext.BarbershopName,
		},
		ExpiresAt: sessionContext.ExpiresAt,
	})
}
