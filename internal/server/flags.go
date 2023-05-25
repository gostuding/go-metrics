package server

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"
)

type ServerOptions struct {
	IpAddress     string
	StoreInterval int
	FileStorePath string
	Restore       bool
}

func GetFlags() (*ServerOptions, error) {
	var options ServerOptions
	flag.StringVar(&options.IpAddress, "a", ":8080", "address and port to run server like address:port")
	flag.IntVar(&options.StoreInterval, "i", 300, "store interval in seconds")
	flag.StringVar(&options.FileStorePath, "f", "/tmp/metrics-db.json", "file path for save the storage")
	flag.BoolVar(&options.Restore, "r", true, "restore storage on start server")
	flag.Parse()
	//-------------------------------------------------------------------------
	if val := os.Getenv("ADDRESS"); val != "" {
		options.IpAddress = val
	}
	if val := os.Getenv("STORE_INTERVAL"); val != "" {
		interval, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
		}
		options.StoreInterval = interval
	}
	if val := os.Getenv("FILE_STORAGE_PATH"); val != "" {
		options.FileStorePath = val
	}
	val := strings.ToLower(os.Getenv("RESTORE"))
	switch val {
	case "true":
		options.Restore = true
	case "false":
		options.Restore = false
	default:
		if val != "" {
			return nil, errors.New("enviroment RESTORE error. Use 'true' or 'false' value")
		}
	}
	//-------------------------------------------------------------------------
	return &options, nil
}
