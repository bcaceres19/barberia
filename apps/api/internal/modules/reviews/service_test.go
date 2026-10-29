package reviews_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"system-barbershop/internal/modules/reviews"
	"system-barbershop/internal/platform/apperr"
)

type fakeRepo struct {
	createHash    string
	createRating  int
	createComment *string
	createCalls   int
	outcome       reviews.CreateOutcome
	err           error

	setHiddenCalls int
	setHiddenFound bool
}

func (f *fakeRepo) CreateByTokenHash(_ context.Context, h string, rating int, c *string) (reviews.Review, reviews.CreateOutcome, error) {
	f.createCalls++
	f.createHash, f.createRating, f.createComment = h, rating, c
	return reviews.Review{ID: "r1", Rating: rating, Comment: c, CreatedAt: time.Unix(0, 0)}, f.outcome, f.err
}
func (f *fakeRepo) ListPublic(context.Context, string, string, int) (reviews.PublicList, bool, error) {
	return reviews.PublicList{}, true, f.err
}
func (f *fakeRepo) ListForBarber(context.Context, string, string, int) (reviews.PanelList, bool, error) {
	return reviews.PanelList{}, true, f.err
}
func (f *fakeRepo) SetHidden(context.Context, string, string, string, bool) (reviews.Review, bool, error) {
	f.setHiddenCalls++
	return reviews.Review{}, f.setHiddenFound, f.err
}

func kindOf(err error) apperr.Kind {
	var e *apperr.Error
	if errors.As(err, &e) {
		return e.Kind
	}
	return ""
}

func ptr(s string) *string { return &s }

func TestCreate_InvalidRating_IsValidationAndNeverTouchesTheRepository(t *testing.T) {
	for _, rating := range []int{0, 6, -1, 100} {
		repo := &fakeRepo{}
		_, err := reviews.NewService(repo).Create(context.Background(), "token", rating, nil)
		if kindOf(err) != apperr.KindValidation {
			t.Fatalf("rating %d: kind = %q, want validation", rating, kindOf(err))
		}
		if repo.createCalls != 0 {
			t.Fatalf("rating %d: el repositorio no debía llamarse", rating)
		}
	}
}

func TestCreate_CommentTooLong_IsValidation(t *testing.T) {
	repo := &fakeRepo{}
	_, err := reviews.NewService(repo).Create(context.Background(), "token", 5, ptr(strings.Repeat("ñ", 501)))
	if kindOf(err) != apperr.KindValidation || repo.createCalls != 0 {
		t.Fatalf("kind = %q calls = %d", kindOf(err), repo.createCalls)
	}
	// 500 caracteres exactos (multibyte) son válidos: se cuentan caracteres, no bytes.
	if _, err := reviews.NewService(repo).Create(context.Background(), "token", 5, ptr(strings.Repeat("ñ", 500))); err != nil {
		t.Fatalf("500 caracteres deben aceptarse: %v", err)
	}
}

func TestCreate_NormalizesTheComment(t *testing.T) {
	cases := []struct {
		name string
		in   *string
		want *string
	}{
		{"nil", nil, nil},
		{"vacío", ptr(""), nil},
		{"solo espacios", ptr("   \t\n "), nil},
		{"recorta extremos", ptr("  Muy bien \n"), ptr("Muy bien")},
		{"conserva el interior", ptr("a  b"), ptr("a  b")},
	}
	for _, c := range cases {
		repo := &fakeRepo{outcome: reviews.Created}
		if _, err := reviews.NewService(repo).Create(context.Background(), "token", 4, c.in); err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		switch {
		case c.want == nil && repo.createComment != nil:
			t.Fatalf("%s: comment = %q, want nil", c.name, *repo.createComment)
		case c.want != nil && (repo.createComment == nil || *repo.createComment != *c.want):
			t.Fatalf("%s: comment = %v, want %q", c.name, repo.createComment, *c.want)
		}
	}
}

func TestCreate_TokenIsHashedAndNeverReachesTheRepositoryInClear(t *testing.T) {
	repo := &fakeRepo{outcome: reviews.Created}
	if _, err := reviews.NewService(repo).Create(context.Background(), " token-en-claro ", 5, nil); err != nil {
		t.Fatal(err)
	}
	if repo.createHash == "" || strings.Contains(repo.createHash, "token-en-claro") || len(repo.createHash) != 64 {
		t.Fatalf("hash = %q, se esperaba un SHA-256 hexadecimal", repo.createHash)
	}
}

func TestCreate_UniformErrors(t *testing.T) {
	// Forma inválida del token: 404 sin tocar el repositorio.
	for _, token := range []string{"", "   ", strings.Repeat("a", reviews.MaxTokenLength+1)} {
		repo := &fakeRepo{}
		_, err := reviews.NewService(repo).Create(context.Background(), token, 5, nil)
		if kindOf(err) != apperr.KindNotFound || repo.createCalls != 0 {
			t.Fatalf("token %q: kind = %q calls = %d", token, kindOf(err), repo.createCalls)
		}
	}
	repo := &fakeRepo{outcome: reviews.TokenNotFound}
	if _, err := reviews.NewService(repo).Create(context.Background(), "x", 5, nil); kindOf(err) != apperr.KindNotFound {
		t.Fatalf("token desconocido: kind = %q", kindOf(err))
	}
	repo = &fakeRepo{outcome: reviews.NotAvailable}
	if _, err := reviews.NewService(repo).Create(context.Background(), "x", 5, nil); kindOf(err) != apperr.KindConflict {
		t.Fatalf("turno no calificable: kind = %q", kindOf(err))
	}
	repo = &fakeRepo{err: errors.New("boom")}
	if _, err := reviews.NewService(repo).Create(context.Background(), "x", 5, nil); kindOf(err) != apperr.KindInternal {
		t.Fatalf("fallo del repositorio: kind = %q", kindOf(err))
	}
}

func TestListsAndSetHidden_MalformedIDs_AreNotFoundWithoutTouchingTheRepository(t *testing.T) {
	repo := &fakeRepo{setHiddenFound: true}
	svc := reviews.NewService(repo)
	if _, err := svc.ListPublic(context.Background(), "slug", "no-es-uuid"); kindOf(err) != apperr.KindNotFound {
		t.Fatalf("ListPublic: %v", err)
	}
	if _, err := svc.ListForBarber(context.Background(), "shop", "no-es-uuid"); kindOf(err) != apperr.KindNotFound {
		t.Fatalf("ListForBarber: %v", err)
	}
	if _, err := svc.SetHidden(context.Background(), "shop", "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4", "nope", true); kindOf(err) != apperr.KindNotFound {
		t.Fatalf("SetHidden: %v", err)
	}
	if repo.setHiddenCalls != 0 {
		t.Fatal("un identificador mal formado no debe llegar al repositorio")
	}
}

func TestSetHidden_NotFoundFromRepository(t *testing.T) {
	repo := &fakeRepo{setHiddenFound: false}
	_, err := reviews.NewService(repo).SetHidden(context.Background(), "shop",
		"8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4", "1b2c3d4e-5f60-4718-8293-a4b5c6d7e8f9", true)
	if kindOf(err) != apperr.KindNotFound {
		t.Fatalf("kind = %q", kindOf(err))
	}
}
