package httpserver

import (
	"errors"
	"net/http"

	"system-barbershop/internal/platform/apperr"
)

// problemSpec fija, para un apperr.Kind, la representación de red estable:
// URI de tipo, título, código de proyecto y estado HTTP. type es una ruta
// relativa bajo /api/v1 (docs/06-api/estandar-openapi.md): el proyecto
// todavía no tiene un dominio público donde publicar documentación de
// errores, así que el identificador se ata a una ruta que el propio API
// puede servir el día que exista ese contenido, en vez de a un dominio
// externo que no existe hoy.
type problemSpec struct {
	typeURI string
	title   string
	code    string
	status  int
}

// Los valores de code siguen docs/06-api/estandar-openapi.md sección 5:
// minúscula con guiones, como el ejemplo normativo "slot-conflict".
var notFoundProblem = problemSpec{
	typeURI: "/api/v1/problems/not-found",
	title:   "Recurso no encontrado",
	code:    "not-found",
	status:  http.StatusNotFound,
}

var internalProblem = problemSpec{
	typeURI: "/api/v1/problems/internal-error",
	title:   "Error interno",
	code:    "internal-error",
	status:  http.StatusInternalServerError,
}

var bodyTooLargeProblem = problemSpec{
	typeURI: "/api/v1/problems/payload-too-large",
	title:   "Cuerpo de la solicitud demasiado grande",
	code:    "payload-too-large",
	status:  http.StatusBadRequest,
}

var invalidRequestProblem = problemSpec{
	typeURI: "/api/v1/problems/invalid-request",
	title:   "Solicitud inválida",
	code:    "invalid-request",
	status:  http.StatusBadRequest,
}

// validationProblem cubre un cuerpo bien formado que incumple una
// validación de campo (docs/06-api/estandar-openapi.md sección 11, fila
// 422): distinto de invalidRequestProblem, que cubre JSON o sintaxis
// inválida detectada antes de poder evaluar los campos.
var validationProblem = problemSpec{
	typeURI: "/api/v1/problems/validation-error",
	title:   "Error de validación",
	code:    "validation-error",
	status:  http.StatusUnprocessableEntity,
}

// unauthorizedProblem cubre HU-005 (CA-005-02, CA-005-04, CA-005-07):
// credencial ausente, inválida o expirada. detail es siempre el mismo
// mensaje genérico para correo inexistente, contraseña incorrecta, usuario
// inactivo, token desconocido, vencido o revocado — apperr.KindUnauthorized
// garantiza que ninguna rama de código pueda distinguirlos en la respuesta.
var unauthorizedProblem = problemSpec{
	typeURI: "/api/v1/problems/unauthorized",
	title:   "No autorizado",
	code:    "unauthorized",
	status:  http.StatusUnauthorized,
}

// idempotencyConflictProblem cubre RN-IDE-01: una clave de idempotencia ya
// usada con contenido u operación distintos. detail distingue el caso
// concreto; type, title y code se mantienen fijos, igual que notFoundProblem.
var idempotencyConflictProblem = problemSpec{
	typeURI: "/api/v1/problems/idempotency-conflict",
	title:   "Conflicto de idempotencia",
	code:    "idempotency-conflict",
	status:  http.StatusConflict,
}

// idempotencyLockedProblem cubre DEC-043: otra llamada concurrente con la
// misma clave sigue en curso. 409 sin espera acotada; el cliente puede
// reintentar más tarde.
var idempotencyLockedProblem = problemSpec{
	typeURI: "/api/v1/problems/idempotency-locked",
	title:   "Operación en curso",
	code:    "idempotency-locked",
	status:  http.StatusConflict,
}

// Translate convierte cualquier error en un Problem seguro para el cliente.
// Es el único punto central de traducción que exige
// docs/04-arquitectura/backend-go.md sección 5 ("los errores de dominio se
// traducen a HTTP en un punto común"):
//
//   - *apperr.Error con Kind distinto de KindInternal expone su Message,
//     que el llamador ya garantizó seguro.
//   - Cualquier otro error —incluido apperr.KindInternal, un *http.MaxBytesError
//     o un error ajeno no reconocido— colapsa a un problema genérico sin
//     reenviar jamás err.Error() al cliente (CA-003-02): un mensaje de
//     driver, una ruta de archivo o una traza no tienen forma de llegar al
//     cuerpo de la respuesta desde aquí.
//   - Un recurso inexistente y uno de otra barbería usan el mismo
//     apperr.KindNotFound, así que ambos producen exactamente el mismo
//     Problem (CA-003-03, RN-TEN-01): no hay una rama de código que pueda
//     distinguirlos.
func Translate(err error, requestID string) Problem {
	if appErr, ok := apperr.As(err); ok {
		switch appErr.Kind {
		case apperr.KindNotFound:
			return newProblem(notFoundProblem, appErr.Message, requestID)
		case apperr.KindInvalid:
			return newProblem(invalidRequestProblem, appErr.Message, requestID)
		case apperr.KindIdempotencyConflict:
			return newProblem(idempotencyConflictProblem, appErr.Message, requestID)
		case apperr.KindIdempotencyLocked:
			return newProblem(idempotencyLockedProblem, appErr.Message, requestID)
		case apperr.KindValidation:
			return newProblem(validationProblem, appErr.Message, requestID)
		case apperr.KindUnauthorized:
			return newProblem(unauthorizedProblem, appErr.Message, requestID)
		}
	}

	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		return newProblem(bodyTooLargeProblem, "", requestID)
	}

	return newProblem(internalProblem, "", requestID)
}

func newProblem(spec problemSpec, detail, requestID string) Problem {
	return Problem{
		Type:      spec.typeURI,
		Title:     spec.title,
		Status:    spec.status,
		Detail:    detail,
		Instance:  requestID,
		Code:      spec.code,
		RequestID: requestID,
	}
}
