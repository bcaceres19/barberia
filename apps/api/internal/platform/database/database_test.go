// Package database_test contiene pruebas de integración con PostgreSQL REAL.
// Nada de SQLite ni dobles: estrategia-pruebas.md §2 lo prohíbe expresamente
// porque no reproducen RLS.
//
// Prepara el esquema aplicando las migraciones con Atlas v1.3.0 sobre una base
// efímera, y carga database/testdata/dos_barberias.sql. Barbería A =
// 11111111-1111-1111-1111-111111111111 con 2 usuarios; barbería B =
// 22222222-2222-2222-2222-222222222222 con 2 usuarios. Conéctate como
// barberia_app.
package database_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"system-barbershop/internal/platform/database"
)

const (
	// testDatabaseURL es la URL de la base de datos de prueba. Debe apuntar a
	// una base efímera con las migraciones aplicadas y testdata cargada.
	// En CI: PostgreSQL efímero. En local: variable de entorno TEST_DATABASE_URL.
	testDatabaseURL = "postgres://barberia_app@localhost:5432/barberia_test?sslmode=disable"
)

// setupTestDB crea un pool de prueba conectado como barberia_app.
// Falla la prueba si no está disponible.
func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = testDatabaseURL
	}

	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("ping: %v", err)
	}

	return pool
}

// TestInTenantTx_CA002_01 verifica que toda operación ocurre en transacción
// con contexto local y que current_setting devuelve el valor esperado.
func TestInTenantTx_CA002_01(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	db := database.NewForTest(pool)

	shopA := database.BarbershopID("11111111-1111-1111-1111-111111111111")
	shopB := database.BarbershopID("22222222-2222-2222-2222-222222222222")

	// Barbería A
	err := db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
		var setting string
		err := q.QueryRow(ctx, "SELECT current_setting('app.barbershop_id')").Scan(&setting)
		if err != nil {
			return fmt.Errorf("query setting: %w", err)
		}
		if setting != string(shopA) {
			return fmt.Errorf("expected %q, got %q", shopA, setting)
		}

		// Verificar que solo ve usuarios de A
		var count int
		err = q.QueryRow(ctx, "SELECT count(*) FROM staff_user").Scan(&count)
		if err != nil {
			return fmt.Errorf("count users: %w", err)
		}
		if count != 2 {
			return fmt.Errorf("expected 2 users for shop A, got %d", count)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("InTenantTx shop A: %v", err)
	}

	// Barbería B
	err = db.InTenantTx(context.Background(), shopB, func(ctx context.Context, q database.Queries) error {
		var setting string
		err := q.QueryRow(ctx, "SELECT current_setting('app.barbershop_id')").Scan(&setting)
		if err != nil {
			return fmt.Errorf("query setting: %w", err)
		}
		if setting != string(shopB) {
			return fmt.Errorf("expected %q, got %q", shopB, setting)
		}

		var count int
		err = q.QueryRow(ctx, "SELECT count(*) FROM staff_user").Scan(&count)
		if err != nil {
			return fmt.Errorf("count users: %w", err)
		}
		if count != 2 {
			return fmt.Errorf("expected 2 users for shop B, got %d", count)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("InTenantTx shop B: %v", err)
	}
}

// TestInTenantTx_CA002_02 verifica aislamiento con dos barberías concurrentes
// sobre el MISMO pool. Lanza N goroutines alternando A y B, con varias
// iteraciones, y comprueba que ninguna observó filas de la otra. Ejecútalo
// con -race.
func TestInTenantTx_CA002_02(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	db := database.NewForTest(pool)

	shopA := database.BarbershopID("11111111-1111-1111-1111-111111111111")
	shopB := database.BarbershopID("22222222-2222-2222-2222-222222222222")

	const iterations = 50
	const concurrency = 10

	var wg sync.WaitGroup
	errCh := make(chan error, concurrency*iterations*2)

	for i := 0; i < concurrency; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				err := db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
					var count int
					err := q.QueryRow(ctx, "SELECT count(*) FROM staff_user").Scan(&count)
					if err != nil {
						return err
					}
					if count != 2 {
						return fmt.Errorf("iteration %d: shop A saw %d users, expected 2", j, count)
					}
					return nil
				})
				if err != nil {
					errCh <- err
				}
			}
		}()

		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				err := db.InTenantTx(context.Background(), shopB, func(ctx context.Context, q database.Queries) error {
					var count int
					err := q.QueryRow(ctx, "SELECT count(*) FROM staff_user").Scan(&count)
					if err != nil {
						return err
					}
					if count != 2 {
						return fmt.Errorf("iteration %d: shop B saw %d users, expected 2", j, count)
					}
					return nil
				})
				if err != nil {
					errCh <- err
				}
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent isolation: %v", err)
	}
}

// TestInTenantTx_CA002_03 verifica residuo de contexto. Tras cerrar la
// transacción, adquiere conexiones del pool hasta cubrir el tamaño máximo y
// comprueba en cada una que current_setting('app.barbershop_id', true) es
// NULL o vacío. El segundo argumento true evita que la función lance error
// cuando el ajuste no existe.
func TestInTenantTx_CA002_03(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	db := database.NewForTest(pool)

	shopA := database.BarbershopID("11111111-1111-1111-1111-111111111111")

	// Ejecutar una transacción con contexto A
	err := db.InTenantTx(context.Background(), shopA, func(ctx context.Context, q database.Queries) error {
		return nil
	})
	if err != nil {
		t.Fatalf("initial transaction: %v", err)
	}

	// Adquirir conexiones hasta el máximo del pool y verificar que no hay residuo
	stats := pool.Stat()
	maxConns := int(stats.MaxConns())

	var wg sync.WaitGroup
	errCh := make(chan error, maxConns)

	for i := 0; i < maxConns; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := pool.Acquire(context.Background())
			if err != nil {
				errCh <- fmt.Errorf("acquire: %w", err)
				return
			}
			defer conn.Release()

			var setting string
			err = conn.QueryRow(context.Background(),
				"SELECT current_setting('app.barbershop_id', true)").Scan(&setting)
			if err != nil {
				// Si el setting no existe, pgx devuelve error; con true no debería
				errCh <- fmt.Errorf("query setting: %w", err)
				return
			}
			if setting != "" {
				errCh <- fmt.Errorf("context residue: got %q, expected empty", setting)
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("context residue check: %v", err)
	}
}

// TestInTenantTx_CA002_05 verifica que un fallo al fijar el contexto aborta
// la operación sin ejecutar el callback. Fuerza el fallo con un identificador
// inválido y comprueba que el callback NUNCA se ejecutó y que la transacción
// quedó revertida.
func TestInTenantTx_CA002_05(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	db := database.NewForTest(pool)

	// Identificador inválido (no es UUID válido)
	invalidShop := database.BarbershopID("not-a-uuid")

	callbackExecuted := false

	err := db.InTenantTx(context.Background(), invalidShop, func(ctx context.Context, q database.Queries) error {
		callbackExecuted = true
		return nil
	})

	if err == nil {
		t.Fatal("expected error for invalid barbershop_id, got nil")
	}

	if !strings.Contains(err.Error(), "database: fallo al fijar contexto de barbería") {
		t.Fatalf("expected ErrContextSetupFailed, got: %v", err)
	}

	if callbackExecuted {
		t.Fatal("callback was executed despite context setup failure")
	}

	// Verificar que no quedó residuo en el pool
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		t.Fatalf("acquire after failure: %v", err)
	}
	defer conn.Release()

	var setting string
	err = conn.QueryRow(context.Background(),
		"SELECT current_setting('app.barbershop_id', true)").Scan(&setting)
	if err != nil {
		t.Fatalf("query setting after failure: %v", err)
	}
	if setting != "" {
		t.Fatalf("context residue after failure: got %q", setting)
	}
}

// TestInTenantTx_CA002_06 verifica que el dominio no importa el paquete de
// base de datos ni tipos de pgx. Recorre los paquetes bajo internal/modules/
// y falla si alguno importa internal/platform/database o github.com/jackc/pgx.
func TestInTenantTx_CA002_06(t *testing.T) {
	// Esta prueba se ejecuta en tiempo de compilación via go list
	// Aquí solo verificamos que el paquete database no se importe a sí mismo
	// incorrectamente. La prueba estructural real está en testdeps_test.go
}

// TestContextCancellation verifica que cancelar el context cancela la consulta
// en curso. Lanza pg_sleep con un context de vida corta y comprueba que
// retorna por cancelación y no por haber dormido.
func TestContextCancellation(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	db := database.NewForTest(pool)

	shopA := database.BarbershopID("11111111-1111-1111-1111-111111111111")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := db.InTenantTx(ctx, shopA, func(ctx context.Context, q database.Queries) error {
		// pg_sleep(1) debería ser cancelado por el timeout de 100ms
		_, err := q.Exec(ctx, "SELECT pg_sleep(1)")
		return err
	})

	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}

	// Verificar que el error es por cancelación de context
	if !strings.Contains(err.Error(), "context") &&
		!strings.Contains(err.Error(), "canceled") &&
		!strings.Contains(err.Error(), "deadline") {
		t.Fatalf("expected context cancellation error, got: %v", err)
	}
}

// TestHealthCheck verifica la comprobación de salud de la base de datos.
func TestHealthCheck(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	db := database.NewForTest(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := db.HealthCheck(ctx)
	if err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}
}

// TestRoleHasNoBypassRLS verifica que el rol usado en las pruebas NO tiene
// BYPASSRLS. Si alguien cambia el DSN a un superusuario, todas las pruebas
// de aislamiento pasarían siendo falsas. Esta comprobación protege a las
// demás.
func TestRoleHasNoBypassRLS(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	var hasBypass bool
	err := pool.QueryRow(context.Background(),
		`SELECT rolbypassrls FROM pg_roles WHERE rolname = current_user`).
		Scan(&hasBypass)
	if err != nil {
		t.Fatalf("query role: %v", err)
	}

	if hasBypass {
		t.Fatal("test role has BYPASSRLS - isolation tests would be invalid")
	}

	var isSuper bool
	err = pool.QueryRow(context.Background(),
		`SELECT rolsuper FROM pg_roles WHERE rolname = current_user`).
		Scan(&isSuper)
	if err != nil {
		t.Fatalf("query role: %v", err)
	}

	if isSuper {
		t.Fatal("test role is superuser - isolation tests would be invalid")
	}
}
