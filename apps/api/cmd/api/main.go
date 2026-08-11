// Command api arranca el proceso HTTP del sistema de agenda. Solo carga
// configuración, construye dependencias, registra rutas y controla el
// apagado: no contiene reglas de negocio ni SQL, según
// docs/04-arquitectura/backend-go.md.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/httpserver"
	"system-barbershop/internal/platform/observability"
)

const shutdownGracePeriod = 10 * time.Second

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := observability.NewLogger(cfg.LogLevel)

	// Las rutas de módulo (auth, shops, staff, catalog, schedule, booking,
	// notification) se registran aquí a medida que existan, siguiendo la
	// estructura de audiencias de docs/04-arquitectura/backend-go.md
	// (/api/v1/public, /api/v1/customer, /api/v1/private). Chi v5 se
	// incorpora en ese momento (DEC-034); el mux estándar basta mientras la
	// única ruta es la comprobación operativa.
	mux := http.NewServeMux()
	mux.Handle("GET /health", httpserver.HealthHandler())

	handler := httpserver.Chain(mux,
		httpserver.RequestID,
		httpserver.Recover(logger),
	)

	server := httpserver.New(cfg.HTTPAddr, handler)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("servidor http iniciado", "addr", cfg.HTTPAddr, "environment", cfg.Environment)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownGracePeriod)
	defer cancel()

	logger.Info("apagando servidor http")
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("apagado forzado del servidor http", "error", err)
		return err
	}

	return nil
}
