package shops

import "context"

// BrandUpdateResult es el desenlace de BrandRepository.Update. Found en
// false cubre una barbería no visible en el tenant vigente (defensivo, mismo
// criterio que UpdateResult.Found).
type BrandUpdateResult struct {
	Brand Brand
	Found bool
}

// BrandRepository es el puerto de persistencia de la marca y el vocabulario
// (issue #292). El núcleo no importa internal/platform/database ni pgx
// (CA-002-06), mismo patrón que Repository y BookingPolicyRepository.
type BrandRepository interface {
	// Get lee la marca vigente de barbershopID dentro de una transacción
	// tenant-aware. found=false cubre una fila no visible en el tenant
	// vigente.
	Get(ctx context.Context, barbershopID string) (brand Brand, found bool, err error)

	// Update reemplaza los seis campos de la marca en UNA sola sentencia
	// filtrada por barbershopID (defensa en profundidad sobre RLS) y
	// devuelve la fila guardada con RETURNING. BrandService ya normalizó y
	// validó la entrada: las restricciones CHECK de la migración son la
	// última línea de defensa, no la primera.
	Update(ctx context.Context, barbershopID string, brand Brand) (BrandUpdateResult, error)
}
