// Package postgres implementa los dos lados de la cola de avisos al cliente
// (HU-130, DEC-142): el Enqueuer, que la API usa dentro de la transacción del
// cambio, y el WorkerRepository, que el worker usa solo mediante funciones
// SECURITY DEFINER (DEC-040).
package postgres

import (
	"context"
	"fmt"

	"system-barbershop/internal/platform/database"
)

// Enqueuer implementa booking/postgres.NoticeHook con UNA sentencia: persiste el
// aviso dentro de la misma transacción que el cambio del turno. booking no
// importa este paquete: la raíz de composición lo inyecta.
//
// Hay como máximo un aviso `pending` por turno (índice único parcial): un cambio
// nuevo lo reemplaza (la cancelación sustituye a una reprogramación pendiente) y
// reinicia la espera de 60 segundos, de modo que dos cambios seguidos dan un solo
// correo (RN-REC-01).
type Enqueuer struct{}

// NewEnqueuer construye el encolador.
func NewEnqueuer() Enqueuer { return Enqueuer{} }

// state_at es el updated_at que el turno tiene tras el cambio, en esta misma
// transacción: identifica qué estado cubre cada aviso.
const enqueueSQL = `
INSERT INTO appointment_notice (barbershop_id, appointment_id, kind, state_at, run_at)
SELECT a.barbershop_id, a.id, $3, a.updated_at, now() + interval '60 seconds'
  FROM appointment a
 WHERE a.barbershop_id = $1 AND a.id = $2
ON CONFLICT (appointment_id) WHERE status = 'pending'
DO UPDATE SET kind = EXCLUDED.kind, state_at = EXCLUDED.state_at, run_at = EXCLUDED.run_at`

// AppointmentNoticeRequired implementa booking/postgres.NoticeHook.
func (Enqueuer) AppointmentNoticeRequired(ctx context.Context, q database.Queries, barbershopID, appointmentID, kind string) error {
	if _, err := q.Exec(ctx, enqueueSQL, barbershopID, appointmentID, kind); err != nil {
		return fmt.Errorf("appointmentnotice/postgres: enqueue: %w", err)
	}
	return nil
}
