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

// Address represents a network location with a host and a gRPC port.
type Address struct {

	// Host specifies the network host address for the gRPC connection.
	Host     string `json:"host"`
	GRPCPort string `json:"grpc_port"`
}

type Keys struct {
	PublicCert string `json:"public_cert"`
}

// New initializes a new instance of ClientConfig, parsing flags, environment variables, and potentially a config file.
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

	if err = config.Validate(); err != nil {
		return nil, fmt.Errorf("error validating config: %w", err)
	}

	cfg = config

	return config, nil
}

// parseFlags configures the command-line flags and parses their values into the ClientConfig structure.
func (a *ClientConfig) parseFlags() {
	// Базовые флаги
	flag.StringVar(&a.Address.Host, "host", "", "Host on which to listen. Example: \"localhost\"")
	flag.StringVar(&a.Address.GRPCPort, "grpc-port", "", "Port on which to listen gRPC requests. Example: \"443\"")

	// Флаги подписи
	flag.StringVar(&a.Keys.PublicCert, "certificate", "", "TLS public cert file")

	// Флаг файла конфигурации
	flag.StringVar(&a.ConfigFile, "config", "", "Config file")

	flag.StringVar(&a.OutputFolder, "files-output", "", "Output folder for downloaded files.")

	_ = flag.Value(a.Address)
	flag.Var(a.Address, "a", "Host and port on which to listen gRPC requests. Example: \"localhost:443\" or \":443\"")

	flag.Parse()
}

// ParseEnv reads configuration values from environment variables and updates the ClientConfig instance accordingly.
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

// InitConfigFile reads the configuration file specified in ClientConfig, parses its contents, and initializes the structure.
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

// UnmarshalJSON parses JSON-encoded data and updates the ClientConfig object.
// It conditionally updates empty fields using the values from the JSON structure.
// Returns an error if the JSON data is invalid or fails to unmarshal.
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

	if a.Address.Host == "" && cfgFile.Address.Host != "" {
		a.Address.Host = cfgFile.Address.Host
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

func (a *ClientConfig) Validate() error {
	if a.Keys.PublicCert == "" {
		return fmt.Errorf("certificate is required")
	}

	// Check if folder exists and is persistent
	info, err := os.Stat(a.OutputFolder)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("output folder does not exist: %w", err)
		}
		return fmt.Errorf("error checking output folder: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("output folder is not a directory")
	}

	return nil
}

// String formats the Address as a string in the format "host:port".
func (a *Address) String() string {
	return a.Host + ":" + a.GRPCPort
}

// Set parses a value in the format "host:port" and updates the Address fields Host and GRPCPort. Returns an error if invalid.
func (a *Address) Set(value string) error {
	values := strings.Split(value, ":")
	if len(values) != 2 {
		return fmt.Errorf("invalid value %q, expected <host:port>:<host:port>", value)
	}

	a.Host = values[0]
	a.GRPCPort = values[1]

	return nil
}

// GetAddress returns the configured Address instance for gRPC communication.
func GetAddress() *Address {
	return cfg.Address
}

// GetKeys retrieves the configured Keys instance, containing public certificate details, from the global configuration.
func GetKeys() *Keys {
	return cfg.Keys
}

// GetOutputFolder returns the output folder path configured in the ClientConfig.
func GetOutputFolder() string { return cfg.OutputFolder }

// NewTestConfig initializes a new ClientConfig instance with default Address and Keys and returns it.
func NewTestConfig() (*ClientConfig, error) {
	config := &ClientConfig{
		Address: &Address{},
		Keys:    &Keys{},
	}

	cfg = config
	return config, nil
}
