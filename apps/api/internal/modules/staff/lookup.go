package staff

import (
	"context"

	"system-barbershop/internal/platform/apperr"
)

// BarberLookup adapta *Service a un puerto mínimo y estable que otro módulo
// (catalog, HU-023) puede consumir sin importar el núcleo ni el repositorio
// de staff: solo esta única operación, expresamente pensada para
// consultarse desde fuera (trabajo requerido §3.1, "una operación pública
// estable de staff"). staff NUNCA importa catalog: la composición ocurre en
// cmd/api, la raíz de composición, que construye un BarberLookup y lo pasa
// donde el otro módulo declare su propio puerto (catalog.BarberPort); Go
// resuelve esa relación de forma puramente estructural, sin que ningún
// paquete de módulo importe al otro.
type BarberLookup struct {
	service *Service
}

// NewBarberLookup construye el adaptador sobre un *Service ya existente
// (el mismo que staff/httpapi usa para sus propios handlers, cmd/api lo
// construye una sola vez).
func NewBarberLookup(service *Service) BarberLookup {
	return BarberLookup{service: service}
}

// Exists informa si existe un barbero con ese id en esa barbería. found es
// false tanto si el barbero no existe como si pertenece a otra barbería
// (RN-TEN-01): Service.Get ya unifica ambos casos en el mismo
// apperr.KindNotFound, que aquí simplemente se traduce a false en vez de
// propagarse como error, porque "no existe para este propósito" no es un
// fallo de la operación que llama a este puerto.
func (l BarberLookup) Exists(ctx context.Context, barbershopID, barberID string) (bool, error) {
	_, err := l.service.Get(ctx, barbershopID, barberID)
	if err != nil {
		if appErr, ok := apperr.As(err); ok && appErr.Kind == apperr.KindNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
