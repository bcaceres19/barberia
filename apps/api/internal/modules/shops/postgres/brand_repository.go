package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"system-barbershop/internal/modules/shops"
	"system-barbershop/internal/platform/database"
)

// BrandRepository implementa shops.BrandRepository sobre database.DB. Lee y
// escribe las columnas de marca de `barbershop` (migración
// 20261003120000_add_barbershop_brand.sql) sin tocar las de HU-020 ni las de
// HU-093, y sin versión de concurrencia: son preferencias de presentación,
// la última escritura gana. Comparte el mismo *database.DB.
type BrandRepository struct {
	db *database.DB
}

// NewBrandRepository construye el repositorio de la marca.
func NewBrandRepository(db *database.DB) *BrandRepository {
	return &BrandRepository{db: db}
}

var _ shops.BrandRepository = (*BrandRepository)(nil)

// Get implementa shops.BrandRepository.Get dentro de una transacción
// tenant-aware. RLS (barbershop_select_tenant_policy) ya restringe la fila
// visible a id = current_setting('app.barbershop_id'); el filtro explícito
// WHERE id = $1 es defensa en profundidad, mismo patrón que Repository.Get.
func (r *BrandRepository) Get(ctx context.Context, barbershopID string) (shops.Brand, bool, error) {
	var (
		b     shops.Brand
		found bool
	)

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		err := q.QueryRow(ctx,
			`SELECT brand_accent, business_term, business_term_gender,
			        professional_term, professional_term_plural, professional_term_gender,
			        panel_profile
			   FROM barbershop
			  WHERE id = $1`,
			barbershopID,
		).Scan(&b.Accent, &b.BusinessTerm, &b.BusinessTermGender,
			&b.ProfessionalTerm, &b.ProfessionalTermPlural, &b.ProfessionalTermGender,
			&b.PanelProfile)
		switch {
		case err == nil:
			found = true
		case errors.Is(err, pgx.ErrNoRows):
			// found queda false: defensivo, ver shops.BrandUpdateResult.Found.
		default:
			return fmt.Errorf("select brand: %w", err)
		}
		return nil
	})
	if err != nil {
		return shops.Brand{}, false, fmt.Errorf("shops/postgres: get brand: %w", err)
	}
	return b, found, nil
}

// Update implementa shops.BrandRepository.Update con un único UPDATE ...
// RETURNING filtrado por barbershopID (defensa en profundidad sobre
// barbershop_update_tenant_policy). Una sola sentencia es atómica: o se
// escriben los seis campos (y el perfil, si viene) o ninguno. Un perfil vacío
// conserva el guardado (DEC-115).
func (r *BrandRepository) Update(ctx context.Context, barbershopID string, brand shops.Brand) (shops.BrandUpdateResult, error) {
	var result shops.BrandUpdateResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var b shops.Brand
		err := q.QueryRow(ctx,
			`UPDATE barbershop
			    SET brand_accent = $2,
			        business_term = $3,
			        business_term_gender = $4,
			        professional_term = $5,
			        professional_term_plural = $6,
			        professional_term_gender = $7,
			        panel_profile = COALESCE(NULLIF($8, ''), panel_profile)
			  WHERE id = $1
			RETURNING brand_accent, business_term, business_term_gender,
			          professional_term, professional_term_plural, professional_term_gender,
			          panel_profile`,
			barbershopID,
			brand.Accent, brand.BusinessTerm, string(brand.BusinessTermGender),
			brand.ProfessionalTerm, brand.ProfessionalTermPlural, string(brand.ProfessionalTermGender),
			string(brand.PanelProfile),
		).Scan(&b.Accent, &b.BusinessTerm, &b.BusinessTermGender,
			&b.ProfessionalTerm, &b.ProfessionalTermPlural, &b.ProfessionalTermGender,
			&b.PanelProfile)
		switch {
		case err == nil:
			result.Found = true
			result.Brand = b
		case errors.Is(err, pgx.ErrNoRows):
			// result.Found queda false: defensivo.
		default:
			return fmt.Errorf("update brand: %w", err)
		}
		return nil
	})
	if err != nil {
		return shops.BrandUpdateResult{}, fmt.Errorf("shops/postgres: update brand: %w", err)
	}
	return result, nil
}
