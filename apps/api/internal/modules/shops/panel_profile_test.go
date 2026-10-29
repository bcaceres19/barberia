package shops_test

import (
	"context"
	"errors"
	"testing"

	"system-barbershop/internal/modules/shops"
)

type fakePanelProfileRepository struct {
	roles []shops.Role
	err   error
	calls [][2]string
}

func (f *fakePanelProfileRepository) RolesOf(_ context.Context, barbershopID, staffUserID string) ([]shops.Role, error) {
	f.calls = append(f.calls, [2]string{barbershopID, staffUserID})
	return f.roles, f.err
}

func TestPanelProfileFor(t *testing.T) {
	cases := []struct {
		name  string
		roles []shops.Role
		want  shops.PanelProfile
	}{
		{"sin roles conserva el panel completo", nil, shops.PanelProfileShop},
		{"empresa", []shops.Role{shops.RoleEmpresa}, shops.PanelProfileShop},
		{"independiente", []shops.Role{shops.RoleIndependiente}, shops.PanelProfileSolo},
		{"empleado", []shops.Role{shops.RoleEmpleado}, shops.PanelProfileEmployee},
		{"empleado e independiente ve el individual completo", []shops.Role{shops.RoleEmpleado, shops.RoleIndependiente}, shops.PanelProfileSolo},
		{"el orden de los roles no importa", []shops.Role{shops.RoleIndependiente, shops.RoleEmpleado}, shops.PanelProfileSolo},
		{"empresa prevalece sobre independiente", []shops.Role{shops.RoleIndependiente, shops.RoleEmpresa}, shops.PanelProfileShop},
		{"empresa prevalece sobre empleado", []shops.Role{shops.RoleEmpleado, shops.RoleEmpresa}, shops.PanelProfileShop},
		{"un rol desconocido se ignora", []shops.Role{"administrador"}, shops.PanelProfileShop},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shops.PanelProfileFor(tc.roles); got != tc.want {
				t.Fatalf("PanelProfileFor(%v) = %q, want %q", tc.roles, got, tc.want)
			}
		})
	}
}

func TestPanelProfileService_For_ReadsTheRolesOfThatUserInThatShop(t *testing.T) {
	repo := &fakePanelProfileRepository{roles: []shops.Role{shops.RoleEmpleado}}

	got, err := shops.NewPanelProfileService(repo).For(context.Background(), shopID, "staff-1")
	if err != nil {
		t.Fatalf("For: %v", err)
	}
	if got != shops.PanelProfileEmployee {
		t.Fatalf("expected employee, got %q", got)
	}
	if len(repo.calls) != 1 || repo.calls[0] != [2]string{shopID, "staff-1"} {
		t.Fatalf("unexpected repository calls: %v", repo.calls)
	}
}

func TestPanelProfileService_For_RepositoryErrorIsInternal(t *testing.T) {
	repo := &fakePanelProfileRepository{err: errors.New("boom")}

	if _, err := shops.NewPanelProfileService(repo).For(context.Background(), shopID, "staff-1"); err == nil {
		t.Fatal("expected an error when the repository fails")
	}
}

func TestPanelProfileService_For_CanceledContextSkipsTheRepository(t *testing.T) {
	repo := &fakePanelProfileRepository{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := shops.NewPanelProfileService(repo).For(ctx, shopID, "staff-1"); err == nil {
		t.Fatal("expected an error with a canceled context")
	}
	if len(repo.calls) != 0 {
		t.Fatalf("a canceled context must not reach the repository, got %v", repo.calls)
	}
}
