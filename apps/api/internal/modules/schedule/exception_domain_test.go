package schedule_test

import (
	"strings"
	"testing"

	"system-barbershop/internal/modules/schedule"
)

func TestValidateEffectiveDate_Valid(t *testing.T) {
	for _, raw := range []string{"2026-01-01", "2026-12-31", "2028-02-29"} {
		got, err := schedule.ValidateEffectiveDate(raw)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", raw, err)
		}
		if got != raw {
			t.Fatalf("expected %q unchanged, got %q", raw, got)
		}
	}
}

func TestValidateEffectiveDate_Invalid(t *testing.T) {
	for _, raw := range []string{"2026-13-01", "2026-02-30", "not-a-date", "", "2026/01/01", "26-01-01"} {
		if _, err := schedule.ValidateEffectiveDate(raw); err == nil {
			t.Fatalf("expected %q to be rejected", raw)
		}
	}
}

func TestValidateReason_NilStaysNil(t *testing.T) {
	got, err := schedule.ValidateReason(nil)
	if err != nil || got != nil {
		t.Fatalf("expected nil, nil, got %v, %v", got, err)
	}
}

func TestValidateReason_EmptyOrWhitespaceNormalizesToNil(t *testing.T) {
	for _, raw := range []string{"", "   ", "\t\n"} {
		got, err := schedule.ValidateReason(&raw)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", raw, err)
		}
		if got != nil {
			t.Fatalf("expected nil for %q, got %q", raw, *got)
		}
	}
}

func TestValidateReason_Trims(t *testing.T) {
	raw := "  Festivo trabajado  "
	got, err := schedule.ValidateReason(&raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || *got != "Festivo trabajado" {
		t.Fatalf("expected trimmed reason, got %v", got)
	}
}

func TestValidateReason_TooLong_Rejected(t *testing.T) {
	raw := strings.Repeat("a", 201)
	if _, err := schedule.ValidateReason(&raw); err == nil {
		t.Fatal("expected a 201-character reason to be rejected")
	}
}

func TestValidateReason_Exactly200Characters_Accepted(t *testing.T) {
	raw := strings.Repeat("a", 200)
	got, err := schedule.ValidateReason(&raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || *got != raw {
		t.Fatal("expected the exact 200-character reason to be accepted unchanged")
	}
}

// --- ValidateExceptionShape: CA-041-04 ------------------------------------

func TestValidateExceptionShape_ClosedWithSegments_Rejected(t *testing.T) {
	_, err := schedule.ValidateExceptionShape(true, []schedule.CreateExceptionSegmentInput{
		{StartsTime: "08:00", DurationMinutes: 60},
	})
	if err == nil {
		t.Fatal("expected a closed exception with segments to be rejected")
	}
}

func TestValidateExceptionShape_ClosedWithoutSegments_Accepted(t *testing.T) {
	segments, err := schedule.ValidateExceptionShape(true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(segments) != 0 {
		t.Fatalf("expected no segments, got %+v", segments)
	}
}

func TestValidateExceptionShape_OpenWithoutSegments_Rejected(t *testing.T) {
	_, err := schedule.ValidateExceptionShape(false, nil)
	if err == nil {
		t.Fatal("expected an open exception without segments to be rejected")
	}
}

func TestValidateExceptionShape_OpenWithNonOverlappingSegments_Accepted(t *testing.T) {
	segments, err := schedule.ValidateExceptionShape(false, []schedule.CreateExceptionSegmentInput{
		{StartsTime: "08:00", DurationMinutes: 120},
		{StartsTime: "14:00", DurationMinutes: 120},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segments))
	}
}

func TestValidateExceptionShape_OverlappingSegments_Rejected(t *testing.T) {
	_, err := schedule.ValidateExceptionShape(false, []schedule.CreateExceptionSegmentInput{
		{StartsTime: "08:00", DurationMinutes: 120},
		{StartsTime: "09:00", DurationMinutes: 60},
	})
	if err == nil {
		t.Fatal("expected overlapping segments to be rejected")
	}
}

func TestValidateExceptionShape_ContiguousSegments_Accepted(t *testing.T) {
	_, err := schedule.ValidateExceptionShape(false, []schedule.CreateExceptionSegmentInput{
		{StartsTime: "08:00", DurationMinutes: 120},
		{StartsTime: "10:00", DurationMinutes: 30},
	})
	if err != nil {
		t.Fatalf("expected contiguous segments to be accepted, got: %v", err)
	}
}

func TestValidateExceptionShape_InvalidSegmentTime_Rejected(t *testing.T) {
	_, err := schedule.ValidateExceptionShape(false, []schedule.CreateExceptionSegmentInput{
		{StartsTime: "8am", DurationMinutes: 60},
	})
	if err == nil {
		t.Fatal("expected an invalid starts time to be rejected")
	}
}

func TestValidateExceptionShape_InvalidSegmentDuration_Rejected(t *testing.T) {
	_, err := schedule.ValidateExceptionShape(false, []schedule.CreateExceptionSegmentInput{
		{StartsTime: "08:00", DurationMinutes: 0},
	})
	if err == nil {
		t.Fatal("expected an invalid duration to be rejected")
	}
}

// --- Cursor de excepciones -------------------------------------------------

func TestEncodeDecodeExceptionCursor_RoundTrips(t *testing.T) {
	original := schedule.ExceptionCursor{EffectiveDate: "2026-12-08", ID: "6f1a2b3c-4d5e-4f60-8172-8394a5b6c7d8"}
	token := schedule.EncodeExceptionCursor(original)
	if token == "" {
		t.Fatal("expected a non-empty cursor token")
	}
	decoded, err := schedule.DecodeExceptionCursor(token)
	if err != nil {
		t.Fatalf("unexpected error decoding cursor: %v", err)
	}
	if decoded != original {
		t.Fatalf("expected %+v, got %+v", original, decoded)
	}
}

func TestDecodeExceptionCursor_Malformed_Rejected(t *testing.T) {
	if _, err := schedule.DecodeExceptionCursor("not-a-valid-cursor!!!"); err == nil {
		t.Fatal("expected a malformed cursor to be rejected")
	}
}
