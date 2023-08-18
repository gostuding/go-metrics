package agent

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	IP             string
	Key            []byte
	RateLimit      int
	Port           int
	PollInterval   int
	ReportInterval int
	GzipCompress   bool
}

func (n *Config) String() string {
	return fmt.Sprintf("%s:%d -r %d -p %d", n.IP, n.Port, n.PollInterval, n.ReportInterval)
}

func (n *Config) Set(value string) error {
	items := strings.Split(value, ":")
	if len(items) != 2 {
		return fmt.Errorf("NetworkAddress ('%s') incorrect. Use value like: 'IP:PORT'", value)
	}

	ip, port, err := net.SplitHostPort(value)
	if err != nil {
		return fmt.Errorf("NetworkAddress ('%s') incorrect. Use value like: 'IP:PORT': %w", value, err)
	}

	n.IP = ip
	val, err := strconv.Atoi(port)
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
	if n.RateLimit <= 0 {
		return errors.New("args error: rate limit must be greater then 0")
	}
	return nil
}

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

func GetFlags() (Config, error) {
	agentArgs := Config{
		IP:             "",
		Port:           8080,
		PollInterval:   2,
		ReportInterval: 10,
		GzipCompress:   true,
		Key:            nil,
		RateLimit:      5,
	}
	var key string
	flag.Var(&agentArgs, "a", "Net address like 'host:port'")
	flag.IntVar(&agentArgs.PollInterval, "p", agentArgs.PollInterval, "Poll metricks interval")
	flag.IntVar(&agentArgs.ReportInterval, "r", agentArgs.ReportInterval, "Report metricks interval")
	flag.IntVar(&agentArgs.RateLimit, "l", agentArgs.RateLimit, "Rate limit")
	flag.BoolVar(&agentArgs.GzipCompress, "gzip", agentArgs.GzipCompress, "Use gzip compress in requests")
	flag.StringVar(&key, "k", "", "Key for SHA256")
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
	key = envToString("KEY", key)
	if key != "" {
		agentArgs.Key = []byte(key)
	}
	return agentArgs, agentArgs.validate()
}
