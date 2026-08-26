package schedule_test

import (
	"testing"
	"time"

	"system-barbershop/internal/modules/schedule"
)

func TestColombianHolidaysForYear_HasEighteenHolidays(t *testing.T) {
	for _, year := range []int{2024, 2025, 2026, 2027, 2028, 2100} {
		holidays := schedule.ColombianHolidaysForYear(year)
		if len(holidays) != 18 {
			t.Fatalf("year %d: expected 18 holidays, got %d: %+v", year, len(holidays), holidays)
		}
	}
}

func TestColombianHolidaysForYear_SortedAscendingByDate(t *testing.T) {
	holidays := schedule.ColombianHolidaysForYear(2026)
	for i := 1; i < len(holidays); i++ {
		if holidays[i].Date.Before(holidays[i-1].Date) {
			t.Fatalf("holidays not sorted ascending: %+v before %+v", holidays[i-1], holidays[i])
		}
	}
}

func TestColombianHolidaysForYear_FixedDatesNeverMove(t *testing.T) {
	holidays := schedule.ColombianHolidaysForYear(2026)
	want := map[string][2]int{
		"Año Nuevo":                       {1, 1},
		"Día del Trabajo":                 {5, 1},
		"Día de la Independencia":         {7, 20},
		"Batalla de Boyacá":               {8, 7},
		"Día de la Inmaculada Concepción": {12, 8},
		"Navidad":                         {12, 25},
	}
	for _, h := range holidays {
		md, ok := want[h.Name]
		if !ok {
			continue
		}
		if int(h.Date.Month()) != md[0] || h.Date.Day() != md[1] {
			t.Fatalf("%s: expected %d-%02d, got %s", h.Name, md[0], md[1], h.Date.Format("2006-01-02"))
		}
	}
}

// TestColombianHolidaysForYear_LeyEmilianiHolidaysAlwaysFallOnMonday cubre
// los diez festivos que la Ley 51 de 1983 traslada al lunes siguiente
// cuando no caen ya en lunes: San José, Reyes Magos, San Pedro y San
// Pablo, Asunción, Día de la Raza, Todos los Santos, Independencia de
// Cartagena, Ascensión, Corpus Christi y Sagrado Corazón.
func TestColombianHolidaysForYear_LeyEmilianiHolidaysAlwaysFallOnMonday(t *testing.T) {
	movable := map[string]bool{
		"Día de los Reyes Magos":     true,
		"Día de San José":            true,
		"Ascensión del Señor":        true,
		"Corpus Christi":             true,
		"Sagrado Corazón de Jesús":   true,
		"San Pedro y San Pablo":      true,
		"Asunción de la Virgen":      true,
		"Día de la Raza":             true,
		"Día de Todos los Santos":    true,
		"Independencia de Cartagena": true,
	}
	for _, year := range []int{2024, 2025, 2026, 2027, 2028} {
		for _, h := range schedule.ColombianHolidaysForYear(year) {
			if movable[h.Name] && h.Date.Weekday() != time.Monday {
				t.Fatalf("year %d: %s should fall on Monday, got %s (%s)", year, h.Name, h.Date.Weekday(), h.Date.Format("2006-01-02"))
			}
		}
	}
}

// TestColombianHolidaysForYear_HolyThursdayAndGoodFriday_NeverMove verifica
// que los dos festivos ligados a Pascua que la ley NO traslada mantienen
// su distancia fija (3 y 2 días antes de Pascua), incluso en años donde
// esa fecha ya cae en lunes/martes.
func TestColombianHolidaysForYear_HolyThursdayAndGoodFriday_NeverMove(t *testing.T) {
	for _, year := range []int{2024, 2025, 2026, 2027, 2028} {
		holidays := schedule.ColombianHolidaysForYear(year)
		var thursday, friday time.Time
		for _, h := range holidays {
			switch h.Name {
			case "Jueves Santo":
				thursday = h.Date
			case "Viernes Santo":
				friday = h.Date
			}
		}
		if thursday.IsZero() || friday.IsZero() {
			t.Fatalf("year %d: missing Jueves Santo/Viernes Santo", year)
		}
		if friday.Sub(thursday) != 24*time.Hour {
			t.Fatalf("year %d: expected exactly one day between Jueves Santo and Viernes Santo, got %s", year, friday.Sub(thursday))
		}
		if thursday.Weekday() != time.Thursday || friday.Weekday() != time.Friday {
			t.Fatalf("year %d: expected Thursday/Friday weekdays, got %s/%s", year, thursday.Weekday(), friday.Weekday())
		}
	}
}

func TestColombianHolidaysForYear_2026KnownDates(t *testing.T) {
	// Pascua 2026 es el 5 de abril (fuente pública conocida): Jueves Santo
	// y Viernes Santo caen el 2 y 3 de abril de 2026.
	holidays := schedule.ColombianHolidaysForYear(2026)
	byName := map[string]time.Time{}
	for _, h := range holidays {
		byName[h.Name] = h.Date
	}
	wantGoodFriday := time.Date(2026, time.April, 3, 0, 0, 0, 0, time.UTC)
	if got := byName["Viernes Santo"]; !got.Equal(wantGoodFriday) {
		t.Fatalf("Viernes Santo 2026: expected %s, got %s", wantGoodFriday.Format("2006-01-02"), got.Format("2006-01-02"))
	}
}

func TestIsColombianHoliday_FixedDate_Found(t *testing.T) {
	christmas := time.Date(2026, time.December, 25, 15, 4, 5, 0, time.UTC)
	holiday, found := schedule.IsColombianHoliday(christmas)
	if !found || holiday.Name != "Navidad" {
		t.Fatalf("expected Navidad on 2026-12-25, got found=%v holiday=%+v", found, holiday)
	}
}

func TestIsColombianHoliday_NonHoliday_NotFound(t *testing.T) {
	// El 2 de enero prácticamente nunca es festivo en Colombia.
	_, found := schedule.IsColombianHoliday(time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC))
	if found {
		t.Fatal("expected January 2 to not be a Colombian holiday")
	}
}
