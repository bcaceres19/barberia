package notification

import (
	"context"
	"errors"
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
