package booking_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"system-barbershop/internal/modules/booking"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/idempotency"
)

// fakeCorrectRepository es un doble de booking.Repository que solo
// implementa CorrectAppointmentStatus: las demás operaciones nunca deben
// llamarse desde CorrectAppointmentStatusService (heredadas de
// fakeManualRepository, que entra en panic si se invocan), mismo criterio
// que fakeCloseRepository en close_test.go.
type fakeCorrectRepository struct {
	fakeManualRepository

	correctFn func(ctx context.Context, barbershopID string, input booking.CorrectAppointmentStatusInput, key idempotency.Key, fingerprint idempotency.Fingerprint) (booking.CorrectAppointmentStatusResult, error)

	lastCorrectInput booking.CorrectAppointmentStatusInput
}

func (f *fakeCorrectRepository) CorrectAppointmentStatus(
	ctx context.Context, barbershopID string, input booking.CorrectAppointmentStatusInput, key idempotency.Key, fingerprint idempotency.Fingerprint,
) (booking.CorrectAppointmentStatusResult, error) {
	f.lastCorrectInput = input
	if f.correctFn == nil {
		return booking.CorrectAppointmentStatusResult{
			Decision:    idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Appointment: booking.CorrectedAppointment{ID: "appt-1", Status: booking.StatusNoShow},
		}, nil
	}
	return f.correctFn(ctx, barbershopID, input, key, fingerprint)
}

const validCorrectAppointmentID = "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"

func validCorrectRequest() booking.CorrectAppointmentStatusRequest {
	return booking.CorrectAppointmentStatusRequest{
		AppointmentID:        validCorrectAppointmentID,
		DestinationStatus:    string(booking.StatusNoShow),
		Reason:               "El barbero marcó completed por error, el cliente nunca llegó.",
		ExpectedVersionToken: "opaque-token",
		ActorStaffUserID:     "staff-1",
	}
}

func TestCorrectAppointmentStatus_MalformedAppointmentID_RejectsWithoutTouchingRepository(t *testing.T) {
	repo := &fakeCorrectRepository{}
	svc := booking.NewCorrectAppointmentStatusService(repo, fakeClock{now: time.Now()})

	req := validCorrectRequest()
	req.AppointmentID = "not-a-uuid"

	_, err := svc.CorrectAppointmentStatus(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindNotFound)
	if repo.lastCorrectInput.AppointmentID != "" {
		t.Fatalf("no debió llamar al repositorio con un id malformado")
	}
}

func TestCorrectAppointmentStatus_DestinationConfirmed_RejectsAsValidation(t *testing.T) {
	repo := &fakeCorrectRepository{}
	svc := booking.NewCorrectAppointmentStatusService(repo, fakeClock{now: time.Now()})

	req := validCorrectRequest()
	req.DestinationStatus = string(booking.StatusConfirmed)

	_, err := svc.CorrectAppointmentStatus(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindValidation)
	if repo.lastCorrectInput.AppointmentID != "" {
		t.Fatalf("no debió llamar al repositorio con destino confirmed")
	}
}

func TestCorrectAppointmentStatus_DestinationUnknown_RejectsAsValidation(t *testing.T) {
	repo := &fakeCorrectRepository{}
	svc := booking.NewCorrectAppointmentStatusService(repo, fakeClock{now: time.Now()})

	req := validCorrectRequest()
	req.DestinationStatus = "archived"

	_, err := svc.CorrectAppointmentStatus(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindValidation)
}

func TestCorrectAppointmentStatus_ReasonBlank_RejectsAsValidation(t *testing.T) {
	repo := &fakeCorrectRepository{}
	svc := booking.NewCorrectAppointmentStatusService(repo, fakeClock{now: time.Now()})

	req := validCorrectRequest()
	req.Reason = "   "

	_, err := svc.CorrectAppointmentStatus(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindValidation)
	if repo.lastCorrectInput.AppointmentID != "" {
		t.Fatalf("no debió llamar al repositorio con motivo en blanco")
	}
}

func TestCorrectAppointmentStatus_ReasonTooLong_RejectsAsValidation(t *testing.T) {
	repo := &fakeCorrectRepository{}
	svc := booking.NewCorrectAppointmentStatusService(repo, fakeClock{now: time.Now()})

	req := validCorrectRequest()
	req.Reason = strings.Repeat("a", 501)

	_, err := svc.CorrectAppointmentStatus(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindValidation)
}

func TestCorrectAppointmentStatus_MissingVersionToken_RejectsAsInvalid(t *testing.T) {
	repo := &fakeCorrectRepository{}
	svc := booking.NewCorrectAppointmentStatusService(repo, fakeClock{now: time.Now()})

	req := validCorrectRequest()
	req.ExpectedVersionToken = "  "

	_, err := svc.CorrectAppointmentStatus(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindInvalid)
}

func TestCorrectAppointmentStatus_MissingActor_RejectsAsValidation(t *testing.T) {
	repo := &fakeCorrectRepository{}
	svc := booking.NewCorrectAppointmentStatusService(repo, fakeClock{now: time.Now()})

	req := validCorrectRequest()
	req.ActorStaffUserID = ""

	_, err := svc.CorrectAppointmentStatus(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindValidation)
}

func TestCorrectAppointmentStatus_ValidRequest_CallsRepositoryWithStaffActorTrimmedReasonAndClockNow(t *testing.T) {
	now := time.Date(2026, 9, 10, 15, 0, 0, 0, time.UTC)
	repo := &fakeCorrectRepository{}
	svc := booking.NewCorrectAppointmentStatusService(repo, fakeClock{now: now})

	req := validCorrectRequest()
	req.Reason = "  motivo con espacios  "
	_, err := svc.CorrectAppointmentStatus(context.Background(), "shop-1", req, "key-1", "fp-1")
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}

	got := repo.lastCorrectInput
	if got.AppointmentID != validCorrectAppointmentID {
		t.Fatalf("AppointmentID = %q, want %q", got.AppointmentID, validCorrectAppointmentID)
	}
	if got.DestinationStatus != booking.StatusNoShow {
		t.Fatalf("DestinationStatus = %q, want %q", got.DestinationStatus, booking.StatusNoShow)
	}
	if got.Reason != "motivo con espacios" {
		t.Fatalf("Reason = %q, want recortado sin espacios", got.Reason)
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

func TestCorrectAppointmentStatus_RepositoryResult_PassesThrough(t *testing.T) {
	want := booking.CorrectAppointmentStatusResult{
		Decision:    idempotency.Decision{Outcome: idempotency.OutcomeProceed},
		Appointment: booking.CorrectedAppointment{ID: "appt-9", VersionToken: "new-token", Status: booking.StatusCancelledByBarber},
	}
	repo := &fakeCorrectRepository{
		correctFn: func(context.Context, string, booking.CorrectAppointmentStatusInput, idempotency.Key, idempotency.Fingerprint) (booking.CorrectAppointmentStatusResult, error) {
			return want, nil
		},
	}
	svc := booking.NewCorrectAppointmentStatusService(repo, fakeClock{now: time.Now()})

	req := validCorrectRequest()
	got, err := svc.CorrectAppointmentStatus(context.Background(), "shop-1", req, "key-1", "fp-1")
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}
	if got.Appointment.ID != "appt-9" || got.Appointment.VersionToken != "new-token" {
		t.Fatalf("resultado no pasado tal cual: %+v", got)
	}
}

func TestCorrectAppointmentStatus_RepositoryError_Propagates(t *testing.T) {
	wantErr := apperr.InvalidState("el turno todavía no tiene un resultado terminal para corregir")
	repo := &fakeCorrectRepository{
		correctFn: func(context.Context, string, booking.CorrectAppointmentStatusInput, idempotency.Key, idempotency.Fingerprint) (booking.CorrectAppointmentStatusResult, error) {
			return booking.CorrectAppointmentStatusResult{}, wantErr
		},
	}
	svc := booking.NewCorrectAppointmentStatusService(repo, fakeClock{now: time.Now()})

	req := validCorrectRequest()
	_, err := svc.CorrectAppointmentStatus(context.Background(), "shop-1", req, "key-1", "fp-1")
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}
