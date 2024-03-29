package agent

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
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
	defPort           = 8080      // default server port
	defPoolInterval   = 2         // default update metrics interval
	defReportInterval = 10        // default send to server interval
	defRateLimit      = 5         // default max gorutines to send messages
	defaultKey        = "default" // Key for hash
	falseStr          = "false"   // internal value
)

// Config contains agent's configuration.
type (
	Config struct {
		PublicKey      *rsa.PublicKey `json:"-"`                         // public key for messages encryption
		PublicKeyPath  string         `json:"crypto_key,omitempty"`      // path to public key
		IP             string         `json:"address,omitempty"`         // server's ip address
		LocalAddress   *net.IP        `json:"-"`                         // agent's local ip address
		gzipCompress   string         `json:"-"`                         //
		HashKey        string         `json:"key,omitempty"`             // key for hashing requests body
		RateLimit      int            `json:"rate_limit,omitempty"`      // max requests in time
		Port           int            `json:"-"`                         // server's port
		PollInterval   int            `json:"poll_interval,omitempty"`   // poll requests interval
		ReportInterval int            `json:"report_interval,omitempty"` // send to server interval
		GzipCompress   bool           `json:"gzip,omitempty"`            // flag to compress requests or not
		SendByRPC      bool           `json:"-"`                         // flag for RPC send using
	}
)

// String convert Config to string.
func (n *Config) String() string {
	return fmt.Sprintf("%s:%d -r %d -p %d", n.IP, n.Port, n.PollInterval, n.ReportInterval)
}

func (n *Config) setDefault() {
	if n.Port == 0 {
		n.Port = defPort
	}
	if n.PollInterval == 0 {
		n.PollInterval = defPoolInterval
	}
	if n.ReportInterval == 0 {
		n.ReportInterval = defReportInterval
	}
	if n.RateLimit == 0 {
		n.RateLimit = defRateLimit
	}
	if n.gzipCompress != falseStr {
		n.GzipCompress = true
	}
	if n.HashKey == "" {
		n.HashKey = defaultKey
	}
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
	if value, ok := os.LookupEnv(envName); !ok {
		return def
	} else {
		return value
	}
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

// lookFileConfig gets config values from file with path.
func lookFileConfig(path string, a *Config) error {
	defer a.setDefault()
	if val, ok := os.LookupEnv("CONFIG"); ok {
		path = val
	}
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("config file read error: %w", err)
	}
	var c Config
	err = json.Unmarshal(data, &c)
	if err != nil {
		return fmt.Errorf("config file convert error: %w", err)
	}
	if a.IP == "" {
		err = a.Set(c.IP)
		if err != nil {
			return fmt.Errorf("config file IP error: %w", err)
		}
	}
	if a.PollInterval == 0 {
		a.PollInterval = c.PollInterval
	}
	if a.ReportInterval == 0 {
		a.ReportInterval = c.ReportInterval
	}
	if a.RateLimit == 0 {
		a.RateLimit = c.RateLimit
	}
	if a.gzipCompress == "" {
		a.GzipCompress = c.GzipCompress
	}
	if a.HashKey == "" {
		a.HashKey = c.HashKey
	}
	if a.PublicKeyPath == "" {
		a.PublicKeyPath = c.PublicKeyPath
	}
	return nil
}

// lookEnviroment gets config values from Enviroment.
func lookEnviroment(a *Config) error {
	if address, ok := os.LookupEnv("ADDRESS"); ok {
		if err := a.Set(address); err != nil {
			return fmt.Errorf("enviroment 'ADDRESS' value error: %w", err)
		}
	}
	var err error
	a.ReportInterval, err = envToInt("REPORT_INTERVAL", a.ReportInterval)
	if err != nil {
		return err
	}
	a.PollInterval, err = envToInt("POLL_INTERVAL", a.PollInterval)
	if err != nil {
		return err
	}
	a.RateLimit, err = envToInt("RATE_LIMIT", a.RateLimit)
	if err != nil {
		return err
	}
	a.HashKey = envToString("KEY", a.HashKey)
	pKey := envToString("CRYPTO_KEY", a.PublicKeyPath)
	if pKey != "" {
		a.PublicKey, err = parcePublicKey(pKey)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetLocalIP is internal function.
func getLocalIP() (*net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, fmt.Errorf("get local address error: %w", err)
	}
	defer conn.Close() //nolint:errcheck //<-senselessly
	localAddress, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return nil, errors.New("can't get local address")
	}
	return &localAddress.IP, nil
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
	agentArgs := Config{}
	l, err := getLocalIP()
	if err != nil {
		return nil, err
	}
	agentArgs.LocalAddress = l
	cfgPath := ""
	if !flag.Parsed() {
		flag.Var(&agentArgs, "a", "Net address like 'host:port'")
		flag.IntVar(&agentArgs.PollInterval, "p", agentArgs.PollInterval, "Poll metricks interval")
		flag.IntVar(&agentArgs.ReportInterval, "r", agentArgs.ReportInterval, "Report metricks interval")
		flag.IntVar(&agentArgs.RateLimit, "l", agentArgs.RateLimit, "Rate limit")
		flag.StringVar(&agentArgs.gzipCompress, "gzip", agentArgs.gzipCompress, "Use gzip compress in requests")
		flag.StringVar(&agentArgs.HashKey, "k", "", "Key for HASHSUMM in SHA256")
		flag.StringVar(&agentArgs.PublicKeyPath, "crypto-key", "", "Path to PUBLIC key file")
		flag.StringVar(&cfgPath, "c", "", "Path to config file")
		flag.StringVar(&cfgPath, "config", cfgPath, "Path to config file (the same as -c)")
		flag.BoolVar(&agentArgs.SendByRPC, "rpc", agentArgs.SendByRPC,
			"Use RPC for send data to server. Sets only by this arg")
		flag.Parse()
	}
	if err := lookFileConfig(cfgPath, &agentArgs); err != nil {
		return nil, err
	}
	if err := lookEnviroment(&agentArgs); err != nil {
		return nil, err
	}
	return &agentArgs, agentArgs.validate()
}
