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

// As extrae un *Error de la cadena de err, igual que errors.As.
func As(err error) (*Error, bool) {
	var appErr *Error
	ok := errors.As(err, &appErr)
	return appErr, ok
}
