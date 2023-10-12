package commons

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func MiddlewareLogger(logger *slog.Logger, filteredServers ...any) grpc.UnaryServerInterceptor {
	filterLookup := make(map[any]struct{})
	for _, srv := range filteredServers {
		filterLookup[srv] = struct{}{}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if _, ok := filterLookup[info.Server]; ok {
			return handler(ctx, req)
		}
		start := time.Now()
		data, err := handler(ctx, req)
		respStatus := status.New(codes.OK, "ok")
		if err != nil {
			respStatus, _ = status.FromError(err)
		}
		logger.Info(
			"handle rpc",
			"method", info.FullMethod,
			"code", respStatus.Code().String(),
			"duration", time.Since(start),
		)
		return data, err
	}
}

func MiddlewareRecover(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(
					"recovered from panic",
					"error", r,
					"stack", string(debug.Stack()),
				)
				err = status.Errorf(codes.Internal, "an unexpected error occurred")
			}
		}()
		return handler(ctx, req)
	}
}
