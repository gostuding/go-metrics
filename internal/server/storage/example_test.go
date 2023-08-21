package storage

import (
	"context"
	"fmt"
	"log"
)

func init() {
	mem, err := NewMemStorage(restoreStorage, defFileName, saveInterval)
	if err != nil {
		fmt.Printf("Create memory storage error: %v", err)
	}
	sqlStrg, err := NewSQLStorage(dsnString)
	if err != nil {
		fmt.Printf("Create sql storage error: %v", err)
	}
	sqlStorage = sqlStrg
	memStorage = mem
	ctx = context.Background()
}

func ExampleMemStorage_Update() {
	// Update metric with gauge type.
	err := memStorage.Update(ctx, gType, defMetricName, "1.0")
	if err != nil {
		fmt.Printf("update %s error: %v", gType, err)
		return
	}
	// Update metric with counter type.
	err = memStorage.Update(ctx, cType, defMetricName, "1")
	if err != nil {
		fmt.Printf("update %s error: %v", cType, err)
		return
	}

	//Output:
	//
}

func ExampleMemStorage_GetMetric() {
	mValue := "1"
	mName := "counter metric name"
	// Add one counter metric.
	err := memStorage.Update(ctx, counterType, mName, mValue)
	if err != nil {
		fmt.Printf("update %s error: %v", counterType, err)
		return
	}
	// Get added metric value.
	val, err := memStorage.GetMetric(ctx, counterType, mName)
	if err != nil {
		fmt.Printf("get value of %s error: %v", counterType, err)
		return
	}
	fmt.Println(val)

	//Output:
	// 1
}

func ExampleMemStorage_UpdateJSON() {
	jsonConterValue := `{"id": "metric name", "type": "counter", "delta": 1}`
	// Add metrics to storage.
	val, err := memStorage.UpdateJSON(ctx, []byte(jsonConterValue))
	if err != nil {
		fmt.Printf("update storage by JSON (%s) error: %v", jsonConterValue, err)
		return
	}
	fmt.Println(string(val))

	// Output:
	// {"id":"metric name","type":"counter","delta":1}
}

func ExampleMemStorage_UpdateJSONSlice() {
	jSlice := `[{"id": "metric name", "type": "counter", "delta": 1}, {"id": "metric name", "type": "gauge", "value": 1}]`
	val, err := memStorage.UpdateJSONSlice(ctx, []byte(jSlice))
	if err != nil {
		fmt.Printf("update storage by JSON slice (%s) error: %v", jSlice, err)
		return
	}
	fmt.Println(string(val))
}

func ExampleMemStorage_GetMetricJSON() {
	// Add metrics to storage.
	jsonAddValue := `{"id": "metric name", "type": "gauge", "value": 1}`
	_, err := memStorage.UpdateJSON(ctx, []byte(jsonAddValue))
	if err != nil {
		fmt.Printf("update storage by JSON (%s) error: %v", jsonAddValue, err)
		return
	}
	// Get metric from storage
	jsonGetValue := `{"id": "metric name", "type": "gauge"}`
	val, err := memStorage.GetMetricJSON(ctx, []byte(jsonGetValue))
	if err != nil {
		fmt.Printf("get metric by JSON (%s) error: %v", jsonGetValue, err)
		return
	}
	fmt.Println(string(val))

	// Output:
	// {"id":"metric name","type":"gauge","value":1}
}

func ExampleMemStorage_Save() {
	memStorage.SavePath = defFileName
	err := memStorage.Save()
	if err != nil {
		fmt.Printf("save storage error: %v", err)
	} else {
		fmt.Println("Storage save success")
	}

	// Output:
	// Storage save success
}

func ExampleMemStorage_Clear() {
	err := memStorage.Update(ctx, gType, defMetricName, "1.0")
	if err != nil {
		fmt.Printf("add metric %s error: %v", gType, err)
		return
	}
	val, err := memStorage.GetMetric(ctx, gType, defMetricName)
	if err != nil {
		fmt.Printf("get metric %s error: %v", gType, err)
		return
	}
	fmt.Println(val)
	// Clearing storage
	err = memStorage.Clear(ctx)
	if err != nil {
		fmt.Printf("clear storage error: %v", err)
		return
	}
	fmt.Println(len(memStorage.Gauges))

	// Output:
	// 1
	// 0
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
}

func ExampleSQLStorage_GetMetric() {
	if err := sqlStorage.PingDB(ctx); err != nil {
		fmt.Println("SQLStorage ping error")
		return
	}
	mValue := "1"
	mName := "counter metric name"
	// Add one counter metric.
	err := sqlStorage.Update(ctx, counterType, mName, mValue)
	if err != nil {
		fmt.Printf("update %s error: %v", counterType, err)
		return
	}
	// Get added metric value.
	val, err := sqlStorage.GetMetric(ctx, counterType, mName)
	if err != nil {
		fmt.Printf("get value of %s error: %v", counterType, err)
		return
	}
	fmt.Println(val)
}

func ExampleSQLStorage_UpdateJSON() {
	if err := sqlStorage.PingDB(ctx); err != nil {
		log.Fatalf("SQLStorage ping error")
	}
	jsonConterValue := `{"id": "metric name", "type": "counter", "delta": 1}`
	// Add metrics to storage.
	_, err := sqlStorage.UpdateJSON(ctx, []byte(jsonConterValue))
	if err != nil {
		log.Fatalf("update storage by JSON (%s) error: %v", jsonConterValue, err)
	}
}

func ExampleSQLStorage_UpdateJSONSlice() {
	if err := sqlStorage.PingDB(ctx); err != nil {
		log.Fatalf("SQLStorage connection error")
	}
	jSlice := `[{"id": "metric name", "type": "counter", "delta": 1}, {"id": "metric name", "type": "gauge", "value": 1}]`
	_, err := sqlStorage.UpdateJSONSlice(ctx, []byte(jSlice))
	if err != nil {
		log.Fatalf("update storage by JSON slice (%s) error: %v", jSlice, err)
	}
}

func ExampleSQLStorage_GetMetricJSON() {
	// Add the metrics to storage.
	jsonAddValue := `{"id": "metric name", "type": "gauge", "value": 1}`
	_, err := sqlStorage.UpdateJSON(ctx, []byte(jsonAddValue))
	if err != nil {
		log.Fatalf("update storage by JSON (%s) error: %v", jsonAddValue, err)
	}
	// Get the metric from storage
	jsonGetValue := `{"id": "metric name", "type": "gauge"}`
	_, err = sqlStorage.GetMetricJSON(ctx, []byte(jsonGetValue))
	if err != nil {
		log.Fatalf("get metric by JSON (%s) error: %v", jsonGetValue, err)
	}
}

func ExampleSQLStorage_PingDB() {
	// Checks connection to database
	err := sqlStorage.PingDB(ctx)
	if err != nil {
		log.Fatalf("database connection error: %v", err)
	}
}
