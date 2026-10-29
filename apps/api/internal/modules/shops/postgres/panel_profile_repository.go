package postgres

import (
	"context"
	"fmt"

	"system-barbershop/internal/modules/shops"
	"system-barbershop/internal/platform/database"
)

// PanelProfileRepository implementa shops.PanelProfileRepository sobre la
// tabla staff_user_role (migración 20261029130000_create_staff_user_role.sql).
// Solo lee: la aplicación no tiene INSERT, UPDATE ni DELETE sobre ella.
type PanelProfileRepository struct {
	db *database.DB
}

// NewPanelProfileRepository construye el repositorio de roles.
func NewPanelProfileRepository(db *database.DB) *PanelProfileRepository {
	return &PanelProfileRepository{db: db}
}

var _ shops.PanelProfileRepository = (*PanelProfileRepository)(nil)

// RolesOf implementa shops.PanelProfileRepository.RolesOf. RLS ya restringe las
// filas al tenant vigente; el filtro por barbershop_id es defensa en
// profundidad, mismo patrón que BrandRepository.Get.
func (r *PanelProfileRepository) RolesOf(ctx context.Context, barbershopID, staffUserID string) ([]shops.Role, error) {
	var roles []shops.Role

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		rows, err := q.Query(ctx,
			`SELECT role
			   FROM staff_user_role
			  WHERE barbershop_id = $1 AND staff_user_id = $2`,
			barbershopID, staffUserID)
		if err != nil {
			return fmt.Errorf("select roles: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var role string
			if err := rows.Scan(&role); err != nil {
				return fmt.Errorf("scan role: %w", err)
			}
			roles = append(roles, shops.Role(role))
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("shops/postgres: roles of staff user: %w", err)
	}
	return roles, nil
}
