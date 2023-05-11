package metrics

import "fmt"

type Storager interface {
	UpdateMetrics()
	SendMetrics(fmt.Stringer)
}
