package agent

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
)

// Default values for Config.
const (
	defPort           = 8080
	defPoolInterval   = 2
	defReportInterval = 10
	defRateLimit      = 5
)

// Config contains agent's configuration.
type Config struct {
	PublicKey      *rsa.PublicKey // public key for messages encryption
	IP             string         // server's ip address
	Key            []byte         // key for hashing requests body
	RateLimit      int            // max requests in time
	Port           int            // server's port
	PollInterval   int            // poll requests interval
	ReportInterval int            // send to server interval
	GzipCompress   bool           // flag to compress requests or not
}

// String convert Config to string.
func (n *Config) String() string {
	return fmt.Sprintf("%s:%d -r %d -p %d", n.IP, n.Port, n.PollInterval, n.ReportInterval)
}

// Set validates and sets server's address.
// Use string like ip:port.
func (n *Config) Set(value string) error {
	ip, port, err := net.SplitHostPort(value)
	if err != nil {
		return fmt.Errorf("NetworkAddress ('%s') incorrect. Use value like: 'IP:PORT': %w", value, err)
	}

	n.IP = ip
	val, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("NetworkAddress Port ('%s') convert error: %w. Use integer type", port, err)
	}
	n.Port = val
	return nil
}

// validate is private func.
// Checks Config's arguments.
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

// envToInt is private func.
// Checks enviroment exists the value and convert in to int.
func envToInt(envName string, def int) (int, error) {
	value, ok := os.LookupEnv(envName)
	if !ok {
		return def, nil
	}
	val, err := strconv.Atoi(value)
	if err != nil {
		return def, fmt.Errorf("enviroment value '%s' of '%s' type error: '%w'", value, envName, err)
	}
	return val, nil
}

// envToString is private func.
func envToString(envName string, def string) string {
	value, ok := os.LookupEnv(envName)
	if !ok {
		return def
	}
	return value
}

// parcePublicKey reads rsa public key from file.
func parcePublicKey(filePath string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("file read error: %w", err)
	}
	block, _ := pem.Decode([]byte(data)) //nolint:all //<-senselessly
	if block == nil {
		return nil, errors.New("failed to parse PEM block with publick key")
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parce public key error: %w", err)
	}
	pub, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("key type is not RSA")
	}
	return pub, nil
}

// NewConfig return's configuration object for agent.
// The list of parameters are taken from startup variables and environment variables.
//
// Enviroment values:
//
//	ADDRESS - server address in format ip:port
//	REPORT_INTERVAL - send request interval in seconds
//	POLL_INTERVAL - update metrics interval in seconds
//	RATE_LIMIT - max requests count
func NewConfig() (*Config, error) {
	agentArgs := Config{
		IP:             "",
		Port:           defPort,
		PollInterval:   defPoolInterval,
		ReportInterval: defReportInterval,
		GzipCompress:   true,
		Key:            nil,
		RateLimit:      defRateLimit,
	}
	var key string
	var criptoKeyPath string
	if !flag.Parsed() {
		flag.Var(&agentArgs, "a", "Net address like 'host:port'")
		flag.IntVar(&agentArgs.PollInterval, "p", agentArgs.PollInterval, "Poll metricks interval")
		flag.IntVar(&agentArgs.ReportInterval, "r", agentArgs.ReportInterval, "Report metricks interval")
		flag.IntVar(&agentArgs.RateLimit, "l", agentArgs.RateLimit, "Rate limit")
		flag.BoolVar(&agentArgs.GzipCompress, "gzip", agentArgs.GzipCompress, "Use gzip compress in requests")
		flag.StringVar(&key, "k", "", "Key for SHA256")
		flag.StringVar(&criptoKeyPath, "crypto-key", "", "Key for PUBLIC key for send data to server")
		flag.Parse()
	}

	if address := os.Getenv("ADDRESS"); address != "" {
		err := agentArgs.Set(address)
		if err != nil {
			return &agentArgs, fmt.Errorf("enviroment 'ADDRESS' value error: %w", err)
		}
	}
	var err error
	agentArgs.ReportInterval, err = envToInt("REPORT_INTERVAL", agentArgs.ReportInterval)
	if err != nil {
		return nil, err
	}
	agentArgs.PollInterval, err = envToInt("POLL_INTERVAL", agentArgs.PollInterval)
	if err != nil {
		return nil, err
	}
	agentArgs.RateLimit, err = envToInt("RATE_LIMIT", agentArgs.RateLimit)
	if err != nil {
		return nil, err
	}
	key = envToString("KEY", key)
	if key != "" {
		agentArgs.Key = []byte(key)
	}
	criptoKeyPath = envToString("CRYPTO_KEY", criptoKeyPath)
	if criptoKeyPath != "" {
		agentArgs.PublicKey, err = parcePublicKey(criptoKeyPath)
		if err != nil {
			return nil, err
		}
	}
	return &agentArgs, agentArgs.validate()
}
