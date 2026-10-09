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
		default:
			return fmt.Errorf("update reminder: %w", err)
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
