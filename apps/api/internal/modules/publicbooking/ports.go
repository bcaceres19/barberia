package publicbooking

import (
	"context"
	"time"

	"system-barbershop/internal/platform/idempotency"
)

// Repository es el puerto de resolución pública del módulo. El núcleo no
// importa internal/platform/database ni pgx (CA-002-06): postgres/ traduce
// entre este contrato y database.DB, exactamente igual que shops.Repository.
type Repository interface {
	// ResolveBySlug resuelve slug (ya recortado, dentro de MaxSlugLength) a
	// la barbería pública correspondiente. found=false cubre un slug con
	// forma inválida, inexistente o de una barbería no publicable
	// (public_slug NULL): las tres causas son indistinguibles a propósito
	// (CA-090-02, RN-TEN-01) y se resuelven en UNA sola llamada mediante
	// public_resolve_barbershop_by_slug + la lectura tenant-aware
	// subsiguiente, sin que el llamador pueda fijar ni inferir
	// barbershopId.
	ResolveBySlug(ctx context.Context, slug string) (profile BarbershopProfile, found bool, err error)

	// ListPublicServices resuelve slug (ya recortado, dentro de
	// MaxSlugLength) EXACTAMENTE como ResolveBySlug -mismo
	// public_resolve_barbershop_by_slug, misma indistinción entre forma
	// inválida/desconocido/no publicable- y, si resuelve, lee una página de
	// los servicios ACTIVOS con al menos una asignación vigente (HU-091,
	// CA-091-01) de esa barbería, ordenada por (created_at, id). cursor es
	// nil para la primera página; limit ya llegó clamped al rango
	// [MinServiceListLimit, MaxServiceListLimit] por Service.
	// barbershopFound=false cubre exactamente el mismo universo de causas
	// que ResolveBySlug found=false; en ese caso ListResult queda vacío. No
	// expone barbershopId: cada llamada vuelve a resolver por slug, igual
	// que ResolveBySlug, así que el núcleo de publicbooking nunca retiene ni
	// recibe un identificador interno entre ambas operaciones.
	ListPublicServices(ctx context.Context, slug string, cursor *ServiceCursor, limit int) (result PublicServiceListResult, barbershopFound bool, err error)

	// ListPublicBarbers resuelve slug EXACTAMENTE como ResolveBySlug/
	// ListPublicServices y, si resuelve, lee los barberos con asignación
	// vigente a serviceID (HU-092, CA-092-02) ordenados por
	// (created_at de la asignación, barber.id). serviceID ya llegó validado
	// como forma de UUID por Service (LooksLikePublicServiceID); un
	// serviceID de otra barbería, inexistente, o de un servicio inactivo
	// produce una lista vacía -NUNCA un error distinto- porque el filtro
	// tenant-aware ya lo excluye igual que "sin barberos asignados"
	// (CA-092-03: misma respuesta uniforme, sin distinguir la causa).
	// barbershopFound=false cubre exactamente el mismo universo de causas
	// que ResolveBySlug found=false.
	ListPublicBarbers(ctx context.Context, slug string, serviceID string) (result PublicBarberListResult, barbershopFound bool, err error)
}

// AvailabilityRepository es el puerto de persistencia adicional que
// AvailabilityService (HU-094) necesita, deliberadamente SEPARADO de
// Repository: ResolveBySlug/ListPublicServices/ListPublicBarbers nunca
// retienen ni reciben un identificador interno entre operaciones (mismo
// slug se vuelve a resolver cada vez), pero HU-094 sí lo necesita para
// orquestar schedule/shops/catalog -que exigen barbershopID, no un slug- en
// una sola operación, algo que esas tres lecturas de una sola tabla nunca
// necesitaron. Mantener este puerto separado evita que las pruebas de
// Service/ResolveBySlug tengan que conocer un identificador que nunca usan.
// El identificador que ResolveBarbershopID devuelve NUNCA sale de
// AvailabilityService hacia ningún DTO ni respuesta HTTP (RN-TEN-01,
// CA-090-04): existe solo para las llamadas server-side subsiguientes
// dentro de la misma solicitud.
type AvailabilityRepository interface {
	// ResolveBarbershopID resuelve slug EXACTAMENTE como
	// Repository.ResolveBySlug (misma public_resolve_barbershop_by_slug,
	// misma indistinción de causas) pero devuelve el identificador interno
	// de la barbería en vez de su perfil público.
	ResolveBarbershopID(ctx context.Context, slug string) (barbershopID string, found bool, err error)

	// ListOccupiedIntervals lee, dentro de barbershopID y barberID, los
	// intervalos [starts_at, ends_at) de las citas que ocupan agenda
	// (columna generada `occupies_schedule`, RN-CON-01/RN-CAN-04: mismo
	// criterio que la restricción de exclusión de la base de datos, para
	// que el cálculo de disponibilidad nunca pueda divergir de ella) cuyo
	// intervalo interseca [from, to). starts/ends son pares paralelos
	// ordenados por starts_at.
	ListOccupiedIntervals(ctx context.Context, barbershopID, barberID string, from, to time.Time) (starts []time.Time, ends []time.Time, err error)
}

// CustomerRepository es el puerto de identidad pública (HU-096) que
// ReconcilePublicCustomer necesita: buscar, dentro de barbershopID, si
// phone y/o email ya pertenecen a un customer existente (DEC-085, resuelve
// DP-PUB-04). Deliberadamente separado de Repository/AvailabilityRepository
// -ninguna de esas operaciones toca la tabla customer- y sin método de
// escritura: HU-096 solo define la política; HU-097 persiste dentro de su
// propia transacción atómica.
type CustomerRepository interface {
	// FindCustomerMatches busca, dentro de barbershopID, un customer no
	// anonimizado cuyo phone coincida (phoneMatchID) y, por separado, uno
	// cuyo email coincida (emailMatchID) -EXACTAMENTE la forma ya
	// normalizada que el cliente dio. Cualquiera de los dos puede volver
	// nil si no hay coincidencia; pueden apuntar al mismo id o a ids
	// distintos (RN-TEN-01: la búsqueda nunca cruza barbershopID).
	FindCustomerMatches(ctx context.Context, barbershopID, phone, email string) (phoneMatchID, emailMatchID *string, err error)
}

// EffectiveDaySegmentsLimit acota cuántos tramos por día puede devolver
// EffectiveDayPort antes de que Service los descarte como una respuesta
// inesperada: ninguna jornada real declara tantos tramos (HU-040/HU-041
// nunca lo permitieron), así que un valor mayor solo puede ser un adaptador
// roto, nunca un caso de negocio legítimo.
const EffectiveDaySegmentsLimit = 64

// EffectiveDayPort resuelve, para una fecha civil ("AAAA-MM-DD") dentro de
// barbershopID, si barberID trabaja y sus tramos como pares paralelos de
// hora civil de inicio ("HH:MM") y duración en minutos (RN-DIS-02,
// CA-041-07). El adaptador real vive en el paquete schedule
// (schedule.NewAvailabilityLookup), que reutiliza Service.ResolveEffectiveDay
// sin duplicar la precedencia excepción manual > festivo automático >
// horario semanal (B2): publicbooking nunca importa schedule. Firma con
// tipos universales únicamente, mismo criterio que booking.BarberServicePort.
type EffectiveDayPort interface {
	ResolveEffectiveDay(ctx context.Context, barbershopID, barberID, dateCivil string) (
		isWorking bool, segmentStarts []string, segmentDurations []int, err error,
	)
}

// BusyBlocksPort resuelve, para barberID dentro de barbershopID y el rango
// de fechas civiles [fromDate, toDate], los instantes absolutos ocupados
// por bloqueos vigentes (puntuales y de serie ya expandidos), como pares
// paralelos de inicio/fin (RN-BLQ-01 a RN-BLQ-04). timezone es la zona IANA
// ya resuelta de la barbería (RN-DIS-07), necesaria para convertir las
// ocurrencias de serie -civiles- a instantes absolutos. El adaptador real
// vive en el paquete schedule (schedule.NewAvailabilityLookup), que
// reutiliza Service.EffectiveBlocks (HU-042) sin duplicar su proyección
// (B2/B3): publicbooking nunca importa schedule.
type BusyBlocksPort interface {
	BusyIntervals(ctx context.Context, barbershopID, barberID, fromDate, toDate, timezone string) (
		starts []time.Time, ends []time.Time, err error,
	)
}

// BookingPolicyPort resuelve la anticipación mínima (minutos), la ventana
// máxima (días) y el paso de rejilla (minutos) vigentes de barbershopID
// (HU-093, RN-DIS-04, RN-DIS-06). El adaptador real vive en el paquete
// shops (shops.NewAvailabilityBookingPolicy), que reutiliza
// BookingPolicyService.Get sin duplicar sus rangos/defaults (DEC-083):
// publicbooking nunca importa shops.
type BookingPolicyPort interface {
	BookingPolicy(ctx context.Context, barbershopID string) (minAdvanceMinutes, maxAdvanceDays, slotGridMinutes int, err error)
}

// ServiceAssignmentPort resuelve, para (barberID, serviceID) dentro de
// barbershopID, si el servicio está activo Y asignado a ese barbero,
// devolviendo su duración planificada (DEC-002). found=false cubre
// servicio/barbero inexistente, ajeno, inactivo o no asignado, sin
// distinguir la causa (RN-TEN-01, CA-092-03: mismo criterio que HU-092).
// Firma idéntica a booking.BarberServicePort a propósito: el mismo
// adaptador ya construido para HU-061 (catalog.NewManualBookingCatalog) la
// satisface de forma puramente estructural, sin código nuevo en catalog.
type ServiceAssignmentPort interface {
	ActiveAssignedService(ctx context.Context, barbershopID, barberID, serviceID string) (
		name string, durationMinutes int, priceAmountCents int64, currency string, found bool, err error,
	)
}

// PublicAppointmentPort ejecuta T1 pública (HU-097) dentro de UNA sola
// transacción atómica protegida por el protocolo de idempotencia
// reutilizable de HU-004 (RN-IDE-01, DEC-043): reconcilia el cliente según
// la decisión ya tomada por ConfirmationService (DEC-085), inserta la cita
// `confirmed` de origen `public` con sus snapshots, el evento
// appointment_created y el token de acceso (DEC-089), todo o nada. Firma
// con tipos universales únicamente (mismo criterio que BarberServicePort/
// ServiceAssignmentPort): el adaptador real vive en el paquete booking
// (booking.NewPublicAppointmentAdapter), que reutiliza su propia primitiva
// transaccional (booking.Repository.CreatePublic) sin duplicar SQL;
// publicbooking nunca importa booking (CA-002-06).
//
// customerExistingID no nil reutiliza ese cliente (sobrescribiendo
// customerUpdatePhone/customerUpdateEmail cuando no son nil, el campo que
// no participó en la coincidencia, DEC-085); nil crea un cliente nuevo con
// customerFullName/customerPhone/customerEmail/customerNote.
// tokenPlain/tokenHash/tokenIssuedAt/tokenExpiresAt ya llegan calculados
// por Service (crypto/rand + SHA-256, DEC-089): este puerto persiste solo
// el hash, pero necesita tokenPlain para incluirlo, una única vez, en la
// respuesta almacenada de idempotencia. Un error cuyo apperr.As expone
// Kind == apperr.KindScheduleConflict señala que la restricción de
// exclusión de PostgreSQL (RN-CON-03) rechazó la franja (RN-CON-05):
// ConfirmationService lo reconoce por ese Kind (apperr es un paquete de
// plataforma, no un módulo hermano), sin que este paquete importe booking.
type PublicAppointmentPort interface {
	CreatePublicAppointment(
		ctx context.Context,
		barbershopID, barbershopName, timezone string,
		barberID, serviceID string,
		startsAt, endsAt time.Time,
		attendeeName string,
		serviceName string,
		durationMinutes int,
		priceAmountCents int64,
		currency string,
		customerNote *string,
		customerExistingID *string,
		customerUpdatePhone *string,
		customerUpdateEmail *string,
		customerFullName, customerPhone, customerEmail string,
		tokenPlain, tokenHash string,
		tokenIssuedAt, tokenExpiresAt time.Time,
		key idempotency.Key,
		fingerprint idempotency.Fingerprint,
	) (decision idempotency.Decision, stored idempotency.StoredResponse, err error)
}

// ConfirmationEmailPort entrega el correo de confirmación de HU-097
// (F-PUB-07, DEC-091): envío síncrono simple, un único intento, sin cola ni
// reintentos, reutilizando el proveedor de correo ya integrado y verificado
// en producción (Resend). El adaptador real vive en el paquete notification
// (notification.NewResendConfirmationEmailSender), que satisface esta
// interfaz de forma puramente estructural (mismo criterio que
// notification.DualChannelRecoverySender frente al puerto de auth):
// publicbooking nunca importa notification. Un error de entrega NUNCA
// revierte la transacción ya comprometida (DEC-091, "dentro o
// inmediatamente después de" la transacción): ConfirmAppointment ya
// devolvió éxito al cliente cuando este puerto se invoca; el llamador HTTP
// solo registra el fallo en el log (mismo criterio que
// auth.RecoveryService.Request frente a RecoveryCodeSender).
type ConfirmationEmailPort interface {
	SendConfirmation(
		ctx context.Context,
		email, barbershopName, serviceName, attendeeName string,
		startsAtLocalFormatted string,
		accessLink string,
	) error
}

// AvailabilityTimezonePort resuelve la zona IANA vigente de barbershopID
// (RN-DIS-07). Firma idéntica a booking.TimezonePort a propósito: el mismo
// adaptador ya construido para HU-061/HU-062/HU-065 (shops.NewTimezoneLookup)
// la satisface de forma puramente estructural, sin código nuevo en shops.
type AvailabilityTimezonePort interface {
	Timezone(ctx context.Context, barbershopID string) (string, error)
}
