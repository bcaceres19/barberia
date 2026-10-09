package postgres

import (
	"context"

	"system-barbershop/internal/platform/database"
)

// SyncHook es el puerto con el que booking permite a una integración (hoy, la
// publicación en Google Calendar, DEC-102) reaccionar a una escritura sobre
// `appointment` DENTRO de la misma transacción, sin que booking conozca ni
// importe a esa integración. La implementación se inyecta en la raíz de
// composición (cmd/api).
//
// Contrato: se invoca después de TODA escritura sobre `appointment` (alta,
// reprogramación y cualquier cambio de estado), con el `q` de esa misma
// transacción. Si devuelve un error, la transacción completa se revierte: es
// preferible no confirmar un cambio a confirmarlo sin encolar su publicación.
// Una escritura nueva sobre `appointment` (por ejemplo la cancelación por el
// cliente, HU-099) DEBE llamarlo; sync_hook_test.go lo comprueba.
type SyncHook interface {
	AppointmentChanged(ctx context.Context, q database.Queries, barbershopID, appointmentID string) error
}

// WithSyncHook registra la integración que reacciona a los cambios de citas.
// Se llama durante la composición, antes de atender solicitudes; sin ella el
// repositorio funciona igual (no hay nada que publicar).
func (r *Repository) WithSyncHook(hook SyncHook) *Repository {
	r.hook = hook
	return r
}

func (r *Repository) appointmentChanged(ctx context.Context, q database.Queries, barbershopID, appointmentID string) error {
	if r.hook == nil {
		return nil
	}
	return r.hook.AppointmentChanged(ctx, q, barbershopID, appointmentID)
}
