package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"system-barbershop/internal/modules/shops"
	"system-barbershop/internal/platform/database"
)

// BookingPolicyRepository implementa shops.BookingPolicyRepository sobre
// database.DB. Separado de Repository (repository.go) porque opera con su
// propia disciplina transaccional de concurrencia optimista (HU-093,
// CA-093-02: `SELECT ... FOR UPDATE` + comparación de token antes de
// escribir), aunque ambos lean/escriban la misma tabla `barbershop` -mismo
// criterio de separación por responsabilidad que
// catalog/postgres.AssignmentRepository frente a Repository. Comparte el
// mismo *database.DB.
type BookingPolicyRepository struct {
	db *database.DB
}

// NewBookingPolicyRepository construye el repositorio de la política de
// reserva.
func NewBookingPolicyRepository(db *database.DB) *BookingPolicyRepository {
	return &BookingPolicyRepository{db: db}
}

var _ shops.BookingPolicyRepository = (*BookingPolicyRepository)(nil)

// Get implementa shops.BookingPolicyRepository.Get leyendo los seis campos
// de la política más `updated_at` (para derivar VersionToken) dentro de una
// transacción tenant-aware. RLS (barbershop_select_tenant_policy) ya
// restringe la fila visible a id = current_setting('app.barbershop_id'); el
// filtro explícito WHERE id = $1 es defensa en profundidad, mismo patrón
// que Repository.Get.
func (r *BookingPolicyRepository) Get(ctx context.Context, barbershopID string) (shops.BookingPolicy, bool, error) {
	var (
		p         shops.BookingPolicy
		updatedAt time.Time
		found     bool
	)

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		err := q.QueryRow(ctx,
			`SELECT min_advance_minutes, max_advance_days, slot_grid_minutes,
			        cancellation_deadline_minutes, late_cancellation_client_allowed,
			        late_cancellation_reason_required, updated_at
			   FROM barbershop
			  WHERE id = $1`,
			barbershopID,
		).Scan(
			&p.MinAdvanceMinutes, &p.MaxAdvanceDays, &p.SlotGridMinutes,
			&p.CancellationDeadlineMinutes, &p.LateCancellationClientAllowed,
			&p.LateCancellationReasonRequired, &updatedAt,
		)
		switch {
		case err == nil:
			found = true
		case errors.Is(err, pgx.ErrNoRows):
			// found queda false: defensivo, ver shops.UpdateResult.Found.
		default:
			return fmt.Errorf("select booking policy: %w", err)
		}
		return nil
	})
	if err != nil {
		return shops.BookingPolicy{}, false, fmt.Errorf("shops/postgres: get booking policy: %w", err)
	}
	if found {
		p.VersionToken = shops.EncodeBookingPolicyVersionToken(barbershopID, updatedAt)
	}
	return p, found, nil
}

// Update implementa shops.BookingPolicyRepository.Update dentro de UNA sola
// transacción tenant-aware: bloquea la fila (`SELECT ... FOR UPDATE`),
// deriva el token vigente de su updated_at y lo compara contra
// input.ExpectedVersionToken ANTES de escribir nada
// (BookingPolicyUpdateOutcomeVersionConflict si no coincide); si coincide,
// ejecuta el UPDATE completo de los seis campos y devuelve el token nuevo a
// partir del updated_at que RETURNING trae de la misma sentencia (sin una
// segunda consulta).
func (r *BookingPolicyRepository) Update(ctx context.Context, barbershopID string, input shops.BookingPolicyUpdateInput) (shops.BookingPolicyUpdateResult, error) {
	var result shops.BookingPolicyUpdateResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var currentUpdatedAt time.Time
		if err := q.QueryRow(ctx,
			`SELECT updated_at FROM barbershop WHERE id = $1 FOR UPDATE`,
			barbershopID,
		).Scan(&currentUpdatedAt); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				result.Outcome = shops.BookingPolicyUpdateOutcomeNotFound
				return nil
			}
			return fmt.Errorf("lock barbershop before booking policy update: %w", err)
		}

		if shops.EncodeBookingPolicyVersionToken(barbershopID, currentUpdatedAt) != input.ExpectedVersionToken {
			result.Outcome = shops.BookingPolicyUpdateOutcomeVersionConflict
			return nil
		}

		var (
			p            shops.BookingPolicy
			newUpdatedAt time.Time
		)
		err := q.QueryRow(ctx,
			`UPDATE barbershop
			    SET min_advance_minutes = $2,
			        max_advance_days = $3,
			        slot_grid_minutes = $4,
			        cancellation_deadline_minutes = $5,
			        late_cancellation_client_allowed = $6,
			        late_cancellation_reason_required = $7
			  WHERE id = $1
			RETURNING min_advance_minutes, max_advance_days, slot_grid_minutes,
			          cancellation_deadline_minutes, late_cancellation_client_allowed,
			          late_cancellation_reason_required, updated_at`,
			barbershopID,
			input.MinAdvanceMinutes, input.MaxAdvanceDays, input.SlotGridMinutes,
			input.CancellationDeadlineMinutes, input.LateCancellationClientAllowed,
			input.LateCancellationReasonRequired,
		).Scan(
			&p.MinAdvanceMinutes, &p.MaxAdvanceDays, &p.SlotGridMinutes,
			&p.CancellationDeadlineMinutes, &p.LateCancellationClientAllowed,
			&p.LateCancellationReasonRequired, &newUpdatedAt,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				result.Outcome = shops.BookingPolicyUpdateOutcomeNotFound
				return nil
			}
			return fmt.Errorf("update booking policy: %w", err)
		}

		p.VersionToken = shops.EncodeBookingPolicyVersionToken(barbershopID, newUpdatedAt)
		result.Outcome = shops.BookingPolicyUpdateOutcomeUpdated
		result.Policy = p
		return nil
	})
	if err != nil {
		return shops.BookingPolicyUpdateResult{}, fmt.Errorf("shops/postgres: update booking policy: %w", err)
	}
	return result, nil
}
