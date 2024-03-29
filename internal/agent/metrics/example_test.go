package metrics

import (
	"fmt"
	"log"
	"net"

	"go.uber.org/zap"
)

var (
	ip        = "localhost"
	port      = 8080
	rateLimit = 1
	key       = []byte("default")
	compress  = false
)

func Example() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalln("create logger error:", err)
	}
	localAddress := net.IP("127.0.0.1")
	storage := NewMemoryStorage(nil, logger, ip, key, port, compress,
		rateLimit, &localAddress, false)
	// Collect metrics.
	storage.UpdateMetrics()
	storage.UpdateAditionalMetrics()
	// All metrics are in MetricsSlice.
	// Count UpdateMetrics repeats are in PollCount.
	fmt.Printf("Update metrics count: %d", *storage.MetricsSlice["PollCount"].Delta)
	// Check that PollCount == 1
	if *storage.MetricsSlice["PollCount"].Delta == 1 {
		// If storage data will be send success, PollCount will be cheched to 0
		// If encription nead, set KEY instead nil.
		storage.SendMetricsSlice()
	}

	// Output:
	// Update metrics count: 1
}
