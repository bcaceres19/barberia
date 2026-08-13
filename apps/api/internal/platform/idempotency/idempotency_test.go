package idempotency_test

import (
	"errors"
	"testing"

	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/idempotency"
)

func TestParseKey_RejectsEmpty(t *testing.T) {
	_, err := idempotency.ParseKey("")

	appErr, ok := apperr.As(err)
	if !ok {
		t.Fatalf("expected an apperr.Error, got %v", err)
	}
	if appErr.Kind != apperr.KindInvalid {
		t.Fatalf("expected KindInvalid, got %q", appErr.Kind)
	}
}

func TestParseKey_RejectsDisallowedCharacters(t *testing.T) {
	cases := []string{
		"clave con espacios",
		"clave/con/barras",
		"clave\ncon\nsaltos",
		"clave;con;punto-y-coma",
		"café", // fuera de A-Za-z0-9._-
	}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			_, err := idempotency.ParseKey(raw)
			if err == nil {
				t.Fatalf("expected an error for %q", raw)
			}
			appErr, ok := apperr.As(err)
			if !ok || appErr.Kind != apperr.KindInvalid {
				t.Fatalf("expected KindInvalid, got %v", err)
			}
		})
	}
}

func TestParseKey_RejectsTooLong(t *testing.T) {
	raw := make([]byte, 256)
	for i := range raw {
		raw[i] = 'a'
	}
	_, err := idempotency.ParseKey(string(raw))
	if err == nil {
		t.Fatal("expected an error for a 256-character key")
	}
}

func TestParseKey_AcceptsValidValue(t *testing.T) {
	key, err := idempotency.ParseKey("create-appointment_2026-08-13.001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(key) != "create-appointment_2026-08-13.001" {
		t.Fatalf("unexpected key value: %q", key)
	}
}

// TestParseKey_ErrorNeverEchoesRawValue confirma que el mensaje de error no
// refleja el valor crudo del cliente: podría contener datos sensibles que
// el cliente eligió incluir en la clave.
func TestParseKey_ErrorNeverEchoesRawValue(t *testing.T) {
	sensitive := "tarjeta-4111111111111111"
	_, err := idempotency.ParseKey(sensitive + " con espacio inválido")

	appErr, ok := apperr.As(err)
	if !ok {
		t.Fatalf("expected an apperr.Error, got %v", err)
	}
	if want := "4111111111111111"; contains(appErr.Message, want) {
		t.Fatalf("el mensaje de error no debe repetir el valor crudo: %q", appErr.Message)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}

func TestKey_Redacted_IsStableAndNeverEqualsTheRawValue(t *testing.T) {
	key, err := idempotency.ParseKey("clave-de-prueba-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	first := key.Redacted()
	second := key.Redacted()
	if first != second {
		t.Fatalf("expected Redacted to be deterministic, got %q then %q", first, second)
	}
	if first == string(key) {
		t.Fatal("Redacted must never equal the raw key value")
	}
	if len(first) == 0 {
		t.Fatal("expected a non-empty redacted reference")
	}
}

func TestKey_Redacted_DiffersForDifferentKeys(t *testing.T) {
	a, _ := idempotency.ParseKey("clave-a")
	b, _ := idempotency.ParseKey("clave-b")

	if a.Redacted() == b.Redacted() {
		t.Fatal("expected different keys to produce different redacted references")
	}
}

func TestComputeFingerprint_IsDeterministic(t *testing.T) {
	f1 := idempotency.ComputeFingerprint("POST", "/api/v1/private/appointments", []byte(`{"a":1}`))
	f2 := idempotency.ComputeFingerprint("POST", "/api/v1/private/appointments", []byte(`{"a":1}`))

	if f1 != f2 {
		t.Fatalf("expected the same input to produce the same fingerprint, got %q vs %q", f1, f2)
	}
}

func TestComputeFingerprint_MatchesStoredFormat(t *testing.T) {
	f := idempotency.ComputeFingerprint("POST", "/api/v1/private/appointments", []byte(`{"a":1}`))

	// idempotency_record_request_fingerprint_ck (DDL-VAL-01): hexadecimal
	// en minúsculas, 32 a 128 caracteres. SHA-256 produce 64.
	if len(f) != 64 {
		t.Fatalf("expected a 64-character SHA-256 hex digest, got %d characters", len(f))
	}
	for _, r := range string(f) {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			t.Fatalf("fingerprint contains a non-lowercase-hex character: %q in %q", r, f)
		}
	}
}

func TestComputeFingerprint_MethodIsCaseInsensitive(t *testing.T) {
	upper := idempotency.ComputeFingerprint("POST", "/x", []byte("body"))
	lower := idempotency.ComputeFingerprint("post", "/x", []byte("body"))

	if upper != lower {
		t.Fatal("expected method casing not to change the fingerprint")
	}
}

func TestComputeFingerprint_DifferentContentProducesDifferentFingerprint(t *testing.T) {
	base := idempotency.ComputeFingerprint("POST", "/api/v1/private/appointments", []byte(`{"a":1}`))

	variants := map[string]idempotency.Fingerprint{
		"method distinto": idempotency.ComputeFingerprint("PUT", "/api/v1/private/appointments", []byte(`{"a":1}`)),
		"ruta distinta":   idempotency.ComputeFingerprint("POST", "/api/v1/private/other", []byte(`{"a":1}`)),
		"body distinto":   idempotency.ComputeFingerprint("POST", "/api/v1/private/appointments", []byte(`{"a":2}`)),
		// El separador "\n" evita que method+path concatenados sin límite
		// colisionen: sin separador, method="POSTX" path="" y method="POST"
		// path="X" producirían la misma concatenación.
		"colision sin separador": idempotency.ComputeFingerprint("POSTX", "", []byte(`{"a":1}`)),
	}
	for name, other := range variants {
		t.Run(name, func(t *testing.T) {
			if other == base {
				t.Fatalf("expected a different fingerprint for %s", name)
			}
		})
	}
}

func TestDecision_AsError_ProceedAndReplayAreNil(t *testing.T) {
	if err := (idempotency.Decision{Outcome: idempotency.OutcomeProceed}).AsError(); err != nil {
		t.Fatalf("expected nil for OutcomeProceed, got %v", err)
	}
	if err := (idempotency.Decision{Outcome: idempotency.OutcomeReplay}).AsError(); err != nil {
		t.Fatalf("expected nil for OutcomeReplay, got %v", err)
	}
}

func TestDecision_AsError_MapsEachConflictOutcomeToTheDocumentedKind(t *testing.T) {
	cases := map[idempotency.Outcome]apperr.Kind{
		idempotency.OutcomeConflictOperation:   apperr.KindIdempotencyConflict,
		idempotency.OutcomeConflictFingerprint: apperr.KindIdempotencyConflict,
		idempotency.OutcomeConflictInProgress:  apperr.KindIdempotencyLocked,
		idempotency.OutcomeLocked:              apperr.KindIdempotencyLocked,
	}

	for outcome, wantKind := range cases {
		t.Run(outcome.String(), func(t *testing.T) {
			err := (idempotency.Decision{Outcome: outcome}).AsError()
			appErr, ok := apperr.As(err)
			if !ok {
				t.Fatalf("expected an apperr.Error, got %v", err)
			}
			if appErr.Kind != wantKind {
				t.Fatalf("expected %q, got %q", wantKind, appErr.Kind)
			}
		})
	}
}

func TestDecision_AsError_UnknownOutcomeIsInternal(t *testing.T) {
	err := (idempotency.Decision{Outcome: idempotency.Outcome(99)}).AsError()

	appErr, ok := apperr.As(err)
	if !ok {
		t.Fatalf("expected an apperr.Error, got %v", err)
	}
	if appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected KindInternal for an unrecognized outcome, got %q", appErr.Kind)
	}
	if !errors.Is(err, appErr.Err) {
		t.Fatal("expected the internal cause to be unwrappable for diagnostics")
	}
}
