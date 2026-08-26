package schedule

import "system-barbershop/internal/platform/apperr"

// errExceptionNotFound cubre una excepción inexistente, de otro barbero o
// de otra barbería (CA-041-06, RN-TEN-01): el mismo Kind para las tres
// causas es lo que impide que la capa HTTP las distinga.
func errExceptionNotFound() error {
	return apperr.NotFound("no existe una excepción de horario con ese identificador")
}

// Mensajes de campo de CA-041-04.
func errEffectiveDateInvalid() error {
	return apperr.Validation("la fecha debe tener formato YYYY-MM-DD y ser una fecha real")
}

func errReasonTooLong() error {
	return apperr.Validation("el motivo excede el largo máximo")
}

func errClosedExceptionHasSegments() error {
	return apperr.Validation("una excepción cerrada no admite tramos")
}

func errOpenExceptionNeedsSegments() error {
	return apperr.Validation("una excepción abierta exige al menos un tramo")
}

func errSegmentsOverlap() error {
	return apperr.Validation("dos tramos de la excepción se solapan entre sí")
}

// errExceptionDateConflict cubre CA-041-05: ya existe otra excepción del
// mismo barbero para esa fecha, o (defensivamente) un solape de tramos que
// la validación previa no capturó. Conflicto de negocio ajeno a
// idempotencia (apperr.KindConflict, mismo criterio que
// errOverlapConflict de HU-040).
func errExceptionDateConflict() error {
	return apperr.Conflict("ya existe una excepción de este barbero para esa fecha")
}

func errExceptionSegmentsConflict() error {
	return apperr.Conflict("dos tramos de la excepción se solapan entre sí")
}
