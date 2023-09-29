package interseptors

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"

	pb "github.com/gostuding/go-metrics/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func GzipInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("gzip")
		if len(values) == 0 {
			return handler(ctx, req)
		}
	}
	data, ok := req.(*pb.MetricsRequest)
	if !ok {
		return nil, status.Error(codes.FailedPrecondition, "req is not bytes")
	}
	reader, err := gzip.NewReader(bytes.NewReader(data.Metrics))
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("new gzip reader create error: %v", err))
	}
	data.Metrics, err = io.ReadAll(reader)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("read error: %v", err))
	}
	return handler(ctx, data)
}
