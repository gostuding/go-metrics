package server

import (
	"flag"
	"os"
)

// проверка флага и переменной окружения
func GetFlags() string {
	ipAddress := flag.String("a", ":8080", "address and port to run server like address:port")
	flag.Parse()
	//-------------------------------------------------------------------------
	if address := os.Getenv("ADDRESS"); address != "" {
		ipAddress = &address
	}
	return *ipAddress
}
