package catalog

import "context"

// BarberPort es el puerto pequeño y explícito que AssignmentService
// consulta para verificar que un barbero pertenece a la barbería vigente,
// sin que el núcleo de catalog importe el núcleo ni el repositorio de staff
// (HU-023, trabajo requerido §3.1: "La existencia/pertenencia de un
// barbero se consulta mediante un puerto explícito ..., no mediante
// acoplamiento del núcleo a su repositorio"). cmd/api (la raíz de
// composición) conecta este puerto con la implementación real
// (staff.NewBarberLookup(staffService)); catalog nunca importa el paquete
// staff, ni siquiera para este propósito: el adaptador satisface esta
// interfaz de forma puramente estructural.
type BarberPort interface {
	// Exists informa si existe un barbero con ese id en esa barbería.
	// found=false cubre tanto "no existe" como "es de otra barbería"
	// (mismo criterio uniforme que el resto del backend, RN-TEN-01).
	Exists(ctx context.Context, barbershopID, barberID string) (bool, error)
}

// AssignmentRepository es el puerto de persistencia de las asignaciones
// barbero-servicio. El núcleo no importa internal/platform/database ni pgx
// (CA-002-06): postgres/ traduce entre este contrato y database.DB, mismo
// patrón que Repository (ports.go).
type AssignmentRepository interface {
	// List lee una página de asignaciones de barberID dentro de
	// barbershopID, ordenada por (created_at, service_id). cursor es nil
	// para la primera página; limit ya llegó clamped al rango
	// [MinListLimit, MaxListLimit] por AssignmentService. No valida que
	// barberID exista: eso ya lo hizo AssignmentService mediante BarberPort
	// antes de llamar aquí.
	List(ctx context.Context, barbershopID, barberID string, cursor *Cursor, limit int) (AssignmentListResult, error)

	// Assign ejecuta, dentro de UNA sola InTenantTx: verificar que
	// serviceID exista en barbershopID (AssignOutcomeServiceNotFound si
	// no), e INSERT ... ON CONFLICT DO NOTHING (AssignOutcomeCreated u
	// AssignOutcomeAlreadyExists). barberID ya se verificó existente
	// mediante BarberPort antes de esta llamada; la FK compuesta de
	// barber_service es además defensa en profundidad (CA-023-04) por si
	// esa verificación previa quedara desincronizada.
	Assign(ctx context.Context, barbershopID, barberID, serviceID string) (AssignResult, error)

	// Exists informa si barberID tiene asignado serviceID dentro de
	// barbershopID (HU-061, DEC-072): una consulta de existencia simple
	// sobre `barber_service`, sin paginar ni cargar la fila completa.
	Exists(ctx context.Context, barbershopID, barberID, serviceID string) (bool, error)

	// Unassign ejecuta, dentro de UNA sola InTenantTx que bloquea la fila
	// de `service` (SELECT ... FOR UPDATE) para resistir la carrera de dos
	// desasignaciones concurrentes de las dos últimas filas de un mismo
	// servicio (DEC-068): verificar que la asociación exista
	// (UnassignOutcomeNotFound si no), verificar que no sea la última fila
	// activa de un servicio activo (UnassignOutcomeLastActiveConflict si lo
	// es), y DELETE (UnassignOutcomeDeleted en caso contrario).
	Unassign(ctx context.Context, barbershopID, barberID, serviceID string) (UnassignResult, error)
}
