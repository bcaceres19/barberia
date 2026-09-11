// Pruebas del caso de uso EN AISLAMIENTO, sin PostgreSQL real: cubren el
// rechazo por forma (largo) y la traducción found=false a apperr.NotFound.
// La resolución real vía public_resolve_barbershop_by_slug, el aislamiento
// de tenant y la lectura del perfil público se prueban contra PostgreSQL
// real con dos tenants en
// internal/modules/publicbooking/postgres/repository_test.go
// (docs/03-desarrollo/estrategia-pruebas.md §2).
package publicbooking_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"system-barbershop/internal/modules/publicbooking"
	"system-barbershop/internal/platform/apperr"
)

type fakeRepository struct {
	profile publicbooking.BarbershopProfile
	found   bool
	err     error
	calls   []string
}

func (f *fakeRepository) ResolveBySlug(_ context.Context, slug string) (publicbooking.BarbershopProfile, bool, error) {
	f.calls = append(f.calls, slug)
	return f.profile, f.found, f.err
}

var _ publicbooking.Repository = (*fakeRepository)(nil)

func assertNotFound(t *testing.T, err error) {
	t.Helper()
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindNotFound {
		t.Fatalf("expected apperr.KindNotFound, got %v", err)
	}
}

func TestResolveBarbershop_Found_ReturnsProfile(t *testing.T) {
	contactEmail := "contacto@ejemplo.test"
	want := publicbooking.BarbershopProfile{
		Name:         "Barbería Ejemplo",
		Timezone:     "America/Bogota",
		ContactEmail: &contactEmail,
	}
	repo := &fakeRepository{profile: want, found: true}
	svc := publicbooking.NewService(repo)

	got, err := svc.ResolveBarbershop(context.Background(), "barberia-ejemplo")
	if err != nil {
		t.Fatalf("ResolveBarbershop: %v", err)
	}
	if got != want {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
	if len(repo.calls) != 1 || repo.calls[0] != "barberia-ejemplo" {
		t.Fatalf("expected exactly one ResolveBySlug call with the trimmed slug, got %v", repo.calls)
	}
}

func TestResolveBarbershop_TrimsSurroundingWhitespaceBeforeCallingRepository(t *testing.T) {
	repo := &fakeRepository{found: true}
	svc := publicbooking.NewService(repo)

	if _, err := svc.ResolveBarbershop(context.Background(), "  barberia-ejemplo  "); err != nil {
		t.Fatalf("ResolveBarbershop: %v", err)
	}
	if len(repo.calls) != 1 || repo.calls[0] != "barberia-ejemplo" {
		t.Fatalf("expected the repository to receive the trimmed slug, got %v", repo.calls)
	}
}

func TestResolveBarbershop_NotFound_ReturnsNotFoundWithoutLeakingCause(t *testing.T) {
	repo := &fakeRepository{found: false}
	svc := publicbooking.NewService(repo)

	_, err := svc.ResolveBarbershop(context.Background(), "no-existe")
	assertNotFound(t, err)
}

func TestResolveBarbershop_EmptySlug_ReturnsNotFoundWithoutCallingRepository(t *testing.T) {
	repo := &fakeRepository{}
	svc := publicbooking.NewService(repo)

	_, err := svc.ResolveBarbershop(context.Background(), "   ")
	assertNotFound(t, err)
	if len(repo.calls) != 0 {
		t.Fatalf("expected zero repository calls for an empty slug, got %v", repo.calls)
	}
}

func TestResolveBarbershop_SlugTooLong_ReturnsNotFoundWithoutCallingRepository(t *testing.T) {
	repo := &fakeRepository{}
	svc := publicbooking.NewService(repo)

	tooLong := strings.Repeat("a", publicbooking.MaxSlugLength+1)
	_, err := svc.ResolveBarbershop(context.Background(), tooLong)
	assertNotFound(t, err)
	if len(repo.calls) != 0 {
		t.Fatalf("expected zero repository calls for an overlong slug, got %v", repo.calls)
	}
}

func TestResolveBarbershop_RepositoryError_ReturnsInternal(t *testing.T) {
	repo := &fakeRepository{err: errors.New("boom")}
	svc := publicbooking.NewService(repo)

	_, err := svc.ResolveBarbershop(context.Background(), "barberia-ejemplo")
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected apperr.KindInternal, got %v", err)
	}
}

func TestResolveBarbershop_ContextCancelled_ReturnsInternalWithoutCallingRepository(t *testing.T) {
	repo := &fakeRepository{}
	svc := publicbooking.NewService(repo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.ResolveBarbershop(ctx, "barberia-ejemplo")
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected apperr.KindInternal, got %v", err)
	}
	if len(repo.calls) != 0 {
		t.Fatalf("expected zero repository calls with a cancelled context, got %v", repo.calls)
	}
}
