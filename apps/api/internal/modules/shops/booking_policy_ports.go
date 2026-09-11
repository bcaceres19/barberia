package shops

import "context"

// BookingPolicyUpdateInput es la forma ya validada por BookingPolicyService
// que llega al repositorio: PUT de contrato cerrado, todos los campos
// siempre presentes (CA-093-02, sin PATCH parcial). ExpectedVersionToken es
// el valor de la cabecera If-Match.
type BookingPolicyUpdateInput struct {
	MinAdvanceMinutes              int
	MaxAdvanceDays                 int
	SlotGridMinutes                int
	CancellationDeadlineMinutes    int
	LateCancellationClientAllowed  bool
	LateCancellationReasonRequired bool
	ExpectedVersionToken           string
}

// BookingPolicyUpdateOutcome distingue, sin ambigüedad, qué pasó dentro de
// la transacción de BookingPolicyRepository.Update.
type BookingPolicyUpdateOutcome int

const (
	// BookingPolicyUpdateOutcomeUpdated: el token coincidió y la fila se
	// escribió.
	BookingPolicyUpdateOutcomeUpdated BookingPolicyUpdateOutcome = iota
	// BookingPolicyUpdateOutcomeVersionConflict: la fila existe pero el
	// token vigente ya no coincide con ExpectedVersionToken; CERO escritura
	// (CA-093-02).
	BookingPolicyUpdateOutcomeVersionConflict
	// BookingPolicyUpdateOutcomeNotFound: la barbería no es visible en el
	// tenant vigente (defensivo, mismo criterio que UpdateResult.Found).
	BookingPolicyUpdateOutcomeNotFound
)

// BookingPolicyUpdateResult es el desenlace completo de
// BookingPolicyRepository.Update.
type BookingPolicyUpdateResult struct {
	Outcome BookingPolicyUpdateOutcome
	Policy  BookingPolicy
}

// BookingPolicyRepository es el puerto de persistencia de la política
// pública de reserva y cancelación (HU-093). El núcleo no importa
// internal/platform/database ni pgx (CA-002-06), mismo patrón que
// Repository (ports.go).
type BookingPolicyRepository interface {
	// Get lee la política vigente de barbershopID dentro de una transacción
	// tenant-aware. found=false cubre una fila no visible en el tenant
	// vigente (defensivo, mismo criterio que Repository.Get).
	Get(ctx context.Context, barbershopID string) (policy BookingPolicy, found bool, err error)

	// Update bloquea la fila (SELECT ... FOR UPDATE) dentro de UNA sola
	// transacción tenant-aware, compara el token de versión vigente contra
	// input.ExpectedVersionToken (BookingPolicyUpdateOutcomeVersionConflict
	// si no coincide, CERO escritura) y, si coincide, ejecuta el UPDATE
	// completo de los seis campos en la MISMA transacción (CA-093-02: sin
	// escritura parcial). BookingPolicyService ya validó los rangos y la
	// coherencia entre campos antes de llamar aquí; las restricciones CHECK
	// de la migración son la última línea de defensa, no la primera.
	Update(ctx context.Context, barbershopID string, input BookingPolicyUpdateInput) (BookingPolicyUpdateResult, error)
}
