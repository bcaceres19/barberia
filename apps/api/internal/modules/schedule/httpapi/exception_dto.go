package httpapi

import (
	"time"

	"system-barbershop/internal/modules/schedule"
)

// HolidayCalendarResponse es la representación canónica del interruptor de
// calendario colombiano de festivos de un barbero (HU-041, CA-041-01/02):
// exactamente enabled.
type HolidayCalendarResponse struct {
	Enabled bool `json:"enabled"`
}

// UpdateHolidayCalendarRequest es el cuerpo de
// PATCH /private/barbers/{barberId}/holiday-calendar (CA-041-01/02).
type UpdateHolidayCalendarRequest struct {
	Enabled bool `json:"enabled"`
}

// ScheduleExceptionSegmentInput es un tramo tal como lo envía el cliente en
// la creación/edición de una excepción (CA-041-04).
type ScheduleExceptionSegmentInput struct {
	StartsTime      string `json:"startsTime"`
	DurationMinutes int    `json:"durationMinutes"`
}

// ScheduleExceptionSegmentResponse es un tramo ya persistido de una
// excepción de jornada abierta.
type ScheduleExceptionSegmentResponse struct {
	ID              string `json:"id"`
	StartsTime      string `json:"startsTime"`
	DurationMinutes int    `json:"durationMinutes"`
}

// ScheduleExceptionResponse es la representación canónica de una excepción
// de jornada (HU-041, CA-041-04/05): exactamente id, effectiveDate,
// isClosed, reason, segments, createdAt, updatedAt. Nunca barbershopId ni
// barberId: el tenant se deriva de SessionCookie y el barbero de la ruta.
//
// ADVERTENCIA: postgres.exceptionResponseWire (schedule/postgres/
// exception_repository.go) declara EXACTAMENTE la misma forma (mismos
// nombres de campo JSON, mismo orden, mismos tipos) para construir el
// cuerpo que la idempotencia persiste en el alta. Un cambio aquí debe
// reflejarse ahí en el mismo commit.
type ScheduleExceptionResponse struct {
	ID            string                             `json:"id"`
	EffectiveDate string                             `json:"effectiveDate"`
	IsClosed      bool                               `json:"isClosed"`
	Reason        *string                            `json:"reason"`
	Segments      []ScheduleExceptionSegmentResponse `json:"segments"`
	CreatedAt     time.Time                          `json:"createdAt"`
	UpdatedAt     time.Time                          `json:"updatedAt"`
}

// ScheduleExceptionListResponse es la página paginada por cursor de
// GET /private/barbers/{barberId}/schedule-exceptions (CA-041-04/05).
type ScheduleExceptionListResponse struct {
	Items      []ScheduleExceptionResponse `json:"items"`
	NextCursor *string                     `json:"nextCursor"`
}

// CreateScheduleExceptionRequest es el cuerpo de
// POST /private/barbers/{barberId}/schedule-exceptions (CA-041-04/05).
// Cerrado: el handler rechaza cualquier campo desconocido. Nunca declara
// barbershopId ni barberId.
type CreateScheduleExceptionRequest struct {
	EffectiveDate string                          `json:"effectiveDate"`
	IsClosed      bool                            `json:"isClosed"`
	Reason        *string                         `json:"reason"`
	Segments      []ScheduleExceptionSegmentInput `json:"segments"`
}

// UpdateScheduleExceptionRequest es el cuerpo de
// PATCH /private/barbers/{barberId}/schedule-exceptions/{exceptionId}
// (CA-041-04/05): reemplaza la excepción completa.
type UpdateScheduleExceptionRequest struct {
	EffectiveDate string                          `json:"effectiveDate"`
	IsClosed      bool                            `json:"isClosed"`
	Reason        *string                         `json:"reason"`
	Segments      []ScheduleExceptionSegmentInput `json:"segments"`
}

// ColombianHolidayResponse es la representación canónica de un festivo
// colombiano calculado para un año (HU-041): exactamente date y name.
type ColombianHolidayResponse struct {
	Date string `json:"date"`
	Name string `json:"name"`
}

// ColombianHolidayListResponse es la colección completa de los festivos
// colombianos de UN año calendario (HU-041): sin paginación por cursor.
type ColombianHolidayListResponse struct {
	Items []ColombianHolidayResponse `json:"items"`
}

func newHolidayCalendarResponse(enabled bool) HolidayCalendarResponse {
	return HolidayCalendarResponse{Enabled: enabled}
}

func toExceptionSegmentInputs(raw []ScheduleExceptionSegmentInput) []schedule.CreateExceptionSegmentInput {
	segments := make([]schedule.CreateExceptionSegmentInput, 0, len(raw))
	for _, seg := range raw {
		segments = append(segments, schedule.CreateExceptionSegmentInput{StartsTime: seg.StartsTime, DurationMinutes: seg.DurationMinutes})
	}
	return segments
}

func newScheduleExceptionResponse(e schedule.ScheduleException) ScheduleExceptionResponse {
	segments := make([]ScheduleExceptionSegmentResponse, 0, len(e.Segments))
	for _, seg := range e.Segments {
		segments = append(segments, ScheduleExceptionSegmentResponse{ID: seg.ID, StartsTime: seg.StartsTime, DurationMinutes: seg.DurationMinutes})
	}
	return ScheduleExceptionResponse{
		ID:            e.ID,
		EffectiveDate: e.EffectiveDate,
		IsClosed:      e.IsClosed,
		Reason:        e.Reason,
		Segments:      segments,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}

func newScheduleExceptionListResponse(result schedule.ExceptionListResult) ScheduleExceptionListResponse {
	items := make([]ScheduleExceptionResponse, 0, len(result.Items))
	for _, e := range result.Items {
		items = append(items, newScheduleExceptionResponse(e))
	}
	var next *string
	if result.NextCursor != "" {
		v := result.NextCursor
		next = &v
	}
	return ScheduleExceptionListResponse{Items: items, NextCursor: next}
}

func newColombianHolidayListResponse(holidays []schedule.ColombianHoliday) ColombianHolidayListResponse {
	items := make([]ColombianHolidayResponse, 0, len(holidays))
	for _, h := range holidays {
		items = append(items, ColombianHolidayResponse{Date: h.Date.Format("2006-01-02"), Name: h.Name})
	}
	return ColombianHolidayListResponse{Items: items}
}
