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

	// ListPublicServices resuelve slug (ya recortado, dentro de
	// MaxSlugLength) EXACTAMENTE como ResolveBySlug -mismo
	// public_resolve_barbershop_by_slug, misma indistinción entre forma
	// inválida/desconocido/no publicable- y, si resuelve, lee una página de
	// los servicios ACTIVOS con al menos una asignación vigente (HU-091,
	// CA-091-01) de esa barbería, ordenada por (created_at, id). cursor es
	// nil para la primera página; limit ya llegó clamped al rango
	// [MinServiceListLimit, MaxServiceListLimit] por Service.
	// barbershopFound=false cubre exactamente el mismo universo de causas
	// que ResolveBySlug found=false; en ese caso ListResult queda vacío. No
	// expone barbershopId: cada llamada vuelve a resolver por slug, igual
	// que ResolveBySlug, así que el núcleo de publicbooking nunca retiene ni
	// recibe un identificador interno entre ambas operaciones.
	ListPublicServices(ctx context.Context, slug string, cursor *ServiceCursor, limit int) (result PublicServiceListResult, barbershopFound bool, err error)

	// ListPublicBarbers resuelve slug EXACTAMENTE como ResolveBySlug/
	// ListPublicServices y, si resuelve, lee los barberos con asignación
	// vigente a serviceID (HU-092, CA-092-02) ordenados por
	// (created_at de la asignación, barber.id). serviceID ya llegó validado
	// como forma de UUID por Service (LooksLikePublicServiceID); un
	// serviceID de otra barbería, inexistente, o de un servicio inactivo
	// produce una lista vacía -NUNCA un error distinto- porque el filtro
	// tenant-aware ya lo excluye igual que "sin barberos asignados"
	// (CA-092-03: misma respuesta uniforme, sin distinguir la causa).
	// barbershopFound=false cubre exactamente el mismo universo de causas
	// que ResolveBySlug found=false.
	ListPublicBarbers(ctx context.Context, slug string, serviceID string) (result PublicBarberListResult, barbershopFound bool, err error)
}
