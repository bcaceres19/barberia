package httpserver_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/httpserver"
	"system-barbershop/internal/platform/idempotency"
)

func TestIdempotencyKeyFromRequest_RejectsMissingHeader(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/x", nil)

	_, err := httpserver.IdempotencyKeyFromRequest(r)

	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInvalid {
		t.Fatalf("expected KindInvalid, got %v", err)
	}
}

func TestIdempotencyKeyFromRequest_AcceptsValidHeader(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/x", nil)
	r.Header.Set(httpserver.IdempotencyKeyHeader, "clave-valida-123")

	key, err := httpserver.IdempotencyKeyFromRequest(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(key) != "clave-valida-123" {
		t.Fatalf("unexpected key: %q", key)
	}
}

func TestIdempotencyFingerprint_MatchesCoreComputation(t *testing.T) {
	body := []byte(`{"a":1}`)
	r := httptest.NewRequest(http.MethodPost, "/api/v1/private/test?ignored=query", strings.NewReader(string(body)))

	got := httpserver.IdempotencyFingerprint(r, body)
	want := idempotency.ComputeFingerprint(http.MethodPost, "/api/v1/private/test", body)

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestWriteStoredResponse_ReproducesStatusContentTypeAndBodyExactly(t *testing.T) {
	stored := idempotency.StoredResponse{
		Status:      http.StatusCreated,
		ContentType: "application/json",
		Body:        `{"id":"r-1"}`,
	}

	w := httptest.NewRecorder()
	httpserver.WriteStoredResponse(w, stored)

	if w.Code != stored.Status {
		t.Fatalf("expected status %d, got %d", stored.Status, w.Code)
	}
	if got := w.Header().Get("Content-Type"); got != stored.ContentType {
		t.Fatalf("expected Content-Type %q, got %q", stored.ContentType, got)
	}
	if got := w.Body.String(); got != stored.Body {
		t.Fatalf("expected body %q, got %q", stored.Body, got)
	}
}

// --- Pruebas de integración: handler de solo-pruebas -----------------------
//
// HU-004 no publica un endpoint de producción (fuera de alcance del
// prompt); testIdempotentHandler existe ÚNICAMENTE en este archivo _test.go
// para ejercer el contrato HTTP completo —cabecera, huella, las tres
// funciones SQL dentro de una sola transacción de negocio y la traducción a
// application/problem+json— contra PostgreSQL real, sin agregar una
// operación de negocio ni tocar el router de producción.

const idempotencyTestDatabaseURL = "postgres://barberia_app@localhost:5432/barberia_test?sslmode=disable"

var idempotencyTestShop = database.BarbershopID("11111111-1111-1111-1111-111111111111")

func setupIdempotencyTestDB(t *testing.T) *database.DB {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = idempotencyTestDatabaseURL
	}

	cfg := config.Config{
		Environment:              "test",
		DatabaseMaxConns:         10,
		DatabaseMinConns:         2,
		DatabaseMaxConnLifetime:  time.Hour,
		DatabaseMaxConnIdleTime:  30 * time.Minute,
		DatabaseConnectTimeout:   5 * time.Second,
		DatabaseStatementTimeout: 10 * time.Second,
	}

	db, err := database.NewDB(config.DatabaseDSN(url), cfg)
	if err != nil {
		t.Fatalf("database.NewDB: %v", err)
	}
	return db
}

// testIdempotentHandler es el patrón documentado en apps/api/README.md: un
// futuro módulo de negocio llama IdempotencyKeyFromRequest,
// IdempotencyFingerprint, y dentro de SU PROPIA InTenantTx, Begin seguido
// de su efecto y Complete. Aquí el "efecto" es ficticio a propósito (HU-004
// no crea citas ni ningún recurso real): solo confirma una respuesta fija.
func testIdempotentHandler(db *database.DB, shop database.BarbershopID) http.HandlerFunc {
	coord := idempotency.NewSQLCoordinator()

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := httpserver.RequestIDFromContext(r.Context())

		key, err := httpserver.IdempotencyKeyFromRequest(r)
		if err != nil {
			httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			httpserver.WriteProblem(w, httpserver.Translate(apperr.Internal(err), requestID))
			return
		}
		fingerprint := httpserver.IdempotencyFingerprint(r, body)

		effectResponse := idempotency.StoredResponse{
			Status:      http.StatusCreated,
			ContentType: "application/json",
			Body:        `{"id":"test-effect-1"}`,
		}

		var decision idempotency.Decision
		txErr := db.InTenantTx(r.Context(), shop, func(ctx context.Context, q database.Queries) error {
			d, err := coord.Begin(ctx, q, shop, key, "test_effect", fingerprint, 60*time.Second)
			if err != nil {
				return err
			}
			decision = d
			if d.Outcome != idempotency.OutcomeProceed {
				return nil
			}
			_, err = coord.Complete(ctx, q, shop, key, effectResponse)
			return err
		})
		if txErr != nil {
			httpserver.WriteProblem(w, httpserver.Translate(txErr, requestID))
			return
		}

		switch decision.Outcome {
		case idempotency.OutcomeProceed:
			httpserver.WriteStoredResponse(w, effectResponse)
		case idempotency.OutcomeReplay:
			httpserver.WriteStoredResponse(w, decision.Response)
		default:
			httpserver.WriteProblem(w, httpserver.Translate(decision.AsError(), requestID))
		}
	}
}

func newIdempotencyTestServer(db *database.DB, shop database.BarbershopID) http.Handler {
	return httpserver.RequestID(testIdempotentHandler(db, shop))
}

func doIdempotentRequest(t *testing.T, h http.Handler, key, body string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, "/test-effect", strings.NewReader(body))
	if key != "" {
		r.Header.Set(httpserver.IdempotencyKeyHeader, key)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestIdempotentHandler_FirstExecution_ThenReplaysExactly(t *testing.T) {
	db := setupIdempotencyTestDB(t)
	defer db.Close()
	h := newIdempotencyTestServer(db, idempotencyTestShop)

	key := "http-first-exec-" + t.Name()
	body := `{"a":1}`

	first := doIdempotentRequest(t, h, key, body)
	if first.Code != http.StatusCreated {
		t.Fatalf("expected 201 on first execution, got %d: %s", first.Code, first.Body.String())
	}

	second := doIdempotentRequest(t, h, key, body)
	if second.Code != first.Code {
		t.Fatalf("expected the same status on replay, got %d vs %d", second.Code, first.Code)
	}
	if second.Body.String() != first.Body.String() {
		t.Fatalf("expected the exact original body on replay, got %q vs %q", second.Body.String(), first.Body.String())
	}
	if second.Header().Get("Content-Type") != first.Header().Get("Content-Type") {
		t.Fatal("expected the exact original Content-Type on replay")
	}
}

func TestIdempotentHandler_DifferentBody_Returns409ProblemJSON(t *testing.T) {
	db := setupIdempotencyTestDB(t)
	defer db.Close()
	h := newIdempotencyTestServer(db, idempotencyTestShop)

	key := "http-conflict-" + t.Name()
	doIdempotentRequest(t, h, key, `{"a":1}`)

	conflict := doIdempotentRequest(t, h, key, `{"a":2}`)
	if conflict.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", conflict.Code, conflict.Body.String())
	}
	if ct := conflict.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("expected application/problem+json, got %q", ct)
	}

	var problem httpserver.Problem
	if err := json.Unmarshal(conflict.Body.Bytes(), &problem); err != nil {
		t.Fatalf("invalid problem+json body: %v", err)
	}
	if problem.Code != "idempotency-conflict" {
		t.Fatalf("expected code idempotency-conflict, got %q", problem.Code)
	}
}

func TestIdempotentHandler_MissingKey_Returns400ProblemJSON(t *testing.T) {
	db := setupIdempotencyTestDB(t)
	defer db.Close()
	h := newIdempotencyTestServer(db, idempotencyTestShop)

	w := doIdempotentRequest(t, h, "", `{"a":1}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}

	var problem httpserver.Problem
	if err := json.Unmarshal(w.Body.Bytes(), &problem); err != nil {
		t.Fatalf("invalid problem+json body: %v", err)
	}
	if problem.Code != "invalid-request" {
		t.Fatalf("expected code invalid-request, got %q", problem.Code)
	}
}

// TestIdempotentHandler_ConcurrentRequests_SecondGetsLockedProblem cubre
// DEC-043 en la capa HTTP completa: la segunda solicitud concurrente con la
// misma clave recibe 409 application/problem+json de inmediato, sin esperar
// a que la primera termine.
func TestIdempotentHandler_ConcurrentRequests_SecondGetsLockedProblem(t *testing.T) {
	db := setupIdempotencyTestDB(t)
	defer db.Close()

	key := "http-locked-" + t.Name()
	body := `{"a":1}`

	aReady := make(chan struct{})
	bDone := make(chan struct{})
	holdingHandler := func(w http.ResponseWriter, r *http.Request) {
		reqKey, err := httpserver.IdempotencyKeyFromRequest(r)
		if err != nil {
			t.Fatalf("unexpected key error: %v", err)
		}
		fp := httpserver.IdempotencyFingerprint(r, []byte(body))
		coord := idempotency.NewSQLCoordinator()

		err = db.InTenantTx(r.Context(), idempotencyTestShop, func(ctx context.Context, q database.Queries) error {
			decision, err := coord.Begin(ctx, q, idempotencyTestShop, reqKey, "test_effect", fp, 60*time.Second)
			if err != nil {
				return err
			}
			if decision.Outcome != idempotency.OutcomeProceed {
				t.Errorf("expected session A to proceed, got %s", decision.Outcome)
			}
			close(aReady)
			<-bDone
			return nil
		})
		if err != nil {
			t.Errorf("session A InTenantTx: %v", err)
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		httpserver.RequestID(http.HandlerFunc(holdingHandler)).ServeHTTP(
			httptest.NewRecorder(), doIdempotentRequestObject(key, body))
	}()

	<-aReady
	h := newIdempotencyTestServer(db, idempotencyTestShop)
	locked := doIdempotentRequest(t, h, key, body)
	close(bDone)
	wg.Wait()

	if locked.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", locked.Code, locked.Body.String())
	}
	var problem httpserver.Problem
	if err := json.Unmarshal(locked.Body.Bytes(), &problem); err != nil {
		t.Fatalf("invalid problem+json body: %v", err)
	}
	if problem.Code != "idempotency-locked" {
		t.Fatalf("expected code idempotency-locked, got %q", problem.Code)
	}
}

func doIdempotentRequestObject(key, body string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/test-effect", strings.NewReader(body))
	r.Header.Set(httpserver.IdempotencyKeyHeader, key)
	return r
}
