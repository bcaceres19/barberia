package notification

import (
	"context"
	"time"
)

// whatsAppStatusRetention es cuánto se conserva un estado de entrega (DEC-126):
// basta para investigar un fallo reciente sin acumular una tabla sin límite.
const whatsAppStatusRetention = 30 * 24 * time.Hour

// WhatsAppPurgeStore borra los datos vencidos del webhook.
type WhatsAppPurgeStore interface {
	PurgeExpired(ctx context.Context, limit, statusRetentionSeconds int) (int, error)
}

// WhatsAppPurgeService purga ventanas vencidas y estados antiguos por lotes.
type WhatsAppPurgeService struct {
	store WhatsAppPurgeStore
	limit int
}

// NewWhatsAppPurgeService construye el servicio; limit acota cada lote.
func NewWhatsAppPurgeService(store WhatsAppPurgeStore, limit int) *WhatsAppPurgeService {
	return &WhatsAppPurgeService{store: store, limit: limit}
}

// PurgeOnce ejecuta un lote y devuelve cuántas filas borró.
func (s *WhatsAppPurgeService) PurgeOnce(ctx context.Context) (int, error) {
	return s.store.PurgeExpired(ctx, s.limit, int(whatsAppStatusRetention.Seconds()))
}
