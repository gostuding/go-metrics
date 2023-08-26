package storage

import (
	"context"
	"fmt"
	"strings"
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
	// {"delta":1,"id":"metric name","type":"counter"}
}

func ExampleMemStorage_UpdateJSONSlice() {
	jSlice := `[{"id": "metric name", "type": "counter", "delta": 1}, {"id": "metric name", "type": "gauge", "value": 1}]`
	val, err := memStorage.UpdateJSONSlice(ctx, []byte(jSlice))
	if err != nil {
		fmt.Printf("update storage by JSON slice (%s) error: %v", jSlice, err)
		return
	}
	fmt.Print(strings.Count(string(val), "SUCCESS"))

	// Output:
	// 2
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
	// {"value":1,"id":"metric name","type":"gauge"}
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
