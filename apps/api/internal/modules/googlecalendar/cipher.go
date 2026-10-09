package googlecalendar

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// keySize es AES-256 (DEC-102).
const keySize = 32

// Cipher cifra y descifra credenciales con AES-256-GCM y un anillo de claves:
// cada texto cifrado se guarda junto al identificador de la clave que lo
// produjo, así que rotar la clave activa no invalida lo ya cifrado mientras la
// anterior siga en el anillo (DEC-102). Es seguro para uso concurrente.
type Cipher struct {
	activeID string
	aeads    map[string]cipher.AEAD
	random   io.Reader
}

// NewCipher construye el cifrador. keys asocia id → clave de 32 bytes; activeID
// es la clave con que se cifra de ahora en adelante.
func NewCipher(activeID string, keys map[string][]byte) (*Cipher, error) {
	if activeID == "" {
		return nil, errors.New("googlecalendar: falta el identificador de la clave activa")
	}
	if _, ok := keys[activeID]; !ok {
		return nil, fmt.Errorf("googlecalendar: la clave activa %q no está en el anillo", activeID)
	}
	aeads := make(map[string]cipher.AEAD, len(keys))
	for id, key := range keys {
		if id == "" {
			return nil, errors.New("googlecalendar: una clave del anillo no tiene identificador")
		}
		if len(key) != keySize {
			return nil, fmt.Errorf("googlecalendar: la clave %q debe medir %d bytes (mide %d)", id, keySize, len(key))
		}
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, fmt.Errorf("googlecalendar: clave %q: %w", id, err)
		}
		aead, err := cipher.NewGCM(block)
		if err != nil {
			return nil, fmt.Errorf("googlecalendar: clave %q: %w", id, err)
		}
		aeads[id] = aead
	}
	return &Cipher{activeID: activeID, aeads: aeads, random: rand.Reader}, nil
}

// ParseKey decodifica una clave de 32 bytes en base64 estándar.
func ParseKey(encoded string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, errors.New("googlecalendar: la clave debe estar en base64")
	}
	if len(key) != keySize {
		return nil, fmt.Errorf("googlecalendar: la clave debe medir %d bytes (mide %d)", keySize, len(key))
	}
	return key, nil
}

// ActiveKeyID devuelve el identificador con que se cifra ahora.
func (c *Cipher) ActiveKeyID() string { return c.activeID }

// Encrypt cifra plaintext con la clave activa y devuelve nonce || texto cifrado
// (con la etiqueta de autenticación) y el identificador de la clave. aad se
// autentica junto al texto —por ejemplo la identidad del recurso— sin guardarse.
func (c *Cipher) Encrypt(plaintext, aad []byte) (ciphertext []byte, keyID string, err error) {
	aead := c.aeads[c.activeID]
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(c.random, nonce); err != nil {
		return nil, "", fmt.Errorf("googlecalendar: generar nonce: %w", err)
	}
	return aead.Seal(nonce, nonce, plaintext, aad), c.activeID, nil
}

// Decrypt descifra con la clave identificada por keyID. Una clave desconocida,
// un texto alterado o un aad distinto producen error, nunca texto parcial.
func (c *Cipher) Decrypt(ciphertext []byte, keyID string, aad []byte) ([]byte, error) {
	aead, ok := c.aeads[keyID]
	if !ok {
		return nil, fmt.Errorf("googlecalendar: clave de cifrado %q desconocida", keyID)
	}
	if len(ciphertext) < aead.NonceSize()+aead.Overhead() {
		return nil, errors.New("googlecalendar: texto cifrado demasiado corto")
	}
	nonce, sealed := ciphertext[:aead.NonceSize()], ciphertext[aead.NonceSize():]
	plaintext, err := aead.Open(nil, nonce, sealed, aad)
	if err != nil {
		return nil, errors.New("googlecalendar: no se pudo descifrar (clave, datos o contexto incorrectos)")
	}
	return plaintext, nil
}
