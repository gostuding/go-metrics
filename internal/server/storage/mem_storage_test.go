package storage

import (
	"net/http"
	"testing"
)

func TestMemStorageAddMetric(t *testing.T) {
	type args struct {
		mType  string
		mName  string
		mValue string
	}
	tests := []struct {
		name    string
		path    args
		want    int
		wantErr bool
	}{
		{name: "Добавление значения метрики Counter", path: args{"counter", "item", "2"}, want: http.StatusOK, wantErr: false},
		{name: "Неправильный путь", path: args{"", "item", "2"}, want: http.StatusBadRequest, wantErr: true},
		{name: "Неправильный тип данных", path: args{"gauge", "item", "2ll"}, want: http.StatusBadRequest, wantErr: true},
	}
	for _, val := range tests {
		tt := val // переопределили переменную чтобы избежать использования ссылки на переменную цикла (есть такая особенность)
		t.Run(tt.name, func(t *testing.T) {
			ms := MemStorage{}
			got, err := ms.Update(tt.path.mType, tt.path.mName, tt.path.mValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Update('%s', '%s', '%s') = %v, want %v", tt.path.mType, tt.path.mName, tt.path.mValue, got, tt.want)
			}
		})
	}
}

func TestMemStorageGetMetric(t *testing.T) {
	type fields struct {
		Gauges   map[string]float64
		Counters map[string]int64
	}

	var gTest = func() map[string]float64 {
		v := make(map[string]float64)
		v["item"] = 0.34
		return v
	}

	var cTest = func() map[string]int64 {
		v := make(map[string]int64)
		v["item"] = 2
		return v
	}

	type args struct {
		mType string
		mName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
		want1  int
	}{
		{name: "Получение Gauges ", fields: fields{Gauges: gTest(), Counters: cTest()}, args: args{mType: "gauge", mName: "item"}, want: "0.34", want1: http.StatusOK},
		{name: "Неправильный тип", fields: fields{Gauges: gTest(), Counters: cTest()}, args: args{mType: "error", mName: "item"}, want: "", want1: http.StatusNotFound},
		{name: "Неправильное имя", fields: fields{Gauges: gTest(), Counters: cTest()}, args: args{mType: "counter", mName: "none"}, want: "", want1: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := MemStorage{
				Gauges:   tt.fields.Gauges,
				Counters: tt.fields.Counters,
			}
			got, got1 := ms.GetMetric(tt.args.mType, tt.args.mName)
			if got != tt.want {
				t.Errorf("MemStorage.GetMetric() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("MemStorage.GetMetric() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
