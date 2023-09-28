package server

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"path/filepath"
)

// Defaulf constans for Config.
const (
	defaultAddress       = ":8080"           // Server address
	defaultFileName      = "metrics-db.json" // MemStorage file name
	defaultKey           = "default"         // Key for hash
	defaultStoreInterval = 300               // Save MemStore interval
	falseString          = "false"
)

// Config is struct, which contains server options.
type (
	Config struct {
		PrivateKey      *rsa.PrivateKey `json:"-"`                        // rsa private key
		PrivateKeyPath  string          `json:"crypto_key,omitempty"`     //
		IPAddress       string          `json:"address,omitempty"`        // server addres in format 'ip:port'.
		FileStorePath   string          `json:"store_file,omitempty"`     // file path if used memory storage type.
		ConnectDBString string          `json:"database_dsn,omitempty"`   // database connection string.
		resString       string          `json:"-"`                        //
		Key             string          `json:"key,omitempty"`            // key for requests hash check.
		TrustedSubnet   string          `json:"trusted_subnet"`           // trusted subnet for agents
		StoreInterval   int             `json:"store_interval,omitempty"` // save storage interval.
		Restore         bool            `json:"restore,omitempty"`        // restore mem storage flag.
		SendByRPC       bool            `json:"-"`                        //
	}
	// Internal struct.
	keysStruct struct {
		HashKey        string
		PrivateKeyPath string
	}
)

// SetDefault values for Config.
func (c *Config) setDefault() {
	if c.IPAddress == "" {
		c.IPAddress = defaultAddress
	}
	if c.FileStorePath == "" {
		c.FileStorePath = filepath.Join(os.TempDir(), defaultFileName)
	}
	if c.Key == "" {
		c.Key = defaultKey
	}
	if c.StoreInterval == 0 {
		c.StoreInterval = defaultStoreInterval
	}
	if c.resString != falseString {
		c.Restore = true
	}
}

// Private func for get Enviroment values.
func stringEnvCheck(val string, name string) string {
	v, ok := os.LookupEnv(name)
	if ok {
		return v
	}
	return val
}

// parcePrivateKey reads rsa private key from file.
func parcePrivateKey(filePath string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("file read error: %w", err)
	}
	block, _ := pem.Decode([]byte(data)) //nolint:all //<-senselessly
	if block == nil {
		return nil, errors.New("failed to parse PEM block with private key")
	}
	pKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parce private key error: %w", err)
	}
	return pKey, nil
}

// lookEnviroment gets options from Enviroment.
func lookEnviroment(cfg *Config, keys *keysStruct) error {
	if val, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		interval, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("STORE INTERVAL enviroment incorrect: %w", err)
		}
		cfg.StoreInterval = interval
	}
	cfg.IPAddress = stringEnvCheck(cfg.IPAddress, "ADDRESS")
	cfg.FileStorePath = stringEnvCheck(cfg.FileStorePath, "FILE_STORAGE_PATH")
	cfg.ConnectDBString = stringEnvCheck(cfg.ConnectDBString, "DATABASE_DSN")
	keys.HashKey = stringEnvCheck(keys.HashKey, "KEY")
	if keys.HashKey != "" {
		cfg.Key = keys.HashKey
	}
	val := strings.ToLower(os.Getenv("RESTORE"))
	switch val {
	case "true":
		cfg.Restore = true
	case "false":
		cfg.Restore = false
	default:
		if val != "" {
			return fmt.Errorf("enviroment RESTORE error. Use 'true' or 'false' value instead of '%s'", val)
		}
	}
	keys.PrivateKeyPath = stringEnvCheck(keys.PrivateKeyPath, "CRYPTO_KEY")
	if keys.PrivateKeyPath != "" {
		key, err := parcePrivateKey(keys.PrivateKeyPath)
		if err != nil {
			return err
		}
		cfg.PrivateKey = key
	}
	cfg.TrustedSubnet = stringEnvCheck(cfg.TrustedSubnet, "TRUSTED_SUBNET")
	return nil
}

// lookFileConfig gets options from json file.
func lookFileConfig(path string, cfg *Config, keys *keysStruct) error {
	defer cfg.setDefault()
	if val, ok := os.LookupEnv("CONFIG"); ok {
		path = val
	}
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("config file read error: %w", err)
	}
	c := Config{Restore: true}
	err = json.Unmarshal(data, &c)
	if err != nil {
		return fmt.Errorf("config file convert error: %w", err)
	}
	if cfg.IPAddress == "" {
		cfg.IPAddress = c.IPAddress
	}
	if cfg.ConnectDBString == "" {
		cfg.ConnectDBString = c.ConnectDBString
	}
	if cfg.StoreInterval == 0 {
		cfg.StoreInterval = c.StoreInterval
	}
	if cfg.FileStorePath == "" {
		cfg.FileStorePath = c.FileStorePath
	}
	if cfg.TrustedSubnet == "" {
		cfg.TrustedSubnet = c.TrustedSubnet
	}
	if cfg.resString == "" && !c.Restore {
		cfg.resString = falseString
	}
	if keys.HashKey == "" {
		keys.HashKey = c.Key
	}
	if keys.PrivateKeyPath == "" {
		keys.PrivateKeyPath = c.PrivateKeyPath
	}
	return nil
}

// NewConfig reads startup parameters and runtime environment variables.
// Returns Config object with server options.
func NewConfig() (*Config, error) {
	cfg := Config{}
	keys := keysStruct{}
	var cfgFilePath string
	if !flag.Parsed() {
		flag.StringVar(&cfg.IPAddress, "a", "", "address and port to run server like address:port")
		flag.IntVar(&cfg.StoreInterval, "i", 0, "store interval in seconds")
		flag.StringVar(&cfg.FileStorePath, "f", "", "file path for save the storage")
		flag.StringVar(&cfg.resString, "r", "", "restore storage on start server (true or false)")
		flag.StringVar(&cfg.ConnectDBString, "d", "", "database connect string")
		flag.StringVar(&cfg.TrustedSubnet, "t", "", "trusted subnet")
		flag.StringVar(&keys.HashKey, "k", "", "Key for SHA256 checks")
		flag.StringVar(&keys.PrivateKeyPath, "crypto-key", "", "path to file with RSA private key")
		flag.StringVar(&cfgFilePath, "c", "", "path to file with config for server")
		flag.BoolVar(&cfg.SendByRPC, "rpc", cfg.SendByRPC, "Use RPC for get data from agents. Sets only by this arg")
		flag.Parse()
	}
	if err := lookFileConfig(cfgFilePath, &cfg, &keys); err != nil {
		return nil, err
	}
	if err := lookEnviroment(&cfg, &keys); err != nil {
		return nil, err
	}
	return &cfg, nil
}
