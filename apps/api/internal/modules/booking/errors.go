package booking

import "system-barbershop/internal/platform/apperr"

func errBarberIDRequired() error {
	return apperr.Validation("el barbero es obligatorio")
}

func errServiceIDRequired() error {
	return apperr.Validation("el servicio es obligatorio")
}

func errAttendeeNameRequired() error {
	return apperr.Validation("el nombre de la persona atendida es obligatorio")
}

func errAttendeeNameTooLong() error {
	return apperr.Validation("el nombre de la persona atendida excede el largo máximo")
}

func errOriginInvalid() error {
	return apperr.Validation("el origen de la cita debe ser public o manual")
}

func errIntervalInvalid() error {
	return apperr.Validation("la hora de fin debe ser posterior a la hora de inicio")
}

func errIntervalDurationMismatch() error {
	return apperr.Validation("el intervalo no coincide con la duración del servicio")
}

func errServiceNameSnapshotRequired() error {
	return apperr.Validation("el nombre del servicio es obligatorio")
}

func errServiceNameSnapshotTooLong() error {
	return apperr.Validation("el nombre del servicio excede el largo máximo")
}

func errDurationSnapshotOutOfRange() error {
	return apperr.Validation("la duración del servicio debe ser un entero entre 1 y 1440 minutos")
}

func errPriceSnapshotNegative() error {
	return apperr.Validation("el precio no puede ser negativo")
}

func errCurrencySnapshotInvalid() error {
	return apperr.Validation("la moneda debe ser un código de tres letras mayúsculas")
}

func errCustomerNoteTooLong() error {
	return apperr.Validation("la nota excede el largo máximo")
}

func errCustomerFullNameRequired() error {
	return apperr.Validation("el nombre del cliente es obligatorio")
}

func errCustomerFullNameTooLong() error {
	return apperr.Validation("el nombre del cliente excede el largo máximo")
}

func errCustomerPhoneInvalid() error {
	return apperr.Validation("el teléfono del cliente tiene un formato inválido")
}

func errCustomerEmailInvalid() error {
	return apperr.Validation("el correo del cliente tiene un formato inválido")
}

// errCustomerInputInvalid cubre CustomerInput con ambos casos presentes o
// ambos ausentes: la primitiva exige exactamente uno (cliente existente
// vinculado por identificador, o datos de un cliente nuevo).
func errCustomerInputInvalid() error {
	return apperr.Validation("debe indicarse exactamente un cliente: existente o nuevo")
}

func errActorInvalid() error {
	return apperr.Validation("el actor del historial tiene una forma inválida")
}

// Los errores de conflicto de agenda y de cliente inexistente/ajeno viven
// en booking/postgres (errScheduleConflict, errCustomerNotFound): solo se
// detectan al traducir un exclusion_violation/foreign_key_violation real de
// PostgreSQL, no como validación de forma en este archivo.
