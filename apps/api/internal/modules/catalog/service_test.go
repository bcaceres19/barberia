package catalog_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"system-barbershop/internal/modules/catalog"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/idempotency"
)

// fakeRepository es un doble de catalog.Repository para probar
// catalog.CatalogService sin PostgreSQL real: solo prueba la orquestación
// (validación antes de tocar el repositorio, traducción de resultados a
// apperr), nunca SQL, RLS ni el unique_violation real (eso vive en
// postgres/repository_test.go contra PostgreSQL real).
type fakeRepository struct {
	listFn       func(ctx context.Context, barbershopID string, cursor *catalog.Cursor, limit int) (catalog.ListResult, error)
	getFn        func(ctx context.Context, barbershopID, serviceID string) (catalog.Service, bool, error)
	createFn     func(ctx context.Context, barbershopID string, input catalog.CreateInput, key idempotency.Key, fingerprint idempotency.Fingerprint) (catalog.CreateResult, error)
	updateFn     func(ctx context.Context, barbershopID, serviceID string, fields catalog.UpdateFields) (catalog.UpdateResult, error)
	deactivateFn func(ctx context.Context, barbershopID, serviceID string, key idempotency.Key, fingerprint idempotency.Fingerprint) (catalog.LifecycleResult, error)
	reactivateFn func(ctx context.Context, barbershopID, serviceID string, key idempotency.Key, fingerprint idempotency.Fingerprint) (catalog.LifecycleResult, error)

	createCalls     int
	updateCalls     int
	deactivateCalls int
	reactivateCalls int
}

func (f *fakeRepository) List(ctx context.Context, barbershopID string, cursor *catalog.Cursor, limit int) (catalog.ListResult, error) {
	return f.listFn(ctx, barbershopID, cursor, limit)
}

func (f *fakeRepository) Get(ctx context.Context, barbershopID, serviceID string) (catalog.Service, bool, error) {
	return f.getFn(ctx, barbershopID, serviceID)
}

func (f *fakeRepository) Create(ctx context.Context, barbershopID string, input catalog.CreateInput, key idempotency.Key, fingerprint idempotency.Fingerprint) (catalog.CreateResult, error) {
	f.createCalls++
	return f.createFn(ctx, barbershopID, input, key, fingerprint)
}

func (f *fakeRepository) Update(ctx context.Context, barbershopID, serviceID string, fields catalog.UpdateFields) (catalog.UpdateResult, error) {
	f.updateCalls++
	return f.updateFn(ctx, barbershopID, serviceID, fields)
}

func (f *fakeRepository) Deactivate(ctx context.Context, barbershopID, serviceID string, key idempotency.Key, fingerprint idempotency.Fingerprint) (catalog.LifecycleResult, error) {
	f.deactivateCalls++
	return f.deactivateFn(ctx, barbershopID, serviceID, key, fingerprint)
}

func (f *fakeRepository) Reactivate(ctx context.Context, barbershopID, serviceID string, key idempotency.Key, fingerprint idempotency.Fingerprint) (catalog.LifecycleResult, error) {
	f.reactivateCalls++
	return f.reactivateFn(ctx, barbershopID, serviceID, key, fingerprint)
}

var _ catalog.Repository = (*fakeRepository)(nil)

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

func failingCreateFn(t *testing.T) func(context.Context, string, catalog.CreateInput, idempotency.Key, idempotency.Fingerprint) (catalog.CreateResult, error) {
	return func(context.Context, string, catalog.CreateInput, idempotency.Key, idempotency.Fingerprint) (catalog.CreateResult, error) {
		t.Fatal("repository must not be called for an invalid input")
		return catalog.CreateResult{}, nil
	}
}

func validCreateRaw() catalog.CreateInputRaw {
	return catalog.CreateInputRaw{Name: "Corte clásico", DurationMinutes: 30, Price: "45000.00"}
}

// --- Create: validación de campos, en el orden fijo del contrato ---------

func TestCreate_EmptyName_RejectedWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{createFn: failingCreateFn(t)}
	svc := catalog.NewService(repo)

	raw := validCreateRaw()
	raw.Name = ""
	_, err := svc.Create(context.Background(), "shop-1", raw, "key-1", "a")
	mustBeValidation(t, err)
	if repo.createCalls != 0 {
		t.Fatalf("expected 0 repository calls, got %d", repo.createCalls)
	}
}

func TestCreate_WhitespaceOnlyName_Rejected(t *testing.T) {
	repo := &fakeRepository{createFn: failingCreateFn(t)}
	svc := catalog.NewService(repo)

	raw := validCreateRaw()
	raw.Name = "   \t\n  "
	_, err := svc.Create(context.Background(), "shop-1", raw, "key-1", "a")
	mustBeValidation(t, err)
}

func TestCreate_NameOver120Characters_Rejected(t *testing.T) {
	repo := &fakeRepository{createFn: failingCreateFn(t)}
	svc := catalog.NewService(repo)

	raw := validCreateRaw()
	raw.Name = strings.Repeat("a", 121)
	_, err := svc.Create(context.Background(), "shop-1", raw, "key-1", "a")
	mustBeValidation(t, err)
}

func TestCreate_NameExactly120Characters_Accepted(t *testing.T) {
	name := strings.Repeat("a", 120)
	repo := &fakeRepository{createFn: func(_ context.Context, _ string, input catalog.CreateInput, _ idempotency.Key, _ idempotency.Fingerprint) (catalog.CreateResult, error) {
		if input.Name != name {
			t.Fatalf("expected repository to receive the trimmed name %q, got %q", name, input.Name)
		}
		return catalog.CreateResult{Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed}}, nil
	}}
	svc := catalog.NewService(repo)

	raw := validCreateRaw()
	raw.Name = name
	_, err := svc.Create(context.Background(), "shop-1", raw, "key-1", "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreate_UnicodeName_CountsRunesNotBytes(t *testing.T) {
	// 120 runas de "ñ" (2 bytes cada una en UTF-8, 240 bytes en total): debe
	// aceptarse porque la regla cuenta caracteres, no bytes (CA-022-04).
	name := strings.Repeat("ñ", 120)
	repo := &fakeRepository{createFn: func(context.Context, string, catalog.CreateInput, idempotency.Key, idempotency.Fingerprint) (catalog.CreateResult, error) {
		return catalog.CreateResult{Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed}}, nil
	}}
	svc := catalog.NewService(repo)

	raw := validCreateRaw()
	raw.Name = name
	_, err := svc.Create(context.Background(), "shop-1", raw, "key-1", "a")
	if err != nil {
		t.Fatalf("unexpected error for a 120-rune unicode name: %v", err)
	}
}

func TestCreate_DescriptionOver500Characters_Rejected(t *testing.T) {
	repo := &fakeRepository{createFn: failingCreateFn(t)}
	svc := catalog.NewService(repo)

	raw := validCreateRaw()
	raw.Description = strings.Repeat("a", 501)
	_, err := svc.Create(context.Background(), "shop-1", raw, "key-1", "a")
	mustBeValidation(t, err)
}

func TestCreate_DescriptionExactly500Characters_Accepted(t *testing.T) {
	description := strings.Repeat("a", 500)
	repo := &fakeRepository{createFn: func(_ context.Context, _ string, input catalog.CreateInput, _ idempotency.Key, _ idempotency.Fingerprint) (catalog.CreateResult, error) {
		if input.Description == nil || *input.Description != description {
			t.Fatalf("expected description of 500 chars, got %v", input.Description)
		}
		return catalog.CreateResult{Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed}}, nil
	}}
	svc := catalog.NewService(repo)

	raw := validCreateRaw()
	raw.Description = description
	_, err := svc.Create(context.Background(), "shop-1", raw, "key-1", "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreate_EmptyDescription_NormalizedToNil(t *testing.T) {
	repo := &fakeRepository{createFn: func(_ context.Context, _ string, input catalog.CreateInput, _ idempotency.Key, _ idempotency.Fingerprint) (catalog.CreateResult, error) {
		if input.Description != nil {
			t.Fatalf("expected nil description, got %q", *input.Description)
		}
		return catalog.CreateResult{Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed}}, nil
	}}
	svc := catalog.NewService(repo)

	raw := validCreateRaw()
	raw.Description = "   "
	_, err := svc.Create(context.Background(), "shop-1", raw, "key-1", "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreate_Durations25_30_45_90_AllAccepted(t *testing.T) {
	for _, minutes := range []int{25, 30, 45, 90} {
		repo := &fakeRepository{createFn: func(_ context.Context, _ string, input catalog.CreateInput, _ idempotency.Key, _ idempotency.Fingerprint) (catalog.CreateResult, error) {
			if input.DurationMinutes != minutes {
				t.Fatalf("expected duration %d, got %d", minutes, input.DurationMinutes)
			}
			return catalog.CreateResult{Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed}}, nil
		}}
		svc := catalog.NewService(repo)

		raw := validCreateRaw()
		raw.DurationMinutes = minutes
		if _, err := svc.Create(context.Background(), "shop-1", raw, "key-1", "a"); err != nil {
			t.Fatalf("unexpected error for %d minutes: %v", minutes, err)
		}
	}
}

func TestCreate_DurationZeroNegativeOrAbove1440_Rejected(t *testing.T) {
	for _, minutes := range []int{0, -1, 1441, 10000} {
		repo := &fakeRepository{createFn: failingCreateFn(t)}
		svc := catalog.NewService(repo)

		raw := validCreateRaw()
		raw.DurationMinutes = minutes
		_, err := svc.Create(context.Background(), "shop-1", raw, "key-1", "a")
		mustBeValidation(t, err)
	}
}

func TestCreate_PriceZeroOrNegative_Rejected(t *testing.T) {
	for _, price := range []string{"0", "0.00", "-1"} {
		repo := &fakeRepository{createFn: failingCreateFn(t)}
		svc := catalog.NewService(repo)

		raw := validCreateRaw()
		raw.Price = price
		_, err := svc.Create(context.Background(), "shop-1", raw, "key-1", "a")
		mustBeValidation(t, err)
	}
}

func TestCreate_PriceInvalidFormat_Rejected(t *testing.T) {
	repo := &fakeRepository{createFn: failingCreateFn(t)}
	svc := catalog.NewService(repo)

	raw := validCreateRaw()
	raw.Price = "not-a-number"
	_, err := svc.Create(context.Background(), "shop-1", raw, "key-1", "a")
	mustBeValidation(t, err)
}

func TestCreate_NameTaken_TranslatesToConflictWithoutWrappingAsInternal(t *testing.T) {
	repo := &fakeRepository{createFn: func(context.Context, string, catalog.CreateInput, idempotency.Key, idempotency.Fingerprint) (catalog.CreateResult, error) {
		return catalog.CreateResult{NameTaken: true}, nil
	}}
	svc := catalog.NewService(repo)

	_, err := svc.Create(context.Background(), "shop-1", validCreateRaw(), "key-1", "a")
	mustBeConflict(t, err)
}

func TestCreate_PropagatesRepositoryResultUntouched(t *testing.T) {
	want := catalog.CreateResult{
		Decision: idempotency.Decision{Outcome: idempotency.OutcomeReplay, Response: idempotency.StoredResponse{Status: 201, Body: `{"id":"x"}`}},
		Response: idempotency.StoredResponse{Status: 201, Body: `{"id":"x"}`},
	}
	repo := &fakeRepository{createFn: func(context.Context, string, catalog.CreateInput, idempotency.Key, idempotency.Fingerprint) (catalog.CreateResult, error) {
		return want, nil
	}}
	svc := catalog.NewService(repo)

	got, err := svc.Create(context.Background(), "shop-1", validCreateRaw(), "key-1", "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Decision.Outcome != idempotency.OutcomeReplay || got.Response.Body != `{"id":"x"}` {
		t.Fatalf("expected the repository result to pass through untouched, got %+v", got)
	}
}

// --- Update: edición parcial, cuerpo vacío, not-found, conflicto ---------

func TestUpdate_EmptyBody_RejectedWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{updateFn: func(context.Context, string, string, catalog.UpdateFields) (catalog.UpdateResult, error) {
		t.Fatal("repository must not be called for an empty update body")
		return catalog.UpdateResult{}, nil
	}}
	svc := catalog.NewService(repo)

	_, err := svc.Update(context.Background(), "shop-1", "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4", catalog.UpdateInputRaw{})
	mustBeValidation(t, err)
	if repo.updateCalls != 0 {
		t.Fatalf("expected 0 repository calls, got %d", repo.updateCalls)
	}
}

func TestUpdate_MalformedID_ReturnsNotFoundWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{updateFn: func(context.Context, string, string, catalog.UpdateFields) (catalog.UpdateResult, error) {
		t.Fatal("repository must not be called for a malformed id")
		return catalog.UpdateResult{}, nil
	}}
	svc := catalog.NewService(repo)

	name := "Nuevo nombre"
	_, err := svc.Update(context.Background(), "shop-1", "not-a-uuid", catalog.UpdateInputRaw{Name: &name})
	mustBeNotFound(t, err)
}

func TestUpdate_OnlyOneFieldPresent_ValidatesOnlyThatField(t *testing.T) {
	minutes := 45
	repo := &fakeRepository{updateFn: func(_ context.Context, _, _ string, fields catalog.UpdateFields) (catalog.UpdateResult, error) {
		if fields.Name != nil || fields.Description.Set || fields.PriceCents != nil {
			t.Fatalf("expected only DurationMinutes present, got %+v", fields)
		}
		if fields.DurationMinutes == nil || *fields.DurationMinutes != minutes {
			t.Fatalf("expected duration %d, got %v", minutes, fields.DurationMinutes)
		}
		return catalog.UpdateResult{Found: true, Service: catalog.Service{ID: "x"}}, nil
	}}
	svc := catalog.NewService(repo)

	_, err := svc.Update(context.Background(), "shop-1", "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4",
		catalog.UpdateInputRaw{DurationMinutes: &minutes})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdate_DescriptionSetToEmpty_ClearsDescription(t *testing.T) {
	empty := ""
	repo := &fakeRepository{updateFn: func(_ context.Context, _, _ string, fields catalog.UpdateFields) (catalog.UpdateResult, error) {
		if !fields.Description.Set || fields.Description.Value != nil {
			t.Fatalf("expected description cleared (Set=true, Value=nil), got %+v", fields.Description)
		}
		return catalog.UpdateResult{Found: true, Service: catalog.Service{ID: "x"}}, nil
	}}
	svc := catalog.NewService(repo)

	_, err := svc.Update(context.Background(), "shop-1", "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4",
		catalog.UpdateInputRaw{DescriptionSet: true, Description: &empty})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdate_DurationOutOfRange_Rejected(t *testing.T) {
	repo := &fakeRepository{updateFn: func(context.Context, string, string, catalog.UpdateFields) (catalog.UpdateResult, error) {
		t.Fatal("repository must not be called for an invalid duration")
		return catalog.UpdateResult{}, nil
	}}
	svc := catalog.NewService(repo)

	minutes := 0
	_, err := svc.Update(context.Background(), "shop-1", "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4",
		catalog.UpdateInputRaw{DurationMinutes: &minutes})
	mustBeValidation(t, err)
}

func TestUpdate_PriceZero_Rejected(t *testing.T) {
	repo := &fakeRepository{updateFn: func(context.Context, string, string, catalog.UpdateFields) (catalog.UpdateResult, error) {
		t.Fatal("repository must not be called for an invalid price")
		return catalog.UpdateResult{}, nil
	}}
	svc := catalog.NewService(repo)

	price := "0"
	_, err := svc.Update(context.Background(), "shop-1", "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4",
		catalog.UpdateInputRaw{Price: &price})
	mustBeValidation(t, err)
}

func TestUpdate_NotFound_TranslatesToNotFound(t *testing.T) {
	name := "Nuevo nombre"
	repo := &fakeRepository{updateFn: func(context.Context, string, string, catalog.UpdateFields) (catalog.UpdateResult, error) {
		return catalog.UpdateResult{Found: false}, nil
	}}
	svc := catalog.NewService(repo)

	_, err := svc.Update(context.Background(), "shop-1", "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4",
		catalog.UpdateInputRaw{Name: &name})
	mustBeNotFound(t, err)
}

func TestUpdate_NameTaken_TranslatesToConflict(t *testing.T) {
	name := "Corte clásico"
	repo := &fakeRepository{updateFn: func(context.Context, string, string, catalog.UpdateFields) (catalog.UpdateResult, error) {
		return catalog.UpdateResult{NameTaken: true}, nil
	}}
	svc := catalog.NewService(repo)

	_, err := svc.Update(context.Background(), "shop-1", "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4",
		catalog.UpdateInputRaw{Name: &name})
	mustBeConflict(t, err)
}

func TestUpdate_Found_ReturnsUpdatedServiceByID_NoDuplication(t *testing.T) {
	id := "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"
	name := "Nuevo nombre"
	updated := catalog.Service{ID: id, Name: "Nuevo nombre", PriceCents: 4500000, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	repo := &fakeRepository{updateFn: func(_ context.Context, _, serviceID string, fields catalog.UpdateFields) (catalog.UpdateResult, error) {
		if serviceID != id {
			t.Fatalf("expected repository to receive id %q, got %q", id, serviceID)
		}
		return catalog.UpdateResult{Found: true, Service: updated}, nil
	}}
	svc := catalog.NewService(repo)

	got, err := svc.Update(context.Background(), "shop-1", id, catalog.UpdateInputRaw{Name: &name})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != id || got.Name != "Nuevo nombre" {
		t.Fatalf("expected the same id with the new name, got %+v", got)
	}
}

// --- Get: not-found unifica inexistente y ajeno (CA-022-06) --------------

func TestGet_MalformedID_ReturnsNotFoundWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{getFn: func(context.Context, string, string) (catalog.Service, bool, error) {
		t.Fatal("repository must not be called for a malformed id")
		return catalog.Service{}, false, nil
	}}
	svc := catalog.NewService(repo)

	_, err := svc.Get(context.Background(), "shop-1", "not-a-uuid")
	mustBeNotFound(t, err)
}

func TestGet_NotFound_TranslatesToNotFound(t *testing.T) {
	repo := &fakeRepository{getFn: func(context.Context, string, string) (catalog.Service, bool, error) {
		return catalog.Service{}, false, nil
	}}
	svc := catalog.NewService(repo)

	_, err := svc.Get(context.Background(), "shop-1", "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4")
	mustBeNotFound(t, err)
}

func TestGet_Found_ReturnsService(t *testing.T) {
	want := catalog.Service{ID: "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4", Name: "Corte clásico", PriceCents: 4500000}
	repo := &fakeRepository{getFn: func(context.Context, string, string) (catalog.Service, bool, error) {
		return want, true, nil
	}}
	svc := catalog.NewService(repo)

	got, err := svc.Get(context.Background(), "shop-1", want.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
}

// --- List: clamping de limit y decodificación de cursor -------------------

func TestList_ZeroLimit_UsesDefault(t *testing.T) {
	repo := &fakeRepository{listFn: func(_ context.Context, _ string, _ *catalog.Cursor, limit int) (catalog.ListResult, error) {
		if limit != catalog.DefaultListLimit {
			t.Fatalf("expected default limit %d, got %d", catalog.DefaultListLimit, limit)
		}
		return catalog.ListResult{}, nil
	}}
	svc := catalog.NewService(repo)

	if _, err := svc.List(context.Background(), "shop-1", "", 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestList_LimitAboveMax_Clamped(t *testing.T) {
	repo := &fakeRepository{listFn: func(_ context.Context, _ string, _ *catalog.Cursor, limit int) (catalog.ListResult, error) {
		if limit != catalog.MaxListLimit {
			t.Fatalf("expected clamped limit %d, got %d", catalog.MaxListLimit, limit)
		}
		return catalog.ListResult{}, nil
	}}
	svc := catalog.NewService(repo)

	if _, err := svc.List(context.Background(), "shop-1", "", 999999); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestList_EmptyCursor_PassesNilCursor(t *testing.T) {
	repo := &fakeRepository{listFn: func(_ context.Context, _ string, cursor *catalog.Cursor, _ int) (catalog.ListResult, error) {
		if cursor != nil {
			t.Fatalf("expected nil cursor for the first page, got %+v", cursor)
		}
		return catalog.ListResult{}, nil
	}}
	svc := catalog.NewService(repo)

	if _, err := svc.List(context.Background(), "shop-1", "", 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestList_ValidCursor_DecodedAndPassedThrough(t *testing.T) {
	original := catalog.Cursor{CreatedAt: time.Now().UTC(), ID: "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"}
	token := catalog.EncodeCursor(original)

	repo := &fakeRepository{listFn: func(_ context.Context, _ string, cursor *catalog.Cursor, _ int) (catalog.ListResult, error) {
		if cursor == nil || cursor.ID != original.ID {
			t.Fatalf("expected decoded cursor with id %q, got %+v", original.ID, cursor)
		}
		return catalog.ListResult{}, nil
	}}
	svc := catalog.NewService(repo)

	if _, err := svc.List(context.Background(), "shop-1", token, 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestList_InvalidCursor_RejectedWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{listFn: func(context.Context, string, *catalog.Cursor, int) (catalog.ListResult, error) {
		t.Fatal("repository must not be called for an invalid cursor")
		return catalog.ListResult{}, nil
	}}
	svc := catalog.NewService(repo)

	_, err := svc.List(context.Background(), "shop-1", "not-a-valid-cursor!!!", 10)
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInvalid {
		t.Fatalf("expected apperr.KindInvalid, got %v", err)
	}
}

// --- PreviewDeactivation/Deactivate/Reactivate (HU-024) --------------------

const validServiceID = "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"

func TestPreviewDeactivation_MalformedID_ReturnsNotFoundWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{getFn: func(context.Context, string, string) (catalog.Service, bool, error) {
		t.Fatal("repository must not be called for a malformed id")
		return catalog.Service{}, false, nil
	}}
	svc := catalog.NewService(repo)

	_, err := svc.PreviewDeactivation(context.Background(), "shop-1", "not-a-uuid")
	mustBeNotFound(t, err)
}

func TestPreviewDeactivation_NotFound_TranslatesToNotFound(t *testing.T) {
	repo := &fakeRepository{getFn: func(context.Context, string, string) (catalog.Service, bool, error) {
		return catalog.Service{}, false, nil
	}}
	svc := catalog.NewService(repo)

	_, err := svc.PreviewDeactivation(context.Background(), "shop-1", validServiceID)
	mustBeNotFound(t, err)
}

func TestPreviewDeactivation_Found_ReturnsZeroAffectedAppointments(t *testing.T) {
	repo := &fakeRepository{getFn: func(context.Context, string, string) (catalog.Service, bool, error) {
		return catalog.Service{ID: validServiceID, IsActive: true}, true, nil
	}}
	svc := catalog.NewService(repo)

	impact, err := svc.PreviewDeactivation(context.Background(), "shop-1", validServiceID)
	if err != nil {
		t.Fatalf("PreviewDeactivation: %v", err)
	}
	// DEC-069: B1 no tiene appointment todavía, así que el conteo real es
	// siempre 0.
	if impact.AffectedAppointments != 0 {
		t.Fatalf("expected AffectedAppointments=0 (DEC-069), got %d", impact.AffectedAppointments)
	}
}

func failingLifecycleFn(t *testing.T) func(context.Context, string, string, idempotency.Key, idempotency.Fingerprint) (catalog.LifecycleResult, error) {
	return func(context.Context, string, string, idempotency.Key, idempotency.Fingerprint) (catalog.LifecycleResult, error) {
		t.Fatal("repository must not be called for a malformed id")
		return catalog.LifecycleResult{}, nil
	}
}

func TestDeactivate_MalformedID_ReturnsNotFoundWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{deactivateFn: failingLifecycleFn(t)}
	svc := catalog.NewService(repo)

	_, err := svc.Deactivate(context.Background(), "shop-1", "not-a-uuid", "key-1", "a")
	mustBeNotFound(t, err)
}

func TestDeactivate_NotFound_TranslatesToNotFound(t *testing.T) {
	repo := &fakeRepository{deactivateFn: func(context.Context, string, string, idempotency.Key, idempotency.Fingerprint) (catalog.LifecycleResult, error) {
		return catalog.LifecycleResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Found:    false,
		}, nil
	}}
	svc := catalog.NewService(repo)

	_, err := svc.Deactivate(context.Background(), "shop-1", validServiceID, "key-1", "a")
	mustBeNotFound(t, err)
}

func TestDeactivate_InvalidTransition_TranslatesToConflict(t *testing.T) {
	repo := &fakeRepository{deactivateFn: func(context.Context, string, string, idempotency.Key, idempotency.Fingerprint) (catalog.LifecycleResult, error) {
		return catalog.LifecycleResult{
			Decision:          idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Found:             true,
			InvalidTransition: true,
		}, nil
	}}
	svc := catalog.NewService(repo)

	_, err := svc.Deactivate(context.Background(), "shop-1", validServiceID, "key-1", "a")
	mustBeConflict(t, err)
}

func TestDeactivate_Proceed_ReturnsResultUntouched(t *testing.T) {
	want := catalog.LifecycleResult{
		Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed},
		Found:    true,
		Service:  catalog.Service{ID: validServiceID, IsActive: false},
	}
	repo := &fakeRepository{deactivateFn: func(context.Context, string, string, idempotency.Key, idempotency.Fingerprint) (catalog.LifecycleResult, error) {
		return want, nil
	}}
	svc := catalog.NewService(repo)

	got, err := svc.Deactivate(context.Background(), "shop-1", validServiceID, "key-1", "a")
	if err != nil {
		t.Fatalf("Deactivate: %v", err)
	}
	if got.Service.ID != want.Service.ID || got.Service.IsActive != want.Service.IsActive {
		t.Fatalf("expected the repository result untouched, got %+v want %+v", got, want)
	}
}

func TestDeactivate_ReplayOutcome_PassedThroughWithoutTranslatingAsError(t *testing.T) {
	repo := &fakeRepository{deactivateFn: func(context.Context, string, string, idempotency.Key, idempotency.Fingerprint) (catalog.LifecycleResult, error) {
		return catalog.LifecycleResult{
			Decision: idempotency.Decision{Outcome: idempotency.OutcomeReplay},
			Response: idempotency.StoredResponse{Status: 200, Body: `{"service":{}}`},
		}, nil
	}}
	svc := catalog.NewService(repo)

	// CA-024-06: una repetición exacta (mismo Outcome=Replay) nunca se
	// traduce a NotFound/Conflict, aunque Found/InvalidTransition queden en
	// su valor cero por defecto (no son significativos fuera de
	// OutcomeProceed).
	got, err := svc.Deactivate(context.Background(), "shop-1", validServiceID, "key-1", "a")
	if err != nil {
		t.Fatalf("Deactivate: %v", err)
	}
	if got.Decision.Outcome != idempotency.OutcomeReplay {
		t.Fatalf("expected OutcomeReplay passed through, got %s", got.Decision.Outcome)
	}
}

func TestReactivate_MalformedID_ReturnsNotFoundWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{reactivateFn: failingLifecycleFn(t)}
	svc := catalog.NewService(repo)

	_, err := svc.Reactivate(context.Background(), "shop-1", "not-a-uuid", "key-1", "a")
	mustBeNotFound(t, err)
}

func TestReactivate_NotFound_TranslatesToNotFound(t *testing.T) {
	repo := &fakeRepository{reactivateFn: func(context.Context, string, string, idempotency.Key, idempotency.Fingerprint) (catalog.LifecycleResult, error) {
		return catalog.LifecycleResult{Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed}, Found: false}, nil
	}}
	svc := catalog.NewService(repo)

	_, err := svc.Reactivate(context.Background(), "shop-1", validServiceID, "key-1", "a")
	mustBeNotFound(t, err)
}

func TestReactivate_InvalidTransition_TranslatesToConflict(t *testing.T) {
	repo := &fakeRepository{reactivateFn: func(context.Context, string, string, idempotency.Key, idempotency.Fingerprint) (catalog.LifecycleResult, error) {
		return catalog.LifecycleResult{
			Decision:          idempotency.Decision{Outcome: idempotency.OutcomeProceed},
			Found:             true,
			InvalidTransition: true,
		}, nil
	}}
	svc := catalog.NewService(repo)

	_, err := svc.Reactivate(context.Background(), "shop-1", validServiceID, "key-1", "a")
	mustBeConflict(t, err)
}

func TestReactivate_Proceed_ReturnsResultUntouched(t *testing.T) {
	want := catalog.LifecycleResult{
		Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed},
		Found:    true,
		Service:  catalog.Service{ID: validServiceID, IsActive: true},
	}
	repo := &fakeRepository{reactivateFn: func(context.Context, string, string, idempotency.Key, idempotency.Fingerprint) (catalog.LifecycleResult, error) {
		return want, nil
	}}
	svc := catalog.NewService(repo)

	got, err := svc.Reactivate(context.Background(), "shop-1", validServiceID, "key-1", "a")
	if err != nil {
		t.Fatalf("Reactivate: %v", err)
	}
	if got.Service.ID != want.Service.ID || !got.Service.IsActive {
		t.Fatalf("expected the repository result untouched, got %+v want %+v", got, want)
	}
}
