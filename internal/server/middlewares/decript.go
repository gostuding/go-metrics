package middlewares

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type ErrorType int

const (
	contentEncoding           = "Content-Encoding"
	contentType               = "Content-Type"
	gzipString                = "gzip"
	applicationJSON           = "application/json"
	textHTML                  = "text/html"
	hashVarName               = "HashSHA256"
	acceptEncoding            = "Accept-Encoding"
	ReadBodyError   ErrorType = iota
	CloseBodyError
	GzipReaderError
)

func getError(t ErrorType, err error) error {
	switch t {
	case ReadBodyError:
		return fmt.Errorf("request body read error: %w", err)
	case CloseBodyError:
		return fmt.Errorf("close request body error: %w", err)
	case GzipReaderError:
		return fmt.Errorf("gzip reader error: %w", err)
	default:
		return err
	}
}

func decriptMessage(key *rsa.PrivateKey, msg []byte) (*[]byte, error) {
	size := key.PublicKey.Size()
	if len(msg)%size != 0 {
		return nil, errors.New("message length error")
	}
	hash := sha256.New()
	dectipted := make([]byte, 0)
	for i := 0; i < len(msg); i += size {
		data, err := rsa.DecryptOAEP(hash, nil, key, msg[i:i+size], []byte(""))
		if err != nil {
			return nil, fmt.Errorf("decription message error: %w", err)
		}
		dectipted = append(dectipted, data...)
	}
	return &dectipted, nil
}

// DecriptMiddleware decripts messages from clients.
func DecriptMiddleware(
	key *rsa.PrivateKey,
	logger *zap.SugaredLogger,
) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if key != nil && r.Method == http.MethodPost {
				data, err := io.ReadAll(r.Body)
				if err != nil {
					logger.Warnf(getError(ReadBodyError, err).Error())
					return
				}
				if err = r.Body.Close(); err != nil {
					logger.Warnf(getError(ReadBodyError, err).Error())
					return
				}
				body, err := decriptMessage(key, data)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					logger.Warnf("decript error: %w", err)
					return
				}
				r.Body = io.NopCloser(bytes.NewReader(*body))
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
