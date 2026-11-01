package postgres

import (
	"context"

	"system-barbershop/internal/platform/database"
)

// Tipos de cambio que originan un aviso al cliente (HU-130, DEC-142). Son los
// valores de `appointment_notice.kind`.
const (
	NoticeCancelledByBarber = "cancelled_by_barber"
	NoticeRescheduled       = "rescheduled"
)

// NoticeHook es el puerto con el que booking permite encolar el aviso al
// cliente DENTRO de la misma transacción del cambio, sin conocer cómo se envía
// (RN-REC-06). La implementación se inyecta en la raíz de composición (cmd/api).
//
// Contrato: se invoca solo cuando el barbero cambia el turno de verdad (la
// cancelación aplicada ahora o la reprogramación a otro intervalo), nunca en una
// repetición idempotente ni en un no-op, con el `q` de esa misma transacción. Si
// devuelve un error, la transacción completa se revierte: es preferible no
// confirmar el cambio a confirmarlo sin dejar el aviso en la cola.
type NoticeHook interface {
	AppointmentNoticeRequired(ctx context.Context, q database.Queries, barbershopID, appointmentID, kind string) error
	// AppointmentReminderRequired programa (o regenera, si el turno cambió de
	// hora) el recordatorio previo al turno (HU-131, DEC-144) en la transacción
	// que lo crea o lo reprograma. Decide por sí mismo si hay algo que recordar.
	AppointmentReminderRequired(ctx context.Context, q database.Queries, barbershopID, appointmentID string) error
}

// WithNoticeHook registra quien encola los avisos al cliente. Se llama durante
// la composición, antes de atender solicitudes; sin él el repositorio funciona
// igual (no se avisa a nadie).
func (r *Repository) WithNoticeHook(hook NoticeHook) *Repository {
	r.notices = hook
	return r
}

func (r *Repository) noticeRequired(ctx context.Context, q database.Queries, barbershopID, appointmentID, kind string) error {
	if r.notices == nil {
		return nil
	}
	return r.notices.AppointmentNoticeRequired(ctx, q, barbershopID, appointmentID, kind)
}

func (r *Repository) reminderRequired(ctx context.Context, q database.Queries, barbershopID, appointmentID string) error {
	if r.notices == nil {
		return nil
	}
	return r.notices.AppointmentReminderRequired(ctx, q, barbershopID, appointmentID)
}
