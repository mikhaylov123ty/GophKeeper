package utils_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/utils"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/config"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

var cfg *config.ClientConfig

func init() {
	var err error
	cfg, err = config.NewTestConfig()
	if err != nil {
		panic(err)
	}
}

func setupTestCert(t *testing.T) (string, func()) {
	// Create a temporary directory for cert
	tempDir := t.TempDir()
	pubKeyPath := filepath.Join(tempDir, "public.pem")

	// Generate RSA key pair for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	assert.NoError(t, err)

	pubPem := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	err = os.WriteFile(pubKeyPath, pubPem, 0644)
	assert.NoError(t, err)

	// Return path and cleanup function
	return pubKeyPath, func() {
		os.Remove(pubKeyPath)
	}
}

func TestEncryptData(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "encrypts data successfully",
			args:    args{data: []byte("hello world")},
			want:    nil, // We can't predict encrypted output, so want nil for now
			wantErr: assert.NoError,
		},
		{
			name:    "returns data as-is when PublicCert is empty",
			args:    args{data: []byte("no encryption")},
			want:    []byte("no encryption"),
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "encrypts data successfully" {
				// Setup a valid cert
				pubPath, cleanup := setupTestCert(t)
				defer cleanup()

				cfg.Keys.PublicCert = pubPath

				fmt.Printf("cfg %+v\n", cfg)

			} else {
				// For the case of empty cert, ensure no config set
				cfg.Keys.PublicCert = ""
			}

			got, err := utils.EncryptData(tt.args.data)
			if !tt.wantErr(t, err, fmt.Sprintf("EncryptData(%v)", tt.args.data)) {
				return
			}
			if tt.want == nil {
				// For encrypted data, just check that output is not equal to input
				assert.NotNil(t, got)
				assert.NotEqual(t, tt.args.data, got)
			} else {
				// For cases where data should be unchanged
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestDeryptData(t *testing.T) {
	type args struct {
		body []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "decrypts data successfully",
			args:    args{}, // will set up below
			want:    []byte("secret message"),
			wantErr: assert.NoError,
		},
		{
			name:    "returns data as-is when PublicCert is empty",
			args:    args{body: []byte("some data")},
			want:    []byte("some data"),
			wantErr: assert.NoError,
		},
		{
			name:    "fails to decrypt with invalid data",
			args:    args{body: []byte("invalid data")},
			want:    nil,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "decrypts data successfully":
				// Generate test cert and key
				pubPath, cleanup := setupTestCert(t)
				defer cleanup()

				// Set config
				cfg.Keys = &config.Keys{
					PublicCert: pubPath,
				}

				// Encrypt data
				plainText := []byte("secret message")
				encrypted, err := utils.EncryptData(plainText)
				assert.NoError(t, err)

				tt.args.body = encrypted

			case "returns data as-is when PublicCert is empty":
				cfg.Keys = &config.Keys{
					PublicCert: "",
				}
				tt.args.body = []byte("some data")
			case "fails to decrypt with invalid data":
				cfg.Keys = &config.Keys{
					PublicCert: "bad cert",
				}
				// Pass invalid base64 data
				tt.args.body = []byte("not base64!!!")
			}

			got, err := utils.DeryptData(tt.args.body)
			if !tt.wantErr(t, err, fmt.Sprintf("DeryptData(%v)", tt.args.body)) {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
