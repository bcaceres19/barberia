package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"system-barbershop/internal/modules/googlecalendar"
	"system-barbershop/internal/platform/database"
)

// WorkerRepository implementa googlecalendar.JobStore contra las funciones
// SECURITY DEFINER concedidas exclusivamente a barberia_worker (DEC-040,
// DEC-102): el worker no tiene acceso directo a ninguna tabla ni fija contexto
// de tenant. db debe ser el pool del proceso worker (cfg.WorkerDatabaseURL).
type WorkerRepository struct {
	db *database.DB
}

// NewWorkerRepository construye el repositorio del worker.
func NewWorkerRepository(db *database.DB) *WorkerRepository { return &WorkerRepository{db: db} }

var _ googlecalendar.JobStore = (*WorkerRepository)(nil)

// ClaimJobs implementa googlecalendar.JobStore.ClaimJobs.
func (r *WorkerRepository) ClaimJobs(ctx context.Context, limit, leaseSeconds int, now time.Time) ([]googlecalendar.Job, error) {
	var jobs []googlecalendar.Job
	err := r.db.CallSecurityDefinerRows(ctx,
		`SELECT job_id::text, claim_token::text, barbershop_id::text, connection_id::text,
		        resource_type, resource_id::text, attempts
		   FROM gcal_claim_jobs($1, $2, $3)`,
		[]any{limit, leaseSeconds, now},
		func(rows pgx.Rows) error {
			for rows.Next() {
				var j googlecalendar.Job
				if err := rows.Scan(&j.ID, &j.ClaimToken, &j.BarbershopID, &j.ConnectionID,
					&j.ResourceType, &j.ResourceID, &j.Attempts); err != nil {
					return err
				}
				jobs = append(jobs, j)
			}
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("googlecalendar/postgres: claim jobs: %w", err)
	}
	return jobs, nil
}

// JobContext implementa googlecalendar.JobStore.JobContext.
func (r *WorkerRepository) JobContext(ctx context.Context, jobID, claimToken string) (googlecalendar.JobContext, bool, error) {
	var jc googlecalendar.JobContext
	found := false
	err := r.db.CallSecurityDefinerRow(ctx,
		`SELECT barber_id::text, connection_status, calendar_id, reminder_minutes, token_ciphertext,
		        coalesce(token_key_id, ''), barbershop_timezone,
		        coalesce(link_event_id, ''), coalesce(link_generation, 0), coalesce(link_visible_hash, ''),
		        resource_type, resource_id::text, resource_found, coalesce(barber_matches, false),
		        coalesce(appt_status, ''), starts_at, ends_at,
		        coalesce(attendee_name, ''), coalesce(service_name, ''), coalesce(customer_email, ''),
		        coalesce(block_type, ''), coalesce(block_source, ''), coalesce(block_deleted, false)
		   FROM gcal_job_context($1::uuid, $2::uuid)`,
		[]any{jobID, claimToken},
		func(row pgx.Row) error {
			var startsAt, endsAt *time.Time
			err := row.Scan(&jc.BarberID, &jc.ConnectionStatus, &jc.CalendarID, &jc.ReminderMinutes, &jc.TokenCiphertext,
				&jc.TokenKeyID, &jc.Timezone,
				&jc.LinkEventID, &jc.LinkGeneration, &jc.LinkVisibleHash,
				&jc.ResourceType, &jc.ResourceID, &jc.ResourceFound, &jc.BarberMatches,
				&jc.ApptStatus, &startsAt, &endsAt,
				&jc.AttendeeName, &jc.ServiceName, &jc.CustomerEmail,
				&jc.BlockType, &jc.BlockSource, &jc.BlockDeleted)
			if errors.Is(err, pgx.ErrNoRows) {
				return nil // la reclamación ya no es vigente
			}
			if err != nil {
				return err
			}
			if startsAt != nil {
				jc.StartsAt = *startsAt
			}
			if endsAt != nil {
				jc.EndsAt = *endsAt
			}
			found = true
			return nil
		})
	if err != nil {
		return googlecalendar.JobContext{}, false, fmt.Errorf("googlecalendar/postgres: job context: %w", err)
	}
	return jc, found, nil
}

// FinishJob implementa googlecalendar.JobStore.FinishJob.
func (r *WorkerRepository) FinishJob(ctx context.Context, in googlecalendar.FinishInput) (bool, error) {
	var ok bool
	err := r.db.CallSecurityDefinerRow(ctx,
		`SELECT gcal_finish_job($1::uuid, $2::uuid, $3, NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, 0),
		                        NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, 0), $10, $11)`,
		[]any{in.JobID, in.ClaimToken, in.Outcome, in.EventID, in.ETag, in.Generation,
			in.VisibleHash, in.ErrorCode, in.RetrySeconds, in.MaxAttempts, in.Now},
		func(row pgx.Row) error { return row.Scan(&ok) })
	if err != nil {
		return false, fmt.Errorf("googlecalendar/postgres: finish job: %w", err)
	}
	return ok, nil
}

// RestoreClaimConnection implementa googlecalendar.JobStore.RestoreClaimConnection.
func (r *WorkerRepository) RestoreClaimConnection(ctx context.Context, intervalSeconds int, now time.Time) (googlecalendar.RestoreTarget, bool, error) {
	var t googlecalendar.RestoreTarget
	found := false
	err := r.db.CallSecurityDefinerRow(ctx,
		`SELECT connection_id::text, barbershop_id::text, barber_id::text, calendar_id, token_ciphertext,
		        coalesce(token_key_id, '')
		   FROM gcal_restore_claim_connection($1, $2)`,
		[]any{intervalSeconds, now},
		func(row pgx.Row) error {
			err := row.Scan(&t.ConnectionID, &t.BarbershopID, &t.BarberID, &t.CalendarID, &t.TokenCiphertext, &t.TokenKeyID)
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			if err != nil {
				return err
			}
			found = true
			return nil
		})
	if err != nil {
		return googlecalendar.RestoreTarget{}, false, fmt.Errorf("googlecalendar/postgres: restore claim: %w", err)
	}
	return t, found, nil
}

// RestoreLinks implementa googlecalendar.JobStore.RestoreLinks.
func (r *WorkerRepository) RestoreLinks(ctx context.Context, connectionID string, now time.Time) ([]googlecalendar.RestoreLink, error) {
	var links []googlecalendar.RestoreLink
	err := r.db.CallSecurityDefinerRows(ctx,
		`SELECT resource_type, resource_id::text, google_event_id FROM gcal_restore_links($1::uuid, $2)`,
		[]any{connectionID, now},
		func(rows pgx.Rows) error {
			for rows.Next() {
				var l googlecalendar.RestoreLink
				if err := rows.Scan(&l.ResourceType, &l.ResourceID, &l.EventID); err != nil {
					return err
				}
				links = append(links, l)
			}
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("googlecalendar/postgres: restore links: %w", err)
	}
	return links, nil
}

// EnqueueMissing implementa googlecalendar.JobStore.EnqueueMissing.
func (r *WorkerRepository) EnqueueMissing(ctx context.Context, connectionID, resourceType, resourceID string) (bool, error) {
	var queued bool
	err := r.db.CallSecurityDefinerRow(ctx,
		`SELECT gcal_enqueue_missing($1::uuid, $2, $3::uuid)`,
		[]any{connectionID, resourceType, resourceID},
		func(row pgx.Row) error { return row.Scan(&queued) })
	if err != nil {
		return false, fmt.Errorf("googlecalendar/postgres: enqueue missing: %w", err)
	}
	return queued, nil
}
