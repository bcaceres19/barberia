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
			        has_physical_location, address, latitude, longitude, offers_home_service,
			        google_place_id
			   FROM barbershop
			  WHERE id = $1`,
			barbershopID,
		).Scan(&b.Accent, &b.BusinessTerm, &b.BusinessTermGender,
			&b.ProfessionalTerm, &b.ProfessionalTermPlural, &b.ProfessionalTermGender,
			&b.Location.HasPhysicalLocation, &b.Location.Address, &b.Location.Latitude, &b.Location.Longitude,
			&b.Location.OffersHomeService, &b.Location.GooglePlaceID)
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
// escriben los seis campos y la ubicación pedida, o ninguno. La ubicación se
// conserva salvo que location la cambie (DEC-130): $8 = interruptor ($8 NULL lo
// conserva), $9 = reemplazar dirección y coordenadas a la vez y $13 = interruptor
// de domicilio (DEC-131, NULL lo conserva), $14 = reemplazar la ficha de Google con $15
// (DEC-129; $15 NULL la quita).
func (r *BrandRepository) Update(ctx context.Context, barbershopID string, brand shops.Brand, location shops.LocationWrite) (shops.BrandUpdateResult, error) {
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
			        has_physical_location = COALESCE($8::boolean, has_physical_location),
			        address   = CASE WHEN $9::boolean THEN $10::text ELSE address END,
			        latitude  = CASE WHEN $9::boolean THEN $11::numeric ELSE latitude END,
			        longitude = CASE WHEN $9::boolean THEN $12::numeric ELSE longitude END,
			        offers_home_service = COALESCE($13::boolean, offers_home_service),
			        google_place_id = CASE WHEN $14::boolean THEN $15::text ELSE google_place_id END
			  WHERE id = $1
			RETURNING brand_accent, business_term, business_term_gender,
			          professional_term, professional_term_plural, professional_term_gender,
			          has_physical_location, address, latitude, longitude, offers_home_service,
			          google_place_id`,
			barbershopID,
			brand.Accent, brand.BusinessTerm, string(brand.BusinessTermGender),
			brand.ProfessionalTerm, brand.ProfessionalTermPlural, string(brand.ProfessionalTermGender),
			location.HasPhysicalLocation, location.ReplaceAddress,
			location.Address, location.Latitude, location.Longitude,
			location.OffersHomeService,
			location.ReplaceGooglePlaceID, location.GooglePlaceID,
		).Scan(&b.Accent, &b.BusinessTerm, &b.BusinessTermGender,
			&b.ProfessionalTerm, &b.ProfessionalTermPlural, &b.ProfessionalTermGender,
			&b.Location.HasPhysicalLocation, &b.Location.Address, &b.Location.Latitude, &b.Location.Longitude,
			&b.Location.OffersHomeService, &b.Location.GooglePlaceID)
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
