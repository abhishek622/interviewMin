package pkg

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

type Crypto struct {
	key []byte
}

func NewCrypto(key string) (*Crypto, error) {
	k := []byte(key)
	if len(k) != 32 {
		return nil, fmt.Errorf("invalid key size: must be 32 bytes")
	}
	return &Crypto{key: k}, nil
}

func (c *Crypto) Encrypt(input string) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}

	// Use AES-GCM for authenticated encryption
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// cipherText includes the auth tag
	cipherText := aesGCM.Seal(nil, nonce, []byte(input), nil)

	// final payload = nonce + ciphertext
	final := append(nonce, cipherText...)

	return base64.StdEncoding.EncodeToString(final), nil
}

// Decrypt takes base64 input and returns plaintext
func (c *Crypto) Decrypt(input string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(raw) < nonceSize {
		return "", fmt.Errorf("invalid encrypted data")
	}

	nonce := raw[:nonceSize]
	cipherText := raw[nonceSize:]

	plainText, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}
