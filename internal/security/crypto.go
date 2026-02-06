// Package security содержит функции шифрования данных.
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

var ErrDecryptFailed = errors.New("decryption failed")

// Crypto обеспечивает шифрование AES-256-GCM.
type Crypto struct {
	aead cipher.AEAD
}

// NewCrypto создаёт Crypto с ключом (32 байта для AES-256).
func NewCrypto(key []byte) (*Crypto, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Crypto{aead: aead}, nil
}

// Encrypt шифрует данные.
func (c *Crypto) Encrypt(data []byte) ([]byte, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return c.aead.Seal(nonce, nonce, data, nil), nil
}

// Decrypt расшифровывает данные.
func (c *Crypto) Decrypt(data []byte) ([]byte, error) {
	if len(data) < c.aead.NonceSize() {
		return nil, ErrDecryptFailed
	}
	nonce, ciphertext := data[:c.aead.NonceSize()], data[c.aead.NonceSize():]
	return c.aead.Open(nil, nonce, ciphertext, nil)
}
