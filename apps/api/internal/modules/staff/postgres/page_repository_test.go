package postgres_test

import (
	"context"
	"testing"
)

func TestListPage_TotalsBoundsAndTenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	ctx := context.Background()
	ours := createBarber(t, repo, shopC, "Página C "+uniqueSuffix(t))
	other := createBarber(t, repo, shopD, "Página D "+uniqueSuffix(t))
	first, err := repo.ListPage(ctx, string(shopC), 1, 3)
	if err != nil {
		t.Fatal(err)
	}
	if first.Page != 1 || first.PageSize != 3 || len(first.Items) != 3 || first.TotalPages != (first.Total+2)/3 {
		t.Fatalf("bad page %+v", first)
	}
	seen := map[string]bool{}
	for p := 1; p <= first.TotalPages; p++ {
		got, err := repo.ListPage(ctx, string(shopC), p, 3)
		if err != nil {
			t.Fatal(err)
		}
		if got.Total != first.Total {
			t.Fatal("total changed without a write")
		}
		for _, b := range got.Items {
			if b.ID == other.ID || seen[b.ID] {
				t.Fatal("tenant leak or duplicate")
			}
			seen[b.ID] = true
		}
	}
	if len(seen) != first.Total || !seen[ours.ID] {
		t.Fatal("missing rows")
	}
	last, err := repo.ListPage(ctx, string(shopC), 999999, 3)
	if err != nil {
		t.Fatal(err)
	}
	if last.Page != first.TotalPages || len(last.Items) == 0 {
		t.Fatal("out of range page must clamp")
	}
	empty, err := repo.ListPage(ctx, "28800000-0000-4000-8000-000000000099", 1, 3)
	if err != nil {
		t.Fatal(err)
	}
	if empty.Total != 0 || empty.Page != 1 || empty.TotalPages != 1 || len(empty.Items) != 0 || empty.Items == nil {
		t.Fatalf("bad empty %+v", empty)
	}
	theirs, err := repo.ListPage(ctx, string(shopD), 1, 50)
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range theirs.Items {
		if b.ID == ours.ID {
			t.Fatal("reverse tenant leak")
		}
	}
}
