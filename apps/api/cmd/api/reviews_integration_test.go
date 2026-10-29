package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"system-barbershop/internal/platform/database"
)

// Recorrido HTTP completo de HU-100 contra el router REAL y PostgreSQL con dos
// barberías (database/testdata/hu100_resenas.sql): el cliente califica con el
// token de su enlace → aparece en la selección pública → la barbería la oculta
// desde el panel → desaparece del promedio público → la otra barbería no puede
// tocarla.
const (
	reviewShopUno    = "01010001-0101-0101-0101-010100010001"
	reviewServiceUno = "01010031-0101-0101-0101-010100310031"
	reviewUserUno    = "01019001-0101-0101-0101-010190019001"
	reviewSlugUno    = "barberia-hu100-uno"

	reviewShopDos = "01010002-0101-0101-0101-010100020002"
	reviewUserDos = "01019002-0101-0101-0101-010190029002"
)

// seedCompletedAppointmentWithToken crea un barbero propio, una cita completada
// y un token vigente; devuelve (barberID, token en claro).
func seedCompletedAppointmentWithToken(t *testing.T, db *database.DB, status string) (barberID, token string) {
	t.Helper()
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		t.Fatal(err)
	}
	token = hex.EncodeToString(buf)
	sum := sha256.Sum256([]byte(token))
	hash := hex.EncodeToString(sum[:])

	err := db.InTenantTx(context.Background(), database.BarbershopID(reviewShopUno), func(ctx context.Context, q database.Queries) error {
		if err := q.QueryRow(ctx, `INSERT INTO barber (barbershop_id, full_name) VALUES ($1, $2) RETURNING id`,
			reviewShopUno, "Barbero reseñas "+strconv.FormatInt(time.Now().UnixNano(), 10)).Scan(&barberID); err != nil {
			return err
		}
		var customerID, appointmentID string
		if err := q.QueryRow(ctx, `INSERT INTO customer (barbershop_id, full_name) VALUES ($1, 'Cliente de prueba') RETURNING id`,
			reviewShopUno).Scan(&customerID); err != nil {
			return err
		}
		starts := time.Date(2033, 1, 1, 8, 0, 0, 0, time.UTC).AddDate(0, 0, int(time.Now().UnixNano()%3000))
		resolved := "now()"
		if status == "confirmed" {
			resolved = "NULL"
		}
		if err := q.QueryRow(ctx,
			`INSERT INTO appointment (barbershop_id, barber_id, service_id, customer_id, attendee_name, starts_at, ends_at,
			                          status, origin, resolved_at, service_name_snapshot, duration_minutes_snapshot,
			                          price_amount_snapshot, price_currency_snapshot)
			 VALUES ($1, $2, $3, $4, 'Cliente de prueba', $5, $6, $7, 'manual', `+resolved+`, 'Servicio', 30, 20000, 'COP')
			 RETURNING id`, reviewShopUno, barberID, reviewServiceUno, customerID, starts, starts.Add(30*time.Minute), status).Scan(&appointmentID); err != nil {
			return err
		}
		issued := time.Now().UTC().Add(-time.Hour)
		_, err := q.Exec(ctx,
			`INSERT INTO appointment_access_token (barbershop_id, appointment_id, token_hash, issued_at, expires_at)
			 VALUES ($1, $2, $3, $4, $5)`, reviewShopUno, appointmentID, hash, issued, issued.Add(90*24*time.Hour))
		return err
	})
	if err != nil {
		t.Fatalf("sembrar cita completada: %v", err)
	}
	return barberID, token
}

func reviewRequest(router *chi.Mux, method, path, cookie, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: cookie})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

type publicReviews struct {
	Rating      *float64 `json:"rating"`
	RatingCount int      `json:"ratingCount"`
	Items       []struct {
		Rating  int     `json:"rating"`
		Comment *string `json:"comment"`
	} `json:"items"`
}

func decodeJSON[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	return v
}

func TestReviews_HTTP_Journey_RateThenPublicThenHideThenGone(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	barberID, token := seedCompletedAppointmentWithToken(t, db, "completed")
	reviewPath := "/api/v1/customer/appointments/" + token + "/review"
	publicPath := "/api/v1/public/barbershops/" + reviewSlugUno + "/barbers/" + barberID + "/reviews"
	panelPath := "/api/v1/private/barbers/" + barberID + "/reviews"
	cookieUno := createSessionCookie(t, db, reviewShopUno, reviewUserUno, "reviews-uno")
	cookieDos := createSessionCookie(t, db, reviewShopDos, reviewUserDos, "reviews-dos")

	// El turno completado ofrece «Calificar» (CA-100-01).
	got := reviewRequest(router, http.MethodGet, "/api/v1/customer/appointments/"+token, "", "")
	if got.Code != http.StatusOK || !strings.Contains(got.Body.String(), `"canReview":true`) {
		t.Fatalf("GET turno: %d %s", got.Code, got.Body.String())
	}

	// Entradas inválidas: 422 sin guardar nada (CA-100-02).
	for _, body := range []string{`{}`, `{"rating":0}`, `{"rating":6}`, `{"rating":5,"comment":"` + strings.Repeat("a", 501) + `"}`} {
		if rec := reviewRequest(router, http.MethodPost, reviewPath, "", body); rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("cuerpo %.40s: esperaba 422, fue %d: %s", body, rec.Code, rec.Body.String())
		}
	}

	// Calificar: 201, y el comentario de solo espacios se guarda como null.
	rec := reviewRequest(router, http.MethodPost, reviewPath, "", `{"rating":2,"comment":"   "}`)
	if rec.Code != http.StatusCreated || !strings.Contains(rec.Body.String(), `"comment":null`) {
		t.Fatalf("POST review: %d %s", rec.Code, rec.Body.String())
	}
	// Un segundo intento sobre el mismo turno: 409 uniforme, sin modificar la existente (CA-100-03).
	if rec := reviewRequest(router, http.MethodPost, reviewPath, "", `{"rating":5,"comment":"otra"}`); rec.Code != http.StatusConflict {
		t.Fatalf("segundo intento: esperaba 409, fue %d", rec.Code)
	}
	got = reviewRequest(router, http.MethodGet, "/api/v1/customer/appointments/"+token, "", "")
	if !strings.Contains(got.Body.String(), `"canReview":false`) || !strings.Contains(got.Body.String(), `"review":{"rating":2`) {
		t.Fatalf("tras calificar: %s", got.Body.String())
	}

	// Visible de inmediato en la selección pública (CA-100-04).
	pub := decodeJSON[publicReviews](t, reviewRequest(router, http.MethodGet, publicPath, "", ""))
	if pub.Rating == nil || *pub.Rating != 2 || pub.RatingCount != 1 || len(pub.Items) != 1 || pub.Items[0].Rating != 2 {
		t.Fatalf("lista pública: %+v", pub)
	}

	// La otra barbería no ve ni modifica nada (RN-TEN-01).
	if rec := reviewRequest(router, http.MethodGet, panelPath, cookieDos, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("panel de Dos sobre un barbero de Uno: esperaba 404, fue %d", rec.Code)
	}
	panel := decodeJSON[struct {
		Items []struct {
			ID       string  `json:"id"`
			HiddenAt *string `json:"hiddenAt"`
		} `json:"items"`
	}](t, reviewRequest(router, http.MethodGet, panelPath, cookieUno, ""))
	if len(panel.Items) != 1 || panel.Items[0].HiddenAt != nil {
		t.Fatalf("panel de Uno: %+v", panel)
	}
	reviewPatch := panelPath + "/" + panel.Items[0].ID
	if rec := reviewRequest(router, http.MethodPatch, reviewPatch, cookieDos, `{"hidden":true}`); rec.Code != http.StatusNotFound {
		t.Fatalf("Dos no debe poder ocultar una reseña de Uno: %d", rec.Code)
	}

	// Ocultar: sale del promedio y de la lista pública, pero el panel la conserva (CA-100-05).
	if rec := reviewRequest(router, http.MethodPatch, reviewPatch, cookieUno, `{"hidden":true}`); rec.Code != http.StatusOK {
		t.Fatalf("ocultar: %d %s", rec.Code, rec.Body.String())
	}
	pub = decodeJSON[publicReviews](t, reviewRequest(router, http.MethodGet, publicPath, "", ""))
	if pub.Rating != nil || pub.RatingCount != 0 || len(pub.Items) != 0 {
		t.Fatalf("tras ocultar: %+v", pub)
	}
	panelAgain := reviewRequest(router, http.MethodGet, panelPath, cookieUno, "")
	if !strings.Contains(panelAgain.Body.String(), `"hiddenAt":"`) {
		t.Fatalf("el panel debe conservar la oculta con su fecha: %s", panelAgain.Body.String())
	}

	// Volver a mostrar la reincorpora.
	if rec := reviewRequest(router, http.MethodPatch, reviewPatch, cookieUno, `{"hidden":false}`); rec.Code != http.StatusOK {
		t.Fatalf("mostrar: %d", rec.Code)
	}
	pub = decodeJSON[publicReviews](t, reviewRequest(router, http.MethodGet, publicPath, "", ""))
	if pub.RatingCount != 1 {
		t.Fatalf("tras mostrarla de nuevo: %+v", pub)
	}
}

func TestReviews_HTTP_NotReviewableAppointmentsAndBadTokens_AreUniform(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { db.Close() })
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}

	_, confirmedToken := seedCompletedAppointmentWithToken(t, db, "confirmed")
	if rec := reviewRequest(router, http.MethodPost, "/api/v1/customer/appointments/"+confirmedToken+"/review", "", `{"rating":5}`); rec.Code != http.StatusConflict {
		t.Fatalf("turno confirmado: esperaba 409, fue %d", rec.Code)
	}
	got := reviewRequest(router, http.MethodGet, "/api/v1/customer/appointments/"+confirmedToken, "", "")
	if !strings.Contains(got.Body.String(), `"canReview":false`) {
		t.Fatalf("un turno confirmado no ofrece «Calificar»: %s", got.Body.String())
	}

	for _, token := range []string{"token-inexistente", strings.Repeat("a", 300)} {
		if rec := reviewRequest(router, http.MethodPost, "/api/v1/customer/appointments/"+token+"/review", "", `{"rating":5}`); rec.Code != http.StatusNotFound {
			t.Fatalf("token %.12s: esperaba 404, fue %d", token, rec.Code)
		}
	}
	// Panel y lista pública sin sesión / con barbero inexistente.
	if rec := reviewRequest(router, http.MethodGet, "/api/v1/private/barbers/8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4/reviews", "", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("panel sin sesión: esperaba 401, fue %d", rec.Code)
	}
	if rec := reviewRequest(router, http.MethodGet, "/api/v1/public/barbershops/"+reviewSlugUno+"/barbers/8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4/reviews", "", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("barbero inexistente: esperaba 404, fue %d", rec.Code)
	}
}
