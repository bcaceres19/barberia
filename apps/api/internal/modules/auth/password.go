package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Parámetros de argon2id (CA-005-03), elegidos de la recomendación mínima de
// OWASP Password Storage Cheat Sheet para Argon2id (m=19 MiB, t=2, p=1),
// documentados y versionados aquí porque no hay una columna de sal separada:
// el valor codificado (formato PHC) incluye algoritmo, versión, parámetros,
// sal y hash en un único texto autocontenido
// (database/modelo-fisico-referencia.sql sección A.1, staff_credential).
// Cambiar cualquiera de estos valores no invalida hashes ya almacenados con
// parámetros distintos: Verify los lee del propio texto codificado, nunca de
// estas constantes.
const (
	argon2Memory      uint32 = 19 * 1024 // KiB (19 MiB)
	argon2Time        uint32 = 2
	argon2Parallelism uint8  = 1
	argon2SaltLen     uint32 = 16
	argon2KeyLen      uint32 = 32
)

// Justificación de la dependencia (estandar-backend-go.md §5.21):
//   - Necesidad: ninguna función de derivación de clave resistente a fuerza
//     bruta existe en la biblioteca estándar de Go; argon2id es el ganador
//     de la Password Hashing Competition y la recomendación vigente de
//     OWASP para almacenamiento de contraseñas (CA-005-03: nunca cifrado
//     reversible ni hash rápido simple).
//   - Mantenimiento: golang.org/x/crypto es mantenido por el equipo de Go,
//     misma politica de compatibilidad que la biblioteca estándar.
//   - Licencia: BSD-3-Clause, igual que Go.
//   - Superficie transitiva: solo golang.org/x/sys, ya transitiva de otras
//     dependencias del proyecto.
//   - Seguridad: implementación de referencia ampliamente auditada; no se
//     reimplementa Argon2 a mano.

// Argon2Hasher implementa [PasswordHasher] con argon2id
// (database/modelo-fisico-referencia.sql: staff_credential.password_algorithm
// admite 'argon2id' o 'bcrypt'; este proyecto usa argon2id).
type Argon2Hasher struct{}

// NewArgon2Hasher construye el hasher de contraseñas. No conserva estado
// propio: los parámetros son constantes del paquete.
func NewArgon2Hasher() Argon2Hasher { return Argon2Hasher{} }

var _ PasswordHasher = Argon2Hasher{}

// Hash implementa [PasswordHasher]. La sal se genera con crypto/rand en cada
// llamada: dos usuarios con la misma contraseña obtienen valores codificados
// distintos.
func (Argon2Hasher) Hash(password string) (string, error) {
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("auth: generar sal: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Parallelism, argon2KeyLen)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argon2Memory, argon2Time, argon2Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

// Verify implementa [PasswordHasher] en tiempo constante respecto al
// contenido de password (subtle.ConstantTimeCompare), con el costo que el
// propio encodedHash declara. Un encodedHash con formato o versión que esta
// función no reconoce se trata como no coincidente en vez de propagar un
// error: el llamador (LoginService) no debe poder distinguir "credencial
// corrupta en base de datos" de "contraseña incorrecta" (CA-005-02).
func (Argon2Hasher) Verify(encodedHash, password string) bool {
	params, salt, hash, ok := parseEncodedHash(encodedHash)
	if !ok {
		return false
	}

	candidate := argon2.IDKey([]byte(password), salt, params.time, params.memory, params.parallelism, uint32(len(hash)))
	return subtle.ConstantTimeCompare(candidate, hash) == 1
}

type argon2Params struct {
	memory      uint32
	time        uint32
	parallelism uint8
}

// parseEncodedHash decodifica el formato PHC que [Argon2Hasher.Hash]
// produce. No es un parser PHC general: acepta exactamente la forma que
// este paquete escribe, y cualquier otra entrada (incluida la fila
// staff_credential.password_algorithm = 'bcrypt', legítima según el CHECK
// de la tabla pero fuera del alcance de HU-005) devuelve ok=false.
func parseEncodedHash(encoded string) (params argon2Params, salt, hash []byte, ok bool) {
	parts := strings.Split(encoded, "$")
	// "$argon2id$v=19$m=..,t=..,p=..$salt$hash" produce
	// ["", "argon2id", "v=19", "m=..,t=..,p=..", "salt", "hash"].
	if len(parts) != 6 || parts[1] != "argon2id" {
		return argon2Params{}, nil, nil, false
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return argon2Params{}, nil, nil, false
	}

	var p argon2Params
	var m, t uint32
	var par uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &par); err != nil {
		return argon2Params{}, nil, nil, false
	}
	p.memory, p.time, p.parallelism = m, t, par

	decodedSalt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return argon2Params{}, nil, nil, false
	}
	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return argon2Params{}, nil, nil, false
	}

	return p, decodedSalt, decodedHash, true
}
