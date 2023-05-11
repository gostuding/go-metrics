package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

// начало работы
func main() {

	ipAddress := flag.String("a", ":8080", "address and port to run server like address:port")
	flag.Parse()
	//-------------------------------------------------------------------------
	if address := os.Getenv("ADDRESS"); address != "" {
		ipAddress = &address
	}
	//-------------------------------------------------------------------------
	fmt.Println("Run server at adress: ", *ipAddress)
	err := http.ListenAndServe(*ipAddress, GetRouter())
	if err != nil {
		panic(err)
	}
}
