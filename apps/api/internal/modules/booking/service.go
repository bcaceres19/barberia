package booking

import "context"

// BookingService orquesta el núcleo de citas. HU-060 solo expone
// CreateInternal: la política de creación manual completa (validar
// jornada, bloqueos, asignación servicio-barbero y decidir la
// reconciliación de cliente de DP-CIT-01) es de HU-061 en adelante; este
// servicio confía en que el llamador ya tomó esas decisiones y solo aplica
// las validaciones de forma e integridad que le corresponden al núcleo.
type BookingService struct {
	repo Repository
}

// NewBookingService construye el servicio a partir de su puerto de
// persistencia.
func NewBookingService(repo Repository) *BookingService {
	return &BookingService{repo: repo}
}

// CreateInternal valida input y delega en el repositorio dentro de una
// única transacción tenant-aware (RN-CON-01, RN-HIS-01). barbershopID ya
// llegó resuelto por el llamador (autenticación o contexto de prueba); este
// servicio no lo valida de nuevo.
func (s *BookingService) CreateInternal(ctx context.Context, barbershopID string, input CreateInternalInput) (CreateInternalResult, error) {
	if err := input.validate(); err != nil {
		return CreateInternalResult{}, err
	}
	return s.repo.CreateInternal(ctx, barbershopID, input)
}
