package server

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"
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
}

func newGzipWriter(r http.ResponseWriter) *myGzipWriter {
	return &myGzipWriter{ResponseWriter: r, isWriting: false}
}

func (r *myGzipWriter) Write(b []byte) (int, error) {
	if !r.isWriting && r.Header().Get("Content-Encoding") == "gzip" {
		r.isWriting = true
		compressor := gzip.NewWriter(r)
		size, err := compressor.Write(b)
		if err != nil {
			Logger.Warnf("compress respons body error: %w \n", err)
			return 0, err
		}
		if err = compressor.Close(); err != nil {
			Logger.Warnf("compress close error: %w \n", err)
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

func gzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//------------------------------------------
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") { // если стоит gzip, то надо распаковывать
			cr, err := newGzipReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				Logger.Warnf("gzip reader create error: %w", err)
				return
			}
			r.Body = cr // подмена интерфейса для чтения данных запроса
			defer cr.Close()
		}
		//-----------------------------------------
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			//выполнение запроса с нашим ResponseWriter
			h.ServeHTTP(newGzipWriter(w), r) // внедряем реализацию http.ResponseWriter
		} else {
			h.ServeHTTP(w, r) // выполняем без изменения
		}
	})
}

func loggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rWriter := newLogWriter(w)
		//выполнение запроса с нашим ResponseWriter
		h.ServeHTTP(rWriter, r) // внедряем реализацию http.ResponseWriter
		// логирование запроса
		Logger.Infow(
			"Server logger",
			"type", "request",
			"uri", r.RequestURI,
			"method", r.Method,
			"duration", time.Since(start),
		)
		// логирование ответа
		defer Logger.Infow(
			"Server logger",
			"type", "responce",
			"uri", r.RequestURI,
			"status", rWriter.status,
			"size", rWriter.size,
		)
	})
}
