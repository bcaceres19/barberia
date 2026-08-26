package schedule

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"system-barbershop/internal/platform/apperr"
)

// MaxReasonLength coincide con working_hour_override_reason_ck
// (20260826090000_create_working_hour_override.sql): el servicio no
// inventa un límite más estricto que la base ya aplica.
const MaxReasonLength = 200

// ExceptionSegment es un tramo especial ya validado de una excepción de
// jornada abierta (HU-041, CA-041-04): mismo par StartsTime/
// DurationMinutes que un tramo de WorkingHour, sin día de la semana -la
// fecha vive en la cabecera, no en cada tramo-.
type ExceptionSegment struct {
	ID              string
	StartsTime      string
	DurationMinutes int
}

// ScheduleException es la excepción de jornada de un barbero para una
// fecha civil concreta (HU-041, RN-BLQ-02, CA-041-04/05). IsClosed decide
// la forma: true significa día completamente cerrado (Segments siempre
// vacío); false significa día abierto con uno o varios tramos propios
// (Segments siempre no vacío). Prevalece sobre el festivo automático y
// sobre el horario semanal para esa fecha (CA-041-03/07).
type ScheduleException struct {
	ID            string
	EffectiveDate string // Fecha civil "YYYY-MM-DD", ya validada/canónica.
	IsClosed      bool
	Reason        *string
	Segments      []ExceptionSegment
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// exceptionIDPattern es la misma forma canónica que barberIDPattern/
// workingHourIDPattern.
var exceptionIDPattern = barberIDPattern

// LooksLikeExceptionID informa si id tiene la forma de un UUID válido.
func LooksLikeExceptionID(id string) bool {
	return exceptionIDPattern.MatchString(id)
}

// ValidateEffectiveDate confirma que raw tiene la forma YYYY-MM-DD de un
// calendario gregoriano válido (RFC 3339 full-date). No normaliza: una
// fecha con forma distinta o inexistente (por ejemplo 2026-02-30) se
// rechaza directamente.
func ValidateEffectiveDate(raw string) (string, error) {
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return "", errEffectiveDateInvalid()
	}
	// time.Parse acepta desbordamientos silenciosos en algunas variantes
	// (no en este layout exacto, pero se reformatea para blindarlo): si el
	// resultado no reproduce el mismo texto, la fecha no era real.
	if parsed.Format("2006-01-02") != raw {
		return "", errEffectiveDateInvalid()
	}
	return raw, nil
}

// ValidateReason recorta espacios y valida el largo máximo (CA-041-04). Un
// valor nil o vacío tras recortar se normaliza a nil: "sin motivo" no se
// distingue de "cadena vacía".
func ValidateReason(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, nil
	}
	if len([]rune(trimmed)) > MaxReasonLength {
		return nil, errReasonTooLong()
	}
	return &trimmed, nil
}

// ValidateExceptionShape aplica CA-041-04: una excepción cerrada no
// admite tramos; una excepción abierta exige al menos un tramo válido,
// sin solapes entre sí. Reutiliza schedule.IntervalsOverlap (HU-040): la
// fecha es fija por cabecera, así que la misma aritmética de minutos
// desde medianoche compara los tramos entre sí sin ambigüedad, incluido
// un tramo nocturno.
func ValidateExceptionShape(isClosed bool, segments []CreateExceptionSegmentInput) ([]CreateExceptionSegmentInput, error) {
	if isClosed {
		if len(segments) > 0 {
			return nil, errClosedExceptionHasSegments()
		}
		return nil, nil
	}

	if len(segments) == 0 {
		return nil, errOpenExceptionNeedsSegments()
	}

	validated := make([]CreateExceptionSegmentInput, 0, len(segments))
	for _, seg := range segments {
		startsTime, err := ValidateStartsTime(seg.StartsTime)
		if err != nil {
			return nil, err
		}
		if seg.DurationMinutes < MinDurationMinutes || seg.DurationMinutes > MaxDurationMinutes {
			return nil, errDurationInvalid()
		}
		validated = append(validated, CreateExceptionSegmentInput{StartsTime: startsTime, DurationMinutes: seg.DurationMinutes})
	}

	for i := 0; i < len(validated); i++ {
		for j := i + 1; j < len(validated); j++ {
			if IntervalsOverlap(minutesOfDay(validated[i].StartsTime), validated[i].DurationMinutes,
				minutesOfDay(validated[j].StartsTime), validated[j].DurationMinutes) {
				return nil, errSegmentsOverlap()
			}
		}
	}

	return validated, nil
}

// --- Cursor de excepciones (fecha efectiva, luego id) ---------------------

// ExceptionCursor es la posición decodificada de una página de la lista de
// excepciones (CA-041-04/05): orden estable por fecha efectiva y luego
// identificador, exactamente lo que idx_working_hour_override_shop_barber_date
// indexa.
type ExceptionCursor struct {
	EffectiveDate string
	ID            string
}

type exceptionCursorWire struct {
	EffectiveDate string `json:"effectiveDate"`
	ID            string `json:"id"`
}

// EncodeExceptionCursor produce el valor opaco que la respuesta expone
// como `nextCursor`.
func EncodeExceptionCursor(c ExceptionCursor) string {
	raw, err := json.Marshal(exceptionCursorWire{EffectiveDate: c.EffectiveDate, ID: c.ID})
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

// DecodeExceptionCursor invierte EncodeExceptionCursor.
func DecodeExceptionCursor(raw string) (ExceptionCursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return ExceptionCursor{}, apperr.Invalid("el parámetro cursor tiene un formato inválido")
	}
	var w exceptionCursorWire
	if err := json.Unmarshal(decoded, &w); err != nil {
		return ExceptionCursor{}, apperr.Invalid("el parámetro cursor tiene un formato inválido")
	}
	if w.ID == "" || w.EffectiveDate == "" {
		return ExceptionCursor{}, apperr.Invalid("el parámetro cursor tiene un formato inválido")
	}
	return ExceptionCursor{EffectiveDate: w.EffectiveDate, ID: w.ID}, nil
}
