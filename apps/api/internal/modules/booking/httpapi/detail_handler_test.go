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

// fakeDetailBarberNamePort/fakeDetailStaffActorNamePort/fakeDetailRepository
// son dobles mínimos para probar GetAppointmentDetailHandler/
// ListAppointmentHistoryHandler en aislamiento (extracción del principal,
// forma de la respuesta, ausencia de campos prohibidos, traducción de
// errores), no PostgreSQL real (esa cobertura vive en
// internal/modules/booking/postgres/detail_repository_test.go) ni la
// orquestación de DetailService (esa vive en
// internal/modules/booking/detail_service_test.go).
type fakeDetailBarberNamePort struct {
	name  string
	found bool
}

func (f fakeDetailBarberNamePort) Name(context.Context, string, string) (string, bool, error) {
	return f.name, f.found, nil
}

type fakeDetailStaffActorNamePort struct{}

func (fakeDetailStaffActorNamePort) Names(context.Context, string, []string) (map[string]string, error) {
	return map[string]string{}, nil
}

type fakeDetailRepository struct {
	fakeRepository

	detail      booking.AppointmentDetail
	detailFound bool
	detailErr   error

	historyItems []booking.HistoryRow
	historyNext  *booking.HistoryCursor
	historyFound bool
	historyErr   error
}

func (f *fakeDetailRepository) GetAppointmentDetail(context.Context, string, string) (booking.AppointmentDetail, bool, error) {
	if f.detailErr != nil {
		return booking.AppointmentDetail{}, false, f.detailErr
	}
	return f.detail, f.detailFound, nil
}

func (f *fakeDetailRepository) ListAppointmentHistory(
	context.Context, string, string, *booking.HistoryCursor, int,
) ([]booking.HistoryRow, *booking.HistoryCursor, bool, error) {
	if f.historyErr != nil {
		return nil, nil, false, f.historyErr
	}
	return f.historyItems, f.historyNext, f.historyFound, nil
}

func (f *fakeDetailRepository) CustomerNames(context.Context, string, []string) (map[string]string, error) {
	return map[string]string{}, nil
}

const validDetailAppointmentID = "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"

func newDetailService(repo *fakeDetailRepository, barbers fakeDetailBarberNamePort) *booking.DetailService {
	return booking.NewDetailService(repo, barbers, fakeDetailStaffActorNamePort{})
}

// ---------------------------------------------------------------------------
// GetAppointmentDetailHandler
// ---------------------------------------------------------------------------

func TestGetAppointmentDetailHandler_ValidRequest_Returns200WithDetail(t *testing.T) {
	phone := "+573001234567"
	repo := &fakeDetailRepository{
		detailFound: true,
		detail: booking.AppointmentDetail{
			ID:                       validDetailAppointmentID,
			BarberID:                 "barber-1",
			AttendeeName:             "Carlos Restrepo",
			CustomerFullName:         "Carlos Restrepo",
			CustomerPhone:            &phone,
			StartsAt:                 time.Date(2026, 8, 28, 14, 30, 0, 0, time.UTC),
			EndsAt:                   time.Date(2026, 8, 28, 15, 0, 0, 0, time.UTC),
			Status:                   booking.StatusConfirmed,
			Origin:                   booking.OriginManual,
			ServiceNameSnapshot:      "Corte clásico",
			DurationMinutesSnapshot:  30,
			PriceAmountCentsSnapshot: 2000000,
			CurrencySnapshot:         "COP",
			VersionToken:             "opaque-token",
			CreatedAt:                time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC),
		},
	}
	h := httpapi.NewGetAppointmentDetailHandler(newDetailService(repo, fakeDetailBarberNamePort{name: "Ana Gómez", found: true}))

	req := requestWithPrincipal(http.MethodGet, "/api/v1/private/appointments/"+validDetailAppointmentID, nil)
	req = httpserver.RequestWithURLParam(req, "appointmentId", validDetailAppointmentID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var got httpapi.AppointmentDetailResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got.ID != validDetailAppointmentID || got.BarberFullName != "Ana Gómez" {
		t.Fatalf("detalle inesperado: %+v", got)
	}
	if got.PriceAmount != "20000.00" || got.Currency != "COP" {
		t.Fatalf("precio/moneda inesperados: %+v", got)
	}
	if got.VersionToken != "opaque-token" {
		t.Fatalf("VersionToken = %q, want %q", got.VersionToken, "opaque-token")
	}

	// Fuera de alcance de HU-064: customerId, IDs de actor, barbershopId ni
	// updatedAt crudo. Verificado leyendo el JSON crudo, no solo el struct
	// tipado.
	var raw map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}
	for _, forbidden := range []string{"customerId", "barbershopId", "updatedAt", "actorStaffUserId", "actorCustomerId"} {
		if _, ok := raw[forbidden]; ok {
			t.Fatalf("la respuesta expuso el campo prohibido %q: %+v", forbidden, raw)
		}
	}
}

func TestGetAppointmentDetailHandler_NotFound_Returns404(t *testing.T) {
	repo := &fakeDetailRepository{detailFound: false}
	h := httpapi.NewGetAppointmentDetailHandler(newDetailService(repo, fakeDetailBarberNamePort{}))

	req := requestWithPrincipal(http.MethodGet, "/api/v1/private/appointments/"+validDetailAppointmentID, nil)
	req = httpserver.RequestWithURLParam(req, "appointmentId", validDetailAppointmentID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetAppointmentDetailHandler_NoPrincipal_Returns500(t *testing.T) {
	repo := &fakeDetailRepository{}
	h := httpapi.NewGetAppointmentDetailHandler(newDetailService(repo, fakeDetailBarberNamePort{}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/private/appointments/"+validDetailAppointmentID, nil)
	req = httpserver.RequestWithURLParam(req, "appointmentId", validDetailAppointmentID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetAppointmentDetailHandler_RepositoryError_Returns500(t *testing.T) {
	repo := &fakeDetailRepository{detailErr: apperr.Internal(nil)}
	h := httpapi.NewGetAppointmentDetailHandler(newDetailService(repo, fakeDetailBarberNamePort{}))

	req := requestWithPrincipal(http.MethodGet, "/api/v1/private/appointments/"+validDetailAppointmentID, nil)
	req = httpserver.RequestWithURLParam(req, "appointmentId", validDetailAppointmentID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// ListAppointmentHistoryHandler
// ---------------------------------------------------------------------------

func TestListAppointmentHistoryHandler_ValidRequest_Returns200WithItems(t *testing.T) {
	reason := "cliente canceló por teléfono"
	repo := &fakeDetailRepository{
		historyFound: true,
		historyItems: []booking.HistoryRow{
			{
				ID:         "history-1",
				EventType:  booking.EventTypeAppointmentCancelledByCustomer,
				ActorType:  booking.ActorTypeSystem,
				Reason:     &reason,
				OccurredAt: time.Date(2026, 8, 28, 16, 0, 0, 0, time.UTC),
			},
		},
	}
	h := httpapi.NewListAppointmentHistoryHandler(newDetailService(repo, fakeDetailBarberNamePort{}))

	req := requestWithPrincipal(http.MethodGet, "/api/v1/private/appointments/"+validDetailAppointmentID+"/history", nil)
	req = httpserver.RequestWithURLParam(req, "appointmentId", validDetailAppointmentID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got httpapi.AppointmentHistoryResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(got.Items) != 1 || got.Items[0].EventType != "appointment_cancelled_by_customer" {
		t.Fatalf("items inesperados: %+v", got.Items)
	}
	if got.NextCursor != nil {
		t.Fatalf("NextCursor = %v, want nil (última página)", got.NextCursor)
	}

	var raw map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}
	rawItems, _ := raw["items"].([]any)
	rawItem, _ := rawItems[0].(map[string]any)
	for _, forbidden := range []string{"actorStaffUserId", "actorCustomerId", "actorEmail"} {
		if _, ok := rawItem[forbidden]; ok {
			t.Fatalf("la respuesta expuso el campo prohibido %q: %+v", forbidden, rawItem)
		}
	}
}

func TestListAppointmentHistoryHandler_NextCursor_IsExposedWhenRepositoryHasMore(t *testing.T) {
	repo := &fakeDetailRepository{
		historyFound: true,
		historyNext:  &booking.HistoryCursor{OccurredAt: time.Now(), ID: "history-2"},
	}
	h := httpapi.NewListAppointmentHistoryHandler(newDetailService(repo, fakeDetailBarberNamePort{}))

	req := requestWithPrincipal(http.MethodGet, "/api/v1/private/appointments/"+validDetailAppointmentID+"/history", nil)
	req = httpserver.RequestWithURLParam(req, "appointmentId", validDetailAppointmentID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var got httpapi.AppointmentHistoryResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got.NextCursor == nil || *got.NextCursor == "" {
		t.Fatalf("NextCursor vacío, want un cursor codificado")
	}
}

func TestListAppointmentHistoryHandler_AppointmentNotFound_Returns404(t *testing.T) {
	repo := &fakeDetailRepository{historyFound: false}
	h := httpapi.NewListAppointmentHistoryHandler(newDetailService(repo, fakeDetailBarberNamePort{}))

	req := requestWithPrincipal(http.MethodGet, "/api/v1/private/appointments/"+validDetailAppointmentID+"/history", nil)
	req = httpserver.RequestWithURLParam(req, "appointmentId", validDetailAppointmentID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListAppointmentHistoryHandler_InvalidLimit_Returns400(t *testing.T) {
	repo := &fakeDetailRepository{historyFound: true}
	h := httpapi.NewListAppointmentHistoryHandler(newDetailService(repo, fakeDetailBarberNamePort{}))

	req := requestWithPrincipal(http.MethodGet, "/api/v1/private/appointments/"+validDetailAppointmentID+"/history?limit=-1", nil)
	req = httpserver.RequestWithURLParam(req, "appointmentId", validDetailAppointmentID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListAppointmentHistoryHandler_InvalidCursor_Returns400(t *testing.T) {
	repo := &fakeDetailRepository{historyFound: true}
	h := httpapi.NewListAppointmentHistoryHandler(newDetailService(repo, fakeDetailBarberNamePort{}))

	req := requestWithPrincipal(http.MethodGet, "/api/v1/private/appointments/"+validDetailAppointmentID+"/history?cursor=not-valid", nil)
	req = httpserver.RequestWithURLParam(req, "appointmentId", validDetailAppointmentID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListAppointmentHistoryHandler_NoPrincipal_Returns500(t *testing.T) {
	repo := &fakeDetailRepository{historyFound: true}
	h := httpapi.NewListAppointmentHistoryHandler(newDetailService(repo, fakeDetailBarberNamePort{}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/private/appointments/"+validDetailAppointmentID+"/history", nil)
	req = httpserver.RequestWithURLParam(req, "appointmentId", validDetailAppointmentID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}
