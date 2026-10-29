package booking

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/idempotency"
)

// MaxCancellationReasonLength es el tope de appointment.cancellation_reason y
// appointment_history.reason (CHECK de la base: 500 caracteres).
const MaxCancellationReasonLength = 500

// CustomerCancellationPolicy es la política vigente de cancelación propia del
// cliente (RN-CAN-01, RN-CAN-02, DEC-083), tal como la barbería la configuró
// en HU-093. Se lee DENTRO de la transacción de cancelación, nunca de una
// lectura previa: un cambio de configuración aplica desde que se confirma.
type CustomerCancellationPolicy struct {
	// DeadlineMinutes es el plazo en minutos antes del inicio. 0 desactiva la
	// cancelación dentro de plazo (DEC-083): toda cancelación del cliente se
	// trata como tardía y la decide LateClientAllowed.
	DeadlineMinutes    int
	LateClientAllowed  bool
	LateReasonRequired bool
}

// Evaluate decide si el cliente puede cancelar un turno que empieza en
// startsAt con el reloj del servidor en now (RN-CAN-01: el límite exacto se
// evalúa contra ese reloj). Devuelve nil si procede; si no, el rechazo
// distinguible por código:
//
//   - Con el turno ya iniciado (now >= startsAt) la cancelación propia nunca
//     procede, ni aun con la política tardía permitida: lo que sigue es del
//     barbero (RN-CAN-03, T6/T7).
//   - Hasta startsAt-DeadlineMinutes INCLUSIVE es dentro de plazo: no exige
//     motivo (RN-CAN-02 solo lo regula fuera de plazo) y uno opcional se
//     conserva.
//   - Después del límite y antes del inicio es tardía: si LateClientAllowed
//     es falso se rechaza (el cliente debe contactar a la barbería); si es
//     verdadero y LateReasonRequired también, un motivo vacío o en blanco se
//     rechaza (CA-099-03).
//
// reason ya viene recortado; vacío significa «sin motivo».
func (p CustomerCancellationPolicy) Evaluate(startsAt, now time.Time, reason string) error {
	if !now.Before(startsAt) {
		return errCancellationWindowClosed()
	}
	withinDeadline := p.DeadlineMinutes > 0 &&
		!now.After(startsAt.Add(-time.Duration(p.DeadlineMinutes)*time.Minute))
	if withinDeadline {
		return nil
	}
	if !p.LateClientAllowed {
		return errCancellationWindowClosed()
	}
	if p.LateReasonRequired && reason == "" {
		return errCancellationReasonRequired()
	}
	return nil
}

func errCancellationWindowClosed() error {
	return apperr.CancellationWindowClosed("el plazo para cancelar por tu cuenta ya venció; contacta a la barbería")
}

func errCancellationReasonRequired() error {
	return apperr.CancellationReasonRequired("cuenta brevemente el motivo de la cancelación")
}

func errCancellationReasonTooLong() error {
	return apperr.Validation("el motivo excede el largo máximo")
}

// CancelAppointmentByCustomerRequest es la entrada cruda de T5 (HU-099). El
// llamador (customeraccess) ya resolvió el token a la barbería y la cita; el
// cliente solo aporta la versión que leyó y, si la política lo pide, el motivo.
// Nunca llevan estado, actor ni tiempo: el servidor los deriva.
type CancelAppointmentByCustomerRequest struct {
	AppointmentID        string
	ExpectedVersionToken string
	Reason               string
}

// CancelAppointmentByCustomerInput es la entrada ya validada que llega al
// repositorio. Now es el reloj del servidor leído por el servicio (inyectable
// en pruebas); Reason ya está recortado.
type CancelAppointmentByCustomerInput struct {
	AppointmentID        string
	ExpectedVersionToken string
	Reason               string
	Now                  time.Time
}

// CancelAppointmentByCustomerResult sigue el criterio de
// CancelAppointmentByBarberResult: Decision distingue la aplicación (o su
// repetición ya cancelada) de un Replay exacto o de un conflicto de
// idempotencia, y Response trae el cuerpo almacenado para reproducirlo.
type CancelAppointmentByCustomerResult struct {
	Decision idempotency.Decision
	Response idempotency.StoredResponse
}

// CustomerCancellationRepository persiste T5 (HU-099). Una sola transacción
// tenant-aware con el protocolo de idempotencia (RN-IDE-01, DEC-043): Begin,
// bloquear la fila, releer política y estado con la fila bloqueada, aplicar
// `cancelled_by_customer` + motivo + evento con actor cliente + revocar el
// enlace, y Complete. Con la fila bloqueada:
//   - ya `cancelled_by_customer` -> no-op exitoso sin historial nuevo;
//   - `confirmed` -> compara versión (VersionConflict), evalúa la política
//     (rechazos 422) y cancela;
//   - cualquier otro estado -> InvalidState (409) sin tocar nada.
type CustomerCancellationRepository interface {
	CancelByCustomer(
		ctx context.Context,
		barbershopID string,
		input CancelAppointmentByCustomerInput,
		key idempotency.Key,
		fingerprint idempotency.Fingerprint,
	) (CancelAppointmentByCustomerResult, error)
}

// CancelAppointmentByCustomerService implementa T5 (HU-099).
type CancelAppointmentByCustomerService struct {
	repo CustomerCancellationRepository
	now  func() time.Time
}

// NewCancelAppointmentByCustomerService construye el caso de uso. now es el
// reloj del servidor; nil usa time.Now.
func NewCancelAppointmentByCustomerService(repo CustomerCancellationRepository, now func() time.Time) *CancelAppointmentByCustomerService {
	if now == nil {
		now = time.Now
	}
	return &CancelAppointmentByCustomerService{repo: repo, now: now}
}

// CancelAppointmentByCustomer valida la forma de la entrada y delega en el
// repositorio. barbershopID y AppointmentID llegan resueltos desde el token
// por customeraccess; aquí nunca se aceptan del cliente.
func (s *CancelAppointmentByCustomerService) CancelAppointmentByCustomer(
	ctx context.Context,
	barbershopID string,
	req CancelAppointmentByCustomerRequest,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (CancelAppointmentByCustomerResult, error) {
	if err := ctx.Err(); err != nil {
		return CancelAppointmentByCustomerResult{}, apperr.Internal(fmt.Errorf("booking: contexto cancelado antes de cancelar el turno: %w", err))
	}

	appointmentID := strings.TrimSpace(req.AppointmentID)
	if !LooksLikeAppointmentID(appointmentID) {
		return CancelAppointmentByCustomerResult{}, errAppointmentNotFound()
	}
	expectedVersionToken := strings.TrimSpace(req.ExpectedVersionToken)
	if expectedVersionToken == "" {
		return CancelAppointmentByCustomerResult{}, errVersionTokenRequired()
	}
	reason := strings.TrimSpace(req.Reason)
	if utf8.RuneCountInString(reason) > MaxCancellationReasonLength {
		return CancelAppointmentByCustomerResult{}, errCancellationReasonTooLong()
	}

	return s.repo.CancelByCustomer(ctx, barbershopID, CancelAppointmentByCustomerInput{
		AppointmentID:        appointmentID,
		ExpectedVersionToken: expectedVersionToken,
		Reason:               reason,
		Now:                  s.now(),
	}, key, fingerprint)
}
