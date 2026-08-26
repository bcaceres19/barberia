package schedule

import (
	"context"
	"time"

	"system-barbershop/internal/platform/idempotency"
)

// BlockListResult es una página de bloqueos puntuales ya ordenada de forma
// estable (starts_at, id). NextCursor es "" cuando esta página es la última.
type BlockListResult struct {
	Items      []TimeBlock
	NextCursor string
}

// CreateBlockInput son los campos de un bloqueo puntual ya validados por
// Service ANTES de llegar al repositorio.
type CreateBlockInput struct {
	BlockType string
	StartsAt  time.Time
	EndsAt    time.Time
	Reason    *string
}

// CreateBlockResult es el desenlace de un alta idempotente (RN-IDE-01,
// DEC-043). A diferencia de working_hour, un bloqueo puntual NUNCA tiene
// Conflict: RN-BLQ-03/DEC-008 exigen que crear un bloqueo jamás falle por
// chocar con otro estado existente (ni con citas, que este módulo ni
// siquiera consulta, ni con otro bloqueo).
type CreateBlockResult struct {
	Decision idempotency.Decision
	Block    TimeBlock
	Response idempotency.StoredResponse
}

// SeriesListResult es una página de series ya ordenada de forma estable
// (effective_from, id), cada una con sus Dates/Exceptions embebidos.
type SeriesListResult struct {
	Items      []TimeBlockSeries
	NextCursor string
}

// CreateSeriesInput son los campos de una serie ya validados por Service.
type CreateSeriesInput struct {
	BlockType       string
	RecurrenceKind  string
	ISOWeekday      *int
	StartsTime      string
	DurationMinutes int
	EffectiveFrom   string
	EffectiveUntil  *string
	Reason          *string
	// ExplicitDates son las fechas iniciales de una serie date_list
	// (CA-042: alta con la lista completa en una sola operación, RN-BLQ-01
	// "bloqueo de varios días debe poder crearse en una sola operación").
	// Vacío para weekly.
	ExplicitDates []string
}

// CreateSeriesResult es el desenlace de un alta idempotente.
type CreateSeriesResult struct {
	Decision idempotency.Decision
	Series   TimeBlockSeries
	Response idempotency.StoredResponse
}

// UpdateSeriesInput es el reemplazo completo de los campos editables de una
// serie (scope whole) o la definición de la serie nueva que nace a partir de
// effectiveDate (scope this_and_following). RecurrenceKind e ISOWeekday
// nunca cambian mediante este tipo: whole los conserva tal cual están
// persistidos; this_and_following siempre hereda RecurrenceKind=weekly de la
// serie original (errSplitOnlyWeekly ya lo garantiza antes de construir este
// valor).
type UpdateSeriesInput struct {
	BlockType       string
	StartsTime      string
	DurationMinutes int
	EffectiveFrom   string
	EffectiveUntil  *string
	Reason          *string
}

// UpdateSeriesResult es el desenlace de un PATCH de serie. Found en false
// cubre "no existe"/"de otro barbero"/"de otra barbería"/"ya retirada"
// (deleted_at IS NOT NULL se trata como no encontrada: RN-BLQ-04 exige que
// una serie retirada no vuelva a aparecer).
type UpdateSeriesResult struct {
	Series TimeBlockSeries
	Found  bool
}

// Repository es el puerto de persistencia de HU-042, agregado al mismo
// Repository de working_hour/schedule-exceptions (mismo módulo, mismo
// tenant/barbero, mismo criterio que HU-041 frente a HU-040): no se crea una
// segunda interfaz ni un segundo Service.
type BlockRepository interface {
	// ListBlocks lee una página de bloqueos puntuales de barberID dentro de
	// barbershopID, ordenada por (starts_at, id). includeDeleted en true
	// también trae los retirados lógicamente (auditoría, RN-BLQ-04); en
	// false (caso normal de la pantalla) los excluye.
	ListBlocks(ctx context.Context, barbershopID, barberID string, cursor *BlockCursor, limit int, includeDeleted bool) (BlockListResult, error)

	// GetBlock lee un bloqueo por id dentro del barbero y tenant vigentes,
	// incluidos los retirados lógicamente (el registro se conserva,
	// RN-BLQ-04): found=false cubre solo "no existe"/"de otro
	// barbero"/"de otra barbería".
	GetBlock(ctx context.Context, barbershopID, barberID, blockID string) (TimeBlock, bool, error)

	// CreateBlock ejecuta, dentro de UNA sola InTenantTx: Begin
	// (idempotencia), INSERT y Complete. Nunca verifica solape contra otro
	// bloqueo ni contra citas (RN-BLQ-03/DEC-008): un bloqueo puntual
	// siempre se crea.
	CreateBlock(
		ctx context.Context,
		barbershopID, barberID string,
		input CreateBlockInput,
		key idempotency.Key,
		fingerprint idempotency.Fingerprint,
	) (CreateBlockResult, error)

	// DeleteBlock marca el bloqueo como retirado (deleted_at=now(),
	// deleted_by=actorID) con un único `UPDATE ... RETURNING`, filtrando por
	// id, barbershopID, barberID y deleted_at IS NULL: nunca hace DELETE
	// físico (RN-BLQ-04). found=false cubre "no existe"/"de otro
	// barbero"/"de otra barbería"/"ya estaba retirado": reintentar tras un
	// 404 es seguro y no reescribe deleted_at/deleted_by.
	DeleteBlock(ctx context.Context, barbershopID, barberID, blockID, actorID string) (found bool, err error)

	// ListSeries lee una página de series de barberID dentro de
	// barbershopID, ordenada por (effective_from, id), cada una con sus
	// Dates/Exceptions embebidos. includeDeleted en true también trae las
	// retiradas lógicamente.
	ListSeries(ctx context.Context, barbershopID, barberID string, cursor *SeriesCursor, limit int, includeDeleted bool) (SeriesListResult, error)

	// GetSeries lee una serie por id dentro del barbero y tenant vigentes,
	// con sus Dates/Exceptions, incluidas las retiradas lógicamente.
	// found=false cubre solo "no existe"/"de otro barbero"/"de otra
	// barbería".
	GetSeries(ctx context.Context, barbershopID, barberID, seriesID string) (TimeBlockSeries, bool, error)

	// CreateSeries ejecuta, dentro de UNA sola InTenantTx: Begin
	// (idempotencia), INSERT de la cabecera, INSERT de cada
	// input.ExplicitDates (date_list) y Complete.
	CreateSeries(
		ctx context.Context,
		barbershopID, barberID string,
		input CreateSeriesInput,
		key idempotency.Key,
		fingerprint idempotency.Fingerprint,
	) (CreateSeriesResult, error)

	// UpdateSeriesWhole reemplaza los campos editables de la cabecera
	// (scope whole) con un único `UPDATE ... RETURNING`, filtrando por id,
	// barbershopID, barberID y deleted_at IS NULL.
	UpdateSeriesWhole(ctx context.Context, barbershopID, barberID, seriesID string, input UpdateSeriesInput) (UpdateSeriesResult, error)

	// SplitSeriesFrom implementa "esta y las siguientes" (scope
	// this_and_following, solo weekly): dentro de UNA sola transacción,
	// bloquea la serie original (SELECT ... FOR UPDATE), la trunca
	// (effective_until = effectiveDate - 1 día) y crea una serie nueva con
	// input, effective_from = effectiveDate y effective_until igual al
	// effective_until ORIGINAL (antes de truncar). Devuelve la serie NUEVA
	// (la definición "futura" que el barbero acaba de editar).
	SplitSeriesFrom(ctx context.Context, barbershopID, barberID, seriesID, effectiveDate string, input UpdateSeriesInput) (UpdateSeriesResult, error)

	// DeleteSeries marca la serie completa como retirada
	// (deleted_at=now()) con un único `UPDATE ... RETURNING`, filtrando por
	// deleted_at IS NULL. No toca time_block_series_date ni
	// time_block_series_exception (RESTRICT/DEC-070: nunca cascada). found
	// en false cubre "no existe"/"ya estaba retirada".
	DeleteSeries(ctx context.Context, barbershopID, barberID, seriesID string) (found bool, err error)

	// AddSeriesDate agrega una fecha explícita a una serie date_list.
	// found=false: la serie no existe/es de otro barbero/de otra barbería.
	// conflict=true: la fecha ya estaba registrada
	// (time_block_series_date_pk) o cae fuera del rango vigente / la serie
	// no es date_list (el disparador time_block_series_date_check_parent lo
	// rechaza; ambas causas se traducen aquí a un único booleano porque
	// Service ya validó recurrenceKind==date_list antes de llamar, así que
	// en la práctica solo puede tratarse del rango).
	AddSeriesDate(ctx context.Context, barbershopID, barberID, seriesID, blockDate string) (found bool, conflict bool, err error)

	// RemoveSeriesDate retira físicamente una fecha explícita (la tabla no
	// tiene eliminación lógica propia: es una fila hija sin ciclo de vida
	// propio, análoga a working_hour_override_segment). found=false cubre
	// "no existía"/"ya se había retirado".
	RemoveSeriesDate(ctx context.Context, barbershopID, barberID, seriesID, blockDate string) (found bool, err error)

	// AddSeriesException agrega una excepción ("esta instancia no") a
	// cualquier serie (weekly o date_list). found=false: la serie no
	// existe. conflict=true: ya existe una excepción para esa fecha en esa
	// serie.
	AddSeriesException(ctx context.Context, barbershopID, barberID, seriesID, excludedDate string, reason *string) (found bool, conflict bool, err error)

	// RemoveSeriesException retira físicamente una excepción ("restaura" la
	// instancia). found=false cubre "no existía"/"ya se había retirado".
	RemoveSeriesException(ctx context.Context, barbershopID, barberID, seriesID, excludedDate string) (found bool, err error)

	// ListEffectiveManualBlocks lee, sin paginar, los bloqueos puntuales NO
	// retirados de barberID cuyo intervalo [starts_at, ends_at) se solapa
	// con [from, until), ordenados por starts_at. Alimenta exclusivamente
	// Service.EffectiveBlocks (CA-042: proyección para B3/B4, "Qué NO hace"
	// documenta que no une esta lista con seriesOccurrences: cada
	// consumidor futuro decide cómo combinarlas).
	ListEffectiveManualBlocks(ctx context.Context, barbershopID, barberID string, from, until time.Time) ([]TimeBlock, error)

	// ListActiveSeriesForProjection lee, sin paginar, las series NO
	// retiradas de barberID cuyo rango efectivo se solapa con [from, until)
	// (fechas civiles), con sus Dates/Exceptions embebidos: Service las
	// expande en memoria (mismo criterio que
	// ResolveEffectiveDay/ListWorkingHoursForWeekday, que también resuelven
	// la recurrencia en Go, no en SQL).
	ListActiveSeriesForProjection(ctx context.Context, barbershopID, barberID, from, until string) ([]TimeBlockSeries, error)
}
