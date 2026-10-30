package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

// LoggingChallengeEmailSender implementa [ChallengeEmailSender] como marcador
// para desarrollo sin Resend: deja constancia de que un envío habría ocurrido,
// sin correo ni código (RN-DAT-02), y nunca falla. cmd/api lo selecciona solo
// cuando Resend no está configurado por completo.
type LoggingChallengeEmailSender struct {
	logger *slog.Logger
}

// NewLoggingChallengeEmailSender construye el remitente placeholder.
func NewLoggingChallengeEmailSender(logger *slog.Logger) LoggingChallengeEmailSender {
	return LoggingChallengeEmailSender{logger: logger}
}

var _ ChallengeEmailSender = LoggingChallengeEmailSender{}

// SendCode implementa [ChallengeEmailSender]. Nunca registra email ni code.
func (s LoggingChallengeEmailSender) SendCode(ctx context.Context, _, _ string) error {
	s.logger.InfoContext(ctx, "auth: envío de código del reto por correo simulado (Resend no está configurado)")
	return nil
}

// CapturingChallengeEmailSender existe solo para flujos E2E locales/test: tras
// una entrega correcta escribe el código en claro en un archivo, igual que
// [CapturingWhatsAppOTPProvider] para WhatsApp. cmd/api lo activa únicamente
// en local/test con APP_PHONE_CHALLENGE_CAPTURE_FILE.
type CapturingChallengeEmailSender struct {
	inner ChallengeEmailSender
	path  string
}

// NewCapturingChallengeEmailSender envuelve inner.
func NewCapturingChallengeEmailSender(inner ChallengeEmailSender, path string) CapturingChallengeEmailSender {
	return CapturingChallengeEmailSender{inner: inner, path: path}
}

var _ ChallengeEmailSender = CapturingChallengeEmailSender{}

// SendCode implementa [ChallengeEmailSender].
func (s CapturingChallengeEmailSender) SendCode(ctx context.Context, email, code string) error {
	if err := s.inner.SendCode(ctx, email, code); err != nil {
		return err
	}
	data, err := json.Marshal(struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}{Email: email, Code: code})
	if err != nil {
		return fmt.Errorf("auth: serializar captura OTP: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("auth: escribir captura OTP: %w", err)
	}
	return nil
}
