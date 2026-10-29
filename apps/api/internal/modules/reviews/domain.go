package reviews

import (
	"strings"
	"time"
	"unicode/utf8"
)

const (
	// MinRating y MaxRating acotan la calificación entera (barber_review_rating_ck).
	MinRating = 1
	MaxRating = 5
	// MaxCommentLength coincide con barber_review_comment_ck.
	MaxCommentLength = 500
	// MaxTokenLength acota el token de la ruta antes de hashearlo: el real mide
	// 43 caracteres (DEC-089); uno más largo nunca coincide con ninguna fila.
	MaxTokenLength = 256
	// PublicListLimit es el máximo de reseñas que devuelve la lista pública: las
	// más recientes, sin paginación (la selección de barbero no es un archivo).
	PublicListLimit = 20
	// PanelListLimit es el máximo que ve quien modera en el panel.
	PanelListLimit = 100
)

// Review es una reseña completa, tal como la ve quien modera en el panel (con
// su fecha de ocultamiento si la tiene). Nunca guarda datos del cliente.
type Review struct {
	ID        string
	BarberID  string
	Rating    int
	Comment   *string
	HiddenAt  *time.Time
	CreatedAt time.Time
}

// PublicReview es lo que ve cualquier visitante: calificación, comentario y
// fecha. Sin identificador de turno ni de cliente.
type PublicReview struct {
	Rating    int
	Comment   *string
	CreatedAt time.Time
}

// Summary es el promedio de 1 decimal y el total sobre las reseñas no ocultas;
// ambos nulos/cero cuando no hay ninguna visible (DEC-128 §5).
type Summary struct {
	Rating      *float64
	RatingCount int
}

// PublicList es la respuesta pública de un barbero: su resumen y las reseñas
// visibles más recientes.
type PublicList struct {
	Summary Summary
	Items   []PublicReview
}

// PanelList es la vista de moderación: resumen público y todas las reseñas,
// incluidas las ocultas, de la más reciente a la más antigua.
type PanelList struct {
	Summary Summary
	Items   []Review
}

// NormalizeComment recorta los extremos y convierte vacío o solo espacios en
// nil (mismo criterio de texto corto que el perfil del barbero). ok=false si
// excede MaxCommentLength, contado en caracteres.
func NormalizeComment(raw *string) (comment *string, ok bool) {
	if raw == nil {
		return nil, true
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, true
	}
	if utf8.RuneCountInString(trimmed) > MaxCommentLength {
		return nil, false
	}
	return &trimmed, true
}

// ValidRating indica si n está en 1..5.
func ValidRating(n int) bool { return n >= MinRating && n <= MaxRating }
