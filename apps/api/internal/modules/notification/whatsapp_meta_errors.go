package notification

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Errores de dominio de la entrega por Meta. Los llamadores los distinguen
// con errors.Is; ninguno incluye el destinatario, el código ni el token.
var (
	// ErrInvalidRecipient: el teléfono no está en formato E.164.
	ErrInvalidRecipient = errors.New("notification: destinatario de WhatsApp inválido")
	// ErrRecipientNotAuthorized: el modo de desarrollo solo envía a la lista
	// APP_META_WHATSAPP_TEST_RECIPIENTS.
	ErrRecipientNotAuthorized = errors.New("notification: destinatario fuera de la lista de pruebas autorizada")
	// ErrConversationWindowClosed: Meta rechazó el texto libre porque el
	// destinatario no escribió en las últimas 24 horas (código 131047).
	ErrConversationWindowClosed = errors.New("notification: ventana de atención de 24 horas cerrada")
	// ErrRecipientNotAllowed: el número no está en la lista de destinatarios
	// de la cuenta de WhatsApp Business mientras esta no tenga acceso
	// completo (código 131030).
	ErrRecipientNotAllowed = errors.New("notification: destinatario no permitido por Meta")
	// ErrRecipientUnreachable: Meta no pudo entregar al número (código 131026).
	ErrRecipientUnreachable = errors.New("notification: destinatario no alcanzable por WhatsApp")
	// ErrMetaCredentials: token inválido, vencido o sin permisos.
	ErrMetaCredentials = errors.New("notification: credenciales o permisos de Meta rechazados")
	// ErrMetaRateLimited: Meta limitó la frecuencia o el volumen de envíos.
	ErrMetaRateLimited = errors.New("notification: límite de envíos de Meta alcanzado")
	// ErrTemplateUnavailable: la plantilla de autenticación no existe, no
	// está aprobada o sus parámetros no coinciden.
	ErrTemplateUnavailable = errors.New("notification: plantilla de WhatsApp no disponible")
)

// MetaAPIError es el error estructurado de Graph API. Conserva únicamente
// los identificadores de diagnóstico: el mensaje que devuelve Meta puede
// incluir el número de destino y por eso nunca se copia (RN-DAT-02).
type MetaAPIError struct {
	HTTPStatus int
	Code       int
	Subcode    int
	Type       string
	TraceID    string
	kind       error
}

func (e *MetaAPIError) Error() string {
	return fmt.Sprintf("%v: meta respondió %d (code %d, subcode %d, trace %s)",
		e.kind, e.HTTPStatus, e.Code, e.Subcode, e.TraceID)
}

func (e *MetaAPIError) Unwrap() error { return e.kind }

type metaErrorEnvelope struct {
	Error struct {
		Type      string `json:"type"`
		Code      int    `json:"code"`
		Subcode   int    `json:"error_subcode"`
		FBTraceID string `json:"fbtrace_id"`
	} `json:"error"`
}

// newMetaAPIError interpreta una respuesta no exitosa. Un cuerpo que no es
// JSON de Graph API se clasifica solo por el estado HTTP.
func newMetaAPIError(status int, body []byte) *MetaAPIError {
	var env metaErrorEnvelope
	_ = json.Unmarshal(body, &env)
	e := &MetaAPIError{
		HTTPStatus: status,
		Code:       env.Error.Code,
		Subcode:    env.Error.Subcode,
		Type:       env.Error.Type,
		TraceID:    env.Error.FBTraceID,
	}
	e.kind = classifyMetaError(status, e.Code)
	return e
}

func classifyMetaError(status, code int) error {
	switch code {
	case 131047:
		return ErrConversationWindowClosed
	case 131030:
		return ErrRecipientNotAllowed
	case 131026:
		return ErrRecipientUnreachable
	case 190:
		return ErrMetaCredentials
	case 4, 17, 80007, 130429, 131048, 131056:
		return ErrMetaRateLimited
	case 132000, 132001, 132012, 132015, 132016:
		return ErrTemplateUnavailable
	}
	switch {
	case status == http.StatusTooManyRequests:
		return ErrMetaRateLimited
	case status == http.StatusUnauthorized || status == http.StatusForbidden || code == 10 || (code >= 200 && code <= 299):
		return ErrMetaCredentials
	}
	return errDeliveryFailed
}
