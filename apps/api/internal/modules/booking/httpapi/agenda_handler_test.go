package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"system-barbershop/internal/modules/booking"
	"system-barbershop/internal/modules/booking/httpapi"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
)

// fakeAgendaBarberPort/fakeAgendaClock/fakeAgendaRepository son dobles
// mínimos para probar ListDailyAgendaHandler en aislamiento (decodificación
// de query params, extracción del principal, forma de la respuesta,
// traducción de errores), no PostgreSQL real (esa cobertura vive en
// internal/modules/booking/postgres/agenda_repository_test.go) ni la
// orquestación de AgendaService (esa vive en
// internal/modules/booking/agenda_test.go).
type fakeAgendaBarberPort struct {
	exists bool
	err    error
}

func (f fakeAgendaBarberPort) Exists(context.Context, string, string) (bool, error) {
	return f.exists, f.err
}

type fakeAgendaClock struct{ now time.Time }

func (f fakeAgendaClock) Now() time.Time { return f.now }

type fakeAgendaRepository struct {
	fakeRepository

	entries []booking.DailyAgendaEntry
	err     error
}

func (f *fakeAgendaRepository) ListDailyAgenda(context.Context, string, string, time.Time, time.Time) ([]booking.DailyAgendaEntry, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.entries, nil
}

const validAgendaBarberID = "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"

func newAgendaHandler(repo *fakeAgendaRepository, barbers fakeAgendaBarberPort, timezones fakeTimezones) *httpapi.ListDailyAgendaHandler {
	service := booking.NewAgendaService(repo, barbers, timezones, fakeAgendaClock{now: time.Now()})
	return httpapi.NewListDailyAgendaHandler(service)
}

func TestListDailyAgendaHandler_ValidRequest_Returns200WithItems(t *testing.T) {
	entries := []booking.DailyAgendaEntry{
		{
			ID:                       "appt-1",
			AttendeeName:             "Carlos Restrepo",
			StartsAt:                 time.Date(2026, 8, 28, 14, 30, 0, 0, time.UTC),
			EndsAt:                   time.Date(2026, 8, 28, 15, 0, 0, 0, time.UTC),
			Status:                   booking.StatusConfirmed,
			Origin:                   booking.OriginManual,
			ServiceNameSnapshot:      "Corte clásico",
			DurationMinutesSnapshot:  30,
			PriceAmountCentsSnapshot: 2000000,
			CurrencySnapshot:         "COP",
		},
	}
	repo := &fakeAgendaRepository{entries: entries}
	h := newAgendaHandler(repo, fakeAgendaBarberPort{exists: true}, fakeTimezones{})

	req := requestWithPrincipal(http.MethodGet, "/api/v1/private/barbers/"+validAgendaBarberID+"/appointments/daily-agenda?date=2026-08-28", nil)
	req = httpserver.RequestWithURLParam(req, "barberId", validAgendaBarberID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var got httpapi.DailyAgendaResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(got.Items) != 1 {
		t.Fatalf("Items = %+v, want 1 entry", got.Items)
	}
	item := got.Items[0]
	if item.ID != "appt-1" || item.AttendeeName != "Carlos Restrepo" || item.ServiceName != "Corte clásico" {
		t.Fatalf("item inesperado: %+v", item)
	}
	if item.PriceAmount != "20000.00" || item.Currency != "COP" {
		t.Fatalf("precio/moneda inesperados: %+v", item)
	}
	if item.Status != "confirmed" || item.Origin != "manual" {
		t.Fatalf("status/origin inesperados: %+v", item)
	}

	// CA-062-05: nunca teléfono, correo, nota, customerId ni barberId por
	// fila (ya fijado por la ruta). Verificado leyendo el JSON crudo, no
	// solo el struct tipado, para detectar un campo agregado por error que
	// el struct de la prueba no declare.
	var raw map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}
	rawItems, _ := raw["items"].([]any)
	if len(rawItems) != 1 {
		t.Fatalf("raw items = %+v", rawItems)
	}
	rawItem, _ := rawItems[0].(map[string]any)
	for _, forbidden := range []string{"customerPhone", "customerEmail", "customerNote", "customerId", "barberId", "phone", "email", "note"} {
		if _, ok := rawItem[forbidden]; ok {
			t.Fatalf("la respuesta expuso el campo prohibido %q: %+v", forbidden, rawItem)
		}
	}
}

func TestListDailyAgendaHandler_EmptyAgenda_Returns200WithEmptyItems(t *testing.T) {
	repo := &fakeAgendaRepository{entries: nil}
	h := newAgendaHandler(repo, fakeAgendaBarberPort{exists: true}, fakeTimezones{})

	req := requestWithPrincipal(http.MethodGet, "/api/v1/private/barbers/"+validAgendaBarberID+"/appointments/daily-agenda", nil)
	req = httpserver.RequestWithURLParam(req, "barberId", validAgendaBarberID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got httpapi.DailyAgendaResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(got.Items) != 0 {
		t.Fatalf("Items = %+v, want empty", got.Items)
	}
}

func TestListDailyAgendaHandler_BarberNotFound_Returns404(t *testing.T) {
	repo := &fakeAgendaRepository{}
	h := newAgendaHandler(repo, fakeAgendaBarberPort{exists: false}, fakeTimezones{})

	req := requestWithPrincipal(http.MethodGet, "/api/v1/private/barbers/"+validAgendaBarberID+"/appointments/daily-agenda", nil)
	req = httpserver.RequestWithURLParam(req, "barberId", validAgendaBarberID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListDailyAgendaHandler_InvalidDate_Returns400(t *testing.T) {
	repo := &fakeAgendaRepository{}
	h := newAgendaHandler(repo, fakeAgendaBarberPort{exists: true}, fakeTimezones{})

	req := requestWithPrincipal(http.MethodGet, "/api/v1/private/barbers/"+validAgendaBarberID+"/appointments/daily-agenda?date=not-a-date", nil)
	req = httpserver.RequestWithURLParam(req, "barberId", validAgendaBarberID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListDailyAgendaHandler_NoPrincipal_Returns500(t *testing.T) {
	repo := &fakeAgendaRepository{}
	h := newAgendaHandler(repo, fakeAgendaBarberPort{exists: true}, fakeTimezones{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/private/barbers/"+validAgendaBarberID+"/appointments/daily-agenda", nil)
	req = httpserver.RequestWithURLParam(req, "barberId", validAgendaBarberID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListDailyAgendaHandler_RepositoryError_Returns500(t *testing.T) {
	repo := &fakeAgendaRepository{err: apperr.Internal(nil)}
	h := newAgendaHandler(repo, fakeAgendaBarberPort{exists: true}, fakeTimezones{})

	req := requestWithPrincipal(http.MethodGet, "/api/v1/private/barbers/"+validAgendaBarberID+"/appointments/daily-agenda", nil)
	req = httpserver.RequestWithURLParam(req, "barberId", validAgendaBarberID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}
