package postgres_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"sync"
	"testing"
	"time"

	"system-barbershop/internal/modules/googlecalendar"
	gcalpostgres "system-barbershop/internal/modules/googlecalendar/postgres"
	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
)

// Pruebas de integración contra PostgreSQL real y DOS barberías (shopC/shopD de
// testdata/hu021_barberos.sql): RLS, claves foráneas compuestas, unicidad y el
// CHECK que liga las credenciales al estado son lo que aísla y protege esta
// integración, así que nada de esto se prueba con dobles.

const (
	testDatabaseURL = "postgres://barberia_app@localhost:5432/barberia_test?sslmode=disable"

	shopC = database.BarbershopID("33333333-3333-3333-3333-333333333333")
	shopD = database.BarbershopID("44444444-4444-4444-4444-444444444444")

	userC1 = "cccccc01-cccc-4ccc-8ccc-cccccccccc01"
	userD1 = "dddddd01-dddd-4ddd-8ddd-dddddddddd01"
)

func setupTestDB(t *testing.T) *database.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = testDatabaseURL
	}
	cfg := config.Config{
		Environment:              "test",
		DatabaseMaxConns:         10,
		DatabaseMinConns:         2,
		DatabaseMaxConnLifetime:  time.Hour,
		DatabaseMaxConnIdleTime:  30 * time.Minute,
		DatabaseConnectTimeout:   5 * time.Second,
		DatabaseStatementTimeout: 10 * time.Second,
	}
	db, err := database.NewDB(config.DatabaseDSN(url), cfg)
	if err != nil {
		t.Fatalf("database.NewDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func randomHex(t *testing.T, n int) string {
	t.Helper()
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(buf)
}

// fixtures crea un barbero y una sesión de la barbería indicada (barberia_app no
// puede DELETE sobre barber, así que cada prueba usa los suyos con nombre único).
type fixtures struct {
	barberID  string
	sessionID string
}

func newFixtures(t *testing.T, db *database.DB, shop database.BarbershopID, staffUserID string) fixtures {
	t.Helper()
	var f fixtures
	err := db.InTenantTx(context.Background(), shop, func(ctx context.Context, q database.Queries) error {
		if err := q.QueryRow(ctx,
			`INSERT INTO barber (barbershop_id, full_name) VALUES ($1, $2) RETURNING id::text`,
			string(shop), "Gcal "+randomHex(t, 6)).Scan(&f.barberID); err != nil {
			return err
		}
		return q.QueryRow(ctx,
			`INSERT INTO staff_session (barbershop_id, staff_user_id, token_hash, expires_at)
			 VALUES ($1, $2, $3, now() + interval '1 hour') RETURNING id::text`,
			string(shop), staffUserID, randomHex(t, 32)).Scan(&f.sessionID)
	})
	if err != nil {
		t.Fatalf("fixtures: %v", err)
	}
	return f
}

func linkBarber(t *testing.T, db *database.DB, shop database.BarbershopID, barberID, staffUserID string) {
	t.Helper()
	err := db.InTenantTx(context.Background(), shop, func(ctx context.Context, q database.Queries) error {
		// El usuario es como máximo un barbero: se libera cualquier vínculo previo.
		if _, err := q.Exec(ctx, `UPDATE barber SET staff_user_id = NULL WHERE barbershop_id = $1 AND staff_user_id = $2`,
			string(shop), staffUserID); err != nil {
			return err
		}
		_, err := q.Exec(ctx, `UPDATE barber SET staff_user_id = $3 WHERE barbershop_id = $1 AND id = $2`,
			string(shop), barberID, staffUserID)
		return err
	})
	if err != nil {
		t.Fatalf("linkBarber: %v", err)
	}
}

func connectedData(at time.Time) googlecalendar.ConnectedData {
	return googlecalendar.ConnectedData{
		AccountEmail: "barbero@ejemplo.test", Credentials: []byte("texto-cifrado-de-prueba"), KeyID: "v1", At: at,
	}
}

func stateFor(shop database.BarbershopID, f fixtures, staffUserID string, now time.Time) googlecalendar.OAuthState {
	return googlecalendar.OAuthState{
		BarbershopID: string(shop), BarberID: f.barberID, StaffUserID: staffUserID, SessionID: f.sessionID,
		StateHash: randomHexString(), VerifierEncrypted: []byte("verificador-cifrado"), VerifierKeyID: "v1",
		ExpiresAt: now.Add(10 * time.Minute),
	}
}

func randomHexString() string {
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func TestSaveConnected_CreatesThenReusesTheRowWithoutDuplicating(t *testing.T) {
	db := setupTestDB(t)
	repo := gcalpostgres.New(db)
	f := newFixtures(t, db, shopC, userC1)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	if _, found, err := repo.GetConnection(ctx, string(shopC), f.barberID); err != nil || found {
		t.Fatalf("antes de conectar no hay fila: found=%v err=%v", found, err)
	}

	first, err := repo.SaveConnected(ctx, string(shopC), f.barberID, connectedData(now))
	if err != nil || first.Status != googlecalendar.StatusConnected || first.AccountEmail != "barbero@ejemplo.test" {
		t.Fatalf("SaveConnected: %+v err=%v", first, err)
	}
	if first.CalendarID != googlecalendar.CalendarPrimary || string(first.Credentials) != "texto-cifrado-de-prueba" || first.KeyID != "v1" {
		t.Fatalf("datos inesperados: %+v", first)
	}

	again := connectedData(now.Add(time.Hour))
	again.AccountEmail = "otra@ejemplo.test"
	second, err := repo.SaveConnected(ctx, string(shopC), f.barberID, again)
	if err != nil || second.ID != first.ID {
		t.Fatalf("reconectar reutiliza la misma fila: %+v err=%v", second, err)
	}
	if second.AccountEmail != "otra@ejemplo.test" {
		t.Fatalf("la cuenta nueva reemplaza a la anterior: %+v", second)
	}
}

func TestSetReminder_PersistsClearsAndRejectsOutOfRangeAtTheDatabase(t *testing.T) {
	db := setupTestDB(t)
	repo := gcalpostgres.New(db)
	f := newFixtures(t, db, shopC, userC1)
	ctx := context.Background()

	if _, found, err := repo.SetReminder(ctx, string(shopC), f.barberID, ptr(30)); err != nil || found {
		t.Fatalf("sin conexión no hay recordatorio que guardar: found=%v err=%v", found, err)
	}
	if _, err := repo.SaveConnected(ctx, string(shopC), f.barberID, connectedData(time.Now().UTC())); err != nil {
		t.Fatal(err)
	}

	conn, found, err := repo.SetReminder(ctx, string(shopC), f.barberID, ptr(40320))
	if err != nil || !found || conn.ReminderMinutes == nil || *conn.ReminderMinutes != 40320 {
		t.Fatalf("SetReminder(40320): %+v found=%v err=%v", conn, found, err)
	}
	conn, _, _ = repo.SetReminder(ctx, string(shopC), f.barberID, nil)
	if conn.ReminderMinutes != nil {
		t.Fatalf("nil debe volver a los recordatorios predeterminados: %v", *conn.ReminderMinutes)
	}
	// El CHECK de la base es la última defensa aunque el servicio ya valida.
	if _, _, err := repo.SetReminder(ctx, string(shopC), f.barberID, ptr(40321)); err == nil {
		t.Fatal("la base debe rechazar un recordatorio mayor de 40320")
	}
	if _, _, err := repo.SetReminder(ctx, string(shopC), f.barberID, ptr(-1)); err == nil {
		t.Fatal("la base debe rechazar un recordatorio negativo")
	}
}

func ptr(v int) *int { return &v }

func TestMarkStatus_DropsCredentialsWhenTheTokenIsNoLongerValid(t *testing.T) {
	db := setupTestDB(t)
	repo := gcalpostgres.New(db)
	f := newFixtures(t, db, shopC, userC1)
	ctx := context.Background()
	now := time.Now().UTC()

	if found, err := repo.MarkStatus(ctx, string(shopC), f.barberID, googlecalendar.StatusDisconnected, "", now); err != nil || found {
		t.Fatalf("sin conexión no hay nada que marcar: found=%v err=%v", found, err)
	}

	if _, err := repo.SaveConnected(ctx, string(shopC), f.barberID, connectedData(now)); err != nil {
		t.Fatal(err)
	}
	if found, err := repo.MarkStatus(ctx, string(shopC), f.barberID, googlecalendar.StatusError, "google_503", now); err != nil || !found {
		t.Fatalf("error transitorio: found=%v err=%v", found, err)
	}
	conn, _, _ := repo.GetConnection(ctx, string(shopC), f.barberID)
	if conn.Status != googlecalendar.StatusError || !conn.HasCredentials() || conn.LastErrorCode != "google_503" {
		t.Fatalf("un error transitorio conserva las credenciales: %+v", conn)
	}

	if _, err := repo.MarkStatus(ctx, string(shopC), f.barberID, googlecalendar.StatusReauthRequired, "token_revoked", now); err != nil {
		t.Fatal(err)
	}
	conn, _, _ = repo.GetConnection(ctx, string(shopC), f.barberID)
	if conn.Status != googlecalendar.StatusReauthRequired || conn.HasCredentials() || conn.KeyID != "" {
		t.Fatalf("reauth_required borra las credenciales: %+v", conn)
	}
	// Un error transitorio no puede «resucitar» una conexión sin credenciales.
	if found, _ := repo.MarkStatus(ctx, string(shopC), f.barberID, googlecalendar.StatusError, "x", now); found {
		t.Fatal("no se pasa a error desde reauth_required")
	}

	if _, err := repo.SaveConnected(ctx, string(shopC), f.barberID, connectedData(now)); err != nil {
		t.Fatal(err)
	}
	if _, found, _ := repo.SetReminder(ctx, string(shopC), f.barberID, ptr(15)); !found {
		t.Fatal("reconectar reactiva la edición del recordatorio")
	}
	if _, err := repo.MarkStatus(ctx, string(shopC), f.barberID, googlecalendar.StatusDisconnected, "", now); err != nil {
		t.Fatal(err)
	}
	conn, _, _ = repo.GetConnection(ctx, string(shopC), f.barberID)
	if conn.Status != googlecalendar.StatusDisconnected || conn.HasCredentials() || conn.AccountEmail != "" || conn.DisconnectedAt == nil {
		t.Fatalf("desconectar borra credenciales y correo y conserva la fila: %+v", conn)
	}
	if conn.ReminderMinutes == nil || *conn.ReminderMinutes != 15 {
		t.Fatalf("desconectar conserva la preferencia de recordatorio: %+v", conn.ReminderMinutes)
	}
	if _, found, _ := repo.SetReminder(ctx, string(shopC), f.barberID, ptr(5)); found {
		t.Fatal("una conexión desconectada no admite cambiar el recordatorio")
	}
}

func TestTwoTenants_CannotSeeNorTouchEachOthersConnection(t *testing.T) {
	db := setupTestDB(t)
	repo := gcalpostgres.New(db)
	fc := newFixtures(t, db, shopC, userC1)
	fd := newFixtures(t, db, shopD, userD1)
	ctx := context.Background()
	now := time.Now().UTC()

	if _, err := repo.SaveConnected(ctx, string(shopC), fc.barberID, connectedData(now)); err != nil {
		t.Fatal(err)
	}

	// D no ve el barbero de C ni su conexión, ni siquiera con el id correcto.
	if _, found, err := repo.GetConnection(ctx, string(shopD), fc.barberID); err != nil || found {
		t.Fatalf("D no debe ver la conexión de C: found=%v err=%v", found, err)
	}
	if found, _ := repo.MarkStatus(ctx, string(shopD), fc.barberID, googlecalendar.StatusDisconnected, "", now); found {
		t.Fatal("D no debe desconectar a C")
	}
	if _, found, _ := repo.SetReminder(ctx, string(shopD), fc.barberID, ptr(5)); found {
		t.Fatal("D no debe cambiar el recordatorio de C")
	}
	conn, found, _ := repo.GetConnection(ctx, string(shopC), fc.barberID)
	if !found || conn.Status != googlecalendar.StatusConnected {
		t.Fatalf("la conexión de C debe seguir intacta: %+v", conn)
	}

	// Crear una conexión de D sobre el barbero de C lo impide la FK compuesta.
	if _, err := repo.SaveConnected(ctx, string(shopD), fc.barberID, connectedData(now)); err == nil {
		t.Fatal("la FK compuesta debe impedir conectar el barbero de otra barbería")
	}
	// Y cada barbería conecta a SUS barberos sin interferir.
	if _, err := repo.SaveConnected(ctx, string(shopD), fd.barberID, connectedData(now)); err != nil {
		t.Fatalf("D conecta a su propio barbero: %v", err)
	}
}

func TestOAuthState_IsSingleUseExpiringAndTenantScoped(t *testing.T) {
	db := setupTestDB(t)
	repo := gcalpostgres.New(db)
	f := newFixtures(t, db, shopC, userC1)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	state := stateFor(shopC, f, userC1, now)
	if err := repo.CreateState(ctx, state, now); err != nil {
		t.Fatalf("CreateState: %v", err)
	}

	// Otra barbería no puede consumirlo, y fallar no lo gasta.
	if _, found, err := repo.ConsumeState(ctx, string(shopD), state.StateHash, now); err != nil || found {
		t.Fatalf("D no debe consumir el state de C: found=%v err=%v", found, err)
	}
	got, found, err := repo.ConsumeState(ctx, string(shopC), state.StateHash, now.Add(time.Minute))
	if err != nil || !found {
		t.Fatalf("ConsumeState: found=%v err=%v", found, err)
	}
	if got.BarberID != f.barberID || got.StaffUserID != userC1 || got.SessionID != f.sessionID ||
		string(got.VerifierEncrypted) != "verificador-cifrado" || got.VerifierKeyID != "v1" {
		t.Fatalf("el estado debe volver íntegro: %+v", got)
	}
	if _, found, _ := repo.ConsumeState(ctx, string(shopC), state.StateHash, now.Add(time.Minute)); found {
		t.Fatal("un state ya consumido no vuelve a servir")
	}

	expired := stateFor(shopC, f, userC1, now)
	if err := repo.CreateState(ctx, expired, now); err != nil {
		t.Fatal(err)
	}
	if _, found, _ := repo.ConsumeState(ctx, string(shopC), expired.StateHash, now.Add(11*time.Minute)); found {
		t.Fatal("un state vencido no debe consumirse")
	}

	// Crear un state nuevo purga los vencidos y consumidos de ese barbero.
	if err := repo.CreateState(ctx, stateFor(shopC, f, userC1, now.Add(time.Hour)), now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	var left int
	_ = db.InTenantTx(ctx, shopC, func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx, `SELECT count(*) FROM google_calendar_oauth_state WHERE barber_id = $1`, f.barberID).Scan(&left)
	})
	if left != 1 {
		t.Fatalf("solo debe quedar el state vigente, hay %d", left)
	}
}

func TestOAuthState_ConcurrentConsumeHasExactlyOneWinner(t *testing.T) {
	db := setupTestDB(t)
	repo := gcalpostgres.New(db)
	f := newFixtures(t, db, shopC, userC1)
	ctx := context.Background()
	now := time.Now().UTC()

	state := stateFor(shopC, f, userC1, now)
	if err := repo.CreateState(ctx, state, now); err != nil {
		t.Fatal(err)
	}

	const attempts = 8
	results := make([]bool, attempts)
	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, results[i], _ = repo.ConsumeState(ctx, string(shopC), state.StateHash, now)
		}()
	}
	wg.Wait()

	winners := 0
	for _, ok := range results {
		if ok {
			winners++
		}
	}
	if winners != 1 {
		t.Fatalf("exactamente un callback debe ganar el state, ganaron %d", winners)
	}
}

func TestBarberOfUser_ReadsTheLinkAndIsTenantScoped(t *testing.T) {
	db := setupTestDB(t)
	repo := gcalpostgres.New(db)
	f := newFixtures(t, db, shopC, userC1)
	ctx := context.Background()
	linkBarber(t, db, shopC, f.barberID, userC1)
	t.Cleanup(func() {
		_ = db.InTenantTx(ctx, shopC, func(ctx context.Context, q database.Queries) error {
			_, err := q.Exec(ctx, `UPDATE barber SET staff_user_id = NULL WHERE barbershop_id = $1 AND staff_user_id = $2`, string(shopC), userC1)
			return err
		})
	})

	id, found, err := repo.BarberOfUser(ctx, string(shopC), userC1)
	if err != nil || !found || id != f.barberID {
		t.Fatalf("BarberOfUser: %q found=%v err=%v", id, found, err)
	}
	if _, found, _ := repo.BarberOfUser(ctx, string(shopD), userC1); found {
		t.Fatal("otra barbería no resuelve al barbero del usuario")
	}
	if _, found, _ := repo.BarberOfUser(ctx, string(shopC), userD1); found {
		t.Fatal("un usuario ajeno no tiene barbero en esta barbería")
	}
}
