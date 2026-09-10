package notification

import "context"

// PhoneChallengeSender adapta el puerto genérico de WhatsApp al caso de uso
// de reto OTP. Mantiene auth independiente del proveedor y reutiliza el
// armado de plantilla de Meta que ya implementa WhatsAppSender.
type PhoneChallengeSender struct {
	sender WhatsAppSender
}

// NewPhoneChallengeSender construye el adaptador de entrega para HU-007.
func NewPhoneChallengeSender(sender WhatsAppSender) PhoneChallengeSender {
	return PhoneChallengeSender{sender: sender}
}

// SendCode delega una única vez la entrega, sin añadir destinatario ni código
// al error que devuelve el proveedor.
func (s PhoneChallengeSender) SendCode(ctx context.Context, phone, code string) error {
	return s.sender.Send(ctx, phone, code)
}
