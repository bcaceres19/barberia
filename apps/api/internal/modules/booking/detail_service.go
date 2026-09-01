package booking

import (
	"context"
	"fmt"
	"strings"

	"system-barbershop/internal/platform/apperr"
)

// DetailService implementa el caso de uso de lectura de HU-064: detalle de
// una cita y su historial inmutable paginado. Colabora con staff/auth
// EXCLUSIVAMENTE mediante BarberNamePort/StaffActorNamePort (mismo criterio
// que AgendaService frente a BarberPort/TimezonePort): ni importa staff, ni
// auth.
type DetailService struct {
	repo        Repository
	barbers     BarberNamePort
	staffActors StaffActorNamePort
}

// NewDetailService construye el caso de uso a partir de sus tres
// colaboradores.
func NewDetailService(repo Repository, barbers BarberNamePort, staffActors StaffActorNamePort) *DetailService {
	return &DetailService{repo: repo, barbers: barbers, staffActors: staffActors}
}

// GetDetail devuelve el detalle de appointmentID dentro de barbershopID. Un
// appointmentID con forma inválida, inexistente o de otra barbería responde
// el mismo apperr.NotFound (RN-TEN-01, CA-064-01, CA-064-07), verificado
// ANTES de tocar el repositorio (mismo criterio que
// AgendaService.ListDailyAgenda frente a barberID).
func (s *DetailService) GetDetail(ctx context.Context, barbershopID, appointmentID string) (AppointmentDetail, error) {
	if err := ctx.Err(); err != nil {
		return AppointmentDetail{}, apperr.Internal(fmt.Errorf("booking: contexto cancelado antes de leer el detalle de la cita: %w", err))
	}

	appointmentID = strings.TrimSpace(appointmentID)
	if !LooksLikeAppointmentID(appointmentID) {
		return AppointmentDetail{}, errAppointmentNotFound()
	}

	detail, found, err := s.repo.GetAppointmentDetail(ctx, barbershopID, appointmentID)
	if err != nil {
		return AppointmentDetail{}, apperr.Internal(fmt.Errorf("booking: leer el detalle de la cita: %w", err))
	}
	if !found {
		return AppointmentDetail{}, errAppointmentNotFound()
	}

	name, foundBarber, err := s.barbers.Name(ctx, barbershopID, detail.BarberID)
	if err != nil {
		return AppointmentDetail{}, apperr.Internal(fmt.Errorf("booking: resolver el nombre del barbero: %w", err))
	}
	if foundBarber {
		detail.BarberFullName = name
	}
	return detail, nil
}

// ListHistory devuelve una página del historial de appointmentID, ordenada
// por instante y luego por identificador (CA-064-05). Un appointmentID con
// forma inválida, inexistente o de otra barbería responde el mismo
// apperr.NotFound que GetDetail. cursorToken vacío pide la primera página;
// limit no positivo usa HistoryDefaultLimit (mismo criterio que
// staff.Service.List).
func (s *DetailService) ListHistory(
	ctx context.Context,
	barbershopID, appointmentID, cursorToken string,
	limit int,
) (HistoryPage, error) {
	if err := ctx.Err(); err != nil {
		return HistoryPage{}, apperr.Internal(fmt.Errorf("booking: contexto cancelado antes de listar el historial: %w", err))
	}

	appointmentID = strings.TrimSpace(appointmentID)
	if !LooksLikeAppointmentID(appointmentID) {
		return HistoryPage{}, errAppointmentNotFound()
	}

	var cursor *HistoryCursor
	if strings.TrimSpace(cursorToken) != "" {
		c, err := DecodeHistoryCursor(cursorToken)
		if err != nil {
			return HistoryPage{}, err
		}
		cursor = &c
	}

	rows, next, found, err := s.repo.ListAppointmentHistory(ctx, barbershopID, appointmentID, cursor, clampHistoryLimit(limit))
	if err != nil {
		return HistoryPage{}, apperr.Internal(fmt.Errorf("booking: listar el historial de la cita: %w", err))
	}
	if !found {
		return HistoryPage{}, errAppointmentNotFound()
	}

	staffNames, err := s.staffActors.Names(ctx, barbershopID, uniqueStaffActorIDs(rows))
	if err != nil {
		return HistoryPage{}, apperr.Internal(fmt.Errorf("booking: resolver nombres de actores staff: %w", err))
	}
	customerNames, err := s.repo.CustomerNames(ctx, barbershopID, uniqueCustomerActorIDs(rows))
	if err != nil {
		return HistoryPage{}, apperr.Internal(fmt.Errorf("booking: resolver nombres de actores cliente: %w", err))
	}

	items := make([]HistoryEntry, 0, len(rows))
	for _, row := range rows {
		items = append(items, HistoryEntry{
			ID:         row.ID,
			EventType:  row.EventType,
			ActorType:  row.ActorType,
			ActorLabel: resolveActorLabel(row, staffNames, customerNames),
			Reason:     row.Reason,
			OccurredAt: row.OccurredAt,
			Changes:    row.Changes,
		})
	}

	nextToken := ""
	if next != nil {
		nextToken = EncodeHistoryCursor(*next)
	}
	return HistoryPage{Items: items, NextCursor: nextToken}, nil
}

// uniqueStaffActorIDs/uniqueCustomerActorIDs recogen, sin duplicados, los
// identificadores de actor de una página completa: exactamente el lote que
// StaffActorNamePort.Names/Repository.CustomerNames necesitan para resolver
// todos los nombres de la página en una sola llamada cada uno (trabajo
// requerido §2.3, "sin generar N+1").
func uniqueStaffActorIDs(rows []HistoryRow) []string {
	seen := make(map[string]struct{})
	var ids []string
	for _, row := range rows {
		if row.ActorStaffUserID == nil {
			continue
		}
		if _, ok := seen[*row.ActorStaffUserID]; ok {
			continue
		}
		seen[*row.ActorStaffUserID] = struct{}{}
		ids = append(ids, *row.ActorStaffUserID)
	}
	return ids
}

func uniqueCustomerActorIDs(rows []HistoryRow) []string {
	seen := make(map[string]struct{})
	var ids []string
	for _, row := range rows {
		if row.ActorCustomerID == nil {
			continue
		}
		if _, ok := seen[*row.ActorCustomerID]; ok {
			continue
		}
		seen[*row.ActorCustomerID] = struct{}{}
		ids = append(ids, *row.ActorCustomerID)
	}
	return ids
}
