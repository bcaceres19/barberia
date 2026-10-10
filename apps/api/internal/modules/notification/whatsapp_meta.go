package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"
)

// metaGraphBaseURL es el host fijo de la API de Meta; solo la versión es
// configurable, para poder actualizarla sin recompilar cuando Meta retire
// una versión (APIVersion en MetaWhatsAppConfig).
const metaGraphBaseURL = "https://graph.facebook.com"

// metaSendTimeout es el límite por intento. No hay reintento automático
// (DEC-066): si la solicitud expira o la respuesta se pierde, Meta pudo haber
// aceptado el mensaje, y repetirlo duplicaría el código ante el usuario.
const metaSendTimeout = 5 * time.Second

// metaMaxResponseBytes acota la lectura de la respuesta de Graph API.
const metaMaxResponseBytes = 64 << 10

var e164Pattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

// MetaWhatsAppMode elige cómo se transporta el código que genera NAVA.
type MetaWhatsAppMode string

const (
	// MetaModeTemplate envía la plantilla de categoría AUTHENTICATION, la
	// única que Meta admite fuera de la ventana de atención de 24 horas.
	MetaModeTemplate MetaWhatsAppMode = "template"
	// MetaModeText envía texto libre (DEC-124): Meta solo lo entrega si el
	// destinatario escribió al número en las últimas 24 horas. Es la
	// alternativa mientras Meta no habilite la plantilla de autenticación.
	MetaModeText MetaWhatsAppMode = "text"
)

// MetaWhatsAppConfig agrupa los valores de despliegue del adaptador de Meta
// WhatsApp Cloud API (DEC-066, DEC-123). PhoneNumberID y AccessToken son
// secretos: se leen de variables de entorno, nunca se comitean ni se
// registran.
type MetaWhatsAppConfig struct {
	APIVersion    string
	PhoneNumberID string
	AccessToken   string
	Mode          MetaWhatsAppMode
	TemplateName  string
	LanguageCode  string
	// TestRecipients, si no está vacía, restringe el envío a esos teléfonos
	// E.164 en cualquier modo.
	TestRecipients []string
	// Logger registra el wamid aceptado; nunca recibe teléfono, código ni
	// token. Con nil no se registra nada.
	Logger *slog.Logger
}

// MetaSendResult es el acuse de que Graph API aceptó el mensaje. Aceptado no
// significa entregado: el estado de entrega llega por webhook, que NAVA aún
// no consume.
type MetaSendResult struct {
	MessageID string
}

// MetaWhatsAppSender implementa [WhatsAppSender] contra Meta WhatsApp Cloud
// API en modo directo. NAVA genera y valida el código; Meta solo lo
// transporta.
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
	if cfg.Mode == "" {
		cfg.Mode = MetaModeTemplate
	}
	return MetaWhatsAppSender{httpClient: httpClient, cfg: cfg}
}

var _ WhatsAppSender = MetaWhatsAppSender{}

type metaMessage struct {
	MessagingProduct string        `json:"messaging_product"`
	RecipientType    string        `json:"recipient_type"`
	To               string        `json:"to"`
	Type             string        `json:"type"`
	Text             *metaText     `json:"text,omitempty"`
	Template         *metaTemplate `json:"template,omitempty"`
}

type metaText struct {
	PreviewURL bool   `json:"preview_url"`
	Body       string `json:"body"`
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

type metaSendResponse struct {
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages"`
}

// Send implementa [WhatsAppSender]. phoneE164 llega con '+' inicial
// (formato staff_user.phone); la API de Meta espera el número sin ese
// prefijo.
func (s MetaWhatsAppSender) Send(ctx context.Context, phoneE164, code string) error {
	_, err := s.SendOTP(ctx, phoneE164, code)
	return err
}

// SendOTP envía el código recibido, sin alterarlo, y devuelve el wamid con el
// que Meta aceptó el mensaje.
func (s MetaWhatsAppSender) SendOTP(ctx context.Context, phoneE164, code string) (MetaSendResult, error) {
	if !e164Pattern.MatchString(phoneE164) {
		return MetaSendResult{}, ErrInvalidRecipient
	}
	if len(s.cfg.TestRecipients) > 0 && !slices.Contains(s.cfg.TestRecipients, phoneE164) {
		return MetaSendResult{}, ErrRecipientNotAuthorized
	}

	body, err := json.Marshal(s.buildMessage(strings.TrimPrefix(phoneE164, "+"), code))
	if err != nil {
		return MetaSendResult{}, fmt.Errorf("notification: codificar mensaje de WhatsApp: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, metaSendTimeout)
	defer cancel()

	url := fmt.Sprintf("%s/%s/%s/messages", metaGraphBaseURL, s.cfg.APIVersion, s.cfg.PhoneNumberID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return MetaSendResult{}, fmt.Errorf("notification: construir solicitud a Meta: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.AccessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		// El resultado es desconocido: Meta pudo procesar la solicitud. No se
		// reintenta y el error no arrastra el cuerpo ni las cabeceras.
		return MetaSendResult{}, fmt.Errorf("%w: sin respuesta de Meta", errDeliveryFailed)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, metaMaxResponseBytes))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return MetaSendResult{}, newMetaAPIError(resp.StatusCode, raw)
	}

	var parsed metaSendResponse
	if err := json.Unmarshal(raw, &parsed); err != nil || len(parsed.Messages) == 0 || parsed.Messages[0].ID == "" {
		return MetaSendResult{}, fmt.Errorf("%w: respuesta de Meta sin identificador de mensaje", errDeliveryFailed)
	}

	result := MetaSendResult{MessageID: parsed.Messages[0].ID}
	if s.cfg.Logger != nil {
		s.cfg.Logger.InfoContext(ctx, "whatsapp aceptado por Meta",
			slog.String("wamid", result.MessageID), slog.String("mode", string(s.cfg.Mode)))
	}
	return result, nil
}

func (s MetaWhatsAppSender) buildMessage(to, code string) metaMessage {
	msg := metaMessage{MessagingProduct: "whatsapp", RecipientType: "individual", To: to}
	if s.cfg.Mode == MetaModeText {
		msg.Type = "text"
		msg.Text = &metaText{Body: textMessage(code)}
		return msg
	}
	// Plantilla AUTHENTICATION con botón COPY_CODE: el cuerpo y el botón
	// reciben el mismo código. Meta exige el botón con sub_type "url".
	msg.Type = "template"
	msg.Template = &metaTemplate{
		Name:     s.cfg.TemplateName,
		Language: metaLanguage{Code: s.cfg.LanguageCode},
		Components: []metaComponent{
			{Type: "body", Parameters: []metaParameter{{Type: "text", Text: code}}},
			{Type: "button", SubType: "url", Index: "0", Parameters: []metaParameter{{Type: "text", Text: code}}},
		},
	}
	return msg
}

func textMessage(code string) string {
	return "NAVA — Tu código de verificación es: " + code +
		"\n\nNo lo compartas con nadie. Si no lo solicitaste, ignora este mensaje."
}
