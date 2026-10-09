package postgres_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"system-barbershop/internal/modules/googlecalendar"
	gcalpostgres "system-barbershop/internal/modules/googlecalendar/postgres"
	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
)

// Pruebas de integración de la cola de publicación (issue #324, DEC-102,
// DDL-CON-01) contra PostgreSQL real y DOS barberías: encolado en la misma
// transacción, un único trabajo pendiente por recurso, reclamo con lease (SKIP
// LOCKED) entre varios workers, finalización por CAS sobre claim_token y el
// aislamiento por tenant. El worker se conecta como barberia_worker y solo puede
// ejecutar las funciones SECURITY DEFINER.

func setupWorkerDB(t *testing.T) *database.DB {
	t.Helper()
	url := os.Getenv("TEST_WORKER_DATABASE_URL")
	if url == "" {
		url = "postgres://barberia_worker@localhost:5432/barberia_test?sslmode=disable"
	}
	cfg := config.Config{
		Environment: "test", DatabaseMaxConns: 8, DatabaseMinConns: 1,
		DatabaseMaxConnLifetime: time.Hour, DatabaseMaxConnIdleTime: 30 * time.Minute,
		DatabaseConnectTimeout: 5 * time.Second, DatabaseStatementTimeout: 10 * time.Second,
	}
	db, err := database.NewDB(config.DatabaseDSN(url), cfg)
	if err != nil {
		t.Fatalf("database.NewDB (worker): %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

type syncFixture struct {
	shop         database.BarbershopID
	app          *database.DB
	repo         *gcalpostgres.Repository
	worker       *gcalpostgres.WorkerRepository
	barberID     string
	connectionID string
	staffUserID  string
}

// newSyncFixture crea un barbero con la conexión ya `connected`.
func newSyncFixture(t *testing.T, app *database.DB, worker *database.DB, shop database.BarbershopID, staffUserID string) syncFixture {
	t.Helper()
	f := newFixtures(t, app, shop, staffUserID)
	repo := gcalpostgres.New(app)
	conn, err := repo.SaveConnected(context.Background(), string(shop), f.barberID, connectedData(time.Now().UTC()))
	if err != nil {
		t.Fatalf("SaveConnected: %v", err)
	}
	return syncFixture{shop: shop, app: app, repo: repo, worker: gcalpostgres.NewWorkerRepository(worker),
		barberID: f.barberID, connectionID: conn.ID, staffUserID: staffUserID}
}

// seedAppointment crea una cita confirmada y la ENCOLA con el hook, igual que lo
// hace booking: dentro de la misma transacción de la escritura.
func (f syncFixture) seedAppointment(t *testing.T, status string, startsIn time.Duration) string {
	t.Helper()
	var id string
	err := f.app.InTenantTx(context.Background(), f.shop, func(ctx context.Context, q database.Queries) error {
		var serviceID, customerID string
		if err := q.QueryRow(ctx,
			`INSERT INTO service (barbershop_id, name, duration_minutes, price_amount) VALUES ($1, $2, 40, 30000)
			 RETURNING id::text`, string(f.shop), "Corte "+randomHex(t, 4)).Scan(&serviceID); err != nil {
			return err
		}
		if err := q.QueryRow(ctx,
			`INSERT INTO customer (barbershop_id, full_name, email) VALUES ($1, $2, $3) RETURNING id::text`,
			string(f.shop), "Cliente Prueba", "cliente."+randomHex(t, 4)+"@ejemplo.test").Scan(&customerID); err != nil {
			return err
		}
		start := time.Now().UTC().Add(startsIn).Truncate(time.Minute)
		resolved := "NULL"
		if status != "confirmed" {
			resolved = "now()"
		}
		if err := q.QueryRow(ctx,
			`INSERT INTO appointment (barbershop_id, barber_id, service_id, customer_id, attendee_name, starts_at, ends_at,
			                          status, resolved_at, origin, service_name_snapshot, duration_minutes_snapshot,
			                          price_amount_snapshot, price_currency_snapshot)
			 VALUES ($1, $2, $3, $4, 'Ana Pérez', $5::timestamptz, $5::timestamptz + interval '40 minutes', $6, `+resolved+`, 'manual', 'Corte clásico', 40, 30000, 'COP')
			 RETURNING id::text`,
			string(f.shop), f.barberID, serviceID, customerID, start, status).Scan(&id); err != nil {
			return err
		}
		return gcalpostgres.NewEnqueuer().AppointmentChanged(ctx, q, string(f.shop), id)
	})
	if err != nil {
		t.Fatalf("seedAppointment: %v", err)
	}
	return id
}

func (f syncFixture) jobs(t *testing.T) []string {
	t.Helper()
	var statuses []string
	err := f.app.InTenantTx(context.Background(), f.shop, func(ctx context.Context, q database.Queries) error {
		rows, err := q.Query(ctx,
			`SELECT status FROM google_calendar_sync_job WHERE connection_id = $1 ORDER BY created_at`, f.connectionID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var s string
			if err := rows.Scan(&s); err != nil {
				return err
			}
			statuses = append(statuses, s)
		}
		return rows.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
	return statuses
}

// claimMine reclama y devuelve solo los trabajos de esta conexión (la base de
// pruebas es compartida con otras pruebas y tenants).
func (f syncFixture) claimMine(t *testing.T, limit int, now time.Time) []googlecalendar.Job {
	t.Helper()
	var mine []googlecalendar.Job
	for range 20 { // la base compartida puede tener trabajos de otras pruebas por delante
		jobs, err := f.worker.ClaimJobs(context.Background(), limit, 60, now)
		if err != nil {
			t.Fatalf("ClaimJobs: %v", err)
		}
		for _, j := range jobs {
			if j.ConnectionID == f.connectionID {
				mine = append(mine, j)
			} else {
				// No es de esta prueba: se libera para no interferir con otras.
				_, _ = f.worker.FinishJob(context.Background(), googlecalendar.FinishInput{
					JobID: j.ID, ClaimToken: j.ClaimToken, Outcome: googlecalendar.OutcomeRetry,
					ErrorCode: "ajeno", RetrySeconds: 1, MaxAttempts: 100, Now: now})
			}
		}
		if len(jobs) == 0 || len(mine) > 0 {
			break
		}
	}
	return mine
}

func TestEnqueue_AppointmentWriteQueuesOneJobAndFoldsRepeatedChanges(t *testing.T) {
	app, worker := setupTestDB(t), setupWorkerDB(t)
	f := newSyncFixture(t, app, worker, shopC, userC1)
	id := f.seedAppointment(t, "confirmed", 48*time.Hour)

	if got := f.jobs(t); len(got) != 1 || got[0] != "pending" {
		t.Fatalf("una cita confirmada de un barbero conectado encola un trabajo pendiente: %v", got)
	}

	// Cinco cambios seguidos del mismo recurso se funden en UN trabajo.
	for range 5 {
		if err := app.InTenantTx(context.Background(), f.shop, func(ctx context.Context, q database.Queries) error {
			return gcalpostgres.NewEnqueuer().AppointmentChanged(ctx, q, string(f.shop), id)
		}); err != nil {
			t.Fatal(err)
		}
	}
	if got := f.jobs(t); len(got) != 1 {
		t.Fatalf("los cambios repetidos se funden en un solo trabajo pendiente: %v", got)
	}
}

func TestEnqueue_NothingWithoutALiveConnection(t *testing.T) {
	app, worker := setupTestDB(t), setupWorkerDB(t)
	f := newSyncFixture(t, app, worker, shopC, userC1)
	ctx := context.Background()

	// Sin conexión viva (desconectada o con permiso revocado) no se encola.
	if _, err := f.repo.MarkStatus(ctx, string(f.shop), f.barberID, googlecalendar.StatusDisconnected, "", time.Now()); err != nil {
		t.Fatal(err)
	}
	f.seedAppointment(t, "confirmed", 48*time.Hour)
	if got := f.jobs(t); len(got) != 0 {
		t.Fatalf("una conexión desconectada no encola nada: %v", got)
	}

	// Un barbero que nunca conectó tampoco.
	other := newFixtures(t, app, shopC, userC1)
	_ = other
	var count int
	_ = app.InTenantTx(ctx, shopC, func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx, `SELECT count(*) FROM google_calendar_sync_job WHERE barbershop_id = $1 AND connection_id NOT IN
			(SELECT id FROM google_calendar_connection WHERE barbershop_id = $1)`, string(shopC)).Scan(&count)
	})
	if count != 0 {
		t.Fatal("ningún trabajo huérfano")
	}
}

func TestEnqueue_TimeBlockRules(t *testing.T) {
	app, worker := setupTestDB(t), setupWorkerDB(t)
	f := newSyncFixture(t, app, worker, shopC, userC1)

	add := func(blockType, source string) {
		t.Helper()
		err := app.InTenantTx(context.Background(), f.shop, func(ctx context.Context, q database.Queries) error {
			var id string
			if err := q.QueryRow(ctx,
				`INSERT INTO time_block (barbershop_id, barber_id, block_type, source, starts_at, ends_at)
				 VALUES ($1, $2, $3, $4, now() + interval '2 days', now() + interval '2 days 1 hour') RETURNING id::text`,
				string(f.shop), f.barberID, blockType, source).Scan(&id); err != nil {
				return err
			}
			return gcalpostgres.NewEnqueuer().TimeBlockChanged(ctx, q, string(f.shop), id)
		})
		if err != nil {
			t.Fatalf("time block %s/%s: %v", blockType, source, err)
		}
	}
	add("lunch", "manual")
	add("holiday", "holiday_calendar")
	add("holiday", "manual")

	if got := f.jobs(t); len(got) != 1 {
		t.Fatalf("solo el bloqueo manual de un tipo publicable se encola (los festivos no): %v", got)
	}
}

func TestClaim_LeaseSkipLockedAndTokenRotation(t *testing.T) {
	app, worker := setupTestDB(t), setupWorkerDB(t)
	f := newSyncFixture(t, app, worker, shopC, userC1)
	f.seedAppointment(t, "confirmed", 48*time.Hour)
	now := time.Now().UTC().Add(5 * time.Second) // los trabajos nacen con run_at = now() de la base: «ahora» debe ser posterior

	first := f.claimMine(t, 50, now)
	if len(first) != 1 || first[0].Attempts != 1 || first[0].ClaimToken == "" {
		t.Fatalf("debe reclamarse una vez con el contador en 1: %+v", first)
	}
	// Mientras el lease está vigente, nadie más lo reclama.
	if again := f.claimMine(t, 50, now.Add(30*time.Second)); len(again) != 0 {
		t.Fatalf("un trabajo con lease vigente no se reclama de nuevo: %+v", again)
	}
	// Vencido el lease (worker caído) otro lo retoma con un token NUEVO.
	retaken := f.claimMine(t, 50, now.Add(2*time.Minute))
	if len(retaken) != 1 || retaken[0].ID != first[0].ID || retaken[0].ClaimToken == first[0].ClaimToken || retaken[0].Attempts != 2 {
		t.Fatalf("con el lease vencido se reclama con otro token: %+v vs %+v", retaken, first)
	}

	// El primer worker ya no puede finalizar: su claim_token dejó de ser vigente (CAS).
	ok, err := f.worker.FinishJob(context.Background(), googlecalendar.FinishInput{
		JobID: first[0].ID, ClaimToken: first[0].ClaimToken, Outcome: googlecalendar.OutcomeSkipped, Now: now})
	if err != nil || ok {
		t.Fatalf("el worker con la reclamación perdida no debe finalizar: ok=%v err=%v", ok, err)
	}
	if _, found, _ := f.worker.JobContext(context.Background(), first[0].ID, first[0].ClaimToken); found {
		t.Fatal("el contexto exige la reclamación vigente")
	}
	ok, err = f.worker.FinishJob(context.Background(), googlecalendar.FinishInput{
		JobID: retaken[0].ID, ClaimToken: retaken[0].ClaimToken, Outcome: googlecalendar.OutcomeSkipped, Now: now})
	if err != nil || !ok {
		t.Fatalf("quien tiene la reclamación vigente sí finaliza: ok=%v err=%v", ok, err)
	}
	if got := f.jobs(t); len(got) != 0 {
		t.Fatalf("un trabajo terminado se borra: %v", got)
	}
}

func TestClaim_SeveralWorkersNeverTakeTheSameJob(t *testing.T) {
	app, worker := setupTestDB(t), setupWorkerDB(t)
	f := newSyncFixture(t, app, worker, shopC, userC1)
	for i := range 12 {
		f.seedAppointment(t, "confirmed", time.Duration(48+i)*time.Hour)
	}
	now := time.Now().UTC().Add(5 * time.Second) // los trabajos nacen con run_at = now() de la base: «ahora» debe ser posterior

	var mu sync.Mutex
	seen := map[string]int{}
	var wg sync.WaitGroup
	for range 6 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			jobs, err := gcalpostgres.NewWorkerRepository(worker).ClaimJobs(context.Background(), 50, 60, now) // lote grande: todos compiten
			if err != nil {
				t.Errorf("claim: %v", err)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			for _, j := range jobs {
				seen[j.ID]++
			}
		}()
	}
	wg.Wait()
	for id, n := range seen {
		if n != 1 {
			t.Fatalf("el trabajo %s fue reclamado %d veces (SKIP LOCKED debe evitarlo)", id, n)
		}
	}
}

func TestContext_ReturnsTheMinimumNeededAndOnlyForAClaimedJob(t *testing.T) {
	app, worker := setupTestDB(t), setupWorkerDB(t)
	f := newSyncFixture(t, app, worker, shopC, userC1)
	apptID := f.seedAppointment(t, "confirmed", 48*time.Hour)

	jobs := f.claimMine(t, 50, time.Now().UTC().Add(5*time.Second))
	if len(jobs) != 1 {
		t.Fatalf("jobs = %+v", jobs)
	}
	jc, found, err := f.worker.JobContext(context.Background(), jobs[0].ID, jobs[0].ClaimToken)
	if err != nil || !found {
		t.Fatalf("JobContext: found=%v err=%v", found, err)
	}
	if jc.ResourceType != "appointment" || jc.ResourceID != apptID || !jc.ResourceFound || !jc.BarberMatches ||
		jc.ApptStatus != "confirmed" || jc.AttendeeName != "Ana Pérez" || jc.ServiceName != "Corte clásico" ||
		jc.CustomerEmail == "" || jc.Timezone == "" || jc.CalendarID != "primary" || jc.ConnectionStatus != "connected" ||
		jc.BarberID != f.barberID || len(jc.TokenCiphertext) == 0 || jc.TokenKeyID != "v1" || jc.LinkEventID != "" {
		t.Fatalf("contexto inesperado: %+v", jc)
	}
	if jc.EndsAt.Sub(jc.StartsAt) != 40*time.Minute {
		t.Fatalf("horario = %v - %v", jc.StartsAt, jc.EndsAt)
	}
	// Con un token ajeno (aunque el id sea real) no se obtiene nada.
	if _, found, _ := f.worker.JobContext(context.Background(), jobs[0].ID, "00000000-0000-4000-8000-000000000000"); found {
		t.Fatal("un claim_token distinto no debe revelar la cita")
	}
}

func TestFinish_Outcomes(t *testing.T) {
	app, worker := setupTestDB(t), setupWorkerDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Add(5 * time.Second) // los trabajos nacen con run_at = now() de la base: «ahora» debe ser posterior

	t.Run("published crea el vínculo, borra el trabajo y marca la sincronización", func(t *testing.T) {
		f := newSyncFixture(t, app, worker, shopC, userC1)
		apptID := f.seedAppointment(t, "confirmed", 48*time.Hour)
		job := f.claimMine(t, 50, now)[0]
		ok, err := f.worker.FinishJob(ctx, googlecalendar.FinishInput{JobID: job.ID, ClaimToken: job.ClaimToken,
			Outcome: googlecalendar.OutcomePublished, EventID: "navaevento0001", ETag: `"e1"`, Generation: 1,
			VisibleHash: "hash-1", Now: now})
		if err != nil || !ok {
			t.Fatalf("published: ok=%v err=%v", ok, err)
		}
		if got := f.jobs(t); len(got) != 0 {
			t.Fatalf("el trabajo debe borrarse: %v", got)
		}
		// El siguiente trabajo ve el vínculo.
		f.seedAppointmentReenqueue(t, apptID)
		next := f.claimMine(t, 50, now)[0]
		jc, _, _ := f.worker.JobContext(ctx, next.ID, next.ClaimToken)
		if jc.LinkEventID != "navaevento0001" || jc.LinkGeneration != 1 || jc.LinkVisibleHash != "hash-1" {
			t.Fatalf("el vínculo persistido decide entre crear y actualizar: %+v", jc)
		}
		conn, _, _ := f.repo.GetConnection(ctx, string(f.shop), f.barberID)
		if conn.LastSyncedAt == nil {
			t.Fatal("published debe fijar last_synced_at")
		}
	})

	t.Run("removed borra el vínculo", func(t *testing.T) {
		f := newSyncFixture(t, app, worker, shopC, userC1)
		apptID := f.seedAppointment(t, "confirmed", 48*time.Hour)
		job := f.claimMine(t, 50, now)[0]
		_, _ = f.worker.FinishJob(ctx, googlecalendar.FinishInput{JobID: job.ID, ClaimToken: job.ClaimToken,
			Outcome: googlecalendar.OutcomePublished, EventID: "navaevento0002", Generation: 1, Now: now})
		f.seedAppointmentReenqueue(t, apptID)
		next := f.claimMine(t, 50, now)[0]
		if ok, err := f.worker.FinishJob(ctx, googlecalendar.FinishInput{JobID: next.ID, ClaimToken: next.ClaimToken,
			Outcome: googlecalendar.OutcomeRemoved, Now: now}); err != nil || !ok {
			t.Fatalf("removed: ok=%v err=%v", ok, err)
		}
		f.seedAppointmentReenqueue(t, apptID)
		last := f.claimMine(t, 50, now)[0]
		if jc, _, _ := f.worker.JobContext(ctx, last.ID, last.ClaimToken); jc.LinkEventID != "" {
			t.Fatalf("el vínculo debe estar cerrado para que no se recree: %+v", jc)
		}
	})

	t.Run("retry reprograma con backoff y al agotar intentos queda failed", func(t *testing.T) {
		f := newSyncFixture(t, app, worker, shopC, userC1)
		f.seedAppointment(t, "confirmed", 48*time.Hour)
		job := f.claimMine(t, 50, now)[0]
		if ok, err := f.worker.FinishJob(ctx, googlecalendar.FinishInput{JobID: job.ID, ClaimToken: job.ClaimToken,
			Outcome: googlecalendar.OutcomeRetry, ErrorCode: "google_unavailable", RetrySeconds: 600, MaxAttempts: 3, Now: now}); err != nil || !ok {
			t.Fatalf("retry: ok=%v err=%v", ok, err)
		}
		if got := f.claimMine(t, 50, now.Add(5*time.Minute)); len(got) != 0 {
			t.Fatalf("antes del backoff no se reclama: %+v", got)
		}
		second := f.claimMine(t, 50, now.Add(11*time.Minute))
		if len(second) != 1 || second[0].Attempts != 2 {
			t.Fatalf("pasado el backoff se reclama: %+v", second)
		}
		_, _ = f.worker.FinishJob(ctx, googlecalendar.FinishInput{JobID: second[0].ID, ClaimToken: second[0].ClaimToken,
			Outcome: googlecalendar.OutcomeRetry, ErrorCode: "x", RetrySeconds: 60, MaxAttempts: 3, Now: now.Add(11 * time.Minute)})
		third := f.claimMine(t, 50, now.Add(13*time.Minute))
		if len(third) != 1 || third[0].Attempts != 3 {
			t.Fatalf("tercer intento: %+v", third)
		}
		_, _ = f.worker.FinishJob(ctx, googlecalendar.FinishInput{JobID: third[0].ID, ClaimToken: third[0].ClaimToken,
			Outcome: googlecalendar.OutcomeRetry, ErrorCode: "google_unavailable", RetrySeconds: 60, MaxAttempts: 3, Now: now.Add(13 * time.Minute)})
		if got := f.jobs(t); len(got) != 1 || got[0] != "failed" {
			t.Fatalf("agotados los intentos queda failed, sin bucle infinito: %v", got)
		}
		if again := f.claimMine(t, 50, now.Add(24*time.Hour)); len(again) != 0 {
			t.Fatalf("un trabajo failed no se reclama: %+v", again)
		}
	})

	t.Run("reauth borra credenciales y vacía la cola de la conexión", func(t *testing.T) {
		f := newSyncFixture(t, app, worker, shopC, userC1)
		f.seedAppointment(t, "confirmed", 48*time.Hour)
		f.seedAppointment(t, "confirmed", 72*time.Hour)
		job := f.claimMine(t, 1, now)[0]
		if ok, err := f.worker.FinishJob(ctx, googlecalendar.FinishInput{JobID: job.ID, ClaimToken: job.ClaimToken,
			Outcome: googlecalendar.OutcomeReauth, ErrorCode: "token_revoked", Now: now}); err != nil || !ok {
			t.Fatalf("reauth: ok=%v err=%v", ok, err)
		}
		conn, _, _ := f.repo.GetConnection(ctx, string(f.shop), f.barberID)
		if conn.Status != googlecalendar.StatusReauthRequired || conn.HasCredentials() || conn.LastErrorCode != "token_revoked" {
			t.Fatalf("conexión = %+v", conn)
		}
		if got := f.jobs(t); len(got) != 0 {
			t.Fatalf("la cola de una conexión sin permiso se vacía: %v", got)
		}
	})

	t.Run("permanent marca el trabajo failed y la conexión en error con el token intacto", func(t *testing.T) {
		f := newSyncFixture(t, app, worker, shopC, userC1)
		f.seedAppointment(t, "confirmed", 48*time.Hour)
		job := f.claimMine(t, 50, now)[0]
		if ok, err := f.worker.FinishJob(ctx, googlecalendar.FinishInput{JobID: job.ID, ClaimToken: job.ClaimToken,
			Outcome: googlecalendar.OutcomePermanent, ErrorCode: "calendar_not_found", Now: now}); err != nil || !ok {
			t.Fatalf("permanent: ok=%v err=%v", ok, err)
		}
		conn, _, _ := f.repo.GetConnection(ctx, string(f.shop), f.barberID)
		if conn.Status != googlecalendar.StatusError || !conn.HasCredentials() || conn.LastErrorCode != "calendar_not_found" {
			t.Fatalf("conexión = %+v", conn)
		}
		if got := f.jobs(t); len(got) != 1 || got[0] != "failed" {
			t.Fatalf("jobs = %v", got)
		}
	})
}

// seedAppointmentReenqueue vuelve a encolar una cita existente (otro cambio).
func (f syncFixture) seedAppointmentReenqueue(t *testing.T, apptID string) {
	t.Helper()
	if err := f.app.InTenantTx(context.Background(), f.shop, func(ctx context.Context, q database.Queries) error {
		return gcalpostgres.NewEnqueuer().AppointmentChanged(ctx, q, string(f.shop), apptID)
	}); err != nil {
		t.Fatal(err)
	}
}

func TestTwoTenants_JobsAndContextNeverCross(t *testing.T) {
	app, worker := setupTestDB(t), setupWorkerDB(t)
	c := newSyncFixture(t, app, worker, shopC, userC1)
	d := newSyncFixture(t, app, worker, shopD, userD1)
	c.seedAppointment(t, "confirmed", 48*time.Hour)
	d.seedAppointment(t, "confirmed", 49*time.Hour)

	if got := c.jobs(t); len(got) != 1 {
		t.Fatalf("C solo ve su cola: %v", got)
	}
	// Desde la sesión de D la cola de C es invisible (RLS) y no se puede alterar.
	var visible int
	_ = app.InTenantTx(context.Background(), shopD, func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx, `SELECT count(*) FROM google_calendar_sync_job WHERE connection_id = $1`, c.connectionID).Scan(&visible)
	})
	if visible != 0 {
		t.Fatal("D no debe ver los trabajos de C")
	}
	if _, found, _ := d.repo.RequeueConnection(context.Background(), string(shopD), c.barberID, time.Now()); found {
		t.Fatal("D no debe poder reordenar la cola del barbero de C")
	}
	// Cada trabajo reclamado trae SU barbería y su conexión.
	// Reclamar los de una barbería libera los ajenos con una espera corta: cada barbería
	// reclama en un instante posterior para no depender de esa espera.
	for i, f := range []syncFixture{c, d} {
		jobs := f.claimMine(t, 50, time.Now().UTC().Add(5*time.Second+time.Duration(i)*time.Hour))
		if len(jobs) != 1 || jobs[0].BarbershopID != string(f.shop) {
			t.Fatalf("el trabajo debe pertenecer a su barbería: %+v", jobs)
		}
	}
}

func TestWorkerRole_HasNoDirectTableAccess(t *testing.T) {
	worker := setupWorkerDB(t)
	for _, table := range []string{"google_calendar_connection", "google_calendar_sync_job", "google_calendar_event_link", "appointment", "customer"} {
		err := worker.CallSecurityDefinerRow(context.Background(), `SELECT count(*) FROM `+table, nil, func(row pgx.Row) error {
			var n int
			return row.Scan(&n)
		})
		if err == nil {
			t.Fatalf("el worker no debe poder leer %s directamente (DEC-040)", table)
		}
	}
}

func TestRestore_ListsOnlyLiveFutureLinksAndRequeuesIdempotently(t *testing.T) {
	app, worker := setupTestDB(t), setupWorkerDB(t)
	f := newSyncFixture(t, app, worker, shopC, userC1)
	ctx := context.Background()
	now := time.Now().UTC().Add(5 * time.Second) // los trabajos nacen con run_at = now() de la base: «ahora» debe ser posterior

	confirmed := f.seedAppointment(t, "confirmed", 48*time.Hour)
	cancelled := f.seedAppointment(t, "confirmed", 72*time.Hour)
	completed := f.seedAppointment(t, "confirmed", 96*time.Hour)
	past := f.seedAppointment(t, "confirmed", 120*time.Hour)

	// Publica las cuatro (vínculos) y luego cambia su estado en la base.
	for _, id := range []string{confirmed, cancelled, completed, past} {
		f.seedAppointmentReenqueue(t, id)
	}
	for range 10 {
		jobs := f.claimMine(t, 50, now)
		if len(jobs) == 0 {
			break
		}
		for _, j := range jobs {
			_, _ = f.worker.FinishJob(ctx, googlecalendar.FinishInput{JobID: j.ID, ClaimToken: j.ClaimToken,
				Outcome: googlecalendar.OutcomePublished, EventID: "nava" + j.ResourceID[:8], Generation: 1, Now: now})
		}
	}
	err := app.InTenantTx(ctx, f.shop, func(ctx context.Context, q database.Queries) error {
		if _, err := q.Exec(ctx, `UPDATE appointment SET status = 'cancelled_by_barber', resolved_at = now() WHERE id = $1`, cancelled); err != nil {
			return err
		}
		if _, err := q.Exec(ctx, `UPDATE appointment SET status = 'completed', resolved_at = now() WHERE id = $1`, completed); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	links, err := f.worker.RestoreLinks(ctx, f.connectionID, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 2 {
		t.Fatalf("solo la confirmada y la futura siguen vigentes (cancelada y terminal no se restauran): %+v", links)
	}
	got := map[string]bool{}
	for _, l := range links {
		got[l.ResourceID] = true
	}
	if !got[confirmed] || !got[past] {
		t.Fatalf("vínculos = %+v", links)
	}
	// Una cita ya pasada no se restaura.
	later, _ := f.worker.RestoreLinks(ctx, f.connectionID, now.Add(200*time.Hour))
	if len(later) != 0 {
		t.Fatalf("lo pasado no se restaura: %+v", later)
	}

	if queued, err := f.worker.EnqueueMissing(ctx, f.connectionID, "appointment", confirmed); err != nil || !queued {
		t.Fatalf("EnqueueMissing: queued=%v err=%v", queued, err)
	}
	if queued, _ := f.worker.EnqueueMissing(ctx, f.connectionID, "appointment", confirmed); queued {
		t.Fatal("reencolar lo ya pendiente no duplica el trabajo")
	}
	if got := f.jobs(t); len(got) != 1 {
		t.Fatalf("un solo trabajo pendiente: %v", got)
	}
}

func TestRestore_ClaimsEachDueConnectionOnceAndSkipsLiveOnes(t *testing.T) {
	app, worker := setupTestDB(t), setupWorkerDB(t)
	f := newSyncFixture(t, app, worker, shopC, userC1)
	now := time.Now().UTC().Add(5 * time.Second) // los trabajos nacen con run_at = now() de la base: «ahora» debe ser posterior

	var mine *googlecalendar.RestoreTarget
	for range 50 {
		target, found, err := f.worker.RestoreClaimConnection(context.Background(), 300, now)
		if err != nil {
			t.Fatal(err)
		}
		if !found {
			break
		}
		if target.ConnectionID == f.connectionID {
			mine = &target
			break
		}
	}
	if mine == nil || mine.BarbershopID != string(f.shop) || mine.BarberID != f.barberID || len(mine.TokenCiphertext) == 0 {
		t.Fatalf("debe reservarse la conexión pendiente de comprobar: %+v", mine)
	}
	// Recién reservada, no vuelve a salir hasta que pase el intervalo.
	for range 50 {
		target, found, _ := f.worker.RestoreClaimConnection(context.Background(), 300, now.Add(time.Minute))
		if !found {
			break
		}
		if target.ConnectionID == f.connectionID {
			t.Fatal("una conexión comprobada hace un minuto no debe reservarse de nuevo")
		}
	}
}

func TestSyncNow_RequeuesFailedAndMakesPendingDueAndIsSafeToRepeat(t *testing.T) {
	app, worker := setupTestDB(t), setupWorkerDB(t)
	f := newSyncFixture(t, app, worker, shopC, userC1)
	ctx := context.Background()
	now := time.Now().UTC().Add(5 * time.Second) // los trabajos nacen con run_at = now() de la base: «ahora» debe ser posterior

	f.seedAppointment(t, "confirmed", 48*time.Hour)
	f.seedAppointment(t, "confirmed", 72*time.Hour)
	jobs := f.claimMine(t, 50, now)
	// Uno falla definitivamente, el otro queda en espera lejana.
	_, _ = f.worker.FinishJob(ctx, googlecalendar.FinishInput{JobID: jobs[0].ID, ClaimToken: jobs[0].ClaimToken,
		Outcome: googlecalendar.OutcomePermanent, ErrorCode: "calendar_not_found", Now: now})
	_, _ = f.worker.FinishJob(ctx, googlecalendar.FinishInput{JobID: jobs[1].ID, ClaimToken: jobs[1].ClaimToken,
		Outcome: googlecalendar.OutcomeRetry, ErrorCode: "google_unavailable", RetrySeconds: 3600, MaxAttempts: 8, Now: now})

	counts, err := f.repo.JobCounts(ctx, string(f.shop), f.barberID)
	if err != nil || counts.Pending != 1 || counts.Failed != 1 {
		t.Fatalf("antes: %+v %v", counts, err)
	}
	for range 3 { // varios clics
		counts, found, err := f.repo.RequeueConnection(ctx, string(f.shop), f.barberID, now)
		if err != nil || !found || counts.Pending != 2 || counts.Failed != 0 {
			t.Fatalf("Sincronizar ahora: %+v found=%v err=%v", counts, found, err)
		}
	}
	conn, _, _ := f.repo.GetConnection(ctx, string(f.shop), f.barberID)
	if conn.Status != googlecalendar.StatusConnected {
		t.Fatalf("reintentar saca la conexión del estado de error: %v", conn.Status)
	}
	if due := f.claimMine(t, 50, now); len(due) != 2 {
		t.Fatalf("ambos vencen ya: %+v", due)
	}
}

func TestDisconnectAndInitialPublication_DropAndRebuildTheQueue(t *testing.T) {
	app, worker := setupTestDB(t), setupWorkerDB(t)
	f := newSyncFixture(t, app, worker, shopC, userC1)
	ctx := context.Background()
	now := time.Now().UTC().Add(5 * time.Second) // los trabajos nacen con run_at = now() de la base: «ahora» debe ser posterior

	f.seedAppointment(t, "confirmed", 48*time.Hour)
	if got := f.jobs(t); len(got) != 1 {
		t.Fatalf("jobs = %v", got)
	}
	if _, err := f.repo.MarkStatus(ctx, string(f.shop), f.barberID, googlecalendar.StatusDisconnected, "", now); err != nil {
		t.Fatal(err)
	}
	if got := f.jobs(t); len(got) != 0 {
		t.Fatalf("desconectar detiene los trabajos pendientes de la conexión: %v", got)
	}

	// Reconectar publica lo vigente de los próximos 6 meses y nada de lo pasado ni lejano.
	f.seedAppointment(t, "confirmed", 24*time.Hour)           // dentro de la ventana
	f.seedAppointment(t, "confirmed", 24*30*8*time.Hour)      // a 8 meses: fuera
	f.seedAppointment(t, "cancelled_by_barber", 30*time.Hour) // cancelada: no se publica
	if _, err := f.repo.SaveConnected(ctx, string(f.shop), f.barberID, connectedData(now)); err != nil {
		t.Fatal(err)
	}
	pending := f.jobs(t)
	// La cita de 48 h, la de 24 h (la de 8 meses y la cancelada quedan fuera).
	if len(pending) != 2 {
		t.Fatalf("la publicación inicial encola solo lo vigente de los próximos 6 meses: %v", pending)
	}
}

func TestReminderChange_RequeuesTheFutureLinkedEvents(t *testing.T) {
	app, worker := setupTestDB(t), setupWorkerDB(t)
	f := newSyncFixture(t, app, worker, shopC, userC1)
	ctx := context.Background()
	now := time.Now().UTC().Add(5 * time.Second) // los trabajos nacen con run_at = now() de la base: «ahora» debe ser posterior

	apptID := f.seedAppointment(t, "confirmed", 48*time.Hour)
	job := f.claimMine(t, 50, now)[0]
	_, _ = f.worker.FinishJob(ctx, googlecalendar.FinishInput{JobID: job.ID, ClaimToken: job.ClaimToken,
		Outcome: googlecalendar.OutcomePublished, EventID: "navaevento0003", Generation: 1, Now: now})
	if got := f.jobs(t); len(got) != 0 {
		t.Fatalf("jobs = %v", got)
	}

	minutes := 15
	if _, found, err := f.repo.SetReminder(ctx, string(f.shop), f.barberID, &minutes); err != nil || !found {
		t.Fatalf("SetReminder: found=%v err=%v", found, err)
	}
	if got := f.jobs(t); len(got) != 1 {
		t.Fatalf("cambiar el recordatorio actualiza los eventos futuros ya publicados: %v (cita %s)", got, apptID)
	}
}
