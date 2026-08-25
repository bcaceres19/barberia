package schedule

import (
	"encoding/base64"
	"encoding/json"
	"regexp"
	"strconv"
	"time"

	"system-barbershop/internal/platform/apperr"
)

// MinISOWeekday y MaxISOWeekday acotan iso_weekday: 1 = lunes … 7 = domingo
// (DEC-020). No se usa 0-6 para no heredar la ambigüedad de qué día es el
// cero.
const (
	MinISOWeekday = 1
	MaxISOWeekday = 7
)

// MinDurationMinutes y MaxDurationMinutes coinciden con
// working_hour_duration_minutes_ck (20260825160000_create_working_hour.sql):
// el servicio no inventa un límite más estricto que la base ya aplica
// (CA-040-04).
const (
	MinDurationMinutes = 1
	MaxDurationMinutes = 1440
)

// DefaultListLimit, MinListLimit y MaxListLimit acotan `limit` en la lista
// paginada (CA-040-01): límites técnicos, no un máximo de negocio de tramos
// por barbero (docs/06-api/estandar-openapi.md §6.10).
const (
	DefaultListLimit = 20
	MinListLimit     = 1
	MaxListLimit     = 50
)

// WorkingHour es un tramo recurrente de la jornada laboral de un barbero,
// para un único día ISO de la semana (HU-040, CA-040-01/02/03). Varios
// tramos del mismo barbero y día representan una jornada partida; un tramo
// cuyo StartsTime + DurationMinutes supera las 24:00 representa una
// jornada nocturna (DEC-020): se almacena inicio + duración, nunca un fin
// calculado, para no ser ambiguo cuando el fin cae al día siguiente.
type WorkingHour struct {
	ID              string
	ISOWeekday      int
	StartsTime      string // Hora civil "HH:MM", ya validada/canónica.
	DurationMinutes int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// barberIDPattern y workingHourIDPattern son la misma forma canónica
// 8-4-4-4-12 que database.ValidBarbershopID/staff.LooksLikeBarberID exigen.
// El núcleo de schedule no importa ninguno de esos dos paquetes
// (CA-002-06 y el mismo criterio de "colaboración solo por puerto
// explícito" de HU-023 frente a staff): duplicar esta única expresión
// regular es la misma duplicación pequeña y clara que catalog.barberIDPattern
// ya documenta frente a staff.barberIDPattern.
var (
	barberIDPattern      = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	workingHourIDPattern = barberIDPattern
)

// LooksLikeBarberID informa si id tiene la forma de un UUID válido. Un
// identificador que no cumple esta forma no puede corresponder a ninguna
// fila real: se trata igual que "no existe" (mismo apperr.NotFound) en vez
// de dejar que la consulta a BarberPort o a PostgreSQL falle con un error
// de tipo.
func LooksLikeBarberID(id string) bool {
	return barberIDPattern.MatchString(id)
}

// LooksLikeWorkingHourID informa si id tiene la forma de un UUID válido.
func LooksLikeWorkingHourID(id string) bool {
	return workingHourIDPattern.MatchString(id)
}

// startsTimePattern exige HH:MM de 24 horas, cero-rellenado (mismo patrón
// que el contrato OpenAPI, api/openapi/components/schemas/
// CreateWorkingHourRequest.yaml): "8:00" o "24:00" se rechazan.
var startsTimePattern = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

// ValidateStartsTime confirma que raw tiene la forma HH:MM de 24 horas. No
// normaliza: a diferencia de un nombre de texto libre, una hora con forma
// distinta no tiene una única forma "recortada" razonable, así que se
// rechaza directamente (CA-040-04).
func ValidateStartsTime(raw string) (string, error) {
	if !startsTimePattern.MatchString(raw) {
		return "", errStartsTimeInvalid()
	}
	return raw, nil
}

// minutesOfDay convierte una hora HH:MM ya validada (ValidateStartsTime) en
// minutos desde medianoche. No revalida el formato: los llamadores dentro
// de este paquete solo lo invocan sobre valores que ya pasaron
// ValidateStartsTime.
func minutesOfDay(startsTime string) int {
	hours, _ := strconv.Atoi(startsTime[0:2])
	minutes, _ := strconv.Atoi(startsTime[3:5])
	return hours*60 + minutes
}

// IntervalsOverlap informa si dos tramos [aStart, aStart+aDuration) y
// [bStart, bStart+bDuration), ambos en minutos desde medianoche sobre el
// MISMO día ISO de la semana, se solapan. Semántica semiabierta [inicio,
// fin): dos tramos contiguos (el fin de uno coincide con el inicio del
// otro) NO se consideran solapados (CA-040-03). aStart/bStart nunca se
// reducen módulo 1440: un tramo nocturno cuyo fin supera 1440 se compara
// con la misma aritmética simple que estandar-base-datos.md documenta con
// el ancla 2000-01-01 para working_hour_override_segment; esta función
// deliberadamente NO resuelve el solape entre el arrastre de un tramo
// nocturno de un día y el tramo del día ISO siguiente (fuera de alcance de
// HU-040, ver el comentario "Qué NO hace esta migración" de
// 20260825160000_create_working_hour.sql).
func IntervalsOverlap(aStart, aDuration, bStart, bDuration int) bool {
	aEnd := aStart + aDuration
	bEnd := bStart + bDuration
	return aStart < bEnd && bStart < aEnd
}

// Cursor es la posición decodificada de una página de la lista (CA-040-01):
// orden estable por día ISO de la semana, luego hora de inicio y luego
// identificador, exactamente los tres campos que
// idx_working_hour_shop_barber_weekday_start indexa.
type Cursor struct {
	ISOWeekday int
	StartsTime string
	ID         string
}

// cursorWire es la forma serializada del cursor. Deliberadamente separada
// de Cursor: el cliente solo debe tratarlo como un valor opaco.
type cursorWire struct {
	ISOWeekday int    `json:"isoWeekday"`
	StartsTime string `json:"startsTime"`
	ID         string `json:"id"`
}

// EncodeCursor produce el valor opaco que la respuesta expone como
// `nextCursor`.
func EncodeCursor(c Cursor) string {
	raw, err := json.Marshal(cursorWire{ISOWeekday: c.ISOWeekday, StartsTime: c.StartsTime, ID: c.ID})
	if err != nil {
		// No debería ocurrir nunca (los campos son int/string): degradar a
		// cursor vacío es preferible a propagar un panic desde una función
		// que el contrato documenta como infalible en la práctica.
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

// DecodeCursor invierte EncodeCursor. Cualquier valor que no haya salido de
// EncodeCursor se rechaza como entrada de cliente inválida, nunca como
// error interno.
func DecodeCursor(raw string) (Cursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return Cursor{}, apperr.Invalid("el parámetro cursor tiene un formato inválido")
	}
	var w cursorWire
	if err := json.Unmarshal(decoded, &w); err != nil {
		return Cursor{}, apperr.Invalid("el parámetro cursor tiene un formato inválido")
	}
	if w.ID == "" || w.StartsTime == "" || w.ISOWeekday < MinISOWeekday || w.ISOWeekday > MaxISOWeekday {
		return Cursor{}, apperr.Invalid("el parámetro cursor tiene un formato inválido")
	}
	return Cursor{ISOWeekday: w.ISOWeekday, StartsTime: w.StartsTime, ID: w.ID}, nil
}
