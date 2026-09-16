package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// confirmationEmailSubject es el asunto fijo del correo de confirmación de
// HU-097 (F-PUB-07, DEC-091): sin configuración por variable de entorno -a
// diferencia del asunto de recuperación (ResendEmailConfig.Subject)-,
// porque ninguna decisión pide que sea personalizable todavía.
const confirmationEmailSubject = "Tu turno fue confirmado"

// ResendConfirmationEmailConfig agrupa los valores de despliegue del correo
// de confirmación (DEC-091): reutiliza las mismas credenciales de Resend ya
// exigidas para HU-008 (APIKey/FromAddress), sin asunto configurable.
type ResendConfirmationEmailConfig struct {
	APIKey      string
	FromAddress string
}

// ResendConfirmationEmailSender satisface, de forma puramente estructural,
// publicbooking.ConfirmationEmailPort (HU-097): reutiliza el mismo endpoint
// REST de Resend que ResendEmailSender, con un cuerpo de texto distinto
// (resumen de la cita + enlace de acceso, en vez de un código corto).
// publicbooking nunca importa notification (CA-002-06).
type ResendConfirmationEmailSender struct {
	httpClient *http.Client
	cfg        ResendConfirmationEmailConfig
}

// NewResendConfirmationEmailSender construye el adaptador. httpClient
// permite inyectar un *http.Client de prueba; nil usa http.DefaultClient.
func NewResendConfirmationEmailSender(cfg ResendConfirmationEmailConfig, httpClient *http.Client) ResendConfirmationEmailSender {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return ResendConfirmationEmailSender{httpClient: httpClient, cfg: cfg}
}

// SendConfirmation implementa publicbooking.ConfirmationEmailPort.
// accessLink puede llegar vacío (config.Config.PublicWebBaseURL sin fijar,
// DEC-091): el correo se envía igual, con una instrucción segura en vez de
// un enlace.
func (s ResendConfirmationEmailSender) SendConfirmation(
	ctx context.Context,
	email, barbershopName, serviceName, attendeeName string,
	startsAtLocalFormatted string,
	accessLink string,
) error {
	ctx, cancel := context.WithTimeout(ctx, resendSendTimeout)
	defer cancel()

	linkLine := "Muy pronto podrás consultar tu turno en línea con el enlace que enviaremos por separado."
	if accessLink != "" {
		linkLine = fmt.Sprintf("Consulta tu turno cuando quieras en este enlace: %s", accessLink)
	}

	text := fmt.Sprintf(
		"Hola %s,\n\nTu turno en %s quedó confirmado.\n\nServicio: %s\nFecha y hora: %s\n\n%s\n\nSi no reservaste este turno, ignora este mensaje.",
		attendeeName, barbershopName, serviceName, startsAtLocalFormatted, linkLine,
	)

	payload := resendEmailRequest{
		From:    s.cfg.FromAddress,
		To:      []string{email},
		Subject: confirmationEmailSubject,
		Text:    text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("notification: codificar correo de confirmación de Resend: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendAPIURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notification: construir solicitud de confirmación a Resend: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("notification: enviar correo de confirmación a Resend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		// El cuerpo del error de Resend puede incluir el destinatario; se
		// descarta explícitamente (RN-DAT-02).
		_, _ = io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("%w: resend respondió %d", errDeliveryFailed, resp.StatusCode)
	}
	return nil
}
