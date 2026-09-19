package notification

import (
	"context"
	"errors"
	"fmt"
)

// DualChannelRecoverySender compone un WhatsAppSender y un EmailSender y
// entrega por el que la persona eligió (DEC-092); satisface el puerto pequeño que auth.RecoveryService consume
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

// SendCode entrega el código únicamente por el canal cuyo destino llega
// lleno (DEC-092): phone para WhatsApp, email para correo. El destino vacío
// significa "canal no elegido" y NUNCA se sustituye por el otro como
// respaldo. El error devuelto es solo para diagnóstico interno (logging sin
// destinatario ni código); nunca decide la respuesta HTTP, que ya se fijó
// antes de llamar aquí (no enumeración, DEC-065).
func (s DualChannelRecoverySender) SendCode(ctx context.Context, phone, email, code string) error {
	switch {
	case phone != "" && email != "":
		// Ningún flujo de recuperación pide hoy ambos, pero el puerto lo
		// admite: se conserva la tolerancia a fallo parcial de DEC-066, un
		// canal no impide el intento por el otro.
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
	case phone != "":
		if err := s.whatsapp.Send(ctx, phone, code); err != nil {
			return fmt.Errorf("notification: canal whatsapp falló: %w", err)
		}
		return nil
	case email != "":
		if err := s.email.Send(ctx, email, code); err != nil {
			return fmt.Errorf("notification: canal correo falló: %w", err)
		}
		return nil
	default:
		return errors.New("notification: código de recuperación sin destino de canal")
	}
}
