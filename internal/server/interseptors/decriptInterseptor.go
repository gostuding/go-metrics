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
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

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
			return nil, status.Error(codes.Internal, fmt.Sprintf("decript error: %v", err))
		}
		dectipted = append(dectipted, data...)
	}
	return dectipted, nil
}

func DecriptInterceptor(key *rsa.PrivateKey) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			values := md.Get("rsa")
			if len(values) == 0 {
				return handler(ctx, req)
			}
		}
		var err error
		data, ok := req.(*pb.MetricsRequest)
		if !ok {
			return nil, status.Error(codes.FailedPrecondition, "req is not bytes")
		}
		data.Metrics, err = decript(key, data.Metrics)
		if err != nil {
			return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("decript error: %v", err))
		}
		return handler(ctx, data)
	}
}
