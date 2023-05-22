package server

import (
	"net/http"
	"time"
)

// структура для подмены интерфейса http.ResponseWriter
type myRWriter struct {
	http.ResponseWriter     // интерфейс http.ResponseWriter
	status              int // статус ответа
	size                int // размер ответа
	body                []byte
}

func (r *myRWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b) // запись данных через стандартный ResponseWriter
	r.size += size                         // получаем размер записанных данных
	r.body = b
	return size, err
}

func (r *myRWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.status = statusCode // захватываем код статуса
	r.ResponseWriter.WriteHeader(statusCode)
}

func serverMiddleware(h http.Handler) http.Handler {
	wrapper := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responceWriter := &myRWriter{ResponseWriter: w, status: 0, size: 0}
		//выполнение запроса с нашим ResponseWriter
		h.ServeHTTP(responceWriter, r) // внедряем реализацию http.ResponseWriter
		// логирование запроса
		requestLog(r.RequestURI, r.Method, time.Since(start))
		// логирование ответа
		defer responseLog(r.RequestURI+string(responceWriter.body), responceWriter.status, responceWriter.size)
	}
	return http.HandlerFunc(wrapper)
}
