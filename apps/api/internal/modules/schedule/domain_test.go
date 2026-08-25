package schedule_test

import (
	"testing"

	"system-barbershop/internal/modules/schedule"
)

func TestLooksLikeBarberID(t *testing.T) {
	if !schedule.LooksLikeBarberID("8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4") {
		t.Fatal("expected a canonical UUID to look like a barber id")
	}
	if schedule.LooksLikeBarberID("not-a-uuid") {
		t.Fatal("expected a malformed value to not look like a barber id")
	}
}

func TestLooksLikeWorkingHourID(t *testing.T) {
	if !schedule.LooksLikeWorkingHourID("6f1a2b3c-4d5e-4f60-8172-8394a5b6c7d8") {
		t.Fatal("expected a canonical UUID to look like a working hour id")
	}
	if schedule.LooksLikeWorkingHourID("") {
		t.Fatal("expected an empty value to not look like a working hour id")
	}
}

func TestValidateStartsTime_Valid(t *testing.T) {
	for _, raw := range []string{"00:00", "08:00", "23:59", "22:00"} {
		got, err := schedule.ValidateStartsTime(raw)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", raw, err)
		}
		if got != raw {
			t.Fatalf("expected %q unchanged, got %q", raw, got)
		}
	}
}

func TestValidateStartsTime_Invalid(t *testing.T) {
	for _, raw := range []string{"8:00", "24:00", "23:60", "8am", "", "08:00:00", "230:00"} {
		if _, err := schedule.ValidateStartsTime(raw); err == nil {
			t.Fatalf("expected %q to be rejected", raw)
		}
	}
}

// --- IntervalsOverlap: CA-040-03/04, semántica semiabierta [inicio, fin) ---

func TestIntervalsOverlap_Contiguous_NotOverlapping(t *testing.T) {
	// [480,720) y [720,960): el fin de uno coincide con el inicio del otro.
	if schedule.IntervalsOverlap(480, 240, 720, 240) {
		t.Fatal("expected contiguous intervals to not overlap")
	}
}

func TestIntervalsOverlap_Identical_Overlaps(t *testing.T) {
	if !schedule.IntervalsOverlap(480, 60, 480, 60) {
		t.Fatal("expected identical intervals to overlap")
	}
}

func TestIntervalsOverlap_PartialOverlap_Detected(t *testing.T) {
	// [480, 600) y [540, 660): se solapan en [540, 600).
	if !schedule.IntervalsOverlap(480, 120, 540, 120) {
		t.Fatal("expected partially overlapping intervals to be detected")
	}
}

func TestIntervalsOverlap_Disjoint_NotOverlapping(t *testing.T) {
	if schedule.IntervalsOverlap(480, 60, 600, 60) {
		t.Fatal("expected disjoint intervals to not overlap")
	}
}

func TestIntervalsOverlap_NightShift_ExtendsPastMidnight(t *testing.T) {
	// 22:00 (1320) + 300 min termina en 1620 (03:00 del día siguiente,
	// DEC-020): un tramo que empieza a las 23:00 (1380) del MISMO día ISO
	// cae dentro de ese rango extendido.
	if !schedule.IntervalsOverlap(1320, 300, 1380, 60) {
		t.Fatal("expected a night-shift interval to overlap a segment starting within its extended range")
	}
}

// --- Cursor: codificación opaca ------------------------------------------

func TestEncodeDecodeCursor_RoundTrips(t *testing.T) {
	original := schedule.Cursor{ISOWeekday: 3, StartsTime: "08:00", ID: "6f1a2b3c-4d5e-4f60-8172-8394a5b6c7d8"}
	token := schedule.EncodeCursor(original)
	if token == "" {
		t.Fatal("expected a non-empty cursor token")
	}

	decoded, err := schedule.DecodeCursor(token)
	if err != nil {
		t.Fatalf("unexpected error decoding cursor: %v", err)
	}
	if decoded != original {
		t.Fatalf("expected %+v, got %+v", original, decoded)
	}
}

func TestDecodeCursor_Malformed_Rejected(t *testing.T) {
	if _, err := schedule.DecodeCursor("not-a-valid-cursor!!!"); err == nil {
		t.Fatal("expected a malformed cursor to be rejected")
	}
}

func TestDecodeCursor_OutOfRangeWeekday_Rejected(t *testing.T) {
	token := schedule.EncodeCursor(schedule.Cursor{ISOWeekday: 9, StartsTime: "08:00", ID: "6f1a2b3c-4d5e-4f60-8172-8394a5b6c7d8"})
	if _, err := schedule.DecodeCursor(token); err == nil {
		t.Fatal("expected a cursor with an out-of-range weekday to be rejected")
	}
}
