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
	"time"

	"system-barbershop/internal/modules/auth"
	authpostgres "system-barbershop/internal/modules/auth/postgres"
	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/observability"
)

// purgeInterval es la cadencia del lote de purga de HU-007
// (login_throttle/auth_phone_challenge, CA-007-06). No es un valor
// aprobado por ninguna decisión normativa (ninguna fuente fija una
// cadencia, solo el tamaño del lote vía config.Config); 5 minutos es
// razonable para tablas de retención corta (horas/días) sin generar carga
// perceptible.
const purgeInterval = 5 * time.Minute

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

	// HU-007/HU-008 (CA-007-06, DEC-064): purga en lote de login_throttle,
	// auth_phone_challenge y staff_recovery_code, exclusiva de
	// barberia_worker (DDL-AUT-01, DEC-040). El reclamo de recordatorios
	// con SKIP LOCKED, el envío por canal y los reintentos se agregan junto
	// con la maquinaria general de notificaciones de B5, según
	// docs/05-backend/estandar-base-datos.md.
	purgeService := auth.NewPurgeService(
		authpostgres.NewPurgeRepository(db),
		cfg.LoginThrottlePurgeLimit,
		cfg.PhoneChallengePurgeLimit,
		cfg.RecoveryCodePurgeLimit,
	)

	logger.Info("worker iniciado", "environment", cfg.Environment, "purge_interval", purgeInterval.String())

	ticker := time.NewTicker(purgeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("worker apagado")
			return nil
		case <-ticker.C:
			loginThrottleDeleted, phoneChallengeDeleted, recoveryCodeDeleted, err := purgeService.PurgeOnce(ctx)
			if err != nil {
				logger.Error("worker: fallo al purgar login_throttle/auth_phone_challenge/staff_recovery_code")
				continue
			}
			if loginThrottleDeleted > 0 || phoneChallengeDeleted > 0 || recoveryCodeDeleted > 0 {
				logger.Info("worker: purga completada",
					"login_throttle_deleted", loginThrottleDeleted,
					"phone_challenge_deleted", phoneChallengeDeleted,
					"recovery_code_deleted", recoveryCodeDeleted,
				)
			}
		}
	}
}
