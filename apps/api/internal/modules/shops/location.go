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
}

// IsZero informa si el parche no pide ningún cambio.
func (p LocationPatch) IsZero() bool {
	return p.HasPhysicalLocation == nil && !p.AddressSet
}

// LocationWrite es lo que BrandRepository.Update escribe sobre la ubicación,
// ya normalizado y geocodificado por BrandService. Es un conjunto de cambios,
// no la ubicación completa: lo que no se pide no se toca.
type LocationWrite struct {
	// HasPhysicalLocation nil conserva el valor actual.
	HasPhysicalLocation *bool
	// ReplaceAddress true sobrescribe Address, Latitude y Longitude juntos
	// (los tres a nil si se quitó la dirección); false deja los tres como están.
	ReplaceAddress bool
	Address        *string
	Latitude       *float64
	Longitude      *float64
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
