package schedule

import (
	"context"

	"system-barbershop/internal/platform/idempotency"
)

// BarberPort es el puerto pequeño y explícito que Service consulta para
// verificar que un barbero pertenece a la barbería vigente, sin que el
// núcleo de schedule importe el núcleo ni el repositorio de staff (mismo
// criterio que catalog.BarberPort frente a HU-023). cmd/api (la raíz de
// composición) conecta este puerto con la implementación real
// (staff.NewBarberLookup(staffService)); schedule nunca importa el paquete
// staff.
type BarberPort interface {
	// Exists informa si existe un barbero con ese id en esa barbería.
	// found=false cubre tanto "no existe" como "es de otra barbería"
	// (RN-TEN-01).
	Exists(ctx context.Context, barbershopID, barberID string) (bool, error)
}

// CreateInput son los tres campos de un tramo ya validados por Service
// ANTES de llegar al repositorio (día, hora, duración): el repositorio
// nunca revalida su forma, solo su solape contra el estado persistido.
type CreateInput struct {
	ISOWeekday      int
	StartsTime      string
	DurationMinutes int
}

// UpdateInput es el reemplazo completo del intervalo de un tramo existente,
// con la misma forma que CreateInput.
type UpdateInput = CreateInput

// ListResult es una página de tramos ya ordenada de forma estable
// (iso_weekday, starts_time, id). NextCursor es "" cuando esta página es la
// última (CA-040-01).
type ListResult struct {
	Items      []WorkingHour
	NextCursor string
}

// CreateResult es el desenlace completo de un intento de alta idempotente
// (RN-IDE-01, DEC-043), con un desenlace adicional propio de HU-040:
//   - Decision.Outcome == OutcomeProceed && !Conflict: WorkingHour y
//     Response reflejan el tramo recién creado.
//   - Decision.Outcome == OutcomeProceed && Conflict: el intervalo se
//     solapaba con otro tramo existente del mismo barbero y día (o repetía
//     su hora de inicio); no se insertó nada y la reclamación de
//     idempotencia se liberó (mismo criterio que
//     catalog.CreateResult.NameTaken). Service.Create traduce este caso a
//     errOverlapConflict, nunca a un CreateResult "exitoso".
//   - Decision.Outcome == OutcomeReplay: Decision.Response ya trae la
//     respuesta original; Response se deja igual a esa misma copia.
//   - Cualquier otro Outcome: Decision.AsError() ya traduce ese caso.
type CreateResult struct {
	Decision    idempotency.Decision
	WorkingHour WorkingHour
	Response    idempotency.StoredResponse
	Conflict    bool
}

// UpdateResult es el desenlace de un intento de edición. Found en false
// cubre un tramo inexistente, de otro barbero o de otra barbería
// (CA-040-05): el UPDATE ... RETURNING no encontró fila que actualizar.
// Conflict en true cubre CA-040-04: el intervalo nuevo se solapa con otro
// tramo existente del mismo barbero y día (excluyendo el propio tramo
// editado); no se persistió ningún cambio.
type UpdateResult struct {
	WorkingHour WorkingHour
	Found       bool
	Conflict    bool
}

// Repository es el puerto de persistencia del módulo schedule. El núcleo no
// importa internal/platform/database ni pgx (CA-002-06): postgres/ traduce
// entre este contrato y database.DB, mismo patrón que staff.Repository.
//
// Repository.Create es la única operación que además coordina
// internal/platform/idempotency: Begin, la verificación de solape, el
// INSERT y Complete DEBEN ejecutarse dentro de UNA sola transacción
// tenant-aware (apps/api/README.md, "Patrón obligatorio: idempotencia
// reutilizable"), así que esa orquestación vive en el adaptador postgres,
// no en Service.
type Repository interface {
	// List lee una página de tramos de barberID dentro de barbershopID,
	// ordenada por (iso_weekday, starts_time, id). cursor es nil para la
	// primera página; limit ya llegó clamped al rango [MinListLimit,
	// MaxListLimit] por Service. No valida que barberID exista: eso ya lo
	// hizo Service mediante BarberPort antes de llamar aquí.
	List(ctx context.Context, barbershopID, barberID string, cursor *Cursor, limit int) (ListResult, error)

	// Get lee un tramo por id dentro del barbero y tenant vigentes.
	// found=false cubre tanto "no existe" como "es de otro barbero o de
	// otra barbería" (CA-040-05): la consulta filtra por barbershopID y
	// barberID además de RLS, sin distinguir la causa en el resultado.
	Get(ctx context.Context, barbershopID, barberID, workingHourID string) (WorkingHour, bool, error)

	// Create ejecuta, dentro de UNA sola InTenantTx: Begin (idempotencia),
	// si procede, bloquea el barbero (SELECT ... FOR UPDATE, para resistir
	// la carrera de dos altas concurrentes sobre un mismo barbero y día
	// cuando el conjunto de tramos existentes parte vacío -"phantom read"-,
	// mismo criterio que DEC-068 bloqueando la fila de service), verifica
	// solape contra los tramos existentes del mismo (barbershopID,
	// barberID, input.ISOWeekday), INSERT si no hay solape, y Complete.
	// input ya llegó validado por Service; key y fingerprint ya fueron
	// interpretados por la capa HTTP.
	Create(
		ctx context.Context,
		barbershopID, barberID string,
		input CreateInput,
		key idempotency.Key,
		fingerprint idempotency.Fingerprint,
	) (CreateResult, error)

	// Update ejecuta, dentro de UNA sola transacción tenant-aware: bloquea
	// el barbero (mismo criterio que Create), verifica solape contra los
	// demás tramos existentes del mismo (barbershopID, barberID,
	// input.ISOWeekday) EXCLUYENDO workingHourID, y `UPDATE ... RETURNING`
	// filtrando por id, barbershopID y barberID: nunca crea ni duplica una
	// fila, y nunca edita la de otro barbero o de otra barbería
	// (CA-040-05).
	Update(ctx context.Context, barbershopID, barberID, workingHourID string, input UpdateInput) (UpdateResult, error)

	// Delete retira físicamente el tramo (working_hour no tiene
	// eliminación lógica), filtrando por id, barbershopID y barberID.
	// found=false cubre tanto "no existía" como "ya se había retirado"
	// (CA-040-05): reintentar la misma operación tras un 404 es seguro.
	Delete(ctx context.Context, barbershopID, barberID, workingHourID string) (found bool, err error)

	// --- HU-041: interruptor de festivos y excepciones de jornada --------

	// GetHolidayCalendarEnabled lee el interruptor de calendario colombiano
	// de festivos del barbero. found=false cubre tanto "no existe" como
	// "es de otra barbería" (RN-TEN-01).
	GetHolidayCalendarEnabled(ctx context.Context, barbershopID, barberID string) (enabled bool, found bool, err error)

	// SetHolidayCalendarEnabled activa o desactiva el interruptor con un
	// único `UPDATE ... RETURNING`, filtrando por id y barbershopID.
	// Result.Found en false cubre "no existe"/"de otra barbería".
	SetHolidayCalendarEnabled(ctx context.Context, barbershopID, barberID string, enabled bool) (HolidayCalendarResult, error)

	// ListExceptions lee una página de excepciones de barberID dentro de
	// barbershopID, ordenada por (effective_date, id). cursor es nil para
	// la primera página; limit ya llegó clamped por Service.
	ListExceptions(ctx context.Context, barbershopID, barberID string, cursor *ExceptionCursor, limit int) (ExceptionListResult, error)

	// GetException lee una excepción por id dentro del barbero y tenant
	// vigentes, con sus tramos. found=false cubre "no existe"/"de otro
	// barbero"/"de otra barbería" (CA-041-06).
	GetException(ctx context.Context, barbershopID, barberID, exceptionID string) (ScheduleException, bool, error)

	// GetExceptionByDate lee la excepción (si existe) de barberID para
	// effectiveDate, con sus tramos. Lo usa exclusivamente
	// ResolveEffectiveDay (el puerto interno de CA-041-07): a diferencia
	// de GetException, no la usa ningún handler HTTP.
	GetExceptionByDate(ctx context.Context, barbershopID, barberID, effectiveDate string) (ScheduleException, bool, error)

	// ListWorkingHoursForWeekday lee, sin paginar, los tramos de
	// working_hour de barberID para un único isoWeekday, ordenados por
	// starts_time. Lo usa exclusivamente ResolveEffectiveDay: la lista
	// paginada de List (CA-040-01) sirve a la pantalla completa de los
	// siete días, no a la resolución de un único día.
	ListWorkingHoursForWeekday(ctx context.Context, barbershopID, barberID string, isoWeekday int) ([]WorkingHour, error)

	// CreateException ejecuta, dentro de UNA sola InTenantTx: Begin
	// (idempotencia), INSERT de la cabecera y de cada tramo, y Complete.
	// input ya llegó validado por Service (CA-041-04); key y fingerprint
	// ya fueron interpretados por la capa HTTP. Una fecha duplicada
	// (working_hour_override_shop_barber_date_uk) o un solape entre
	// tramos (working_hour_override_segment_no_overlap_excl) se traducen a
	// CreateExceptionResult.Conflict, con ROLLBACK completo (incluida la
	// reclamación de idempotencia, CA-004-06).
	CreateException(
		ctx context.Context,
		barbershopID, barberID string,
		input CreateExceptionInput,
		key idempotency.Key,
		fingerprint idempotency.Fingerprint,
	) (CreateExceptionResult, error)

	// UpdateException reemplaza la excepción completa (fecha, forma y
	// tramos): borra los tramos existentes y crea los nuevos dentro de la
	// misma transacción que el `UPDATE` de la cabecera, filtrando por id,
	// barbershopID y barberID. Found en false cubre "no existe"/"de otro
	// barbero"/"de otra barbería" (CA-041-06); Conflict cubre fecha
	// duplicada o solape de tramos.
	UpdateException(ctx context.Context, barbershopID, barberID, exceptionID string, input UpdateExceptionInput) (UpdateExceptionResult, error)

	// DeleteException retira físicamente la cabecera y sus tramos (sin
	// eliminación lógica), filtrando por id, barbershopID y barberID.
	// found=false cubre "no existía"/"ya se había retirado" (CA-041-06):
	// reintentar tras un 404 es seguro.
	DeleteException(ctx context.Context, barbershopID, barberID, exceptionID string) (found bool, err error)
}
