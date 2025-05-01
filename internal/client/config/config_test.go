package config_test

import (
	"encoding/json"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

// TestAddress_Set tests the Address.Set method
func TestAddress_Set(t *testing.T) {
	addr := &config.Address{}

	// valid input
	err := addr.Set("localhost:1234")
	require.NoError(t, err)
	assert.Equal(t, "localhost", addr.Host)
	assert.Equal(t, "1234", addr.GRPCPort)

	// invalid input
	err = addr.Set("invalid")
	assert.Error(t, err)
}

// TestUnmarshalJSON tests the custom unmarshalling logic
func TestUnmarshalJSON(t *testing.T) {
	var cfgFile struct {
		Address        *config.Address `json:"address"`
		ReportInterval string          `json:"report_interval"`
		PollInterval   string          `json:"poll_interval"`
		PublicCert     string          `json:"public_cert"`
		OutputFolder   string          `json:"files_output_folder"`
	}

	jsonData := `{
        "address": {"host": "127.0.0.1", "grpc_port": "8080"},
        "report_interval": "10s",
        "poll_interval": "5s",
        "public_cert": "cert.pem",
        "files_output_folder": "/tmp/output"
    }`

	err := json.Unmarshal([]byte(jsonData), &cfgFile)
	require.NoError(t, err)

	assert.Equal(t, "127.0.0.1", cfgFile.Address.Host)
	assert.Equal(t, "8080", cfgFile.Address.GRPCPort)
	assert.Equal(t, "cert.pem", cfgFile.PublicCert)
	assert.Equal(t, "/tmp/output", cfgFile.OutputFolder)
}

// TestParseEnv overrides environment variables
func TestParseEnv(t *testing.T) {
	os.Setenv("ADDRESS", "envhost:5555")
	os.Setenv("PUBLIC_CERT", "envcert.pem")
	os.Setenv("CONFIG", "config.json")
	os.Setenv("GRPC_PORT", "9999")
	os.Setenv("OUTPUT_FOLDER", "/env/output")
	defer func() {
		os.Unsetenv("ADDRESS")
		os.Unsetenv("PUBLIC_CERT")
		os.Unsetenv("CONFIG")
		os.Unsetenv("GRPC_PORT")
		os.Unsetenv("OUTPUT_FOLDER")
	}()

	cfg := &config.ClientConfig{
		Address: &config.Address{},
		Keys:    &config.Keys{},
	}

	err := cfg.ParseEnv()
	require.NoError(t, err)

	assert.Equal(t, "envhost", cfg.Address.Host)
	assert.Equal(t, "9999", cfg.Address.GRPCPort)
	assert.Equal(t, "envcert.pem", cfg.Keys.PublicCert)
	assert.Equal(t, "config.json", cfg.ConfigFile)
	assert.Equal(t, "/env/output", cfg.OutputFolder)
}

// TestInitConfigFile reads a sample config file
func TestInitConfigFile(t *testing.T) {
	// Create a temporary file with JSON content
	tmpFile, err := os.CreateTemp("", "config*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	jsonContent := `{
        "address": {"host": "localhost", "grpc_port": "7777"},
        "public_cert": "certfile.pem",
        "files_output_folder": "/tmp/output"
    }`

	_, err = tmpFile.Write([]byte(jsonContent))
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	cfg := &config.ClientConfig{
		ConfigFile: tmpFile.Name(),
		Address:    &config.Address{},
		Keys:       &config.Keys{},
	}

	err = cfg.InitConfigFile()
	require.NoError(t, err)

	assert.Equal(t, "7777", cfg.Address.GRPCPort)
	assert.Equal(t, "certfile.pem", cfg.Keys.PublicCert)
	assert.Equal(t, "/tmp/output", cfg.OutputFolder)
}

// TestNew combines multiple parts
func TestNew(t *testing.T) {
	// Set flags
	os.Args = []string{"cmd", "-host=127.0.0.1", "-grpc-port=8888", "-cert=mycert.pem", "-files-output=./temp/"}

	// Clear environment variables
	os.Unsetenv("ADDRESS")
	os.Unsetenv("PUBLIC_CERT")
	os.Unsetenv("CONFIG")
	os.Unsetenv("GRPC_PORT")
	os.Unsetenv("OUTPUT_FOLDER")

	cfg, err := config.New()
	require.NoError(t, err)

	assert.Equal(t, "127.0.0.1", cfg.Address.Host)
	assert.Equal(t, "8888", cfg.Address.GRPCPort)
	assert.Equal(t, "mycert.pem", cfg.Keys.PublicCert)
	assert.Equal(t, "./temp/", cfg.OutputFolder)
}
