package postgres_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"system-barbershop/internal/modules/appointmentnotice"
	noticepostgres "system-barbershop/internal/modules/appointmentnotice/postgres"
	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
)

// Pruebas de integración de la cola de avisos al cliente (HU-130, DEC-142)
// contra PostgreSQL real y DOS barberías: encolado en la transacción del cambio,
// un único aviso pendiente por turno, aislamiento por tenant, y el ciclo del
// worker (reclamo, contexto con el estado vigente, cierre por CAS y evidencia).
// El worker se conecta como barberia_worker y solo puede ejecutar las funciones
// SECURITY DEFINER.

const (
	shopA = database.BarbershopID("11111111-1111-1111-1111-111111111111")
	shopB = database.BarbershopID("22222222-2222-2222-2222-222222222222")
)

func openDB(t *testing.T, envVar, fallback string) *database.DB {
	t.Helper()
	url := os.Getenv(envVar)
	if url == "" {
		url = fallback
	}
	cfg := config.Config{
		Environment: "test", DatabaseMaxConns: 8, DatabaseMinConns: 1,
		DatabaseMaxConnLifetime: time.Hour, DatabaseMaxConnIdleTime: 30 * time.Minute,
		DatabaseConnectTimeout: 5 * time.Second, DatabaseStatementTimeout: 10 * time.Second,
	}
	db, err := database.NewDB(config.DatabaseDSN(url), cfg)
	if err != nil {
		t.Fatalf("database.NewDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func appDB(t *testing.T) *database.DB {
	return openDB(t, "TEST_DATABASE_URL", "postgres://barberia_app@localhost:5432/barberia_test?sslmode=disable")
}

func workerDB(t *testing.T) *database.DB {
	return openDB(t, "TEST_WORKER_DATABASE_URL", "postgres://barberia_worker@localhost:5432/barberia_test?sslmode=disable")
}

func hex8(t *testing.T) string {
	t.Helper()
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(b)
}

// shopFixture es el barbero y el servicio que las pruebas de este paquete comparten
// por barbería. Crear uno nuevo por prueba inflaba el catálogo de la barbería de
// fixtures (otras suites listan solo 50 servicios), así que se crean una vez.
type shopFixture struct{ barberID, serviceID string }

var (
	fixturesMu sync.Mutex
	fixtures   = map[database.BarbershopID]shopFixture{}
)

func sharedFixture(t *testing.T, db *database.DB, shop database.BarbershopID) shopFixture {
	t.Helper()
	fixturesMu.Lock()
	defer fixturesMu.Unlock()
	if f, ok := fixtures[shop]; ok {
		return f
	}
	var f shopFixture
	err := db.InTenantTx(context.Background(), shop, func(ctx context.Context, q database.Queries) error {
		if err := q.QueryRow(ctx, `INSERT INTO barber (barbershop_id, full_name) VALUES ($1, $2) RETURNING id::text`,
			string(shop), "Barbero aviso "+hex8(t)).Scan(&f.barberID); err != nil {
			return err
		}
		return q.QueryRow(ctx,
			`INSERT INTO service (barbershop_id, name, duration_minutes, price_amount) VALUES ($1, $2, 40, 30000) RETURNING id::text`,
			string(shop), "Corte "+hex8(t)).Scan(&f.serviceID)
	})
	if err != nil {
		t.Fatalf("sharedFixture: %v", err)
	}
	fixtures[shop] = f
	return f
}

// seedAppointment crea un cliente (con o sin correo) y una cita confirmada con el
// barbero y el servicio compartidos de la barbería, y devuelve su id y el correo.
// El inicio es aleatorio dentro de un año para que ninguna cita se solape con otra
// del mismo barbero.
func seedAppointment(t *testing.T, db *database.DB, shop database.BarbershopID, withEmail bool) (id, email string) {
	t.Helper()
	f := sharedFixture(t, db, shop)
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	offsetMinutes := int(b[0])<<16 | int(b[1])<<8 | int(b[2]) // hasta ~16,7 M de minutos
	offsetMinutes %= 365 * 24 * 60
	err := db.InTenantTx(context.Background(), shop, func(ctx context.Context, q database.Queries) error {
		var customerID string
		var emailArg *string
		if withEmail {
			email = "cliente." + hex8(t) + "@ejemplo.test"
			emailArg = &email
		}
		if err := q.QueryRow(ctx,
			`INSERT INTO customer (barbershop_id, full_name, email) VALUES ($1, 'Cliente Aviso', $2) RETURNING id::text`,
			string(shop), emailArg).Scan(&customerID); err != nil {
			return err
		}
		start := time.Now().UTC().Add(72*time.Hour + time.Duration(offsetMinutes)*time.Minute).Truncate(time.Minute)
		return q.QueryRow(ctx,
			`INSERT INTO appointment (barbershop_id, barber_id, service_id, customer_id, attendee_name, starts_at, ends_at,
			                          status, origin, service_name_snapshot, duration_minutes_snapshot,
			                          price_amount_snapshot, price_currency_snapshot)
			 VALUES ($1, $2, $3, $4, 'Ana Pérez', $5::timestamptz, $5::timestamptz + interval '40 minutes',
			         'confirmed', 'manual', 'Corte clásico', 40, 30000, 'COP')
			 RETURNING id::text`,
			string(shop), f.barberID, f.serviceID, customerID, start).Scan(&id)
	})
	if err != nil {
		t.Fatalf("seedAppointment: %v", err)
	}
	return id, email
}

func enqueue(t *testing.T, db *database.DB, shop database.BarbershopID, appointmentID, kind string) {
	t.Helper()
	err := db.InTenantTx(context.Background(), shop, func(ctx context.Context, q database.Queries) error {
		return noticepostgres.NewEnqueuer().AppointmentNoticeRequired(ctx, q, string(shop), appointmentID, kind)
	})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
}

type noticeRow struct {
	ID, Kind, Status, Code string
	Attempts               int
}

func notices(t *testing.T, db *database.DB, shop database.BarbershopID, appointmentID string) []noticeRow {
	t.Helper()
	var out []noticeRow
	err := db.InTenantTx(context.Background(), shop, func(ctx context.Context, q database.Queries) error {
		rows, err := q.Query(ctx,
			`SELECT id::text, kind, status, coalesce(outcome_code, ''), attempts
			   FROM appointment_notice WHERE appointment_id = $1 ORDER BY created_at`, appointmentID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var n noticeRow
			if err := rows.Scan(&n.ID, &n.Kind, &n.Status, &n.Code, &n.Attempts); err != nil {
				return err
			}
			out = append(out, n)
		}
		return rows.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// claimMine reclama y devuelve solo el aviso de esta prueba (la base de pruebas
// es compartida): los ajenos se devuelven a la cola con una espera larga.
func claimMine(t *testing.T, w *noticepostgres.WorkerRepository, noticeID string, now time.Time) (appointmentnotice.Notice, bool) {
	t.Helper()
	for range 20 {
		claimed, err := w.ClaimNotices(context.Background(), 50, 60, 5, now)
		if err != nil {
			t.Fatalf("ClaimNotices: %v", err)
		}
		var mine appointmentnotice.Notice
		found := false
		for _, n := range claimed {
			if n.ID == noticeID {
				mine, found = n, true
				continue
			}
			_, _ = w.FinishNotice(context.Background(), appointmentnotice.FinishInput{
				NoticeID: n.ID, ClaimToken: n.ClaimToken, Outcome: appointmentnotice.OutcomeRetry,
				Code: "ajeno", RetrySeconds: 3600, MaxAttempts: 100, Now: now})
		}
		if found {
			return mine, true
		}
		if len(claimed) == 0 {
			return appointmentnotice.Notice{}, false
		}
	}
	return appointmentnotice.Notice{}, false
}

func TestEnqueue_WritesTheNoticeInTheSameTransactionAsTheChange(t *testing.T) {
	app := appDB(t)
	id, _ := seedAppointment(t, app, shopA, true)

	// Si la transacción del cambio se revierte, no queda aviso (CA-130-01).
	err := app.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
		if err := noticepostgres.NewEnqueuer().AppointmentNoticeRequired(ctx, q, string(shopA), id, "rescheduled"); err != nil {
			return err
		}
		return context.Canceled // fuerza el rollback
	})
	if err == nil {
		t.Fatal("la transacción debía fallar")
	}
	if got := notices(t, app, shopA, id); len(got) != 0 {
		t.Fatalf("un cambio revertido no deja aviso: %+v", got)
	}

	enqueue(t, app, shopA, id, "rescheduled")
	got := notices(t, app, shopA, id)
	if len(got) != 1 || got[0].Status != "pending" || got[0].Kind != "rescheduled" {
		t.Fatalf("un cambio confirmado deja un aviso pendiente: %+v", got)
	}
}

func TestEnqueue_ConsecutiveChangesLeaveExactlyOnePendingNotice(t *testing.T) {
	app := appDB(t)
	id, _ := seedAppointment(t, app, shopA, true)

	enqueue(t, app, shopA, id, "rescheduled")
	enqueue(t, app, shopA, id, "rescheduled")
	enqueue(t, app, shopA, id, "cancelled_by_barber") // la cancelación sustituye (CA-130-03)

	got := notices(t, app, shopA, id)
	if len(got) != 1 || got[0].Status != "pending" || got[0].Kind != "cancelled_by_barber" {
		t.Fatalf("debe quedar un solo aviso pendiente, el más reciente: %+v", got)
	}
}

func TestEnqueue_OtherTenantCannotQueueOrSeeTheNotice(t *testing.T) {
	app := appDB(t)
	id, _ := seedAppointment(t, app, shopA, true)

	// Desde el tenant B, el turno de A no existe: no se encola nada (CA-130-09).
	enqueue(t, app, shopB, id, "rescheduled")
	if got := notices(t, app, shopA, id); len(got) != 0 {
		t.Fatalf("B no puede encolar sobre un turno de A: %+v", got)
	}

	enqueue(t, app, shopA, id, "rescheduled")
	if got := notices(t, app, shopB, id); len(got) != 0 {
		t.Fatalf("B no debe ver el aviso de A: %+v", got)
	}
}

func TestWorker_FullCycle_ReadsTheCurrentStateAndLeavesEvidenceWithoutContact(t *testing.T) {
	app, worker := appDB(t), noticepostgres.NewWorkerRepository(workerDB(t))
	ctx := context.Background()
	id, email := seedAppointment(t, app, shopA, true)
	enqueue(t, app, shopA, id, "rescheduled")
	n := notices(t, app, shopA, id)[0]

	// Antes de la espera de 60 s no vence.
	if _, ok := claimMine(t, worker, n.ID, time.Now()); ok {
		t.Fatal("el aviso no debe reclamarse antes de su espera")
	}

	now := time.Now().Add(2 * time.Minute)
	claimed, ok := claimMine(t, worker, n.ID, now)
	if !ok || claimed.Attempts != 1 {
		t.Fatalf("vencida la espera debe reclamarse con intentos=1: %+v ok=%v", claimed, ok)
	}

	c, found, err := worker.NoticeContext(ctx, claimed.ID, claimed.ClaimToken)
	if err != nil || !found {
		t.Fatalf("NoticeContext: found=%v err=%v", found, err)
	}
	if c.Kind != "rescheduled" || c.AppointmentStatus != "confirmed" || c.CustomerEmail != email ||
		c.AttendeeName != "Ana Pérez" || c.ServiceName != "Corte clásico" || c.Timezone == "" || c.AlreadyCovered {
		t.Fatalf("contexto inesperado: %+v", c)
	}
	if _, found, _ := worker.NoticeContext(ctx, claimed.ID, "00000000-0000-4000-8000-000000000000"); found {
		t.Fatal("un token ajeno no debe leer el contexto")
	}

	// CAS: el token vencido no cierra; el vigente sí, y solo una vez.
	ok, err = worker.FinishNotice(ctx, appointmentnotice.FinishInput{
		NoticeID: claimed.ID, ClaimToken: "00000000-0000-4000-8000-000000000000",
		Outcome: appointmentnotice.OutcomeSent, CoversStateAt: c.StateAt, MaxAttempts: 5, Now: now})
	if err != nil || ok {
		t.Fatalf("un token ajeno no debe cerrar: ok=%v err=%v", ok, err)
	}
	fin := appointmentnotice.FinishInput{NoticeID: claimed.ID, ClaimToken: claimed.ClaimToken,
		Outcome: appointmentnotice.OutcomeSent, CoversStateAt: c.StateAt, MaxAttempts: 5, Now: now}
	if ok, err = worker.FinishNotice(ctx, fin); err != nil || !ok {
		t.Fatalf("el token vigente debe cerrar: ok=%v err=%v", ok, err)
	}
	if ok, _ = worker.FinishNotice(ctx, fin); ok {
		t.Fatal("un aviso ya cerrado no se cierra de nuevo")
	}

	if got := notices(t, app, shopA, id); len(got) != 1 || got[0].Status != "sent" || got[0].Attempts != 1 {
		t.Fatalf("el aviso debe quedar sent: %+v", got)
	}

	// Evidencia visible para la barbería y sin datos de contacto (CA-130-05).
	var evidence string
	err = app.InTenantTx(ctx, shopA, func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx, `SELECT row_to_json(t)::text FROM appointment_notice_attempt t WHERE notice_id = $1`, n.ID).Scan(&evidence)
	})
	if err != nil {
		t.Fatalf("la evidencia debe existir: %v", err)
	}
	if strings.Contains(evidence, email) || strings.Contains(evidence, "Ana") || !strings.Contains(evidence, `"outcome":"sent"`) {
		t.Fatalf("evidencia inesperada o con datos de contacto: %s", evidence)
	}
	var otherTenant int
	_ = app.InTenantTx(ctx, shopB, func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx, `SELECT count(*) FROM appointment_notice_attempt WHERE notice_id = $1`, n.ID).Scan(&otherTenant)
	})
	if otherTenant != 0 {
		t.Fatal("B no debe ver la evidencia de A")
	}
}

func TestWorker_CustomerWithoutEmail_ContextHasNoAddress(t *testing.T) {
	app, worker := appDB(t), noticepostgres.NewWorkerRepository(workerDB(t))
	id, _ := seedAppointment(t, app, shopA, false)
	enqueue(t, app, shopA, id, "cancelled_by_barber")
	n := notices(t, app, shopA, id)[0]

	claimed, ok := claimMine(t, worker, n.ID, time.Now().Add(2*time.Minute))
	if !ok {
		t.Fatal("el aviso debe reclamarse")
	}
	c, found, err := worker.NoticeContext(context.Background(), claimed.ID, claimed.ClaimToken)
	if err != nil || !found || c.CustomerEmail != "" {
		t.Fatalf("sin correo el contexto no inventa destinatario: %+v found=%v err=%v", c, found, err)
	}
	// Se cierra como omitido para no dejar el aviso reclamado en la base compartida.
	_, _ = worker.FinishNotice(context.Background(), appointmentnotice.FinishInput{
		NoticeID: claimed.ID, ClaimToken: claimed.ClaimToken, Outcome: appointmentnotice.OutcomeSkipped,
		Code: appointmentnotice.CodeNoEmail, MaxAttempts: 5, Now: time.Now()})
	if got := notices(t, app, shopA, id); got[0].Status != "skipped" || got[0].Code != "no_email" {
		t.Fatalf("debe quedar skipped/no_email: %+v", got)
	}
}

func TestWorker_ExpiredLease_IsReclaimedWithANewTokenAndTheOldOneCannotClose(t *testing.T) {
	app, worker := appDB(t), noticepostgres.NewWorkerRepository(workerDB(t))
	id, _ := seedAppointment(t, app, shopA, true)
	enqueue(t, app, shopA, id, "rescheduled")
	n := notices(t, app, shopA, id)[0]

	t0 := time.Now().Add(2 * time.Minute)
	first, ok := claimMine(t, worker, n.ID, t0)
	if !ok {
		t.Fatal("primer reclamo")
	}
	// Con el lease vigente nadie más lo reclama (CA-130-08).
	if _, again := claimMine(t, worker, n.ID, t0.Add(30*time.Second)); again {
		t.Fatal("un aviso con lease vigente no debe reclamarse")
	}
	second, ok := claimMine(t, worker, n.ID, t0.Add(5*time.Minute))
	if !ok || second.ClaimToken == first.ClaimToken || second.Attempts != 2 {
		t.Fatalf("vencido el lease debe reclamarse con otro token e intentos=2: %+v ok=%v", second, ok)
	}
	if closed, _ := worker.FinishNotice(context.Background(), appointmentnotice.FinishInput{
		NoticeID: first.ID, ClaimToken: first.ClaimToken, Outcome: appointmentnotice.OutcomeSent,
		CoversStateAt: time.Now(), MaxAttempts: 5, Now: t0}); closed {
		t.Fatal("el primer worker ya no puede cerrar el aviso")
	}
	_, _ = worker.FinishNotice(context.Background(), appointmentnotice.FinishInput{
		NoticeID: second.ID, ClaimToken: second.ClaimToken, Outcome: appointmentnotice.OutcomeSkipped,
		Code: appointmentnotice.CodePast, MaxAttempts: 5, Now: t0})
}

type recordingMailer struct {
	calls []string // kind|email|hora
	err   error
}

func (m *recordingMailer) SendChangeNotice(_ context.Context, kind, email, _, _, _, when string) error {
	m.calls = append(m.calls, kind+"|"+email+"|"+when)
	return m.err
}

type laterClock struct{ at time.Time }

func (c laterClock) Now() time.Time { return c.at }

// runDispatcherOnMine ejecuta el despachador real sobre la cola real y devuelve
// solo lo que le tocó a esta prueba: los avisos ajenos de la base compartida se
// devuelven con una espera larga antes de procesar.
func runDispatcherOnMine(t *testing.T, w *noticepostgres.WorkerRepository, noticeID string, mailer *recordingMailer, at time.Time) {
	t.Helper()
	claimed, ok := claimMine(t, w, noticeID, at)
	if !ok {
		t.Fatal("el aviso debe reclamarse")
	}
	appointmentnotice.NewDispatcher(appointmentnotice.Deps{
		Store: w, Mailer: mailer, Clock: laterClock{at: at}, Config: appointmentnotice.DefaultConfig(),
	}).Process(context.Background(), claimed)
}

func TestDispatcher_OverTheRealQueue_SendsTheCurrentStateAndRecordsIt(t *testing.T) {
	app, worker := appDB(t), noticepostgres.NewWorkerRepository(workerDB(t))
	id, email := seedAppointment(t, app, shopA, true)
	enqueue(t, app, shopA, id, "rescheduled")
	n := notices(t, app, shopA, id)[0]

	mailer := &recordingMailer{}
	runDispatcherOnMine(t, worker, n.ID, mailer, time.Now().Add(2*time.Minute))

	if len(mailer.calls) != 1 || !strings.HasPrefix(mailer.calls[0], "rescheduled|"+email+"|") {
		t.Fatalf("debe enviar el aviso al correo del cliente: %v", mailer.calls)
	}
	if got := notices(t, app, shopA, id); got[0].Status != "sent" {
		t.Fatalf("el aviso debe quedar sent: %+v", got)
	}
}

func TestDispatcher_OverTheRealQueue_RescheduleOfACancelledAppointmentIsNotSent(t *testing.T) {
	app, worker := appDB(t), noticepostgres.NewWorkerRepository(workerDB(t))
	id, _ := seedAppointment(t, app, shopA, true)
	enqueue(t, app, shopA, id, "rescheduled")
	n := notices(t, app, shopA, id)[0]

	// El cliente cancela mientras el aviso espera (RN-REC-02).
	err := app.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
		_, err := q.Exec(ctx,
			`UPDATE appointment SET status = 'cancelled_by_customer', resolved_at = now() WHERE barbershop_id = $1 AND id = $2`,
			string(shopA), id)
		return err
	})
	if err != nil {
		t.Fatalf("cancelar el turno: %v", err)
	}

	mailer := &recordingMailer{}
	runDispatcherOnMine(t, worker, n.ID, mailer, time.Now().Add(2*time.Minute))

	if len(mailer.calls) != 0 {
		t.Fatalf("no se avisa una hora vieja de un turno cancelado: %v", mailer.calls)
	}
	if got := notices(t, app, shopA, id); got[0].Status != "skipped" || got[0].Code != "not_confirmed" {
		t.Fatalf("debe quedar skipped/not_confirmed: %+v", got)
	}
}

func TestDispatcher_OverTheRealQueue_TransientFailureRetriesAndPermanentStops(t *testing.T) {
	app, worker := appDB(t), noticepostgres.NewWorkerRepository(workerDB(t))
	id, _ := seedAppointment(t, app, shopA, true)
	enqueue(t, app, shopA, id, "rescheduled")
	n := notices(t, app, shopA, id)[0]

	t0 := time.Now().Add(2 * time.Minute)
	runDispatcherOnMine(t, worker, n.ID, &recordingMailer{err: context.DeadlineExceeded}, t0)
	if got := notices(t, app, shopA, id); got[0].Status != "pending" || got[0].Code != "provider_unavailable" || got[0].Attempts != 1 {
		t.Fatalf("un fallo transitorio vuelve a pending con su causa: %+v", got)
	}

	// Vencida la espera de 60 s se reintenta; esta vez el proveedor lo rechaza.
	runDispatcherOnMine(t, worker, n.ID, &recordingMailer{err: permanent{}}, t0.Add(2*time.Minute))
	got := notices(t, app, shopA, id)
	if got[0].Status != "failed" || got[0].Code != "provider_rejected" || got[0].Attempts != 2 {
		t.Fatalf("un rechazo permanente termina en failed sin más intentos: %+v", got)
	}

	var attempts []string
	err := app.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
		rows, err := q.Query(ctx, `SELECT attempt_no::text || ':' || outcome FROM appointment_notice_attempt WHERE notice_id = $1 ORDER BY attempt_no`, n.ID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var s string
			if err := rows.Scan(&s); err != nil {
				return err
			}
			attempts = append(attempts, s)
		}
		return rows.Err()
	})
	if err != nil || strings.Join(attempts, ",") != "1:retry,2:failed" {
		t.Fatalf("cada intento deja su evidencia: %v err=%v", attempts, err)
	}
}

type permanent struct{}

func (permanent) Error() string   { return "rechazo" }
func (permanent) Permanent() bool { return true }

// ---- HU-131 / DEC-144: recordatorio 30 minutos antes ----

func enqueueReminder(t *testing.T, db *database.DB, shop database.BarbershopID, appointmentID string) {
	t.Helper()
	err := db.InTenantTx(context.Background(), shop, func(ctx context.Context, q database.Queries) error {
		return noticepostgres.NewEnqueuer().AppointmentReminderRequired(ctx, q, string(shop), appointmentID)
	})
	if err != nil {
		t.Fatalf("enqueueReminder: %v", err)
	}
}

// startsAt lee el inicio vigente del turno y, del recordatorio pendiente, cuándo vence.
func startsAt(t *testing.T, db *database.DB, shop database.BarbershopID, appointmentID string) time.Time {
	t.Helper()
	var at time.Time
	err := db.InTenantTx(context.Background(), shop, func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx, `SELECT starts_at FROM appointment WHERE barbershop_id = $1 AND id = $2`,
			string(shop), appointmentID).Scan(&at)
	})
	if err != nil {
		t.Fatal(err)
	}
	return at
}

func reminderRunAt(t *testing.T, db *database.DB, shop database.BarbershopID, appointmentID string) (time.Time, int) {
	t.Helper()
	var runAt time.Time
	var n int
	err := db.InTenantTx(context.Background(), shop, func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx,
			`SELECT coalesce(max(run_at), 'epoch'), count(*) FROM appointment_notice
			  WHERE barbershop_id = $1 AND appointment_id = $2 AND kind = 'reminder' AND status = 'pending'`,
			string(shop), appointmentID).Scan(&runAt, &n)
	})
	if err != nil {
		t.Fatal(err)
	}
	return runAt, n
}

func setStart(t *testing.T, db *database.DB, shop database.BarbershopID, appointmentID string, start time.Time, status string) {
	t.Helper()
	err := db.InTenantTx(context.Background(), shop, func(ctx context.Context, q database.Queries) error {
		if status == "confirmed" {
			_, err := q.Exec(ctx, `UPDATE appointment SET starts_at = $3::timestamptz, ends_at = $3::timestamptz + interval '40 minutes'
			                        WHERE barbershop_id = $1 AND id = $2`, string(shop), appointmentID, start)
			return err
		}
		_, err := q.Exec(ctx, `UPDATE appointment SET status = $3, resolved_at = now() WHERE barbershop_id = $1 AND id = $2`,
			string(shop), appointmentID, status)
		return err
	})
	if err != nil {
		t.Fatalf("setStart: %v", err)
	}
}

func TestReminder_IsDueThirtyMinutesBeforeAndCoexistsWithAChangeNotice(t *testing.T) {
	app := appDB(t)
	id, _ := seedAppointment(t, app, shopA, true)
	start := startsAt(t, app, shopA, id)

	enqueueReminder(t, app, shopA, id)
	runAt, n := reminderRunAt(t, app, shopA, id)
	if n != 1 || !runAt.Equal(start.Add(-30*time.Minute)) {
		t.Fatalf("un recordatorio pendiente a 30 minutos del turno: n=%d run_at=%s start=%s", n, runAt, start)
	}

	// Una reprogramación pendiente no sustituye al recordatorio ni al revés (CA-131-04).
	enqueue(t, app, shopA, id, "rescheduled")
	enqueueReminder(t, app, shopA, id)
	enqueueReminder(t, app, shopA, id)
	var kinds []string
	for _, row := range notices(t, app, shopA, id) {
		kinds = append(kinds, row.Kind+"/"+row.Status)
	}
	if len(kinds) != 2 {
		t.Fatalf("un recordatorio y un aviso de cambio, no más (RN-REC-01): %v", kinds)
	}
}

func TestReminder_IsNotScheduledWhenThereIsNothingToRemind(t *testing.T) {
	app := appDB(t)

	soon, _ := seedAppointment(t, app, shopA, true)
	setStart(t, app, shopA, soon, time.Now().UTC().Add(10*time.Minute), "confirmed")
	enqueueReminder(t, app, shopA, soon)
	if _, n := reminderRunAt(t, app, shopA, soon); n != 0 {
		t.Fatal("un turno a menos de 30 minutos no genera recordatorio")
	}

	cancelled, _ := seedAppointment(t, app, shopA, true)
	setStart(t, app, shopA, cancelled, time.Time{}, "cancelled_by_customer")
	enqueueReminder(t, app, shopA, cancelled)
	if _, n := reminderRunAt(t, app, shopA, cancelled); n != 0 {
		t.Fatal("un turno no confirmado no genera recordatorio (RN-REC-02)")
	}
}

func TestReminder_MovingTheAppointmentRegeneratesOneAndTheOldHourIsNeverSent(t *testing.T) {
	app, worker := appDB(t), noticepostgres.NewWorkerRepository(workerDB(t))
	id, _ := seedAppointment(t, app, shopA, true)
	enqueueReminder(t, app, shopA, id)
	oldRun, _ := reminderRunAt(t, app, shopA, id)
	old := notices(t, app, shopA, id)[0]

	// El turno se mueve 3 horas después y el recordatorio viejo vence antes de regenerarse.
	setStart(t, app, shopA, id, startsAt(t, app, shopA, id).Add(3*time.Hour), "confirmed")
	mailer := &recordingMailer{}
	runDispatcherOnMine(t, worker, old.ID, mailer, oldRun.Add(time.Minute))
	if len(mailer.calls) != 0 {
		t.Fatalf("no se recuerda la hora vieja (RN-REC-03): %v", mailer.calls)
	}
	if got := notices(t, app, shopA, id)[0]; got.Status != "skipped" || got.Code != "superseded" {
		t.Fatalf("el recordatorio de la hora vieja queda skipped/superseded: %+v", got)
	}

	// La reprogramación programa el nuevo: uno solo vigente.
	enqueueReminder(t, app, shopA, id)
	enqueueReminder(t, app, shopA, id)
	newRun, n := reminderRunAt(t, app, shopA, id)
	if n != 1 || !newRun.Equal(oldRun.Add(3*time.Hour)) {
		t.Fatalf("un solo recordatorio vigente a la hora nueva: n=%d run_at=%s", n, newRun)
	}
}

func TestReminder_SentOnceWithTheCurrentHourAndNotRepeatedForTheSameHour(t *testing.T) {
	app, worker := appDB(t), noticepostgres.NewWorkerRepository(workerDB(t))
	id, email := seedAppointment(t, app, shopA, true)
	enqueueReminder(t, app, shopA, id)
	first := notices(t, app, shopA, id)[0]
	runAt, _ := reminderRunAt(t, app, shopA, id)

	mailer := &recordingMailer{}
	runDispatcherOnMine(t, worker, first.ID, mailer, runAt.Add(time.Second))
	if len(mailer.calls) != 1 || !strings.HasPrefix(mailer.calls[0], "reminder|"+email+"|") {
		t.Fatalf("debe enviarse el recordatorio al correo del cliente: %v", mailer.calls)
	}
	if got := notices(t, app, shopA, id)[0]; got.Status != "sent" {
		t.Fatalf("debe quedar sent: %+v", got)
	}

	// Un segundo recordatorio para la misma hora ya está cubierto: no se repite.
	enqueueReminder(t, app, shopA, id)
	rows := notices(t, app, shopA, id)
	second := rows[len(rows)-1]
	runDispatcherOnMine(t, worker, second.ID, mailer, runAt.Add(2*time.Second))
	if len(mailer.calls) != 1 {
		t.Fatalf("no se repite el recordatorio de la misma hora: %v", mailer.calls)
	}
	if got := notices(t, app, shopA, id); got[len(got)-1].Code != "superseded" {
		t.Fatalf("el duplicado queda skipped/superseded: %+v", got)
	}
}

func TestReminder_CancelledWhileWaitingIsNotSent(t *testing.T) {
	app, worker := appDB(t), noticepostgres.NewWorkerRepository(workerDB(t))
	id, _ := seedAppointment(t, app, shopA, true)
	enqueueReminder(t, app, shopA, id)
	n := notices(t, app, shopA, id)[0]
	runAt, _ := reminderRunAt(t, app, shopA, id)

	setStart(t, app, shopA, id, time.Time{}, "cancelled_by_customer")
	mailer := &recordingMailer{}
	runDispatcherOnMine(t, worker, n.ID, mailer, runAt.Add(time.Second))
	if len(mailer.calls) != 0 {
		t.Fatalf("no se recuerda un turno cancelado (RN-REC-02): %v", mailer.calls)
	}
	if got := notices(t, app, shopA, id)[0]; got.Status != "skipped" || got.Code != "not_confirmed" {
		t.Fatalf("debe quedar skipped/not_confirmed: %+v", got)
	}
}

func TestReminder_OtherTenantCannotScheduleForMyAppointment(t *testing.T) {
	app := appDB(t)
	id, _ := seedAppointment(t, app, shopA, true)
	enqueueReminder(t, app, shopB, id) // sesión de B sobre el turno de A: el INSERT…SELECT no ve nada
	if _, n := reminderRunAt(t, app, shopA, id); n != 0 {
		t.Fatal("otra barbería no puede programar recordatorios de un turno ajeno (RN-TEN-01)")
	}
}
