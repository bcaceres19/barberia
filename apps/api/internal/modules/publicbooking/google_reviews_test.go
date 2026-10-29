// Pruebas del servicio de reseñas de Google (HU-101, DEC-129) en aislamiento: un
// resolvedor y un proveedor dobles. La caché y el adaptador HTTP se prueban en
// publicbooking/google; el SQL, en postgres/repository_test.go.
package publicbooking_test

import (
	"context"
	"errors"
	"testing"

	"system-barbershop/internal/modules/publicbooking"
	"system-barbershop/internal/platform/apperr"
)

type fakeResolver struct {
	profile publicbooking.BarbershopProfile
	err     error
}

func (f fakeResolver) ResolveBarbershop(context.Context, string) (publicbooking.BarbershopProfile, error) {
	return f.profile, f.err
}

type fakeReviewsProvider struct {
	reviews publicbooking.GoogleReviews
	err     error
	calls   []string
}

func (f *fakeReviewsProvider) Fetch(_ context.Context, placeID string) (publicbooking.GoogleReviews, error) {
	f.calls = append(f.calls, placeID)
	return f.reviews, f.err
}

func placeID(s string) *string { return &s }

func sampleReviews() publicbooking.GoogleReviews {
	rating := 4.7
	return publicbooking.GoogleReviews{Rating: &rating, RatingCount: 128, Reviews: []publicbooking.GoogleReview{{AuthorName: "Camila R.", Rating: 5}}}
}

func TestGoogleReviews_WithoutPlaceID_IsEmptyAndNeverCallsGoogle(t *testing.T) {
	provider := &fakeReviewsProvider{reviews: sampleReviews()}
	svc := publicbooking.NewGoogleReviewsService(fakeResolver{profile: publicbooking.BarbershopProfile{}}, provider, nil)

	got, err := svc.Get(context.Background(), "barberia-ejemplo")
	if err != nil || !got.Empty() {
		t.Fatalf("CA-101-01: sin ficha debe ser vacío sin error, obtuvo %+v err=%v", got, err)
	}
	if len(provider.calls) != 0 {
		t.Fatalf("sin ficha no se consulta a Google, hubo %d llamadas", len(provider.calls))
	}
}

func TestGoogleReviews_WithoutProvider_IsEmpty(t *testing.T) {
	svc := publicbooking.NewGoogleReviewsService(fakeResolver{profile: publicbooking.BarbershopProfile{GooglePlaceID: placeID("ChIJabc")}}, nil, nil)

	got, err := svc.Get(context.Background(), "barberia-ejemplo")
	if err != nil || !got.Empty() {
		t.Fatalf("sin credencial la integración queda desactivada: %+v err=%v", got, err)
	}
}

func TestGoogleReviews_WithPlaceID_ReturnsWhatTheProviderReturned(t *testing.T) {
	provider := &fakeReviewsProvider{reviews: sampleReviews()}
	svc := publicbooking.NewGoogleReviewsService(fakeResolver{profile: publicbooking.BarbershopProfile{GooglePlaceID: placeID("ChIJabc")}}, provider, nil)

	got, err := svc.Get(context.Background(), "barberia-ejemplo")
	if err != nil || got.Rating == nil || *got.Rating != 4.7 || len(got.Reviews) != 1 {
		t.Fatalf("CA-101-02: debe devolver la calificación y las reseñas, obtuvo %+v err=%v", got, err)
	}
	if len(provider.calls) != 1 || provider.calls[0] != "ChIJabc" {
		t.Fatalf("debe consultar la ficha configurada: %v", provider.calls)
	}
}

func TestGoogleReviews_ProviderFailure_IsEmptyAndNeverAnError(t *testing.T) {
	for name, cause := range map[string]error{
		"cuota o red":      publicbooking.ErrGoogleReviewsUnavailable,
		"error inesperado": errors.New("boom"),
	} {
		t.Run(name, func(t *testing.T) {
			provider := &fakeReviewsProvider{err: cause}
			svc := publicbooking.NewGoogleReviewsService(fakeResolver{profile: publicbooking.BarbershopProfile{GooglePlaceID: placeID("ChIJabc")}}, provider, nil)

			got, err := svc.Get(context.Background(), "barberia-ejemplo")
			if err != nil || !got.Empty() {
				t.Fatalf("CA-101-03: un fallo no se propaga al cliente: %+v err=%v", got, err)
			}
		})
	}
}

func TestGoogleReviews_UnknownOrUnpublishedSlug_KeepsTheUniform404(t *testing.T) {
	provider := &fakeReviewsProvider{reviews: sampleReviews()}
	svc := publicbooking.NewGoogleReviewsService(fakeResolver{err: apperr.NotFound("barbería no disponible")}, provider, nil)

	_, err := svc.Get(context.Background(), "no-existe")
	var appErr *apperr.Error
	if !errors.As(err, &appErr) || appErr.Kind != apperr.KindNotFound {
		t.Fatalf("un slug no publicable debe ser el 404 uniforme: %v", err)
	}
	if len(provider.calls) != 0 {
		t.Fatal("sin barbería no se consulta a Google")
	}
}
