// Package config carga la configuración de los procesos api y worker desde
// variables de entorno. Es el único paquete autorizado a leer el entorno del
// sistema operativo; el resto de la aplicación recibe un [Config] ya validado.
package config

import (
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
	for _, cidr := range cfg.TrustedProxies {
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			return Config{}, fmt.Errorf("config: APP_TRUSTED_PROXIES contiene un CIDR inválido %q: %w", cidr, err)
		}
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
