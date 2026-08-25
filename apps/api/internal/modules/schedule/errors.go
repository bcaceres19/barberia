package schedule

import "system-barbershop/internal/platform/apperr"

// errBarberNotFound cubre tanto un barbero inexistente como uno de otra
// barbería (RN-TEN-01): el mismo Kind para ambos casos es lo que impide que
// la capa HTTP los distinga, igual que el resto del backend.
func errBarberNotFound() error {
	return apperr.NotFound("no existe un barbero con ese identificador")
}

// errWorkingHourNotFound cubre un tramo inexistente, de otro barbero o de
// otra barbería (CA-040-05, RN-TEN-01): el mismo Kind para las tres causas
// es lo que impide que la capa HTTP las distinga.
func errWorkingHourNotFound() error {
	return apperr.NotFound("no existe un tramo de horario con ese identificador")
}

// Mensajes de campo de CA-040-04: describen un único campo, seguros para el
// cliente (nunca SQL ni el valor crudo enviado). Compartidos por Create y
// Update: ambos validan los mismos tres campos con la misma regla.
func errISOWeekdayInvalid() error {
	return apperr.Validation("el día de la semana debe estar entre 1 y 7")
}

func errStartsTimeInvalid() error {
	return apperr.Validation("la hora de inicio debe tener formato HH:MM de 24 horas")
}

func errDurationInvalid() error {
	return apperr.Validation("la duración debe ser mayor que cero y no exceder 1440 minutos")
}

// errOverlapConflict cubre CA-040-04: el intervalo solicitado se solapa con
// otro tramo existente del mismo barbero y día, o repite exactamente su
// hora de inicio. Conflicto de negocio ajeno a idempotencia
// (apperr.KindConflict, mismo criterio que catalog.errLastActiveAssignment):
// no depende de la forma del cuerpo, depende del estado ya persistido de
// otras filas.
func errOverlapConflict() error {
	return apperr.Conflict("el tramo se solapa con otro existente del mismo barbero y día")
}
