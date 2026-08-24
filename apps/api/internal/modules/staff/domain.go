package staff

import (
	"encoding/base64"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"system-barbershop/internal/platform/apperr"
)

// FullNameMaxLength coincide con barber_full_name_ck
// (20260823130000_create_barber.sql): el servicio no inventa un límite más
// estricto que la base ya aplica (CA-021-03).
const FullNameMaxLength = 120

// DefaultListLimit, MinListLimit y MaxListLimit acotan `limit` en la lista
// paginada (CA-021-02): límites técnicos, no un máximo de negocio de
// barberos por barbería (docs/06-api/estandar-openapi.md §6.10).
const (
	DefaultListLimit = 20
	MinListLimit     = 1
	MaxListLimit     = 50
)

// Barber es un barbero de la barbería activa: exactamente los campos que
// HU-021 autoriza (DEC-047). Una barbería unipersonal y una de varios
// barberos usan este mismo tipo, sin una forma especial para "barbero
// único" (CA-021-01).
type Barber struct {
	ID        string
	FullName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// barberIDPattern es la misma forma canónica 8-4-4-4-12 que
// database.ValidBarbershopID exige (el núcleo no puede importar ese
// paquete, CA-002-06, así que se repite la validación aquí). Un
// identificador que no cumple esta forma no puede corresponder a ninguna
// fila real: se trata igual que "no existe" (mismo apperr.NotFound,
// CA-021-05) en vez de dejar que la consulta SQL falle con un error de
// tipo (que sería un 500, no un 404).
var barberIDPattern = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
)

// LooksLikeBarberID informa si id tiene la forma de un UUID válido.
func LooksLikeBarberID(id string) bool {
	return barberIDPattern.MatchString(id)
}

// NormalizeFullName recorta espacios, igual que barber_full_name_ck espera
// (btrim). No trunca ni rechaza: el largo y la vacuidad se validan aparte
// (Service.Create/Rename), y preserva cualquier carácter Unicode válido
// restante (CA-021-03: "preserva caracteres Unicode válidos").
func NormalizeFullName(raw string) string {
	return strings.TrimSpace(raw)
}

// Cursor es la posición decodificada de una página de la lista (CA-021-02):
// orden estable por fecha de alta y luego por identificador, exactamente
// los dos campos que idx_barber_shop_created_id indexa.
type Cursor struct {
	CreatedAt time.Time
	ID        string
}

// cursorWire es la forma serializada del cursor. Deliberadamente separada
// de Cursor: el cliente solo debe tratarlo como un valor opaco (nunca un
// contrato que decodifique él mismo), así que esta forma puede cambiar sin
// que eso sea un cambio incompatible del contrato HTTP (que solo promete
// "cadena opaca").
type cursorWire struct {
	CreatedAt time.Time `json:"createdAt"`
	ID        string    `json:"id"`
}

// EncodeCursor produce el valor opaco que la respuesta expone como
// `nextCursor`.
func EncodeCursor(c Cursor) string {
	raw, err := json.Marshal(cursorWire{CreatedAt: c.CreatedAt, ID: c.ID})
	if err != nil {
		// No debería ocurrir nunca (los campos son time.Time y string):
		// degradar a cursor vacío es preferible a propagar un panic desde
		// una función que el contrato documenta como infalible en la
		// práctica.
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

// DecodeCursor invierte EncodeCursor. Cualquier valor que no haya salido de
// EncodeCursor (manipulado a mano, truncado, de otra operación) se rechaza
// como entrada de cliente inválida, nunca como error interno: el cliente
// solo puede haberlo obtenido de una respuesta anterior o haberlo
// inventado, y ambos casos merecen la misma respuesta segura.
func DecodeCursor(raw string) (Cursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return Cursor{}, apperr.Invalid("el parámetro cursor tiene un formato inválido")
	}
	var w cursorWire
	if err := json.Unmarshal(decoded, &w); err != nil {
		return Cursor{}, apperr.Invalid("el parámetro cursor tiene un formato inválido")
	}
	if w.ID == "" || w.CreatedAt.IsZero() {
		return Cursor{}, apperr.Invalid("el parámetro cursor tiene un formato inválido")
	}
	return Cursor{CreatedAt: w.CreatedAt, ID: w.ID}, nil
}
