// Pruebas del caso de uso EN AISLAMIENTO, sin PostgreSQL real: cubren el
// rechazo por forma (vacío/demasiado largo), la traducción found=false a
// apperr.NotFound (sin distinguir causa, CA-098-02) y que el hash enviado a
// Repository es SHA-256 hexadecimal del token en claro, nunca el valor
// original. La resolución real vía public_resolve_appointment_token_tenant,
// el aislamiento de tenant y la lectura de la proyección se prueban contra
// PostgreSQL real con dos tenants en
// internal/modules/customeraccess/postgres/repository_test.go
// (docs/03-desarrollo/estrategia-pruebas.md §2).
package customeraccess_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"system-barbershop/internal/modules/customeraccess"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/idempotency"
)

type fakeRepository struct {
	view  customeraccess.AppointmentView
	found bool
	err   error
	calls []string

	target      customeraccess.CancellationTarget
	targetFound bool
	targetCalls []string
}

func (f *fakeRepository) ResolveCancellationTarget(_ context.Context, tokenHash string) (customeraccess.CancellationTarget, bool, error) {
	f.targetCalls = append(f.targetCalls, tokenHash)
	return f.target, f.targetFound, f.err
}

func (f *fakeRepository) ResolveAppointmentByTokenHash(_ context.Context, tokenHash string) (customeraccess.AppointmentView, bool, error) {
	f.calls = append(f.calls, tokenHash)
	return f.view, f.found, f.err
}

func hashOf(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func TestServiceGetAppointmentByTokenHashesBeforeCallingRepository(t *testing.T) {
	view := customeraccess.AppointmentView{BarbershopName: "Barbería Ejemplo"}
	repo := &fakeRepository{view: view, found: true}
	svc := customeraccess.NewService(repo)

	got, err := svc.GetAppointmentByToken(context.Background(), "token-en-claro")
	if err != nil {
		t.Fatalf("GetAppointmentByToken: %v", err)
	}
	if got != view {
		t.Fatalf("view = %+v, want %+v", got, view)
	}
	if len(repo.calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(repo.calls))
	}
	want := hashOf("token-en-claro")
	if repo.calls[0] != want {
		t.Fatalf("hash enviado = %q, want %q (nunca el valor en claro)", repo.calls[0], want)
	}
	if strings.Contains(repo.calls[0], "token-en-claro") {
		t.Fatalf("el hash no debe contener el token en claro: %q", repo.calls[0])
	}
}

func TestServiceGetAppointmentByTokenRejectsEmptyWithoutCallingRepository(t *testing.T) {
	repo := &fakeRepository{}
	svc := customeraccess.NewService(repo)

	_, err := svc.GetAppointmentByToken(context.Background(), "   ")
	assertUniformNotFound(t, err)
	if len(repo.calls) != 0 {
		t.Fatalf("no debía consultar el repositorio con un token vacío, calls = %d", len(repo.calls))
	}
}

func TestServiceGetAppointmentByTokenRejectsOversizedWithoutCallingRepository(t *testing.T) {
	repo := &fakeRepository{}
	svc := customeraccess.NewService(repo)

	oversized := strings.Repeat("a", customeraccess.MaxTokenLength+1)
	_, err := svc.GetAppointmentByToken(context.Background(), oversized)
	assertUniformNotFound(t, err)
	if len(repo.calls) != 0 {
		t.Fatalf("no debía consultar el repositorio con un token demasiado largo, calls = %d", len(repo.calls))
	}
}

func TestServiceGetAppointmentByTokenNotFoundIsUniform(t *testing.T) {
	repo := &fakeRepository{found: false}
	svc := customeraccess.NewService(repo)

	_, err := svc.GetAppointmentByToken(context.Background(), "cualquier-token")
	assertUniformNotFound(t, err)
}

func TestServiceGetAppointmentByTokenRepositoryErrorIsInternal(t *testing.T) {
	repo := &fakeRepository{err: errors.New("fallo de conexión")}
	svc := customeraccess.NewService(repo)

	_, err := svc.GetAppointmentByToken(context.Background(), "cualquier-token")
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("err = %v, want apperr.KindInternal", err)
	}
}

func assertUniformNotFound(t *testing.T, err error) {
	t.Helper()
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindNotFound {
		t.Fatalf("err = %v, want apperr.KindNotFound", err)
	}
}

type fakeCanceller struct {
	calls []cancelCall
	err   error
}

type cancelCall struct {
	shop, appointment, version, reason string
	key                                idempotency.Key
}

func (f *fakeCanceller) CancelAppointment(
	_ context.Context, barbershopID, appointmentID, expectedVersionToken, reason string,
	key idempotency.Key, _ idempotency.Fingerprint,
) (idempotency.Decision, idempotency.StoredResponse, error) {
	f.calls = append(f.calls, cancelCall{barbershopID, appointmentID, expectedVersionToken, reason, key})
	if f.err != nil {
		return idempotency.Decision{}, idempotency.StoredResponse{}, f.err
	}
	return idempotency.Decision{Outcome: idempotency.OutcomeProceed}, idempotency.StoredResponse{Status: 200, Body: `{"status":"cancelled_by_customer"}`}, nil
}

func TestServiceGetAppointmentByTokenEmitsVersionTokenFromInjectedEncoder(t *testing.T) {
	updated := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	repo := &fakeRepository{found: true, view: customeraccess.AppointmentView{AppointmentID: "appt-1", UpdatedAt: updated}}
	svc := customeraccess.NewService(repo, customeraccess.WithVersionTokens(func(id string, at time.Time) string {
		return id + "@" + at.Format(time.RFC3339)
	}))

	got, err := svc.GetAppointmentByToken(context.Background(), "token-en-claro")
	if err != nil {
		t.Fatalf("GetAppointmentByToken: %v", err)
	}
	if got.VersionToken != "appt-1@2026-11-01T10:00:00Z" {
		t.Fatalf("VersionToken = %q", got.VersionToken)
	}
}

func TestServiceCancelAppointmentResolvesTargetByHashAndDelegates(t *testing.T) {
	repo := &fakeRepository{target: customeraccess.CancellationTarget{BarbershopID: "shop-1", AppointmentID: "appt-1"}, targetFound: true}
	canceller := &fakeCanceller{}
	svc := customeraccess.NewService(repo, customeraccess.WithCanceller(canceller))

	decision, stored, err := svc.CancelAppointment(context.Background(), " token-en-claro ", "v1", "me surgió algo", "key-1", "fp-1")
	if err != nil {
		t.Fatalf("CancelAppointment: %v", err)
	}
	if decision.Outcome != idempotency.OutcomeProceed || stored.Status != 200 {
		t.Fatalf("decision=%+v stored=%+v", decision, stored)
	}
	if len(repo.targetCalls) != 1 || repo.targetCalls[0] != hashOf("token-en-claro") {
		t.Fatalf("hash enviado = %v, want SHA-256 del token recortado", repo.targetCalls)
	}
	want := cancelCall{"shop-1", "appt-1", "v1", "me surgió algo", "key-1"}
	if len(canceller.calls) != 1 || canceller.calls[0] != want {
		t.Fatalf("calls = %+v, want %+v", canceller.calls, want)
	}
}

func TestServiceCancelAppointmentHidesWhyTheTokenIsNotUsable(t *testing.T) {
	for name, token := range map[string]string{
		"vacío":         "   ",
		"demasiado":     strings.Repeat("a", customeraccess.MaxTokenLength+1),
		"inexistente":   "token-que-no-existe",
		"hash revocado": "token-revocado",
	} {
		t.Run(name, func(t *testing.T) {
			repo := &fakeRepository{targetFound: false}
			canceller := &fakeCanceller{}
			svc := customeraccess.NewService(repo, customeraccess.WithCanceller(canceller))

			_, _, err := svc.CancelAppointment(context.Background(), token, "v1", "", "key-1", "fp-1")
			appErr, ok := apperr.As(err)
			if !ok || appErr.Kind != apperr.KindNotFound {
				t.Fatalf("esperaba NotFound uniforme, obtuvo %v", err)
			}
			if len(canceller.calls) != 0 {
				t.Fatal("no debió intentar cancelar nada")
			}
		})
	}
}

func TestServiceCancelAppointmentPropagatesPolicyRejection(t *testing.T) {
	repo := &fakeRepository{target: customeraccess.CancellationTarget{BarbershopID: "shop-1", AppointmentID: "appt-1"}, targetFound: true}
	canceller := &fakeCanceller{err: apperr.CancellationWindowClosed("fuera de plazo")}
	svc := customeraccess.NewService(repo, customeraccess.WithCanceller(canceller))

	_, _, err := svc.CancelAppointment(context.Background(), "token", "v1", "", "key-1", "fp-1")
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindCancellationWindowClosed {
		t.Fatalf("esperaba el rechazo de política, obtuvo %v", err)
	}
}

func TestServiceCancelAppointmentRepositoryFailureIsInternal(t *testing.T) {
	repo := &fakeRepository{err: errors.New("boom")}
	svc := customeraccess.NewService(repo, customeraccess.WithCanceller(&fakeCanceller{}))

	_, _, err := svc.CancelAppointment(context.Background(), "token", "v1", "", "key-1", "fp-1")
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("esperaba Internal, obtuvo %v", err)
	}
}
