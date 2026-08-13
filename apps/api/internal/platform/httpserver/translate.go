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
	if appErr, ok := apperr.As(err); ok && appErr.Kind == apperr.KindNotFound {
		return newProblem(notFoundProblem, appErr.Message, requestID)
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
