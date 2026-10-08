package shops_test

import (
	"context"
	"errors"
	"testing"

	"system-barbershop/internal/modules/shops"
	"system-barbershop/internal/platform/apperr"
)

type fakePublicLinkRepository struct {
	slug      string
	found     bool
	err       error
	gotShopID string
}

func (f *fakePublicLinkRepository) Ensure(_ context.Context, barbershopID string) (string, bool, error) {
	f.gotShopID = barbershopID
	return f.slug, f.found, f.err
}

var _ shops.PublicLinkRepository = (*fakePublicLinkRepository)(nil)

func TestPublicLinkService_Get_ReturnsTheSlugOfTheSessionBarbershop(t *testing.T) {
	repo := &fakePublicLinkRepository{slug: "cortefinok7x2m9q4", found: true}
	got, err := shops.NewPublicLinkService(repo).Get(context.Background(), "shop-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Slug != "cortefinok7x2m9q4" || repo.gotShopID != "shop-1" {
		t.Fatalf("unexpected result %+v for shop %q", got, repo.gotShopID)
	}
}

func TestPublicLinkService_Get_NotFound_MapsToNotFound(t *testing.T) {
	_, err := shops.NewPublicLinkService(&fakePublicLinkRepository{}).Get(context.Background(), "shop-1")
	if appErr, ok := apperr.As(err); !ok || appErr.Kind != apperr.KindNotFound {
		t.Fatalf("expected a not-found apperr, got %v", err)
	}
}

func TestPublicLinkService_Get_RepositoryFailure_IsInternal(t *testing.T) {
	_, err := shops.NewPublicLinkService(&fakePublicLinkRepository{err: errors.New("boom")}).Get(context.Background(), "shop-1")
	if appErr, ok := apperr.As(err); !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected an internal apperr, got %v", err)
	}
}

func TestPublicLinkService_Get_CancelledContext_DoesNotReachTheRepository(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repo := &fakePublicLinkRepository{slug: "x-abcdef", found: true}
	if _, err := shops.NewPublicLinkService(repo).Get(ctx, "shop-1"); err == nil {
		t.Fatal("expected an error for a cancelled context")
	}
	if repo.gotShopID != "" {
		t.Fatal("the repository must not be called with a cancelled context")
	}
}
