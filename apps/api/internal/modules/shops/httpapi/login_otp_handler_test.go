package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"system-barbershop/internal/modules/shops"
	"system-barbershop/internal/modules/shops/httpapi"
)

type fakeLoginOTPRepository struct {
	channel   shops.LoginOTPChannel
	found     bool
	gotShopID string
	updates   []shops.LoginOTPChannel
}

func (f *fakeLoginOTPRepository) Get(_ context.Context, shopID string) (shops.LoginOTPChannel, bool, error) {
	f.gotShopID = shopID
	return f.channel, f.found, nil
}

func (f *fakeLoginOTPRepository) Update(_ context.Context, shopID string, c shops.LoginOTPChannel) (shops.LoginOTPChannel, bool, error) {
	f.gotShopID = shopID
	f.updates = append(f.updates, c)
	return c, f.found, nil
}

const loginOTPTarget = "/api/v1/private/settings/login-otp"

func TestGetLoginOTPHandler_ReturnsOnlyTheChannelAndForbidsCaching(t *testing.T) {
	repo := &fakeLoginOTPRepository{channel: shops.LoginOTPChannelBoth, found: true}
	h := httpapi.NewGetLoginOTPHandler(shops.NewLoginOTPService(repo))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, requestWithPrincipal(http.MethodGet, loginOTPTarget, nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuvo %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body) != 1 || body["channel"] != "both" {
		t.Fatalf("esperaba exactamente {channel}, obtuvo %v", body)
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("esperaba no-store, obtuvo %q", rec.Header().Get("Cache-Control"))
	}
	if repo.gotShopID != testShopID {
		t.Fatalf("la barbería sale de la sesión: esperaba %q, obtuvo %q", testShopID, repo.gotShopID)
	}
}

func TestLoginOTPHandlers_NoPrincipal_Is500(t *testing.T) {
	svc := shops.NewLoginOTPService(&fakeLoginOTPRepository{found: true})
	for name, h := range map[string]http.Handler{
		"GET":   httpapi.NewGetLoginOTPHandler(svc),
		"PATCH": httpapi.NewUpdateLoginOTPHandler(svc),
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPatch, loginOTPTarget, strings.NewReader(`{"channel":"email"}`)))
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("%s sin principal: esperaba 500, obtuvo %d", name, rec.Code)
		}
	}
}

func TestUpdateLoginOTPHandler_Success_StoresForTheSessionBarbershop(t *testing.T) {
	repo := &fakeLoginOTPRepository{found: true}
	h := httpapi.NewUpdateLoginOTPHandler(shops.NewLoginOTPService(repo))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, requestWithPrincipal(http.MethodPatch, loginOTPTarget, []byte(`{"channel":"whatsapp"}`)))

	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuvo %d: %s", rec.Code, rec.Body.String())
	}
	if len(repo.updates) != 1 || repo.updates[0] != shops.LoginOTPChannelWhatsApp || repo.gotShopID != testShopID {
		t.Fatalf("updates=%v shop=%q", repo.updates, repo.gotShopID)
	}
}

// DEC-081: ningún destino viaja en la petición; cualquier campo extra es un 400
// y no escribe nada.
func TestUpdateLoginOTPHandler_RejectsAnyDestinationOrUnknownField(t *testing.T) {
	for name, body := range map[string]string{
		"correo":       `{"channel":"email","email":"otro@ejemplo.test"}`,
		"telefono":     `{"channel":"whatsapp","phone":"+573000000000"}`,
		"barbershopId": `{"channel":"email","barbershopId":"22222222-2222-2222-2222-222222222222"}`,
		"vacio":        `{"channel":"email"} {"channel":"both"}`,
		"no-json":      `channel=email`,
	} {
		t.Run(name, func(t *testing.T) {
			repo := &fakeLoginOTPRepository{found: true}
			rec := httptest.NewRecorder()
			httpapi.NewUpdateLoginOTPHandler(shops.NewLoginOTPService(repo)).
				ServeHTTP(rec, requestWithPrincipal(http.MethodPatch, loginOTPTarget, []byte(body)))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("esperaba 400, obtuvo %d: %s", rec.Code, rec.Body.String())
			}
			if len(repo.updates) != 0 {
				t.Fatal("una petición inválida no debe escribir")
			}
		})
	}
}

func TestUpdateLoginOTPHandler_ChannelOutsideTheList_Is422(t *testing.T) {
	for _, channel := range []string{"sms", "", "Email", "email,whatsapp"} {
		repo := &fakeLoginOTPRepository{found: true}
		rec := httptest.NewRecorder()
		httpapi.NewUpdateLoginOTPHandler(shops.NewLoginOTPService(repo)).
			ServeHTTP(rec, requestWithPrincipal(http.MethodPatch, loginOTPTarget, []byte(`{"channel":"`+channel+`"}`)))
		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("canal %q: esperaba 422, obtuvo %d", channel, rec.Code)
		}
		if len(repo.updates) != 0 {
			t.Errorf("canal %q: no debe escribir", channel)
		}
	}
}

func TestUpdateLoginOTPHandler_BarbershopNotVisible_Is404(t *testing.T) {
	rec := httptest.NewRecorder()
	httpapi.NewUpdateLoginOTPHandler(shops.NewLoginOTPService(&fakeLoginOTPRepository{found: false})).
		ServeHTTP(rec, requestWithPrincipal(http.MethodPatch, loginOTPTarget, []byte(`{"channel":"email"}`)))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperaba 404, obtuvo %d", rec.Code)
	}
}

// El enum del contrato y el del dominio no pueden divergir.
func TestContract_LoginOtpChannelEnum_MatchesShopsChannels(t *testing.T) {
	doc := loadYAML[enumSchemaDoc](t, "api/openapi/components/schemas/LoginOtpChannel.yaml")
	domain := []shops.LoginOTPChannel{shops.LoginOTPChannelEmail, shops.LoginOTPChannelWhatsApp, shops.LoginOTPChannelBoth}
	if len(doc.Enum) != len(domain) {
		t.Fatalf("el dominio tiene %d canales, el contrato documenta %v", len(domain), doc.Enum)
	}
	for _, c := range domain {
		if !containsString(doc.Enum, string(c)) {
			t.Errorf("el dominio tiene %q pero el contrato no lo documenta", c)
		}
	}
}

// El cuerpo solo declara `channel` (sin destinos ni barbershopId).
func TestContract_LoginOtpSchemas_OnlyDeclareChannel(t *testing.T) {
	for _, file := range []string{"LoginOtpSettings", "UpdateLoginOtpRequest"} {
		schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/"+file+".yaml")
		if len(schema.Properties) != 1 {
			t.Fatalf("%s declara %d propiedades, se esperaba solo channel", file, len(schema.Properties))
		}
		requireProps(t, schema, []string{"channel"})
		if !containsString(schema.Required, "channel") {
			t.Errorf("%s no marca channel como required", file)
		}
	}
}
