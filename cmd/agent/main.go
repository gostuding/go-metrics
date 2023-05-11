package main

import ( // "runtime"
	// "runtime/metrics"
	// "time"
	"fmt"
	"time"

	"github.com/gostuding/go-metrics/cmd/agent/metrics"
)

func main() {
	Storage := metrics.NewMetricStorage()
	index := 1

	for {
		time.Sleep(1 * time.Second)
		if index%2 == 0 {
			Storage.UpdateMetrics()
		}
		if index%10 == 0 {
			Storage.SendMetrics()
			fmt.Println("Send")
		}
		index += 1
	}
}
