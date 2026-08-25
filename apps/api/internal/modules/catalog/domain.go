package catalog

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"system-barbershop/internal/platform/apperr"
)

// NameMaxLength coincide con service_name_ck
// (database/migrations/20260824000000_create_service.sql): el servicio no
// inventa un límite más estricto que la base ya aplica (CA-022-04).
const NameMaxLength = 120

// DescriptionMaxLength coincide con service_description_ck.
const DescriptionMaxLength = 500

// MinDurationMinutes y MaxDurationMinutes acotan duration_minutes
// (CA-022-03): un límite TÉCNICO que coincide con service_duration_minutes_ck,
// no una lista cerrada de valores de negocio (RN-SER-02). 25, 30, 45 y 90
// minutos son igual de válidos que cualquier otro entero en este rango.
const (
	MinDurationMinutes = 1
	MaxDurationMinutes = 1440
)

// CurrencyCOP es la única moneda que HU-022 admite (DEC-067): fija, sin
// campo editable. Ninguna solicitud de cliente puede cambiar este valor;
// el contrato ni siquiera declara un campo de moneda de entrada.
const CurrencyCOP = "COP"

// DefaultListLimit, MinListLimit y MaxListLimit acotan `limit` en la lista
// paginada (CA-022-01), mismo criterio técnico que staff.DefaultListLimit
// (docs/06-api/estandar-openapi.md §6.10): la colección no tiene un máximo
// de negocio de servicios por barbería.
const (
	DefaultListLimit = 20
	MinListLimit     = 1
	MaxListLimit     = 50
)

// Service es un servicio del catálogo de la barbería activa. PriceCents
// representa el precio en centavos (entero exacto, nunca coma flotante,
// docs/03-desarrollo/estandar-backend-go.md, docs/05-backend/
// estandar-base-datos.md §5); Currency es siempre CurrencyCOP (DEC-067).
// Description es nil cuando el servicio no tiene descripción. IsActive y
// DeactivatedAt reflejan el ciclo de vida de HU-024 (RN-SER-03):
// DeactivatedAt es nil mientras IsActive es true, y nunca nil cuando es
// false (mismo invariante que service_deactivated_at_ck).
type Service struct {
	ID              string
	Name            string
	Description     *string
	DurationMinutes int
	PriceCents      int64
	Currency        string
	IsActive        bool
	DeactivatedAt   *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// DeactivationImpact es el resultado de previsualizar o confirmar el efecto
// de desactivar un servicio (CA-024-01, CA-024-04). AffectedAppointments es
// siempre 0 en B1 (DEC-069): `appointment` no existe todavía en la cadena
// migrada, así que cero citas futuras existen para ningún servicio -es el
// conteo real del sistema actual, no un valor por defecto ni un dato
// simulado. B3 sustituye el cálculo por una consulta real contra
// `appointment` sin cambiar esta forma.
type DeactivationImpact struct {
	AffectedAppointments int
}

// currentDeactivationImpact es el único punto donde B1 decide
// AffectedAppointments (DEC-069). No es un puerto ni un adaptador que
// simule consultar una tabla: es la verdad honesta y literal del sistema
// actual, porque appointment no existe todavía en ningún ambiente
// migrado -no puede haber una sola cita futura, para ningún servicio, de
// ninguna barbería. Cuando B3 cree appointment, esta función (y solo esta)
// se reemplaza por una consulta real al repositorio de esa capacidad, sin
// tocar el resto de esta historia.
func currentDeactivationImpact() DeactivationImpact {
	return DeactivationImpact{AffectedAppointments: 0}
}

// serviceIDPattern es la misma forma canónica 8-4-4-4-12 que
// database.ValidBarbershopID exige (el núcleo no puede importar ese
// paquete, CA-002-06, así que se repite aquí, mismo criterio que
// staff.LooksLikeBarberID). Un identificador que no cumple esta forma no
// puede corresponder a ninguna fila real: se trata igual que "no existe"
// (mismo apperr.NotFound) en vez de dejar que la consulta SQL falle con un
// error de tipo (500, no 404).
var serviceIDPattern = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
)

// LooksLikeServiceID informa si id tiene la forma de un UUID válido.
func LooksLikeServiceID(id string) bool {
	return serviceIDPattern.MatchString(id)
}

// NormalizeName recorta espacios, igual que service_name_ck espera (btrim).
// No trunca ni rechaza: el largo y la vacuidad se validan aparte
// (Service.Create/Update), y preserva cualquier carácter Unicode válido
// restante.
func NormalizeName(raw string) string {
	return strings.TrimSpace(raw)
}

// NormalizeDescription recorta espacios y devuelve nil cuando el resultado
// queda vacío: una descripción compuesta solo por espacios se trata igual
// que una ausente, nunca como cadena vacía persistida.
func NormalizeDescription(raw string) *string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// pricePattern exige un decimal estrictamente positivo con como máximo dos
// dígitos de fracción y como máximo diez dígitos enteros (coherente con
// numeric(12,2), docs/06-api/estandar-openapi.md §6: "importe exacto:
// string decimal más código de moneda"). Sin signo: un precio nunca es
// negativo ni lleva '+' explícito.
var pricePattern = regexp.MustCompile(`^[0-9]{1,10}(\.[0-9]{1,2})?$`)

// ParsePriceCOP interpreta raw (el valor crudo de `price` en el cuerpo)
// como un número exacto de centavos de peso colombiano, sin pasar nunca por
// coma flotante. Rechaza formato inválido (letras, signo, más de dos
// decimales, notación científica) y valores menores o iguales a cero
// (DEC-067: "no se permiten servicios gratuitos").
func ParsePriceCOP(raw string) (int64, error) {
	trimmed := strings.TrimSpace(raw)
	if !pricePattern.MatchString(trimmed) {
		return 0, errPriceInvalidFormat()
	}

	parts := strings.SplitN(trimmed, ".", 2)
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, errPriceInvalidFormat()
	}

	frac := int64(0)
	if len(parts) == 2 {
		fracDigits := parts[1]
		if len(fracDigits) == 1 {
			fracDigits += "0"
		}
		frac, err = strconv.ParseInt(fracDigits, 10, 64)
		if err != nil {
			return 0, errPriceInvalidFormat()
		}
	}

	cents := whole*100 + frac
	if cents <= 0 {
		return 0, errPriceMustBePositive()
	}
	return cents, nil
}

// FormatPriceCOP formatea cents (siempre > 0 en un Service válido) como el
// string decimal de dos dígitos que el contrato expone en `price` (mismo
// formato exacto que ParsePriceCOP acepta de vuelta, sin pérdida).
func FormatPriceCOP(cents int64) string {
	whole := cents / 100
	frac := cents % 100
	return fmt.Sprintf("%d.%02d", whole, frac)
}

// Cursor es la posición decodificada de una página de la lista (CA-022-01):
// orden estable por fecha de alta y luego por identificador, mismo patrón
// que staff.Cursor (duplicación pequeña y clara preferible a una
// abstracción cruzada entre módulos, docs/03-desarrollo/
// estandar-backend-go.md §5.23).
type Cursor struct {
	CreatedAt time.Time
	ID        string
}

// cursorWire es la forma serializada del cursor: el cliente solo debe
// tratarlo como un valor opaco.
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
// como entrada de cliente inválida, nunca como error interno.
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
