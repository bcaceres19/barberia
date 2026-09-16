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

	"system-barbershop/internal/modules/customeraccess"
	"system-barbershop/internal/platform/apperr"
)

type fakeRepository struct {
	view  customeraccess.AppointmentView
	found bool
	err   error
	calls []string
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
