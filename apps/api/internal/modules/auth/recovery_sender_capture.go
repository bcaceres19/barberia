package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// CapturingRecoveryCodeSender envuelve otro [RecoveryCodeSender] y ADEMÁS
// escribe `{"phone": "...", "email": "...", "code": "..."}` en un archivo.
// Existe EXCLUSIVAMENTE para que una prueba de sistema pueda completar el
// recorrido de verificación real sin depender de credenciales reales de
// Meta/Resend, mismo mecanismo que CapturingPhoneCodeSender de HU-007.
// cmd/api/main.go SOLO la construye cuando APP_RECOVERY_CAPTURE_FILE está
// presente Y cfg.Environment es "local" o "test" (verificación redundante
// en ambos archivos a propósito, igual que HU-007).
type CapturingRecoveryCodeSender struct {
	inner RecoveryCodeSender
	path  string
}

// NewCapturingRecoveryCodeSender construye el remitente de captura.
func NewCapturingRecoveryCodeSender(inner RecoveryCodeSender, path string) CapturingRecoveryCodeSender {
	return CapturingRecoveryCodeSender{inner: inner, path: path}
}

var _ RecoveryCodeSender = CapturingRecoveryCodeSender{}

// SendCode implementa [RecoveryCodeSender]: delega primero al remitente
// interno y solo después escribe la captura, para que un fallo de entrega
// real siga siendo un fallo real y no quede enmascarado por la captura de
// prueba.
func (s CapturingRecoveryCodeSender) SendCode(ctx context.Context, phone, email, code string) error {
	if err := s.inner.SendCode(ctx, phone, email, code); err != nil {
		return err
	}

	data, err := json.Marshal(struct {
		Phone string `json:"phone"`
		Email string `json:"email"`
		Code  string `json:"code"`
	}{Phone: phone, Email: email, Code: code})
	if err != nil {
		return fmt.Errorf("auth: serializar captura de prueba de recuperación: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("auth: escribir captura de prueba de recuperación: %w", err)
	}
	return nil
}
