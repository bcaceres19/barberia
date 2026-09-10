package booking_test

import (
	"context"
	"testing"
	"time"

	"system-barbershop/internal/modules/booking"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/idempotency"
)

// fakeManualRepository es un doble de booking.Repository para probar solo
// la orquestación de ManualBookingService (validación de forma,
// interpretación de la zona/instante civil, consulta a los puertos ANTES
// de tocar el repositorio, decisión de reconciliación de DEC-071), nunca
// SQL/RLS/atomicidad real (eso vive en postgres/manual_repository_test.go
// contra PostgreSQL real).
type fakeManualRepository struct {
	findFn   func(ctx context.Context, barbershopID string, phone, email *string) (booking.Customer, bool, error)
	createFn func(ctx context.Context, barbershopID string, input booking.CreateInternalInput, key idempotency.Key, fingerprint idempotency.Fingerprint) (booking.CreateManualResult, error)

	lastInput booking.CreateInternalInput
}

func (f *fakeManualRepository) CreateInternal(context.Context, string, booking.CreateInternalInput) (booking.CreateInternalResult, error) {
	panic("no usado por ManualBookingService")
}

func (f *fakeManualRepository) FindCustomerForReconciliation(ctx context.Context, barbershopID string, phone, email *string) (booking.Customer, bool, error) {
	if f.findFn == nil {
		return booking.Customer{}, false, nil
	}
	return f.findFn(ctx, barbershopID, phone, email)
}

func (f *fakeManualRepository) CreateManual(ctx context.Context, barbershopID string, input booking.CreateInternalInput, key idempotency.Key, fingerprint idempotency.Fingerprint) (booking.CreateManualResult, error) {
	f.lastInput = input
	if f.createFn == nil {
		return booking.CreateManualResult{
			Decision:    idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Appointment: booking.Appointment{ID: "appt-1"},
		}, nil
	}
	return f.createFn(ctx, barbershopID, input, key, fingerprint)
}

func (f *fakeManualRepository) ListDailyAgenda(context.Context, string, string, time.Time, time.Time) ([]booking.DailyAgendaEntry, error) {
	panic("no usado por ManualBookingService")
}

func (f *fakeManualRepository) GetAppointmentDetail(context.Context, string, string) (booking.AppointmentDetail, bool, error) {
	panic("no usado por ManualBookingService")
}

func (f *fakeManualRepository) ListAppointmentHistory(context.Context, string, string, *booking.HistoryCursor, int) ([]booking.HistoryRow, *booking.HistoryCursor, bool, error) {
	panic("no usado por ManualBookingService")
}

func (f *fakeManualRepository) CustomerNames(context.Context, string, []string) (map[string]string, error) {
	panic("no usado por ManualBookingService")
}

func (f *fakeManualRepository) Reschedule(
	context.Context, string, booking.RescheduleInput, idempotency.Key, idempotency.Fingerprint,
) (booking.RescheduleResult, error) {
	panic("no usado por ManualBookingService")
}

func (f *fakeManualRepository) CancelByBarber(
	context.Context, string, booking.CancelAppointmentByBarberInput, idempotency.Key, idempotency.Fingerprint,
) (booking.CancelAppointmentByBarberResult, error) {
	panic("no usado por ManualBookingService")
}

var _ booking.Repository = (*fakeManualRepository)(nil)

type fakeBarberServicePort struct {
	found           bool
	name            string
	durationMinutes int
	priceCents      int64
	currency        string
	err             error
}

func (f fakeBarberServicePort) ActiveAssignedService(context.Context, string, string, string) (string, int, int64, string, bool, error) {
	if f.err != nil {
		return "", 0, 0, "", false, f.err
	}
	return f.name, f.durationMinutes, f.priceCents, f.currency, f.found, nil
}

type fakeBlockCheckPort struct {
	blocked bool
	err     error
}

func (f fakeBlockCheckPort) HasActiveBlock(context.Context, string, string, time.Time, time.Time, string) (bool, error) {
	return f.blocked, f.err
}

type fakeTimezonePort struct {
	timezone string
	err      error
}

func (f fakeTimezonePort) Timezone(context.Context, string) (string, error) {
	return f.timezone, f.err
}

func validManualService() fakeBarberServicePort {
	return fakeBarberServicePort{found: true, name: "Corte clásico", durationMinutes: 30, priceCents: 2000000, currency: "COP"}
}

func validManualRequest() booking.CreateManualAppointmentRequest {
	return booking.CreateManualAppointmentRequest{
		BarberID:         "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4",
		ServiceID:        "6f1a2b3c-4d5e-4f60-8172-8394a5b6c7d8",
		AttendeeName:     "Carlos Restrepo",
		CustomerFullName: "Carlos Restrepo",
		StartsAtLocal:    "2026-09-03T14:30:00",
		ActorStaffUserID: "staff-1",
	}
}

func TestCreateManualAppointment_ValidRequest_BuildsManualOriginAndCallsRepository(t *testing.T) {
	repo := &fakeManualRepository{}
	svc := booking.NewManualBookingService(repo, validManualService(), fakeBlockCheckPort{}, fakeTimezonePort{timezone: "America/Bogota"})

	result, err := svc.CreateManualAppointment(context.Background(), "shop-1", validManualRequest(), "key-1", "fp-1")
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}
	if result.Appointment.ID != "appt-1" {
		t.Fatalf("Appointment.ID = %q, want appt-1", result.Appointment.ID)
	}
	if repo.lastInput.Origin != booking.OriginManual {
		t.Fatalf("Origin = %q, want manual", repo.lastInput.Origin)
	}
	if repo.lastInput.Customer.New == nil || repo.lastInput.Customer.New.FullName != "Carlos Restrepo" {
		t.Fatalf("Customer.New inesperado: %+v", repo.lastInput.Customer)
	}
	wantStart := time.Date(2026, 9, 3, 14, 30, 0, 0, mustLoadLocation(t, "America/Bogota"))
	if !repo.lastInput.StartsAt.Equal(wantStart) {
		t.Fatalf("StartsAt = %v, want %v", repo.lastInput.StartsAt, wantStart)
	}
	wantEnd := wantStart.Add(30 * time.Minute)
	if !repo.lastInput.EndsAt.Equal(wantEnd) {
		t.Fatalf("EndsAt = %v, want %v", repo.lastInput.EndsAt, wantEnd)
	}
	if repo.lastInput.Actor.Type != booking.ActorTypeStaff || repo.lastInput.Actor.StaffUserID == nil || *repo.lastInput.Actor.StaffUserID != "staff-1" {
		t.Fatalf("Actor inesperado: %+v", repo.lastInput.Actor)
	}
}

func TestCreateManualAppointment_MissingBarberID_RejectsWithoutTouchingPorts(t *testing.T) {
	repo := &fakeManualRepository{}
	svc := booking.NewManualBookingService(repo, fakeBarberServicePort{err: errUnexpectedCall(t)}, fakeBlockCheckPort{}, fakeTimezonePort{timezone: "America/Bogota"})

	req := validManualRequest()
	req.BarberID = "  "
	_, err := svc.CreateManualAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindValidation)
}

func TestCreateManualAppointment_StartsAtWithTimezoneSuffix_RejectsAsInvalid(t *testing.T) {
	repo := &fakeManualRepository{}
	svc := booking.NewManualBookingService(repo, validManualService(), fakeBlockCheckPort{}, fakeTimezonePort{timezone: "America/Bogota"})

	req := validManualRequest()
	req.StartsAtLocal = "2026-09-03T14:30:00Z"
	_, err := svc.CreateManualAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	assertKind(t, err, apperr.KindInvalid)
}

func TestCreateManualAppointment_ServiceNotAssigned_ReturnsNotFound(t *testing.T) {
	repo := &fakeManualRepository{}
	svc := booking.NewManualBookingService(repo, fakeBarberServicePort{found: false}, fakeBlockCheckPort{}, fakeTimezonePort{timezone: "America/Bogota"})

	_, err := svc.CreateManualAppointment(context.Background(), "shop-1", validManualRequest(), "key-1", "fp-1")
	assertKind(t, err, apperr.KindNotFound)
}

func TestCreateManualAppointment_ActiveBlock_ReturnsConflictWithoutCallingRepository(t *testing.T) {
	repo := &fakeManualRepository{}
	svc := booking.NewManualBookingService(repo, validManualService(), fakeBlockCheckPort{blocked: true}, fakeTimezonePort{timezone: "America/Bogota"})

	_, err := svc.CreateManualAppointment(context.Background(), "shop-1", validManualRequest(), "key-1", "fp-1")
	assertKind(t, err, apperr.KindConflict)
	if repo.lastInput.BarberID != "" {
		t.Fatalf("no debía llamar a CreateManual con un bloqueo vigente")
	}
}

func TestCreateManualAppointment_PhoneGiven_Found_UsesExistingCustomer(t *testing.T) {
	repo := &fakeManualRepository{
		findFn: func(_ context.Context, _ string, phone, email *string) (booking.Customer, bool, error) {
			if phone == nil || *phone != "+573001234567" {
				t.Fatalf("phone inesperado: %v", phone)
			}
			return booking.Customer{ID: "cust-existente"}, true, nil
		},
	}
	svc := booking.NewManualBookingService(repo, validManualService(), fakeBlockCheckPort{}, fakeTimezonePort{timezone: "America/Bogota"})

	req := validManualRequest()
	phone := "+573001234567"
	req.CustomerPhone = &phone
	_, err := svc.CreateManualAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}
	if repo.lastInput.Customer.ExistingID == nil || *repo.lastInput.Customer.ExistingID != "cust-existente" {
		t.Fatalf("Customer.ExistingID inesperado: %+v", repo.lastInput.Customer)
	}
}

func TestCreateManualAppointment_PhoneGiven_NotFound_CreatesNewCustomer(t *testing.T) {
	repo := &fakeManualRepository{
		findFn: func(context.Context, string, *string, *string) (booking.Customer, bool, error) {
			return booking.Customer{}, false, nil
		},
	}
	svc := booking.NewManualBookingService(repo, validManualService(), fakeBlockCheckPort{}, fakeTimezonePort{timezone: "America/Bogota"})

	req := validManualRequest()
	phone := "+573001234567"
	req.CustomerPhone = &phone
	_, err := svc.CreateManualAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}
	if repo.lastInput.Customer.New == nil || repo.lastInput.Customer.New.Phone == nil || *repo.lastInput.Customer.New.Phone != "+573001234567" {
		t.Fatalf("Customer.New inesperado: %+v", repo.lastInput.Customer)
	}
}

func TestCreateManualAppointment_NoPhoneWithEmail_ReconciliatesByEmail(t *testing.T) {
	called := false
	repo := &fakeManualRepository{
		findFn: func(_ context.Context, _ string, phone, email *string) (booking.Customer, bool, error) {
			called = true
			if phone != nil {
				t.Fatalf("phone debía ser nil cuando el cliente no lo dio")
			}
			if email == nil || *email != "carlos@example.com" {
				t.Fatalf("email inesperado: %v", email)
			}
			return booking.Customer{ID: "cust-por-correo"}, true, nil
		},
	}
	svc := booking.NewManualBookingService(repo, validManualService(), fakeBlockCheckPort{}, fakeTimezonePort{timezone: "America/Bogota"})

	req := validManualRequest()
	email := "Carlos@Example.com"
	req.CustomerEmail = &email
	_, err := svc.CreateManualAppointment(context.Background(), "shop-1", req, "key-1", "fp-1")
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}
	if !called {
		t.Fatal("debía consultar la reconciliación por correo (DEC-071)")
	}
	if repo.lastInput.Customer.ExistingID == nil || *repo.lastInput.Customer.ExistingID != "cust-por-correo" {
		t.Fatalf("Customer.ExistingID inesperado: %+v", repo.lastInput.Customer)
	}
}

func TestCreateManualAppointment_NoPhoneNoEmail_NeverReconciles_AlwaysNewCustomer(t *testing.T) {
	repo := &fakeManualRepository{
		findFn: func(context.Context, string, *string, *string) (booking.Customer, bool, error) {
			t.Fatal("no debía consultar reconciliación sin teléfono ni correo (DEC-071)")
			return booking.Customer{}, false, nil
		},
	}
	svc := booking.NewManualBookingService(repo, validManualService(), fakeBlockCheckPort{}, fakeTimezonePort{timezone: "America/Bogota"})

	_, err := svc.CreateManualAppointment(context.Background(), "shop-1", validManualRequest(), "key-1", "fp-1")
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}
	if repo.lastInput.Customer.New == nil {
		t.Fatalf("Customer.New inesperado: %+v", repo.lastInput.Customer)
	}
}

func mustLoadLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatalf("time.LoadLocation(%q): %v", name, err)
	}
	return loc
}

func assertKind(t *testing.T, err error, kind apperr.Kind) {
	t.Helper()
	if err == nil {
		t.Fatal("err inesperadamente nil")
	}
	appErr, ok := apperr.As(err)
	if !ok {
		t.Fatalf("err no es *apperr.Error: %v", err)
	}
	if appErr.Kind != kind {
		t.Fatalf("Kind = %q, want %q (err: %v)", appErr.Kind, kind, err)
	}
}

func errUnexpectedCall(t *testing.T) error {
	t.Helper()
	return apperr.Internal(nil)
}
