// Command worker arranca el proceso trabajador que enviará recordatorios ya
// persistidos. Reutiliza servicios y repositorios de los módulos, no
// handlers HTTP, según docs/04-arquitectura/stack-despliegue-operacion.md.
package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"system-barbershop/internal/modules/auth"
	authpostgres "system-barbershop/internal/modules/auth/postgres"
	"system-barbershop/internal/modules/googlecalendar"
	googlecalendargoogle "system-barbershop/internal/modules/googlecalendar/google"
	googlecalendarpostgres "system-barbershop/internal/modules/googlecalendar/postgres"
	"system-barbershop/internal/modules/notification"
	notificationpostgres "system-barbershop/internal/modules/notification/postgres"
	"system-barbershop/internal/platform/clock"
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

	// Issue #355 (DEC-126): purga de ventanas vencidas y estados de entrega de
	// WhatsApp. Sin webhook configurado las tablas están vacías y el lote no
	// borra nada.
	whatsAppPurge := notification.NewWhatsAppPurgeService(
		notificationpostgres.NewWhatsAppPurgeRepository(db), cfg.PhoneChallengePurgeLimit,
	)

	// Issue #324 (DEC-102): publicación en Google Calendar. Con las credenciales
	// ausentes el worker sigue purgando y la integración queda desactivada.
	if cfg.GoogleCalendarEnabled() {
		publisher, err := buildGoogleCalendarPublisher(db, logger, cfg)
		if err != nil {
			logger.Error("no se pudo iniciar el publicador de Google Calendar")
			return err
		}
		go publisher.Run(ctx)
		logger.Info("worker: publicación en Google Calendar activa")
	} else if cfg.GoogleCalendarPartiallyConfigured() {
		logger.Warn("google_calendar.config.incomplete",
			"detail", "faltan variables GOOGLE_CALENDAR_*; la publicación queda desactivada")
	}

	logger.Info("worker iniciado", "environment", cfg.Environment, "purge_interval", purgeInterval.String())

	ticker := time.NewTicker(purgeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("worker apagado")
			return nil
		case <-ticker.C:
			if whatsAppDeleted, err := whatsAppPurge.PurgeOnce(ctx); err != nil {
				logger.Error("worker: fallo al purgar los datos del webhook de WhatsApp")
			} else if whatsAppDeleted > 0 {
				logger.Info("worker: purga de WhatsApp completada", "whatsapp_deleted", whatsAppDeleted)
			}
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

// buildGoogleCalendarPublisher arma el publicador de Google Calendar. db es el
// pool del worker (barberia_worker): solo ejecuta las funciones de reclamo y
// finalización (DEC-040, DEC-102).
func buildGoogleCalendarPublisher(db *database.DB, logger *slog.Logger, cfg config.Config) (*googlecalendar.Publisher, error) {
	cipher, err := googlecalendar.NewCipherFromKeys(cfg.GoogleCalendarTokenKeyID,
		cfg.GoogleCalendarTokenEncryptionKey, cfg.GoogleCalendarPreviousKeys)
	if err != nil {
		return nil, errors.New("googlecalendar: fallo al iniciar el cifrador")
	}
	return googlecalendar.NewPublisher(googlecalendar.PublisherDeps{
		Store: googlecalendarpostgres.NewWorkerRepository(db),
		API:   googlecalendargoogle.NewEventsClient("", nil),
		Provider: googlecalendargoogle.New(googlecalendargoogle.Config{
			ClientID:     cfg.GoogleCalendarClientID,
			ClientSecret: cfg.GoogleCalendarClientSecret,
			RedirectURL:  cfg.GoogleCalendarRedirectURI,
		}),
		Cipher: cipher,
		Clock:  clock.System{},
		Config: googlecalendar.DefaultPublisherConfig(),
		Logger: logger,
	}), nil
}
