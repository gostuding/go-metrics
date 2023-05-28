package server

import (
	"flag"
	"os"
)

// проверка флага a и переменной окружения ADDRESS
func GetFlags() string {
	ipAddress := flag.String("a", ":8080", "address and port to run server like address:port")
	flag.Parse()
	//-------------------------------------------------------------------------
	if address := os.Getenv("ADDRESS"); address != "" {
		ipAddress = &address
	}
	return *ipAddress
}
