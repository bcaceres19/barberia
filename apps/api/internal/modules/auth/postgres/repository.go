// Package postgres es el adaptador PostgreSQL del módulo auth: traduce
// entre el puerto auth.Repository y database.DB.InTenantTx/ResolveTenant,
// sin que el núcleo del módulo (domain.go, ports.go, service.go) importe
// pgx ni internal/platform/database (CA-002-06). Las consultas SQL viven
// aquí, visibles y revisables, según docs/03-desarrollo/estandar-backend-go.md
// §4.
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

// decoyStaffUserID acompaña a decoyBarbershopID (service.go) cuando
// ResolveLoginTenant no resolvió tenant: nunca coincide con un
// staff_user_id real (gen_random_uuid() jamás produce el UUID nulo), pero
// permite ejecutar exactamente la misma segunda consulta
// (auth_get_credential) que el camino con tenant resuelto, para igualar la
// forma del trabajo de base de datos entre ambos casos (CA-005-02).
const decoyStaffUserID = "00000000-0000-0000-0000-000000000000"

// Repository implementa auth.Repository contra PostgreSQL real.
type Repository struct {
	db *database.DB
}

// New construye el repositorio. db es el pool tenant-aware ya abierto por
// cmd/api (database.NewDB), conectado como barberia_app.
func New(db *database.DB) *Repository {
	return &Repository{db: db}
}

var _ auth.Repository = (*Repository)(nil)

// ResolveLoginTenant implementa auth.Repository llamando a la función
// SECURITY DEFINER authn_resolve_login_tenant, ANTES de que exista contexto
// de tenant (modelo-fisico-referencia.sql sección A.0), mediante la
// excepción angosta database.DB.ResolveTenant.
func (r *Repository) ResolveLoginTenant(ctx context.Context, email string) (string, bool, error) {
	shop, found, err := r.db.ResolveTenant(ctx,
		"SELECT authn_resolve_login_tenant($1)", email)
	if err != nil {
		return "", false, fmt.Errorf("auth/postgres: resolve login tenant: %w", err)
	}
	return string(shop), found, nil
}

// LookupCredential implementa auth.Repository. Ejecuta, dentro de una única
// transacción tenant-aware (el tenant real o el señuelo, según resolved),
// EXACTAMENTE las mismas dos consultas en ambos casos: buscar el
// staff_user_id activo por correo dentro del tenant, y leer su credencial
// mediante la función estrecha auth_get_credential. Cuando el correo no
// aparece (tenant señuelo, o tenant real pero fila ausente por alguna
// inconsistencia), se sustituye por decoyStaffUserID para conservar la
// misma forma de segunda consulta (CA-005-02): auth_get_credential con un
// id que nunca existe devuelve cero filas, igual de rápido que con uno que
// sí existe.
func (r *Repository) LookupCredential(ctx context.Context, barbershopID string, resolved bool, email string) (auth.Credential, error) {
	var cred auth.Credential

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		staffUserID := decoyStaffUserID
		userFound := false

		if resolved {
			var id string
			err := q.QueryRow(ctx,
				`SELECT id FROM staff_user WHERE email = $1`, email,
			).Scan(&id)
			switch {
			case err == nil:
				staffUserID = id
				userFound = true
			case errors.Is(err, pgx.ErrNoRows):
				// Se conserva el id señuelo para igualar la forma.
			default:
				return fmt.Errorf("select staff_user: %w", err)
			}
		}

		var hash, algorithm string
		err := q.QueryRow(ctx,
			`SELECT password_hash, password_algorithm FROM auth_get_credential($1)`, staffUserID,
		).Scan(&hash, &algorithm)
		switch {
		case err == nil:
			cred = auth.Credential{
				Found:             userFound,
				StaffUserID:       staffUserID,
				PasswordHash:      hash,
				PasswordAlgorithm: algorithm,
			}
		case errors.Is(err, pgx.ErrNoRows):
			cred = auth.Credential{}
		default:
			return fmt.Errorf("auth_get_credential: %w", err)
		}
		return nil
	})
	if err != nil {
		return auth.Credential{}, fmt.Errorf("auth/postgres: lookup credential: %w", err)
	}
	return cred, nil
}

// CreateSession implementa auth.Repository insertando la fila de
// staff_session en su propia transacción tenant-aware. issuedAt y expiresAt
// llegan ya calculados por el servicio (reloj inyectado, no now() de
// PostgreSQL), para que las pruebas puedan congelar el tiempo.
func (r *Repository) CreateSession(ctx context.Context, barbershopID, staffUserID, tokenHash string, issuedAt, expiresAt time.Time) error {
	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		_, err := q.Exec(ctx,
			`INSERT INTO staff_session
			   (barbershop_id, staff_user_id, token_hash, issued_at, expires_at, last_used_at)
			 VALUES ($1, $2, $3, $4, $5, $4)`,
			barbershopID, staffUserID, tokenHash, issuedAt, expiresAt,
		)
		return err
	})
	if err != nil {
		return fmt.Errorf("auth/postgres: create session: %w", err)
	}
	return nil
}

// StaffUserNames implementa auth.Repository.StaffUserNames (HU-064): una
// sola consulta por lote con `= ANY($2)`, tenant-aware por barbershop_id
// además de RLS, que nunca selecciona email ni ninguna otra columna. Un id
// sin coincidencia simplemente está ausente del mapa devuelto.
func (r *Repository) StaffUserNames(ctx context.Context, barbershopID string, staffUserIDs []string) (map[string]string, error) {
	names := make(map[string]string, len(staffUserIDs))
	if len(staffUserIDs) == 0 {
		return names, nil
	}

	shop, err := database.ValidBarbershopID(barbershopID)
	if err != nil {
		return nil, fmt.Errorf("auth/postgres: %w", err)
	}

	err = r.db.InTenantTx(ctx, shop, func(ctx context.Context, q database.Queries) error {
		rows, err := q.Query(ctx,
			`SELECT id, full_name FROM staff_user WHERE barbershop_id = $1 AND id = ANY($2)`,
			barbershopID, staffUserIDs)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var id, name string
			if err := rows.Scan(&id, &name); err != nil {
				return err
			}
			names[id] = name
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("auth/postgres: staff user names: %w", err)
	}
	return names, nil
}
