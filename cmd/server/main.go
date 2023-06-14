package main

import (
	"log"
	"os"

	"github.com/gostuding/go-metrics/internal/server"
	"github.com/gostuding/go-metrics/internal/server/storage"
	"go.uber.org/zap"
)

func checkForError(err error, logger *zap.SugaredLogger) {
	if err != nil {
		if logger != nil {
			logger.Warnln(err)
			os.Exit(1)
		}
		log.Fatalln(err)
	}
}

func main() {
	cfg, err := server.GetFlags()
	checkForError(err, nil)
	logger, err := server.InitLogger()
	checkForError(err, nil)

	if cfg.ConnectDBString != "" {
		strg, err := storage.NewSQLStorage(cfg.ConnectDBString, logger)
		checkForError(err, logger)
		checkForError(server.RunServer(cfg, strg, logger), logger)
	} else {
		storage, err := storage.NewMemStorage(cfg.Restore, cfg.FileStorePath, cfg.StoreInterval, cfg.ConnectDBString)
		checkForError(err, logger)
		checkForError(server.RunServer(cfg, storage, logger), logger)
	}

}
