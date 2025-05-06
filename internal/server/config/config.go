// Модуль config инициализирует конфигрурацию сервера
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

var cfg *ServerConfig

// ServerConfig represents the main server configuration structure.
// It includes settings for address, logging, database, cryptographic keys, and configuration file location.
type ServerConfig struct {
	Address    *Address `json:"address"`
	Logger     *Logger  `json:"logger"`
	DB         *DB      `json:"db"`
	Keys       *Keys    `json:"keys"`
	ConfigFile string   `json:"config_file"`
}

// Address represents a network address with a host and a gRPC port.
type Address struct {
	Host     string `json:"host"`
	GRPCPort string `json:"grpc_port"`
}

// Logger represents the logging configuration, including the log level setting and output format specifier.
type Logger struct {
	LogLevel  string `json:"log_level"`
	LogFormat string `json:"log_format"`
}

// DB represents the database configuration, including the data source name and the migrations directory.
type DB struct {
	DSN           string `json:"dsn"`
	MigrationsDir string `json:"migrations_dir"`
}

// Keys represents a container for cryptographic and JWT keys used for secure operations.
type Keys struct {

	// CryptoKeys represents a set of cryptographic keys including a private key and a corresponding certificate.
	CryptoKeys *CryptoKeys `json:"crypto_keys"`
	JWTKey     string      `json:"jwt_key"`
}

type CryptoKeys struct {
	Certificate string `json:"certificate"`
	PrivateKey  string `json:"private_key"`
}

// Init initializes and validates the server configuration, including flags, environment variables, and optional config file.
func Init() (*ServerConfig, error) {
	var err error
	config := &ServerConfig{
		Address: &Address{},
		Logger:  &Logger{},
		DB:      &DB{},
		Keys: &Keys{
			CryptoKeys: &CryptoKeys{},
		},
	}

	// Парсинг флагов
	config.parseFlags()

	// Инициализация конфига из файла
	if config.ConfigFile != "" {
		if err = config.InitConfigFile(); err != nil {
			return nil, fmt.Errorf("failed init config file: %w", err)
		}
	}

	// Парсинг переменных окружения
	if err = config.ParseEnv(); err != nil {
		return nil, fmt.Errorf("error parsing environment variables: %w", err)
	}

	if err = config.Validate(); err != nil {
		return nil, fmt.Errorf("error validating config: %w", err)
	}

	cfg = config

	return config, nil
}

// parseFlags parses command-line flags and initializes the fields of ServerConfig accordingly.
func (s *ServerConfig) parseFlags() {
	// Базовые флаги
	flag.StringVar(&s.Address.Host, "host", "", "Host on which to listen. Example: \"localhost\"")
	flag.StringVar(&s.Address.GRPCPort, "grpc-port", "", "Port on which to listen gRPC requests. Example: \"4443\"")

	// Флаги логирования
	flag.StringVar(&s.Logger.LogLevel, "l", "", "Log level. Example: \"info\"")

	// Флаги БД
	flag.StringVar(&s.DB.DSN, "d", "", "Host which to connect to DB. Example: \"postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable\"")
	flag.StringVar(&s.DB.MigrationsDir, "m", "", "Migrations directory. Example: \"file://./migrations\"")

	// Флаги приватного и публичного ключей
	flag.StringVar(&s.Keys.CryptoKeys.PrivateKey, "private-key", "", "Path to private key file")
	flag.StringVar(&s.Keys.CryptoKeys.Certificate, "certificate", "", "Path to public cert file")

	// Флаги приватного и публичного ключей
	flag.StringVar(&s.Keys.JWTKey, "jwt-key", "", "jwt key")

	// Флаг файла конфигурации
	flag.StringVar(&s.ConfigFile, "config", "", "Config file")

	_ = flag.Value(s.Address)
	flag.Var(s.Address, "a", "Host and port on which to listen. Example: \"localhost:8081\" or \":8081\"")

	flag.Parse()
}

// ParseEnv parses environment variables to configure the ServerConfig fields and returns an error if parsing fails.
func (s *ServerConfig) ParseEnv() error {
	var err error
	if address := os.Getenv("ADDRESS"); address != "" {
		if err = s.Address.Set(address); err != nil {
			return err
		}
	}

	if grpcPort := os.Getenv("GRPC_PORT"); grpcPort != "" {
		s.Address.GRPCPort = grpcPort
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		s.Logger.LogLevel = logLevel
	}

	if dsn := os.Getenv("DATABASE_DSN"); dsn != "" {
		s.DB.DSN = dsn
	}

	if migrationsDir := os.Getenv("MIGRATIONS_DIR"); migrationsDir != "" {
		s.DB.MigrationsDir = migrationsDir
	}

	if privateKey := os.Getenv("PRIVATE_KEY"); privateKey != "" {
		s.Keys.CryptoKeys.PrivateKey = privateKey
	}

	if certificate := os.Getenv("CERTIFICATE"); certificate != "" {
		s.Keys.CryptoKeys.Certificate = certificate
	}

	if jwtKey := os.Getenv("JWT_KEY"); jwtKey != "" {
		s.Keys.JWTKey = jwtKey
	}

	if config := os.Getenv("CONFIG_FILE"); config != "" {
		s.ConfigFile = config
	}

	return nil
}

// InitConfigFile initializes the ServerConfig by reading and unmarshaling the configuration file specified by ConfigFile.
func (s *ServerConfig) InitConfigFile() error {
	fileData, err := os.ReadFile(s.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err = json.Unmarshal(fileData, s); err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface to customize the unmarshalling process for ServerConfig.
// It merges configuration data from JSON into the current ServerConfig object while preserving prior field values.
// This includes parsing sub-objects like Address, Logger, DB, and Keys, if they are provided in the JSON.
// Returns an error if JSON unmarshalling fails or if there are inconsistencies in the provided data.
func (s *ServerConfig) UnmarshalJSON(b []byte) error {
	var err error
	var cfgFile struct {
		Address *Address `json:"address"`
		DB      *DB      `json:"db"`
		Logger  *Logger  `json:"logger"`
		Keys    *Keys    `json:"keys"`
	}

	if err = json.Unmarshal(b, &cfgFile); err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	if s.Address.Host == "" && cfgFile.Address.Host != "" {
		s.Address.Host = cfgFile.Address.Host
	}

	if s.Address.GRPCPort == "" && cfgFile.Address.GRPCPort != "" {
		s.Address.GRPCPort = cfgFile.Address.GRPCPort
	}

	// DB config file parsing
	if s.DB.DSN == "" && cfgFile.DB.DSN != "" {
		s.DB.DSN = cfgFile.DB.DSN
	}

	if s.DB.MigrationsDir == "" && cfgFile.DB.MigrationsDir != "" {
		s.DB.MigrationsDir = cfgFile.DB.MigrationsDir
	}

	//TLS keys file parsing
	if s.Keys.CryptoKeys.PrivateKey == "" && cfgFile.Keys.CryptoKeys.PrivateKey != "" {
		s.Keys.CryptoKeys.PrivateKey = cfgFile.Keys.CryptoKeys.PrivateKey
	}
	if s.Keys.CryptoKeys.Certificate == "" && cfgFile.Keys.CryptoKeys.Certificate != "" {
		s.Keys.CryptoKeys.Certificate = cfgFile.Keys.CryptoKeys.Certificate
	}

	//JWT key file parsing
	if s.Keys.JWTKey == "" && cfgFile.Keys.JWTKey != "" {
		s.Keys.JWTKey = cfgFile.Keys.JWTKey
	}

	// Logger config file parsing
	if s.Logger.LogLevel == "" && cfgFile.Logger.LogLevel != "" {
		s.Logger.LogLevel = cfgFile.Logger.LogLevel
	}
	if s.Logger.LogFormat == "" && cfgFile.Logger.LogFormat != "" {
		s.Logger.LogFormat = cfgFile.Logger.LogFormat
	}

	return nil
}

func (s *ServerConfig) Validate() error {
	if s.Keys.JWTKey == "" {
		return fmt.Errorf("JWT key is required")
	}

	if s.Keys.CryptoKeys.PrivateKey == "" {
		return fmt.Errorf("private key is required")
	} else {
		if _, err := os.ReadFile(s.Keys.CryptoKeys.PrivateKey); err != nil {
			return fmt.Errorf("private file missing: %w", err)
		}
	}

	if s.Keys.CryptoKeys.Certificate == "" {
		return fmt.Errorf("certificate is required")
	}

	return nil
}

// String returns the Address in the "host:port" format.
func (a *Address) String() string {
	return a.Host + ":" + a.GRPCPort
}

// Set parses a string in the format "host:port" and updates the Address fields Host and GRPCPort. Returns an error if the format is invalid.
func (a *Address) Set(value string) error {
	values := strings.Split(value, ":")
	if len(values) != 2 {
		return fmt.Errorf("invalid value %q, expected <host:port>:<host:port>", value)
	}

	a.Host = values[0]
	a.GRPCPort = values[1]

	return nil
}

// GetAddress retrieves the server's network address configuration and returns it as an Address pointer.
func GetAddress() *Address {
	return cfg.Address
}

// GetLogger returns the Logger instance from the server configuration.
func GetLogger() *Logger {
	return cfg.Logger
}

// GetDB retrieves the database configuration from the global server configuration and returns it as a *DB pointer.
func GetDB() *DB {
	return cfg.DB
}

// GetKeys retrieves the configuration's cryptographic and JWT keys for secure operations.
func GetKeys() *Keys {
	return cfg.Keys
}

// NewTestConfig initializes a new ServerConfig instance with default values and assigns it to the global cfg variable.
func NewTestConfig() (*ServerConfig, error) {
	config := &ServerConfig{
		Address: &Address{},
		Keys: &Keys{
			CryptoKeys: &CryptoKeys{},
		},
		Logger: &Logger{},
		DB:     &DB{},
	}

	cfg = config
	return config, nil
}
