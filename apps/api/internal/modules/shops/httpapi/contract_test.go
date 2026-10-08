// Pruebas de contrato de HU-020: comparan el handler real contra el
// OpenAPI fuente (api/openapi/paths/settings.yaml y sus schemas), sin
// depender de Node/Redocly (mismo criterio que
// internal/modules/auth/httpapi/contract_test.go): se lee el YAML fuente
// directamente, no el bundle de dist/.
package httpapi_test

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

	"system-barbershop/internal/modules/shops"
)

type schemaDoc struct {
	Properties map[string]any `yaml:"properties"`
	Required   []string       `yaml:"required"`
}

type operation struct {
	OperationID string         `yaml:"operationId"`
	Security    []any          `yaml:"security"`
	Responses   map[string]any `yaml:"responses"`
}

type refParam struct {
	Ref  string `yaml:"$ref"`
	Name string `yaml:"name"`
}

type bookingPolicyOperation struct {
	OperationID string         `yaml:"operationId"`
	Security    []any          `yaml:"security"`
	Parameters  []refParam     `yaml:"parameters"`
	Responses   map[string]any `yaml:"responses"`
}

type barbershopSettingsPathFile struct {
	BarbershopSettings struct {
		Get   operation `yaml:"get"`
		Patch operation `yaml:"patch"`
	} `yaml:"/private/settings/barbershop"`
}

type bookingPolicyPathFile struct {
	BookingPolicy struct {
		Get bookingPolicyOperation `yaml:"get"`
		Put bookingPolicyOperation `yaml:"put"`
	} `yaml:"/private/settings/booking-policy"`
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("no se pudo obtener el directorio de trabajo: %v", err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "redocly.yaml")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("no se encontró redocly.yaml subiendo desde el directorio de la prueba")
		}
		dir = parent
	}
}

func loadYAML[T any](t *testing.T, relPath string) T {
	t.Helper()
	full := filepath.Join(findRepoRoot(t), filepath.FromSlash(relPath))
	raw, err := os.ReadFile(full)
	if err != nil {
		t.Fatalf("no se pudo leer %s: %v", full, err)
	}
	var v T
	if err := yaml.Unmarshal(raw, &v); err != nil {
		t.Fatalf("no se pudo interpretar %s: %v", full, err)
	}
	return v
}

func requireProps(t *testing.T, schema schemaDoc, want []string) {
	t.Helper()
	if len(schema.Properties) != len(want) {
		t.Fatalf("expected %d properties, schema has %d: %v", len(want), len(schema.Properties), schema.Properties)
	}
	for _, prop := range want {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("schema no declara la propiedad %q", prop)
		}
	}
	for _, wantReq := range want {
		found := false
		for _, r := range schema.Required {
			if r == wantReq {
				found = true
			}
		}
		if !found {
			t.Errorf("schema no marca %q como required", wantReq)
		}
	}
}

// TestContract_BarbershopSettingsResponseSchema_MatchesDTOFields verifica
// que httpapi.BarbershopSettingsResponse (name, timezone, contactEmail,
// contactPhone) coincide exactamente con las propiedades y los required de
// BarbershopSettingsResponse.yaml (CA-020-07: exactamente esos cuatro
// campos, nunca public_slug ni logo).
func TestContract_BarbershopSettingsResponseSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/BarbershopSettingsResponse.yaml")
	requireProps(t, schema, []string{"name", "timezone", "contactEmail", "contactPhone"})
}

// TestContract_UpdateBarbershopSettingsRequestSchema_MatchesDTOFields
// verifica httpapi.UpdateBarbershopSettingsRequest contra
// UpdateBarbershopSettingsRequest.yaml. barbershopId NUNCA aparece
// (CA-020-05): esta comprobación de "exactamente 4 propiedades" es la
// defensa de contrato contra que alguien lo agregue sin darse cuenta.
func TestContract_UpdateBarbershopSettingsRequestSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/UpdateBarbershopSettingsRequest.yaml")
	requireProps(t, schema, []string{"name", "timezone", "contactEmail", "contactPhone"})
	if _, ok := schema.Properties["barbershopId"]; ok {
		t.Fatal("CA-020-05: el contrato de PATCH nunca debe declarar barbershopId")
	}
}

// TestContract_GetBarbershopSettingsOperation_MethodPathSecurityAndResponses
// verifica que el path fuente declara exactamente la operación de lectura:
// SessionCookie y los códigos que el handler realmente produce (200, 401,
// 404 defensivo, 500), sin inventar códigos no ejercitados.
func TestContract_GetBarbershopSettingsOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[barbershopSettingsPathFile](t, "api/openapi/paths/settings.yaml")
	op := doc.BarbershopSettings.Get

	if op.OperationID != "getBarbershopSettings" {
		t.Fatalf("expected operationId getBarbershopSettings, got %q", op.OperationID)
	}
	if len(op.Security) != 1 {
		t.Fatalf("expected exactly one security requirement (SessionCookie), got %v", op.Security)
	}

	wantStatuses := []string{"200", "401", "404", "500"}
	for _, s := range wantStatuses {
		if _, ok := op.Responses[s]; !ok {
			t.Errorf("el contrato no documenta la respuesta %s, pero el handler la produce", s)
		}
	}
	if len(op.Responses) != len(wantStatuses) {
		t.Errorf("el contrato documenta %d respuestas, se esperaban exactamente %d (%v); revisa que no sobre ni falte una",
			len(op.Responses), len(wantStatuses), wantStatuses)
	}
}

// TestContract_UpdateBarbershopSettingsOperation_MethodPathSecurityAndResponses
// cubre la operación de actualización: 200 éxito, 400 JSON/campo
// desconocido, 401, 404 defensivo, 422 validación de campo, 500.
func TestContract_UpdateBarbershopSettingsOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[barbershopSettingsPathFile](t, "api/openapi/paths/settings.yaml")
	op := doc.BarbershopSettings.Patch

	if op.OperationID != "updateBarbershopSettings" {
		t.Fatalf("expected operationId updateBarbershopSettings, got %q", op.OperationID)
	}
	if len(op.Security) != 1 {
		t.Fatalf("expected exactly one security requirement (SessionCookie), got %v", op.Security)
	}

	wantStatuses := []string{"200", "400", "401", "404", "422", "500"}
	for _, s := range wantStatuses {
		if _, ok := op.Responses[s]; !ok {
			t.Errorf("el contrato no documenta la respuesta %s, pero el handler la produce", s)
		}
	}
	if len(op.Responses) != len(wantStatuses) {
		t.Errorf("el contrato documenta %d respuestas, se esperaban exactamente %d (%v); revisa que no sobre ni falte una",
			len(op.Responses), len(wantStatuses), wantStatuses)
	}
}

// TestContract_OpenAPIYAML_RegistersBarbershopSettingsPath confirma que
// openapi.yaml registra el path bajo el mismo documento raíz que las demás
// operaciones privadas, con la misma técnica de referencia JSON pointer.
func TestContract_OpenAPIYAML_RegistersBarbershopSettingsPath(t *testing.T) {
	type pathsDoc struct {
		Paths map[string]any `yaml:"paths"`
	}
	doc := loadYAML[pathsDoc](t, "api/openapi/openapi.yaml")

	if _, ok := doc.Paths["/private/settings/barbershop"]; !ok {
		t.Fatal("openapi.yaml no registra paths./private/settings/barbershop")
	}
}

// TestContract_BookingPolicyResponseSchema_MatchesDTOFields verifica que
// httpapi.BookingPolicyResponse (HU-093) coincide exactamente con las
// propiedades y los required de BookingPolicyResponse.yaml.
func TestContract_BookingPolicyResponseSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/BookingPolicyResponse.yaml")
	requireProps(t, schema, []string{
		"minAdvanceMinutes", "maxAdvanceDays", "slotGridMinutes",
		"cancellationDeadlineMinutes", "lateCancellationClientAllowed",
		"lateCancellationReasonRequired", "versionToken",
	})
}

// TestContract_UpdateBookingPolicyRequestSchema_MatchesDTOFields verifica
// httpapi.UpdateBookingPolicyRequest contra UpdateBookingPolicyRequest.yaml.
// barbershopId y versionToken NUNCA aparecen en el cuerpo (CA-093-03: el
// tenant se deriva de la sesión; la precondición de versión viaja en
// If-Match, no en el cuerpo).
func TestContract_UpdateBookingPolicyRequestSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/UpdateBookingPolicyRequest.yaml")
	requireProps(t, schema, []string{
		"minAdvanceMinutes", "maxAdvanceDays", "slotGridMinutes",
		"cancellationDeadlineMinutes", "lateCancellationClientAllowed",
		"lateCancellationReasonRequired",
	})
	for _, forbidden := range []string{"barbershopId", "versionToken"} {
		if _, ok := schema.Properties[forbidden]; ok {
			t.Fatalf("CA-093-03: el contrato de PUT nunca debe declarar %q", forbidden)
		}
	}
}

// TestContract_GetBookingPolicyOperation_MethodPathSecurityAndResponses
// verifica que el path fuente declara exactamente la operación de lectura:
// SessionCookie y los códigos que el handler realmente produce (200, 401,
// 404 defensivo, 500), sin inventar códigos no ejercitados.
func TestContract_GetBookingPolicyOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[bookingPolicyPathFile](t, "api/openapi/paths/settings.yaml")
	op := doc.BookingPolicy.Get

	if op.OperationID != "getBookingPolicy" {
		t.Fatalf("expected operationId getBookingPolicy, got %q", op.OperationID)
	}
	if len(op.Security) != 1 {
		t.Fatalf("expected exactly one security requirement (SessionCookie), got %v", op.Security)
	}

	wantStatuses := []string{"200", "401", "404", "500"}
	for _, s := range wantStatuses {
		if _, ok := op.Responses[s]; !ok {
			t.Errorf("el contrato no documenta la respuesta %s, pero el handler la produce", s)
		}
	}
	if len(op.Responses) != len(wantStatuses) {
		t.Errorf("el contrato documenta %d respuestas, se esperaban exactamente %d (%v); revisa que no sobre ni falte una",
			len(op.Responses), len(wantStatuses), wantStatuses)
	}
}

// TestContract_UpdateBookingPolicyOperation_MethodPathSecurityAndResponses
// cubre la operación de actualización: If-Match obligatorio (CA-093-02),
// 200 éxito, 400 JSON/campo desconocido, 401, 404 defensivo, 409 conflicto
// de versión, 422 validación de rango/coherencia, 500.
func TestContract_UpdateBookingPolicyOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[bookingPolicyPathFile](t, "api/openapi/paths/settings.yaml")
	op := doc.BookingPolicy.Put

	if op.OperationID != "updateBookingPolicy" {
		t.Fatalf("expected operationId updateBookingPolicy, got %q", op.OperationID)
	}
	if len(op.Security) != 1 {
		t.Fatalf("expected exactly one security requirement (SessionCookie), got %v", op.Security)
	}
	if len(op.Parameters) != 1 || op.Parameters[0].Ref != "../components/parameters/IfMatch.yaml" {
		t.Fatalf("expected exactly one parameter referencing IfMatch.yaml, got %+v", op.Parameters)
	}

	wantStatuses := []string{"200", "400", "401", "404", "409", "422", "500"}
	for _, s := range wantStatuses {
		if _, ok := op.Responses[s]; !ok {
			t.Errorf("el contrato no documenta la respuesta %s, pero el handler la produce", s)
		}
	}
	if len(op.Responses) != len(wantStatuses) {
		t.Errorf("el contrato documenta %d respuestas, se esperaban exactamente %d (%v); revisa que no sobre ni falte una",
			len(op.Responses), len(wantStatuses), wantStatuses)
	}
}

// TestContract_OpenAPIYAML_RegistersBookingPolicyPath confirma que
// openapi.yaml registra el path bajo el mismo documento raíz que las demás
// operaciones privadas.
func TestContract_OpenAPIYAML_RegistersBookingPolicyPath(t *testing.T) {
	type pathsDoc struct {
		Paths map[string]any `yaml:"paths"`
	}
	doc := loadYAML[pathsDoc](t, "api/openapi/openapi.yaml")

	if _, ok := doc.Paths["/private/settings/booking-policy"]; !ok {
		t.Fatal("openapi.yaml no registra paths./private/settings/booking-policy")
	}
}

// --- Marca y vocabulario (issue #292, DEC-110) ------------------------------

type brandPathFile struct {
	Brand struct {
		Get   operation `yaml:"get"`
		Patch operation `yaml:"patch"`
	} `yaml:"/private/settings/brand"`
}

type enumSchemaDoc struct {
	Enum []string `yaml:"enum"`
}

var brandFields = []string{
	"accent", "businessTerm", "businessTermGender",
	"professionalTerm", "professionalTermPlural", "professionalTermGender",
}

// panelProfileField es el séptimo campo (issue #294, DEC-115): siempre presente
// en la respuesta y optativo en la solicitud.
const panelProfileField = "panelProfile"

// TestContract_BrandSchemas_MatchDTOFields verifica que BrandResponse y
// UpdateBrandRequest declaran exactamente los campos de los DTO y que ninguno
// declara barbershopId (el tenant sale solo de la sesión). panelProfile es
// required en la respuesta y optativo en la solicitud: un cliente que no lo
// conoce no debe reiniciar el perfil guardado.
func TestContract_BrandSchemas_MatchDTOFields(t *testing.T) {
	response := loadYAML[schemaDoc](t, "api/openapi/components/schemas/BrandResponse.yaml")
	requireProps(t, response, append(append([]string{}, brandFields...), panelProfileField))

	request := loadYAML[schemaDoc](t, "api/openapi/components/schemas/UpdateBrandRequest.yaml")
	if _, ok := request.Properties[panelProfileField]; !ok {
		t.Fatalf("UpdateBrandRequest no declara %s", panelProfileField)
	}
	if len(request.Properties) != len(brandFields)+1 {
		t.Fatalf("UpdateBrandRequest declara %d propiedades, se esperaban %d", len(request.Properties), len(brandFields)+1)
	}
	for _, field := range brandFields {
		if !containsString(request.Required, field) {
			t.Errorf("UpdateBrandRequest no marca %q como required", field)
		}
	}
	if containsString(request.Required, panelProfileField) {
		t.Errorf("UpdateBrandRequest no debe exigir %s: es optativo (DEC-115)", panelProfileField)
	}

	for name, schema := range map[string]schemaDoc{"BrandResponse": response, "UpdateBrandRequest": request} {
		if _, ok := schema.Properties["barbershopId"]; ok {
			t.Fatalf("%s nunca debe declarar barbershopId", name)
		}
	}
}

func containsString(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}

// TestContract_PanelProfileEnum_MatchesShopsProfiles ata el contrato a los
// perfiles del dominio (y estos, a barbershop_panel_profile_ck).
func TestContract_PanelProfileEnum_MatchesShopsProfiles(t *testing.T) {
	doc := loadYAML[enumSchemaDoc](t, "api/openapi/components/schemas/PanelProfile.yaml")
	if len(doc.Enum) != 2 {
		t.Fatalf("se esperaban 2 perfiles, el contrato documenta %v", doc.Enum)
	}
	for _, p := range doc.Enum {
		if !shops.PanelProfile(p).IsValid() {
			t.Errorf("el contrato documenta %q pero el dominio no lo reconoce", p)
		}
	}
}

// TestContract_BrandAccentEnum_MatchesShopsAllowedAccents ata el contrato a la
// lista cerrada del dominio (y esta, a barbershop_brand_accent_ck): si una de
// las tres cambia sin las otras, esta prueba falla.
func TestContract_BrandAccentEnum_MatchesShopsAllowedAccents(t *testing.T) {
	doc := loadYAML[enumSchemaDoc](t, "api/openapi/components/schemas/BrandAccent.yaml")
	if len(doc.Enum) != len(shops.AllowedAccents) {
		t.Fatalf("el enum documenta %d acentos y el dominio permite %d", len(doc.Enum), len(shops.AllowedAccents))
	}
	for _, key := range doc.Enum {
		if !shops.IsAllowedAccent(key) {
			t.Errorf("el contrato documenta %q pero shops.AllowedAccents no lo permite", key)
		}
	}
}

func TestContract_BrandGenderEnum_MatchesShopsGenders(t *testing.T) {
	doc := loadYAML[enumSchemaDoc](t, "api/openapi/components/schemas/BrandTermGender.yaml")
	if len(doc.Enum) != 2 {
		t.Fatalf("se esperaban 2 géneros, el contrato documenta %v", doc.Enum)
	}
	for _, g := range doc.Enum {
		if !shops.Gender(g).IsValid() {
			t.Errorf("el contrato documenta %q pero el dominio no lo reconoce", g)
		}
	}
}

func TestContract_BrandTermLimits_MatchShopsTermLength(t *testing.T) {
	type termDoc struct {
		MinLength int `yaml:"minLength"`
		MaxLength int `yaml:"maxLength"`
	}
	doc := loadYAML[termDoc](t, "api/openapi/components/schemas/BrandTerm.yaml")
	if doc.MinLength != shops.TermMinLength || doc.MaxLength != shops.TermMaxLength {
		t.Fatalf("el contrato declara %d-%d y el dominio %d-%d",
			doc.MinLength, doc.MaxLength, shops.TermMinLength, shops.TermMaxLength)
	}
}

func TestContract_BrandOperations_SecurityAndResponses(t *testing.T) {
	doc := loadYAML[brandPathFile](t, "api/openapi/paths/settings.yaml")

	cases := []struct {
		op       operation
		id       string
		statuses []string
	}{
		{doc.Brand.Get, "getBrand", []string{"200", "401", "404", "500"}},
		{doc.Brand.Patch, "updateBrand", []string{"200", "400", "401", "404", "422", "500"}},
	}
	for _, c := range cases {
		if c.op.OperationID != c.id {
			t.Fatalf("expected operationId %s, got %q", c.id, c.op.OperationID)
		}
		if len(c.op.Security) != 1 {
			t.Fatalf("%s: expected exactly one security requirement, got %v", c.id, c.op.Security)
		}
		for _, s := range c.statuses {
			if _, ok := c.op.Responses[s]; !ok {
				t.Errorf("%s: el contrato no documenta la respuesta %s, pero el handler la produce", c.id, s)
			}
		}
		if len(c.op.Responses) != len(c.statuses) {
			t.Errorf("%s: el contrato documenta %d respuestas, se esperaban %d (%v)", c.id, len(c.op.Responses), len(c.statuses), c.statuses)
		}
	}
}

func TestContract_OpenAPIYAML_RegistersBrandPath(t *testing.T) {
	type pathsDoc struct {
		Paths map[string]any `yaml:"paths"`
	}
	doc := loadYAML[pathsDoc](t, "api/openapi/openapi.yaml")

	if _, ok := doc.Paths["/private/settings/brand"]; !ok {
		t.Fatal("openapi.yaml no registra paths./private/settings/brand")
	}
}

// --- Enlace público de reservas (issue #304, DEC-117) -----------------------

type publicLinkPathFile struct {
	PublicLink struct {
		Get operation `yaml:"get"`
	} `yaml:"/private/settings/public-link"`
}

type slugSchemaDoc struct {
	Properties map[string]struct {
		MinLength int    `yaml:"minLength"`
		MaxLength int    `yaml:"maxLength"`
		Pattern   string `yaml:"pattern"`
	} `yaml:"properties"`
}

// TestContract_PublicLinkSchema_MatchesDTOAndDatabaseLimits: la respuesta
// declara solo `slug` (ningún identificador interno) y sus límites son los de
// barbershop_public_slug_ck / shops.SlugMinLength-SlugMaxLength.
func TestContract_PublicLinkSchema_MatchesDTOAndDatabaseLimits(t *testing.T) {
	response := loadYAML[schemaDoc](t, "api/openapi/components/schemas/PublicLinkResponse.yaml")
	requireProps(t, response, []string{"slug"})
	if _, ok := response.Properties["barbershopId"]; ok {
		t.Fatal("PublicLinkResponse nunca debe declarar barbershopId")
	}

	slug := loadYAML[slugSchemaDoc](t, "api/openapi/components/schemas/PublicLinkResponse.yaml").Properties["slug"]
	if slug.MinLength != shops.SlugMinLength || slug.MaxLength != shops.SlugMaxLength {
		t.Fatalf("slug declara [%d,%d], shops espera [%d,%d]", slug.MinLength, slug.MaxLength, shops.SlugMinLength, shops.SlugMaxLength)
	}
	if slug.Pattern != `^[a-z0-9]([a-z0-9-]{1,38}[a-z0-9])$` {
		t.Fatalf("slug.pattern no coincide con barbershop_public_slug_ck: %q", slug.Pattern)
	}
}

func TestContract_PublicLinkOperation_SecurityAndResponses(t *testing.T) {
	doc := loadYAML[publicLinkPathFile](t, "api/openapi/paths/settings.yaml")
	op := doc.PublicLink.Get

	if op.OperationID != "getPublicLink" {
		t.Fatalf("expected operationId getPublicLink, got %q", op.OperationID)
	}
	if len(op.Security) != 1 {
		t.Fatalf("expected exactly one security requirement, got %v", op.Security)
	}
	want := []string{"200", "401", "404", "500"}
	for _, s := range want {
		if _, ok := op.Responses[s]; !ok {
			t.Errorf("el contrato no documenta la respuesta %s, pero el handler la produce", s)
		}
	}
	if len(op.Responses) != len(want) {
		t.Errorf("el contrato documenta %d respuestas, se esperaban %d", len(op.Responses), len(want))
	}
}

func TestContract_OpenAPIYAML_RegistersPublicLinkPath(t *testing.T) {
	type pathsDoc struct {
		Paths map[string]any `yaml:"paths"`
	}
	doc := loadYAML[pathsDoc](t, "api/openapi/openapi.yaml")

	if _, ok := doc.Paths["/private/settings/public-link"]; !ok {
		t.Fatal("openapi.yaml no registra paths./private/settings/public-link")
	}
}
