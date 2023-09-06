package main

import (
	"os"
)

func main() {
	other()
	os.Exit(1) // want "using 'os.Exit' function in main package and main function detected"
}

func other() {
	os.Exit(1)
}
