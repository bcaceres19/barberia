package notification

import "context"

// EmailOnlyRecoverySender compone únicamente un EmailSender y satisface el
// mismo puerto pequeño que [DualChannelRecoverySender]
// (SendCode(ctx, phone, email, code) error), sin que auth conozca Resend
// (DEC-066). Ignora phone deliberadamente: existe EXCLUSIVAMENTE para la
// excepción de correo único en APP_ENVIRONMENT=local/test descrita en
// docs/10-backlog/prompts/test/issue-86-otp-correo-resend.md, cuando no hay
// credenciales de Meta WhatsApp configuradas. cmd/api/main.go es el único
// lugar que decide construirla, y solo en local/test; en pilot/production
// esa decisión sigue exigiendo el remitente dual (DEC-051/DEC-066).
type EmailOnlyRecoverySender struct {
	email EmailSender
}

// NewEmailOnlyRecoverySender construye el remitente de solo correo.
func NewEmailOnlyRecoverySender(email EmailSender) EmailOnlyRecoverySender {
	return EmailOnlyRecoverySender{email: email}
}

// SendCode implementa el puerto de auth ignorando phone.
func (s EmailOnlyRecoverySender) SendCode(ctx context.Context, phone, email, code string) error {
	return s.email.Send(ctx, email, code)
}
