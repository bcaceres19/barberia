package notification

import (
	"context"
	"errors"
	"time"
)

// WhatsAppSender entrega un código corto por WhatsApp a un teléfono en
// formato E.164. Puerto angosto: ningún llamador conoce el proveedor
// concreto detrás de la implementación.
type WhatsAppSender interface {
	Send(ctx context.Context, phoneE164, code string) error
}

// EmailSender entrega un código corto por correo. Puerto angosto, mismo
// criterio que [WhatsAppSender].
type EmailSender interface {
	Send(ctx context.Context, email, code string) error
}

// errDeliveryFailed marca cualquier fallo de entrega de un adaptador cuyo
// resultado no tiene una clasificación más específica.
var errDeliveryFailed = errors.New("notification: fallo de entrega")

// WhatsAppEventStore guarda lo que Meta notifica por webhook (DEC-126). El
// teléfono llega ya convertido en su HMAC: ningún adaptador ve el número en
// claro.
type WhatsAppEventStore interface {
	// RecordInbound registra un mensaje entrante, que abre la ventana de 24 h.
	RecordInbound(ctx context.Context, phoneHash string, at time.Time) error
	// RecordStatus registra un estado de entrega y devuelve false si ya estaba
	// registrado (Meta repite notificaciones).
	RecordStatus(ctx context.Context, wamid, status string, at time.Time, errorCode *int) (bool, error)
	// ConversationOpen informa si hay un mensaje entrante en las últimas 24 h.
	ConversationOpen(ctx context.Context, phoneHash string) (bool, error)
}
