package middlewares

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_myGzipWriter_Write(t *testing.T) {
	data := []byte("test data write")
	mock := newWMock()
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err, "create logger error")
	writer := NewGzipWriter(mock, logger.Sugar())
	writer.Header().Set(contentEncoding, gzipString)
	_, err = writer.Write(data)
	assert.NoError(t, err, "write data error")
	if reflect.DeepEqual(mock.Body, data) {
		t.Errorf("write gzip error, data is equal to body")
	}
}

func Test_myGzipWriter_WriteHeader(t *testing.T) {
	mock := newWMock()
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err, "create logger error")
	writer := NewGzipWriter(mock, logger.Sugar())
	writer.Header().Add(contentType, applicationJSON)
	writer.WriteHeader(http.StatusOK)
	if writer.Header().Get(contentEncoding) != gzipString {
		t.Errorf("gzip add content type in header error")
	}
	writer.Header().Set(contentEncoding, "")
	writer.WriteHeader(http.StatusBadRequest)
	if writer.Header().Get(contentEncoding) == gzipString {
		t.Errorf("gzip incorret add content type in header")
	}
}

func Test_myGzipWriter_Header(t *testing.T) {
	mock := newWMock()
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err, "create logger error")
	writer := NewGzipWriter(mock, logger.Sugar())
	writer.Header().Add(contentType, applicationJSON)
	if writer.Header().Get(contentType) != applicationJSON {
		t.Errorf("gzip header get error")
	}
}

func Test_gzipReader_Read(t *testing.T) {
	m := newWMock()
	logger, err := zap.NewDevelopment()
	write := []byte("data")
	if err != nil {
		fmt.Printf("create logger errror: %v", err)
		return
	}
	w := NewGzipWriter(m, logger.Sugar())
	w.Header().Set(contentType, textHTML)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(write)
	if err != nil {
		t.Errorf("write gzip data error %v", err)
		return
	}
	r, err := NewGzipReader(m)
	if err != nil {
		fmt.Printf("create reader error: %v", err)
		return
	}
	data, err := io.ReadAll(r)
	if err != nil {
		t.Errorf("read data error %v", err)
		return
	}
	if !reflect.DeepEqual(data, write) {
		t.Errorf("read data error. Read (%s) not equal to: %s", string(data), string(write))
		return
	}
}

func Test_gzipReader_Close(t *testing.T) {
	m := newWMock()
	r, err := NewGzipReader(m)
	if err != nil {
		fmt.Printf("gzipReader.Close() create reader error: %v", err)
		return
	}
	if err := r.Close(); err != nil {
		t.Errorf("gzipReader.Close() error: %v", err)
	}
}
