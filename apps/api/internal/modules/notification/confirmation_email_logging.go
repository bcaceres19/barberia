package notification

import (
	"context"
	"log/slog"
)

// LoggingConfirmationEmailSender satisface, de forma puramente estructural,
// publicbooking.ConfirmationEmailPort como marcador de posición documentado
// para despliegues sin credenciales reales de Resend configuradas
// (típicamente local/test, mismo criterio que LoggingRecoveryCodeSender):
// cmd/api/main.go lo construye solo cuando APP_RESEND_API_KEY/
// APP_RESEND_FROM_ADDRESS faltan. Deja constancia en el log de que un envío
// HABRÍA ocurrido, sin destinatario, código ni enlace (RN-DAT-02), y nunca
// falla.
type LoggingConfirmationEmailSender struct {
	logger *slog.Logger
}

// NewLoggingConfirmationEmailSender construye el remitente placeholder.
func NewLoggingConfirmationEmailSender(logger *slog.Logger) LoggingConfirmationEmailSender {
	return LoggingConfirmationEmailSender{logger: logger}
}

// SendConfirmation implementa publicbooking.ConfirmationEmailPort. Nunca
// registra email, barbershopName, serviceName, attendeeName ni accessLink.
func (s LoggingConfirmationEmailSender) SendConfirmation(
	ctx context.Context,
	_, _, _, _ string,
	_ string,
	_ string,
) error {
	s.logger.InfoContext(ctx, "publicbooking: envío de correo de confirmación simulado (sin credenciales de Resend configuradas)")
	return nil
}
