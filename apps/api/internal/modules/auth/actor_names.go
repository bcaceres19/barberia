package auth

import "context"

// StaffActorNameLookup adapta un Repository ya existente a un puerto mínimo
// y estable (booking.StaffActorNamePort, HU-064) para que otro módulo
// resuelva, en un solo lote, el nombre visible de cada staff_user_id que
// aparece como actor en un historial: mismo criterio de adaptador pequeño
// que staff.BarberLookup frente a catalog.BarberPort. auth NUNCA importa
// booking: la composición ocurre en cmd/api.
type StaffActorNameLookup struct {
	repo Repository
}

// NewStaffActorNameLookup construye el adaptador sobre un Repository ya
// existente (el mismo que auth.NewLoginService/NewSessionService reciben).
func NewStaffActorNameLookup(repo Repository) StaffActorNameLookup {
	return StaffActorNameLookup{repo: repo}
}

// Names resuelve el nombre visible de cada id de staffUserIDs, dentro de esa
// barbería, en una sola consulta por lote (RN-HIS-01, HU-064, "sin generar
// N+1"). Un id sin coincidencia (no debería ocurrir: la FK de
// appointment_history es ON DELETE RESTRICT) simplemente está ausente del
// mapa devuelto; el llamador decide entonces la etiqueta segura de reserva.
func (l StaffActorNameLookup) Names(ctx context.Context, barbershopID string, staffUserIDs []string) (map[string]string, error) {
	return l.repo.StaffUserNames(ctx, barbershopID, staffUserIDs)
}
