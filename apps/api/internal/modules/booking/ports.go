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
