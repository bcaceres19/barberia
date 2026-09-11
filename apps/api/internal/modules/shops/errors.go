package shops

import (
	"fmt"

	"system-barbershop/internal/platform/apperr"
)

// Mensajes de campo de CA-020-03/CA-020-06: cada uno describe un único
// campo, seguros para el cliente (nunca SQL ni el valor crudo enviado,
// CA-003-02). PATCH evalúa los campos en el orden fijo de la firma del
// contrato (name, timezone, contactEmail, contactPhone) y devuelve el
// primer error de campo que encuentra, igual que LoginService hace con la
// forma del request.
func errNameRequired() error {
	return apperr.Validation("el nombre de la barbería es obligatorio")
}

func errNameTooLong() error {
	return apperr.Validation("el nombre de la barbería excede el largo máximo")
}

func errTimezoneRequired() error {
	return apperr.Validation("la zona horaria es obligatoria")
}

func errTimezoneTooLong() error {
	return apperr.Validation("la zona horaria excede el largo máximo")
}

// errTimezoneInvalid es el error de CA-020-03: el texto no corresponde a
// ninguna zona de pg_timezone_names. Se construye después de que el
// repositorio confirma la ausencia dentro de la misma transacción que
// intentó el UPDATE (cero escritura), no antes.
func errTimezoneInvalid() error {
	return apperr.Validation("la zona horaria no es una zona IANA reconocida")
}

func errContactEmailInvalid() error {
	return apperr.Validation("el correo de contacto no tiene un formato válido")
}

func errContactPhoneInvalid() error {
	return apperr.Validation("el teléfono de contacto no tiene un formato E.164 válido")
}

// errBookingPolicyVersionTokenRequired cubre HU-093 (CA-093-02): la
// cabecera If-Match (precondición de versión) es obligatoria para
// actualizar la política de reserva, mismo criterio que
// booking.errVersionTokenRequired.
func errBookingPolicyVersionTokenRequired() error {
	return apperr.Invalid("falta la cabecera If-Match con el token de versión de la política de reserva")
}

// errBookingPolicyVersionConflict cubre HU-093 (CA-093-02): el token de
// versión que el cliente envió (cabecera If-Match) ya no coincide con la
// representación vigente porque otra escritura tocó la barbería primero.
func errBookingPolicyVersionConflict() error {
	return apperr.VersionConflict("la política de reserva cambió desde que se leyó; recarga antes de reintentar")
}

// errBookingPolicyFieldOutOfRange cubre HU-093 (CA-093-02): field está fuera
// del rango permitido por DEC-083. Nunca revela el rango ni el valor
// enviado en el mensaje (mismo criterio de mensajes seguros de errNameTooLong
// y similares): el cliente ya conoce el rango porque el formulario lo
// muestra (CA-093-05).
func errBookingPolicyFieldOutOfRange(field string) error {
	return apperr.Validation(fmt.Sprintf("%s está fuera del rango permitido", field))
}

// errBookingPolicyIncoherentCancellationPolicy cubre HU-093 (CA-093-02):
// exigir motivo para una cancelación tardía que el cliente ni siquiera
// puede hacer es una combinación incoherente (DEC-083).
func errBookingPolicyIncoherentCancellationPolicy() error {
	return apperr.Validation("no se puede exigir motivo de cancelación tardía si el cliente no puede cancelar tarde")
}

// errBookingPolicyAdvanceExceedsWindow cubre HU-093 (CA-093-02,
// barbershop_min_advance_vs_window_ck): la anticipación mínima no puede
// alcanzar ni superar toda la ventana pública de reserva, o ninguna franja
// quedaría reservable.
func errBookingPolicyAdvanceExceedsWindow() error {
	return apperr.Validation("la anticipación mínima no puede alcanzar ni superar la ventana máxima de reserva")
}
