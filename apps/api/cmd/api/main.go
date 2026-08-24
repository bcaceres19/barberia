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
	"system-barbershop/internal/modules/notification"
	"system-barbershop/internal/modules/shops"
	shopshttpapi "system-barbershop/internal/modules/shops/httpapi"
	shopspostgres "system-barbershop/internal/modules/shops/postgres"
	"system-barbershop/internal/modules/staff"
	staffhttpapi "system-barbershop/internal/modules/staff/httpapi"
	staffpostgres "system-barbershop/internal/modules/staff/postgres"
	"system-barbershop/internal/platform/clientip"
	"system-barbershop/internal/platform/clock"
	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/httpserver"
	"system-barbershop/internal/platform/idempotency"
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

	router, err := buildRouter(db, logger, cfg)
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
func buildRouter(db *database.DB, logger *slog.Logger, cfg config.Config) (*chi.Mux, error) {
	// httpserver.NewRouter monta las tres audiencias de
	// docs/04-arquitectura/backend-go.md (/api/v1/public, /api/v1/customer,
	// /api/v1/private) sobre Chi v5 (DEC-034) con el middleware base ya
	// aplicado, y devuelve el subrouter real de /api/v1/private: es el
	// ÚNICO lugar donde se monta middleware y rutas privadas (paso 8,
	// autenticación) para que CA-006-04 se cumpla sin excepciones.
	router, private := httpserver.NewRouter(logger)
	router.Get("/health", httpserver.HealthHandler())
	router.Get("/health/db", httpserver.DatabaseHealthHandler(db))

	trustedProxies, err := clientip.ParseTrustedProxies(cfg.TrustedProxies)
	if err != nil {
		return nil, errors.New("clientip: fallo al parsear APP_TRUSTED_PROXIES")
	}
	hmacSecret := []byte(cfg.AuthHMACSecret)

	// HU-007: defensa escalonada contra abuso del login. throttleService
	// registra el intento y decide el escalamiento ANTES de que
	// LoginService resuelva tenant o evalúe contraseña (DEC-061, CA-007-02).
	throttleService := auth.NewThrottleService(
		authpostgres.NewThrottleRepository(db),
		auth.ThrottleConfig{
			WindowSeconds:     cfg.LoginThrottleWindowSeconds,
			EscalationSeconds: cfg.LoginThrottleEscalationSeconds,
			Threshold:         cfg.LoginThrottleThreshold,
			RetentionSeconds:  cfg.LoginThrottleRetentionSeconds,
		},
		hmacSecret,
		clock.System{},
	)

	// HU-005: inicio de sesión, público (DEC-055, sin middleware de
	// autenticación: sería circular, CT-003 resuelta). El servicio no
	// importa Chi ni PostgreSQL; solo el repositorio (authpostgres) y el
	// handler (authhttpapi) lo hacen.
	loginService, err := auth.NewLoginService(
		authpostgres.New(db),
		auth.NewArgon2Hasher(),
		auth.NewCryptoTokenGenerator(),
		clock.System{},
		throttleService,
	)
	if err != nil {
		return nil, errors.New("auth: fallo al iniciar LoginService")
	}
	loginHandler := authhttpapi.NewLoginHandler(loginService, authhttpapi.DefaultCookieConfig(), trustedProxies)
	router.Post("/api/v1/public/auth/login", loginHandler.ServeHTTP)

	// HU-007: reto telefónico que desbloquea el login tras el escalamiento
	// (DEC-062). sender es un marcador de posición (auth.LoggingPhoneCodeSender):
	// el adaptador real de WhatsApp es DEC-066, decisión de HU-008.
	var phoneSender auth.PhoneCodeSender = auth.NewLoggingPhoneCodeSender(logger)
	if capturePath := os.Getenv("APP_PHONE_CHALLENGE_CAPTURE_FILE"); capturePath != "" {
		// Doble candado de entorno (aquí y en el propio nombre de la
		// variable): la captura de código en claro en un archivo NUNCA
		// puede activarse fuera de local/test, sin importar qué valor
		// llegue por variable de entorno en un despliegue real.
		if cfg.Environment != "local" && cfg.Environment != "test" {
			return nil, errors.New(
				"auth: APP_PHONE_CHALLENGE_CAPTURE_FILE solo puede usarse en local/test (uso exclusivo de e2e)")
		}
		phoneSender = auth.NewCapturingPhoneCodeSender(phoneSender, capturePath)
	}
	phoneChallengeService := auth.NewPhoneChallengeService(
		authpostgres.NewPhoneChallengeRepository(db),
		auth.NewCryptoPhoneCodeGenerator(),
		phoneSender,
		auth.PhoneChallengeConfig{
			ExpiresSeconds:        cfg.PhoneChallengeExpiresSeconds,
			RateWindowSeconds:     cfg.PhoneChallengeRateWindowSeconds,
			RateMaxActive:         cfg.PhoneChallengeRateMaxActive,
			ResendCooldownSeconds: cfg.PhoneChallengeResendCooldownSeconds,
		},
		hmacSecret,
	)
	challengeHandler := authhttpapi.NewChallengeHandler(phoneChallengeService, throttleService, trustedProxies)
	challengeVerifyHandler := authhttpapi.NewChallengeVerifyHandler(phoneChallengeService, throttleService, trustedProxies)
	router.Post("/api/v1/public/auth/challenge", challengeHandler.ServeHTTP)
	router.Post("/api/v1/public/auth/challenge/verify", challengeVerifyHandler.ServeHTTP)

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
	// HU-012 (DEC-060): lectura no destructiva de contexto de sesión, para
	// que el frontend rehidrate barbería activa/expiración al abrir o
	// recargar la aplicación. Se registra sobre el mismo subrouter privado,
	// después del middleware de sesión, igual que logout (CA-006-04).
	sessionContextHandler := authhttpapi.NewSessionContextHandler(sessionService)

	private.Use(sessionMiddleware.RequireSession)
	private.Post("/auth/logout", logoutHandler.ServeHTTP)
	private.Get("/auth/session", sessionContextHandler.ServeHTTP)

	// HU-020: configuración básica de la barbería activa (nombre, zona
	// IANA, contacto opcional). Registrado sobre el mismo subrouter
	// privado, después del middleware de sesión, igual que las rutas de
	// auth (CA-006-04).
	shopService := shops.NewService(shopspostgres.New(db))
	getBarbershopSettingsHandler := shopshttpapi.NewGetBarbershopSettingsHandler(shopService)
	updateBarbershopSettingsHandler := shopshttpapi.NewUpdateBarbershopSettingsHandler(shopService)
	private.Get("/settings/barbershop", getBarbershopSettingsHandler.ServeHTTP)
	private.Patch("/settings/barbershop", updateBarbershopSettingsHandler.ServeHTTP)

	// HU-021: registro y listado de barberos de la barbería activa.
	// staffpostgres.New recibe el mismo idempotency.SQLCoordinator real que
	// protege el alta (RN-IDE-01, DEC-043), coordinado dentro de la misma
	// InTenantTx que el INSERT (staff/postgres/repository.go).
	staffService := staff.NewService(staffpostgres.New(db, idempotency.NewSQLCoordinator()))
	listBarbersHandler := staffhttpapi.NewListBarbersHandler(staffService)
	getBarberHandler := staffhttpapi.NewGetBarberHandler(staffService)
	createBarberHandler := staffhttpapi.NewCreateBarberHandler(staffService)
	renameBarberHandler := staffhttpapi.NewRenameBarberHandler(staffService)
	private.Get("/barbers", listBarbersHandler.ServeHTTP)
	private.Post("/barbers", createBarberHandler.ServeHTTP)
	private.Get("/barbers/{barberId}", getBarberHandler.ServeHTTP)
	private.Patch("/barbers/{barberId}", renameBarberHandler.ServeHTTP)

	// HU-008 (DEC-063-066): recuperación de acceso con código de un solo
	// uso. sender es el adaptador dual de Meta WhatsApp Cloud API + Resend
	// cuando hay credenciales configuradas; sin ellas (típicamente
	// local/test, donde config.Load no las exige) se usa el marcador de
	// posición documentado, igual patrón que el reto telefónico de HU-007.
	var recoverySender auth.RecoveryCodeSender
	if cfg.MetaWhatsAppPhoneNumberID != "" && cfg.MetaWhatsAppAccessToken != "" && cfg.MetaWhatsAppTemplateName != "" &&
		cfg.ResendAPIKey != "" && cfg.ResendFromAddress != "" {
		recoverySender = notification.NewDualChannelRecoverySender(
			notification.NewMetaWhatsAppSender(notification.MetaWhatsAppConfig{
				APIVersion:    cfg.MetaWhatsAppAPIVersion,
				PhoneNumberID: cfg.MetaWhatsAppPhoneNumberID,
				AccessToken:   cfg.MetaWhatsAppAccessToken,
				TemplateName:  cfg.MetaWhatsAppTemplateName,
				LanguageCode:  cfg.MetaWhatsAppLanguageCode,
			}, nil),
			notification.NewResendEmailSender(notification.ResendEmailConfig{
				APIKey:      cfg.ResendAPIKey,
				FromAddress: cfg.ResendFromAddress,
				Subject:     cfg.ResendSubject,
			}, nil),
		)
	} else {
		recoverySender = auth.NewLoggingRecoveryCodeSender(logger)
	}
	if capturePath := os.Getenv("APP_RECOVERY_CAPTURE_FILE"); capturePath != "" {
		// Mismo doble candado de entorno que APP_PHONE_CHALLENGE_CAPTURE_FILE
		// de HU-007: la captura en claro NUNCA puede activarse fuera de
		// local/test, sin importar qué valor llegue por variable de entorno
		// en un despliegue real.
		if cfg.Environment != "local" && cfg.Environment != "test" {
			return nil, errors.New(
				"auth: APP_RECOVERY_CAPTURE_FILE solo puede usarse en local/test (uso exclusivo de pruebas de sistema)")
		}
		recoverySender = auth.NewCapturingRecoveryCodeSender(recoverySender, capturePath)
	}
	recoveryService := auth.NewRecoveryService(
		authpostgres.NewRecoveryRepository(db),
		auth.NewCryptoPhoneCodeGenerator(),
		auth.NewCryptoTokenGenerator(),
		recoverySender,
		auth.NewArgon2Hasher(),
		auth.RecoveryConfig{
			CodeExpiresSeconds:       cfg.RecoveryCodeExpiresSeconds,
			CodeMaxAttempts:          cfg.RecoveryCodeMaxAttempts,
			ResendCooldownSeconds:    cfg.RecoveryResendCooldownSeconds,
			ResendWindowSeconds:      cfg.RecoveryResendWindowSeconds,
			ResendMaxPerWindow:       cfg.RecoveryResendMaxPerWindow,
			ResetTokenExpiresSeconds: cfg.RecoveryResetTokenExpiresSeconds,
		},
		hmacSecret,
	)
	recoveryRequestHandler := authhttpapi.NewRecoveryRequestHandler(recoveryService, logger)
	recoveryVerifyHandler := authhttpapi.NewRecoveryVerifyHandler(recoveryService)
	recoveryResetPasswordHandler := authhttpapi.NewRecoveryResetPasswordHandler(recoveryService)
	router.Post("/api/v1/public/auth/recovery/request", recoveryRequestHandler.ServeHTTP)
	router.Post("/api/v1/public/auth/recovery/verify", recoveryVerifyHandler.ServeHTTP)
	router.Post("/api/v1/public/auth/recovery/reset-password", recoveryResetPasswordHandler.ServeHTTP)

	return router, nil
}
