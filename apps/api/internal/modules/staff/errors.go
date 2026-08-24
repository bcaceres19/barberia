package staff

import "system-barbershop/internal/platform/apperr"

// Mensajes de campo de CA-021-03: describen un único campo, seguros para el
// cliente (nunca SQL ni el valor crudo enviado, CA-003-02). Compartidos por
// Create y Rename: ambos validan `fullName` con la misma regla.
func errFullNameRequired() error {
	return apperr.Validation("el nombre del barbero es obligatorio")
}

func errFullNameTooLong() error {
	return apperr.Validation("el nombre del barbero excede el largo máximo")
}

// errBarberNotFound cubre tanto un barbero inexistente como uno de otra
// barbería: el mismo Kind para ambos casos es lo que impide que la capa
// HTTP los distinga (CA-021-05, RN-TEN-01), igual que el resto del
// backend.
func errBarberNotFound() error {
	return apperr.NotFound("no existe un barbero con ese identificador")
}
