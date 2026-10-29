// Pruebas de contrato de HU-099: comparan los DTO y los códigos del handler de
// cancelación con el OpenAPI fuente (api/openapi), sin depender de Node/Redocly.
package httpapi_test

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"system-barbershop/internal/modules/customeraccess/httpapi"
)

type contractSchema struct {
	Properties           map[string]any `yaml:"properties"`
	Required             []string       `yaml:"required"`
	AdditionalProperties any            `yaml:"additionalProperties"`
}

func contractRoot(t *testing.T) string {
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

func loadYAML[T any](t *testing.T, rel string) T {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(contractRoot(t), filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	var v T
	if err := yaml.Unmarshal(raw, &v); err != nil {
		t.Fatalf("%s: %v", rel, err)
	}
	return v
}

func jsonFields(t *testing.T, v any) []string {
	t.Helper()
	typ := reflect.TypeOf(v)
	var names []string
	for i := 0; i < typ.NumField(); i++ {
		tag := typ.Field(i).Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		names = append(names, strings.Split(tag, ",")[0])
	}
	sort.Strings(names)
	return names
}

func schemaProps(s contractSchema) []string {
	names := make([]string, 0, len(s.Properties))
	for k := range s.Properties {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func TestContract_CustomerAppointmentResponse_MatchesDTO(t *testing.T) {
	schema := loadYAML[contractSchema](t, "api/openapi/components/schemas/CustomerAppointmentResponse.yaml")
	if got, want := jsonFields(t, httpapi.CustomerAppointmentResponse{}), schemaProps(schema); !reflect.DeepEqual(got, want) {
		t.Fatalf("DTO %v != OpenAPI %v", got, want)
	}
	for _, field := range []string{"versionToken", "barbershopContactEmail", "barbershopContactPhone"} {
		found := false
		for _, r := range schema.Required {
			found = found || r == field
		}
		if !found {
			t.Errorf("CustomerAppointmentResponse no marca %q como required", field)
		}
	}
}

func TestContract_CancelRequestIsClosedAndOnlyCarriesReason(t *testing.T) {
	schema := loadYAML[contractSchema](t, "api/openapi/components/schemas/CancelCustomerAppointmentRequest.yaml")
	if got, want := jsonFields(t, httpapi.CancelAppointmentRequest{}), schemaProps(schema); !reflect.DeepEqual(got, want) {
		t.Fatalf("DTO %v != OpenAPI %v", got, want)
	}
	if schema.AdditionalProperties != false {
		t.Fatal("el cuerpo debe ser cerrado (additionalProperties: false)")
	}
	if len(schema.Required) != 0 {
		t.Fatalf("el motivo es opcional en el contrato: %v", schema.Required)
	}
}

func TestContract_CancelOperation_DeclaresEverySharedStatusAndNoSession(t *testing.T) {
	ops := loadYAML[map[string]struct {
		Post struct {
			OperationID string         `yaml:"operationId"`
			Security    []any          `yaml:"security"`
			Responses   map[string]any `yaml:"responses"`
		} `yaml:"post"`
	}](t, "api/openapi/paths/customer-appointments.yaml")

	op, ok := ops["/customer/appointments/{token}/cancel"]
	if !ok {
		t.Fatal("falta /customer/appointments/{token}/cancel")
	}
	if op.Post.OperationID != "cancelCustomerAppointment" {
		t.Fatalf("operationId = %q", op.Post.OperationID)
	}
	if len(op.Post.Security) != 0 {
		t.Fatalf("security debe ser [] (sin sesión): %v", op.Post.Security)
	}
	for _, status := range []string{"200", "400", "404", "409", "422", "500"} {
		if _, ok := op.Post.Responses[status]; !ok {
			t.Errorf("falta la respuesta %s", status)
		}
	}
}
