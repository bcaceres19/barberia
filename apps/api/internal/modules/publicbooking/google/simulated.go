package google

import (
	"context"

	"system-barbershop/internal/modules/publicbooking"
)

// Simulated es el proveedor de desarrollo (local y test, sin clave): devuelve la
// misma ficha ficticia para cualquier Place ID, sin tocar la red, para poder ver
// y probar la sección sin credenciales. Los autores son inventados.
type Simulated struct{}

var _ publicbooking.GoogleReviewsProvider = Simulated{}

// Fetch implementa publicbooking.GoogleReviewsProvider.
func (Simulated) Fetch(_ context.Context, _ string) (publicbooking.GoogleReviews, error) {
	rating := 4.7
	maps := "https://maps.google.com/?cid=0"
	text1, text2 := "Excelente atención y muy puntuales.", "Me encantó el corte, volveré."
	t1, t2, t3 := "hace un mes", "hace 3 meses", "hace 6 meses"
	return publicbooking.GoogleReviews{
		Rating:      &rating,
		RatingCount: 128,
		MapsURL:     &maps,
		Reviews: []publicbooking.GoogleReview{
			{AuthorName: "Camila R. (simulada)", Rating: 5, Text: &text1, RelativeTime: &t1},
			{AuthorName: "Andrés M. (simulado)", Rating: 4, RelativeTime: &t2},
			{AuthorName: "Lucía P. (simulada)", Rating: 5, Text: &text2, RelativeTime: &t3},
		},
	}, nil
}
