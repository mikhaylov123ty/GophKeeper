// Модуль config инициализирует конфигрурацию агента
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

// ClientConfig - структура конфигурации агента
type ClientConfig struct {
	Address    *Address
	ConfigFile string
	Key        string
}
type Address struct {
	Host     string
	GRPCPort string
}

// New - конструктор конфигурации агента
func New() (*ClientConfig, error) {
	var err error
	config := &ClientConfig{Address: &Address{}}

	// Парсинг флагов
	config.parseFlags()

	// Инициализация конфига из файла
	if config.ConfigFile != "" {
		if err = config.initConfigFile(); err != nil {
			return nil, fmt.Errorf("failed init config file: %w", err)
		}
	}

	// Парсинг переменных окружения
	if err = config.parseEnv(); err != nil {
		return nil, fmt.Errorf("error parsing environment variables: %w", err)
	}

	return config, nil
}

// parseFlags - Парсинг инструкций флагов агента
func (a *ClientConfig) parseFlags() {
	// Базовые флаги
	flag.StringVar(&a.Address.Host, "host", "localhost", "Host on which to listen. Example: \"localhost\"")
	flag.StringVar(&a.Address.GRPCPort, "grpc-port", "", "Port on which to listen gRPC requests. Example: \"443\"")

	// Флаги подписи
	flag.StringVar(&a.Key, "k", "", "Key")

	// Флаг файла конфигурации
	flag.StringVar(&a.ConfigFile, "config", "", "Config file")

	_ = flag.Value(a.Address)
	flag.Var(a.Address, "a", "Host and port on which to listen gRPC requests. Example: \"localhost:443\" or \":443\"")

	flag.Parse()
}

// parseEnv - Парсинг инструкций переменных окружений агента
func (a *ClientConfig) parseEnv() error {
	if address := os.Getenv("ADDRESS"); address != "" {
		if err := a.Address.Set(address); err != nil {
			return fmt.Errorf("error setting ADDRESS: %w", err)
		}
	}

	if key := os.Getenv("KEY"); key != "" {
		a.Key = key
	}

	if config := os.Getenv("CONFIG"); config != "" {
		a.ConfigFile = config
	}

	if grpcPort := os.Getenv("GRPC_PORT"); grpcPort != "" {
		a.Address.GRPCPort = grpcPort
	}

	return nil
}

// initConfigFile читает и инициализирует файл конфигурации
func (a *ClientConfig) initConfigFile() error {
	fileData, err := os.ReadFile(a.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err = json.Unmarshal(fileData, a); err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return nil
}

// UnmarshalJSON реализует интерфейс Unmarshaler
// позволяет десериализировать файл конфига с условиями
func (a *ClientConfig) UnmarshalJSON(b []byte) error {
	var err error
	var cfg struct {
		GRPCPort       string `json:"grpc_port"`
		ReportInterval string `json:"report_interval"`
		PollInterval   string `json:"poll_interval"`
		CryptoKey      string `json:"crypto_key"`
	}

	if err = json.Unmarshal(b, &cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	if a.Address.GRPCPort == "" && cfg.GRPCPort != "" {
		a.Address.GRPCPort = cfg.GRPCPort
	}

	return nil
}

// String реализует интерфейс flag.Value
func (a *Address) String() string {
	return a.Host + ":" + a.GRPCPort
}

// Set реализует интерфейс flag.Value
func (a *Address) Set(value string) error {
	values := strings.Split(value, ":")
	if len(values) != 2 {
		return fmt.Errorf("invalid value %q, expected <host:port>:<host:port>", value)
	}

	a.Host = values[0]
	a.GRPCPort = values[1]

	return nil
}
