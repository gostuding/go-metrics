package main

import (
	"flag"
	"fmt"
	"net/http"
)

func main() {

	ip_address := flag.String("a", ":8080", "address and port to run server")
	flag.Parse()
	fmt.Println("Run server at adress: ", *ip_address)
	err := http.ListenAndServe(*ip_address, GetRouter())
	if err != nil {
		panic(err)
	}
}
