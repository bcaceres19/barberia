package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"system-barbershop/internal/modules/staff"
	"system-barbershop/internal/modules/staff/httpapi"
	"testing"
)

func TestListBarbers_NumberedParametersAndWire(t *testing.T) {
	repo := &fakeRepository{listPageFn: func(_ context.Context, shop string, page, size int) (staff.PageResult, error) {
		if shop != testShopID || page != 2 || size != 3 {
			t.Fatalf("wrong request: %s %d %d", shop, page, size)
		}
		return staff.PageResult{Items: []staff.Barber{}, Page: 2, PageSize: 3, Total: 4, TotalPages: 2}, nil
	}}
	h := httpapi.NewListBarbersHandler(staff.NewService(repo))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, requestWithPrincipal(http.MethodGet, "/api/v1/private/barbers?page=2&pageSize=3", nil))
	if rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Page, PageSize, Total, TotalPages int
		NextCursor                        *string
		Items                             []httpapi.BarberResponse
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Page != 2 || body.PageSize != 3 || body.Total != 4 || body.TotalPages != 2 || body.NextCursor != nil || body.Items == nil {
		t.Fatalf("bad wire %+v", body)
	}
}
func TestListBarbers_NumberedRejectsInvalidOrMixedParameters(t *testing.T) {
	repo := &fakeRepository{listPageFn: func(context.Context, string, int, int) (staff.PageResult, error) {
		t.Fatal("invalid request reached repository")
		return staff.PageResult{}, nil
	}}
	h := httpapi.NewListBarbersHandler(staff.NewService(repo))
	for _, query := range []string{"page=0", "page=-1", "page=no", "page=", "pageSize=0", "pageSize=51", "pageSize=-1", "pageSize=no", "page=1&cursor=", "page=1&limit=2", "pageSize=3&cursor=x"} {
		t.Run(query, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, requestWithPrincipal(http.MethodGet, "/api/v1/private/barbers?"+query, nil))
			if rec.Code != 400 {
				t.Fatalf("expected 400, got %d", rec.Code)
			}
		})
	}
}
