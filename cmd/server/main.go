package main

import (
	"log"

	"github.com/gostuding/go-metrics/internal/server"
	"github.com/gostuding/go-metrics/internal/server/storage"
)

func run() error {
	cfg, err := server.GetFlags()
	if err != nil {
		return err
	}
	logger, err := server.InitLogger()
	if err != nil {
		return err
	}

	var strg server.Storage
	var strErr error
	if cfg.ConnectDBString == "" {
		strg, strErr = storage.NewMemStorage(cfg.Restore, cfg.FileStorePath, cfg.StoreInterval)
	} else {
		strg, strErr = storage.NewSQLStorage(cfg.ConnectDBString, logger)
	}
	if strErr != nil {
		return strErr
	}
	return server.RunServer(cfg, strg, logger)
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
