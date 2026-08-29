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

// errBarberNotFound cubre HU-062 (CA-062-02): barberID con forma inválida,
// inexistente o de otra barbería, sin distinguir la causa (RN-TEN-01).
func errBarberNotFound() error {
	return apperr.NotFound("no existe un barbero con ese identificador")
}

// errDateInvalid cubre HU-062: el parámetro date no es una fecha civil
// AAAA-MM-DD válida.
func errDateInvalid() error {
	return apperr.Invalid("date debe ser una fecha civil con el formato AAAA-MM-DD")
}

// Los errores de conflicto de agenda y de cliente inexistente/ajeno viven
// en booking/postgres (errScheduleConflict, errCustomerNotFound): solo se
// detectan al traducir un exclusion_violation/foreign_key_violation real de
// PostgreSQL, no como validación de forma en este archivo.
