package postgres_test

import (
	"context"
	"slices"
	"testing"

	"system-barbershop/internal/modules/shops"
	shopspostgres "system-barbershop/internal/modules/shops/postgres"
	"system-barbershop/internal/platform/database"
)

// Usuarios de database/testdata/staff_user_role_365.sql sobre el par dedicado
// de este paquete (shopA/shopB).
const (
	userEmpleado      = "365a0001-0000-4000-8000-000000000001"
	userMixto         = "365a0001-0000-4000-8000-000000000002"
	userIndependiente = "365a0001-0000-4000-8000-000000000003"
	userEmpresa       = "365a0001-0000-4000-8000-000000000004"
	userSinRol        = "365a0001-0000-4000-8000-000000000005"
	userDeLaOtraShop  = "365a0001-0000-4000-8000-000000000006"
)

func TestPanelProfileRepository_RolesOf_ReturnsEveryRoleOfTheUser(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.NewPanelProfileRepository(db)

	cases := []struct {
		name string
		user string
		want []shops.Role
	}{
		{"un rol", userEmpleado, []shops.Role{shops.RoleEmpleado}},
		{"empleado e independiente a la vez", userMixto, []shops.Role{shops.RoleEmpleado, shops.RoleIndependiente}},
		{"independiente", userIndependiente, []shops.Role{shops.RoleIndependiente}},
		{"empresa", userEmpresa, []shops.Role{shops.RoleEmpresa}},
		{"sin ningún rol", userSinRol, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repo.RolesOf(context.Background(), shopA, tc.user)
			if err != nil {
				t.Fatalf("RolesOf: %v", err)
			}
			slices.Sort(got)
			if !slices.Equal(got, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

// El panel que resulta de cada combinación de roles, de punta a punta con la base.
func TestPanelProfileService_AgainstPostgres(t *testing.T) {
	db := setupTestDB(t)
	svc := shops.NewPanelProfileService(shopspostgres.NewPanelProfileRepository(db))

	want := map[string]shops.PanelProfile{
		userEmpleado:      shops.PanelProfileEmployee,
		userMixto:         shops.PanelProfileSolo,
		userIndependiente: shops.PanelProfileSolo,
		userEmpresa:       shops.PanelProfileShop,
		userSinRol:        shops.PanelProfileShop,
	}
	for user, profile := range want {
		got, err := svc.For(context.Background(), shopA, user)
		if err != nil || got != profile {
			t.Errorf("user %s: expected %q, got %q err=%v", user, profile, got, err)
		}
	}
}

// RN-TEN-01: dos barberías. Un usuario de shopB no tiene roles visibles desde
// shopA, ni siquiera consultando su identificador, y viceversa.
func TestPanelProfileRepository_RolesOf_TwoTenants_NeverCrossBarbershops(t *testing.T) {
	db := setupTestDB(t)
	repo := shopspostgres.NewPanelProfileRepository(db)

	fromA, err := repo.RolesOf(context.Background(), shopA, userDeLaOtraShop)
	if err != nil || len(fromA) != 0 {
		t.Fatalf("shopA must not see the roles of a shopB user, got %v err=%v", fromA, err)
	}
	fromB, err := repo.RolesOf(context.Background(), shopB, userDeLaOtraShop)
	if err != nil || !slices.Equal(fromB, []shops.Role{shops.RoleIndependiente}) {
		t.Fatalf("shopB must see its own user's role, got %v err=%v", fromB, err)
	}
	crossed, err := repo.RolesOf(context.Background(), shopB, userIndependiente)
	if err != nil || len(crossed) != 0 {
		t.Fatalf("shopB must not see the roles of a shopA user, got %v err=%v", crossed, err)
	}
}

// La aplicación solo lee los roles: sin INSERT, UPDATE ni DELETE (DEC-134).
func TestPanelProfile_TheApplicationRoleCannotWriteRoles(t *testing.T) {
	db := setupTestDB(t)

	for name, sql := range map[string]string{
		"insert": `INSERT INTO staff_user_role (barbershop_id, staff_user_id, role) VALUES ($1, $2, 'empresa')`,
		"update": `UPDATE staff_user_role SET role = 'empresa' WHERE barbershop_id = $1 AND staff_user_id = $2`,
		"delete": `DELETE FROM staff_user_role WHERE barbershop_id = $1 AND staff_user_id = $2`,
	} {
		t.Run(name, func(t *testing.T) {
			err := db.InTenantTx(context.Background(), database.BarbershopID(shopA), func(ctx context.Context, q database.Queries) error {
				_, err := q.Exec(ctx, sql, shopA, userSinRol)
				return err
			})
			if err == nil {
				t.Fatalf("%s: the application role must not be able to write roles", name)
			}
		})
	}
}
