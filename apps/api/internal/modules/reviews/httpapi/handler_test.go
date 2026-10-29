// Pruebas del adaptador HTTP de reviews en aislamiento (decodificación estricta,
// códigos uniformes, cabeceras de privacidad). La persistencia real vive en
// reviews/postgres/repository_test.go.
package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/modules/reviews"
	"system-barbershop/internal/modules/reviews/httpapi"
	"system-barbershop/internal/platform/httpserver"
)

const (
	barberID = "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"
	reviewID = "1b2c3d4e-5f60-4718-8293-a4b5c6d7e8f9"
)

type fakeRepo struct {
	outcome reviews.CreateOutcome
	created reviews.Review
	calls   int
	hidden  *bool

	public   reviews.PublicList
	panel    reviews.PanelList
	found    bool
	setFound bool
}

func (f *fakeRepo) CreateByTokenHash(_ context.Context, _ string, rating int, comment *string) (reviews.Review, reviews.CreateOutcome, error) {
	f.calls++
	r := f.created
	r.Rating, r.Comment = rating, comment
	return r, f.outcome, nil
}
func (f *fakeRepo) ListPublic(context.Context, string, string, int) (reviews.PublicList, bool, error) {
	return f.public, f.found, nil
}
func (f *fakeRepo) ListForBarber(context.Context, string, string, int) (reviews.PanelList, bool, error) {
	return f.panel, f.found, nil
}
func (f *fakeRepo) SetHidden(_ context.Context, _, _, _ string, hidden bool) (reviews.Review, bool, error) {
	f.hidden = &hidden
	r := f.created
	if hidden {
		now := time.Now()
		r.HiddenAt = &now
	}
	return r, f.setFound, nil
}

func jsonRequest(method, target, body string, params map[string]string) *http.Request {
	req := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range params {
		req = httpserver.RequestWithURLParam(req, k, v)
	}
	return req
}

func withPrincipal(req *http.Request) *http.Request {
	return req.WithContext(auth.ContextWithPrincipal(req.Context(), auth.Principal{BarbershopID: "shop-1"}))
}

func TestCreateHandler_Created_ReturnsOnlyTheCustomerSafeShapeWithNoStore(t *testing.T) {
	repo := &fakeRepo{outcome: reviews.Created, created: reviews.Review{ID: "interno", BarberID: barberID, CreatedAt: time.Date(2027, 3, 15, 16, 10, 0, 0, time.UTC)}}
	h := httpapi.NewCreateReviewHandler(reviews.NewService(repo))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, jsonRequest(http.MethodPost, "/x", `{"rating":5,"comment":"  Muy bien  "}`, map[string]string{"token": "tok"}))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("Cache-Control = %q", rec.Header().Get("Cache-Control"))
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body) != 3 || body["rating"] != float64(5) || body["comment"] != "Muy bien" || body["createdAt"] == nil {
		t.Fatalf("cuerpo = %v, se esperaban exactamente rating, comment (recortado) y createdAt", body)
	}
	if strings.Contains(rec.Body.String(), "interno") || strings.Contains(rec.Body.String(), barberID) {
		t.Fatalf("la respuesta del cliente no debe exponer identificadores: %s", rec.Body.String())
	}
}

func TestCreateHandler_Errors(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		outcome reviews.CreateOutcome
		want    int
		calls   int
	}{
		{"calificación ausente", `{"comment":"hola"}`, reviews.Created, http.StatusUnprocessableEntity, 0},
		{"calificación 0", `{"rating":0}`, reviews.Created, http.StatusUnprocessableEntity, 0},
		{"calificación 6", `{"rating":6}`, reviews.Created, http.StatusUnprocessableEntity, 0},
		{"comentario de 501", `{"rating":3,"comment":"` + strings.Repeat("a", 501) + `"}`, reviews.Created, http.StatusUnprocessableEntity, 0},
		{"campo desconocido", `{"rating":3,"barberId":"x"}`, reviews.Created, http.StatusBadRequest, 0},
		{"JSON inválido", `{`, reviews.Created, http.StatusBadRequest, 0},
		{"calificación decimal", `{"rating":4.5}`, reviews.Created, http.StatusBadRequest, 0},
		{"token desconocido", `{"rating":3}`, reviews.TokenNotFound, http.StatusNotFound, 1},
		{"turno no calificable", `{"rating":3}`, reviews.NotAvailable, http.StatusConflict, 1},
	}
	for _, c := range cases {
		repo := &fakeRepo{outcome: c.outcome}
		rec := httptest.NewRecorder()
		httpapi.NewCreateReviewHandler(reviews.NewService(repo)).ServeHTTP(rec,
			jsonRequest(http.MethodPost, "/x", c.body, map[string]string{"token": "tok"}))
		if rec.Code != c.want || repo.calls != c.calls {
			t.Errorf("%s: status = %d (want %d), repo calls = %d (want %d): %s", c.name, rec.Code, c.want, repo.calls, c.calls, rec.Body.String())
		}
		if rec.Header().Get("Cache-Control") != "no-store" {
			t.Errorf("%s: toda respuesta del cliente lleva Cache-Control: no-store", c.name)
		}
	}
}

func TestListPublicHandler_ShapeAndNotFound(t *testing.T) {
	rating := 4.5
	comment := "Bien"
	repo := &fakeRepo{found: true, public: reviews.PublicList{
		Summary: reviews.Summary{Rating: &rating, RatingCount: 2},
		Items:   []reviews.PublicReview{{Rating: 5, Comment: &comment, CreatedAt: time.Unix(0, 0).UTC()}, {Rating: 4}},
	}}
	h := httpapi.NewListPublicReviewsHandler(reviews.NewService(repo))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, jsonRequest(http.MethodGet, "/x", "", map[string]string{"slug": "s", "barberId": barberID}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Rating      *float64         `json:"rating"`
		RatingCount int              `json:"ratingCount"`
		Items       []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Rating == nil || *body.Rating != 4.5 || body.RatingCount != 2 || len(body.Items) != 2 {
		t.Fatalf("cuerpo = %s", rec.Body.String())
	}
	for _, it := range body.Items {
		if len(it) != 3 {
			t.Fatalf("una reseña pública solo lleva rating, comment y createdAt: %v", it)
		}
	}

	// Sin reseñas: items es [] y rating null (nunca null/ausente en items).
	repo.public = reviews.PublicList{}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, jsonRequest(http.MethodGet, "/x", "", map[string]string{"slug": "s", "barberId": barberID}))
	if !strings.Contains(rec.Body.String(), `"items":[]`) || !strings.Contains(rec.Body.String(), `"rating":null`) {
		t.Fatalf("sin reseñas: %s", rec.Body.String())
	}

	repo.found = false
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, jsonRequest(http.MethodGet, "/x", "", map[string]string{"slug": "s", "barberId": barberID}))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("barbero desconocido: status = %d", rec.Code)
	}
}

func TestPanelHandlers_RequireSessionAndModerate(t *testing.T) {
	hidden := time.Date(2027, 3, 16, 9, 0, 0, 0, time.UTC)
	repo := &fakeRepo{found: true, setFound: true,
		created: reviews.Review{ID: reviewID, BarberID: barberID, CreatedAt: time.Unix(0, 0).UTC()},
		panel: reviews.PanelList{Items: []reviews.Review{
			{ID: reviewID, Rating: 1, HiddenAt: &hidden, CreatedAt: time.Unix(0, 0).UTC()},
		}},
	}
	svc := reviews.NewService(repo)
	params := map[string]string{"barberId": barberID, "reviewId": reviewID}

	// Sin principal de sesión: defecto de wiring → 500, nunca datos.
	rec := httptest.NewRecorder()
	httpapi.NewListReviewsHandler(svc).ServeHTTP(rec, jsonRequest(http.MethodGet, "/x", "", params))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("sin sesión: status = %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	httpapi.NewListReviewsHandler(svc).ServeHTTP(rec, withPrincipal(jsonRequest(http.MethodGet, "/x", "", params)))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"hiddenAt":"2027-03-16T09:00:00Z"`) {
		t.Fatalf("listado: status = %d: %s", rec.Code, rec.Body.String())
	}

	// Ocultar y volver a mostrar.
	for _, want := range []bool{true, false} {
		body := `{"hidden":false}`
		if want {
			body = `{"hidden":true}`
		}
		rec = httptest.NewRecorder()
		httpapi.NewUpdateReviewHandler(svc).ServeHTTP(rec, withPrincipal(jsonRequest(http.MethodPatch, "/x", body, params)))
		if rec.Code != http.StatusOK || repo.hidden == nil || *repo.hidden != want {
			t.Fatalf("hidden=%v: status = %d hidden recibido = %v: %s", want, rec.Code, repo.hidden, rec.Body.String())
		}
		if (strings.Contains(rec.Body.String(), `"hiddenAt":null`)) == want {
			t.Fatalf("hidden=%v: hiddenAt inconsistente: %s", want, rec.Body.String())
		}
	}

	// `hidden` es obligatorio; un campo ajeno se rechaza.
	for _, body := range []string{`{}`, `{"hidden":true,"rating":1}`} {
		rec = httptest.NewRecorder()
		httpapi.NewUpdateReviewHandler(svc).ServeHTTP(rec, withPrincipal(jsonRequest(http.MethodPatch, "/x", body, params)))
		if rec.Code != http.StatusUnprocessableEntity && rec.Code != http.StatusBadRequest {
			t.Fatalf("cuerpo %s: status = %d", body, rec.Code)
		}
	}

	// Reseña inexistente → 404.
	repo.setFound = false
	rec = httptest.NewRecorder()
	httpapi.NewUpdateReviewHandler(svc).ServeHTTP(rec, withPrincipal(jsonRequest(http.MethodPatch, "/x", `{"hidden":true}`, params)))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("reseña inexistente: status = %d", rec.Code)
	}
}
