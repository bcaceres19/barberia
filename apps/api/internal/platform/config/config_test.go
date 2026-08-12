package config_test

import (
	"fmt"
	"strings"
	"testing"

	"system-barbershop/internal/platform/config"
)

// withEnv fija variables de entorno para la duración de la prueba y las
// limpia al terminar, evitando fugas entre subpruebas.
func withEnv(t *testing.T, kv map[string]string, fn func()) {
	t.Helper()
	for k, v := range kv {
		t.Setenv(k, v)
	}
	fn()
}

func baseLocalEnv() map[string]string {
	return map[string]string{
		"APP_ENVIRONMENT":  "local",
		"APP_DATABASE_URL": "postgres://barberia_app:secret@localhost:5432/barberia?sslmode=disable",
	}
}

func TestLoad_DefaultsLocal(t *testing.T) {
	withEnv(t, baseLocalEnv(), func() {
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.DatabaseMaxConns != 20 || cfg.DatabaseMinConns != 2 {
			t.Fatalf("unexpected pool defaults: max=%d min=%d", cfg.DatabaseMaxConns, cfg.DatabaseMinConns)
		}
	})
}

func TestLoad_InvalidIntRejected(t *testing.T) {
	env := baseLocalEnv()
	env["APP_DATABASE_MAX_CONNS"] = "abc"
	withEnv(t, env, func() {
		_, err := config.Load()
		if err == nil {
			t.Fatal("expected error for invalid APP_DATABASE_MAX_CONNS, got nil")
		}
	})
}

func TestLoad_InvalidDurationRejected(t *testing.T) {
	env := baseLocalEnv()
	env["APP_DATABASE_STATEMENT_TIMEOUT"] = "not-a-duration"
	withEnv(t, env, func() {
		_, err := config.Load()
		if err == nil {
			t.Fatal("expected error for invalid APP_DATABASE_STATEMENT_TIMEOUT, got nil")
		}
	})
}

func TestLoad_MinConnsGreaterThanMaxConnsRejected(t *testing.T) {
	env := baseLocalEnv()
	env["APP_DATABASE_MIN_CONNS"] = "50"
	env["APP_DATABASE_MAX_CONNS"] = "10"
	withEnv(t, env, func() {
		_, err := config.Load()
		if err == nil {
			t.Fatal("expected error when minConns > maxConns, got nil")
		}
	})
}

func TestLoad_TLSNotRequiredInLocalOrTest(t *testing.T) {
	for _, env := range []string{"local", "test"} {
		e := baseLocalEnv()
		e["APP_ENVIRONMENT"] = env
		withEnv(t, e, func() {
			if _, err := config.Load(); err != nil {
				t.Fatalf("unexpected error in %s: %v", env, err)
			}
		})
	}
}

func TestLoad_TLSRequiredOutsideLocalTest(t *testing.T) {
	for _, env := range []string{"pilot", "production"} {
		e := map[string]string{
			"APP_ENVIRONMENT":         env,
			"APP_DATABASE_URL":        "postgres://barberia_app:secret@db:5432/barberia?sslmode=disable",
			"APP_WORKER_DATABASE_URL": "postgres://barberia_worker:secret@db:5432/barberia?sslmode=require",
		}
		withEnv(t, e, func() {
			_, err := config.Load()
			if err == nil {
				t.Fatalf("expected TLS error in %s, got nil", env)
			}
		})
	}
}

func TestLoad_WorkerDSNRequiredOutsideLocalTest(t *testing.T) {
	env := map[string]string{
		"APP_ENVIRONMENT":  "production",
		"APP_DATABASE_URL": "postgres://barberia_app:secret@db:5432/barberia?sslmode=require",
	}
	withEnv(t, env, func() {
		_, err := config.Load()
		if err == nil {
			t.Fatal("expected error for missing APP_WORKER_DATABASE_URL, got nil")
		}
	})
}

func TestLoad_WorkerDSNMustDifferFromAPIOutsideLocalTest(t *testing.T) {
	same := "postgres://barberia_app:secret@db:5432/barberia?sslmode=require"
	env := map[string]string{
		"APP_ENVIRONMENT":         "production",
		"APP_DATABASE_URL":        same,
		"APP_WORKER_DATABASE_URL": same,
	}
	withEnv(t, env, func() {
		_, err := config.Load()
		if err == nil {
			t.Fatal("expected error when api and worker share the same DSN, got nil")
		}
	})
}

func TestLoad_ProductionValidConfig(t *testing.T) {
	env := map[string]string{
		"APP_ENVIRONMENT":         "production",
		"APP_DATABASE_URL":        "postgres://barberia_app:secret@db:5432/barberia?sslmode=require",
		"APP_WORKER_DATABASE_URL": "postgres://barberia_worker:secret@db:5432/barberia?sslmode=require",
	}
	withEnv(t, env, func() {
		if _, err := config.Load(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// TestDatabaseDSN_RedactsPassword verifica que ni %s/%v ni %+v sobre un
// struct que contenga un DatabaseDSN filtren la contraseña.
func TestDatabaseDSN_RedactsPassword(t *testing.T) {
	dsn := config.DatabaseDSN("postgres://barberia_app:supersecreto@db:5432/barberia?sslmode=require")

	rendered := fmt.Sprintf("%s", dsn)
	if strings.Contains(rendered, "supersecreto") {
		t.Fatalf("password leaked via %%s: %s", rendered)
	}

	type wrapper struct{ DatabaseURL config.DatabaseDSN }
	w := wrapper{DatabaseURL: dsn}
	renderedStruct := fmt.Sprintf("%+v", w)
	if strings.Contains(renderedStruct, "supersecreto") {
		t.Fatalf("password leaked via %%+v: %s", renderedStruct)
	}
}

func TestDatabaseDSN_RedactsKeywordFormat(t *testing.T) {
	dsn := config.DatabaseDSN("host=db user=barberia_app password=supersecreto dbname=barberia sslmode=require")
	rendered := fmt.Sprintf("%s", dsn)
	if strings.Contains(rendered, "supersecreto") {
		t.Fatalf("password leaked via keyword-format DSN: %s", rendered)
	}
}
