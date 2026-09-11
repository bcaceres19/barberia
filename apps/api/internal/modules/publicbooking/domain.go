package publicbooking

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"system-barbershop/internal/platform/apperr"
)

// BarbershopProfile es la representación pública mínima de una barbería
// habilitada (HU-090, CA-090-01, CA-090-04): nombre, zona horaria IANA y
// contacto público opcional. Nunca incluye un identificador interno, el
// slug mismo, ni ningún otro dato de configuración privada.
type BarbershopProfile struct {
	Name         string
	Timezone     string
	ContactEmail *string
	ContactPhone *string
}

// MaxSlugLength acota la longitud aceptada del identificador antes de
// consultar la base de datos: cualquier valor más largo se rechaza de
// inmediato con la misma respuesta uniforme que un slug desconocido
// (CA-090-02), sin gastar una consulta. Igual a shops.SlugMaxLength
// (barbershop_public_slug_ck): ningún slug generado supera ese largo, así
// que uno más largo nunca puede coincidir.
const MaxSlugLength = 40

// PublicService es la proyección pública mínima de un servicio ofrecido por
// la barbería resuelta (HU-091, CA-091-01): nombre, descripción, duración,
// precio y moneda. Nunca isActive, deactivatedAt, createdAt, updatedAt ni
// ningún otro dato de auditoría o ciclo de vida -esos campos son alcance
// exclusivo de catalog (HU-022/HU-024); el núcleo de publicbooking no
// importa ese paquete (CA-002-06), así que repite su propia proyección
// mínima en vez de reutilizar catalog.Service.
type PublicService struct {
	ID              string
	Name            string
	Description     *string
	DurationMinutes int
	PriceCents      int64
	Currency        string
}

// FormatServicePriceCOP formatea cents como el string decimal de dos
// dígitos que el contrato público expone en `price`. Duplica
// catalog.FormatPriceCOP a propósito (mismo criterio de CA-002-06 que
// centsFromNumeric en publicbooking/postgres): el núcleo de publicbooking
// no importa catalog.
func FormatServicePriceCOP(cents int64) string {
	whole := cents / 100
	frac := cents % 100
	return fmt.Sprintf("%d.%02d", whole, frac)
}

// Límites de paginación del catálogo público (HU-091), mismo criterio
// técnico que catalog.DefaultListLimit/MinListLimit/MaxListLimit: la
// colección no tiene un máximo de negocio de servicios por barbería.
const (
	DefaultServiceListLimit = 20
	MinServiceListLimit     = 1
	MaxServiceListLimit     = 50
)

// PublicServiceListResult es una página de PublicService ya ordenada de
// forma estable (created_at, id). NextCursor es "" cuando esta página es la
// última.
type PublicServiceListResult struct {
	Items      []PublicService
	NextCursor string
}

// ServiceCursor es la posición decodificada de una página del catálogo
// público, mismo patrón que catalog.Cursor (el núcleo de publicbooking no
// importa catalog, CA-002-06: duplicación pequeña y clara preferible a una
// abstracción cruzada entre módulos, docs/03-desarrollo/
// estandar-backend-go.md §5.23).
type ServiceCursor struct {
	CreatedAt time.Time
	ID        string
}

type serviceCursorWire struct {
	CreatedAt time.Time `json:"createdAt"`
	ID        string    `json:"id"`
}

// EncodeServiceCursor produce el valor opaco que la respuesta expone como
// `nextCursor`.
func EncodeServiceCursor(c ServiceCursor) string {
	raw, err := json.Marshal(serviceCursorWire{CreatedAt: c.CreatedAt, ID: c.ID})
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

// DecodeServiceCursor invierte EncodeServiceCursor. Cualquier valor que no
// haya salido de EncodeServiceCursor se rechaza como entrada de cliente
// inválida, nunca como error interno.
func DecodeServiceCursor(raw string) (ServiceCursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return ServiceCursor{}, apperr.Invalid("el parámetro cursor tiene un formato inválido")
	}
	var w serviceCursorWire
	if err := json.Unmarshal(decoded, &w); err != nil {
		return ServiceCursor{}, apperr.Invalid("el parámetro cursor tiene un formato inválido")
	}
	if w.ID == "" || w.CreatedAt.IsZero() {
		return ServiceCursor{}, apperr.Invalid("el parámetro cursor tiene un formato inválido")
	}
	return ServiceCursor{CreatedAt: w.CreatedAt, ID: w.ID}, nil
}
