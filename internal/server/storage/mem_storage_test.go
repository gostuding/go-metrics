package storage

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
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
			ms, err := NewMemStorage(false, "", 300)
			assert.NoError(t, err, "error making new memStorage")
			err = ms.Update(context.Background(), tt.path.mType, tt.path.mName, tt.path.mValue)
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
		v["item"] = float64(0.34)
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
			ms, err := NewMemStorage(false, "", 300)
			assert.NoError(t, err, "error making new memStorage")
			ms.Counters = tt.fields.Counters
			ms.Gauges = tt.fields.Gauges
			got, err := ms.GetMetric(context.Background(), tt.args.mType, tt.args.mName)
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

func TestMemStorage_GetMetricJSON(t *testing.T) {
	type fields struct {
		Gauges   map[string]float64
		Counters map[string]int64
	}
	fieldsTest := fields{Gauges: map[string]float64{"RandomValue": 0.222}, Counters: map[string]int64{"PollCount": 1}}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "Получение значения",
			fields:  fieldsTest,
			args:    args{[]byte(`{"id":"PollCount","type":"counter"}`)},
			want:    []byte(`{"id":"PollCount","type":"counter","delta":1}`),
			wantErr: false,
		},
		{
			name:    "Неправильный тип значения",
			fields:  fieldsTest,
			args:    args{[]byte(`{"id":"PollCount","type":"counter1"}`)},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Неправильное имя значения",
			fields:  fieldsTest,
			args:    args{[]byte(`{"id":"PollCount1","type":"counter"}`)},
			want:    []byte(""),
			wantErr: true,
		},
		{
			name:    "Неправильная сериализация json",
			fields:  fieldsTest,
			args:    args{[]byte(`{"id": PollCount1","type":"counter"}`)},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ms, err := NewMemStorage(false, "", 300)
			assert.NoError(t, err, "error making new memStorage")
			ms.Counters = tt.fields.Counters
			ms.Gauges = tt.fields.Gauges
			got, err := ms.GetMetricJSON(context.Background(), tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("memStorage.GetMetricJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("memStorage.GetMetricJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemStorage_UpdateJSON(t *testing.T) {
	type fields struct {
		Gauges   map[string]float64
		Counters map[string]int64
	}
	fieldsTest := fields{
		Gauges:   map[string]float64{"RandomValue": 0.222},
		Counters: map[string]int64{"PollCount": 1},
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "Обновление значения counter",
			fields:  fieldsTest,
			args:    args{[]byte(`{"id":"PollCount","type":"counter","delta":1}`)},
			want:    []byte(`{"id":"PollCount","type":"counter","delta":2}`),
			wantErr: false,
		},
		{
			name:    "Значение delta для counter не определено",
			fields:  fieldsTest,
			args:    args{[]byte(`{"id":"PollCount","type":"counter","value":1}`)},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Значение value для gauge не определено",
			fields:  fieldsTest,
			args:    args{[]byte(`{"id":"PollCount","type":"gauge","delta":1}`)},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Тип не определён",
			fields:  fieldsTest,
			args:    args{[]byte(`{"id":"PollCount","delta":1,"value":0.222}`)},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Неправильный формат json",
			fields:  fieldsTest,
			args:    args{[]byte(`{"id":"PollCount" "delta":1,"value":0.222}`)},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ms, err := NewMemStorage(false, "", 300)
			assert.NoError(t, err, "error making new memStorage")
			ms.Counters = tt.fields.Counters
			ms.Gauges = tt.fields.Gauges
			got, err := ms.UpdateJSON(context.Background(), tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("memStorage.UpdateJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("memStorage.UpdateJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
