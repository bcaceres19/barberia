package auth

import (
	"context"
	"log/slog"
)

// LoggingRecoveryCodeSender implementa [RecoveryCodeSender] como marcador
// de posición documentado para despliegues sin credenciales reales de
// Meta/Resend configuradas (típicamente local/test): cmd/api/main.go lo
// construye solo cuando la configuración de proveedor está incompleta,
// nunca en un despliegue con requiresHardening activo (esa combinación
// falla al arrancar, ver config.Load). Deja constancia en el log de que un
// envío HABRÍA ocurrido, sin phone, email ni code (RN-DAT-02), y nunca
// falla.
type LoggingRecoveryCodeSender struct {
	logger *slog.Logger
}

// NewLoggingRecoveryCodeSender construye el remitente placeholder.
func NewLoggingRecoveryCodeSender(logger *slog.Logger) LoggingRecoveryCodeSender {
	return LoggingRecoveryCodeSender{logger: logger}
}

var _ RecoveryCodeSender = LoggingRecoveryCodeSender{}

// SendCode implementa [RecoveryCodeSender]. Nunca registra phone, email ni code.
func (s LoggingRecoveryCodeSender) SendCode(ctx context.Context, phone, email, code string) error {
	s.logger.InfoContext(ctx, "auth: envío de código de recuperación simulado (sin credenciales de proveedor configuradas)")
	return nil
}
