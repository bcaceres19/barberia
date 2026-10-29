// Pruebas de contrato de HU-100: comparan los DTO y los códigos de los handlers
// con el OpenAPI fuente (api/openapi), sin depender de Node/Redocly.
package httpapi_test

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

type schemaDoc struct {
	Properties           map[string]any `yaml:"properties"`
	Required             []string       `yaml:"required"`
	AdditionalProperties any            `yaml:"additionalProperties"`
}

type operation struct {
	OperationID string         `yaml:"operationId"`
	Security    []any          `yaml:"security"`
	Responses   map[string]any `yaml:"responses"`
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "redocly.yaml")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("no se encontró redocly.yaml")
		}
		dir = parent
	}
}

func load[T any](t *testing.T, rel string) T {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	var v T
	if err := yaml.Unmarshal(raw, &v); err != nil {
		t.Fatalf("%s: %v", rel, err)
	}
	return v
}

func requireExactly(t *testing.T, name string, schema schemaDoc, props, required []string) {
	t.Helper()
	if len(schema.Properties) != len(props) {
		t.Fatalf("%s declara %d propiedades, se esperaban %d: %v", name, len(schema.Properties), len(props), schema.Properties)
	}
	for _, p := range props {
		if _, ok := schema.Properties[p]; !ok {
			t.Errorf("%s no declara %q", name, p)
		}
	}
	for _, r := range required {
		found := false
		for _, x := range schema.Required {
			found = found || x == r
		}
		if !found {
			t.Errorf("%s no marca %q como required", name, r)
		}
	}
}

func TestContract_ReviewSchemas_MatchDTOFields(t *testing.T) {
	create := load[schemaDoc](t, "api/openapi/components/schemas/CreateReviewRequest.yaml")
	requireExactly(t, "CreateReviewRequest", create, []string{"rating", "comment"}, []string{"rating"})
	if create.AdditionalProperties != false {
		t.Error("CreateReviewRequest debe ser cerrado (additionalProperties: false)")
	}
	update := load[schemaDoc](t, "api/openapi/components/schemas/UpdateReviewRequest.yaml")
	requireExactly(t, "UpdateReviewRequest", update, []string{"hidden"}, []string{"hidden"})
	if update.AdditionalProperties != false {
		t.Error("UpdateReviewRequest debe ser cerrado (additionalProperties: false)")
	}

	all := []string{"rating", "comment", "createdAt"}
	requireExactly(t, "CustomerReview", load[schemaDoc](t, "api/openapi/components/schemas/CustomerReview.yaml"), all, all)
	requireExactly(t, "PublicReview", load[schemaDoc](t, "api/openapi/components/schemas/PublicReview.yaml"), all, all)
	requireExactly(t, "Review", load[schemaDoc](t, "api/openapi/components/schemas/Review.yaml"),
		[]string{"id", "rating", "comment", "hiddenAt", "createdAt"}, []string{"id", "rating", "comment", "hiddenAt", "createdAt"})
	list := []string{"rating", "ratingCount", "items"}
	requireExactly(t, "PublicReviewList", load[schemaDoc](t, "api/openapi/components/schemas/PublicReviewList.yaml"), list, list)
	requireExactly(t, "ReviewList", load[schemaDoc](t, "api/openapi/components/schemas/ReviewList.yaml"), list, list)
}

func TestContract_CustomerAppointmentResponse_AddsCanReviewAndReview(t *testing.T) {
	schema := load[schemaDoc](t, "api/openapi/components/schemas/CustomerAppointmentResponse.yaml")
	for _, name := range []string{"canReview", "review"} {
		if _, ok := schema.Properties[name]; !ok {
			t.Errorf("CustomerAppointmentResponse no declara %q", name)
		}
		found := false
		for _, r := range schema.Required {
			found = found || r == name
		}
		if !found {
			t.Errorf("CustomerAppointmentResponse debe exigir %q", name)
		}
	}
}

func statuses(op operation) map[string]bool {
	out := map[string]bool{}
	for k := range op.Responses {
		out[k] = true
	}
	return out
}

func requireStatuses(t *testing.T, name string, op operation, want ...string) {
	t.Helper()
	got := statuses(op)
	if len(got) != len(want) {
		t.Errorf("%s documenta %d respuestas, se esperaban %v", name, len(got), want)
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("%s no documenta %s", name, w)
		}
	}
}

func TestContract_ReviewOperations_SecurityAndStatuses(t *testing.T) {
	customer := load[struct {
		Review struct {
			Post operation `yaml:"post"`
		} `yaml:"/customer/appointments/{token}/review"`
	}](t, "api/openapi/paths/customer-appointments.yaml")
	if customer.Review.Post.OperationID != "createCustomerAppointmentReview" || len(customer.Review.Post.Security) != 0 {
		t.Errorf("POST review del cliente: operationId=%q security=%v (debe ser público: security: [])",
			customer.Review.Post.OperationID, customer.Review.Post.Security)
	}
	requireStatuses(t, "POST /customer/appointments/{token}/review", customer.Review.Post, "201", "400", "404", "409", "422", "500")

	public := load[struct {
		Reviews struct {
			Get operation `yaml:"get"`
		} `yaml:"/public/barbershops/{slug}/barbers/{barberId}/reviews"`
	}](t, "api/openapi/paths/public-booking.yaml")
	if public.Reviews.Get.OperationID != "listPublicBarberReviews" || len(public.Reviews.Get.Security) != 0 {
		t.Errorf("GET público de reseñas: operationId=%q security=%v", public.Reviews.Get.OperationID, public.Reviews.Get.Security)
	}
	requireStatuses(t, "GET /public/.../reviews", public.Reviews.Get, "200", "404", "500")

	staff := load[struct {
		List struct {
			Get operation `yaml:"get"`
		} `yaml:"/private/barbers/{barberId}/reviews"`
		Item struct {
			Patch operation `yaml:"patch"`
		} `yaml:"/private/barbers/{barberId}/reviews/{reviewId}"`
	}](t, "api/openapi/paths/staff.yaml")
	if len(staff.List.Get.Security) == 0 || len(staff.Item.Patch.Security) == 0 {
		t.Error("las operaciones del panel exigen sesión")
	}
	requireStatuses(t, "GET /private/barbers/{barberId}/reviews", staff.List.Get, "200", "401", "404", "500")
	requireStatuses(t, "PATCH /private/barbers/{barberId}/reviews/{reviewId}", staff.Item.Patch, "200", "400", "401", "404", "422", "500")
}

func TestContract_OpenAPIYAML_RegistersTheReviewPaths(t *testing.T) {
	doc := load[struct {
		Paths map[string]any `yaml:"paths"`
	}](t, "api/openapi/openapi.yaml")
	for _, p := range []string{
		"/customer/appointments/{token}/review",
		"/public/barbershops/{slug}/barbers/{barberId}/reviews",
		"/private/barbers/{barberId}/reviews",
		"/private/barbers/{barberId}/reviews/{reviewId}",
	} {
		if _, ok := doc.Paths[p]; !ok {
			t.Errorf("openapi.yaml no registra %s", p)
		}
	}
}
