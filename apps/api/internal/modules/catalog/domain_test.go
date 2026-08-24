package catalog_test

import (
	"testing"
	"time"

	"system-barbershop/internal/modules/catalog"
	"system-barbershop/internal/platform/apperr"
)

func TestNormalizeName_TrimsSurroundingSpaces(t *testing.T) {
	got := catalog.NormalizeName("  Corte clásico  ")
	if got != "Corte clásico" {
		t.Fatalf("expected trimmed name, got %q", got)
	}
}

func TestNormalizeDescription_EmptyOrWhitespaceBecomesNil(t *testing.T) {
	for _, raw := range []string{"", "   ", "\t\n"} {
		if got := catalog.NormalizeDescription(raw); got != nil {
			t.Fatalf("expected nil for %q, got %q", raw, *got)
		}
	}
}

func TestNormalizeDescription_TrimsAndPreservesUnicode(t *testing.T) {
	got := catalog.NormalizeDescription("  Corte con máquina y tijera  ")
	if got == nil || *got != "Corte con máquina y tijera" {
		t.Fatalf("expected trimmed unicode description, got %v", got)
	}
}

func TestLooksLikeServiceID_AcceptsCanonicalUUID(t *testing.T) {
	if !catalog.LooksLikeServiceID("8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4") {
		t.Fatal("expected a canonical UUID to be accepted")
	}
}

func TestLooksLikeServiceID_RejectsMalformedValues(t *testing.T) {
	cases := []string{"", "not-a-uuid", "8f3ac2b1e4d546f6a7c8d9e0f1a2b3c4", "'; DROP TABLE service; --"}
	for _, c := range cases {
		if catalog.LooksLikeServiceID(c) {
			t.Fatalf("expected %q to be rejected", c)
		}
	}
}

func TestEncodeDecodeCursor_RoundTrips(t *testing.T) {
	original := catalog.Cursor{
		CreatedAt: time.Date(2026, 8, 24, 15, 4, 5, 0, time.UTC),
		ID:        "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4",
	}

	encoded := catalog.EncodeCursor(original)
	if encoded == "" {
		t.Fatal("expected a non-empty encoded cursor")
	}

	decoded, err := catalog.DecodeCursor(encoded)
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
		_, err := catalog.DecodeCursor(c)
		if err == nil {
			t.Fatalf("expected %q to be rejected", c)
		}
		appErr, ok := apperr.As(err)
		if !ok || appErr.Kind != apperr.KindInvalid {
			t.Fatalf("expected apperr.KindInvalid for %q, got %v", c, err)
		}
	}
}

// --- ParsePriceCOP/FormatPriceCOP: dinero exacto, nunca coma flotante -----

func TestParsePriceCOP_AcceptsIntegerAndTwoDecimals(t *testing.T) {
	cases := map[string]int64{
		"1":             100,
		"45000":         4500000,
		"45000.5":       4500050,
		"45000.50":      4500050,
		"45000.99":      4500099,
		"0.01":          1,
		"9999999999.99": 999999999999,
	}
	for raw, want := range cases {
		got, err := catalog.ParsePriceCOP(raw)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", raw, err)
		}
		if got != want {
			t.Fatalf("expected %d cents for %q, got %d", want, raw, got)
		}
	}
}

func TestParsePriceCOP_RejectsZeroAndNegative(t *testing.T) {
	for _, raw := range []string{"0", "0.00", "-1", "-45000.00"} {
		_, err := catalog.ParsePriceCOP(raw)
		if err == nil {
			t.Fatalf("expected %q to be rejected (DEC-067: price > 0)", raw)
		}
		appErr, ok := apperr.As(err)
		if !ok || appErr.Kind != apperr.KindValidation {
			t.Fatalf("expected apperr.KindValidation for %q, got %v", raw, err)
		}
	}
}

func TestParsePriceCOP_RejectsInvalidFormat(t *testing.T) {
	cases := []string{"", "abc", "45,000", "45000.123", "1e10", "+45000", "45000.", ".50"}
	for _, raw := range cases {
		_, err := catalog.ParsePriceCOP(raw)
		if err == nil {
			t.Fatalf("expected %q to be rejected as invalid format", raw)
		}
		appErr, ok := apperr.As(err)
		if !ok || appErr.Kind != apperr.KindValidation {
			t.Fatalf("expected apperr.KindValidation for %q, got %v", raw, err)
		}
	}
}

func TestParsePriceCOP_TrimsSurroundingWhitespace(t *testing.T) {
	got, err := catalog.ParsePriceCOP("  45000.00  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 4500000 {
		t.Fatalf("expected 4500000 cents, got %d", got)
	}
}

func TestFormatPriceCOP_AlwaysTwoDecimals(t *testing.T) {
	cases := map[int64]string{
		100:          "1.00",
		4500000:      "45000.00",
		4500050:      "45000.50",
		1:            "0.01",
		999999999999: "9999999999.99",
	}
	for cents, want := range cases {
		got := catalog.FormatPriceCOP(cents)
		if got != want {
			t.Fatalf("expected %q for %d cents, got %q", want, cents, got)
		}
	}
}

func TestParseFormatPriceCOP_RoundTrips(t *testing.T) {
	cases := []string{"45000.00", "1.00", "0.01", "9999999999.99", "50000.50"}
	for _, raw := range cases {
		cents, err := catalog.ParsePriceCOP(raw)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", raw, err)
		}
		if got := catalog.FormatPriceCOP(cents); got != raw {
			t.Fatalf("expected round-trip %q, got %q", raw, got)
		}
	}
}
