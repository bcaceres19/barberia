package googlecalendar_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"system-barbershop/internal/modules/googlecalendar"
)

// --- dobles -----------------------------------------------------------------

type fakeStore struct {
	mu        sync.Mutex
	jobs      []googlecalendar.Job
	contexts  map[string]googlecalendar.JobContext
	finished  []googlecalendar.FinishInput
	finishOK  bool
	contextOK bool

	restoreTarget *googlecalendar.RestoreTarget
	restoreLinks  []googlecalendar.RestoreLink
	enqueued      []googlecalendar.RestoreLink
}

func newFakeStore() *fakeStore {
	return &fakeStore{contexts: map[string]googlecalendar.JobContext{}, finishOK: true, contextOK: true}
}

func (s *fakeStore) ClaimJobs(context.Context, int, int, time.Time) ([]googlecalendar.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	jobs := s.jobs
	s.jobs = nil
	return jobs, nil
}

func (s *fakeStore) JobContext(_ context.Context, jobID, _ string) (googlecalendar.JobContext, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	jc, ok := s.contexts[jobID]
	return jc, ok && s.contextOK, nil
}

func (s *fakeStore) FinishJob(_ context.Context, in googlecalendar.FinishInput) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.finished = append(s.finished, in)
	return s.finishOK, nil
}

func (s *fakeStore) RestoreClaimConnection(context.Context, int, time.Time) (googlecalendar.RestoreTarget, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.restoreTarget == nil {
		return googlecalendar.RestoreTarget{}, false, nil
	}
	target := *s.restoreTarget
	s.restoreTarget = nil
	return target, true, nil
}

func (s *fakeStore) RestoreLinks(context.Context, string, time.Time) ([]googlecalendar.RestoreLink, error) {
	return s.restoreLinks, nil
}

func (s *fakeStore) EnqueueMissing(_ context.Context, _, resourceType, resourceID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enqueued = append(s.enqueued, googlecalendar.RestoreLink{ResourceType: resourceType, ResourceID: resourceID})
	return true, nil
}

type apiCall struct {
	op       string
	token    string
	calendar string
	eventID  string
	event    googlecalendar.Event
	notify   bool
}

type fakeAPI struct {
	mu    sync.Mutex
	calls []apiCall

	insertErrs []error // se consumen en orden; nil = éxito
	patchErrs  []error
	deleteErr  error
	listIDs    map[string]bool
	listErr    error
}

func (a *fakeAPI) record(c apiCall) {
	a.mu.Lock()
	a.calls = append(a.calls, c)
	a.mu.Unlock()
}

func pop(errs *[]error) error {
	if len(*errs) == 0 {
		return nil
	}
	err := (*errs)[0]
	*errs = (*errs)[1:]
	return err
}

func (a *fakeAPI) Insert(_ context.Context, token, cal string, ev googlecalendar.Event, notify bool) (googlecalendar.Published, error) {
	a.record(apiCall{op: "insert", token: token, calendar: cal, eventID: ev.ID, event: ev, notify: notify})
	a.mu.Lock()
	err := pop(&a.insertErrs)
	a.mu.Unlock()
	if err != nil {
		return googlecalendar.Published{}, err
	}
	return googlecalendar.Published{EventID: ev.ID, ETag: `"etag-1"`}, nil
}

func (a *fakeAPI) Patch(_ context.Context, token, cal, eventID string, ev googlecalendar.Event, notify bool) (googlecalendar.Published, error) {
	a.record(apiCall{op: "patch", token: token, calendar: cal, eventID: eventID, event: ev, notify: notify})
	a.mu.Lock()
	err := pop(&a.patchErrs)
	a.mu.Unlock()
	if err != nil {
		return googlecalendar.Published{}, err
	}
	return googlecalendar.Published{EventID: eventID, ETag: `"etag-2"`}, nil
}

func (a *fakeAPI) Delete(_ context.Context, token, cal, eventID string, notify bool) error {
	a.record(apiCall{op: "delete", token: token, calendar: cal, eventID: eventID, notify: notify})
	return a.deleteErr
}

func (a *fakeAPI) ListEventIDs(context.Context, string, string, string, time.Time) (map[string]bool, error) {
	return a.listIDs, a.listErr
}

func (a *fakeAPI) ops() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]string, len(a.calls))
	for i, c := range a.calls {
		out[i] = c.op
	}
	return out
}

type pubHarness struct {
	pub      *googlecalendar.Publisher
	store    *fakeStore
	api      *fakeAPI
	provider *fakeProvider
	clock    *fakeClock
	cipher   *googlecalendar.Cipher
}

const (
	connID     = "c0000000-0000-4000-8000-0000000000c1"
	apptID     = "a0000000-0000-4000-8000-0000000000a1"
	blockID    = "b0000000-0000-4000-8000-0000000000b1"
	jobID      = "10b00000-0000-4000-8000-000000000001"
	claimToken = "c1a10000-0000-4000-8000-000000000001"
)

func newPubHarness(t *testing.T) *pubHarness {
	t.Helper()
	h := &pubHarness{
		store:    newFakeStore(),
		api:      &fakeAPI{},
		provider: &fakeProvider{},
		clock:    &fakeClock{now: time.Date(2026, 10, 10, 14, 0, 0, 0, time.UTC)},
		cipher:   mustCipher(t, "v1", map[string][]byte{"v1": key(1)}),
	}
	h.pub = googlecalendar.NewPublisher(googlecalendar.PublisherDeps{
		Store: h.store, API: h.api, Provider: h.provider, Cipher: h.cipher, Clock: h.clock,
		Config: googlecalendar.DefaultPublisherConfig(),
	})
	return h
}

// apptContext es una cita confirmada en el futuro con conexión sana, sin vínculo.
func (h *pubHarness) apptContext(t *testing.T) googlecalendar.JobContext {
	t.Helper()
	token, keyID, err := h.cipher.Encrypt([]byte("1//refresh-de-prueba"), []byte("gcal-refresh|"+shopA+"|"+barberA))
	if err != nil {
		t.Fatal(err)
	}
	reminder := 30
	return googlecalendar.JobContext{
		BarberID: barberA, ConnectionStatus: "connected", CalendarID: "primary", ReminderMinutes: &reminder,
		TokenCiphertext: token, TokenKeyID: keyID, Timezone: "America/Bogota",
		ResourceType: googlecalendar.ResourceAppointment, ResourceID: apptID, ResourceFound: true, BarberMatches: true,
		ApptStatus: "confirmed",
		StartsAt:   h.clock.now.Add(24 * time.Hour), EndsAt: h.clock.now.Add(24*time.Hour + 40*time.Minute),
		AttendeeName: "Ana Pérez", ServiceName: "Corte clásico", CustomerEmail: "ana@ejemplo.test",
	}
}

func (h *pubHarness) run(t *testing.T, jc googlecalendar.JobContext, attempts int) googlecalendar.FinishInput {
	t.Helper()
	job := googlecalendar.Job{
		ID: jobID, ClaimToken: claimToken, BarbershopID: shopA, ConnectionID: connID,
		ResourceType: jc.ResourceType, ResourceID: jc.ResourceID, Attempts: attempts,
	}
	h.store.contexts[jobID] = jc
	h.store.finished = nil
	h.pub.Process(context.Background(), job)
	if len(h.store.finished) != 1 {
		t.Fatalf("el trabajo debe finalizarse una vez, hubo %d", len(h.store.finished))
	}
	return h.store.finished[0]
}

// --- publicar ---------------------------------------------------------------

func TestPublish_NewAppointment_CreatesOneEventWithTheAgreedContent(t *testing.T) {
	h := newPubHarness(t)
	jc := h.apptContext(t)

	fin := h.run(t, jc, 1)

	if fin.Outcome != googlecalendar.OutcomePublished || fin.EventID == "" || fin.Generation != 1 {
		t.Fatalf("finalización inesperada: %+v", fin)
	}
	if ops := h.api.ops(); len(ops) != 1 || ops[0] != "insert" {
		t.Fatalf("una cita nueva crea exactamente un evento, ops=%v", ops)
	}
	call := h.api.calls[0]
	if call.token != "access-token-de-prueba" || call.calendar != "primary" {
		t.Fatalf("debe usar el access token renovado y el calendario de la conexión: %+v", call)
	}
	ev := call.event
	if ev.Summary != "Ana Pérez — Corte clásico" {
		t.Fatalf("título = %q (DEC-101: «Nombre — Servicio»)", ev.Summary)
	}
	if ev.Description != "Servicio: Corte clásico\nEstado: Confirmada" {
		t.Fatalf("descripción = %q", ev.Description)
	}
	if !ev.StartsAt.Equal(jc.StartsAt) || !ev.EndsAt.Equal(jc.EndsAt) || ev.TimeZone != "America/Bogota" {
		t.Fatalf("el evento representa el mismo instante en la zona de la barbería: %+v", ev)
	}
	if ev.ReminderMinutes == nil || *ev.ReminderMinutes != 30 {
		t.Fatalf("recordatorio = %v, se esperaba el override de 30 minutos", ev.ReminderMinutes)
	}
	want := map[string]string{
		googlecalendar.PropResourceType: "appointment",
		googlecalendar.PropResourceID:   apptID,
		googlecalendar.PropConnectionID: connID,
	}
	for k, v := range want {
		if ev.Private[k] != v {
			t.Fatalf("propiedad %s = %q, se esperaba %q", k, ev.Private[k], v)
		}
	}
	if ev.ID != googlecalendar.EventID(connID, "appointment", apptID, 1) || fin.EventID != ev.ID {
		t.Fatalf("el id propuesto debe ser determinista: %q vs %q", ev.ID, fin.EventID)
	}
	if ev.AttendeeEmail != "ana@ejemplo.test" || !call.notify {
		t.Fatalf("DEC-122: se invita al cliente y Google le envía la invitación: %+v notify=%v", ev.AttendeeEmail, call.notify)
	}
	// El correo solo viaja como asistente.
	for name, value := range map[string]string{"título": ev.Summary, "descripción": ev.Description} {
		if strings.Contains(value, "ana@ejemplo.test") {
			t.Fatalf("el correo no debe aparecer en el %s", name)
		}
	}
	for k, v := range ev.Private {
		if strings.Contains(v, "ana@ejemplo.test") || strings.Contains(k, "mail") {
			t.Fatalf("el correo no debe aparecer en las propiedades extendidas: %s=%s", k, v)
		}
	}
	if fin.VisibleHash == "" || fin.MaxAttempts != 8 {
		t.Fatalf("debe guardar la huella visible y respetar el máximo de intentos: %+v", fin)
	}
}

func TestPublish_WithoutReminderUsesCalendarDefaults(t *testing.T) {
	h := newPubHarness(t)
	jc := h.apptContext(t)
	jc.ReminderMinutes = nil
	h.run(t, jc, 1)
	if h.api.calls[0].event.ReminderMinutes != nil {
		t.Fatal("sin reminder_minutes se usan los recordatorios predeterminados (useDefault)")
	}
}

func TestPublish_Invitee_OnlyWithEmailAndAHealthyConnection(t *testing.T) {
	cases := map[string]func(jc *googlecalendar.JobContext){
		"sin correo":           func(jc *googlecalendar.JobContext) { jc.CustomerEmail = "" },
		"conexión con error":   func(jc *googlecalendar.JobContext) { jc.ConnectionStatus = "error" },
		"correo solo espacios": func(jc *googlecalendar.JobContext) { jc.CustomerEmail = "   " },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			h := newPubHarness(t)
			jc := h.apptContext(t)
			mutate(&jc)
			h.run(t, jc, 1)
			call := h.api.calls[0]
			if call.event.AttendeeEmail != "" || call.notify {
				t.Fatalf("no se invita ni se notifica: %+v", call)
			}
		})
	}
}

func TestPublish_Reschedule_PatchesTheSameEventAndOnlyNotifiesWhatTheGuestSees(t *testing.T) {
	h := newPubHarness(t)
	jc := h.apptContext(t)
	first := h.run(t, jc, 1)

	// Reprogramar: mismo evento, el invitado sí se entera (cambió el horario).
	jc.LinkEventID, jc.LinkGeneration, jc.LinkVisibleHash = first.EventID, first.Generation, first.VisibleHash
	jc.StartsAt = jc.StartsAt.Add(2 * time.Hour)
	jc.EndsAt = jc.EndsAt.Add(2 * time.Hour)
	moved := h.run(t, jc, 1)
	if ops := h.api.ops(); len(ops) != 2 || ops[1] != "patch" {
		t.Fatalf("reprogramar actualiza el MISMO evento, ops=%v", ops)
	}
	if h.api.calls[1].eventID != first.EventID || moved.EventID != first.EventID {
		t.Fatalf("debe conservar el id del evento: %+v", h.api.calls[1])
	}
	if !h.api.calls[1].notify {
		t.Fatal("cambió el horario que el invitado ve: Google debe avisarle")
	}

	// Cambiar solo el recordatorio del barbero NO notifica al invitado.
	jc.LinkVisibleHash = moved.VisibleHash
	reminder := 10
	jc.ReminderMinutes = &reminder
	h.run(t, jc, 1)
	if h.api.calls[2].op != "patch" || h.api.calls[2].notify {
		t.Fatalf("refrescar el recordatorio no envía correos (DEC-122): %+v", h.api.calls[2])
	}
}

func TestPublish_EventDeletedInGoogle_RecreatesWithANewGeneration(t *testing.T) {
	h := newPubHarness(t)
	jc := h.apptContext(t)
	jc.LinkEventID = googlecalendar.EventID(connID, "appointment", apptID, 1)
	jc.LinkGeneration = 1
	h.api.patchErrs = []error{&googlecalendar.APIError{Status: 404}}

	fin := h.run(t, jc, 1)

	if ops := h.api.ops(); len(ops) != 2 || ops[0] != "patch" || ops[1] != "insert" {
		t.Fatalf("un 404 al actualizar recrea el evento, ops=%v", ops)
	}
	want := googlecalendar.EventID(connID, "appointment", apptID, 2)
	if fin.Outcome != googlecalendar.OutcomePublished || fin.EventID != want || fin.Generation != 2 {
		t.Fatalf("debe recrearse con la generación 2 (Google no reutiliza ids borrados): %+v", fin)
	}
}

func TestPublish_CrashBetweenCreateAndLink_AdoptsTheExistingEventWithoutDuplicating(t *testing.T) {
	h := newPubHarness(t)
	jc := h.apptContext(t)
	h.api.insertErrs = []error{&googlecalendar.APIError{Status: 409}} // el id ya existe

	fin := h.run(t, jc, 2)

	if ops := h.api.ops(); len(ops) != 2 || ops[0] != "insert" || ops[1] != "patch" {
		t.Fatalf("un 409 se adopta con un PATCH del mismo id, ops=%v", ops)
	}
	if fin.Outcome != googlecalendar.OutcomePublished || fin.Generation != 1 ||
		fin.EventID != googlecalendar.EventID(connID, "appointment", apptID, 1) {
		t.Fatalf("debe quedar vinculado el evento ya creado, sin segunda generación: %+v", fin)
	}
}

func TestPublish_ConflictWithADeletedEventID_MovesToTheNextGeneration(t *testing.T) {
	h := newPubHarness(t)
	jc := h.apptContext(t)
	h.api.insertErrs = []error{&googlecalendar.APIError{Status: 409}}
	h.api.patchErrs = []error{&googlecalendar.APIError{Status: 410}} // el id es de un evento borrado

	fin := h.run(t, jc, 1)

	if fin.Outcome != googlecalendar.OutcomePublished || fin.Generation != 2 {
		t.Fatalf("debe pasar a la generación 2: %+v", fin)
	}
}

// --- retirar y conservar ----------------------------------------------------

func TestRemove_CancelledAppointment_DeletesTheEventAndNotifiesTheGuest(t *testing.T) {
	for _, status := range []string{"cancelled_by_customer", "cancelled_by_barber"} {
		t.Run(status, func(t *testing.T) {
			h := newPubHarness(t)
			jc := h.apptContext(t)
			jc.ApptStatus = status
			jc.LinkEventID = "navaevento1"
			jc.LinkGeneration = 1

			fin := h.run(t, jc, 1)

			if ops := h.api.ops(); len(ops) != 1 || ops[0] != "delete" || h.api.calls[0].eventID != "navaevento1" {
				t.Fatalf("cancelar elimina el evento vinculado, ops=%v", ops)
			}
			if !h.api.calls[0].notify {
				t.Fatal("el cliente invitado recibe la cancelación de Google (DEC-122)")
			}
			if fin.Outcome != googlecalendar.OutcomeRemoved {
				t.Fatalf("outcome = %s", fin.Outcome)
			}
		})
	}
}

func TestRemove_AlreadyGoneInGoogle_IsSuccess(t *testing.T) {
	h := newPubHarness(t)
	jc := h.apptContext(t)
	jc.ApptStatus = "cancelled_by_barber"
	jc.LinkEventID = "navaevento1"
	h.api.deleteErr = &googlecalendar.APIError{Status: 410}

	if fin := h.run(t, jc, 1); fin.Outcome != googlecalendar.OutcomeRemoved {
		t.Fatalf("un 404/410 durante la cancelación es éxito: %+v", fin)
	}
}

func TestNoLink_CancelledOrTerminalOrPast_DoesNothingInGoogle(t *testing.T) {
	cases := map[string]func(jc *googlecalendar.JobContext){
		"cancelada sin vínculo": func(jc *googlecalendar.JobContext) { jc.ApptStatus = "cancelled_by_barber" },
		"completada":            func(jc *googlecalendar.JobContext) { jc.ApptStatus = "completed" },
		"no_show":               func(jc *googlecalendar.JobContext) { jc.ApptStatus = "no_show" },
		"confirmada ya pasada": func(jc *googlecalendar.JobContext) {
			jc.StartsAt = jc.StartsAt.Add(-72 * time.Hour)
			jc.EndsAt = jc.EndsAt.Add(-72 * time.Hour)
		},
		"conexión desconectada":    func(jc *googlecalendar.JobContext) { jc.ConnectionStatus = "disconnected" },
		"conexión reauth_required": func(jc *googlecalendar.JobContext) { jc.ConnectionStatus = "reauth_required" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			h := newPubHarness(t)
			jc := h.apptContext(t)
			mutate(&jc)
			fin := h.run(t, jc, 1)
			if len(h.api.calls) != 0 || fin.Outcome != googlecalendar.OutcomeSkipped {
				t.Fatalf("no debe llamar a Google ni crear nada: ops=%v outcome=%s", h.api.ops(), fin.Outcome)
			}
		})
	}
}

func TestTerminalStates_KeepTheExistingEventUntouched(t *testing.T) {
	for _, status := range []string{"completed", "no_show"} {
		t.Run(status, func(t *testing.T) {
			h := newPubHarness(t)
			jc := h.apptContext(t)
			jc.ApptStatus = status
			jc.LinkEventID = "navaevento1"
			fin := h.run(t, jc, 1)
			if len(h.api.calls) != 0 || fin.Outcome != googlecalendar.OutcomeSkipped {
				t.Fatalf("%s conserva el evento sin tocarlo: ops=%v outcome=%s", status, h.api.ops(), fin.Outcome)
			}
		})
	}
}

func TestMissingOrForeignResource_RemovesTheLinkedEvent(t *testing.T) {
	for name, mutate := range map[string]func(jc *googlecalendar.JobContext){
		"recurso inexistente":  func(jc *googlecalendar.JobContext) { jc.ResourceFound = false },
		"cita de otro barbero": func(jc *googlecalendar.JobContext) { jc.BarberMatches = false },
	} {
		t.Run(name, func(t *testing.T) {
			h := newPubHarness(t)
			jc := h.apptContext(t)
			jc.LinkEventID = "navaevento1"
			mutate(&jc)
			if fin := h.run(t, jc, 1); fin.Outcome != googlecalendar.OutcomeRemoved {
				t.Fatalf("outcome = %s", fin.Outcome)
			}
		})
	}
}

// --- bloqueos ---------------------------------------------------------------

func blockContext(h *pubHarness, t *testing.T) googlecalendar.JobContext {
	jc := h.apptContext(t)
	jc.ResourceType, jc.ResourceID = googlecalendar.ResourceTimeBlock, blockID
	jc.ApptStatus, jc.AttendeeName, jc.ServiceName, jc.CustomerEmail = "", "", "", ""
	jc.BlockType, jc.BlockSource = "lunch", "manual"
	return jc
}

func TestBlocks_ManualBlockPublishesAnEventAndNeverInvitesAnyone(t *testing.T) {
	h := newPubHarness(t)
	fin := h.run(t, blockContext(h, t), 1)
	if fin.Outcome != googlecalendar.OutcomePublished {
		t.Fatalf("outcome = %s", fin.Outcome)
	}
	ev := h.api.calls[0].event
	if ev.Summary != "Almuerzo" || ev.AttendeeEmail != "" || h.api.calls[0].notify {
		t.Fatalf("un bloqueo se titula por su tipo y no invita a nadie: %+v", ev)
	}
	if ev.Private[googlecalendar.PropResourceType] != "time_block" {
		t.Fatalf("propiedades = %v", ev.Private)
	}
}

func TestBlocks_RemovedOrNotPublishable_DeletesTheEventOrSkips(t *testing.T) {
	h := newPubHarness(t)
	jc := blockContext(h, t)
	jc.BlockDeleted = true
	jc.LinkEventID = "navabloque1"
	if fin := h.run(t, jc, 1); fin.Outcome != googlecalendar.OutcomeRemoved || h.api.calls[0].notify {
		t.Fatalf("un bloqueo retirado elimina su evento sin notificar: %+v %+v", fin, h.api.calls)
	}

	for _, blockType := range []string{"holiday", "inventado"} {
		h2 := newPubHarness(t)
		jc := blockContext(h2, t)
		jc.BlockType = blockType
		if fin := h2.run(t, jc, 1); fin.Outcome != googlecalendar.OutcomeSkipped || len(h2.api.calls) != 0 {
			t.Fatalf("%s no se publica: %+v", blockType, fin)
		}
	}
	h3 := newPubHarness(t)
	jc = blockContext(h3, t)
	jc.BlockSource = "holiday_calendar"
	if fin := h3.run(t, jc, 1); fin.Outcome != googlecalendar.OutcomeSkipped || len(h3.api.calls) != 0 {
		t.Fatalf("un bloqueo del calendario automático no se publica: %+v", fin)
	}
}

// --- fallos de Google -------------------------------------------------------

func TestFailures_AreTranslatedToTheRightOutcome(t *testing.T) {
	cases := map[string]struct {
		err      error
		outcome  string
		code     string
		minRetry int
	}{
		"503":                 {&googlecalendar.APIError{Status: 503}, googlecalendar.OutcomeRetry, "google_unavailable", 30},
		"429 con Retry-After": {&googlecalendar.APIError{Status: 429, RetryAfter: 10 * time.Minute}, googlecalendar.OutcomeRetry, "google_unavailable", 600},
		"403 de cuota":        {&googlecalendar.APIError{Status: 403, Reason: "rateLimitExceeded"}, googlecalendar.OutcomeRetry, "google_unavailable", 30},
		"403 sin permiso":     {&googlecalendar.APIError{Status: 403, Reason: "forbidden"}, googlecalendar.OutcomePermanent, "permission_denied", 0},
		"404 calendario":      {&googlecalendar.APIError{Status: 404}, googlecalendar.OutcomePermanent, "calendar_not_found", 0},
		"400 evento inválido": {&googlecalendar.APIError{Status: 400}, googlecalendar.OutcomePermanent, "invalid_event", 0},
		"tiempo agotado":      {context.DeadlineExceeded, googlecalendar.OutcomeRetry, "transient", 30},
		"error de red":        {errors.New("connection reset"), googlecalendar.OutcomeRetry, "transient", 30},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			h := newPubHarness(t)
			h.api.insertErrs = []error{tc.err}
			fin := h.run(t, h.apptContext(t), 1)
			if fin.Outcome != tc.outcome || fin.ErrorCode != tc.code {
				t.Fatalf("outcome=%s code=%s, se esperaba %s/%s", fin.Outcome, fin.ErrorCode, tc.outcome, tc.code)
			}
			if tc.outcome == googlecalendar.OutcomeRetry && fin.RetrySeconds < tc.minRetry {
				t.Fatalf("espera = %d s, mínimo %d", fin.RetrySeconds, tc.minRetry)
			}
		})
	}
}

func TestFailures_BackoffGrowsExponentiallyAndIsCapped(t *testing.T) {
	want := []time.Duration{30 * time.Second, time.Minute, 2 * time.Minute, 4 * time.Minute, 8 * time.Minute,
		16 * time.Minute, 32 * time.Minute, time.Hour, time.Hour}
	for i, expected := range want {
		if got := googlecalendar.Backoff(i + 1); got != expected {
			t.Fatalf("Backoff(%d) = %v, se esperaba %v", i+1, got, expected)
		}
	}
	if googlecalendar.Backoff(0) != 30*time.Second || googlecalendar.Backoff(100) != time.Hour {
		t.Fatal("los extremos deben quedar acotados")
	}

	h := newPubHarness(t)
	h.api.insertErrs = []error{&googlecalendar.APIError{Status: 503}}
	if fin := h.run(t, h.apptContext(t), 4); fin.RetrySeconds != 240 {
		t.Fatalf("el intento 4 espera 4 minutos, esperó %d s", fin.RetrySeconds)
	}
}

func TestGoogleDown_NeverLosesTheAppointmentOrLoopsForever(t *testing.T) {
	h := newPubHarness(t)
	h.api.insertErrs = []error{&googlecalendar.APIError{Status: 503}}
	fin := h.run(t, h.apptContext(t), 8)
	// El publicador pide reintentar; es la base (gcal_finish_job) quien, al
	// alcanzar MaxAttempts, deja el trabajo `failed` en vez de reintentar sin fin.
	if fin.Outcome != googlecalendar.OutcomeRetry || fin.MaxAttempts != 8 {
		t.Fatalf("debe pedir reintento acotado por MaxAttempts: %+v", fin)
	}
}

func TestToken_RevokedOrMissingCredentials_RequiresReauthWithoutCallingGoogle(t *testing.T) {
	t.Run("Google revocó el refresh token", func(t *testing.T) {
		h := newPubHarness(t)
		h.provider.refreshErr = googlecalendar.ErrTokenRevoked
		fin := h.run(t, h.apptContext(t), 1)
		if fin.Outcome != googlecalendar.OutcomeReauth || fin.ErrorCode != "token_revoked" || len(h.api.calls) != 0 {
			t.Fatalf("%+v ops=%v", fin, h.api.ops())
		}
	})
	t.Run("sin credenciales guardadas", func(t *testing.T) {
		h := newPubHarness(t)
		jc := h.apptContext(t)
		jc.TokenCiphertext, jc.TokenKeyID = nil, ""
		if fin := h.run(t, jc, 1); fin.Outcome != googlecalendar.OutcomeReauth || len(h.api.calls) != 0 {
			t.Fatalf("%+v", fin)
		}
	})
	t.Run("refresh transitorio", func(t *testing.T) {
		h := newPubHarness(t)
		h.provider.refreshErr = errors.New("503 del token endpoint")
		if fin := h.run(t, h.apptContext(t), 1); fin.Outcome != googlecalendar.OutcomeRetry {
			t.Fatalf("un fallo del refresh no es una revocación: %+v", fin)
		}
	})
	t.Run("clave retirada del anillo", func(t *testing.T) {
		h := newPubHarness(t)
		jc := h.apptContext(t)
		jc.TokenKeyID = "clave-perdida"
		if fin := h.run(t, jc, 1); fin.Outcome != googlecalendar.OutcomeRetry {
			t.Fatalf("no poder descifrar es un error de configuración, no una revocación: %+v", fin)
		}
	})
}

func TestToken_ExpiredAccessToken_ForcesOneRefreshAndRetries(t *testing.T) {
	h := newPubHarness(t)
	h.api.insertErrs = []error{&googlecalendar.APIError{Status: 401}}
	fin := h.run(t, h.apptContext(t), 1)
	if fin.Outcome != googlecalendar.OutcomePublished {
		t.Fatalf("un 401 con token caducado se resuelve renovando: %+v", fin)
	}
	if h.provider.refreshed != 2 || len(h.api.calls) != 2 {
		t.Fatalf("debe renovar una vez más y reintentar: refrescos=%d llamadas=%d", h.provider.refreshed, len(h.api.calls))
	}

	h2 := newPubHarness(t)
	h2.api.insertErrs = []error{&googlecalendar.APIError{Status: 401}, &googlecalendar.APIError{Status: 401}}
	if fin := h2.run(t, h2.apptContext(t), 1); fin.Outcome != googlecalendar.OutcomeReauth {
		t.Fatalf("un 401 persistente significa que el permiso ya no sirve: %+v", fin)
	}
}

func TestToken_IsCachedAcrossJobsOfTheSameConnection(t *testing.T) {
	h := newPubHarness(t)
	h.run(t, h.apptContext(t), 1)
	h.run(t, h.apptContext(t), 1)
	if h.provider.refreshed != 1 {
		t.Fatalf("el access token vigente se reutiliza entre trabajos: %d renovaciones", h.provider.refreshed)
	}
}

// --- reclamación ------------------------------------------------------------

func TestClaim_LostLease_WritesNothing(t *testing.T) {
	h := newPubHarness(t)
	job := googlecalendar.Job{ID: jobID, ClaimToken: claimToken, BarbershopID: shopA, ConnectionID: connID,
		ResourceType: "appointment", ResourceID: apptID, Attempts: 1}
	h.store.contexts[jobID] = h.apptContext(t)
	h.store.contextOK = false // otro worker lo retomó

	h.pub.Process(context.Background(), job)

	if len(h.store.finished) != 0 || len(h.api.calls) != 0 {
		t.Fatal("sin la reclamación vigente no se llama a Google ni se finaliza")
	}
}

func TestRunOnce_ProcessesTheClaimedBatch(t *testing.T) {
	h := newPubHarness(t)
	h.store.jobs = []googlecalendar.Job{{ID: jobID, ClaimToken: claimToken, BarbershopID: shopA, ConnectionID: connID,
		ResourceType: "appointment", ResourceID: apptID, Attempts: 1}}
	h.store.contexts[jobID] = h.apptContext(t)

	if n := h.pub.RunOnce(context.Background()); n != 1 {
		t.Fatalf("procesó %d trabajos, se esperaba 1", n)
	}
	if len(h.store.finished) != 1 || h.store.finished[0].Outcome != googlecalendar.OutcomePublished {
		t.Fatalf("finalizaciones = %+v", h.store.finished)
	}
}

// --- restauración de eventos borrados ---------------------------------------

func (h *pubHarness) restoreTarget(t *testing.T) *googlecalendar.RestoreTarget {
	jc := h.apptContext(t)
	return &googlecalendar.RestoreTarget{ConnectionID: connID, BarbershopID: shopA, BarberID: barberA, CalendarID: "primary",
		TokenCiphertext: jc.TokenCiphertext, TokenKeyID: jc.TokenKeyID}
}

func TestRestore_RequeuesOnlyTheLinkedResourcesWhoseEventIsMissing(t *testing.T) {
	h := newPubHarness(t)
	h.store.restoreTarget = h.restoreTarget(t)
	h.store.restoreLinks = []googlecalendar.RestoreLink{
		{ResourceType: "appointment", ResourceID: "a-presente", EventID: "evento-presente"},
		{ResourceType: "appointment", ResourceID: "a-borrada", EventID: "evento-borrado"},
		{ResourceType: "time_block", ResourceID: "b-borrado", EventID: "evento-bloque-borrado"},
	}
	h.api.listIDs = map[string]bool{"evento-presente": true, "evento-ajeno": true}

	if !h.pub.RestoreOnce(context.Background()) {
		t.Fatal("había una conexión por comprobar")
	}
	if len(h.store.enqueued) != 2 || h.store.enqueued[0].ResourceID != "a-borrada" || h.store.enqueued[1].ResourceID != "b-borrado" {
		t.Fatalf("solo se reencolan los eventos que faltan: %+v", h.store.enqueued)
	}
	if len(h.api.calls) != 0 {
		t.Fatal("el chequeo solo lista; recrear lo hace el trabajo encolado")
	}
}

func TestRestore_NothingDueGoogleFailureOrNoToken_DoesNotRequeue(t *testing.T) {
	h := newPubHarness(t)
	if h.pub.RestoreOnce(context.Background()) {
		t.Fatal("sin conexiones vencidas no hay nada que comprobar")
	}

	h.store.restoreTarget = h.restoreTarget(t)
	h.store.restoreLinks = []googlecalendar.RestoreLink{{ResourceType: "appointment", ResourceID: "a1", EventID: "e1"}}
	h.api.listErr = errors.New("503")
	h.pub.RestoreOnce(context.Background())
	if len(h.store.enqueued) != 0 {
		t.Fatal("si no se pudo listar no se asume que faltan eventos")
	}

	h2 := newPubHarness(t)
	h2.store.restoreTarget = h2.restoreTarget(t)
	h2.store.restoreLinks = h.store.restoreLinks
	h2.provider.refreshErr = googlecalendar.ErrTokenRevoked
	h2.pub.RestoreOnce(context.Background())
	if len(h2.store.enqueued) != 0 {
		t.Fatal("sin token válido no se reencola nada")
	}
}

func TestEventID_IsDeterministicAndUsesOnlyBase32HexCharacters(t *testing.T) {
	a := googlecalendar.EventID(connID, "appointment", apptID, 1)
	if a != googlecalendar.EventID(connID, "appointment", apptID, 1) {
		t.Fatal("el id debe ser determinista")
	}
	if a == googlecalendar.EventID(connID, "appointment", apptID, 2) || a == googlecalendar.EventID(connID, "time_block", apptID, 1) {
		t.Fatal("la generación y el tipo cambian el id")
	}
	if len(a) < 5 || len(a) > 1024 {
		t.Fatalf("largo %d fuera de lo que Google admite", len(a))
	}
	for _, r := range a {
		if !(r >= '0' && r <= '9') && !(r >= 'a' && r <= 'v') {
			t.Fatalf("carácter %q fuera de base32hex (0-9, a-v)", r)
		}
	}
}

func TestRun_StopsWhenTheContextEnds(t *testing.T) {
	h := newPubHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		h.pub.Run(ctx)
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run debe terminar cuando el contexto termina")
	}
}

func TestFinish_LostCASDoesNotCrashAndIsReportedOnlyAsAWarning(t *testing.T) {
	h := newPubHarness(t)
	h.store.finishOK = false // otro worker ya finalizó este trabajo
	fin := h.run(t, h.apptContext(t), 1)
	if fin.Outcome != googlecalendar.OutcomePublished {
		t.Fatalf("el publicador igual calcula su resultado: %+v", fin)
	}
}
