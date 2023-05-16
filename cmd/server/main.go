package main

import (
	"log"

	"github.com/gostuding/go-metrics/internal/server"
	"github.com/gostuding/go-metrics/internal/server/storage"
)

func main() {
	address := server.GetFlags()
	storage := &storage.MemStorage{}
	err := server.RunServer(address, storage)
	if err != nil {
		log.Fatalln("run server error: ", err)
	}
}
