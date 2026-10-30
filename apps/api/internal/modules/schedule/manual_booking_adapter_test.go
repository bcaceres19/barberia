package schedule_test

import (
	"context"
	"testing"
	"time"

	"system-barbershop/internal/modules/schedule"
)

// WithinWorkingHours (CA-061-02, CA-065-04): el intervalo debe caber completo en
// un segmento de la jornada efectiva, en la zona de la barbería.
func TestManualBookingBlocks_WithinWorkingHours(t *testing.T) {
	bogota, err := time.LoadLocation("America/Bogota")
	if err != nil {
		t.Fatal(err)
	}
	// 2026-09-15 es martes y no es festivo colombiano.
	at := func(day, hour, minute int) time.Time {
		return time.Date(2026, time.September, day, hour, minute, 0, 0, bogota)
	}
	weekly := []schedule.WorkingHour{{StartsTime: "09:00", DurationMinutes: 480}}
	split := []schedule.WorkingHour{
		{StartsTime: "09:00", DurationMinutes: 180},
		{StartsTime: "14:00", DurationMinutes: 180},
	}
	overnight := []schedule.WorkingHour{{StartsTime: "22:00", DurationMinutes: 480}}
	closed := &schedule.ScheduleException{IsClosed: true}

	cases := []struct {
		name       string
		exception  *schedule.ScheduleException
		weekly     []schedule.WorkingHour
		start, end time.Time
		want       bool
	}{
		{"inside the day", nil, weekly, at(15, 10, 0), at(15, 10, 30), true},
		{"exactly the whole segment", nil, weekly, at(15, 9, 0), at(15, 17, 0), true},
		{"ends exactly at closing time", nil, weekly, at(15, 16, 30), at(15, 17, 0), true},
		{"before opening", nil, weekly, at(15, 7, 0), at(15, 7, 30), false},
		{"starts before opening", nil, weekly, at(15, 8, 45), at(15, 9, 15), false},
		{"ends after closing", nil, weekly, at(15, 16, 45), at(15, 17, 15), false},
		{"late evening", nil, weekly, at(15, 21, 0), at(15, 21, 30), false},
		{"small hours", nil, weekly, at(15, 3, 0), at(15, 3, 30), false},
		{"day without hours", nil, nil, at(15, 10, 0), at(15, 10, 30), false},
		{"manual closed day", closed, weekly, at(15, 10, 0), at(15, 10, 30), false},
		{"straddles the break of a split day", nil, split, at(15, 11, 30), at(15, 12, 30), false},
		{"inside the second block of a split day", nil, split, at(15, 14, 30), at(15, 15, 0), true},
		{"inside the break of a split day", nil, split, at(15, 12, 30), at(15, 13, 0), false},
		{"night shift, before midnight", nil, overnight, at(15, 23, 0), at(15, 23, 30), true},
		{"night shift, across midnight", nil, overnight, at(15, 23, 45), at(16, 0, 15), true},
		{"night shift, small hours of the next day", nil, overnight, at(16, 2, 0), at(16, 2, 30), true},
		{"night shift, after it ended", nil, overnight, at(16, 7, 0), at(16, 7, 30), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newResolveRepo(t, tc.exception, false, tc.weekly)
			adapter := schedule.NewManualBookingBlocks(schedule.NewService(repo, fakeBarberPort{}))

			got, err := adapter.WithinWorkingHours(context.Background(), "shop-1", validBarberID, tc.start, tc.end, "America/Bogota")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("WithinWorkingHours = %v, want %v", got, tc.want)
			}
		})
	}
}

// El instante se interpreta en la zona de la barbería, no en la del servidor:
// 15:00 UTC es 10:00 en Bogotá (dentro de la jornada), no 15:00 de ninguna parte.
func TestManualBookingBlocks_WithinWorkingHours_UsesTheShopTimezone(t *testing.T) {
	weekly := []schedule.WorkingHour{{StartsTime: "09:00", DurationMinutes: 480}}
	repo := newResolveRepo(t, nil, false, weekly)
	adapter := schedule.NewManualBookingBlocks(schedule.NewService(repo, fakeBarberPort{}))

	start := time.Date(2026, time.September, 15, 15, 0, 0, 0, time.UTC)
	got, err := adapter.WithinWorkingHours(context.Background(), "shop-1", validBarberID, start, start.Add(30*time.Minute), "America/Bogota")
	if err != nil || !got {
		t.Fatalf("expected 10:00 Bogotá to be inside the day, got %v (err %v)", got, err)
	}
	got, err = adapter.WithinWorkingHours(context.Background(), "shop-1", validBarberID, start, start.Add(30*time.Minute), "Asia/Tokyo")
	if err != nil || got {
		t.Fatalf("expected 00:00 Tokyo next day to be outside, got %v (err %v)", got, err)
	}
}
