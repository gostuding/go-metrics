package middlewares

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// Internal types.
type (
	// Struct for write gzip data in response.
	myGzipWriter struct {
		http.ResponseWriter
		logger    *zap.SugaredLogger
		isWriting bool
	}
	// Struct for read gzip data from request.
	gzipReader struct {
		r    io.ReadCloser
		gzip *gzip.Reader
	}
)

// NewGzipWriter creates new writer.
func NewGzipWriter(r http.ResponseWriter, logger *zap.SugaredLogger) *myGzipWriter {
	return &myGzipWriter{ResponseWriter: r, isWriting: false, logger: logger}
}

// Write data in response by gzip writer.
func (r *myGzipWriter) Write(b []byte) (int, error) {
	if !r.isWriting && r.Header().Get(contentEncoding) == gzipString {
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
		return size, nil
	}
	return r.ResponseWriter.Write(b) //nolint:wrapcheck //<-senselessly
}

// WriteHeader checks Content-Type and sets Content-Encoding data.
func (r *myGzipWriter) WriteHeader(statusCode int) {
	contentType := r.Header().Get(contentType) == applicationJSON || r.Header().Get(contentType) == textHTML
	if statusCode == http.StatusOK && contentType {
		r.Header().Set(contentEncoding, gzipString)
	}
	r.ResponseWriter.WriteHeader(statusCode)
}

// Header returns response headers map.
func (r *myGzipWriter) Header() http.Header {
	return r.ResponseWriter.Header()
}

// NewGzipReader creates new gzip reader.
func NewGzipReader(r io.ReadCloser) (*gzipReader, error) {
	reader, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("new gzip reader create error: %w", err)
	}
	return &gzipReader{r: r, gzip: reader}, nil
}

// Read and ungzip data.
func (c gzipReader) Read(p []byte) (n int, err error) {
	size, err := c.gzip.Read(p)
	if errors.Is(err, io.EOF) {
		return size, err //nolint:wrapcheck //<-senselessly
	}
	if err != nil {
		return 0, fmt.Errorf("gzip read error: %w", err)
	}
	return size, nil
}

// Close gzip reader.
func (c *gzipReader) Close() error {
	if err := c.r.Close(); err != nil {
		return fmt.Errorf("close request interface error: %w", err)
	}
	if err := c.gzip.Close(); err != nil {
		return fmt.Errorf("gzip reader close error: %w", err)
	}
	return nil
}

// GzipMiddleware usefull for gzip support enable in server.
// Is using when request contains Content-Encoding: gzip in Headers.
func GzipMiddleware(logger *zap.SugaredLogger) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.Header.Get(contentEncoding), gzipString) {
				cr, err := NewGzipReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logger.Warnf(getError(GzipReaderError, err).Error())
					return
				}
				r.Body = cr
				defer cr.Close() //nolint:errcheck //<-senselessly
			}
			if strings.Contains(r.Header.Get(acceptEncoding), gzipString) {
				next.ServeHTTP(NewGzipWriter(w, logger), r)
			} else {
				next.ServeHTTP(w, r)
			}
		}
		return http.HandlerFunc(fn)
	}
}
