package main

import (
	"github.com/gostuding/go-metrics/internal/server"
	"github.com/gostuding/go-metrics/internal/server/storage"
)

func main() {
	options, err := server.GetFlags()
	if err != nil {
		panic(err)
	}
	storage, err := storage.NewMemStorage(options.Restore, options.FileStorePath)
	if err != nil {
		panic(err)
	}

	err = server.RunServer(options, storage)
	if err != nil {
		panic(err)
	}
}
