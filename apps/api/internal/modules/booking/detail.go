package booking

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// appointmentIDPattern refleja la forma de un UUID, mismo criterio que
// barberIDPattern (CA-002-06: cada módulo repite su propia validación de
// forma). Un identificador que no cumple esta forma no puede corresponder a
// ninguna fila real: se trata igual que "no existe" (mismo apperr.NotFound)
// en vez de dejar que la consulta SQL falle con un error de tipo.
var appointmentIDPattern = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
)

// LooksLikeAppointmentID informa si id tiene la forma de un UUID válido.
func LooksLikeAppointmentID(id string) bool {
	return appointmentIDPattern.MatchString(id)
}

// Límites de la página de historial (HU-064, CA-064-05), mismo criterio de
// límites técnicos (no de negocio) que staff.DefaultListLimit/MinListLimit/
// MaxListLimit.
const (
	HistoryDefaultLimit = 20
	HistoryMinLimit     = 1
	HistoryMaxLimit     = 50
)

// AppointmentDetail es la lectura privada de una cita (HU-064, CA-064-01 a
// CA-064-04): persona atendida, cliente que reservó (con su contacto
// opcional y nota), barbero, origen, intervalo, estado y snapshots de
// servicio (DEC-004, nunca el catálogo vigente). Deliberadamente NUNCA
// incluye CustomerID ni BarbershopID (fuera de alcance del prompt de
// HU-064): BarberFullName se resuelve aparte, vía BarberNamePort, porque el
// repositorio de booking no puede unir contra la tabla `barber` (dueña de
// staff, CA-002-06).
type AppointmentDetail struct {
	ID                       string
	BarberID                 string
	BarberFullName           string
	AttendeeName             string
	CustomerFullName         string
	CustomerPhone            *string
	CustomerEmail            *string
	CustomerNote             *string
	StartsAt                 time.Time
	EndsAt                   time.Time
	Status                   Status
	Origin                   Origin
	ServiceNameSnapshot      string
	DurationMinutesSnapshot  int
	PriceAmountCentsSnapshot int64
	CurrencySnapshot         string
	// VersionToken es el token opaco de concurrencia (trabajo requerido
	// §2.4 del prompt de HU-064): determinista a partir de ID y del instante
	// de actualización vigente, pero irreversible (hash, no una codificación
	// de updated_at), para que una mutación futura pueda usarlo como
	// precondición sin que updated_at se vuelva una regla de negocio pública.
	VersionToken string
	CreatedAt    time.Time
}

// EncodeVersionToken produce el token opaco de concurrencia de una cita a
// partir de su identificador y su updated_at vigente (irreversible: SHA-256,
// no una codificación reversible de updated_at). Dos lecturas de la misma
// fila sin ninguna escritura entre medias producen el mismo token; cualquier
// escritura que toque updated_at (disparador *_set_updated_at) produce uno
// distinto.
func EncodeVersionToken(appointmentID string, updatedAt time.Time) string {
	sum := sha256.Sum256([]byte(appointmentID + "|" + strconv.FormatInt(updatedAt.UTC().UnixNano(), 10)))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// HistoryChange es un campo modificado con su valor anterior y nuevo
// (RN-HIS-01), proyección directa de appointment_history_change. HU-064 solo
// lee: nunca inserta ninguna fila.
type HistoryChange struct {
	FieldName     string
	PreviousValue *string
	NewValue      *string
}

// HistoryRow es la fila cruda que el repositorio devuelve para una entrada
// de historial: conserva los identificadores de actor SIN resolver
// (ActorStaffUserID/ActorCustomerID) porque batchear su resolución en un
// solo lote por página (trabajo requerido §2.3, "sin generar N+1") es
// responsabilidad de DetailService, no del repositorio.
type HistoryRow struct {
	ID               string
	EventType        EventType
	ActorType        ActorType
	ActorStaffUserID *string
	ActorCustomerID  *string
	Reason           *string
	OccurredAt       time.Time
	Changes          []HistoryChange
}

// HistoryEntry es la fila de historial ya lista para exponer (HU-064,
// CA-064-05): ActorLabel es un nombre visible seguro (nunca correo ni un
// identificador interno), resuelto por DetailService a partir de
// HistoryRow.ActorStaffUserID/ActorCustomerID.
type HistoryEntry struct {
	ID         string
	EventType  EventType
	ActorType  ActorType
	ActorLabel string
	Reason     *string
	OccurredAt time.Time
	Changes    []HistoryChange
}

// HistoryPage es una página del historial, orden estable (occurred_at, id).
// NextCursor es "" cuando esta página es la última.
type HistoryPage struct {
	Items      []HistoryEntry
	NextCursor string
}

// HistoryCursor es la posición decodificada de una página de historial:
// orden estable por instante y luego por identificador, exactamente los dos
// campos finales que idx_appointment_history_shop_appointment_occurred
// indexa junto con barbershop_id/appointment_id (mismo criterio que
// staff.Cursor frente a idx_barber_shop_created_id).
type HistoryCursor struct {
	OccurredAt time.Time
	ID         string
}

// historyCursorWire es la forma serializada del cursor: deliberadamente
// separada de HistoryCursor para que el cliente solo la trate como un valor
// opaco (mismo criterio que staff.cursorWire).
type historyCursorWire struct {
	OccurredAt time.Time `json:"occurredAt"`
	ID         string    `json:"id"`
}

// EncodeHistoryCursor produce el valor opaco que la respuesta expone como
// `nextCursor`.
func EncodeHistoryCursor(c HistoryCursor) string {
	raw, err := json.Marshal(historyCursorWire{OccurredAt: c.OccurredAt, ID: c.ID})
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

// DecodeHistoryCursor invierte EncodeHistoryCursor. Cualquier valor que no
// haya salido de EncodeHistoryCursor se rechaza como entrada de cliente
// inválida (mismo criterio que staff.DecodeCursor).
func DecodeHistoryCursor(raw string) (HistoryCursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return HistoryCursor{}, errHistoryCursorInvalid()
	}
	var w historyCursorWire
	if err := json.Unmarshal(decoded, &w); err != nil {
		return HistoryCursor{}, errHistoryCursorInvalid()
	}
	if w.ID == "" || w.OccurredAt.IsZero() {
		return HistoryCursor{}, errHistoryCursorInvalid()
	}
	return HistoryCursor{OccurredAt: w.OccurredAt, ID: w.ID}, nil
}

// clampHistoryLimit aplica el mismo criterio de límites técnicos que
// staff.Service.List: un limit ausente o no positivo usa el valor por
// defecto; un limit fuera de [HistoryMinLimit, HistoryMaxLimit] se recorta,
// nunca se rechaza.
func clampHistoryLimit(limit int) int {
	if limit <= 0 {
		return HistoryDefaultLimit
	}
	if limit < HistoryMinLimit {
		return HistoryMinLimit
	}
	if limit > HistoryMaxLimit {
		return HistoryMaxLimit
	}
	return limit
}

// safeActorLabel produce la etiqueta de reserva cuando el nombre autorizado
// del actor no está disponible (trabajo requerido §2.3: "si falta un nombre
// autorizado, usa una etiqueta segura coherente con el tipo de actor"). En
// la práctica no debería ocurrir para 'staff'/'customer' (las FK de
// appointment_history son ON DELETE RESTRICT: la fila referenciada nunca
// desaparece), pero el codigo no depende de esa garantía para no fallar de
// forma insegura.
func safeActorLabel(actorType ActorType) string {
	switch actorType {
	case ActorTypeStaff:
		return "Miembro del equipo"
	case ActorTypeCustomer:
		return "Cliente"
	default:
		return "Sistema"
	}
}

// resolveActorLabel decide el nombre visible de una fila de historial a
// partir de los lotes ya resueltos (staffNames/customerNames), sin volver a
// consultar nada (trabajo requerido §2.3, "sin generar N+1").
func resolveActorLabel(row HistoryRow, staffNames, customerNames map[string]string) string {
	switch row.ActorType {
	case ActorTypeStaff:
		if row.ActorStaffUserID != nil {
			if name, ok := staffNames[*row.ActorStaffUserID]; ok && strings.TrimSpace(name) != "" {
				return name
			}
		}
		return safeActorLabel(ActorTypeStaff)
	case ActorTypeCustomer:
		if row.ActorCustomerID != nil {
			if name, ok := customerNames[*row.ActorCustomerID]; ok && strings.TrimSpace(name) != "" {
				return name
			}
		}
		return safeActorLabel(ActorTypeCustomer)
	default:
		return safeActorLabel(ActorTypeSystem)
	}
}
