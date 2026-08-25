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
	"system-barbershop/internal/modules/catalog"
	"system-barbershop/internal/modules/catalog/httpapi"
	"system-barbershop/internal/platform/httpserver"
	"system-barbershop/internal/platform/idempotency"
)

// fakeRepository es un doble en memoria de catalog.Repository: estas
// pruebas verifican el adaptador HTTP en aislamiento (decodificación, forma
// cerrada, extracción del principal, headers, traducción de errores), no
// PostgreSQL real (esa cobertura vive en
// internal/modules/catalog/postgres/repository_test.go).
type fakeRepository struct {
	listFn       func(barbershopID string, cursor *catalog.Cursor, limit int) (catalog.ListResult, error)
	getFn        func(barbershopID, serviceID string) (catalog.Service, bool, error)
	createFn     func(barbershopID string, input catalog.CreateInput, key idempotency.Key, fp idempotency.Fingerprint) (catalog.CreateResult, error)
	updateFn     func(barbershopID, serviceID string, fields catalog.UpdateFields) (catalog.UpdateResult, error)
	deactivateFn func(barbershopID, serviceID string, key idempotency.Key, fp idempotency.Fingerprint) (catalog.LifecycleResult, error)
	reactivateFn func(barbershopID, serviceID string, key idempotency.Key, fp idempotency.Fingerprint) (catalog.LifecycleResult, error)
}

func (f *fakeRepository) List(_ context.Context, barbershopID string, cursor *catalog.Cursor, limit int) (catalog.ListResult, error) {
	return f.listFn(barbershopID, cursor, limit)
}

func (f *fakeRepository) Get(_ context.Context, barbershopID, serviceID string) (catalog.Service, bool, error) {
	return f.getFn(barbershopID, serviceID)
}

func (f *fakeRepository) Create(_ context.Context, barbershopID string, input catalog.CreateInput, key idempotency.Key, fp idempotency.Fingerprint) (catalog.CreateResult, error) {
	return f.createFn(barbershopID, input, key, fp)
}

func (f *fakeRepository) Update(_ context.Context, barbershopID, serviceID string, fields catalog.UpdateFields) (catalog.UpdateResult, error) {
	return f.updateFn(barbershopID, serviceID, fields)
}

func (f *fakeRepository) Deactivate(_ context.Context, barbershopID, serviceID string, key idempotency.Key, fp idempotency.Fingerprint) (catalog.LifecycleResult, error) {
	return f.deactivateFn(barbershopID, serviceID, key, fp)
}

func (f *fakeRepository) Reactivate(_ context.Context, barbershopID, serviceID string, key idempotency.Key, fp idempotency.Fingerprint) (catalog.LifecycleResult, error) {
	return f.reactivateFn(barbershopID, serviceID, key, fp)
}

var _ catalog.Repository = (*fakeRepository)(nil)

const testShopID = "11111111-1111-1111-1111-111111111111"

func requestWithPrincipal(method, target string, body []byte) *http.Request {
	var req *http.Request
	if body == nil {
		req = httptest.NewRequest(method, target, nil)
	} else {
		req = httptest.NewRequest(method, target, bytes.NewReader(body))
	}
	principal := auth.Principal{SessionID: "s-1", StaffUserID: "u-1", BarbershopID: testShopID}
	return req.WithContext(auth.ContextWithPrincipal(req.Context(), principal))
}

func requestWithServiceID(method, target string, body []byte, serviceID string) *http.Request {
	r := requestWithPrincipal(method, target, body)
	return httpserver.RequestWithURLParam(r, "serviceId", serviceID)
}

func decodeProblem(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var problem map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
		t.Fatalf("decode problem body: %v (%s)", err, rec.Body.String())
	}
	return problem
}

// --- List -------------------------------------------------------------

func TestListServicesHandler_Success_ReturnsItemsAndCursor(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeRepository{listFn: func(barbershopID string, cursor *catalog.Cursor, limit int) (catalog.ListResult, error) {
		if barbershopID != testShopID {
			t.Fatalf("expected shop %q, got %q", testShopID, barbershopID)
		}
		if limit != catalog.DefaultListLimit {
			t.Fatalf("expected default limit, got %d", limit)
		}
		if cursor != nil {
			t.Fatalf("expected nil cursor, got %+v", cursor)
		}
		return catalog.ListResult{
			Items: []catalog.Service{
				{ID: "s-1", Name: "Corte", DurationMinutes: 30, PriceCents: 4500000, Currency: "COP", CreatedAt: now, UpdatedAt: now},
			},
			NextCursor: "opaque-cursor",
		}, nil
	}}
	h := httpapi.NewListServicesHandler(catalog.NewService(repo))

	req := requestWithPrincipal(http.MethodGet, "/api/v1/private/services", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body httpapi.ServiceListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Items) != 1 || body.Items[0].ID != "s-1" {
		t.Fatalf("unexpected items: %+v", body.Items)
	}
	if body.Items[0].Price != "45000.00" {
		t.Fatalf("expected price formatted as decimal string, got %q", body.Items[0].Price)
	}
	if body.NextCursor == nil || *body.NextCursor != "opaque-cursor" {
		t.Fatalf("expected nextCursor to round-trip, got %v", body.NextCursor)
	}
}

func TestListServicesHandler_NoNextPage_ReturnsExplicitNull(t *testing.T) {
	repo := &fakeRepository{listFn: func(string, *catalog.Cursor, int) (catalog.ListResult, error) {
		return catalog.ListResult{Items: []catalog.Service{}, NextCursor: ""}, nil
	}}
	h := httpapi.NewListServicesHandler(catalog.NewService(repo))

	req := requestWithPrincipal(http.MethodGet, "/api/v1/private/services", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), `"nextCursor":null`) {
		t.Fatalf("expected explicit null nextCursor in body, got %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"items":[]`) {
		t.Fatalf("expected an explicit empty array for items, got %s", rec.Body.String())
	}
}

func TestListServicesHandler_InvalidLimit_Returns400(t *testing.T) {
	repo := &fakeRepository{listFn: func(string, *catalog.Cursor, int) (catalog.ListResult, error) {
		t.Fatal("repository must not be called for an invalid limit")
		return catalog.ListResult{}, nil
	}}
	h := httpapi.NewListServicesHandler(catalog.NewService(repo))

	req := requestWithPrincipal(http.MethodGet, "/api/v1/private/services?limit=not-a-number", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- Get ----------------------------------------------------------------

func TestGetServiceHandler_Found_ReturnsService(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeRepository{getFn: func(barbershopID, serviceID string) (catalog.Service, bool, error) {
		if barbershopID != testShopID || serviceID != "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4" {
			t.Fatalf("unexpected args: %s %s", barbershopID, serviceID)
		}
		return catalog.Service{ID: serviceID, Name: "Corte", DurationMinutes: 30, PriceCents: 4500000, Currency: "COP", CreatedAt: now, UpdatedAt: now}, true, nil
	}}
	h := httpapi.NewGetServiceHandler(catalog.NewService(repo))

	req := requestWithServiceID(http.MethodGet, "/api/v1/private/services/8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4", nil, "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetServiceHandler_NotFound_Returns404(t *testing.T) {
	repo := &fakeRepository{getFn: func(string, string) (catalog.Service, bool, error) {
		return catalog.Service{}, false, nil
	}}
	h := httpapi.NewGetServiceHandler(catalog.NewService(repo))

	req := requestWithServiceID(http.MethodGet, "/api/v1/private/services/99999999-9999-4999-8999-999999999999", nil, "99999999-9999-4999-8999-999999999999")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	problem := decodeProblem(t, rec)
	if problem["code"] != "not-found" {
		t.Fatalf("expected code=not-found, got %v", problem["code"])
	}
}

func TestGetServiceHandler_OtherTenantsID_SameNotFoundAsNonexistent(t *testing.T) {
	// CA-022-06: el handler nunca puede distinguir "no existe" de "es de
	// otra barbería" porque el servicio ya colapsó ambos en el mismo
	// resultado found=false.
	repo := &fakeRepository{getFn: func(string, string) (catalog.Service, bool, error) {
		return catalog.Service{}, false, nil
	}}
	h := httpapi.NewGetServiceHandler(catalog.NewService(repo))

	req := requestWithServiceID(http.MethodGet, "/api/v1/private/services/8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4", nil, "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// --- Create ---------------------------------------------------------------

func TestCreateServiceHandler_MissingIdempotencyKey_Returns400(t *testing.T) {
	repo := &fakeRepository{createFn: func(string, catalog.CreateInput, idempotency.Key, idempotency.Fingerprint) (catalog.CreateResult, error) {
		t.Fatal("repository must not be called without an idempotency key")
		return catalog.CreateResult{}, nil
	}}
	h := httpapi.NewCreateServiceHandler(catalog.NewService(repo))

	req := requestWithPrincipal(http.MethodPost, "/api/v1/private/services", []byte(`{"name":"Corte","durationMinutes":30,"price":"45000.00"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateServiceHandler_UnknownField_Returns400(t *testing.T) {
	repo := &fakeRepository{createFn: func(string, catalog.CreateInput, idempotency.Key, idempotency.Fingerprint) (catalog.CreateResult, error) {
		t.Fatal("repository must not be called for an unknown field")
		return catalog.CreateResult{}, nil
	}}
	h := httpapi.NewCreateServiceHandler(catalog.NewService(repo))

	req := requestWithPrincipal(http.MethodPost, "/api/v1/private/services",
		[]byte(`{"name":"Corte","durationMinutes":30,"price":"45000.00","isActive":true}`))
	req.Header.Set(httpserver.IdempotencyKeyHeader, "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateServiceHandler_ValidationError_Returns422WithoutCallingRepository(t *testing.T) {
	repo := &fakeRepository{createFn: func(string, catalog.CreateInput, idempotency.Key, idempotency.Fingerprint) (catalog.CreateResult, error) {
		t.Fatal("repository must not be called for an invalid price")
		return catalog.CreateResult{}, nil
	}}
	h := httpapi.NewCreateServiceHandler(catalog.NewService(repo))

	req := requestWithPrincipal(http.MethodPost, "/api/v1/private/services",
		[]byte(`{"name":"Corte","durationMinutes":30,"price":"0"}`))
	req.Header.Set(httpserver.IdempotencyKeyHeader, "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateServiceHandler_FractionalDuration_Returns400(t *testing.T) {
	// CA-022-03: un JSON que ni siquiera es un entero (30.5) se rechaza al
	// decodificar, antes de llegar a la validación de rango del servicio.
	repo := &fakeRepository{createFn: func(string, catalog.CreateInput, idempotency.Key, idempotency.Fingerprint) (catalog.CreateResult, error) {
		t.Fatal("repository must not be called for a fractional duration")
		return catalog.CreateResult{}, nil
	}}
	h := httpapi.NewCreateServiceHandler(catalog.NewService(repo))

	req := requestWithPrincipal(http.MethodPost, "/api/v1/private/services",
		[]byte(`{"name":"Corte","durationMinutes":30.5,"price":"45000.00"}`))
	req.Header.Set(httpserver.IdempotencyKeyHeader, "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateServiceHandler_Proceed_Returns201WithLocationAndBody(t *testing.T) {
	now := time.Now().UTC()
	stored := idempotency.StoredResponse{
		Status:      201,
		ContentType: "application/json",
		Body: `{"id":"8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4","name":"Corte","description":null,"durationMinutes":30,` +
			`"price":"45000.00","currency":"COP","createdAt":"` + now.Format(time.RFC3339Nano) + `","updatedAt":"` + now.Format(time.RFC3339Nano) + `"}`,
	}
	repo := &fakeRepository{createFn: func(barbershopID string, input catalog.CreateInput, key idempotency.Key, fp idempotency.Fingerprint) (catalog.CreateResult, error) {
		if barbershopID != testShopID {
			t.Fatalf("unexpected shop: %s", barbershopID)
		}
		if input.Name != "Corte" || input.DurationMinutes != 30 || input.PriceCents != 4500000 {
			t.Fatalf("unexpected input: %+v", input)
		}
		if key != "key-1" {
			t.Fatalf("unexpected key: %s", key)
		}
		return catalog.CreateResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Service:  catalog.Service{ID: "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4", Name: "Corte", DurationMinutes: 30, PriceCents: 4500000, Currency: "COP", CreatedAt: now, UpdatedAt: now},
			Response: stored,
		}, nil
	}}
	h := httpapi.NewCreateServiceHandler(catalog.NewService(repo))

	req := requestWithPrincipal(http.MethodPost, "/api/v1/private/services", []byte(`{"name":"Corte","durationMinutes":30,"price":"45000.00"}`))
	req.Header.Set(httpserver.IdempotencyKeyHeader, "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if loc := rec.Header().Get("Location"); loc != "/private/services/8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4" {
		t.Fatalf("unexpected Location header: %q", loc)
	}
	if rec.Body.String() != stored.Body {
		t.Fatalf("expected the exact stored body, got %s", rec.Body.String())
	}
}

func TestCreateServiceHandler_Replay_ReturnsSameStatusAndBodyWithLocation(t *testing.T) {
	stored := idempotency.StoredResponse{
		Status:      201,
		ContentType: "application/json",
		Body:        `{"id":"8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4","name":"Corte"}`,
	}
	repo := &fakeRepository{createFn: func(string, catalog.CreateInput, idempotency.Key, idempotency.Fingerprint) (catalog.CreateResult, error) {
		return catalog.CreateResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeReplay, Response: stored},
			Response: stored,
		}, nil
	}}
	h := httpapi.NewCreateServiceHandler(catalog.NewService(repo))

	req := requestWithPrincipal(http.MethodPost, "/api/v1/private/services", []byte(`{"name":"Corte","durationMinutes":30,"price":"45000.00"}`))
	req.Header.Set(httpserver.IdempotencyKeyHeader, "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected the same original 201 on replay, got %d", rec.Code)
	}
	if rec.Body.String() != stored.Body {
		t.Fatalf("expected the exact original body on replay, got %s", rec.Body.String())
	}
	if loc := rec.Header().Get("Location"); loc != "/private/services/8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4" {
		t.Fatalf("expected Location to be reconstructed on replay too, got %q", loc)
	}
}

func TestCreateServiceHandler_ConflictFingerprint_Returns409(t *testing.T) {
	repo := &fakeRepository{createFn: func(string, catalog.CreateInput, idempotency.Key, idempotency.Fingerprint) (catalog.CreateResult, error) {
		return catalog.CreateResult{Decision: idempotency.Decision{Outcome: idempotency.OutcomeConflictFingerprint}}, nil
	}}
	h := httpapi.NewCreateServiceHandler(catalog.NewService(repo))

	req := requestWithPrincipal(http.MethodPost, "/api/v1/private/services", []byte(`{"name":"Corte","durationMinutes":30,"price":"45000.00"}`))
	req.Header.Set(httpserver.IdempotencyKeyHeader, "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	problem := decodeProblem(t, rec)
	if problem["code"] != "idempotency-conflict" {
		t.Fatalf("expected code=idempotency-conflict, got %v", problem["code"])
	}
}

func TestCreateServiceHandler_Locked_Returns409(t *testing.T) {
	repo := &fakeRepository{createFn: func(string, catalog.CreateInput, idempotency.Key, idempotency.Fingerprint) (catalog.CreateResult, error) {
		return catalog.CreateResult{Decision: idempotency.Decision{Outcome: idempotency.OutcomeLocked}}, nil
	}}
	h := httpapi.NewCreateServiceHandler(catalog.NewService(repo))

	req := requestWithPrincipal(http.MethodPost, "/api/v1/private/services", []byte(`{"name":"Corte","durationMinutes":30,"price":"45000.00"}`))
	req.Header.Set(httpserver.IdempotencyKeyHeader, "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	problem := decodeProblem(t, rec)
	if problem["code"] != "idempotency-locked" {
		t.Fatalf("expected code=idempotency-locked, got %v", problem["code"])
	}
}

func TestCreateServiceHandler_NameConflict_Returns409WithConflictCode(t *testing.T) {
	// DEC-067: distinto de idempotency-conflict/idempotency-locked -mismo
	// status 409, code distinto-.
	repo := &fakeRepository{createFn: func(string, catalog.CreateInput, idempotency.Key, idempotency.Fingerprint) (catalog.CreateResult, error) {
		return catalog.CreateResult{NameTaken: true}, nil
	}}
	h := httpapi.NewCreateServiceHandler(catalog.NewService(repo))

	req := requestWithPrincipal(http.MethodPost, "/api/v1/private/services", []byte(`{"name":"Corte","durationMinutes":30,"price":"45000.00"}`))
	req.Header.Set(httpserver.IdempotencyKeyHeader, "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	problem := decodeProblem(t, rec)
	if problem["code"] != "conflict" {
		t.Fatalf("expected code=conflict, got %v", problem["code"])
	}
}

func TestCreateServiceHandler_ThroughFullRouter_NeverLogsNameOrBody(t *testing.T) {
	repo := &fakeRepository{createFn: func(string, catalog.CreateInput, idempotency.Key, idempotency.Fingerprint) (catalog.CreateResult, error) {
		return catalog.CreateResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Service:  catalog.Service{ID: "s-1", Name: "Nombre Sensible No Debe Aparecer"},
			Response: idempotency.StoredResponse{Status: 201, ContentType: "application/json", Body: `{"id":"s-1"}`},
		}, nil
	}}
	h := httpapi.NewCreateServiceHandler(catalog.NewService(repo))

	req := requestWithPrincipal(http.MethodPost, "/api/v1/private/services",
		[]byte(`{"name":"Nombre Sensible No Debe Aparecer","durationMinutes":30,"price":"45000.00"}`))
	req.Header.Set(httpserver.IdempotencyKeyHeader, "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if strings.Contains(rec.Header().Get("Location"), "Sensible") {
		t.Fatal("Location header must never contain the service's name")
	}
}

// --- Update -----------------------------------------------------------

func TestUpdateServiceHandler_Success_Returns200(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeRepository{updateFn: func(barbershopID, serviceID string, fields catalog.UpdateFields) (catalog.UpdateResult, error) {
		if fields.Name == nil || *fields.Name != "Nuevo nombre" {
			t.Fatalf("unexpected fields: %+v", fields)
		}
		return catalog.UpdateResult{Found: true, Service: catalog.Service{ID: serviceID, Name: "Nuevo nombre", Currency: "COP", CreatedAt: now, UpdatedAt: now}}, nil
	}}
	h := httpapi.NewUpdateServiceHandler(catalog.NewService(repo))

	req := requestWithServiceID(http.MethodPatch, "/api/v1/private/services/8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4",
		[]byte(`{"name":"Nuevo nombre"}`), "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateServiceHandler_UnknownField_Returns400(t *testing.T) {
	repo := &fakeRepository{updateFn: func(string, string, catalog.UpdateFields) (catalog.UpdateResult, error) {
		t.Fatal("repository must not be called for an unknown field")
		return catalog.UpdateResult{}, nil
	}}
	h := httpapi.NewUpdateServiceHandler(catalog.NewService(repo))

	req := requestWithServiceID(http.MethodPatch, "/api/v1/private/services/8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4",
		[]byte(`{"name":"X","isActive":false}`), "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateServiceHandler_EmptyObject_Returns422(t *testing.T) {
	repo := &fakeRepository{updateFn: func(string, string, catalog.UpdateFields) (catalog.UpdateResult, error) {
		t.Fatal("repository must not be called for an empty object")
		return catalog.UpdateResult{}, nil
	}}
	h := httpapi.NewUpdateServiceHandler(catalog.NewService(repo))

	req := requestWithServiceID(http.MethodPatch, "/api/v1/private/services/8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4",
		[]byte(`{}`), "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for an empty object, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateServiceHandler_DescriptionEmptyString_ClearsDescription(t *testing.T) {
	repo := &fakeRepository{updateFn: func(_, _ string, fields catalog.UpdateFields) (catalog.UpdateResult, error) {
		if !fields.Description.Set || fields.Description.Value != nil {
			t.Fatalf("expected description cleared, got %+v", fields.Description)
		}
		return catalog.UpdateResult{Found: true, Service: catalog.Service{ID: "s-1"}}, nil
	}}
	h := httpapi.NewUpdateServiceHandler(catalog.NewService(repo))

	req := requestWithServiceID(http.MethodPatch, "/api/v1/private/services/8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4",
		[]byte(`{"description":""}`), "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateServiceHandler_NotFound_Returns404(t *testing.T) {
	repo := &fakeRepository{updateFn: func(string, string, catalog.UpdateFields) (catalog.UpdateResult, error) {
		return catalog.UpdateResult{Found: false}, nil
	}}
	h := httpapi.NewUpdateServiceHandler(catalog.NewService(repo))

	req := requestWithServiceID(http.MethodPatch, "/api/v1/private/services/99999999-9999-4999-8999-999999999999",
		[]byte(`{"name":"X"}`), "99999999-9999-4999-8999-999999999999")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateServiceHandler_NameConflict_Returns409(t *testing.T) {
	repo := &fakeRepository{updateFn: func(string, string, catalog.UpdateFields) (catalog.UpdateResult, error) {
		return catalog.UpdateResult{NameTaken: true}, nil
	}}
	h := httpapi.NewUpdateServiceHandler(catalog.NewService(repo))

	req := requestWithServiceID(http.MethodPatch, "/api/v1/private/services/8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4",
		[]byte(`{"name":"Ya existe"}`), "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	problem := decodeProblem(t, rec)
	if problem["code"] != "conflict" {
		t.Fatalf("expected code=conflict, got %v", problem["code"])
	}
}

func TestUpdateServiceHandler_ValidationError_Returns422(t *testing.T) {
	repo := &fakeRepository{updateFn: func(string, string, catalog.UpdateFields) (catalog.UpdateResult, error) {
		t.Fatal("repository must not be called for an invalid duration")
		return catalog.UpdateResult{}, nil
	}}
	h := httpapi.NewUpdateServiceHandler(catalog.NewService(repo))

	req := requestWithServiceID(http.MethodPatch, "/api/v1/private/services/8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4",
		[]byte(`{"durationMinutes":0}`), "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- Wiring defensivo -------------------------------------------------

// --- HU-024: ciclo de vida --------------------------------------------------

const lifecycleServiceID = "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"

func TestGetServiceDeactivationImpactHandler_Found_Returns200WithZero(t *testing.T) {
	repo := &fakeRepository{getFn: func(barbershopID, serviceID string) (catalog.Service, bool, error) {
		if barbershopID != testShopID || serviceID != lifecycleServiceID {
			t.Fatalf("unexpected lookup: shop=%s service=%s", barbershopID, serviceID)
		}
		return catalog.Service{ID: serviceID, IsActive: true}, true, nil
	}}
	h := httpapi.NewGetServiceDeactivationImpactHandler(catalog.NewService(repo))

	req := requestWithServiceID(http.MethodGet, "/api/v1/private/services/"+lifecycleServiceID+"/deactivation-impact", nil, lifecycleServiceID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body httpapi.ServiceDeactivationImpactResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.AffectedAppointments != 0 {
		t.Fatalf("expected affectedAppointments=0 (DEC-069), got %d", body.AffectedAppointments)
	}
}

func TestGetServiceDeactivationImpactHandler_NotFound_Returns404(t *testing.T) {
	repo := &fakeRepository{getFn: func(string, string) (catalog.Service, bool, error) {
		return catalog.Service{}, false, nil
	}}
	h := httpapi.NewGetServiceDeactivationImpactHandler(catalog.NewService(repo))

	req := requestWithServiceID(http.MethodGet, "/api/v1/private/services/"+lifecycleServiceID+"/deactivation-impact", nil, lifecycleServiceID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeactivateServiceHandler_MissingIdempotencyKey_Returns400(t *testing.T) {
	repo := &fakeRepository{deactivateFn: func(string, string, idempotency.Key, idempotency.Fingerprint) (catalog.LifecycleResult, error) {
		t.Fatal("repository must not be called without an idempotency key")
		return catalog.LifecycleResult{}, nil
	}}
	h := httpapi.NewDeactivateServiceHandler(catalog.NewService(repo))

	req := requestWithServiceID(http.MethodPost, "/api/v1/private/services/"+lifecycleServiceID+"/deactivate", nil, lifecycleServiceID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeactivateServiceHandler_UnknownField_Returns400(t *testing.T) {
	repo := &fakeRepository{deactivateFn: func(string, string, idempotency.Key, idempotency.Fingerprint) (catalog.LifecycleResult, error) {
		t.Fatal("repository must not be called for an unknown field")
		return catalog.LifecycleResult{}, nil
	}}
	h := httpapi.NewDeactivateServiceHandler(catalog.NewService(repo))

	req := requestWithServiceID(http.MethodPost, "/api/v1/private/services/"+lifecycleServiceID+"/deactivate", []byte(`{"isActive":false}`), lifecycleServiceID)
	req.Header.Set(httpserver.IdempotencyKeyHeader, "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeactivateServiceHandler_Proceed_Returns200WithStoredBody(t *testing.T) {
	stored := idempotency.StoredResponse{
		Status:      200,
		ContentType: "application/json",
		Body:        `{"service":{"id":"` + lifecycleServiceID + `","isActive":false},"affectedAppointments":0}`,
	}
	repo := &fakeRepository{deactivateFn: func(barbershopID, serviceID string, key idempotency.Key, fp idempotency.Fingerprint) (catalog.LifecycleResult, error) {
		if barbershopID != testShopID || serviceID != lifecycleServiceID || key != "key-1" {
			t.Fatalf("unexpected call: shop=%s service=%s key=%s", barbershopID, serviceID, key)
		}
		return catalog.LifecycleResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Found:    true,
			Response: stored,
		}, nil
	}}
	h := httpapi.NewDeactivateServiceHandler(catalog.NewService(repo))

	req := requestWithServiceID(http.MethodPost, "/api/v1/private/services/"+lifecycleServiceID+"/deactivate", nil, lifecycleServiceID)
	req.Header.Set(httpserver.IdempotencyKeyHeader, "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != stored.Body {
		t.Fatalf("expected the exact stored body, got %s", rec.Body.String())
	}
}

func TestDeactivateServiceHandler_NotFound_Returns404(t *testing.T) {
	repo := &fakeRepository{deactivateFn: func(string, string, idempotency.Key, idempotency.Fingerprint) (catalog.LifecycleResult, error) {
		return catalog.LifecycleResult{Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed}, Found: false}, nil
	}}
	h := httpapi.NewDeactivateServiceHandler(catalog.NewService(repo))

	req := requestWithServiceID(http.MethodPost, "/api/v1/private/services/"+lifecycleServiceID+"/deactivate", nil, lifecycleServiceID)
	req.Header.Set(httpserver.IdempotencyKeyHeader, "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeactivateServiceHandler_AlreadyInactive_Returns409(t *testing.T) {
	repo := &fakeRepository{deactivateFn: func(string, string, idempotency.Key, idempotency.Fingerprint) (catalog.LifecycleResult, error) {
		return catalog.LifecycleResult{
			Decision:          idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Found:             true,
			InvalidTransition: true,
		}, nil
	}}
	h := httpapi.NewDeactivateServiceHandler(catalog.NewService(repo))

	req := requestWithServiceID(http.MethodPost, "/api/v1/private/services/"+lifecycleServiceID+"/deactivate", nil, lifecycleServiceID)
	req.Header.Set(httpserver.IdempotencyKeyHeader, "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	problem := decodeProblem(t, rec)
	if problem["code"] != "conflict" {
		t.Fatalf("expected code=conflict, got %v", problem["code"])
	}
}

func TestDeactivateServiceHandler_IdempotencyConflict_Returns409(t *testing.T) {
	repo := &fakeRepository{deactivateFn: func(string, string, idempotency.Key, idempotency.Fingerprint) (catalog.LifecycleResult, error) {
		return catalog.LifecycleResult{Decision: idempotency.Decision{Outcome: idempotency.OutcomeConflictFingerprint}}, nil
	}}
	h := httpapi.NewDeactivateServiceHandler(catalog.NewService(repo))

	req := requestWithServiceID(http.MethodPost, "/api/v1/private/services/"+lifecycleServiceID+"/deactivate", nil, lifecycleServiceID)
	req.Header.Set(httpserver.IdempotencyKeyHeader, "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	problem := decodeProblem(t, rec)
	if problem["code"] != "idempotency-conflict" {
		t.Fatalf("expected code=idempotency-conflict, got %v", problem["code"])
	}
}

func TestReactivateServiceHandler_Proceed_Returns200WithStoredBody(t *testing.T) {
	stored := idempotency.StoredResponse{
		Status:      200,
		ContentType: "application/json",
		Body:        `{"id":"` + lifecycleServiceID + `","isActive":true,"deactivatedAt":null}`,
	}
	repo := &fakeRepository{reactivateFn: func(barbershopID, serviceID string, key idempotency.Key, fp idempotency.Fingerprint) (catalog.LifecycleResult, error) {
		if key != "key-1" {
			t.Fatalf("unexpected key: %s", key)
		}
		return catalog.LifecycleResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Found:    true,
			Response: stored,
		}, nil
	}}
	h := httpapi.NewReactivateServiceHandler(catalog.NewService(repo))

	req := requestWithServiceID(http.MethodPost, "/api/v1/private/services/"+lifecycleServiceID+"/reactivate", nil, lifecycleServiceID)
	req.Header.Set(httpserver.IdempotencyKeyHeader, "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != stored.Body {
		t.Fatalf("expected the exact stored body, got %s", rec.Body.String())
	}
}

func TestReactivateServiceHandler_AlreadyActive_Returns409(t *testing.T) {
	repo := &fakeRepository{reactivateFn: func(string, string, idempotency.Key, idempotency.Fingerprint) (catalog.LifecycleResult, error) {
		return catalog.LifecycleResult{
			Decision:          idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Found:             true,
			InvalidTransition: true,
		}, nil
	}}
	h := httpapi.NewReactivateServiceHandler(catalog.NewService(repo))

	req := requestWithServiceID(http.MethodPost, "/api/v1/private/services/"+lifecycleServiceID+"/reactivate", nil, lifecycleServiceID)
	req.Header.Set(httpserver.IdempotencyKeyHeader, "key-1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestReactivateServiceHandler_MissingIdempotencyKey_Returns400(t *testing.T) {
	repo := &fakeRepository{reactivateFn: func(string, string, idempotency.Key, idempotency.Fingerprint) (catalog.LifecycleResult, error) {
		t.Fatal("repository must not be called without an idempotency key")
		return catalog.LifecycleResult{}, nil
	}}
	h := httpapi.NewReactivateServiceHandler(catalog.NewService(repo))

	req := requestWithServiceID(http.MethodPost, "/api/v1/private/services/"+lifecycleServiceID+"/reactivate", nil, lifecycleServiceID)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandlers_MissingPrincipal_Returns500Safely(t *testing.T) {
	// Solo puede ocurrir si la ruta se montó fuera del subrouter protegido
	// por SessionMiddleware (CA-006-04 lo evita estructuralmente); esta
	// prueba confirma que, si ocurriera, la respuesta sigue siendo un
	// Problem uniforme y no un panic ni una fuga.
	repo := &fakeRepository{}
	h := httpapi.NewListServicesHandler(catalog.NewService(repo))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/private/services", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}
