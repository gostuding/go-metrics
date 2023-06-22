package server

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// структура для подмены интерфейса http.ResponseWriter
type myLogWriter struct {
	http.ResponseWriter     // интерфейс http.ResponseWriter
	status              int // статус ответа
	size                int // размер ответа
}

func newLogWriter(w http.ResponseWriter) *myLogWriter {
	return &myLogWriter{ResponseWriter: w, status: 0, size: 0}
}

func (r *myLogWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b) // запись данных через стандартный ResponseWriter
	r.size += size
	return size, err // получаем размер записанных данных
}

func (r *myLogWriter) WriteHeader(statusCode int) {
	r.status = statusCode // запоминаем код статуса
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *myLogWriter) Header() http.Header {
	return r.ResponseWriter.Header()
}

// структура для создания middleware gzip
type myGzipWriter struct {
	http.ResponseWriter // интерфейс http.ResponseWriter
	isWriting           bool
	logger              *zap.SugaredLogger
}

func newGzipWriter(r http.ResponseWriter, logger *zap.SugaredLogger) *myGzipWriter {
	return &myGzipWriter{ResponseWriter: r, isWriting: false, logger: logger}
}

func (r *myGzipWriter) Write(b []byte) (int, error) {
	if !r.isWriting && r.Header().Get("Content-Encoding") == "gzip" {
		r.isWriting = true
		compressor := gzip.NewWriter(r)
		size, err := compressor.Write(b)
		if err != nil {
			r.logger.Warnf("compress respons body error: %w \n", err)
			return 0, err
		}
		if err = compressor.Close(); err != nil {
			r.logger.Warnf("compress close error: %w \n", err)
			return 0, err
		}
		r.isWriting = false
		return size, err
	}
	return r.ResponseWriter.Write(b)
}

func (r *myGzipWriter) WriteHeader(statusCode int) {
	contentType := r.Header().Get("Content-Type") == "application/json" || r.Header().Get("Content-Type") == "text/html"
	if statusCode == 200 && contentType { // проверяем здесь, т.к. после этого заголовок уже не изменить
		r.Header().Set("Content-Encoding", "gzip")
	}
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *myGzipWriter) Header() http.Header {
	return r.ResponseWriter.Header()
}

// структура для чтения данных из Body через gzip
type gzipReader struct {
	r    io.ReadCloser
	gzip *gzip.Reader
}

func newGzipReader(r io.ReadCloser) (*gzipReader, error) {
	reader, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &gzipReader{r: r, gzip: reader}, nil
}

func (c gzipReader) Read(p []byte) (n int, err error) {
	return c.gzip.Read(p) // чтени данных и их распаковка
}

func (c *gzipReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.gzip.Close()
}

//----------------------------------------------------------------------

func gzipMiddleware(logger *zap.SugaredLogger) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			//------------------------------------------
			if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") { // если стоит gzip, то надо распаковывать
				cr, err := newGzipReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logger.Warnf("gzip reader create error: %w", err)
					return
				}
				r.Body = cr // подмена интерфейса для чтения данных запроса
				defer cr.Close()
			}
			//-----------------------------------------
			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				//выполнение запроса с нашим ResponseWriter
				next.ServeHTTP(newGzipWriter(w, logger), r) // внедряем реализацию http.ResponseWriter
			} else {
				next.ServeHTTP(w, r) // выполняем без изменения
			}
		}
		return http.HandlerFunc(fn)
	}
}

func loggerMiddleware(logger *zap.SugaredLogger) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			rWriter := newLogWriter(w)
			//выполнение запроса с нашим ResponseWriter
			start := time.Now()
			next.ServeHTTP(rWriter, r) // внедряем реализацию http.ResponseWriter
			// логирование запроса
			logger.Infow(
				"Server logger",
				"type", "request",
				"uri", r.RequestURI,
				"method", r.Method,
				"duration", time.Since(start),
			)
			// логирование ответа
			defer logger.Infow(
				"Server logger",
				"type", "responce",
				"uri", r.RequestURI,
				"status", rWriter.status,
				"size", rWriter.size,
			)
		}
		return http.HandlerFunc(fn)
	}
}

// ----------------------------------------------------------------------

// структура для проверки hash в заголовке
type hashReader struct {
	reader *bytes.Reader
}

func newHashReader(r *http.Request, key []byte) (*hashReader, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("read request body error: %w", err)
	}
	err = r.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("close request body error: %w", err)
	}
	if len(data) > 0 {
		headerHash := r.Header.Get("HashSHA256")
		if headerHash == "" {
			return nil, fmt.Errorf("header HashSHA256 undefined, data: %x", data)
		}
		h := hmac.New(sha256.New, key)
		_, err = h.Write(data)
		if err != nil {
			return nil, fmt.Errorf("write hash summ error: %w", err)
		}
		hashSum := fmt.Sprintf("%x", h.Sum(nil))
		if headerHash != hashSum {
			return nil, fmt.Errorf("request body hash check error. hash must be: %s, get: %s", hashSum, headerHash)
		}
	}
	return &hashReader{reader: bytes.NewReader(data)}, nil
}

func (c *hashReader) Read(p []byte) (n int, err error) {
	return c.reader.Read(p) // чтени данных и их распаковка
}

func (c *hashReader) Close() error {
	return nil
}

func hashCheckMiddleware(key string, logger *zap.SugaredLogger) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if key != "" && r.Method == "POST" {
				reader, err := newHashReader(r, []byte(key))
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					logger.Warnf("hash (%s) checker error: %w", err)
					return
				}
				r.Body = reader
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
