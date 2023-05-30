package agent

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	IP             string
	Port           int
	PollInterval   int
	ReportInterval int
	GzipCompress   bool
}

// функция для удовлетворения интерфейсу flag.Value
func (n *Config) String() string {
	return fmt.Sprintf("%s:%d -r %d -p %d", n.IP, n.Port, n.PollInterval, n.ReportInterval)
}

// функция для удовлетворения интерфейсу flag.Value
func (n *Config) Set(value string) error {
	items := strings.Split(value, ":")
	if len(items) != 2 {
		return fmt.Errorf("NetworkAddress ('%s') incorrect. Use value like: 'IP:PORT'", value)
	}
	n.IP = items[0]
	val, err := strconv.Atoi(items[1])
	if err != nil {
		return fmt.Errorf("NetworkAddress Port ('%s') convert error: %w. Use integer type", items[1], err)
	}
	n.Port = val
	return nil
}

func (n *Config) validate() error {
	if n.Port <= 1 {
		return errors.New("args error: Port must be greater then 0")
	}
	if n.ReportInterval <= 0 {
		return errors.New("args error: REPORT_INTERVAL must be greater then 0")
	}
	if n.PollInterval <= 0 {
		return errors.New("args error: POLL_INTERVAL must be greater then 0")
	}
	return nil
}

// проверка значения переменных окружения на тип int
func strToInt(name string, str string) (int, error) {
	val, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("enviroment value '%s' of '%s' type error: '%w'", str, name, err)
	}
	return val, nil
}

// получение и проверка флагов и переменных окружения
func GetFlags() (Config, error) {
	agentArgs := Config{"", 8080, 2, 10, false}
	flag.Var(&agentArgs, "a", "Net address like 'host:port'")
	flag.IntVar(&agentArgs.PollInterval, "p", 2, "Poll metricks interval")
	flag.IntVar(&agentArgs.ReportInterval, "r", 10, "Report metricks interval")
	flag.BoolVar(&agentArgs.GzipCompress, "gzip", false, "Use gzip compress in requests")
	flag.Parse()

	if address := os.Getenv("ADDRESS"); address != "" {
		err := agentArgs.Set(address)
		if err != nil {
			return agentArgs, fmt.Errorf("enviroment 'ADDRESS' value error: %w", err)
		}
	}
	if report := os.Getenv("REPORT_INTERVAL"); report != "" {
		send, err := strToInt("REPORT_INTERVAL", report)
		if err != nil {
			return agentArgs, err
		}
		agentArgs.ReportInterval = send
	}
	if poll := os.Getenv("POLL_INTERVAL"); poll != "" {
		update, err := strToInt("POLL_INTERVAL", poll)
		if err != nil {
			return agentArgs, err
		}
		agentArgs.PollInterval = update
	}
	return agentArgs, agentArgs.validate()
}
