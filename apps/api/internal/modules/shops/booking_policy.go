package shops

import (
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"time"
)

// BookingPolicy es la configuración pública de reserva y cancelación de la
// barbería activa (HU-093, RN-DIS-04, RN-DIS-06, RN-CAN-01, RN-CAN-02).
// VersionToken es el token opaco de concurrencia (mismo patrón que
// booking.EncodeVersionToken): determinista a partir del identificador de
// la barbería y su updated_at vigente, irreversible.
type BookingPolicy struct {
	MinAdvanceMinutes              int
	MaxAdvanceDays                 int
	SlotGridMinutes                int
	CancellationDeadlineMinutes    int
	LateCancellationClientAllowed  bool
	LateCancellationReasonRequired bool
	VersionToken                   string
}

// Rangos permitidos de DEC-083 (resuelve DP-PUB-02): generosos alrededor de
// los valores iniciales de DEC-005/DEC-006/DEC-010/DEC-018, exactamente los
// mismos límites que las restricciones CHECK de
// 20260911060000_add_barbershop_booking_policy.sql (el servicio no inventa
// un rango distinto al que la base ya aplica, CA-093-02).
const (
	MinAdvanceMinutesFloor = 0
	MinAdvanceMinutesCeil  = 1440

	MaxAdvanceDaysFloor = 1
	MaxAdvanceDaysCeil  = 90

	CancellationDeadlineMinutesFloor = 0
	CancellationDeadlineMinutesCeil  = 10080
)

// AllowedSlotGridMinutes es el conjunto discreto que DEC-083 autoriza para
// la rejilla (RN-DIS-06, barbershop_slot_grid_minutes_ck): un paso no
// listado aquí se rechaza, nunca se redondea ni se acepta silenciosamente.
var AllowedSlotGridMinutes = [...]int{5, 10, 15, 20, 30, 60}

// IsAllowedSlotGridMinutes informa si v pertenece al conjunto discreto que
// DEC-083 autoriza para la rejilla.
func IsAllowedSlotGridMinutes(v int) bool {
	for _, allowed := range AllowedSlotGridMinutes {
		if v == allowed {
			return true
		}
	}
	return false
}

// EncodeBookingPolicyVersionToken produce el token opaco de concurrencia de
// la política de reserva de una barbería a partir de su identificador y su
// updated_at vigente (HU-093, CA-093-02). Duplica
// booking.EncodeVersionToken a propósito (CA-002-06: el núcleo de shops no
// importa booking): irreversible (SHA-256, no una codificación reversible
// de updated_at), para que una mutación futura pueda usarlo como
// precondición sin que updated_at se vuelva una regla de negocio pública.
// Dos lecturas de la misma fila sin ninguna escritura entre medias
// producen el mismo token; cualquier escritura que toque updated_at
// (barbershop_set_updated_at) produce uno distinto -incluida una
// escritura de HU-020 sobre name/timezone/contacto, que también invalida
// el token de la política: ambas comparten la misma fila y el mismo
// updated_at, así que una edición concurrente de cualquiera de las dos se
// detecta igual.
func EncodeBookingPolicyVersionToken(barbershopID string, updatedAt time.Time) string {
	sum := sha256.Sum256([]byte(barbershopID + "|" + strconv.FormatInt(updatedAt.UTC().UnixNano(), 10)))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
