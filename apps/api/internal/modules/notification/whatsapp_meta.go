package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// metaGraphBaseURL es el host fijo de la API de Meta; solo la versión es
// configurable, para poder actualizarla sin recompilar cuando Meta retire
// una versión (metaAPIVersion en MetaWhatsAppConfig).
const metaGraphBaseURL = "https://graph.facebook.com"

// metaSendTimeout es el límite por intento, sin reintento síncrono dentro
// de la solicitud HTTP del cliente (DEC-066): un fallo de Meta nunca debe
// mantener abierta la respuesta al barbero más allá de este plazo.
const metaSendTimeout = 5 * time.Second

// MetaWhatsAppConfig agrupa los valores de despliegue del adaptador de Meta
// WhatsApp Cloud API (DEC-066). PhoneNumberID y AccessToken son secretos:
// se leen de variables de entorno, nunca se comitean ni se registran.
type MetaWhatsAppConfig struct {
	APIVersion    string
	PhoneNumberID string
	AccessToken   string
	TemplateName  string
	LanguageCode  string
}

// MetaWhatsAppSender implementa [WhatsAppSender] contra Meta WhatsApp Cloud
// API en modo directo (DEC-066), usando una plantilla de categoría
// "Authentication" pre-aprobada: es el único tipo de plantilla que Meta
// permite enviar fuera de la ventana de conversación de 24 horas sin
// recargo de "marketing", y su formato está pensado exactamente para OTP.
//
// El nombre/idioma exactos de la plantilla y la forma precisa de sus
// componentes (cuerpo + botón de "copiar código") dependen de la plantilla
// que el propietario apruebe en Meta Business Manager para el despliegue
// real; esta implementación sigue el formato documentado de Meta Cloud API
// v19+ con un componente de cuerpo (el código como único parámetro) y, si
// el nombre de plantilla sigue la convención de plantillas de
// autenticación con botón de copia, un componente de botón "otp" en el
// índice 0. Verificar contra la plantilla real antes de activarla en
// producción.
type MetaWhatsAppSender struct {
	httpClient *http.Client
	cfg        MetaWhatsAppConfig
}

// NewMetaWhatsAppSender construye el adaptador. httpClient permite inyectar
// un *http.Client de prueba (con Transport interceptado); nil usa
// http.DefaultClient.
func NewMetaWhatsAppSender(cfg MetaWhatsAppConfig, httpClient *http.Client) MetaWhatsAppSender {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return MetaWhatsAppSender{httpClient: httpClient, cfg: cfg}
}

var _ WhatsAppSender = MetaWhatsAppSender{}

type metaTemplateMessage struct {
	MessagingProduct string       `json:"messaging_product"`
	To               string       `json:"to"`
	Type             string       `json:"type"`
	Template         metaTemplate `json:"template"`
}

type metaTemplate struct {
	Name       string          `json:"name"`
	Language   metaLanguage    `json:"language"`
	Components []metaComponent `json:"components"`
}

type metaLanguage struct {
	Code string `json:"code"`
}

type metaComponent struct {
	Type       string          `json:"type"`
	SubType    string          `json:"sub_type,omitempty"`
	Index      string          `json:"index,omitempty"`
	Parameters []metaParameter `json:"parameters"`
}

type metaParameter struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Send implementa [WhatsAppSender]. phoneE164 llega con '+' inicial
// (formato staff_user.phone); la API de Meta espera el número sin ese
// prefijo.
func (s MetaWhatsAppSender) Send(ctx context.Context, phoneE164, code string) error {
	ctx, cancel := context.WithTimeout(ctx, metaSendTimeout)
	defer cancel()

	to := strings.TrimPrefix(phoneE164, "+")

	payload := metaTemplateMessage{
		MessagingProduct: "whatsapp",
		To:               to,
		Type:             "template",
		Template: metaTemplate{
			Name:     s.cfg.TemplateName,
			Language: metaLanguage{Code: s.cfg.LanguageCode},
			Components: []metaComponent{
				{
					Type:       "body",
					Parameters: []metaParameter{{Type: "text", Text: code}},
				},
				{
					Type:       "button",
					SubType:    "otp",
					Index:      "0",
					Parameters: []metaParameter{{Type: "text", Text: code}},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("notification: codificar plantilla de WhatsApp: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s/messages", metaGraphBaseURL, s.cfg.APIVersion, s.cfg.PhoneNumberID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notification: construir solicitud a Meta: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.AccessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("notification: enviar solicitud a Meta: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		// El cuerpo del error de Meta puede incluir el número de destino;
		// se descarta explícitamente del mensaje de error (RN-DAT-02). Solo
		// el status code queda disponible para diagnóstico.
		_, _ = io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("%w: meta respondió %d", errDeliveryFailed, resp.StatusCode)
	}
	return nil
}

var errDeliveryFailed = errors.New("notification: fallo de entrega")
