package httpapi

import "time"

// TimeBlockResponse es la representación canónica de un bloqueo puntual
// (HU-042): exactamente id, blockType, source, startsAt, endsAt, reason,
// deletedAt, deletedBy, createdAt, updatedAt. Nunca barbershopId ni
// barberId: el tenant se deriva de SessionCookie y el barbero de la ruta.
//
// ADVERTENCIA: postgres.timeBlockResponseWire (schedule/postgres/
// block_repository.go) declara EXACTAMENTE la misma forma para construir el
// cuerpo que la idempotencia persiste en el alta. Un cambio aquí debe
// reflejarse ahí en el mismo commit.
type TimeBlockResponse struct {
	ID        string     `json:"id"`
	BlockType string     `json:"blockType"`
	Source    string     `json:"source"`
	StartsAt  time.Time  `json:"startsAt"`
	EndsAt    time.Time  `json:"endsAt"`
	Reason    *string    `json:"reason"`
	DeletedAt *time.Time `json:"deletedAt"`
	DeletedBy *string    `json:"deletedBy"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// TimeBlockListResponse es la página paginada por cursor de
// GET /private/barbers/{barberId}/time-blocks.
type TimeBlockListResponse struct {
	Items      []TimeBlockResponse `json:"items"`
	NextCursor *string             `json:"nextCursor"`
}

// CreateTimeBlockRequest es el cuerpo de
// POST /private/barbers/{barberId}/time-blocks. Cerrado: el handler rechaza
// cualquier campo desconocido.
type CreateTimeBlockRequest struct {
	BlockType string  `json:"blockType"`
	StartsAt  string  `json:"startsAt"`
	EndsAt    string  `json:"endsAt"`
	Reason    *string `json:"reason"`
}

// SeriesDateResponse es una fecha explícita embebida en
// TimeBlockSeriesResponse.
type SeriesDateResponse struct {
	BlockDate string `json:"blockDate"`
}

// SeriesExceptionResponse es una excepción embebida en
// TimeBlockSeriesResponse ("esta instancia no").
type SeriesExceptionResponse struct {
	ExcludedDate string    `json:"excludedDate"`
	Reason       *string   `json:"reason"`
	CreatedAt    time.Time `json:"createdAt"`
}

// TimeBlockSeriesResponse es la representación canónica de una definición
// recurrente (HU-042): weekly o date_list, con sus fechas/excepciones
// embebidas. Nunca barbershopId ni barberId.
//
// ADVERTENCIA: postgres.timeBlockSeriesResponseWire declara EXACTAMENTE la
// misma forma para el cuerpo que la idempotencia persiste en el alta.
type TimeBlockSeriesResponse struct {
	ID              string                    `json:"id"`
	BlockType       string                    `json:"blockType"`
	RecurrenceKind  string                    `json:"recurrenceKind"`
	ISOWeekday      *int                      `json:"isoWeekday"`
	StartsTime      string                    `json:"startsTime"`
	DurationMinutes int                       `json:"durationMinutes"`
	EffectiveFrom   string                    `json:"effectiveFrom"`
	EffectiveUntil  *string                   `json:"effectiveUntil"`
	Reason          *string                   `json:"reason"`
	DeletedAt       *time.Time                `json:"deletedAt"`
	CreatedAt       time.Time                 `json:"createdAt"`
	UpdatedAt       time.Time                 `json:"updatedAt"`
	Dates           []SeriesDateResponse      `json:"dates"`
	Exceptions      []SeriesExceptionResponse `json:"exceptions"`
}

// TimeBlockSeriesListResponse es la página paginada por cursor de
// GET /private/barbers/{barberId}/time-block-series.
type TimeBlockSeriesListResponse struct {
	Items      []TimeBlockSeriesResponse `json:"items"`
	NextCursor *string                   `json:"nextCursor"`
}

// CreateTimeBlockSeriesRequest es el cuerpo de
// POST /private/barbers/{barberId}/time-block-series. explicitDates solo se
// admite cuando recurrenceKind es date_list (RN-BLQ-01: "bloqueo de varios
// días debe poder crearse en una sola operación").
type CreateTimeBlockSeriesRequest struct {
	BlockType       string   `json:"blockType"`
	RecurrenceKind  string   `json:"recurrenceKind"`
	ISOWeekday      *int     `json:"isoWeekday"`
	StartsTime      string   `json:"startsTime"`
	DurationMinutes int      `json:"durationMinutes"`
	EffectiveFrom   string   `json:"effectiveFrom"`
	EffectiveUntil  *string  `json:"effectiveUntil"`
	Reason          *string  `json:"reason"`
	ExplicitDates   []string `json:"explicitDates"`
}

// UpdateTimeBlockSeriesRequest es el cuerpo de
// PATCH /private/barbers/{barberId}/time-block-series/{seriesId}
// (RN-BLQ-01: "esta y las siguientes" o "toda la serie"). effectiveDate
// solo es obligatorio cuando scope es this_and_following: es el punto de
// corte donde nace la serie nueva.
type UpdateTimeBlockSeriesRequest struct {
	Scope           string  `json:"scope"`
	BlockType       string  `json:"blockType"`
	StartsTime      string  `json:"startsTime"`
	DurationMinutes int     `json:"durationMinutes"`
	EffectiveFrom   string  `json:"effectiveFrom"`
	EffectiveUntil  *string `json:"effectiveUntil"`
	Reason          *string `json:"reason"`
	EffectiveDate   string  `json:"effectiveDate"`
}

// AddSeriesDateRequest es el cuerpo de
// POST .../time-block-series/{seriesId}/dates.
type AddSeriesDateRequest struct {
	BlockDate string `json:"blockDate"`
}

// AddSeriesExceptionRequest es el cuerpo de
// POST .../time-block-series/{seriesId}/exceptions.
type AddSeriesExceptionRequest struct {
	ExcludedDate string  `json:"excludedDate"`
	Reason       *string `json:"reason"`
}

// SeriesOccurrenceResponse es una ocurrencia expandida de una serie dentro
// del rango consultado (civil: fecha + hora local de la barbería).
type SeriesOccurrenceResponse struct {
	SeriesID        string  `json:"seriesId"`
	BlockType       string  `json:"blockType"`
	Date            string  `json:"date"`
	StartsTime      string  `json:"startsTime"`
	DurationMinutes int     `json:"durationMinutes"`
	Reason          *string `json:"reason"`
}

// EffectiveBlocksResponse es la respuesta de
// GET /private/barbers/{barberId}/time-blocks/effective. manualBlocks y
// seriesOccurrences vienen deliberadamente separados: unirlos en una sola
// línea de tiempo exige la zona IANA de la barbería, que ningún consumidor
// necesita todavía (ver schedule.EffectiveBlocksResult).
type EffectiveBlocksResponse struct {
	ManualBlocks      []TimeBlockResponse        `json:"manualBlocks"`
	SeriesOccurrences []SeriesOccurrenceResponse `json:"seriesOccurrences"`
}
