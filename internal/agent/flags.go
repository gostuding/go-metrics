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
	IP                  string
	Port                int
	PollInterval        int
	ReportInterval      int
	ReportSliceInterval int
	GzipCompress        bool
	Key                 string
	RateLimit           int
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
	if n.ReportSliceInterval <= 0 {
		return errors.New("args error: report metric by slice must be greater then 0")
	}
	if n.RateLimit <= 0 {
		return errors.New("args error: rate limit must be greater then 0")
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

func envToInt(envName string, def int) (int, error) {
	if value := os.Getenv(envName); value != "" {
		send, err := strToInt(envName, value)
		if err != nil {
			return def, err
		}
		def = send
	}
	return def, nil
}

func envToString(envName string, def string) string {
	if value := os.Getenv(envName); value != "" {
		return value
	}
	return def
}

// получение и проверка флагов и переменных окружения
func GetFlags() (Config, error) {
	agentArgs := Config{"", 8080, 2, 10, 3, false, "", 5}
	flag.Var(&agentArgs, "a", "Net address like 'host:port'")
	flag.IntVar(&agentArgs.PollInterval, "p", 2, "Poll metricks interval")
	flag.IntVar(&agentArgs.ReportInterval, "r", 10, "Report metricks interval")
	flag.IntVar(&agentArgs.ReportSliceInterval, "rs", 25, "Report metricks by slice interval")
	flag.IntVar(&agentArgs.RateLimit, "l", 5, "Rate limit")
	flag.BoolVar(&agentArgs.GzipCompress, "gzip", true, "Use gzip compress in requests")
	flag.StringVar(&agentArgs.Key, "k", "", "Key for SHA256")
	flag.Parse()

	if address := os.Getenv("ADDRESS"); address != "" {
		err := agentArgs.Set(address)
		if err != nil {
			return agentArgs, fmt.Errorf("enviroment 'ADDRESS' value error: %w", err)
		}
	}

	var err error
	agentArgs.ReportInterval, err = envToInt("REPORT_INTERVAL", agentArgs.ReportInterval)
	if err != nil {
		return agentArgs, err
	}
	agentArgs.PollInterval, err = envToInt("POLL_INTERVAL", agentArgs.PollInterval)
	if err != nil {
		return agentArgs, err
	}
	agentArgs.RateLimit, err = envToInt("RATE_LIMIT", agentArgs.RateLimit)
	if err != nil {
		return agentArgs, err
	}
	agentArgs.Key = envToString("KEY", agentArgs.Key)
	return agentArgs, agentArgs.validate()
}
