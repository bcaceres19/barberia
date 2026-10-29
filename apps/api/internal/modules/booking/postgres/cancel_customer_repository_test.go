// Pruebas de integración de HU-099 (T5, cancelación por el cliente) contra
// PostgreSQL REAL con dos barberías (shopQ y shopR): mismos fixtures que
// cancel_repository_test.go. La política se cambia por barbería con el
// rol de aplicación dentro de su propio tenant, como lo hace HU-093. El reloj
// llega por CancelAppointmentByCustomerInput.Now para fijar «antes, exacto y
// después» del límite sin dormir.
package postgres_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"testing"
	"time"

	"system-barbershop/internal/modules/booking"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

type customerCancelFixture struct {
	db   *database.DB
	repo booking.CustomerCancellationRepository
	full interface {
		CancelByBarber(context.Context, string, booking.CancelAppointmentByBarberInput, idempotency.Key, idempotency.Fingerprint) (booking.CancelAppointmentByBarberResult, error)
	}
	suffix      string
	start       time.Time
	appointment string
	version     string
	tokenHash   string
}

// newCustomerCancelFixture crea una cita futura de shopQ con un enlace vigente
// y deja la política de shopQ en (20, tardía permitida, motivo obligatorio).
func newCustomerCancelFixture(t *testing.T) customerCancelFixture {
	t.Helper()
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	start := baseStart(t)

	created, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, start, 30, suffix))
	if err != nil {
		t.Fatalf("CreateInternal: %v", err)
	}
	detail, found, err := repo.GetAppointmentDetail(context.Background(), string(shopQ), created.Appointment.ID)
	if err != nil || !found {
		t.Fatalf("GetAppointmentDetail: found=%v err=%v", found, err)
	}
	f := customerCancelFixture{
		db: db, repo: repo, full: repo, suffix: suffix, start: start,
		appointment: created.Appointment.ID, version: detail.VersionToken,
		tokenHash: sha256Hex("token-" + suffix),
	}
	err = db.InTenantTx(context.Background(), shopQ, func(ctx context.Context, q database.Queries) error {
		_, err := q.Exec(ctx,
			`INSERT INTO appointment_access_token (barbershop_id, appointment_id, token_hash, issued_at, expires_at)
			 VALUES ($1, $2, $3, now() - interval '1 minute', now() + interval '90 days')`,
			shopQ, created.Appointment.ID, f.tokenHash)
		return err
	})
	if err != nil {
		t.Fatalf("insert token: %v", err)
	}
	f.setPolicy(t, shopQ, 20, true, true)
	return f
}

func (f customerCancelFixture) setPolicy(t *testing.T, shop database.BarbershopID, minutes int, allowed, reasonRequired bool) {
	t.Helper()
	err := f.db.InTenantTx(context.Background(), shop, func(ctx context.Context, q database.Queries) error {
		_, err := q.Exec(ctx,
			`UPDATE barbershop
			    SET cancellation_deadline_minutes = $2, late_cancellation_client_allowed = $3, late_cancellation_reason_required = $4
			  WHERE id = $1`, shop, minutes, allowed, reasonRequired)
		return err
	})
	if err != nil {
		t.Fatalf("setPolicy: %v", err)
	}
}

func (f customerCancelFixture) input(now time.Time, reason string) booking.CancelAppointmentByCustomerInput {
	return booking.CancelAppointmentByCustomerInput{AppointmentID: f.appointment, ExpectedVersionToken: f.version, Reason: reason, Now: now}
}

func (f customerCancelFixture) cancel(t *testing.T, shop database.BarbershopID, in booking.CancelAppointmentByCustomerInput, key string) (booking.CancelAppointmentByCustomerResult, error) {
	t.Helper()
	return f.repo.CancelByCustomer(context.Background(), string(shop), in, idempotency.Key(key+"-"+f.suffix), cancelFingerprint(key+"-"+f.suffix))
}

type customerCancelState struct {
	status, reason string
	hasReason      bool
	tokenRevoked   bool
	events         int
	eventReason    string
	eventHasReason bool
	actorType      string
	actorCustomer  bool
}

func (f customerCancelFixture) state(t *testing.T) customerCancelState {
	t.Helper()
	var s customerCancelState
	err := f.db.InTenantTx(context.Background(), shopQ, func(ctx context.Context, q database.Queries) error {
		var reason, eventReason *string
		if err := q.QueryRow(ctx, `SELECT status, cancellation_reason FROM appointment WHERE barbershop_id = $1 AND id = $2`,
			shopQ, f.appointment).Scan(&s.status, &reason); err != nil {
			return err
		}
		if reason != nil {
			s.hasReason, s.reason = true, *reason
		}
		if err := q.QueryRow(ctx, `SELECT revoked_at IS NOT NULL FROM appointment_access_token WHERE token_hash = $1`, f.tokenHash).Scan(&s.tokenRevoked); err != nil {
			return err
		}
		if err := q.QueryRow(ctx, `SELECT count(*) FROM appointment_history WHERE barbershop_id = $1 AND appointment_id = $2 AND event_type = 'appointment_cancelled_by_customer'`,
			shopQ, f.appointment).Scan(&s.events); err != nil {
			return err
		}
		if s.events > 0 {
			if err := q.QueryRow(ctx, `SELECT reason, actor_type, actor_customer_id IS NOT NULL FROM appointment_history
			   WHERE barbershop_id = $1 AND appointment_id = $2 AND event_type = 'appointment_cancelled_by_customer' LIMIT 1`,
				shopQ, f.appointment).Scan(&eventReason, &s.actorType, &s.actorCustomer); err != nil {
				return err
			}
			if eventReason != nil {
				s.eventHasReason, s.eventReason = true, *eventReason
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("state: %v", err)
	}
	return s
}

func (f customerCancelFixture) assertUntouched(t *testing.T) {
	t.Helper()
	s := f.state(t)
	if s.status != "confirmed" || s.hasReason || s.tokenRevoked || s.events != 0 {
		t.Fatalf("un rechazo no debe modificar nada: %+v", s)
	}
}

// CA-099-01: un único cambio de estado, motivo, un único evento con actor
// cliente y enlace revocado, todo en la misma transacción.
func TestCancelByCustomer_WithinDeadline_CancelsOnceWithCustomerActor(t *testing.T) {
	f := newCustomerCancelFixture(t)
	now := f.start.Add(-2 * time.Hour)

	result, err := f.cancel(t, shopQ, f.input(now, "cambio de planes"), "ok")
	if err != nil {
		t.Fatalf("CancelByCustomer: %v", err)
	}
	if result.Decision.Outcome != idempotency.OutcomeProceed || result.Response.Status != 200 ||
		result.Response.Body != `{"status":"cancelled_by_customer"}` {
		t.Fatalf("result = %+v", result)
	}
	s := f.state(t)
	if s.status != "cancelled_by_customer" || !s.hasReason || s.reason != "cambio de planes" || !s.tokenRevoked ||
		s.events != 1 || s.actorType != "customer" || !s.actorCustomer || !s.eventHasReason || s.eventReason != "cambio de planes" {
		t.Fatalf("state = %+v", s)
	}
}

func TestCancelByCustomer_WithinDeadlineWithoutReason_StoresNoReason(t *testing.T) {
	f := newCustomerCancelFixture(t)
	if _, err := f.cancel(t, shopQ, f.input(f.start.Add(-time.Hour), ""), "noreason"); err != nil {
		t.Fatalf("CancelByCustomer: %v", err)
	}
	s := f.state(t)
	if s.status != "cancelled_by_customer" || s.hasReason || s.eventHasReason {
		t.Fatalf("no se debe inventar un motivo: %+v", s)
	}
}

// CA-099-02 y CA-099-03 contra la base: antes, exacto y después del límite ×
// política tardía × motivo, con la política leída de la barbería.
func TestCancelByCustomer_PolicyBoundaryTable(t *testing.T) {
	cases := []struct {
		name           string
		offset         time.Duration // respecto al límite startsAt-20min
		allowed        bool
		reasonRequired bool
		reason         string
		want           apperr.Kind // vacío = éxito
	}{
		{"antes del límite", -time.Minute, false, false, "", ""},
		{"exactamente en el límite", 0, false, false, "", ""},
		{"un instante después, tardía prohibida", time.Nanosecond, false, false, "", apperr.KindCancellationWindowClosed},
		{"un minuto después, tardía permitida sin motivo obligatorio", time.Minute, true, false, "", ""},
		{"un minuto después, permitida con motivo y con motivo", time.Minute, true, true, "no llego", ""},
		{"un minuto después, permitida con motivo obligatorio sin motivo", time.Minute, true, true, "", apperr.KindCancellationReasonRequired},
		{"ya iniciada", 20 * time.Minute, true, false, "tarde", apperr.KindCancellationWindowClosed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newCustomerCancelFixture(t)
			f.setPolicy(t, shopQ, 20, tc.allowed, tc.reasonRequired)
			now := f.start.Add(-20 * time.Minute).Add(tc.offset)

			_, err := f.cancel(t, shopQ, f.input(now, tc.reason), "table")
			if tc.want == "" {
				if err != nil {
					t.Fatalf("esperaba éxito: %v", err)
				}
				if f.state(t).status != "cancelled_by_customer" {
					t.Fatal("la cita debió quedar cancelada")
				}
				return
			}
			appErr, ok := apperr.As(err)
			if !ok || appErr.Kind != tc.want {
				t.Fatalf("esperaba %s, obtuvo %v", tc.want, err)
			}
			f.assertUntouched(t)
		})
	}
}

// Un cambio de política aplica desde que se confirma, y un rechazo no consume
// la clave: reintentar con la MISMA clave ya con la política abierta funciona.
func TestCancelByCustomer_RejectionDoesNotConsumeKey_PolicyChangeApplies(t *testing.T) {
	f := newCustomerCancelFixture(t)
	f.setPolicy(t, shopQ, 20, false, false)
	late := f.start.Add(-5 * time.Minute)

	_, err := f.cancel(t, shopQ, f.input(late, ""), "retry")
	if appErr, ok := apperr.As(err); !ok || appErr.Kind != apperr.KindCancellationWindowClosed {
		t.Fatalf("esperaba plazo vencido, obtuvo %v", err)
	}
	f.assertUntouched(t)

	f.setPolicy(t, shopQ, 20, true, false)
	if _, err := f.cancel(t, shopQ, f.input(late, ""), "retry"); err != nil {
		t.Fatalf("la misma clave debió poder reintentarse tras un rechazo: %v", err)
	}
	if f.state(t).status != "cancelled_by_customer" {
		t.Fatal("la cita debió quedar cancelada")
	}
}

// CA-099-05: repetición exacta reproduce la respuesta sin duplicar historial.
func TestCancelByCustomer_ExactRepeat_ReplaysWithoutDuplicateEvent(t *testing.T) {
	f := newCustomerCancelFixture(t)
	in := f.input(f.start.Add(-time.Hour), "")

	first, err := f.cancel(t, shopQ, in, "replay")
	if err != nil {
		t.Fatalf("CancelByCustomer(1): %v", err)
	}
	second, err := f.cancel(t, shopQ, in, "replay")
	if err != nil {
		t.Fatalf("CancelByCustomer(2): %v", err)
	}
	if second.Decision.Outcome != idempotency.OutcomeReplay || second.Response.Body != first.Response.Body {
		t.Fatalf("la repetición no reprodujo la respuesta: %+v", second)
	}
	if got := f.state(t).events; got != 1 {
		t.Fatalf("eventos de cancelación = %d, want 1", got)
	}
}

// CA-099-05: clave divergente (misma clave, otra huella) no modifica nada.
func TestCancelByCustomer_SameKeyDifferentFingerprint_IsIdempotencyConflict(t *testing.T) {
	f := newCustomerCancelFixture(t)
	in := f.input(f.start.Add(-time.Hour), "")
	key := idempotency.Key("diverge-" + f.suffix)

	if _, err := f.repo.CancelByCustomer(context.Background(), string(shopQ), in, key, cancelFingerprint("a-"+f.suffix)); err != nil {
		t.Fatalf("CancelByCustomer(1): %v", err)
	}
	result, err := f.repo.CancelByCustomer(context.Background(), string(shopQ), in, key, cancelFingerprint("b-"+f.suffix))
	if err != nil {
		t.Fatalf("CancelByCustomer(2): %v", err)
	}
	if result.Decision.Outcome == idempotency.OutcomeProceed || result.Decision.Outcome == idempotency.OutcomeReplay {
		t.Fatalf("una huella divergente no puede proceder ni reproducir: %v", result.Decision.Outcome)
	}
	if got := f.state(t).events; got != 1 {
		t.Fatalf("eventos de cancelación = %d, want 1", got)
	}
}

// CA-099-05: versión obsoleta no modifica nada.
func TestCancelByCustomer_StaleVersion_ReturnsVersionConflictAndChangesNothing(t *testing.T) {
	f := newCustomerCancelFixture(t)
	in := f.input(f.start.Add(-time.Hour), "")
	in.ExpectedVersionToken = "version-que-ya-no-es"

	_, err := f.cancel(t, shopQ, in, "stale")
	if appErr, ok := apperr.As(err); !ok || appErr.Kind != apperr.KindVersionConflict {
		t.Fatalf("esperaba VersionConflict, obtuvo %v", err)
	}
	f.assertUntouched(t)
}

// CA-099-05: otro estado terminal (aquí, cancelada por el barbero) responde
// InvalidState y no agrega un evento del cliente.
func TestCancelByCustomer_OtherTerminalState_ReturnsInvalidState(t *testing.T) {
	f := newCustomerCancelFixture(t)
	barberInput := booking.CancelAppointmentByBarberInput{
		AppointmentID: f.appointment, ExpectedVersionToken: f.version,
		Actor: booking.Actor{Type: booking.ActorTypeStaff, StaffUserID: strPtr(staffQ)},
	}
	if _, err := f.full.CancelByBarber(context.Background(), string(shopQ), barberInput,
		idempotency.Key("barber-"+f.suffix), cancelFingerprint("barber-"+f.suffix)); err != nil {
		t.Fatalf("CancelByBarber: %v", err)
	}

	_, err := f.cancel(t, shopQ, f.input(f.start.Add(-time.Hour), ""), "terminal")
	if appErr, ok := apperr.As(err); !ok || appErr.Kind != apperr.KindInvalidState {
		t.Fatalf("esperaba InvalidState, obtuvo %v", err)
	}
	if s := f.state(t); s.status != "cancelled_by_barber" || s.events != 0 {
		t.Fatalf("state = %+v", s)
	}
}

// RN-TEN-01: el identificador de la cita de una barbería no existe en otra, y
// la política de la otra barbería nunca se aplica.
func TestCancelByCustomer_TwoTenants_OtherShopSeesNotFoundAndKeepsItsOwnPolicy(t *testing.T) {
	f := newCustomerCancelFixture(t)
	f.setPolicy(t, shopR, 0, false, false)

	_, err := f.cancel(t, shopR, f.input(f.start.Add(-time.Hour), ""), "tenant")
	if appErr, ok := apperr.As(err); !ok || appErr.Kind != apperr.KindNotFound {
		t.Fatalf("esperaba NotFound desde otra barbería, obtuvo %v", err)
	}
	f.assertUntouched(t)

	// La política restrictiva de shopR no afecta a shopQ.
	if _, err := f.cancel(t, shopQ, f.input(f.start.Add(-time.Hour), ""), "own"); err != nil {
		t.Fatalf("la política de otra barbería no debe aplicar: %v", err)
	}
}

// CA-099-04: tras cancelar, la franja exacta queda libre para otra cita.
func TestCancelByCustomer_ReleasesSlot(t *testing.T) {
	f := newCustomerCancelFixture(t)
	repo := newRepository(f.db)

	if _, err := f.cancel(t, shopQ, f.input(f.start.Add(-time.Hour), ""), "free"); err != nil {
		t.Fatalf("CancelByCustomer: %v", err)
	}
	if _, err := repo.CreateInternal(context.Background(), string(shopQ), validInput(t, barberQ1, serviceQ, f.start, 30, f.suffix+"-new")); err != nil {
		t.Fatalf("la franja liberada debió aceptar una cita nueva: %v", err)
	}
}

// Carrera: dos cancelaciones concurrentes con claves distintas. FOR UPDATE las
// serializa; ambas tienen éxito y el historial conserva un único evento.
func TestCancelByCustomer_ConcurrentDoubleCancel_ExactlyOneEvent(t *testing.T) {
	f := newCustomerCancelFixture(t)
	in := f.input(f.start.Add(-time.Hour), "")

	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i := range errs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = f.cancel(t, shopQ, in, "race-"+string(rune('a'+i)))
		}()
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("cancelación %d: %v", i, err)
		}
	}
	if got := f.state(t).events; got != 1 {
		t.Fatalf("eventos de cancelación = %d, want 1", got)
	}
}
