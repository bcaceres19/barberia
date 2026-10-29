package customeraccess

import (
	"context"

	"system-barbershop/internal/platform/idempotency"
)

// Repository resuelve un token de acceso ya hasheado a la proyección
// mínima del turno correspondiente. found=false cubre, sin distinguirlas
// (CA-098-02, RN-TEN-01), las tres causas posibles: el hash no existe,
// existe pero ya venció, o existe pero ya fue revocado.
type Repository interface {
	ResolveAppointmentByTokenHash(ctx context.Context, tokenHash string) (view AppointmentView, found bool, err error)

	// ResolveCancellationTarget resuelve el mismo token vigente a la
	// barbería y la cita sobre las que HU-099 ejecuta la cancelación. found
	// sigue la misma regla uniforme que ResolveAppointmentByTokenHash.
	ResolveCancellationTarget(ctx context.Context, tokenHash string) (target CancellationTarget, found bool, err error)
}

// AppointmentCanceller ejecuta T5 sobre un turno ya resuelto desde su token
// (HU-099). Lo implementa un adaptador de booking: customeraccess nunca importa
// booking (CA-002-06), así que el puerto solo habla con tipos universales. La
// decisión de idempotencia y la respuesta almacenada se reproducen tal cual;
// los rechazos de versión, estado y política vuelven como error.
type AppointmentCanceller interface {
	CancelAppointment(
		ctx context.Context,
		barbershopID, appointmentID, expectedVersionToken, reason string,
		key idempotency.Key,
		fingerprint idempotency.Fingerprint,
	) (idempotency.Decision, idempotency.StoredResponse, error)
}
