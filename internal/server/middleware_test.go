package server

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type rwMock struct {
	Head http.Header
}

func (r *rwMock) Write(b []byte) (int, error) {
	return len(b), nil
}

func (r *rwMock) WriteHeader(statusCode int) {

}
func (r *rwMock) Header() http.Header {
	return r.Head

}

func Test_myLogWriter_Write(t *testing.T) {
	b := []byte("123")
	mock := rwMock{}
	logWriter := newLogWriter(&mock)
	count, err := logWriter.Write(b)
	assert.NoError(t, err, "Write error")
	if count != len(b) {
		t.Errorf("myLogWriter_Write size count = %d, want %d", count, len(b))
	}
}

func Test_myLogWriter_WriteHeader(t *testing.T) {
	mock := rwMock{Head: make(http.Header)}
	logWriter := newLogWriter(&mock)
	logWriter.WriteHeader(http.StatusAlreadyReported)
	if logWriter.status != http.StatusAlreadyReported {
		t.Errorf("myLogWriter_WriteHeader status = %d, want %d", logWriter.status, http.StatusAlreadyReported)
	}
}

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
				ResponseWriter: &rwMock{Head: make(http.Header)},
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
				ResponseWriter: &rwMock{Head: make(http.Header)},
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
