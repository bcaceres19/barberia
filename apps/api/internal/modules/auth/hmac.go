package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// HMACHex calcula HMAC-SHA256(secret, value) en hexadecimal minúsculas (64
// caracteres), la representación que login_throttle.ip_hash y
// auth_phone_challenge.ip_hash/code_hash exigen (HU-007, DEC-062). Nunca se
// usa SHA-256 simple para IP o código: ambos son espacios de baja entropía
// (~2^32 direcciones IPv4, 10^6 códigos de 6 dígitos) reconstruibles por
// fuerza bruta offline sin un secreto; HMAC con la clave de despliegue
// (nunca en el repositorio) no lo es.
func HMACHex(value string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}
