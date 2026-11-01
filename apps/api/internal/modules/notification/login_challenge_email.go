package notification

import (
	"context"
	"net/http"
)

// loginChallengeEmailSubject es el asunto fijo del correo del reto de acceso
// (DEC-139). A diferencia del de recuperación no es configurable: no hay un
// segundo texto que desplegar por ambiente.
const loginChallengeEmailSubject = "Código de verificación de acceso"

// ResendLoginChallengeEmailSender entrega por Resend el código del reto
// adicional de acceso (HU-007, DEC-139). Satisface auth.ChallengeEmailSender
// sin que auth conozca el proveedor (DEC-066) y comparte credenciales y manejo
// de errores con el correo de recuperación.
type ResendLoginChallengeEmailSender struct {
	httpClient *http.Client
	cfg        ResendEmailConfig
}

// NewResendLoginChallengeEmailSender construye el adaptador; httpClient nil usa
// http.DefaultClient. Subject de cfg se ignora: el asunto es fijo.
func NewResendLoginChallengeEmailSender(cfg ResendEmailConfig, httpClient *http.Client) ResendLoginChallengeEmailSender {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return ResendLoginChallengeEmailSender{httpClient: httpClient, cfg: cfg}
}

// SendCode entrega el código al correo de la cuenta. El error devuelto no
// incluye el destinatario ni el código (RN-DAT-02).
func (s ResendLoginChallengeEmailSender) SendCode(ctx context.Context, email, code string) error {
	return postResendEmail(ctx, s.httpClient, s.cfg, email, loginChallengeEmailSubject, emailContent{
		Preheader:  "Tu código de verificación para entrar a NAVA.",
		Title:      "Verifica que eres tú",
		Paragraphs: []string{"Tu código de verificación para entrar es el siguiente. Vence en pocos minutos."},
		Code:       code,
		Footnotes:  []string{"Si no intentaste entrar, ignora este mensaje y considera cambiar tu contraseña."},
	})
}
