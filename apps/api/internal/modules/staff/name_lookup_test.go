package staff_test

import (
	"context"
	"errors"
	"testing"

	"system-barbershop/internal/modules/staff"
	"system-barbershop/internal/platform/apperr"
)

// TestBarberNameLookup_Name_KnownBarber_ReturnsFullName confirma que
// staff.BarberNameLookup (el puerto que HU-064/booking consulta) delega en
// Service.Get sin necesitar más que ese único método público, mismo
// criterio que BarberLookup.Exists.
func TestBarberNameLookup_Name_KnownBarber_ReturnsFullName(t *testing.T) {
	repo := &fakeRepository{
		getFn: func(ctx context.Context, barbershopID, barberID string) (staff.Barber, bool, error) {
			return staff.Barber{ID: barberID, FullName: "Ana Gómez"}, true, nil
		},
	}
	svc := staff.NewService(repo)
	lookup := staff.NewBarberNameLookup(svc)

	name, found, err := lookup.Name(context.Background(), "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222")
	if err != nil {
		t.Fatalf("Name: %v", err)
	}
	if !found || name != "Ana Gómez" {
		t.Fatalf("Name = %q, found = %v, want %q/true", name, found, "Ana Gómez")
	}
}

func TestBarberNameLookup_Name_UnknownOrCrossTenantBarber_ReturnsFoundFalseWithoutError(t *testing.T) {
	repo := &fakeRepository{
		getFn: func(context.Context, string, string) (staff.Barber, bool, error) {
			return staff.Barber{}, false, nil
		},
	}
	svc := staff.NewService(repo)
	lookup := staff.NewBarberNameLookup(svc)

	name, found, err := lookup.Name(context.Background(), "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222")
	if err != nil {
		t.Fatalf("Name: %v", err)
	}
	if found || name != "" {
		t.Fatalf("Name = %q, found = %v, want vacío/false", name, found)
	}
}

func TestBarberNameLookup_Name_InfrastructureError_Propagates(t *testing.T) {
	repo := &fakeRepository{
		getFn: func(context.Context, string, string) (staff.Barber, bool, error) {
			return staff.Barber{}, false, errors.New("boom")
		},
	}
	svc := staff.NewService(repo)
	lookup := staff.NewBarberNameLookup(svc)

	_, _, err := lookup.Name(context.Background(), "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222")
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected apperr.KindInternal to propagate, got %v", err)
	}
}
