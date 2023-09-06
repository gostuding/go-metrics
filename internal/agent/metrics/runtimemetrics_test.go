package metrics

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.uber.org/zap"
)

func Test_makeMetric(t *testing.T) {
	floatValue := float64(1)
	intValue := int64(1)
	type args struct {
		id    string
		value any
	}
	tests := []struct {
		name    string
		args    args
		want    *metrics
		wantErr bool
	}{
		{
			name:    "Test gauge",
			args:    args{id: "gauge value", value: floatValue},
			want:    &metrics{ID: "gauge value", MType: "gauge", Value: &floatValue},
			wantErr: false,
		},
		{
			name:    "Test counter",
			args:    args{id: "counter value", value: intValue},
			want:    &metrics{ID: "counter value", MType: "counter", Delta: &intValue},
			wantErr: false,
		},
		{
			name:    "Test error",
			args:    args{id: "counter value", value: "123"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeMetric(tt.args.id, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeMetric() = '%v:%v', want '%v:%v'", got.ID, got.MType, tt.want.ID, tt.want.MType)
			}
		})
	}
}

func Test_metricsStorage_addMetric(t *testing.T) {
	gaugeValue := float64(10)
	counterValue := int64(10)
	ms := NewMemoryStorage(&zap.Logger{}, "", []byte(""), 0, false, 1)
	type args struct {
		name  string
		value any
	}
	tests := []struct {
		name string
		args args
		want metrics
	}{
		{
			name: "Add gauge",
			args: args{name: "gauge value", value: gaugeValue},
			want: metrics{ID: "gauge value", MType: "gauge", Value: &gaugeValue},
		},
		{
			name: "Add counter",
			args: args{name: "counter value", value: counterValue},
			want: metrics{ID: "counter value", MType: "counter", Delta: &counterValue},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ms.addMetric(tt.args.name, tt.args.value)
			if !reflect.DeepEqual(ms.MetricsSlice[tt.args.name], tt.want) {
				t.Errorf("addMetric() value error. want: %v, got %v", tt.want, ms.MetricsSlice[tt.args.name])
				return
			}
		})
	}
}

func Test_metricsStorage_UpdateMetrics(t *testing.T) {
	ms := NewMemoryStorage(&zap.Logger{}, "", []byte(""), 0, false, 1)
	ms.UpdateMetrics()
	pollCount := ms.MetricsSlice["PollCount"].Delta
	ms.UpdateMetrics()
	t.Run("pollCountChange", func(t *testing.T) {
		if pollCount == ms.MetricsSlice["PollCount"].Delta {
			t.Error("UpdateMetrics pollCount change error")
		}
	})
	t.Run("additionalMetricsChange", func(t *testing.T) {
		ms.UpdateAditionalMetrics()
		if ms.MetricsSlice["TotalMemory"].Value == nil {
			t.Error("UpdateAditionalMetrics change error")
		}
	})
}

func Test_makeMap(t *testing.T) {
	var r runtime.MemStats
	runtime.ReadMemStats(&r)
	var p int64 = 10
	mName := "PollCount"
	tests := []struct {
		name      string
		pollCount *int64
		want      int64
	}{
		{
			name:      "Make map with nil pollCount",
			pollCount: nil,
			want:      1,
		},
		{
			name:      "Make map with pollCount = 10",
			pollCount: &p,
			want:      11,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			m := makeMap(&r, tt.pollCount)
			got, err := strconv.ParseInt(fmt.Sprint(m[mName]), 10, 64)
			assert.NoError(t, err, "value convert error")
			if got != tt.want {
				t.Errorf("makeMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
