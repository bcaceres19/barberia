package shops

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

// MaxAddressLength es el largo máximo de la dirección, igual que
// barbershop_address_ck: el servicio no inventa un límite más laxo que la base.
const MaxAddressLength = 200

// Location es la ubicación física opcional de la barbería (DEC-130, HU-102).
// HasPhysicalLocation es un ajuste explícito: nunca se deriva de si Address
// está vacía, y apagarlo conserva Address y las coordenadas. Latitude y
// Longitude son nil o ambas no nil, y solo existen junto a una Address
// (barbershop_coordinates_ck): son el resultado de geocodificar una vez en el
// servidor al guardar, no un dato que el cliente envíe.
type Location struct {
	HasPhysicalLocation bool
	Address             *string
	Latitude            *float64
	Longitude           *float64
	// OffersHomeService es el interruptor general del servicio a domicilio
	// (DEC-131, HU-103): independiente de la ubicación del local. Apagarlo no
	// toca los turnos ya creados con domicilio.
	OffersHomeService bool
	// GooglePlaceID es el Place ID de la ficha de Google de la barbería
	// (DEC-129, HU-101): nil = sin ficha, y la reserva pública no muestra reseñas
	// de Google. Lo escribe el panel; la calificación y las reseñas nunca se
	// guardan (viven en la caché de memoria del adaptador).
	GooglePlaceID *string
}

// Visible informa si la reserva pública debe mostrar el mapa: el interruptor
// activo y coordenadas guardadas (CA-102-02, CA-102-03).
func (l Location) Visible() bool {
	return l.HasPhysicalLocation && l.Address != nil && l.Latitude != nil && l.Longitude != nil
}

// LocationPatch es el cambio pedido sobre la ubicación en una actualización de
// la marca. Todo es opcional: HasPhysicalLocation nil y AddressSet false
// significan "no tocar". AddressSet true con Address nil o vacía quita la
// dirección y sus coordenadas; con una dirección nueva la geocodifica.
type LocationPatch struct {
	HasPhysicalLocation *bool
	AddressSet          bool
	Address             *string
	// OffersHomeService nil conserva el interruptor de domicilio (DEC-131).
	OffersHomeService *bool
	// GooglePlaceIDSet true con GooglePlaceID nil o vacío quita la ficha de
	// Google; con un valor lo guarda (DEC-129). False conserva la actual.
	GooglePlaceIDSet bool
	GooglePlaceID    *string
}

// IsZero informa si el parche no pide ningún cambio.
func (p LocationPatch) IsZero() bool {
	return p.HasPhysicalLocation == nil && !p.AddressSet && p.OffersHomeService == nil && !p.GooglePlaceIDSet
}

// LocationWrite es lo que BrandRepository.Update escribe sobre la ubicación,
// ya normalizado y geocodificado por BrandService. Es un conjunto de cambios,
// no la ubicación completa: lo que no se pide no se toca.
type LocationWrite struct {
	// HasPhysicalLocation nil conserva el valor actual.
	HasPhysicalLocation *bool
	// OffersHomeService nil conserva el interruptor de domicilio (DEC-131).
	OffersHomeService *bool
	// ReplaceAddress true sobrescribe Address, Latitude y Longitude juntos
	// (los tres a nil si se quitó la dirección); false deja los tres como están.
	ReplaceAddress bool
	Address        *string
	Latitude       *float64
	Longitude      *float64
	// ReplaceGooglePlaceID true sobrescribe la ficha de Google con
	// GooglePlaceID (nil la quita); false la deja como está (DEC-129).
	ReplaceGooglePlaceID bool
	GooglePlaceID        *string
}

// Coordinates es un punto geográfico en grados decimales.
type Coordinates struct {
	Latitude  float64
	Longitude float64
}

// Geocoder es el puerto que resuelve una dirección en texto a coordenadas
// (DEC-130). El núcleo no importa ningún cliente de Google: lo implementa
// shops/google y las pruebas usan un doble. Se llama UNA vez al guardar una
// dirección nueva, nunca al leer.
type Geocoder interface {
	// Geocode devuelve las coordenadas de address. ErrAddressUnresolvable si el
	// proveedor no la encuentra o no es lo bastante precisa;
	// ErrGeocoderUnavailable si no pudo consultarse (red, cuota, credencial).
	Geocode(ctx context.Context, address string) (Coordinates, error)
}

// GeocoderFunc adapta una función al puerto Geocoder.
type GeocoderFunc func(ctx context.Context, address string) (Coordinates, error)

// Geocode implementa Geocoder.
func (f GeocoderFunc) Geocode(ctx context.Context, address string) (Coordinates, error) {
	return f(ctx, address)
}

var (
	// ErrAddressUnresolvable indica que la dirección no pudo resolverse a un
	// punto fiable (no existe, es ambigua o solo se ubica a nivel de ciudad).
	ErrAddressUnresolvable = errors.New("shops: la dirección no se pudo geocodificar")
	// ErrGeocoderUnavailable indica que la consulta al proveedor falló por una
	// causa ajena a la dirección (sin credencial, cuota agotada, red).
	ErrGeocoderUnavailable = errors.New("shops: el geocodificador no está disponible")
)

// NormalizeAddress recorta y colapsa los espacios repetidos de la dirección.
// Devuelve nil si queda vacía: sin dirección es nil, nunca una cadena vacía.
func NormalizeAddress(raw string) *string {
	normalized := strings.Join(strings.Fields(raw), " ")
	if normalized == "" {
		return nil
	}
	return &normalized
}

// IsValidAddressLength informa si address, ya normalizada, cabe en la base.
func IsValidAddressLength(address string) bool {
	return utf8.RuneCountInString(address) <= MaxAddressLength
}

// MaxGooglePlaceIDLength es el largo máximo del Place ID, igual que
// barbershop_google_place_id_ck.
const MaxGooglePlaceIDLength = 200

// NormalizeGooglePlaceID recorta el Place ID y valida su forma (letras, dígitos,
// `_` y `-`, de 1 a 200 caracteres), la misma que exige la base: el valor viaja
// luego como segmento de una URL de Google, así que nunca debe admitir `/`, `?`
// ni espacios. Devuelve (nil, true) si queda vacío (quitar la ficha) y
// (nil, false) si la forma no es válida.
func NormalizeGooglePlaceID(raw string) (*string, bool) {
	id := strings.TrimSpace(raw)
	if id == "" {
		return nil, true
	}
	if len(id) > MaxGooglePlaceIDLength {
		return nil, false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-':
		default:
			return nil, false
		}
	}
	return &id, true
}
