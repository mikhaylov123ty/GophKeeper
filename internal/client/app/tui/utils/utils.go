package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"os"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/config"
)

func EncryptData(data []byte) ([]byte, error) {
	// Пропуск обработки, если флаг не задан
	if config.GetKeys().PublicCert == "" {
		return data, nil
	}

	var err error
	var publicPEM []byte

	publicPEM, err = os.ReadFile(config.GetKeys().PublicCert)
	if err != nil {
		return nil, fmt.Errorf("failed to read public certificate: %w", err)
	}

	publicCertBlock, _ := pem.Decode(publicPEM)
	certHash := []byte(createHash(publicCertBlock.Bytes))

	block, err := aes.NewCipher(certHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return []byte(base64.StdEncoding.EncodeToString(ciphertext)), nil
}

func DeryptData(body []byte) ([]byte, error) {
	if config.GetKeys().PublicCert == "" {
		return body, nil
	}

	var err error
	var publicPEM []byte

	publicPEM, err = os.ReadFile(config.GetKeys().PublicCert)
	if err != nil {
		return nil, fmt.Errorf("failed to read public certificate: %w", err)
	}

	publicCertBlock, _ := pem.Decode(publicPEM)
	certHash := []byte(createHash(publicCertBlock.Bytes))

	block, err := aes.NewCipher(certHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	decodedBody, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to decode body: %w", err)
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := decodedBody[:nonceSize], decodedBody[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}

func createHash(key []byte) string {
	hasher := sha256.New()
	hasher.Write(key)
	return base64.StdEncoding.EncodeToString(hasher.Sum(nil))[:32]
}
