package notification

import (
	"context"
	"fmt"
)

// DualChannelRecoverySender compone un WhatsAppSender y un EmailSender y
// satisface el puerto pequeño que auth.RecoveryService consume
// (SendCode(ctx, phone, email, code) error), sin que auth conozca ninguno
// de los dos proveedores concretos (DEC-066).
type DualChannelRecoverySender struct {
	whatsapp WhatsAppSender
	email    EmailSender
}

// NewDualChannelRecoverySender construye el remitente compuesto.
func NewDualChannelRecoverySender(whatsapp WhatsAppSender, email EmailSender) DualChannelRecoverySender {
	return DualChannelRecoverySender{whatsapp: whatsapp, email: email}
}

// SendCode intenta ambos canales, siempre los dos, sin importar si el
// primero falla: un fallo de WhatsApp no debe impedir el intento por
// correo, ni viceversa (DEC-066, tolerancia a fallo parcial). El error
// devuelto es solo para diagnóstico interno (logging sin destinatario ni
// código); nunca decide la respuesta HTTP, que ya se fijó antes de llamar
// aquí (no enumeración, DEC-065).
func (s DualChannelRecoverySender) SendCode(ctx context.Context, phone, email, code string) error {
	whatsappErr := s.whatsapp.Send(ctx, phone, code)
	emailErr := s.email.Send(ctx, email, code)

	switch {
	case whatsappErr != nil && emailErr != nil:
		return fmt.Errorf("notification: ambos canales fallaron: whatsapp=%w email=%w", whatsappErr, emailErr)
	case whatsappErr != nil:
		return fmt.Errorf("notification: canal whatsapp falló: %w", whatsappErr)
	case emailErr != nil:
		return fmt.Errorf("notification: canal correo falló: %w", emailErr)
	}
	return nil
}
