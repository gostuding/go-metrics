package agent

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type NetworkAdress struct {
	IP   string
	Port int
}

// функция для удовлетворения интерфейсу flag.Value
func (n NetworkAdress) String() string {
	return fmt.Sprintf("%s:%d", n.IP, n.Port)
}

// функция для удовлетворения интерфейсу flag.Value
func (n *NetworkAdress) Set(value string) error {
	items := strings.Split(value, ":")
	if len(items) == 2 {
		n.IP = items[0]
		val, err := strconv.Atoi(items[1])
		if err != nil {
			return fmt.Errorf("NetworkAddress Port ('%s') convert error: %s. Use integer type", items[1], err)
		}
		n.Port = val
	} else {
		return fmt.Errorf("NetworkAddress ('%s') incorrect. Use value like: 'IP:PORT'", value)
	}
	return nil
}

// проверка значения переменных окружения на тип int
func strToInt(name string, str string) int {
	val, err := strconv.Atoi(str)
	if err != nil {
		panic(fmt.Sprintf("enviroment value '%s' of '%s' type error: '%s'", str, name, err))
	}
	return val
}

// получение и проверка флагов и переменных окружения
func GetFlags() (NetworkAdress, int, int) {
	sendAddress, updateTime, sendTime := NetworkAdress{IP: "", Port: 8080}, 0, 0
	flag.Var(&sendAddress, "a", "Net address like 'host:port'")
	flag.IntVar(&updateTime, "p", 2, "Poll metricks interval")
	flag.IntVar(&sendTime, "r", 10, "Report metricks interval")
	flag.Parse()

	if address := os.Getenv("ADDRESS"); address != "" {
		err := sendAddress.Set(address)
		if err != nil {
			panic(fmt.Sprintf("enviroment 'ADDRESS' value error: %s", err))
		}
	}
	if upd := os.Getenv("REPORT_INTERVAL"); upd != "" {
		sendTime = strToInt("REPORT_INTERVAL", upd)
	}
	if upd := os.Getenv("POLL_INTERVAL"); upd != "" {
		updateTime = strToInt("POLL_INTERVAL", upd)
	}

	if updateTime <= 0 || sendTime <= 0 {
		panic("POLL_INTERVAL and REPORT_INTERVAL must be greater then 0!")
	}

	return sendAddress, updateTime, sendTime
}
