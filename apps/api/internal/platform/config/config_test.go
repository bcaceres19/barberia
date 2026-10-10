package config_test

import (
	"bytes"
	"encoding/base64"
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

func TestLoad_UnsupportedOTPProviderRejected(t *testing.T) {
	for _, provider := range []string{"sms", "otro"} {
		env := baseLocalEnv()
		env["OTP_PROVIDER"] = provider
		withEnv(t, env, func() {
			if _, err := config.Load(); err == nil {
				t.Fatalf("expected OTP_PROVIDER=%s to be rejected", provider)
			}
		})
	}
}

func TestLoad_MetaDefaultsToTemplateModeAndRecentGraphVersion(t *testing.T) {
	withEnv(t, baseLocalEnv(), func() {
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.MetaWhatsAppMode != "template" || cfg.MetaWhatsAppAPIVersion != "v24.0" {
			t.Fatalf("unexpected defaults: mode=%q version=%q", cfg.MetaWhatsAppMode, cfg.MetaWhatsAppAPIVersion)
		}
		if cfg.MetaWhatsAppConfigured() {
			t.Fatal("expected Meta to be unconfigured without credentials")
		}
	})
}

func productionTextEnv() map[string]string {
	return map[string]string{
		"APP_ENVIRONMENT":                   "production",
		"APP_DATABASE_URL":                  "postgres://barberia_app:secret@db:5432/barberia?sslmode=require",
		"APP_WORKER_DATABASE_URL":           "postgres://barberia_worker:secret@db:5432/barberia?sslmode=require",
		"APP_AUTH_HMAC_SECRET":              testHMACSecret,
		"APP_META_WHATSAPP_MODE":            "text",
		"APP_META_WHATSAPP_PHONE_NUMBER_ID": "1234567890",
		"APP_META_WHATSAPP_ACCESS_TOKEN":    "meta-access-token-de-prueba",
		"APP_RESEND_API_KEY":                "resend-api-key-de-prueba",
		"APP_RESEND_FROM_ADDRESS":           "no-responder@barberia.test",
	}
}

func TestLoad_MetaTextModeLoadsInLocalWithRecipients(t *testing.T) {
	env := baseLocalEnv()
	env["APP_META_WHATSAPP_MODE"] = "text"
	env["APP_META_WHATSAPP_PHONE_NUMBER_ID"] = "1234567890"
	env["APP_META_WHATSAPP_ACCESS_TOKEN"] = "meta-access-token-de-prueba"
	env["APP_META_WHATSAPP_TEST_RECIPIENTS"] = "+573001234567, +573009876543"
	withEnv(t, env, func() {
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !cfg.MetaWhatsAppConfigured() || len(cfg.MetaWhatsAppTestRecipients) != 2 {
			t.Fatalf("expected text mode to be configured with two recipients, cfg=%+v", cfg)
		}
	})
}

func TestLoad_MetaTextModeRequiresRecipientsInLocalAndTest(t *testing.T) {
	for _, environment := range []string{"local", "test"} {
		env := baseLocalEnv()
		env["APP_ENVIRONMENT"] = environment
		env["APP_META_WHATSAPP_MODE"] = "text"
		env["APP_META_WHATSAPP_PHONE_NUMBER_ID"] = "1234567890"
		env["APP_META_WHATSAPP_ACCESS_TOKEN"] = "meta-access-token-de-prueba"
		withEnv(t, env, func() {
			if _, err := config.Load(); err == nil {
				t.Fatalf("expected text mode without recipients to be rejected in %s", environment)
			}
		})
	}
}

func TestLoad_MetaTextModeAllowedInProductionWithoutRecipientsOrTemplate(t *testing.T) {
	withEnv(t, productionTextEnv(), func() {
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("text mode must load in production: %v", err)
		}
		if !cfg.MetaWhatsAppConfigured() {
			t.Fatal("expected text mode to count as configured in production without recipients")
		}
	})
}

func TestLoad_MetaTextModeInProductionStillRequiresCredentials(t *testing.T) {
	for _, missing := range []string{"APP_META_WHATSAPP_PHONE_NUMBER_ID", "APP_META_WHATSAPP_ACCESS_TOKEN"} {
		t.Run(missing, func(t *testing.T) {
			env := productionTextEnv()
			delete(env, missing)
			withEnv(t, env, func() {
				if _, err := config.Load(); err == nil {
					t.Fatalf("expected production without %s to be rejected", missing)
				}
			})
		})
	}
}

func TestLoad_MetaTextModeRecipientsMustBeE164(t *testing.T) {
	for _, recipients := range []string{"3001234567", "+57300abc"} {
		env := productionTextEnv()
		env["APP_META_WHATSAPP_TEST_RECIPIENTS"] = recipients
		withEnv(t, env, func() {
			if _, err := config.Load(); err == nil {
				t.Fatalf("expected recipients %q to be rejected", recipients)
			}
		})
	}
}

func TestLoad_MetaRejectsRetiredDevelopmentMode(t *testing.T) {
	env := baseLocalEnv()
	env["APP_META_WHATSAPP_MODE"] = "development"
	withEnv(t, env, func() {
		if _, err := config.Load(); err == nil {
			t.Fatal("expected the retired development mode to be rejected")
		}
	})
}

func TestLoad_MetaRejectsUnknownMode(t *testing.T) {
	env := baseLocalEnv()
	env["APP_META_WHATSAPP_MODE"] = "texto"
	withEnv(t, env, func() {
		if _, err := config.Load(); err == nil {
			t.Fatal("expected unknown mode to be rejected")
		}
	})
}

func TestLoad_MetaWhatsAppPartialConfigurationRejectedInEveryEnvironment(t *testing.T) {
	for _, environment := range []string{"local", "test", "pilot", "production"} {
		for _, partial := range []map[string]string{
			{"APP_META_WHATSAPP_PHONE_NUMBER_ID": "fake-phone-number-id"},
			{"APP_META_WHATSAPP_ACCESS_TOKEN": "fake-access-token"},
			{"APP_META_WHATSAPP_TEMPLATE_NAME": "fake-template"},
			{
				"APP_META_WHATSAPP_PHONE_NUMBER_ID": "fake-phone-number-id",
				"APP_META_WHATSAPP_ACCESS_TOKEN":    "fake-access-token",
			},
		} {
			t.Run(environment, func(t *testing.T) {
				env := baseLocalEnv()
				env["APP_ENVIRONMENT"] = environment
				env["APP_META_WHATSAPP_PHONE_NUMBER_ID"] = ""
				env["APP_META_WHATSAPP_ACCESS_TOKEN"] = ""
				env["APP_META_WHATSAPP_TEMPLATE_NAME"] = ""
				if environment == "pilot" || environment == "production" {
					env["APP_DATABASE_URL"] = "postgres://barberia_app:secret@db:5432/barberia?sslmode=require"
					env["APP_WORKER_DATABASE_URL"] = "postgres://barberia_worker:secret@db:5432/barberia?sslmode=require"
				}
				for key, value := range partial {
					env[key] = value
				}
				withEnv(t, env, func() {
					if _, err := config.Load(); err == nil {
						t.Fatal("expected partial Meta WhatsApp configuration to be rejected")
					}
				})
			})
		}
	}
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

// --- Google Calendar (DEC-099, DEC-102) -------------------------------------

func googleKey(fill byte) string {
	return base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{fill}, 32))
}

func googleEnv() map[string]string {
	env := baseLocalEnv()
	env["GOOGLE_CALENDAR_CLIENT_ID"] = "cliente.apps.googleusercontent.com"
	env["GOOGLE_CALENDAR_CLIENT_SECRET"] = "secreto-de-prueba"
	env["GOOGLE_CALENDAR_REDIRECT_URI"] = "http://localhost:5173/panel/barberia/google-calendar/callback"
	env["GOOGLE_CALENDAR_TOKEN_ENCRYPTION_KEY"] = googleKey(1)
	return env
}

func productionGoogleEnv() map[string]string {
	env := googleEnv()
	env["APP_ENVIRONMENT"] = "production"
	env["APP_DATABASE_URL"] = "postgres://barberia_app:secret@db:5432/barberia?sslmode=require"
	env["APP_WORKER_DATABASE_URL"] = "postgres://barberia_worker:secret@db:5432/barberia?sslmode=require"
	env["APP_META_WHATSAPP_PHONE_NUMBER_ID"] = "1234567890"
	env["APP_META_WHATSAPP_ACCESS_TOKEN"] = "meta-access-token-de-prueba"
	env["APP_META_WHATSAPP_TEMPLATE_NAME"] = "recuperacion_acceso"
	env["APP_RESEND_API_KEY"] = "resend-api-key-de-prueba"
	env["APP_RESEND_FROM_ADDRESS"] = "no-responder@barberia.test"
	env["GOOGLE_CALENDAR_REDIRECT_URI"] = "https://app.ejemplo.test/panel/barberia/google-calendar/callback"
	return env
}

func TestLoad_GoogleCalendar_AbsentVariablesDisableTheIntegrationWithoutFailing(t *testing.T) {
	withEnv(t, baseLocalEnv(), func() {
		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("la ausencia de variables no debe fallar el arranque: %v", err)
		}
		if cfg.GoogleCalendarEnabled() || cfg.GoogleCalendarPartiallyConfigured() {
			t.Fatal("sin variables la integración está desactivada y no es una configuración parcial")
		}
	})
}

func TestLoad_GoogleCalendar_CompleteConfigurationEnablesIt(t *testing.T) {
	env := googleEnv()
	env["GOOGLE_CALENDAR_TOKEN_ENCRYPTION_KEY_PREVIOUS"] = "v0:" + googleKey(9)
	withEnv(t, env, func() {
		cfg, err := config.Load()
		if err != nil {
			t.Fatal(err)
		}
		if !cfg.GoogleCalendarEnabled() || cfg.GoogleCalendarTokenKeyID != "v1" || len(cfg.GoogleCalendarPreviousKeys) != 1 {
			t.Fatalf("configuración inesperada: %+v", cfg.GoogleCalendarPreviousKeys)
		}
	})
}

func TestLoad_GoogleCalendar_MissingAnyOfTheFourDisablesAndFlagsPartial(t *testing.T) {
	for _, missing := range []string{
		"GOOGLE_CALENDAR_CLIENT_ID", "GOOGLE_CALENDAR_CLIENT_SECRET",
		"GOOGLE_CALENDAR_REDIRECT_URI", "GOOGLE_CALENDAR_TOKEN_ENCRYPTION_KEY",
	} {
		t.Run(missing, func(t *testing.T) {
			env := googleEnv()
			env[missing] = "" // getEnv trata una variable vacía como ausente
			withEnv(t, env, func() {
				cfg, err := config.Load()
				if err != nil {
					t.Fatalf("faltar una variable no debe fallar el arranque: %v", err)
				}
				if cfg.GoogleCalendarEnabled() || !cfg.GoogleCalendarPartiallyConfigured() {
					t.Fatal("con una variable ausente la integración queda desactivada y se avisa como parcial")
				}
			})
		})
	}
}

func TestLoad_GoogleCalendar_MalformedValuesFailLoudly(t *testing.T) {
	cases := map[string]struct {
		base     func() map[string]string
		override map[string]string
		wantErr  string
	}{
		"clave de 16 bytes": {googleEnv, map[string]string{
			"GOOGLE_CALENDAR_TOKEN_ENCRYPTION_KEY": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 16))}, "32 bytes"},
		"clave no base64": {googleEnv, map[string]string{"GOOGLE_CALENDAR_TOKEN_ENCRYPTION_KEY": "###"}, "32 bytes"},
		"redirect http en producción": {productionGoogleEnv, map[string]string{
			"GOOGLE_CALENDAR_REDIRECT_URI": "http://app.ejemplo.test/callback"}, "HTTPS"},
		"redirect http a un host que no es localhost": {googleEnv, map[string]string{
			"GOOGLE_CALENDAR_REDIRECT_URI": "http://app.ejemplo.test/callback"}, "HTTPS"},
		"redirect sin host": {googleEnv, map[string]string{"GOOGLE_CALENDAR_REDIRECT_URI": "/panel/callback"}, "URL válida"},
		"anterior sin id":   {googleEnv, map[string]string{"GOOGLE_CALENDAR_TOKEN_ENCRYPTION_KEY_PREVIOUS": googleKey(2)}, "id:clave"},
		"anterior repite el id activo": {googleEnv, map[string]string{
			"GOOGLE_CALENDAR_TOKEN_ENCRYPTION_KEY_PREVIOUS": "v1:" + googleKey(2)}, "repite"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			env := tc.base()
			for k, v := range tc.override {
				env[k] = v
			}
			withEnv(t, env, func() {
				_, err := config.Load()
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("un valor presente pero mal formado debe impedir el arranque con %q, fue: %v", tc.wantErr, err)
				}
			})
		})
	}
}

func TestLoad_GoogleCalendar_HTTPSRedirectIsValidInProduction(t *testing.T) {
	withEnv(t, productionGoogleEnv(), func() {
		cfg, err := config.Load()
		if err != nil || !cfg.GoogleCalendarEnabled() {
			t.Fatalf("HTTPS en producción es válido: %v", err)
		}
	})
}
