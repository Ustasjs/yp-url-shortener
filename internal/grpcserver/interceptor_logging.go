package grpcserver

import (
	"context"
	"time"

	"Ustasjs/yp-url-shortener/internal/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor is the gRPC version of logger.LoggerMiddleware. It logs the
// method, how long it took and the resulting status code.
func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		logger.Log.Info("grpc request",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", time.Since(start)),
			zap.String("code", status.Code(err).String()),
		)

		return resp, err
	}
}
