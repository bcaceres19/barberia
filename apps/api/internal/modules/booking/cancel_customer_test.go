package booking_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"system-barbershop/internal/modules/booking"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/idempotency"
)

// TestCustomerCancellationPolicy_Evaluate recorre la tabla de HU-099
// (CA-099-02, CA-099-03): instante (antes / exacto / después del límite /
// ya iniciado) × política tardía (permitida o no) × motivo (obligatorio o no,
// presente o no).
func TestCustomerCancellationPolicy_Evaluate(t *testing.T) {
	startsAt := time.Date(2026, 11, 20, 15, 0, 0, 0, time.UTC)
	limit := startsAt.Add(-20 * time.Minute) // 14:40

	const (
		okCase     = ""
		closedCase = "closed"
		reasonCase = "reason"
	)
	policy := func(deadline int, allowed, reasonRequired bool) booking.CustomerCancellationPolicy {
		return booking.CustomerCancellationPolicy{DeadlineMinutes: deadline, LateClientAllowed: allowed, LateReasonRequired: reasonRequired}
	}

	cases := []struct {
		name   string
		policy booking.CustomerCancellationPolicy
		now    time.Time
		reason string
		want   string
	}{
		{"antes del límite, sin motivo, tardía prohibida", policy(20, false, true), limit.Add(-time.Second), "", okCase},
		{"antes del límite con motivo opcional", policy(20, false, true), limit.Add(-time.Hour), "cambio de planes", okCase},
		{"exactamente en el límite cuenta como dentro de plazo", policy(20, false, true), limit, "", okCase},
		{"un instante después del límite, tardía prohibida", policy(20, false, false), limit.Add(time.Nanosecond), "", closedCase},
		{"después del límite, tardía prohibida aunque haya motivo", policy(20, false, false), limit.Add(5 * time.Minute), "no puedo ir", closedCase},
		{"después del límite, permitida sin motivo obligatorio", policy(20, true, false), limit.Add(5 * time.Minute), "", okCase},
		{"después del límite, permitida con motivo obligatorio y motivo", policy(20, true, true), limit.Add(5 * time.Minute), "no puedo ir", okCase},
		{"después del límite, permitida con motivo obligatorio y sin motivo", policy(20, true, true), limit.Add(5 * time.Minute), "", reasonCase},
		{"turno ya iniciado nunca procede (instante exacto)", policy(20, true, false), startsAt, "", closedCase},
		{"turno ya iniciado nunca procede (después)", policy(20, true, false), startsAt.Add(time.Minute), "tarde", closedCase},
		{"plazo 0 trata todo como tardío: prohibida", policy(0, false, true), startsAt.Add(-24 * time.Hour), "", closedCase},
		{"plazo 0 trata todo como tardío: permitida con motivo", policy(0, true, true), startsAt.Add(-24 * time.Hour), "otro día", okCase},
		{"plazo 0 permitida y motivo obligatorio sin motivo", policy(0, true, true), startsAt.Add(-24 * time.Hour), "", reasonCase},
		{"plazo máximo de 7 días", policy(10080, false, true), startsAt.Add(-10080 * time.Minute), "", okCase},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.policy.Evaluate(startsAt, tc.now, tc.reason)
			switch tc.want {
			case okCase:
				if err != nil {
					t.Fatalf("esperaba éxito, obtuvo %v", err)
				}
			case closedCase:
				assertKind(t, err, apperr.KindCancellationWindowClosed)
			case reasonCase:
				assertKind(t, err, apperr.KindCancellationReasonRequired)
			}
		})
	}
}

type fakeCustomerCancelRepository struct {
	last     booking.CancelAppointmentByCustomerInput
	lastShop string
	calls    int
}

func (f *fakeCustomerCancelRepository) CancelByCustomer(
	_ context.Context, barbershopID string, input booking.CancelAppointmentByCustomerInput, _ idempotency.Key, _ idempotency.Fingerprint,
) (booking.CancelAppointmentByCustomerResult, error) {
	f.calls++
	f.last = input
	f.lastShop = barbershopID
	return booking.CancelAppointmentByCustomerResult{Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed}}, nil
}

func TestCancelAppointmentByCustomer_ValidatesShapeBeforeTouchingRepository(t *testing.T) {
	good := booking.CancelAppointmentByCustomerRequest{AppointmentID: validCancelAppointmentID, ExpectedVersionToken: "v1"}
	cases := []struct {
		name string
		mut  func(*booking.CancelAppointmentByCustomerRequest)
		kind apperr.Kind
	}{
		{"id malformado", func(r *booking.CancelAppointmentByCustomerRequest) { r.AppointmentID = "x" }, apperr.KindNotFound},
		{"sin versión", func(r *booking.CancelAppointmentByCustomerRequest) { r.ExpectedVersionToken = "  " }, apperr.KindInvalid},
		{"motivo demasiado largo", func(r *booking.CancelAppointmentByCustomerRequest) {
			r.Reason = strings.Repeat("ñ", booking.MaxCancellationReasonLength+1)
		}, apperr.KindValidation},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeCustomerCancelRepository{}
			svc := booking.NewCancelAppointmentByCustomerService(repo, nil)
			req := good
			tc.mut(&req)
			_, err := svc.CancelAppointmentByCustomer(context.Background(), "shop-1", req, "key-1", "fp-1")
			assertKind(t, err, tc.kind)
			if repo.calls != 0 {
				t.Fatal("no debió llamar al repositorio")
			}
		})
	}
}

func TestCancelAppointmentByCustomer_PassesTrimmedReasonAndInjectedClock(t *testing.T) {
	fixed := time.Date(2026, 11, 20, 14, 40, 0, 0, time.UTC)
	repo := &fakeCustomerCancelRepository{}
	svc := booking.NewCancelAppointmentByCustomerService(repo, func() time.Time { return fixed })

	_, err := svc.CancelAppointmentByCustomer(context.Background(), "shop-1", booking.CancelAppointmentByCustomerRequest{
		AppointmentID:        validCancelAppointmentID,
		ExpectedVersionToken: " v1 ",
		Reason:               "  me surgió algo  ",
	}, "key-1", "fp-1")
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if repo.lastShop != "shop-1" || repo.last.Reason != "me surgió algo" || repo.last.ExpectedVersionToken != "v1" || !repo.last.Now.Equal(fixed) {
		t.Fatalf("entrada inesperada al repositorio: %+v (shop %q)", repo.last, repo.lastShop)
	}
}
