package middlewares

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type myLogWriter struct {
	http.ResponseWriter
	Status int
	Size   int
}

func NewLogWriter(w http.ResponseWriter) *myLogWriter {
	return &myLogWriter{ResponseWriter: w, Status: 0, Size: 0}
}

// Write sets b size in myLogWriter struct.
func (r *myLogWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.Size += size
	return size, err //nolint:wrapcheck //<-not need
}

// WriteHeader sets statusCode in myLogWriter struct.
func (r *myLogWriter) WriteHeader(statusCode int) {
	r.Status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *myLogWriter) Header() http.Header {
	return r.ResponseWriter.Header()
}

// LoggerMiddleware writes in logger status code and size of responce data.
func LoggerMiddleware(logger *zap.SugaredLogger) func(h http.Handler) http.Handler {
	var (
		typeString = "type"
		urlString  = "url"
	)
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			rWriter := NewLogWriter(w)
			start := time.Now()
			next.ServeHTTP(rWriter, r)
			logger.Infow(
				"Request logger",
				typeString, "request",
				urlString, r.RequestURI,
				"method", r.Method,
				"duration", time.Since(start),
			)
			defer logger.Infow(
				"Response logger",
				typeString, "responce",
				urlString, r.RequestURI,
				"status", rWriter.Status,
				"size", rWriter.Size,
			)
		}
		return http.HandlerFunc(fn)
	}
}
