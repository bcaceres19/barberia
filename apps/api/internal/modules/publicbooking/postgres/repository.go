// Package postgres es el adaptador de persistencia de
// publicbooking.Repository sobre database.DB (CA-002-06: el núcleo de
// publicbooking no importa este paquete ni pgx; es este paquete el que
// importa publicbooking).
package postgres

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"system-barbershop/internal/modules/publicbooking"
	"system-barbershop/internal/platform/database"
)

// Repository implementa publicbooking.Repository sobre database.DB.
type Repository struct {
	db *database.DB
}

// New construye el repositorio.
func New(db *database.DB) *Repository {
	return &Repository{db: db}
}

var _ publicbooking.Repository = (*Repository)(nil)
var _ publicbooking.AvailabilityRepository = (*Repository)(nil)

// ResolveBySlug implementa publicbooking.Repository.ResolveBySlug en dos
// pasos, exactamente el patrón ya establecido por
// authn_resolve_login_tenant/authn_resolve_session_tenant (HU-005,
// database.DB.ResolveTenant): primero resuelve slug a un identificador de
// barbería SIN contexto de tenant, mediante la función SECURITY DEFINER
// estrecha public_resolve_barbershop_by_slug (que responde igual -NULL- a
// forma inválida, slug desconocido o barbería no publicable); si resuelve,
// abre una transacción tenant-aware normal (InTenantTx) sobre ESE
// identificador para leer los campos públicos mínimos ya protegidos por
// barbershop_select_tenant_policy, sin necesitar ningún privilegio
// adicional sobre la tabla.
func (r *Repository) ResolveBySlug(ctx context.Context, slug string) (publicbooking.BarbershopProfile, bool, error) {
	barbershopID, found, err := r.db.ResolveTenant(ctx,
		`SELECT public_resolve_barbershop_by_slug($1)`,
		slug,
	)
	if err != nil {
		return publicbooking.BarbershopProfile{}, false, fmt.Errorf("publicbooking/postgres: resolve tenant by slug: %w", err)
	}
	if !found {
		return publicbooking.BarbershopProfile{}, false, nil
	}

	var profile publicbooking.BarbershopProfile
	profileFound := false
	err = r.db.InTenantTx(ctx, barbershopID, func(ctx context.Context, q database.Queries) error {
		err := q.QueryRow(ctx,
			`SELECT name, timezone, contact_email, contact_phone
			   FROM barbershop
			  WHERE id = $1`,
			string(barbershopID),
		).Scan(&profile.Name, &profile.Timezone, &profile.ContactEmail, &profile.ContactPhone)
		switch {
		case err == nil:
			profileFound = true
		case errors.Is(err, pgx.ErrNoRows):
			// profileFound queda false: defensivo (la función SECURITY
			// DEFINER ya confirmó la fila en la misma sentencia SQL que
			// esta transacción re-lee; una desaparición entre ambas solo
			// podría venir de un borrado físico que este esquema no
			// permite para barbershop).
		default:
			return fmt.Errorf("select public barbershop profile: %w", err)
		}
		return nil
	})
	if err != nil {
		return publicbooking.BarbershopProfile{}, false, fmt.Errorf("publicbooking/postgres: read public profile: %w", err)
	}
	return profile, profileFound, nil
}

// ResolveBarbershopID implementa publicbooking.AvailabilityRepository.
// ResolveBarbershopID reutilizando el mismo primer paso de dos que
// ResolveBySlug (public_resolve_barbershop_by_slug vía db.ResolveTenant),
// pero se detiene ahí: HU-094 necesita el identificador para orquestar
// schedule/shops/catalog, no un perfil público.
func (r *Repository) ResolveBarbershopID(ctx context.Context, slug string) (string, bool, error) {
	barbershopID, found, err := r.db.ResolveTenant(ctx,
		`SELECT public_resolve_barbershop_by_slug($1)`,
		slug,
	)
	if err != nil {
		return "", false, fmt.Errorf("publicbooking/postgres: resolve tenant by slug: %w", err)
	}
	if !found {
		return "", false, nil
	}
	return string(barbershopID), true, nil
}

// ListOccupiedIntervals implementa
// publicbooking.AvailabilityRepository.ListOccupiedIntervals: lee, dentro
// de una transacción tenant-aware normal (InTenantTx, RLS ya acota a
// barbershopID), los intervalos de las citas que ocupan agenda
// (`occupies_schedule`, estados-citas.md §11 -mismo criterio derivado que
// la restricción de exclusión `appointment_barber_interval_excl`, para que
// esta lectura nunca pueda divergir de ella) cuyo intervalo interseca
// [from, to). barberID se filtra explícitamente además de RLS (mismo
// criterio que las demás consultas de este paquete).
func (r *Repository) ListOccupiedIntervals(ctx context.Context, barbershopID, barberID string, from, to time.Time) ([]time.Time, []time.Time, error) {
	var starts, ends []time.Time
	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		rows, err := q.Query(ctx, `
			SELECT starts_at, ends_at
			  FROM appointment
			 WHERE barbershop_id = $1
			   AND barber_id = $2
			   AND occupies_schedule
			   AND starts_at < $4
			   AND ends_at > $3
			 ORDER BY starts_at`,
			barbershopID, barberID, from, to,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var s, e time.Time
			if err := rows.Scan(&s, &e); err != nil {
				return err
			}
			starts = append(starts, s)
			ends = append(ends, e)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, nil, fmt.Errorf("publicbooking/postgres: listar ocupación de agenda: %w", err)
	}
	return starts, ends, nil
}

// centsFromNumeric convierte price_amount (pgtype.Numeric) a centavos
// exactos, sin pasar nunca por coma flotante. Duplica
// catalog/postgres.centsFromNumeric a propósito: el núcleo de publicbooking
// no importa catalog (CA-002-06, docs/03-desarrollo/estandar-backend-go.md
// §5.23, mismo criterio ya aplicado por publicbooking.ServiceCursor frente
// a catalog.Cursor).
func centsFromNumeric(n pgtype.Numeric) (int64, error) {
	if !n.Valid || n.NaN || n.Int == nil {
		return 0, fmt.Errorf("publicbooking/postgres: price_amount nulo, NaN o sin valor")
	}
	exp := n.Exp + 2
	result := new(big.Int).Set(n.Int)
	switch {
	case exp == 0:
		// ya está en centavos.
	case exp > 0:
		scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(exp)), nil)
		result.Mul(result, scale)
	default:
		scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-exp)), nil)
		result.Quo(result, scale)
	}
	if !result.IsInt64() {
		return 0, fmt.Errorf("publicbooking/postgres: price_amount fuera de rango representable")
	}
	return result.Int64(), nil
}

// ListPublicServices implementa publicbooking.Repository.ListPublicServices
// con el mismo patrón de dos pasos que ResolveBySlug: resuelve slug SIN
// contexto de tenant mediante public_resolve_barbershop_by_slug y, si
// resuelve, abre una transacción tenant-aware (InTenantTx) para leer la
// página de servicios activos con al menos una asignación vigente
// (HU-091, CA-091-01). El filtro EXISTS sobre barber_service ya queda
// acotado al mismo tenant por barbershop_service_select_tenant_policy
// (RLS), pero se repite explícitamente por bs.barbershop_id = s.barbershop_id
// siguiendo el mismo criterio que catalog/postgres.Repository.List
// ("cada consulta filtra explícitamente por barbershopID además de RLS").
// Paginación "pedir uno de más" (mismo patrón que catalog.Repository.List).
func (r *Repository) ListPublicServices(ctx context.Context, slug string, cursor *publicbooking.ServiceCursor, limit int) (publicbooking.PublicServiceListResult, bool, error) {
	barbershopID, found, err := r.db.ResolveTenant(ctx,
		`SELECT public_resolve_barbershop_by_slug($1)`,
		slug,
	)
	if err != nil {
		return publicbooking.PublicServiceListResult{}, false, fmt.Errorf("publicbooking/postgres: resolve tenant by slug: %w", err)
	}
	if !found {
		return publicbooking.PublicServiceListResult{}, false, nil
	}

	const baseQuery = `
		SELECT s.id, s.name, s.description, s.duration_minutes, s.price_amount, s.price_currency,
		       s.created_at
		  FROM service s
		 WHERE s.barbershop_id = $1
		   AND s.is_active = true
		   AND EXISTS (
		       SELECT 1 FROM barber_service bs
		        WHERE bs.barbershop_id = s.barbershop_id AND bs.service_id = s.id
		   )`

	var result publicbooking.PublicServiceListResult
	err = r.db.InTenantTx(ctx, barbershopID, func(ctx context.Context, q database.Queries) error {
		var (
			rows pgx.Rows
			err  error
		)
		fetchLimit := limit + 1

		if cursor == nil {
			rows, err = q.Query(ctx,
				baseQuery+`
			 ORDER BY s.created_at, s.id
			 LIMIT $2`,
				string(barbershopID), fetchLimit,
			)
		} else {
			rows, err = q.Query(ctx,
				baseQuery+`
			   AND (s.created_at, s.id) > ($2, $3)
			 ORDER BY s.created_at, s.id
			 LIMIT $4`,
				string(barbershopID), cursor.CreatedAt, cursor.ID, fetchLimit,
			)
		}
		if err != nil {
			return fmt.Errorf("list public services: query: %w", err)
		}
		defer rows.Close()

		items := make([]publicbooking.PublicService, 0, fetchLimit)
		createdAts := make([]time.Time, 0, fetchLimit)
		for rows.Next() {
			var (
				svc          publicbooking.PublicService
				priceNumeric pgtype.Numeric
				createdAt    time.Time
			)
			if err := rows.Scan(&svc.ID, &svc.Name, &svc.Description, &svc.DurationMinutes, &priceNumeric, &svc.Currency, &createdAt); err != nil {
				return fmt.Errorf("list public services: scan: %w", err)
			}
			cents, err := centsFromNumeric(priceNumeric)
			if err != nil {
				return fmt.Errorf("list public services: %w", err)
			}
			svc.PriceCents = cents
			items = append(items, svc)
			createdAts = append(createdAts, createdAt)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("list public services: rows: %w", err)
		}

		hasMore := len(items) > limit
		if hasMore {
			items = items[:limit]
			createdAts = createdAts[:limit]
		}

		result.Items = items
		if hasMore {
			last := len(items) - 1
			result.NextCursor = publicbooking.EncodeServiceCursor(publicbooking.ServiceCursor{
				CreatedAt: createdAts[last],
				ID:        items[last].ID,
			})
		}
		return nil
	})
	if err != nil {
		return publicbooking.PublicServiceListResult{}, false, fmt.Errorf("publicbooking/postgres: read public services: %w", err)
	}
	return result, true, nil
}

// ListPublicBarbers implementa publicbooking.Repository.ListPublicBarbers
// con el mismo patrón de dos pasos que ListPublicServices: resuelve slug SIN
// contexto de tenant mediante public_resolve_barbershop_by_slug y, si
// resuelve, abre una transacción tenant-aware (InTenantTx) para leer los
// barberos con asignación vigente al servicio activo serviceID (HU-092,
// CA-092-02). serviceID sin forma de UUID se descarta ANTES de construir la
// consulta (evita un error de tipo de PostgreSQL sobre una columna `uuid`),
// pero SOLO después de resolver la barbería: found sigue reflejando
// exclusivamente si la barbería existe, nunca si serviceID es válido
// (CA-092-03: un serviceID inválido, ajeno, inexistente o de un servicio
// inactivo produce la MISMA lista vacía que uno sin barberos asignados).
func (r *Repository) ListPublicBarbers(ctx context.Context, slug string, serviceID string) (publicbooking.PublicBarberListResult, bool, error) {
	barbershopID, found, err := r.db.ResolveTenant(ctx,
		`SELECT public_resolve_barbershop_by_slug($1)`,
		slug,
	)
	if err != nil {
		return publicbooking.PublicBarberListResult{}, false, fmt.Errorf("publicbooking/postgres: resolve tenant by slug: %w", err)
	}
	if !found {
		return publicbooking.PublicBarberListResult{}, false, nil
	}

	var result publicbooking.PublicBarberListResult
	if !publicbooking.LooksLikePublicServiceID(serviceID) {
		return result, true, nil
	}

	err = r.db.InTenantTx(ctx, barbershopID, func(ctx context.Context, q database.Queries) error {
		rows, err := q.Query(ctx,
			`SELECT b.id, b.full_name
			   FROM barber b
			   JOIN barber_service bs
			     ON bs.barbershop_id = b.barbershop_id AND bs.barber_id = b.id
			   JOIN service s
			     ON s.barbershop_id = bs.barbershop_id AND s.id = bs.service_id
			  WHERE b.barbershop_id = $1
			    AND bs.service_id = $2
			    AND s.is_active = true
			  ORDER BY bs.created_at, b.id`,
			string(barbershopID), serviceID,
		)
		if err != nil {
			return fmt.Errorf("list public barbers: query: %w", err)
		}
		defer rows.Close()

		items := make([]publicbooking.PublicBarber, 0)
		for rows.Next() {
			var barber publicbooking.PublicBarber
			if err := rows.Scan(&barber.ID, &barber.FullName); err != nil {
				return fmt.Errorf("list public barbers: scan: %w", err)
			}
			items = append(items, barber)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("list public barbers: rows: %w", err)
		}

		result.Items = items
		return nil
	})
	if err != nil {
		return publicbooking.PublicBarberListResult{}, false, fmt.Errorf("publicbooking/postgres: read public barbers: %w", err)
	}
	return result, true, nil
}
