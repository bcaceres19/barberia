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
ON CONFLICT (appointment_id, (kind = 'reminder')) WHERE status = 'pending'
DO UPDATE SET kind = EXCLUDED.kind, state_at = EXCLUDED.state_at, run_at = EXCLUDED.run_at`

// La anticipación (30 minutos) es la inicial del recordatorio (RN-REC-01, DEC-144).
// No es configurable todavía (RN-REC-05 queda fuera del primer tramo).
//
// El recordatorio guarda como state_at el instante del turno: si el turno cambia
// de hora, el recordatorio viejo ya no coincide y el worker lo descarta al enviar
// (RN-REC-01, RN-REC-03). No se programa si faltan menos de reminderLead para el
// turno: no habría nada que recordar con antelación (un turno tomado "para ya").
// Un turno no confirmado no se programa. El índice único admite un recordatorio
// pendiente por turno: reprogramar lo actualiza (regenera) en vez de duplicarlo.
const reminderSQL = `
INSERT INTO appointment_notice (barbershop_id, appointment_id, kind, state_at, run_at)
SELECT a.barbershop_id, a.id, 'reminder', a.starts_at, a.starts_at - interval '30 minutes'
  FROM appointment a
 WHERE a.barbershop_id = $1 AND a.id = $2
   AND a.status = 'confirmed'
   AND a.starts_at - interval '30 minutes' > now()
ON CONFLICT (appointment_id, (kind = 'reminder')) WHERE status = 'pending'
DO UPDATE SET state_at = EXCLUDED.state_at, run_at = EXCLUDED.run_at`

// AppointmentReminderRequired implementa booking/postgres.NoticeHook.
func (Enqueuer) AppointmentReminderRequired(ctx context.Context, q database.Queries, barbershopID, appointmentID string) error {
	if _, err := q.Exec(ctx, reminderSQL, barbershopID, appointmentID); err != nil {
		return fmt.Errorf("appointmentnotice/postgres: enqueue reminder: %w", err)
	}
	return nil
}

// AppointmentNoticeRequired implementa booking/postgres.NoticeHook.
func (Enqueuer) AppointmentNoticeRequired(ctx context.Context, q database.Queries, barbershopID, appointmentID, kind string) error {
	if _, err := q.Exec(ctx, enqueueSQL, barbershopID, appointmentID, kind); err != nil {
		return fmt.Errorf("appointmentnotice/postgres: enqueue: %w", err)
	}
	return nil
}
