package main

import (
	"flag"
	"fmt"
	"net/http"
)

func main() {

	ipAddress := flag.String("a", ":8080", "address and port to run server")
	flag.Parse()
	fmt.Println("Run server at adress: ", *ipAddress)
	err := http.ListenAndServe(*ipAddress, GetRouter())
	if err != nil {
		panic(err)
	}
}
