package staff_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"system-barbershop/internal/modules/staff"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/idempotency"
)

// fakeRepository es un doble de staff.Repository para probar staff.Service
// sin PostgreSQL real: solo prueba la orquestación (validación antes de
// tocar el repositorio, traducción de resultados a apperr), nunca SQL ni
// RLS (eso vive en postgres/repository_test.go contra PostgreSQL real).
type fakeRepository struct {
	listFn   func(ctx context.Context, barbershopID string, cursor *staff.Cursor, limit int) (staff.ListResult, error)
	getFn    func(ctx context.Context, barbershopID, barberID string) (staff.Barber, bool, error)
	createFn func(ctx context.Context, barbershopID, fullName string, key idempotency.Key, fingerprint idempotency.Fingerprint) (staff.CreateResult, error)
	renameFn func(ctx context.Context, barbershopID, barberID, fullName string) (staff.RenameResult, error)

	createCalls int
}

func (f *fakeRepository) List(ctx context.Context, barbershopID string, cursor *staff.Cursor, limit int) (staff.ListResult, error) {
	return f.listFn(ctx, barbershopID, cursor, limit)
}

func (f *fakeRepository) Get(ctx context.Context, barbershopID, barberID string) (staff.Barber, bool, error) {
	return f.getFn(ctx, barbershopID, barberID)
}

func (f *fakeRepository) Create(ctx context.Context, barbershopID, fullName string, key idempotency.Key, fingerprint idempotency.Fingerprint) (staff.CreateResult, error) {
	f.createCalls++
	return f.createFn(ctx, barbershopID, fullName, key, fingerprint)
}

func (f *fakeRepository) Rename(ctx context.Context, barbershopID, barberID, fullName string) (staff.RenameResult, error) {
	return f.renameFn(ctx, barbershopID, barberID, fullName)
}

var _ staff.Repository = (*fakeRepository)(nil)

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

// --- Create: validación de nombre (CA-021-03) --------------------------

func TestCreate_EmptyName_RejectedWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{createFn: func(context.Context, string, string, idempotency.Key, idempotency.Fingerprint) (staff.CreateResult, error) {
		t.Fatal("repository must not be called for an invalid name")
		return staff.CreateResult{}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Create(context.Background(), "shop-1", "", "key-1", "a")
	mustBeValidation(t, err)
	if repo.createCalls != 0 {
		t.Fatalf("expected 0 repository calls, got %d", repo.createCalls)
	}
}

func TestCreate_WhitespaceOnlyName_Rejected(t *testing.T) {
	repo := &fakeRepository{createFn: func(context.Context, string, string, idempotency.Key, idempotency.Fingerprint) (staff.CreateResult, error) {
		t.Fatal("repository must not be called")
		return staff.CreateResult{}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Create(context.Background(), "shop-1", "   \t\n  ", "key-1", "a")
	mustBeValidation(t, err)
}

func TestCreate_NameOver120Characters_Rejected(t *testing.T) {
	repo := &fakeRepository{createFn: func(context.Context, string, string, idempotency.Key, idempotency.Fingerprint) (staff.CreateResult, error) {
		t.Fatal("repository must not be called")
		return staff.CreateResult{}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Create(context.Background(), "shop-1", strings.Repeat("a", 121), "key-1", "a")
	mustBeValidation(t, err)
}

func TestCreate_NameExactly120Characters_Accepted(t *testing.T) {
	name := strings.Repeat("a", 120)
	repo := &fakeRepository{createFn: func(_ context.Context, _ string, fullName string, _ idempotency.Key, _ idempotency.Fingerprint) (staff.CreateResult, error) {
		if fullName != name {
			t.Fatalf("expected repository to receive the trimmed name %q, got %q", name, fullName)
		}
		return staff.CreateResult{Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed}}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Create(context.Background(), "shop-1", name, "key-1", "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreate_TrimsNameBeforeCallingRepository(t *testing.T) {
	repo := &fakeRepository{createFn: func(_ context.Context, _ string, fullName string, _ idempotency.Key, _ idempotency.Fingerprint) (staff.CreateResult, error) {
		if fullName != "Carlos Ramírez" {
			t.Fatalf("expected trimmed name, got %q", fullName)
		}
		return staff.CreateResult{Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed}}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Create(context.Background(), "shop-1", "  Carlos Ramírez  ", "key-1", "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreate_UnicodeName_CountsRunesNotBytes(t *testing.T) {
	// 120 runas de "ñ" (2 bytes cada una en UTF-8, 240 bytes en total): debe
	// aceptarse porque la regla cuenta caracteres, no bytes (CA-021-03).
	name := strings.Repeat("ñ", 120)
	repo := &fakeRepository{createFn: func(context.Context, string, string, idempotency.Key, idempotency.Fingerprint) (staff.CreateResult, error) {
		return staff.CreateResult{Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed}}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Create(context.Background(), "shop-1", name, "key-1", "a")
	if err != nil {
		t.Fatalf("unexpected error for a 120-rune unicode name: %v", err)
	}
}

func TestCreate_PropagatesRepositoryResultUntouched(t *testing.T) {
	want := staff.CreateResult{
		Decision: idempotency.Decision{Outcome: idempotency.OutcomeReplay, Response: idempotency.StoredResponse{Status: 201, Body: `{"id":"x"}`}},
		Response: idempotency.StoredResponse{Status: 201, Body: `{"id":"x"}`},
	}
	repo := &fakeRepository{createFn: func(context.Context, string, string, idempotency.Key, idempotency.Fingerprint) (staff.CreateResult, error) {
		return want, nil
	}}
	svc := staff.NewService(repo)

	got, err := svc.Create(context.Background(), "shop-1", "Carlos", "key-1", "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Decision.Outcome != idempotency.OutcomeReplay || got.Response.Body != `{"id":"x"}` {
		t.Fatalf("expected the repository result to pass through untouched, got %+v", got)
	}
}

// --- Rename: mismas reglas de nombre, además de not-found ----------------

func TestRename_EmptyName_RejectedWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{renameFn: func(context.Context, string, string, string) (staff.RenameResult, error) {
		t.Fatal("repository must not be called for an invalid name")
		return staff.RenameResult{}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Rename(context.Background(), "shop-1", "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4", "")
	mustBeValidation(t, err)
}

func TestRename_MalformedID_ReturnsNotFoundWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{renameFn: func(context.Context, string, string, string) (staff.RenameResult, error) {
		t.Fatal("repository must not be called for a malformed id")
		return staff.RenameResult{}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Rename(context.Background(), "shop-1", "not-a-uuid", "Nombre Válido")
	mustBeNotFound(t, err)
}

func TestRename_NotFound_TranslatesToNotFound(t *testing.T) {
	repo := &fakeRepository{renameFn: func(context.Context, string, string, string) (staff.RenameResult, error) {
		return staff.RenameResult{Found: false}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Rename(context.Background(), "shop-1", "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4", "Nombre Válido")
	mustBeNotFound(t, err)
}

func TestRename_Found_ReturnsUpdatedBarberByID_NoDuplication(t *testing.T) {
	id := "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"
	updated := staff.Barber{ID: id, FullName: "Nuevo Nombre", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	repo := &fakeRepository{renameFn: func(_ context.Context, _, barberID, fullName string) (staff.RenameResult, error) {
		if barberID != id {
			t.Fatalf("expected repository to receive id %q, got %q", id, barberID)
		}
		if fullName != "Nuevo Nombre" {
			t.Fatalf("expected trimmed name, got %q", fullName)
		}
		return staff.RenameResult{Found: true, Barber: updated}, nil
	}}
	svc := staff.NewService(repo)

	got, err := svc.Rename(context.Background(), "shop-1", id, "  Nuevo Nombre  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != id || got.FullName != "Nuevo Nombre" {
		t.Fatalf("expected the same id with the new name, got %+v", got)
	}
}

// --- Get: not-found unifica inexistente y ajeno (CA-021-05) --------------

func TestGet_MalformedID_ReturnsNotFoundWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{getFn: func(context.Context, string, string) (staff.Barber, bool, error) {
		t.Fatal("repository must not be called for a malformed id")
		return staff.Barber{}, false, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Get(context.Background(), "shop-1", "not-a-uuid")
	mustBeNotFound(t, err)
}

func TestGet_NotFound_TranslatesToNotFound(t *testing.T) {
	repo := &fakeRepository{getFn: func(context.Context, string, string) (staff.Barber, bool, error) {
		return staff.Barber{}, false, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Get(context.Background(), "shop-1", "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4")
	mustBeNotFound(t, err)
}

func TestGet_Found_ReturnsBarber(t *testing.T) {
	want := staff.Barber{ID: "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4", FullName: "Carlos"}
	repo := &fakeRepository{getFn: func(context.Context, string, string) (staff.Barber, bool, error) {
		return want, true, nil
	}}
	svc := staff.NewService(repo)

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
	repo := &fakeRepository{listFn: func(_ context.Context, _ string, _ *staff.Cursor, limit int) (staff.ListResult, error) {
		if limit != staff.DefaultListLimit {
			t.Fatalf("expected default limit %d, got %d", staff.DefaultListLimit, limit)
		}
		return staff.ListResult{}, nil
	}}
	svc := staff.NewService(repo)

	if _, err := svc.List(context.Background(), "shop-1", "", 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestList_LimitAboveMax_Clamped(t *testing.T) {
	repo := &fakeRepository{listFn: func(_ context.Context, _ string, _ *staff.Cursor, limit int) (staff.ListResult, error) {
		if limit != staff.MaxListLimit {
			t.Fatalf("expected clamped limit %d, got %d", staff.MaxListLimit, limit)
		}
		return staff.ListResult{}, nil
	}}
	svc := staff.NewService(repo)

	if _, err := svc.List(context.Background(), "shop-1", "", 999999); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestList_NegativeLimit_ClampedToMin(t *testing.T) {
	// limit<=0 se trata como "no especificado" y usa el valor por defecto,
	// no el mínimo: un cliente que nunca envía limit debe obtener una
	// página razonable, no una de un solo elemento.
	repo := &fakeRepository{listFn: func(_ context.Context, _ string, _ *staff.Cursor, limit int) (staff.ListResult, error) {
		if limit != staff.DefaultListLimit {
			t.Fatalf("expected default limit for a negative input, got %d", limit)
		}
		return staff.ListResult{}, nil
	}}
	svc := staff.NewService(repo)

	if _, err := svc.List(context.Background(), "shop-1", "", -5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestList_EmptyCursor_PassesNilCursor(t *testing.T) {
	repo := &fakeRepository{listFn: func(_ context.Context, _ string, cursor *staff.Cursor, _ int) (staff.ListResult, error) {
		if cursor != nil {
			t.Fatalf("expected nil cursor for the first page, got %+v", cursor)
		}
		return staff.ListResult{}, nil
	}}
	svc := staff.NewService(repo)

	if _, err := svc.List(context.Background(), "shop-1", "", 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestList_ValidCursor_DecodedAndPassedThrough(t *testing.T) {
	original := staff.Cursor{CreatedAt: time.Now().UTC(), ID: "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"}
	token := staff.EncodeCursor(original)

	repo := &fakeRepository{listFn: func(_ context.Context, _ string, cursor *staff.Cursor, _ int) (staff.ListResult, error) {
		if cursor == nil || cursor.ID != original.ID {
			t.Fatalf("expected decoded cursor with id %q, got %+v", original.ID, cursor)
		}
		return staff.ListResult{}, nil
	}}
	svc := staff.NewService(repo)

	if _, err := svc.List(context.Background(), "shop-1", token, 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestList_InvalidCursor_RejectedWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{listFn: func(context.Context, string, *staff.Cursor, int) (staff.ListResult, error) {
		t.Fatal("repository must not be called for an invalid cursor")
		return staff.ListResult{}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.List(context.Background(), "shop-1", "not-a-valid-cursor!!!", 10)
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInvalid {
		t.Fatalf("expected apperr.KindInvalid, got %v", err)
	}
}
