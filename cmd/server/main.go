package main

import (
	"log"

	"github.com/gostuding/go-metrics/internal/server"
	"github.com/gostuding/go-metrics/internal/server/storage"
)

func main() {
	cfg, err := server.GetFlags()
	if err != nil {
		log.Fatalln(err)
	}
	logger, err := server.InitLogger()
	if err != nil {
		panic(err)
	}
	storage, err := storage.NewMemStorage(cfg.Restore, cfg.FileStorePath, cfg.StoreInterval)
	if err != nil {
		panic(err)
	}
	err = server.RunServer(cfg, storage, logger)
	if err != nil {
		panic(err)
	}
}
