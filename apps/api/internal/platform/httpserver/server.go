// Package httpserver construye el servidor HTTP y el middleware base
// compartido por las rutas públicas, de cliente y privadas. No conoce Chi ni
// las rutas de un módulo concreto: cmd/api compone routers de módulo sobre
// el handler que este paquete produce, según docs/04-arquitectura/backend-go.md.
package httpserver

import (
	"net/http"
	"time"
)

// Valores de tiempo conservadores para el servidor base. Un módulo con una
// necesidad distinta la documenta en su propio middleware, no aquí.
const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
)

// New construye un *http.Server listo para ejecutar, con los timeouts base
// del proyecto. El llamador decide cuándo iniciarlo y cómo apagarlo.
func New(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}

// Chain aplica middlewares a un handler en el orden dado: el primero de la
// lista es el más externo. Existe para que cmd/api componga el orden
// documentado en docs/04-arquitectura/backend-go.md sin repetir el patrón de
// envoltura en cada proceso.
func Chain(h http.Handler, mw ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}
