// Recorrido real de HU-099 sin dobles: handlers HTTP de customeraccess →
// Service → booking (adaptador, servicio y repositorio PostgreSQL) con dos
// barberías. Prueba el cableado que cmd/api hace: versionToken emitido por
// booking.EncodeVersionToken, token → cita, política, historial y revocación.
package postgres_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"system-barbershop/internal/modules/booking"
	bookingpostgres "system-barbershop/internal/modules/booking/postgres"
	"system-barbershop/internal/modules/customeraccess"
	"system-barbershop/internal/modules/customeraccess/httpapi"
	customeraccesspostgres "system-barbershop/internal/modules/customeraccess/postgres"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/httpserver"
	"system-barbershop/internal/platform/idempotency"
)

func realCancelHandlers(t *testing.T, db *database.DB) (*httpapi.GetAppointmentHandler, *httpapi.CancelAppointmentHandler) {
	t.Helper()
	bookingRepo := bookingpostgres.New(db, idempotency.NewSQLCoordinator())
	svc := customeraccess.NewService(
		customeraccesspostgres.New(db),
		customeraccess.WithVersionTokens(booking.EncodeVersionToken),
		customeraccess.WithCanceller(booking.NewCustomerAppointmentCancellerAdapter(
			booking.NewCancelAppointmentByCustomerService(bookingRepo, time.Now))),
	)
	return httpapi.NewGetAppointmentHandler(svc), httpapi.NewCancelAppointmentHandler(svc)
}

func doGet(h http.Handler, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/customer/appointments/"+token, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httpserver.RequestWithURLParam(req, "token", token))
	return rec
}

func doCancel(h http.Handler, token, body, key, version string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/customer/appointments/"+token+"/cancel", strings.NewReader(body))
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	if version != "" {
		req.Header.Set("If-Match", version)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httpserver.RequestWithURLParam(req, "token", token))
	return rec
}

func TestCancelFlow_ReadCancelThenLinkStillShowsCancelled(t *testing.T) {
	db := setupTestDB(t)
	get, cancel := realCancelHandlers(t, db)

	appointment := createTestAppointment(t, db, shopUno, barberUno, serviceUno, booking.StatusConfirmed)
	token := insertAccessToken(t, db, shopUno, appointment.ID, tokenState{expiresIn: 90 * 24 * time.Hour})

	read := doGet(get, token)
	var view httpapi.CustomerAppointmentResponse
	if read.Code != http.StatusOK || json.Unmarshal(read.Body.Bytes(), &view) != nil || view.VersionToken == "" || view.Status != "confirmed" {
		t.Fatalf("GET: %d %s", read.Code, read.Body.String())
	}

	key := "0b7d5c8e-3f4a-4d6b-9a1e-" + uniqueDigits(t, 12)
	first := doCancel(cancel, token, `{"reason":"cambio de planes"}`, key, view.VersionToken)
	if first.Code != http.StatusOK || strings.TrimSpace(first.Body.String()) != `{"status":"cancelled_by_customer"}` {
		t.Fatalf("cancelación: %d %s", first.Code, first.Body.String())
	}

	// DEC-136: cancelar NO revoca el enlace; sigue mostrando el turno como
	// cancelado (sin acciones). Repetir la llamada con la misma clave es una
	// réplica idempotente que no modifica nada.
	after := doGet(get, token)
	var cancelledView httpapi.CustomerAppointmentResponse
	if after.Code != http.StatusOK || json.Unmarshal(after.Body.Bytes(), &cancelledView) != nil || cancelledView.Status != "cancelled_by_customer" || cancelledView.CanReview {
		t.Fatalf("GET tras cancelar: %d %s", after.Code, after.Body.String())
	}
	if again := doCancel(cancel, token, `{"reason":"cambio de planes"}`, key, view.VersionToken); again.Code != http.StatusOK {
		t.Fatalf("repetición con la misma clave: %d %s", again.Code, again.Body.String())
	}

	var status string
	var events int
	err := db.InTenantTx(context.Background(), database.BarbershopID(shopUno), func(ctx context.Context, q database.Queries) error {
		if err := q.QueryRow(ctx, `SELECT status FROM appointment WHERE barbershop_id = $1 AND id = $2`, shopUno, appointment.ID).Scan(&status); err != nil {
			return err
		}
		return q.QueryRow(ctx, `SELECT count(*) FROM appointment_history WHERE barbershop_id = $1 AND appointment_id = $2 AND event_type = 'appointment_cancelled_by_customer'`,
			shopUno, appointment.ID).Scan(&events)
	})
	if err != nil || status != "cancelled_by_customer" || events != 1 {
		t.Fatalf("estado=%q eventos=%d err=%v", status, events, err)
	}
}

func TestCancelFlow_StaleVersionAndForeignTokenChangeNothing(t *testing.T) {
	db := setupTestDB(t)
	get, cancel := realCancelHandlers(t, db)

	apptUno := createTestAppointment(t, db, shopUno, barberUno, serviceUno, booking.StatusConfirmed)
	tokenUno := insertAccessToken(t, db, shopUno, apptUno.ID, tokenState{expiresIn: 90 * 24 * time.Hour})
	apptDos := createTestAppointment(t, db, shopDos, barberDos, serviceDos, booking.StatusConfirmed)
	tokenDos := insertAccessToken(t, db, shopDos, apptDos.ID, tokenState{expiresIn: 90 * 24 * time.Hour})

	var viewDos httpapi.CustomerAppointmentResponse
	if err := json.Unmarshal(doGet(get, tokenDos).Body.Bytes(), &viewDos); err != nil {
		t.Fatalf("GET shopDos: %v", err)
	}

	key := "1c8e6d9f-4a5b-4e7c-8b2f-" + uniqueDigits(t, 12)
	// Versión de OTRA cita: obsoleta para esta -> 409 sin cambios.
	if rec := doCancel(cancel, tokenUno, "", key, viewDos.VersionToken); rec.Code != http.StatusConflict {
		t.Fatalf("versión ajena: %d %s, se esperaba 409", rec.Code, rec.Body.String())
	}
	// Token inexistente -> 404 uniforme.
	if rec := doCancel(cancel, "token-que-no-existe", "", key, viewDos.VersionToken); rec.Code != http.StatusNotFound {
		t.Fatalf("token inexistente: %d, se esperaba 404", rec.Code)
	}
	// Ambas citas siguen confirmadas y la de shopDos no se tocó.
	for _, tc := range []struct{ shop, id string }{{shopUno, apptUno.ID}, {shopDos, apptDos.ID}} {
		var status string
		err := db.InTenantTx(context.Background(), database.BarbershopID(tc.shop), func(ctx context.Context, q database.Queries) error {
			return q.QueryRow(ctx, `SELECT status FROM appointment WHERE barbershop_id = $1 AND id = $2`, tc.shop, tc.id).Scan(&status)
		})
		if err != nil || status != "confirmed" {
			t.Fatalf("cita %s: estado=%q err=%v", tc.id, status, err)
		}
	}
}
