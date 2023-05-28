package server

import (
	"log"
	"net/http"
)

// запуск сервера на прослушку
func RunServer(ipAddress string, storage Storage) error {
	log.Println("Run server at adress: ", ipAddress)
	err := http.ListenAndServe(ipAddress, makeRouter(storage))
	if err != nil {
		return err
	}
	return nil
}
