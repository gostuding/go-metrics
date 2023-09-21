package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type hashWriter struct {
	http.ResponseWriter
	key  []byte
	body []byte
}

// NewHashWriter creates new hash writer.
func NewHashWriter(r http.ResponseWriter, key []byte) *hashWriter {
	return &hashWriter{ResponseWriter: r, key: key, body: nil}
}

// Write sets hash summ for bosy in response header.
func (r *hashWriter) Write(b []byte) (int, error) {
	if r.key != nil {
		data := append(r.body[:], b[:]...) //nolint:gocritic //<-should be
		h := hmac.New(sha256.New, r.key)
		_, err := h.Write(data)
		if err != nil {
			return 0, fmt.Errorf("write body hash summ error: %w", err)
		}
		r.body = data
		r.Header().Set(hashVarName, hex.EncodeToString(h.Sum(nil)))
	}
	size, err := r.ResponseWriter.Write(b)
	if err != nil {
		return 0, fmt.Errorf("response write error: %w", err)
	}
	return size, nil
}

func checkHash(data, key []byte, hash string) error {
	if len(data) > 0 && hash != "" {
		h := hmac.New(sha256.New, key)
		_, err := h.Write(data)
		if err != nil {
			return fmt.Errorf("write hash summ error: %w", err)
		}
		if hash != hex.EncodeToString(h.Sum(nil)) {
			return fmt.Errorf("incorrect hash summ: %s", hash)
		}
	}
	return nil
}

// HashCheckMiddleware checks hash summ for request body.
// Hash must be in request Header: "HashSHA256": "...hassumm...".
func HashCheckMiddleware(
	hashKey []byte,
	logger *zap.SugaredLogger,
) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if len(hashKey) > 0 && r.Method == http.MethodPost {
				data, err := io.ReadAll(r.Body)
				if err != nil {
					logger.Warnf(getError(ReadBodyError, err).Error())
					return
				}
				if err = r.Body.Close(); err != nil {
					logger.Warnf(getError(CloseBodyError, err).Error())
					return
				}
				err = checkHash(data, hashKey, r.Header.Get(hashVarName))
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					logger.Warnf("hash checker error: %w", err)
					return
				}
				r.Body = io.NopCloser(bytes.NewReader(data))
				next.ServeHTTP(NewHashWriter(w, hashKey), r)
			} else {
				next.ServeHTTP(w, r)
			}
		}
		return http.HandlerFunc(fn)
	}
}
