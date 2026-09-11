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
