// Package postgres implementa los almacenes de notification contra PostgreSQL.
// Las tablas del webhook de WhatsApp no tienen barbershop_id ni se conceden a
// ningún rol: todo pasa por funciones SECURITY DEFINER estrechas (DEC-126).
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"system-barbershop/internal/modules/notification"
	"system-barbershop/internal/platform/database"
)

// WhatsAppEventRepository implementa notification.WhatsAppEventStore con el
// pool de la API (barberia_app).
type WhatsAppEventRepository struct {
	db *database.DB
}

// NewWhatsAppEventRepository construye el repositorio.
func NewWhatsAppEventRepository(db *database.DB) *WhatsAppEventRepository {
	return &WhatsAppEventRepository{db: db}
}

var _ notification.WhatsAppEventStore = (*WhatsAppEventRepository)(nil)

// RecordInbound implementa notification.WhatsAppEventStore.
func (r *WhatsAppEventRepository) RecordInbound(ctx context.Context, phoneHash string, at time.Time) error {
	var one int
	err := r.db.CallSecurityDefinerRow(ctx,
		`SELECT 1 FROM (SELECT whatsapp_window_record($1, $2)) AS recorded`,
		[]any{phoneHash, at},
		func(row pgx.Row) error { return row.Scan(&one) },
	)
	if err != nil {
		return fmt.Errorf("notification/postgres: record inbound message: %w", err)
	}
	return nil
}

// RecordStatus implementa notification.WhatsAppEventStore.
func (r *WhatsAppEventRepository) RecordStatus(ctx context.Context, wamid, status string, at time.Time, errorCode *int) (bool, error) {
	var inserted bool
	err := r.db.CallSecurityDefinerRow(ctx,
		`SELECT whatsapp_status_record($1, $2, $3, $4)`,
		[]any{wamid, status, at, errorCode},
		func(row pgx.Row) error { return row.Scan(&inserted) },
	)
	if err != nil {
		return false, fmt.Errorf("notification/postgres: record message status: %w", err)
	}
	return inserted, nil
}

// ConversationOpen implementa notification.WhatsAppEventStore.
func (r *WhatsAppEventRepository) ConversationOpen(ctx context.Context, phoneHash string) (bool, error) {
	var open bool
	err := r.db.CallSecurityDefinerRow(ctx,
		`SELECT whatsapp_window_is_open($1)`, []any{phoneHash},
		func(row pgx.Row) error { return row.Scan(&open) },
	)
	if err != nil {
		return false, fmt.Errorf("notification/postgres: read conversation window: %w", err)
	}
	return open, nil
}
