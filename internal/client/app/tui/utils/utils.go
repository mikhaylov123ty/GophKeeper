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

	"github.com/charmbracelet/lipgloss"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/config"
)

var (
	CursorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))                     // Bright purple
	SelectedStyle   = lipgloss.NewStyle().Padding(1).Foreground(lipgloss.Color("2")).Bold(true) // Green
	UnselectedStyle = lipgloss.NewStyle().Padding(1).Foreground(lipgloss.Color("7"))            // White
	BackgroundStyle = lipgloss.NewStyle().Background(lipgloss.Color("245"))                     // Grey background
	TitleStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("65"))           // Bold yellow
	SeparatorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("235"))                     // Light grey

	ListHeight = 15

	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorRed    = "\033[31m"
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

func NavigateFooter() string {
	return BackgroundStyle.Render(SeparatorStyle.Render("\nUse arrow keys to navigate and enter to select.\n")) // Navigation instructions
}

func AddItemsFooter() string {
	return BackgroundStyle.Render(SeparatorStyle.Render("\nPress Enter to save, CTRL+Q to cancel, or Backspace to delete the last character.\n"))
}

func ListItemsFooter() string {
	return BackgroundStyle.Render(SeparatorStyle.Render("\nUse arrow keys to navigate. E to edit. D to delete. Enter to select. CTRL+Q to cancel.\n"))
}

func ItemDataFooter() string {
	return BackgroundStyle.Render(SeparatorStyle.Render("\nPress CTRL+Q to return.\n"))
}

func BinaryItemDataFooter() string {
	return BackgroundStyle.Render(SeparatorStyle.Render("\nPress D to download file to output folder. CTRL+Q to return.\n"))
}

func AuthFooter() string {
	return BackgroundStyle.Render(SeparatorStyle.Render("\nPress Tab to switch fields, Enter to submit, or CTRL+Q to exit.\n"))
}
