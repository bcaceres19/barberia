package schedule_test

import (
	"testing"
	"time"

	"system-barbershop/internal/modules/schedule"
)

func TestValidateBlockType(t *testing.T) {
	valid := []string{"break", "lunch", "unavailable", "day_off", "holiday", "vacation", "emergency"}
	for _, bt := range valid {
		if _, err := schedule.ValidateBlockType(bt); err != nil {
			t.Errorf("ValidateBlockType(%q) inesperadamente rechazado: %v", bt, err)
		}
	}
	invalid := []string{"", "BREAK", "holidays", "appointment"}
	for _, bt := range invalid {
		if _, err := schedule.ValidateBlockType(bt); err == nil {
			t.Errorf("ValidateBlockType(%q) debería rechazarse", bt)
		}
	}
}

func TestValidateRecurrenceKind(t *testing.T) {
	for _, rk := range []string{"weekly", "date_list"} {
		if _, err := schedule.ValidateRecurrenceKind(rk); err != nil {
			t.Errorf("ValidateRecurrenceKind(%q) inesperadamente rechazado: %v", rk, err)
		}
	}
	for _, rk := range []string{"", "daily", "WEEKLY"} {
		if _, err := schedule.ValidateRecurrenceKind(rk); err == nil {
			t.Errorf("ValidateRecurrenceKind(%q) debería rechazarse", rk)
		}
	}
}

func TestValidateCivilDate(t *testing.T) {
	if _, err := schedule.ValidateCivilDate("2026-12-08"); err != nil {
		t.Fatalf("fecha válida rechazada: %v", err)
	}
	invalid := []string{"2026-02-30", "08-12-2026", "2026/12/08", "not-a-date", ""}
	for _, d := range invalid {
		if _, err := schedule.ValidateCivilDate(d); err == nil {
			t.Errorf("ValidateCivilDate(%q) debería rechazarse", d)
		}
	}
}

func TestValidateInstant(t *testing.T) {
	if _, err := schedule.ValidateInstant("2026-07-20T15:00:00-05:00"); err != nil {
		t.Fatalf("instante válido rechazado: %v", err)
	}
	invalid := []string{"2026-07-20 15:00:00", "2026-07-20", "not-a-timestamp", ""}
	for _, raw := range invalid {
		if _, err := schedule.ValidateInstant(raw); err == nil {
			t.Errorf("ValidateInstant(%q) debería rechazarse", raw)
		}
	}
}

func TestValidateBlockInterval(t *testing.T) {
	base := time.Date(2026, 7, 20, 15, 0, 0, 0, time.UTC)
	if err := schedule.ValidateBlockInterval(base, base.Add(time.Hour)); err != nil {
		t.Errorf("intervalo válido rechazado: %v", err)
	}
	// Cruce de medianoche: sigue siendo válido, sin rama especial.
	if err := schedule.ValidateBlockInterval(base, base.Add(20*time.Hour)); err != nil {
		t.Errorf("intervalo que cruza medianoche rechazado: %v", err)
	}
	if err := schedule.ValidateBlockInterval(base, base); err == nil {
		t.Error("ends == starts debería rechazarse (semiabierto)")
	}
	if err := schedule.ValidateBlockInterval(base, base.Add(-time.Hour)); err == nil {
		t.Error("ends antes de starts debería rechazarse")
	}
}

func TestValidateWeekdayShape(t *testing.T) {
	one := 1
	if err := schedule.ValidateWeekdayShape(schedule.RecurrenceKindWeekly, &one); err != nil {
		t.Errorf("weekly con isoWeekday válido rechazado: %v", err)
	}
	if err := schedule.ValidateWeekdayShape(schedule.RecurrenceKindWeekly, nil); err == nil {
		t.Error("weekly sin isoWeekday debería rechazarse")
	}
	zero := 0
	if err := schedule.ValidateWeekdayShape(schedule.RecurrenceKindWeekly, &zero); err == nil {
		t.Error("weekly con isoWeekday 0 debería rechazarse")
	}
	if err := schedule.ValidateWeekdayShape(schedule.RecurrenceKindDateList, nil); err != nil {
		t.Errorf("date_list sin isoWeekday rechazado: %v", err)
	}
	if err := schedule.ValidateWeekdayShape(schedule.RecurrenceKindDateList, &one); err == nil {
		t.Error("date_list con isoWeekday debería rechazarse")
	}
}

func TestValidateEffectiveRange(t *testing.T) {
	until := "2026-12-31"
	if err := schedule.ValidateEffectiveRange("2026-01-01", &until); err != nil {
		t.Errorf("rango válido rechazado: %v", err)
	}
	if err := schedule.ValidateEffectiveRange("2026-01-01", nil); err != nil {
		t.Errorf("sin effectiveUntil (indefinido) rechazado: %v", err)
	}
	before := "2025-12-31"
	if err := schedule.ValidateEffectiveRange("2026-01-01", &before); err == nil {
		t.Error("effectiveUntil anterior a effectiveFrom debería rechazarse")
	}
}

func TestBlockCursorRoundTrip(t *testing.T) {
	original := schedule.BlockCursor{StartsAt: "2026-07-20T15:00:00Z", ID: "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"}
	encoded := schedule.EncodeBlockCursor(original)
	if encoded == "" {
		t.Fatal("EncodeBlockCursor no debería producir cadena vacía")
	}
	decoded, err := schedule.DecodeBlockCursor(encoded)
	if err != nil {
		t.Fatalf("DecodeBlockCursor: %v", err)
	}
	if decoded != original {
		t.Errorf("round-trip = %+v, quiero %+v", decoded, original)
	}
	if _, err := schedule.DecodeBlockCursor("no-es-base64-válido!!"); err == nil {
		t.Error("cursor corrupto debería rechazarse")
	}
}

func TestSeriesCursorRoundTrip(t *testing.T) {
	original := schedule.SeriesCursor{EffectiveFrom: "2026-07-20", ID: "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"}
	encoded := schedule.EncodeSeriesCursor(original)
	decoded, err := schedule.DecodeSeriesCursor(encoded)
	if err != nil {
		t.Fatalf("DecodeSeriesCursor: %v", err)
	}
	if decoded != original {
		t.Errorf("round-trip = %+v, quiero %+v", decoded, original)
	}
	if _, err := schedule.DecodeSeriesCursor("no-es-base64-válido!!"); err == nil {
		t.Error("cursor corrupto debería rechazarse")
	}
}

func TestLooksLikeTimeBlockIDAndSeriesID(t *testing.T) {
	valid := "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"
	if !schedule.LooksLikeTimeBlockID(valid) {
		t.Error("UUID válido rechazado por LooksLikeTimeBlockID")
	}
	if !schedule.LooksLikeSeriesID(valid) {
		t.Error("UUID válido rechazado por LooksLikeSeriesID")
	}
	if schedule.LooksLikeTimeBlockID("no-es-un-uuid") {
		t.Error("cadena inválida aceptada por LooksLikeTimeBlockID")
	}
}
