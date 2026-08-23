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

// testHMACSecret cumple el largo mínimo exigido (32) sin ser un secreto
// real: solo se usa dentro de pruebas con t.Setenv, nunca persiste.
const testHMACSecret = "prueba-no-es-un-secreto-real-0123456789"

func baseLocalEnv() map[string]string {
	return map[string]string{
		"APP_ENVIRONMENT":      "local",
		"APP_DATABASE_URL":     "postgres://barberia_app:secret@localhost:5432/barberia?sslmode=disable",
		"APP_AUTH_HMAC_SECRET": testHMACSecret,
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
			"APP_AUTH_HMAC_SECRET":    testHMACSecret,
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
		"APP_ENVIRONMENT":      "production",
		"APP_DATABASE_URL":     "postgres://barberia_app:secret@db:5432/barberia?sslmode=require",
		"APP_AUTH_HMAC_SECRET": testHMACSecret,
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
		"APP_AUTH_HMAC_SECRET":    testHMACSecret,
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
		"APP_ENVIRONMENT":                   "production",
		"APP_DATABASE_URL":                  "postgres://barberia_app:secret@db:5432/barberia?sslmode=require",
		"APP_WORKER_DATABASE_URL":           "postgres://barberia_worker:secret@db:5432/barberia?sslmode=require",
		"APP_AUTH_HMAC_SECRET":              testHMACSecret,
		"APP_META_WHATSAPP_PHONE_NUMBER_ID": "1234567890",
		"APP_META_WHATSAPP_ACCESS_TOKEN":    "meta-access-token-de-prueba",
		"APP_META_WHATSAPP_TEMPLATE_NAME":   "recuperacion_acceso",
		"APP_RESEND_API_KEY":                "resend-api-key-de-prueba",
		"APP_RESEND_FROM_ADDRESS":           "no-responder@barberia.test",
	}
	withEnv(t, env, func() {
		if _, err := config.Load(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// TestLoad_ProductionMissingMetaWhatsAppCredentialsRejected cubre DEC-066:
// fuera de local/test, HU-008 exige el adaptador real de WhatsApp, no el
// marcador de posición.
func TestLoad_ProductionMissingMetaWhatsAppCredentialsRejected(t *testing.T) {
	env := map[string]string{
		"APP_ENVIRONMENT":         "production",
		"APP_DATABASE_URL":        "postgres://barberia_app:secret@db:5432/barberia?sslmode=require",
		"APP_WORKER_DATABASE_URL": "postgres://barberia_worker:secret@db:5432/barberia?sslmode=require",
		"APP_AUTH_HMAC_SECRET":    testHMACSecret,
		"APP_RESEND_API_KEY":      "resend-api-key-de-prueba",
		"APP_RESEND_FROM_ADDRESS": "no-responder@barberia.test",
	}
	withEnv(t, env, func() {
		if _, err := config.Load(); err == nil {
			t.Fatal("expected error when Meta WhatsApp credentials are missing outside local/test")
		}
	})
}

// TestLoad_ProductionMissingResendCredentialsRejected cubre DEC-066 para el
// canal de correo.
func TestLoad_ProductionMissingResendCredentialsRejected(t *testing.T) {
	env := map[string]string{
		"APP_ENVIRONMENT":                   "production",
		"APP_DATABASE_URL":                  "postgres://barberia_app:secret@db:5432/barberia?sslmode=require",
		"APP_WORKER_DATABASE_URL":           "postgres://barberia_worker:secret@db:5432/barberia?sslmode=require",
		"APP_AUTH_HMAC_SECRET":              testHMACSecret,
		"APP_META_WHATSAPP_PHONE_NUMBER_ID": "1234567890",
		"APP_META_WHATSAPP_ACCESS_TOKEN":    "meta-access-token-de-prueba",
		"APP_META_WHATSAPP_TEMPLATE_NAME":   "recuperacion_acceso",
	}
	withEnv(t, env, func() {
		if _, err := config.Load(); err == nil {
			t.Fatal("expected error when Resend credentials are missing outside local/test")
		}
	})
}

// TestLoad_LocalEnvironment_MissingProviderCredentials_StillLoads cubre que
// local/test NO exige credenciales de proveedor (main.go usa el marcador
// de posición documentado en ese caso).
func TestLoad_LocalEnvironment_MissingProviderCredentials_StillLoads(t *testing.T) {
	env := baseLocalEnv()
	withEnv(t, env, func() {
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.MetaWhatsAppPhoneNumberID != "" || cfg.ResendAPIKey != "" {
			t.Fatal("expected empty provider credentials when none are set")
		}
	})
}

func TestLoad_RecoveryParamsOutOfRangeRejected(t *testing.T) {
	env := baseLocalEnv()
	env["APP_RECOVERY_CODE_MAX_ATTEMPTS"] = "0"
	withEnv(t, env, func() {
		if _, err := config.Load(); err == nil {
			t.Fatal("expected error for APP_RECOVERY_CODE_MAX_ATTEMPTS=0")
		}
	})
}

func TestLoad_AuthHMACSecretTooShortRejected(t *testing.T) {
	env := baseLocalEnv()
	env["APP_AUTH_HMAC_SECRET"] = "corto"
	withEnv(t, env, func() {
		if _, err := config.Load(); err == nil {
			t.Fatal("expected error for short APP_AUTH_HMAC_SECRET, got nil")
		}
	})
}

func TestLoad_AuthHMACSecretMissingRejected(t *testing.T) {
	env := baseLocalEnv()
	delete(env, "APP_AUTH_HMAC_SECRET")
	withEnv(t, env, func() {
		if _, err := config.Load(); err == nil {
			t.Fatal("expected error for missing APP_AUTH_HMAC_SECRET, got nil")
		}
	})
}

func TestLoad_LoginThrottleRetentionBelowEscalationRejected(t *testing.T) {
	env := baseLocalEnv()
	env["APP_LOGIN_THROTTLE_ESCALATION_SECONDS"] = "86400"
	env["APP_LOGIN_THROTTLE_RETENTION_SECONDS"] = "3600"
	withEnv(t, env, func() {
		if _, err := config.Load(); err == nil {
			t.Fatal("expected error when retention < escalation, got nil")
		}
	})
}

func TestLoad_LoginThrottleDefaults(t *testing.T) {
	withEnv(t, baseLocalEnv(), func() {
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.LoginThrottleWindowSeconds != 900 || cfg.LoginThrottleEscalationSeconds != 86400 ||
			cfg.LoginThrottleThreshold != 5 || cfg.LoginThrottleRetentionSeconds != 172800 {
			t.Fatalf("unexpected throttle defaults: %+v", cfg)
		}
		if cfg.PhoneChallengeExpiresSeconds != 300 || cfg.PhoneChallengeMaxAttempts != 5 ||
			cfg.PhoneChallengeRateWindowSeconds != 900 || cfg.PhoneChallengeRateMaxActive != 3 ||
			cfg.PhoneChallengeResendCooldownSeconds != 60 {
			t.Fatalf("unexpected phone challenge defaults: %+v", cfg)
		}
		if len(cfg.TrustedProxies) != 0 {
			t.Fatalf("expected no trusted proxies by default, got %v", cfg.TrustedProxies)
		}
	})
}

func TestLoad_TrustedProxiesParsed(t *testing.T) {
	env := baseLocalEnv()
	env["APP_TRUSTED_PROXIES"] = "10.0.0.0/8, 172.16.0.0/12"
	withEnv(t, env, func() {
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cfg.TrustedProxies) != 2 || cfg.TrustedProxies[0] != "10.0.0.0/8" || cfg.TrustedProxies[1] != "172.16.0.0/12" {
			t.Fatalf("unexpected trusted proxies: %v", cfg.TrustedProxies)
		}
	})
}

func TestLoad_TrustedProxiesInvalidCIDRRejected(t *testing.T) {
	env := baseLocalEnv()
	env["APP_TRUSTED_PROXIES"] = "not-a-cidr"
	withEnv(t, env, func() {
		if _, err := config.Load(); err == nil {
			t.Fatal("expected error for invalid CIDR in APP_TRUSTED_PROXIES, got nil")
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
