// Package postgres es el adaptador de persistencia de
// customeraccess.Repository sobre database.DB (mismo criterio de
// independencia que publicbooking/postgres: el núcleo de customeraccess no
// importa este paquete ni pgx).
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"system-barbershop/internal/modules/customeraccess"
	"system-barbershop/internal/platform/database"
)

// Repository implementa customeraccess.Repository sobre database.DB.
type Repository struct {
	db *database.DB
}

// New construye el repositorio.
func New(db *database.DB) *Repository {
	return &Repository{db: db}
}

var _ customeraccess.Repository = (*Repository)(nil)

// ResolveAppointmentByTokenHash implementa customeraccess.Repository en dos
// pasos, exactamente el patrón ya establecido por
// publicbooking.Repository.ResolveBySlug (public_resolve_barbershop_by_slug,
// HU-090, database.DB.ResolveTenant): primero resuelve el hash a un
// identificador de barbería SIN contexto de tenant, mediante la función
// SECURITY DEFINER estrecha public_resolve_appointment_token_tenant (que
// responde igual -NULL- a un hash inexistente, vencido o revocado,
// CA-098-02); si resuelve, abre una transacción tenant-aware normal
// (InTenantTx) sobre ESE identificador para leer la proyección mínima ya
// protegida por RLS, exigiendo de nuevo el mismo hash vigente -defensa en
// profundidad frente a una revocación ocurrida entre ambos pasos, sin
// necesitar ningún privilegio adicional sobre appointment/barber/
// barbershop.
func (r *Repository) ResolveAppointmentByTokenHash(ctx context.Context, tokenHash string) (customeraccess.AppointmentView, bool, error) {
	barbershopID, found, err := r.db.ResolveTenant(ctx,
		`SELECT public_resolve_appointment_token_tenant($1)`,
		tokenHash,
	)
	if err != nil {
		return customeraccess.AppointmentView{}, false, fmt.Errorf("customeraccess/postgres: resolve tenant by token: %w", err)
	}
	if !found {
		return customeraccess.AppointmentView{}, false, nil
	}

	var view customeraccess.AppointmentView
	viewFound := false
	err = r.db.InTenantTx(ctx, barbershopID, func(ctx context.Context, q database.Queries) error {
		err := q.QueryRow(ctx,
			`SELECT bs.name, bs.timezone,
			        ap.attendee_name, ap.service_name_snapshot, ap.duration_minutes_snapshot,
			        br.full_name, ap.starts_at, ap.ends_at, ap.status,
			        bs.cancellation_deadline_minutes, bs.late_cancellation_client_allowed,
			        bs.late_cancellation_reason_required
			   FROM appointment_access_token t
			   JOIN appointment ap ON ap.barbershop_id = t.barbershop_id AND ap.id = t.appointment_id
			   JOIN barbershop bs ON bs.id = t.barbershop_id
			   JOIN barber br    ON br.barbershop_id = t.barbershop_id AND br.id = ap.barber_id
			  WHERE t.token_hash = $1
			    AND t.revoked_at IS NULL
			    AND (t.expires_at IS NULL OR t.expires_at > now())`,
			tokenHash,
		).Scan(
			&view.BarbershopName, &view.Timezone,
			&view.AttendeeName, &view.ServiceName, &view.DurationMinutes,
			&view.BarberName, &view.StartsAt, &view.EndsAt, &view.Status,
			&view.CancellationDeadlineMinutes, &view.LateCancellationClientAllowed,
			&view.LateCancellationReasonRequired,
		)
		switch {
		case err == nil:
			viewFound = true
		case errors.Is(err, pgx.ErrNoRows):
			// viewFound queda false: defensivo (la función SECURITY
			// DEFINER ya confirmó la fila vigente en la misma sentencia
			// SQL que esta transacción re-lee; una revocación entre ambas
			// solo podría venir de HU-099 o de un futuro worker de
			// anonimización -CA-098-02 exige responder igual, no un error).
		default:
			return fmt.Errorf("select appointment by token: %w", err)
		}
		return nil
	})
	if err != nil {
		return customeraccess.AppointmentView{}, false, fmt.Errorf("customeraccess/postgres: read appointment view: %w", err)
	}
	return view, viewFound, nil
}
