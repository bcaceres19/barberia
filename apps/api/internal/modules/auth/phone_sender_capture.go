package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// CapturingPhoneCodeSender envuelve otro [PhoneCodeSender] y ADEMÁS escribe
// `{"phone": "...", "code": "..."}` en un archivo. Existe EXCLUSIVAMENTE
// para que la suite E2E de HU-007 pueda completar el recorrido de
// verificación real sin depender de un proveedor de WhatsApp real
// ("terceros se interceptan; no envíes mensajes reales", igual que la
// interceptación de red que ya usa e2e/acceso.spec.ts): sin este mecanismo
// no hay forma de conocer el código de 6 dígitos generado del lado del
// servidor. cmd/api/main.go SOLO la construye cuando
// APP_PHONE_CHALLENGE_CAPTURE_FILE está presente Y cfg.Environment es
// "local" o "test" (verificación redundante en ambos archivos a propósito:
// ver comentario en main.go).
type CapturingPhoneCodeSender struct {
	inner PhoneCodeSender
	path  string
}

// NewCapturingPhoneCodeSender construye el remitente de captura.
func NewCapturingPhoneCodeSender(inner PhoneCodeSender, path string) CapturingPhoneCodeSender {
	return CapturingPhoneCodeSender{inner: inner, path: path}
}

var _ PhoneCodeSender = CapturingPhoneCodeSender{}

// SendCode implementa [PhoneCodeSender]: delega primero al remitente
// interno y solo después escribe la captura, para que un fallo de entrega
// real siga siendo un fallo real y no quede enmascarado por la captura de
// prueba.
func (s CapturingPhoneCodeSender) SendCode(ctx context.Context, phone, code string) error {
	if err := s.inner.SendCode(ctx, phone, code); err != nil {
		return err
	}

	data, err := json.Marshal(struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}{Phone: phone, Code: code})
	if err != nil {
		return fmt.Errorf("auth: serializar captura de prueba del reto telefónico: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("auth: escribir captura de prueba del reto telefónico: %w", err)
	}
	return nil
}
