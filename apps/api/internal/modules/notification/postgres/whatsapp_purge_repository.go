package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"system-barbershop/internal/modules/notification"
	"system-barbershop/internal/platform/database"
)

// WhatsAppPurgeRepository implementa notification.WhatsAppPurgeStore. La
// función está concedida exclusivamente a barberia_worker (DEC-040): db debe
// ser el pool del proceso worker, nunca el de la API.
type WhatsAppPurgeRepository struct {
	db *database.DB
}

// NewWhatsAppPurgeRepository construye el repositorio.
func NewWhatsAppPurgeRepository(db *database.DB) *WhatsAppPurgeRepository {
	return &WhatsAppPurgeRepository{db: db}
}

var _ notification.WhatsAppPurgeStore = (*WhatsAppPurgeRepository)(nil)

// PurgeExpired implementa notification.WhatsAppPurgeStore.
func (r *WhatsAppPurgeRepository) PurgeExpired(ctx context.Context, limit, statusRetentionSeconds int) (int, error) {
	var deleted int
	err := r.db.CallSecurityDefinerRow(ctx,
		`SELECT whatsapp_webhook_purge_expired($1, $2)`, []any{limit, statusRetentionSeconds},
		func(row pgx.Row) error { return row.Scan(&deleted) },
	)
	if err != nil {
		return 0, fmt.Errorf("notification/postgres: purge whatsapp webhook data: %w", err)
	}
	return deleted, nil
}
