// Package config carga la configuración de los procesos api y worker desde
// variables de entorno. Es el único paquete autorizado a leer el entorno del
// sistema operativo; el resto de la aplicación recibe un [Config] ya validado.
package config

import (
	"fmt"
	"os"
	"time"
)

// Config reúne los valores de configuración que necesitan los procesos api y
// worker. Los campos se agregan solo cuando un componente concreto los
// consume; este struct no es un depósito genérico de variables de entorno.
type Config struct {
	// Environment identifica el ambiente en ejecución: local, test, pilot o
	// production. No determina reglas de negocio, solo comportamiento
	// operativo como el nivel de log.
	Environment string

	// HTTPAddr es la dirección donde escucha el servidor HTTP del proceso api.
	HTTPAddr string

	// LogLevel controla el nivel mínimo de severidad que emite el logger de
	// observabilidad.
	LogLevel string

	// DatabaseURL es la URL de conexión a PostgreSQL. Es un secreto: su tipo
	// implementa String() redactando la contraseña. Una URL completa en un
	// log de arranque es una fuga de credenciales.
	DatabaseURL string

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
}

// Load lee la configuración desde variables de entorno y aplica valores por
// defecto seguros para desarrollo local. Devuelve un error si un valor
// obligatorio falta o no es válido.
func Load() (Config, error) {
	cfg := Config{
		Environment:              getEnv("APP_ENVIRONMENT", "local"),
		HTTPAddr:                 getEnv("APP_HTTP_ADDR", ":8080"),
		LogLevel:                 getEnv("APP_LOG_LEVEL", "info"),
		DatabaseURL:              getEnv("APP_DATABASE_URL", ""),
		DatabaseMaxConns:         getEnvInt("APP_DATABASE_MAX_CONNS", 20),
		DatabaseMinConns:         getEnvInt("APP_DATABASE_MIN_CONNS", 2),
		DatabaseMaxConnLifetime:  getEnvDuration("APP_DATABASE_MAX_CONN_LIFETIME", "1h"),
		DatabaseMaxConnIdleTime:  getEnvDuration("APP_DATABASE_MAX_CONN_IDLE_TIME", "30m"),
		DatabaseConnectTimeout:   getEnvDuration("APP_DATABASE_CONNECT_TIMEOUT", "5s"),
		DatabaseStatementTimeout: getEnvDuration("APP_DATABASE_STATEMENT_TIMEOUT", "10s"),
	}

	if cfg.Environment == "" {
		return Config{}, fmt.Errorf("config: APP_ENVIRONMENT no puede quedar vacío")
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("config: APP_DATABASE_URL no puede quedar vacío")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		var result int
		_, err := fmt.Sscanf(value, "%d", &result)
		if err == nil {
			return result
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback string) time.Duration {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		d, err := time.ParseDuration(value)
		if err == nil {
			return d
		}
	}
	d, _ := time.ParseDuration(fallback)
	return d
}
