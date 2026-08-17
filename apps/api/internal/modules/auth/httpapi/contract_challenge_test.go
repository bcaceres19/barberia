// Pruebas de contrato de HU-007: comparan los handlers del reto telefónico
// contra api/openapi/paths/public-auth.yaml, mismo criterio que
// contract_test.go (HU-005).
package httpapi_test

import (
	"testing"
)

type challengePathFile struct {
	Challenge struct {
		Post operation `yaml:"post"`
	} `yaml:"/public/auth/challenge"`
	ChallengeVerify struct {
		Post operation `yaml:"post"`
	} `yaml:"/public/auth/challenge/verify"`
}

func TestContract_ChallengeRequestSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/ChallengeRequest.yaml")

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

func TestContract_ChallengeVerifyRequestSchema_MatchesDTOFields(t *testing.T) {
	schema := loadYAML[schemaDoc](t, "api/openapi/components/schemas/ChallengeVerifyRequest.yaml")

	wantProps := map[string]bool{"email": true, "code": true}
	if len(schema.Properties) != len(wantProps) {
		t.Fatalf("expected %d properties, schema has %d: %v", len(wantProps), len(schema.Properties), schema.Properties)
	}
	for prop := range wantProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("schema no declara la propiedad %q", prop)
		}
	}
	for _, want := range []string{"email", "code"} {
		found := false
		for _, r := range schema.Required {
			if r == want {
				found = true
			}
		}
		if !found {
			t.Errorf("schema no marca %q como required", want)
		}
	}
}

func TestContract_ChallengeOperations_MethodPathSecurityAndResponses(t *testing.T) {
	doc := loadYAML[challengePathFile](t, "api/openapi/paths/public-auth.yaml")

	requestOp := doc.Challenge.Post
	if requestOp.OperationID != "requestPhoneChallenge" {
		t.Fatalf("expected operationId requestPhoneChallenge, got %q", requestOp.OperationID)
	}
	if requestOp.Security == nil || len(requestOp.Security) != 0 {
		t.Fatalf("expected security: [] (operación pública, no enumeración), got %v", requestOp.Security)
	}
	wantRequestStatuses := []string{"202", "400", "422", "500"}
	for _, s := range wantRequestStatuses {
		if _, ok := requestOp.Responses[s]; !ok {
			t.Errorf("el contrato de /auth/challenge no documenta la respuesta %s", s)
		}
	}
	if len(requestOp.Responses) != len(wantRequestStatuses) {
		t.Errorf("el contrato de /auth/challenge documenta %d respuestas, se esperaban %d (%v)",
			len(requestOp.Responses), len(wantRequestStatuses), wantRequestStatuses)
	}

	verifyOp := doc.ChallengeVerify.Post
	if verifyOp.OperationID != "verifyPhoneChallenge" {
		t.Fatalf("expected operationId verifyPhoneChallenge, got %q", verifyOp.OperationID)
	}
	if verifyOp.Security == nil || len(verifyOp.Security) != 0 {
		t.Fatalf("expected security: [] (operación pública), got %v", verifyOp.Security)
	}
	wantVerifyStatuses := []string{"204", "400", "401", "422", "500"}
	for _, s := range wantVerifyStatuses {
		if _, ok := verifyOp.Responses[s]; !ok {
			t.Errorf("el contrato de /auth/challenge/verify no documenta la respuesta %s", s)
		}
	}
	if len(verifyOp.Responses) != len(wantVerifyStatuses) {
		t.Errorf("el contrato de /auth/challenge/verify documenta %d respuestas, se esperaban %d (%v)",
			len(verifyOp.Responses), len(wantVerifyStatuses), wantVerifyStatuses)
	}
}
