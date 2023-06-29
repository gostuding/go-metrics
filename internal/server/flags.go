package server

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	IPAddress       string
	StoreInterval   int
	FileStorePath   string
	Restore         bool
	ConnectDBString string
	// Key             string
	Key []byte
}

func stringEnvCheck(val string, name string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return val
}

func GetFlags() (*Config, error) {
	var options Config
	var key string
	flag.StringVar(&options.IPAddress, "a", ":8080", "address and port to run server like address:port")
	flag.IntVar(&options.StoreInterval, "i", 300, "store interval in seconds")
	flag.StringVar(&options.FileStorePath, "f", "/tmp/metrics-db.json", "file path for save the storage")
	flag.BoolVar(&options.Restore, "r", true, "restore storage on start server")
	flag.StringVar(&options.ConnectDBString, "d", "", "database connect string")
	flag.StringVar(&key, "k", "", "Key for SHA256 checks")
	flag.Parse()
	//-------------------------------------------------------------------------
	if val := os.Getenv("STORE_INTERVAL"); val != "" {
		interval, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
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
	//-------------------------------------------------------------------------
	return &options, nil
}
