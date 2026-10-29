package reviews

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"system-barbershop/internal/platform/apperr"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func looksLikeUUID(s string) bool { return uuidPattern.MatchString(s) }

// Service orquesta HU-100: valida lo que el cliente envía, resuelve el turno por
// el token y delega la escritura atómica en Repository.
type Service struct {
	repo Repository
}

// NewService construye el servicio de reseñas.
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// Create guarda la reseña del turno del token (CA-100-01 a CA-100-03). Un token
// inexistente, vencido o revocado responde 404 uniforme; un turno no completado,
// cancelado o ya calificado responde 409 uniforme, sin distinguir la causa. La
// validación de la calificación y del comentario va primero: un 422 no revela
// nada del turno.
func (s *Service) Create(ctx context.Context, tokenRaw string, rating int, comment *string) (Review, error) {
	if !ValidRating(rating) {
		return Review{}, apperr.Validation("la calificación debe ser un número entero de 1 a 5")
	}
	normalized, ok := NormalizeComment(comment)
	if !ok {
		return Review{}, apperr.Validation(fmt.Sprintf("el comentario no puede superar %d caracteres", MaxCommentLength))
	}

	token := strings.TrimSpace(tokenRaw)
	if token == "" || len(token) > MaxTokenLength {
		return Review{}, errTokenNotFound()
	}
	sum := sha256.Sum256([]byte(token))

	review, outcome, err := s.repo.CreateByTokenHash(ctx, hex.EncodeToString(sum[:]), rating, normalized)
	if err != nil {
		return Review{}, apperr.Internal(fmt.Errorf("reviews: guardar reseña: %w", err))
	}
	switch outcome {
	case Created:
		return review, nil
	case TokenNotFound:
		return Review{}, errTokenNotFound()
	default:
		return Review{}, apperr.Conflict("este turno no admite una reseña")
	}
}

// ListPublic devuelve las reseñas visibles de un barbero para el visitante de la
// reserva pública (CA-100-04).
func (s *Service) ListPublic(ctx context.Context, slug, barberID string) (PublicList, error) {
	if !looksLikeUUID(barberID) {
		return PublicList{}, errBarberNotFound()
	}
	list, found, err := s.repo.ListPublic(ctx, slug, barberID, PublicListLimit)
	if err != nil {
		return PublicList{}, apperr.Internal(fmt.Errorf("reviews: listar reseñas públicas: %w", err))
	}
	if !found {
		return PublicList{}, errBarberNotFound()
	}
	return list, nil
}

// ListForBarber devuelve a quien modera todas las reseñas del barbero, ocultas
// incluidas (CA-100-05).
func (s *Service) ListForBarber(ctx context.Context, barbershopID, barberID string) (PanelList, error) {
	if !looksLikeUUID(barberID) {
		return PanelList{}, errBarberNotFound()
	}
	list, found, err := s.repo.ListForBarber(ctx, barbershopID, barberID, PanelListLimit)
	if err != nil {
		return PanelList{}, apperr.Internal(fmt.Errorf("reviews: listar reseñas del panel: %w", err))
	}
	if !found {
		return PanelList{}, errBarberNotFound()
	}
	return list, nil
}

// SetHidden oculta o vuelve a mostrar una reseña (CA-100-05). Es idempotente por
// naturaleza: repetir la misma orden deja el mismo estado.
func (s *Service) SetHidden(ctx context.Context, barbershopID, barberID, reviewID string, hidden bool) (Review, error) {
	if !looksLikeUUID(barberID) || !looksLikeUUID(reviewID) {
		return Review{}, errReviewNotFound()
	}
	review, found, err := s.repo.SetHidden(ctx, barbershopID, barberID, reviewID, hidden)
	if err != nil {
		return Review{}, apperr.Internal(fmt.Errorf("reviews: ocultar reseña: %w", err))
	}
	if !found {
		return Review{}, errReviewNotFound()
	}
	return review, nil
}

func errTokenNotFound() error {
	return apperr.NotFound("no existe un turno accesible con ese enlace")
}

func errBarberNotFound() error { return apperr.NotFound("no existe ese barbero") }

func errReviewNotFound() error { return apperr.NotFound("no existe esa reseña") }
