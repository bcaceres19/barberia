// Package config carga la configuración de los procesos api y worker desde
// variables de entorno. Es el único paquete autorizado a leer el entorno del
// sistema operativo; el resto de la aplicación recibe un [Config] ya validado.
package config

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// entornosSinTLSObligatorio son los únicos ambientes donde una URL de base
// de datos sin TLS (o el DSN de worker ausente) no hace fallar Load(). Fuera
// de estos dos, un despliegue sin TLS es un error de configuración, no un
// valor por defecto silencioso.
var entornosSinTLSObligatorio = map[string]bool{
	"local": true,
	"test":  true,
}

// DatabaseDSN es una cadena de conexión a PostgreSQL. No es un string
// desnudo: implementa String() y LogValue() para que ni fmt.Sprintf("%+v",
// cfg) ni un logger estructurado puedan filtrar la contraseña. Una URL de
// conexión completa en un log de arranque es una fuga de credenciales.
type DatabaseDSN string

// String redacta la contraseña de la URL antes de imprimirla. Soporta el
// formato URL (postgres://usuario:contraseña@host/...) y el formato de
// palabras clave (host=... password=...); si no reconoce ninguno de los dos
// pero el valor no está vacío, devuelve un marcador fijo en vez de arriesgar
// una fuga parcial.
func (d DatabaseDSN) String() string {
	return redactDSN(string(d))
}

// LogValue implementa slog.LogValuer: garantiza la misma redacción cuando
// el valor se registra con el logger estructurado del proyecto, no solo con
// fmt.
func (d DatabaseDSN) LogValue() any {
	return d.String()
}

var dsnPasswordKeyword = regexp.MustCompile(`(?i)(password|pwd)=\S+`)

func redactDSN(raw string) string {
	if raw == "" {
		return ""
	}
	if u, err := url.Parse(raw); err == nil && u.User != nil {
		if _, hasPassword := u.User.Password(); hasPassword {
			u.User = url.UserPassword(u.User.Username(), "xxxxx")
		}
		return u.String()
	}
	if dsnPasswordKeyword.MatchString(raw) {
		return dsnPasswordKeyword.ReplaceAllString(raw, "$1=xxxxx")
	}
	return "[dsn redactado]"
}

// Config reúne los valores de configuración que necesitan los procesos api y
// worker. Los campos se agregan solo cuando un componente concreto los
// consume; este struct no es un depósito genérico de variables de entorno.
type Config struct {
	// Environment identifica el ambiente en ejecución: local, test, pilot o
	// production. No determina reglas de negocio, pero SÍ determina qué
	// validaciones de seguridad son obligatorias (TLS, DSN de worker
	// separado): fuera de local/test se exigen ambas.
	Environment string

	// HTTPAddr es la dirección donde escucha el servidor HTTP del proceso api.
	HTTPAddr string

	// LogLevel controla el nivel mínimo de severidad que emite el logger de
	// observabilidad.
	LogLevel string

	// DatabaseURL es la URL de conexión a PostgreSQL que usa el proceso api,
	// conectado como barberia_app.
	DatabaseURL DatabaseDSN

	// WorkerDatabaseURL es la URL de conexión a PostgreSQL que usa el
	// proceso worker, conectado como barberia_worker (DEC-040: roles
	// separados). Deliberadamente distinta de DatabaseURL: nada en este
	// paquete permite que ambos procesos compartan la misma variable.
	// Obligatoria fuera de local/test.
	WorkerDatabaseURL DatabaseDSN

	// DatabaseMaxConns es el número máximo de conexiones en el pool.
	DatabaseMaxConns int

	// DatabaseMinConns es el número mínimo de conexiones mantenidas en el pool.
	DatabaseMinConns int

	// DatabaseMaxConnLifetime es el tiempo máximo de vida de una conexión.
	DatabaseMaxConnLifetime time.Duration

	// DatabaseMaxConnIdleTime es el tiempo máximo de inactividad antes de
	// cerrar una conexión.
	DatabaseMaxConnIdleTime time.Duration

	// DatabaseConnectTimeout es el timeout para establecer una conexión.
	DatabaseConnectTimeout time.Duration

	// DatabaseStatementTimeout es el timeout por sentencia a nivel de
	// conexión. Una consulta que se cuelga sin límite bloquea una conexión
	// del pool y degrada todo el proceso.
	DatabaseStatementTimeout time.Duration

	// AuthHMACSecret firma el HMAC-SHA256 de la IP (login_throttle.ip_hash)
	// y del código del reto telefónico (auth_phone_challenge.code_hash),
	// HU-007 (DEC-062). Nunca se registra ni se expone; Load exige un largo
	// mínimo para que no sea un valor trivial de adivinar.
	AuthHMACSecret string

	// TrustedProxies son los CIDR de los únicos proxies inmediatos cuya
	// cabecera X-Forwarded-For se acepta al resolver la IP real de una
	// solicitud (HU-007). Vacío por defecto: sin proxies confiables, se usa
	// siempre RemoteAddr, nunca una cabecera que el cliente puede falsificar.
	TrustedProxies []string

	// LoginThrottleWindowSeconds es la ventana deslizante del contador por
	// IP (CA-007-05, valor inicial 900 = 15 min, DEC-052).
	LoginThrottleWindowSeconds int
	// LoginThrottleEscalationSeconds es cuánto dura el escalamiento una vez
	// activado (valor inicial 86400 = 24 h, DEC-052).
	LoginThrottleEscalationSeconds int
	// LoginThrottleThreshold es la cantidad de solicitudes permitidas SIN
	// reto dentro de la ventana; la solicitud threshold+1 lo exige
	// (valor inicial 5, DEC-061).
	LoginThrottleThreshold int
	// LoginThrottleRetentionSeconds es cuánto se conserva la fila de
	// contador antes de purgarse (debe ser >= LoginThrottleEscalationSeconds).
	LoginThrottleRetentionSeconds int
	// LoginThrottlePurgeLimit acota cuántas filas vencidas purga el worker
	// por lote (login_throttle_purge_expired).
	LoginThrottlePurgeLimit int

	// PhoneChallengeExpiresSeconds es la vigencia del código del reto
	// (valor inicial 300 = 5 min, DEC-062).
	PhoneChallengeExpiresSeconds int
	// PhoneChallengeMaxAttempts es el máximo de intentos fallidos antes de
	// invalidar el código (valor inicial 5, DEC-062).
	PhoneChallengeMaxAttempts int
	// PhoneChallengeRateWindowSeconds es la ventana en la que se cuentan las
	// solicitudes activas del reto por IP (valor inicial 900 = 15 min,
	// DEC-062, misma ventana que el umbral de login).
	PhoneChallengeRateWindowSeconds int
	// PhoneChallengeRateMaxActive es el máximo de solicitudes del reto por
	// IP dentro de esa ventana (valor inicial 3, DEC-062).
	PhoneChallengeRateMaxActive int
	// PhoneChallengeResendCooldownSeconds es el mínimo entre dos solicitudes
	// consecutivas del reto desde la misma IP (valor inicial 60, DEC-062).
	PhoneChallengeResendCooldownSeconds int
	// PhoneChallengePurgeLimit acota cuántos retos vencidos purga el worker
	// por lote (auth_phone_challenge_purge_expired).
	PhoneChallengePurgeLimit int

	// RecoveryCodeExpiresSeconds es la vigencia del código de recuperación
	// (valor inicial 900 = 15 min, DEC-064).
	RecoveryCodeExpiresSeconds int
	// RecoveryCodeMaxAttempts es el máximo de intentos fallidos antes de
	// invalidar el código (valor inicial 5, DEC-064).
	RecoveryCodeMaxAttempts int
	// RecoveryResendCooldownSeconds es el mínimo entre dos solicitudes
	// consecutivas para la misma cuenta (valor inicial 60, DEC-064).
	RecoveryResendCooldownSeconds int
	// RecoveryResendWindowSeconds es la ventana en la que se cuentan los
	// códigos solicitados por cuenta (valor inicial 3600 = 1 h, DEC-064).
	RecoveryResendWindowSeconds int
	// RecoveryResendMaxPerWindow es el máximo de códigos por cuenta dentro
	// de esa ventana (valor inicial 3, DEC-064).
	RecoveryResendMaxPerWindow int
	// RecoveryResetTokenExpiresSeconds es la vigencia del token de reinicio
	// emitido tras verificar el código (valor inicial 300 = 5 min, DEC-064).
	RecoveryResetTokenExpiresSeconds int
	// RecoveryCodePurgeLimit acota cuántos códigos de recuperación vencidos
	// purga el worker por lote (auth_recovery_purge_expired).
	RecoveryCodePurgeLimit int

	// MetaWhatsAppAPIVersion es la versión de Meta Graph API a usar
	// (DEC-066). No es secreto.
	MetaWhatsAppAPIVersion string
	// MetaWhatsAppPhoneNumberID identifica el número emisor en Meta
	// WhatsApp Cloud API. Secreto operativo (no una contraseña, pero
	// específico del despliegue): nunca se registra.
	MetaWhatsAppPhoneNumberID string
	// MetaWhatsAppAccessToken autentica contra Meta Graph API. Secreto:
	// nunca se registra ni se comitea.
	MetaWhatsAppAccessToken string
	// MetaWhatsAppMode es "template" (plantilla AUTHENTICATION, valor por
	// defecto) o "development" (texto libre). "development" solo se acepta
	// en local/test (DEC-123).
	MetaWhatsAppMode string
	// MetaWhatsAppTemplateName es el nombre de la plantilla de categoría
	// "Authentication" aprobada por Meta para el código de recuperación
	// (DEC-066). Obligatoria en modo "template".
	MetaWhatsAppTemplateName string
	// MetaWhatsAppLanguageCode es el código de idioma de esa plantilla
	// (p. ej. "es" o "es_CO").
	MetaWhatsAppLanguageCode string
	// MetaWhatsAppTestRecipients lista los teléfonos E.164 a los que el modo
	// "development" puede escribir. Obligatoria en ese modo.
	MetaWhatsAppTestRecipients []string

	// OTPProvider selecciona el proveedor de OTP WhatsApp. El único valor
	// aceptado es "meta"; existe para rechazar de forma explícita un
	// despliegue que aún declare el proveedor retirado (DEC-123).
	OTPProvider string

	// ResendAPIKey autentica contra la API de Resend (DEC-066). Secreto:
	// nunca se registra ni se comitea.
	ResendAPIKey string
	// ResendFromAddress es el remitente verificado del correo de
	// recuperación.
	ResendFromAddress string
	// ResendSubject es el asunto fijo del correo de recuperación.
	ResendSubject string

	// PublicWebBaseURL es el origen del frontend público (sin ruta ni barra
	// final) que HU-097 usa para construir el enlace de acceso al turno
	// dentro del correo de confirmación (F-PUB-07, DEC-089/DEC-091). Cadena
	// vacía es una configuración válida (desarrollo sin frontend público
	// desplegado todavía): el correo se envía sin enlace clicable en ese
	// caso, nunca se bloquea el arranque ni la confirmación por su
	// ausencia.
	PublicWebBaseURL string

	// GoogleCalendar* configuran la conexión de cada barbero con SU Google
	// Calendar (DEC-099, DEC-102). Las cuatro primeras son obligatorias para
	// activar la integración: si falta alguna, la integración queda
	// desactivada y el resto del producto funciona (GoogleCalendarEnabled).
	// Secretos: nunca se registran ni se comitean.
	GoogleCalendarClientID     string
	GoogleCalendarClientSecret string
	// GoogleCalendarRedirectURI es la URL registrada en Google Cloud a la que
	// Google devuelve al navegador: la pantalla de retorno del frontend. Exige
	// HTTPS fuera de local/test (DEC-102).
	GoogleCalendarRedirectURI string
	// GoogleCalendarTokenEncryptionKey es la clave AES-256 (32 bytes en
	// base64) con la que se cifra el refresh token (DEC-102).
	GoogleCalendarTokenEncryptionKey string
	// GoogleCalendarTokenKeyID identifica esa clave en cada texto cifrado
	// ("v1" por defecto); cambiarla junto con la clave permite rotar.
	GoogleCalendarTokenKeyID string
	// GoogleCalendarPreviousKeys son claves anteriores, "id:base64" separadas
	// por coma, que siguen sirviendo para DESCIFRAR lo guardado antes de rotar.
	GoogleCalendarPreviousKeys []string
}

// GoogleCalendarEnabled informa si están las cuatro variables que activan la
// integración con Google Calendar. Con alguna ausente la integración queda
// desactivada sin impedir el arranque (DEC-102).
func (c Config) GoogleCalendarEnabled() bool {
	return c.GoogleCalendarClientID != "" && c.GoogleCalendarClientSecret != "" &&
		c.GoogleCalendarRedirectURI != "" && c.GoogleCalendarTokenEncryptionKey != ""
}

// GoogleCalendarPartiallyConfigured informa si hay alguna variable de Google
// Calendar sin que la integración esté completa: casi siempre es un olvido.
func (c Config) GoogleCalendarPartiallyConfigured() bool {
	any := c.GoogleCalendarClientID != "" || c.GoogleCalendarClientSecret != "" ||
		c.GoogleCalendarRedirectURI != "" || c.GoogleCalendarTokenEncryptionKey != ""
	return any && !c.GoogleCalendarEnabled()
}

// Load lee la configuración desde variables de entorno y aplica valores por
// defecto seguros para desarrollo local. Devuelve un error si un valor
// obligatorio falta, o si un valor presente no es válido: una variable mal
// escrita (p. ej. APP_DATABASE_MAX_CONNS=abc) debe impedir el arranque, no
// caer en silencio al valor por defecto.
func Load() (Config, error) {
	environment := getEnv("APP_ENVIRONMENT", "local")
	if environment == "" {
		return Config{}, fmt.Errorf("config: APP_ENVIRONMENT no puede quedar vacío")
	}
	otpProvider := getEnv("OTP_PROVIDER", "meta")

	maxConns, err := getEnvInt("APP_DATABASE_MAX_CONNS", 20)
	if err != nil {
		return Config{}, err
	}
	minConns, err := getEnvInt("APP_DATABASE_MIN_CONNS", 2)
	if err != nil {
		return Config{}, err
	}
	if minConns > maxConns {
		return Config{}, fmt.Errorf(
			"config: APP_DATABASE_MIN_CONNS (%d) no puede ser mayor que APP_DATABASE_MAX_CONNS (%d)",
			minConns, maxConns,
		)
	}

	maxConnLifetime, err := getEnvDuration("APP_DATABASE_MAX_CONN_LIFETIME", "1h")
	if err != nil {
		return Config{}, err
	}
	maxConnIdleTime, err := getEnvDuration("APP_DATABASE_MAX_CONN_IDLE_TIME", "30m")
	if err != nil {
		return Config{}, err
	}
	connectTimeout, err := getEnvDuration("APP_DATABASE_CONNECT_TIMEOUT", "5s")
	if err != nil {
		return Config{}, err
	}
	statementTimeout, err := getEnvDuration("APP_DATABASE_STATEMENT_TIMEOUT", "10s")
	if err != nil {
		return Config{}, err
	}

	throttleWindow, err := getEnvInt("APP_LOGIN_THROTTLE_WINDOW_SECONDS", 900)
	if err != nil {
		return Config{}, err
	}
	throttleEscalation, err := getEnvInt("APP_LOGIN_THROTTLE_ESCALATION_SECONDS", 86400)
	if err != nil {
		return Config{}, err
	}
	throttleThreshold, err := getEnvInt("APP_LOGIN_THROTTLE_THRESHOLD", 5)
	if err != nil {
		return Config{}, err
	}
	throttleRetention, err := getEnvInt("APP_LOGIN_THROTTLE_RETENTION_SECONDS", 172800)
	if err != nil {
		return Config{}, err
	}
	throttlePurgeLimit, err := getEnvInt("APP_LOGIN_THROTTLE_PURGE_LIMIT", 500)
	if err != nil {
		return Config{}, err
	}

	challengeExpires, err := getEnvInt("APP_PHONE_CHALLENGE_EXPIRES_SECONDS", 300)
	if err != nil {
		return Config{}, err
	}
	challengeMaxAttempts, err := getEnvInt("APP_PHONE_CHALLENGE_MAX_ATTEMPTS", 5)
	if err != nil {
		return Config{}, err
	}
	challengeRateWindow, err := getEnvInt("APP_PHONE_CHALLENGE_RATE_WINDOW_SECONDS", 900)
	if err != nil {
		return Config{}, err
	}
	challengeRateMaxActive, err := getEnvInt("APP_PHONE_CHALLENGE_RATE_MAX_ACTIVE", 3)
	if err != nil {
		return Config{}, err
	}
	challengeResendCooldown, err := getEnvInt("APP_PHONE_CHALLENGE_RESEND_COOLDOWN_SECONDS", 60)
	if err != nil {
		return Config{}, err
	}
	challengePurgeLimit, err := getEnvInt("APP_PHONE_CHALLENGE_PURGE_LIMIT", 500)
	if err != nil {
		return Config{}, err
	}

	recoveryCodeDefault := 900
	recoveryCodeExpires, err := getEnvInt("APP_RECOVERY_CODE_EXPIRES_SECONDS", recoveryCodeDefault)
	if err != nil {
		return Config{}, err
	}
	recoveryMaxAttempts, err := getEnvInt("APP_RECOVERY_CODE_MAX_ATTEMPTS", 5)
	if err != nil {
		return Config{}, err
	}
	recoveryResendCooldown, err := getEnvInt("APP_RECOVERY_RESEND_COOLDOWN_SECONDS", 60)
	if err != nil {
		return Config{}, err
	}
	recoveryResendWindow, err := getEnvInt("APP_RECOVERY_RESEND_WINDOW_SECONDS", 3600)
	if err != nil {
		return Config{}, err
	}
	recoveryResendMaxPerWindow, err := getEnvInt("APP_RECOVERY_RESEND_MAX_PER_WINDOW", 3)
	if err != nil {
		return Config{}, err
	}
	recoveryResetTokenExpires, err := getEnvInt("APP_RECOVERY_RESET_TOKEN_EXPIRES_SECONDS", 300)
	if err != nil {
		return Config{}, err
	}
	recoveryCodePurgeLimit, err := getEnvInt("APP_RECOVERY_CODE_PURGE_LIMIT", 500)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Environment:              environment,
		HTTPAddr:                 getEnv("APP_HTTP_ADDR", ":8080"),
		LogLevel:                 getEnv("APP_LOG_LEVEL", "info"),
		DatabaseURL:              DatabaseDSN(getEnv("APP_DATABASE_URL", "")),
		WorkerDatabaseURL:        DatabaseDSN(getEnv("APP_WORKER_DATABASE_URL", "")),
		DatabaseMaxConns:         maxConns,
		DatabaseMinConns:         minConns,
		DatabaseMaxConnLifetime:  maxConnLifetime,
		DatabaseMaxConnIdleTime:  maxConnIdleTime,
		DatabaseConnectTimeout:   connectTimeout,
		DatabaseStatementTimeout: statementTimeout,

		AuthHMACSecret: getEnv("APP_AUTH_HMAC_SECRET", ""),
		TrustedProxies: getEnvCSV("APP_TRUSTED_PROXIES"),

		LoginThrottleWindowSeconds:     throttleWindow,
		LoginThrottleEscalationSeconds: throttleEscalation,
		LoginThrottleThreshold:         throttleThreshold,
		LoginThrottleRetentionSeconds:  throttleRetention,
		LoginThrottlePurgeLimit:        throttlePurgeLimit,

		PhoneChallengeExpiresSeconds:        challengeExpires,
		PhoneChallengeMaxAttempts:           challengeMaxAttempts,
		PhoneChallengeRateWindowSeconds:     challengeRateWindow,
		PhoneChallengeRateMaxActive:         challengeRateMaxActive,
		PhoneChallengeResendCooldownSeconds: challengeResendCooldown,
		PhoneChallengePurgeLimit:            challengePurgeLimit,

		RecoveryCodeExpiresSeconds:       recoveryCodeExpires,
		RecoveryCodeMaxAttempts:          recoveryMaxAttempts,
		RecoveryResendCooldownSeconds:    recoveryResendCooldown,
		RecoveryResendWindowSeconds:      recoveryResendWindow,
		RecoveryResendMaxPerWindow:       recoveryResendMaxPerWindow,
		RecoveryResetTokenExpiresSeconds: recoveryResetTokenExpires,
		RecoveryCodePurgeLimit:           recoveryCodePurgeLimit,

		MetaWhatsAppAPIVersion:     getEnv("APP_META_WHATSAPP_API_VERSION", "v24.0"),
		MetaWhatsAppPhoneNumberID:  getEnv("APP_META_WHATSAPP_PHONE_NUMBER_ID", ""),
		MetaWhatsAppAccessToken:    getEnv("APP_META_WHATSAPP_ACCESS_TOKEN", ""),
		MetaWhatsAppMode:           getEnv("APP_META_WHATSAPP_MODE", "template"),
		MetaWhatsAppTemplateName:   getEnv("APP_META_WHATSAPP_TEMPLATE_NAME", ""),
		MetaWhatsAppTestRecipients: getEnvCSV("APP_META_WHATSAPP_TEST_RECIPIENTS"),
		MetaWhatsAppLanguageCode:   getEnv("APP_META_WHATSAPP_LANGUAGE_CODE", "es"),

		OTPProvider: otpProvider,

		ResendAPIKey:      getEnv("APP_RESEND_API_KEY", ""),
		ResendFromAddress: getEnv("APP_RESEND_FROM_ADDRESS", ""),
		ResendSubject:     getEnv("APP_RESEND_SUBJECT", "Código de recuperación de acceso"),

		PublicWebBaseURL: getEnv("APP_PUBLIC_WEB_BASE_URL", ""),

		GoogleCalendarClientID:           getEnv("GOOGLE_CALENDAR_CLIENT_ID", ""),
		GoogleCalendarClientSecret:       getEnv("GOOGLE_CALENDAR_CLIENT_SECRET", ""),
		GoogleCalendarRedirectURI:        getEnv("GOOGLE_CALENDAR_REDIRECT_URI", ""),
		GoogleCalendarTokenEncryptionKey: getEnv("GOOGLE_CALENDAR_TOKEN_ENCRYPTION_KEY", ""),
		GoogleCalendarTokenKeyID:         getEnv("GOOGLE_CALENDAR_TOKEN_ENCRYPTION_KEY_ID", "v1"),
		GoogleCalendarPreviousKeys:       getEnvCSV("GOOGLE_CALENDAR_TOKEN_ENCRYPTION_KEY_PREVIOUS"),
	}
	if err := validateGoogleCalendar(cfg); err != nil {
		return Config{}, err
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("config: APP_DATABASE_URL no puede quedar vacío")
	}

	// minHMACSecretLen: por debajo de este largo, HMAC-SHA256 no ofrece
	// margen de seguridad razonable contra un atacante que solo necesita
	// reconstruir la clave, no invertir el hash (HU-007, DDL-AUT-01).
	const minHMACSecretLen = 32
	if len(cfg.AuthHMACSecret) < minHMACSecretLen {
		return Config{}, fmt.Errorf(
			"config: APP_AUTH_HMAC_SECRET debe tener al menos %d caracteres", minHMACSecretLen,
		)
	}
	if cfg.LoginThrottleWindowSeconds < 1 || cfg.LoginThrottleEscalationSeconds < 1 ||
		cfg.LoginThrottleThreshold < 1 {
		return Config{}, fmt.Errorf("config: parámetros de APP_LOGIN_THROTTLE_* fuera de rango")
	}
	if cfg.LoginThrottleRetentionSeconds < cfg.LoginThrottleEscalationSeconds {
		return Config{}, fmt.Errorf(
			"config: APP_LOGIN_THROTTLE_RETENTION_SECONDS no puede ser menor que " +
				"APP_LOGIN_THROTTLE_ESCALATION_SECONDS (una fila purgada perdería un escalamiento vigente)",
		)
	}
	if cfg.PhoneChallengeExpiresSeconds < 1 || cfg.PhoneChallengeMaxAttempts < 1 ||
		cfg.PhoneChallengeRateWindowSeconds < 1 || cfg.PhoneChallengeRateMaxActive < 1 ||
		cfg.PhoneChallengeResendCooldownSeconds < 1 {
		return Config{}, fmt.Errorf("config: parámetros de APP_PHONE_CHALLENGE_* fuera de rango")
	}
	if cfg.LoginThrottlePurgeLimit < 1 || cfg.LoginThrottlePurgeLimit > 1000 ||
		cfg.PhoneChallengePurgeLimit < 1 || cfg.PhoneChallengePurgeLimit > 1000 {
		return Config{}, fmt.Errorf(
			"config: APP_LOGIN_THROTTLE_PURGE_LIMIT/APP_PHONE_CHALLENGE_PURGE_LIMIT deben estar entre 1 y 1000",
		)
	}
	if cfg.RecoveryCodeExpiresSeconds < 1 || cfg.RecoveryCodeMaxAttempts < 1 || cfg.RecoveryCodeMaxAttempts > 10 ||
		cfg.RecoveryResendCooldownSeconds < 1 || cfg.RecoveryResendWindowSeconds < 1 ||
		cfg.RecoveryResendMaxPerWindow < 1 || cfg.RecoveryResetTokenExpiresSeconds < 1 {
		return Config{}, fmt.Errorf("config: parámetros de APP_RECOVERY_* fuera de rango")
	}
	if cfg.RecoveryCodePurgeLimit < 1 || cfg.RecoveryCodePurgeLimit > 1000 {
		return Config{}, fmt.Errorf("config: APP_RECOVERY_CODE_PURGE_LIMIT debe estar entre 1 y 1000")
	}
	for _, cidr := range cfg.TrustedProxies {
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			return Config{}, fmt.Errorf("config: APP_TRUSTED_PROXIES contiene un CIDR inválido %q: %w", cidr, err)
		}
	}

	if cfg.OTPProvider != "meta" {
		return Config{}, fmt.Errorf("config: OTP_PROVIDER solo admite meta (Twilio fue retirado, DEC-123)")
	}
	if err := validateMetaWhatsApp(cfg); err != nil {
		return Config{}, err
	}

	requiresHardening := !entornosSinTLSObligatorio[cfg.Environment]

	if requiresHardening {
		if cfg.WorkerDatabaseURL == "" {
			return Config{}, fmt.Errorf(
				"config: APP_WORKER_DATABASE_URL es obligatorio fuera de local/test " +
					"(DEC-040: api y worker no comparten credencial)",
			)
		}
		if cfg.WorkerDatabaseURL == cfg.DatabaseURL {
			return Config{}, fmt.Errorf(
				"config: APP_DATABASE_URL y APP_WORKER_DATABASE_URL no pueden ser iguales " +
					"fuera de local/test (DEC-040: api y worker usan roles distintos)",
			)
		}
		if err := requireTLS(string(cfg.DatabaseURL), "APP_DATABASE_URL"); err != nil {
			return Config{}, err
		}
		if err := requireTLS(string(cfg.WorkerDatabaseURL), "APP_WORKER_DATABASE_URL"); err != nil {
			return Config{}, err
		}

		// DEC-066: fuera de local/test, HU-008 exige el adaptador real de
		// entrega (Meta WhatsApp Cloud API + Resend), no el marcador de
		// posición que solo registra en el log. Faltar cualquiera de estos
		// valores debe impedir el arranque, igual que un DSN sin TLS.
		if cfg.OTPProvider == "meta" && (cfg.MetaWhatsAppPhoneNumberID == "" || cfg.MetaWhatsAppAccessToken == "" || cfg.MetaWhatsAppTemplateName == "") {
			return Config{}, fmt.Errorf(
				"config: APP_META_WHATSAPP_PHONE_NUMBER_ID/APP_META_WHATSAPP_ACCESS_TOKEN/" +
					"APP_META_WHATSAPP_TEMPLATE_NAME son obligatorios fuera de local/test (DEC-066)",
			)
		}
		if cfg.ResendAPIKey == "" || cfg.ResendFromAddress == "" {
			return Config{}, fmt.Errorf(
				"config: APP_RESEND_API_KEY/APP_RESEND_FROM_ADDRESS son obligatorios fuera de local/test (DEC-066)",
			)
		}
	}

	return cfg, nil
}

// requireTLS rechaza sslmode=disable (o su ausencia interpretada como tal
// por libpq) fuera de local/test. No valida el certificado en sí: eso lo
// hace el driver al conectar. Aquí solo se evita que un despliegue real
// quede configurado, por defecto u omisión, sin TLS.
func requireTLS(dsn, varName string) error {
	u, err := url.Parse(dsn)
	if err != nil {
		return fmt.Errorf("config: %s no es una URL válida: %w", varName, err)
	}
	sslmode := u.Query().Get("sslmode")
	if sslmode == "" || sslmode == "disable" {
		return fmt.Errorf(
			"config: %s requiere sslmode distinto de 'disable' fuera de local/test", varName,
		)
	}
	return nil
}

// MetaWhatsAppConfigured indica si hay credenciales suficientes para el modo
// elegido. Sin ellas, local/test cae en el remitente que solo registra.
func (c Config) MetaWhatsAppConfigured() bool {
	if c.MetaWhatsAppPhoneNumberID == "" || c.MetaWhatsAppAccessToken == "" {
		return false
	}
	if c.MetaWhatsAppMode == "development" {
		return len(c.MetaWhatsAppTestRecipients) > 0
	}
	return c.MetaWhatsAppTemplateName != ""
}

var metaTestRecipientPattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

// validateMetaWhatsApp impide un despliegue a medias: las credenciales se
// configuran juntas y el texto libre de desarrollo nunca llega a un ambiente
// con TLS obligatorio (DEC-123).
func validateMetaWhatsApp(cfg Config) error {
	if cfg.MetaWhatsAppMode != "template" && cfg.MetaWhatsAppMode != "development" {
		return fmt.Errorf("config: APP_META_WHATSAPP_MODE debe ser template o development")
	}
	if cfg.MetaWhatsAppMode == "development" {
		if cfg.Environment != "local" && cfg.Environment != "test" {
			return fmt.Errorf("config: APP_META_WHATSAPP_MODE=development solo puede usarse en local/test; use template")
		}
		if len(cfg.MetaWhatsAppTestRecipients) == 0 {
			return fmt.Errorf("config: APP_META_WHATSAPP_TEST_RECIPIENTS es obligatoria con APP_META_WHATSAPP_MODE=development")
		}
	}
	for _, phone := range cfg.MetaWhatsAppTestRecipients {
		if !metaTestRecipientPattern.MatchString(phone) {
			return fmt.Errorf("config: APP_META_WHATSAPP_TEST_RECIPIENTS debe contener teléfonos E.164 separados por coma")
		}
	}

	needsTemplate := cfg.MetaWhatsAppMode == "template"
	anySet := cfg.MetaWhatsAppPhoneNumberID != "" || cfg.MetaWhatsAppAccessToken != "" ||
		(needsTemplate && cfg.MetaWhatsAppTemplateName != "")
	complete := cfg.MetaWhatsAppPhoneNumberID != "" && cfg.MetaWhatsAppAccessToken != "" &&
		(!needsTemplate || cfg.MetaWhatsAppTemplateName != "")
	if anySet && !complete {
		return fmt.Errorf(
			"config: APP_META_WHATSAPP_PHONE_NUMBER_ID/APP_META_WHATSAPP_ACCESS_TOKEN" +
				" (y APP_META_WHATSAPP_TEMPLATE_NAME en modo template) deben configurarse juntas o permanecer ausentes",
		)
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

// getEnvCSV parsea una variable de entorno como lista separada por comas,
// recortando espacios y descartando elementos vacíos (una coma sobrante no
// produce un CIDR "" que después fallaría a validar). Ausente o vacía
// devuelve una lista vacía, nunca nil vs. []string{} de forma inconsistente.
func getEnvCSV(key string) []string {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// getEnvInt parsea una variable de entorno como entero. Si la variable está
// ausente o vacía, devuelve el valor por defecto; si está presente pero no
// es un entero válido, devuelve error en vez de caer al valor por defecto
// en silencio.
func getEnvInt(key string, fallback int) (int, error) {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback, nil
	}
	var result int
	if _, err := fmt.Sscanf(value, "%d", &result); err != nil {
		return 0, fmt.Errorf("config: %s=%q no es un entero válido: %w", key, value, err)
	}
	return result, nil
}

// getEnvDuration parsea una variable de entorno como duración de Go. Si la
// variable está ausente o vacía, devuelve el valor por defecto; si está
// presente pero no es una duración válida, devuelve error.
func getEnvDuration(key string, fallback string) (time.Duration, error) {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		d, err := time.ParseDuration(fallback)
		if err != nil {
			return 0, fmt.Errorf("config: valor por defecto de %s (%q) inválido: %w", key, fallback, err)
		}
		return d, nil
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("config: %s=%q no es una duración válida: %w", key, value, err)
	}
	return d, nil
}

// validateGoogleCalendar rechaza valores PRESENTES pero mal formados: una clave
// que no mide 32 bytes o una URI de retorno sin HTTPS fuera de local/test no
// deben degradarse en silencio a «integración desactivada». La AUSENCIA de
// variables sí desactiva la integración sin fallar (DEC-102).
func validateGoogleCalendar(cfg Config) error {
	if cfg.GoogleCalendarRedirectURI != "" {
		u, err := url.Parse(cfg.GoogleCalendarRedirectURI)
		if err != nil || u.Host == "" || u.Fragment != "" {
			return fmt.Errorf("config: GOOGLE_CALENDAR_REDIRECT_URI no es una URL válida")
		}
		localHTTP := entornosSinTLSObligatorio[cfg.Environment] && u.Scheme == "http" &&
			(u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1")
		if u.Scheme != "https" && !localHTTP {
			return fmt.Errorf(
				"config: GOOGLE_CALENDAR_REDIRECT_URI debe usar HTTPS (http://localhost solo se admite en local/test)")
		}
	}
	if cfg.GoogleCalendarTokenEncryptionKey != "" {
		if !validGoogleKey(cfg.GoogleCalendarTokenEncryptionKey) {
			return fmt.Errorf("config: GOOGLE_CALENDAR_TOKEN_ENCRYPTION_KEY debe ser una clave de 32 bytes en base64")
		}
		if strings.TrimSpace(cfg.GoogleCalendarTokenKeyID) == "" {
			return fmt.Errorf("config: GOOGLE_CALENDAR_TOKEN_ENCRYPTION_KEY_ID no puede quedar vacío")
		}
	}
	seen := map[string]bool{cfg.GoogleCalendarTokenKeyID: true}
	for _, entry := range cfg.GoogleCalendarPreviousKeys {
		id, key, ok := strings.Cut(entry, ":")
		if !ok || id == "" || !validGoogleKey(key) {
			return fmt.Errorf("config: GOOGLE_CALENDAR_TOKEN_ENCRYPTION_KEY_PREVIOUS debe tener la forma id:clave-base64 con claves de 32 bytes")
		}
		if seen[id] {
			return fmt.Errorf("config: GOOGLE_CALENDAR_TOKEN_ENCRYPTION_KEY_PREVIOUS repite el identificador de clave %q", id)
		}
		seen[id] = true
	}
	return nil
}

func validGoogleKey(encoded string) bool {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	return err == nil && len(raw) == 32
}
