package publicbooking

import (
	"math"
	"strings"
	"unicode/utf8"

	"system-barbershop/internal/platform/apperr"
)

// HomeServiceMaxAddressLength es el largo máximo de la dirección a domicilio,
// igual que appointment_home_service_ck: el servicio no inventa un límite más
// laxo que la base.
const HomeServiceMaxAddressLength = 200

// HomeServiceInput es la dirección que el cliente eligió con el autocompletado
// de Google Places para un turno a domicilio (DEC-131, HU-103), tal como llegó
// en el cuerpo, todavía sin normalizar. Es dato personal (RN-DAT-02): ningún
// log, error ni mensaje la incluye.
type HomeServiceInput struct {
	Address   string
	Latitude  float64
	Longitude float64
}

// HomeService es la dirección ya normalizada y validada.
type HomeService struct {
	Address   string
	Latitude  float64
	Longitude float64
}

// NormalizeHomeService convierte la dirección recibida en la que se persiste:
// recorta y colapsa los espacios repetidos, exige una dirección no vacía de
// hasta HomeServiceMaxAddressLength caracteres y coordenadas finitas dentro de
// rango. Un punto (0, 0) con dirección se acepta: la selección del
// autocompletado es la fuente de verdad, no una heurística del servidor.
func NormalizeHomeService(in HomeServiceInput) (HomeService, error) {
	address := strings.Join(strings.Fields(in.Address), " ")
	if address == "" {
		return HomeService{}, apperr.Validation("la dirección a domicilio es obligatoria")
	}
	if utf8.RuneCountInString(address) > HomeServiceMaxAddressLength {
		return HomeService{}, apperr.Validation("la dirección a domicilio excede el largo máximo")
	}
	if !finiteWithin(in.Latitude, -90, 90) || !finiteWithin(in.Longitude, -180, 180) {
		return HomeService{}, apperr.Validation("las coordenadas de la dirección a domicilio no son válidas")
	}
	return HomeService{Address: address, Latitude: in.Latitude, Longitude: in.Longitude}, nil
}

func finiteWithin(v, lo, hi float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= lo && v <= hi
}

// errHomeServiceNotOffered cubre CA-103-05: el cliente pidió domicilio y la
// barbería lo tiene apagado en el momento de confirmar, ya sea porque nunca lo
// ofreció o porque lo apagó entre la elección y el envío. No crea el turno.
func errHomeServiceNotOffered() error {
	return apperr.Validation("esta barbería no ofrece servicio a domicilio")
}
