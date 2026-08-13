package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/platform/database"
)

var _ auth.SessionRepository = (*Repository)(nil)

// ResolveSessionTenant implementa auth.SessionRepository llamando a la
// función SECURITY DEFINER authn_resolve_session_tenant, ya creada por la
// migración de HU-005 (20260813120000_create_auth_credentials_and_sessions.sql)
// pero sin consumidor hasta ahora, ANTES de que exista contexto de tenant,
// mediante la misma excepción angosta database.DB.ResolveTenant que usa
// ResolveLoginTenant.
func (r *Repository) ResolveSessionTenant(ctx context.Context, tokenHash string) (string, bool, error) {
	shop, found, err := r.db.ResolveTenant(ctx,
		"SELECT authn_resolve_session_tenant($1)", tokenHash)
	if err != nil {
		return "", false, fmt.Errorf("auth/postgres: resolve session tenant: %w", err)
	}
	return string(shop), found, nil
}

// ValidateAndRenewSession revalida la sesión DENTRO de la transacción
// tenant-aware (RLS) con una única sentencia UPDATE ... FROM ... RETURNING:
// reconfirma revoked_at, expires_at y que el usuario siga activo, y si
// sigue siendo válida extiende last_used_at/expires_at en la MISMA
// operación atómica.
//
// No hace falta un SELECT ... FOR UPDATE explícito por separado: un UPDATE
// ya toma el bloqueo de fila necesario. Si un logout concurrente confirma
// primero, esta sentencia -bloqueada brevemente por ese bloqueo de fila-
// vuelve a evaluar su WHERE contra la fila ya comprometida (revoked_at no
// nulo) al desbloquearse, y no encuentra ninguna fila que actualizar
// (found=false): el resultado final siempre queda revocado, y esta
// renovación nunca reabre ni revive una sesión que acaba de revocarse.
func (r *Repository) ValidateAndRenewSession(ctx context.Context, barbershopID, tokenHash string, now, newExpiresAt time.Time) (auth.Principal, bool, error) {
	var principal auth.Principal
	found := false

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var sessionID, staffUserID string
		err := q.QueryRow(ctx, `
			UPDATE staff_session s
			   SET last_used_at = $2, expires_at = $3
			  FROM staff_user u
			 WHERE s.token_hash = $1
			   AND u.id = s.staff_user_id
			   AND u.barbershop_id = s.barbershop_id
			   AND s.revoked_at IS NULL
			   AND s.expires_at > $2
			   AND u.is_active
			RETURNING s.id, s.staff_user_id`,
			tokenHash, now, newExpiresAt,
		).Scan(&sessionID, &staffUserID)
		switch {
		case err == nil:
			principal = auth.Principal{
				SessionID:    sessionID,
				StaffUserID:  staffUserID,
				BarbershopID: barbershopID,
			}
			found = true
		case errors.Is(err, pgx.ErrNoRows):
			// Vencida, revocada, de un usuario inactivo o desajustada de
			// tenant: found queda false, sin error.
		default:
			return fmt.Errorf("renew staff_session: %w", err)
		}
		return nil
	})
	if err != nil {
		return auth.Principal{}, false, fmt.Errorf("auth/postgres: validate and renew session: %w", err)
	}
	return principal, found, nil
}

// RevokeSession fija revoked_at = now en la fila sessionID, solo si
// pertenece a barbershopID y no estaba revocada todavía. No distingue ni
// falla si la fila ya estaba revocada o no existe dentro del tenant: el
// efecto observable de logout es idempotente (CA-006-02). RLS
// (staff_session_update_tenant_policy) y el filtro explícito de
// barbershop_id hacen estructuralmente imposible que esta sentencia revoque
// una fila de otro tenant, incluso si sessionID llegara corrupto -en la
// práctica nunca ocurre: sessionID proviene siempre de un auth.Principal ya
// validado por el middleware de sesión, jamás de un parámetro que el
// cliente controle (CA-006-07, DEC-058).
func (r *Repository) RevokeSession(ctx context.Context, barbershopID, sessionID string, now time.Time) error {
	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		_, err := q.Exec(ctx, `
			UPDATE staff_session
			   SET revoked_at = $3
			 WHERE id = $1
			   AND barbershop_id = $2
			   AND revoked_at IS NULL`,
			sessionID, barbershopID, now,
		)
		return err
	})
	if err != nil {
		return fmt.Errorf("auth/postgres: revoke session: %w", err)
	}
	return nil
}
