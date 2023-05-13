package server

import (
	"log"
	"net/http"
)

// запуск сервера на прослушку
func RunServer(ipAddress string, storage Storage) {
	log.Println("Run server at adress: ", ipAddress)
	err := http.ListenAndServe(ipAddress, makeRouter(storage))
	if err != nil {
		log.Fatal("run server error: ", err)
	}
}
