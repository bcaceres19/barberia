package customeraccess

import "context"

// Repository resuelve un token de acceso ya hasheado a la proyección
// mínima del turno correspondiente. found=false cubre, sin distinguirlas
// (CA-098-02, RN-TEN-01), las tres causas posibles: el hash no existe,
// existe pero ya venció, o existe pero ya fue revocado.
type Repository interface {
	ResolveAppointmentByTokenHash(ctx context.Context, tokenHash string) (view AppointmentView, found bool, err error)
}
