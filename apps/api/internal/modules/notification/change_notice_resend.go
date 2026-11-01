package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Tipos de aviso que el adaptador sabe redactar. Son los mismos valores que
// appointmentnotice.KindCancelledByBarber y KindRescheduled: este paquete no
// importa appointmentnotice (se satisfacen estructuralmente, como el correo de
// confirmación con publicbooking).
const (
	changeNoticeCancelled   = "cancelled_by_barber"
	changeNoticeRescheduled = "rescheduled"
)

// ResendChangeNoticeConfig agrupa los valores de despliegue del aviso de cambio
// (HU-130): las mismas credenciales de Resend que el resto de correos.
type ResendChangeNoticeConfig struct {
	APIKey      string
	FromAddress string
}

// ResendChangeNoticeSender satisface appointmentnotice.Mailer: avisa al cliente
// de que el barbero canceló o movió su turno. El texto describe el estado
// vigente que recibe, nunca uno guardado al programar (RN-REC-03).
type ResendChangeNoticeSender struct {
	httpClient *http.Client
	cfg        ResendChangeNoticeConfig
}

// NewResendChangeNoticeSender construye el adaptador. httpClient permite inyectar
// un *http.Client de prueba; nil usa http.DefaultClient.
func NewResendChangeNoticeSender(cfg ResendChangeNoticeConfig, httpClient *http.Client) ResendChangeNoticeSender {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return ResendChangeNoticeSender{httpClient: httpClient, cfg: cfg}
}

// deliveryError es un fallo de entrega con el estado HTTP del proveedor. Un 4xx
// distinto de 408 y 429 es un rechazo permanente (el mensaje o la dirección no
// serán aceptados nunca): reintentarlo solo repite el rechazo. No incluye el
// destinatario (RN-DAT-02).
type deliveryError struct{ status int }

func (e deliveryError) Error() string {
	return fmt.Sprintf("notification: resend respondió %d", e.status)
}

func (e deliveryError) Unwrap() error { return errDeliveryFailed }

// Permanent indica que el proveedor rechazó el mensaje de forma definitiva.
func (e deliveryError) Permanent() bool {
	return e.status >= 400 && e.status < 500 && e.status != http.StatusRequestTimeout && e.status != http.StatusTooManyRequests
}

// SendChangeNotice implementa appointmentnotice.Mailer. startsAtLocal llega ya
// formateado en la zona de la barbería.
func (s ResendChangeNoticeSender) SendChangeNotice(
	ctx context.Context,
	kind, email, barbershopName, serviceName, attendeeName, startsAtLocal string,
) error {
	var subject string
	content := emailContent{
		Brand:    barbershopName,
		Greeting: fmt.Sprintf("Hola %s,", attendeeName),
		Details: []emailDetail{
			{Label: "Servicio", Value: serviceName},
			{Label: "Fecha y hora", Value: startsAtLocal},
		},
	}
	switch kind {
	case changeNoticeCancelled:
		subject = "Tu turno fue cancelado"
		content.Preheader = fmt.Sprintf("%s canceló tu turno.", barbershopName)
		content.Badge = &emailBadge{Label: "Turno cancelado", Tone: toneDanger}
		content.Title = "Tu turno fue cancelado"
		content.Paragraphs = []string{
			fmt.Sprintf("%s canceló tu turno. Estos son los datos del turno cancelado:", barbershopName),
			"Si quieres otro horario, comunícate con la barbería o reserva de nuevo.",
		}
	case changeNoticeRescheduled:
		subject = "Tu turno cambió de horario"
		content.Preheader = fmt.Sprintf("%s cambió el horario de tu turno.", barbershopName)
		content.Badge = &emailBadge{Label: "Horario actualizado", Tone: toneWarning}
		content.Title = "Tu turno cambió de horario"
		content.Paragraphs = []string{
			fmt.Sprintf("%s cambió el horario de tu turno. Este es el horario vigente:", barbershopName),
			"Si no puedes asistir, comunícate con la barbería.",
		}
	default:
		return errors.New("notification: tipo de aviso desconocido")
	}
	content.Footnotes = []string{"Si no esperabas este mensaje, ignóralo."}

	htmlBody, err := content.HTML()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, resendSendTimeout)
	defer cancel()

	body, err := json.Marshal(resendEmailRequest{From: s.cfg.FromAddress, To: []string{email}, Subject: subject, HTML: htmlBody, Text: content.Text()})
	if err != nil {
		return fmt.Errorf("notification: codificar aviso de cambio de Resend: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendAPIURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notification: construir solicitud del aviso de cambio a Resend: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		// El error de transporte puede citar la URL pero no el cuerpo; aun así no
		// se envuelve el original para no arrastrar nada del destinatario.
		return errors.New("notification: enviar aviso de cambio a Resend: fallo de transporte")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		// El cuerpo del error de Resend puede incluir el destinatario (RN-DAT-02).
		_, _ = io.Copy(io.Discard, resp.Body)
		return deliveryError{status: resp.StatusCode}
	}
	return nil
}
