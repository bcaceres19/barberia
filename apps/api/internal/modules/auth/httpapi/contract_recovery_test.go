// Pruebas de contrato de HU-008: comparan los handlers de recuperación de
// acceso contra api/openapi/paths/public-auth.yaml, mismo criterio que
// contract_challenge_test.go (HU-007).
package httpapi_test

import (
	"testing"
)

type recoveryPathFile struct {
	Request struct {
		Post operation `yaml:"post"`
	} `yaml:"/public/auth/recovery/request"`
	Verify struct {
		Post operation `yaml:"post"`
	} `yaml:"/public/auth/recovery/verify"`
	ResetPassword struct {
		Post operation `yaml:"post"`
	} `yaml:"/public/auth/recovery/reset-password"`
}

func TestContract_RecoveryRequestSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/RecoveryRequestRequest.yaml")

	if _, ok := schema.Properties["email"]; !ok {
		t.Fatal("schema no declara email")
	}
	if len(schema.Properties) != 1 {
		t.Fatalf("expected exactly 1 property (email), got %v", schema.Properties)
	}
	if len(schema.Required) != 1 || schema.Required[0] != "email" {
		t.Fatalf("expected email to be required, got %v", schema.Required)
	}
}

func TestContract_RecoveryVerifyRequestSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/RecoveryVerifyRequest.yaml")

	wantProps := map[string]bool{"email": true, "code": true}
	if len(schema.Properties) != len(wantProps) {
		t.Fatalf("expected %d properties, schema has %d: %v", len(wantProps), len(schema.Properties), schema.Properties)
	}
	for prop := range wantProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("schema no declara la propiedad %q", prop)
		}
	}
}

func TestContract_RecoveryVerifyResponseSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/RecoveryVerifyResponse.yaml")

	wantProps := map[string]bool{"resetToken": true, "maskedPhone": true, "maskedEmail": true}
	if len(schema.Properties) != len(wantProps) {
		t.Fatalf("expected %d properties, schema has %d: %v", len(wantProps), len(schema.Properties), schema.Properties)
	}
	for prop := range wantProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("schema no declara la propiedad %q", prop)
		}
	}
}

func TestContract_RecoveryResetPasswordRequestSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/RecoveryResetPasswordRequest.yaml")

	wantProps := map[string]bool{"email": true, "resetToken": true, "newPassword": true}
	if len(schema.Properties) != len(wantProps) {
		t.Fatalf("expected %d properties, schema has %d: %v", len(wantProps), len(schema.Properties), schema.Properties)
	}
	for prop := range wantProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("schema no declara la propiedad %q", prop)
		}
	}
	if len(schema.Required) != 3 {
		t.Fatalf("expected all 3 fields required, got %v", schema.Required)
	}
}

func TestContract_RecoveryOperations_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[recoveryPathFile](t, "api/openapi/paths/public-auth.yaml")

	requestOp := doc.Request.Post
	if requestOp.OperationID != "requestRecovery" {
		t.Fatalf("expected operationId requestRecovery, got %q", requestOp.OperationID)
	}
	if requestOp.Security == nil || len(requestOp.Security) != 0 {
		t.Fatalf("expected security: [] (operación pública, no enumeración), got %v", requestOp.Security)
	}
	wantRequestStatuses := []string{"202", "400", "422", "500"}
	for _, s := range wantRequestStatuses {
		if _, ok := requestOp.Responses[s]; !ok {
			t.Errorf("el contrato de /auth/recovery/request no documenta la respuesta %s", s)
		}
	}
	if len(requestOp.Responses) != len(wantRequestStatuses) {
		t.Errorf("el contrato de /auth/recovery/request documenta %d respuestas, se esperaban %d (%v)",
			len(requestOp.Responses), len(wantRequestStatuses), wantRequestStatuses)
	}

	verifyOp := doc.Verify.Post
	if verifyOp.OperationID != "verifyRecovery" {
		t.Fatalf("expected operationId verifyRecovery, got %q", verifyOp.OperationID)
	}
	wantVerifyStatuses := []string{"200", "400", "401", "422", "500"}
	for _, s := range wantVerifyStatuses {
		if _, ok := verifyOp.Responses[s]; !ok {
			t.Errorf("el contrato de /auth/recovery/verify no documenta la respuesta %s", s)
		}
	}
	if len(verifyOp.Responses) != len(wantVerifyStatuses) {
		t.Errorf("el contrato de /auth/recovery/verify documenta %d respuestas, se esperaban %d (%v)",
			len(verifyOp.Responses), len(wantVerifyStatuses), wantVerifyStatuses)
	}

	resetOp := doc.ResetPassword.Post
	if resetOp.OperationID != "resetPasswordWithRecoveryToken" {
		t.Fatalf("expected operationId resetPasswordWithRecoveryToken, got %q", resetOp.OperationID)
	}
	wantResetStatuses := []string{"204", "400", "401", "422", "500"}
	for _, s := range wantResetStatuses {
		if _, ok := resetOp.Responses[s]; !ok {
			t.Errorf("el contrato de /auth/recovery/reset-password no documenta la respuesta %s", s)
		}
	}
	if len(resetOp.Responses) != len(wantResetStatuses) {
		t.Errorf("el contrato de /auth/recovery/reset-password documenta %d respuestas, se esperaban %d (%v)",
			len(resetOp.Responses), len(wantResetStatuses), wantResetStatuses)
	}
}
