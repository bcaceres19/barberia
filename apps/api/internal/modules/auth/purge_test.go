package auth_test

import (
	"context"
	"errors"
	"testing"

	"system-barbershop/internal/modules/auth"
)

type purgeCall struct{ limit int }

type fakePurgeRepository struct {
	loginThrottleDeleted  int
	phoneChallengeDeleted int
	recoveryCodeDeleted   int
	loginThrottleErr      error
	phoneChallengeErr     error
	recoveryCodeErr       error
	loginThrottleCalls    []purgeCall
	phoneChallengeCalls   []purgeCall
	recoveryCodeCalls     []purgeCall
}

func (f *fakePurgeRepository) PurgeLoginThrottle(_ context.Context, limit int) (int, error) {
	f.loginThrottleCalls = append(f.loginThrottleCalls, purgeCall{limit})
	return f.loginThrottleDeleted, f.loginThrottleErr
}

func (f *fakePurgeRepository) PurgePhoneChallenges(_ context.Context, limit int) (int, error) {
	f.phoneChallengeCalls = append(f.phoneChallengeCalls, purgeCall{limit})
	return f.phoneChallengeDeleted, f.phoneChallengeErr
}

func (f *fakePurgeRepository) PurgeRecoveryCodes(_ context.Context, limit int) (int, error) {
	f.recoveryCodeCalls = append(f.recoveryCodeCalls, purgeCall{limit})
	return f.recoveryCodeDeleted, f.recoveryCodeErr
}

func TestPurgeService_PurgeOnce_ReturnsAllThreeCounts(t *testing.T) {
	repo := &fakePurgeRepository{loginThrottleDeleted: 3, phoneChallengeDeleted: 7, recoveryCodeDeleted: 2}
	svc := auth.NewPurgeService(repo, 500, 500, 500)

	loginDeleted, challengeDeleted, recoveryDeleted, err := svc.PurgeOnce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loginDeleted != 3 || challengeDeleted != 7 || recoveryDeleted != 2 {
		t.Fatalf("expected (3,7,2), got (%d,%d,%d)", loginDeleted, challengeDeleted, recoveryDeleted)
	}
	if repo.loginThrottleCalls[0].limit != 500 || repo.phoneChallengeCalls[0].limit != 500 || repo.recoveryCodeCalls[0].limit != 500 {
		t.Fatal("expected the configured limits to be passed through")
	}
}

func TestPurgeService_LoginThrottleError_SkipsRest(t *testing.T) {
	repo := &fakePurgeRepository{loginThrottleErr: errors.New("db unavailable")}
	svc := auth.NewPurgeService(repo, 500, 500, 500)

	_, _, _, err := svc.PurgeOnce(context.Background())
	if err == nil {
		t.Fatal("expected an error when PurgeLoginThrottle fails")
	}
	if len(repo.phoneChallengeCalls) != 0 || len(repo.recoveryCodeCalls) != 0 {
		t.Fatal("expected PurgePhoneChallenges/PurgeRecoveryCodes to be skipped after a login_throttle failure")
	}
}

func TestPurgeService_PhoneChallengeError_SkipsRecoveryButKeepsLoginThrottleCount(t *testing.T) {
	repo := &fakePurgeRepository{loginThrottleDeleted: 5, phoneChallengeErr: errors.New("db unavailable")}
	svc := auth.NewPurgeService(repo, 500, 500, 500)

	loginDeleted, _, _, err := svc.PurgeOnce(context.Background())
	if err == nil {
		t.Fatal("expected an error when PurgePhoneChallenges fails")
	}
	if loginDeleted != 5 {
		t.Fatalf("expected the login_throttle count to still be reported, got %d", loginDeleted)
	}
	if len(repo.recoveryCodeCalls) != 0 {
		t.Fatal("expected PurgeRecoveryCodes to be skipped after a phone_challenge failure")
	}
}

func TestPurgeService_RecoveryCodeError_StillReturnsEarlierCounts(t *testing.T) {
	repo := &fakePurgeRepository{loginThrottleDeleted: 5, phoneChallengeDeleted: 4, recoveryCodeErr: errors.New("db unavailable")}
	svc := auth.NewPurgeService(repo, 500, 500, 500)

	loginDeleted, challengeDeleted, _, err := svc.PurgeOnce(context.Background())
	if err == nil {
		t.Fatal("expected an error when PurgeRecoveryCodes fails")
	}
	if loginDeleted != 5 || challengeDeleted != 4 {
		t.Fatalf("expected the earlier counts to still be reported, got (%d,%d)", loginDeleted, challengeDeleted)
	}
}

func TestPurgeService_CancelledContext_NeverCallsRepository(t *testing.T) {
	repo := &fakePurgeRepository{}
	svc := auth.NewPurgeService(repo, 500, 500, 500)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, _, _ = svc.PurgeOnce(ctx)
	if len(repo.loginThrottleCalls) != 0 || len(repo.phoneChallengeCalls) != 0 || len(repo.recoveryCodeCalls) != 0 {
		t.Fatal("expected no repository call once the context was already cancelled")
	}
}
