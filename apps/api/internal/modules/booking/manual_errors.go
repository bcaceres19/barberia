package booking

import (
	"errors"

	"system-barbershop/internal/platform/apperr"
)

// errServiceNotAssigned cubre DEC-072: el servicio no existe, está
// inactivo o no está asignado al barbero elegido. Un único 404 sin
// distinguir la causa (mismo criterio uniforme que RN-TEN-01).
func errServiceNotAssigned() error {
	return apperr.NotFound("no existe un servicio activo asignado a ese barbero")
}

// errBlockedInterval cubre DEC-073: el intervalo elegido se solapa con un
// bloqueo vigente del barbero. Mismo tratamiento HTTP que un cruce de
// citas (409).
func errBlockedInterval() error {
	return apperr.Conflict("el barbero tiene un bloqueo vigente en ese intervalo")
}

// errTimezoneUnresolvable cubre el caso defensivo de una zona IANA
// persistida que time.LoadLocation ya no reconoce: no debería ocurrir
// nunca (shops.Repository.Update ya la valida contra pg_timezone_names
// antes de persistirla), pero un caso defensivo se distingue de un pánico.
func errTimezoneUnresolvable() error {
	return apperr.Internal(errors.New("booking: zona IANA de la barbería no reconocida por time.LoadLocation"))
}

func errStartsAtInvalid() error {
	return apperr.Invalid("startsAt debe ser una fecha y hora civiles con el formato AAAA-MM-DDTHH:MM:SS, sin zona")
}
