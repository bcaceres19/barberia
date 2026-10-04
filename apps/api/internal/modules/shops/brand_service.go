package shops

import (
	"context"
	"fmt"

	"system-barbershop/internal/platform/apperr"
)

// BrandService implementa el caso de uso del issue #292: leer y actualizar la
// marca y el vocabulario de la barbería activa. Vive en shops porque shops es
// el dueño de la configuración de la barbería, pero no comparte Service
// (HU-020) para no ampliar su contrato cerrado de cuatro campos.
type BrandService struct {
	repo BrandRepository
}

// NewBrandService construye el servicio.
func NewBrandService(repo BrandRepository) *BrandService {
	return &BrandService{repo: repo}
}

// Get lee la marca de la barbería barbershopID. barbershopID llega siempre de
// un auth.Principal ya autenticado; este método no acepta ningún
// identificador que el cliente pueda controlar.
func (s *BrandService) Get(ctx context.Context, barbershopID string) (Brand, error) {
	if err := ctx.Err(); err != nil {
		return Brand{}, apperr.Internal(fmt.Errorf("shops: contexto cancelado antes de leer la marca: %w", err))
	}

	brand, found, err := s.repo.Get(ctx, barbershopID)
	if err != nil {
		return Brand{}, apperr.Internal(fmt.Errorf("shops: leer marca: %w", err))
	}
	if !found {
		return Brand{}, apperr.NotFound("barbería no encontrada")
	}
	return brand, nil
}

// Update normaliza y valida input y, si pasa todas las comprobaciones de
// campo, delega en el repositorio la escritura. Los campos se evalúan en el
// orden fijo del contrato; el primer error es el que se devuelve, sin
// ejecutar ninguna escritura (mismo criterio que Service.Update).
func (s *BrandService) Update(ctx context.Context, barbershopID string, input Brand) (Brand, error) {
	if err := ctx.Err(); err != nil {
		return Brand{}, apperr.Internal(fmt.Errorf("shops: contexto cancelado antes de actualizar la marca: %w", err))
	}

	if !IsAllowedAccent(input.Accent) {
		return Brand{}, errBrandAccentInvalid()
	}

	brand := Brand{
		Accent:                 input.Accent,
		BusinessTerm:           NormalizeTerm(input.BusinessTerm),
		BusinessTermGender:     input.BusinessTermGender,
		ProfessionalTerm:       NormalizeTerm(input.ProfessionalTerm),
		ProfessionalTermPlural: NormalizeTerm(input.ProfessionalTermPlural),
		ProfessionalTermGender: input.ProfessionalTermGender,
	}

	switch {
	case !IsValidTerm(brand.BusinessTerm):
		return Brand{}, errBrandTermInvalid("businessTerm")
	case !brand.BusinessTermGender.IsValid():
		return Brand{}, errBrandGenderInvalid("businessTermGender")
	case !IsValidTerm(brand.ProfessionalTerm):
		return Brand{}, errBrandTermInvalid("professionalTerm")
	case !IsValidTerm(brand.ProfessionalTermPlural):
		return Brand{}, errBrandTermInvalid("professionalTermPlural")
	case !brand.ProfessionalTermGender.IsValid():
		return Brand{}, errBrandGenderInvalid("professionalTermGender")
	}

	result, err := s.repo.Update(ctx, barbershopID, brand)
	if err != nil {
		return Brand{}, apperr.Internal(fmt.Errorf("shops: actualizar marca: %w", err))
	}
	if !result.Found {
		return Brand{}, apperr.NotFound("barbería no encontrada")
	}
	return result.Brand, nil
}
