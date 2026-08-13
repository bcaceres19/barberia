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
	"system-barbershop/internal/platform/database"
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

	// La aplicación se conecta como barberia_app (cfg.DatabaseURL), nunca
	// como barberia_migrator ni superusuario. Un fallo aquí impide arrancar:
	// el mensaje no incluye el error crudo del driver para no arriesgar un
	// fragmento del DSN en el log de arranque (mismo criterio que
	// DB.HealthCheck).
	db, err := database.NewDB(cfg.DatabaseURL, cfg)
	if err != nil {
		logger.Error("no se pudo iniciar el pool de base de datos")
		return errors.New("database: fallo al iniciar el pool")
	}
	defer db.Close()

	// httpserver.NewRouter monta las tres audiencias de
	// docs/04-arquitectura/backend-go.md (/api/v1/public, /api/v1/customer,
	// /api/v1/private) sobre Chi v5 (DEC-034) con el middleware base ya
	// aplicado. Las rutas de módulo (auth, shops, staff, catalog, schedule,
	// booking, notification) se registran aquí a medida que existan.
	router := httpserver.NewRouter(logger)
	router.Get("/health", httpserver.HealthHandler())
	router.Get("/health/db", httpserver.DatabaseHealthHandler(db))

	server := httpserver.New(cfg.HTTPAddr, router)

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
