package google_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"system-barbershop/internal/modules/googlecalendar"
	"system-barbershop/internal/modules/googlecalendar/google"
)

type recordedRequest struct {
	method string
	path   string
	query  url.Values
	auth   string
	body   map[string]any
}

type eventsStub struct {
	mu       sync.Mutex
	requests []recordedRequest
	respond  func(r *http.Request, w http.ResponseWriter)
	server   *httptest.Server
}

func newEventsStub(t *testing.T) *eventsStub {
	t.Helper()
	s := &eventsStub{}
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		s.mu.Lock()
		s.requests = append(s.requests, recordedRequest{
			method: r.Method, path: r.URL.EscapedPath(), query: r.URL.Query(), auth: r.Header.Get("Authorization"), body: body,
		})
		respond := s.respond
		s.mu.Unlock()
		if respond != nil {
			respond(r, w)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"evento-1","etag":"\"etag-1\""}`))
	}))
	t.Cleanup(s.server.Close)
	return s
}

func (s *eventsStub) client() *google.EventsClient {
	return google.NewEventsClient(s.server.URL, nil)
}

func sampleEvent() googlecalendar.Event {
	reminder := 30
	return googlecalendar.Event{
		ID:          "nava0123456789abcdef0123456789abcdef",
		Summary:     "Ana Pérez — Corte clásico",
		Description: "Servicio: Corte clásico\nEstado: Confirmada",
		StartsAt:    time.Date(2026, 10, 11, 15, 0, 0, 0, time.UTC),
		EndsAt:      time.Date(2026, 10, 11, 15, 40, 0, 0, time.UTC),
		TimeZone:    "America/Bogota",
		Private: map[string]string{
			googlecalendar.PropResourceType: "appointment",
			googlecalendar.PropResourceID:   "a-1",
			googlecalendar.PropConnectionID: "c-1",
		},
		ReminderMinutes: &reminder,
		AttendeeEmail:   "ana@ejemplo.test",
	}
}

func TestInsert_SendsTheAgreedEventShapeWithBearerTokenAndInvitation(t *testing.T) {
	stub := newEventsStub(t)
	published, err := stub.client().Insert(context.Background(), "token-de-acceso", "primary", sampleEvent(), true)
	if err != nil || published.EventID != "evento-1" || published.ETag != `"etag-1"` {
		t.Fatalf("Insert: %+v %v", published, err)
	}

	req := stub.requests[0]
	if req.method != http.MethodPost || req.path != "/calendars/primary/events" || req.auth != "Bearer token-de-acceso" {
		t.Fatalf("solicitud inesperada: %+v", req)
	}
	if req.query.Get("sendUpdates") != "all" {
		t.Fatalf("sendUpdates = %q, el cliente invitado debe recibir la invitación (DEC-122)", req.query.Get("sendUpdates"))
	}
	body := req.body
	if body["id"] != "nava0123456789abcdef0123456789abcdef" || body["status"] != "confirmed" ||
		body["summary"] != "Ana Pérez — Corte clásico" {
		t.Fatalf("cuerpo = %v", body)
	}
	start := body["start"].(map[string]any)
	if start["dateTime"] != "2026-10-11T10:00:00-05:00" || start["timeZone"] != "America/Bogota" {
		t.Fatalf("el evento representa el mismo instante en la zona de la barbería: %v", start)
	}
	props := body["extendedProperties"].(map[string]any)["private"].(map[string]any)
	if props[googlecalendar.PropConnectionID] != "c-1" || props[googlecalendar.PropResourceID] != "a-1" {
		t.Fatalf("propiedades privadas = %v", props)
	}
	rem := body["reminders"].(map[string]any)
	overrides := rem["overrides"].([]any)
	if rem["useDefault"] != false || len(overrides) != 1 ||
		overrides[0].(map[string]any)["minutes"] != float64(30) || overrides[0].(map[string]any)["method"] != "popup" {
		t.Fatalf("un único recordatorio emergente de 30 minutos: %v", rem)
	}
	attendees := body["attendees"].([]any)
	if len(attendees) != 1 || attendees[0].(map[string]any)["email"] != "ana@ejemplo.test" {
		t.Fatalf("attendees = %v", attendees)
	}
	for _, flag := range []string{"guestsCanModify", "guestsCanInviteOthers", "guestsCanSeeOtherGuests"} {
		if body[flag] != false {
			t.Fatalf("%s debe ser false para el invitado (DEC-122): %v", flag, body[flag])
		}
	}
	// El correo no aparece en ningún otro lugar del cuerpo.
	raw, _ := json.Marshal(body)
	if strings.Count(string(raw), "ana@ejemplo.test") != 1 {
		t.Fatalf("el correo solo viaja como asistente: %s", raw)
	}
}

func TestPatch_WithoutGuestClearsAttendeesAndUsesCalendarDefaults(t *testing.T) {
	stub := newEventsStub(t)
	ev := sampleEvent()
	ev.AttendeeEmail = ""
	ev.ReminderMinutes = nil

	if _, err := stub.client().Patch(context.Background(), "t", "primary", "evento-1", ev, false); err != nil {
		t.Fatal(err)
	}
	req := stub.requests[0]
	if req.method != http.MethodPatch || req.path != "/calendars/primary/events/evento-1" || req.query.Get("sendUpdates") != "none" {
		t.Fatalf("solicitud inesperada: %+v", req)
	}
	if _, hasID := req.body["id"]; hasID {
		t.Fatal("un PATCH nunca reenvía el id")
	}
	if attendees := req.body["attendees"].([]any); len(attendees) != 0 {
		t.Fatalf("sin invitado se envía attendees vacío para retirar al anterior: %v", attendees)
	}
	if _, has := req.body["guestsCanModify"]; has {
		t.Fatal("sin invitado no se envían las banderas de invitados")
	}
	if rem := req.body["reminders"].(map[string]any); rem["useDefault"] != true {
		t.Fatalf("sin reminder_minutes: useDefault = true, fue %v", rem)
	}
}

func TestDelete_EscapesTheCalendarIDAndHonoursNotify(t *testing.T) {
	stub := newEventsStub(t)
	stub.respond = func(_ *http.Request, w http.ResponseWriter) { w.WriteHeader(http.StatusNoContent) }

	if err := stub.client().Delete(context.Background(), "t", "barbero@ejemplo.test", "evento-1", true); err != nil {
		t.Fatal(err)
	}
	req := stub.requests[0]
	if req.method != http.MethodDelete || req.path != "/calendars/barbero@ejemplo.test/events/evento-1" && req.path != "/calendars/barbero%40ejemplo.test/events/evento-1" {
		t.Fatalf("solicitud inesperada: %+v", req)
	}
	if req.query.Get("sendUpdates") != "all" {
		t.Fatalf("la cancelación se notifica al invitado: %v", req.query)
	}
}

func TestErrors_AreReducedToStatusReasonAndRetryAfterWithoutBodyOrToken(t *testing.T) {
	stub := newEventsStub(t)
	stub.respond = func(_ *http.Request, w http.ResponseWriter) {
		w.Header().Set("Retry-After", "120")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":{"code":403,"message":"detalle con ana@ejemplo.test","errors":[{"reason":"userRateLimitExceeded"}]}}`))
	}
	_, err := stub.client().Insert(context.Background(), "token-secreto", "primary", sampleEvent(), false)
	apiErr, ok := asAPI(err)
	if !ok || apiErr.Status != 403 || apiErr.Reason != "userRateLimitExceeded" || apiErr.RetryAfter != 2*time.Minute {
		t.Fatalf("error = %#v", err)
	}
	for _, leak := range []string{"token-secreto", "ana@ejemplo.test", "detalle"} {
		if strings.Contains(err.Error(), leak) {
			t.Fatalf("el error no debe contener %q: %v", leak, err)
		}
	}
}

func TestStatusMapping_GoneAndConflictAreRecognised(t *testing.T) {
	for status, check := range map[int]func(error) bool{
		http.StatusNotFound: googlecalendar.IsGone, http.StatusGone: googlecalendar.IsGone, http.StatusConflict: googlecalendar.IsConflict,
	} {
		stub := newEventsStub(t)
		stub.respond = func(_ *http.Request, w http.ResponseWriter) { w.WriteHeader(status) }
		_, err := stub.client().Patch(context.Background(), "t", "primary", "e", sampleEvent(), false)
		if !check(err) {
			t.Fatalf("el estado %d debe reconocerse: %v", status, err)
		}
	}
}

func TestListEventIDs_FiltersByConnectionPropertyPagesAndSkipsCancelled(t *testing.T) {
	stub := newEventsStub(t)
	stub.respond = func(r *http.Request, w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("pageToken") == "" {
			_, _ = w.Write([]byte(`{"items":[{"id":"e1","status":"confirmed"},{"id":"e2","status":"cancelled"}],"nextPageToken":"p2"}`))
			return
		}
		_, _ = w.Write([]byte(`{"items":[{"id":"e3","status":"confirmed"}]}`))
	}
	timeMin := time.Date(2026, 10, 10, 14, 0, 0, 0, time.UTC)

	ids, err := stub.client().ListEventIDs(context.Background(), "t", "primary", "c-1", timeMin)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 || !ids["e1"] || !ids["e3"] || ids["e2"] {
		t.Fatalf("ids = %v (un evento cancelado equivale a borrado)", ids)
	}
	if len(stub.requests) != 2 {
		t.Fatalf("debe recorrer las dos páginas, hizo %d solicitudes", len(stub.requests))
	}
	q := stub.requests[0].query
	if q.Get("privateExtendedProperty") != "navaConnectionId=c-1" || q.Get("timeMin") != "2026-10-10T14:00:00Z" ||
		q.Get("showDeleted") != "false" || q.Has("syncToken") {
		t.Fatalf("solo los eventos de NAVA de esa conexión, sin syncToken: %v", q)
	}
	if stub.requests[1].query.Get("pageToken") != "p2" {
		t.Fatalf("segunda página: %v", stub.requests[1].query)
	}
}

func TestListEventIDs_FailureAndRunawayPagination(t *testing.T) {
	stub := newEventsStub(t)
	stub.respond = func(_ *http.Request, w http.ResponseWriter) { w.WriteHeader(http.StatusServiceUnavailable) }
	if _, err := stub.client().ListEventIDs(context.Background(), "t", "primary", "c-1", time.Now()); err == nil {
		t.Fatal("un 503 al listar es un error: nunca se asume que faltan eventos")
	}

	stub2 := newEventsStub(t)
	stub2.respond = func(_ *http.Request, w http.ResponseWriter) {
		_, _ = w.Write([]byte(`{"items":[],"nextPageToken":"siempre"}`))
	}
	if _, err := stub2.client().ListEventIDs(context.Background(), "t", "primary", "c-1", time.Now()); err == nil {
		t.Fatal("la paginación sin fin debe cortarse")
	}
}

func asAPI(err error) (*googlecalendar.APIError, bool) {
	apiErr, ok := err.(*googlecalendar.APIError)
	return apiErr, ok
}
