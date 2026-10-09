package postgres

import (
	"context"
	"fmt"

	"system-barbershop/internal/platform/database"
)

// Enqueuer implementa, con UNA sola sentencia SQL cada una, los puertos de
// enganche que definen booking (SyncHook) y schedule (SyncHook): encola un
// trabajo «reconcilia este recurso» DENTRO de la misma transacción que el
// cambio de negocio (DEC-102). Ni booking ni schedule importan este paquete: la
// raíz de composición inyecta el Enqueuer.
//
// Solo encola si el barbero tiene una conexión que sigue publicando (`connected`
// o `error`); sin ella, o con la integración desactivada, no hace nada (DEC-122,
// punto 5). Además de la conexión del barbero actual, encola cualquier conexión
// que YA tenga un vínculo con el recurso: si una cita cambiara de barbero, la
// conexión anterior retira su evento. Varios cambios seguidos se funden en un
// trabajo pendiente (índice único parcial).
type Enqueuer struct{}

// NewEnqueuer construye el encolador.
func NewEnqueuer() Enqueuer { return Enqueuer{} }

const enqueueAppointmentSQL = `
INSERT INTO google_calendar_sync_job (barbershop_id, connection_id, resource_type, resource_id)
SELECT c.barbershop_id, c.id, 'appointment', a.id
  FROM appointment a
  JOIN google_calendar_connection c
    ON c.barbershop_id = a.barbershop_id AND c.barber_id = a.barber_id
 WHERE a.barbershop_id = $1 AND a.id = $2 AND c.status IN ('connected', 'error')
UNION
SELECT l.barbershop_id, l.connection_id, 'appointment', l.resource_id
  FROM google_calendar_event_link l
  JOIN google_calendar_connection c
    ON c.barbershop_id = l.barbershop_id AND c.id = l.connection_id
 WHERE l.barbershop_id = $1 AND l.resource_type = 'appointment' AND l.resource_id = $2
   AND c.status IN ('connected', 'error')
ON CONFLICT (connection_id, resource_type, resource_id) WHERE status = 'pending' DO NOTHING`

const enqueueTimeBlockSQL = `
INSERT INTO google_calendar_sync_job (barbershop_id, connection_id, resource_type, resource_id)
SELECT c.barbershop_id, c.id, 'time_block', t.id
  FROM time_block t
  JOIN google_calendar_connection c
    ON c.barbershop_id = t.barbershop_id AND c.barber_id = t.barber_id
 WHERE t.barbershop_id = $1 AND t.id = $2 AND c.status IN ('connected', 'error')
   AND t.source = 'manual' AND t.block_type <> 'holiday'
UNION
SELECT l.barbershop_id, l.connection_id, 'time_block', l.resource_id
  FROM google_calendar_event_link l
  JOIN google_calendar_connection c
    ON c.barbershop_id = l.barbershop_id AND c.id = l.connection_id
 WHERE l.barbershop_id = $1 AND l.resource_type = 'time_block' AND l.resource_id = $2
   AND c.status IN ('connected', 'error')
ON CONFLICT (connection_id, resource_type, resource_id) WHERE status = 'pending' DO NOTHING`

// AppointmentChanged implementa booking/postgres.SyncHook.
func (Enqueuer) AppointmentChanged(ctx context.Context, q database.Queries, barbershopID, appointmentID string) error {
	if _, err := q.Exec(ctx, enqueueAppointmentSQL, barbershopID, appointmentID); err != nil {
		return fmt.Errorf("googlecalendar/postgres: enqueue appointment: %w", err)
	}
	return nil
}

// TimeBlockChanged implementa schedule/postgres.SyncHook.
func (Enqueuer) TimeBlockChanged(ctx context.Context, q database.Queries, barbershopID, blockID string) error {
	if _, err := q.Exec(ctx, enqueueTimeBlockSQL, barbershopID, blockID); err != nil {
		return fmt.Errorf("googlecalendar/postgres: enqueue time block: %w", err)
	}
	return nil
}
