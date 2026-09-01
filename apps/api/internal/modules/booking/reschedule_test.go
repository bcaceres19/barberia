package booking_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"system-barbershop/internal/modules/booking"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/idempotency"
)

// fakeRescheduleRepository es un doble de booking.Repository que solo
// implementa GetAppointmentDetail/Reschedule: las demás operaciones nunca
// deben llamarse desde RescheduleService (heredadas de
// fakeManualRepository, que entra en panic si se invocan).
type fakeRescheduleRepository struct {
	fakeManualRepository

	detail      booking.AppointmentDetail
	detailFound bool
	detailErr   error

	rescheduleFn func(ctx context.Context, barbershopID string, input booking.RescheduleInput, key idempotency.Key, fingerprint idempotency.Fingerprint) (booking.RescheduleResult, error)

	lastRescheduleInput booking.RescheduleInput
}

func (f *fakeRescheduleRepository) GetAppointmentDetail(
	_ context.Context, _, _ string,
) (booking.AppointmentDetail, bool, error) {
	if f.detailErr != nil {
		return booking.AppointmentDetail{}, false, f.detailErr
	}
	return f.detail, f.detailFound, nil
}

func (f *fakeRescheduleRepository) Reschedule(
	ctx context.Context, barbershopID string, input booking.RescheduleInput, key idempotency.Key, fingerprint idempotency.Fingerprint,
) (booking.RescheduleResult, error) {
	f.lastRescheduleInput = input
	if f.rescheduleFn == nil {
		return booking.RescheduleResult{
			Decision:    idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Appointment: booking.RescheduleAppointment{ID: "appt-1"},
		}, nil
	}
	return f.rescheduleFn(ctx, barbershopID, input, key, fingerprint)
}

const validRescheduleAppointmentID = "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"

func validRescheduleRequest(t *testing.T, startsAt time.Time) booking.RescheduleAppointmentRequest {
	t.Helper()
	return booking.RescheduleAppointmentRequest{
		AppointmentID:        validRescheduleAppointmentID,
		NewStartsAtLocal:     startsAt.Format("2006-01-02T15:04:05"),
		ExpectedVersionToken: "opaque-token",
		ActorStaffUserID:     "staff-1",
	}
}

func TestRescheduleAppointment_MalformedAppointmentID_RejectsWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRescheduleRepository{}
	svc := booking.NewRescheduleService(repo, fakeBlockCheckPort{}, fakeTimezonePort{timezone: "America/Bogota"}, fakeClock{now: time.Now()})

	req := validRescheduleRequest(t, time.Now().Add(48*time.Hour))
	req.AppointmentID = "not-a-uuid"

	_, err := svc.RescheduleAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindNotFound)
}

func TestRescheduleAppointment_MissingVersionToken_RejectsAsInvalid(t *testing.T) {
	repo := &fakeRescheduleRepository{}
	svc := booking.NewRescheduleService(repo, fakeBlockCheckPort{}, fakeTimezonePort{timezone: "America/Bogota"}, fakeClock{now: time.Now()})

	req := validRescheduleRequest(t, time.Now().Add(48*time.Hour))
	req.ExpectedVersionToken = "  "

	_, err := svc.RescheduleAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindInvalid)
}

func TestRescheduleAppointment_MissingActor_RejectsAsValidation(t *testing.T) {
	repo := &fakeRescheduleRepository{}
	svc := booking.NewRescheduleService(repo, fakeBlockCheckPort{}, fakeTimezonePort{timezone: "America/Bogota"}, fakeClock{now: time.Now()})

	req := validRescheduleRequest(t, time.Now().Add(48*time.Hour))
	req.ActorStaffUserID = ""

	_, err := svc.RescheduleAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindValidation)
}

func TestRescheduleAppointment_InvalidStartsAtFormat_RejectsAsInvalid(t *testing.T) {
	repo := &fakeRescheduleRepository{}
	svc := booking.NewRescheduleService(repo, fakeBlockCheckPort{}, fakeTimezonePort{timezone: "America/Bogota"}, fakeClock{now: time.Now()})

	req := validRescheduleRequest(t, time.Now().Add(48*time.Hour))
	req.NewStartsAtLocal = "2026-09-10T14:30:00Z" // sufijo de zona no admitido

	_, err := svc.RescheduleAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindInvalid)
}

func TestRescheduleAppointment_StartsAtNotFuture_RejectsAsValidation(t *testing.T) {
	// El reloj fijo y el nuevo inicio se construyen en la MISMA zona que el
	// servicio usará para interpretar el string civil (America/Bogota):
	// formatear "ahora - 1h" en UTC y reinterpretarlo en Bogota (UTC-5)
	// desplazaría el instante 5 horas, pudiendo volverlo "futuro" por
	// accidente.
	loc := mustLoadLocation(t, "America/Bogota")
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, loc)
	repo := &fakeRescheduleRepository{}
	svc := booking.NewRescheduleService(repo, fakeBlockCheckPort{}, fakeTimezonePort{timezone: "America/Bogota"}, fakeClock{now: now})

	req := validRescheduleRequest(t, now.Add(-1*time.Hour))

	_, err := svc.RescheduleAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindValidation)
}

func TestRescheduleAppointment_AppointmentNotFound_ReturnsNotFound(t *testing.T) {
	repo := &fakeRescheduleRepository{detailFound: false}
	svc := booking.NewRescheduleService(repo, fakeBlockCheckPort{}, fakeTimezonePort{timezone: "America/Bogota"}, fakeClock{now: time.Now()})

	req := validRescheduleRequest(t, time.Now().Add(48*time.Hour))
	_, err := svc.RescheduleAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindNotFound)
}

func TestRescheduleAppointment_GetAppointmentDetailError_ReturnsInternal(t *testing.T) {
	repo := &fakeRescheduleRepository{detailErr: errors.New("boom")}
	svc := booking.NewRescheduleService(repo, fakeBlockCheckPort{}, fakeTimezonePort{timezone: "America/Bogota"}, fakeClock{now: time.Now()})

	req := validRescheduleRequest(t, time.Now().Add(48*time.Hour))
	_, err := svc.RescheduleAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindInternal)
}

func TestRescheduleAppointment_TimezonePortError_Propagates(t *testing.T) {
	repo := &fakeRescheduleRepository{detailFound: true, detail: booking.AppointmentDetail{BarberID: "barber-1", DurationMinutesSnapshot: 30}}
	wantErr := apperr.Internal(errors.New("shops: fallo simulado"))
	svc := booking.NewRescheduleService(repo, fakeBlockCheckPort{}, fakeTimezonePort{err: wantErr}, fakeClock{now: time.Now()})

	req := validRescheduleRequest(t, time.Now().Add(48*time.Hour))
	_, err := svc.RescheduleAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}

func TestRescheduleAppointment_BlockedInterval_ReturnsConflictWithoutCallingRepository(t *testing.T) {
	repo := &fakeRescheduleRepository{detailFound: true, detail: booking.AppointmentDetail{BarberID: "barber-1", DurationMinutesSnapshot: 30}}
	svc := booking.NewRescheduleService(repo, fakeBlockCheckPort{blocked: true}, fakeTimezonePort{timezone: "America/Bogota"}, fakeClock{now: time.Now()})

	req := validRescheduleRequest(t, time.Now().Add(48*time.Hour))
	_, err := svc.RescheduleAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindConflict)
	if repo.lastRescheduleInput.AppointmentID != "" {
		t.Fatalf("no debió llamar al repositorio con un bloqueo vigente (DEC-076)")
	}
}

func TestRescheduleAppointment_BlockCheckPortError_Propagates(t *testing.T) {
	repo := &fakeRescheduleRepository{detailFound: true, detail: booking.AppointmentDetail{BarberID: "barber-1", DurationMinutesSnapshot: 30}}
	wantErr := apperr.Internal(errors.New("schedule: fallo simulado"))
	svc := booking.NewRescheduleService(repo, fakeBlockCheckPort{err: wantErr}, fakeTimezonePort{timezone: "America/Bogota"}, fakeClock{now: time.Now()})

	req := validRescheduleRequest(t, time.Now().Add(48*time.Hour))
	_, err := svc.RescheduleAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}

func TestRescheduleAppointment_ValidRequest_ComputesEndsAtFromDurationSnapshotAndCallsRepository(t *testing.T) {
	repo := &fakeRescheduleRepository{
		detailFound: true,
		detail:      booking.AppointmentDetail{BarberID: "barber-1", DurationMinutesSnapshot: 45},
	}
	svc := booking.NewRescheduleService(repo, fakeBlockCheckPort{}, fakeTimezonePort{timezone: "America/Bogota"}, fakeClock{now: time.Now()})

	newStarts := time.Now().Add(48 * time.Hour)
	req := validRescheduleRequest(t, newStarts)

	_, err := svc.RescheduleAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}

	got := repo.lastRescheduleInput
	if got.AppointmentID != validRescheduleAppointmentID {
		t.Fatalf("AppointmentID = %q, want %q", got.AppointmentID, validRescheduleAppointmentID)
	}
	wantDuration := 45 * time.Minute
	if got.NewEndsAt.Sub(got.NewStartsAt) != wantDuration {
		t.Fatalf("duración derivada = %v, want %v (nunca del catálogo ni del cliente)", got.NewEndsAt.Sub(got.NewStartsAt), wantDuration)
	}
	if got.ExpectedVersionToken != "opaque-token" {
		t.Fatalf("ExpectedVersionToken = %q, want %q", got.ExpectedVersionToken, "opaque-token")
	}
	if got.Actor.Type != booking.ActorTypeStaff || got.Actor.StaffUserID == nil || *got.Actor.StaffUserID != "staff-1" {
		t.Fatalf("Actor = %+v, want staff-1", got.Actor)
	}
}

func TestRescheduleAppointment_RepositoryResult_PassesThrough(t *testing.T) {
	want := booking.RescheduleResult{
		Decision:    idempotency.Decision{Outcome: idempotency.OutcomeProceed},
		Appointment: booking.RescheduleAppointment{ID: "appt-9", VersionToken: "new-token"},
	}
	repo := &fakeRescheduleRepository{
		detailFound: true,
		detail:      booking.AppointmentDetail{BarberID: "barber-1", DurationMinutesSnapshot: 30},
		rescheduleFn: func(context.Context, string, booking.RescheduleInput, idempotency.Key, idempotency.Fingerprint) (booking.RescheduleResult, error) {
			return want, nil
		},
	}
	svc := booking.NewRescheduleService(repo, fakeBlockCheckPort{}, fakeTimezonePort{timezone: "America/Bogota"}, fakeClock{now: time.Now()})

	req := validRescheduleRequest(t, time.Now().Add(48*time.Hour))
	got, err := svc.RescheduleAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}
	if got.Appointment.ID != "appt-9" || got.Appointment.VersionToken != "new-token" {
		t.Fatalf("resultado no pasado tal cual: %+v", got)
	}
}
