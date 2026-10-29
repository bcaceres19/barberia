package shops

import (
	"context"
	"fmt"

	"system-barbershop/internal/platform/apperr"
)

// Role es un rol de un usuario del área privada (DEC-134), igual que la lista
// cerrada de staff_user_role_role_ck. Solo decide la presentación del panel.
type Role string

const (
	RoleEmpresa       Role = "empresa"
	RoleEmpleado      Role = "empleado"
	RoleIndependiente Role = "independiente"
)

// PanelProfileFor deriva el panel de un usuario de sus roles (DEC-134). La
// prioridad es empresa, independiente y empleado: quien administra la barbería
// conserva el panel completo, y quien es empleado e independiente a la vez ve
// el individual completo. Sin ningún rol (usuario creado sin asignar), el panel
// completo, que es el comportamiento anterior a los roles.
func PanelProfileFor(roles []Role) PanelProfile {
	var empresa, independiente, empleado bool
	for _, r := range roles {
		switch r {
		case RoleEmpresa:
			empresa = true
		case RoleIndependiente:
			independiente = true
		case RoleEmpleado:
			empleado = true
		}
	}
	switch {
	case empresa:
		return PanelProfileShop
	case independiente:
		return PanelProfileSolo
	case empleado:
		return PanelProfileEmployee
	default:
		return PanelProfileShop
	}
}

// PanelProfileService calcula el panel del usuario autenticado a partir de sus
// roles. No decide permisos: es solo presentación (DEC-134).
type PanelProfileService struct {
	repo PanelProfileRepository
}

// NewPanelProfileService construye el servicio.
func NewPanelProfileService(repo PanelProfileRepository) *PanelProfileService {
	return &PanelProfileService{repo: repo}
}

// For devuelve el panel de staffUserID en barbershopID.
func (s *PanelProfileService) For(ctx context.Context, barbershopID, staffUserID string) (PanelProfile, error) {
	if err := ctx.Err(); err != nil {
		return "", apperr.Internal(fmt.Errorf("shops: contexto cancelado antes de leer los roles: %w", err))
	}
	roles, err := s.repo.RolesOf(ctx, barbershopID, staffUserID)
	if err != nil {
		return "", apperr.Internal(fmt.Errorf("shops: leer roles: %w", err))
	}
	return PanelProfileFor(roles), nil
}
