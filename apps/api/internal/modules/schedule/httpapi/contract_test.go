// Pruebas de contrato de HU-040: comparan los handlers reales contra el
// OpenAPI fuente (api/openapi/paths/schedules.yaml y sus schemas), sin
// depender de Node/Redocly, mismo criterio que
// staff/httpapi/contract_test.go: se lee el YAML fuente directamente, no
// el bundle de dist/.
package httpapi_test

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
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

type schedulePathsFile struct {
	Collection struct {
		Get  operation `yaml:"get"`
		Post operation `yaml:"post"`
	} `yaml:"/private/barbers/{barberId}/working-hours"`
	Item struct {
		Get    operation `yaml:"get"`
		Patch  operation `yaml:"patch"`
		Delete operation `yaml:"delete"`
	} `yaml:"/private/barbers/{barberId}/working-hours/{workingHourId}"`
}

type scheduleExceptionPathsFile struct {
	HolidayCalendar struct {
		Get   operation `yaml:"get"`
		Patch operation `yaml:"patch"`
	} `yaml:"/private/barbers/{barberId}/holiday-calendar"`
	Collection struct {
		Get  operation `yaml:"get"`
		Post operation `yaml:"post"`
	} `yaml:"/private/barbers/{barberId}/schedule-exceptions"`
	Item struct {
		Get    operation `yaml:"get"`
		Patch  operation `yaml:"patch"`
		Delete operation `yaml:"delete"`
	} `yaml:"/private/barbers/{barberId}/schedule-exceptions/{exceptionId}"`
	ColombianHolidays struct {
		Get operation `yaml:"get"`
	} `yaml:"/private/schedule/colombian-holidays"`
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

func requireResponses(t *testing.T, op operation, wantOperationID string, wantStatuses []string) {
	t.Helper()
	if op.OperationID != wantOperationID {
		t.Fatalf("expected operationId %q, got %q", wantOperationID, op.OperationID)
	}
	if len(op.Security) != 1 {
		t.Fatalf("expected exactly one security requirement (SessionCookie), got %v", op.Security)
	}
	for _, s := range wantStatuses {
		if _, ok := op.Responses[s]; !ok {
			t.Errorf("el contrato no documenta la respuesta %s para %s, pero el handler la produce", s, wantOperationID)
		}
	}
	if len(op.Responses) != len(wantStatuses) {
		t.Errorf("%s documenta %d respuestas, se esperaban exactamente %d (%v); revisa que no sobre ni falte una",
			wantOperationID, len(op.Responses), len(wantStatuses), wantStatuses)
	}
}

// --- Schemas -------------------------------------------------------------

func TestContract_WorkingHourResponseSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/WorkingHourResponse.yaml")
	requireProps(t, schema, []string{"id", "isoWeekday", "startsTime", "durationMinutes", "createdAt", "updatedAt"})
}

func TestContract_CreateWorkingHourRequestSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/CreateWorkingHourRequest.yaml")
	requireProps(t, schema, []string{"isoWeekday", "startsTime", "durationMinutes"})
	if _, ok := schema.Properties["barbershopId"]; ok {
		t.Fatal("el contrato de alta nunca debe declarar barbershopId")
	}
	if _, ok := schema.Properties["barberId"]; ok {
		t.Fatal("el contrato de alta nunca debe declarar barberId en el cuerpo")
	}
}

func TestContract_UpdateWorkingHourRequestSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/UpdateWorkingHourRequest.yaml")
	requireProps(t, schema, []string{"isoWeekday", "startsTime", "durationMinutes"})
}

func TestContract_WorkingHourListResponseSchema_HasItemsAndNextCursor(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/WorkingHourListResponse.yaml")
	requireProps(t, schema, []string{"items", "nextCursor"})
}

// --- Operaciones -----------------------------------------------------------

func TestContract_ListWorkingHoursOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[schedulePathsFile](t, "api/openapi/paths/schedules.yaml")
	requireResponses(t, doc.Collection.Get, "listWorkingHours", []string{"200", "400", "401", "404", "500"})
}

func TestContract_CreateWorkingHourOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[schedulePathsFile](t, "api/openapi/paths/schedules.yaml")
	requireResponses(t, doc.Collection.Post, "createWorkingHour", []string{"201", "400", "401", "404", "409", "422", "500"})
}

func TestContract_GetWorkingHourOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[schedulePathsFile](t, "api/openapi/paths/schedules.yaml")
	requireResponses(t, doc.Item.Get, "getWorkingHour", []string{"200", "401", "404", "500"})
}

func TestContract_UpdateWorkingHourOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[schedulePathsFile](t, "api/openapi/paths/schedules.yaml")
	requireResponses(t, doc.Item.Patch, "updateWorkingHour", []string{"200", "400", "401", "404", "409", "422", "500"})
}

func TestContract_DeleteWorkingHourOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[schedulePathsFile](t, "api/openapi/paths/schedules.yaml")
	requireResponses(t, doc.Item.Delete, "deleteWorkingHour", []string{"204", "401", "404", "500"})
}

// TestContract_OpenAPIYAML_RegistersSchedulePaths confirma que openapi.yaml
// registra ambos paths bajo el mismo documento raíz que las demás
// operaciones privadas, con la misma técnica de referencia JSON pointer.
func TestContract_OpenAPIYAML_RegistersSchedulePaths(t *testing.T) {
	type pathsDoc struct {
		Paths map[string]any `yaml:"paths"`
	}
	doc := loadYAML[pathsDoc](t, "api/openapi/openapi.yaml")

	if _, ok := doc.Paths["/private/barbers/{barberId}/working-hours"]; !ok {
		t.Fatal("openapi.yaml no registra paths./private/barbers/{barberId}/working-hours")
	}
	if _, ok := doc.Paths["/private/barbers/{barberId}/working-hours/{workingHourId}"]; !ok {
		t.Fatal("openapi.yaml no registra paths./private/barbers/{barberId}/working-hours/{workingHourId}")
	}
}

// --- HU-041: excepciones de jornada y festivos ------------------------------

func TestContract_HolidayCalendarResponseSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/HolidayCalendarResponse.yaml")
	requireProps(t, schema, []string{"enabled"})
}

func TestContract_UpdateHolidayCalendarRequestSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/UpdateHolidayCalendarRequest.yaml")
	requireProps(t, schema, []string{"enabled"})
}

func TestContract_ScheduleExceptionResponseSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/ScheduleExceptionResponse.yaml")
	requireProps(t, schema, []string{"id", "effectiveDate", "isClosed", "reason", "segments", "createdAt", "updatedAt"})
}

func TestContract_ScheduleExceptionListResponseSchema_HasItemsAndNextCursor(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/ScheduleExceptionListResponse.yaml")
	requireProps(t, schema, []string{"items", "nextCursor"})
}

// requireExceptionRequestShape verifica la forma común de
// CreateScheduleExceptionRequest/UpdateScheduleExceptionRequest: cuatro
// propiedades (effectiveDate, isClosed, reason, segments), pero solo las
// dos primeras son obligatorias -reason y segments son opcionales en el
// contrato, mismo criterio que ValidateExceptionShape validándolos según
// isClosed en tiempo de ejecución, no en el esquema-.
func requireExceptionRequestShape(t *testing.T, schema schemaDoc) {
	t.Helper()
	want := []string{"effectiveDate", "isClosed", "reason", "segments"}
	if len(schema.Properties) != len(want) {
		t.Fatalf("expected %d properties, schema has %d: %v", len(want), len(schema.Properties), schema.Properties)
	}
	for _, prop := range want {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("schema no declara la propiedad %q", prop)
		}
	}
	for _, wantReq := range []string{"effectiveDate", "isClosed"} {
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
	if len(schema.Required) != 2 {
		t.Errorf("expected exactly 2 required properties (effectiveDate, isClosed), got %v", schema.Required)
	}
	if _, ok := schema.Properties["barbershopId"]; ok {
		t.Fatal("el contrato nunca debe declarar barbershopId")
	}
	if _, ok := schema.Properties["barberId"]; ok {
		t.Fatal("el contrato nunca debe declarar barberId en el cuerpo")
	}
}

func TestContract_CreateScheduleExceptionRequestSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/CreateScheduleExceptionRequest.yaml")
	requireExceptionRequestShape(t, schema)
}

func TestContract_UpdateScheduleExceptionRequestSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/UpdateScheduleExceptionRequest.yaml")
	requireExceptionRequestShape(t, schema)
}

func TestContract_ColombianHolidayListResponseSchema_HasItems(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/ColombianHolidayListResponse.yaml")
	requireProps(t, schema, []string{"items"})
}

func TestContract_GetHolidayCalendarOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[scheduleExceptionPathsFile](t, "api/openapi/paths/schedule-exceptions.yaml")
	requireResponses(t, doc.HolidayCalendar.Get, "getHolidayCalendar", []string{"200", "401", "404", "500"})
}

func TestContract_UpdateHolidayCalendarOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[scheduleExceptionPathsFile](t, "api/openapi/paths/schedule-exceptions.yaml")
	requireResponses(t, doc.HolidayCalendar.Patch, "updateHolidayCalendar", []string{"200", "400", "401", "404", "500"})
}

func TestContract_ListScheduleExceptionsOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[scheduleExceptionPathsFile](t, "api/openapi/paths/schedule-exceptions.yaml")
	requireResponses(t, doc.Collection.Get, "listScheduleExceptions", []string{"200", "400", "401", "404", "500"})
}

func TestContract_CreateScheduleExceptionOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[scheduleExceptionPathsFile](t, "api/openapi/paths/schedule-exceptions.yaml")
	requireResponses(t, doc.Collection.Post, "createScheduleException", []string{"201", "400", "401", "404", "409", "422", "500"})
}

func TestContract_GetScheduleExceptionOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[scheduleExceptionPathsFile](t, "api/openapi/paths/schedule-exceptions.yaml")
	requireResponses(t, doc.Item.Get, "getScheduleException", []string{"200", "401", "404", "500"})
}

func TestContract_UpdateScheduleExceptionOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[scheduleExceptionPathsFile](t, "api/openapi/paths/schedule-exceptions.yaml")
	requireResponses(t, doc.Item.Patch, "updateScheduleException", []string{"200", "400", "401", "404", "409", "422", "500"})
}

func TestContract_DeleteScheduleExceptionOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[scheduleExceptionPathsFile](t, "api/openapi/paths/schedule-exceptions.yaml")
	requireResponses(t, doc.Item.Delete, "deleteScheduleException", []string{"204", "401", "404", "500"})
}

func TestContract_ListColombianHolidaysOperation_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[scheduleExceptionPathsFile](t, "api/openapi/paths/schedule-exceptions.yaml")
	requireResponses(t, doc.ColombianHolidays.Get, "listColombianHolidays", []string{"200", "400", "401", "500"})
}

// TestContract_OpenAPIYAML_RegistersScheduleExceptionPaths confirma que
// openapi.yaml registra los cuatro paths nuevos de HU-041.
func TestContract_OpenAPIYAML_RegistersScheduleExceptionPaths(t *testing.T) {
	type pathsDoc struct {
		Paths map[string]any `yaml:"paths"`
	}
	doc := loadYAML[pathsDoc](t, "api/openapi/openapi.yaml")

	for _, p := range []string{
		"/private/barbers/{barberId}/holiday-calendar",
		"/private/barbers/{barberId}/schedule-exceptions",
		"/private/barbers/{barberId}/schedule-exceptions/{exceptionId}",
		"/private/schedule/colombian-holidays",
	} {
		if _, ok := doc.Paths[p]; !ok {
			t.Fatalf("openapi.yaml no registra paths.%s", p)
		}
	}
}
