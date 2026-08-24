// Package postgres es el adaptador de persistencia de catalog.Repository
// sobre database.DB (CA-002-06: el núcleo de catalog no importa este
// paquete ni pgx; es este paquete el que importa catalog), mismo patrón que
// staff/postgres.
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"system-barbershop/internal/modules/catalog"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

// createServiceOperation identifica, dentro de una barbería, la operación
// de idempotencia del alta de servicios (RN-IDE-01): una clave ya usada
// para OTRA operación nunca se confunde con esta.
const createServiceOperation idempotency.Operation = "create_service"

// createServiceIdempotencyTTL es la vigencia de una reclamación de
// idempotencia sobre el alta de un servicio antes de tratarse como si nunca
// hubiera existido (CA-004-05), mismo valor que staff (24 horas cubre con
// margen cualquier reintento de red razonable).
const createServiceIdempotencyTTL = 24 * time.Hour

// nameUniqueConstraint es el nombre exacto del índice único parcial que
// DEC-067 exige (database/migrations/20260824000000_create_service.sql):
// solo un unique_violation sobre ESTE constraint se traduce a
// CreateResult.NameTaken/UpdateResult.NameTaken; cualquier otro
// unique_violation (que hoy no debería existir en `service`) se propaga
// como error interno en vez de camuflarse como conflicto de nombre.
const nameUniqueConstraint = "idx_service_active_name"

// errNameConflictInternal es un error interno (sentinela) que solo existe
// para atravesar InTenantTx y provocar su ROLLBACK normal cuando el INSERT
// o el UPDATE chocan con nameUniqueConstraint. Nunca se expone fuera de
// este paquete: Create/Update lo detectan con errors.Is y lo traducen al
// flag NameTaken en el resultado, no en el error devuelto.
var errNameConflictInternal = errors.New("catalog/postgres: nombre de servicio en conflicto")

// Repository implementa catalog.Repository sobre database.DB.
type Repository struct {
	db    *database.DB
	coord idempotency.Coordinator
}

// New construye el repositorio con el coordinador de idempotencia real
// (SQLCoordinator), mismo criterio que staffpostgres.New.
func New(db *database.DB, coord idempotency.Coordinator) *Repository {
	return &Repository{db: db, coord: coord}
}

var _ catalog.Repository = (*Repository)(nil)

// serviceResponseWire es la forma JSON EXACTA que también usa
// httpapi.ServiceResponse (mismos nombres de campo, mismo orden, mismos
// tipos): la respuesta que Complete persiste para una repetición exacta
// (CA-004-01) debe ser BYTE A BYTE idéntica a la que el handler ya
// construyó para la primera ejecución, y ambas deben tener EXACTAMENTE la
// misma forma que un GET/PATCH de este mismo servicio. Si cambia
// httpapi.ServiceResponse, este struct debe cambiar igual en el mismo
// commit (ver TestCreate_StoredResponseBody_MatchesHTTPAPIWireShape).
type serviceResponseWire struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     *string   `json:"description"`
	DurationMinutes int       `json:"durationMinutes"`
	Price           string    `json:"price"`
	Currency        string    `json:"currency"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func marshalServiceResponse(svc catalog.Service) ([]byte, error) {
	return json.Marshal(serviceResponseWire{
		ID:              svc.ID,
		Name:            svc.Name,
		Description:     svc.Description,
		DurationMinutes: svc.DurationMinutes,
		Price:           catalog.FormatPriceCOP(svc.PriceCents),
		Currency:        svc.Currency,
		CreatedAt:       svc.CreatedAt,
		UpdatedAt:       svc.UpdatedAt,
	})
}

// numericFromCents construye el valor pgtype.Numeric que representa cents
// centavos como cantidad exacta en pesos (dos decimales), sin pasar nunca
// por coma flotante: Int = cents, Exp = -2 (docs/05-backend/
// estandar-base-datos.md §5).
func numericFromCents(cents int64) pgtype.Numeric {
	return pgtype.Numeric{Int: big.NewInt(cents), Exp: -2, Valid: true}
}

// numericFromCentsPtr construye el mismo valor que numericFromCents, o un
// pgtype.Numeric inválido (NULL en SQL) cuando cents es nil: usado en
// UPDATE para que COALESCE($n, price_amount) conserve el valor existente
// cuando el cliente no envió `price` en el PATCH.
func numericFromCentsPtr(cents *int64) pgtype.Numeric {
	if cents == nil {
		return pgtype.Numeric{}
	}
	return numericFromCents(*cents)
}

// centsFromNumeric extrae el número exacto de centavos representado por n
// (n.Int * 10^n.Exp pesos), sin pasar por coma flotante. numeric(12,2)
// siempre devuelve n.Exp = -2, pero la conversión es correcta para
// cualquier escala no positiva.
func centsFromNumeric(n pgtype.Numeric) (int64, error) {
	if !n.Valid || n.NaN || n.Int == nil {
		return 0, fmt.Errorf("catalog/postgres: price_amount nulo, NaN o sin valor")
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
		return 0, fmt.Errorf("catalog/postgres: price_amount fuera de rango representable")
	}
	return result.Int64(), nil
}

// isNameConflict informa si err es el unique_violation (23505) de
// nameUniqueConstraint (DEC-067). Cualquier otro error (incluida cualquier
// otra violación de restricción) devuelve false: solo este constraint
// concreto se traduce a un conflicto de negocio de nombre.
func isNameConflict(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23505" && pgErr.ConstraintName == nameUniqueConstraint
}

// rowScanner es el subconjunto de pgx.Row/pgx.Rows que scanServiceWithID
// necesita, para no acoplar la firma a un tipo concreto de pgx.
type rowScanner interface {
	Scan(dest ...any) error
}

// scanServiceWithID lee id, name, description, duration_minutes,
// price_amount, price_currency, created_at, updated_at (en ese orden, el
// mismo que List/Get/Create/Update proyectan) hacia un catalog.Service
// nuevo, convirtiendo price_amount de pgtype.Numeric a centavos exactos sin
// pasar por coma flotante.
func scanServiceWithID(row rowScanner) (catalog.Service, error) {
	var svc catalog.Service
	var priceNumeric pgtype.Numeric
	if err := row.Scan(
		&svc.ID, &svc.Name, &svc.Description, &svc.DurationMinutes, &priceNumeric, &svc.Currency,
		&svc.CreatedAt, &svc.UpdatedAt,
	); err != nil {
		return catalog.Service{}, err
	}
	cents, err := centsFromNumeric(priceNumeric)
	if err != nil {
		return catalog.Service{}, err
	}
	svc.PriceCents = cents
	return svc, nil
}

// List implementa catalog.Repository.List: orden estable (created_at, id)
// dentro del tenant vigente, cursor opaco decodificado por el núcleo
// (catalog.Cursor), paginación "pedir uno de más" para saber si hay página
// siguiente sin una segunda consulta COUNT (mismo patrón que
// staff/postgres.Repository.List).
func (r *Repository) List(ctx context.Context, barbershopID string, cursor *catalog.Cursor, limit int) (catalog.ListResult, error) {
	var result catalog.ListResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var (
			rows pgx.Rows
			err  error
		)
		fetchLimit := limit + 1

		if cursor == nil {
			rows, err = q.Query(ctx,
				`SELECT id, name, description, duration_minutes, price_amount, price_currency,
				        created_at, updated_at
				   FROM service
				  WHERE barbershop_id = $1
				  ORDER BY created_at, id
				  LIMIT $2`,
				barbershopID, fetchLimit,
			)
		} else {
			rows, err = q.Query(ctx,
				`SELECT id, name, description, duration_minutes, price_amount, price_currency,
				        created_at, updated_at
				   FROM service
				  WHERE barbershop_id = $1
				    AND (created_at, id) > ($2, $3)
				  ORDER BY created_at, id
				  LIMIT $4`,
				barbershopID, cursor.CreatedAt, cursor.ID, fetchLimit,
			)
		}
		if err != nil {
			return fmt.Errorf("list services: query: %w", err)
		}
		defer rows.Close()

		items := make([]catalog.Service, 0, fetchLimit)
		for rows.Next() {
			svc, err := scanServiceWithID(rows)
			if err != nil {
				return fmt.Errorf("list services: scan: %w", err)
			}
			items = append(items, svc)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("list services: rows: %w", err)
		}

		hasMore := len(items) > limit
		if hasMore {
			items = items[:limit]
		}

		result.Items = items
		if hasMore {
			last := items[len(items)-1]
			result.NextCursor = catalog.EncodeCursor(catalog.Cursor{CreatedAt: last.CreatedAt, ID: last.ID})
		}
		return nil
	})
	if err != nil {
		return catalog.ListResult{}, fmt.Errorf("catalog/postgres: list services: %w", err)
	}
	return result, nil
}

// Get implementa catalog.Repository.Get. RLS (service_select_tenant_policy)
// ya restringe la fila visible a barbershop_id = current_setting(...); el
// filtro explícito WHERE barbershop_id = $2 es defensa en profundidad,
// mismo patrón que staff/postgres.Repository.Get.
func (r *Repository) Get(ctx context.Context, barbershopID, serviceID string) (catalog.Service, bool, error) {
	var svc catalog.Service
	found := false

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		svc.ID = serviceID
		var priceNumeric pgtype.Numeric
		err := q.QueryRow(ctx,
			`SELECT name, description, duration_minutes, price_amount, price_currency,
			        created_at, updated_at
			   FROM service
			  WHERE id = $1 AND barbershop_id = $2`,
			serviceID, barbershopID,
		).Scan(&svc.Name, &svc.Description, &svc.DurationMinutes, &priceNumeric, &svc.Currency,
			&svc.CreatedAt, &svc.UpdatedAt)
		switch {
		case err == nil:
			cents, convErr := centsFromNumeric(priceNumeric)
			if convErr != nil {
				return convErr
			}
			svc.PriceCents = cents
			found = true
		case errors.Is(err, pgx.ErrNoRows):
			// found queda false: cubre tanto "no existe" como "es de otra
			// barbería" (CA-022-06), sin distinguir la causa.
		default:
			return fmt.Errorf("select service: %w", err)
		}
		return nil
	})
	if err != nil {
		return catalog.Service{}, false, fmt.Errorf("catalog/postgres: get service: %w", err)
	}
	if !found {
		return catalog.Service{}, false, nil
	}
	return svc, true, nil
}

// Create implementa catalog.Repository.Create: Begin, INSERT y Complete
// ocurren dentro de la MISMA InTenantTx (apps/api/README.md, "Patrón
// obligatorio: idempotencia reutilizable"). Un unique_violation de
// nameUniqueConstraint durante el INSERT hace ROLLBACK de toda la
// transacción -incluida la reclamación de idempotencia que Begin ya había
// tomado- y se traduce a CreateResult.NameTaken, dejando la clave libre
// para un reintento legítimo con un nombre distinto (mismo criterio que
// CA-004-06).
func (r *Repository) Create(
	ctx context.Context,
	barbershopID string,
	input catalog.CreateInput,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (catalog.CreateResult, error) {
	var result catalog.CreateResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		decision, err := r.coord.Begin(ctx, q, database.BarbershopID(barbershopID), key, createServiceOperation, fingerprint, createServiceIdempotencyTTL)
		if err != nil {
			return fmt.Errorf("idempotency begin: %w", err)
		}
		result.Decision = decision

		if decision.Outcome != idempotency.OutcomeProceed {
			if decision.Outcome == idempotency.OutcomeReplay {
				result.Response = decision.Response
			}
			return nil
		}

		row := q.QueryRow(ctx,
			`INSERT INTO service (barbershop_id, name, description, duration_minutes, price_amount, price_currency)
			      VALUES ($1, $2, $3, $4, $5, 'COP')
			   RETURNING id, name, description, duration_minutes, price_amount, price_currency,
			             created_at, updated_at`,
			barbershopID, input.Name, input.Description, input.DurationMinutes, numericFromCents(input.PriceCents),
		)
		svc, err := scanServiceWithID(row)
		if err != nil {
			if isNameConflict(err) {
				result.NameTaken = true
				return errNameConflictInternal
			}
			return fmt.Errorf("insert service: %w", err)
		}

		body, err := marshalServiceResponse(svc)
		if err != nil {
			return fmt.Errorf("marshal created service: %w", err)
		}
		stored := idempotency.StoredResponse{
			Status:      201,
			ContentType: "application/json",
			Body:        string(body),
		}

		ok, err := r.coord.Complete(ctx, q, database.BarbershopID(barbershopID), key, stored)
		if err != nil {
			return fmt.Errorf("idempotency complete: %w", err)
		}
		if !ok {
			return fmt.Errorf("idempotency complete: la reclamación ya no estaba in_progress")
		}

		result.Service = svc
		result.Response = stored
		return nil
	})
	if err != nil {
		if errors.Is(err, errNameConflictInternal) {
			return result, nil
		}
		return catalog.CreateResult{}, fmt.Errorf("catalog/postgres: create service: %w", err)
	}
	return result, nil
}

// Update implementa catalog.Repository.Update: un único
// `UPDATE ... RETURNING`, filtrando por id Y barbershop_id (defensa en
// profundidad sobre service_update_tenant_policy). Cada campo ausente en
// fields se resuelve con COALESCE hacia el valor actual, salvo
// description, que usa una bandera explícita (para poder "borrar" con
// NULL, distinto de "no tocar"). Un unique_violation de
// nameUniqueConstraint se traduce a UpdateResult.NameTaken; cero filas
// afectadas cubre tanto "no existe" como "es de otra barbería"
// (CA-022-06).
func (r *Repository) Update(ctx context.Context, barbershopID, serviceID string, fields catalog.UpdateFields) (catalog.UpdateResult, error) {
	var result catalog.UpdateResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var svc catalog.Service
		svc.ID = serviceID
		var priceNumeric pgtype.Numeric

		err := q.QueryRow(ctx,
			`UPDATE service
			    SET name             = COALESCE($3::text, name),
			        description      = CASE WHEN $4::boolean THEN $5::text ELSE description END,
			        duration_minutes = COALESCE($6::integer, duration_minutes),
			        price_amount     = COALESCE($7::numeric(12,2), price_amount)
			  WHERE id = $1 AND barbershop_id = $2
			RETURNING name, description, duration_minutes, price_amount, price_currency,
			          created_at, updated_at`,
			serviceID, barbershopID,
			fields.Name,
			fields.Description.Set, fields.Description.Value,
			fields.DurationMinutes,
			numericFromCentsPtr(fields.PriceCents),
		).Scan(&svc.Name, &svc.Description, &svc.DurationMinutes, &priceNumeric, &svc.Currency,
			&svc.CreatedAt, &svc.UpdatedAt)
		switch {
		case err == nil:
			cents, convErr := centsFromNumeric(priceNumeric)
			if convErr != nil {
				return convErr
			}
			svc.PriceCents = cents
			result.Found = true
			result.Service = svc
		case errors.Is(err, pgx.ErrNoRows):
			// result.Found queda false: defensivo, ver catalog.UpdateResult.Found.
		case isNameConflict(err):
			result.NameTaken = true
			return errNameConflictInternal
		default:
			return fmt.Errorf("update service: %w", err)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, errNameConflictInternal) {
			return result, nil
		}
		return catalog.UpdateResult{}, fmt.Errorf("catalog/postgres: update service: %w", err)
	}
	return result, nil
}
