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

// errAppointmentNotFound cubre HU-064 (CA-064-01, CA-064-07): appointmentId
// con forma inválida, inexistente o de otra barbería, sin distinguir la
// causa (RN-TEN-01), mismo criterio que errBarberNotFound.
func errAppointmentNotFound() error {
	return apperr.NotFound("no existe una cita con ese identificador")
}

// errHistoryCursorInvalid cubre HU-064: el parámetro cursor del historial no
// es un valor opaco válido producido por EncodeHistoryCursor.
func errHistoryCursorInvalid() error {
	return apperr.Invalid("el parámetro cursor tiene un formato inválido")
}

// errVersionTokenRequired cubre HU-065: la cabecera If-Match (precondición
// de versión) es obligatoria para reprogramar, mismo criterio de
// obligatoriedad que Idempotency-Key.
func errVersionTokenRequired() error {
	return apperr.Invalid("falta la cabecera If-Match con el token de versión del turno")
}

// errStartsAtNotFuture cubre HU-065 (CA-065-*): T2 exige un nuevo inicio
// estrictamente futuro en la zona de la barbería.
func errStartsAtNotFuture() error {
	return apperr.Validation("el nuevo inicio debe ser un instante futuro")
}

// errAppointmentNotConfirmed cubre HU-065: T2 solo aplica sobre una cita
// `confirmed`; una cita terminal (completada, cancelada, no_show) ya no
// admite reprogramación.
func errAppointmentNotConfirmed() error {
	return apperr.InvalidState("el turno ya no está confirmado; recarga para ver su estado actual")
}

// errVersionConflict cubre HU-065: el token opaco de versión que el cliente
// envió (cabecera If-Match) ya no coincide con la representación vigente
// porque otra escritura tocó el turno primero.
func errVersionConflict() error {
	return apperr.VersionConflict("el turno cambió desde que se leyó; recarga antes de reintentar")
}

// errAppointmentNotStarted cubre HU-067 (CA-067-03): T4 manual y T7 solo
// aplican sobre una cita `confirmed` cuyo starts_at ya pasó según el reloj
// del servidor; antes de ese instante exacto la operación es inválida.
func errAppointmentNotStarted() error {
	return apperr.Validation("el turno todavía no comienza; espera hasta su hora de inicio")
}

// errAppointmentAlreadyClosed cubre HU-067 (CA-067-05): el turno ya tiene
// un resultado terminal distinto del que el comando intenta aplicar
// (completed/no_show contrario, o cancelled_by_barber/cancelled_by_customer).
// Repetir el MISMO resultado nunca produce este error: es un no-op exitoso
// resuelto dentro del repositorio, antes de llegar a esta rama.
func errAppointmentAlreadyClosed() error {
	return apperr.InvalidState("el turno ya tiene un resultado terminal registrado; una futura corrección (T8) permitirá cambiarlo")
}

// Los errores de conflicto de agenda y de cliente inexistente/ajeno viven
// en booking/postgres (errScheduleConflict, errCustomerNotFound): solo se
// detectan al traducir un exclusion_violation/foreign_key_violation real de
// PostgreSQL, no como validación de forma en este archivo.
