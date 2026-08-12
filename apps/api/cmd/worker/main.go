// Command worker arranca el proceso trabajador que enviará recordatorios ya
// persistidos. Reutiliza servicios y repositorios de los módulos, no
// handlers HTTP, según docs/04-arquitectura/stack-despliegue-operacion.md.
package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
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

	// El worker se conecta como barberia_worker (cfg.WorkerDatabaseURL), NO
	// como cfg.DatabaseURL (esa es la credencial del proceso api): DEC-040
	// exige roles separados, y config.Load ya rechaza que ambas URLs sean
	// iguales fuera de local/test. WorkerDatabaseURL solo puede llegar vacío
	// en local/test (Load lo exige en cualquier otro ambiente); ahí se
	// reutiliza la URL del api para no exigir una segunda variable en
	// desarrollo local.
	workerDSN := cfg.WorkerDatabaseURL
	if workerDSN == "" {
		workerDSN = cfg.DatabaseURL
	}
	db, err := database.NewDB(workerDSN, cfg)
	if err != nil {
		logger.Error("no se pudo iniciar el pool de base de datos")
		return errors.New("database: fallo al iniciar el pool")
	}
	defer db.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// El reclamo de lotes con SKIP LOCKED, el envío por canal y los
	// reintentos se agregan junto con el módulo notification, según
	// docs/05-backend/estandar-base-datos.md.
	logger.Info("worker iniciado, sin trabajos programados todavía", "environment", cfg.Environment)

	<-ctx.Done()
	logger.Info("worker apagado")

	return nil
}
