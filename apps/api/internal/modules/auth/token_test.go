package auth_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"system-barbershop/internal/modules/auth"
)

func TestCryptoTokenGenerator_New_ProducesDistinctTokens(t *testing.T) {
	g := auth.NewCryptoTokenGenerator()

	a, err := g.New()
	if err != nil {
		t.Fatalf("New (a): %v", err)
	}
	b, err := g.New()
	if err != nil {
		t.Fatalf("New (b): %v", err)
	}

	if a == b {
		t.Fatal("expected two independent tokens to differ")
	}
	if len(a) < 32 {
		t.Fatalf("expected a token with substantial length, got %d chars", len(a))
	}
}

func TestHashToken_MatchesSHA256HexOf64Chars(t *testing.T) {
	got := auth.HashToken("token-de-prueba")

	sum := sha256.Sum256([]byte("token-de-prueba"))
	want := hex.EncodeToString(sum[:])

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
	if len(got) != 64 {
		t.Fatalf("expected 64 hex chars (staff_session_token_hash_ck), got %d", len(got))
	}
}

func TestHashToken_IsDeterministic(t *testing.T) {
	if auth.HashToken("mismo-valor") != auth.HashToken("mismo-valor") {
		t.Fatal("expected HashToken to be deterministic for the same input")
	}
}
