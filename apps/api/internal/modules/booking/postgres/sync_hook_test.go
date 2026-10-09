package postgres_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"system-barbershop/internal/modules/booking"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

// Pruebas del puerto SyncHook (issue #324, DEC-102): booking avisa, dentro de la
// MISMA transacción, de toda escritura sobre `appointment`.

// TestSyncHook_EveryAppointmentWriteStatementIsFollowedByAHookCall es la guarda de
// código fuente: si alguien agrega una escritura nueva sobre `appointment` (por
// ejemplo la cancelación por el cliente, HU-099) sin llamar al gancho, la cita
// cambiaría sin publicarse en Google Calendar. Cuenta las sentencias de escritura
// del paquete y exige al menos una llamada al gancho por cada una.
func TestSyncHook_EveryAppointmentWriteStatementIsFollowedByAHookCall(t *testing.T) {
	writes := regexp.MustCompile(`(?i)(INSERT\s+INTO\s+appointment\s*\(|UPDATE\s+appointment\s*\n|UPDATE\s+appointment\s+SET)`)
	calls := regexp.MustCompile(`r\.appointmentChanged\(`)

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	totalWrites, totalCalls := 0, 0
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(filepath.Join(".", name))
		if err != nil {
			t.Fatal(err)
		}
		totalWrites += len(writes.FindAll(src, -1))
		totalCalls += len(calls.FindAll(src, -1))
	}
	if totalWrites == 0 {
		t.Fatal("la guarda no encontró ninguna escritura sobre appointment: el patrón dejó de ser válido")
	}
	if totalCalls < totalWrites {
		t.Fatalf("hay %d escrituras sobre `appointment` y solo %d llamadas a r.appointmentChanged: toda escritura debe avisar al gancho (DEC-102)",
			totalWrites, totalCalls)
	}
}

type recordingHook struct {
	mu    sync.Mutex
	calls []string
	err   error
}

func (h *recordingHook) AppointmentChanged(_ context.Context, q database.Queries, barbershopID, appointmentID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.calls = append(h.calls, barbershopID+"|"+appointmentID)
	return h.err
}

func TestSyncHook_CreateAndStatusChanges_NotifyTheHookInTheSameTransaction(t *testing.T) {
	db := setupTestDB(t)
	hook := &recordingHook{}
	repo := newRepository(db).WithSyncHook(hook)
	suffix := uniqueSuffix(t)

	input := manualInput(t, barberQ1, serviceQ, suffix, booking.CustomerInput{
		New: &booking.NewCustomerInput{FullName: "Cliente gancho " + suffix},
	})
	result, err := repo.CreateManual(context.Background(), string(shopQ), input,
		idempotency.Key("hook-"+suffix), idempotency.Fingerprint("2123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("CreateManual: %v", err)
	}
	want := string(shopQ) + "|" + result.Appointment.ID
	if len(hook.calls) != 1 || hook.calls[0] != want {
		t.Fatalf("crear una cita avisa al gancho una vez con su barbería e id: %v", hook.calls)
	}
}

func TestSyncHook_HookFailureRollsBackTheBusinessChange(t *testing.T) {
	db := setupTestDB(t)
	hook := &recordingHook{err: errors.New("no se pudo encolar")}
	repo := newRepository(db).WithSyncHook(hook)
	suffix := uniqueSuffix(t)

	input := manualInput(t, barberQ1, serviceQ, suffix, booking.CustomerInput{
		New: &booking.NewCustomerInput{FullName: "Cliente revertido " + suffix},
	})
	if _, err := repo.CreateManual(context.Background(), string(shopQ), input,
		idempotency.Key("hook-fail-"+suffix), idempotency.Fingerprint("3123456789abcdef0123456789abcdef")); err == nil {
		t.Fatal("si el gancho falla, la operación falla")
	}

	var count int
	if err := db.InTenantTx(context.Background(), shopQ, func(ctx context.Context, q database.Queries) error {
		return q.QueryRow(ctx, `SELECT count(*) FROM appointment WHERE barbershop_id = $1 AND attendee_name = $2`,
			string(shopQ), "Persona atendida "+suffix).Scan(&count)
	}); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("es preferible no confirmar la cita a confirmarla sin encolar su publicación: la transacción debe revertirse")
	}
}

func TestSyncHook_WithoutAHookTheRepositoryWorksAsBefore(t *testing.T) {
	db := setupTestDB(t)
	repo := newRepository(db)
	suffix := uniqueSuffix(t)
	input := manualInput(t, barberQ1, serviceQ, suffix, booking.CustomerInput{
		New: &booking.NewCustomerInput{FullName: "Cliente sin gancho " + suffix},
	})
	if _, err := repo.CreateManual(context.Background(), string(shopQ), input,
		idempotency.Key("nohook-"+suffix), idempotency.Fingerprint("4123456789abcdef0123456789abcdef")); err != nil {
		t.Fatalf("sin gancho registrado todo sigue funcionando: %v", err)
	}
}
