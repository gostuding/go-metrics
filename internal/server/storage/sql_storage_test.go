//go:build sql_storage
// +build sql_storage

package storage

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkSQLStorage(b *testing.B) {
	ms, err := NewSQLStorage(testsDefDSN)
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
