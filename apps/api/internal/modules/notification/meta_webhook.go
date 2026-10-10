package notification

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"system-barbershop/internal/modules/auth"
)

// metaSignaturePrefix es el esquema con el que Meta firma el cuerpo exacto de
// cada notificación en X-Hub-Signature-256.
const metaSignaturePrefix = "sha256="

// waIDPattern es el formato con el que Meta identifica a quien escribe: el
// número en E.164 sin el '+'.
var waIDPattern = regexp.MustCompile(`^[1-9][0-9]{7,14}$`)

// VerifyMetaSignature comprueba en tiempo constante que header sea el HMAC-SHA256
// de body con el secreto de la aplicación de Meta. Debe ejecutarse antes de
// interpretar el cuerpo.
func VerifyMetaSignature(appSecret string, body []byte, header string) bool {
	if appSecret == "" || !strings.HasPrefix(header, metaSignaturePrefix) {
		return false
	}
	got, err := hex.DecodeString(strings.TrimPrefix(header, metaSignaturePrefix))
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write(body)
	return hmac.Equal(got, mac.Sum(nil))
}

// MetaWebhookPayload recoge solo lo que NAVA usa de una notificación de Meta.
// Deliberadamente no decodifica el contenido de los mensajes entrantes: no hace
// falta y evita guardar o registrar texto de personas (RN-DAT-02).
type MetaWebhookPayload struct {
	Object string `json:"object"`
	Entry  []struct {
		Changes []struct {
			Field string           `json:"field"`
			Value MetaWebhookValue `json:"value"`
		} `json:"changes"`
	} `json:"entry"`
}

// MetaWebhookValue es el cuerpo de un cambio del campo «messages».
type MetaWebhookValue struct {
	Metadata struct {
		PhoneNumberID string `json:"phone_number_id"`
	} `json:"metadata"`
	Messages []struct {
		ID        string `json:"id"`
		From      string `json:"from"`
		Timestamp string `json:"timestamp"`
	} `json:"messages"`
	Statuses []struct {
		ID        string `json:"id"`
		Status    string `json:"status"`
		Timestamp string `json:"timestamp"`
		Errors    []struct {
			Code int `json:"code"`
		} `json:"errors"`
	} `json:"statuses"`
}

// ErrInvalidMetaPayload indica un cuerpo que no es una notificación de Meta.
var ErrInvalidMetaPayload = errors.New("notification: notificación de Meta inválida")

// ParseMetaWebhook decodifica body. Un objeto distinto de
// whatsapp_business_account es válido pero no tiene nada que procesar.
func ParseMetaWebhook(body []byte) (MetaWebhookPayload, error) {
	var payload MetaWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return MetaWebhookPayload{}, ErrInvalidMetaPayload
	}
	return payload, nil
}

// MetaWebhookService traduce las notificaciones de Meta en registros de
// ventana y de estados de entrega. No decide ningún envío: un webhook perdido
// no cierra la ventana real, así que Meta sigue siendo quien valida (DEC-126).
type MetaWebhookService struct {
	store         WhatsAppEventStore
	phoneNumberID string
	hashSecret    []byte
	logger        *slog.Logger
}

// NewMetaWebhookService construye el servicio. phoneNumberID filtra los eventos
// de otros números de la misma cuenta; hashSecret firma el HMAC de los teléfonos.
func NewMetaWebhookService(store WhatsAppEventStore, phoneNumberID string, hashSecret []byte, logger *slog.Logger) *MetaWebhookService {
	return &MetaWebhookService{store: store, phoneNumberID: phoneNumberID, hashSecret: hashSecret, logger: logger}
}

// PhoneHash es la única forma en que un teléfono llega a la base de datos.
// El prefijo separa este uso del HMAC de otros valores con el mismo secreto.
func (s *MetaWebhookService) PhoneHash(phoneE164 string) string {
	return auth.HMACHex("whatsapp-window:"+phoneE164, s.hashSecret)
}

// ConversationOpen informa si hay un mensaje entrante registrado en las
// últimas 24 horas. false significa «sin registro vigente», no «Meta lo
// rechazará».
func (s *MetaWebhookService) ConversationOpen(ctx context.Context, phoneE164 string) (bool, error) {
	if !e164Pattern.MatchString(phoneE164) {
		return false, ErrInvalidRecipient
	}
	return s.store.ConversationOpen(ctx, s.PhoneHash(phoneE164))
}

// Process aplica cada evento del número configurado. Un fallo del almacén no
// detiene los demás eventos; se devuelve al final para que Meta reintente la
// notificación, lo que es seguro porque cada registro es idempotente.
func (s *MetaWebhookService) Process(ctx context.Context, payload MetaWebhookPayload) error {
	if payload.Object != "whatsapp_business_account" {
		return nil
	}
	var failures []error
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			if change.Field != "messages" || change.Value.Metadata.PhoneNumberID != s.phoneNumberID {
				continue
			}
			failures = append(failures, s.processValue(ctx, change.Value)...)
		}
	}
	return errors.Join(failures...)
}

func (s *MetaWebhookService) processValue(ctx context.Context, value MetaWebhookValue) []error {
	var failures []error

	for _, message := range value.Messages {
		at, ok := parseMetaTimestamp(message.Timestamp)
		if !ok || !waIDPattern.MatchString(message.From) {
			s.logger.WarnContext(ctx, "whatsapp webhook: mensaje entrante descartado por formato")
			continue
		}
		if err := s.store.RecordInbound(ctx, s.PhoneHash("+"+message.From), at); err != nil {
			failures = append(failures, fmt.Errorf("notification: registrar mensaje entrante: %w", err))
		}
	}

	for _, status := range value.Statuses {
		if !isTrackedStatus(status.Status) {
			continue
		}
		at, ok := parseMetaTimestamp(status.Timestamp)
		if !ok || status.ID == "" || len(status.ID) > 255 {
			s.logger.WarnContext(ctx, "whatsapp webhook: estado descartado por formato")
			continue
		}
		var errorCode *int
		if len(status.Errors) > 0 {
			errorCode = &status.Errors[0].Code
		}
		inserted, err := s.store.RecordStatus(ctx, status.ID, status.Status, at, errorCode)
		if err != nil {
			failures = append(failures, fmt.Errorf("notification: registrar estado de entrega: %w", err))
			continue
		}
		if inserted && status.Status == "failed" {
			attrs := []any{slog.String("wamid", status.ID)}
			if errorCode != nil {
				attrs = append(attrs, slog.Int("error_code", *errorCode))
			}
			s.logger.WarnContext(ctx, "whatsapp: Meta no pudo entregar el mensaje", attrs...)
		}
	}
	return failures
}

func isTrackedStatus(status string) bool {
	switch status {
	case "sent", "delivered", "read", "failed":
		return true
	}
	return false
}

// parseMetaTimestamp convierte los segundos Unix en texto que usa Meta.
func parseMetaTimestamp(raw string) (time.Time, bool) {
	seconds, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || seconds <= 0 {
		return time.Time{}, false
	}
	return time.Unix(seconds, 0).UTC(), true
}
