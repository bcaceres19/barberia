package catalog

import "system-barbershop/internal/platform/apperr"

// Mensajes de campo de CA-022-04: describen un único campo, seguros para el
// cliente (nunca SQL ni el valor crudo enviado). Compartidos por Create y
// Update: ambos validan los mismos campos con las mismas reglas.
func errNameRequired() error {
	return apperr.Validation("el nombre del servicio es obligatorio")
}

func errNameTooLong() error {
	return apperr.Validation("el nombre del servicio excede el largo máximo")
}

func errDescriptionTooLong() error {
	return apperr.Validation("la descripción del servicio excede el largo máximo")
}

func errDurationOutOfRange() error {
	return apperr.Validation("la duración del servicio debe ser un entero entre 1 y 1440 minutos")
}

func errPriceInvalidFormat() error {
	return apperr.Validation("el precio debe ser un número decimal con hasta dos cifras")
}

func errPriceMustBePositive() error {
	return apperr.Validation("el precio debe ser mayor que cero")
}

// errServiceNotFound cubre tanto un servicio inexistente como uno de otra
// barbería: el mismo Kind para ambos casos es lo que impide que la capa
// HTTP los distinga (CA-022-06, RN-TEN-01), igual que el resto del backend.
func errServiceNotFound() error {
	return apperr.NotFound("no existe un servicio con ese identificador")
}

// errNameConflict cubre DEC-067: dos servicios activos de la misma
// barbería no pueden compartir nombre. apperr.KindConflict, no
// apperr.KindValidation: no es un defecto de forma del campo, es un
// conflicto con otra fila ya persistida.
func errNameConflict() error {
	return apperr.Conflict("ya existe un servicio activo con ese nombre en esta barbería")
}

// errUpdateEmptyBody cubre CA-022-04/CA-022-07: PATCH exige al menos un
// campo de catálogo para editar; un objeto vacío no tiene ningún efecto que
// aplicar y se rechaza antes de tocar el repositorio.
func errUpdateEmptyBody() error {
	return apperr.Validation("debe incluir al menos un campo del catálogo para editar")
}

// errServiceAlreadyInactive cubre HU-024: una clave de idempotencia nueva
// que intenta desactivar un servicio que ya está inactivo (CA-024-06,
// "transición inválida"). apperr.KindConflict, no KindValidation: no es un
// defecto de forma del cuerpo, es un conflicto con el estado real ya
// persistido de ese mismo recurso.
func errServiceAlreadyInactive() error {
	return apperr.Conflict("el servicio ya está inactivo")
}

// errServiceAlreadyActive es el mismo criterio que errServiceAlreadyInactive,
// en sentido inverso (reactivar un servicio que ya está activo).
func errServiceAlreadyActive() error {
	return apperr.Conflict("el servicio ya está activo")
}
