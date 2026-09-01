// Pruebas de integración de StaffUserNames (HU-064). Requiere PostgreSQL
// REAL con database/testdata/dos_barberias.sql cargado, mismo criterio que
// repository_test.go.
package postgres_test

import (
	"context"
	"testing"

	authpostgres "system-barbershop/internal/modules/auth/postgres"
)

func TestStaffUserNames_ReturnsFullNameForRequestedIDs(t *testing.T) {
	db := setupTestDB(t)
	repo := authpostgres.New(db)

	names, err := repo.StaffUserNames(context.Background(), shopA, []string{staffUserActiveA, "00000000-0000-0000-0000-000000000000"})
	if err != nil {
		t.Fatalf("StaffUserNames: %v", err)
	}
	if got := names[staffUserActiveA]; got != "Dueña A" {
		t.Fatalf("names[%q] = %q, want %q", staffUserActiveA, got, "Dueña A")
	}
	if _, ok := names["00000000-0000-0000-0000-000000000000"]; ok {
		t.Fatalf("un id inexistente no debió aparecer en el mapa")
	}
}

// TestStaffUserNames_CrossTenantID_NeverAppears verifica RN-TEN-01: un
// staff_user real de shopB nunca aparece al consultar shopA, aunque el id
// coincida literalmente.
func TestStaffUserNames_CrossTenantID_NeverAppears(t *testing.T) {
	db := setupTestDB(t)
	repo := authpostgres.New(db)

	const staffUserActiveB = "bbbbbbb1-bbbb-bbbb-bbbb-bbbbbbbbbbb1"
	names, err := repo.StaffUserNames(context.Background(), shopA, []string{staffUserActiveB})
	if err != nil {
		t.Fatalf("StaffUserNames: %v", err)
	}
	if _, ok := names[staffUserActiveB]; ok {
		t.Fatalf("un staff_user de shopB no debió verse al consultar shopA (RN-TEN-01): %+v", names)
	}
}

func TestStaffUserNames_EmptyInput_ReturnsEmptyMapWithoutQuerying(t *testing.T) {
	db := setupTestDB(t)
	repo := authpostgres.New(db)

	names, err := repo.StaffUserNames(context.Background(), shopA, nil)
	if err != nil {
		t.Fatalf("StaffUserNames: %v", err)
	}
	if len(names) != 0 {
		t.Fatalf("names = %+v, want vacío", names)
	}
}
