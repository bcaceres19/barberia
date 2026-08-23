package shops

import "context"

// UpdateInput es la forma ya normalizada y validada en cuanto a campo que
// [Service.Update] pasa al repositorio: Name recortado, Timezone recortado
// (todavía no confirmado contra pg_timezone_names, eso lo hace el
// repositorio dentro de la transacción), y ContactEmail/ContactPhone ya
// normalizados a nil cuando el formulario los deja vacíos.
type UpdateInput struct {
	Name         string
	Timezone     string
	ContactEmail *string
	ContactPhone *string
}

// UpdateResult es el desenlace de un intento de actualización. TimezoneValid
// en false significa cero escritura (CA-020-03): ni Name ni Contact* se
// tocaron, y Barbershop queda en su valor cero. Found en false cubre una
// barbería no visible en el tenant vigente (defensivo: un principal
// autenticado siempre tiene su propia fila, pero el contrato documenta 404
// para este caso igual que GET, sin asumir que nunca puede ocurrir).
type UpdateResult struct {
	Barbershop    Barbershop
	TimezoneValid bool
	Found         bool
}

// Repository es el puerto de persistencia del módulo shops. El núcleo no
// importa internal/platform/database ni pgx (CA-002-06): postgres/ traduce
// entre este contrato y database.DB, exactamente igual que auth.Repository.
type Repository interface {
	// Get lee los cuatro campos autorizados de la barbería barbershopID
	// dentro de una transacción tenant-aware. found=false cubre una fila no
	// visible en el tenant vigente (defensivo, ver [UpdateResult.Found]).
	Get(ctx context.Context, barbershopID string) (barbershop Barbershop, found bool, err error)

	// Update comprueba timezone contra pg_timezone_names y, si es válida,
	// ejecuta el UPDATE (filtrando también por barbershopID, defensa en
	// profundidad sobre RLS) dentro de UNA sola transacción tenant-aware:
	// zona inválida implica cero escritura (CA-020-03), nunca una escritura
	// parcial de Name o Contact*.
	Update(ctx context.Context, barbershopID string, input UpdateInput) (UpdateResult, error)
}
