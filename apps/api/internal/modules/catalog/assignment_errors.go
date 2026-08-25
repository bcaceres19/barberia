package catalog

import "system-barbershop/internal/platform/apperr"

// errAssignmentBarberNotFound cubre tanto un barbero inexistente como uno de
// otra barbería (CA-023-04, RN-TEN-01): el mismo Kind para ambos casos es lo
// que impide que la capa HTTP los distinga, igual que el resto del backend
// (mismo criterio que errServiceNotFound en errors.go).
func errAssignmentBarberNotFound() error {
	return apperr.NotFound("no existe un barbero con ese identificador")
}

// errAssignmentNotFound cubre servicio inexistente/ajeno, barbero
// inexistente/ajeno (cuando la desasignación lo descubre por ausencia de
// fila, no por BarberPort) y la asociación concreta inexistente: las tres
// causas producen el mismo 404 uniforme (CA-023-04).
func errAssignmentNotFound() error {
	return apperr.NotFound("no existe esa asignación de servicio para ese barbero")
}

// errLastActiveAssignment cubre DEC-068: retirar la última asignación
// activa de un servicio activo se rechaza con un conflicto de negocio ajeno
// a idempotencia (apperr.KindConflict, igual que errNameConflict), no una
// validación de campo: no depende de la forma del cuerpo, depende del
// estado ya persistido de otras filas.
func errLastActiveAssignment() error {
	return apperr.Conflict("el servicio quedaría sin ningún barbero asignado; asigna otro barbero antes de retirar este")
}
