package interseptors

import (
	"bytes"
	"compress/gzip"
	"context"
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
		return nil, status.Error(codes.Canceled, makeError(NotByteError, nil).Error()) //nolint:wrapcheck //<-
	}
	reader, err := gzip.NewReader(bytes.NewReader(data.Metrics))
	if err != nil {
		return nil, status.Error(codes.Internal, makeError(GzipCreateError, err).Error()) //nolint:wrapcheck //<-
	}
	data.Metrics, err = io.ReadAll(reader)
	if err != nil {
		return nil, status.Error(codes.Internal, makeError(ReadError, err).Error()) //nolint:wrapcheck //<-
	}
	return handler(ctx, data)
}
