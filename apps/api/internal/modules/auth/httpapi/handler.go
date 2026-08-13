// Package httpapi es el adaptador HTTP del módulo auth: decodifica,
// valida la forma, invoca auth.LoginService y traduce el resultado a HTTP
// (docs/03-desarrollo/estandar-backend-go.md §4). No importa Chi: la
// composición de rutas es responsabilidad exclusiva de cmd/api sobre el
// *chi.Mux que internal/platform/httpserver ya construye
// (internal/platform/archtest.TestDomainAndServicesDoNotImportChi solo
// autoriza a internal/platform/httpserver a importar Chi), igual que
// router.Get("/health", ...) ya hace en cmd/api/main.go.
package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
)

// Límites de forma, comprobados ANTES de invocar el caso de uso: coinciden
// con staff_user_email_ck (254) y con un margen razonable para una
// contraseña (256), sin acoplar el handler a la representación de
// PostgreSQL.
const (
	maxEmailLen    = 254
	maxPasswordLen = 256
)

// CookieName, CookiePath, CookieSameSite y CookieSecure fijan los atributos
// exactos de la cookie de sesión (DEC-050). SameSite, Path, Domain y el
// nombre NO están fijados por ninguna fuente normativa ni por el issue #44:
// ver DP-SEG-07 en docs/00-control/dudas-pendientes.md, que registra este
// vacío y documenta por qué se eligió Lax/host-only/estos valores como
// elección PROVISIONAL, pendiente de confirmación explícita del
// propietario.
const (
	CookieName = "barberia_session"
	CookiePath = "/api/v1"
)

// CookieConfig agrupa los atributos configurables de la cookie de sesión,
// para que las pruebas HTTP puedan variar Secure sin depender de TLS real.
type CookieConfig struct {
	Name     string
	Path     string
	SameSite http.SameSite
	Secure   bool
}

// DefaultCookieConfig devuelve la configuración provisional de DP-SEG-07:
// SameSite=Lax (cookie de sesión de primer nivel, no necesita enviarse en
// navegación cross-site), Path=/api/v1 (cubre login público y rutas
// privadas futuras sin exponerla a otras rutas del host), Secure=true
// (DEC-050, no negociable), sin Domain explícito (host-only, el navegador
// no la envía a subdominios no revisados).
func DefaultCookieConfig() CookieConfig {
	return CookieConfig{
		Name:     CookieName,
		Path:     CookiePath,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	}
}

// LoginHandler decodifica, valida la forma, invoca el caso de uso de login y
// traduce el resultado a HTTP. No contiene reglas de negocio ni SQL
// (docs/03-desarrollo/estandar-backend-go.md §4).
type LoginHandler struct {
	service *auth.LoginService
	cookie  CookieConfig
}

// NewLoginHandler construye el handler de login.
func NewLoginHandler(service *auth.LoginService, cookie CookieConfig) *LoginHandler {
	return &LoginHandler{service: service, cookie: cookie}
}

// ServeHTTP implementa http.Handler.
func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())

	var req LoginRequest
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
		// tratamiento que un JSON malformado.
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Invalid("cuerpo JSON inválido o con un campo desconocido"), requestID))
		return
	}

	if problem := validateLoginRequest(req); problem != "" {
		httpserver.WriteProblem(w, httpserver.Translate(apperr.Validation(problem), requestID))
		return
	}

	session, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.cookie.Name,
		Value:    session.Token,
		Path:     h.cookie.Path,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: h.cookie.SameSite,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(LoginResponse{ExpiresAt: session.ExpiresAt})
}

// validateLoginRequest comprueba la forma del request (campos presentes y
// dentro de un largo razonable) ANTES de invocar el caso de uso. Devuelve
// una cadena vacía cuando la forma es válida. No valida que el correo
// "parezca" un correo real: cualquier valor con forma razonable llega a
// LoginService, que lo trata de manera uniforme sin importar si existe
// (CA-005-02) — una validación de formato más estricta aquí no aportaría
// seguridad, solo una superficie adicional que mantener sincronizada con
// staff_user_email_ck.
func validateLoginRequest(req LoginRequest) string {
	switch {
	case req.Email == "":
		return "email es obligatorio"
	case len(req.Email) > maxEmailLen:
		return "email excede el largo máximo"
	case req.Password == "":
		return "password es obligatorio"
	case len(req.Password) > maxPasswordLen:
		return "password excede el largo máximo"
	default:
		return ""
	}
}
