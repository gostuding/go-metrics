package interseptors

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	pb "github.com/gostuding/go-metrics/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	hashVarName = "HashSHA256" // Header name for hash check.
)

func checkHash(data, key []byte, hash string) error {
	if len(data) > 0 && hash != "" {
		h := hmac.New(sha256.New, key)
		_, err := h.Write(data)
		if err != nil {
			return makeError(WriteError, err)
		}
		if hash != hex.EncodeToString(h.Sum(nil)) {
			return makeError(HashIncorrectError, hash)
		}
	}
	return nil
}

func HashInterceptor(key []byte) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if key == nil {
			return handler(ctx, req)
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Internal, "get value from context error") //nolint:wrapcheck //<-
		}
		values := md.Get(hashVarName)
		if len(values) == 0 {
			return nil, status.Error(codes.InvalidArgument, "hash undefined") //nolint:wrapcheck //<-
		}
		var err error
		data, ok := req.(*pb.MetricsRequest)
		if !ok {
			return nil, status.Error(codes.Canceled, makeError(NotByteError, nil).Error()) //nolint:wrapcheck //<-
		}
		err = checkHash(data.Metrics, key, values[0])
		if err != nil {
			return nil, status.Error(codes.Aborted, "incorrect hash") //nolint:wrapcheck //<-
		}
		return handler(ctx, data)
	}
}
