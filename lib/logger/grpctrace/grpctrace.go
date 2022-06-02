// grpctrace provides a simple Interceptor for emitting the contents of a request and the time taken to process it
package grpctrace

import (
	"context"
	"time"

	"biller/lib/errorhandler"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// UnaryInterceptor wraps a unary GRPC handler to log calls and any resulting server errors
func UnaryInterceptor(log *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()
		res, err := handler(ctx, req)
		timeTaken := time.Since(start)

		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.Duration("latency", timeTaken),
		}

		if err != nil {
			fields = append(fields, zap.Error(err))
			if !errorhandler.IsClientError(err) {
				log.Error("request error", fields...)
				return res, err
			}
		}

		log.Debug("response", fields...)

		return res, err
	}

}
