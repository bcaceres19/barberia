package schedule_test

import (
	"context"
	"testing"
	"time"

	"system-barbershop/internal/modules/schedule"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/idempotency"
)

// fakeRepository es un doble de schedule.Repository para probar
// schedule.Service sin PostgreSQL real: solo prueba la orquestación
// (validación antes de tocar el repositorio, traducción de resultados a
// apperr), nunca SQL, RLS ni la carrera de dos altas concurrentes (eso vive
// en postgres/repository_test.go contra PostgreSQL real).
type fakeRepository struct {
	listFn   func(ctx context.Context, barbershopID, barberID string, cursor *schedule.Cursor, limit int) (schedule.ListResult, error)
	getFn    func(ctx context.Context, barbershopID, barberID, workingHourID string) (schedule.WorkingHour, bool, error)
	createFn func(ctx context.Context, barbershopID, barberID string, input schedule.CreateInput, key idempotency.Key, fingerprint idempotency.Fingerprint) (schedule.CreateResult, error)
	updateFn func(ctx context.Context, barbershopID, barberID, workingHourID string, input schedule.UpdateInput) (schedule.UpdateResult, error)
	deleteFn func(ctx context.Context, barbershopID, barberID, workingHourID string) (bool, error)

	createCalls int
	updateCalls int
	deleteCalls int

	// HU-041: excepciones de jornada y calendario de festivos.
	getHolidayCalendarEnabledFn  func(ctx context.Context, barbershopID, barberID string) (bool, bool, error)
	setHolidayCalendarEnabledFn  func(ctx context.Context, barbershopID, barberID string, enabled bool) (schedule.HolidayCalendarResult, error)
	listExceptionsFn             func(ctx context.Context, barbershopID, barberID string, cursor *schedule.ExceptionCursor, limit int) (schedule.ExceptionListResult, error)
	getExceptionFn               func(ctx context.Context, barbershopID, barberID, exceptionID string) (schedule.ScheduleException, bool, error)
	getExceptionByDateFn         func(ctx context.Context, barbershopID, barberID, effectiveDate string) (schedule.ScheduleException, bool, error)
	createExceptionFn            func(ctx context.Context, barbershopID, barberID string, input schedule.CreateExceptionInput, key idempotency.Key, fingerprint idempotency.Fingerprint) (schedule.CreateExceptionResult, error)
	updateExceptionFn            func(ctx context.Context, barbershopID, barberID, exceptionID string, input schedule.UpdateExceptionInput) (schedule.UpdateExceptionResult, error)
	deleteExceptionFn            func(ctx context.Context, barbershopID, barberID, exceptionID string) (bool, error)
	listWorkingHoursForWeekdayFn func(ctx context.Context, barbershopID, barberID string, isoWeekday int) ([]schedule.WorkingHour, error)

	// HU-042: bloqueos de agenda. Implementadas en block_service_test.go
	// para mantener este archivo enfocado en HU-040/HU-041.
	listBlocksFn                    func(ctx context.Context, barbershopID, barberID string, cursor *schedule.BlockCursor, limit int, includeDeleted bool) (schedule.BlockListResult, error)
	getBlockFn                      func(ctx context.Context, barbershopID, barberID, blockID string) (schedule.TimeBlock, bool, error)
	createBlockFn                   func(ctx context.Context, barbershopID, barberID string, input schedule.CreateBlockInput, key idempotency.Key, fingerprint idempotency.Fingerprint) (schedule.CreateBlockResult, error)
	deleteBlockFn                   func(ctx context.Context, barbershopID, barberID, blockID, actorID string) (bool, error)
	listSeriesFn                    func(ctx context.Context, barbershopID, barberID string, cursor *schedule.SeriesCursor, limit int, includeDeleted bool) (schedule.SeriesListResult, error)
	getSeriesFn                     func(ctx context.Context, barbershopID, barberID, seriesID string) (schedule.TimeBlockSeries, bool, error)
	createSeriesFn                  func(ctx context.Context, barbershopID, barberID string, input schedule.CreateSeriesInput, key idempotency.Key, fingerprint idempotency.Fingerprint) (schedule.CreateSeriesResult, error)
	updateSeriesWholeFn             func(ctx context.Context, barbershopID, barberID, seriesID string, input schedule.UpdateSeriesInput) (schedule.UpdateSeriesResult, error)
	splitSeriesFromFn               func(ctx context.Context, barbershopID, barberID, seriesID, effectiveDate string, input schedule.UpdateSeriesInput) (schedule.UpdateSeriesResult, error)
	deleteSeriesFn                  func(ctx context.Context, barbershopID, barberID, seriesID string) (bool, error)
	addSeriesDateFn                 func(ctx context.Context, barbershopID, barberID, seriesID, blockDate string) (bool, bool, error)
	removeSeriesDateFn              func(ctx context.Context, barbershopID, barberID, seriesID, blockDate string) (bool, error)
	addSeriesExceptionFn            func(ctx context.Context, barbershopID, barberID, seriesID, excludedDate string, reason *string) (bool, bool, error)
	removeSeriesExceptionFn         func(ctx context.Context, barbershopID, barberID, seriesID, excludedDate string) (bool, error)
	listEffectiveManualBlocksFn     func(ctx context.Context, barbershopID, barberID string, from, until time.Time) ([]schedule.TimeBlock, error)
	listActiveSeriesForProjectionFn func(ctx context.Context, barbershopID, barberID, from, until string) ([]schedule.TimeBlockSeries, error)
}

func (f *fakeRepository) List(ctx context.Context, barbershopID, barberID string, cursor *schedule.Cursor, limit int) (schedule.ListResult, error) {
	return f.listFn(ctx, barbershopID, barberID, cursor, limit)
}

func (f *fakeRepository) Get(ctx context.Context, barbershopID, barberID, workingHourID string) (schedule.WorkingHour, bool, error) {
	return f.getFn(ctx, barbershopID, barberID, workingHourID)
}

func (f *fakeRepository) Create(ctx context.Context, barbershopID, barberID string, input schedule.CreateInput, key idempotency.Key, fingerprint idempotency.Fingerprint) (schedule.CreateResult, error) {
	f.createCalls++
	return f.createFn(ctx, barbershopID, barberID, input, key, fingerprint)
}

func (f *fakeRepository) Update(ctx context.Context, barbershopID, barberID, workingHourID string, input schedule.UpdateInput) (schedule.UpdateResult, error) {
	f.updateCalls++
	return f.updateFn(ctx, barbershopID, barberID, workingHourID, input)
}

func (f *fakeRepository) Delete(ctx context.Context, barbershopID, barberID, workingHourID string) (bool, error) {
	f.deleteCalls++
	return f.deleteFn(ctx, barbershopID, barberID, workingHourID)
}

func (f *fakeRepository) GetHolidayCalendarEnabled(ctx context.Context, barbershopID, barberID string) (bool, bool, error) {
	return f.getHolidayCalendarEnabledFn(ctx, barbershopID, barberID)
}

func (f *fakeRepository) SetHolidayCalendarEnabled(ctx context.Context, barbershopID, barberID string, enabled bool) (schedule.HolidayCalendarResult, error) {
	return f.setHolidayCalendarEnabledFn(ctx, barbershopID, barberID, enabled)
}

func (f *fakeRepository) ListExceptions(ctx context.Context, barbershopID, barberID string, cursor *schedule.ExceptionCursor, limit int) (schedule.ExceptionListResult, error) {
	return f.listExceptionsFn(ctx, barbershopID, barberID, cursor, limit)
}

func (f *fakeRepository) GetException(ctx context.Context, barbershopID, barberID, exceptionID string) (schedule.ScheduleException, bool, error) {
	return f.getExceptionFn(ctx, barbershopID, barberID, exceptionID)
}

func (f *fakeRepository) GetExceptionByDate(ctx context.Context, barbershopID, barberID, effectiveDate string) (schedule.ScheduleException, bool, error) {
	return f.getExceptionByDateFn(ctx, barbershopID, barberID, effectiveDate)
}

func (f *fakeRepository) CreateException(ctx context.Context, barbershopID, barberID string, input schedule.CreateExceptionInput, key idempotency.Key, fingerprint idempotency.Fingerprint) (schedule.CreateExceptionResult, error) {
	return f.createExceptionFn(ctx, barbershopID, barberID, input, key, fingerprint)
}

func (f *fakeRepository) UpdateException(ctx context.Context, barbershopID, barberID, exceptionID string, input schedule.UpdateExceptionInput) (schedule.UpdateExceptionResult, error) {
	return f.updateExceptionFn(ctx, barbershopID, barberID, exceptionID, input)
}

func (f *fakeRepository) DeleteException(ctx context.Context, barbershopID, barberID, exceptionID string) (bool, error) {
	return f.deleteExceptionFn(ctx, barbershopID, barberID, exceptionID)
}

func (f *fakeRepository) ListWorkingHoursForWeekday(ctx context.Context, barbershopID, barberID string, isoWeekday int) ([]schedule.WorkingHour, error) {
	return f.listWorkingHoursForWeekdayFn(ctx, barbershopID, barberID, isoWeekday)
}

var _ schedule.Repository = (*fakeRepository)(nil)

// fakeBarberPort es un doble de schedule.BarberPort.
type fakeBarberPort struct {
	exists bool
	err    error
}

func (f fakeBarberPort) Exists(context.Context, string, string) (bool, error) {
	return f.exists, f.err
}

var _ schedule.BarberPort = fakeBarberPort{}

const validBarberID = "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"
const validWorkingHourID = "6f1a2b3c-4d5e-4f60-8172-8394a5b6c7d8"

func mustBeValidation(t *testing.T, err error) {
	t.Helper()
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindValidation {
		t.Fatalf("expected apperr.KindValidation, got %v", err)
	}
}

func mustBeNotFound(t *testing.T, err error) {
	t.Helper()
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindNotFound {
		t.Fatalf("expected apperr.KindNotFound, got %v", err)
	}
}

func mustBeConflict(t *testing.T, err error) {
	t.Helper()
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindConflict {
		t.Fatalf("expected apperr.KindConflict, got %v", err)
	}
}

// --- Create: validación de campos (CA-040-04), antes de BarberPort/repo --

func TestCreate_InvalidWeekday_RejectedWithoutTouchingBarberPortOrRepository(t *testing.T) {
	repo := &fakeRepository{createFn: failCreateIfCalled(t)}
	svc := schedule.NewService(repo, fakeBarberPort{})

	for _, day := range []int{0, 8, -1} {
		_, err := svc.Create(context.Background(), "shop-1", validBarberID, day, "08:00", 60, "key-1", "a")
		mustBeValidation(t, err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected 0 repository calls, got %d", repo.createCalls)
	}
}

func TestCreate_InvalidStartsTime_Rejected(t *testing.T) {
	repo := &fakeRepository{createFn: failCreateIfCalled(t)}
	svc := schedule.NewService(repo, fakeBarberPort{})

	_, err := svc.Create(context.Background(), "shop-1", validBarberID, 1, "8am", 60, "key-1", "a")
	mustBeValidation(t, err)
}

func TestCreate_InvalidDuration_Rejected(t *testing.T) {
	repo := &fakeRepository{createFn: failCreateIfCalled(t)}
	svc := schedule.NewService(repo, fakeBarberPort{})

	for _, duration := range []int{0, -1, 1441} {
		_, err := svc.Create(context.Background(), "shop-1", validBarberID, 1, "08:00", duration, "key-1", "a")
		mustBeValidation(t, err)
	}
}

func TestCreate_MalformedBarberID_ReturnsNotFoundWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{createFn: failCreateIfCalled(t)}
	svc := schedule.NewService(repo, fakeBarberPort{exists: true})

	_, err := svc.Create(context.Background(), "shop-1", "not-a-uuid", 1, "08:00", 60, "key-1", "a")
	mustBeNotFound(t, err)
}

func TestCreate_BarberNotFound_ReturnsNotFoundWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{createFn: failCreateIfCalled(t)}
	svc := schedule.NewService(repo, fakeBarberPort{exists: false})

	_, err := svc.Create(context.Background(), "shop-1", validBarberID, 1, "08:00", 60, "key-1", "a")
	mustBeNotFound(t, err)
}

func TestCreate_RepositoryConflict_TranslatesToConflict(t *testing.T) {
	repo := &fakeRepository{createFn: func(context.Context, string, string, schedule.CreateInput, idempotency.Key, idempotency.Fingerprint) (schedule.CreateResult, error) {
		return schedule.CreateResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Conflict: true,
		}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{exists: true})

	_, err := svc.Create(context.Background(), "shop-1", validBarberID, 1, "08:00", 60, "key-1", "a")
	mustBeConflict(t, err)
}

func TestCreate_Success_PassesValidatedInputToRepository(t *testing.T) {
	repo := &fakeRepository{createFn: func(_ context.Context, _, barberID string, input schedule.CreateInput, _ idempotency.Key, _ idempotency.Fingerprint) (schedule.CreateResult, error) {
		if barberID != validBarberID {
			t.Fatalf("expected barberID %q, got %q", validBarberID, barberID)
		}
		if input.ISOWeekday != 1 || input.StartsTime != "08:00" || input.DurationMinutes != 60 {
			t.Fatalf("unexpected input: %+v", input)
		}
		return schedule.CreateResult{Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed}}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{exists: true})

	_, err := svc.Create(context.Background(), "shop-1", validBarberID, 1, "08:00", 60, "key-1", "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Get: not-found unifica inexistente, ajeno y de otro barbero ---------

func TestGet_MalformedIDs_ReturnsNotFoundWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{getFn: func(context.Context, string, string, string) (schedule.WorkingHour, bool, error) {
		t.Fatal("repository must not be called for a malformed id")
		return schedule.WorkingHour{}, false, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	_, err := svc.Get(context.Background(), "shop-1", "not-a-uuid", validWorkingHourID)
	mustBeNotFound(t, err)

	_, err = svc.Get(context.Background(), "shop-1", validBarberID, "not-a-uuid")
	mustBeNotFound(t, err)
}

func TestGet_NotFound_TranslatesToNotFound(t *testing.T) {
	repo := &fakeRepository{getFn: func(context.Context, string, string, string) (schedule.WorkingHour, bool, error) {
		return schedule.WorkingHour{}, false, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	_, err := svc.Get(context.Background(), "shop-1", validBarberID, validWorkingHourID)
	mustBeNotFound(t, err)
}

func TestGet_Found_ReturnsWorkingHour(t *testing.T) {
	want := schedule.WorkingHour{ID: validWorkingHourID, ISOWeekday: 2, StartsTime: "09:00", DurationMinutes: 120}
	repo := &fakeRepository{getFn: func(context.Context, string, string, string) (schedule.WorkingHour, bool, error) {
		return want, true, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	got, err := svc.Get(context.Background(), "shop-1", validBarberID, validWorkingHourID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
}

// --- Update: mismas reglas de validación, además de conflicto ------------

func TestUpdate_InvalidInterval_RejectedWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{updateFn: func(context.Context, string, string, string, schedule.UpdateInput) (schedule.UpdateResult, error) {
		t.Fatal("repository must not be called for an invalid interval")
		return schedule.UpdateResult{}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	_, err := svc.Update(context.Background(), "shop-1", validBarberID, validWorkingHourID, 8, "08:00", 60)
	mustBeValidation(t, err)
}

func TestUpdate_NotFound_TranslatesToNotFound(t *testing.T) {
	repo := &fakeRepository{updateFn: func(context.Context, string, string, string, schedule.UpdateInput) (schedule.UpdateResult, error) {
		return schedule.UpdateResult{Found: false}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	_, err := svc.Update(context.Background(), "shop-1", validBarberID, validWorkingHourID, 1, "08:00", 60)
	mustBeNotFound(t, err)
}

func TestUpdate_Conflict_TranslatesToConflict(t *testing.T) {
	repo := &fakeRepository{updateFn: func(context.Context, string, string, string, schedule.UpdateInput) (schedule.UpdateResult, error) {
		return schedule.UpdateResult{Found: true, Conflict: true}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	_, err := svc.Update(context.Background(), "shop-1", validBarberID, validWorkingHourID, 1, "08:00", 60)
	mustBeConflict(t, err)
}

func TestUpdate_Success_ReturnsUpdatedWorkingHour(t *testing.T) {
	updated := schedule.WorkingHour{ID: validWorkingHourID, ISOWeekday: 3, StartsTime: "10:00", DurationMinutes: 90}
	repo := &fakeRepository{updateFn: func(context.Context, string, string, string, schedule.UpdateInput) (schedule.UpdateResult, error) {
		return schedule.UpdateResult{Found: true, WorkingHour: updated}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	got, err := svc.Update(context.Background(), "shop-1", validBarberID, validWorkingHourID, 3, "10:00", 90)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != updated {
		t.Fatalf("expected %+v, got %+v", updated, got)
	}
}

// --- Delete: not-found uniforme, idempotente ante reintento --------------

func TestDelete_MalformedIDs_ReturnsNotFoundWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{deleteFn: func(context.Context, string, string, string) (bool, error) {
		t.Fatal("repository must not be called for a malformed id")
		return false, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	err := svc.Delete(context.Background(), "shop-1", "not-a-uuid", validWorkingHourID)
	mustBeNotFound(t, err)
}

func TestDelete_NotFound_TranslatesToNotFound(t *testing.T) {
	repo := &fakeRepository{deleteFn: func(context.Context, string, string, string) (bool, error) {
		return false, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	err := svc.Delete(context.Background(), "shop-1", validBarberID, validWorkingHourID)
	mustBeNotFound(t, err)
}

func TestDelete_Found_Succeeds(t *testing.T) {
	repo := &fakeRepository{deleteFn: func(context.Context, string, string, string) (bool, error) {
		return true, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	if err := svc.Delete(context.Background(), "shop-1", validBarberID, validWorkingHourID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- List: clamping de limit, cursor y BarberPort -------------------------

func TestList_BarberNotFound_ReturnsNotFoundWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{listFn: func(context.Context, string, string, *schedule.Cursor, int) (schedule.ListResult, error) {
		t.Fatal("repository must not be called when the barber does not exist")
		return schedule.ListResult{}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{exists: false})

	_, err := svc.List(context.Background(), "shop-1", validBarberID, "", 10)
	mustBeNotFound(t, err)
}

func TestList_ZeroLimit_UsesDefault(t *testing.T) {
	repo := &fakeRepository{listFn: func(_ context.Context, _, _ string, _ *schedule.Cursor, limit int) (schedule.ListResult, error) {
		if limit != schedule.DefaultListLimit {
			t.Fatalf("expected default limit %d, got %d", schedule.DefaultListLimit, limit)
		}
		return schedule.ListResult{}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{exists: true})

	if _, err := svc.List(context.Background(), "shop-1", validBarberID, "", 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestList_LimitAboveMax_Clamped(t *testing.T) {
	repo := &fakeRepository{listFn: func(_ context.Context, _, _ string, _ *schedule.Cursor, limit int) (schedule.ListResult, error) {
		if limit != schedule.MaxListLimit {
			t.Fatalf("expected clamped limit %d, got %d", schedule.MaxListLimit, limit)
		}
		return schedule.ListResult{}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{exists: true})

	if _, err := svc.List(context.Background(), "shop-1", validBarberID, "", 999999); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestList_ValidCursor_DecodedAndPassedThrough(t *testing.T) {
	original := schedule.Cursor{ISOWeekday: 4, StartsTime: "08:00", ID: validWorkingHourID}
	token := schedule.EncodeCursor(original)

	repo := &fakeRepository{listFn: func(_ context.Context, _, _ string, cursor *schedule.Cursor, _ int) (schedule.ListResult, error) {
		if cursor == nil || *cursor != original {
			t.Fatalf("expected decoded cursor %+v, got %+v", original, cursor)
		}
		return schedule.ListResult{}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{exists: true})

	if _, err := svc.List(context.Background(), "shop-1", validBarberID, token, 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func failCreateIfCalled(t *testing.T) func(context.Context, string, string, schedule.CreateInput, idempotency.Key, idempotency.Fingerprint) (schedule.CreateResult, error) {
	t.Helper()
	return func(context.Context, string, string, schedule.CreateInput, idempotency.Key, idempotency.Fingerprint) (schedule.CreateResult, error) {
		t.Fatal("repository must not be called")
		return schedule.CreateResult{}, nil
	}
}
