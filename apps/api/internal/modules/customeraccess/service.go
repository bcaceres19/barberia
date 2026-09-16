package customeraccess

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"system-barbershop/internal/platform/apperr"
)

// Service orquesta HU-098: valida la forma del token recibido, lo hashea y
// delega en Repository la resolución tenant-aware. httpapi nunca calcula el
// hash por su cuenta: solo este paquete toca el valor en claro antes de
// descartarlo (CA-098-02, el valor en claro nunca se registra ni persiste).
type Service struct {
	repo Repository
}

// NewService construye el servicio de acceso del cliente.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetAppointmentByToken resuelve el turno correspondiente al token en claro
// de la URL (CA-098-01). Un token vacío, demasiado largo, inexistente,
// vencido o revocado produce EXACTAMENTE el mismo error NotFound uniforme
// (CA-098-02, RN-TEN-01): la causa nunca se distingue, mismo criterio que
// publicbooking.errBarbershopNotPublic frente a un slug inválido.
func (s *Service) GetAppointmentByToken(ctx context.Context, tokenRaw string) (AppointmentView, error) {
	if err := ctx.Err(); err != nil {
		return AppointmentView{}, apperr.Internal(fmt.Errorf("customeraccess: contexto cancelado antes de resolver token: %w", err))
	}

	token := strings.TrimSpace(tokenRaw)
	if token == "" || len(token) > MaxTokenLength {
		return AppointmentView{}, errAppointmentAccessDenied()
	}

	sum := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(sum[:])

	view, found, err := s.repo.ResolveAppointmentByTokenHash(ctx, tokenHash)
	if err != nil {
		return AppointmentView{}, apperr.Internal(fmt.Errorf("customeraccess: resolver turno por token: %w", err))
	}
	if !found {
		return AppointmentView{}, errAppointmentAccessDenied()
	}
	return view, nil
}

// errAppointmentAccessDenied cubre las tres causas de CA-098-02 (token
// inexistente, vencido o revocado) con un único 404 uniforme, mismo
// criterio que publicbooking.errBarbershopNotPublic.
func errAppointmentAccessDenied() error {
	return apperr.NotFound("no existe un turno accesible con ese enlace")
}
