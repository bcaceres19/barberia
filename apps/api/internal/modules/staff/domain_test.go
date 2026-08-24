package staff_test

import (
	"testing"
	"time"

	"system-barbershop/internal/modules/staff"
	"system-barbershop/internal/platform/apperr"
)

func TestNormalizeFullName_TrimsSurroundingSpaces(t *testing.T) {
	got := staff.NormalizeFullName("  Carlos Ramírez  ")
	if got != "Carlos Ramírez" {
		t.Fatalf("expected trimmed name, got %q", got)
	}
}

func TestNormalizeFullName_PreservesInternalUnicode(t *testing.T) {
	got := staff.NormalizeFullName("José Núñez")
	if got != "José Núñez" {
		t.Fatalf("expected unicode preserved, got %q", got)
	}
}

func TestLooksLikeBarberID_AcceptsCanonicalUUID(t *testing.T) {
	if !staff.LooksLikeBarberID("8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4") {
		t.Fatal("expected a canonical UUID to be accepted")
	}
}

func TestLooksLikeBarberID_RejectsMalformedValues(t *testing.T) {
	cases := []string{"", "not-a-uuid", "8f3ac2b1e4d546f6a7c8d9e0f1a2b3c4", "'; DROP TABLE barber; --"}
	for _, c := range cases {
		if staff.LooksLikeBarberID(c) {
			t.Fatalf("expected %q to be rejected", c)
		}
	}
}

func TestEncodeDecodeCursor_RoundTrips(t *testing.T) {
	original := staff.Cursor{
		CreatedAt: time.Date(2026, 8, 23, 15, 4, 5, 0, time.UTC),
		ID:        "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4",
	}

	encoded := staff.EncodeCursor(original)
	if encoded == "" {
		t.Fatal("expected a non-empty encoded cursor")
	}

	decoded, err := staff.DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("unexpected error decoding: %v", err)
	}
	if !decoded.CreatedAt.Equal(original.CreatedAt) || decoded.ID != original.ID {
		t.Fatalf("expected round-trip to preserve the cursor, got %+v", decoded)
	}
}

func TestDecodeCursor_RejectsGarbageAsClientError(t *testing.T) {
	cases := []string{"", "not-base64!!!", "AAAA", "e30"} // "e30" == base64("{}")
	for _, c := range cases {
		_, err := staff.DecodeCursor(c)
		if err == nil {
			t.Fatalf("expected %q to be rejected", c)
		}
		appErr, ok := apperr.As(err)
		if !ok || appErr.Kind != apperr.KindInvalid {
			t.Fatalf("expected apperr.KindInvalid for %q, got %v", c, err)
		}
	}
}

func TestDecodeCursor_DoesNotAcceptAnArbitraryForgedCursor(t *testing.T) {
	// Un cursor "inventado" con forma correcta pero que nunca salió de
	// EncodeCursor sigue siendo aceptado como VÁLIDO en cuanto a forma (el
	// cursor es solo una posición de recorrido, no un secreto): lo que
	// importa es que nunca revela ni modifica un recurso ajeno, verificado
	// en postgres/repository_test.go con dos tenants reales, no aquí.
	forged := staff.EncodeCursor(staff.Cursor{CreatedAt: time.Now(), ID: "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"})
	if _, err := staff.DecodeCursor(forged); err != nil {
		t.Fatalf("unexpected error decoding a well-formed cursor: %v", err)
	}
}
