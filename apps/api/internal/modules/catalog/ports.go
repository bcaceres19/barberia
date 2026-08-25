package catalog

import (
	"context"

	"system-barbershop/internal/platform/idempotency"
)

// ListResult es una página de servicios ya ordenada de forma estable
// (created_at, id). NextCursor es "" cuando esta página es la última
// (CA-022-01).
type ListResult struct {
	Items      []Service
	NextCursor string
}

// CreateInput es la entrada ya normalizada y validada de un alta de
// servicio (CatalogService.Create la construye; Repository.Create nunca vuelve a
// validar formato, solo aplica las restricciones que dependen de otras
// filas, como el nombre único entre activos).
type CreateInput struct {
	Name            string
	Description     *string
	DurationMinutes int
	PriceCents      int64
}

// CreateResult es el desenlace completo de un intento de alta idempotente
// (RN-IDE-01, DEC-043), mismo patrón que staff.CreateResult:
//   - OutcomeProceed: Service y Response reflejan la fila recién creada.
//   - OutcomeReplay: Decision.Response ya trae la respuesta original;
//     Response se deja igual a esa misma copia por conveniencia del
//     llamador, Service queda en su valor cero.
//   - Cualquier otro Outcome (incluido NameTaken): Decision.AsError() o el
//     apperr correspondiente a NameTaken ya traduce el caso; Service y
//     Response quedan en su valor cero.
//
// NameTaken es true cuando el INSERT chocó con idx_service_active_name
// (DEC-067): la transacción entera hizo ROLLBACK, incluida la reclamación
// de idempotencia que Begin había tomado, así que la clave queda libre para
// un reintento legítimo con un nombre distinto (mismo criterio que
// CA-004-06).
type CreateResult struct {
	Decision  idempotency.Decision
	Service   Service
	Response  idempotency.StoredResponse
	NameTaken bool
}

// UpdateFields es la entrada ya normalizada y validada de una edición
// parcial (CatalogService.Update la construye). Un puntero nil significa "el
// cliente no envió este campo, no lo toques"; Description usa
// OptionalDescription para distinguir además "enviado como vacío: borrar la
// descripción existente" de "no enviado".
type UpdateFields struct {
	Name            *string
	Description     OptionalDescription
	DurationMinutes *int
	PriceCents      *int64
}

// HasAny informa si al menos un campo de catálogo viene presente (CA-022-04:
// un PATCH sin ningún campo se rechaza antes de tocar el repositorio).
func (f UpdateFields) HasAny() bool {
	return f.Name != nil || f.Description.Set || f.DurationMinutes != nil || f.PriceCents != nil
}

// OptionalDescription representa el valor de `description` en un PATCH:
// Set=false significa "campo ausente, no tocar"; Set=true con Value=nil
// significa "borrar la descripción existente"; Set=true con Value=&s
// significa "reemplazar por s".
type OptionalDescription struct {
	Set   bool
	Value *string
}

// UpdateResult es el desenlace de un intento de edición. Found en false
// cubre un servicio inexistente o de otra barbería (CA-022-06). NameTaken
// es true cuando el UPDATE chocó con idx_service_active_name al cambiar el
// nombre (DEC-067); en ese caso Found también queda false porque el UPDATE
// no aplicó ningún cambio.
type UpdateResult struct {
	Service   Service
	Found     bool
	NameTaken bool
}

// LifecycleResult es el desenlace completo de un intento de
// desactivación/reactivación idempotente (RN-IDE-01, DEC-043), mismo patrón
// que CreateResult:
//   - OutcomeProceed: Service refleja la fila ya transicionada.
//   - OutcomeReplay: Decision.Response ya trae la respuesta original.
//   - Cualquier otro Outcome (incluido InvalidTransition): Decision.AsError()
//     o el apperr correspondiente a InvalidTransition ya traduce el caso.
//
// Found es false cuando serviceID no existe o pertenece a otra barbería
// (mismo apperr.NotFound uniforme que Get/Update, CA-024-07); solo es
// significativo cuando Decision.Outcome == OutcomeProceed. InvalidTransition
// es true cuando el servicio SÍ existe pero el UPDATE condicionado
// (is_active = NOT destino) no afectó ninguna fila porque ya estaba en el
// estado destino (una clave de idempotencia nueva sobre una transición que
// ya no aplica): la transacción entera hace ROLLBACK, incluida la
// reclamación de idempotencia que Begin había tomado, así que la clave
// queda libre para un reintento legítimo (mismo criterio que
// CreateResult.NameTaken).
type LifecycleResult struct {
	Decision          idempotency.Decision
	Service           Service
	Response          idempotency.StoredResponse
	Found             bool
	InvalidTransition bool
}

// Repository es el puerto de persistencia del módulo catalog. El núcleo no
// importa internal/platform/database ni pgx (CA-002-06): postgres/ traduce
// entre este contrato y database.DB, mismo patrón que staff.Repository.
//
// Repository.Create es la única operación que además coordina
// internal/platform/idempotency: Begin, el INSERT y Complete DEBEN
// ejecutarse dentro de UNA sola transacción tenant-aware (apps/api/README.md,
// "Patrón obligatorio: idempotencia reutilizable").
type Repository interface {
	// List lee una página de servicios de barbershopID, ordenada por
	// (created_at, id). cursor es nil para la primera página; limit ya
	// llegó clamped al rango [MinListLimit, MaxListLimit] por Service. Cada
	// consulta filtra explícitamente por barbershopID además de RLS.
	List(ctx context.Context, barbershopID string, cursor *Cursor, limit int) (ListResult, error)

	// Get lee un servicio por id dentro del tenant vigente. found=false
	// cubre tanto "no existe" como "es de otra barbería" (CA-022-06): la
	// consulta filtra por barbershopID además de RLS, sin distinguir la
	// causa en el resultado.
	Get(ctx context.Context, barbershopID, serviceID string) (service Service, found bool, err error)

	// Create ejecuta el protocolo completo de idempotencia (Begin, INSERT,
	// Complete) dentro de UNA sola InTenantTx. input ya llegó normalizado y
	// validado por Service; key y fingerprint ya fueron interpretados por
	// la capa HTTP.
	Create(
		ctx context.Context,
		barbershopID string,
		input CreateInput,
		key idempotency.Key,
		fingerprint idempotency.Fingerprint,
	) (CreateResult, error)

	// Update ejecuta `UPDATE ... RETURNING` dentro de una transacción
	// tenant-aware, filtrando por barbershopID además de RLS: nunca crea ni
	// duplica una fila, y nunca edita la de otra barbería (CA-022-06).
	// fields ya llegó normalizado y validado por Service.
	Update(ctx context.Context, barbershopID, serviceID string, fields UpdateFields) (UpdateResult, error)

	// Deactivate ejecuta el protocolo completo de idempotencia (Begin,
	// UPDATE condicionado a is_active = true, Complete) dentro de UNA sola
	// InTenantTx (HU-024, CA-024-02, CA-024-03, CA-024-06). Un UPDATE que no
	// afecta ninguna fila porque el servicio ya está inactivo se traduce a
	// LifecycleResult.InvalidTransition; found=false (dentro de Service
	// vacío) cubre "no existe" o "es de otra barbería" mediante el mismo
	// apperr.NotFound uniforme que Get.
	Deactivate(
		ctx context.Context,
		barbershopID, serviceID string,
		key idempotency.Key,
		fingerprint idempotency.Fingerprint,
	) (LifecycleResult, error)

	// Reactivate es el mismo protocolo que Deactivate, en sentido inverso
	// (HU-024, CA-024-05, CA-024-06): UPDATE condicionado a is_active =
	// false. Nunca crea otra fila ni toca duration_minutes, price_amount ni
	// barber_service.
	Reactivate(
		ctx context.Context,
		barbershopID, serviceID string,
		key idempotency.Key,
		fingerprint idempotency.Fingerprint,
	) (LifecycleResult, error)
}
