package staff_test

import (
	"context"
	"errors"
	"testing"

	"system-barbershop/internal/modules/staff"
	"system-barbershop/internal/platform/apperr"
)

// TestBarberLookup_Exists_KnownBarber_ReturnsTrue confirma que
// staff.BarberLookup (el puerto que HU-023/catalog consulta) delega en
// Service.Get sin necesitar más que ese único método público.
func TestBarberLookup_Exists_KnownBarber_ReturnsTrue(t *testing.T) {
	repo := &fakeRepository{
		getFn: func(ctx context.Context, barbershopID, barberID string) (staff.Barber, bool, error) {
			return staff.Barber{ID: barberID, FullName: "Alguien"}, true, nil
		},
	}
	svc := staff.NewService(repo)
	lookup := staff.NewBarberLookup(svc)

	exists, err := lookup.Exists(context.Background(), "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !exists {
		t.Fatal("expected exists=true for a known barber")
	}
}

func TestBarberLookup_Exists_UnknownOrCrossTenantBarber_ReturnsFalseWithoutError(t *testing.T) {
	// CA-023-04: un barbero inexistente o de otra barbería no es un error
	// de infraestructura para quien consulta este puerto, es simplemente
	// "no existe" -- Service.Get ya unifica ambos casos (RN-TEN-01).
	repo := &fakeRepository{
		getFn: func(context.Context, string, string) (staff.Barber, bool, error) {
			return staff.Barber{}, false, nil
		},
	}
	svc := staff.NewService(repo)
	lookup := staff.NewBarberLookup(svc)

	exists, err := lookup.Exists(context.Background(), "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if exists {
		t.Fatal("expected exists=false for an unknown/cross-tenant barber")
	}
}

func TestBarberLookup_Exists_MalformedID_ReturnsFalseWithoutError(t *testing.T) {
	// Service.Get ya traduce una forma inválida al mismo apperr.NotFound
	// que "no existe" (staff/service.go); BarberLookup no necesita repetir
	// esa validación.
	svc := staff.NewService(&fakeRepository{})
	lookup := staff.NewBarberLookup(svc)

	exists, err := lookup.Exists(context.Background(), "11111111-1111-1111-1111-111111111111", "not-a-uuid")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if exists {
		t.Fatal("expected exists=false for a malformed id")
	}
}

func TestBarberLookup_Exists_InfrastructureError_Propagates(t *testing.T) {
	repo := &fakeRepository{
		getFn: func(context.Context, string, string) (staff.Barber, bool, error) {
			return staff.Barber{}, false, errors.New("boom")
		},
	}
	svc := staff.NewService(repo)
	lookup := staff.NewBarberLookup(svc)

	_, err := lookup.Exists(context.Background(), "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222")
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected apperr.KindInternal to propagate, got %v", err)
	}
}
