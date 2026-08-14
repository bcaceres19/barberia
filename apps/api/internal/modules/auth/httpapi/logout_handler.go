package httpapi

import (
	"errors"
	"net/http"
	"time"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
)

// LogoutHandler decodifica nada (sin cuerpo), invoca
// auth.SessionService.Logout con el auth.Principal que
// [SessionMiddleware.RequireSession] ya dejó en el contexto, y limpia la
// cookie de sesión con los MISMOS atributos con que HU-005 la emitió
// (Path, SameSite, Secure). No contiene reglas de negocio ni SQL
// (docs/03-desarrollo/estandar-backend-go.md §4).
type LogoutHandler struct {
	service *auth.SessionService
	cookie  CookieConfig
}

// NewLogoutHandler construye el handler de cierre de sesión.
func NewLogoutHandler(service *auth.SessionService, cookie CookieConfig) *LogoutHandler {
	return &LogoutHandler{service: service, cookie: cookie}
}

// ServeHTTP implementa http.Handler.
func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())

	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		// Solo puede ocurrir si esta ruta se montó fuera del subrouter
		// protegido por SessionMiddleware: un defecto de wiring, no una
		// entrada de cliente. CA-006-04 lo evita estructuralmente; aquí se
		// trata como error interno en vez de asumir una forma de bypass
		// silenciosa.
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Internal(errors.New("auth: falta el principal de sesión en el contexto")), requestID))
		return
	}

	if err := h.service.Logout(r.Context(), principal); err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.cookie.Name,
		Value:    "",
		Path:     h.cookie.Path,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: h.cookie.SameSite,
	})

	w.WriteHeader(http.StatusNoContent)
}
