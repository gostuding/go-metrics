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

type myLogWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func newLogWriter(w http.ResponseWriter) *myLogWriter {
	return &myLogWriter{ResponseWriter: w, status: 0, size: 0}
}

func (r *myLogWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.size += size
	return size, err //nolint:wrapcheck //<-not need
}

func (r *myLogWriter) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *myLogWriter) Header() http.Header {
	return r.ResponseWriter.Header()
}

type myGzipWriter struct {
	http.ResponseWriter
	logger    *zap.SugaredLogger
	isWriting bool
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
			return 0, fmt.Errorf("compress respons body error: %w", err)
		}
		if err = compressor.Close(); err != nil {
			return 0, fmt.Errorf("compress close error: %w", err)
		}
		r.isWriting = false
		return size, err
	}
	return r.ResponseWriter.Write(b)
}

func (r *myGzipWriter) WriteHeader(statusCode int) {
	contentType := r.Header().Get("Content-Type") == "application/json" || r.Header().Get("Content-Type") == "text/html"
	if statusCode == 200 && contentType {
		r.Header().Set("Content-Encoding", "gzip")
	}
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *myGzipWriter) Header() http.Header {
	return r.ResponseWriter.Header()
}

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
	return c.gzip.Read(p)
}

func (c *gzipReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.gzip.Close()
}

func gzipMiddleware(logger *zap.SugaredLogger) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
				cr, err := newGzipReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logger.Warnf("gzip reader create error: %w", err)
					return
				}
				r.Body = cr
				defer cr.Close()
			}
			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(newGzipWriter(w, logger), r)
			} else {
				next.ServeHTTP(w, r)
			}
		}
		return http.HandlerFunc(fn)
	}
}

func loggerMiddleware(logger *zap.SugaredLogger) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			rWriter := newLogWriter(w)
			start := time.Now()
			next.ServeHTTP(rWriter, r)
			logger.Infow(
				"Server logger",
				"type", "request",
				"uri", r.RequestURI,
				"method", r.Method,
				"duration", time.Since(start),
			)
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

type hashWriter struct {
	http.ResponseWriter
	key  []byte
	body []byte
}

func newHashWriter(r http.ResponseWriter, key []byte, wh bool) *hashWriter {
	return &hashWriter{ResponseWriter: r, key: key, body: nil}
}

func (r *hashWriter) Write(b []byte) (int, error) {
	if r.key != nil {
		data := append(r.body[:], b[:]...)
		h := hmac.New(sha256.New, r.key)
		_, err := h.Write(data)
		if err != nil {
			return 0, fmt.Errorf("write body hash summ error: %w", err)
		}
		r.body = data
		r.Header().Set("HashSHA256", fmt.Sprintf("%x", h.Sum(nil)))
	}
	return r.ResponseWriter.Write(b)
}

func checkBodyHash(r *http.Request, key []byte) (*[]byte, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("read request body error: %w", err)
	}
	defer r.Body.Close()
	if len(data) > 0 {
		headerHash := r.Header.Get("HashSHA256")
		if headerHash != "" {
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
	}
	return &data, nil
}

func hashCheckMiddleware(
	key []byte,
	logger *zap.SugaredLogger,
	writeHeaderStatus bool,
) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if len(key) > 0 && r.Method == "POST" {
				body, err := checkBodyHash(r, key)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					logger.Warnf("hash checker error: %w", err)
					return
				}
				r.Body = io.NopCloser(bytes.NewReader(*body))
				next.ServeHTTP(newHashWriter(w, key, writeHeaderStatus), r)
			} else {
				next.ServeHTTP(w, r)
			}
		}
		return http.HandlerFunc(fn)
	}
}
