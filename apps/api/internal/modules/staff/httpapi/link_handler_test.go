package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"system-barbershop/internal/modules/staff"
	"system-barbershop/internal/modules/staff/httpapi"
)

const myBarberID = "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"

func TestGetMyBarberHandler_UsesThePrincipalAndReturnsTheBarber(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeRepository{getLinkedFn: func(shop, user string) (staff.Barber, bool, error) {
		if shop != testShopID || user != "u-1" {
			t.Fatalf("debe consultar con el principal de la sesión, recibió %q/%q", shop, user)
		}
		return staff.Barber{ID: myBarberID, FullName: "Carlos", CreatedAt: now, UpdatedAt: now}, true, nil
	}}
	rec := httptest.NewRecorder()
	httpapi.NewGetMyBarberHandler(staff.NewService(repo)).ServeHTTP(rec,
		requestWithPrincipal(http.MethodGet, "/api/v1/private/me/barber", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["id"] != myBarberID {
		t.Fatalf("unexpected body: %v", body)
	}
	if _, leaked := body["staffUserId"]; leaked {
		t.Fatal("la respuesta nunca debe exponer staffUserId")
	}
}

func TestGetMyBarberHandler_WithoutLink_Is404(t *testing.T) {
	repo := &fakeRepository{getLinkedFn: func(string, string) (staff.Barber, bool, error) { return staff.Barber{}, false, nil }}
	rec := httptest.NewRecorder()
	httpapi.NewGetMyBarberHandler(staff.NewService(repo)).ServeHTTP(rec,
		requestWithPrincipal(http.MethodGet, "/api/v1/private/me/barber", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestLinkMyBarberHandler_Success(t *testing.T) {
	repo := &fakeRepository{linkFn: func(shop, user, barber string) (staff.LinkResult, error) {
		if shop != testShopID || user != "u-1" || barber != myBarberID {
			t.Fatalf("argumentos inesperados %q/%q/%q", shop, user, barber)
		}
		return staff.LinkResult{Found: true, Barber: staff.Barber{ID: myBarberID, FullName: "Carlos"}}, nil
	}}
	rec := httptest.NewRecorder()
	httpapi.NewLinkMyBarberHandler(staff.NewService(repo)).ServeHTTP(rec,
		requestWithPrincipal(http.MethodPut, "/api/v1/private/me/barber", []byte(`{"barberId":"`+myBarberID+`"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLinkMyBarberHandler_RejectsAnyoneElsesUserID(t *testing.T) {
	repo := &fakeRepository{linkFn: func(string, string, string) (staff.LinkResult, error) {
		t.Fatal("un cuerpo con staffUserId nunca debe llegar al repositorio")
		return staff.LinkResult{}, nil
	}}
	for name, body := range map[string]string{
		"staffUserId":    `{"barberId":"` + myBarberID + `","staffUserId":"otro-usuario"}`,
		"campo extra":    `{"barberId":"` + myBarberID + `","x":1}`,
		"JSON inválido":  `{`,
		"dos documentos": `{"barberId":"` + myBarberID + `"}{}`,
	} {
		t.Run(name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			httpapi.NewLinkMyBarberHandler(staff.NewService(repo)).ServeHTTP(rec,
				requestWithPrincipal(http.MethodPut, "/api/v1/private/me/barber", []byte(body)))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestLinkMyBarberHandler_TakenBarber_Is409WithoutRevealingTheOwner(t *testing.T) {
	repo := &fakeRepository{linkFn: func(string, string, string) (staff.LinkResult, error) {
		return staff.LinkResult{Found: true, Taken: true}, nil
	}}
	rec := httptest.NewRecorder()
	httpapi.NewLinkMyBarberHandler(staff.NewService(repo)).ServeHTTP(rec,
		requestWithPrincipal(http.MethodPut, "/api/v1/private/me/barber", []byte(`{"barberId":"`+myBarberID+`"}`)))
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	if problem := decodeProblem(t, rec); problem["code"] != "conflict" {
		t.Fatalf("unexpected problem: %v", problem)
	}
}

func TestLinkMyBarberHandler_UnknownBarber_Is404(t *testing.T) {
	repo := &fakeRepository{linkFn: func(string, string, string) (staff.LinkResult, error) { return staff.LinkResult{}, nil }}
	rec := httptest.NewRecorder()
	httpapi.NewLinkMyBarberHandler(staff.NewService(repo)).ServeHTTP(rec,
		requestWithPrincipal(http.MethodPut, "/api/v1/private/me/barber", []byte(`{"barberId":"`+myBarberID+`"}`)))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestUnlinkMyBarberHandler_Is204AndIdempotent(t *testing.T) {
	for _, released := range []string{myBarberID, ""} {
		repo := &fakeRepository{unlinkFn: func(shop, user string) (string, error) {
			if shop != testShopID || user != "u-1" {
				t.Fatalf("debe usar el principal, recibió %q/%q", shop, user)
			}
			return released, nil
		}}
		rec := httptest.NewRecorder()
		httpapi.NewUnlinkMyBarberHandler(staff.NewService(repo)).ServeHTTP(rec,
			requestWithPrincipal(http.MethodDelete, "/api/v1/private/me/barber", nil))
		if rec.Code != http.StatusNoContent || strings.TrimSpace(rec.Body.String()) != "" {
			t.Fatalf("expected an empty 204, got %d %q", rec.Code, rec.Body.String())
		}
	}
}
