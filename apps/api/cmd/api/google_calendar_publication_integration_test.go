// Pruebas de integración de la PUBLICACIÓN en Google Calendar (issue #324,
// DEC-099, DEC-101, DEC-102, DEC-122) de punta a punta: router real,
// PostgreSQL real con dos barberías, el worker real (barberia_worker + el
// publicador) y un Google FALSO con almacén de eventos en memoria. Ninguna
// prueba toca la red real.
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/oauth2"

	"system-barbershop/internal/modules/googlecalendar"
	googlecalendargoogle "system-barbershop/internal/modules/googlecalendar/google"
	googlecalendarpostgres "system-barbershop/internal/modules/googlecalendar/postgres"
	"system-barbershop/internal/platform/clock"
	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
)

// fakeCalendar es el almacén de eventos del Google falso.
type fakeCalendar struct {
	mu       sync.Mutex
	events   map[string]map[string]any
	requests []calendarRequest
	// failWith, si es distinto de cero, hace que toda llamada responda ese estado
	// (simula una caída de Google).
	failWith int
}

type calendarRequest struct {
	method       string
	eventID      string
	sendUpdates  string
	body         map[string]any
	responseCode int
}

func (c *fakeCalendar) handle(w http.ResponseWriter, r *http.Request) {
	raw, _ := io.ReadAll(r.Body)
	var body map[string]any
	_ = json.Unmarshal(raw, &body)

	path := strings.TrimPrefix(r.URL.Path, "/calendar/v3/calendars/")
	parts := strings.Split(path, "/") // {calendario}/events[/{id}]
	eventID := ""
	if len(parts) == 3 {
		eventID = parts[2]
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.events == nil {
		c.events = map[string]map[string]any{}
	}
	record := calendarRequest{method: r.Method, eventID: eventID, sendUpdates: r.URL.Query().Get("sendUpdates"), body: body}
	if c.failWith != 0 {
		record.responseCode = c.failWith
		c.requests = append(c.requests, record)
		w.WriteHeader(c.failWith)
		return
	}
	respond := func(code int, payload any) {
		record.responseCode = code
		c.requests = append(c.requests, record)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		if payload != nil {
			_ = json.NewEncoder(w).Encode(payload)
		}
	}

	switch r.Method {
	case http.MethodPost:
		id, _ := body["id"].(string)
		if _, exists := c.events[id]; exists {
			respond(http.StatusConflict, map[string]any{"error": map[string]any{"code": 409}})
			return
		}
		record.eventID = id
		c.events[id] = body
		respond(http.StatusOK, map[string]any{"id": id, "etag": `"1"`})
	case http.MethodPatch:
		existing, ok := c.events[eventID]
		if !ok {
			respond(http.StatusNotFound, map[string]any{"error": map[string]any{"code": 404}})
			return
		}
		for k, v := range body {
			existing[k] = v
		}
		respond(http.StatusOK, map[string]any{"id": eventID, "etag": `"2"`})
	case http.MethodDelete:
		if _, ok := c.events[eventID]; !ok {
			respond(http.StatusGone, map[string]any{"error": map[string]any{"code": 410}})
			return
		}
		delete(c.events, eventID)
		respond(http.StatusNoContent, nil)
	case http.MethodGet:
		wantProp := r.URL.Query().Get("privateExtendedProperty") // navaConnectionId=<id>
		items := []map[string]any{}
		for id, ev := range c.events {
			props, _ := ev["extendedProperties"].(map[string]any)
			private, _ := props["private"].(map[string]any)
			if strings.TrimPrefix(wantProp, "navaConnectionId=") == asString(private["navaConnectionId"]) {
				items = append(items, map[string]any{"id": id, "status": "confirmed"})
			}
		}
		respond(http.StatusOK, map[string]any{"items": items})
	default:
		respond(http.StatusMethodNotAllowed, nil)
	}
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

// eventBy devuelve el cuerpo del evento publicado para un recurso de NAVA.
func (c *fakeCalendar) eventFor(resourceID string) (string, map[string]any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for id, ev := range c.events {
		props, _ := ev["extendedProperties"].(map[string]any)
		private, _ := props["private"].(map[string]any)
		if private["navaResourceId"] == resourceID {
			return id, ev, true
		}
	}
	return "", nil, false
}

func (c *fakeCalendar) callsFor(eventID string) []calendarRequest {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []calendarRequest
	for _, r := range c.requests {
		if r.eventID == eventID {
			out = append(out, r)
		}
	}
	return out
}

func (c *fakeCalendar) dropEvent(eventID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.events, eventID)
}

func setupWorkerTestDB(t *testing.T) *database.DB {
	t.Helper()
	dsn := os.Getenv("TEST_WORKER_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://barberia_worker@localhost:5432/barberia_test?sslmode=disable"
	}
	cfg := config.Config{
		Environment: "test", DatabaseMaxConns: 5, DatabaseMinConns: 1,
		DatabaseMaxConnLifetime: time.Hour, DatabaseMaxConnIdleTime: 30 * time.Minute,
		DatabaseConnectTimeout: 5 * time.Second, DatabaseStatementTimeout: 10 * time.Second,
	}
	db, err := database.NewDB(config.DatabaseDSN(dsn), cfg)
	if err != nil {
		t.Fatalf("database.NewDB (worker): %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

type publication struct {
	router    http.Handler
	raw       string
	google    *googleStub
	calendar  *fakeCalendar
	publisher *googlecalendar.Publisher
	barber    barberBody
	service   serviceBody
}

// newPublication conecta a un barbero de la barbería A con el Google falso y
// arma el publicador real contra la base del worker.
func newPublication(t *testing.T, db *database.DB, label string) *publication {
	t.Helper()
	stub := newGoogleStub(t)
	calendar := &fakeCalendar{}
	stub.server.Config.Handler.(*http.ServeMux).HandleFunc("/calendar/v3/", calendar.handle)

	router := googleCalendarRouter(t, db, stub)
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, label)
	t.Cleanup(func() {
		doGCal(router, raw, http.MethodDelete, "", "")
		doMyBarberRequest(router, raw, http.MethodDelete, "")
	})

	barber := createBarberViaRouter(t, router, raw, "Pub "+uniqueToken(t, "n"))
	service := createServiceViaRouter(t, router, raw, "Corte "+uniqueToken(t, "s"))
	if rec := doAssignServiceRequest(router, raw, barber.ID, service.ID); rec.Code != http.StatusCreated {
		t.Fatalf("asignar servicio: %d %s", rec.Code, rec.Body.String())
	}
	linkMe(t, router, raw, barber.ID)
	state := startConnection(t, router, raw)
	if rec := doGCal(router, raw, http.MethodPost, "/callback", `{"state":"`+state+`","code":"codigo"}`); rec.Code != http.StatusOK {
		t.Fatalf("conectar: %d %s", rec.Code, rec.Body.String())
	}

	cipher, err := googlecalendar.NewCipherFromKeys("v1", base64.StdEncoding.EncodeToString(bytesRepeat(5, 32)), nil)
	if err != nil {
		t.Fatal(err)
	}
	publisher := googlecalendar.NewPublisher(googlecalendar.PublisherDeps{
		Store: googlecalendarpostgres.NewWorkerRepository(setupWorkerTestDB(t)),
		API:   googlecalendargoogle.NewEventsClient(stub.server.URL+"/calendar/v3", nil),
		Provider: googlecalendargoogle.New(googlecalendargoogle.Config{
			ClientID: "cliente-de-prueba", ClientSecret: "secreto-de-prueba",
			RedirectURL: "http://localhost:5173/panel/barberia/google-calendar/callback",
			Endpoint:    &oauth2.Endpoint{TokenURL: stub.server.URL + "/token"},
			RevokeURL:   stub.server.URL + "/revoke",
		}),
		Cipher: cipher,
		Clock:  clock.System{},
		Config: googlecalendar.DefaultPublisherConfig(),
	})
	return &publication{router: router, raw: raw, google: stub, calendar: calendar, publisher: publisher, barber: barber, service: service}
}

func bytesRepeat(b byte, n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = b
	}
	return out
}

// drain procesa la cola hasta dejarla vacía.
func (p *publication) drain(t *testing.T) {
	t.Helper()
	for range 20 {
		if p.publisher.RunOnce(context.Background()) == 0 {
			return
		}
	}
}

// ageRestoreCheck hace que el chequeo de eventos borrados de esta conexión venza
// ya (el ciclo normal lo repite cada cinco minutos).
func (p *publication) ageRestoreCheck(t *testing.T) {
	t.Helper()
	db := setupTestDB(t)
	err := db.InTenantTx(context.Background(), database.BarbershopID(shopA), func(ctx context.Context, q database.Queries) error {
		_, err := q.Exec(ctx,
			`UPDATE google_calendar_connection SET last_restore_check_at = now() - interval '1 hour'
			  WHERE barbershop_id = $1 AND barber_id = $2`, shopA, p.barber.ID)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}

// restoreAll ejecuta el chequeo de eventos borrados sobre todas las conexiones
// vencidas (la base de pruebas es compartida, así que hay que recorrerlas todas).
func (p *publication) restoreAll() {
	for range 500 {
		if !p.publisher.RestoreOnce(context.Background()) {
			return
		}
	}
}

// createAppointment registra un turno manual (HU-061) con un cliente con correo.
func (p *publication) createAppointment(t *testing.T, startsAt string) (id string, version string) {
	t.Helper()
	body := `{"barberId":"` + p.barber.ID + `","serviceId":"` + p.service.ID + `","attendeeName":"Ana Pérez",` +
		`"customerFullName":"Ana Pérez","customerEmail":"ana.` + strings.ToLower(uniqueToken(t, "e")) + `@ejemplo.test","startsAt":"` + startsAt + `"}`
	rec := doJSONRequest(p.router, http.MethodPost, p.raw, "/appointments", uniqueToken(t, "appt"), body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("crear turno: %d %s", rec.Code, rec.Body.String())
	}
	var created struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	return created.ID, p.version(t, created.ID)
}

func (p *publication) version(t *testing.T, appointmentID string) string {
	t.Helper()
	rec := doJSONRequest(p.router, http.MethodGet, p.raw, "/appointments/"+appointmentID, "", "")
	var detail struct {
		VersionToken string `json:"versionToken"`
	}
	if rec.Code != http.StatusOK || json.Unmarshal(rec.Body.Bytes(), &detail) != nil || detail.VersionToken == "" {
		t.Fatalf("detalle del turno: %d %s", rec.Code, rec.Body.String())
	}
	return detail.VersionToken
}

func (p *publication) post(t *testing.T, path, ifMatch, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/private"+path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("If-Match", ifMatch)
	req.Header.Set("Idempotency-Key", "k-"+base64.RawURLEncoding.EncodeToString([]byte(time.Now().Format(time.RFC3339Nano))))
	req.AddCookie(&http.Cookie{Name: cookieName, Value: p.raw})
	rec := httptest.NewRecorder()
	p.router.ServeHTTP(rec, req)
	return rec
}

func futureCivil(days, hour int) string {
	d := time.Now().UTC().AddDate(0, 0, days)
	return time.Date(d.Year(), d.Month(), d.Day(), hour, 0, 0, 0, time.UTC).Format("2006-01-02T15:04:05")
}

func TestPublication_Appointment_CreateRescheduleCancel_FullJourney(t *testing.T) {
	db := setupTestDB(t)
	p := newPublication(t, db, "pub-journey")

	// 1) Reserva → un evento en el calendario del barbero, con el cliente invitado.
	apptID, version := p.createAppointment(t, futureCivil(9, 10))
	p.drain(t)
	eventID, event, ok := p.calendar.eventFor(apptID)
	if !ok {
		t.Fatalf("la cita debe publicarse como un evento (1 cita, 1 evento)")
	}
	if !strings.HasPrefix(asString(event["summary"]), "Ana Pérez — Corte ") {
		t.Fatalf("título = %v (DEC-101: «Nombre — Servicio»)", event["summary"])
	}
	attendees, _ := event["attendees"].([]any)
	if len(attendees) != 1 || !strings.HasPrefix(asString(attendees[0].(map[string]any)["email"]), "ana.") {
		t.Fatalf("DEC-122: el cliente queda invitado como asistente: %v", event["attendees"])
	}
	for _, flag := range []string{"guestsCanModify", "guestsCanInviteOthers", "guestsCanSeeOtherGuests"} {
		if event[flag] != false {
			t.Fatalf("%s debe ser false: %v", flag, event[flag])
		}
	}
	created := p.calendar.callsFor(eventID)
	if len(created) != 1 || created[0].method != http.MethodPost || created[0].sendUpdates != "all" {
		t.Fatalf("una sola creación con invitación al cliente: %+v", created)
	}
	if got := strings.ToLower(asString(event["summary"]) + asString(event["description"])); strings.Contains(got, "@") {
		t.Fatal("el correo del cliente no debe aparecer en el título ni en la descripción")
	}
	status := gcalJSON(t, doGCal(p.router, p.raw, http.MethodGet, "", ""))
	if status["lastSyncedAt"] == nil || status["pendingSyncJobs"] != float64(0) || status["failedSyncJobs"] != float64(0) {
		t.Fatalf("tras publicar, la cola queda vacía y se registra la sincronización: %v", status)
	}

	// 2) Reprogramar desde la app → el MISMO evento se actualiza y se avisa al cliente.
	rec := p.post(t, "/appointments/"+apptID+"/reschedule", version, `{"startsAt":"`+futureCivil(9, 14)+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("reprogramar: %d %s", rec.Code, rec.Body.String())
	}
	p.drain(t)
	again := p.calendar.callsFor(eventID)
	if len(again) != 2 || again[1].method != http.MethodPatch || again[1].sendUpdates != "all" {
		t.Fatalf("reprogramar actualiza el mismo evento y notifica el cambio de horario: %+v", again)
	}
	if id2, _, _ := p.calendar.eventFor(apptID); id2 != eventID {
		t.Fatal("el evento no debe cambiar de identidad")
	}

	// 3) Cancelar desde la app → el evento se elimina y el vínculo se cierra.
	version = p.version(t, apptID)
	if rec := p.post(t, "/appointments/"+apptID+"/cancel", version, `{}`); rec.Code != http.StatusOK {
		t.Fatalf("cancelar: %d %s", rec.Code, rec.Body.String())
	}
	p.drain(t)
	if _, _, ok := p.calendar.eventFor(apptID); ok {
		t.Fatal("cancelar elimina el evento de Google")
	}
	last := p.calendar.callsFor(eventID)
	if got := last[len(last)-1]; got.method != http.MethodDelete || got.sendUpdates != "all" {
		t.Fatalf("la cancelación se notifica al invitado: %+v", got)
	}

	// 4) Nada de lo cancelado se recrea, ni por el chequeo periódico.
	before := len(p.calendar.requests)
	p.ageRestoreCheck(t)
	p.restoreAll()
	p.drain(t)
	if _, _, ok := p.calendar.eventFor(apptID); ok || len(p.calendar.requests) > before+1 {
		t.Fatalf("una cita cancelada en la app no se recrea: %d solicitudes nuevas", len(p.calendar.requests)-before)
	}
}

func TestPublication_EventDeletedInGoogle_IsRestoredWithoutTouchingNAVA(t *testing.T) {
	db := setupTestDB(t)
	p := newPublication(t, db, "pub-restore")
	apptID, _ := p.createAppointment(t, futureCivil(10, 11))
	p.drain(t)
	oldID, _, ok := p.calendar.eventFor(apptID)
	if !ok {
		t.Fatal("setup: el evento debe existir")
	}

	// El barbero lo borra por error en Google: NAVA no cambia su cita…
	p.calendar.dropEvent(oldID)
	detail := doJSONRequest(p.router, http.MethodGet, p.raw, "/appointments/"+apptID, "", "")
	if detail.Code != http.StatusOK || !strings.Contains(detail.Body.String(), `"status":"confirmed"`) {
		t.Fatalf("un cambio hecho en Google no altera la cita de NAVA: %d %s", detail.Code, detail.Body.String())
	}

	// …y el chequeo periódico lo detecta y lo recrea (una sola vez, sin duplicados).
	p.ageRestoreCheck(t)
	p.restoreAll()
	p.drain(t)
	newID, _, ok := p.calendar.eventFor(apptID)
	if !ok || newID == oldID {
		t.Fatalf("debe recrearse con un id nuevo (Google no reutiliza el de un evento borrado): %q vs %q", newID, oldID)
	}
	count := 0
	for _, r := range p.calendar.callsFor(newID) {
		if r.method == http.MethodPost {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("se recrea una sola vez, hubo %d creaciones", count)
	}
}

func TestPublication_TimeBlock_PublishesAndDeletes(t *testing.T) {
	db := setupTestDB(t)
	p := newPublication(t, db, "pub-block")

	key := uniqueToken(t, "pub-block")
	body := `{"blockType":"lunch","startsAt":"` + futureCivil(12, 12) + `-05:00","endsAt":"` + futureCivil(12, 13) + `-05:00","reason":"Nota privada del barbero"}`
	rec := doJSONRequest(p.router, http.MethodPost, p.raw, "/barbers/"+p.barber.ID+"/time-blocks", key, body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("crear bloqueo: %d %s", rec.Code, rec.Body.String())
	}
	var block timeBlockBody
	_ = json.Unmarshal(rec.Body.Bytes(), &block)
	p.drain(t)

	eventID, event, ok := p.calendar.eventFor(block.ID)
	if !ok || event["summary"] != "Almuerzo" {
		t.Fatalf("un bloqueo manual se publica con el título de su tipo: %v", event)
	}
	if strings.Contains(asString(event["description"]), "Nota privada") {
		t.Fatal("el motivo del bloqueo es una nota del barbero y no se publica")
	}
	if attendees, _ := event["attendees"].([]any); len(attendees) != 0 {
		t.Fatal("un bloqueo nunca invita a nadie")
	}

	if rec := doJSONRequest(p.router, http.MethodDelete, p.raw, "/barbers/"+p.barber.ID+"/time-blocks/"+block.ID, "", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("retirar bloqueo: %d", rec.Code)
	}
	p.drain(t)
	if _, _, ok := p.calendar.eventFor(block.ID); ok {
		t.Fatal("retirar el bloqueo elimina su evento")
	}
	_ = eventID
}

func TestPublication_GoogleDown_NeverLosesTheAppointment_AndSyncNowRetries(t *testing.T) {
	db := setupTestDB(t)
	p := newPublication(t, db, "pub-down")

	// Google responde 503 a toda llamada de eventos.
	p.calendar.mu.Lock()
	p.calendar.failWith = http.StatusServiceUnavailable
	p.calendar.mu.Unlock()

	apptID, _ := p.createAppointment(t, futureCivil(11, 10))
	p.drain(t)

	// La cita existe aunque Google no haya publicado nada, y el cambio sigue en la cola
	// con un reintento programado (backoff), sin bucle infinito.
	if detail := doJSONRequest(p.router, http.MethodGet, p.raw, "/appointments/"+apptID, "", ""); detail.Code != http.StatusOK {
		t.Fatalf("la cita debe existir sin depender de Google: %d", detail.Code)
	}
	if _, _, ok := p.calendar.eventFor(apptID); ok {
		t.Fatal("con Google caído no se publicó nada")
	}
	status := gcalJSON(t, doGCal(p.router, p.raw, http.MethodGet, "", ""))
	if status["pendingSyncJobs"] != float64(1) || status["failedSyncJobs"] != float64(0) || status["status"] != "connected" {
		t.Fatalf("el cambio espera su reintento y la conexión sigue sana: %v", status)
	}
	attempts := len(p.calendar.requests)
	p.drain(t) // dentro del backoff no se vuelve a llamar a Google
	if len(p.calendar.requests) != attempts {
		t.Fatalf("antes del backoff no se reintenta: %d llamadas más", len(p.calendar.requests)-attempts)
	}

	// Google vuelve. «Sincronizar ahora» solo reordena la cola de esta conexión y es
	// segura ante varios clics.
	p.calendar.mu.Lock()
	p.calendar.failWith = 0
	p.calendar.mu.Unlock()
	for range 3 {
		rec := doGCal(p.router, p.raw, http.MethodPost, "/sync", "")
		if body := gcalJSON(t, rec); rec.Code != http.StatusOK || body["pendingSyncJobs"] != float64(1) || body["failedSyncJobs"] != float64(0) {
			t.Fatalf("sync: %d %v", rec.Code, body)
		}
	}
	p.drain(t)
	if _, _, ok := p.calendar.eventFor(apptID); !ok {
		t.Fatal("tras sincronizar, el evento se publica")
	}
	if status := gcalJSON(t, doGCal(p.router, p.raw, http.MethodGet, "", "")); status["pendingSyncJobs"] != float64(0) {
		t.Fatalf("la cola queda vacía: %v", status)
	}
}

func TestPublication_RevokedPermission_MarksReauthRequiredAndDropsTheQueue(t *testing.T) {
	db := setupTestDB(t)
	p := newPublication(t, db, "pub-revoked")
	p.createAppointment(t, futureCivil(13, 10))

	// Google ya no acepta el refresh token.
	p.google.mu.Lock()
	p.google.tokenStatus, p.google.tokenBody = http.StatusBadRequest, `{"error":"invalid_grant"}`
	p.google.mu.Unlock()
	p.drain(t)

	status := gcalJSON(t, doGCal(p.router, p.raw, http.MethodGet, "", ""))
	if status["status"] != "reauth_required" || status["pendingSyncJobs"] != float64(0) {
		t.Fatalf("un permiso revocado deja la conexión en reauth_required y la cola vacía: %v", status)
	}
	// Las citas nuevas ya no encolan nada mientras no haya conexión viva.
	p.createAppointment(t, futureCivil(14, 10))
	if status := gcalJSON(t, doGCal(p.router, p.raw, http.MethodGet, "", "")); status["pendingSyncJobs"] != float64(0) {
		t.Fatalf("sin permiso no se encola: %v", status)
	}
}

func TestPublication_TwoTenants_APublicationNeverTouchesAnotherBarbershop(t *testing.T) {
	db := setupTestDB(t)
	p := newPublication(t, db, "pub-tenant-a")
	rawB := createSessionCookie(t, db, shopB, staffUserActiveB, "pub-tenant-b")
	t.Cleanup(func() { doMyBarberRequest(p.router, rawB, http.MethodDelete, "") })

	apptID, _ := p.createAppointment(t, futureCivil(15, 10))
	p.drain(t)
	if _, _, ok := p.calendar.eventFor(apptID); !ok {
		t.Fatal("la cita de A se publica")
	}
	// B ni ve la integración de A ni su cola.
	status := gcalJSON(t, doGCal(p.router, rawB, http.MethodGet, "", ""))
	if status["status"] != "not_connected" || status["pendingSyncJobs"] != float64(0) {
		t.Fatalf("B no ve la integración de A: %v", status)
	}
	if rec := doGCal(p.router, rawB, http.MethodPost, "/sync", ""); rec.Code != http.StatusConflict && rec.Code != http.StatusNotFound {
		t.Fatalf("B no puede sincronizar la conexión de A: %d", rec.Code)
	}
}
