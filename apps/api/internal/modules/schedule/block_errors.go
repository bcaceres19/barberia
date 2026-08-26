package schedule

import "system-barbershop/internal/platform/apperr"

// errTimeBlockNotFound cubre un bloqueo puntual inexistente, de otro
// barbero o de otra barbería (RN-TEN-01): el mismo Kind para las tres
// causas impide que la capa HTTP las distinga.
func errTimeBlockNotFound() error {
	return apperr.NotFound("no existe un bloqueo con ese identificador")
}

// errTimeBlockSeriesNotFound cubre una serie inexistente, de otro barbero o
// de otra barbería.
func errTimeBlockSeriesNotFound() error {
	return apperr.NotFound("no existe una definición de bloqueo recurrente con ese identificador")
}

func errBlockTypeInvalid() error {
	return apperr.Validation("el tipo de bloqueo debe ser uno de: break, lunch, unavailable, day_off, holiday, vacation, emergency")
}

func errRecurrenceKindInvalid() error {
	return apperr.Validation("la recurrencia debe ser weekly o date_list")
}

func errDateInvalid() error {
	return apperr.Validation("la fecha debe tener formato YYYY-MM-DD y ser una fecha calendario válida")
}

func errDateTimeInvalid() error {
	return apperr.Validation("la marca de tiempo debe tener formato RFC 3339 con offset explícito")
}

func errIntervalInvalid() error {
	return apperr.Validation("la hora de fin debe ser posterior a la hora de inicio")
}

func errWeekdayShapeInvalid() error {
	return apperr.Validation("una recurrencia weekly exige isoWeekday entre 1 y 7; una date_list no admite isoWeekday")
}

func errEffectiveRangeInvalid() error {
	return apperr.Validation("effectiveUntil no puede ser anterior a effectiveFrom")
}

func errCursorInvalid() error {
	return apperr.Invalid("el parámetro cursor tiene un formato inválido")
}

// errDateOutOfSeriesRange cubre CA-042: una fecha explícita fuera del rango
// vigente de su serie, o una serie que no es date_list. Mismo Kind que la
// excepción del disparador time_block_series_date_check_parent, traducido a
// un error de negocio legible en vez de propagar el texto crudo de
// PostgreSQL.
func errDateOutOfSeriesRange() error {
	return apperr.Validation("la fecha debe pertenecer al rango vigente de una serie date_list")
}

// errDuplicateSeriesDate cubre CA-042: la fecha ya está registrada en esa
// serie (time_block_series_date_pk).
func errDuplicateSeriesDate() error {
	return apperr.Conflict("esa fecha ya está registrada en la serie")
}

// errSeriesDateNotFound cubre "no existía"/"ya se había retirado"
// (reintentar tras un 404 es seguro, mismo criterio que
// DeleteScheduleException).
func errSeriesDateNotFound() error {
	return apperr.NotFound("no existe esa fecha explícita en la serie")
}

// errDuplicateSeriesException cubre CA-042: ya existe una excepción para
// esa fecha en esa serie (time_block_series_exception_pk).
func errDuplicateSeriesException() error {
	return apperr.Conflict("ya existe una excepción para esa fecha en la serie")
}

func errSeriesExceptionNotFound() error {
	return apperr.NotFound("no existe esa excepción en la serie")
}

// errUpdateScopeInvalid cubre un scope fuera de whole/this_and_following.
func errUpdateScopeInvalid() error {
	return apperr.Validation("scope debe ser whole o this_and_following")
}

// errSplitOnlyWeekly cubre CA-042: this_and_following solo tiene sentido en
// una recurrencia weekly. Una serie date_list ya expone cada fecha como un
// recurso individual (dates/exceptions): dividir la serie no aporta nada
// que esos dos sub-recursos no resuelvan ya.
func errSplitOnlyWeekly() error {
	return apperr.Validation("this_and_following solo aplica a series weekly; edita fechas individuales con el sub-recurso dates")
}

// errSplitDateOutOfRange cubre CA-042: effectiveDate debe caer dentro del
// rango vigente de la serie ORIGINAL y dejar al menos un día para la serie
// que queda truncada (effectiveDate > effectiveFrom original).
func errSplitDateOutOfRange() error {
	return apperr.Validation("effectiveDate debe ser posterior a effectiveFrom y, si existe, no posterior a effectiveUntil")
}

// errRecurrenceKindImmutable cubre CA-042: un PATCH nunca convierte una
// serie weekly en date_list ni viceversa; crear una serie nueva expresa esa
// intención sin ambigüedad.
func errRecurrenceKindImmutable() error {
	return apperr.Validation("recurrenceKind no puede cambiar al editar una serie existente")
}

// errExplicitDatesOnlyDateList cubre CA-042: explicitDates solo tiene
// sentido cuando recurrenceKind es date_list.
func errExplicitDatesOnlyDateList() error {
	return apperr.Validation("explicitDates solo se admite cuando recurrenceKind es date_list")
}

// errQueryRangeInvalid cubre la proyección efectiva (CA-042: to anterior a from).
func errQueryRangeInvalid() error {
	return apperr.Invalid("el parámetro to no puede ser anterior a from")
}
