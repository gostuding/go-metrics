package main

import (
	"net/http"
)

func main() {
	err := http.ListenAndServe(`:8080`, GetRouter())
	if err != nil {
		panic(err)
	}
}
