//go:build sql_storage
// +build sql_storage

package storage

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	dbDSN = ""
)

func init() {
	for _, item := range os.Args {
		if strings.HasPrefix(item, "dsn=") {
			dbDSN = strings.Replace(item, "dsn=", "", 1)
		}
	}
}

func ExampleNewSQLStorage() {
	// Create SQL Storage example.
	// Database structure is checking when NewSQLStorage is calling.
	sqlStrg, err := NewSQLStorage(dbDSN)
	if err != nil {
		fmt.Printf("Create sql storage error: %v", err)
		return
	}
	// Do any actions with storage...
	// ...
	// ...
	// Stop storage before work finish.
	err = sqlStrg.Stop()
	if err != nil {
		fmt.Printf("Stop sql storage error: %v", err)
	} else {
		fmt.Println("SQL storage stop success")
	}
	// Output:
	// SQL storage stop success
}

func ExampleSQLStorage_Update() {
	sqlStrg, err := NewSQLStorage(dbDSN)
	if err != nil {
		fmt.Printf("Create sql storage error: %v", err)
		return
	}
	// Update metric with gauge type.
	err = sqlStrg.Update(context.Background(), gType, defMetricName, "1.0")
	if err != nil {
		fmt.Printf("update %s error: %v", gType, err)
		return
	}
	// Update metric with counter type.
	err = sqlStrg.Update(context.Background(), cType, defMetricName, "1")
	if err != nil {
		fmt.Printf("update %s error: %v", cType, err)
		return
	}
	fmt.Println("update success")

	// Output:
	// update success
}

func ExampleSQLStorage_GetMetric() {
	sqlStrg, err := NewSQLStorage(dbDSN)
	if err != nil {
		fmt.Printf("Create sql storage error: %v", err)
		return
	}
	mValue := "1.1"
	mName := "metric name 1"
	err = sqlStrg.Clear(ctx)
	if err != nil {
		fmt.Printf("storage clear error: %v", err)
		return
	}
	// Add one counter metric.
	err = sqlStrg.Update(ctx, gaugeType, mName, mValue)
	if err != nil {
		fmt.Printf("update %s error: %v", gaugeType, err)
		return
	}
	// Get added metric value.
	val, err := sqlStrg.GetMetric(ctx, gaugeType, mName)
	if err != nil {
		fmt.Printf("get value of %s error: %v", gaugeType, err)
		return
	}
	fmt.Println(val)

	// Output:
	// 1.100000
}

func ExampleSQLStorage_UpdateJSON() {
	sqlStrg, err := NewSQLStorage(dbDSN)
	if err != nil {
		fmt.Printf("Create sql storage error: %v", err)
		return
	}
	if err := sqlStrg.Clear(ctx); err != nil {
		fmt.Printf("storage clear error: %v", err)
		return
	}
	jsonConterValue := `{"id": "metric update json name", "type": "gauge", "value": 1}`
	// Add metrics to storage.
	val, err := sqlStrg.UpdateJSON(ctx, []byte(jsonConterValue))
	if err != nil {
		fmt.Printf("update storage by JSON (%s) error: %v", jsonConterValue, err)
	} else {
		fmt.Println(string(val))
	}

	// Output:
	// {"value":1,"id":"metric update json name","type":"gauge"}
}

func ExampleSQLStorage_UpdateJSONSlice() {
	sqlStrg, err := NewSQLStorage(dbDSN)
	if err != nil {
		fmt.Printf("Create sql storage error: %v", err)
		return
	}
	jSlice := `[{"id": "metric name", "type": "counter", "delta": 1}, {"id": "metric name", "type": "gauge", "value": 1}]`
	_, err = sqlStrg.UpdateJSONSlice(ctx, []byte(jSlice))
	if err != nil {
		fmt.Printf("update storage by JSON slice (%s) error: %v", jSlice, err)
	} else {
		fmt.Println("Update json slice success")
	}

	// Output:
	// Update json slice success
}

func ExampleSQLStorage_GetMetricJSON() {
	sqlStrg, err := NewSQLStorage(dbDSN)
	if err != nil {
		fmt.Printf("Create sql storage error: %v", err)
		return
	}
	// Add the metrics to storage.
	jsonAddValue := `{"id": "metric name", "type": "gauge", "value": 1}`
	_, err = sqlStrg.UpdateJSON(ctx, []byte(jsonAddValue))
	if err != nil {
		fmt.Printf("update storage by JSON (%s) error: %v", jsonAddValue, err)
		return
	}
	// Get the metric from storage
	jsonGetValue := `{"id": "metric name", "type": "gauge"}`
	val, err := sqlStrg.GetMetricJSON(ctx, []byte(jsonGetValue))
	if err != nil {
		fmt.Printf("get metric by JSON (%s) error: %v", jsonGetValue, err)
	} else {
		fmt.Println(string(val))
	}

	// Output:
	// {"value":1,"id":"metric name","type":"gauge"}
}

func ExampleSQLStorage_PingDB() {
	sqlStrg, err := NewSQLStorage(dbDSN)
	if err != nil {
		fmt.Printf("Create sql storage error: %v", err)
		return
	}
	// Checks connection to database
	err = sqlStrg.PingDB(ctx)
	if err != nil {
		fmt.Printf("database connection error: %v", err)
	} else {
		fmt.Println("ping success")
	}

	// Output:
	// ping success
}

func BenchmarkSQLStorage(b *testing.B) {
	ms, err := NewSQLStorage(dbDSN)
	if !assert.NoError(b, err, "create sql storage error") {
		return
	}
	val := int64(1)
	m := metric{ID: "test", MType: counterType, Delta: &val}
	mString := `{"id": "test", "type": "counter", "value": 1}`
	mStringSlice := strings.Repeat(mString, 2) + `{"id": "test", "type": "gauge", "value": 1.0}`

	b.Run("add metric", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ms.Update(ctx, m.MType, m.ID, "1") //nolint:all //<-senselessly
		}
	})

	b.Run("update one metric", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ms.updateOneMetric(ctx, m, ms.con) //nolint:all //<-senselessly
		}
	})

	b.ResetTimer()
	b.Run("get metric", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ms.GetMetric(ctx, m.MType, m.ID) //nolint:all //<-senselessly
		}
	})

	b.Run("get metric json", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ms.GetMetricJSON(ctx, []byte(mString)) //nolint:all //<-senselessly
		}
	})

	b.Run("update metric json", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ms.UpdateJSON(ctx, []byte(mString)) //nolint:all //<-senselessly
		}
	})

	b.Run("update metric json slice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ms.UpdateJSONSlice(ctx, []byte(mStringSlice)) //nolint:all //<-senselessly
		}
	})

	b.Run("get metrics HTML", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ms.GetMetricsHTML(ctx) //nolint:all //<-senselessly
		}
	})
}
