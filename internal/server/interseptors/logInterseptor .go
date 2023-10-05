package interseptors

import (
	"context"
	"time"

	pb "github.com/gostuding/go-metrics/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func LogInterceptor(logger *zap.SugaredLogger) grpc.UnaryServerInterceptor {
	var (
		urlString  = "url"
		respLogger = "Response logger"
		answer     = "answer"
	)
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		reqSize := 0
		if v, ok := req.(*pb.MetricsRequest); ok {
			reqSize = len(v.Metrics)
		}
		logger.Infow(
			"Request logger",
			urlString, info.FullMethod,
			"duration", time.Since(start),
			"size", reqSize,
		)
		resp, err := handler(ctx, req)
		if err != nil {
			logger.Infow(
				respLogger,
				urlString, info.FullMethod,
				"error", err.Error(),
			)
		} else {
			if v, ok := resp.(*pb.MetricsResponse); ok {
				logger.Infow(
					respLogger,
					urlString, info.FullMethod,
					answer, v.Error,
				)
			} else {
				logger.Infow(
					respLogger,
					urlString, info.FullMethod,
					answer, "undefined type",
				)
			}
		}
		return resp, err
	}
}
