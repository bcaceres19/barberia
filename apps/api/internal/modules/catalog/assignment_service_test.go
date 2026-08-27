package catalog_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"system-barbershop/internal/modules/catalog"
	"system-barbershop/internal/platform/apperr"
)

// fakeAssignmentRepository es un doble de catalog.AssignmentRepository:
// prueba solo la orquestación de catalog.AssignmentService (validación de
// forma, consulta al puerto de barbero ANTES de tocar el repositorio,
// traducción de Outcome a apperr), nunca SQL, RLS ni el bloqueo de fila
// real de DEC-068 (eso vive en postgres/assignment_repository_test.go
// contra PostgreSQL real).
type fakeAssignmentRepository struct {
	listFn     func(ctx context.Context, barbershopID, barberID string, cursor *catalog.Cursor, limit int) (catalog.AssignmentListResult, error)
	assignFn   func(ctx context.Context, barbershopID, barberID, serviceID string) (catalog.AssignResult, error)
	unassignFn func(ctx context.Context, barbershopID, barberID, serviceID string) (catalog.UnassignResult, error)
	existsFn   func(ctx context.Context, barbershopID, barberID, serviceID string) (bool, error)

	assignCalls   int
	unassignCalls int
}

func (f *fakeAssignmentRepository) Exists(ctx context.Context, barbershopID, barberID, serviceID string) (bool, error) {
	if f.existsFn == nil {
		return false, nil
	}
	return f.existsFn(ctx, barbershopID, barberID, serviceID)
}

func (f *fakeAssignmentRepository) List(ctx context.Context, barbershopID, barberID string, cursor *catalog.Cursor, limit int) (catalog.AssignmentListResult, error) {
	return f.listFn(ctx, barbershopID, barberID, cursor, limit)
}

func (f *fakeAssignmentRepository) Assign(ctx context.Context, barbershopID, barberID, serviceID string) (catalog.AssignResult, error) {
	f.assignCalls++
	return f.assignFn(ctx, barbershopID, barberID, serviceID)
}

func (f *fakeAssignmentRepository) Unassign(ctx context.Context, barbershopID, barberID, serviceID string) (catalog.UnassignResult, error) {
	f.unassignCalls++
	return f.unassignFn(ctx, barbershopID, barberID, serviceID)
}

var _ catalog.AssignmentRepository = (*fakeAssignmentRepository)(nil)

// fakeBarberPort es un doble de catalog.BarberPort: nunca toca staff ni
// PostgreSQL.
type fakeBarberPort struct {
	existsFn func(ctx context.Context, barbershopID, barberID string) (bool, error)
	calls    int
}

func (f *fakeBarberPort) Exists(ctx context.Context, barbershopID, barberID string) (bool, error) {
	f.calls++
	return f.existsFn(ctx, barbershopID, barberID)
}

var _ catalog.BarberPort = (*fakeBarberPort)(nil)

const (
	fakeShop    = "11111111-1111-1111-1111-111111111111"
	fakeBarber  = "22222222-2222-2222-2222-222222222222"
	fakeService = "33333333-3333-3333-3333-333333333333"
)

func alwaysExists() *fakeBarberPort {
	return &fakeBarberPort{existsFn: func(context.Context, string, string) (bool, error) { return true, nil }}
}

// --- List -------------------------------------------------------------

func TestAssignmentService_List_UnknownBarberID_NotFoundWithoutTouchingRepository(t *testing.T) {
	repo := &fakeAssignmentRepository{}
	barbers := &fakeBarberPort{existsFn: func(context.Context, string, string) (bool, error) { return false, nil }}
	svc := catalog.NewAssignmentService(repo, barbers)

	_, err := svc.List(context.Background(), fakeShop, fakeBarber, "", 0)
	mustBeNotFound(t, err)
	if barbers.calls != 1 {
		t.Fatalf("expected exactly one BarberPort.Exists call, got %d", barbers.calls)
	}
}

func TestAssignmentService_List_MalformedBarberID_NotFoundWithoutCallingPort(t *testing.T) {
	repo := &fakeAssignmentRepository{}
	barbers := &fakeBarberPort{existsFn: func(context.Context, string, string) (bool, error) {
		t.Fatal("BarberPort.Exists should not be called for a malformed id")
		return false, nil
	}}
	svc := catalog.NewAssignmentService(repo, barbers)

	_, err := svc.List(context.Background(), fakeShop, "not-a-uuid", "", 0)
	mustBeNotFound(t, err)
}

func TestAssignmentService_List_KnownBarber_ReturnsRepositoryResult(t *testing.T) {
	want := catalog.AssignmentListResult{
		Items: []catalog.Assignment{{BarberID: fakeBarber, ServiceID: fakeService, CreatedAt: time.Now()}},
	}
	repo := &fakeAssignmentRepository{
		listFn: func(ctx context.Context, barbershopID, barberID string, cursor *catalog.Cursor, limit int) (catalog.AssignmentListResult, error) {
			if barbershopID != fakeShop || barberID != fakeBarber {
				t.Fatalf("unexpected args: %s %s", barbershopID, barberID)
			}
			return want, nil
		},
	}
	svc := catalog.NewAssignmentService(repo, alwaysExists())

	got, err := svc.List(context.Background(), fakeShop, fakeBarber, "", 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got.Items) != 1 || got.Items[0].ServiceID != fakeService {
		t.Fatalf("unexpected result: %+v", got)
	}
}

// --- Assign -------------------------------------------------------------

func TestAssignmentService_Assign_MalformedBarberID_NotFoundWithoutCallingPortOrRepo(t *testing.T) {
	repo := &fakeAssignmentRepository{}
	barbers := &fakeBarberPort{existsFn: func(context.Context, string, string) (bool, error) {
		t.Fatal("BarberPort.Exists should not be called for a malformed id")
		return false, nil
	}}
	svc := catalog.NewAssignmentService(repo, barbers)

	_, err := svc.Assign(context.Background(), fakeShop, "not-a-uuid", fakeService)
	mustBeNotFound(t, err)
	if repo.assignCalls != 0 {
		t.Fatalf("expected Repository.Assign never called, got %d calls", repo.assignCalls)
	}
}

func TestAssignmentService_Assign_MalformedServiceID_NotFoundWithoutCallingPortOrRepo(t *testing.T) {
	repo := &fakeAssignmentRepository{}
	barbers := &fakeBarberPort{existsFn: func(context.Context, string, string) (bool, error) {
		t.Fatal("BarberPort.Exists should not be called when serviceID is already malformed")
		return false, nil
	}}
	svc := catalog.NewAssignmentService(repo, barbers)

	_, err := svc.Assign(context.Background(), fakeShop, fakeBarber, "not-a-uuid")
	mustBeNotFound(t, err)
	if repo.assignCalls != 0 {
		t.Fatalf("expected Repository.Assign never called, got %d calls", repo.assignCalls)
	}
}

func TestAssignmentService_Assign_UnknownBarber_NotFoundWithoutTouchingRepository(t *testing.T) {
	repo := &fakeAssignmentRepository{}
	barbers := &fakeBarberPort{existsFn: func(context.Context, string, string) (bool, error) { return false, nil }}
	svc := catalog.NewAssignmentService(repo, barbers)

	_, err := svc.Assign(context.Background(), fakeShop, fakeBarber, fakeService)
	mustBeNotFound(t, err)
	if repo.assignCalls != 0 {
		t.Fatalf("expected Repository.Assign never called when the barber does not exist, got %d calls", repo.assignCalls)
	}
}

func TestAssignmentService_Assign_UnknownService_NotFound(t *testing.T) {
	repo := &fakeAssignmentRepository{
		assignFn: func(context.Context, string, string, string) (catalog.AssignResult, error) {
			return catalog.AssignResult{Outcome: catalog.AssignOutcomeServiceNotFound}, nil
		},
	}
	svc := catalog.NewAssignmentService(repo, alwaysExists())

	_, err := svc.Assign(context.Background(), fakeShop, fakeBarber, fakeService)
	mustBeNotFound(t, err)
}

func TestAssignmentService_Assign_New_ReturnsCreated(t *testing.T) {
	want := catalog.Assignment{BarberID: fakeBarber, ServiceID: fakeService, CreatedAt: time.Now()}
	repo := &fakeAssignmentRepository{
		assignFn: func(context.Context, string, string, string) (catalog.AssignResult, error) {
			return catalog.AssignResult{Outcome: catalog.AssignOutcomeCreated, Assignment: want}, nil
		},
	}
	svc := catalog.NewAssignmentService(repo, alwaysExists())

	got, err := svc.Assign(context.Background(), fakeShop, fakeBarber, fakeService)
	if err != nil {
		t.Fatalf("Assign: %v", err)
	}
	if got.Outcome != catalog.AssignOutcomeCreated || got.Assignment.ServiceID != fakeService {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestAssignmentService_Assign_Repeated_ReturnsAlreadyExistsWithoutError(t *testing.T) {
	// CA-023-02: repetir exactamente la misma operación no crea una segunda
	// fila; el resultado sigue siendo un éxito, no un conflicto.
	repo := &fakeAssignmentRepository{
		assignFn: func(context.Context, string, string, string) (catalog.AssignResult, error) {
			return catalog.AssignResult{Outcome: catalog.AssignOutcomeAlreadyExists}, nil
		},
	}
	svc := catalog.NewAssignmentService(repo, alwaysExists())

	got, err := svc.Assign(context.Background(), fakeShop, fakeBarber, fakeService)
	if err != nil {
		t.Fatalf("Assign: %v", err)
	}
	if got.Outcome != catalog.AssignOutcomeAlreadyExists {
		t.Fatalf("expected AssignOutcomeAlreadyExists, got %v", got.Outcome)
	}
}

func TestAssignmentService_Assign_RepositoryError_WrapsAsInternal(t *testing.T) {
	repo := &fakeAssignmentRepository{
		assignFn: func(context.Context, string, string, string) (catalog.AssignResult, error) {
			return catalog.AssignResult{}, errors.New("boom")
		},
	}
	svc := catalog.NewAssignmentService(repo, alwaysExists())

	_, err := svc.Assign(context.Background(), fakeShop, fakeBarber, fakeService)
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected apperr.KindInternal, got %v", err)
	}
}

// --- Unassign -----------------------------------------------------------

func TestAssignmentService_Unassign_MalformedIDs_NotFoundWithoutTouchingRepository(t *testing.T) {
	repo := &fakeAssignmentRepository{}
	svc := catalog.NewAssignmentService(repo, alwaysExists())

	_, err := svc.Unassign(context.Background(), fakeShop, "not-a-uuid", fakeService)
	mustBeNotFound(t, err)
	if repo.unassignCalls != 0 {
		t.Fatalf("expected Repository.Unassign never called, got %d calls", repo.unassignCalls)
	}
}

func TestAssignmentService_Unassign_NoSuchAssignment_NotFound(t *testing.T) {
	repo := &fakeAssignmentRepository{
		unassignFn: func(context.Context, string, string, string) (catalog.UnassignResult, error) {
			return catalog.UnassignResult{Outcome: catalog.UnassignOutcomeNotFound}, nil
		},
	}
	svc := catalog.NewAssignmentService(repo, alwaysExists())

	_, err := svc.Unassign(context.Background(), fakeShop, fakeBarber, fakeService)
	mustBeNotFound(t, err)
}

func TestAssignmentService_Unassign_LastActiveAssignment_Conflict(t *testing.T) {
	// DEC-068/CA-023-05: rechazo con un error de conflicto de negocio, no
	// una validación de campo.
	repo := &fakeAssignmentRepository{
		unassignFn: func(context.Context, string, string, string) (catalog.UnassignResult, error) {
			return catalog.UnassignResult{Outcome: catalog.UnassignOutcomeLastActiveConflict}, nil
		},
	}
	svc := catalog.NewAssignmentService(repo, alwaysExists())

	_, err := svc.Unassign(context.Background(), fakeShop, fakeBarber, fakeService)
	mustBeConflict(t, err)
}

func TestAssignmentService_Unassign_Deleted_Succeeds(t *testing.T) {
	repo := &fakeAssignmentRepository{
		unassignFn: func(context.Context, string, string, string) (catalog.UnassignResult, error) {
			return catalog.UnassignResult{Outcome: catalog.UnassignOutcomeDeleted}, nil
		},
	}
	svc := catalog.NewAssignmentService(repo, alwaysExists())

	got, err := svc.Unassign(context.Background(), fakeShop, fakeBarber, fakeService)
	if err != nil {
		t.Fatalf("Unassign: %v", err)
	}
	if got.Outcome != catalog.UnassignOutcomeDeleted {
		t.Fatalf("expected UnassignOutcomeDeleted, got %v", got.Outcome)
	}
}

func TestAssignmentService_Unassign_DoesNotConsultBarberPort(t *testing.T) {
	// A diferencia de List/Assign, Unassign descubre un barbero ajeno o
	// inexistente por la ausencia de la fila de asociación (Repository ya
	// lo traduce a UnassignOutcomeNotFound): no necesita BarberPort.
	repo := &fakeAssignmentRepository{
		unassignFn: func(context.Context, string, string, string) (catalog.UnassignResult, error) {
			return catalog.UnassignResult{Outcome: catalog.UnassignOutcomeDeleted}, nil
		},
	}
	barbers := &fakeBarberPort{existsFn: func(context.Context, string, string) (bool, error) {
		t.Fatal("BarberPort.Exists should not be called by Unassign")
		return false, nil
	}}
	svc := catalog.NewAssignmentService(repo, barbers)

	if _, err := svc.Unassign(context.Background(), fakeShop, fakeBarber, fakeService); err != nil {
		t.Fatalf("Unassign: %v", err)
	}
}
