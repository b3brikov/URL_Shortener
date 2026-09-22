package interceptor

import (
	"context"
	"log/slog"
	"runtime/debug"

	"google.golang.org/grpc"
)

func RecoverInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		defer func() {
			if p := recover(); p != nil {
				logger.Error("panic recovered",
					slog.Any("panic", p),
					slog.String("stack", string(debug.Stack())))
			}
		}()
		return handler(ctx, req)
	}
}
