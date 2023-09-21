package middlewares

import (
	"net/http"
	"testing"

	"github.com/gostuding/go-metrics/internal/server/middlewares/mocks"
)

func Test_hashWriter_Write(t *testing.T) {
	type fields struct {
		ResponseWriter http.ResponseWriter
		key            []byte
		body           []byte
	}
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "key nil",
			fields: fields{
				ResponseWriter: mocks.NewWMock(),
				key:            nil,
				body:           nil,
			},
			args:    args{[]byte("test")},
			want:    "",
			wantErr: false,
		},
		{
			name: "key default",
			fields: fields{
				ResponseWriter: mocks.NewWMock(),
				key:            []byte("default"),
				body:           nil,
			},
			args:    args{[]byte("test")},
			want:    "de79cc62d7da11c1f3049dbf73ba060497e3d4e7a07029fa6f48e75cfc681042",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r := &hashWriter{
				ResponseWriter: tt.fields.ResponseWriter,
				key:            tt.fields.key,
				body:           tt.fields.body,
			}
			_, err := r.Write(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("hashWriter.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.fields.ResponseWriter.Header().Get(hashVarName) != tt.want {
				t.Errorf("hashWriter.Write() = %v, want %v", tt.fields.ResponseWriter.Header().Get(hashVarName), tt.want)
			}
		})
	}
}

func Test_checkHash(t *testing.T) {
	type args struct {
		data []byte
		key  []byte
		hash string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Null data",
			args:    args{data: nil, key: nil, hash: ""},
			wantErr: false,
		},
		{
			name: "Test data",
			args: args{
				data: []byte("test"),
				key:  []byte("default"),
				hash: "de79cc62d7da11c1f3049dbf73ba060497e3d4e7a07029fa6f48e75cfc681042",
			},
			wantErr: false,
		},
		{
			name:    "Bad hash",
			args:    args{data: []byte("test"), key: []byte("default"), hash: "d1"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkHash(tt.args.data, tt.args.key, tt.args.hash); (err != nil) != tt.wantErr {
				t.Errorf("checkHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
