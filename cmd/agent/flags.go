package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
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
func GetFlags() {
	flag.Var(&SendAddress, "a", "Net address like 'host:port'")
	update := flag.Int("p", 2, "Update metricks interval")
	send := flag.Int("r", 10, "Send metricks interval")
	flag.Parse()
	UpdateTime = *update
	SendTime = *send
}
