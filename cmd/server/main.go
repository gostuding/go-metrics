package main

import (
	"github.com/gostuding/go-metrics/internal/server"
	"github.com/gostuding/go-metrics/internal/server/storage"
)

func main() {
	cfg, err := server.GetFlags()
	if err != nil {
		panic(err)
	}
	storage, err := storage.NewMemStorage(cfg.Restore, cfg.FileStorePath, cfg.StoreInterval)
	if err != nil {
		panic(err)
	}

	err = server.RunServer(cfg, storage)
	if err != nil {
		panic(err)
	}
}
