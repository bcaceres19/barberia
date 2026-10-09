// Pruebas de integración del vínculo barbero–usuario (DEC-100, issue #322)
// contra el router REAL (buildRouter) y PostgreSQL real con dos barberías.
// Usan los usuarios de database/testdata/dos_barberias.sql: staffUserActiveA
// y el barbero de la A (aaaaaaa2) como segundo usuario de la misma barbería.
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

const staffUserBarberA = "aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2"

func doMyBarberRequest(router http.Handler, rawToken, method, rawBody string) *httptest.ResponseRecorder {
	var body *bytes.Reader
	if rawBody != "" {
		body = bytes.NewReader([]byte(rawBody))
	} else {
		body = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, "/api/v1/private/me/barber", body)
	req.Header.Set("Content-Type", "application/json")
	if rawToken != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: rawToken})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func createBarberVia(t *testing.T, router http.Handler, raw, name string) barberBody {
	t.Helper()
	rec := doCreateBarberRequest(router, raw, uniqueToken(t, "link-key"), `{"fullName":"`+name+`"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 creating %q, got %d: %s", name, rec.Code, rec.Body.String())
	}
	var b barberBody
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return b
}

func TestMyBarber_HTTP_LinkReadSwitchAndUnlinkJourney(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	raw := createSessionCookie(t, db, shopA, staffUserActiveA, "link-journey-a")
	t.Cleanup(func() { doMyBarberRequest(router, raw, http.MethodDelete, "") })
	first := createBarberVia(t, router, raw, "Vínculo uno "+uniqueToken(t, "n"))
	second := createBarberVia(t, router, raw, "Vínculo dos "+uniqueToken(t, "n"))

	if rec := doMyBarberRequest(router, raw, http.MethodGet, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("sin vínculo se espera 404, got %d: %s", rec.Code, rec.Body.String())
	}

	rec := doMyBarberRequest(router, raw, http.MethodPut, `{"barberId":"`+first.ID+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("vincular: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var linked map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &linked); err != nil {
		t.Fatal(err)
	}
	if linked["id"] != first.ID {
		t.Fatalf("expected the first barber, got %v", linked)
	}
	if _, leaked := linked["staffUserId"]; leaked {
		t.Fatal("la respuesta nunca debe exponer staffUserId")
	}

	if rec := doMyBarberRequest(router, raw, http.MethodPut, `{"barberId":"`+first.ID+`"}`); rec.Code != http.StatusOK {
		t.Fatalf("repetir la selección es idempotente, got %d", rec.Code)
	}

	if rec := doMyBarberRequest(router, raw, http.MethodPut, `{"barberId":"`+second.ID+`"}`); rec.Code != http.StatusOK {
		t.Fatalf("cambiar de barbero: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = doMyBarberRequest(router, raw, http.MethodGet, "")
	var current barberBody
	if rec.Code != http.StatusOK || json.Unmarshal(rec.Body.Bytes(), &current) != nil || current.ID != second.ID {
		t.Fatalf("tras cambiar debe leerse el segundo barbero: %d %s", rec.Code, rec.Body.String())
	}

	if rec := doMyBarberRequest(router, raw, http.MethodDelete, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("quitar: expected 204, got %d", rec.Code)
	}
	if rec := doMyBarberRequest(router, raw, http.MethodDelete, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("quitar sin vínculo es idempotente, got %d", rec.Code)
	}
	if rec := doMyBarberRequest(router, raw, http.MethodGet, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("tras quitarlo se espera 404, got %d", rec.Code)
	}
	// Quitar el vínculo no borra al barbero.
	if rec := doGetBarberRequest(router, raw, second.ID); rec.Code != http.StatusOK {
		t.Fatalf("el barbero debe seguir existiendo, got %d", rec.Code)
	}
}

func TestMyBarber_HTTP_BarberOfAnotherUser_Is409AndPreservesPreviousLink(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	rawOne := createSessionCookie(t, db, shopA, staffUserActiveA, "link-conflict-1")
	rawTwo := createSessionCookie(t, db, shopA, staffUserBarberA, "link-conflict-2")
	t.Cleanup(func() {
		doMyBarberRequest(router, rawOne, http.MethodDelete, "")
		doMyBarberRequest(router, rawTwo, http.MethodDelete, "")
	})
	mine := createBarberVia(t, router, rawOne, "Mío "+uniqueToken(t, "n"))
	theirs := createBarberVia(t, router, rawOne, "Ajeno "+uniqueToken(t, "n"))

	if rec := doMyBarberRequest(router, rawOne, http.MethodPut, `{"barberId":"`+mine.ID+`"}`); rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec := doMyBarberRequest(router, rawTwo, http.MethodPut, `{"barberId":"`+theirs.ID+`"}`); rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	rec := doMyBarberRequest(router, rawOne, http.MethodPut, `{"barberId":"`+theirs.ID+`"}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	var current barberBody
	rec = doMyBarberRequest(router, rawOne, http.MethodGet, "")
	if rec.Code != http.StatusOK || json.Unmarshal(rec.Body.Bytes(), &current) != nil || current.ID != mine.ID {
		t.Fatalf("el vínculo anterior debe conservarse: %d %s", rec.Code, rec.Body.String())
	}
}

func TestMyBarber_HTTP_TwoTenants_AndNoCookie(t *testing.T) {
	db := setupTestDB(t)
	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	rawA := createSessionCookie(t, db, shopA, staffUserActiveA, "link-tenant-a")
	rawB := createSessionCookie(t, db, shopB, staffUserActiveB, "link-tenant-b")
	t.Cleanup(func() {
		doMyBarberRequest(router, rawA, http.MethodDelete, "")
		doMyBarberRequest(router, rawB, http.MethodDelete, "")
	})
	barberB := createBarberVia(t, router, rawB, "Solo B "+uniqueToken(t, "n"))

	// Un usuario de A pidiendo el barbero de B recibe el mismo 404 que un id inexistente.
	cross := doMyBarberRequest(router, rawA, http.MethodPut, `{"barberId":"`+barberB.ID+`"}`)
	missing := doMyBarberRequest(router, rawA, http.MethodPut, `{"barberId":"99999999-9999-4999-8999-999999999999"}`)
	if cross.Code != http.StatusNotFound || missing.Code != http.StatusNotFound {
		t.Fatalf("expected 404/404, got %d/%d", cross.Code, missing.Code)
	}
	if rec := doMyBarberRequest(router, rawA, http.MethodGet, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("A no debe quedar vinculado a un barbero de B, got %d", rec.Code)
	}

	// El cuerpo nunca acepta el usuario de otra persona.
	own := createBarberVia(t, router, rawA, "Propio A "+uniqueToken(t, "n"))
	forged := doMyBarberRequest(router, rawA, http.MethodPut,
		`{"barberId":"`+own.ID+`","staffUserId":"`+staffUserBarberA+`"}`)
	if forged.Code != http.StatusBadRequest {
		t.Fatalf("un staffUserId en el cuerpo debe rechazarse con 400, got %d", forged.Code)
	}

	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		if rec := doMyBarberRequest(router, "", method, `{}`); rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s sin cookie: expected 401, got %d", method, rec.Code)
		}
	}
}
