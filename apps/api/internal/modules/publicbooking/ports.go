package publicbooking

import "context"

// Repository es el puerto de resolución pública del módulo. El núcleo no
// importa internal/platform/database ni pgx (CA-002-06): postgres/ traduce
// entre este contrato y database.DB, exactamente igual que shops.Repository.
type Repository interface {
	// ResolveBySlug resuelve slug (ya recortado, dentro de MaxSlugLength) a
	// la barbería pública correspondiente. found=false cubre un slug con
	// forma inválida, inexistente o de una barbería no publicable
	// (public_slug NULL): las tres causas son indistinguibles a propósito
	// (CA-090-02, RN-TEN-01) y se resuelven en UNA sola llamada mediante
	// public_resolve_barbershop_by_slug + la lectura tenant-aware
	// subsiguiente, sin que el llamador pueda fijar ni inferir
	// barbershopId.
	ResolveBySlug(ctx context.Context, slug string) (profile BarbershopProfile, found bool, err error)
}
