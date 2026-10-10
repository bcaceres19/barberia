package postgres_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	notificationpostgres "system-barbershop/internal/modules/notification/postgres"
	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
)

// Pruebas de integración del webhook de WhatsApp (issue #355, DEC-126) contra
// PostgreSQL real. Las tablas no tienen barbershop_id: el aislamiento que se
// comprueba es que ni la API ni el worker las tocan directamente.

func openDB(t *testing.T, envVar, fallback string) *database.DB {
	t.Helper()
	url := os.Getenv(envVar)
	if url == "" {
		url = fallback
	}
	cfg := config.Config{
		Environment: "test", DatabaseMaxConns: 8, DatabaseMinConns: 1,
		DatabaseMaxConnLifetime: time.Hour, DatabaseMaxConnIdleTime: 30 * time.Minute,
		DatabaseConnectTimeout: 5 * time.Second, DatabaseStatementTimeout: 10 * time.Second,
	}
	db, err := database.NewDB(config.DatabaseDSN(url), cfg)
	if err != nil {
		t.Fatalf("database.NewDB: %v", err)
	}
	t.Cleanup(db.Close)
	return db
}

func appDB(t *testing.T) *database.DB {
	return openDB(t, "TEST_DATABASE_URL", "postgres://barberia_app@localhost:5432/barberia_test?sslmode=disable")
}

func workerDB(t *testing.T) *database.DB {
	return openDB(t, "TEST_WORKER_DATABASE_URL", "postgres://barberia_worker@localhost:5432/barberia_test?sslmode=disable")
}

func hashOf(seed string) string {
	sum := sha256.Sum256([]byte(seed + time.Now().String()))
	return hex.EncodeToString(sum[:])
}

func TestWhatsAppEventRepository_InboundOpensTheWindowPerPhone(t *testing.T) {
	repo := notificationpostgres.NewWhatsAppEventRepository(appDB(t))
	ctx := context.Background()
	phoneA, phoneB := hashOf("telefono-a"), hashOf("telefono-b")

	if open, err := repo.ConversationOpen(ctx, phoneA); err != nil || open {
		t.Fatalf("without an inbound message the window must be closed: open=%v err=%v", open, err)
	}
	if err := repo.RecordInbound(ctx, phoneA, time.Now().Add(-time.Hour)); err != nil {
		t.Fatalf("RecordInbound: %v", err)
	}
	if open, err := repo.ConversationOpen(ctx, phoneA); err != nil || !open {
		t.Fatalf("a message from an hour ago must open the window: open=%v err=%v", open, err)
	}
	if open, err := repo.ConversationOpen(ctx, phoneB); err != nil || open {
		t.Fatalf("another phone must not see an open window: open=%v err=%v", open, err)
	}
}

func TestWhatsAppEventRepository_OldInboundDoesNotOpenAndRepeatsAreIdempotent(t *testing.T) {
	repo := notificationpostgres.NewWhatsAppEventRepository(appDB(t))
	ctx := context.Background()
	phone := hashOf("telefono-antiguo")

	if err := repo.RecordInbound(ctx, phone, time.Now().Add(-26*time.Hour)); err != nil {
		t.Fatalf("RecordInbound: %v", err)
	}
	if open, err := repo.ConversationOpen(ctx, phone); err != nil || open {
		t.Fatalf("a message older than 24h must not open the window: open=%v err=%v", open, err)
	}

	recent := time.Now().Add(-time.Minute)
	for i := 0; i < 2; i++ {
		if err := repo.RecordInbound(ctx, phone, recent); err != nil {
			t.Fatalf("RecordInbound repeat %d: %v", i, err)
		}
	}
	// Una notificación tardía con una fecha anterior no cierra la ventana.
	if err := repo.RecordInbound(ctx, phone, time.Now().Add(-25*time.Hour)); err != nil {
		t.Fatalf("RecordInbound late: %v", err)
	}
	if open, err := repo.ConversationOpen(ctx, phone); err != nil || !open {
		t.Fatalf("a late duplicate must not roll the window back: open=%v err=%v", open, err)
	}
}

func TestWhatsAppEventRepository_StatusIsRecordedOnce(t *testing.T) {
	repo := notificationpostgres.NewWhatsAppEventRepository(appDB(t))
	ctx := context.Background()
	wamid := "wamid." + hashOf("estado")[:16]
	code := 131047

	first, err := repo.RecordStatus(ctx, wamid, "failed", time.Now(), &code)
	if err != nil || !first {
		t.Fatalf("the first status must be new: inserted=%v err=%v", first, err)
	}
	again, err := repo.RecordStatus(ctx, wamid, "failed", time.Now(), &code)
	if err != nil || again {
		t.Fatalf("a repeated status must not be inserted again: inserted=%v err=%v", again, err)
	}
	other, err := repo.RecordStatus(ctx, wamid, "sent", time.Now(), nil)
	if err != nil || !other {
		t.Fatalf("another status of the same message must be new: inserted=%v err=%v", other, err)
	}
	if _, err := repo.RecordStatus(ctx, wamid, "inventado", time.Now(), nil); err == nil {
		t.Fatal("an unknown status must be rejected by the database")
	}
}

func TestWhatsAppTables_AreNotReadableByAppOrWorker(t *testing.T) {
	for name, db := range map[string]*database.DB{"barberia_app": appDB(t), "barberia_worker": workerDB(t)} {
		for _, table := range []string{"whatsapp_conversation_window", "whatsapp_message_status"} {
			err := db.CallSecurityDefinerRow(context.Background(), "SELECT count(*) FROM "+table, nil,
				func(row pgx.Row) error { var n int; return row.Scan(&n) })
			if err == nil {
				t.Fatalf("%s must not read %s directly (DDL-AUT-01)", name, table)
			}
		}
	}
}

func TestWhatsAppPurge_IsWorkerOnlyAndKeepsCurrentData(t *testing.T) {
	app := appDB(t)
	events := notificationpostgres.NewWhatsAppEventRepository(app)
	ctx := context.Background()
	phone := hashOf("telefono-vigente")
	if err := events.RecordInbound(ctx, phone, time.Now().Add(-time.Minute)); err != nil {
		t.Fatalf("RecordInbound: %v", err)
	}

	if _, err := notificationpostgres.NewWhatsAppPurgeRepository(app).PurgeExpired(ctx, 100, 2592000); err == nil {
		t.Fatal("barberia_app must not run the purge")
	}
	if _, err := notificationpostgres.NewWhatsAppPurgeRepository(workerDB(t)).PurgeExpired(ctx, 100, 2592000); err != nil {
		t.Fatalf("barberia_worker must run the purge: %v", err)
	}
	if open, err := events.ConversationOpen(ctx, phone); err != nil || !open {
		t.Fatalf("the purge must keep a current window: open=%v err=%v", open, err)
	}
	if _, err := notificationpostgres.NewWhatsAppPurgeRepository(workerDB(t)).PurgeExpired(ctx, 0, 2592000); err == nil {
		t.Fatal("an out-of-range limit must be rejected")
	}
}
