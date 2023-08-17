package main

import (
	"log"

	"github.com/gostuding/go-metrics/internal/server"
	"github.com/gostuding/go-metrics/internal/server/storage"
	"go.uber.org/zap"
)

func run(logger *zap.SugaredLogger) error {
	var strg server.Storage
	var strErr error

	cfg, err := server.GetFlags()
	if err != nil {
		return err
	}
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
	logger, err := server.InitLogger()
	if err != nil {
		log.Fatal(err)
	}

	err = run(logger)
	if err != nil {
		logger.Fatalln(err)
	}
}
