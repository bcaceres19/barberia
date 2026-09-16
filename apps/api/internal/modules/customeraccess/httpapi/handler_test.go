// Pruebas del adaptador HTTP en aislamiento (extracción del parámetro de
// ruta, traducción de errores, forma de la respuesta, cabeceras de
// privacidad): no PostgreSQL real, esa cobertura vive en
// internal/modules/customeraccess/postgres/repository_test.go.
package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"system-barbershop/internal/modules/customeraccess"
	"system-barbershop/internal/modules/customeraccess/httpapi"
	"system-barbershop/internal/platform/httpserver"
)

// fakeRepository es un doble en memoria de customeraccess.Repository: solo
// verifica que el handler pasa el hash recibido a Service, nunca reimplementa
// la resolución real.
type fakeRepository struct {
	view  customeraccess.AppointmentView
	found bool
	err   error
	calls []string
}

func (f *fakeRepository) ResolveAppointmentByTokenHash(_ context.Context, tokenHash string) (customeraccess.AppointmentView, bool, error) {
	f.calls = append(f.calls, tokenHash)
	return f.view, f.found, f.err
}

var _ customeraccess.Repository = (*fakeRepository)(nil)

func requestWithToken(token string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/customer/appointments/"+token, nil)
	return httpserver.RequestWithURLParam(req, "token", token)
}

func TestGetAppointmentHandler_Success_ReturnsAppointmentView(t *testing.T) {
	starts := time.Date(2027, 3, 15, 15, 0, 0, 0, time.UTC)
	repo := &fakeRepository{
		found: true,
		view: customeraccess.AppointmentView{
			BarbershopName:                 "Barbería Ejemplo",
			Timezone:                       "America/Bogota",
			AttendeeName:                   "Cliente Ejemplo",
			ServiceName:                    "Corte clásico",
			DurationMinutes:                30,
			BarberName:                     "Barbero Ejemplo",
			StartsAt:                       starts,
			EndsAt:                         starts.Add(30 * time.Minute),
			Status:                         "confirmed",
			CancellationDeadlineMinutes:    20,
			LateCancellationClientAllowed:  true,
			LateCancellationReasonRequired: true,
		},
	}
	h := httpapi.NewGetAppointmentHandler(customeraccess.NewService(repo))

	req := requestWithToken("token-en-claro")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("expected Cache-Control: no-store, got %q", rec.Header().Get("Cache-Control"))
	}
	var body httpapi.CustomerAppointmentResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body.BarbershopName != "Barbería Ejemplo" || body.BarberName != "Barbero Ejemplo" {
		t.Fatalf("unexpected response: %+v", body)
	}
	if body.Status != "confirmed" {
		t.Fatalf("Status = %q", body.Status)
	}
	if len(repo.calls) != 1 {
		t.Fatalf("expected exactly one repository call, got %d", len(repo.calls))
	}
	// CA-098-02: el repositorio nunca recibe el valor en claro, solo su hash.
	if repo.calls[0] == "token-en-claro" {
		t.Fatal("el repositorio no debe recibir el token en claro, solo su hash")
	}

	// CA-098-03: el cuerpo nunca menciona un identificador interno.
	raw := strings.ToLower(rec.Body.String())
	for _, forbidden := range []string{"\"id\"", "barbershopid", "appointmentid", "customerid"} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("CA-098-03: response body must never mention %q: %s", forbidden, raw)
		}
	}
}

func TestGetAppointmentHandler_NotFound_ReturnsUniformProblem(t *testing.T) {
	repo := &fakeRepository{found: false}
	h := httpapi.NewGetAppointmentHandler(customeraccess.NewService(repo))

	req := requestWithToken("token-inexistente")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("expected Cache-Control: no-store on error too, got %q", rec.Header().Get("Cache-Control"))
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json", ct)
	}
}

func TestGetAppointmentHandler_RepositoryError_ReturnsInternalProblem(t *testing.T) {
	repo := &fakeRepository{err: errors.New("fallo de conexión")}
	h := httpapi.NewGetAppointmentHandler(customeraccess.NewService(repo))

	req := requestWithToken("cualquier-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetAppointmentHandler_EmptyToken_ReturnsNotFoundWithoutCallingRepository(t *testing.T) {
	repo := &fakeRepository{}
	h := httpapi.NewGetAppointmentHandler(customeraccess.NewService(repo))

	req := requestWithToken("")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(repo.calls) != 0 {
		t.Fatalf("no debía consultar el repositorio con un token vacío, calls = %d", len(repo.calls))
	}
}
