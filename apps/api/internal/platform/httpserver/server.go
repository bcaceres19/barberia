// Package httpserver construye el router Chi v5, el middleware base y el
// *http.Server compartidos por las rutas públicas, de cliente y privadas.
// cmd/api monta sobre el router que [NewRouter] produce las rutas concretas
// de cada módulo a medida que existen, según
// docs/04-arquitectura/backend-go.md.
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
