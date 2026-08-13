package httpapi

import (
	"net/http"
	"regexp"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/platform/httpserver"
)

// maxSessionCookieLen acota el valor de la cookie de sesión ANTES de
// calcular su hash o tocar PostgreSQL (CA-006-05). CryptoTokenGenerator
// produce base64.RawURLEncoding de 32 bytes (43 caracteres); se acepta un
// margen amplio para no acoplar el middleware a ese tamaño exacto si el
// generador cambia algún día, sin dejar de rechazar de inmediato un valor
// evidentemente hostil (varios KB) antes de cualquier trabajo criptográfico
// o de base de datos.
const maxSessionCookieLen = 512

// sessionCookiePattern es el alfabeto que produce base64.RawURLEncoding:
// letras, dígitos, '-' y '_', sin relleno. Un valor fuera de este alfabeto
// nunca puede coincidir con un hash almacenado; rechazarlo aquí evita el
// costo de SHA-256 y la consulta a PostgreSQL para una cookie con forma
// evidentemente inválida.
var sessionCookiePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// SessionMiddleware implementa el paso 8 del orden de middleware
// (docs/04-arquitectura/backend-go.md sección 6): autenticación para rutas
// privadas. Se monta UNA sola vez sobre el subrouter de /api/v1/private
// (DEC-055 ya dejó ese subrouter libre del login, CT-003 resuelta):
// ninguna ruta de ese subrouter se registra sin pasar por aquí, sin
// excepción alguna (CA-006-04).
type SessionMiddleware struct {
	service *auth.SessionService
	cookie  CookieConfig
}

// NewSessionMiddleware construye el middleware de sesión.
func NewSessionMiddleware(service *auth.SessionService, cookie CookieConfig) *SessionMiddleware {
	return &SessionMiddleware{service: service, cookie: cookie}
}

// RequireSession extrae la cookie de sesión, la valida (con renovación
// deslizante incluida) y propaga [auth.Principal] en el contexto de la
// solicitud, o responde 401 uniforme sin ejecutar next en absoluto. La
// forma de la cookie se comprueba ANTES de invocar el servicio: ausencia,
// forma inválida y tamaño excesivo nunca llegan a SHA-256 ni a PostgreSQL
// (CA-006-05), y producen exactamente el mismo problema que un token
// desconocido, vencido, revocado o de un usuario inactivo (CA-006-02):
// ninguna rama de este middleware distingue esos casos en la respuesta.
func (m *SessionMiddleware) RequireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := httpserver.RequestIDFromContext(r.Context())

		c, err := r.Cookie(m.cookie.Name)
		if err != nil || !validSessionCookieShape(c.Value) {
			httpserver.WriteProblem(w, httpserver.Translate(auth.ErrInvalidSession(), requestID))
			return
		}

		principal, err := m.service.Validate(r.Context(), c.Value)
		if err != nil {
			httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
			return
		}

		ctx := auth.ContextWithPrincipal(r.Context(), principal)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validSessionCookieShape comprueba forma y tamaño únicamente; nunca
// interpreta el valor como un token real de una cuenta existente (eso lo
// decide PostgreSQL dentro de [auth.SessionService.Validate]).
func validSessionCookieShape(v string) bool {
	return v != "" && len(v) <= maxSessionCookieLen && sessionCookiePattern.MatchString(v)
}
