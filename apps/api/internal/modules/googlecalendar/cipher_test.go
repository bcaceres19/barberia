package googlecalendar_test

import (
	"bytes"
	"encoding/base64"
	"testing"

	"system-barbershop/internal/modules/googlecalendar"
)

func key(fill byte) []byte { return bytes.Repeat([]byte{fill}, 32) }

func mustCipher(t *testing.T, active string, keys map[string][]byte) *googlecalendar.Cipher {
	t.Helper()
	c, err := googlecalendar.NewCipher(active, keys)
	if err != nil {
		t.Fatalf("NewCipher: %v", err)
	}
	return c
}

func TestCipher_RoundTripNeverStoresPlaintextAndUsesFreshNonces(t *testing.T) {
	c := mustCipher(t, "v1", map[string][]byte{"v1": key(1)})
	secret := []byte("1//0g-refresh-token-de-prueba")

	first, id, err := c.Encrypt(secret, []byte("ctx"))
	if err != nil || id != "v1" {
		t.Fatalf("Encrypt: id=%q err=%v", id, err)
	}
	second, _, _ := c.Encrypt(secret, []byte("ctx"))
	if bytes.Equal(first, second) {
		t.Fatal("dos cifrados del mismo texto deben diferir (nonce nuevo)")
	}
	if bytes.Contains(first, secret) {
		t.Fatal("el texto cifrado no debe contener el secreto en claro")
	}
	plain, err := c.Decrypt(first, "v1", []byte("ctx"))
	if err != nil || !bytes.Equal(plain, secret) {
		t.Fatalf("Decrypt: %q err=%v", plain, err)
	}
}

func TestCipher_RejectsWrongKeyTamperingAndWrongContext(t *testing.T) {
	c := mustCipher(t, "v1", map[string][]byte{"v1": key(1), "v9": key(9)})
	sealed, _, _ := c.Encrypt([]byte("secreto"), []byte("barbero-A"))

	if _, err := c.Decrypt(sealed, "v9", []byte("barbero-A")); err == nil {
		t.Fatal("una clave distinta no debe descifrar")
	}
	if _, err := c.Decrypt(sealed, "desconocida", []byte("barbero-A")); err == nil {
		t.Fatal("un key_id desconocido no debe descifrar")
	}
	if _, err := c.Decrypt(sealed, "v1", []byte("barbero-B")); err == nil {
		t.Fatal("un contexto distinto (otro barbero) no debe descifrar")
	}
	tampered := append([]byte(nil), sealed...)
	tampered[len(tampered)-1] ^= 0xFF
	if _, err := c.Decrypt(tampered, "v1", []byte("barbero-A")); err == nil {
		t.Fatal("un texto alterado no debe descifrar")
	}
	if _, err := c.Decrypt([]byte("corto"), "v1", nil); err == nil {
		t.Fatal("un texto demasiado corto no debe descifrar")
	}
}

func TestCipher_RotationKeepsOldCiphertextReadable(t *testing.T) {
	old := mustCipher(t, "v1", map[string][]byte{"v1": key(1)})
	sealedOld, idOld, _ := old.Encrypt([]byte("token-viejo"), nil)

	rotated := mustCipher(t, "v2", map[string][]byte{"v1": key(1), "v2": key(2)})
	if rotated.ActiveKeyID() != "v2" {
		t.Fatalf("la clave activa debe ser v2, es %q", rotated.ActiveKeyID())
	}
	plain, err := rotated.Decrypt(sealedOld, idOld, nil)
	if err != nil || string(plain) != "token-viejo" {
		t.Fatalf("lo cifrado con v1 debe seguir legible tras rotar: %q %v", plain, err)
	}
	sealedNew, idNew, _ := rotated.Encrypt([]byte("token-nuevo"), nil)
	if idNew != "v2" {
		t.Fatalf("lo nuevo se cifra con v2, no con %q", idNew)
	}

	// Sin la clave anterior en el anillo, lo viejo deja de ser legible.
	onlyNew := mustCipher(t, "v2", map[string][]byte{"v2": key(2)})
	if _, err := onlyNew.Decrypt(sealedOld, idOld, nil); err == nil {
		t.Fatal("sin v1 en el anillo, lo cifrado con v1 no debe descifrarse")
	}
	if _, err := onlyNew.Decrypt(sealedNew, idNew, nil); err != nil {
		t.Fatalf("lo cifrado con v2 sí: %v", err)
	}
}

func TestNewCipher_ValidatesTheKeyring(t *testing.T) {
	cases := map[string]struct {
		active string
		keys   map[string][]byte
	}{
		"sin clave activa":        {"", map[string][]byte{"v1": key(1)}},
		"activa fuera del anillo": {"v2", map[string][]byte{"v1": key(1)}},
		"clave de 16 bytes":       {"v1", map[string][]byte{"v1": bytes.Repeat([]byte{1}, 16)}},
		"identificador vacío":     {"v1", map[string][]byte{"v1": key(1), "": key(2)}},
		"anillo vacío":            {"v1", map[string][]byte{}},
		"clave de 33 bytes":       {"v1", map[string][]byte{"v1": bytes.Repeat([]byte{1}, 33)}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := googlecalendar.NewCipher(tc.active, tc.keys); err == nil {
				t.Fatal("se esperaba un error")
			}
		})
	}
}

func TestParseKey(t *testing.T) {
	good := base64.StdEncoding.EncodeToString(key(7))
	if got, err := googlecalendar.ParseKey(good); err != nil || !bytes.Equal(got, key(7)) {
		t.Fatalf("clave válida: %v", err)
	}
	for name, bad := range map[string]string{
		"no base64": "###",
		"16 bytes":  base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 16)),
		"vacía":     "",
	} {
		if _, err := googlecalendar.ParseKey(bad); err == nil {
			t.Fatalf("%s: se esperaba un error", name)
		}
	}
}
