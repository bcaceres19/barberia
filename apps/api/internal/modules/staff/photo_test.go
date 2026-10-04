package staff_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"

	"system-barbershop/internal/modules/staff"
)

// solidImage codifica una imagen lisa de w×h como JPEG o PNG. Las pruebas
// construyen bytes reales en vez de cabeceras a mano: ValidatePhoto decide el
// formato por image.DecodeConfig, así que el fixture debe ser una imagen de
// verdad.
func solidImage(t *testing.T, format string, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			img.Set(x, y, color.RGBA{R: 184, G: 149, B: 90, A: 255})
		}
	}
	var buf bytes.Buffer
	var err error
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80})
	case "png":
		err = png.Encode(&buf, img)
	default:
		t.Fatalf("formato de fixture desconocido: %s", format)
	}
	if err != nil {
		t.Fatalf("codificar fixture %s: %v", format, err)
	}
	return buf.Bytes()
}

func TestValidatePhoto_ValidJPEGAndPNG_Accepted(t *testing.T) {
	cases := []struct {
		name, format, declared string
	}{
		{"jpeg", "jpeg", staff.PhotoContentTypeJPEG},
		{"png", "png", staff.PhotoContentTypePNG},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := solidImage(t, tc.format, 512, 512)
			photo, err := staff.ValidatePhoto(tc.declared, data)
			if err != nil {
				t.Fatalf("expected a valid photo, got %v", err)
			}
			if photo.ContentType != tc.declared || !bytes.Equal(photo.Data, data) {
				t.Fatalf("unexpected photo: %+v", photo.ContentType)
			}
		})
	}
}

func TestValidatePhoto_Empty_Rejected(t *testing.T) {
	_, err := staff.ValidatePhoto(staff.PhotoContentTypeJPEG, nil)
	mustBeValidation(t, err)
}

func TestValidatePhoto_OverSizeLimit_Rejected(t *testing.T) {
	// Un JPEG válido seguido de relleno: excede MaxPhotoBytes sin necesitar
	// una imagen enorme, y el tamaño se comprueba ANTES de decodificar.
	data := append(solidImage(t, "jpeg", 128, 128), bytes.Repeat([]byte{0}, staff.MaxPhotoBytes)...)
	_, err := staff.ValidatePhoto(staff.PhotoContentTypeJPEG, data)
	mustBeValidation(t, err)
	if !strings.Contains(err.Error(), "512 KiB") {
		t.Fatalf("expected the size message, got %v", err)
	}
}

func TestValidatePhoto_NotAnImage_Rejected(t *testing.T) {
	// Un HTML que solo declara ser JPEG: el formato se decide por los bytes.
	_, err := staff.ValidatePhoto(staff.PhotoContentTypeJPEG, []byte("<html><script>alert(1)</script></html>"))
	mustBeValidation(t, err)
}

func TestValidatePhoto_DeclaredTypeMismatch_Rejected(t *testing.T) {
	data := solidImage(t, "png", 128, 128)
	_, err := staff.ValidatePhoto(staff.PhotoContentTypeJPEG, data)
	mustBeValidation(t, err)
}

func TestValidatePhoto_DimensionsOutOfRange_Rejected(t *testing.T) {
	cases := []struct {
		name string
		w, h int
	}{
		{"demasiado pequeña", staff.MinPhotoSide - 1, staff.MinPhotoSide - 1},
		{"un lado pequeño", 512, staff.MinPhotoSide - 1},
		{"demasiado grande", staff.MaxPhotoSide + 1, 128},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := staff.ValidatePhoto(staff.PhotoContentTypePNG, solidImage(t, "png", tc.w, tc.h))
			mustBeValidation(t, err)
		})
	}
}

func TestValidatePhoto_BoundaryDimensions_Accepted(t *testing.T) {
	for _, side := range []int{staff.MinPhotoSide, staff.MaxPhotoSide} {
		if _, err := staff.ValidatePhoto(staff.PhotoContentTypePNG, solidImage(t, "png", side, side)); err != nil {
			t.Fatalf("side %d should be accepted, got %v", side, err)
		}
	}
}
