package server

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

// структура для подмены интерфейса http.ResponseWriter
type myRWriter struct {
	http.ResponseWriter     // интерфейс http.ResponseWriter
	status              int // статус ответа
	size                int // размер ответа
	isGzipSupport       bool
	gzResponseWriter    io.Writer
}

func (r *myRWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter или сжимая его
	size, err := 0, errors.New("")
	if r.isGzipSupport && r.Header().Get("Content-Encoding") == "gzip" {
		r.isGzipSupport = false // для избежания зацикливания рекурсии
		compressor := gzip.NewWriter(r)
		size, err = compressor.Write(b)
		if err != nil {
			Logger.Warnf("compress responыe body error: %v \n", err)
			return 0, err
		}
		if err = compressor.Close(); err != nil {
			Logger.Warnf("compress close error: %v \n", err)
			return 0, err
		}
	} else {
		size, err = r.ResponseWriter.Write(b) // запись данных через стандартный ResponseWriter
		r.size += size                        // получаем размер записанных данных
	}
	return size, err
}

func (r *myRWriter) WriteHeader(statusCode int) {
	contentType := r.Header().Get("Content-Type") == "application/json" || r.Header().Get("Content-Type") == "text/html"
	if statusCode < 300 && r.isGzipSupport && contentType { // проверяем здесь, т.к. после этого заголовок уже не изменить
		r.Header().Set("Content-Encoding", "gzip")
	}
	r.status = statusCode // запоминаем код статуса
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *myRWriter) Header() http.Header {
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

func loggerMiddleware(h http.Handler) http.Handler {
	wrapper := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responceWriter := &myRWriter{
			ResponseWriter:   w,
			status:           0,
			size:             0,
			isGzipSupport:    strings.Contains(r.Header.Get("Accept-Encoding"), "gzip"),
			gzResponseWriter: gzip.NewWriter(w),
		}
		//------------------------------------------
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") { // если стоит gzip, то надо распаковывать
			cr, err := newGzipReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				Logger.Warnf("gzip reader create error: %s", err)
				return
			}
			Logger.Infoln("read data with gzip support")
			r.Body = cr // подмена интерфейса для чтения данных запроса
			defer cr.Close()
		}

		//-----------------------------------------
		//выполнение запроса с нашим ResponseWriter
		h.ServeHTTP(responceWriter, r) // внедряем реализацию http.ResponseWriter
		// логирование запроса
		requestLog(r.RequestURI, r.Method, time.Since(start))
		// логирование ответа
		defer responseLog(r.RequestURI, responceWriter.status, responceWriter.size)
	}
	return http.HandlerFunc(wrapper)
}

// type gzipWriter struct {
// 	response http.ResponseWriter
// 	gzip     *gzip.Writer
// }

// func newGzipWriter(w http.ResponseWriter) *gzipWriter {
// 	return &gzipWriter{response: w, gzip: gzip.NewWriter(w)}
// }

// func (c *gzipWriter) Header() http.Header {
// 	return c.response.Header()
// }

// func (c *gzipWriter) Write(p []byte) (int, error) {
// 	return c.gzip.Write(p)
// }

// func (c *gzipWriter) WriteHeader(statusCode int) {
// 	c.response.WriteHeader(statusCode)
// }

// func (c *gzipWriter) Close() error {
// 	return c.gzip.Close()
// }

// func gzipMiddleware(h http.Handler) http.Handler {
// 	wrapper := func(w http.ResponseWriter, r *http.Request) {
// 		response := w
// 		// проверка на необходимость распаковки тела запроса
// 		contentEncoding := r.Header.Get("Content-Encoding") // если стоит gzip, то надо распаковывать
// 		sendsGzip := strings.Contains(contentEncoding, "gzip")
// 		if sendsGzip {
// 			cr, err := newGzipReader(r.Body)
// 			if err != nil {
// 				w.WriteHeader(http.StatusInternalServerError)
// 				Logger.Warnf("gzip reader create error: %s", err)
// 				return
// 			}
// 			Logger.Infoln("read data with gzip support")
// 			r.Body = cr // подмена интерфейса для чтения данных запроса
// 			defer cr.Close()
// 		}
// 		// подмена ResponseWriter на gzipWriter при поддержке запакованных данных
// 		supportsGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
// 		supportsGzip = (strings.Contains(r.Header.Get("Content-Type"), "application/json") || strings.Contains(r.Header.Get("Content-Type"), "text/html"))
// 		if supportsGzip {
// 			gzip := newGzipWriter(w)
// 			response = gzip // подмена интерфейса для записи ответа
// 			defer gzip.Close()
// 			response.Header().Set("Content-Encoding", "gzip")
// 			Logger.Infoln("write data by gzip")
// 		} else {
// 			Logger.Infoln("write data without gzip support: ", r.Header.Get("Accept-Encoding"))
// 		}

// 		h.ServeHTTP(response, r)
// 	}
// 	return http.HandlerFunc(wrapper)
// }
