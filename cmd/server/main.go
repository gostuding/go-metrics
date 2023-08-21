package main

import (
	"fmt"
	"log"

	"github.com/gostuding/go-metrics/internal/server"
	"github.com/gostuding/go-metrics/internal/server/storage"
	"go.uber.org/zap"
)

func run(logger *zap.SugaredLogger) error {
	var strg server.Storage
	var strErr error

	cfg, err := server.NewConfig()
	if err != nil {
		return fmt.Errorf("create config error: %w", err)
	}
	if cfg.ConnectDBString == "" {
		strg, strErr = storage.NewMemStorage(cfg.Restore, cfg.FileStorePath, cfg.StoreInterval)
	} else {
		strg, strErr = storage.NewSQLStorage(cfg.ConnectDBString)
	}
	if strErr != nil {
		return fmt.Errorf("storage error: %w", err)
	}
	srv := server.NewServer(cfg, logger, strg)
	return srv.RunServer() //nolint:wrapcheck //<-senselessly
}

func main() {
	logger, err := server.NewLogger()
	if err != nil {
		log.Fatal(err)
	}
	err = run(logger)
	if err != nil {
		logger.Fatalln(err)
	}
}
