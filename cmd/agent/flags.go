package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"os"
)

type NetWorkAdress struct {
	IP   string
	Port int
}

// интерфецсы для flag.Value
func (n NetWorkAdress) String() string {
	return fmt.Sprintf("%s:%d", n.IP, n.Port)
}

func (n *NetWorkAdress) Set(value string) error {
	items := strings.Split(value, ":")
	n.IP = items[0]
	n.Port = 8080
	if len(items) == 2 {
		val, err := strconv.Atoi(items[1])
		if err != nil {
			return err
		}
		n.Port = val
	}
	return nil
}

// -----------------------------------------------------------
// переменные для программы
// -----------------------------------------------------------
var SendAddress = NetWorkAdress{IP: "", Port: 8080}
var UpdateTime int = 0
var SendTime int = 0

// -----------------------------------------------------------

func strToInt(str string) int {
	val, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return val
}

func GetFlags() {
	flag.Var(&SendAddress, "a", "Net address like 'host:port'")
	update := flag.Int("p", 2, "Update metricks interval")
	send := flag.Int("r", 10, "Send metricks interval")
	flag.Parse()
	UpdateTime = *update
	SendTime = *send

	if address := os.Getenv("ADDRESS"); address != "" {
		SendAddress.Set(address)
	}
	if upd := os.Getenv("REPORT_INTERVAL"); strToInt(upd) > 0 {
		SendTime = strToInt(upd)
	}
	if upd := os.Getenv("POLL_INTERVAL"); strToInt(upd) > 0 {
		UpdateTime = strToInt(upd)
	}
}
