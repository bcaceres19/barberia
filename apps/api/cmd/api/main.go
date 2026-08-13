// Command api arranca el proceso HTTP del sistema de agenda. Solo carga
// configuración, construye dependencias, registra rutas y controla el
// apagado: no contiene reglas de negocio ni SQL, según
// docs/04-arquitectura/backend-go.md.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"system-barbershop/internal/modules/auth"
	authhttpapi "system-barbershop/internal/modules/auth/httpapi"
	authpostgres "system-barbershop/internal/modules/auth/postgres"
	"system-barbershop/internal/platform/clock"
	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/httpserver"
	"system-barbershop/internal/platform/observability"

	"github.com/go-chi/chi/v5"
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

	router, err := buildRouter(db, logger)
	if err != nil {
		logger.Error(err.Error())
		return errors.New("http: fallo al ensamblar el router")
	}

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

// buildRouter ensambla el *chi.Mux completo de la aplicación: middleware
// base, las tres audiencias y las rutas de cada módulo ya implementado.
// Extraído de run() (sin escuchar ni apagar el servidor HTTP) para que la
// prueba estructural de HU-006 (TestPrivateRouteInventory_* en
// router_inventory_test.go) pueda construir el router REAL de producción,
// caminarlo con chi.Walk y confirmar que toda ruta bajo /api/v1/private
// exige sesión válida (CA-006-04), sin volver a levantar el servidor.
func buildRouter(db *database.DB, logger *slog.Logger) (*chi.Mux, error) {
	// httpserver.NewRouter monta las tres audiencias de
	// docs/04-arquitectura/backend-go.md (/api/v1/public, /api/v1/customer,
	// /api/v1/private) sobre Chi v5 (DEC-034) con el middleware base ya
	// aplicado, y devuelve el subrouter real de /api/v1/private: es el
	// ÚNICO lugar donde se monta middleware y rutas privadas (paso 8,
	// autenticación) para que CA-006-04 se cumpla sin excepciones.
	router, private := httpserver.NewRouter(logger)
	router.Get("/health", httpserver.HealthHandler())
	router.Get("/health/db", httpserver.DatabaseHealthHandler(db))

	// HU-005: inicio de sesión, público (DEC-055, sin middleware de
	// autenticación: sería circular, CT-003 resuelta). El servicio no
	// importa Chi ni PostgreSQL; solo el repositorio (authpostgres) y el
	// handler (authhttpapi) lo hacen.
	loginService, err := auth.NewLoginService(
		authpostgres.New(db),
		auth.NewArgon2Hasher(),
		auth.NewCryptoTokenGenerator(),
		clock.System{},
	)
	if err != nil {
		return nil, errors.New("auth: fallo al iniciar LoginService")
	}
	loginHandler := authhttpapi.NewLoginHandler(loginService, authhttpapi.DefaultCookieConfig())
	router.Post("/api/v1/public/auth/login", loginHandler.ServeHTTP)

	// HU-006: middleware de sesión (paso 8) montado UNA sola vez sobre el
	// subrouter privado, ANTES de registrar ninguna ruta sobre él (chi
	// exige que Use() preceda a cualquier Get/Post en ese mismo Router).
	// Toda ruta privada futura (agenda, servicios, horario, ajustes) se
	// registra sobre `private`, nunca con el patrón completo sobre
	// `router`: así hereda el middleware sin que nadie tenga que acordarse
	// de repetirlo (CA-006-04).
	sessionService := auth.NewSessionService(authpostgres.New(db), clock.System{})
	sessionMiddleware := authhttpapi.NewSessionMiddleware(sessionService, authhttpapi.DefaultCookieConfig())
	logoutHandler := authhttpapi.NewLogoutHandler(sessionService, authhttpapi.DefaultCookieConfig())

	private.Use(sessionMiddleware.RequireSession)
	private.Post("/auth/logout", logoutHandler.ServeHTTP)

	return router, nil
}
