package postgres

import (
	"context"

	"system-barbershop/internal/platform/database"
)

// SyncHook es el puerto con el que schedule permite a una integración (hoy, la
// publicación en Google Calendar, DEC-102) reaccionar a una escritura sobre
// `time_block` DENTRO de la misma transacción, sin que schedule conozca ni
// importe a esa integración.
//
// Contrato: se invoca después de TODA escritura sobre `time_block` (alta y retiro
// lógico), con el `q` de esa misma transacción; un error revierte la transacción
// completa. Una escritura nueva sobre `time_block` DEBE llamarlo;
// sync_hook_test.go lo comprueba.
type SyncHook interface {
	TimeBlockChanged(ctx context.Context, q database.Queries, barbershopID, blockID string) error
}

// WithSyncHook registra la integración que reacciona a los cambios de bloqueos.
// Se llama durante la composición; sin ella el repositorio funciona igual.
func (r *Repository) WithSyncHook(hook SyncHook) *Repository {
	r.hook = hook
	return r
}

func (r *Repository) timeBlockChanged(ctx context.Context, q database.Queries, barbershopID, blockID string) error {
	if r.hook == nil {
		return nil
	}
	return r.hook.TimeBlockChanged(ctx, q, barbershopID, blockID)
}
