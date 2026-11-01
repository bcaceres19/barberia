package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// resendAPIURL es el único endpoint que este adaptador usa.
const resendAPIURL = "https://api.resend.com/emails"

// resendSendTimeout es el límite por intento, sin reintento síncrono
// (DEC-066), igual criterio que [metaSendTimeout].
const resendSendTimeout = 5 * time.Second

// ResendEmailConfig agrupa los valores de despliegue del adaptador de
// Resend (DEC-066). APIKey es secreto: se lee de variable de entorno,
// nunca se comitea ni se registra. Un despliegue que ya use AWS puede
// sustituir este adaptador por uno de SES equivalente sin requerir una
// nueva decisión (DEC-066 lo autoriza explícitamente); esta implementación
// cubre el caso por defecto.
type ResendEmailConfig struct {
	APIKey      string
	FromAddress string
	Subject     string
}

// ResendEmailSender implementa [EmailSender] contra la API REST de Resend.
type ResendEmailSender struct {
	httpClient *http.Client
	cfg        ResendEmailConfig
}

// NewResendEmailSender construye el adaptador. httpClient permite inyectar
// un *http.Client de prueba; nil usa http.DefaultClient.
func NewResendEmailSender(cfg ResendEmailConfig, httpClient *http.Client) ResendEmailSender {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return ResendEmailSender{httpClient: httpClient, cfg: cfg}
}

var _ EmailSender = ResendEmailSender{}

type resendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	// HTML y Text viajan juntos (issue #465): el cliente de correo elige; Text
	// es además la alternativa para quien no renderiza HTML.
	HTML string `json:"html,omitempty"`
	Text string `json:"text"`
}

// Send implementa [EmailSender].
func (s ResendEmailSender) Send(ctx context.Context, email, code string) error {
	return postResendEmail(ctx, s.httpClient, s.cfg, email, s.cfg.Subject, emailContent{
		Preheader:  "Tu código para recuperar el acceso a NAVA.",
		Title:      "Recupera tu acceso",
		Paragraphs: []string{"Usa este código para recuperar el acceso a tu cuenta. Vence en pocos minutos."},
		Code:       code,
		Footnotes:  []string{"Si no lo solicitaste, ignora este mensaje: tu cuenta sigue segura."},
	})
}

// postResendEmail envía por Resend un correo con la plantilla de marca (HTML y
// texto). Comparte el manejo de
// errores de los correos de código (recuperación y reto de acceso): sin
// reintento síncrono y sin copiar el cuerpo de error del proveedor, que puede
// incluir el destinatario (RN-DAT-02).
func postResendEmail(ctx context.Context, httpClient *http.Client, cfg ResendEmailConfig, to, subject string, content emailContent) error {
	htmlBody, err := content.HTML()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, resendSendTimeout)
	defer cancel()

	payload := resendEmailRequest{
		From:    cfg.FromAddress,
		To:      []string{to},
		Subject: subject,
		HTML:    htmlBody,
		Text:    content.Text(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("notification: codificar correo de Resend: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendAPIURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notification: construir solicitud a Resend: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("notification: enviar solicitud a Resend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("%w: resend respondió %d", errDeliveryFailed, resp.StatusCode)
	}
	return nil
}
