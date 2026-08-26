package schedule_test

import (
	"context"
	"testing"
	"time"

	"system-barbershop/internal/modules/schedule"
	"system-barbershop/internal/platform/idempotency"
)

const validExceptionID = "7a1b2c3d-4e5f-4061-8273-8495a6b7c8d9"

// --- Calendario de festivos (CA-041-01/02) --------------------------------

func TestGetHolidayCalendar_MalformedID_ReturnsNotFoundWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{getHolidayCalendarEnabledFn: func(context.Context, string, string) (bool, bool, error) {
		t.Fatal("repository must not be called for a malformed id")
		return false, false, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	_, err := svc.GetHolidayCalendar(context.Background(), "shop-1", "not-a-uuid")
	mustBeNotFound(t, err)
}

func TestGetHolidayCalendar_BarberNotFound_ReturnsNotFound(t *testing.T) {
	repo := &fakeRepository{getHolidayCalendarEnabledFn: func(context.Context, string, string) (bool, bool, error) {
		return false, false, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	_, err := svc.GetHolidayCalendar(context.Background(), "shop-1", validBarberID)
	mustBeNotFound(t, err)
}

func TestGetHolidayCalendar_Found_ReturnsEnabled(t *testing.T) {
	repo := &fakeRepository{getHolidayCalendarEnabledFn: func(context.Context, string, string) (bool, bool, error) {
		return true, true, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	enabled, err := svc.GetHolidayCalendar(context.Background(), "shop-1", validBarberID)
	if err != nil || !enabled {
		t.Fatalf("expected enabled=true, got %v (err=%v)", enabled, err)
	}
}

func TestSetHolidayCalendar_BarberNotFound_ReturnsNotFound(t *testing.T) {
	repo := &fakeRepository{setHolidayCalendarEnabledFn: func(context.Context, string, string, bool) (schedule.HolidayCalendarResult, error) {
		return schedule.HolidayCalendarResult{Found: false}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	_, err := svc.SetHolidayCalendar(context.Background(), "shop-1", validBarberID, true)
	mustBeNotFound(t, err)
}

func TestSetHolidayCalendar_Success_ReturnsNewValue(t *testing.T) {
	repo := &fakeRepository{setHolidayCalendarEnabledFn: func(_ context.Context, _, _ string, enabled bool) (schedule.HolidayCalendarResult, error) {
		return schedule.HolidayCalendarResult{Enabled: enabled, Found: true}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	got, err := svc.SetHolidayCalendar(context.Background(), "shop-1", validBarberID, true)
	if err != nil || !got {
		t.Fatalf("expected true, got %v (err=%v)", got, err)
	}
}

// --- CreateException: validación de forma (CA-041-04), antes de repo -----

func TestCreateException_InvalidDate_RejectedWithoutTouchingBarberPortOrRepository(t *testing.T) {
	repo := &fakeRepository{createExceptionFn: func(context.Context, string, string, schedule.CreateExceptionInput, idempotency.Key, idempotency.Fingerprint) (schedule.CreateExceptionResult, error) {
		t.Fatal("repository must not be called")
		return schedule.CreateExceptionResult{}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	_, err := svc.CreateException(context.Background(), "shop-1", validBarberID, "not-a-date", true, nil, nil, "key-1", "a")
	mustBeValidation(t, err)
}

func TestCreateException_ClosedWithSegments_Rejected(t *testing.T) {
	repo := &fakeRepository{createExceptionFn: func(context.Context, string, string, schedule.CreateExceptionInput, idempotency.Key, idempotency.Fingerprint) (schedule.CreateExceptionResult, error) {
		t.Fatal("repository must not be called")
		return schedule.CreateExceptionResult{}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	segments := []schedule.CreateExceptionSegmentInput{{StartsTime: "08:00", DurationMinutes: 60}}
	_, err := svc.CreateException(context.Background(), "shop-1", validBarberID, "2026-12-08", true, nil, segments, "key-1", "a")
	mustBeValidation(t, err)
}

func TestCreateException_BarberNotFound_ReturnsNotFoundWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{createExceptionFn: func(context.Context, string, string, schedule.CreateExceptionInput, idempotency.Key, idempotency.Fingerprint) (schedule.CreateExceptionResult, error) {
		t.Fatal("repository must not be called")
		return schedule.CreateExceptionResult{}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{exists: false})

	_, err := svc.CreateException(context.Background(), "shop-1", validBarberID, "2026-12-08", true, nil, nil, "key-1", "a")
	mustBeNotFound(t, err)
}

func TestCreateException_RepositoryConflict_TranslatesToConflict(t *testing.T) {
	repo := &fakeRepository{createExceptionFn: func(context.Context, string, string, schedule.CreateExceptionInput, idempotency.Key, idempotency.Fingerprint) (schedule.CreateExceptionResult, error) {
		return schedule.CreateExceptionResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Conflict: true,
		}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{exists: true})

	_, err := svc.CreateException(context.Background(), "shop-1", validBarberID, "2026-12-08", true, nil, nil, "key-1", "a")
	mustBeConflict(t, err)
}

func TestCreateException_Success_PassesValidatedInputToRepository(t *testing.T) {
	reason := "Festivo trabajado"
	repo := &fakeRepository{createExceptionFn: func(_ context.Context, _, barberID string, input schedule.CreateExceptionInput, _ idempotency.Key, _ idempotency.Fingerprint) (schedule.CreateExceptionResult, error) {
		if barberID != validBarberID {
			t.Fatalf("expected barberID %q, got %q", validBarberID, barberID)
		}
		if input.EffectiveDate != "2026-07-20" || input.IsClosed || input.Reason == nil || *input.Reason != reason {
			t.Fatalf("unexpected input: %+v", input)
		}
		if len(input.Segments) != 1 {
			t.Fatalf("expected 1 segment, got %d", len(input.Segments))
		}
		return schedule.CreateExceptionResult{Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed}}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{exists: true})

	segments := []schedule.CreateExceptionSegmentInput{{StartsTime: "09:00", DurationMinutes: 180}}
	_, err := svc.CreateException(context.Background(), "shop-1", validBarberID, "2026-07-20", false, &reason, segments, "key-1", "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- UpdateException/DeleteException/GetException/ListExceptions ----------

func TestUpdateException_NotFound_TranslatesToNotFound(t *testing.T) {
	repo := &fakeRepository{updateExceptionFn: func(context.Context, string, string, string, schedule.UpdateExceptionInput) (schedule.UpdateExceptionResult, error) {
		return schedule.UpdateExceptionResult{Found: false}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	_, err := svc.UpdateException(context.Background(), "shop-1", validBarberID, validExceptionID, "2026-12-08", true, nil, nil)
	mustBeNotFound(t, err)
}

func TestUpdateException_Conflict_TranslatesToConflict(t *testing.T) {
	repo := &fakeRepository{updateExceptionFn: func(context.Context, string, string, string, schedule.UpdateExceptionInput) (schedule.UpdateExceptionResult, error) {
		return schedule.UpdateExceptionResult{Found: true, Conflict: true}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	_, err := svc.UpdateException(context.Background(), "shop-1", validBarberID, validExceptionID, "2026-12-08", true, nil, nil)
	mustBeConflict(t, err)
}

func TestDeleteException_MalformedIDs_ReturnsNotFoundWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{deleteExceptionFn: func(context.Context, string, string, string) (bool, error) {
		t.Fatal("repository must not be called for a malformed id")
		return false, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	err := svc.DeleteException(context.Background(), "shop-1", validBarberID, "not-a-uuid")
	mustBeNotFound(t, err)
}

func TestDeleteException_NotFound_TranslatesToNotFound(t *testing.T) {
	repo := &fakeRepository{deleteExceptionFn: func(context.Context, string, string, string) (bool, error) {
		return false, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	err := svc.DeleteException(context.Background(), "shop-1", validBarberID, validExceptionID)
	mustBeNotFound(t, err)
}

func TestListExceptions_BarberNotFound_ReturnsNotFoundWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{listExceptionsFn: func(context.Context, string, string, *schedule.ExceptionCursor, int) (schedule.ExceptionListResult, error) {
		t.Fatal("repository must not be called when the barber does not exist")
		return schedule.ExceptionListResult{}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{exists: false})

	_, err := svc.ListExceptions(context.Background(), "shop-1", validBarberID, "", 10)
	mustBeNotFound(t, err)
}

// --- ResolveEffectiveDay: precedencia de CA-041-07 -------------------------

func newResolveRepo(t *testing.T, exception *schedule.ScheduleException, holidayEnabled bool, weekly []schedule.WorkingHour) *fakeRepository {
	t.Helper()
	return &fakeRepository{
		getExceptionByDateFn: func(context.Context, string, string, string) (schedule.ScheduleException, bool, error) {
			if exception == nil {
				return schedule.ScheduleException{}, false, nil
			}
			return *exception, true, nil
		},
		getHolidayCalendarEnabledFn: func(context.Context, string, string) (bool, bool, error) {
			return holidayEnabled, true, nil
		},
		listWorkingHoursForWeekdayFn: func(context.Context, string, string, int) ([]schedule.WorkingHour, error) {
			return weekly, nil
		},
	}
}

func TestResolveEffectiveDay_ManualClosed_PrevailsOverEverything(t *testing.T) {
	exception := &schedule.ScheduleException{IsClosed: true}
	repo := newResolveRepo(t, exception, true, []schedule.WorkingHour{{StartsTime: "08:00", DurationMinutes: 480}})
	svc := schedule.NewService(repo, fakeBarberPort{})

	got, err := svc.ResolveEffectiveDay(context.Background(), "shop-1", validBarberID, time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.IsWorking || got.Source != schedule.EffectiveDaySourceManualClosed {
		t.Fatalf("expected manual-closed, got %+v", got)
	}
}

func TestResolveEffectiveDay_ManualOpen_PrevailsOverHolidayAndWeekly(t *testing.T) {
	exception := &schedule.ScheduleException{
		IsClosed: false,
		Segments: []schedule.ExceptionSegment{{StartsTime: "09:00", DurationMinutes: 180}},
	}
	repo := newResolveRepo(t, exception, true, nil)
	svc := schedule.NewService(repo, fakeBarberPort{})

	got, err := svc.ResolveEffectiveDay(context.Background(), "shop-1", validBarberID, time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.IsWorking || got.Source != schedule.EffectiveDaySourceManualOpen || len(got.Segments) != 1 {
		t.Fatalf("expected manual-open with 1 segment, got %+v", got)
	}
}

func TestResolveEffectiveDay_HolidayAuto_WhenEnabledAndNoException(t *testing.T) {
	repo := newResolveRepo(t, nil, true, []schedule.WorkingHour{{StartsTime: "08:00", DurationMinutes: 480}})
	svc := schedule.NewService(repo, fakeBarberPort{})

	// 2026-07-20 es festivo colombiano (Día de la Independencia).
	got, err := svc.ResolveEffectiveDay(context.Background(), "shop-1", validBarberID, time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.IsWorking || got.Source != schedule.EffectiveDaySourceHolidayAuto {
		t.Fatalf("expected holiday-auto, got %+v", got)
	}
}

func TestResolveEffectiveDay_HolidayIgnored_WhenCalendarDisabled(t *testing.T) {
	repo := newResolveRepo(t, nil, false, []schedule.WorkingHour{{StartsTime: "08:00", DurationMinutes: 480}})
	svc := schedule.NewService(repo, fakeBarberPort{})

	got, err := svc.ResolveEffectiveDay(context.Background(), "shop-1", validBarberID, time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.IsWorking || got.Source != schedule.EffectiveDaySourceWeekly {
		t.Fatalf("expected weekly (calendar disabled, a real holiday must not block), got %+v", got)
	}
}

func TestResolveEffectiveDay_Weekly_WhenNoExceptionAndNotHoliday(t *testing.T) {
	repo := newResolveRepo(t, nil, true, []schedule.WorkingHour{{StartsTime: "08:00", DurationMinutes: 480}})
	svc := schedule.NewService(repo, fakeBarberPort{})

	// 2026-01-02 no es festivo colombiano.
	got, err := svc.ResolveEffectiveDay(context.Background(), "shop-1", validBarberID, time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.IsWorking || got.Source != schedule.EffectiveDaySourceWeekly || len(got.Segments) != 1 {
		t.Fatalf("expected weekly with 1 segment, got %+v", got)
	}
}

func TestResolveEffectiveDay_BarberNotFound_ReturnsNotFound(t *testing.T) {
	repo := &fakeRepository{
		getExceptionByDateFn: func(context.Context, string, string, string) (schedule.ScheduleException, bool, error) {
			return schedule.ScheduleException{}, false, nil
		},
		getHolidayCalendarEnabledFn: func(context.Context, string, string) (bool, bool, error) {
			return false, false, nil
		},
	}
	svc := schedule.NewService(repo, fakeBarberPort{})

	_, err := svc.ResolveEffectiveDay(context.Background(), "shop-1", validBarberID, time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC))
	mustBeNotFound(t, err)
}
