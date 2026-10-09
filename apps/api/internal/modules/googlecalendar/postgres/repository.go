// Package postgres es el adaptador de persistencia de googlecalendar.Repository
// sobre database.DB: el núcleo no importa este paquete ni pgx. Toda operación
// corre en una transacción con el tenant fijado y filtra además por
// barbershop_id (defensa en profundidad sobre RLS).
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

// Repository implementa googlecalendar.Repository.
type Repository struct {
	db *database.DB
}

// New construye el repositorio.
func New(db *database.DB) *Repository { return &Repository{db: db} }

var _ googlecalendar.Repository = (*Repository)(nil)

const connectionColumns = `id::text, barbershop_id::text, barber_id::text, status,
	coalesce(google_account_email, ''), calendar_id, reminder_minutes,
	connected_at, disconnected_at, last_synced_at, coalesce(last_error_code, ''),
	refresh_token_ciphertext, coalesce(token_key_id, '')`

func scanConnection(row pgx.Row) (googlecalendar.Connection, error) {
	var c googlecalendar.Connection
	var status string
	err := row.Scan(&c.ID, &c.BarbershopID, &c.BarberID, &status, &c.AccountEmail, &c.CalendarID,
		&c.ReminderMinutes, &c.ConnectedAt, &c.DisconnectedAt, &c.LastSyncedAt, &c.LastErrorCode,
		&c.Credentials, &c.KeyID)
	c.Status = googlecalendar.Status(status)
	return c, err
}

// GetConnection implementa googlecalendar.Repository.GetConnection.
func (r *Repository) GetConnection(ctx context.Context, barbershopID, barberID string) (googlecalendar.Connection, bool, error) {
	var conn googlecalendar.Connection
	found := false
	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		c, err := scanConnection(q.QueryRow(ctx,
			`SELECT `+connectionColumns+` FROM google_calendar_connection
			  WHERE barbershop_id = $1 AND barber_id = $2`, barbershopID, barberID))
		switch {
		case err == nil:
			conn, found = c, true
		case errors.Is(err, pgx.ErrNoRows):
		default:
			return fmt.Errorf("select connection: %w", err)
		}
		return nil
	})
	if err != nil {
		return googlecalendar.Connection{}, false, fmt.Errorf("googlecalendar/postgres: get connection: %w", err)
	}
	return conn, found, nil
}

// SaveConnected implementa googlecalendar.Repository.SaveConnected. La
// restricción UNIQUE (barbershop_id, barber_id) convierte el INSERT en el
// «crear o reutilizar» de un barbero: reconectar no deja filas duplicadas.
func (r *Repository) SaveConnected(ctx context.Context, barbershopID, barberID string, data googlecalendar.ConnectedData) (googlecalendar.Connection, error) {
	var conn googlecalendar.Connection
	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		c, err := scanConnection(q.QueryRow(ctx,
			`INSERT INTO google_calendar_connection AS g
			    (barbershop_id, barber_id, status, google_account_email, refresh_token_ciphertext,
			     token_key_id, connected_at)
			 VALUES ($1, $2, 'connected', NULLIF($3, ''), $4, $5, $6)
			 ON CONFLICT (barbershop_id, barber_id) DO UPDATE SET
			    status = 'connected',
			    google_account_email = NULLIF($3, ''),
			    refresh_token_ciphertext = $4,
			    token_key_id = $5,
			    connected_at = $6,
			    disconnected_at = NULL,
			    last_error_code = NULL
			 RETURNING `+connectionColumns,
			barbershopID, barberID, data.AccountEmail, data.Credentials, data.KeyID, data.At))
		if err != nil {
			return fmt.Errorf("upsert connection: %w", err)
		}
		conn = c
		if _, err := q.Exec(ctx, enqueueInitialSQL, barbershopID, barberID, data.At); err != nil {
			return fmt.Errorf("enqueue initial publication: %w", err)
		}
		return nil
	})
	if err != nil {
		return googlecalendar.Connection{}, fmt.Errorf("googlecalendar/postgres: save connected: %w", err)
	}
	return conn, nil
}

// MarkStatus implementa googlecalendar.Repository.MarkStatus.
func (r *Repository) MarkStatus(ctx context.Context, barbershopID, barberID string, status googlecalendar.Status, errorCode string, at time.Time) (bool, error) {
	var query string
	args := []any{barbershopID, barberID, errorCode}
	switch status {
	case googlecalendar.StatusDisconnected:
		// Desconectar borra credenciales y correo de la cuenta; conserva el
		// barbero, la preferencia de recordatorio y las fechas.
		query = `UPDATE google_calendar_connection
		            SET status = 'disconnected', refresh_token_ciphertext = NULL, token_key_id = NULL,
		                google_account_email = NULL, disconnected_at = $4, last_error_code = NULLIF($3, '')
		          WHERE barbershop_id = $1 AND barber_id = $2`
		args = append(args, at)
	case googlecalendar.StatusReauthRequired:
		query = `UPDATE google_calendar_connection
		            SET status = 'reauth_required', refresh_token_ciphertext = NULL, token_key_id = NULL,
		                last_error_code = NULLIF($3, '')
		          WHERE barbershop_id = $1 AND barber_id = $2`
	case googlecalendar.StatusError:
		query = `UPDATE google_calendar_connection
		            SET status = 'error', last_error_code = NULLIF($3, '')
		          WHERE barbershop_id = $1 AND barber_id = $2 AND status IN ('connected', 'error')`
	default:
		return false, fmt.Errorf("googlecalendar/postgres: estado %q no se asigna con MarkStatus", status)
	}

	found := false
	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		tag, err := q.Exec(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("update status: %w", err)
		}
		found = tag.RowsAffected() > 0
		// Una conexión que ya no publica detiene su cola (DEC-101.11); al volver a
		// conectar, la publicación inicial reencola lo vigente.
		if found && (status == googlecalendar.StatusDisconnected || status == googlecalendar.StatusReauthRequired) {
			if _, err := q.Exec(ctx, deleteConnectionJobsSQL, barbershopID, barberID); err != nil {
				return fmt.Errorf("drop connection jobs: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("googlecalendar/postgres: mark status: %w", err)
	}
	return found, nil
}

// SetReminder implementa googlecalendar.Repository.SetReminder.
func (r *Repository) SetReminder(ctx context.Context, barbershopID, barberID string, minutes *int) (googlecalendar.Connection, bool, error) {
	var conn googlecalendar.Connection
	found := false
	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		c, err := scanConnection(q.QueryRow(ctx,
			`UPDATE google_calendar_connection SET reminder_minutes = $3
			  WHERE barbershop_id = $1 AND barber_id = $2 AND status <> 'disconnected'
			  RETURNING `+connectionColumns, barbershopID, barberID, minutes))
		switch {
		case err == nil:
			conn, found = c, true
		case errors.Is(err, pgx.ErrNoRows):
			return nil
		default:
			return fmt.Errorf("update reminder: %w", err)
		}
		if _, err := q.Exec(ctx, enqueueReminderRefreshSQL, barbershopID, barberID, time.Now().UTC()); err != nil {
			return fmt.Errorf("enqueue reminder refresh: %w", err)
		}
		return nil
	})
	if err != nil {
		return googlecalendar.Connection{}, false, fmt.Errorf("googlecalendar/postgres: set reminder: %w", err)
	}
	return conn, found, nil
}

// CreateState implementa googlecalendar.Repository.CreateState.
func (r *Repository) CreateState(ctx context.Context, state googlecalendar.OAuthState, now time.Time) error {
	err := r.db.InTenantTx(ctx, database.BarbershopID(state.BarbershopID), func(ctx context.Context, q database.Queries) error {
		// Purga oportunista: los estados vencidos o ya usados de este barbero
		// no se acumulan aunque nadie ejecute un trabajo de limpieza.
		if _, err := q.Exec(ctx,
			`DELETE FROM google_calendar_oauth_state
			  WHERE barbershop_id = $1 AND barber_id = $2
			    AND (expires_at <= $3 OR consumed_at IS NOT NULL)`,
			state.BarbershopID, state.BarberID, now); err != nil {
			return fmt.Errorf("purge states: %w", err)
		}
		if _, err := q.Exec(ctx,
			`INSERT INTO google_calendar_oauth_state
			    (barbershop_id, barber_id, staff_user_id, session_id, state_hash,
			     code_verifier_ciphertext, verifier_key_id, expires_at, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			state.BarbershopID, state.BarberID, state.StaffUserID, state.SessionID, state.StateHash,
			state.VerifierEncrypted, state.VerifierKeyID, state.ExpiresAt, now); err != nil {
			return fmt.Errorf("insert state: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("googlecalendar/postgres: create state: %w", err)
	}
	return nil
}

// ConsumeState implementa googlecalendar.Repository.ConsumeState. El UPDATE
// condicional es atómico: de dos callbacks con el mismo state solo uno recibe
// la fila.
func (r *Repository) ConsumeState(ctx context.Context, barbershopID, stateHash string, now time.Time) (googlecalendar.OAuthState, bool, error) {
	var state googlecalendar.OAuthState
	found := false
	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		err := q.QueryRow(ctx,
			`UPDATE google_calendar_oauth_state SET consumed_at = $3
			  WHERE barbershop_id = $1 AND state_hash = $2
			    AND consumed_at IS NULL AND expires_at > $3
			  RETURNING barbershop_id::text, barber_id::text, staff_user_id::text, session_id::text,
			            state_hash, code_verifier_ciphertext, verifier_key_id, expires_at`,
			barbershopID, stateHash, now,
		).Scan(&state.BarbershopID, &state.BarberID, &state.StaffUserID, &state.SessionID,
			&state.StateHash, &state.VerifierEncrypted, &state.VerifierKeyID, &state.ExpiresAt)
		switch {
		case err == nil:
			found = true
		case errors.Is(err, pgx.ErrNoRows):
		default:
			return fmt.Errorf("consume state: %w", err)
		}
		return nil
	})
	if err != nil {
		return googlecalendar.OAuthState{}, false, fmt.Errorf("googlecalendar/postgres: consume state: %w", err)
	}
	return state, found, nil
}

// BarberOfUser implementa googlecalendar.Repository.BarberOfUser. Lee el
// vínculo de DEC-100 directamente: el módulo dueño de `barber` (staff) no se
// importa, igual que el resto de módulos que consultan datos ajenos.
func (r *Repository) BarberOfUser(ctx context.Context, barbershopID, staffUserID string) (string, bool, error) {
	barberID := ""
	found := false
	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		err := q.QueryRow(ctx,
			`SELECT id::text FROM barber WHERE barbershop_id = $1 AND staff_user_id = $2`,
			barbershopID, staffUserID).Scan(&barberID)
		switch {
		case err == nil:
			found = true
		case errors.Is(err, pgx.ErrNoRows):
		default:
			return fmt.Errorf("select linked barber: %w", err)
		}
		return nil
	})
	if err != nil {
		return "", false, fmt.Errorf("googlecalendar/postgres: barber of user: %w", err)
	}
	return barberID, found, nil
}

const enqueueInitialSQL = `
INSERT INTO google_calendar_sync_job (barbershop_id, connection_id, resource_type, resource_id)
SELECT c.barbershop_id, c.id, 'appointment', a.id
  FROM google_calendar_connection c
  JOIN appointment a ON a.barbershop_id = c.barbershop_id AND a.barber_id = c.barber_id
 WHERE c.barbershop_id = $1 AND c.barber_id = $2 AND c.status = 'connected'
   AND a.status = 'confirmed' AND a.ends_at > $3::timestamptz
   AND a.starts_at < $3::timestamptz + interval '6 months'
UNION
SELECT c.barbershop_id, c.id, 'time_block', t.id
  FROM google_calendar_connection c
  JOIN time_block t ON t.barbershop_id = c.barbershop_id AND t.barber_id = c.barber_id
 WHERE c.barbershop_id = $1 AND c.barber_id = $2 AND c.status = 'connected'
   AND t.deleted_at IS NULL AND t.source = 'manual' AND t.block_type <> 'holiday'
   AND t.ends_at > $3::timestamptz AND t.starts_at < $3::timestamptz + interval '6 months'
UNION
SELECT l.barbershop_id, l.connection_id, l.resource_type, l.resource_id
  FROM google_calendar_event_link l
  JOIN google_calendar_connection c ON c.barbershop_id = l.barbershop_id AND c.id = l.connection_id
 WHERE c.barbershop_id = $1 AND c.barber_id = $2 AND c.status = 'connected'
ON CONFLICT (connection_id, resource_type, resource_id) WHERE status = 'pending' DO NOTHING`

const enqueueReminderRefreshSQL = `
INSERT INTO google_calendar_sync_job (barbershop_id, connection_id, resource_type, resource_id)
SELECT l.barbershop_id, l.connection_id, l.resource_type, l.resource_id
  FROM google_calendar_event_link l
  JOIN google_calendar_connection c ON c.barbershop_id = l.barbershop_id AND c.id = l.connection_id
  LEFT JOIN appointment a ON l.resource_type = 'appointment' AND a.barbershop_id = l.barbershop_id AND a.id = l.resource_id
  LEFT JOIN time_block t ON l.resource_type = 'time_block' AND t.barbershop_id = l.barbershop_id AND t.id = l.resource_id
 WHERE c.barbershop_id = $1 AND c.barber_id = $2 AND c.status IN ('connected', 'error')
   AND ((a.id IS NOT NULL AND a.status = 'confirmed' AND a.ends_at > $3::timestamptz)
     OR (t.id IS NOT NULL AND t.deleted_at IS NULL AND t.ends_at > $3::timestamptz))
ON CONFLICT (connection_id, resource_type, resource_id) WHERE status = 'pending' DO NOTHING`

const deleteConnectionJobsSQL = `
DELETE FROM google_calendar_sync_job
 WHERE barbershop_id = $1
   AND connection_id = (SELECT id FROM google_calendar_connection WHERE barbershop_id = $1 AND barber_id = $2)`

// JobCounts implementa googlecalendar.Repository.JobCounts.
func (r *Repository) JobCounts(ctx context.Context, barbershopID, barberID string) (googlecalendar.JobCounts, error) {
	var counts googlecalendar.JobCounts
	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx,
			`SELECT count(*) FILTER (WHERE j.status IN ('pending', 'processing')),
			        count(*) FILTER (WHERE j.status = 'failed')
			   FROM google_calendar_sync_job j
			   JOIN google_calendar_connection c ON c.barbershop_id = j.barbershop_id AND c.id = j.connection_id
			  WHERE c.barbershop_id = $1 AND c.barber_id = $2`,
			barbershopID, barberID).Scan(&counts.Pending, &counts.Failed)
	})
	if err != nil {
		return googlecalendar.JobCounts{}, fmt.Errorf("googlecalendar/postgres: job counts: %w", err)
	}
	return counts, nil
}

// RequeueConnection implementa googlecalendar.Repository.RequeueConnection.
func (r *Repository) RequeueConnection(ctx context.Context, barbershopID, barberID string, now time.Time) (googlecalendar.JobCounts, bool, error) {
	var counts googlecalendar.JobCounts
	found := false
	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var connectionID string
		err := q.QueryRow(ctx,
			`SELECT id::text FROM google_calendar_connection
			  WHERE barbershop_id = $1 AND barber_id = $2 AND status IN ('connected', 'error')`,
			barbershopID, barberID).Scan(&connectionID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("select connection: %w", err)
		}
		found = true

		// Un fallido con un gemelo más nuevo ya no aporta nada: se descarta para
		// que reactivar el resto no choque con el índice de un único pendiente.
		if _, err := q.Exec(ctx,
			`DELETE FROM google_calendar_sync_job f
			  WHERE f.barbershop_id = $1 AND f.connection_id = $2 AND f.status = 'failed'
			    AND EXISTS (
			      SELECT 1 FROM google_calendar_sync_job o
			       WHERE o.connection_id = f.connection_id AND o.resource_type = f.resource_type
			         AND o.resource_id = f.resource_id AND o.id <> f.id
			         AND (o.status = 'pending'
			              OR (o.status = 'failed' AND (o.created_at, o.id) > (f.created_at, f.id))))`,
			barbershopID, connectionID); err != nil {
			return fmt.Errorf("drop superseded failed jobs: %w", err)
		}
		if _, err := q.Exec(ctx,
			`UPDATE google_calendar_sync_job
			    SET status = 'pending', attempts = 0, run_at = $3, last_error_code = NULL
			  WHERE barbershop_id = $1 AND connection_id = $2 AND status = 'failed'`,
			barbershopID, connectionID, now); err != nil {
			return fmt.Errorf("retry failed jobs: %w", err)
		}
		if _, err := q.Exec(ctx,
			`UPDATE google_calendar_sync_job SET run_at = LEAST(run_at, $3)
			  WHERE barbershop_id = $1 AND connection_id = $2 AND status = 'pending'`,
			barbershopID, connectionID, now); err != nil {
			return fmt.Errorf("make pending jobs due: %w", err)
		}
		// Reintentar tras un error permanente: si vuelve a fallar, el worker lo marca de nuevo.
		if _, err := q.Exec(ctx,
			`UPDATE google_calendar_connection SET status = 'connected', last_error_code = NULL
			  WHERE barbershop_id = $1 AND id = $2 AND status = 'error'`,
			barbershopID, connectionID); err != nil {
			return fmt.Errorf("reset connection error: %w", err)
		}
		return q.QueryRow(ctx,
			`SELECT count(*) FILTER (WHERE status IN ('pending', 'processing')),
			        count(*) FILTER (WHERE status = 'failed')
			   FROM google_calendar_sync_job WHERE barbershop_id = $1 AND connection_id = $2`,
			barbershopID, connectionID).Scan(&counts.Pending, &counts.Failed)
	})
	if err != nil {
		return googlecalendar.JobCounts{}, false, fmt.Errorf("googlecalendar/postgres: requeue connection: %w", err)
	}
	return counts, found, nil
}
