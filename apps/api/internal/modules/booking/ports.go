package booking

import "context"

// Repository es el puerto de persistencia del módulo booking. El núcleo no
// importa internal/platform/database ni pgx (CA-002-06): postgres/ traduce
// entre este contrato y database.DB, mismo patrón que catalog.Repository.
type Repository interface {
	// CreateInternal ejecuta, dentro de UNA sola transacción tenant-aware,
	// la creación o vinculación del cliente que input.Customer ya decidió,
	// el INSERT de la cita `confirmed` y el INSERT del evento
	// appointment_created (§2.3 del prompt de HU-060): un fallo en
	// cualquier paso revierte los tres. input ya llegó validado por
	// BookingService.
	CreateInternal(ctx context.Context, barbershopID string, input CreateInternalInput) (CreateInternalResult, error)
}
