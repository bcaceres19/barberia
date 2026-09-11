// Pruebas del caso de uso EN AISLAMIENTO, sin PostgreSQL real: cubren
// rangos, coherencia entre campos y la traducción de los desenlaces del
// repositorio. El aislamiento de tenant, el bloqueo optimista real
// (SELECT ... FOR UPDATE) y la atomicidad de la escritura se prueban
// contra PostgreSQL real en
// internal/modules/shops/postgres/booking_policy_repository_test.go
// (docs/03-desarrollo/estrategia-pruebas.md §2).
package shops_test

import (
	"context"
	"errors"
	"testing"

	"system-barbershop/internal/modules/shops"
	"system-barbershop/internal/platform/apperr"
)

type bookingPolicyUpdateCall struct {
	barbershopID string
	input        shops.BookingPolicyUpdateInput
}

type fakeBookingPolicyRepository struct {
	getPolicy shops.BookingPolicy
	getFound  bool
	getErr    error
	getCalls  []string

	updateResult shops.BookingPolicyUpdateResult
	updateErr    error
	updateCalls  []bookingPolicyUpdateCall
}

func (f *fakeBookingPolicyRepository) Get(_ context.Context, barbershopID string) (shops.BookingPolicy, bool, error) {
	f.getCalls = append(f.getCalls, barbershopID)
	return f.getPolicy, f.getFound, f.getErr
}

func (f *fakeBookingPolicyRepository) Update(_ context.Context, barbershopID string, input shops.BookingPolicyUpdateInput) (shops.BookingPolicyUpdateResult, error) {
	f.updateCalls = append(f.updateCalls, bookingPolicyUpdateCall{barbershopID, input})
	return f.updateResult, f.updateErr
}

var _ shops.BookingPolicyRepository = (*fakeBookingPolicyRepository)(nil)

const bookingPolicyShopID = "11111111-1111-1111-1111-111111111111"

// validPolicyInput son los defaults de DEC-083 (CA-093-01), siempre válidos
// como punto de partida: cada prueba de rechazo cambia exactamente un campo.
func validPolicyInput() shops.BookingPolicyInput {
	return shops.BookingPolicyInput{
		MinAdvanceMinutes:              60,
		MaxAdvanceDays:                 3,
		SlotGridMinutes:                15,
		CancellationDeadlineMinutes:    20,
		LateCancellationClientAllowed:  true,
		LateCancellationReasonRequired: true,
		ExpectedVersionToken:           "tok-1",
	}
}

func assertBookingPolicyValidation(t *testing.T, err error) {
	t.Helper()
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindValidation {
		t.Fatalf("expected apperr.KindValidation, got %v", err)
	}
}

func TestBookingPolicyGet_Found_ReturnsPolicy(t *testing.T) {
	want := shops.BookingPolicy{MinAdvanceMinutes: 60, MaxAdvanceDays: 3, SlotGridMinutes: 15, CancellationDeadlineMinutes: 20, LateCancellationClientAllowed: true, LateCancellationReasonRequired: true, VersionToken: "tok-1"}
	repo := &fakeBookingPolicyRepository{getPolicy: want, getFound: true}
	svc := shops.NewBookingPolicyService(repo)

	got, err := svc.Get(context.Background(), bookingPolicyShopID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != want {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
	if len(repo.getCalls) != 1 || repo.getCalls[0] != bookingPolicyShopID {
		t.Fatalf("expected exactly one Get call with barbershopID, got %v", repo.getCalls)
	}
}

func TestBookingPolicyGet_NotFound_ReturnsNotFound(t *testing.T) {
	repo := &fakeBookingPolicyRepository{getFound: false}
	svc := shops.NewBookingPolicyService(repo)

	_, err := svc.Get(context.Background(), bookingPolicyShopID)
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindNotFound {
		t.Fatalf("expected apperr.KindNotFound, got %v", err)
	}
}

func TestBookingPolicyGet_RepositoryError_ReturnsInternal(t *testing.T) {
	repo := &fakeBookingPolicyRepository{getErr: errors.New("boom")}
	svc := shops.NewBookingPolicyService(repo)

	_, err := svc.Get(context.Background(), bookingPolicyShopID)
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected apperr.KindInternal, got %v", err)
	}
}

func TestBookingPolicyGet_ContextCancelled_ReturnsInternalWithoutCallingRepository(t *testing.T) {
	repo := &fakeBookingPolicyRepository{}
	svc := shops.NewBookingPolicyService(repo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.Get(ctx, bookingPolicyShopID)
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected apperr.KindInternal, got %v", err)
	}
	if len(repo.getCalls) != 0 {
		t.Fatalf("expected zero repository calls with a cancelled context, got %v", repo.getCalls)
	}
}

func TestBookingPolicyUpdate_ValidInput_DelegatesToRepository(t *testing.T) {
	updated := shops.BookingPolicy{MinAdvanceMinutes: 60, MaxAdvanceDays: 3, SlotGridMinutes: 15, CancellationDeadlineMinutes: 20, LateCancellationClientAllowed: true, LateCancellationReasonRequired: true, VersionToken: "tok-2"}
	repo := &fakeBookingPolicyRepository{updateResult: shops.BookingPolicyUpdateResult{Outcome: shops.BookingPolicyUpdateOutcomeUpdated, Policy: updated}}
	svc := shops.NewBookingPolicyService(repo)

	got, err := svc.Update(context.Background(), bookingPolicyShopID, validPolicyInput())
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got != updated {
		t.Fatalf("expected %+v, got %+v", updated, got)
	}
	if len(repo.updateCalls) != 1 || repo.updateCalls[0].input.ExpectedVersionToken != "tok-1" {
		t.Fatalf("expected exactly one Update call carrying the version token, got %+v", repo.updateCalls)
	}
}

func TestBookingPolicyUpdate_MissingVersionToken_ReturnsInvalidWithoutCallingRepository(t *testing.T) {
	repo := &fakeBookingPolicyRepository{}
	svc := shops.NewBookingPolicyService(repo)

	input := validPolicyInput()
	input.ExpectedVersionToken = ""
	_, err := svc.Update(context.Background(), bookingPolicyShopID, input)
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInvalid {
		t.Fatalf("expected apperr.KindInvalid, got %v", err)
	}
	if len(repo.updateCalls) != 0 {
		t.Fatalf("expected zero repository calls, got %v", repo.updateCalls)
	}
}

func TestBookingPolicyUpdate_FieldOutOfRange_ReturnsValidationWithoutCallingRepository(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*shops.BookingPolicyInput)
	}{
		{"minAdvanceMinutes below floor", func(i *shops.BookingPolicyInput) { i.MinAdvanceMinutes = -1 }},
		{"minAdvanceMinutes above ceiling", func(i *shops.BookingPolicyInput) { i.MinAdvanceMinutes = 1441 }},
		{"maxAdvanceDays below floor", func(i *shops.BookingPolicyInput) { i.MaxAdvanceDays = 0 }},
		{"maxAdvanceDays above ceiling", func(i *shops.BookingPolicyInput) { i.MaxAdvanceDays = 91 }},
		{"slotGridMinutes not in allowed set", func(i *shops.BookingPolicyInput) { i.SlotGridMinutes = 7 }},
		{"cancellationDeadlineMinutes below floor", func(i *shops.BookingPolicyInput) { i.CancellationDeadlineMinutes = -1 }},
		{"cancellationDeadlineMinutes above ceiling", func(i *shops.BookingPolicyInput) { i.CancellationDeadlineMinutes = 10081 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeBookingPolicyRepository{}
			svc := shops.NewBookingPolicyService(repo)

			input := validPolicyInput()
			tc.mutate(&input)
			_, err := svc.Update(context.Background(), bookingPolicyShopID, input)
			assertBookingPolicyValidation(t, err)
			if len(repo.updateCalls) != 0 {
				t.Fatalf("expected zero repository calls, got %v", repo.updateCalls)
			}
		})
	}
}

func TestBookingPolicyUpdate_AllowedSlotGridValues_AllPass(t *testing.T) {
	for _, grid := range shops.AllowedSlotGridMinutes {
		t.Run("", func(t *testing.T) {
			repo := &fakeBookingPolicyRepository{updateResult: shops.BookingPolicyUpdateResult{Outcome: shops.BookingPolicyUpdateOutcomeUpdated}}
			svc := shops.NewBookingPolicyService(repo)

			input := validPolicyInput()
			input.SlotGridMinutes = grid
			if _, err := svc.Update(context.Background(), bookingPolicyShopID, input); err != nil {
				t.Fatalf("Update with grid=%d: %v", grid, err)
			}
		})
	}
}

func TestBookingPolicyUpdate_ReasonRequiredWithoutClientAllowed_ReturnsValidation(t *testing.T) {
	repo := &fakeBookingPolicyRepository{}
	svc := shops.NewBookingPolicyService(repo)

	input := validPolicyInput()
	input.LateCancellationClientAllowed = false
	input.LateCancellationReasonRequired = true
	_, err := svc.Update(context.Background(), bookingPolicyShopID, input)
	assertBookingPolicyValidation(t, err)
	if len(repo.updateCalls) != 0 {
		t.Fatalf("expected zero repository calls for an incoherent policy, got %v", repo.updateCalls)
	}
}

func TestBookingPolicyUpdate_ClientNotAllowedWithoutReason_IsCoherentAndPasses(t *testing.T) {
	repo := &fakeBookingPolicyRepository{updateResult: shops.BookingPolicyUpdateResult{Outcome: shops.BookingPolicyUpdateOutcomeUpdated}}
	svc := shops.NewBookingPolicyService(repo)

	input := validPolicyInput()
	input.LateCancellationClientAllowed = false
	input.LateCancellationReasonRequired = false
	if _, err := svc.Update(context.Background(), bookingPolicyShopID, input); err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestBookingPolicyUpdate_AdvanceEqualsOrExceedsWindow_ReturnsValidation(t *testing.T) {
	repo := &fakeBookingPolicyRepository{}
	svc := shops.NewBookingPolicyService(repo)

	input := validPolicyInput()
	input.MinAdvanceMinutes = 1440
	input.MaxAdvanceDays = 1 // 1440 minutos de ventana: la anticipación mínima la alcanza exactamente.
	_, err := svc.Update(context.Background(), bookingPolicyShopID, input)
	assertBookingPolicyValidation(t, err)
	if len(repo.updateCalls) != 0 {
		t.Fatalf("expected zero repository calls, got %v", repo.updateCalls)
	}
}

func TestBookingPolicyUpdate_AdvanceBelowWindow_Passes(t *testing.T) {
	repo := &fakeBookingPolicyRepository{updateResult: shops.BookingPolicyUpdateResult{Outcome: shops.BookingPolicyUpdateOutcomeUpdated}}
	svc := shops.NewBookingPolicyService(repo)

	input := validPolicyInput()
	input.MinAdvanceMinutes = 1439
	input.MaxAdvanceDays = 1
	if _, err := svc.Update(context.Background(), bookingPolicyShopID, input); err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestBookingPolicyUpdate_VersionConflict_ReturnsVersionConflict(t *testing.T) {
	repo := &fakeBookingPolicyRepository{updateResult: shops.BookingPolicyUpdateResult{Outcome: shops.BookingPolicyUpdateOutcomeVersionConflict}}
	svc := shops.NewBookingPolicyService(repo)

	_, err := svc.Update(context.Background(), bookingPolicyShopID, validPolicyInput())
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindVersionConflict {
		t.Fatalf("expected apperr.KindVersionConflict, got %v", err)
	}
}

func TestBookingPolicyUpdate_NotFound_ReturnsNotFound(t *testing.T) {
	repo := &fakeBookingPolicyRepository{updateResult: shops.BookingPolicyUpdateResult{Outcome: shops.BookingPolicyUpdateOutcomeNotFound}}
	svc := shops.NewBookingPolicyService(repo)

	_, err := svc.Update(context.Background(), bookingPolicyShopID, validPolicyInput())
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindNotFound {
		t.Fatalf("expected apperr.KindNotFound, got %v", err)
	}
}

func TestBookingPolicyUpdate_RepositoryError_ReturnsInternal(t *testing.T) {
	repo := &fakeBookingPolicyRepository{updateErr: errors.New("boom")}
	svc := shops.NewBookingPolicyService(repo)

	_, err := svc.Update(context.Background(), bookingPolicyShopID, validPolicyInput())
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected apperr.KindInternal, got %v", err)
	}
}

func TestBookingPolicyUpdate_ContextCancelled_ReturnsInternalWithoutCallingRepository(t *testing.T) {
	repo := &fakeBookingPolicyRepository{}
	svc := shops.NewBookingPolicyService(repo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.Update(ctx, bookingPolicyShopID, validPolicyInput())
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected apperr.KindInternal, got %v", err)
	}
	if len(repo.updateCalls) != 0 {
		t.Fatalf("expected zero repository calls with a cancelled context, got %v", repo.updateCalls)
	}
}
