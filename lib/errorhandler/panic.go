package errorhandler

import (
	"context"
	"runtime/debug"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func PanicUnaryInterceptor(log *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		panicked := true

		defer func() {
			if r := recover(); r != nil || panicked {
				message := "handler panic recovery"
				fields := []zap.Field{
					zap.StackSkip("stacktrace", 4),
				}
				if err, ok := r.(error); ok {
					message = err.Error()
				} else {
					fields = append(fields, zap.Any("panic", r))
				}
				log.Error(message, fields...)
				err = status.Errorf(codes.Internal, "%v", r)
			}
		}()

		resp, err := handler(ctx, req)
		panicked = false
		return resp, err
	}
}

func PanicStreamInterceptor(log *zap.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		panicked := true

		defer func() {
			if r := recover(); r != nil || panicked {
				debug.PrintStack()
				log.Error(string(debug.Stack()), zap.Error(err))
				err = status.Errorf(codes.Internal, "%v", r)
			}
		}()

		err = handler(srv, stream)
		panicked = false
		return err
	}
}

// ClientErrors are errors we show to the user and not report. Other errors will be masked as Unknown(2)
var ClientErrors = map[codes.Code]bool{
	codes.Canceled:           true,
	codes.InvalidArgument:    true,
	codes.DeadlineExceeded:   true,
	codes.NotFound:           true,
	codes.AlreadyExists:      true,
	codes.PermissionDenied:   true,
	codes.ResourceExhausted:  true,
	codes.FailedPrecondition: true,
	codes.Unauthenticated:    true,
}

func IsClientError(err error) bool {
	st, _ := status.FromError(err)

	return ClientErrors[st.Code()]
}
