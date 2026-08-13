package auth_test

import (
	"strings"
	"testing"

	"system-barbershop/internal/modules/auth"
)

func TestArgon2Hasher_HashThenVerify_Succeeds(t *testing.T) {
	h := auth.NewArgon2Hasher()

	encoded, err := h.Hash("una-contraseña-de-prueba")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if !strings.HasPrefix(encoded, "$argon2id$") {
		t.Fatalf("expected a PHC-encoded argon2id hash, got %q", encoded)
	}
	if !h.Verify(encoded, "una-contraseña-de-prueba") {
		t.Fatal("expected the correct password to verify")
	}
}

func TestArgon2Hasher_Verify_WrongPasswordFails(t *testing.T) {
	h := auth.NewArgon2Hasher()

	encoded, err := h.Hash("contraseña-correcta")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if h.Verify(encoded, "contraseña-incorrecta") {
		t.Fatal("expected the wrong password to fail verification")
	}
}

// TestArgon2Hasher_TwoUsersSamePassword_ProduceDistinctEncodedValues cubre
// la prueba obligatoria del prompt: dos usuarios con la misma contraseña
// producen valores codificados distintos (sal aleatoria por llamada) y
// ambos verifican correctamente.
func TestArgon2Hasher_TwoUsersSamePassword_ProduceDistinctEncodedValues(t *testing.T) {
	h := auth.NewArgon2Hasher()
	const password = "misma-contraseña-para-dos-usuarios"

	encodedA, err := h.Hash(password)
	if err != nil {
		t.Fatalf("Hash A: %v", err)
	}
	encodedB, err := h.Hash(password)
	if err != nil {
		t.Fatalf("Hash B: %v", err)
	}

	if encodedA == encodedB {
		t.Fatal("expected distinct encoded hashes for two independent Hash calls with the same password")
	}
	if !h.Verify(encodedA, password) {
		t.Fatal("expected encodedA to verify")
	}
	if !h.Verify(encodedB, password) {
		t.Fatal("expected encodedB to verify")
	}
}

func TestArgon2Hasher_Verify_MalformedEncodedHash_NeverPanicsOrErrors(t *testing.T) {
	h := auth.NewArgon2Hasher()

	malformed := []string{
		"",
		"no-es-phc-en-absoluto",
		"$argon2id$v=19$m=19456,t=2,p=1$sal-invalida-no-base64!!$hash",
		"$bcrypt$v=1$costo$sal$hash",
		"$argon2id$v=1$m=19456,t=2,p=1$" + strings.Repeat("A", 22) + "$" + strings.Repeat("B", 43),
	}

	for _, encoded := range malformed {
		if h.Verify(encoded, "cualquier-contraseña") {
			t.Errorf("expected Verify to reject malformed hash %q, got true", encoded)
		}
	}
}
