package schedule

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

// LooksLikeTimeBlockID y LooksLikeSeriesID reutilizan la misma forma
// canónica que barberIDPattern/workingHourIDPattern (domain.go): un
// identificador que no la cumple no puede corresponder a ninguna fila real.
func LooksLikeTimeBlockID(id string) bool { return barberIDPattern.MatchString(id) }
func LooksLikeSeriesID(id string) bool    { return barberIDPattern.MatchString(id) }

// BlockType enumera los siete tipos que el criterio de salida de B2 exige
// (RN-BLQ-01): descanso, almuerzo, hora no disponible, día libre, festivo,
// vacaciones y emergencia. Mismo vocabulario cerrado que
// time_block_block_type_ck/time_block_series_block_type_ck
// (20260826100000_create_time_block.sql).
const (
	BlockTypeBreak       = "break"
	BlockTypeLunch       = "lunch"
	BlockTypeUnavailable = "unavailable"
	BlockTypeDayOff      = "day_off"
	BlockTypeHoliday     = "holiday"
	BlockTypeVacation    = "vacation"
	BlockTypeEmergency   = "emergency"
)

var validBlockTypes = map[string]bool{
	BlockTypeBreak:       true,
	BlockTypeLunch:       true,
	BlockTypeUnavailable: true,
	BlockTypeDayOff:      true,
	BlockTypeHoliday:     true,
	BlockTypeVacation:    true,
	BlockTypeEmergency:   true,
}

// RecurrenceKind distingue una serie semanal (un único día ISO) de una lista
// explícita de fechas (DEC-020), mismo vocabulario que
// time_block_series_recurrence_kind_ck.
const (
	RecurrenceKindWeekly   = "weekly"
	RecurrenceKindDateList = "date_list"
)

// BlockSource distingue un bloqueo manual de uno generado por el calendario
// colombiano de festivos (HU-041), mismo vocabulario que time_block_source_ck.
// HU-042 solo produce bloqueos "manual"; "holiday_calendar" queda reservado
// para cuando HU-041 empiece a materializar sus festivos aquí (fuera de
// alcance de esta HU).
const BlockSourceManual = "manual"

// ValidateBlockType confirma que raw es uno de los siete tipos cerrados.
func ValidateBlockType(raw string) (string, error) {
	if !validBlockTypes[raw] {
		return "", errBlockTypeInvalid()
	}
	return raw, nil
}

// ValidateRecurrenceKind confirma que raw es "weekly" o "date_list".
func ValidateRecurrenceKind(raw string) (string, error) {
	if raw != RecurrenceKindWeekly && raw != RecurrenceKindDateList {
		return "", errRecurrenceKindInvalid()
	}
	return raw, nil
}

// ValidateReason (recorte y límite de 200 caracteres) ya existe en
// exception_domain.go (MaxReasonLength, CA-041-04): time_block_reason_ck y
// time_block_series_reason_ck comparten el mismo límite, así que HU-042
// reutiliza esa función en vez de duplicarla.

// dateLayout es la forma civil "YYYY-MM-DD" que también usa
// schedule-exceptions (effectiveDate) para no depender de zona horaria en el
// borde del contrato: el cliente ya resolvió a qué fecha civil de la
// barbería corresponde antes de enviarla.
const dateLayout = "2006-01-02"

// ValidateCivilDate confirma que raw tiene la forma YYYY-MM-DD y corresponde
// a una fecha calendario real (rechaza "2026-02-30").
func ValidateCivilDate(raw string) (string, error) {
	parsed, err := time.Parse(dateLayout, raw)
	if err != nil || parsed.Format(dateLayout) != raw {
		return "", errDateInvalid()
	}
	return raw, nil
}

// ValidateInstant confirma que raw es una marca de tiempo RFC 3339 con
// offset explícito: un bloqueo puntual es un instante absoluto, no una hora
// civil recurrente (a diferencia de working_hour/schedule-exceptions). El
// cliente ya conoce la zona IANA de la barbería (GET .../barbershop) y
// calcula el offset correcto antes de enviarlo (RN-DIS-05/RN-DIS-07).
func ValidateInstant(raw string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, errDateTimeInvalid()
	}
	return parsed, nil
}

// ValidateBlockInterval exige el intervalo semiabierto [startsAt, endsAt)
// con fin estrictamente posterior (RN-DIS-05), incluido el cruce de
// medianoche: endsAt puede caer al día siguiente sin ninguna rama especial,
// porque ambos son instantes absolutos, no horas civiles.
func ValidateBlockInterval(startsAt, endsAt time.Time) error {
	if !endsAt.After(startsAt) {
		return errIntervalInvalid()
	}
	return nil
}

// ValidateWeekdayShape aplica la misma regla que
// time_block_series_weekday_shape_ck: weekly exige iso_weekday 1-7;
// date_list exige ausencia de iso_weekday.
func ValidateWeekdayShape(recurrenceKind string, isoWeekday *int) error {
	switch recurrenceKind {
	case RecurrenceKindWeekly:
		if isoWeekday == nil || *isoWeekday < MinISOWeekday || *isoWeekday > MaxISOWeekday {
			return errWeekdayShapeInvalid()
		}
	case RecurrenceKindDateList:
		if isoWeekday != nil {
			return errWeekdayShapeInvalid()
		}
	}
	return nil
}

// ValidateEffectiveRange exige que effectiveUntil, si no es nil, no sea
// anterior a effectiveFrom (time_block_series_effective_range_ck). La
// comparación lexicográfica es válida porque ambas ya pasaron
// ValidateCivilDate (forma YYYY-MM-DD).
func ValidateEffectiveRange(effectiveFrom string, effectiveUntil *string) error {
	if effectiveUntil != nil && *effectiveUntil < effectiveFrom {
		return errEffectiveRangeInvalid()
	}
	return nil
}

// TimeBlock es un bloqueo puntual de agenda, incluida la emergencia
// (RN-BLQ-01, RN-BLQ-03). DeletedAt/DeletedBy != nil indica retiro lógico
// (RN-BLQ-04): el registro se conserva pero deja de restar disponibilidad.
type TimeBlock struct {
	ID        string
	BlockType string
	Source    string
	StartsAt  time.Time
	EndsAt    time.Time
	Reason    *string
	DeletedAt *time.Time
	DeletedBy *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SeriesDate es una fecha explícita de una serie date_list
// (time_block_series_date).
type SeriesDate struct {
	BlockDate string
}

// SeriesException es una instancia suprimida de una serie recurrente
// ("esta instancia no", RN-BLQ-01).
type SeriesException struct {
	ExcludedDate string
	Reason       *string
	CreatedAt    time.Time
}

// TimeBlockSeries es una definición recurrente semanal o una lista explícita
// de fechas (RN-BLQ-01, DEC-020). Nunca se materializa en TimeBlock: la
// disponibilidad la expande en la consulta del rango pedido
// (Service.EffectiveBlocks). Dates y Exceptions vienen embebidos en la
// representación: ninguna pantalla necesita paginarlos por separado, un
// bloqueo recurrente real rara vez acumula más de unas pocas decenas.
type TimeBlockSeries struct {
	ID              string
	BlockType       string
	RecurrenceKind  string
	ISOWeekday      *int
	StartsTime      string
	DurationMinutes int
	EffectiveFrom   string
	EffectiveUntil  *string
	Reason          *string
	DeletedAt       *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Dates           []SeriesDate
	Exceptions      []SeriesException
}

// BlockCursor es la posición decodificada de una página de bloqueos
// puntuales: orden estable (starts_at, id), el mismo prefijo que
// idx_time_block_shop_barber_starts_at indexa.
type BlockCursor struct {
	StartsAt string // RFC 3339, ya canónico.
	ID       string
}

type blockCursorWire struct {
	StartsAt string `json:"startsAt"`
	ID       string `json:"id"`
}

// EncodeBlockCursor produce el valor opaco `nextCursor` de la lista de
// bloqueos puntuales.
func EncodeBlockCursor(c BlockCursor) string {
	raw, err := json.Marshal(blockCursorWire{StartsAt: c.StartsAt, ID: c.ID})
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

// DecodeBlockCursor invierte EncodeBlockCursor.
func DecodeBlockCursor(raw string) (BlockCursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return BlockCursor{}, errCursorInvalid()
	}
	var w blockCursorWire
	if err := json.Unmarshal(decoded, &w); err != nil {
		return BlockCursor{}, errCursorInvalid()
	}
	if w.ID == "" || w.StartsAt == "" {
		return BlockCursor{}, errCursorInvalid()
	}
	return BlockCursor{StartsAt: w.StartsAt, ID: w.ID}, nil
}

// SeriesCursor es la posición decodificada de una página de series: orden
// estable (effective_from, id), el mismo prefijo que
// idx_time_block_series_shop_barber indexa.
type SeriesCursor struct {
	EffectiveFrom string
	ID            string
}

type seriesCursorWire struct {
	EffectiveFrom string `json:"effectiveFrom"`
	ID            string `json:"id"`
}

// EncodeSeriesCursor produce el valor opaco `nextCursor` de la lista de series.
func EncodeSeriesCursor(c SeriesCursor) string {
	raw, err := json.Marshal(seriesCursorWire{EffectiveFrom: c.EffectiveFrom, ID: c.ID})
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

// DecodeSeriesCursor invierte EncodeSeriesCursor.
func DecodeSeriesCursor(raw string) (SeriesCursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return SeriesCursor{}, errCursorInvalid()
	}
	var w seriesCursorWire
	if err := json.Unmarshal(decoded, &w); err != nil {
		return SeriesCursor{}, errCursorInvalid()
	}
	if w.ID == "" || w.EffectiveFrom == "" {
		return SeriesCursor{}, errCursorInvalid()
	}
	return SeriesCursor{EffectiveFrom: w.EffectiveFrom, ID: w.ID}, nil
}

// UpdateSeriesScope distingue "toda la serie" de "esta y las siguientes"
// (RN-BLQ-01, caso límite de edición de una recurrencia). "Esta instancia"
// se resuelve con el sub-recurso de excepciones (POST .../exceptions): no es
// un alcance de este PATCH porque no reemplaza ningún campo, solo suprime
// una fecha puntual.
const (
	UpdateScopeWhole            = "whole"
	UpdateScopeThisAndFollowing = "this_and_following"
)
