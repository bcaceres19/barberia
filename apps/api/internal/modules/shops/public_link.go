package shops

import (
	"context"
	"fmt"

	"system-barbershop/internal/platform/apperr"
)

// PublicLink es el enlace público de reservas de la barbería activa (DEC-117):
// solo el slug. La URL absoluta la compone el cliente con su propio origen,
// porque el API no conoce el dominio del frontend.
type PublicLink struct {
	Slug string
}

// PublicLinkRepository es el puerto de persistencia del enlace público. El
// núcleo no importa internal/platform/database ni pgx (CA-002-06).
type PublicLinkRepository interface {
	// Ensure devuelve el public_slug de barbershopID y, si todavía es NULL, lo
	// genera `<nombre>-<código>` en la misma transacción tenant-aware, de modo
	// que dos llamadas concurrentes no producen dos slugs. found=false cubre
	// una fila no visible en el tenant vigente (defensivo, mismo criterio que
	// UpdateResult.Found). Nunca reescribe un slug existente.
	Ensure(ctx context.Context, barbershopID string) (slug string, found bool, err error)
}

// PublicLinkService implementa el caso de uso del issue #304: entregar al
// dueño el enlace público de su barbería, generándolo si aún no existe.
type PublicLinkService struct {
	repo PublicLinkRepository
}

// NewPublicLinkService construye el servicio.
func NewPublicLinkService(repo PublicLinkRepository) *PublicLinkService {
	return &PublicLinkService{repo: repo}
}

// Get devuelve el enlace de la barbería barbershopID. barbershopID llega
// siempre de un auth.Principal ya autenticado; este método no acepta ningún
// identificador que el cliente pueda controlar, así que no existe forma de
// pedir el enlace de otra barbería (RN-TEN-01).
func (s *PublicLinkService) Get(ctx context.Context, barbershopID string) (PublicLink, error) {
	if err := ctx.Err(); err != nil {
		return PublicLink{}, apperr.Internal(fmt.Errorf("shops: contexto cancelado antes de leer el enlace público: %w", err))
	}

	slug, found, err := s.repo.Ensure(ctx, barbershopID)
	if err != nil {
		return PublicLink{}, apperr.Internal(fmt.Errorf("shops: obtener enlace público: %w", err))
	}
	if !found {
		return PublicLink{}, apperr.NotFound("barbería no encontrada")
	}
	return PublicLink{Slug: slug}, nil
}
