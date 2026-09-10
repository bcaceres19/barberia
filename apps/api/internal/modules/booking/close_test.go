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

// fakeCloseRepository es un doble de booking.Repository que solo implementa
// CompleteAppointment/MarkNoShow: las demás operaciones nunca deben llamarse
// desde CompleteAppointmentService/MarkNoShowService (heredadas de
// fakeManualRepository, que entra en panic si se invocan), mismo criterio
// que fakeCancelRepository en cancel_test.go.
type fakeCloseRepository struct {
	fakeManualRepository

	completeFn func(ctx context.Context, barbershopID string, input booking.CloseAppointmentInput, key idempotency.Key, fingerprint idempotency.Fingerprint) (booking.CompleteAppointmentResult, error)
	noShowFn   func(ctx context.Context, barbershopID string, input booking.CloseAppointmentInput, key idempotency.Key, fingerprint idempotency.Fingerprint) (booking.MarkNoShowResult, error)

	lastCompleteInput booking.CloseAppointmentInput
	lastNoShowInput   booking.CloseAppointmentInput
}

func (f *fakeCloseRepository) CompleteAppointment(
	ctx context.Context, barbershopID string, input booking.CloseAppointmentInput, key idempotency.Key, fingerprint idempotency.Fingerprint,
) (booking.CompleteAppointmentResult, error) {
	f.lastCompleteInput = input
	if f.completeFn == nil {
		return booking.CompleteAppointmentResult{
			Decision:    idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Appointment: booking.ClosedAppointment{ID: "appt-1", Status: booking.StatusCompleted},
		}, nil
	}
	return f.completeFn(ctx, barbershopID, input, key, fingerprint)
}

func (f *fakeCloseRepository) MarkNoShow(
	ctx context.Context, barbershopID string, input booking.CloseAppointmentInput, key idempotency.Key, fingerprint idempotency.Fingerprint,
) (booking.MarkNoShowResult, error) {
	f.lastNoShowInput = input
	if f.noShowFn == nil {
		return booking.MarkNoShowResult{
			Decision:    idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Appointment: booking.ClosedAppointment{ID: "appt-1", Status: booking.StatusNoShow},
		}, nil
	}
	return f.noShowFn(ctx, barbershopID, input, key, fingerprint)
}

const validCloseAppointmentID = "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"

func validCloseRequest() booking.CloseAppointmentRequest {
	return booking.CloseAppointmentRequest{
		AppointmentID:        validCloseAppointmentID,
		ExpectedVersionToken: "opaque-token",
		ActorStaffUserID:     "staff-1",
	}
}

func TestCompleteAppointment_MalformedAppointmentID_RejectsWithoutTouchingRepository(t *testing.T) {
	repo := &fakeCloseRepository{}
	svc := booking.NewCompleteAppointmentService(repo, fakeClock{now: time.Now()})

	req := validCloseRequest()
	req.AppointmentID = "not-a-uuid"

	_, err := svc.CompleteAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindNotFound)
	if repo.lastCompleteInput.AppointmentID != "" {
		t.Fatalf("no debió llamar al repositorio con un id malformado")
	}
}

func TestCompleteAppointment_MissingVersionToken_RejectsAsInvalid(t *testing.T) {
	repo := &fakeCloseRepository{}
	svc := booking.NewCompleteAppointmentService(repo, fakeClock{now: time.Now()})

	req := validCloseRequest()
	req.ExpectedVersionToken = "  "

	_, err := svc.CompleteAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindInvalid)
}

func TestCompleteAppointment_MissingActor_RejectsAsValidation(t *testing.T) {
	repo := &fakeCloseRepository{}
	svc := booking.NewCompleteAppointmentService(repo, fakeClock{now: time.Now()})

	req := validCloseRequest()
	req.ActorStaffUserID = ""

	_, err := svc.CompleteAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindValidation)
}

func TestCompleteAppointment_ValidRequest_CallsRepositoryWithStaffActorAndClockNow(t *testing.T) {
	now := time.Date(2026, 9, 10, 15, 0, 0, 0, time.UTC)
	repo := &fakeCloseRepository{}
	svc := booking.NewCompleteAppointmentService(repo, fakeClock{now: now})

	req := validCloseRequest()
	_, err := svc.CompleteAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}

	got := repo.lastCompleteInput
	if got.AppointmentID != validCloseAppointmentID {
		t.Fatalf("AppointmentID = %q, want %q", got.AppointmentID, validCloseAppointmentID)
	}
	if got.ExpectedVersionToken != "opaque-token" {
		t.Fatalf("ExpectedVersionToken = %q, want %q", got.ExpectedVersionToken, "opaque-token")
	}
	if got.Actor.Type != booking.ActorTypeStaff || got.Actor.StaffUserID == nil || *got.Actor.StaffUserID != "staff-1" {
		t.Fatalf("Actor = %+v, want staff-1", got.Actor)
	}
	if !got.Now.Equal(now) {
		t.Fatalf("Now = %v, want %v (resuelto por el clock inyectado, nunca time.Now directo)", got.Now, now)
	}
}

func TestCompleteAppointment_RepositoryResult_PassesThrough(t *testing.T) {
	want := booking.CompleteAppointmentResult{
		Decision:    idempotency.Decision{Outcome: idempotency.OutcomeProceed},
		Appointment: booking.ClosedAppointment{ID: "appt-9", VersionToken: "new-token", Status: booking.StatusCompleted},
	}
	repo := &fakeCloseRepository{
		completeFn: func(context.Context, string, booking.CloseAppointmentInput, idempotency.Key, idempotency.Fingerprint) (booking.CompleteAppointmentResult, error) {
			return want, nil
		},
	}
	svc := booking.NewCompleteAppointmentService(repo, fakeClock{now: time.Now()})

	req := validCloseRequest()
	got, err := svc.CompleteAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}
	if got.Appointment.ID != "appt-9" || got.Appointment.VersionToken != "new-token" {
		t.Fatalf("resultado no pasado tal cual: %+v", got)
	}
}

func TestCompleteAppointment_RepositoryError_Propagates(t *testing.T) {
	wantErr := apperr.Validation("el turno todavía no comienza; espera hasta su hora de inicio")
	repo := &fakeCloseRepository{
		completeFn: func(context.Context, string, booking.CloseAppointmentInput, idempotency.Key, idempotency.Fingerprint) (booking.CompleteAppointmentResult, error) {
			return booking.CompleteAppointmentResult{}, wantErr
		},
	}
	svc := booking.NewCompleteAppointmentService(repo, fakeClock{now: time.Now()})

	req := validCloseRequest()
	_, err := svc.CompleteAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}

func TestMarkNoShow_MalformedAppointmentID_RejectsWithoutTouchingRepository(t *testing.T) {
	repo := &fakeCloseRepository{}
	svc := booking.NewMarkNoShowService(repo, fakeClock{now: time.Now()})

	req := validCloseRequest()
	req.AppointmentID = "not-a-uuid"

	_, err := svc.MarkNoShow(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindNotFound)
	if repo.lastNoShowInput.AppointmentID != "" {
		t.Fatalf("no debió llamar al repositorio con un id malformado")
	}
}

func TestMarkNoShow_MissingVersionToken_RejectsAsInvalid(t *testing.T) {
	repo := &fakeCloseRepository{}
	svc := booking.NewMarkNoShowService(repo, fakeClock{now: time.Now()})

	req := validCloseRequest()
	req.ExpectedVersionToken = "  "

	_, err := svc.MarkNoShow(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindInvalid)
}

func TestMarkNoShow_MissingActor_RejectsAsValidation(t *testing.T) {
	repo := &fakeCloseRepository{}
	svc := booking.NewMarkNoShowService(repo, fakeClock{now: time.Now()})

	req := validCloseRequest()
	req.ActorStaffUserID = ""

	_, err := svc.MarkNoShow(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindValidation)
}

func TestMarkNoShow_ValidRequest_CallsRepositoryWithStaffActorAndClockNow(t *testing.T) {
	now := time.Date(2026, 9, 10, 15, 0, 0, 0, time.UTC)
	repo := &fakeCloseRepository{}
	svc := booking.NewMarkNoShowService(repo, fakeClock{now: now})

	req := validCloseRequest()
	_, err := svc.MarkNoShow(context.Background(), "shop-1", req, "key-1", "fp-1")
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}

	got := repo.lastNoShowInput
	if got.AppointmentID != validCloseAppointmentID {
		t.Fatalf("AppointmentID = %q, want %q", got.AppointmentID, validCloseAppointmentID)
	}
	if got.Actor.Type != booking.ActorTypeStaff || got.Actor.StaffUserID == nil || *got.Actor.StaffUserID != "staff-1" {
		t.Fatalf("Actor = %+v, want staff-1", got.Actor)
	}
	if !got.Now.Equal(now) {
		t.Fatalf("Now = %v, want %v (resuelto por el clock inyectado, nunca time.Now directo)", got.Now, now)
	}
}

func TestMarkNoShow_RepositoryResult_PassesThrough(t *testing.T) {
	want := booking.MarkNoShowResult{
		Decision:    idempotency.Decision{Outcome: idempotency.OutcomeProceed},
		Appointment: booking.ClosedAppointment{ID: "appt-9", VersionToken: "new-token", Status: booking.StatusNoShow},
	}
	repo := &fakeCloseRepository{
		noShowFn: func(context.Context, string, booking.CloseAppointmentInput, idempotency.Key, idempotency.Fingerprint) (booking.MarkNoShowResult, error) {
			return want, nil
		},
	}
	svc := booking.NewMarkNoShowService(repo, fakeClock{now: time.Now()})

	req := validCloseRequest()
	got, err := svc.MarkNoShow(context.Background(), "shop-1", req, "key-1", "fp-1")
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}
	if got.Appointment.ID != "appt-9" || got.Appointment.VersionToken != "new-token" {
		t.Fatalf("resultado no pasado tal cual: %+v", got)
	}
}

func TestMarkNoShow_RepositoryError_Propagates(t *testing.T) {
	wantErr := apperr.InvalidState("el turno ya tiene un resultado terminal registrado; una futura corrección (T8) permitirá cambiarlo")
	repo := &fakeCloseRepository{
		noShowFn: func(context.Context, string, booking.CloseAppointmentInput, idempotency.Key, idempotency.Fingerprint) (booking.MarkNoShowResult, error) {
			return booking.MarkNoShowResult{}, wantErr
		},
	}
	svc := booking.NewMarkNoShowService(repo, fakeClock{now: time.Now()})

	req := validCloseRequest()
	_, err := svc.MarkNoShow(context.Background(), "shop-1", req, "key-1", "fp-1")
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}
