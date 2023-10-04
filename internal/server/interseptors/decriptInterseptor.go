package interseptors

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"fmt"

	pb "github.com/gostuding/go-metrics/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrType int

const (
	DecriptError ErrType = iota
	NotByteError
	GzipCreateError
	ReadError
	HashIncorrectError
	WriteError
)

func makeError(t ErrType, args ...any) error {
	switch t {
	case DecriptError:
		return fmt.Errorf("decript error: %v", args...)
	case NotByteError:
		return errors.New("req is not bytes")
	case GzipCreateError:
		return fmt.Errorf("new gzip reader create error: %v", args...)
	case ReadError:
		return fmt.Errorf("read error: %v", args)
	case HashIncorrectError:
		return fmt.Errorf("incorrect hash summ: %s", args...)
	case WriteError:
		return fmt.Errorf("write summ error: %w", args...)
	default:
		return errors.New("undefined error type")
	}
}

func decript(key *rsa.PrivateKey, msg []byte) ([]byte, error) {
	size := key.PublicKey.Size()
	if len(msg)%size != 0 {
		return nil, errors.New("message length error")
	}
	hash := sha256.New()
	dectipted := make([]byte, 0)
	for i := 0; i < len(msg); i += size {
		data, err := rsa.DecryptOAEP(hash, nil, key, msg[i:i+size], []byte(""))
		if err != nil {
			return nil, status.Error(codes.Internal, makeError(DecriptError, err).Error()) //nolint:wrapcheck //<-
		}
		dectipted = append(dectipted, data...)
	}
	return dectipted, nil
}

func DecriptInterceptor(key *rsa.PrivateKey) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if key == nil {
			return handler(ctx, req)
		}
		var err error
		data, ok := req.(*pb.MetricsRequest)
		if !ok {
			return nil, status.Error(codes.FailedPrecondition, makeError(NotByteError, nil).Error()) //nolint:wrapcheck //<-
		}
		data.Metrics, err = decript(key, data.Metrics)
		if err != nil {
			return nil, status.Error(codes.FailedPrecondition, makeError(DecriptError, err).Error()) //nolint:wrapcheck //<-
		}
		return handler(ctx, data)
	}
}
