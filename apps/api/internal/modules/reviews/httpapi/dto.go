package httpapi

import "time"

// CreateReviewRequest es el cuerpo de POST /customer/appointments/{token}/review
// (CA-100-02). rating es puntero para distinguir «ausente» (422) de un valor
// fuera de rango (422); comment es opcional y nullable.
type CreateReviewRequest struct {
	Rating  *int    `json:"rating"`
	Comment *string `json:"comment"`
}

// CustomerReviewResponse es lo que el cliente recibe de vuelta al calificar y lo
// que `GET /customer/appointments/{token}` anida en `review` (forma de
// CustomerReview.yaml): sin identificadores internos.
type CustomerReviewResponse struct {
	Rating    int       `json:"rating"`
	Comment   *string   `json:"comment"`
	CreatedAt time.Time `json:"createdAt"`
}

// PublicReviewResponse es una reseña visible (PublicReview.yaml).
type PublicReviewResponse struct {
	Rating    int       `json:"rating"`
	Comment   *string   `json:"comment"`
	CreatedAt time.Time `json:"createdAt"`
}

// PublicReviewListResponse es la respuesta de
// GET /public/barbershops/{slug}/barbers/{barberId}/reviews.
type PublicReviewListResponse struct {
	Rating      *float64               `json:"rating"`
	RatingCount int                    `json:"ratingCount"`
	Items       []PublicReviewResponse `json:"items"`
}

// ReviewResponse es una reseña vista por quien modera (Review.yaml): con su
// identificador y la fecha de ocultamiento (null = visible).
type ReviewResponse struct {
	ID        string     `json:"id"`
	Rating    int        `json:"rating"`
	Comment   *string    `json:"comment"`
	HiddenAt  *time.Time `json:"hiddenAt"`
	CreatedAt time.Time  `json:"createdAt"`
}

// ReviewListResponse es la respuesta de GET /private/barbers/{barberId}/reviews.
type ReviewListResponse struct {
	Rating      *float64         `json:"rating"`
	RatingCount int              `json:"ratingCount"`
	Items       []ReviewResponse `json:"items"`
}

// UpdateReviewRequest es el cuerpo de PATCH /private/barbers/{barberId}/reviews/{reviewId}.
type UpdateReviewRequest struct {
	Hidden *bool `json:"hidden"`
}
