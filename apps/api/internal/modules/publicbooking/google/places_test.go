package google_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"system-barbershop/internal/modules/publicbooking"
	"system-barbershop/internal/modules/publicbooking/google"
)

const testKey = "clave-de-servidor-ficticia"

// fakeGoogle es un servidor de Google Places falso que cuenta las llamadas.
type fakeGoogle struct {
	*httptest.Server
	calls   atomic.Int32
	status  int
	body    string
	lastReq atomic.Pointer[http.Request]
}

func newFakeGoogle(t *testing.T, status int, body string) *fakeGoogle {
	t.Helper()
	f := &fakeGoogle{status: status, body: body}
	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.calls.Add(1)
		f.lastReq.Store(r.Clone(r.Context()))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(f.status)
		_, _ = fmt.Fprint(w, f.body)
	}))
	t.Cleanup(f.Close)
	return f
}

const okBody = `{
  "rating": 4.66, "userRatingCount": 128, "googleMapsUri": "https://maps.google.com/?cid=123",
  "reviews": [
    {"rating": 5, "relativePublishTimeDescription": "hace un mes", "text": {"text": "Excelente atención."},
     "authorAttribution": {"displayName": "  Camila   R. ", "uri": "https://www.google.com/maps/contrib/1"}},
    {"rating": 4, "relativePublishTimeDescription": "hace 3 meses",
     "authorAttribution": {"displayName": "Andrés M.", "uri": "javascript:alert(1)"}}
  ]}`

type clock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *clock) now() time.Time { c.mu.Lock(); defer c.mu.Unlock(); return c.t }
func (c *clock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

func newClient(f *fakeGoogle, clk *clock) *google.Client {
	return google.New(google.Config{APIKey: testKey, BaseURL: f.URL, Now: clk.now})
}

func TestFetch_Success_MapsFieldsAndSendsTheServerKeyOnlyInAHeader(t *testing.T) {
	f := newFakeGoogle(t, http.StatusOK, okBody)
	got, err := newClient(f, &clock{t: time.Unix(0, 0)}).Fetch(context.Background(), "ChIJabc_1-X")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if got.Rating == nil || *got.Rating != 4.7 || got.RatingCount != 128 {
		t.Fatalf("calificación mal leída: %+v", got)
	}
	if got.MapsURL == nil || *got.MapsURL != "https://maps.google.com/?cid=123" {
		t.Fatalf("enlace a la ficha mal leído: %v", got.MapsURL)
	}
	if len(got.Reviews) != 2 {
		t.Fatalf("esperaba 2 reseñas, hay %d", len(got.Reviews))
	}
	first := got.Reviews[0]
	if first.AuthorName != "Camila R." || first.Rating != 5 || first.Text == nil || *first.Text != "Excelente atención." ||
		first.AuthorURL == nil || first.RelativeTime == nil || *first.RelativeTime != "hace un mes" {
		t.Fatalf("primera reseña mal leída (el nombre debe normalizarse): %+v", first)
	}
	second := got.Reviews[1]
	if second.AuthorURL != nil {
		t.Fatalf("un enlace que no es https (javascript:) nunca debe llegar al navegador: %q", *second.AuthorURL)
	}
	if second.Text != nil {
		t.Fatalf("sin texto, Text debe ser nil: %v", second.Text)
	}

	req := f.lastReq.Load()
	if req.URL.Path != "/places/ChIJabc_1-X" || req.URL.Query().Get("languageCode") != "es" {
		t.Fatalf("petición inesperada: %s", req.URL.String())
	}
	if req.Header.Get("X-Goog-Api-Key") != testKey || !strings.Contains(req.Header.Get("X-Goog-FieldMask"), "reviews") {
		t.Fatalf("faltan la clave o la máscara de campos: %v", req.Header)
	}
	if strings.Contains(req.URL.String(), testKey) {
		t.Fatal("CA-101-04: la clave de servidor nunca debe ir en la URL")
	}
}

func TestFetch_SecondCallInsideTheWindowDoesNotHitGoogle_AndExpiresAfterSixHours(t *testing.T) {
	f := newFakeGoogle(t, http.StatusOK, okBody)
	clk := &clock{t: time.Unix(1_700_000_000, 0)}
	c := newClient(f, clk)

	for i := 0; i < 3; i++ {
		if _, err := c.Fetch(context.Background(), "ChIJabc"); err != nil {
			t.Fatalf("Fetch %d: %v", i, err)
		}
	}
	if f.calls.Load() != 1 {
		t.Fatalf("CA-101-05: tres consultas dentro de la ventana deben costar 1 llamada, costaron %d", f.calls.Load())
	}

	clk.advance(5*time.Hour + 59*time.Minute)
	if _, err := c.Fetch(context.Background(), "ChIJabc"); err != nil || f.calls.Load() != 1 {
		t.Fatalf("a 5 h 59 min sigue en caché: calls=%d err=%v", f.calls.Load(), err)
	}
	clk.advance(2 * time.Minute)
	if _, err := c.Fetch(context.Background(), "ChIJabc"); err != nil || f.calls.Load() != 2 {
		t.Fatalf("pasadas las 6 h debe consultar otra vez: calls=%d err=%v", f.calls.Load(), err)
	}

	// Otra ficha tiene su propia entrada.
	if _, err := c.Fetch(context.Background(), "ChIJotra"); err != nil || f.calls.Load() != 3 {
		t.Fatalf("otra ficha debe consultarse aparte: calls=%d err=%v", f.calls.Load(), err)
	}
}

func TestFetch_Failures_AreUnavailable_AndRememberedBrieflyWithoutRetrying(t *testing.T) {
	cases := map[string]int{
		"cuota agotada":          http.StatusTooManyRequests,
		"identificador inválido": http.StatusNotFound,
		"credencial rechazada":   http.StatusForbidden,
		"error de Google":        http.StatusInternalServerError,
		"petición inválida":      http.StatusBadRequest,
	}
	for name, status := range cases {
		t.Run(name, func(t *testing.T) {
			f := newFakeGoogle(t, status, `{"error":{"message":"no"}}`)
			clk := &clock{t: time.Unix(1_700_000_000, 0)}
			c := newClient(f, clk)

			_, err := c.Fetch(context.Background(), "ChIJabc")
			if !errors.Is(err, publicbooking.ErrGoogleReviewsUnavailable) {
				t.Fatalf("debe ser ErrGoogleReviewsUnavailable, fue %v", err)
			}
			if strings.Contains(err.Error(), testKey) {
				t.Fatal("el error nunca debe contener la clave")
			}
			// Sin reintento síncrono ni una llamada por visita mientras dure el fallo.
			if _, err := c.Fetch(context.Background(), "ChIJabc"); err == nil || f.calls.Load() != 1 {
				t.Fatalf("el fallo debe recordarse: calls=%d err=%v", f.calls.Load(), err)
			}
			clk.advance(2 * time.Minute)
			if _, _ = c.Fetch(context.Background(), "ChIJabc"); f.calls.Load() != 2 {
				t.Fatalf("pasado el minuto debe volver a intentar: calls=%d", f.calls.Load())
			}
		})
	}
}

func TestFetch_NetworkFailureAndGarbage_AreUnavailable(t *testing.T) {
	f := newFakeGoogle(t, http.StatusOK, `esto no es JSON`)
	if _, err := newClient(f, &clock{t: time.Unix(0, 0)}).Fetch(context.Background(), "ChIJabc"); !errors.Is(err, publicbooking.ErrGoogleReviewsUnavailable) {
		t.Fatalf("una respuesta ilegible debe ser no disponible, fue %v", err)
	}

	f2 := newFakeGoogle(t, http.StatusOK, okBody)
	c := newClient(f2, &clock{t: time.Unix(0, 0)})
	f2.Close()
	if _, err := c.Fetch(context.Background(), "ChIJabc"); !errors.Is(err, publicbooking.ErrGoogleReviewsUnavailable) {
		t.Fatalf("sin red debe ser no disponible, fue %v", err)
	}
}

func TestFetch_NoKey_IsUnavailableWithoutCallingGoogle(t *testing.T) {
	f := newFakeGoogle(t, http.StatusOK, okBody)
	c := google.New(google.Config{BaseURL: f.URL})
	if _, err := c.Fetch(context.Background(), "ChIJabc"); !errors.Is(err, publicbooking.ErrGoogleReviewsUnavailable) {
		t.Fatalf("sin credencial debe ser no disponible, fue %v", err)
	}
	if f.calls.Load() != 0 {
		t.Fatalf("sin credencial no debe llamar a Google, llamó %d veces", f.calls.Load())
	}
}

func TestFetch_PlaceWithoutRatingsOrWithBadData_IsEmptyNotAnError(t *testing.T) {
	cases := map[string]string{
		"sin resultados":         `{}`,
		"sin calificaciones":     `{"userRatingCount": 0, "reviews": []}`,
		"calificación imposible": `{"rating": 9, "userRatingCount": 3, "reviews": []}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			f := newFakeGoogle(t, http.StatusOK, body)
			got, err := newClient(f, &clock{t: time.Unix(0, 0)}).Fetch(context.Background(), "ChIJabc")
			if err != nil {
				t.Fatalf("no es un fallo: %v", err)
			}
			if !got.Empty() || got.Reviews == nil || len(got.Reviews) != 0 {
				t.Fatalf("debe ser la estructura vacía explícita: %+v", got)
			}
		})
	}
}

func TestFetch_CapsAtFiveReviews_DropsInvalidOnes_AndTruncatesLongText(t *testing.T) {
	var b strings.Builder
	b.WriteString(`{"rating": 4, "userRatingCount": 9, "reviews": [`)
	b.WriteString(`{"rating": 0, "authorAttribution": {"displayName": "Calificación cero"}},`)
	b.WriteString(`{"rating": 5, "authorAttribution": {"displayName": "   "}},`)
	fmt.Fprintf(&b, `{"rating": 5, "text": {"text": %q}, "authorAttribution": {"displayName": "Larga"}}`, strings.Repeat("ñ", 2500))
	for i := 0; i < 8; i++ {
		fmt.Fprintf(&b, `,{"rating": 4, "authorAttribution": {"displayName": "Autor %d"}}`, i)
	}
	b.WriteString(`]}`)

	f := newFakeGoogle(t, http.StatusOK, b.String())
	got, err := newClient(f, &clock{t: time.Unix(0, 0)}).Fetch(context.Background(), "ChIJabc")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(got.Reviews) != publicbooking.MaxGoogleReviews {
		t.Fatalf("debe entregar como máximo %d reseñas, entregó %d", publicbooking.MaxGoogleReviews, len(got.Reviews))
	}
	if got.Reviews[0].AuthorName != "Larga" || got.Reviews[0].Text == nil || len([]rune(*got.Reviews[0].Text)) != 2000 {
		t.Fatalf("las inválidas se descartan y el texto largo se recorta a 2000: %+v", got.Reviews[0])
	}
}

func TestFetch_ConcurrentVisitsShareASingleCall(t *testing.T) {
	release := make(chan struct{})
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		<-release
		_, _ = fmt.Fprint(w, okBody)
	}))
	t.Cleanup(srv.Close)
	c := google.New(google.Config{APIKey: testKey, BaseURL: srv.URL})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := c.Fetch(context.Background(), "ChIJabc"); err != nil {
				t.Errorf("Fetch: %v", err)
			}
		}()
	}
	time.Sleep(50 * time.Millisecond)
	close(release)
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatalf("diez visitas simultáneas deben compartir una llamada, hubo %d", calls.Load())
	}
}

func TestSimulated_ReturnsAFictionalListingWithoutTheNetwork(t *testing.T) {
	got, err := google.Simulated{}.Fetch(context.Background(), "cualquiera")
	if err != nil || got.Empty() || len(got.Reviews) == 0 || len(got.Reviews) > publicbooking.MaxGoogleReviews {
		t.Fatalf("el simulado debe devolver una ficha válida: %+v err=%v", got, err)
	}
	for _, r := range got.Reviews {
		if !strings.Contains(r.AuthorName, "simulad") {
			t.Fatalf("los autores simulados deben decir que lo son: %q", r.AuthorName)
		}
	}
}
