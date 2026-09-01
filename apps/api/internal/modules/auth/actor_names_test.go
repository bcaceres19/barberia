package auth_test

import (
	"context"
	"errors"
	"testing"

	"system-barbershop/internal/modules/auth"
)

// TestStaffActorNameLookup_Names_DelegatesToRepository confirma que
// auth.StaffActorNameLookup (el puerto que HU-064/booking consulta) delega
// en Repository.StaffUserNames sin transformar el resultado.
func TestStaffActorNameLookup_Names_DelegatesToRepository(t *testing.T) {
	repo := &fakeRepository{staffNames: map[string]string{"staff-1": "Ana Gómez"}}
	lookup := auth.NewStaffActorNameLookup(repo)

	names, err := lookup.Names(context.Background(), "shop-1", []string{"staff-1", "staff-2"})
	if err != nil {
		t.Fatalf("Names: %v", err)
	}
	if names["staff-1"] != "Ana Gómez" {
		t.Fatalf("names = %+v, want staff-1 = Ana Gómez", names)
	}
	if _, ok := names["staff-2"]; ok {
		t.Fatalf("names = %+v, want staff-2 ausente (sin coincidencia)", names)
	}
	if len(repo.staffNamesLastIDs) != 2 {
		t.Fatalf("repo.StaffUserNames llamado con %v, want los 2 ids pedidos", repo.staffNamesLastIDs)
	}
}

func TestStaffActorNameLookup_Names_RepositoryError_Propagates(t *testing.T) {
	repo := &fakeRepository{staffNamesErr: errors.New("boom")}
	lookup := auth.NewStaffActorNameLookup(repo)

	_, err := lookup.Names(context.Background(), "shop-1", []string{"staff-1"})
	if err == nil {
		t.Fatalf("Names no propagó el error del repositorio")
	}
}
