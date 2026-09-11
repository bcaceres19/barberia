package publicbooking

import (
	"context"
	"fmt"
	"strings"

	"system-barbershop/internal/platform/apperr"
)

// Service implementa el caso de uso de HU-090: resolver el enlace público
// de reservas de una barbería a partir de un identificador (slug) no
// confiado, sin sesión y sin que el cliente pueda fijar barbershopId.
type Service struct {
	repo Repository
}

// NewService construye el servicio.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ResolveBarbershop resuelve rawSlug -tal como llega del parámetro de ruta,
// sin decodificar de más ni confiar en su forma- a la barbería pública
// correspondiente. Un slug vacío, demasiado largo, mal formado, desconocido
// o de una barbería no publicable produce EXACTAMENTE el mismo
// apperr.NotFound (CA-090-02): esta función nunca distingue la causa ni en
// el valor de retorno ni en cuánto tarda en decidir "demasiado largo" (se
// rechaza antes de tocar la base, igual de rápido que cualquier otro
// rechazo por forma).
func (s *Service) ResolveBarbershop(ctx context.Context, rawSlug string) (BarbershopProfile, error) {
	if err := ctx.Err(); err != nil {
		return BarbershopProfile{}, apperr.Internal(fmt.Errorf("publicbooking: contexto cancelado antes de resolver slug: %w", err))
	}

	slug := strings.TrimSpace(rawSlug)
	if slug == "" || len(slug) > MaxSlugLength {
		return BarbershopProfile{}, errBarbershopNotPublic()
	}

	profile, found, err := s.repo.ResolveBySlug(ctx, slug)
	if err != nil {
		return BarbershopProfile{}, apperr.Internal(fmt.Errorf("publicbooking: resolver slug: %w", err))
	}
	if !found {
		return BarbershopProfile{}, errBarbershopNotPublic()
	}
	return profile, nil
}

// errBarbershopNotPublic es el único apperr.NotFound que este servicio
// devuelve (CA-090-02): mismo mensaje seguro para forma inválida,
// desconocido y no publicable, sin revelar cuál de los tres ocurrió.
func errBarbershopNotPublic() error {
	return apperr.NotFound("no existe una barbería pública con ese enlace")
}

// ListPublicServices resuelve rawSlug -mismo criterio de recorte y forma
// que ResolveBarbershop- y devuelve una página de los servicios activos y
// asignados de esa barbería (HU-091, CA-091-01). Un slug inválido,
// desconocido o de una barbería no publicable produce el mismo
// apperr.NotFound uniforme que ResolveBarbershop (CA-090-02, reutilizado
// también aquí porque la causa es indistinguible en ambas operaciones).
func (s *Service) ListPublicServices(ctx context.Context, rawSlug string, cursorToken string, limit int) (PublicServiceListResult, error) {
	if err := ctx.Err(); err != nil {
		return PublicServiceListResult{}, apperr.Internal(fmt.Errorf("publicbooking: contexto cancelado antes de listar servicios públicos: %w", err))
	}

	slug := strings.TrimSpace(rawSlug)
	if slug == "" || len(slug) > MaxSlugLength {
		return PublicServiceListResult{}, errBarbershopNotPublic()
	}

	switch {
	case limit <= 0:
		limit = DefaultServiceListLimit
	case limit < MinServiceListLimit:
		limit = MinServiceListLimit
	case limit > MaxServiceListLimit:
		limit = MaxServiceListLimit
	}

	var cursor *ServiceCursor
	if cursorToken != "" {
		decoded, err := DecodeServiceCursor(cursorToken)
		if err != nil {
			return PublicServiceListResult{}, err
		}
		cursor = &decoded
	}

	result, found, err := s.repo.ListPublicServices(ctx, slug, cursor, limit)
	if err != nil {
		return PublicServiceListResult{}, apperr.Internal(fmt.Errorf("publicbooking: listar servicios públicos: %w", err))
	}
	if !found {
		return PublicServiceListResult{}, errBarbershopNotPublic()
	}
	return result, nil
}
