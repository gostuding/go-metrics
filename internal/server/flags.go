package server

import (
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
)

// Config is struct, which contains server options.
type Config struct {
	IPAddress       string // server addres in format 'ip:port'.
	FileStorePath   string // file path if used memory storage type.
	ConnectDBString string // dsn for database connect if used sql storage type.
	Key             []byte // key for requests hash check
	StoreInterval   int    // save storage interval. Used only in memory storage type.
	Restore         bool   // flag to restore storage. Used only in memory type.
}

// Private func for get Enviroment values.
func stringEnvCheck(val string, name string) string {
	v, ok := os.LookupEnv(name)
	if ok {
		return v
	}
	return val
}

// NewConfig reads startup parameters and runtime environment variables.
// Returns Config object with server options.
func NewConfig() (*Config, error) {
	options := Config{
		IPAddress:       defaultAddress,
		FileStorePath:   filepath.Join(os.TempDir(), defaultFileName),
		ConnectDBString: "",
		Key:             []byte(defaultKey),
		StoreInterval:   defaultStoreInterval,
		Restore:         true,
	}
	var key string
	if !flag.Parsed() {
		flag.StringVar(&options.IPAddress, "a", options.IPAddress, "address and port to run server like address:port")
		flag.IntVar(&options.StoreInterval, "i", options.StoreInterval, "store interval in seconds")
		flag.StringVar(&options.FileStorePath, "f", options.FileStorePath, "file path for save the storage")
		flag.BoolVar(&options.Restore, "r", options.Restore, "restore storage on start server")
		flag.StringVar(&options.ConnectDBString, "d", options.ConnectDBString, "database connect string")
		flag.StringVar(&key, "k", "", "Key for SHA256 checks")
		flag.Parse()
	}

	if val, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		interval, err := strconv.Atoi(val)
		if err != nil {
			return nil, fmt.Errorf("STORE INTERVAL enviroment incorrect: %w", err)
		}
		options.StoreInterval = interval
	}
	options.IPAddress = stringEnvCheck(options.IPAddress, "ADDRESS")
	options.FileStorePath = stringEnvCheck(options.FileStorePath, "FILE_STORAGE_PATH")
	options.ConnectDBString = stringEnvCheck(options.ConnectDBString, "DATABASE_DSN")
	key = stringEnvCheck(key, "KEY")
	if key != "" {
		options.Key = []byte(key)
	}
	val := strings.ToLower(os.Getenv("RESTORE"))
	switch val {
	case "true":
		options.Restore = true
	case "false":
		options.Restore = false
	default:
		if val != "" {
			return nil, fmt.Errorf("enviroment RESTORE error. Use 'true' or 'false' value instead of '%s'", val)
		}
	}
	return &options, nil
}
