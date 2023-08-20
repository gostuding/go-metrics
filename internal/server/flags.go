package server

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
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

func stringEnvCheck(val string, name string) string {
	v, ok := os.LookupEnv(name)
	if ok {
		return v
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
	return &options, nil
}
