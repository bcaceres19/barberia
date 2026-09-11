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

	listResult publicbooking.PublicServiceListResult
	listFound  bool
	listErr    error
	listCalls  []listCall

	barbersResult publicbooking.PublicBarberListResult
	barbersFound  bool
	barbersErr    error
	barbersCalls  []barbersCall
}

type listCall struct {
	slug   string
	cursor *publicbooking.ServiceCursor
	limit  int
}

type barbersCall struct {
	slug      string
	serviceID string
}

func (f *fakeRepository) ResolveBySlug(_ context.Context, slug string) (publicbooking.BarbershopProfile, bool, error) {
	f.calls = append(f.calls, slug)
	return f.profile, f.found, f.err
}

func (f *fakeRepository) ListPublicServices(_ context.Context, slug string, cursor *publicbooking.ServiceCursor, limit int) (publicbooking.PublicServiceListResult, bool, error) {
	f.listCalls = append(f.listCalls, listCall{slug: slug, cursor: cursor, limit: limit})
	return f.listResult, f.listFound, f.listErr
}

func (f *fakeRepository) ListPublicBarbers(_ context.Context, slug string, serviceID string) (publicbooking.PublicBarberListResult, bool, error) {
	f.barbersCalls = append(f.barbersCalls, barbersCall{slug: slug, serviceID: serviceID})
	return f.barbersResult, f.barbersFound, f.barbersErr
}

var _ publicbooking.Repository = (*fakeRepository)(nil)

const validServiceID = "11111111-1111-1111-1111-111111111111"

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

func TestListPublicServices_Found_ReturnsResult(t *testing.T) {
	want := publicbooking.PublicServiceListResult{
		Items: []publicbooking.PublicService{
			{ID: "1", Name: "Corte", DurationMinutes: 30, PriceCents: 3500000, Currency: "COP"},
		},
	}
	repo := &fakeRepository{listResult: want, listFound: true}
	svc := publicbooking.NewService(repo)

	got, err := svc.ListPublicServices(context.Background(), "barberia-ejemplo", "", 0)
	if err != nil {
		t.Fatalf("ListPublicServices: %v", err)
	}
	if len(got.Items) != 1 || got.Items[0] != want.Items[0] {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
	if len(repo.listCalls) != 1 || repo.listCalls[0].slug != "barberia-ejemplo" || repo.listCalls[0].limit != publicbooking.DefaultServiceListLimit {
		t.Fatalf("expected one call with trimmed slug and default limit, got %+v", repo.listCalls)
	}
}

func TestListPublicServices_NotFound_ReturnsNotFoundWithoutLeakingCause(t *testing.T) {
	repo := &fakeRepository{listFound: false}
	svc := publicbooking.NewService(repo)

	_, err := svc.ListPublicServices(context.Background(), "no-existe", "", 0)
	assertNotFound(t, err)
}

func TestListPublicServices_EmptySlug_ReturnsNotFoundWithoutCallingRepository(t *testing.T) {
	repo := &fakeRepository{}
	svc := publicbooking.NewService(repo)

	_, err := svc.ListPublicServices(context.Background(), "   ", "", 0)
	assertNotFound(t, err)
	if len(repo.listCalls) != 0 {
		t.Fatalf("expected zero repository calls for an empty slug, got %v", repo.listCalls)
	}
}

func TestListPublicServices_SlugTooLong_ReturnsNotFoundWithoutCallingRepository(t *testing.T) {
	repo := &fakeRepository{}
	svc := publicbooking.NewService(repo)

	tooLong := strings.Repeat("a", publicbooking.MaxSlugLength+1)
	_, err := svc.ListPublicServices(context.Background(), tooLong, "", 0)
	assertNotFound(t, err)
	if len(repo.listCalls) != 0 {
		t.Fatalf("expected zero repository calls for an overlong slug, got %v", repo.listCalls)
	}
}

func TestListPublicServices_LimitClamping(t *testing.T) {
	cases := []struct {
		name  string
		input int
		want  int
	}{
		{"zero uses default", 0, publicbooking.DefaultServiceListLimit},
		{"negative uses default", -5, publicbooking.DefaultServiceListLimit},
		{"above maximum clamps down", publicbooking.MaxServiceListLimit + 100, publicbooking.MaxServiceListLimit},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepository{listFound: true}
			svc := publicbooking.NewService(repo)
			if _, err := svc.ListPublicServices(context.Background(), "barberia-ejemplo", "", tc.input); err != nil {
				t.Fatalf("ListPublicServices: %v", err)
			}
			if len(repo.listCalls) != 1 || repo.listCalls[0].limit != tc.want {
				t.Fatalf("expected limit %d, got %+v", tc.want, repo.listCalls)
			}
		})
	}
}

func TestListPublicServices_InvalidCursor_ReturnsInvalidWithoutCallingRepository(t *testing.T) {
	repo := &fakeRepository{}
	svc := publicbooking.NewService(repo)

	_, err := svc.ListPublicServices(context.Background(), "barberia-ejemplo", "no-es-un-cursor-valido", 0)
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInvalid {
		t.Fatalf("expected apperr.KindInvalid, got %v", err)
	}
	if len(repo.listCalls) != 0 {
		t.Fatalf("expected zero repository calls for an invalid cursor, got %v", repo.listCalls)
	}
}

func TestListPublicServices_RepositoryError_ReturnsInternal(t *testing.T) {
	repo := &fakeRepository{listErr: errors.New("boom")}
	svc := publicbooking.NewService(repo)

	_, err := svc.ListPublicServices(context.Background(), "barberia-ejemplo", "", 0)
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected apperr.KindInternal, got %v", err)
	}
}

func TestListPublicServices_ContextCancelled_ReturnsInternalWithoutCallingRepository(t *testing.T) {
	repo := &fakeRepository{}
	svc := publicbooking.NewService(repo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.ListPublicServices(ctx, "barberia-ejemplo", "", 0)
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected apperr.KindInternal, got %v", err)
	}
	if len(repo.listCalls) != 0 {
		t.Fatalf("expected zero repository calls with a cancelled context, got %v", repo.listCalls)
	}
}

func TestListPublicBarbers_Found_ReturnsResult(t *testing.T) {
	want := publicbooking.PublicBarberListResult{
		Items: []publicbooking.PublicBarber{{ID: "1", FullName: "Juan Pérez"}},
	}
	repo := &fakeRepository{barbersResult: want, barbersFound: true}
	svc := publicbooking.NewService(repo)

	got, err := svc.ListPublicBarbers(context.Background(), "barberia-ejemplo", validServiceID)
	if err != nil {
		t.Fatalf("ListPublicBarbers: %v", err)
	}
	if len(got.Items) != 1 || got.Items[0] != want.Items[0] {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
	if len(repo.barbersCalls) != 1 || repo.barbersCalls[0].slug != "barberia-ejemplo" || repo.barbersCalls[0].serviceID != validServiceID {
		t.Fatalf("expected one call with the trimmed slug and serviceID, got %+v", repo.barbersCalls)
	}
}

func TestListPublicBarbers_TrimsSurroundingWhitespaceBeforeCallingRepository(t *testing.T) {
	repo := &fakeRepository{barbersFound: true}
	svc := publicbooking.NewService(repo)

	if _, err := svc.ListPublicBarbers(context.Background(), "  barberia-ejemplo  ", validServiceID); err != nil {
		t.Fatalf("ListPublicBarbers: %v", err)
	}
	if len(repo.barbersCalls) != 1 || repo.barbersCalls[0].slug != "barberia-ejemplo" {
		t.Fatalf("expected the repository to receive the trimmed slug, got %v", repo.barbersCalls)
	}
}

// Cero barberos elegibles (CA-092-03, extremo "0" de la matriz 0/1/N) no es
// un error: el repositorio ya resolvió la barbería, así que el resultado es
// una lista vacía exitosa, indistinguible de un serviceID ajeno o inválido.
func TestListPublicBarbers_ZeroBarbers_ReturnsEmptySuccessResult(t *testing.T) {
	repo := &fakeRepository{barbersResult: publicbooking.PublicBarberListResult{Items: []publicbooking.PublicBarber{}}, barbersFound: true}
	svc := publicbooking.NewService(repo)

	got, err := svc.ListPublicBarbers(context.Background(), "barberia-ejemplo", validServiceID)
	if err != nil {
		t.Fatalf("ListPublicBarbers: %v", err)
	}
	if len(got.Items) != 0 {
		t.Fatalf("expected zero items, got %+v", got.Items)
	}
}

func TestListPublicBarbers_BarbershopNotFound_ReturnsNotFoundWithoutLeakingCause(t *testing.T) {
	repo := &fakeRepository{barbersFound: false}
	svc := publicbooking.NewService(repo)

	_, err := svc.ListPublicBarbers(context.Background(), "no-existe", validServiceID)
	assertNotFound(t, err)
}

func TestListPublicBarbers_EmptySlug_ReturnsNotFoundWithoutCallingRepository(t *testing.T) {
	repo := &fakeRepository{}
	svc := publicbooking.NewService(repo)

	_, err := svc.ListPublicBarbers(context.Background(), "   ", validServiceID)
	assertNotFound(t, err)
	if len(repo.barbersCalls) != 0 {
		t.Fatalf("expected zero repository calls for an empty slug, got %v", repo.barbersCalls)
	}
}

func TestListPublicBarbers_SlugTooLong_ReturnsNotFoundWithoutCallingRepository(t *testing.T) {
	repo := &fakeRepository{}
	svc := publicbooking.NewService(repo)

	tooLong := strings.Repeat("a", publicbooking.MaxSlugLength+1)
	_, err := svc.ListPublicBarbers(context.Background(), tooLong, validServiceID)
	assertNotFound(t, err)
	if len(repo.barbersCalls) != 0 {
		t.Fatalf("expected zero repository calls for an overlong slug, got %v", repo.barbersCalls)
	}
}

// Un serviceID malformado SÍ llega al repositorio (que resuelve la barbería
// primero y descarta el formato después, sin tocar la base con un valor no
// UUID): CA-092-03 exige que esta causa sea indistinguible de "ajeno" o "ya
// no asignado" en el nivel de Service, nunca del universo distinto de
// "barbería desconocida".
func TestListPublicBarbers_MalformedServiceID_StillResolvesBarbershop(t *testing.T) {
	repo := &fakeRepository{barbersFound: true}
	svc := publicbooking.NewService(repo)

	if _, err := svc.ListPublicBarbers(context.Background(), "barberia-ejemplo", "no-es-un-uuid"); err != nil {
		t.Fatalf("ListPublicBarbers: %v", err)
	}
	if len(repo.barbersCalls) != 1 || repo.barbersCalls[0].serviceID != "no-es-un-uuid" {
		t.Fatalf("expected the repository to receive the raw serviceID, got %v", repo.barbersCalls)
	}
}

func TestListPublicBarbers_RepositoryError_ReturnsInternal(t *testing.T) {
	repo := &fakeRepository{barbersErr: errors.New("boom")}
	svc := publicbooking.NewService(repo)

	_, err := svc.ListPublicBarbers(context.Background(), "barberia-ejemplo", validServiceID)
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected apperr.KindInternal, got %v", err)
	}
}

func TestListPublicBarbers_ContextCancelled_ReturnsInternalWithoutCallingRepository(t *testing.T) {
	repo := &fakeRepository{}
	svc := publicbooking.NewService(repo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.ListPublicBarbers(ctx, "barberia-ejemplo", validServiceID)
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected apperr.KindInternal, got %v", err)
	}
	if len(repo.barbersCalls) != 0 {
		t.Fatalf("expected zero repository calls with a cancelled context, got %v", repo.barbersCalls)
	}
}
