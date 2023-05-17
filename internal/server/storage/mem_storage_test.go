package storage

import (
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
		wantErr bool
	}{
		{name: "Добавление значения метрики Counter", path: args{"counter", "item", "2"}, wantErr: false},
		{name: "Неправильный путь", path: args{"", "item", "2"}, wantErr: true},
		{name: "Неправильный тип данных", path: args{"gauge", "item", "2ll"}, wantErr: true},
	}
	for _, val := range tests {
		tt := val // переопределили переменную чтобы избежать использования ссылки на переменную цикла (есть такая особенность)
		t.Run(tt.name, func(t *testing.T) {
			ms := MemStorage{}
			err := ms.Update(tt.path.mType, tt.path.mName, tt.path.mValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
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
		name      string
		fields    fields
		args      args
		want      string
		wantError bool
	}{
		{name: "Получение Gauges ", fields: fields{Gauges: gTest(), Counters: cTest()}, args: args{mType: "gauge", mName: "item"}, want: "0.34", wantError: false},
		{name: "Неправильный тип", fields: fields{Gauges: gTest(), Counters: cTest()}, args: args{mType: "error", mName: "item"}, want: "", wantError: true},
		{name: "Неправильное имя", fields: fields{Gauges: gTest(), Counters: cTest()}, args: args{mType: "counter", mName: "none"}, want: "", wantError: true},
	}
	for _, val := range tests {
		tt := val
		t.Run(tt.name, func(t *testing.T) {
			ms := MemStorage{
				Gauges:   tt.fields.Gauges,
				Counters: tt.fields.Counters,
			}
			got, err := ms.GetMetric(tt.args.mType, tt.args.mName)
			if got != tt.want {
				t.Errorf("function GetMetric() got = %v, want %v", got, tt.want)
			}
			if tt.wantError && err == nil {
				t.Errorf("function 'GetMetric()' in test '%s' return's error: %v", tt.name, err)
			} else if !tt.wantError && err != nil {
				t.Errorf("function 'GetMetric()' in test '%s' unexpected error: %v", tt.name, err)
			}
		})
	}
}
