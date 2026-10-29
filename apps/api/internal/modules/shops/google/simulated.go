package google

import (
	"context"
	"hash/fnv"
	"strings"

	"system-barbershop/internal/modules/shops"
)

// Simulated es un geocodificador determinista SOLO para desarrollo local y
// pruebas, cuando no hay clave de Google. No consulta ninguna red: ubica
// cualquier dirección cerca del centro de Bogotá con un desplazamiento
// derivado de su texto (la misma dirección da siempre el mismo punto). Así la
// pantalla de Configuración y la reserva pública se pueden ver y probar sin
// credenciales. main.go lo instala únicamente en los ambientes local y test.
//
// Rechaza una dirección sin ningún dígito ni palabra (menos de 5 caracteres
// útiles) como ErrAddressUnresolvable, para poder ejercitar el camino de 422.
type Simulated struct{}

var _ shops.Geocoder = Simulated{}

// Geocode implementa shops.Geocoder.Geocode.
func (Simulated) Geocode(_ context.Context, address string) (shops.Coordinates, error) {
	trimmed := strings.TrimSpace(address)
	if len([]rune(trimmed)) < 5 {
		return shops.Coordinates{}, shops.ErrAddressUnresolvable
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(strings.ToLower(trimmed)))
	sum := h.Sum32()
	// ±0,02° (~2 km) alrededor de 4,6097 N, 74,0817 O.
	dLat := (float64(sum%2001) - 1000) / 50000
	dLng := (float64((sum/2001)%2001) - 1000) / 50000
	return shops.Coordinates{Latitude: 4.6097 + dLat, Longitude: -74.0817 + dLng}, nil
}
