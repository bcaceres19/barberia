package booking

import (
	"context"
	"time"

	"system-barbershop/internal/platform/idempotency"
)

// Repository es el puerto de persistencia del módulo booking. El núcleo no
// importa internal/platform/database ni pgx (CA-002-06): postgres/ traduce
// entre este contrato y database.DB, mismo patrón que catalog.Repository.
type Repository interface {
	// CreateInternal ejecuta, dentro de UNA sola transacción tenant-aware,
	// la creación o vinculación del cliente que input.Customer ya decidió,
	// el INSERT de la cita `confirmed` y el INSERT del evento
	// appointment_created (§2.3 del prompt de HU-060): un fallo en
	// cualquier paso revierte los tres. input ya llegó validado por
	// BookingService.
	CreateInternal(ctx context.Context, barbershopID string, input CreateInternalInput) (CreateInternalResult, error)

	// FindCustomerForReconciliation busca, dentro de barbershopID, un
	// cliente ya persistido reutilizable según DEC-071: por teléfono si
	// phone no es nil (DEC-045), si no por correo si email no es nil
	// (DEC-046). found=false cuando no hay coincidencia o cuando ambos
	// llegan nil (el llamador entonces construye un cliente nuevo sin
	// llamar aquí). ManualBookingService la usa ANTES de decidir
	// CustomerInput; CreateManual la usa de nuevo dentro de su propia
	// transacción para acotar la ventana de una carrera entre dos altas
	// concurrentes con el mismo contacto.
	FindCustomerForReconciliation(ctx context.Context, barbershopID string, phone, email *string) (Customer, bool, error)

	// CreateManual ejecuta, dentro de UNA sola transacción tenant-aware
	// protegida por el protocolo de idempotencia reutilizable de HU-004
	// (RN-IDE-01, DEC-043): Begin, la reconciliación de cliente de
	// DEC-071 (repetida aquí para acotar la carrera entre el Find previo
	// de ManualBookingService y este INSERT), el INSERT de la cita
	// `confirmed`, el INSERT del evento appointment_created y Complete.
	// input.Customer ya llegó con la decisión de ManualBookingService
	// (ExistingID si el Find previo encontró coincidencia, New si no);
	// este método repite el Find solo cuando ExistingID es nil, nunca
	// vuelve a decidir con un identificador que el llamador ya resolvió.
	CreateManual(
		ctx context.Context,
		barbershopID string,
		input CreateInternalInput,
		key idempotency.Key,
		fingerprint idempotency.Fingerprint,
	) (CreateManualResult, error)

	// ListDailyAgenda lee, dentro de barbershopID y del único barberID
	// pedido (DEC-074), las citas cuyo intervalo [starts_at, ends_at)
	// interseca [rangeStart, rangeEnd) (DEC-075: rangeStart/rangeEnd ya
	// llegan como los dos instantes civiles resueltos por AgendaService en
	// la zona de la barbería, nunca calculados aquí). Ordenadas por
	// starts_at y luego por id para un orden estable. barberID ya fue
	// verificado por AgendaService antes de llamar aquí (mismo criterio que
	// schedule.Repository.List frente a schedule.Service.List).
	ListDailyAgenda(ctx context.Context, barbershopID, barberID string, rangeStart, rangeEnd time.Time) ([]DailyAgendaEntry, error)

	// GetAppointmentDetail lee el detalle de una cita dentro de barbershopID
	// (HU-064, CA-064-01 a CA-064-04). found=false cubre appointmentID
	// inexistente o de otra barbería, sin distinguir la causa (RN-TEN-01);
	// en ese caso AppointmentDetail queda en su valor cero.
	// AppointmentDetail.BarberFullName SIEMPRE llega vacío: se resuelve
	// aparte, vía BarberNamePort, porque este repositorio no puede unir
	// contra `barber` (CA-002-06).
	GetAppointmentDetail(ctx context.Context, barbershopID, appointmentID string) (AppointmentDetail, bool, error)

	// ListAppointmentHistory lee una página del historial de appointmentID,
	// orden estable (occurred_at, id) (HU-064, CA-064-05). cursor es nil
	// para la primera página; limit ya llegó clamped por DetailService.
	// found=false cubre appointmentID inexistente o de otra barbería (mismo
	// criterio que GetAppointmentDetail); en ese caso items/nextCursor
	// quedan en su valor cero. nextCursor es nil cuando esta página es la
	// última. Cada HistoryRow.Changes ya viene resuelto (una sola consulta
	// adicional por página, nunca una por fila).
	ListAppointmentHistory(
		ctx context.Context,
		barbershopID, appointmentID string,
		cursor *HistoryCursor,
		limit int,
	) (items []HistoryRow, nextCursor *HistoryCursor, found bool, err error)

	// CustomerNames resuelve, en un solo lote, el nombre visible de cada
	// customerID pedido dentro de barbershopID (trabajo requerido §2.3, "sin
	// generar N+1"): customer es tabla propia de booking (HU-060), así que
	// esta lectura no necesita ningún puerto adicional. Un id sin
	// coincidencia simplemente está ausente del mapa devuelto.
	CustomerNames(ctx context.Context, barbershopID string, customerIDs []string) (map[string]string, error)

	// Reschedule ejecuta T2 (HU-065) dentro de UNA sola transacción
	// tenant-aware protegida por el protocolo de idempotencia reutilizable
	// de HU-004 (RN-IDE-01, DEC-043): Begin, bloquear la fila (`FOR
	// UPDATE`), verificar de nuevo estado/versión con la fila ya
	// bloqueada (nunca confiar en la lectura previa de
	// RescheduleService), aplicar el nuevo intervalo (o detectar el no-op
	// del mismo intervalo, sin tocar `updated_at` ni el historial),
	// insertar el evento `appointment_rescheduled` con sus cambios
	// anterior/nuevo, y Complete. input.NewEndsAt ya llegó derivado de
	// duration_minutes_snapshot; este método nunca la recalcula.
	Reschedule(
		ctx context.Context,
		barbershopID string,
		input RescheduleInput,
		key idempotency.Key,
		fingerprint idempotency.Fingerprint,
	) (RescheduleResult, error)

	// CancelByBarber ejecuta T6 (HU-066) dentro de UNA sola transacción
	// tenant-aware protegida por el protocolo de idempotencia reutilizable
	// de HU-004 (RN-IDE-01, DEC-043): Begin, bloquear la fila (`FOR
	// UPDATE`), verificar de nuevo con la fila ya bloqueada (nunca confiar
	// en una lectura previa), aplicar `cancelled_by_barber` + insertar
	// `appointment_cancelled_by_barber`, y Complete. Semántica exacta de
	// los tres desenlaces posibles con la fila ya bloqueada
	// (CA-066-01/04/05):
	//   - ya está `cancelled_by_barber` -> no-op exitoso (sin comparar
	//     versionToken: una cancelación repetida por CUALQUIER barbero de
	//     la barbería siempre tiene éxito, aunque el token que el cliente
	//     conserva haya quedado obsoleto tras la primera cancelación), sin
	//     UPDATE ni historial nuevo.
	//   - `confirmed` -> compara versionToken (mismatch = VersionConflict),
	//     luego UPDATE status/resolved_at + INSERT historial.
	//   - cualquier otro estado (`completed`, `no_show`,
	//     `cancelled_by_customer`) -> InvalidState (409), sin tocar nada.
	CancelByBarber(
		ctx context.Context,
		barbershopID string,
		input CancelAppointmentByBarberInput,
		key idempotency.Key,
		fingerprint idempotency.Fingerprint,
	) (CancelAppointmentByBarberResult, error)

	// CompleteAppointment ejecuta T4 manual (HU-067) dentro de UNA sola
	// transacción tenant-aware protegida por el protocolo de idempotencia
	// reutilizable de HU-004 (RN-IDE-01, DEC-043): Begin, bloquear la fila
	// (`FOR UPDATE`), verificar de nuevo con la fila ya bloqueada (nunca
	// confiar en una lectura previa), aplicar `completed` + insertar
	// `appointment_completed`, y Complete. Semántica exacta de los
	// desenlaces posibles con la fila ya bloqueada (CA-067-01/03/05):
	//   - ya está `completed` -> no-op exitoso, sin UPDATE ni historial
	//     nuevo.
	//   - `confirmed` con starts_at > input.Now -> Validation (422), sin
	//     tocar nada.
	//   - `confirmed` con starts_at <= input.Now -> compara versionToken
	//     (mismatch = VersionConflict), luego UPDATE status/resolved_at +
	//     INSERT historial.
	//   - cualquier otro estado (`no_show`, `cancelled_by_barber`,
	//     `cancelled_by_customer`) -> InvalidState (409).
	CompleteAppointment(
		ctx context.Context,
		barbershopID string,
		input CloseAppointmentInput,
		key idempotency.Key,
		fingerprint idempotency.Fingerprint,
	) (CompleteAppointmentResult, error)

	// MarkNoShow ejecuta T7 (HU-067) con la misma estructura y el mismo
	// protocolo de idempotencia/bloqueo que CompleteAppointment, aplicando
	// `no_show` + `appointment_no_show` en vez de `completed` +
	// `appointment_completed`. Repetir sobre una cita ya `no_show` es el
	// no-op de este comando; una cita ya `completed` cae en InvalidState
	// (409), mismo criterio simétrico.
	MarkNoShow(
		ctx context.Context,
		barbershopID string,
		input CloseAppointmentInput,
		key idempotency.Key,
		fingerprint idempotency.Fingerprint,
	) (MarkNoShowResult, error)
}

// BarberNamePort resuelve el nombre visible de un barbero por id, dentro de
// barbershopID (RN-TEN-01). found=false cubre inexistente o de otra
// barbería, mismo criterio que BarberPort.Exists. El adaptador real vive en
// el paquete staff (staff.NewBarberNameLookup): booking nunca importa staff.
type BarberNamePort interface {
	Name(ctx context.Context, barbershopID, barberID string) (fullName string, found bool, err error)
}

// StaffActorNamePort resuelve, en un solo lote, el nombre visible de cada
// staff_user_id pedido (RN-HIS-01, HU-064): nunca expone correo ni ningún
// identificador interno. Un id sin coincidencia simplemente está ausente del
// mapa devuelto; DetailService decide entonces la etiqueta segura de
// reserva. El adaptador real vive en el paquete auth
// (auth.NewStaffActorNameLookup), porque staff_user es una tabla propia de
// auth (identidad de sesión), distinta de barber (BarberNamePort, arriba):
// booking nunca importa auth.
type StaffActorNamePort interface {
	Names(ctx context.Context, barbershopID string, staffUserIDs []string) (map[string]string, error)
}

// BarberPort confirma que barberID existe dentro de barbershopID
// (RN-TEN-01), mismo puerto mínimo que schedule.BarberPort. cmd/api lo
// satisface con staff.NewBarberLookup, sin que booking importe staff.
type BarberPort interface {
	Exists(ctx context.Context, barbershopID, barberID string) (bool, error)
}

// CreateManualResult es el desenlace de un intento de alta manual
// idempotente (RN-IDE-01, DEC-043): Decision.Outcome distingue Proceed
// (recién creada, Appointment/Customer reflejan la fila nueva),
// Replay (repetición exacta, Appointment/Customer reflejan la fila
// original) y cualquier otro desenlace que Decision.AsError() ya traduce.
type CreateManualResult struct {
	Decision    idempotency.Decision
	Appointment Appointment
	Customer    Customer
	// Response es la respuesta almacenada: la recién producida cuando
	// Decision.Outcome == OutcomeProceed, o una copia de
	// Decision.Response cuando Decision.Outcome == OutcomeReplay (mismo
	// criterio que schedule.CreateResult.Response).
	Response idempotency.StoredResponse
}

// BarberServicePort resuelve, para (barberID, serviceID) dentro de
// barbershopID, si el servicio está activo Y asignado a ese barbero
// (DEC-072). found=false cubre barbero inexistente, servicio inexistente,
// inactivo o no asignado, sin distinguir la causa (RN-TEN-01). Los cuatro
// valores devueltos son exactamente los que ServiceSnapshot necesita para
// congelar el servicio en la cita: la firma usa solo tipos universales
// (string, int, int64, bool, error) para que el adaptador real (en el
// paquete catalog) satisfaga esta interfaz de forma puramente estructural,
// sin importar el paquete booking (mismo criterio que
// staff.NewBarberLookup frente a catalog.BarberPort).
type BarberServicePort interface {
	ActiveAssignedService(ctx context.Context, barbershopID, barberID, serviceID string) (
		name string, durationMinutes int, priceAmountCents int64, currency string, found bool, err error,
	)
}

// BlockCheckPort informa si barberID tiene, dentro de barbershopID, un
// bloqueo vigente (puntual o de serie, sin distinguir el tipo) que se
// solape con el intervalo semiabierto [startsAt, endsAt) (DEC-073,
// RN-DIS-05). timezone es la zona IANA ya resuelta de la barbería
// (RN-DIS-07): el adaptador real (en el paquete schedule) la necesita para
// convertir las ocurrencias de serie -civiles- a instantes absolutos antes
// de comparar. Firma con tipos universales únicamente, mismo criterio que
// BarberServicePort.
type BlockCheckPort interface {
	HasActiveBlock(ctx context.Context, barbershopID, barberID string, startsAt, endsAt time.Time, timezone string) (bool, error)
}

// TimezonePort resuelve la zona IANA vigente de barbershopID (RN-DIS-07).
// El adaptador real vive en el paquete shops.
type TimezonePort interface {
	Timezone(ctx context.Context, barbershopID string) (string, error)
}
