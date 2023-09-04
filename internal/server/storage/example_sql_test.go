//go:build sql_storage
// +build sql_storage

package storage

import (
	"context"
	"fmt"
)

var (
	testsDefDSN = "host=localhost user=postgres database=metrics"
	sqlStorage  *SQLStorage // Storage used in SQLStorage's tests
)

func init() {
	s, err := NewSQLStorage(testsDefDSN)
	if err != nil {
		fmt.Printf("Create sql storage error: %v", err)
	}
	sqlStorage = s
	ctx = context.Background()
}

func ExampleNewSQLStorage() {
	// Create SQL Storage example.
	// Database structure is checking when NewSQLStorage is calling.
	sqlStrg, err := NewSQLStorage(testsDefDSN)
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
	if err := sqlStorage.PingDB(ctx); err != nil {
		fmt.Println("SQLStorage ping error")
		return
	}
	// Update metric with gauge type.
	err := sqlStorage.Update(context.Background(), gType, defMetricName, "1.0")
	if err != nil {
		fmt.Printf("update %s error: %v", gType, err)
		return
	}
	// Update metric with counter type.
	err = sqlStorage.Update(context.Background(), cType, defMetricName, "1")
	if err != nil {
		fmt.Printf("update %s error: %v", cType, err)
		return
	}
	fmt.Println("update success")

	// Output:
	// update success
}

func ExampleSQLStorage_GetMetric() {
	mValue := "1.1"
	mName := "metric name 1"
	err := sqlStorage.Clear(ctx)
	if err != nil {
		fmt.Printf("storage clear error: %v", err)
		return
	}
	// Add one counter metric.
	err = sqlStorage.Update(ctx, gaugeType, mName, mValue)
	if err != nil {
		fmt.Printf("update %s error: %v", gaugeType, err)
		return
	}
	// Get added metric value.
	val, err := sqlStorage.GetMetric(ctx, gaugeType, mName)
	if err != nil {
		fmt.Printf("get value of %s error: %v", gaugeType, err)
		return
	}
	fmt.Println(val)

	// Output:
	// 1.100000
}

func ExampleSQLStorage_UpdateJSON() {
	if err := sqlStorage.Clear(ctx); err != nil {
		fmt.Printf("storage clear error: %v", err)
		return
	}
	jsonConterValue := `{"id": "metric update json name", "type": "gauge", "value": 1}`
	// Add metrics to storage.
	val, err := sqlStorage.UpdateJSON(ctx, []byte(jsonConterValue))
	if err != nil {
		fmt.Printf("update storage by JSON (%s) error: %v", jsonConterValue, err)
	} else {
		fmt.Println(string(val))
	}

	// Output:
	// {"value":1,"id":"metric update json name","type":"gauge"}
}

func ExampleSQLStorage_UpdateJSONSlice() {
	jSlice := `[{"id": "metric name", "type": "counter", "delta": 1}, {"id": "metric name", "type": "gauge", "value": 1}]`
	_, err := sqlStorage.UpdateJSONSlice(ctx, []byte(jSlice))
	if err != nil {
		fmt.Printf("update storage by JSON slice (%s) error: %v", jSlice, err)
	} else {
		fmt.Println("Update json slice success")
	}

	// Output:
	// Update json slice success
}

func ExampleSQLStorage_GetMetricJSON() {
	// Add the metrics to storage.
	jsonAddValue := `{"id": "metric name", "type": "gauge", "value": 1}`
	_, err := sqlStorage.UpdateJSON(ctx, []byte(jsonAddValue))
	if err != nil {
		fmt.Printf("update storage by JSON (%s) error: %v", jsonAddValue, err)
		return
	}
	// Get the metric from storage
	jsonGetValue := `{"id": "metric name", "type": "gauge"}`
	val, err := sqlStorage.GetMetricJSON(ctx, []byte(jsonGetValue))
	if err != nil {
		fmt.Printf("get metric by JSON (%s) error: %v", jsonGetValue, err)
	} else {
		fmt.Println(string(val))
	}

	// Output:
	// {"value":1,"id":"metric name","type":"gauge"}
}

func ExampleSQLStorage_PingDB() {
	// Checks connection to database
	err := sqlStorage.PingDB(ctx)
	if err != nil {
		fmt.Printf("database connection error: %v", err)
	} else {
		fmt.Println("ping success")
	}

	// Output:
	// ping success
}
