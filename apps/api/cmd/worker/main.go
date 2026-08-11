// Command worker arranca el proceso trabajador que enviará recordatorios ya
// persistidos. Reutiliza servicios y repositorios de los módulos, no
// handlers HTTP, según docs/04-arquitectura/stack-despliegue-operacion.md.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/observability"
)

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// El reclamo de lotes con SKIP LOCKED, el envío por canal y los
	// reintentos se agregan junto con el módulo notification y
	// internal/platform/database, según docs/05-backend/estandar-base-datos.md.
	logger.Info("worker iniciado, sin trabajos programados todavía", "environment", cfg.Environment)

	<-ctx.Done()
	logger.Info("worker apagado")

	return nil
}
