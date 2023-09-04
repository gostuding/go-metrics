package storage

import (
	"reflect"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
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
		{name: "Добавление значения метрики Counter", path: args{counterType, "item", "2"}, wantErr: false},
		{name: "Неправильный путь", path: args{"", "item", "2"}, wantErr: true},
		{name: "Неправильный тип данных", path: args{gaugeType, "item", "2ll"}, wantErr: true},
	}
	for _, val := range tests {
		tt := val
		t.Run(tt.name, func(t *testing.T) {
			ms, err := NewMemStorage(restoreStorage, defFileName, saveInterval)
			assert.NoError(t, err, "error making new MemStorage")
			err = ms.Update(ctx, tt.path.mType, tt.path.mName, tt.path.mValue)
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
		{
			name: "Получение Gauges ",
			fields: fields{
				Gauges:   gTest(),
				Counters: cTest()},
			args: args{
				mType: gaugeType,
				mName: "item",
			},
			want:      "0.34",
			wantError: false,
		},
		{
			name: "Неправильный тип",
			fields: fields{
				Gauges:   gTest(),
				Counters: cTest()},
			args: args{
				mType: "error",
				mName: "item",
			},
			want:      "",
			wantError: true,
		},
		{
			name: "Неправильное имя",
			fields: fields{
				Gauges:   gTest(),
				Counters: cTest(),
			},
			args: args{
				mType: counterType,
				mName: "none",
			},
			want:      "",
			wantError: true,
		},
	}
	for _, val := range tests {
		tt := val
		t.Run(tt.name, func(t *testing.T) {
			ms, err := NewMemStorage(restoreStorage, defFileName, saveInterval)
			assert.NoError(t, err, "error making new MemStorage")
			ms.Counters = tt.fields.Counters
			ms.Gauges = tt.fields.Gauges
			got, err := ms.GetMetric(ctx, tt.args.mType, tt.args.mName)
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
			want:    []byte(`{"delta":1,"id":"PollCount","type":"counter"}`),
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
			ms, err := NewMemStorage(restoreStorage, defFileName, saveInterval)
			assert.NoError(t, err, "error making new MemStorage")
			ms.Counters = tt.fields.Counters
			ms.Gauges = tt.fields.Gauges
			got, err := ms.GetMetricJSON(ctx, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemStorage.GetMetricJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemStorage.GetMetricJSON() = %v, want %v", got, tt.want)
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
			want:    []byte(`{"delta":2,"id":"PollCount","type":"counter"}`),
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
			assert.NoError(t, err, "error making new MemStorage")
			ms.Counters = tt.fields.Counters
			ms.Gauges = tt.fields.Gauges
			got, err := ms.UpdateJSON(ctx, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemStorage.UpdateJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemStorage.UpdateJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemStorage_UpdateJSONSlice(t *testing.T) {
	ms, err := NewMemStorage(restoreStorage, defFileName, saveInterval)
	assert.NoError(t, err, "create mem storage error")
	tests := []struct {
		name    string
		args    []byte
		want    string
		wantErr bool
	}{
		{
			name:    "добваление списка метрик",
			args:    []byte(`[{"id": "1", "type": "gauge", "value": 1}, {"id": "2", "type": "counter", "delta": 1}]`),
			want:    `1. '1' update SUCCESS2. '2' update SUCCESS`,
			wantErr: false,
		},
		{
			name:    "ошибка добваления списка метрик",
			args:    []byte("error"),
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := ms.UpdateJSONSlice(ctx, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemStorage.UpdateJSONSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			res := strings.ReplaceAll(string(got), " \n", "")
			if tt.want != res {
				t.Errorf("MemStorage.UpdateJSONSlice() want = '%s', got '%s'", tt.want, res)
			}
		})
	}
}

func BenchmarkMemStorage(b *testing.B) {
	ms, err := NewMemStorage(restoreStorage, defFileName, saveInterval)
	assert.NoError(b, err, "error making new MemStorage")
	err = ms.Update(ctx, counterType, "test", "0")
	assert.NoError(b, err, "add initial metric error")
	val := int64(1)
	m := metric{ID: "test", MType: counterType, Delta: &val}
	mString := `{"id": "test", "type": "counter", "value": 1}`
	mStringSlice := strings.Repeat(mString, 2) + `{"id": "test", "type": "gauge", "value": 1.0}`

	b.ResetTimer()
	b.Run("add metric", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ms.Update(ctx, m.MType, m.ID, "1") //nolint:all //<-senselessly
		}
	})

	b.Run("update one metric", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ms.updateOneMetric(m) //nolint:all //<-senselessly
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

func Test_getSortedKeysFloat(t *testing.T) {
	args := make(map[string]float64)
	args["2"] = 1
	args["1"] = 2
	var want []string
	want = append(want, "1", "2")
	t.Run("sort test", func(t *testing.T) {
		if got := getSortedKeysFloat(args); !reflect.DeepEqual(got, want) {
			t.Errorf("getSortedKeysFloat() = %v, want %v", got, want)
		}
	})
}

func Test_getSortedKeysInt(t *testing.T) {
	args := make(map[string]int64)
	args["2"] = 1
	args["1"] = 2
	var want []string
	want = append(want, "1", "2")
	t.Run("sort test", func(t *testing.T) {
		if got := getSortedKeysInt(args); !reflect.DeepEqual(got, want) {
			t.Errorf("getSortedKeysInt() = %v, want %v", got, want)
		}
	})
}

func TestMemStorage_Clear(t *testing.T) {
	ms, err := NewMemStorage(restoreStorage, defFileName, saveInterval)
	assert.NoError(t, err, "create storage error")
	ms.Gauges["1"] = 1
	t.Run("clear storage", func(t *testing.T) {
		if err := ms.Clear(ctx); (err != nil) != false {
			t.Errorf("MemStorage.Clear() error = %v, wantErr %v", err, false)
		}
	})
}

func TestMemStorage_Save(t *testing.T) {
	msSuccess, err := NewMemStorage(restoreStorage, defFileName, saveInterval)
	assert.NoError(t, err, "success storage create error")
	msError, err := NewMemStorage(restoreStorage, "/_1.._1ww_", saveInterval)
	assert.NoError(t, err, "error storage create error")
	tests := []struct {
		name    string
		ms      *MemStorage
		wantErr bool
	}{
		{
			name:    "успешное сохранение",
			ms:      msSuccess,
			wantErr: false,
		},
		{
			name:    "ошибка сохранения",
			ms:      msError,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.ms.Save(); (err != nil) != tt.wantErr {
				t.Errorf("MemStorage.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemStorage_restore(t *testing.T) {
	t.Run("MemStorage restore test", func(t *testing.T) {
		mem, err := NewMemStorage(false, defFileName, 0)
		if err != nil {
			t.Errorf("create storage error: %v", err)
			return
		}
		err = mem.Update(ctx, cType, "test", "1")
		if err != nil {
			t.Errorf("add metric in storage error: %v", err)
			return
		}
		mem, err = NewMemStorage(true, defFileName, saveInterval)
		if err != nil {
			t.Errorf("restore storage error: %v", err)
			return
		}
		if len(mem.Counters) != 1 || len(mem.Gauges) > 0 {
			t.Errorf("restore count metrics size error: counter: %d, gauges: %d", len(mem.Counters), len(mem.Gauges))
			return
		}
	})
}
