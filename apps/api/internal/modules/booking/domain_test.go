package booking_test

import (
	"context"
	"testing"
	"time"

	"system-barbershop/internal/modules/booking"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/idempotency"
)

func strPtr(s string) *string { return &s }

func validSnapshot() booking.ServiceSnapshot {
	return booking.ServiceSnapshot{
		Name:             "Corte clásico",
		DurationMinutes:  30,
		PriceAmountCents: 2000000,
		Currency:         "COP",
	}
}

func validInput() booking.CreateInternalInput {
	start := time.Date(2027, 6, 1, 10, 0, 0, 0, time.UTC)
	return booking.CreateInternalInput{
		BarberID:     "b1",
		ServiceID:    "s1",
		AttendeeName: "Cliente de prueba",
		StartsAt:     start,
		EndsAt:       start.Add(30 * time.Minute),
		Origin:       booking.OriginManual,
		Service:      validSnapshot(),
		Customer: booking.CustomerInput{
			New: &booking.NewCustomerInput{FullName: "Cliente de prueba"},
		},
		Actor: booking.Actor{Type: booking.ActorTypeStaff, StaffUserID: strPtr("staff1")},
	}
}

func expectValidationErr(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected a validation error, got nil")
	}
	appErr, ok := apperr.As(err)
	if !ok {
		t.Fatalf("expected an apperr.Error, got %T: %v", err, err)
	}
	if appErr.Kind != apperr.KindValidation {
		t.Fatalf("expected KindValidation, got %v", appErr.Kind)
	}
}

// TestStatus_OccupiesSchedule espeja el criterio de estados-citas.md §4:
// confirmed/completed/no_show ocupan agenda; las dos cancelaciones no.
func TestStatus_OccupiesSchedule(t *testing.T) {
	cases := map[booking.Status]bool{
		booking.StatusConfirmed:           true,
		booking.StatusCompleted:           true,
		booking.StatusNoShow:              true,
		booking.StatusCancelledByCustomer: false,
		booking.StatusCancelledByBarber:   false,
	}
	for status, want := range cases {
		if got := status.OccupiesSchedule(); got != want {
			t.Errorf("Status(%q).OccupiesSchedule() = %v, want %v", status, got, want)
		}
	}
}

func TestBookingService_CreateInternal_ValidInputPassesValidation(t *testing.T) {
	stub := &stubRepository{}
	svc := booking.NewBookingService(stub)
	_, err := svc.CreateInternal(context.Background(), "shop1", validInput())
	if err != nil {
		t.Fatalf("unexpected validation error for a valid input: %v", err)
	}
	if !stub.called {
		t.Fatalf("expected the repository to be called once validation passes")
	}
}

func TestBookingService_CreateInternal_RejectsInvalidInputWithoutCallingRepository(t *testing.T) {
	cases := map[string]func(booking.CreateInternalInput) booking.CreateInternalInput{
		"empty barber id": func(in booking.CreateInternalInput) booking.CreateInternalInput {
			in.BarberID = ""
			return in
		},
		"empty service id": func(in booking.CreateInternalInput) booking.CreateInternalInput {
			in.ServiceID = ""
			return in
		},
		"blank attendee name": func(in booking.CreateInternalInput) booking.CreateInternalInput {
			in.AttendeeName = "   "
			return in
		},
		"attendee name too long": func(in booking.CreateInternalInput) booking.CreateInternalInput {
			in.AttendeeName = repeatChar('a', 121)
			return in
		},
		"invalid origin": func(in booking.CreateInternalInput) booking.CreateInternalInput {
			in.Origin = "walk_in"
			return in
		},
		"ends before starts": func(in booking.CreateInternalInput) booking.CreateInternalInput {
			in.EndsAt = in.StartsAt.Add(-time.Minute)
			return in
		},
		"ends equal starts": func(in booking.CreateInternalInput) booking.CreateInternalInput {
			in.EndsAt = in.StartsAt
			return in
		},
		"duration mismatch": func(in booking.CreateInternalInput) booking.CreateInternalInput {
			in.Service.DurationMinutes = 45
			return in
		},
		"duration out of range": func(in booking.CreateInternalInput) booking.CreateInternalInput {
			in.Service.DurationMinutes = 1500
			in.EndsAt = in.StartsAt.Add(1500 * time.Minute)
			return in
		},
		"negative price": func(in booking.CreateInternalInput) booking.CreateInternalInput {
			in.Service.PriceAmountCents = -1
			return in
		},
		"invalid currency": func(in booking.CreateInternalInput) booking.CreateInternalInput {
			in.Service.Currency = "cop"
			return in
		},
		"customer note too long": func(in booking.CreateInternalInput) booking.CreateInternalInput {
			note := repeatChar('n', 501)
			in.CustomerNote = &note
			return in
		},
		"blank service name snapshot": func(in booking.CreateInternalInput) booking.CreateInternalInput {
			in.Service.Name = "   "
			return in
		},
		"customer input both existing and new": func(in booking.CreateInternalInput) booking.CreateInternalInput {
			in.Customer = booking.CustomerInput{
				ExistingID: strPtr("c1"),
				New:        &booking.NewCustomerInput{FullName: "Cliente de prueba"},
			}
			return in
		},
		"customer input neither existing nor new": func(in booking.CreateInternalInput) booking.CreateInternalInput {
			in.Customer = booking.CustomerInput{}
			return in
		},
		"new customer blank name": func(in booking.CreateInternalInput) booking.CreateInternalInput {
			in.Customer = booking.CustomerInput{New: &booking.NewCustomerInput{FullName: "   "}}
			return in
		},
		"new customer invalid phone": func(in booking.CreateInternalInput) booking.CreateInternalInput {
			in.Customer = booking.CustomerInput{New: &booking.NewCustomerInput{
				FullName: "Cliente de prueba", Phone: strPtr("no-es-telefono"),
			}}
			return in
		},
		"new customer invalid email": func(in booking.CreateInternalInput) booking.CreateInternalInput {
			in.Customer = booking.CustomerInput{New: &booking.NewCustomerInput{
				FullName: "Cliente de prueba", Email: strPtr("Mayusculas@Ejemplo.test"),
			}}
			return in
		},
		"actor staff without id": func(in booking.CreateInternalInput) booking.CreateInternalInput {
			in.Actor = booking.Actor{Type: booking.ActorTypeStaff}
			return in
		},
		"actor system with id": func(in booking.CreateInternalInput) booking.CreateInternalInput {
			in.Actor = booking.Actor{Type: booking.ActorTypeSystem, StaffUserID: strPtr("staff1")}
			return in
		},
		"actor invalid type": func(in booking.CreateInternalInput) booking.CreateInternalInput {
			in.Actor = booking.Actor{Type: "owner"}
			return in
		},
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			stub := &stubRepository{}
			svc := booking.NewBookingService(stub)
			_, err := svc.CreateInternal(context.Background(), "shop1", mutate(validInput()))
			expectValidationErr(t, err)
			if stub.called {
				t.Fatalf("expected the repository to NOT be called when validation fails")
			}
		})
	}
}

func repeatChar(c byte, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = c
	}
	return string(b)
}

type stubRepository struct {
	called bool
}

func (s *stubRepository) CreateInternal(
	_ context.Context, _ string, _ booking.CreateInternalInput,
) (booking.CreateInternalResult, error) {
	s.called = true
	return booking.CreateInternalResult{}, nil
}

func (s *stubRepository) FindCustomerForReconciliation(
	_ context.Context, _ string, _, _ *string,
) (booking.Customer, bool, error) {
	return booking.Customer{}, false, nil
}

func (s *stubRepository) CreateManual(
	_ context.Context, _ string, _ booking.CreateInternalInput, _ idempotency.Key, _ idempotency.Fingerprint,
) (booking.CreateManualResult, error) {
	s.called = true
	return booking.CreateManualResult{}, nil
}

func (s *stubRepository) ListDailyAgenda(
	_ context.Context, _, _ string, _, _ time.Time,
) ([]booking.DailyAgendaEntry, error) {
	panic("no usado")
}
