package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"system-barbershop/internal/modules/staff"
	"system-barbershop/internal/modules/staff/httpapi"
)

const photoBarberID = "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"

// jpegFixture devuelve un JPEG real de 128×128: el servicio decide el formato
// por los bytes, así que el doble no puede ser un texto cualquiera.
func jpegFixture(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 128, 128))
	for x := 0; x < 128; x++ {
		for y := 0; y < 128; y++ {
			img.Set(x, y, color.RGBA{R: 16, G: 27, B: 43, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	return buf.Bytes()
}

func putPhotoRequest(contentType string, body []byte) *http.Request {
	req := requestWithBarberID(http.MethodPut, "/api/v1/private/barbers/"+photoBarberID+"/photo", body, photoBarberID)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return req
}

func TestPutBarberPhotoHandler_Success_Returns200WithPhotoUpdatedAt(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	data := jpegFixture(t)
	repo := &fakeRepository{putPhotoFn: func(shop, id string, photo staff.Photo) (staff.PhotoResult, error) {
		if shop != testShopID || id != photoBarberID {
			t.Fatalf("unexpected scope %s/%s", shop, id)
		}
		if photo.ContentType != "image/jpeg" || !bytes.Equal(photo.Data, data) {
			t.Fatalf("unexpected photo forwarded")
		}
		return staff.PhotoResult{Found: true, Barber: staff.Barber{
			ID: id, FullName: "Carlos", CreatedAt: now, UpdatedAt: now, PhotoUpdatedAt: &now,
		}}, nil
	}}
	h := httpapi.NewPutBarberPhotoHandler(staff.NewService(repo))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, putPhotoRequest("image/jpeg", data))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body httpapi.BarberResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.PhotoUpdatedAt == nil || !body.PhotoUpdatedAt.Equal(now) {
		t.Fatalf("expected photoUpdatedAt %v, got %v", now, body.PhotoUpdatedAt)
	}
}

func TestPutBarberPhotoHandler_UnsupportedOrMissingContentType_Returns422WithoutCallingRepository(t *testing.T) {
	repo := &fakeRepository{putPhotoFn: func(string, string, staff.Photo) (staff.PhotoResult, error) {
		t.Fatal("repository must not be called")
		return staff.PhotoResult{}, nil
	}}
	h := httpapi.NewPutBarberPhotoHandler(staff.NewService(repo))

	for _, contentType := range []string{"", "application/json", "image/gif", "text/plain"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, putPhotoRequest(contentType, jpegFixture(t)))
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("content-type %q: expected 422, got %d", contentType, rec.Code)
		}
	}
}

func TestPutBarberPhotoHandler_ParametersInContentType_AreTolerated(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeRepository{putPhotoFn: func(_, id string, _ staff.Photo) (staff.PhotoResult, error) {
		return staff.PhotoResult{Found: true, Barber: staff.Barber{ID: id, PhotoUpdatedAt: &now}}, nil
	}}
	h := httpapi.NewPutBarberPhotoHandler(staff.NewService(repo))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, putPhotoRequest("image/jpeg; charset=binary", jpegFixture(t)))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPutBarberPhotoHandler_BodyThatIsNotAnImage_Returns422(t *testing.T) {
	repo := &fakeRepository{putPhotoFn: func(string, string, staff.Photo) (staff.PhotoResult, error) {
		t.Fatal("repository must not be called")
		return staff.PhotoResult{}, nil
	}}
	h := httpapi.NewPutBarberPhotoHandler(staff.NewService(repo))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, putPhotoRequest("image/jpeg", []byte("<html></html>")))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
	if problem := decodeProblem(t, rec); problem["code"] != "validation-error" {
		t.Fatalf("unexpected problem: %v", problem)
	}
}

func TestPutBarberPhotoHandler_BarberMissing_Returns404(t *testing.T) {
	repo := &fakeRepository{putPhotoFn: func(string, string, staff.Photo) (staff.PhotoResult, error) {
		return staff.PhotoResult{Found: false}, nil
	}}
	h := httpapi.NewPutBarberPhotoHandler(staff.NewService(repo))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, putPhotoRequest("image/jpeg", jpegFixture(t)))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func getPhotoRequest(ifNoneMatch string) *http.Request {
	req := requestWithBarberID(http.MethodGet, "/api/v1/private/barbers/"+photoBarberID+"/photo", nil, photoBarberID)
	if ifNoneMatch != "" {
		req.Header.Set("If-None-Match", ifNoneMatch)
	}
	return req
}

func TestGetBarberPhotoHandler_Found_ServesBytesWithETagAndPrivateCache(t *testing.T) {
	data := jpegFixture(t)
	updated := time.Unix(1790769600, 0).UTC()
	repo := &fakeRepository{getPhotoFn: func(shop, id string) (staff.StoredPhoto, bool, error) {
		if shop != testShopID || id != photoBarberID {
			t.Fatalf("unexpected scope %s/%s", shop, id)
		}
		return staff.StoredPhoto{ContentType: "image/jpeg", Data: data, UpdatedAt: updated}, true, nil
	}}
	h := httpapi.NewGetBarberPhotoHandler(staff.NewService(repo))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, getPhotoRequest(""))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "image/jpeg" {
		t.Fatalf("expected image/jpeg, got %q", got)
	}
	if got := rec.Header().Get("Cache-Control"); got != "private, max-age=3600" {
		t.Fatalf("a portrait must never be cacheable by a shared cache, got %q", got)
	}
	if got := rec.Header().Get("Content-Length"); got != strconv.Itoa(len(data)) {
		t.Fatalf("unexpected content-length %q", got)
	}
	wantETag := `"` + strconv.FormatInt(updated.UnixMicro(), 10) + `"`
	if got := rec.Header().Get("ETag"); got != wantETag {
		t.Fatalf("expected ETag %s, got %s", wantETag, got)
	}
	if !bytes.Equal(rec.Body.Bytes(), data) {
		t.Fatal("the body must be the stored bytes")
	}
}

func TestGetBarberPhotoHandler_MatchingIfNoneMatch_Returns304WithoutBody(t *testing.T) {
	updated := time.Unix(1790769600, 0).UTC()
	repo := &fakeRepository{getPhotoFn: func(string, string) (staff.StoredPhoto, bool, error) {
		return staff.StoredPhoto{ContentType: "image/jpeg", Data: jpegFixture(t), UpdatedAt: updated}, true, nil
	}}
	h := httpapi.NewGetBarberPhotoHandler(staff.NewService(repo))

	etag := `"` + strconv.FormatInt(updated.UnixMicro(), 10) + `"`
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, getPhotoRequest(etag))

	if rec.Code != http.StatusNotModified {
		t.Fatalf("expected 304, got %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatal("a 304 must carry no body")
	}
}

func TestGetBarberPhotoHandler_NotFound_Returns404(t *testing.T) {
	repo := &fakeRepository{getPhotoFn: func(string, string) (staff.StoredPhoto, bool, error) {
		return staff.StoredPhoto{}, false, nil
	}}
	h := httpapi.NewGetBarberPhotoHandler(staff.NewService(repo))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, getPhotoRequest(""))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func deletePhotoRequest() *http.Request {
	return requestWithBarberID(http.MethodDelete, "/api/v1/private/barbers/"+photoBarberID+"/photo", nil, photoBarberID)
}

func TestDeleteBarberPhotoHandler_Returns204AndNotFoundForMissingBarber(t *testing.T) {
	exists := true
	repo := &fakeRepository{deletePhotoFn: func(shop, id string) (bool, error) {
		if shop != testShopID || id != photoBarberID {
			t.Fatalf("unexpected scope %s/%s", shop, id)
		}
		return exists, nil
	}}
	h := httpapi.NewDeleteBarberPhotoHandler(staff.NewService(repo))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, deletePhotoRequest())
	if rec.Code != http.StatusNoContent || rec.Body.Len() != 0 {
		t.Fatalf("expected an empty 204, got %d (%q)", rec.Code, rec.Body.String())
	}

	exists = false
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, deletePhotoRequest())
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for a missing barber, got %d", rec.Code)
	}
}

func TestPhotoHandlers_MissingPrincipal_Returns500Safely(t *testing.T) {
	svc := staff.NewService(&fakeRepository{})
	handlers := map[string]http.Handler{
		"put":    httpapi.NewPutBarberPhotoHandler(svc),
		"get":    httpapi.NewGetBarberPhotoHandler(svc),
		"delete": httpapi.NewDeleteBarberPhotoHandler(svc),
	}
	for name, h := range handlers {
		req := httptest.NewRequest(http.MethodGet, "/x", nil).WithContext(context.Background())
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("%s: expected 500, got %d", name, rec.Code)
		}
	}
}
