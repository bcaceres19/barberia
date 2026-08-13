package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// tokenBytes es la cantidad de bytes de entropía cruda del token opaco de
// sesión antes de codificar: 32 bytes (256 bits) supera con margen amplio
// cualquier umbral razonable de adivinación en línea (DEC-050).
const tokenBytes = 32

// CryptoTokenGenerator implementa [TokenGenerator] con crypto/rand:
// biblioteca estándar de Go, sin dependencia adicional.
type CryptoTokenGenerator struct{}

// NewCryptoTokenGenerator construye el generador de tokens de sesión.
func NewCryptoTokenGenerator() CryptoTokenGenerator { return CryptoTokenGenerator{} }

var _ TokenGenerator = CryptoTokenGenerator{}

// New implementa [TokenGenerator]. base64.RawURLEncoding produce un valor
// apto para viajar como valor de cookie sin necesitar escape adicional (sin
// '+', '/' ni '=' de relleno).
func (CryptoTokenGenerator) New() (string, error) {
	buf := make([]byte, tokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("auth: generar token de sesión: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// HashToken calcula el hash SHA-256 hexadecimal en minúsculas de un token en
// claro, exactamente la forma que exige
// staff_session_token_hash_ck (64 caracteres) y que
// authn_resolve_session_tenant compara. El valor en claro nunca se persiste;
// solo este hash viaja al repositorio.
func HashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
