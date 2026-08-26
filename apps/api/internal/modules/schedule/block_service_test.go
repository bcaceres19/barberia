package schedule_test

import (
	"context"
	"testing"
	"time"

	"system-barbershop/internal/modules/schedule"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/idempotency"
)

// --- Métodos de fakeRepository para schedule.BlockRepository (HU-042) -----

func (f *fakeRepository) ListBlocks(ctx context.Context, barbershopID, barberID string, cursor *schedule.BlockCursor, limit int, includeDeleted bool) (schedule.BlockListResult, error) {
	return f.listBlocksFn(ctx, barbershopID, barberID, cursor, limit, includeDeleted)
}

func (f *fakeRepository) GetBlock(ctx context.Context, barbershopID, barberID, blockID string) (schedule.TimeBlock, bool, error) {
	return f.getBlockFn(ctx, barbershopID, barberID, blockID)
}

func (f *fakeRepository) CreateBlock(ctx context.Context, barbershopID, barberID string, input schedule.CreateBlockInput, key idempotency.Key, fingerprint idempotency.Fingerprint) (schedule.CreateBlockResult, error) {
	return f.createBlockFn(ctx, barbershopID, barberID, input, key, fingerprint)
}

func (f *fakeRepository) DeleteBlock(ctx context.Context, barbershopID, barberID, blockID, actorID string) (bool, error) {
	return f.deleteBlockFn(ctx, barbershopID, barberID, blockID, actorID)
}

func (f *fakeRepository) ListSeries(ctx context.Context, barbershopID, barberID string, cursor *schedule.SeriesCursor, limit int, includeDeleted bool) (schedule.SeriesListResult, error) {
	return f.listSeriesFn(ctx, barbershopID, barberID, cursor, limit, includeDeleted)
}

func (f *fakeRepository) GetSeries(ctx context.Context, barbershopID, barberID, seriesID string) (schedule.TimeBlockSeries, bool, error) {
	return f.getSeriesFn(ctx, barbershopID, barberID, seriesID)
}

func (f *fakeRepository) CreateSeries(ctx context.Context, barbershopID, barberID string, input schedule.CreateSeriesInput, key idempotency.Key, fingerprint idempotency.Fingerprint) (schedule.CreateSeriesResult, error) {
	return f.createSeriesFn(ctx, barbershopID, barberID, input, key, fingerprint)
}

func (f *fakeRepository) UpdateSeriesWhole(ctx context.Context, barbershopID, barberID, seriesID string, input schedule.UpdateSeriesInput) (schedule.UpdateSeriesResult, error) {
	return f.updateSeriesWholeFn(ctx, barbershopID, barberID, seriesID, input)
}

func (f *fakeRepository) SplitSeriesFrom(ctx context.Context, barbershopID, barberID, seriesID, effectiveDate string, input schedule.UpdateSeriesInput) (schedule.UpdateSeriesResult, error) {
	return f.splitSeriesFromFn(ctx, barbershopID, barberID, seriesID, effectiveDate, input)
}

func (f *fakeRepository) DeleteSeries(ctx context.Context, barbershopID, barberID, seriesID string) (bool, error) {
	return f.deleteSeriesFn(ctx, barbershopID, barberID, seriesID)
}

func (f *fakeRepository) AddSeriesDate(ctx context.Context, barbershopID, barberID, seriesID, blockDate string) (bool, bool, error) {
	return f.addSeriesDateFn(ctx, barbershopID, barberID, seriesID, blockDate)
}

func (f *fakeRepository) RemoveSeriesDate(ctx context.Context, barbershopID, barberID, seriesID, blockDate string) (bool, error) {
	return f.removeSeriesDateFn(ctx, barbershopID, barberID, seriesID, blockDate)
}

func (f *fakeRepository) AddSeriesException(ctx context.Context, barbershopID, barberID, seriesID, excludedDate string, reason *string) (bool, bool, error) {
	return f.addSeriesExceptionFn(ctx, barbershopID, barberID, seriesID, excludedDate, reason)
}

func (f *fakeRepository) RemoveSeriesException(ctx context.Context, barbershopID, barberID, seriesID, excludedDate string) (bool, error) {
	return f.removeSeriesExceptionFn(ctx, barbershopID, barberID, seriesID, excludedDate)
}

func (f *fakeRepository) ListEffectiveManualBlocks(ctx context.Context, barbershopID, barberID string, from, until time.Time) ([]schedule.TimeBlock, error) {
	return f.listEffectiveManualBlocksFn(ctx, barbershopID, barberID, from, until)
}

func (f *fakeRepository) ListActiveSeriesForProjection(ctx context.Context, barbershopID, barberID, from, until string) ([]schedule.TimeBlockSeries, error) {
	return f.listActiveSeriesForProjectionFn(ctx, barbershopID, barberID, from, until)
}

var _ schedule.BlockRepository = (*fakeRepository)(nil)

const validSeriesID = "7a1b2c3d-4e5f-4061-8273-8495a6b7c8d9"

// --- CreateBlock: validación de campos (CA-042), antes de BarberPort/repo -

func TestCreateBlock_InvalidBlockType_RejectedWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{createBlockFn: func(context.Context, string, string, schedule.CreateBlockInput, idempotency.Key, idempotency.Fingerprint) (schedule.CreateBlockResult, error) {
		t.Fatal("repository no debe llamarse con un tipo inválido")
		return schedule.CreateBlockResult{}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	_, err := svc.CreateBlock(context.Background(), "shop-1", validBarberID, "not-a-type", "2026-07-20T15:00:00-05:00", "2026-07-20T16:00:00-05:00", nil, "key-1", "a")
	mustBeValidation(t, err)
}

func TestCreateBlock_EndsBeforeStarts_RejectedWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{createBlockFn: func(context.Context, string, string, schedule.CreateBlockInput, idempotency.Key, idempotency.Fingerprint) (schedule.CreateBlockResult, error) {
		t.Fatal("repository no debe llamarse con un intervalo inválido")
		return schedule.CreateBlockResult{}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	_, err := svc.CreateBlock(context.Background(), "shop-1", validBarberID, schedule.BlockTypeEmergency, "2026-07-20T16:00:00-05:00", "2026-07-20T15:00:00-05:00", nil, "key-1", "a")
	mustBeValidation(t, err)
}

func TestCreateBlock_BarberNotFound_ReturnsNotFound(t *testing.T) {
	repo := &fakeRepository{createBlockFn: func(context.Context, string, string, schedule.CreateBlockInput, idempotency.Key, idempotency.Fingerprint) (schedule.CreateBlockResult, error) {
		t.Fatal("repository no debe llamarse si el barbero no existe")
		return schedule.CreateBlockResult{}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{exists: false})

	_, err := svc.CreateBlock(context.Background(), "shop-1", validBarberID, schedule.BlockTypeEmergency, "2026-07-20T15:00:00-05:00", "2026-07-20T16:00:00-05:00", nil, "key-1", "a")
	mustBeNotFound(t, err)
}

func TestCreateBlock_NeverConflicts_EvenOverlappingIntervalsSucceed(t *testing.T) {
	// RN-BLQ-03/DEC-008: crear un bloqueo NUNCA falla por chocar con otro
	// estado existente. El repositorio nunca reporta conflicto (no existe
	// un campo Conflict en CreateBlockResult): esta prueba documenta esa
	// garantía a nivel de Service, no solo de tipos.
	repo := &fakeRepository{createBlockFn: func(_ context.Context, _, _ string, input schedule.CreateBlockInput, _ idempotency.Key, _ idempotency.Fingerprint) (schedule.CreateBlockResult, error) {
		return schedule.CreateBlockResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Block:    schedule.TimeBlock{ID: "block-1", BlockType: input.BlockType, StartsAt: input.StartsAt, EndsAt: input.EndsAt},
		}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{exists: true})

	result, err := svc.CreateBlock(context.Background(), "shop-1", validBarberID, schedule.BlockTypeEmergency, "2026-07-20T15:00:00-05:00", "2026-07-20T19:00:00-05:00", nil, "key-1", "a")
	if err != nil {
		t.Fatalf("esperaba éxito, obtuve %v", err)
	}
	if result.Block.ID != "block-1" {
		t.Errorf("Block.ID = %q, quiero block-1", result.Block.ID)
	}
}

// --- DeleteBlock: retiro lógico, reintento seguro ---------------------

func TestDeleteBlock_NotFound_ReturnsNotFound(t *testing.T) {
	repo := &fakeRepository{deleteBlockFn: func(context.Context, string, string, string, string) (bool, error) {
		return false, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	err := svc.DeleteBlock(context.Background(), "shop-1", validBarberID, validWorkingHourID, "actor-1")
	mustBeNotFound(t, err)
}

func TestDeleteBlock_MalformedID_RejectedWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{deleteBlockFn: func(context.Context, string, string, string, string) (bool, error) {
		t.Fatal("repository no debe llamarse con un id malformado")
		return false, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	err := svc.DeleteBlock(context.Background(), "shop-1", validBarberID, "not-a-uuid", "actor-1")
	mustBeNotFound(t, err)
}

// --- CreateSeries: forma weekly/date_list ------------------------------

func TestCreateSeries_WeeklyMissingWeekday_Rejected(t *testing.T) {
	repo := &fakeRepository{createSeriesFn: func(context.Context, string, string, schedule.CreateSeriesInput, idempotency.Key, idempotency.Fingerprint) (schedule.CreateSeriesResult, error) {
		t.Fatal("repository no debe llamarse con una forma inválida")
		return schedule.CreateSeriesResult{}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	_, err := svc.CreateSeries(context.Background(), "shop-1", validBarberID, schedule.BlockTypeLunch, schedule.RecurrenceKindWeekly, nil, "13:00", 60, "2026-01-01", nil, nil, nil, "key-1", "a")
	mustBeValidation(t, err)
}

func TestCreateSeries_DateListWithWeekday_Rejected(t *testing.T) {
	repo := &fakeRepository{createSeriesFn: func(context.Context, string, string, schedule.CreateSeriesInput, idempotency.Key, idempotency.Fingerprint) (schedule.CreateSeriesResult, error) {
		t.Fatal("repository no debe llamarse con una forma inválida")
		return schedule.CreateSeriesResult{}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	monday := 1
	_, err := svc.CreateSeries(context.Background(), "shop-1", validBarberID, schedule.BlockTypeVacation, schedule.RecurrenceKindDateList, &monday, "00:00", 1440, "2026-01-01", nil, nil, nil, "key-1", "a")
	mustBeValidation(t, err)
}

func TestCreateSeries_ExplicitDateOutOfRange_Rejected(t *testing.T) {
	repo := &fakeRepository{createSeriesFn: func(context.Context, string, string, schedule.CreateSeriesInput, idempotency.Key, idempotency.Fingerprint) (schedule.CreateSeriesResult, error) {
		t.Fatal("repository no debe llamarse con una fecha fuera de rango")
		return schedule.CreateSeriesResult{}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{exists: true})

	until := "2026-12-31"
	_, err := svc.CreateSeries(context.Background(), "shop-1", validBarberID, schedule.BlockTypeVacation, schedule.RecurrenceKindDateList, nil, "00:00", 1440, "2026-12-15", &until, nil, []string{"2027-01-05"}, "key-1", "a")
	mustBeValidation(t, err)
}

func TestCreateSeries_ExplicitDatesOnWeekly_Rejected(t *testing.T) {
	repo := &fakeRepository{createSeriesFn: func(context.Context, string, string, schedule.CreateSeriesInput, idempotency.Key, idempotency.Fingerprint) (schedule.CreateSeriesResult, error) {
		t.Fatal("repository no debe llamarse: explicitDates no aplica a weekly")
		return schedule.CreateSeriesResult{}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{exists: true})

	monday := 1
	_, err := svc.CreateSeries(context.Background(), "shop-1", validBarberID, schedule.BlockTypeLunch, schedule.RecurrenceKindWeekly, &monday, "13:00", 60, "2026-01-01", nil, nil, []string{"2026-01-05"}, "key-1", "a")
	mustBeValidation(t, err)
}

func TestCreateSeries_Valid_PassesThroughToRepository(t *testing.T) {
	repo := &fakeRepository{createSeriesFn: func(_ context.Context, _, _ string, input schedule.CreateSeriesInput, _ idempotency.Key, _ idempotency.Fingerprint) (schedule.CreateSeriesResult, error) {
		return schedule.CreateSeriesResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Series:   schedule.TimeBlockSeries{ID: "series-1", RecurrenceKind: input.RecurrenceKind},
		}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{exists: true})

	monday := 1
	result, err := svc.CreateSeries(context.Background(), "shop-1", validBarberID, schedule.BlockTypeLunch, schedule.RecurrenceKindWeekly, &monday, "13:00", 60, "2026-01-01", nil, nil, nil, "key-1", "a")
	if err != nil {
		t.Fatalf("esperaba éxito, obtuve %v", err)
	}
	if result.Series.ID != "series-1" {
		t.Errorf("Series.ID = %q, quiero series-1", result.Series.ID)
	}
}

// --- UpdateSeries: scope whole / this_and_following --------------------

func TestUpdateSeries_InvalidScope_Rejected(t *testing.T) {
	repo := &fakeRepository{}
	svc := schedule.NewService(repo, fakeBarberPort{})

	_, err := svc.UpdateSeries(context.Background(), "shop-1", validBarberID, validSeriesID, "not-a-scope", schedule.BlockTypeLunch, "13:00", 60, "2026-01-01", nil, nil, "")
	mustBeValidation(t, err)
}

func TestUpdateSeries_Whole_NotFound_ReturnsNotFound(t *testing.T) {
	repo := &fakeRepository{updateSeriesWholeFn: func(context.Context, string, string, string, schedule.UpdateSeriesInput) (schedule.UpdateSeriesResult, error) {
		return schedule.UpdateSeriesResult{Found: false}, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	_, err := svc.UpdateSeries(context.Background(), "shop-1", validBarberID, validSeriesID, schedule.UpdateScopeWhole, schedule.BlockTypeLunch, "13:00", 60, "2026-01-01", nil, nil, "")
	mustBeNotFound(t, err)
}

func TestUpdateSeries_ThisAndFollowing_OnDateListSeries_Rejected(t *testing.T) {
	repo := &fakeRepository{getSeriesFn: func(context.Context, string, string, string) (schedule.TimeBlockSeries, bool, error) {
		return schedule.TimeBlockSeries{ID: validSeriesID, RecurrenceKind: schedule.RecurrenceKindDateList, EffectiveFrom: "2026-01-01"}, true, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	_, err := svc.UpdateSeries(context.Background(), "shop-1", validBarberID, validSeriesID, schedule.UpdateScopeThisAndFollowing, schedule.BlockTypeVacation, "00:00", 1440, "2026-01-01", nil, nil, "2026-06-01")
	mustBeValidation(t, err)
}

func TestUpdateSeries_ThisAndFollowing_SplitDateNotAfterEffectiveFrom_Rejected(t *testing.T) {
	repo := &fakeRepository{getSeriesFn: func(context.Context, string, string, string) (schedule.TimeBlockSeries, bool, error) {
		return schedule.TimeBlockSeries{ID: validSeriesID, RecurrenceKind: schedule.RecurrenceKindWeekly, EffectiveFrom: "2026-01-01"}, true, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	_, err := svc.UpdateSeries(context.Background(), "shop-1", validBarberID, validSeriesID, schedule.UpdateScopeThisAndFollowing, schedule.BlockTypeLunch, "13:00", 60, "2026-01-01", nil, nil, "2026-01-01")
	mustBeValidation(t, err)
}

func TestUpdateSeries_ThisAndFollowing_Valid_DelegatesToSplitSeriesFrom(t *testing.T) {
	var capturedSplitDate string
	repo := &fakeRepository{
		getSeriesFn: func(context.Context, string, string, string) (schedule.TimeBlockSeries, bool, error) {
			until := "2026-12-31"
			return schedule.TimeBlockSeries{ID: validSeriesID, RecurrenceKind: schedule.RecurrenceKindWeekly, EffectiveFrom: "2026-01-01", EffectiveUntil: &until}, true, nil
		},
		splitSeriesFromFn: func(_ context.Context, _, _, _, effectiveDate string, _ schedule.UpdateSeriesInput) (schedule.UpdateSeriesResult, error) {
			capturedSplitDate = effectiveDate
			return schedule.UpdateSeriesResult{Found: true, Series: schedule.TimeBlockSeries{ID: "series-2"}}, nil
		},
	}
	svc := schedule.NewService(repo, fakeBarberPort{})

	result, err := svc.UpdateSeries(context.Background(), "shop-1", validBarberID, validSeriesID, schedule.UpdateScopeThisAndFollowing, schedule.BlockTypeLunch, "13:00", 60, "2026-01-01", nil, nil, "2026-06-01")
	if err != nil {
		t.Fatalf("esperaba éxito, obtuve %v", err)
	}
	if result.ID != "series-2" {
		t.Errorf("Series.ID = %q, quiero series-2", result.ID)
	}
	if capturedSplitDate != "2026-06-01" {
		t.Errorf("SplitSeriesFrom recibió effectiveDate=%q, quiero 2026-06-01", capturedSplitDate)
	}
}

// --- AddSeriesDate: solo date_list, dentro del rango --------------------

func TestAddSeriesDate_OnWeeklySeries_Rejected(t *testing.T) {
	repo := &fakeRepository{getSeriesFn: func(context.Context, string, string, string) (schedule.TimeBlockSeries, bool, error) {
		return schedule.TimeBlockSeries{ID: validSeriesID, RecurrenceKind: schedule.RecurrenceKindWeekly, EffectiveFrom: "2026-01-01"}, true, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	err := svc.AddSeriesDate(context.Background(), "shop-1", validBarberID, validSeriesID, "2026-01-05")
	mustBeValidation(t, err)
}

func TestAddSeriesDate_OutOfEffectiveRange_Rejected(t *testing.T) {
	repo := &fakeRepository{getSeriesFn: func(context.Context, string, string, string) (schedule.TimeBlockSeries, bool, error) {
		until := "2026-01-31"
		return schedule.TimeBlockSeries{ID: validSeriesID, RecurrenceKind: schedule.RecurrenceKindDateList, EffectiveFrom: "2026-01-01", EffectiveUntil: &until}, true, nil
	}}
	svc := schedule.NewService(repo, fakeBarberPort{})

	err := svc.AddSeriesDate(context.Background(), "shop-1", validBarberID, validSeriesID, "2026-02-15")
	mustBeValidation(t, err)
}

func TestAddSeriesDate_Duplicate_ReturnsConflict(t *testing.T) {
	repo := &fakeRepository{
		getSeriesFn: func(context.Context, string, string, string) (schedule.TimeBlockSeries, bool, error) {
			return schedule.TimeBlockSeries{ID: validSeriesID, RecurrenceKind: schedule.RecurrenceKindDateList, EffectiveFrom: "2026-01-01"}, true, nil
		},
		addSeriesDateFn: func(context.Context, string, string, string, string) (bool, bool, error) {
			return true, true, nil
		},
	}
	svc := schedule.NewService(repo, fakeBarberPort{})

	err := svc.AddSeriesDate(context.Background(), "shop-1", validBarberID, validSeriesID, "2026-01-05")
	mustBeConflict(t, err)
}

// --- EffectiveBlocks: proyección --------------------------------------

func TestEffectiveBlocks_ToBeforeFrom_Rejected(t *testing.T) {
	svc := schedule.NewService(&fakeRepository{}, fakeBarberPort{})

	// to anterior a from es un parámetro de consulta malformado
	// (apperr.KindInvalid), mismo Kind que un cursor corrupto o un limit
	// no numérico: no depende del estado persistido, así que no es
	// apperr.KindValidation.
	_, err := svc.EffectiveBlocks(context.Background(), "shop-1", validBarberID, "2026-07-20", "2026-07-01")
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInvalid {
		t.Fatalf("expected apperr.KindInvalid, got %v", err)
	}
}

func TestEffectiveBlocks_ExpandsWeeklySeriesAndExcludesExceptions(t *testing.T) {
	monday := 1
	repo := &fakeRepository{
		listEffectiveManualBlocksFn: func(context.Context, string, string, time.Time, time.Time) ([]schedule.TimeBlock, error) {
			return []schedule.TimeBlock{{ID: "manual-1"}}, nil
		},
		listActiveSeriesForProjectionFn: func(context.Context, string, string, string, string) ([]schedule.TimeBlockSeries, error) {
			return []schedule.TimeBlockSeries{
				{
					ID: "series-1", RecurrenceKind: schedule.RecurrenceKindWeekly, ISOWeekday: &monday,
					StartsTime: "09:00", DurationMinutes: 60, EffectiveFrom: "2026-01-01",
					Exceptions: []schedule.SeriesException{{ExcludedDate: "2026-01-12"}},
				},
			}, nil
		},
	}
	svc := schedule.NewService(repo, fakeBarberPort{exists: true})

	// 2026-01-05 y 2026-01-12 y 2026-01-19 son lunes; 2026-01-12 está excluido.
	result, err := svc.EffectiveBlocks(context.Background(), "shop-1", validBarberID, "2026-01-01", "2026-01-19")
	if err != nil {
		t.Fatalf("esperaba éxito, obtuve %v", err)
	}
	if len(result.ManualBlocks) != 1 || result.ManualBlocks[0].ID != "manual-1" {
		t.Errorf("ManualBlocks = %+v, quiero un elemento manual-1", result.ManualBlocks)
	}
	wantDates := map[string]bool{"2026-01-05": true, "2026-01-19": true}
	if len(result.SeriesOccurrences) != len(wantDates) {
		t.Fatalf("SeriesOccurrences = %+v, quiero %d ocurrencias", result.SeriesOccurrences, len(wantDates))
	}
	for _, occ := range result.SeriesOccurrences {
		if !wantDates[occ.Date] {
			t.Errorf("fecha inesperada en la proyección: %s", occ.Date)
		}
		if occ.Date == "2026-01-12" {
			t.Error("2026-01-12 está excluido por Exceptions y no debería aparecer")
		}
	}
}

func TestEffectiveBlocks_DateListSeries_OnlyDatesWithinQueryRange(t *testing.T) {
	repo := &fakeRepository{
		listEffectiveManualBlocksFn: func(context.Context, string, string, time.Time, time.Time) ([]schedule.TimeBlock, error) {
			return nil, nil
		},
		listActiveSeriesForProjectionFn: func(context.Context, string, string, string, string) ([]schedule.TimeBlockSeries, error) {
			return []schedule.TimeBlockSeries{
				{
					ID: "series-2", RecurrenceKind: schedule.RecurrenceKindDateList, EffectiveFrom: "2026-01-01",
					StartsTime: "00:00", DurationMinutes: 1440,
					Dates: []schedule.SeriesDate{{BlockDate: "2026-01-10"}, {BlockDate: "2026-02-10"}},
				},
			}, nil
		},
	}
	svc := schedule.NewService(repo, fakeBarberPort{exists: true})

	result, err := svc.EffectiveBlocks(context.Background(), "shop-1", validBarberID, "2026-01-01", "2026-01-31")
	if err != nil {
		t.Fatalf("esperaba éxito, obtuve %v", err)
	}
	if len(result.SeriesOccurrences) != 1 || result.SeriesOccurrences[0].Date != "2026-01-10" {
		t.Errorf("SeriesOccurrences = %+v, quiero solo 2026-01-10", result.SeriesOccurrences)
	}
}
