package staff

import (
	"context"

	"system-barbershop/internal/platform/idempotency"
)

// ListResult es una página de barberos ya ordenada de forma estable
// (created_at, id). NextCursor es "" cuando esta página es la última
// (CA-021-02).
type ListResult struct {
	Items      []Barber
	NextCursor string
}

// CreateResult es el desenlace completo de un intento de alta idempotente
// (RN-IDE-01, DEC-043). Decision.Outcome decide qué hizo el llamador:
//   - OutcomeProceed: Barber y Response reflejan la fila recién creada.
//   - OutcomeReplay: Decision.Response ya trae la respuesta original;
//     Response se deja igual a esa misma copia por conveniencia del
//     llamador (evita que el handler tenga que mirar dos campos distintos
//     según el Outcome), Barber queda en su valor cero (no hubo INSERT).
//   - Cualquier otro Outcome: no se ejecutó ningún efecto; Decision.AsError()
//     ya traduce ese caso a apperr, Barber y Response quedan en su valor
//     cero.
type CreateResult struct {
	Decision idempotency.Decision
	Barber   Barber
	Response idempotency.StoredResponse
}

// RenameResult es el desenlace de un intento de renombrado. Found en false
// cubre un barbero inexistente o de otra barbería (CA-021-05): el
// UPDATE ... RETURNING no encontró fila que actualizar, sin distinguir la
// causa.
type RenameResult struct {
	Barber Barber
	Found  bool
}

// Repository es el puerto de persistencia del módulo staff. El núcleo no
// importa internal/platform/database ni pgx (CA-002-06): postgres/ traduce
// entre este contrato y database.DB, exactamente igual que shops.Repository.
//
// Repository.Create es la única operación que además coordina
// internal/platform/idempotency: Begin, el INSERT y Complete DEBEN
// ejecutarse dentro de UNA sola transacción tenant-aware
// (apps/api/README.md, "Patrón obligatorio: idempotencia reutilizable"), así
// que esa orquestación vive en el adaptador postgres (que ya importa
// database.DB), no en Service.
type Repository interface {
	// List lee una página de barberos de barbershopID, ordenada por
	// (created_at, id). cursor es nil para la primera página; limit ya
	// llegó clamped al rango [MinListLimit, MaxListLimit] por Service. Cada
	// consulta filtra explícitamente por barbershopID además de RLS.
	List(ctx context.Context, barbershopID string, cursor *Cursor, limit int) (ListResult, error)

	// Get lee un barbero por id dentro del tenant vigente. found=false
	// cubre tanto "no existe" como "es de otra barbería" (CA-021-05): la
	// consulta filtra por barbershopID además de RLS, sin distinguir la
	// causa en el resultado.
	Get(ctx context.Context, barbershopID, barberID string) (barber Barber, found bool, err error)

	// Create ejecuta el protocolo completo de idempotencia (Begin, INSERT,
	// Complete) dentro de UNA sola InTenantTx. fullName ya llegó
	// normalizado y validado por Service; key y fingerprint ya fueron
	// interpretados por la capa HTTP (httpserver.IdempotencyKeyFromRequest/
	// IdempotencyFingerprint) antes de llegar aquí.
	Create(
		ctx context.Context,
		barbershopID string,
		fullName string,
		key idempotency.Key,
		fingerprint idempotency.Fingerprint,
	) (CreateResult, error)

	// Rename ejecuta `UPDATE ... RETURNING` dentro de una transacción
	// tenant-aware, filtrando por barbershopID además de RLS: nunca crea ni
	// duplica una fila (CA-021-04), y nunca renombra la de otra barbería
	// (CA-021-05).
	Rename(ctx context.Context, barbershopID, barberID, fullName string) (RenameResult, error)
}
