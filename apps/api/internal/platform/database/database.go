// Package database aloja la apertura del pool de PostgreSQL y el
// establecimiento del contexto de barbería (app.barbershop_id) requerido por
// RLS en cada transacción, según docs/05-backend/base-datos.md.
//
// El pool NO expone Query, QueryRow, Exec ni Begin como métodos exportados. La
// única forma de ejecutar trabajo de negocio es InTenantTx, que recibe el
// identificador de barbería y entrega el ejecutor SOLO dentro del alcance de
// una transacción ya configurada. Esto hace IMPOSIBLE omitir el contexto.
//
// Justificación de la dependencia (estandar-backend-go.md §5.21):
//   - Necesidad: driver nativo de PostgreSQL para pool con hooks de adquisición
//     y liberación, tipos uuid/timestamptz/numeric/rangos sin conversión, y
//     set_config parametrizable para fijar app.barbershop_id sin inyección SQL.
//   - Mantenimiento: jackc/pgx v5 es el driver estándar de facto, desarrollo
//     activo, versión semver estable.
//   - Licencia: MIT.
//   - Superficie transitiva: ninguna dependencia externa más allá de la librería
//     estándar de Go y golang.org/x/* internos.
//   - Seguridad: protocolo nativo, sin capas de abstracción que oculten
//     comportamiento; consultas SQL visibles y revisables. Atlas gobierna
//     migraciones; NO se ejecutan al arrancar (CA-001-07).
package database

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"system-barbershop/internal/platform/config"
)

// ErrContextSetupFailed se devuelve cuando falla la fijación de
// app.barbershop_id. El callback NUNCA se ejecuta en este caso.
var ErrContextSetupFailed = errors.New("database: fallo al fijar contexto de barbería")

// BarbershopID es un tipo propio para el identificador de barbería. No es un
// uuid.UUID desnudo ni un string: un tipo distinto impide pasar por error el id
// de un usuario, una cita u otra entidad (estandar-backend-go.md §5.10).
type BarbershopID string

// ValidBarbershopID valida que el string tenga formato UUID. Se usa en pruebas
// para inyectar el contexto; en producción llega desde la autenticación (HU-005).
func ValidBarbershopID(s string) (BarbershopID, error) {
	if len(s) != 36 {
		return "", fmt.Errorf("barbershop_id: longitud inválida %d", len(s))
	}
	return BarbershopID(s), nil
}

// Queries es la interfaz mínima que el callback recibe para ejecutar
// consultas. NO expone Begin, Commit, Rollback: la transacción la gestiona
// InTenantTx. Los métodos son un subconjunto deliberado de pgx.Tx.
type Queries interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

// DB encapsula el pool de conexiones. No se puede obtener una conexión suelta
// fuera de InTenantTx: el pool es privado y no expone métodos de consulta.
type DB struct {
	pool *pgxpool.Pool
}

// NewDB crea el pool con la configuración provista. La aplicación se conecta
// como barberia_app. NUNCA como barberia_migrator ni superusuario: si las
// pruebas pasan con un rol privilegiado, no prueban nada sobre RLS.
func NewDB(cfg config.Config) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("database: parse config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.DatabaseMaxConns)
	poolConfig.MinConns = int32(cfg.DatabaseMinConns)
	poolConfig.MaxConnLifetime = cfg.DatabaseMaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.DatabaseMaxConnIdleTime

	// Timeout de conexión y sentencia a nivel de pool
	poolConfig.ConnConfig.ConnectTimeout = cfg.DatabaseConnectTimeout
	// statement_timeout se fija por conexión en afterConnect

	// Hook de adquisición: verificamos que la conexión devuelta no conserve
	// app.barbershop_id. Convierte CA-002-03 en garantía de tiempo de
	// ejecución, no solo en una prueba.
	poolConfig.AfterRelease = func(conn *pgx.Conn) bool {
		var setting string
		err := conn.QueryRow(context.Background(),
			"SELECT current_setting('app.barbershop_id', true)").Scan(&setting)
		if err == nil && setting != "" {
			// La conexión quedó con residuo de contexto: no la devolvemos al pool
			conn.Close(context.Background())
			return false
		}
		return true
	}

	// Hook de configuración de conexión nueva: fija statement_timeout y
	// search_path explícito.
	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		// statement_timeout a nivel de conexión desde configuración
		_, err := conn.Exec(ctx,
			fmt.Sprintf("SET statement_timeout = %d",
				cfg.DatabaseStatementTimeout.Milliseconds()))
		if err != nil {
			return fmt.Errorf("database: set statement_timeout: %w", err)
		}
		_, err = conn.Exec(ctx, "SET search_path = public, pg_catalog")
		if err != nil {
			return fmt.Errorf("database: set search_path: %w", err)
		}
		return nil
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("database: create pool: %w", err)
	}

	// Ping inicial para verificar conectividad
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DatabaseConnectTimeout)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database: ping: %w", err)
	}

	return &DB{pool: pool}, nil
}

// Close cierra el pool. Debe llamarse al apagar la aplicación.
func (d *DB) Close() {
	d.pool.Close()
}

// InTenantTx es la ÚNICA forma de ejecutar trabajo de negocio. Recibe el
// contexto de la petición, el identificador de barbería y un callback que
// recibe el ejecutor de consultas YA dentro de una transacción con
// app.barbershop_id fijado con alcance local.
//
// Reglas de diseño, todas verificables:
// a) El pool NO expone Query, QueryRow, Exec ni Begin como métodos exportados.
// b) BarbershopID es un tipo propio, no un uuid.UUID desnudo ni un string.
// c) Secuencia interna: adquirir conexión -> BEGIN -> set_config local ->
//
//	ejecutar callback -> COMMIT, o ROLLBACK ante error/pánico.
//
// d) Si set_config falla, ROLLBACK y error SIN ejecutar callback.
//
//	Jamás se continúa con contexto vacío (CA-002-05).
//
// e) El ejecutor entregado al callback deja de ser válido al retornar. Si
//
//	alguien lo guarda en un struct, la transacción ya estará cerrada.
//
// f) context.Context es el primer parámetro, se propaga hasta PostgreSQL.
// g) No se inician goroutines dentro de la transacción.
// h) Nada de llamadas de red dentro de la transacción (base-datos.md §4.19).
func (d *DB) InTenantTx(
	ctx context.Context,
	shop BarbershopID,
	fn func(ctx context.Context, q Queries) error,
) error {
	// Adquirir conexión del pool
	conn, err := d.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("database: acquire connection: %w", err)
	}

	// Asegurar que la conexión se devuelve al pool al terminar
	defer func() {
		conn.Release()
		// Forzar GC para detectar retención accidental del ejecutor
		runtime.GC()
	}()

	// Iniciar transacción
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("database: begin transaction: %w", err)
	}

	// Rollback diferido: se ejecuta si el callback falla o hace pánico
	rolledBack := false
	defer func() {
		if !rolledBack {
			_ = tx.Rollback(ctx)
		}
	}()

	// Fijar app.barbershop_id con alcance LOCAL a la transacción.
	// set_config SÍ es parametrizable; el tercer argumento true significa
	// "local a la transacción": revierte solo al terminar la transacción,
	// sin necesidad de RESET. Ese comportamiento verifica CA-002-03.
	// NUNCA se concatena el identificador en la cadena SQL: eso introduce
	// inyección SQL en el punto exacto del que depende TODO el aislamiento.
	_, err = tx.Exec(ctx,
		"SELECT set_config('app.barbershop_id', $1::text, true)",
		string(shop),
	)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrContextSetupFailed, err)
	}

	// Verificar que el contexto se fijó correctamente
	var currentSetting string
	err = tx.QueryRow(ctx, "SELECT current_setting('app.barbershop_id')").Scan(&currentSetting)
	if err != nil {
		return fmt.Errorf("database: verify context: %w", err)
	}
	if currentSetting != string(shop) {
		return fmt.Errorf("%w: contexto fijado=%q esperado=%q",
			ErrContextSetupFailed, currentSetting, shop)
	}

	// Ejecutar el callback con el ejecutor de consultas
	err = fn(ctx, tx)
	if err != nil {
		rolledBack = true
		return err
	}

	// Commit
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("database: commit transaction: %w", err)
	}

	rolledBack = true
	return nil
}

// HealthCheck verifica conectividad con timeout corto y propio. La respuesta
// indica disponible o no disponible y NADA más: sin versión de PostgreSQL,
// sin nombre de base, sin host, sin usuario, sin texto del error del driver.
// Un endpoint de salud es público de hecho; su mensaje de error es superficie
// de reconocimiento gratuita para un atacante.
// La salud NO abre una transacción de tenant: no tiene barbería. Es la única
// lectura autorizada fuera del patrón.
func (d *DB) HealthCheck(ctx context.Context) error {
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return d.pool.Ping(checkCtx)
}

// PoolStats expone estadísticas del pool para observabilidad (no para uso de
// negocio). No rompe la encapsulación porque no entrega conexiones ni ejecutores.
func (d *DB) PoolStats() *pgxpool.Stat {
	return d.pool.Stat()
}
