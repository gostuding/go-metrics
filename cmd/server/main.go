package main

import (
	"github.com/gostuding/go-metrics/internal/server"
	"github.com/gostuding/go-metrics/internal/server/storage"
)

// для запуска сервера необходимо передать адрес и объект хранилища. Логика обработки скрыта в internal
func main() {
	address := server.GetFlags()
	storage := &storage.MemStorage{}
	server.RunServer(address, storage)
}
