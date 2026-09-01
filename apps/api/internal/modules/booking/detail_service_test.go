package booking_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"system-barbershop/internal/modules/booking"
	"system-barbershop/internal/platform/apperr"
)

const validAppointmentID = "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4"

// fakeBarberNamePort es un doble de booking.BarberNamePort para probar
// DetailService en aislamiento, sin PostgreSQL ni el módulo staff real.
type fakeBarberNamePort struct {
	name  string
	found bool
	err   error

	lastBarbershopID string
	lastBarberID     string
}

func (f *fakeBarberNamePort) Name(_ context.Context, barbershopID, barberID string) (string, bool, error) {
	f.lastBarbershopID = barbershopID
	f.lastBarberID = barberID
	if f.err != nil {
		return "", false, f.err
	}
	return f.name, f.found, nil
}

// fakeStaffActorNamePort es un doble de booking.StaffActorNamePort.
type fakeStaffActorNamePort struct {
	names map[string]string
	err   error

	lastBarbershopID string
	lastIDs          []string
}

func (f *fakeStaffActorNamePort) Names(_ context.Context, barbershopID string, ids []string) (map[string]string, error) {
	f.lastBarbershopID = barbershopID
	f.lastIDs = ids
	if f.err != nil {
		return nil, f.err
	}
	if f.names == nil {
		return map[string]string{}, nil
	}
	return f.names, nil
}

// fakeDetailRepository es un doble de booking.Repository que solo implementa
// GetAppointmentDetail/ListAppointmentHistory/CustomerNames: las demás
// operaciones nunca deben llamarse desde DetailService (heredadas de
// fakeManualRepository, que entra en panic si se invocan).
type fakeDetailRepository struct {
	fakeManualRepository

	detail                  booking.AppointmentDetail
	detailFound             bool
	detailErr               error
	lastDetailAppointmentID string

	historyItems []booking.HistoryRow
	historyNext  *booking.HistoryCursor
	historyFound bool
	historyErr   error

	lastHistoryAppointmentID string
	lastHistoryCursor        *booking.HistoryCursor
	lastHistoryLimit         int

	customerNames    map[string]string
	customerNamesErr error
	lastCustomerIDs  []string
}

func (f *fakeDetailRepository) GetAppointmentDetail(
	_ context.Context, _, appointmentID string,
) (booking.AppointmentDetail, bool, error) {
	f.lastDetailAppointmentID = appointmentID
	if f.detailErr != nil {
		return booking.AppointmentDetail{}, false, f.detailErr
	}
	return f.detail, f.detailFound, nil
}

func (f *fakeDetailRepository) ListAppointmentHistory(
	_ context.Context, _, appointmentID string, cursor *booking.HistoryCursor, limit int,
) ([]booking.HistoryRow, *booking.HistoryCursor, bool, error) {
	f.lastHistoryAppointmentID = appointmentID
	f.lastHistoryCursor = cursor
	f.lastHistoryLimit = limit
	if f.historyErr != nil {
		return nil, nil, false, f.historyErr
	}
	return f.historyItems, f.historyNext, f.historyFound, nil
}

func (f *fakeDetailRepository) CustomerNames(_ context.Context, _ string, ids []string) (map[string]string, error) {
	f.lastCustomerIDs = ids
	if f.customerNamesErr != nil {
		return nil, f.customerNamesErr
	}
	if f.customerNames == nil {
		return map[string]string{}, nil
	}
	return f.customerNames, nil
}

// ---------------------------------------------------------------------------
// GetDetail
// ---------------------------------------------------------------------------

func TestGetDetail_MalformedAppointmentID_RejectsWithoutTouchingRepository(t *testing.T) {
	repo := &fakeDetailRepository{}
	svc := booking.NewDetailService(repo, &fakeBarberNamePort{}, &fakeStaffActorNamePort{})

	_, err := svc.GetDetail(context.Background(), "shop-1", "not-a-uuid")
	assertKind(t, err, apperr.KindNotFound)
	if repo.lastDetailAppointmentID != "" {
		t.Fatalf("no debió tocar el repositorio con un appointmentId malformado")
	}
}

func TestGetDetail_NotFound_ReturnsNotFound(t *testing.T) {
	repo := &fakeDetailRepository{detailFound: false}
	svc := booking.NewDetailService(repo, &fakeBarberNamePort{}, &fakeStaffActorNamePort{})

	_, err := svc.GetDetail(context.Background(), "shop-1", validAppointmentID)
	assertKind(t, err, apperr.KindNotFound)
}

func TestGetDetail_RepositoryError_ReturnsInternal(t *testing.T) {
	repo := &fakeDetailRepository{detailErr: errors.New("boom")}
	svc := booking.NewDetailService(repo, &fakeBarberNamePort{}, &fakeStaffActorNamePort{})

	_, err := svc.GetDetail(context.Background(), "shop-1", validAppointmentID)
	assertKind(t, err, apperr.KindInternal)
}

func TestGetDetail_Found_ResolvesBarberFullName(t *testing.T) {
	repo := &fakeDetailRepository{
		detailFound: true,
		detail:      booking.AppointmentDetail{ID: validAppointmentID, BarberID: "barber-1", AttendeeName: "Carlos"},
	}
	barbers := &fakeBarberNamePort{name: "Ana Gómez", found: true}
	svc := booking.NewDetailService(repo, barbers, &fakeStaffActorNamePort{})

	got, err := svc.GetDetail(context.Background(), "shop-1", validAppointmentID)
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}
	if got.BarberFullName != "Ana Gómez" {
		t.Fatalf("BarberFullName = %q, want %q", got.BarberFullName, "Ana Gómez")
	}
	if barbers.lastBarbershopID != "shop-1" || barbers.lastBarberID != "barber-1" {
		t.Fatalf("BarberNamePort llamado con tenant/barbero inesperados: %q/%q", barbers.lastBarbershopID, barbers.lastBarberID)
	}
}

// TestGetDetail_BarberNameNotFound_LeavesFullNameEmpty cubre la etiqueta de
// reserva de la sección 2.3 del prompt: en la práctica no debería ocurrir
// (appointment_barbershop_id_barber_id_fk es ON DELETE RESTRICT), pero el
// caso de uso no debe fallar de forma insegura si el puerto no encuentra el
// barbero.
func TestGetDetail_BarberNameNotFound_LeavesFullNameEmpty(t *testing.T) {
	repo := &fakeDetailRepository{
		detailFound: true,
		detail:      booking.AppointmentDetail{ID: validAppointmentID, BarberID: "barber-1"},
	}
	svc := booking.NewDetailService(repo, &fakeBarberNamePort{found: false}, &fakeStaffActorNamePort{})

	got, err := svc.GetDetail(context.Background(), "shop-1", validAppointmentID)
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}
	if got.BarberFullName != "" {
		t.Fatalf("BarberFullName = %q, want vacío", got.BarberFullName)
	}
}

func TestGetDetail_BarberNamePortError_ReturnsInternal(t *testing.T) {
	repo := &fakeDetailRepository{
		detailFound: true,
		detail:      booking.AppointmentDetail{ID: validAppointmentID, BarberID: "barber-1"},
	}
	svc := booking.NewDetailService(repo, &fakeBarberNamePort{err: errors.New("boom")}, &fakeStaffActorNamePort{})

	_, err := svc.GetDetail(context.Background(), "shop-1", validAppointmentID)
	assertKind(t, err, apperr.KindInternal)
}

// ---------------------------------------------------------------------------
// ListHistory
// ---------------------------------------------------------------------------

func TestListHistory_MalformedAppointmentID_RejectsWithoutTouchingRepository(t *testing.T) {
	repo := &fakeDetailRepository{}
	svc := booking.NewDetailService(repo, &fakeBarberNamePort{}, &fakeStaffActorNamePort{})

	_, err := svc.ListHistory(context.Background(), "shop-1", "not-a-uuid", "", 0)
	assertKind(t, err, apperr.KindNotFound)
	if repo.lastHistoryAppointmentID != "" {
		t.Fatalf("no debió tocar el repositorio con un appointmentId malformado")
	}
}

func TestListHistory_InvalidCursor_ReturnsInvalidWithoutTouchingRepository(t *testing.T) {
	repo := &fakeDetailRepository{}
	svc := booking.NewDetailService(repo, &fakeBarberNamePort{}, &fakeStaffActorNamePort{})

	_, err := svc.ListHistory(context.Background(), "shop-1", validAppointmentID, "no-es-base64url-válido-!!", 0)
	assertKind(t, err, apperr.KindInvalid)
	if repo.lastHistoryAppointmentID != "" {
		t.Fatalf("no debió tocar el repositorio con un cursor inválido")
	}
}

func TestListHistory_AppointmentNotFound_ReturnsNotFound(t *testing.T) {
	repo := &fakeDetailRepository{historyFound: false}
	svc := booking.NewDetailService(repo, &fakeBarberNamePort{}, &fakeStaffActorNamePort{})

	_, err := svc.ListHistory(context.Background(), "shop-1", validAppointmentID, "", 0)
	assertKind(t, err, apperr.KindNotFound)
}

func TestListHistory_RepositoryError_ReturnsInternal(t *testing.T) {
	repo := &fakeDetailRepository{historyErr: errors.New("boom")}
	svc := booking.NewDetailService(repo, &fakeBarberNamePort{}, &fakeStaffActorNamePort{})

	_, err := svc.ListHistory(context.Background(), "shop-1", validAppointmentID, "", 0)
	assertKind(t, err, apperr.KindInternal)
}

func TestListHistory_LimitNotPositive_UsesDefaultLimit(t *testing.T) {
	repo := &fakeDetailRepository{historyFound: true}
	svc := booking.NewDetailService(repo, &fakeBarberNamePort{}, &fakeStaffActorNamePort{})

	if _, err := svc.ListHistory(context.Background(), "shop-1", validAppointmentID, "", 0); err != nil {
		t.Fatalf("err inesperado: %v", err)
	}
	if repo.lastHistoryLimit != booking.HistoryDefaultLimit {
		t.Fatalf("limit = %d, want %d", repo.lastHistoryLimit, booking.HistoryDefaultLimit)
	}
}

func TestListHistory_LimitAboveMax_ClampsToMax(t *testing.T) {
	repo := &fakeDetailRepository{historyFound: true}
	svc := booking.NewDetailService(repo, &fakeBarberNamePort{}, &fakeStaffActorNamePort{})

	if _, err := svc.ListHistory(context.Background(), "shop-1", validAppointmentID, "", 10_000); err != nil {
		t.Fatalf("err inesperado: %v", err)
	}
	if repo.lastHistoryLimit != booking.HistoryMaxLimit {
		t.Fatalf("limit = %d, want %d", repo.lastHistoryLimit, booking.HistoryMaxLimit)
	}
}

func TestListHistory_ValidCursorToken_DecodesAndForwardsToRepository(t *testing.T) {
	occurredAt := time.Date(2026, 8, 30, 10, 0, 0, 0, time.UTC)
	cursor := booking.HistoryCursor{OccurredAt: occurredAt, ID: "history-1"}
	token := booking.EncodeHistoryCursor(cursor)

	repo := &fakeDetailRepository{historyFound: true}
	svc := booking.NewDetailService(repo, &fakeBarberNamePort{}, &fakeStaffActorNamePort{})

	if _, err := svc.ListHistory(context.Background(), "shop-1", validAppointmentID, token, 5); err != nil {
		t.Fatalf("err inesperado: %v", err)
	}
	if repo.lastHistoryCursor == nil {
		t.Fatalf("no reenvió el cursor decodificado al repositorio")
	}
	if repo.lastHistoryCursor.ID != "history-1" || !repo.lastHistoryCursor.OccurredAt.Equal(occurredAt) {
		t.Fatalf("cursor decodificado = %+v, want %+v", repo.lastHistoryCursor, cursor)
	}
}

func TestListHistory_NextCursorFromRepository_IsEncodedInThePage(t *testing.T) {
	next := &booking.HistoryCursor{OccurredAt: time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC), ID: "history-2"}
	repo := &fakeDetailRepository{historyFound: true, historyNext: next}
	svc := booking.NewDetailService(repo, &fakeBarberNamePort{}, &fakeStaffActorNamePort{})

	page, err := svc.ListHistory(context.Background(), "shop-1", validAppointmentID, "", 0)
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}
	if page.NextCursor == "" {
		t.Fatalf("NextCursor vacío, want el cursor codificado de %+v", next)
	}
	decoded, err := booking.DecodeHistoryCursor(page.NextCursor)
	if err != nil {
		t.Fatalf("NextCursor no decodifica: %v", err)
	}
	if decoded.ID != next.ID {
		t.Fatalf("NextCursor decodificado = %+v, want ID %q", decoded, next.ID)
	}
}

func TestListHistory_NoNextCursorFromRepository_PageIsLast(t *testing.T) {
	repo := &fakeDetailRepository{historyFound: true, historyNext: nil}
	svc := booking.NewDetailService(repo, &fakeBarberNamePort{}, &fakeStaffActorNamePort{})

	page, err := svc.ListHistory(context.Background(), "shop-1", validAppointmentID, "", 0)
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}
	if page.NextCursor != "" {
		t.Fatalf("NextCursor = %q, want vacío (última página)", page.NextCursor)
	}
}

// TestListHistory_ResolvesActorLabels_InOneBatchPerKind cubre el trabajo
// requerido §2.3 ("sin generar N+1"): dos filas con el mismo actor staff
// deben producir UNA sola llamada a StaffActorNamePort.Names con un único id
// deduplicado, nunca una por fila.
func TestListHistory_ResolvesActorLabels_InOneBatchPerKind(t *testing.T) {
	staffID := "staff-1"
	customerID := "customer-1"
	repo := &fakeDetailRepository{
		historyFound: true,
		historyItems: []booking.HistoryRow{
			{
				ID: "history-1", EventType: booking.EventTypeAppointmentCreated,
				ActorType: booking.ActorTypeStaff, ActorStaffUserID: &staffID,
				OccurredAt: time.Date(2026, 8, 30, 9, 0, 0, 0, time.UTC),
			},
			{
				ID: "history-2", EventType: booking.EventTypeAppointmentRescheduled,
				ActorType: booking.ActorTypeStaff, ActorStaffUserID: &staffID,
				OccurredAt: time.Date(2026, 8, 30, 10, 0, 0, 0, time.UTC),
			},
			{
				ID: "history-3", EventType: booking.EventTypeAppointmentCancelledByCustomer,
				ActorType: booking.ActorTypeCustomer, ActorCustomerID: &customerID,
				OccurredAt: time.Date(2026, 8, 30, 11, 0, 0, 0, time.UTC),
			},
			{
				ID: "history-4", EventType: booking.EventTypeAppointmentNoShow,
				ActorType:  booking.ActorTypeSystem,
				OccurredAt: time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC),
			},
		},
	}
	staffActors := &fakeStaffActorNamePort{names: map[string]string{staffID: "Ana Gómez"}}
	repo.customerNames = map[string]string{customerID: "Carlos Restrepo"}
	svc := booking.NewDetailService(repo, &fakeBarberNamePort{}, staffActors)

	page, err := svc.ListHistory(context.Background(), "shop-1", validAppointmentID, "", 0)
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}
	if len(page.Items) != 4 {
		t.Fatalf("len(items) = %d, want 4", len(page.Items))
	}
	if page.Items[0].ActorLabel != "Ana Gómez" || page.Items[1].ActorLabel != "Ana Gómez" {
		t.Fatalf("etiquetas staff = %q/%q, want ambas 'Ana Gómez'", page.Items[0].ActorLabel, page.Items[1].ActorLabel)
	}
	if page.Items[2].ActorLabel != "Carlos Restrepo" {
		t.Fatalf("etiqueta customer = %q, want 'Carlos Restrepo'", page.Items[2].ActorLabel)
	}
	if page.Items[3].ActorLabel != "Sistema" {
		t.Fatalf("etiqueta system = %q, want 'Sistema'", page.Items[3].ActorLabel)
	}

	if len(staffActors.lastIDs) != 1 || staffActors.lastIDs[0] != staffID {
		t.Fatalf("StaffActorNamePort.Names llamado con %v, want un único id deduplicado [%q]", staffActors.lastIDs, staffID)
	}
	if len(repo.lastCustomerIDs) != 1 || repo.lastCustomerIDs[0] != customerID {
		t.Fatalf("CustomerNames llamado con %v, want un único id deduplicado [%q]", repo.lastCustomerIDs, customerID)
	}
}

// TestListHistory_MissingAuthorizedName_UsesSafeLabel cubre "si falta un
// nombre autorizado, usa una etiqueta segura coherente con el tipo de
// actor" (trabajo requerido §2.3).
func TestListHistory_MissingAuthorizedName_UsesSafeLabel(t *testing.T) {
	staffID := "staff-ausente"
	repo := &fakeDetailRepository{
		historyFound: true,
		historyItems: []booking.HistoryRow{
			{
				ID: "history-1", EventType: booking.EventTypeAppointmentCreated,
				ActorType: booking.ActorTypeStaff, ActorStaffUserID: &staffID,
				OccurredAt: time.Date(2026, 8, 30, 9, 0, 0, 0, time.UTC),
			},
		},
	}
	svc := booking.NewDetailService(repo, &fakeBarberNamePort{}, &fakeStaffActorNamePort{})

	page, err := svc.ListHistory(context.Background(), "shop-1", validAppointmentID, "", 0)
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}
	if page.Items[0].ActorLabel != "Miembro del equipo" {
		t.Fatalf("ActorLabel = %q, want la etiqueta segura de reserva", page.Items[0].ActorLabel)
	}
}

func TestListHistory_ChangesArePreservedPerEntry(t *testing.T) {
	prev, next := "confirmed", "cancelled_by_barber"
	repo := &fakeDetailRepository{
		historyFound: true,
		historyItems: []booking.HistoryRow{
			{
				ID: "history-1", EventType: booking.EventTypeAppointmentStatusCorrected,
				ActorType:  booking.ActorTypeSystem,
				OccurredAt: time.Date(2026, 8, 30, 9, 0, 0, 0, time.UTC),
				Changes:    []booking.HistoryChange{{FieldName: "status", PreviousValue: &prev, NewValue: &next}},
			},
		},
	}
	svc := booking.NewDetailService(repo, &fakeBarberNamePort{}, &fakeStaffActorNamePort{})

	page, err := svc.ListHistory(context.Background(), "shop-1", validAppointmentID, "", 0)
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}
	if len(page.Items[0].Changes) != 1 || page.Items[0].Changes[0].FieldName != "status" {
		t.Fatalf("Changes = %+v, want un cambio de status", page.Items[0].Changes)
	}
}

func TestListHistory_StaffActorNamePortError_ReturnsInternal(t *testing.T) {
	staffID := "staff-1"
	repo := &fakeDetailRepository{
		historyFound: true,
		historyItems: []booking.HistoryRow{
			{ID: "history-1", ActorType: booking.ActorTypeStaff, ActorStaffUserID: &staffID, OccurredAt: time.Now()},
		},
	}
	svc := booking.NewDetailService(repo, &fakeBarberNamePort{}, &fakeStaffActorNamePort{err: errors.New("boom")})

	_, err := svc.ListHistory(context.Background(), "shop-1", validAppointmentID, "", 0)
	assertKind(t, err, apperr.KindInternal)
}

func TestListHistory_CustomerNamesError_ReturnsInternal(t *testing.T) {
	customerID := "customer-1"
	repo := &fakeDetailRepository{
		historyFound: true,
		historyItems: []booking.HistoryRow{
			{ID: "history-1", ActorType: booking.ActorTypeCustomer, ActorCustomerID: &customerID, OccurredAt: time.Now()},
		},
		customerNamesErr: errors.New("boom"),
	}
	svc := booking.NewDetailService(repo, &fakeBarberNamePort{}, &fakeStaffActorNamePort{})

	_, err := svc.ListHistory(context.Background(), "shop-1", validAppointmentID, "", 0)
	assertKind(t, err, apperr.KindInternal)
}
