package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"log"

	"golang.org/x/crypto/argon2"
)

func DeriveKeyFromPassphrase(passphrase string) ([]byte, error) {
	salt := make([]byte, 16)

	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	b := argon2.IDKey([]byte(passphrase), salt, 3, 64*1024, 4, 32)
	return append(salt, b...), nil
}

func Encrypt(key []byte, data []byte) ([]byte, error) {
	if len(key) < 16+32 {
		return nil, errors.New("Invalid Key")
	}
	salt, key := key[:16], key[16:]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return append(salt, ciphertext...), nil
}

func Decrypt(key, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < 16 {
		return nil, errors.New("Invalid Ciphertext")
	}
	salt, ciphertext := ciphertext[:16], ciphertext[16:]

	log.Printf("[DEBUG] Decrypt: salt len=%d, first 4 bytes: %x", len(salt), salt[:4])
	subkey := argon2.IDKey(key, salt, 3, 64*1024, 4, 32)
	defer ZeroKey(subkey)
	block, err := aes.NewCipher(subkey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("Invalid Ciphertext")
	}
	nonce, cipciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	log.Printf("[DEBUG] Decrypt: nonce len=%d, ciphertext len=%d", len(nonce), len(ciphertext))
	plaintext, err := gcm.Open(nil, nonce, cipciphertext, nil)
	if err != nil {
		log.Printf("[ERROR] GCM Open failed: %v", err)
		return nil, err
	}
	return plaintext, nil
}

func ZeroKey(key []byte) {
	for i := range key {
		key[i] = 0
	}

}
