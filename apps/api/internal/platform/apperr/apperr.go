// Package apperr define los tipos de error de aplicación que el dominio y
// los servicios usan para señalar un resultado no exitoso sin conocer HTTP,
// Chi ni el formato de respuesta. Solo la capa httpserver sabe traducir un
// *Error a application/problem+json, según docs/04-arquitectura/backend-go.md
// sección 5 ("los errores de dominio se traducen a HTTP en un punto común").
package apperr

import (
	"errors"
	"fmt"
)

// Kind agrupa los errores de aplicación en el pequeño catálogo estable que
// HU-003 necesita. Un recurso inexistente y un recurso de otra barbería usan
// deliberadamente el mismo Kind: no existe una forma de invocar este paquete
// que distinga ambos casos, así que la traducción HTTP no puede filtrar esa
// diferencia por accidente (CA-003-03, RN-TEN-01).
type Kind string

const (
	// KindNotFound cubre tanto un recurso inexistente como uno que
	// pertenece a otra barbería.
	KindNotFound Kind = "not_found"
	// KindInternal cubre cualquier fallo inesperado. Message se ignora
	// deliberadamente en la traducción HTTP: solo Err viaja para
	// diagnóstico interno (logs), nunca al cliente.
	KindInternal Kind = "internal"
	// KindInvalid cubre una entrada de cliente malformada (JSON, parámetro
	// o cabecera con sintaxis o formato inválido) detectada antes de
	// ejecutar cualquier efecto. Genérico y reutilizable: no es específico
	// de idempotencia (docs/06-api/estandar-openapi.md sección 11, fila 400).
	KindInvalid Kind = "invalid"
	// KindIdempotencyConflict cubre una clave de idempotencia ya usada con
	// contenido u operación distintos (RN-IDE-01). No se reutiliza para
	// otro tipo de conflicto (por ejemplo, de agenda): un conflicto futuro
	// no relacionado con idempotencia necesitará su propio Kind, porque el
	// código de proyecto que Translate expone al cliente debe distinguirlos.
	KindIdempotencyConflict Kind = "idempotency_conflict"
	// KindIdempotencyLocked cubre una segunda llamada concurrente con la
	// misma clave de idempotencia mientras la primera sigue en curso
	// (DEC-043: pg_try_advisory_xact_lock sin espera acotada). El cliente
	// puede reintentar más tarde; no es un error permanente.
	KindIdempotencyLocked Kind = "idempotency_locked"
	// KindValidation cubre un cuerpo bien formado (JSON válido, campos
	// conocidos) que incumple una validación de campo o de negocio
	// procesable, distinto de KindInvalid (docs/06-api/estandar-openapi.md
	// sección 11, filas 400 y 422).
	KindValidation Kind = "validation"
	// KindUnauthorized cubre una credencial ausente, inválida o expirada
	// (HU-005/HU-006). Un correo inexistente y una contraseña incorrecta
	// para un correo existente comparten deliberadamente este mismo Kind y
	// el mismo Message: no existe una forma de invocar este paquete que
	// distinga los dos casos, igual que KindNotFound para tenant cruzado
	// (CA-005-02).
	KindUnauthorized Kind = "unauthorized"
	// KindChallengeRequired cubre una IP que superó el umbral de intentos
	// de acceso y debe completar el reto telefónico antes de que la
	// contraseña se evalúe (HU-007, DEC-061/DEC-062). No se reutiliza
	// KindUnauthorized: el cliente necesita distinguir "credenciales
	// incorrectas" de "falta completar el reto" para mostrar la
	// experiencia correcta, y ambos ya tienen semántica HTTP distinta
	// (401 frente a 429).
	KindChallengeRequired Kind = "challenge_required"
	// KindConflict cubre un conflicto de estado de negocio ajeno a
	// idempotencia: otra fila ya persistida vuelve inválida la operación
	// solicitada (por ejemplo, HU-022 CA-022-04: un nombre de servicio ya
	// usado por otro servicio activo de la misma barbería). No se
	// reutiliza KindIdempotencyConflict, reservado a RN-IDE-01, porque este
	// conflicto no depende de ninguna cabecera Idempotency-Key ni de un
	// reintento: existiría igual en la primera y única solicitud.
	KindConflict Kind = "conflict"
)

// Error es el error de aplicación que domain/servicios devuelven.
type Error struct {
	Kind Kind
	// Message es texto seguro para el cliente: nunca SQL, rutas de
	// archivo, nombres de proveedor ni versiones de dependencias
	// (CA-003-02). Se ignora para KindInternal.
	Message string
	// Err es la causa interna, útil para logs; la capa HTTP nunca la
	// vuelca en la respuesta.
	Err error
	// RetryAfterSeconds acompaña KindChallengeRequired con el segundo en
	// que el escalamiento deja de estar vigente (HU-007). Cero para
	// cualquier otro Kind: la capa HTTP solo añade la cabecera Retry-After
	// cuando este campo es positivo.
	RetryAfterSeconds int
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("apperr[%s]: %s: %v", e.Kind, e.Message, e.Err)
	}
	return fmt.Sprintf("apperr[%s]: %s", e.Kind, e.Message)
}

func (e *Error) Unwrap() error { return e.Err }

// NotFound construye un error de recurso ausente o de otra barbería. message
// es el detalle seguro que puede llegar al cliente.
func NotFound(message string) *Error {
	return &Error{Kind: KindNotFound, Message: message}
}

// Internal construye un error inesperado a partir de una causa interna. La
// causa nunca se expone al cliente; solo sirve para diagnóstico en logs.
func Internal(cause error) *Error {
	return &Error{Kind: KindInternal, Err: cause}
}

// Invalid construye un error de entrada de cliente malformada. message es el
// detalle seguro que puede llegar al cliente (nunca el valor crudo enviado,
// si ese valor pudiera ser sensible).
func Invalid(message string) *Error {
	return &Error{Kind: KindInvalid, Message: message}
}

// IdempotencyConflict construye el error de una clave de idempotencia
// reutilizada con contenido u operación distintos (RN-IDE-01). message es el
// detalle seguro que puede llegar al cliente.
func IdempotencyConflict(message string) *Error {
	return &Error{Kind: KindIdempotencyConflict, Message: message}
}

// IdempotencyLocked construye el error de una clave de idempotencia cuya
// ejecución sigue en curso en otra transacción concurrente (DEC-043).
// message es el detalle seguro que puede llegar al cliente.
func IdempotencyLocked(message string) *Error {
	return &Error{Kind: KindIdempotencyLocked, Message: message}
}

// Validation construye un error de cuerpo bien formado que incumple una
// validación de campo. message es el detalle seguro que puede llegar al
// cliente.
func Validation(message string) *Error {
	return &Error{Kind: KindValidation, Message: message}
}

// Unauthorized construye el error uniforme de credencial ausente, inválida o
// expirada. message es el detalle seguro que puede llegar al cliente: nunca
// distingue correo inexistente de contraseña incorrecta, ni token
// desconocido de vencido, revocado o de usuario inactivo (CA-005-02,
// CA-005-07).
func Unauthorized(message string) *Error {
	return &Error{Kind: KindUnauthorized, Message: message}
}

// ChallengeRequired construye el error de escalamiento de HU-007: la
// contraseña no se evaluó porque la IP superó el umbral de intentos.
// retryAfterSeconds debe ser positivo (segundos hasta que el escalamiento
// deje de estar vigente); la capa HTTP lo traduce en la cabecera
// Retry-After.
func ChallengeRequired(message string, retryAfterSeconds int) *Error {
	return &Error{Kind: KindChallengeRequired, Message: message, RetryAfterSeconds: retryAfterSeconds}
}

// Conflict construye el error de un conflicto de negocio ajeno a
// idempotencia (KindConflict): otra fila ya persistida vuelve inválida la
// operación. message es el detalle seguro que puede llegar al cliente.
func Conflict(message string) *Error {
	return &Error{Kind: KindConflict, Message: message}
}

// As extrae un *Error de la cadena de err, igual que errors.As.
func As(err error) (*Error, bool) {
	var appErr *Error
	ok := errors.As(err, &appErr)
	return appErr, ok
}
