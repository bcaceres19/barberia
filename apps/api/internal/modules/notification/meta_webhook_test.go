package notification_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"system-barbershop/internal/modules/notification"
)

const (
	webhookAppSecret = "app-secret-de-prueba-0123456789"
	webhookNumberID  = "1315949954943476"
)

func signBody(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyMetaSignature(t *testing.T) {
	body := []byte(`{"object":"whatsapp_business_account"}`)
	valid := signBody(webhookAppSecret, body)

	tests := []struct {
		name   string
		secret string
		body   []byte
		header string
		want   bool
	}{
		{"firma válida", webhookAppSecret, body, valid, true},
		{"cuerpo alterado", webhookAppSecret, []byte(`{"object":"otro"}`), valid, false},
		{"secreto distinto", "otro-secreto-de-prueba-0123456", body, valid, false},
		{"sin cabecera", webhookAppSecret, body, "", false},
		{"sin prefijo", webhookAppSecret, body, strings.TrimPrefix(valid, "sha256="), false},
		{"hexadecimal inválido", webhookAppSecret, body, "sha256=zzzz", false},
		{"prefijo de otro algoritmo", webhookAppSecret, body, strings.Replace(valid, "sha256=", "sha1=", 1), false},
		{"secreto vacío", "", body, signBody("", body), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := notification.VerifyMetaSignature(tt.secret, tt.body, tt.header); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

type inboundCall struct {
	hash string
	at   time.Time
}

type statusCall struct {
	wamid, status string
	at            time.Time
	errorCode     *int
}

type fakeEventStore struct {
	inbound     []inboundCall
	statuses    []statusCall
	seen        map[string]bool
	inboundErr  error
	statusErr   error
	openByHash  map[string]bool
	openQueried []string
}

func (f *fakeEventStore) RecordInbound(_ context.Context, hash string, at time.Time) error {
	if f.inboundErr != nil {
		return f.inboundErr
	}
	f.inbound = append(f.inbound, inboundCall{hash, at})
	return nil
}

func (f *fakeEventStore) RecordStatus(_ context.Context, wamid, status string, at time.Time, code *int) (bool, error) {
	if f.statusErr != nil {
		return false, f.statusErr
	}
	if f.seen == nil {
		f.seen = map[string]bool{}
	}
	key := wamid + "|" + status
	if f.seen[key] {
		return false, nil
	}
	f.seen[key] = true
	f.statuses = append(f.statuses, statusCall{wamid, status, at, code})
	return true, nil
}

func (f *fakeEventStore) ConversationOpen(_ context.Context, hash string) (bool, error) {
	f.openQueried = append(f.openQueried, hash)
	return f.openByHash[hash], nil
}

func newWebhookService(store *fakeEventStore, logs *bytes.Buffer) *notification.MetaWebhookService {
	return notification.NewMetaWebhookService(store, webhookNumberID, []byte("secreto-hmac-de-prueba-0123456789"),
		slog.New(slog.NewTextHandler(logs, nil)))
}

func webhookPayload(t *testing.T, numberID, value string) notification.MetaWebhookPayload {
	t.Helper()
	body := `{"object":"whatsapp_business_account","entry":[{"id":"957400180752495","changes":[{"field":"messages","value":{` +
		`"messaging_product":"whatsapp","metadata":{"display_phone_number":"573108953493","phone_number_id":"` + numberID + `"},` + value + `}}]}]}`
	payload, err := notification.ParseMetaWebhook([]byte(body))
	if err != nil {
		t.Fatalf("ParseMetaWebhook: %v", err)
	}
	return payload
}

func TestMetaWebhook_ParseRejectsNonJSON(t *testing.T) {
	if _, err := notification.ParseMetaWebhook([]byte("no es json")); !errors.Is(err, notification.ErrInvalidMetaPayload) {
		t.Fatalf("expected ErrInvalidMetaPayload, got %v", err)
	}
}

func TestMetaWebhook_InboundMessageRecordsHashedPhoneAndNeverTheText(t *testing.T) {
	store := &fakeEventStore{}
	var logs bytes.Buffer
	svc := newWebhookService(store, &logs)

	payload := webhookPayload(t, webhookNumberID,
		`"contacts":[{"wa_id":"573016823624","profile":{"name":"Persona de prueba"}}],`+
			`"messages":[{"from":"573016823624","id":"wamid.in1","timestamp":"1790000000","type":"text","text":{"body":"mensaje privado"}}]`)
	if err := svc.Process(context.Background(), payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(store.inbound) != 1 {
		t.Fatalf("expected one inbound registration, got %d", len(store.inbound))
	}
	got := store.inbound[0]
	if got.hash != svc.PhoneHash("+573016823624") || len(got.hash) != 64 || strings.Contains(got.hash, "573016823624") {
		t.Fatalf("expected the phone to arrive only as its HMAC, got %q", got.hash)
	}
	if !got.at.Equal(time.Unix(1790000000, 0).UTC()) {
		t.Fatalf("unexpected timestamp %v", got.at)
	}
	for _, secret := range []string{"573016823624", "mensaje privado", "Persona de prueba"} {
		if strings.Contains(logs.String(), secret) {
			t.Fatalf("log leaked %q: %s", secret, logs.String())
		}
	}
}

func TestMetaWebhook_PhoneHashIsKeyedAndStable(t *testing.T) {
	a := newWebhookService(&fakeEventStore{}, &bytes.Buffer{})
	b := notification.NewMetaWebhookService(&fakeEventStore{}, webhookNumberID, []byte("otro-secreto-hmac-de-prueba-012345"),
		slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))

	if a.PhoneHash("+573016823624") != a.PhoneHash("+573016823624") {
		t.Fatal("the hash must be stable")
	}
	if a.PhoneHash("+573016823624") == b.PhoneHash("+573016823624") {
		t.Fatal("a different secret must produce a different hash")
	}
	if a.PhoneHash("+573016823624") == a.PhoneHash("+573016823625") {
		t.Fatal("different phones must not collide")
	}
}

func TestMetaWebhook_IgnoresEventsFromAnotherPhoneNumberID(t *testing.T) {
	store := &fakeEventStore{}
	svc := newWebhookService(store, &bytes.Buffer{})

	payload := webhookPayload(t, "999999999999",
		`"messages":[{"from":"573016823624","id":"wamid.x","timestamp":"1790000000","type":"text"}],`+
			`"statuses":[{"id":"wamid.y","status":"sent","timestamp":"1790000000"}]`)
	if err := svc.Process(context.Background(), payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.inbound) != 0 || len(store.statuses) != 0 {
		t.Fatalf("events of another number must be ignored, got %+v %+v", store.inbound, store.statuses)
	}
}

func TestMetaWebhook_IgnoresOtherObjectsAndFields(t *testing.T) {
	store := &fakeEventStore{}
	svc := newWebhookService(store, &bytes.Buffer{})

	other, _ := notification.ParseMetaWebhook([]byte(`{"object":"page","entry":[]}`))
	if err := svc.Process(context.Background(), other); err != nil || len(store.inbound) != 0 {
		t.Fatalf("a non-WhatsApp object must be ignored: %v", err)
	}
	field, _ := notification.ParseMetaWebhook([]byte(`{"object":"whatsapp_business_account","entry":[{"changes":[{"field":"account_update","value":{"metadata":{"phone_number_id":"` + webhookNumberID + `"},"messages":[{"from":"573016823624","timestamp":"1790000000"}]}}]}]}`))
	if err := svc.Process(context.Background(), field); err != nil || len(store.inbound) != 0 {
		t.Fatalf("a field other than messages must be ignored: %v", err)
	}
}

func TestMetaWebhook_DiscardsMalformedEvents(t *testing.T) {
	store := &fakeEventStore{}
	svc := newWebhookService(store, &bytes.Buffer{})

	payload := webhookPayload(t, webhookNumberID,
		`"messages":[`+
			`{"from":"abc","id":"a","timestamp":"1790000000"},`+
			`{"from":"573016823624","id":"b","timestamp":"no-es-numero"},`+
			`{"from":"573016823624","id":"c","timestamp":"-5"}],`+
			`"statuses":[`+
			`{"id":"","status":"sent","timestamp":"1790000000"},`+
			`{"id":"wamid.z","status":"deleted","timestamp":"1790000000"},`+
			`{"id":"wamid.z","status":"sent","timestamp":"x"}]`)
	if err := svc.Process(context.Background(), payload); err != nil {
		t.Fatalf("malformed events must not fail the notification: %v", err)
	}
	if len(store.inbound) != 0 || len(store.statuses) != 0 {
		t.Fatalf("nothing malformed may be stored, got %+v %+v", store.inbound, store.statuses)
	}
}

func TestMetaWebhook_StatusesAreRecordedOnceAndFailuresAreLoggedWithoutPersonalData(t *testing.T) {
	store := &fakeEventStore{}
	var logs bytes.Buffer
	svc := newWebhookService(store, &logs)

	failed := `"statuses":[{"id":"wamid.f1","status":"failed","timestamp":"1790000000","recipient_id":"573016823624",` +
		`"errors":[{"code":131047,"title":"Re-engagement message","message":"texto con 573016823624"}]}]`
	for i := 0; i < 2; i++ {
		if err := svc.Process(context.Background(), webhookPayload(t, webhookNumberID, failed)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if len(store.statuses) != 1 {
		t.Fatalf("a repeated notification must not duplicate the status, got %d", len(store.statuses))
	}
	if store.statuses[0].errorCode == nil || *store.statuses[0].errorCode != 131047 {
		t.Fatalf("expected the Meta error code to be kept, got %+v", store.statuses[0].errorCode)
	}
	if strings.Count(logs.String(), "no pudo entregar") != 1 || !strings.Contains(logs.String(), "wamid.f1") || !strings.Contains(logs.String(), "131047") {
		t.Fatalf("expected a single failure log with wamid and code, got %q", logs.String())
	}
	for _, secret := range []string{"573016823624", "texto con"} {
		if strings.Contains(logs.String(), secret) {
			t.Fatalf("log leaked %q: %s", secret, logs.String())
		}
	}
}

func TestMetaWebhook_StoreFailureIsReturnedButOtherEventsStillProcess(t *testing.T) {
	store := &fakeEventStore{inboundErr: errors.New("db caída")}
	svc := newWebhookService(store, &bytes.Buffer{})

	payload := webhookPayload(t, webhookNumberID,
		`"messages":[{"from":"573016823624","id":"a","timestamp":"1790000000"}],`+
			`"statuses":[{"id":"wamid.s1","status":"delivered","timestamp":"1790000000"}]`)
	if err := svc.Process(context.Background(), payload); err == nil {
		t.Fatal("a store failure must be returned so Meta retries")
	}
	if len(store.statuses) != 1 {
		t.Fatal("a failing inbound registration must not drop the status events of the same notification")
	}
}

func TestMetaWebhook_ConversationOpenUsesTheHashAndValidatesThePhone(t *testing.T) {
	store := &fakeEventStore{openByHash: map[string]bool{}}
	svc := newWebhookService(store, &bytes.Buffer{})
	store.openByHash[svc.PhoneHash("+573016823624")] = true

	open, err := svc.ConversationOpen(context.Background(), "+573016823624")
	if err != nil || !open {
		t.Fatalf("expected an open conversation, got %v, %v", open, err)
	}
	if len(store.openQueried) != 1 || store.openQueried[0] != svc.PhoneHash("+573016823624") {
		t.Fatalf("the store must only see the hash, got %v", store.openQueried)
	}
	if open, _ := svc.ConversationOpen(context.Background(), "+573000000000"); open {
		t.Fatal("an unknown phone must not have an open conversation")
	}
	if _, err := svc.ConversationOpen(context.Background(), "3016823624"); !errors.Is(err, notification.ErrInvalidRecipient) {
		t.Fatalf("expected ErrInvalidRecipient, got %v", err)
	}
}
