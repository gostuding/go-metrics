package server

import (
	"fmt"
	"net/http"
)

// запуск сервера на прослушку
func RunServer(ipAddress string, storage Storage) error {
	err := InitLogger()
	if err != nil {
		return fmt.Errorf("create logger error: %v", err)
	}
	Logger.Infoln("Run server at adress: ", ipAddress)
	err = http.ListenAndServe(ipAddress, makeRouter(storage))
	if err != nil {
		return fmt.Errorf("run server error: %v", err)
	}
	return nil
}
