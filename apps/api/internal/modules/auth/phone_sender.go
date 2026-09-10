package auth

import (
	"context"
	"log/slog"
)

// LoggingPhoneCodeSender implementa [PhoneCodeSender] como marcador para
// desarrollo sin configuración completa de Meta. Solo deja constancia de que
// un envío habría ocurrido, sin phone ni code (RN-DAT-02), y nunca falla:
// permite que las pruebas locales y E2E no dependan de un tercero. cmd/api
// lo selecciona exclusivamente cuando Meta está totalmente ausente en
// local/test; en cualquier configuración completa se usa el adaptador oficial.
type LoggingPhoneCodeSender struct {
	logger *slog.Logger
}

// NewLoggingPhoneCodeSender construye el remitente placeholder.
func NewLoggingPhoneCodeSender(logger *slog.Logger) LoggingPhoneCodeSender {
	return LoggingPhoneCodeSender{logger: logger}
}

var _ PhoneCodeSender = LoggingPhoneCodeSender{}

// SendCode implementa [PhoneCodeSender]. Nunca registra phone ni code.
func (s LoggingPhoneCodeSender) SendCode(ctx context.Context, phone, code string) error {
	s.logger.InfoContext(ctx, "auth: envío de código de reto telefónico simulado (adaptador real pendiente de HU-008/DEC-066)")
	return nil
}
