package staff_test

import (
	"bytes"
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
	listPageFn    func(context.Context, string, int, int) (staff.PageResult, error)
	listFn        func(ctx context.Context, barbershopID string, cursor *staff.Cursor, limit int) (staff.ListResult, error)
	getFn         func(ctx context.Context, barbershopID, barberID string) (staff.Barber, bool, error)
	createFn      func(ctx context.Context, barbershopID, fullName string, description *string, key idempotency.Key, fingerprint idempotency.Fingerprint) (staff.CreateResult, error)
	updateFn      func(ctx context.Context, barbershopID, barberID, fullName string, description staff.DescriptionPatch) (staff.UpdateResult, error)
	putPhotoFn    func(ctx context.Context, barbershopID, barberID string, photo staff.Photo) (staff.PhotoResult, error)
	getPhotoFn    func(ctx context.Context, barbershopID, barberID string) (staff.StoredPhoto, bool, error)
	deletePhotoFn func(ctx context.Context, barbershopID, barberID string) (bool, error)
	getLinkedFn   func(ctx context.Context, barbershopID, staffUserID string) (staff.Barber, bool, error)
	linkFn        func(ctx context.Context, barbershopID, staffUserID, barberID string) (staff.LinkResult, error)
	unlinkFn      func(ctx context.Context, barbershopID, staffUserID string) (string, error)

	createCalls int
}

func (f *fakeRepository) List(ctx context.Context, barbershopID string, cursor *staff.Cursor, limit int) (staff.ListResult, error) {
	return f.listFn(ctx, barbershopID, cursor, limit)
}

func (f *fakeRepository) Get(ctx context.Context, barbershopID, barberID string) (staff.Barber, bool, error) {
	return f.getFn(ctx, barbershopID, barberID)
}

func (f *fakeRepository) Create(ctx context.Context, barbershopID, fullName string, description *string, key idempotency.Key, fingerprint idempotency.Fingerprint) (staff.CreateResult, error) {
	f.createCalls++
	return f.createFn(ctx, barbershopID, fullName, description, key, fingerprint)
}

func (f *fakeRepository) Update(ctx context.Context, barbershopID, barberID, fullName string, description staff.DescriptionPatch) (staff.UpdateResult, error) {
	return f.updateFn(ctx, barbershopID, barberID, fullName, description)
}

func (f *fakeRepository) PutPhoto(ctx context.Context, barbershopID, barberID string, photo staff.Photo) (staff.PhotoResult, error) {
	return f.putPhotoFn(ctx, barbershopID, barberID, photo)
}

func (f *fakeRepository) GetPhoto(ctx context.Context, barbershopID, barberID string) (staff.StoredPhoto, bool, error) {
	return f.getPhotoFn(ctx, barbershopID, barberID)
}

func (f *fakeRepository) DeletePhoto(ctx context.Context, barbershopID, barberID string) (bool, error) {
	return f.deletePhotoFn(ctx, barbershopID, barberID)
}

func (f *fakeRepository) GetLinked(ctx context.Context, barbershopID, staffUserID string) (staff.Barber, bool, error) {
	return f.getLinkedFn(ctx, barbershopID, staffUserID)
}

func (f *fakeRepository) Link(ctx context.Context, barbershopID, staffUserID, barberID string) (staff.LinkResult, error) {
	return f.linkFn(ctx, barbershopID, staffUserID, barberID)
}

func (f *fakeRepository) Unlink(ctx context.Context, barbershopID, staffUserID string) (string, error) {
	return f.unlinkFn(ctx, barbershopID, staffUserID)
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
	repo := &fakeRepository{createFn: func(context.Context, string, string, *string, idempotency.Key, idempotency.Fingerprint) (staff.CreateResult, error) {
		t.Fatal("repository must not be called for an invalid name")
		return staff.CreateResult{}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Create(context.Background(), "shop-1", "", nil, "key-1", "a")
	mustBeValidation(t, err)
	if repo.createCalls != 0 {
		t.Fatalf("expected 0 repository calls, got %d", repo.createCalls)
	}
}

func TestCreate_WhitespaceOnlyName_Rejected(t *testing.T) {
	repo := &fakeRepository{createFn: func(context.Context, string, string, *string, idempotency.Key, idempotency.Fingerprint) (staff.CreateResult, error) {
		t.Fatal("repository must not be called")
		return staff.CreateResult{}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Create(context.Background(), "shop-1", "   \t\n  ", nil, "key-1", "a")
	mustBeValidation(t, err)
}

func TestCreate_NameOver120Characters_Rejected(t *testing.T) {
	repo := &fakeRepository{createFn: func(context.Context, string, string, *string, idempotency.Key, idempotency.Fingerprint) (staff.CreateResult, error) {
		t.Fatal("repository must not be called")
		return staff.CreateResult{}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Create(context.Background(), "shop-1", strings.Repeat("a", 121), nil, "key-1", "a")
	mustBeValidation(t, err)
}

func TestCreate_NameExactly120Characters_Accepted(t *testing.T) {
	name := strings.Repeat("a", 120)
	repo := &fakeRepository{createFn: func(_ context.Context, _ string, fullName string, _ *string, _ idempotency.Key, _ idempotency.Fingerprint) (staff.CreateResult, error) {
		if fullName != name {
			t.Fatalf("expected repository to receive the trimmed name %q, got %q", name, fullName)
		}
		return staff.CreateResult{Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed}}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Create(context.Background(), "shop-1", name, nil, "key-1", "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreate_TrimsNameBeforeCallingRepository(t *testing.T) {
	repo := &fakeRepository{createFn: func(_ context.Context, _ string, fullName string, _ *string, _ idempotency.Key, _ idempotency.Fingerprint) (staff.CreateResult, error) {
		if fullName != "Carlos Ramírez" {
			t.Fatalf("expected trimmed name, got %q", fullName)
		}
		return staff.CreateResult{Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed}}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Create(context.Background(), "shop-1", "  Carlos Ramírez  ", nil, "key-1", "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreate_UnicodeName_CountsRunesNotBytes(t *testing.T) {
	// 120 runas de "ñ" (2 bytes cada una en UTF-8, 240 bytes en total): debe
	// aceptarse porque la regla cuenta caracteres, no bytes (CA-021-03).
	name := strings.Repeat("ñ", 120)
	repo := &fakeRepository{createFn: func(context.Context, string, string, *string, idempotency.Key, idempotency.Fingerprint) (staff.CreateResult, error) {
		return staff.CreateResult{Decision: idempotency.Decision{Outcome: idempotency.OutcomeProceed}}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Create(context.Background(), "shop-1", name, nil, "key-1", "a")
	if err != nil {
		t.Fatalf("unexpected error for a 120-rune unicode name: %v", err)
	}
}

func TestCreate_PropagatesRepositoryResultUntouched(t *testing.T) {
	want := staff.CreateResult{
		Decision: idempotency.Decision{Outcome: idempotency.OutcomeReplay, Response: idempotency.StoredResponse{Status: 201, Body: `{"id":"x"}`}},
		Response: idempotency.StoredResponse{Status: 201, Body: `{"id":"x"}`},
	}
	repo := &fakeRepository{createFn: func(context.Context, string, string, *string, idempotency.Key, idempotency.Fingerprint) (staff.CreateResult, error) {
		return want, nil
	}}
	svc := staff.NewService(repo)

	got, err := svc.Create(context.Background(), "shop-1", "Carlos", nil, "key-1", "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Decision.Outcome != idempotency.OutcomeReplay || got.Response.Body != `{"id":"x"}` {
		t.Fatalf("expected the repository result to pass through untouched, got %+v", got)
	}
}

// --- Rename: mismas reglas de nombre, además de not-found ----------------

func TestRename_EmptyName_RejectedWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{updateFn: func(context.Context, string, string, string, staff.DescriptionPatch) (staff.UpdateResult, error) {
		t.Fatal("repository must not be called for an invalid name")
		return staff.UpdateResult{}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Update(context.Background(), "shop-1", "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4", "", staff.DescriptionPatch{})
	mustBeValidation(t, err)
}

func TestRename_MalformedID_ReturnsNotFoundWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{updateFn: func(context.Context, string, string, string, staff.DescriptionPatch) (staff.UpdateResult, error) {
		t.Fatal("repository must not be called for a malformed id")
		return staff.UpdateResult{}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Update(context.Background(), "shop-1", "not-a-uuid", "Nombre Válido", staff.DescriptionPatch{})
	mustBeNotFound(t, err)
}

func TestRename_NotFound_TranslatesToNotFound(t *testing.T) {
	repo := &fakeRepository{updateFn: func(context.Context, string, string, string, staff.DescriptionPatch) (staff.UpdateResult, error) {
		return staff.UpdateResult{Found: false}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Update(context.Background(), "shop-1", "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4", "Nombre Válido", staff.DescriptionPatch{})
	mustBeNotFound(t, err)
}

func TestRename_Found_ReturnsUpdatedBarberByID_NoDuplication(t *testing.T) {
	id := "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"
	updated := staff.Barber{ID: id, FullName: "Nuevo Nombre", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	repo := &fakeRepository{updateFn: func(_ context.Context, _, barberID, fullName string, _ staff.DescriptionPatch) (staff.UpdateResult, error) {
		if barberID != id {
			t.Fatalf("expected repository to receive id %q, got %q", id, barberID)
		}
		if fullName != "Nuevo Nombre" {
			t.Fatalf("expected trimmed name, got %q", fullName)
		}
		return staff.UpdateResult{Found: true, Barber: updated}, nil
	}}
	svc := staff.NewService(repo)

	got, err := svc.Update(context.Background(), "shop-1", id, "  Nuevo Nombre  ", staff.DescriptionPatch{})
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

// --- Fotografía (DEC-104) -------------------------------------------------

const validBarberID = "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"

func TestSetPhoto_InvalidImage_RejectedWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{putPhotoFn: func(context.Context, string, string, staff.Photo) (staff.PhotoResult, error) {
		t.Fatal("repository must not be called for an invalid image")
		return staff.PhotoResult{}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.SetPhoto(context.Background(), "shop-1", validBarberID, staff.PhotoContentTypeJPEG, []byte("no es una imagen"))
	mustBeValidation(t, err)
}

func TestSetPhoto_MalformedBarberID_NotFoundWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{putPhotoFn: func(context.Context, string, string, staff.Photo) (staff.PhotoResult, error) {
		t.Fatal("repository must not be called for a malformed id")
		return staff.PhotoResult{}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.SetPhoto(context.Background(), "shop-1", "no-uuid", staff.PhotoContentTypePNG, solidImage(t, "png", 128, 128))
	mustBeNotFound(t, err)
}

func TestSetPhoto_BarberMissing_NotFound(t *testing.T) {
	repo := &fakeRepository{putPhotoFn: func(context.Context, string, string, staff.Photo) (staff.PhotoResult, error) {
		return staff.PhotoResult{Found: false}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.SetPhoto(context.Background(), "shop-1", validBarberID, staff.PhotoContentTypePNG, solidImage(t, "png", 128, 128))
	mustBeNotFound(t, err)
}

func TestSetPhoto_Success_ForwardsValidatedPhotoAndReturnsBarber(t *testing.T) {
	data := solidImage(t, "jpeg", 256, 256)
	now := time.Now().UTC()
	var got staff.Photo
	repo := &fakeRepository{putPhotoFn: func(_ context.Context, shop, id string, photo staff.Photo) (staff.PhotoResult, error) {
		got = photo
		if shop != "shop-1" || id != validBarberID {
			t.Fatalf("unexpected scope: %s %s", shop, id)
		}
		return staff.PhotoResult{Found: true, Barber: staff.Barber{ID: id, FullName: "Carlos", PhotoUpdatedAt: &now}}, nil
	}}
	svc := staff.NewService(repo)

	barber, err := svc.SetPhoto(context.Background(), "shop-1", validBarberID, staff.PhotoContentTypeJPEG, data)
	if err != nil {
		t.Fatalf("SetPhoto: %v", err)
	}
	if got.ContentType != staff.PhotoContentTypeJPEG || !bytes.Equal(got.Data, data) {
		t.Fatalf("the validated photo was not forwarded: %s", got.ContentType)
	}
	if barber.PhotoUpdatedAt == nil {
		t.Fatal("expected photoUpdatedAt on the returned barber")
	}
}

func TestPhoto_NotFound_CoversMissingBarberAndMissingPhotoAlike(t *testing.T) {
	repo := &fakeRepository{getPhotoFn: func(context.Context, string, string) (staff.StoredPhoto, bool, error) {
		return staff.StoredPhoto{}, false, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Photo(context.Background(), "shop-1", validBarberID)
	mustBeNotFound(t, err)
	_, err = svc.Photo(context.Background(), "shop-1", "no-uuid")
	mustBeNotFound(t, err)
}

func TestRemovePhoto_MissingBarber_NotFound_ButMissingPhotoIsSuccess(t *testing.T) {
	found := false
	repo := &fakeRepository{deletePhotoFn: func(context.Context, string, string) (bool, error) { return found, nil }}
	svc := staff.NewService(repo)

	mustBeNotFound(t, svc.RemovePhoto(context.Background(), "shop-1", validBarberID))

	found = true
	if err := svc.RemovePhoto(context.Background(), "shop-1", validBarberID); err != nil {
		t.Fatalf("removing when the barber exists must succeed, got %v", err)
	}
}

func (f *fakeRepository) ListPage(ctx context.Context, shop string, page, size int) (staff.PageResult, error) {
	if f.listPageFn != nil {
		return f.listPageFn(ctx, shop, page, size)
	}
	return staff.PageResult{}, nil
}

func TestListPage_DefaultsBoundsAndCanceledContext(t *testing.T) {
	for _, tc := range []struct{ page, size, wantPage, wantSize int }{{0, 0, 1, 20}, {2, 999, 2, 50}, {4, 1, 4, 1}} {
		called := false
		repo := &fakeRepository{listPageFn: func(_ context.Context, shop string, page, size int) (staff.PageResult, error) {
			called = true
			if shop != "shop-1" || page != tc.wantPage || size != tc.wantSize {
				t.Fatal("wrong normalized parameters")
			}
			return staff.PageResult{}, nil
		}}
		svc := staff.NewService(repo)
		if _, err := svc.ListPage(context.Background(), "shop-1", tc.page, tc.size); err != nil {
			t.Fatal(err)
		}
		if !called {
			t.Fatal("repository not called")
		}
		called = false
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := svc.ListPage(ctx, "shop-1", tc.page, tc.size); err == nil || called {
			t.Fatal("canceled context reached repository")
		}
	}
}

// --- Descripción del barbero (DEC-127) -------------------------------------

func descPtr(s string) *string { return &s }

func TestCreate_Description_IsTrimmedBeforeReachingRepository(t *testing.T) {
	var got *string
	repo := &fakeRepository{createFn: func(_ context.Context, _, _ string, description *string, _ idempotency.Key, _ idempotency.Fingerprint) (staff.CreateResult, error) {
		got = description
		return staff.CreateResult{}, nil
	}}
	svc := staff.NewService(repo)

	if _, err := svc.Create(context.Background(), "shop-1", "Carlos", descPtr("  Fade y barba  "), "key-1", "a"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || *got != "Fade y barba" {
		t.Fatalf("expected the trimmed description, got %v", got)
	}
}

func TestCreate_BlankDescription_BecomesNoDescription(t *testing.T) {
	for name, raw := range map[string]*string{"ausente": nil, "vacía": descPtr(""), "solo espacios": descPtr(" \t\n ")} {
		t.Run(name, func(t *testing.T) {
			called := false
			repo := &fakeRepository{createFn: func(_ context.Context, _, _ string, description *string, _ idempotency.Key, _ idempotency.Fingerprint) (staff.CreateResult, error) {
				called = true
				if description != nil {
					t.Fatalf("a blank description must reach the repository as nil, got %q", *description)
				}
				return staff.CreateResult{}, nil
			}}
			svc := staff.NewService(repo)

			if _, err := svc.Create(context.Background(), "shop-1", "Carlos", raw, "key-1", "a"); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !called {
				t.Fatal("expected the repository to be called")
			}
		})
	}
}

func TestCreate_DescriptionTooLong_RejectedWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{createFn: func(context.Context, string, string, *string, idempotency.Key, idempotency.Fingerprint) (staff.CreateResult, error) {
		t.Fatal("repository must not be called for an invalid description")
		return staff.CreateResult{}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Create(context.Background(), "shop-1", "Carlos", descPtr(strings.Repeat("a", staff.DescriptionMaxLength+1)), "key-1", "a")
	mustBeValidation(t, err)
	if repo.createCalls != 0 {
		t.Fatalf("expected 0 repository calls, got %d", repo.createCalls)
	}
}

// El límite se cuenta en caracteres Unicode, no en bytes: 280 «ñ» (560 bytes)
// son válidas y 281 no.
func TestCreate_DescriptionLimit_CountsRunesNotBytes(t *testing.T) {
	repo := &fakeRepository{createFn: func(context.Context, string, string, *string, idempotency.Key, idempotency.Fingerprint) (staff.CreateResult, error) {
		return staff.CreateResult{}, nil
	}}
	svc := staff.NewService(repo)

	if _, err := svc.Create(context.Background(), "shop-1", "Carlos", descPtr(strings.Repeat("ñ", staff.DescriptionMaxLength)), "key-1", "a"); err != nil {
		t.Fatalf("280 runes must be accepted: %v", err)
	}
	_, err := svc.Create(context.Background(), "shop-1", "Carlos", descPtr(strings.Repeat("ñ", staff.DescriptionMaxLength+1)), "key-2", "b")
	mustBeValidation(t, err)
}

func TestUpdate_Description_NotSetKeepsAndSetNormalizes(t *testing.T) {
	id := "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"
	var got staff.DescriptionPatch
	repo := &fakeRepository{updateFn: func(_ context.Context, _, _, _ string, description staff.DescriptionPatch) (staff.UpdateResult, error) {
		got = description
		return staff.UpdateResult{Found: true, Barber: staff.Barber{ID: id}}, nil
	}}
	svc := staff.NewService(repo)

	if _, err := svc.Update(context.Background(), "shop-1", id, "Carlos", staff.DescriptionPatch{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Set {
		t.Fatal("a patch without description must not ask the repository to write it")
	}

	if _, err := svc.Update(context.Background(), "shop-1", id, "Carlos", staff.DescriptionPatch{Set: true, Value: descPtr("  Navaja  ")}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Set || got.Value == nil || *got.Value != "Navaja" {
		t.Fatalf("expected a trimmed description to be written, got %+v", got)
	}

	if _, err := svc.Update(context.Background(), "shop-1", id, "Carlos", staff.DescriptionPatch{Set: true, Value: descPtr("   ")}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Set || got.Value != nil {
		t.Fatalf("a blank description must clear it (Set with nil), got %+v", got)
	}
}

func TestUpdate_DescriptionTooLong_RejectedWithoutTouchingRepository(t *testing.T) {
	repo := &fakeRepository{updateFn: func(context.Context, string, string, string, staff.DescriptionPatch) (staff.UpdateResult, error) {
		t.Fatal("repository must not be called for an invalid description")
		return staff.UpdateResult{}, nil
	}}
	svc := staff.NewService(repo)

	_, err := svc.Update(context.Background(), "shop-1", "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4", "Carlos",
		staff.DescriptionPatch{Set: true, Value: descPtr(strings.Repeat("a", staff.DescriptionMaxLength+1))})
	mustBeValidation(t, err)
}
