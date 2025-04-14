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

// ServerConfig - структура конфигурации сервера
type ServerConfig struct {
	Address    *Address
	Logger     *Logger
	DB         *DB
	Keys       *Keys
	configFile string
}

type Address struct {
	Host     string
	GRPCPort string
}

// Logger - структура конфигруации логгера
type Logger struct {
	LogLevel  string
	LogFormat string
}

// DB - структура конфигруации БД
type DB struct {
	Address       string
	Name          string
	MigrationsDir string
}

type Keys struct {
	HashKey   string
	CryptoKey string
	JWTKey    string
}

// Init - конструктор конфигурации сервера
func Init() (*ServerConfig, error) {
	var err error
	config := &ServerConfig{
		Address: &Address{},
		Logger:  &Logger{},
		DB:      &DB{},
	}

	// Парсинг флагов
	config.parseFlags()

	// Инициализация конфига из файла
	if config.configFile != "" {
		if err = config.initConfigFile(); err != nil {
			return nil, fmt.Errorf("failed init config file: %w", err)
		}
	}

	// Парсинг переменных окружения
	if err = config.parseEnv(); err != nil {
		return nil, fmt.Errorf("error parsing environment variables: %w", err)
	}

	cfg = config

	return config, nil
}

// Парсинг инструкций флагов сервера
func (s *ServerConfig) parseFlags() {
	// Базовые флаги
	flag.StringVar(&s.Address.Host, "host", "localhost", "Host on which to listen. Example: \"localhost\"")
	flag.StringVar(&s.Address.GRPCPort, "grpc-port", "", "Port on which to listen gRPC requests. Example: \"4443\"")

	// Флаги логирования
	flag.StringVar(&s.Logger.LogLevel, "l", "info", "Log level. Example: \"info\"")

	// Флаги БД
	flag.StringVar(&s.DB.Address, "d", "", "Host which to connect to DB. Example: \"postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable\"")

	// Флаги подписи и шифрования
	flag.StringVar(&s.Keys.HashKey, "hash-key", "", "Key")

	// Флаги приватного и публичного ключей
	flag.StringVar(&s.Keys.CryptoKey, "crypto-key", "", "Path to private crypto key file")

	// Флаги приватного и публичного ключей
	flag.StringVar(&s.Keys.CryptoKey, "jwt-key", "", "jwt key")

	// Флаг файла конфигурации
	flag.StringVar(&s.configFile, "config", "", "Config file")

	_ = flag.Value(s.Address)
	flag.Var(s.Address, "a", "Host and port on which to listen. Example: \"localhost:8081\" or \":8081\"")

	flag.Parse()
}

// Парсинг инструкций переменных окружений сервера
func (s *ServerConfig) parseEnv() error {
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

	if address := os.Getenv("DATABASE_DSN"); address != "" {
		s.DB.Address = address
	}

	if key := os.Getenv("HASH_KEY"); key != "" {
		s.Keys.HashKey = key
	}

	if privateKey := os.Getenv("CRYPTO_KEY"); privateKey != "" {
		s.Keys.CryptoKey = privateKey
	}

	if jwtKey := os.Getenv("JWT_KEY"); jwtKey != "" {
		s.Keys.JWTKey = jwtKey
	}

	if config := os.Getenv("CONFIG_FILE"); config != "" {
		s.configFile = config
	}

	return nil
}

// initConfigFile читает и инициализирует файл конфигурации
func (s *ServerConfig) initConfigFile() error {
	fileData, err := os.ReadFile(s.configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err = json.Unmarshal(fileData, s); err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return nil
}

// UnmarshalJSON реализует интерфейс Unmarshaler
// позволяет десериализировать файл конфига с условиями
func (s *ServerConfig) UnmarshalJSON(b []byte) error {
	var err error
	var cfgFile struct {
		Address *Address `json:"address"`
		DB      *DB      `json:"db"`
	}

	if err = json.Unmarshal(b, &cfgFile); err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	if s.Address.GRPCPort == "" && cfgFile.Address.GRPCPort != "" {
		s.Address.GRPCPort = cfgFile.Address.GRPCPort
	}

	// DB config file parsing
	if s.DB.Address == "" && cfgFile.DB.Address != "" {
		s.DB.Address = cfgFile.DB.Address
	}
	if s.DB.Name == "" && cfgFile.DB.Name != "" {
		s.DB.Name = cfgFile.DB.Name
	}
	if s.DB.MigrationsDir == "" && cfgFile.DB.MigrationsDir != "" {
		s.DB.MigrationsDir = cfgFile.DB.MigrationsDir
	}

	return nil
}

// String реализаует интерфейс flag.Value
func (a *Address) String() string {
	return a.Host + ":" + a.GRPCPort
}

// Set реализует интерфейса flag.Value
func (a *Address) Set(value string) error {
	values := strings.Split(value, ":")
	if len(values) != 2 {
		return fmt.Errorf("invalid value %q, expected <host:port>:<host:port>", value)
	}

	a.Host = values[0]
	a.GRPCPort = values[1]

	return nil
}

//type ServerConfig struct {
//	Address    *Address
//	Logger     *Logger
//	DB         *DB
//	Key        string
//	CryptoKey  string
//	ConfigFile string
//}

func GetAddress() *Address {
	return cfg.Address
}

func GetLogger() *Logger {
	return cfg.Logger
}

func GetDB() *DB {
	return cfg.DB
}

func GetKeys() *Keys {
	return cfg.Keys
}
