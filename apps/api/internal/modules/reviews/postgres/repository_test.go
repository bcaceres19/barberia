// Package postgres_test (pruebas de integración, HU-100) requiere PostgreSQL REAL
// con todas las migraciones aplicadas (incluida
// 20261031120000_create_barber_review.sql) y database/testdata/hu100_resenas.sql
// cargado. Cada prueba crea sus propias citas con booking/postgres.CreateInternal
// sobre las barberías dedicadas y su propia fila appointment_access_token; las
// reseñas y las citas no se pueden borrar (RN-HIS-02), así que cada prueba usa
// un barbero propio y verifica diferencias, no totales absolutos.
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/reviews/postgres/...
package postgres_test

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"system-barbershop/internal/modules/booking"
	bookingpostgres "system-barbershop/internal/modules/booking/postgres"
	"system-barbershop/internal/modules/reviews"
	reviewspostgres "system-barbershop/internal/modules/reviews/postgres"
	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

const (
	testDatabaseURL = "postgres://barberia_app@localhost:5432/barberia_test?sslmode=disable"

	shopUno    = "01010001-0101-0101-0101-010100010001"
	serviceUno = "01010031-0101-0101-0101-010100310031"
	slugUno    = "barberia-hu100-uno"

	shopDos    = "01010002-0101-0101-0101-010100020002"
	serviceDos = "01010032-0101-0101-0101-010100320032"
	slugDos    = "barberia-hu100-dos"
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

// newBarber crea un barbero propio de la prueba: las reseñas no se pueden borrar,
// así que reutilizar uno entre corridas arrastraría su promedio.
func newBarber(t *testing.T, db *database.DB, shopID string) string {
	t.Helper()
	var id string
	err := db.InTenantTx(context.Background(), database.BarbershopID(shopID), func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx,
			`INSERT INTO barber (barbershop_id, full_name) VALUES ($1, $2) RETURNING id`,
			shopID, "Barbero de prueba HU-100 "+strconv.FormatInt(time.Now().UnixNano(), 10)).Scan(&id)
	})
	if err != nil {
		t.Fatalf("crear barbero: %v", err)
	}
	return id
}

func randInt(t *testing.T, n int64) int64 {
	t.Helper()
	v, err := rand.Int(rand.Reader, big.NewInt(n))
	if err != nil {
		t.Fatalf("rand.Int: %v", err)
	}
	return v.Int64()
}

// newAppointment crea una cita real y la deja en el estado pedido.
func newAppointment(t *testing.T, db *database.DB, shopID, barberID, serviceID string, status booking.Status) string {
	t.Helper()
	repo := bookingpostgres.New(db, idempotency.NewSQLCoordinator())
	starts := time.Date(2027, 3, 15, 8, 0, 0, 0, time.UTC).
		AddDate(0, 0, int(randInt(t, 3650))).
		Add(time.Duration(randInt(t, 20)) * time.Hour)
	created, err := repo.CreateInternal(context.Background(), shopID, booking.CreateInternalInput{
		BarberID:     barberID,
		ServiceID:    serviceID,
		AttendeeName: "Cliente de prueba HU-100",
		StartsAt:     starts,
		EndsAt:       starts.Add(30 * time.Minute),
		Origin:       booking.OriginPublic,
		Service: booking.ServiceSnapshot{
			Name: "Corte de prueba HU-100", DurationMinutes: 30, PriceAmountCents: 2000000, Currency: "COP",
		},
		Customer: booking.CustomerInput{New: &booking.NewCustomerInput{FullName: "Cliente de prueba HU-100"}},
		Actor:    booking.Actor{Type: booking.ActorTypeSystem},
	})
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	id := created.Appointment.ID
	if status != booking.StatusConfirmed {
		err := db.InTenantTx(context.Background(), database.BarbershopID(shopID), func(ctx context.Context, q database.Queries) error {
			_, err := q.Exec(ctx, `UPDATE appointment SET status = $1, resolved_at = now() WHERE id = $2`, string(status), id)
			return err
		})
		if err != nil {
			t.Fatalf("forzar estado %q: %v", status, err)
		}
	}
	return id
}

// newToken inserta un token vigente para la cita y devuelve (valor en claro, hash).
func newToken(t *testing.T, db *database.DB, shopID, appointmentID string) (plain, hash string) {
	t.Helper()
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		t.Fatal(err)
	}
	plain = hex.EncodeToString(buf)
	sum := sha256.Sum256([]byte(plain))
	hash = hex.EncodeToString(sum[:])
	issued := time.Now().UTC().Add(-time.Hour)
	err := db.InTenantTx(context.Background(), database.BarbershopID(shopID), func(ctx context.Context, q database.Queries) error {
		_, err := q.Exec(ctx,
			`INSERT INTO appointment_access_token (barbershop_id, appointment_id, token_hash, issued_at, expires_at)
			 VALUES ($1, $2, $3, $4, $5)`,
			shopID, appointmentID, hash, issued, issued.Add(90*24*time.Hour))
		return err
	})
	if err != nil {
		t.Fatalf("insertar token: %v", err)
	}
	return plain, hash
}

// review crea una cita completada, su token y la reseña; devuelve la reseña.
func review(t *testing.T, db *database.DB, repo *reviewspostgres.Repository, shopID, barberID, serviceID string, rating int, comment *string) reviews.Review {
	t.Helper()
	appt := newAppointment(t, db, shopID, barberID, serviceID, booking.StatusCompleted)
	_, hash := newToken(t, db, shopID, appt)
	got, outcome, err := repo.CreateByTokenHash(context.Background(), hash, rating, comment)
	if err != nil || outcome != reviews.Created {
		t.Fatalf("CreateByTokenHash: outcome=%v err=%v", outcome, err)
	}
	return got
}

func ptr(s string) *string { return &s }

func TestCreate_CompletedAppointment_StoresOneReviewForTheAppointmentsBarber(t *testing.T) {
	db := setupTestDB(t)
	repo := reviewspostgres.New(db)
	barberCreate := newBarber(t, db, shopUno)
	appt := newAppointment(t, db, shopUno, barberCreate, serviceUno, booking.StatusCompleted)
	_, hash := newToken(t, db, shopUno, appt)

	got, outcome, err := repo.CreateByTokenHash(context.Background(), hash, 5, ptr("Excelente"))
	if err != nil || outcome != reviews.Created {
		t.Fatalf("outcome=%v err=%v", outcome, err)
	}
	if got.BarberID != barberCreate || got.Rating != 5 || got.Comment == nil || *got.Comment != "Excelente" || got.HiddenAt != nil {
		t.Fatalf("reseña inesperada: %+v", got)
	}

	// CA-100-03: un segundo intento sobre el mismo turno no crea ni modifica nada.
	_, outcome, err = repo.CreateByTokenHash(context.Background(), hash, 1, ptr("Otra"))
	if err != nil || outcome != reviews.NotAvailable {
		t.Fatalf("segundo intento: outcome=%v err=%v, se esperaba NotAvailable", outcome, err)
	}
	var count, rating int
	err = db.InTenantTx(context.Background(), database.BarbershopID(shopUno), func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx, `SELECT count(*), max(rating) FROM barber_review WHERE appointment_id = $1`, appt).Scan(&count, &rating)
	})
	if err != nil || count != 1 || rating != 5 {
		t.Fatalf("count=%d rating=%d err=%v, se esperaba 1 reseña de 5 estrellas", count, rating, err)
	}
}

func TestCreate_NotCompletedOrUnknownToken_NeverCreatesAReview(t *testing.T) {
	db := setupTestDB(t)
	repo := reviewspostgres.New(db)
	barberCreate := newBarber(t, db, shopUno)

	for _, status := range []booking.Status{booking.StatusConfirmed, booking.StatusCancelledByCustomer, booking.StatusNoShow} {
		appt := newAppointment(t, db, shopUno, barberCreate, serviceUno, status)
		_, hash := newToken(t, db, shopUno, appt)
		_, outcome, err := repo.CreateByTokenHash(context.Background(), hash, 5, nil)
		if err != nil || outcome != reviews.NotAvailable {
			t.Fatalf("estado %q: outcome=%v err=%v, se esperaba NotAvailable", status, outcome, err)
		}
	}

	sum := sha256.Sum256([]byte("token-que-no-existe"))
	_, outcome, err := repo.CreateByTokenHash(context.Background(), hex.EncodeToString(sum[:]), 5, nil)
	if err != nil || outcome != reviews.TokenNotFound {
		t.Fatalf("token desconocido: outcome=%v err=%v", outcome, err)
	}
}

func TestCreate_RevokedOrExpiredToken_IsTokenNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := reviewspostgres.New(db)
	barberCreate := newBarber(t, db, shopUno)
	appt := newAppointment(t, db, shopUno, barberCreate, serviceUno, booking.StatusCompleted)
	_, hash := newToken(t, db, shopUno, appt)
	err := db.InTenantTx(context.Background(), database.BarbershopID(shopUno), func(ctx context.Context, q database.Queries) error {
		_, err := q.Exec(ctx, `UPDATE appointment_access_token SET revoked_at = now() WHERE token_hash = $1`, hash)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, outcome, err := repo.CreateByTokenHash(context.Background(), hash, 5, nil); err != nil || outcome != reviews.TokenNotFound {
		t.Fatalf("token revocado: outcome=%v err=%v", outcome, err)
	}
}

func TestSummary_AveragesVisibleReviewsOnly_AndHidingIsReversible(t *testing.T) {
	db := setupTestDB(t)
	repo := reviewspostgres.New(db)
	ctx := context.Background()
	barberSummary := newBarber(t, db, shopUno)

	// Sin reseñas: promedio nulo y total cero (CA-100-04).
	empty, found, err := repo.ListPublic(ctx, slugUno, barberSummary, reviews.PublicListLimit)
	if err != nil || !found {
		t.Fatalf("ListPublic: found=%v err=%v", found, err)
	}
	if empty.Summary.Rating != nil || empty.Summary.RatingCount != 0 || len(empty.Items) != 0 {
		t.Fatalf("sin reseñas: %+v", empty)
	}

	_ = review(t, db, repo, shopUno, barberSummary, serviceUno, 5, ptr("Genial"))
	_ = review(t, db, repo, shopUno, barberSummary, serviceUno, 4, nil)
	r1 := review(t, db, repo, shopUno, barberSummary, serviceUno, 1, ptr("Comentario ofensivo"))

	list, _, err := repo.ListPublic(ctx, slugUno, barberSummary, reviews.PublicListLimit)
	if err != nil {
		t.Fatal(err)
	}
	// (5+4+1)/3 = 3.333… → 3.3
	if list.Summary.Rating == nil || *list.Summary.Rating != 3.3 || list.Summary.RatingCount != 3 || len(list.Items) != 3 {
		t.Fatalf("con tres reseñas: %+v", list)
	}

	// CA-100-05: ocultar retira del promedio y de la lista pública…
	hidden, ok, err := repo.SetHidden(ctx, shopUno, barberSummary, r1.ID, true)
	if err != nil || !ok || hidden.HiddenAt == nil {
		t.Fatalf("SetHidden(true): ok=%v err=%v %+v", ok, err, hidden)
	}
	list, _, _ = repo.ListPublic(ctx, slugUno, barberSummary, reviews.PublicListLimit)
	if list.Summary.Rating == nil || *list.Summary.Rating != 4.5 || list.Summary.RatingCount != 2 || len(list.Items) != 2 {
		t.Fatalf("con una oculta: %+v", list)
	}
	for _, it := range list.Items {
		if it.Comment != nil && *it.Comment == "Comentario ofensivo" {
			t.Fatal("una reseña oculta no debe aparecer en la lista pública")
		}
	}
	// …pero el panel la conserva con su fecha.
	panel, found, err := repo.ListForBarber(ctx, shopUno, barberSummary, reviews.PanelListLimit)
	if err != nil || !found || len(panel.Items) != 3 {
		t.Fatalf("panel: found=%v err=%v items=%d", found, err, len(panel.Items))
	}
	var seenHidden bool
	for _, it := range panel.Items {
		if it.ID == r1.ID {
			seenHidden = it.HiddenAt != nil && it.Comment != nil && *it.Comment == "Comentario ofensivo"
		}
	}
	if !seenHidden {
		t.Fatal("el panel debe conservar la reseña oculta con su fecha y contenido")
	}

	// Ocultar de nuevo conserva la fecha original.
	again, _, _ := repo.SetHidden(ctx, shopUno, barberSummary, r1.ID, true)
	if again.HiddenAt == nil || !again.HiddenAt.Equal(*hidden.HiddenAt) {
		t.Fatalf("ocultar una oculta debe conservar su fecha: %v vs %v", again.HiddenAt, hidden.HiddenAt)
	}

	// Volver a mostrar la reincorpora sin perder el contenido.
	shown, ok, err := repo.SetHidden(ctx, shopUno, barberSummary, r1.ID, false)
	if err != nil || !ok || shown.HiddenAt != nil || shown.Comment == nil || *shown.Comment != "Comentario ofensivo" {
		t.Fatalf("SetHidden(false): ok=%v err=%v %+v", ok, err, shown)
	}
	list, _, _ = repo.ListPublic(ctx, slugUno, barberSummary, reviews.PublicListLimit)
	if list.Summary.RatingCount != 3 {
		t.Fatalf("al mostrarla de nuevo debe contar otra vez: %+v", list.Summary)
	}
}

func TestTwoTenants_NeverSeeOrModifyEachOthersReviews(t *testing.T) {
	db := setupTestDB(t)
	repo := reviewspostgres.New(db)
	ctx := context.Background()
	barberIsolate := newBarber(t, db, shopUno)
	barberOther := newBarber(t, db, shopUno)
	barberDos := newBarber(t, db, shopDos)

	mine := review(t, db, repo, shopUno, barberIsolate, serviceUno, 5, ptr("Reseña de Uno"))
	_ = review(t, db, repo, shopDos, barberDos, serviceDos, 2, ptr("Reseña de Dos"))

	// La barbería Dos no puede ocultar la reseña de Uno ni verla por su barbero.
	if _, ok, err := repo.SetHidden(ctx, shopDos, barberIsolate, mine.ID, true); err != nil || ok {
		t.Fatalf("Dos ocultó una reseña de Uno: ok=%v err=%v", ok, err)
	}
	if _, found, err := repo.ListForBarber(ctx, shopDos, barberIsolate, reviews.PanelListLimit); err != nil || found {
		t.Fatalf("Dos listó un barbero de Uno: found=%v err=%v", found, err)
	}
	// Por el slug de Dos tampoco existe el barbero de Uno (mismo resultado que un barbero inexistente).
	if _, found, err := repo.ListPublic(ctx, slugDos, barberIsolate, reviews.PublicListLimit); err != nil || found {
		t.Fatalf("slug de Dos mostró un barbero de Uno: found=%v err=%v", found, err)
	}
	// Reseña de Uno con el barbero equivocado dentro del mismo tenant: no existe.
	if _, ok, _ := repo.SetHidden(ctx, shopUno, barberOther, mine.ID, true); ok {
		t.Fatal("una reseña no debe poder moderarse bajo el barbero de otra persona")
	}
	// Slug inexistente o mal formado.
	if _, found, err := repo.ListPublic(ctx, "slug-que-no-existe", barberIsolate, 5); err != nil || found {
		t.Fatalf("slug inexistente: found=%v err=%v", found, err)
	}
	// Y la reseña de Uno sigue visible.
	list, _, _ := repo.ListPublic(ctx, slugUno, barberIsolate, reviews.PublicListLimit)
	if list.Summary.RatingCount != 1 || len(list.Items) != 1 || list.Items[0].Comment == nil || !strings.Contains(*list.Items[0].Comment, "Uno") {
		t.Fatalf("la reseña de Uno debe seguir intacta: %+v", list)
	}
}

func TestSetHidden_UnknownReview_IsNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := reviewspostgres.New(db)
	barberHide := newBarber(t, db, shopUno)
	if _, ok, err := repo.SetHidden(context.Background(), shopUno, barberHide, "ffffffff-ffff-4fff-8fff-ffffffffffff", true); err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}
