package googlecalendar

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// Tipos de recurso publicables.
const (
	ResourceAppointment = "appointment"
	ResourceTimeBlock   = "time_block"
)

// Event es el evento de Google Calendar que NAVA publica. Es un tipo propio del
// módulo: el adaptador de Google lo serializa, el núcleo nunca ve el SDK ni el
// formato de la API.
type Event struct {
	// ID es el identificador que se propone al crear (determinista, ver EventID);
	// vacío al actualizar.
	ID          string
	Summary     string
	Description string
	StartsAt    time.Time
	EndsAt      time.Time
	TimeZone    string
	// Private son las propiedades extendidas privadas con identificadores
	// opacos (navaResourceType, navaResourceId, navaConnectionId): el único modo
	// de reconocer los eventos publicados por NAVA, nunca por título ni por hora.
	Private map[string]string
	// ReminderMinutes es la anticipación del único recordatorio emergente; nil
	// usa los recordatorios predeterminados del calendario (DEC-101).
	ReminderMinutes *int
	// AttendeeEmail invita al cliente como asistente (DEC-122); vacío = sin
	// invitado. El correo solo viaja como asistente: nunca en título,
	// descripción ni propiedades extendidas.
	AttendeeEmail string
}

// Claves de las propiedades extendidas privadas.
const (
	PropResourceType = "navaResourceType"
	PropResourceID   = "navaResourceId"
	PropConnectionID = "navaConnectionId"
)

// blockTitles son los títulos de los bloqueos por tipo (el motivo es una nota
// del propio barbero y no se publica).
var blockTitles = map[string]string{
	"break":       "Descanso",
	"lunch":       "Almuerzo",
	"unavailable": "No disponible",
	"day_off":     "Día libre",
	"vacation":    "Vacaciones",
	"emergency":   "Emergencia",
}

// IsPublishableBlock informa si un bloqueo puntual se publica: los manuales de
// los tipos de DEC-101; `holiday` (calendario automático) no.
func IsPublishableBlock(blockType, source string) bool {
	if source != "manual" {
		return false
	}
	_, ok := blockTitles[blockType]
	return ok
}

// BuildEvent arma el evento y la huella de lo que ve el invitado a partir del
// estado ACTUAL del recurso (la cola es a nivel de estado). Privacidad
// (DEC-101.6): el título de una cita es «Nombre — Servicio» y la descripción
// solo el servicio y el estado; nunca teléfono, notas, tokens ni identificadores
// internos fuera de las propiedades extendidas privadas.
func BuildEvent(jc JobContext, connectionID string) (Event, string) {
	ev := Event{
		StartsAt: jc.StartsAt,
		EndsAt:   jc.EndsAt,
		TimeZone: jc.Timezone,
		Private: map[string]string{
			PropResourceType: jc.ResourceType,
			PropResourceID:   jc.ResourceID,
			PropConnectionID: connectionID,
		},
		ReminderMinutes: jc.ReminderMinutes,
	}

	switch jc.ResourceType {
	case ResourceAppointment:
		ev.Summary = fmt.Sprintf("%s — %s", strings.TrimSpace(jc.AttendeeName), strings.TrimSpace(jc.ServiceName))
		ev.Description = fmt.Sprintf("Servicio: %s\nEstado: Confirmada", strings.TrimSpace(jc.ServiceName))
		// DEC-122: se invita al cliente solo si la conexión está sana y hay correo.
		if jc.ConnectionStatus == string(StatusConnected) {
			ev.AttendeeEmail = strings.TrimSpace(jc.CustomerEmail)
		}
	default:
		ev.Summary = blockTitles[jc.BlockType]
		if ev.Summary == "" {
			ev.Summary = "Bloqueo"
		}
		ev.Description = "Bloqueo de agenda publicado por NAVA"
	}
	return ev, visibleHash(ev)
}

// visibleHash es la huella de lo que el invitado percibe: si no cambia, refrescar
// el evento (por ejemplo, al cambiar el recordatorio del barbero) no le envía un
// correo (DEC-122).
func visibleHash(ev Event) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		ev.Summary,
		ev.StartsAt.UTC().Format(time.RFC3339),
		ev.EndsAt.UTC().Format(time.RFC3339),
		ev.AttendeeEmail,
	}, "\n")))
	return hex.EncodeToString(sum[:])
}

// EventID es el identificador que NAVA propone al crear un evento: determinista
// en (conexión, recurso, generación). Así, si el worker cae entre «crear en
// Google» y «guardar el vínculo», el reintento propone el MISMO id y Google
// responde «ya existe» en vez de crear un segundo evento. Solo usa caracteres
// de base32hex (0-9, a-v), como exige Google. La generación sube cuando hay que
// recrear el evento: Google no permite reutilizar el id de uno borrado.
func EventID(connectionID, resourceType, resourceID string, generation int) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s|%d", connectionID, resourceType, resourceID, generation)))
	return "nava" + hex.EncodeToString(sum[:])[:32]
}
