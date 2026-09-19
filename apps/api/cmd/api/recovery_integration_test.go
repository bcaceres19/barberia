// Pruebas de integración/sistema de HU-008 contra el router REAL de
// producción (buildRouter) y PostgreSQL real. Cubren el recorrido completo
// de recuperación de acceso (solicitar → verificar → establecer
// contraseña) con el proveedor de entrega interceptado vía
// APP_RECOVERY_CAPTURE_FILE (mismo mecanismo que HU-007 usa para el reto
// telefónico): nunca se llama a Meta/Resend reales. El E2E visual de tres
// pasos pertenece a HU-011; esta prueba cubre el backend de punta a punta.
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./cmd/api/...
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"system-barbershop/internal/platform/database"
)

// recoveryEmailA/B son los correos de duena.a/dueno.b (mismas cuentas que
// staffUserActiveA/B, shopA/shopB), verificados por
// testdata/hu007_reto_telefonico.sql.
const (
	recoveryEmailA = "duena.a@ejemplo.test"
	recoveryEmailB = "dueno.b@ejemplo.test"

	// recoveryEmailO/recoveryPhoneO: cuenta con teléfono único verificado por
	// testdata/hu008_recuperacion_canal.sql, reservada para el recorrido por
	// WhatsApp (DEC-092). recoveryPhoneShared lo comparten dueno.p y dueno.q
	// en barberías distintas: ambiguo, no identifica ninguna cuenta
	// (DEC-093, DP-SEG-14).
	recoveryEmailO      = "dueno.o@ejemplo.test"
	recoveryPhoneO      = "+573000000011"
	recoveryPhoneShared = "+573000000012"
)

// recoveryTarget es el cuerpo común de los tres pasos: canal + valor.
type recoveryTarget map[string]string

func emailRecoveryTarget(email string) recoveryTarget {
	return recoveryTarget{"channel": "email", "email": email}
}

func whatsAppRecoveryTarget(phone string) recoveryTarget {
	return recoveryTarget{"channel": "whatsapp", "phone": phone}
}

// with devuelve una copia del destino con campos extra (code, resetToken…).
func (t recoveryTarget) with(extra map[string]string) map[string]string {
	body := make(map[string]string, len(t)+len(extra))
	for k, v := range t {
		body[k] = v
	}
	for k, v := range extra {
		body[k] = v
	}
	return body
}

// recoveryCapture es la forma que escribe auth.CapturingRecoveryCodeSender.
type recoveryCapture struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
	Code  string `json:"code"`
}

// buildRouterWithRecoveryCapture construye el router real con
// APP_RECOVERY_CAPTURE_FILE apuntando a un archivo temporal del propio
// test, para poder leer el código generado del lado del servidor sin
// depender de un proveedor real.
func buildRouterWithRecoveryCapture(t *testing.T, db *database.DB) (*chi.Mux, string) {
	t.Helper()
	capturePath := filepath.Join(t.TempDir(), "recovery-capture.json")
	t.Setenv("APP_RECOVERY_CAPTURE_FILE", capturePath)

	router, err := buildRouter(db, discardLogger(), testRouterConfig())
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	return router, capturePath
}

func doRecoveryRequest(router *chi.Mux, target recoveryTarget) *httptest.ResponseRecorder {
	body, _ := json.Marshal(target)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/auth/recovery/request", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func doRecoveryVerify(router *chi.Mux, target recoveryTarget, code string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(target.with(map[string]string{"code": code}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/auth/recovery/verify", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// doRecoveryRequestEventually reintenta la solicitud hasta que el
// proveedor interceptado realmente capture un código, con un plazo
// generoso. recoveryEmailA/B son las ÚNICAS dos cuentas con teléfono
// verificado del fixture compartido, y
// internal/modules/auth/postgres/recovery_repository_test.go ejercita el
// mismo recorrido de bajo nivel sobre las MISMAS cuentas; `go test ./...`
// corre paquetes en paralelo contra el MISMO PostgreSQL, así que el
// cooldown real de 60 s (DEC-064) puede rechazar momentáneamente la primera
// solicitud si ese otro paquete acaba de crear un código para la misma
// cuenta. La respuesta HTTP siempre es 202 sin importar el resultado
// interno (DEC-065), así que la única señal observable de una solicitud
// realmente aceptada es la captura del proveedor interceptado.
func doRecoveryRequestEventually(t *testing.T, router *chi.Mux, target recoveryTarget, capturePath string) recoveryCapture {
	t.Helper()
	deadline := time.Now().Add(65 * time.Second)
	for {
		rec := doRecoveryRequest(router, target)
		if rec.Code != http.StatusAccepted {
			t.Fatalf("expected 202 on request, got %d: %s", rec.Code, rec.Body.String())
		}
		if data, err := os.ReadFile(capturePath); err == nil {
			var c recoveryCapture
			if err := json.Unmarshal(data, &c); err != nil {
				t.Fatalf("unmarshal capture: %v", err)
			}
			return c
		}
		if time.Now().After(deadline) {
			t.Fatalf("recovery request: no se capturó ningún envío dentro del plazo de espera (contención de fixture entre paquetes de prueba)")
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func doRecoveryResetPassword(router *chi.Mux, target recoveryTarget, resetToken, newPassword string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(target.with(map[string]string{"resetToken": resetToken, "newPassword": newPassword}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/auth/recovery/reset-password", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// TestRecovery_System_FullJourney_RequestVerifyResetLogsInWithNewPassword
// cubre el recorrido completo de HU-008 contra el router y PostgreSQL
// reales: solicitar (202 genérico), verificar (200 con token de reinicio y
// destino enmascarado), establecer la contraseña nueva (204, revoca la
// sesión existente), y confirma que la contraseña ANTERIOR deja de
// autenticar mientras la NUEVA sí lo hace, vía POST /auth/login real
// (CA-008-05).
func TestRecovery_System_FullJourney_RequestVerifyResetLogsInWithNewPassword(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	router, capturePath := buildRouterWithRecoveryCapture(t, db)

	capture := doRecoveryRequestEventually(t, router, emailRecoveryTarget(recoveryEmailA), capturePath)
	// CA-008-09: elegir correo entrega solo por correo, nunca por WhatsApp.
	if capture.Code == "" || capture.Phone != "" || capture.Email != recoveryEmailA {
		t.Fatalf("unexpected capture: %+v", capture)
	}

	verifyRec := doRecoveryVerify(router, emailRecoveryTarget(recoveryEmailA), capture.Code)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on verify, got %d: %s", verifyRec.Code, verifyRec.Body.String())
	}
	var verifyResp struct {
		ResetToken  string `json:"resetToken"`
		MaskedPhone string `json:"maskedPhone"`
		MaskedEmail string `json:"maskedEmail"`
	}
	if err := json.Unmarshal(verifyRec.Body.Bytes(), &verifyResp); err != nil {
		t.Fatalf("invalid verify JSON: %v", err)
	}
	if verifyResp.ResetToken == "" {
		t.Fatal("expected a non-empty reset token")
	}
	if verifyResp.MaskedPhone == "" || strings.Contains(verifyResp.MaskedPhone, "3000000001") {
		// El destino enmascarado nunca debe ser igual al valor completo.
		t.Fatalf("expected a masked phone, got %q", verifyResp.MaskedPhone)
	}

	const newPassword = "contrasena-nueva-de-prueba-suficientemente-larga"
	resetRec := doRecoveryResetPassword(router, emailRecoveryTarget(recoveryEmailA), verifyResp.ResetToken, newPassword)
	if resetRec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 on reset-password, got %d: %s", resetRec.Code, resetRec.Body.String())
	}

	// CA-008-05 de punta a punta: la contraseña anterior (fixture) deja de
	// autenticar; la nueva sí lo hace. Se prueba contra el login REAL.
	oldLoginBody, _ := json.Marshal(map[string]string{"email": recoveryEmailA, "password": "fixture-no-es-un-hash-real-de-prueba-duena-a-0001"})
	oldLoginReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/auth/login", bytes.NewReader(oldLoginBody))
	oldLoginRec := httptest.NewRecorder()
	router.ServeHTTP(oldLoginRec, oldLoginReq)
	if oldLoginRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected the OLD password to be rejected after reset, got %d", oldLoginRec.Code)
	}

	newLoginBody, _ := json.Marshal(map[string]string{"email": recoveryEmailA, "password": newPassword})
	newLoginReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/auth/login", bytes.NewReader(newLoginBody))
	newLoginRec := httptest.NewRecorder()
	router.ServeHTTP(newLoginRec, newLoginReq)
	if newLoginRec.Code != http.StatusOK {
		t.Fatalf("expected the NEW password to authenticate, got %d: %s", newLoginRec.Code, newLoginRec.Body.String())
	}
}

// TestRecovery_System_NonenumerationThenWrongCodeRejected combina dos
// escenarios sobre la MISMA cuenta (recoveryEmailB) para no competir por
// cooldown con el recorrido feliz de recoveryEmailA de la prueba anterior:
// primero confirma que solicitar para una cuenta real y una inexistente
// produce la misma respuesta (CA-008-01, no enumeración); luego, sobre esa
// misma solicitud real, confirma que un código incorrecto nunca avanza a
// un token de reinicio.
func TestRecovery_System_NonenumerationThenWrongCodeRejected(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	router, capturePath := buildRouterWithRecoveryCapture(t, db)

	// Compara no enumeración (CA-008-01) en cada intento, incluido uno que
	// pudiera rechazarse internamente por contención de cooldown con el
	// paquete internal/modules/auth/postgres (ver doRecoveryRequestEventually):
	// la respuesta observable es idéntica sin importar el resultado interno
	// (DEC-065), así que esta comparación es válida en cada vuelta del
	// reintento, no solo en la que finalmente logra una captura real.
	var capture recoveryCapture
	deadline := time.Now().Add(65 * time.Second)
	for {
		realRec := doRecoveryRequest(router, emailRecoveryTarget(recoveryEmailB))
		nonexistentRec := doRecoveryRequest(router, emailRecoveryTarget("no-existe-recovery@ejemplo.test"))

		if realRec.Code != nonexistentRec.Code {
			t.Fatalf("expected identical status, got %d vs %d", realRec.Code, nonexistentRec.Code)
		}
		if realRec.Body.String() != nonexistentRec.Body.String() {
			t.Fatalf("expected identical body, got %q vs %q", realRec.Body.String(), nonexistentRec.Body.String())
		}

		if data, err := os.ReadFile(capturePath); err == nil {
			if err := json.Unmarshal(data, &capture); err != nil {
				t.Fatalf("unmarshal capture: %v", err)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("recovery request: no se capturó ningún envío dentro del plazo de espera (contención de fixture entre paquetes de prueba)")
		}
		time.Sleep(500 * time.Millisecond)
	}
	if capture.Email != recoveryEmailB {
		t.Fatalf("expected the capture to correspond to the real account, got %+v", capture)
	}

	wrongCode := "000000"
	if wrongCode == capture.Code {
		wrongCode = "111111"
	}
	verifyRec := doRecoveryVerify(router, emailRecoveryTarget(recoveryEmailB), wrongCode)
	if verifyRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for a wrong code, got %d: %s", verifyRec.Code, verifyRec.Body.String())
	}
	if strings.Contains(verifyRec.Body.String(), "resetToken") {
		t.Fatal("expected no reset token to leak on a failed verification")
	}
}

// TestRecovery_System_WhatsAppJourney_IdentifiesByPhoneAndSendsOnlyToPhone
// cubre CA-008-01, CA-008-09 y CA-008-10 de punta a punta con el canal
// WhatsApp (DEC-092, DEC-093): el número identifica la cuenta única, el
// código sale solo hacia el teléfono almacenado y los pasos 2 y 3 reutilizan
// el mismo canal y valor.
func TestRecovery_System_WhatsAppJourney_IdentifiesByPhoneAndSendsOnlyToPhone(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	router, capturePath := buildRouterWithRecoveryCapture(t, db)

	// El valor se escribe con separadores de presentación; el servidor lo
	// normaliza a E.164 antes de identificar la cuenta.
	target := whatsAppRecoveryTarget("+57 300 000 0011")
	capture := doRecoveryRequestEventually(t, router, target, capturePath)
	if capture.Code == "" || capture.Phone != recoveryPhoneO || capture.Email != "" {
		t.Fatalf("expected a send to the stored phone only, got %+v", capture)
	}

	verifyRec := doRecoveryVerify(router, target, capture.Code)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on verify by phone, got %d: %s", verifyRec.Code, verifyRec.Body.String())
	}
	var verifyResp struct {
		ResetToken  string `json:"resetToken"`
		MaskedPhone string `json:"maskedPhone"`
		MaskedEmail string `json:"maskedEmail"`
	}
	if err := json.Unmarshal(verifyRec.Body.Bytes(), &verifyResp); err != nil {
		t.Fatalf("invalid verify JSON: %v", err)
	}
	if verifyResp.ResetToken == "" {
		t.Fatal("expected a non-empty reset token")
	}
	if strings.Contains(verifyRec.Body.String(), "3000000011") || strings.Contains(verifyRec.Body.String(), recoveryEmailO) {
		t.Fatalf("the response must only carry masked destinations, got %s", verifyRec.Body.String())
	}

	resetRec := doRecoveryResetPassword(router, target, verifyResp.ResetToken, "contrasena-nueva-whatsapp-suficientemente-larga")
	if resetRec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 on reset-password by phone, got %d: %s", resetRec.Code, resetRec.Body.String())
	}
}

// TestRecovery_System_WhatsApp_AmbiguousOrUnknownPhone_SameResponseNoSend
// cubre CA-008-10, DEC-065 y DP-SEG-14: un número compartido por dos
// barberías, un número sin cuenta y un número con cuenta responden
// exactamente igual, y los dos primeros no envían nada.
func TestRecovery_System_WhatsApp_AmbiguousOrUnknownPhone_SameResponseNoSend(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	router, capturePath := buildRouterWithRecoveryCapture(t, db)

	ambiguous := doRecoveryRequest(router, whatsAppRecoveryTarget(recoveryPhoneShared))
	unknown := doRecoveryRequest(router, whatsAppRecoveryTarget("+573009999999"))

	if ambiguous.Code != http.StatusAccepted || unknown.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for both, got %d and %d", ambiguous.Code, unknown.Code)
	}
	if ambiguous.Body.String() != unknown.Body.String() {
		t.Fatalf("expected identical bodies, got %q vs %q", ambiguous.Body.String(), unknown.Body.String())
	}
	if _, err := os.ReadFile(capturePath); err == nil {
		t.Fatal("an ambiguous or unknown phone must not send any code")
	}

	// Pasos 2 y 3 con un número ambiguo: mismo 401 uniforme.
	verifyRec := doRecoveryVerify(router, whatsAppRecoveryTarget(recoveryPhoneShared), "123456")
	if verifyRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on verify with an ambiguous phone, got %d: %s", verifyRec.Code, verifyRec.Body.String())
	}
	resetRec := doRecoveryResetPassword(router, whatsAppRecoveryTarget(recoveryPhoneShared), "token-que-no-existe-0123456789", "contrasena-nueva-valida-larga")
	if resetRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on reset with an ambiguous phone, got %d: %s", resetRec.Code, resetRec.Body.String())
	}
}
