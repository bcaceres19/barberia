package publicbooking

import (
	"context"
	"errors"
	"log/slog"
)

// MaxGoogleReviews es el tope de reseñas que se entregan (DEC-129, HU-101).
const MaxGoogleReviews = 5

// GoogleReview es una reseña pública de la ficha de Google del negocio. La
// atribución al autor es obligatoria al mostrarla; nunca se persiste.
type GoogleReview struct {
	AuthorName string
	// AuthorURL es el perfil del autor en Google Maps (https), nil si no hay.
	AuthorURL    *string
	Rating       int
	Text         *string
	RelativeTime *string
}

// GoogleReviews es la calificación y las reseñas de la ficha de Google. El valor
// cero es la estructura vacía explícita: «no hay nada que mostrar» (CA-101-01,
// CA-101-03), que la reserva pública traduce en ocultar la sección.
type GoogleReviews struct {
	// Rating es el promedio de Google (1–5), nil sin datos.
	Rating      *float64
	RatingCount int
	// MapsURL es la ficha real en Google Maps (https), destino de la atribución.
	MapsURL *string
	Reviews []GoogleReview
}

// Empty informa si no hay nada que mostrar.
func (g GoogleReviews) Empty() bool {
	return g.Rating == nil && len(g.Reviews) == 0
}

// GoogleReviewsProvider es el puerto hacia Google Places (DEC-129). El núcleo no
// importa ningún cliente de Google: lo implementa publicbooking/google y las
// pruebas usan un doble. Cualquier error significa «no se pudo consultar» y la
// reserva pública no lo ve (CA-101-03).
type GoogleReviewsProvider interface {
	// Fetch devuelve la calificación y hasta MaxGoogleReviews reseñas de la
	// ficha placeID (ya validado en su forma). Las implementaciones cachean.
	Fetch(ctx context.Context, placeID string) (GoogleReviews, error)
}

// ErrGoogleReviewsUnavailable indica que Google no pudo consultarse o no
// reconoce la ficha (red, cuota agotada, credencial, identificador inexistente).
var ErrGoogleReviewsUnavailable = errors.New("publicbooking: reseñas de Google no disponibles")

// BarbershopResolver resuelve el slug público a su perfil: lo satisface *Service.
type BarbershopResolver interface {
	ResolveBarbershop(ctx context.Context, rawSlug string) (BarbershopProfile, error)
}

// GoogleReviewsService entrega las reseñas de Google de la barbería de un
// enlace público (HU-101). La resolución del slug es la misma de HU-090, con su
// 404 uniforme; todo lo demás se degrada en silencio a la estructura vacía.
type GoogleReviewsService struct {
	resolver BarbershopResolver
	provider GoogleReviewsProvider
	logger   *slog.Logger
}

// NewGoogleReviewsService construye el servicio. provider nil desactiva la
// integración por completo (sin credencial de Google): siempre responde vacío.
func NewGoogleReviewsService(resolver BarbershopResolver, provider GoogleReviewsProvider, logger *slog.Logger) *GoogleReviewsService {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &GoogleReviewsService{resolver: resolver, provider: provider, logger: logger}
}

// Get resuelve el slug y devuelve las reseñas de Google de su barbería. Solo un
// slug inválido o no publicable es un error (el 404 uniforme de HU-090): sin
// ficha, sin credencial o ante una consulta fallida devuelve la estructura vacía
// y registra la causa solo en el servidor (CA-101-01, CA-101-03).
func (s *GoogleReviewsService) Get(ctx context.Context, rawSlug string) (GoogleReviews, error) {
	profile, err := s.resolver.ResolveBarbershop(ctx, rawSlug)
	if err != nil {
		return GoogleReviews{}, err
	}
	if s.provider == nil || profile.GooglePlaceID == nil {
		return GoogleReviews{}, nil
	}
	reviews, err := s.provider.Fetch(ctx, *profile.GooglePlaceID)
	if err != nil {
		s.logger.WarnContext(ctx, "publicbooking.google_reviews.fallo", "error", err)
		return GoogleReviews{}, nil
	}
	return reviews, nil
}
