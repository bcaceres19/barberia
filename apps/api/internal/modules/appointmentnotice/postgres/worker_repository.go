package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"system-barbershop/internal/modules/appointmentnotice"
	"system-barbershop/internal/platform/database"
)

// WorkerRepository implementa appointmentnotice.Store contra las funciones
// SECURITY DEFINER concedidas exclusivamente a barberia_worker (DEC-040): el
// worker no tiene acceso directo a ninguna tabla ni fija contexto de tenant. db
// debe ser el pool del proceso worker (cfg.WorkerDatabaseURL).
type WorkerRepository struct {
	db *database.DB
}

// NewWorkerRepository construye el repositorio del worker.
func NewWorkerRepository(db *database.DB) *WorkerRepository { return &WorkerRepository{db: db} }

var _ appointmentnotice.Store = (*WorkerRepository)(nil)

// ClaimNotices implementa appointmentnotice.Store.ClaimNotices.
func (r *WorkerRepository) ClaimNotices(ctx context.Context, limit, leaseSeconds, maxAttempts int, now time.Time) ([]appointmentnotice.Notice, error) {
	var notices []appointmentnotice.Notice
	err := r.db.CallSecurityDefinerRows(ctx,
		`SELECT notice_id::text, claim_token::text, attempts FROM notice_claim($1, $2, $3, $4)`,
		[]any{limit, leaseSeconds, maxAttempts, now},
		func(rows pgx.Rows) error {
			for rows.Next() {
				var n appointmentnotice.Notice
				if err := rows.Scan(&n.ID, &n.ClaimToken, &n.Attempts); err != nil {
					return err
				}
				notices = append(notices, n)
			}
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("appointmentnotice/postgres: claim notices: %w", err)
	}
	return notices, nil
}

// NoticeContext implementa appointmentnotice.Store.NoticeContext.
func (r *WorkerRepository) NoticeContext(ctx context.Context, noticeID, claimToken string) (appointmentnotice.Context, bool, error) {
	var c appointmentnotice.Context
	found := false
	err := r.db.CallSecurityDefinerRow(ctx,
		`SELECT kind, appointment_status, starts_at, appointment_state_at, coalesce(customer_email, ''),
		        attendee_name, service_name, barbershop_name, barbershop_timezone, already_covered
		   FROM notice_context($1::uuid, $2::uuid)`,
		[]any{noticeID, claimToken},
		func(row pgx.Row) error {
			err := row.Scan(&c.Kind, &c.AppointmentStatus, &c.StartsAt, &c.StateAt, &c.CustomerEmail,
				&c.AttendeeName, &c.ServiceName, &c.BarbershopName, &c.Timezone, &c.AlreadyCovered)
			if errors.Is(err, pgx.ErrNoRows) {
				return nil // la reclamación ya no es vigente
			}
			if err != nil {
				return err
			}
			found = true
			return nil
		})
	if err != nil {
		return appointmentnotice.Context{}, false, fmt.Errorf("appointmentnotice/postgres: notice context: %w", err)
	}
	return c, found, nil
}

// FinishNotice implementa appointmentnotice.Store.FinishNotice.
func (r *WorkerRepository) FinishNotice(ctx context.Context, in appointmentnotice.FinishInput) (bool, error) {
	var covers *time.Time
	if !in.CoversStateAt.IsZero() {
		covers = &in.CoversStateAt
	}
	var ok bool
	err := r.db.CallSecurityDefinerRow(ctx,
		`SELECT notice_finish($1::uuid, $2::uuid, $3, NULLIF($4, ''), NULLIF($5, 0), $6, $7, $8)`,
		[]any{in.NoticeID, in.ClaimToken, in.Outcome, in.Code, in.RetrySeconds, in.MaxAttempts, covers, in.Now},
		func(row pgx.Row) error { return row.Scan(&ok) })
	if err != nil {
		return false, fmt.Errorf("appointmentnotice/postgres: finish notice: %w", err)
	}
	return ok, nil
}
