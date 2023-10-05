package middlewares

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/gostuding/go-metrics/internal/server/middlewares/mocks"
	"go.uber.org/zap"
)

func ExampleNewLogWriter() {
	r := mocks.NewWMock()
	lw := NewLogWriter(r)
	_, err := lw.Write([]byte("data"))
	if err != nil {
		fmt.Printf("write error: %v", err)
		return
	}
	lw.WriteHeader(http.StatusOK)
	fmt.Printf("Status: %d, Size: %d", lw.Status, lw.Size)

	// Output:
	// Status: 200, Size: 4
}

func ExampleNewHashWriter() {
	r := mocks.NewWMock()
	key := []byte("key")
	w := NewHashWriter(r, key)
	_, err := w.Write([]byte("data"))
	if err != nil {
		fmt.Printf("write error: %v", err)
		return
	}
	fmt.Printf("Hash: %s", w.Header().Get(hashVarName))

	// Output:
	// Hash: 5031fe3d989c6d1537a013fa6e739da23463fdaec3b70137d828e36ace221bd0
}

func ExampleNewGzipWriter() {
	data := []byte(strings.Repeat("data", 1000))
	r := mocks.NewWMock()
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("create logger errror: %v", err)
		return
	}
	w := NewGzipWriter(r, logger.Sugar())
	w.Header().Set(contentType, textHTML)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(data)
	if err != nil {
		fmt.Printf("write error: %v", err)
		return
	}
	fmt.Printf("Body length: %d, data legth: %d", len(r.Body), len(data))

	// Output:
	// Body length: 48, data legth: 4000
}

func ExampleNewGzipReader() {
	// Create and fill mock args for example.
	m := mocks.NewWMock()
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("create logger errror: %v", err)
		return
	}
	w := NewGzipWriter(m, logger.Sugar())
	w.Header().Set(contentType, textHTML)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("data"))
	if err != nil {
		fmt.Printf("write data error: %v", err)
		return
	}
	// Creates reader and read gzip data
	r, err := NewGzipReader(m)
	if err != nil {
		fmt.Printf("create reader error: %v", err)
		return
	}
	data, err := io.ReadAll(r)
	if err != nil {
		fmt.Printf("read error: %v", err)
		return
	}
	fmt.Printf("Read length: %d, data: %s", len(data), string(data))

	// Output:
	// Read length: 4, data: data
}

func ExampleGzipMiddleware() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("create logger errror: %v", err)
		return
	}
	GzipMiddleware(logger.Sugar())

	// Output:
	//
}

func ExampleHashCheckMiddleware() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("create logger errror: %v", err)
		return
	}
	HashCheckMiddleware([]byte("key"), logger.Sugar())

	// Output:
	//
}

func ExampleLoggerMiddleware() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("create logger errror: %v", err)
		return
	}
	LoggerMiddleware(logger.Sugar())
	// Output:
	//
}

func ExampleDecriptMiddleware() {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("create key errror: %v", err)
		return
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("create logger errror: %v", err)
		return
	}
	DecriptMiddleware(key, logger.Sugar())

	// Output:
	//
}

func ExampleSubNetCheckMiddleware() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("create logger errror: %v", err)
		return
	}
	_, net, err := net.ParseCIDR("127.0.0.1/24")
	if err != nil {
		fmt.Printf("parce ip error: %v", err)
		return
	}
	SubNetCheckMiddleware(net, logger.Sugar())
	fmt.Printf("subnet string: %s", net.String())

	// Output:
	// subnet string: 127.0.0.0/24
}
