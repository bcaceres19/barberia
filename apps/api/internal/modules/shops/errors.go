package shops

import "system-barbershop/internal/platform/apperr"

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
