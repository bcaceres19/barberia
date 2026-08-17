package auth

import (
	"context"
	"log/slog"
)

// LoggingPhoneCodeSender implementa [PhoneCodeSender] como marcador de
// posición documentado, NO como un adaptador real de WhatsApp: el proveedor
// concreto (Meta WhatsApp Cloud API, DEC-066) es una decisión de HU-008, y
// HU-007 no puede adelantarla sin ampliar su propio alcance ("Elegir el
// proveedor/canal... sin decisión expresa" queda fuera de HU-007). Esta
// implementación deja constancia en el log de que un envío HABRÍA ocurrido,
// sin phone ni code (RN-DAT-02), y nunca falla: permite demostrar
// CA-007-02 de punta a punta contra HTTP/PostgreSQL reales sin depender de
// una cuenta de WhatsApp real. cmd/api/main.go debe reemplazarla por el
// adaptador real cuando HU-008 lo defina.
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
