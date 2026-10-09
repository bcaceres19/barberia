package postgres_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestSyncHook_EveryTimeBlockWriteStatementIsFollowedByAHookCall es la guarda de
// código fuente del puerto SyncHook de schedule (issue #324, DEC-102): toda
// escritura sobre `time_block` (alta y retiro lógico) debe avisar al gancho, o el
// bloqueo cambiaría sin publicarse en Google Calendar.
func TestSyncHook_EveryTimeBlockWriteStatementIsFollowedByAHookCall(t *testing.T) {
	writes := regexp.MustCompile(`(?i)(INSERT\s+INTO\s+time_block\s*\(|UPDATE\s+time_block\s*\n|UPDATE\s+time_block\s+SET|DELETE\s+FROM\s+time_block\b)`)
	calls := regexp.MustCompile(`r\.timeBlockChanged\(`)

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
		t.Fatal("la guarda no encontró ninguna escritura sobre time_block: el patrón dejó de ser válido")
	}
	if totalCalls < totalWrites {
		t.Fatalf("hay %d escrituras sobre `time_block` y solo %d llamadas a r.timeBlockChanged: toda escritura debe avisar al gancho (DEC-102)",
			totalWrites, totalCalls)
	}
}
