package staff

import (
	"bytes"
	"image"
	// Registran los decodificadores que image.DecodeConfig reconoce: sin estos
	// imports el formato real de una imagen nunca se identifica. Solo se lee la
	// cabecera (DecodeConfig), nunca se decodifican los píxeles.
	_ "image/jpeg"
	_ "image/png"
	"time"
)

// Límites de la fotografía de un barbero (DEC-104). MaxPhotoBytes coincide
// con barber_photo_image_size_ck (20260930120000_create_barber_photo.sql): el
// servicio no inventa un límite más laxo que la base ya aplica. Los tamaños
// son técnicos: el cliente reduce cada imagen a 512 px antes de enviarla, así
// que estos topes solo frenan un cliente que no lo haga.
const (
	MaxPhotoBytes = 512 * 1024
	// MinPhotoSide y MaxPhotoSide acotan ambos lados. MaxPhotoSide evita que
	// una cabecera PNG de miles de píxeles obligue a un consumidor futuro a
	// reservar memoria desproporcionada (una imagen pequeña en bytes puede
	// declarar dimensiones enormes).
	MinPhotoSide = 64
	MaxPhotoSide = 1024
)

// Tipos de contenido admitidos, en la misma forma que barber_photo_content_type_ck.
const (
	PhotoContentTypeJPEG = "image/jpeg"
	PhotoContentTypePNG  = "image/png"
)

// Photo es una fotografía ya validada, lista para persistirse. Data nunca se
// construye a partir de la entrada cruda sin pasar por ValidatePhoto.
type Photo struct {
	ContentType string
	Data        []byte
}

// StoredPhoto es la fotografía persistida de un barbero. UpdatedAt sirve de
// versión: la API lo expone como `photoUpdatedAt` y como ETag.
type StoredPhoto struct {
	ContentType string
	Data        []byte
	UpdatedAt   time.Time
}

// ValidatePhoto comprueba que data sea de verdad una imagen JPEG o PNG dentro
// de los límites y que coincida con el tipo declarado. El formato se decide por
// los bytes (image.DecodeConfig), nunca por la cabecera ni la extensión: un
// archivo que solo dice ser una imagen no llega a la base. declaredType es el
// Content-Type ya normalizado (sin parámetros) que envió el cliente.
func ValidatePhoto(declaredType string, data []byte) (Photo, error) {
	switch {
	case len(data) == 0:
		return Photo{}, errPhotoRequired()
	case len(data) > MaxPhotoBytes:
		return Photo{}, errPhotoTooLarge()
	}

	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return Photo{}, errPhotoFormat()
	}

	var contentType string
	switch format {
	case "jpeg":
		contentType = PhotoContentTypeJPEG
	case "png":
		contentType = PhotoContentTypePNG
	default:
		return Photo{}, errPhotoFormat()
	}
	if declaredType != contentType {
		return Photo{}, errPhotoFormat()
	}

	if cfg.Width < MinPhotoSide || cfg.Height < MinPhotoSide ||
		cfg.Width > MaxPhotoSide || cfg.Height > MaxPhotoSide {
		return Photo{}, errPhotoDimensions()
	}

	return Photo{ContentType: contentType, Data: data}, nil
}
