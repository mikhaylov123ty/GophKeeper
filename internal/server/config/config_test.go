package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mikhaylov123ty/GophKeeper/internal/server/config"
)

func TestParseEnv(t *testing.T) {
	type args struct {
		env map[string]string
	}
	tests := []struct {
		name            string
		args            args
		wantAddressHost string
		wantGRPCPort    string
		wantLogLevel    string
		wantDSN         string
		wantCert        string
		wantPrivateKey  string
		wantJWTKey      string
		wantConfigFile  string
	}{
		{
			name: "set env variables correctly",
			args: args{
				env: map[string]string{
					"ADDRESS":      "127.0.0.1:9090",
					"GRPC_PORT":    "8081",
					"LOG_LEVEL":    "debug",
					"DATABASE_DSN": "dsn_value",
					"PRIVATE_KEY":  "./key.key",
					"CERTIFICATE":  "./cert.crt",
					"JWT_KEY":      "jwt",
					"CONFIG_FILE":  "/tmp/config.json",
				},
			},
			wantAddressHost: "127.0.0.1",
			wantGRPCPort:    "8081",
			wantLogLevel:    "debug",
			wantDSN:         "dsn_value",
			wantCert:        "./cert.crt",
			wantPrivateKey:  "./key.key",
			wantJWTKey:      "jwt",
			wantConfigFile:  "/tmp/config.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment
			os.Clearenv()
			// Set environment variables
			for k, v := range tt.args.env {
				os.Setenv(k, v)
			}

			cfg, err := config.NewTestConfig()
			if err != nil {
				panic(err)
			}
			err = cfg.ParseEnv()
			assert.NoError(t, err)
			assert.Equal(t, tt.wantAddressHost, cfg.Address.Host)
			assert.Equal(t, tt.wantGRPCPort, cfg.Address.GRPCPort)
			assert.Equal(t, tt.wantLogLevel, cfg.Logger.LogLevel)
			assert.Equal(t, tt.wantDSN, cfg.DB.DSN)
			assert.Equal(t, tt.wantCert, cfg.Keys.CryptoKeys.Certificate)
			assert.Equal(t, tt.wantPrivateKey, cfg.Keys.CryptoKeys.PrivateKey)
			assert.Equal(t, tt.wantJWTKey, cfg.Keys.JWTKey)
			assert.Equal(t, tt.wantConfigFile, cfg.ConfigFile)
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	type args struct {
		jsonData string
	}
	tests := []struct {
		name            string
		args            args
		wantAddressHost string
		wantGRPCPort    string
		wantLogLevel    string
		wantLogFormat   string
		wantDSN         string
		wantDBName      string
		wantPrivateKey  string
		wantCert        string
		wantJWTKey      string
		wantMigrations  string
	}{
		{
			name: "correct JSON unmarshalling",
			args: args{
				jsonData: `{
                    "address": {"host": "json_host", "grpc_port": "7777"},
                    "logger": {"log_level": "warn", "log_format": "text"},
                    "db": {"dsn": "json_dsn", "name": "json_db", "migrations_dir": "/json_migrations"},
					"keys": {"crypto_keys": {"private_key": "./key.key","certificate": "./cert.crt"}, "jwt_key": "jwt"}
                }`,
			},
			wantAddressHost: "json_host",
			wantGRPCPort:    "7777",
			wantLogLevel:    "warn",
			wantLogFormat:   "text",
			wantDSN:         "json_dsn",
			wantDBName:      "json_db",
			wantPrivateKey:  "./key.key",
			wantCert:        "./cert.crt",
			wantJWTKey:      "jwt",
			wantMigrations:  "/json_migrations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.NewTestConfig()
			if err != nil {
				panic(err)
			}
			err = cfg.UnmarshalJSON([]byte(tt.args.jsonData))
			assert.NoError(t, err)
			assert.Equal(t, tt.wantAddressHost, cfg.Address.Host)
			assert.Equal(t, tt.wantGRPCPort, cfg.Address.GRPCPort)
			assert.Equal(t, tt.wantLogLevel, cfg.Logger.LogLevel)
			assert.Equal(t, tt.wantLogFormat, cfg.Logger.LogFormat)
			assert.Equal(t, tt.wantDSN, cfg.DB.DSN)
			assert.Equal(t, tt.wantMigrations, cfg.DB.MigrationsDir)
			assert.Equal(t, tt.wantCert, cfg.Keys.CryptoKeys.Certificate)
			assert.Equal(t, tt.wantPrivateKey, cfg.Keys.CryptoKeys.PrivateKey)
			assert.Equal(t, tt.wantJWTKey, cfg.Keys.JWTKey)
		})
	}
}

func TestInitConfigFile(t *testing.T) {
	type args struct {
		fileContent string
	}
	tests := []struct {
		name              string
		args              args
		wantAddressHost   string
		wantGRPCPort      string
		wantLogLevel      string
		wantLogFormat     string
		wantDSN           string
		wantDBName        string
		wantPrivateKey    string
		wantCert          string
		wantJWTKey        string
		wantMigrationsDir string
		expectError       bool
	}{
		{
			name: "read and unmarshal config file successfully",
			args: args{
				fileContent: `{
                    "address": {"host": "file_host", "grpc_port": "8888"},
                    "logger": {"log_level": "error", "log_format": "json"},
                    "db": {"dsn": "file_dsn", "name": "file_name", "migrations_dir": "/file_migrations"},
					"keys": {"crypto_keys": {"private_key": "./key.key","certificate": "./cert.crt"}, "jwt_key": "jwt"}
                }`,
			},
			wantAddressHost:   "file_host",
			wantGRPCPort:      "8888",
			wantLogLevel:      "error",
			wantLogFormat:     "json",
			wantDSN:           "file_dsn",
			wantDBName:        "file_name",
			wantMigrationsDir: "/file_migrations",
			wantPrivateKey:    "./key.key",
			wantCert:          "./cert.crt",
			wantJWTKey:        "jwt",
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temp file with content
			tmpFile, err := os.CreateTemp("", "config*.json")
			assert.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.Write([]byte(tt.args.fileContent))
			assert.NoError(t, err)
			assert.NoError(t, tmpFile.Close())

			cfg, err := config.NewTestConfig()
			if err != nil {
				panic(err)
			}

			cfg.ConfigFile = tmpFile.Name()

			err = cfg.InitConfigFile()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantAddressHost, cfg.Address.Host)
				assert.Equal(t, tt.wantGRPCPort, cfg.Address.GRPCPort)
				assert.Equal(t, tt.wantLogLevel, cfg.Logger.LogLevel)
				assert.Equal(t, tt.wantLogFormat, cfg.Logger.LogFormat)
				assert.Equal(t, tt.wantDSN, cfg.DB.DSN)
				assert.Equal(t, tt.wantMigrationsDir, cfg.DB.MigrationsDir)
				assert.Equal(t, tt.wantCert, cfg.Keys.CryptoKeys.Certificate)
				assert.Equal(t, tt.wantPrivateKey, cfg.Keys.CryptoKeys.PrivateKey)
				assert.Equal(t, tt.wantJWTKey, cfg.Keys.JWTKey)
			}
		})
	}
}

func TestAddress_SetAndString(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name     string
		args     args
		wantHost string
		wantPort string
		wantErr  bool
	}{
		{
			name:     "valid address set",
			args:     args{value: "localhost:9090"},
			wantHost: "localhost",
			wantPort: "9090",
			wantErr:  false,
		},
		{
			name:    "invalid address format",
			args:    args{value: "invalidformat"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr := &config.Address{}
			err := addr.Set(tt.args.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantHost, addr.Host)
				assert.Equal(t, tt.wantPort, addr.GRPCPort)
				assert.Equal(t, tt.wantHost+":"+tt.wantPort, addr.String())
			}
		})
	}
}
