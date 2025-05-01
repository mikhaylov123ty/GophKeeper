// Модуль config инициализирует конфигрурацию агента
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

var cfg *ClientConfig

// ClientConfig - структура конфигурации агента
type ClientConfig struct {
	Address      *Address `json:"address"`
	ConfigFile   string
	Keys         *Keys
	OutputFolder string
}
type Address struct {
	Host     string `json:"host"`
	GRPCPort string `json:"grpc_port"`
}

type Keys struct {
	PublicCert string `json:"public_cert"`
}

// New - конструктор конфигурации агента
func New() (*ClientConfig, error) {
	var err error
	config := &ClientConfig{
		Address: &Address{},
		Keys:    &Keys{},
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

	cfg = config

	return config, nil
}

// parseFlags - Парсинг инструкций флагов агента
func (a *ClientConfig) parseFlags() {
	// Базовые флаги
	flag.StringVar(&a.Address.Host, "host", "localhost", "Host on which to listen. Example: \"localhost\"")
	flag.StringVar(&a.Address.GRPCPort, "grpc-port", "", "Port on which to listen gRPC requests. Example: \"443\"")

	// Флаги подписи
	flag.StringVar(&a.Keys.PublicCert, "cert", "", "TLS public cert file")

	// Флаг файла конфигурации
	flag.StringVar(&a.ConfigFile, "config", "", "Config file")

	flag.StringVar(&a.OutputFolder, "files-output", "", "Output folder for downloaded files.")

	_ = flag.Value(a.Address)
	flag.Var(a.Address, "a", "Host and port on which to listen gRPC requests. Example: \"localhost:443\" or \":443\"")

	flag.Parse()
}

// parseEnv - Парсинг инструкций переменных окружений агента
func (a *ClientConfig) ParseEnv() error {
	if address := os.Getenv("ADDRESS"); address != "" {
		if err := a.Address.Set(address); err != nil {
			return fmt.Errorf("error setting ADDRESS: %w", err)
		}
	}

	if publicCert := os.Getenv("PUBLIC_CERT"); publicCert != "" {
		a.Keys.PublicCert = publicCert
	}

	if config := os.Getenv("CONFIG"); config != "" {
		a.ConfigFile = config
	}

	if grpcPort := os.Getenv("GRPC_PORT"); grpcPort != "" {
		a.Address.GRPCPort = grpcPort
	}

	if outputFolder := os.Getenv("OUTPUT_FOLDER"); outputFolder != "" {
		a.OutputFolder = outputFolder
	}

	return nil
}

// initConfigFile читает и инициализирует файл конфигурации
func (a *ClientConfig) InitConfigFile() error {
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
	var cfgFile struct {
		Address      *Address `json:"address"`
		PublicCert   string   `json:"public_cert"`
		OutputFolder string   `json:"files_output_folder"`
	}

	if err = json.Unmarshal(b, &cfgFile); err != nil {
		return fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	if a.Address.GRPCPort == "" && cfgFile.Address.GRPCPort != "" {
		a.Address.GRPCPort = cfgFile.Address.GRPCPort
	}

	if a.Keys.PublicCert == "" && cfgFile.PublicCert != "" {
		a.Keys.PublicCert = cfgFile.PublicCert
	}

	if a.OutputFolder == "" && cfgFile.OutputFolder != "" {
		a.OutputFolder = cfgFile.OutputFolder
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

func GetAddress() *Address {
	return cfg.Address
}

func GetKeys() *Keys {
	fmt.Printf("config %+v\n", cfg)
	return cfg.Keys
}

func GetOutputFolder() string { return cfg.OutputFolder }

func NewTestConfig() (*ClientConfig, error) {
	config := &ClientConfig{
		Address: &Address{},
		Keys:    &Keys{},
	}

	cfg = config
	return config, nil
}
