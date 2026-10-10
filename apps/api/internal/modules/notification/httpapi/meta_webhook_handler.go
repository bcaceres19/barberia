// Package httpapi expone el webhook de Meta WhatsApp Cloud API (DEC-126). Es
// una ruta pública sin sesión: su única autenticación es el token de
// verificación (GET) y la firma HMAC del cuerpo (POST).
package httpapi

import (
	"crypto/sha256"
	"crypto/subtle"
	"io"
	"log/slog"
	"net/http"
	"regexp"

	"system-barbershop/internal/modules/notification"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
)

// signatureHeader es la cabecera con la que Meta firma el cuerpo exacto.
const signatureHeader = "X-Hub-Signature-256"

// challengePattern acota lo que se devuelve en la verificación: Meta envía un
// entero, y no se refleja ningún otro contenido del solicitante.
var challengePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

// MetaWebhookHandler atiende la verificación y las notificaciones de Meta.
type MetaWebhookHandler struct {
	service     *notification.MetaWebhookService
	verifyToken string
	appSecret   string
	logger      *slog.Logger
}

// NewMetaWebhookHandler construye el handler. verifyToken y appSecret son
// secretos distintos: el primero prueba la propiedad del endpoint durante la
// suscripción; el segundo autentica cada notificación.
func NewMetaWebhookHandler(service *notification.MetaWebhookService, verifyToken, appSecret string, logger *slog.Logger) *MetaWebhookHandler {
	return &MetaWebhookHandler{service: service, verifyToken: verifyToken, appSecret: appSecret, logger: logger}
}

// Verify atiende GET: responde el desafío solo si el modo y el token coinciden.
func (h *MetaWebhookHandler) Verify(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	query := r.URL.Query()
	challenge := query.Get("hub.challenge")

	if query.Get("hub.mode") != "subscribe" || !tokensEqual(query.Get("hub.verify_token"), h.verifyToken) {
		httpserver.WriteProblem(w, httpserver.Translate(apperr.Unauthorized("verificación rechazada"), requestID))
		return
	}
	if !challengePattern.MatchString(challenge) {
		httpserver.WriteProblem(w, httpserver.Translate(apperr.Invalid("hub.challenge inválido"), requestID))
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, challenge)
}

// Receive atiende POST. La firma se comprueba antes de interpretar el cuerpo;
// una firma ausente o inválida no toca la base de datos.
func (h *MetaWebhookHandler) Receive(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())

	body, err := io.ReadAll(r.Body)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	if !notification.VerifyMetaSignature(h.appSecret, body, r.Header.Get(signatureHeader)) {
		h.logger.WarnContext(r.Context(), "whatsapp webhook: firma inválida", "requestId", requestID)
		httpserver.WriteProblem(w, httpserver.Translate(apperr.Unauthorized("firma inválida"), requestID))
		return
	}

	payload, err := notification.ParseMetaWebhook(body)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(apperr.Invalid("cuerpo JSON inválido"), requestID))
		return
	}
	if err := h.service.Process(r.Context(), payload); err != nil {
		// Un 5xx hace que Meta reintente; es seguro porque cada registro es
		// idempotente. El detalle no se devuelve ni incluye datos personales.
		h.logger.ErrorContext(r.Context(), "whatsapp webhook: no se pudo registrar la notificación", "requestId", requestID)
		httpserver.WriteProblem(w, httpserver.Translate(apperr.Internal(err), requestID))
		return
	}
	w.WriteHeader(http.StatusOK)
}

// tokensEqual compara en tiempo constante sin filtrar la longitud del secreto.
func tokensEqual(got, want string) bool {
	a, b := sha256.Sum256([]byte(got)), sha256.Sum256([]byte(want))
	return subtle.ConstantTimeCompare(a[:], b[:]) == 1
}
