package main

import (
	"github.com/gostuding/go-metrics/internal/server"
	"github.com/gostuding/go-metrics/internal/server/storage"
)

func main() {
	address := server.GetFlags()
	storage := storage.NewMemStorage()
	err := server.RunServer(address, storage)
	if err != nil {
		panic(err)
	}
}
