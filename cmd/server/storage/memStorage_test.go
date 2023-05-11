package storage

import (
	"net/http"
	"testing"
)

func TestMemStorage_AddMetric(t *testing.T) {
	type fields struct {
		Gauges   map[string]float64
		Counters map[string]int64
	}
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		path    string
		want    int
		wantErr bool
	}{
		{name: "Добавление значения метрики Counter", path: "/update/counter/item/1", want: http.StatusOK, wantErr: false},
		{name: "Неправильный путь", path: "/update/item/1", want: http.StatusNotFound, wantErr: true},
		{name: "Неправильный тип данных", path: "/update/guage/1/ddd", want: http.StatusBadRequest, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				Gauges:   tt.fields.Gauges,
				Counters: tt.fields.Counters,
			}
			got, err := ms.AddMetric(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemStorage.AddMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MemStorage.AddMetric() = %v, want %v", got, tt.want)
			}
		})
	}
}
