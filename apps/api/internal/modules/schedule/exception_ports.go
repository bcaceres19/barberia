package schedule

import "system-barbershop/internal/platform/idempotency"

// CreateExceptionSegmentInput es un tramo de una excepción ya validado por
// Service (día/hora/duración), antes de tocar el repositorio.
type CreateExceptionSegmentInput struct {
	StartsTime      string
	DurationMinutes int
}

// CreateExceptionInput son los campos de una excepción ya validados por
// Service (CA-041-04) ANTES de llegar al repositorio: el repositorio nunca
// revalida su forma, solo la fecha duplicada y el solape de tramos contra
// el estado persistido.
type CreateExceptionInput struct {
	EffectiveDate string
	IsClosed      bool
	Reason        *string
	Segments      []CreateExceptionSegmentInput
}

// UpdateExceptionInput es el reemplazo completo de una excepción
// existente, con la misma forma que CreateExceptionInput.
type UpdateExceptionInput = CreateExceptionInput

// HolidayCalendarResult es el desenlace de leer o actualizar el
// interruptor de calendario colombiano de festivos de un barbero.
type HolidayCalendarResult struct {
	Enabled bool
	Found   bool
}

// ExceptionListResult es una página de excepciones ya ordenada de forma
// estable (effective_date, id). NextCursor es "" cuando esta página es la
// última (CA-041-04).
type ExceptionListResult struct {
	Items      []ScheduleException
	NextCursor string
}

// CreateExceptionResult es el desenlace completo de un intento de alta
// idempotente (RN-IDE-01, DEC-043), con un desenlace adicional propio de
// HU-041:
//   - Decision.Outcome == OutcomeProceed && !Conflict: Exception y
//     Response reflejan la excepción recién creada.
//   - Decision.Outcome == OutcomeProceed && Conflict: la fecha ya tenía
//     otra excepción de este barbero, o dos tramos se solapaban; no se
//     insertó nada y la reclamación de idempotencia se liberó (mismo
//     criterio que schedule.CreateResult.Conflict de HU-040).
//   - Decision.Outcome == OutcomeReplay: Decision.Response ya trae la
//     respuesta original.
//   - Cualquier otro Outcome: Decision.AsError() ya lo traduce.
type CreateExceptionResult struct {
	Decision  idempotency.Decision
	Exception ScheduleException
	Response  idempotency.StoredResponse
	Conflict  bool
}

// UpdateExceptionResult es el desenlace de un intento de edición. Found en
// false cubre una excepción inexistente, de otro barbero o de otra
// barbería (CA-041-06). Conflict en true cubre CA-041-05: la fecha nueva
// ya tiene otra excepción de este barbero (excluyendo la propia excepción
// editada), o dos tramos nuevos se solapan.
type UpdateExceptionResult struct {
	Exception ScheduleException
	Found     bool
	Conflict  bool
}
