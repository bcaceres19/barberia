package catalog

import (
	"context"
	"fmt"

	"system-barbershop/internal/platform/apperr"
)

// AssignmentService implementa los tres casos de uso de HU-023: listar los
// servicios asignados a un barbero, asignar y desasignar. Vive en el
// paquete catalog porque catalog es el módulo dueño de la intención "qué
// servicios se prestan" (trabajo requerido §3.1); colabora con staff
// exclusivamente a través de BarberPort, un puerto pequeño y explícito que
// cmd/api conecta con la implementación real (staff.NewBarberLookup),
// nunca importando el núcleo ni el repositorio de staff.
type AssignmentService struct {
	repo    AssignmentRepository
	barbers BarberPort
}

// NewAssignmentService construye el servicio de casos de uso.
func NewAssignmentService(repo AssignmentRepository, barbers BarberPort) *AssignmentService {
	return &AssignmentService{repo: repo, barbers: barbers}
}

// List lee una página de servicios asignados a barberID dentro de
// barbershopID (CA-023-01). Un barberID inexistente o de otra barbería
// produce apperr.NotFound (mismo criterio uniforme que staff.Service.Get),
// verificado mediante BarberPort antes de tocar el repositorio de
// asignaciones.
func (s *AssignmentService) List(ctx context.Context, barbershopID, barberID, cursorToken string, limit int) (AssignmentListResult, error) {
	if err := ctx.Err(); err != nil {
		return AssignmentListResult{}, apperr.Internal(fmt.Errorf("catalog: contexto cancelado antes de listar asignaciones: %w", err))
	}
	if !LooksLikeBarberID(barberID) {
		return AssignmentListResult{}, errAssignmentBarberNotFound()
	}

	exists, err := s.barbers.Exists(ctx, barbershopID, barberID)
	if err != nil {
		return AssignmentListResult{}, apperr.Internal(fmt.Errorf("catalog: verificar barbero para listar asignaciones: %w", err))
	}
	if !exists {
		return AssignmentListResult{}, errAssignmentBarberNotFound()
	}

	switch {
	case limit <= 0:
		limit = DefaultListLimit
	case limit < MinListLimit:
		limit = MinListLimit
	case limit > MaxListLimit:
		limit = MaxListLimit
	}

	var cursor *Cursor
	if cursorToken != "" {
		decoded, err := DecodeCursor(cursorToken)
		if err != nil {
			return AssignmentListResult{}, err
		}
		cursor = &decoded
	}

	result, err := s.repo.List(ctx, barbershopID, barberID, cursor, limit)
	if err != nil {
		return AssignmentListResult{}, apperr.Internal(fmt.Errorf("catalog: listar asignaciones: %w", err))
	}
	return result, nil
}

// Assign crea la asociación entre barberID y serviceID dentro de
// barbershopID (CA-023-02, CA-023-03): repetir exactamente la misma
// operación no crea una segunda fila (AssignOutcomeAlreadyExists). Un
// barberID o serviceID inexistente o de otra barbería produce el mismo
// apperr.NotFound uniforme (CA-023-04): el primero se descubre mediante
// BarberPort ANTES de tocar el repositorio (trabajo requerido §3.1); el
// segundo lo descubre Repository.Assign, dueño de la tabla `service`.
func (s *AssignmentService) Assign(ctx context.Context, barbershopID, barberID, serviceID string) (AssignResult, error) {
	if err := ctx.Err(); err != nil {
		return AssignResult{}, apperr.Internal(fmt.Errorf("catalog: contexto cancelado antes de asignar servicio: %w", err))
	}
	if !LooksLikeBarberID(barberID) {
		return AssignResult{}, errAssignmentBarberNotFound()
	}
	if !LooksLikeServiceID(serviceID) {
		return AssignResult{}, errServiceNotFound()
	}

	exists, err := s.barbers.Exists(ctx, barbershopID, barberID)
	if err != nil {
		return AssignResult{}, apperr.Internal(fmt.Errorf("catalog: verificar barbero para asignar servicio: %w", err))
	}
	if !exists {
		return AssignResult{}, errAssignmentBarberNotFound()
	}

	result, err := s.repo.Assign(ctx, barbershopID, barberID, serviceID)
	if err != nil {
		return AssignResult{}, apperr.Internal(fmt.Errorf("catalog: asignar servicio: %w", err))
	}
	if result.Outcome == AssignOutcomeServiceNotFound {
		return AssignResult{}, errServiceNotFound()
	}
	return result, nil
}

// IsAssigned informa si barberID tiene asignado serviceID dentro de
// barbershopID (HU-061, DEC-072), sin verificar barbero ni servicio por
// separado: ManualBookingCatalog ya combina esto con CatalogService.Get
// para decidir found de forma uniforme.
func (s *AssignmentService) IsAssigned(ctx context.Context, barbershopID, barberID, serviceID string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, apperr.Internal(fmt.Errorf("catalog: contexto cancelado antes de verificar asignación: %w", err))
	}
	exists, err := s.repo.Exists(ctx, barbershopID, barberID, serviceID)
	if err != nil {
		return false, apperr.Internal(fmt.Errorf("catalog: verificar asignación: %w", err))
	}
	return exists, nil
}

// Unassign retira la asociación entre barberID y serviceID dentro de
// barbershopID (CA-023-05, CA-023-06, DEC-068). No existe tal asignación
// (barbero o servicio inexistente/ajeno, o la asociación concreta nunca
// existió) produce apperr.NotFound uniforme; ser la última asignación
// activa de un servicio activo produce apperr.Conflict, verificado dentro
// de la misma transacción que el DELETE (Repository.Unassign), resistente
// a la carrera de dos desasignaciones concurrentes de las dos últimas filas
// de un mismo servicio.
func (s *AssignmentService) Unassign(ctx context.Context, barbershopID, barberID, serviceID string) (UnassignResult, error) {
	if err := ctx.Err(); err != nil {
		return UnassignResult{}, apperr.Internal(fmt.Errorf("catalog: contexto cancelado antes de retirar servicio: %w", err))
	}
	if !LooksLikeBarberID(barberID) || !LooksLikeServiceID(serviceID) {
		// Una forma inválida no puede corresponder a ninguna fila real: se
		// trata igual que "no existe esa asignación" (CA-023-04), sin
		// necesitar BarberPort para este caso (a diferencia de List/Assign,
		// donde el barbero puede tener forma válida pero pertenecer a otra
		// barbería): Repository.Unassign ya distingue esa situación por la
		// ausencia de la fila de asociación.
		return UnassignResult{}, errAssignmentNotFound()
	}

	result, err := s.repo.Unassign(ctx, barbershopID, barberID, serviceID)
	if err != nil {
		return UnassignResult{}, apperr.Internal(fmt.Errorf("catalog: retirar servicio: %w", err))
	}
	switch result.Outcome {
	case UnassignOutcomeNotFound:
		return UnassignResult{}, errAssignmentNotFound()
	case UnassignOutcomeLastActiveConflict:
		return UnassignResult{}, errLastActiveAssignment()
	}
	return result, nil
}
