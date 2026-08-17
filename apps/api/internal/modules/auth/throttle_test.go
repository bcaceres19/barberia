package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/platform/apperr"
)

type registerAttemptCall struct {
	ipHash string
	cfg    auth.ThrottleConfig
}

type fakeThrottleRepository struct {
	attemptCount int
	escalated    bool
	retryAfter   time.Time
	err          error
	calls        []registerAttemptCall
}

func (f *fakeThrottleRepository) RegisterAttempt(_ context.Context, ipHash string, cfg auth.ThrottleConfig) (int, bool, time.Time, error) {
	f.calls = append(f.calls, registerAttemptCall{ipHash, cfg})
	return f.attemptCount, f.escalated, f.retryAfter, f.err
}

func testThrottleCfg() auth.ThrottleConfig {
	return auth.ThrottleConfig{WindowSeconds: 900, EscalationSeconds: 86400, Threshold: 5, RetentionSeconds: 172800}
}

func TestThrottleService_BelowThreshold_ReturnsNil(t *testing.T) {
	repo := &fakeThrottleRepository{attemptCount: 3, escalated: false}
	svc := auth.NewThrottleService(repo, testThrottleCfg(), []byte("secreto-de-prueba-suficientemente-largo"), fixedClock{time.Now()})

	if err := svc.Check(context.Background(), "203.0.113.1"); err != nil {
		t.Fatalf("expected nil error below threshold, got %v", err)
	}
	if len(repo.calls) != 1 {
		t.Fatalf("expected exactly one RegisterAttempt call, got %d", len(repo.calls))
	}
}

func TestThrottleService_Escalated_ReturnsChallengeRequired(t *testing.T) {
	now := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
	repo := &fakeThrottleRepository{attemptCount: 6, escalated: true, retryAfter: now.Add(24 * time.Hour)}
	svc := auth.NewThrottleService(repo, testThrottleCfg(), []byte("secreto-de-prueba-suficientemente-largo"), fixedClock{now})

	err := svc.Check(context.Background(), "203.0.113.1")

	appErr, ok := apperr.As(err)
	if !ok {
		t.Fatalf("expected *apperr.Error, got %v (%T)", err, err)
	}
	if appErr.Kind != apperr.KindChallengeRequired {
		t.Fatalf("expected KindChallengeRequired, got %v", appErr.Kind)
	}
	wantRetry := int(24 * time.Hour / time.Second)
	if appErr.RetryAfterSeconds != wantRetry {
		t.Fatalf("expected RetryAfterSeconds=%d, got %d", wantRetry, appErr.RetryAfterSeconds)
	}
}

func TestThrottleService_Escalated_RetryAfterNeverBelowOne(t *testing.T) {
	now := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
	// retryAfter en el pasado (reloj del repositorio ligeramente distinto
	// del reloj del servicio): no debe producir un Retry-After negativo o
	// cero, que confundiría a cualquier cliente HTTP.
	repo := &fakeThrottleRepository{attemptCount: 6, escalated: true, retryAfter: now.Add(-1 * time.Second)}
	svc := auth.NewThrottleService(repo, testThrottleCfg(), []byte("secreto-de-prueba-suficientemente-largo"), fixedClock{now})

	err := svc.Check(context.Background(), "203.0.113.1")

	appErr, _ := apperr.As(err)
	if appErr.RetryAfterSeconds < 1 {
		t.Fatalf("expected RetryAfterSeconds >= 1, got %d", appErr.RetryAfterSeconds)
	}
}

func TestThrottleService_HashIP_StableAndKeyed(t *testing.T) {
	secretA := []byte("secreto-de-prueba-suficientemente-largo-a")
	secretB := []byte("secreto-de-prueba-suficientemente-largo-b")
	svcA1 := auth.NewThrottleService(&fakeThrottleRepository{}, testThrottleCfg(), secretA, fixedClock{time.Now()})
	svcA2 := auth.NewThrottleService(&fakeThrottleRepository{}, testThrottleCfg(), secretA, fixedClock{time.Now()})
	svcB := auth.NewThrottleService(&fakeThrottleRepository{}, testThrottleCfg(), secretB, fixedClock{time.Now()})

	if svcA1.HashIP("203.0.113.1") != svcA2.HashIP("203.0.113.1") {
		t.Fatal("expected the same IP+secret to hash identically across instances")
	}
	if svcA1.HashIP("203.0.113.1") == svcB.HashIP("203.0.113.1") {
		t.Fatal("expected different secrets to produce different hashes for the same IP")
	}
	if svcA1.HashIP("203.0.113.1") == svcA1.HashIP("203.0.113.2") {
		t.Fatal("expected different IPs to hash differently")
	}
	hash := svcA1.HashIP("203.0.113.1")
	if len(hash) != 64 {
		t.Fatalf("expected a 64-char hex HMAC-SHA256, got %d chars: %q", len(hash), hash)
	}
	if hash == "203.0.113.1" {
		t.Fatal("the raw IP must never appear as its own hash")
	}
}

func TestThrottleService_RepositoryError_ReturnsInternal(t *testing.T) {
	repo := &fakeThrottleRepository{err: errors.New("db unavailable")}
	svc := auth.NewThrottleService(repo, testThrottleCfg(), []byte("secreto-de-prueba-suficientemente-largo"), fixedClock{time.Now()})

	err := svc.Check(context.Background(), "203.0.113.1")
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected KindInternal, got %v (%T)", err, err)
	}
}

func TestThrottleService_CancelledContext_NeverCallsRepository(t *testing.T) {
	repo := &fakeThrottleRepository{}
	svc := auth.NewThrottleService(repo, testThrottleCfg(), []byte("secreto-de-prueba-suficientemente-largo"), fixedClock{time.Now()})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_ = svc.Check(ctx, "203.0.113.1")
	if len(repo.calls) != 0 {
		t.Fatal("expected no repository call once the context was already cancelled")
	}
}
